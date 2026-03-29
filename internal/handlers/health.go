package handlers

import (
	"golang-boilerplate/internal/config"
	"golang-boilerplate/internal/db"
	"golang-boilerplate/internal/dtos"
	"time"

	"github.com/labstack/echo/v4"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	BaseHandler
	cfg *config.Config
	db  *db.PostgresDB
}

// NewHealthHandler creates a new health handler
func ProvideHealthHandler(cfg *config.Config, db *db.PostgresDB) *HealthHandler {
	return &HealthHandler{
		BaseHandler: *NewBaseHandler(),
		cfg:         cfg,
		db:          db,
	}
}

// HealthCheck godoc
// @Summary Health Check
// @Description Check if the service is running
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.HealthResponse}
// @Router / [get]
func (h *HealthHandler) HealthCheck(c echo.Context) error {
	healthResponse := dtos.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Version:   h.cfg.AppVersion,
		Service:   h.cfg.AppName,
	}

	response := h.SuccessResponse(c, "Service is healthy", healthResponse, nil)
	return response
}

// DatabaseHealthCheck godoc
// @Summary Database Health Check
// @Description Check if the database connection is healthy
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} object{meta=dtos.Meta,data=object}
// @Router /health/database [get]
func (h *HealthHandler) DatabaseHealthCheck(c echo.Context) error {
	if h.db == nil {
		return h.InternalErrorResponse(c, "Database not initialized", nil)
	}

	// Use fast, cached health checks to avoid excessive load on the database
	healthStatus := h.db.FastHealthCheck()
	metrics := h.db.GetMetrics()

	response := map[string]interface{}{
		"database_health":    healthStatus,
		"connection_metrics": metrics,
		"timestamp":          time.Now().UTC(),
	}

	if healthStatus.IsHealthy {
		return h.SuccessResponse(c, "Database is healthy", response, nil)
	}
	return h.InternalErrorResponse(c, "Database is unhealthy", nil)
}

// DatabaseMetrics godoc
// @Summary Database Connection Metrics
// @Description Get detailed database connection metrics
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} object{meta=dtos.Meta,data=object}
// @Router /health/metrics [get]
func (h *HealthHandler) DatabaseMetrics(c echo.Context) error {
	if h.db == nil {
		return h.InternalErrorResponse(c, "Database not initialized", nil)
	}

	metrics := h.db.GetMetrics()
	healthStatus := h.db.HealthCheck()

	response := map[string]interface{}{
		"connection_metrics": metrics,
		"health_status":      healthStatus,
		"timestamp":          time.Now().UTC(),
		"configuration": map[string]interface{}{
			"max_open_connections": h.cfg.DatabaseMaxOpenConns,
			"max_idle_connections": h.cfg.DatabaseMaxIdleConns,
			"conn_max_lifetime":    h.cfg.DatabaseConnMaxLifetime.String(),
			"conn_max_idle_time":   h.cfg.DatabaseConnMaxIdleTime.String(),
			"connect_timeout":      h.cfg.DatabaseConnectTimeout.String(),
			"query_timeout":        h.cfg.DatabaseQueryTimeout.String(),
		},
	}

	return h.SuccessResponse(c, "Database metrics retrieved successfully", response, nil)
}
