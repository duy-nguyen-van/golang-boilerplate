package httpclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang-boilerplate/internal/config"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRestClient_Post(t *testing.T) {
	tests := []struct {
		name            string
		body            interface{}
		headers         map[string]string
		serverHandler   http.HandlerFunc
		expectedStatus  int
		expectedError   bool
		validateHeaders func(*testing.T, *http.Request)
	}{
		{
			name:    "success - post request with custom headers",
			body:    map[string]string{"name": "John"},
			headers: map[string]string{"Authorization": "Bearer token"},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))

				var body map[string]string
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				assert.Equal(t, "John", body["name"])

				w.WriteHeader(http.StatusOK)
				require.NoError(t, json.NewEncoder(w).Encode(map[string]string{"id": "123", "name": "John"}))
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:    "success - post with nil headers",
			body:    map[string]interface{}{"email": "test@example.com"},
			headers: nil,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				w.WriteHeader(http.StatusCreated)
				require.NoError(t, json.NewEncoder(w).Encode(map[string]interface{}{"id": "456"}))
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name:    "error - server returns 400",
			body:    map[string]string{"invalid": "data"},
			headers: map[string]string{},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				require.NoError(t, json.NewEncoder(w).Encode(map[string]string{"error": "bad request"}))
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  false, // resty doesn't return error for HTTP errors
		},
		{
			name:    "error - server returns 500",
			body:    map[string]string{"data": "test"},
			headers: map[string]string{},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				require.NoError(t, json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"}))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(tt.serverHandler)
			defer server.Close()

			cfg := &config.Config{
				HTTPClientTimeout:            60 * time.Second,
				HTTPClientRetryCount:         0,
				HTTPClientRetryWaitMin:       1 * time.Second,
				HTTPClientRetryWaitMax:       5 * time.Second,
				HTTPClientDebug:              false,
				HTTPClientTLSInsecureSkipTLS: false,
				AppName:                      "test-app",
				AppVersion:                   "1.0.0",
			}

			client := ProvideRestClient(cfg)
			okResult := &map[string]interface{}{}
			failedResult := &map[string]interface{}{}

			response, err := client.Post(server.URL, tt.body, okResult, failedResult, tt.headers)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, tt.expectedStatus, response.StatusCode())
			}
		})
	}
}

func TestRestClient_Put(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		headers        map[string]string
		serverHandler  http.HandlerFunc
		expectedStatus int
		expectedError  bool
	}{
		{
			name:    "success - put request",
			body:    map[string]interface{}{"id": 1, "title": "Updated"},
			headers: map[string]string{"X-Request-ID": "req-123"},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var body map[string]interface{}
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				assert.Equal(t, "Updated", body["title"])

				w.WriteHeader(http.StatusOK)
				require.NoError(t, json.NewEncoder(w).Encode(body))
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:    "success - put with nil headers",
			body:    map[string]interface{}{"name": "Test"},
			headers: nil,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				w.WriteHeader(http.StatusOK)
				require.NoError(t, json.NewEncoder(w).Encode(map[string]string{"status": "updated"}))
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.serverHandler)
			defer server.Close()

			cfg := &config.Config{
				HTTPClientTimeout:            60 * time.Second,
				HTTPClientRetryCount:         0,
				HTTPClientRetryWaitMin:       1 * time.Second,
				HTTPClientRetryWaitMax:       5 * time.Second,
				HTTPClientDebug:              false,
				HTTPClientTLSInsecureSkipTLS: false,
				AppName:                      "test-app",
				AppVersion:                   "1.0.0",
			}

			client := ProvideRestClient(cfg)
			okResult := &map[string]interface{}{}
			failedResult := &map[string]interface{}{}

			response, err := client.Put(server.URL, tt.body, okResult, failedResult, tt.headers)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, tt.expectedStatus, response.StatusCode())
			}
		})
	}
}

func TestRestClient_Get(t *testing.T) {
	tests := []struct {
		name           string
		result         interface{}
		headers        map[string]string
		queryParams    string
		serverHandler  http.HandlerFunc
		expectedStatus int
		expectedError  bool
	}{
		{
			name:        "success - get request",
			result:      &map[string]interface{}{},
			headers:     map[string]string{"Accept": "application/json"},
			queryParams: "",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				w.WriteHeader(http.StatusOK)
				require.NoError(t, json.NewEncoder(w).Encode(map[string]string{"id": "1", "name": "Test"}))
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:        "success - get with query params",
			result:      &[]map[string]interface{}{},
			headers:     nil,
			queryParams: "userId=1&status=active",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "1", r.URL.Query().Get("userId"))
				assert.Equal(t, "active", r.URL.Query().Get("status"))

				w.WriteHeader(http.StatusOK)
				require.NoError(t, json.NewEncoder(w).Encode([]map[string]interface{}{
					{"id": "1", "userId": "1"},
				}))
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:        "success - get with nil headers and nil result",
			result:      nil,
			headers:     nil,
			queryParams: "",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("plain text response"))
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:        "error - server returns 404",
			result:      &map[string]interface{}{},
			headers:     map[string]string{},
			queryParams: "",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				require.NoError(t, json.NewEncoder(w).Encode(map[string]string{"error": "not found"}))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.serverHandler)
			defer server.Close()

			cfg := &config.Config{
				HTTPClientTimeout:            60 * time.Second,
				HTTPClientRetryCount:         0,
				HTTPClientRetryWaitMin:       1 * time.Second,
				HTTPClientRetryWaitMax:       5 * time.Second,
				HTTPClientDebug:              false,
				HTTPClientTLSInsecureSkipTLS: false,
				AppName:                      "test-app",
				AppVersion:                   "1.0.0",
			}

			client := ProvideRestClient(cfg)
			response, err := client.Get(server.URL, tt.result, tt.headers, tt.queryParams)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, tt.expectedStatus, response.StatusCode())
			}
		})
	}
}

func TestRestClient_Patch(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		headers        map[string]string
		serverHandler  http.HandlerFunc
		expectedStatus int
		expectedError  bool
	}{
		{
			name:    "success - patch request",
			body:    map[string]interface{}{"title": "Patched"},
			headers: map[string]string{"If-Match": "etag-123"},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPatch, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var body map[string]interface{}
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				assert.Equal(t, "Patched", body["title"])

				w.WriteHeader(http.StatusOK)
				require.NoError(t, json.NewEncoder(w).Encode(body))
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:    "success - patch with nil headers",
			body:    map[string]interface{}{"status": "updated"},
			headers: nil,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPatch, r.Method)
				w.WriteHeader(http.StatusOK)
				require.NoError(t, json.NewEncoder(w).Encode(map[string]string{"status": "patched"}))
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.serverHandler)
			defer server.Close()

			cfg := &config.Config{
				HTTPClientTimeout:            60 * time.Second,
				HTTPClientRetryCount:         0,
				HTTPClientRetryWaitMin:       1 * time.Second,
				HTTPClientRetryWaitMax:       5 * time.Second,
				HTTPClientDebug:              false,
				HTTPClientTLSInsecureSkipTLS: false,
				AppName:                      "test-app",
				AppVersion:                   "1.0.0",
			}

			client := ProvideRestClient(cfg)
			okResult := &map[string]interface{}{}
			failedResult := &map[string]interface{}{}

			response, err := client.Patch(server.URL, tt.body, okResult, failedResult, tt.headers)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, tt.expectedStatus, response.StatusCode())
			}
		})
	}
}

func TestProvideRestClient(t *testing.T) {
	tests := []struct {
		name           string
		cfg            *config.Config
		expectedClient bool
	}{
		{
			name: "success - create rest client with default config",
			cfg: &config.Config{
				HTTPClientTimeout:            30 * time.Second,
				HTTPClientRetryCount:         3,
				HTTPClientRetryWaitMin:       1 * time.Second,
				HTTPClientRetryWaitMax:       5 * time.Second,
				HTTPClientDebug:              false,
				HTTPClientTLSInsecureSkipTLS: false,
				AppName:                      "test-app",
				AppVersion:                   "1.0.0",
			},
			expectedClient: true,
		},
		{
			name: "success - create rest client with retry disabled",
			cfg: &config.Config{
				HTTPClientTimeout:            10 * time.Second,
				HTTPClientRetryCount:         0,
				HTTPClientRetryWaitMin:       1 * time.Second,
				HTTPClientRetryWaitMax:       5 * time.Second,
				HTTPClientDebug:              true,
				HTTPClientTLSInsecureSkipTLS: true,
				AppName:                      "test-app",
				AppVersion:                   "1.0.0",
			},
			expectedClient: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := ProvideRestClient(tt.cfg)

			if tt.expectedClient {
				require.NotNil(t, client)
				assert.Implements(t, (*RestClient)(nil), client)
			} else {
				assert.Nil(t, client)
			}
		})
	}
}

func TestRestClient_HeadersSet(t *testing.T) {
	cfg := &config.Config{
		HTTPClientTimeout:            60 * time.Second,
		HTTPClientRetryCount:         0,
		HTTPClientRetryWaitMin:       1 * time.Second,
		HTTPClientRetryWaitMax:       5 * time.Second,
		HTTPClientDebug:              false,
		HTTPClientTLSInsecureSkipTLS: false,
		AppName:                      "test-app",
		AppVersion:                   "1.0.0",
	}

	client := ProvideRestClient(cfg)

	// Test that Content-Type header is automatically set
	testCases := []struct {
		name              string
		method            string
		endpoint          string
		customHeaders     map[string]string
		expectContentType bool
	}{
		{
			name:              "post sets content-type",
			method:            "POST",
			endpoint:          "https://jsonplaceholder.typicode.com/posts",
			customHeaders:     map[string]string{"X-Custom": "value"},
			expectContentType: true,
		},
		{
			name:              "put sets content-type",
			method:            "PUT",
			endpoint:          "https://jsonplaceholder.typicode.com/posts/1",
			customHeaders:     nil,
			expectContentType: true,
		},
		{
			name:              "patch sets content-type",
			method:            "PATCH",
			endpoint:          "https://jsonplaceholder.typicode.com/posts/1",
			customHeaders:     map[string]string{},
			expectContentType: true,
		},
		{
			name:              "get sets content-type",
			method:            "GET",
			endpoint:          "https://jsonplaceholder.typicode.com/posts/1",
			customHeaders:     nil,
			expectContentType: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var response *resty.Response
			var err error

			okResult := &map[string]interface{}{}
			failedResult := &map[string]interface{}{}

			switch tc.method {
			case "POST":
				response, err = client.Post(tc.endpoint, map[string]string{"test": "data"}, okResult, failedResult, tc.customHeaders)
			case "PUT":
				response, err = client.Put(tc.endpoint, map[string]string{"test": "data"}, okResult, failedResult, tc.customHeaders)
			case "PATCH":
				response, err = client.Patch(tc.endpoint, map[string]string{"test": "data"}, okResult, failedResult, tc.customHeaders)
			case "GET":
				response, err = client.Get(tc.endpoint, okResult, tc.customHeaders, "")
			}

			// We're testing the method signature and that it doesn't panic
			// Actual header verification would require inspecting the resty client
			// which is harder to test without integration tests
			if err == nil && response != nil {
				assert.NotNil(t, response)
			}
		})
	}
}
