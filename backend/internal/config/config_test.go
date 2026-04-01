package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Server.Port != "8080" {
		t.Errorf("expected server port 8080, got %s", cfg.Server.Port)
	}
	if cfg.Database.Host != "127.0.0.1" {
		t.Errorf("expected db host 127.0.0.1, got %s", cfg.Database.Host)
	}
	if cfg.Database.Port != "3306" {
		t.Errorf("expected db port 3306, got %s", cfg.Database.Port)
	}
	if cfg.Database.Name != "restaurant_kds" {
		t.Errorf("expected db name restaurant_kds, got %s", cfg.Database.Name)
	}
	if cfg.Database.MaxIdleConns != 10 {
		t.Errorf("expected max idle conns 10, got %d", cfg.Database.MaxIdleConns)
	}
	if cfg.Database.MaxOpenConns != 100 {
		t.Errorf("expected max open conns 100, got %d", cfg.Database.MaxOpenConns)
	}
	if cfg.Database.ConnMaxLifetimeMin != 60 {
		t.Errorf("expected conn max lifetime 60, got %d", cfg.Database.ConnMaxLifetimeMin)
	}
	if cfg.Business.MergeWindowMinutes != 5 {
		t.Errorf("expected merge window 5, got %d", cfg.Business.MergeWindowMinutes)
	}
	if cfg.Business.RefundPermissions.NotMade != "waiter" {
		t.Errorf("expected not_made permission waiter, got %s", cfg.Business.RefundPermissions.NotMade)
	}
	if cfg.Business.RefundPermissions.MadeNotServed != "leader" {
		t.Errorf("expected made_not_served permission leader, got %s", cfg.Business.RefundPermissions.MadeNotServed)
	}
	if cfg.Business.RefundPermissions.Served != "manager" {
		t.Errorf("expected served permission manager, got %s", cfg.Business.RefundPermissions.Served)
	}
}

func TestDSN(t *testing.T) {
	cfg := DefaultConfig()
	expected := "root:@tcp(127.0.0.1:3306)/restaurant_kds?charset=utf8mb4&parseTime=True&loc=Local"
	if got := cfg.DSN(); got != expected {
		t.Errorf("DSN mismatch\n  got:  %s\n  want: %s", got, expected)
	}
}

func TestLoadFromBytes_PartialOverride(t *testing.T) {
	yaml := []byte(`
server:
  port: "9090"
business:
  merge_window_minutes: 10
`)
	cfg := LoadFromBytes(yaml)

	// Overridden values
	if cfg.Server.Port != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.Server.Port)
	}
	if cfg.Business.MergeWindowMinutes != 10 {
		t.Errorf("expected merge window 10, got %d", cfg.Business.MergeWindowMinutes)
	}

	// Defaults preserved
	if cfg.Database.Host != "127.0.0.1" {
		t.Errorf("expected default db host 127.0.0.1, got %s", cfg.Database.Host)
	}
	if cfg.Database.MaxIdleConns != 10 {
		t.Errorf("expected default max idle conns 10, got %d", cfg.Database.MaxIdleConns)
	}
	if cfg.Business.RefundPermissions.Served != "manager" {
		t.Errorf("expected default served permission manager, got %s", cfg.Business.RefundPermissions.Served)
	}
}

func TestLoadFromBytes_Empty(t *testing.T) {
	cfg := LoadFromBytes(nil)
	def := DefaultConfig()

	if cfg.Server.Port != def.Server.Port {
		t.Errorf("expected default port %s, got %s", def.Server.Port, cfg.Server.Port)
	}
	if cfg.Business.MergeWindowMinutes != def.Business.MergeWindowMinutes {
		t.Errorf("expected default merge window %d, got %d", def.Business.MergeWindowMinutes, cfg.Business.MergeWindowMinutes)
	}
}

func TestLoadFromBytes_RefundPermissionOverride(t *testing.T) {
	yaml := []byte(`
business:
  refund_permissions:
    made_not_served: "manager"
    served: "manager"
`)
	cfg := LoadFromBytes(yaml)

	if cfg.Business.RefundPermissions.MadeNotServed != "manager" {
		t.Errorf("expected made_not_served manager, got %s", cfg.Business.RefundPermissions.MadeNotServed)
	}
	if cfg.Business.RefundPermissions.Served != "manager" {
		t.Errorf("expected served manager, got %s", cfg.Business.RefundPermissions.Served)
	}
}

func TestEnvOverrides(t *testing.T) {
	// Set env vars
	os.Setenv("DB_HOST", "db.example.com")
	os.Setenv("SERVER_PORT", "3000")
	os.Setenv("MERGE_WINDOW_MINUTES", "15")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("MERGE_WINDOW_MINUTES")
	}()

	cfg := DefaultConfig()
	applyEnvOverrides(cfg)

	if cfg.Database.Host != "db.example.com" {
		t.Errorf("expected db host db.example.com, got %s", cfg.Database.Host)
	}
	if cfg.Server.Port != "3000" {
		t.Errorf("expected server port 3000, got %s", cfg.Server.Port)
	}
	if cfg.Business.MergeWindowMinutes != 15 {
		t.Errorf("expected merge window 15, got %d", cfg.Business.MergeWindowMinutes)
	}
}
