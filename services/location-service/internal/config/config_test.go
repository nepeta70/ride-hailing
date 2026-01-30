package config_test

import (
	"encoding/json"
	"os"
	"testing"

	locationconfig "github.com/nepeta70/ride-hailing/services/location-service/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := locationconfig.DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}
	if cfg.Logic.GeohashPrecision != 7 {
		t.Errorf("Expected GeohashPrecision 7, got %d", cfg.Logic.GeohashPrecision)
	}
	if cfg.Logic.LocationTTLSeconds != 300 {
		t.Errorf("Expected LocationTTLSeconds 300, got %d", cfg.Logic.LocationTTLSeconds)
	}
	if cfg.Logic.TopKNearby != 5 {
		t.Errorf("Expected TopKNearby 5, got %d", cfg.Logic.TopKNearby)
	}
}

func TestLoad_JSON(t *testing.T) {
	file, err := os.CreateTemp("", "location-cfg-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(file.Name())
	json.NewEncoder(file).Encode(map[string]any{
		"logic": map[string]any{
			"geohash_precision":    9,
			"location_ttl_seconds": 100,
			"top_k_nearby":         10,
		},
	})
	file.Close()
	cfg, err := locationconfig.Load(file.Name())
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Logic.GeohashPrecision != 9 {
		t.Errorf("Expected GeohashPrecision 9, got %d", cfg.Logic.GeohashPrecision)
	}
	if cfg.Logic.LocationTTLSeconds != 100 {
		t.Errorf("Expected LocationTTLSeconds 100, got %d", cfg.Logic.LocationTTLSeconds)
	}
	if cfg.Logic.TopKNearby != 10 {
		t.Errorf("Expected TopKNearby 10, got %d", cfg.Logic.TopKNearby)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	cfg, err := locationconfig.Load("/nonexistent/location-cfg.json")
	if err == nil {
		t.Error("Expected error for missing file, got nil")
	}
	if cfg != nil {
		t.Error("Expected nil config for missing file")
	}
}
