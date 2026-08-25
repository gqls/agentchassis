package config

// The delivery listener's opt-in arrives as an ENVIRONMENT VARIABLE from the
// production overlay, and nothing in configs/core-manager.yaml names the key.
//
// That combination is a known viper trap: Unmarshal populates from viper's known
// key set, and AutomaticEnv alone does not add a key that appears in neither the
// config file nor a default. A key that nothing reads looks exactly like a key
// that reads empty — and here "empty" means no delivery listener, so every
// customer link would 404 at the box with the overlay looking correct and no
// error anywhere in the cluster.
//
// This test exists because that failure is invisible at every other layer.

import (
	"os"
	"path/filepath"
	"testing"
)

func writeMinimalConfig(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "svc.yaml")
	// Deliberately does NOT mention delivery_port — this mirrors
	// configs/core-manager.yaml, where the key is absent on purpose.
	body := "service_info:\n  name: \"core-manager\"\nserver:\n  port: \"8088\"\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestDeliveryPortIsReadFromTheEnvironment(t *testing.T) {
	t.Setenv("SERVICE_SERVER_DELIVERY_PORT", "8090")

	cfg, err := Load(writeMinimalConfig(t))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Server.DeliveryPort != "8090" {
		t.Fatalf("Server.DeliveryPort = %q, want %q: the production overlay sets "+
			"SERVICE_SERVER_DELIVERY_PORT and the config file never names the key, "+
			"so if this does not bind, the delivery listener never starts and every "+
			"customer link 404s at the box", cfg.Server.DeliveryPort, "8090")
	}
	if cfg.Server.Port != "8088" {
		t.Errorf("Server.Port = %q, want 8088 (control: the file's own value must survive)", cfg.Server.Port)
	}
}

func TestDeliveryPortDefaultsOffWhenUnset(t *testing.T) {
	cfg, err := Load(writeMinimalConfig(t))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Server.DeliveryPort != "" {
		t.Errorf("Server.DeliveryPort = %q with nothing set, want empty: the "+
			"delivery listener's default must be OFF", cfg.Server.DeliveryPort)
	}
}
