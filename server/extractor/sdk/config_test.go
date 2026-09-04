package sdk_test

import (
	"testing"

	"github.com/asciimoo/hister/server/extractor/sdk"
)

func TestConfigSupportZeroValue(t *testing.T) {
	var store sdk.ConfigSupport

	config := store.GetConfig()
	if !config.Enable {
		t.Fatal("zero value store should be enabled")
	}
	if config.Options == nil {
		t.Fatal("zero value store should return a nonnil options map")
	}
	if len(config.Options) != 0 {
		t.Fatalf("zero value store returned options: %v", config.Options)
	}
}

func TestConfigSupportCustomDefaults(t *testing.T) {
	store := sdk.NewConfigSupport(sdk.Config{
		Enable: false,
		Options: map[string]any{
			"timeout": 10,
		},
	}, "optional")

	first := store.GetConfig()
	if first.Enable {
		t.Fatal("custom default should be disabled")
	}
	if first.Options["timeout"] != 10 {
		t.Fatalf("unexpected default options: %v", first.Options)
	}

	first.Options["timeout"] = 20
	second := store.GetConfig()
	if second.Options["timeout"] != 10 {
		t.Fatalf("default options were mutated: %v", second.Options)
	}

	configured := &sdk.Config{
		Enable: true,
		Options: map[string]any{
			"timeout":  30,
			"optional": true,
		},
	}
	if err := store.SetConfig(configured); err != nil {
		t.Fatalf("SetConfig returned an error: %v", err)
	}
	if store.GetConfig() != configured {
		t.Fatal("SetConfig should retain the supplied config pointer")
	}
}

func TestConfigSupportRejectsUnknownOptions(t *testing.T) {
	store := sdk.NewConfigSupport(sdk.Config{
		Enable:  true,
		Options: map[string]any{"known": 1},
	})

	configured := &sdk.Config{
		Enable:  true,
		Options: map[string]any{"unknown": 1},
	}
	if err := store.SetConfig(configured); err == nil {
		t.Fatal("SetConfig accepted an unknown option")
	}
	if store.GetConfig() == configured {
		t.Fatal("SetConfig stored a rejected config")
	}
}
