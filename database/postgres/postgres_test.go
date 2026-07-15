package postgres

import (
	"testing"
	"time"
)

func TestPoolSettings(t *testing.T) {
	maxIdle, maxOpen, maxLifetime := poolSettings(&DBConn{})
	if maxIdle != _defaultMaxIdleConns || maxOpen != _defaultMaxOpenConns || maxLifetime != _defaultMaxLifeTime {
		t.Fatalf("poolSettings defaults = (%d, %d, %s)", maxIdle, maxOpen, maxLifetime)
	}

	maxIdle, maxOpen, maxLifetime = poolSettings(&DBConn{
		MaxIdleConns:    3,
		MaxOpenConns:    7,
		ConnMaxLifetime: time.Minute,
	})
	if maxIdle != 3 || maxOpen != 7 || maxLifetime != time.Minute {
		t.Fatalf("poolSettings overrides = (%d, %d, %s)", maxIdle, maxOpen, maxLifetime)
	}
}

func TestNewDoesNotConnectDuringInitialization(t *testing.T) {
	if _, err := New(nil); err == nil {
		t.Fatal("New(nil) returned nil error")
	}

	cfg := ConnectionConfig{Host: "127.0.0.1", Port: "1", UserName: "app"}
	db, err := New(&DBConn{
		Database: "app",
		Master:   cfg,
		Replicas: []ConnectionConfig{cfg},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("DB() error = %v", err)
	}
	defer sqlDB.Close()
	if got := sqlDB.Stats().OpenConnections; got != 0 {
		t.Fatalf("open connections = %d, want 0", got)
	}
}
