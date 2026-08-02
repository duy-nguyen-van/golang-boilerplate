package handlers

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"golang-boilerplate/internal/config"
	"golang-boilerplate/internal/db"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvideHealthHandler(t *testing.T) {
	h := ProvideHealthHandler(&config.Config{AppName: "app", AppVersion: "1.0.0"}, nil)
	require.NotNil(t, h)
	assert.Nil(t, h.db)
}

func TestHealthHandler_HealthCheck(t *testing.T) {
	h := ProvideHealthHandler(&config.Config{AppName: "golang-boilerplate", AppVersion: "2.0.0"}, nil)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	require.NoError(t, h.HealthCheck(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "golang-boilerplate")
	assert.Contains(t, rec.Body.String(), "2.0.0")
}

func TestHealthHandler_DatabaseHealthCheck(t *testing.T) {
	cfg := &config.Config{AppName: "app", AppVersion: "1.0.0"}

	t.Run("nil db", func(t *testing.T) {
		h := ProvideHealthHandler(cfg, nil)
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/health/database", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		require.NoError(t, h.DatabaseHealthCheck(c))
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "Database not initialized")
	})

	t.Run("unhealthy db", func(t *testing.T) {
		h := ProvideHealthHandler(cfg, &db.PostgresDB{})
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/health/database", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		require.NoError(t, h.DatabaseHealthCheck(c))
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "Database is unhealthy")
	})

	t.Run("healthy db", func(t *testing.T) {
		h := ProvideHealthHandler(cfg, postgresDBWithHealth(true))
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/health/database", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		require.NoError(t, h.DatabaseHealthCheck(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Database is healthy")
	})
}

func TestHealthHandler_DatabaseMetrics(t *testing.T) {
	cfg := &config.Config{
		AppName:                 "app",
		AppVersion:              "1.0.0",
		DatabaseMaxOpenConns:    25,
		DatabaseMaxIdleConns:    5,
		DatabaseConnMaxLifetime: 5 * time.Minute,
		DatabaseConnMaxIdleTime: time.Minute,
		DatabaseConnectTimeout:  30 * time.Second,
		DatabaseQueryTimeout:    30 * time.Second,
	}

	t.Run("nil db", func(t *testing.T) {
		h := ProvideHealthHandler(cfg, nil)
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/health/metrics", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		require.NoError(t, h.DatabaseMetrics(c))
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("with db", func(t *testing.T) {
		// Nil manager: GetMetrics/HealthCheck return zero values without touching a live DB.
		h := ProvideHealthHandler(cfg, &db.PostgresDB{})
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/health/metrics", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		require.NoError(t, h.DatabaseMetrics(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "connection_metrics")
		assert.Contains(t, rec.Body.String(), "max_open_connections")
	})
}

// postgresDBWithHealth builds a PostgresDB with an injected DatabaseManager so FastHealthCheck /
// GetMetrics / HealthCheck can return cached values without a live database.
func postgresDBWithHealth(healthy bool) *db.PostgresDB {
	pdb := &db.PostgresDB{}
	mgrType := reflect.TypeOf((*db.DatabaseManager)(nil)).Elem()
	mgrPtr := reflect.New(mgrType)
	mgrVal := mgrPtr.Elem()

	hsField := mgrVal.FieldByName("healthStatus")
	hs := db.HealthStatus{
		IsHealthy: healthy,
		LastCheck: time.Now(),
	}
	reflect.NewAt(hsField.Type(), unsafe.Pointer(hsField.UnsafeAddr())).Elem().Set(reflect.ValueOf(hs))

	metricsField := mgrVal.FieldByName("metrics")
	metrics := &db.ConnectionMetrics{OpenConnections: 1, MaxOpenConnections: 25}
	reflect.NewAt(metricsField.Type(), unsafe.Pointer(metricsField.UnsafeAddr())).Elem().Set(reflect.ValueOf(metrics))

	pdbVal := reflect.ValueOf(pdb).Elem()
	managerField := pdbVal.FieldByName("manager")
	reflect.NewAt(managerField.Type(), unsafe.Pointer(managerField.UnsafeAddr())).Elem().Set(mgrPtr)

	return pdb
}
