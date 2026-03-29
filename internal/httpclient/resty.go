package httpclient

import (
	"crypto/tls"
	"net/http"

	"golang-boilerplate/internal/config"

	"github.com/go-resty/resty/v2"
	"github.com/labstack/echo/v4"
)

type RestClient interface {
	Post(endpoint string, body, okResult, failedResult interface{}, headers map[string]string) (*resty.Response, error)
	Put(endpoint string, body, okResult, failedResult interface{}, headers map[string]string) (*resty.Response, error)
	Get(endpoint string, result interface{}, headers map[string]string, queryParams string) (*resty.Response, error)
	Patch(endpoint string, body, okResult, failedResult interface{}, headers map[string]string) (*resty.Response, error)
}

type restClient struct {
	client resty.Client
}

func ProvideRestClient(cfg *config.Config) RestClient {
	// Create a new Resty client with timeout configuration
	c := resty.New().
		SetTimeout(cfg.HTTPClientTimeout)

	// Default headers
	c = c.SetHeaders(map[string]string{
		echo.HeaderAccept:      echo.MIMEApplicationJSON,
		echo.HeaderContentType: echo.MIMEApplicationJSON,
		"User-Agent":           cfg.AppName + "/" + cfg.AppVersion,
		"Accept-Charset":       "utf-8",
	})

	// Retry policy
	if cfg.HTTPClientRetryCount > 0 {
		c = c.
			SetRetryCount(cfg.HTTPClientRetryCount).
			SetRetryWaitTime(cfg.HTTPClientRetryWaitMin).
			SetRetryMaxWaitTime(cfg.HTTPClientRetryWaitMax).
			AddRetryCondition(func(r *resty.Response, err error) bool {
				if err != nil {
					return true
				}
				if r == nil {
					return true
				}
				status := r.StatusCode()
				// Retry on 5xx and 429
				if status == http.StatusTooManyRequests || (status >= 500 && status < 600) {
					return true
				}
				return false
			})
	}

	// Debug logging
	c = c.SetDebug(cfg.HTTPClientDebug)

	// TLS options
	if cfg.HTTPClientTLSInsecureSkipTLS {
		//nolint:gosec // G402: explicit opt-in via HTTPClientTLSInsecureSkipTLS for dev/test only
		c = c.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return &restClient{client: *c}
}

func (r *restClient) Post(
	endpoint string, body,
	okResult,
	failedResult interface{},
	headers map[string]string,
) (*resty.Response, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	headers[echo.HeaderContentType] = echo.MIMEApplicationJSON

	return r.client.R().
		SetHeaders(headers).
		SetBody(body).
		SetResult(okResult).
		SetError(failedResult).
		Post(endpoint)
}

func (r *restClient) Put(
	endpoint string, body,
	okResult,
	failedResult interface{},
	headers map[string]string,
) (*resty.Response, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	headers[echo.HeaderContentType] = echo.MIMEApplicationJSON

	return r.client.R().
		SetHeaders(headers).
		SetBody(body).
		SetResult(okResult).
		SetError(failedResult).
		Put(endpoint)
}

func (r *restClient) Get(
	endpoint string,
	result interface{},
	headers map[string]string,
	queryParams string,
) (*resty.Response, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	headers[echo.HeaderContentType] = echo.MIMEApplicationJSON

	request := r.client.R().
		SetHeaders(headers).
		SetQueryString(queryParams)

	// Only set result and error if they are not nil
	if result != nil {
		request = request.SetResult(result).SetError(result)
	}

	return request.Get(endpoint)
}

func (r *restClient) Patch(
	endpoint string, body,
	okResult,
	failedResult interface{},
	headers map[string]string,
) (*resty.Response, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	headers[echo.HeaderContentType] = echo.MIMEApplicationJSON

	return r.client.R().
		SetHeaders(headers).
		SetBody(body).
		SetResult(okResult).
		SetError(failedResult).
		Patch(endpoint)
}
