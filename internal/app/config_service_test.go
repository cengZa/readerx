package app

import (
	"testing"

	"github.com/heybox/readerx/internal/storage"
)

func TestConfigServiceSetsAndGetsAllowedConfig(t *testing.T) {
	store, err := storage.OpenSQLite(t.TempDir() + "/reader.db")
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	defer store.Close()

	service := NewConfigService(store)
	if err := service.Set("search.limit", "25"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	value, err := service.Get("search.limit")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if value != "25" {
		t.Fatalf("value = %q, want 25", value)
	}
}

func TestConfigServiceRejectsUnknownKey(t *testing.T) {
	store, err := storage.OpenSQLite(t.TempDir() + "/reader.db")
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	defer store.Close()

	err = NewConfigService(store).Set("unknown.key", "value")
	if err == nil {
		t.Fatalf("Set unknown key returned nil error")
	}
}
