package db

import (
	"golang-boilerplate/internal/config"

	_ "github.com/lib/pq" // PostgreSQL driver for database/sql
	"gorm.io/gorm"
)

// PostgresDB wraps the database connection
type PostgresDB struct {
	*gorm.DB
	manager *DatabaseManager
}

// NewPostgresDB creates a new PostgreSQL database connection using the enhanced manager
func (r *PostgresDB) NewPostgresDB(c *config.Config) error {
	// Create the database manager
	manager, err := NewDatabaseManager(c)
	if err != nil {
		return err
	}

	// Get the GORM instance from the manager
	r.DB = manager.GetDB()
	r.manager = manager

	return nil
}

// GetManager returns the database manager for advanced operations
func (r *PostgresDB) GetManager() *DatabaseManager {
	return r.manager
}

// HealthCheck performs a health check on the database connection
func (r *PostgresDB) HealthCheck() HealthStatus {
	if r.manager != nil {
		return r.manager.HealthCheck()
	}
	return HealthStatus{IsHealthy: false}
}

// FastHealthCheck performs a quick health check without database ping
func (r *PostgresDB) FastHealthCheck() HealthStatus {
	if r.manager != nil {
		return r.manager.FastHealthCheck()
	}
	return HealthStatus{IsHealthy: false}
}

// GetMetrics returns the current connection metrics
func (r *PostgresDB) GetMetrics() ConnectionMetrics {
	if r.manager != nil {
		return r.manager.GetMetrics()
	}
	return ConnectionMetrics{}
}

// Close gracefully closes the database connection
func (r *PostgresDB) Close() error {
	if r.manager != nil {
		return r.manager.Close()
	}
	return nil
}
