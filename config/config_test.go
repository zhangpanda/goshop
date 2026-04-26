package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	content := `
server:
  port: 9090
  mode: release
db:
  host: 127.0.0.1
  port: 3306
  user: root
  password: "test"
  dbname: testdb
redis:
  host: 127.0.0.1
  port: 6379
  password: ""
  db: 0
jwt:
  secret: "test-secret"
  expire: 24
`
	f, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(content)
	f.Close()

	cfg, err := Load(f.Name())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("Port = %d, want 9090", cfg.Server.Port)
	}
	if cfg.Server.Mode != "release" {
		t.Errorf("Mode = %s, want release", cfg.Server.Mode)
	}
	if cfg.DB.DBName != "testdb" {
		t.Errorf("DBName = %s, want testdb", cfg.DB.DBName)
	}
	if cfg.JWT.Secret != "test-secret" {
		t.Errorf("JWT.Secret = %s, want test-secret", cfg.JWT.Secret)
	}
	if cfg.JWT.Expire != 24 {
		t.Errorf("JWT.Expire = %d, want 24", cfg.JWT.Expire)
	}
}

func TestLoadConfigNotFound(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Error("Load should fail with nonexistent file")
	}
}

func TestLoadConfigInvalid(t *testing.T) {
	f, _ := os.CreateTemp("", "config-*.yaml")
	defer os.Remove(f.Name())
	f.WriteString("{{invalid yaml")
	f.Close()

	_, err := Load(f.Name())
	if err == nil {
		t.Error("Load should fail with invalid yaml")
	}
}
