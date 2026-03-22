//go:build integration

package db

import (
	"context"
	"testing"
	"time"

	"golang-boilerplate/internal/config"

	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

// TestDatabaseManager_Postgres_Health is a minimal integration test that
// exercises DatabaseManager against a real Postgres instance using
// testcontainers. It is guarded by the `integration` build tag so it
// only runs when explicitly requested.
func TestDatabaseManager_Postgres_Health(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	container, err := tcpostgres.RunContainer(ctx,
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("testuser"),
		tcpostgres.WithPassword("testpassword"),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	defer func() {
		_ = container.Terminate(ctx)
	}()

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get container host: %v", err)
	}

	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatalf("failed to get mapped port: %v", err)
	}

	cfg := &config.Config{
		DatabaseHost:           host,
		DatabasePort:           port.Port(),
		DatabaseUsername:       "testuser",
		DatabasePassword:       "testpassword",
		DatabaseName:           "testdb",
		DatabaseEnableDebug:    false,
		DatabaseMaxOpenConns:   5,
		DatabaseMaxIdleConns:   5,
		DatabaseConnMaxLifetime: 5 * time.Minute,
		DatabaseConnMaxIdleTime: 1 * time.Minute,
		DatabaseConnectTimeout:  10 * time.Second,
		DatabaseQueryTimeout:    5 * time.Second,
		DatabaseHealthTimeout:   3 * time.Second,
		DatabaseRetryAttempts:   1,
		DatabaseRetryDelay:      1 * time.Second,
		DatabaseSSLMode:         "disable",
		DatabaseTimezone:        "UTC",
	}

	manager, err := NewDatabaseManager(cfg)
	if err != nil {
		t.Fatalf("failed to create database manager: %v", err)
	}

	status := manager.HealthCheck()
	if !status.IsHealthy {
		t.Fatalf("expected database to be healthy, got: %+v", status)
	}
}

