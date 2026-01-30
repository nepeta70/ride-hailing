package config_test

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/nepeta70/ride-hailing/internal/pkg/config"
)

type ConfigMock struct {
	Valid bool
}

func (d *ConfigMock) Validate() error {
	if !d.Valid {
		return os.ErrInvalid
	}
	return nil
}

func (d *ConfigMock) Init() error {
	return nil
}

func TestDefaultBaseConfig(t *testing.T) {
	cfg := config.DefaultBaseConfig()
	if reflect.DeepEqual(cfg, config.BaseConfig{}) {
		t.Error("DefaultBaseConfig should not return zero value")
	}
}

func TestLoadGeneric_FileNotFound(t *testing.T) {
	type TestConfig struct {
		A string `json:"a"`
	}
	var cfg TestConfig
	_, err := config.LoadGeneric("nonexistent.json", &cfg)
	if err != nil && !os.IsNotExist(err) {
		t.Errorf("Expected file not found or nil error, got: %v", err)
	}
}

func TestLoadGeneric_Validator(t *testing.T) {
	type TestConfig struct {
		Dummy ConfigMock
	}
	// Write a config file with valid=false
	tmp, err := os.CreateTemp("", "testcfg-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmp.Name())
	json.NewEncoder(tmp).Encode(map[string]any{"Dummy": map[string]any{"Valid": false}})
	tmp.Close()
	var cfg TestConfig
	_, err = config.LoadGeneric(tmp.Name(), &cfg)
	if err == nil {
		t.Error("Expected error from invalid validator, got nil")
	}
	// Now test with valid=true
	tmp2, err := os.CreateTemp("", "testcfg-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmp2.Name())
	json.NewEncoder(tmp2).Encode(map[string]any{"Dummy": map[string]any{"Valid": true}})
	tmp2.Close()
	var cfg2 TestConfig
	_, err = config.LoadGeneric(tmp2.Name(), &cfg2)
	if err != nil {
		t.Errorf("Expected nil error from valid validator, got: %v", err)
	}
}

func TestLoadGeneric_JSONUnmarshal(t *testing.T) {
	type TestConfig struct {
		A string `json:"a"`
	}
	file, err := os.CreateTemp("", "testcfg-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(file.Name())
	json.NewEncoder(file).Encode(map[string]string{"a": "hello"})
	file.Close()
	var cfg TestConfig
	_, err = config.LoadGeneric(file.Name(), &cfg)
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}
	if cfg.A != "hello" {
		t.Errorf("Expected 'hello', got: %v", cfg.A)
	}
}
