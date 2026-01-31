package common

import (
	"os"
	"path/filepath"
	"testing"
)

// helper to create a config populated across sections for round-trip testing.
func buildFullConfig() *Config {
	return &Config{
		Server: ServerConfig{Address: ":8081"},
		Client: ClientConfig{URL: "http://client"},
		Broker: BrokerConfig{LogFile: "broker.log", LogsDir: "logs", DetectBins: true},
		Proc:   ProcConfig{MaxMessageSizeBytes: 1024, RPCInputFD: 3, RPCOutputFD: 4},
		Local:  LocalConfig{}, // Local config is now build-time only
		Meta:   MetaConfig{SessionDBPath: "session.db"},
		Remote: RemoteConfig{NVDAPIKey: "key", ViewFetchURL: "http://views"},
		Assets: AssetsConfig{}, // Assets config is now build-time only
		Capec:  CapecConfig{}, // CAPEC config is now build-time only

		Logging: LoggingConfig{Level: "debug", Dir: "logdir"},
		Access:  AccessConfig{RPCTimeoutSeconds: 5, ShutdownTimeoutSeconds: 2, StaticDir: "site"},
	}
}

func TestSaveConfig_FailsWhenPathIsDirectory(t *testing.T) {
	dir := t.TempDir()
	// Attempt to write to the directory path itself should fail.
	err := SaveConfig(&Config{}, dir)
	if err == nil {
		t.Fatalf("expected SaveConfig to fail when target is directory")
	}
}

func TestLoadConfig_FailsWhenPathIsDirectory(t *testing.T) {
	dir := t.TempDir()
	_, err := LoadConfig(dir)
	if err == nil {
		t.Fatalf("expected LoadConfig to fail when target is directory")
	}
}

func TestSaveLoadConfig_RoundTripFullConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	original := buildFullConfig()

	if err := SaveConfig(original, path); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	loaded, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Ensure selected fields survived round-trip.
	if loaded.Server.Address != original.Server.Address || loaded.Client.URL != original.Client.URL {
		t.Fatalf("round-trip mismatch in server/client: %+v vs %+v", loaded, original)
	}
	if loaded.Broker.LogFile != original.Broker.LogFile || loaded.Broker.LogsDir != original.Broker.LogsDir {
		t.Fatalf("round-trip mismatch in broker: %+v vs %+v", loaded.Broker, original.Broker)
	}
	if loaded.Access.StaticDir != original.Access.StaticDir || loaded.Access.RPCTimeoutSeconds != original.Access.RPCTimeoutSeconds {
		t.Fatalf("round-trip mismatch in access: %+v vs %+v", loaded.Access, original.Access)
	}
}

func TestSaveConfig_DefaultFileOverwrites(t *testing.T) {
	dir := t.TempDir()
	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(original)
	})
	if err := os.WriteFile(DefaultConfigFile, []byte(`{"server":{"address":"old"}}`), 0644); err != nil {
		t.Fatalf("prewrite failed: %v", err)
	}

	cfg := &Config{Server: ServerConfig{Address: "new"}}
	if err := SaveConfig(cfg, ""); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	loaded, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if loaded.Server.Address != "new" {
		t.Fatalf("expected overwritten address 'new', got %q", loaded.Server.Address)
	}
}
