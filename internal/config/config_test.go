package config

import (
	"testing"
)

func TestLoadDefault(t *testing.T) {
	cfg := LoadDefault()

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Expected host 0.0.0.0, got %s", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Database.Database != "growth_tracker" {
		t.Errorf("Expected database growth_tracker, got %s", cfg.Database.Database)
	}
	if cfg.JWT.ExpireTime != 86400*7 {
		t.Errorf("Expected expire time 604800, got %d", cfg.JWT.ExpireTime)
	}
}

func TestDatabaseConfig_GetDSN(t *testing.T) {
	cfg := &DatabaseConfig{
		User:     "root",
		Password: "password",
		Host:     "localhost",
		Port:     3306,
		Database: "test_db",
		Charset:  "utf8mb4",
	}

	dsn := cfg.GetDSN()
	expected := "root:password@tcp(localhost:3306)/test_db?charset=utf8mb4&parseTime=True&loc=Local"

	if dsn != expected {
		t.Errorf("Expected DSN %s, got %s", expected, dsn)
	}
}

func TestRedisConfig_GetAddr(t *testing.T) {
	cfg := &RedisConfig{
		Host: "localhost",
		Port: 6379,
	}

	addr := cfg.GetAddr()
	expected := "localhost:6379"

	if addr != expected {
		t.Errorf("Expected addr %s, got %s", expected, addr)
	}
}

func TestServerConfig_GetAddr(t *testing.T) {
	cfg := &ServerConfig{
		Host: "0.0.0.0",
		Port: 8080,
	}

	addr := cfg.GetAddr()
	expected := "0.0.0.0:8080"

	if addr != expected {
		t.Errorf("Expected addr %s, got %s", expected, addr)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("nonexistent.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}
