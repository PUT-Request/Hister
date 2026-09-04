package sdk

import (
	"fmt"
	"maps"
)

// ConfigSupport provides the standard configuration behavior for an extractor.
// Its zero value is enabled by default and rejects every option key.
// Embed it in an extractor to provide GetConfig and SetConfig automatically.
type ConfigSupport struct {
	config         *Config
	defaults       Config
	allowedOptions map[string]struct{}
	hasDefaults    bool
}

// NewConfigSupport returns a configuration store with custom defaults. Option
// keys present in defaults or allowedOptions are accepted by SetConfig.
func NewConfigSupport(defaults Config, allowedOptions ...string) ConfigSupport {
	allowed := make(map[string]struct{}, len(defaults.Options)+len(allowedOptions))
	for key := range defaults.Options {
		allowed[key] = struct{}{}
	}
	for _, key := range allowedOptions {
		allowed[key] = struct{}{}
	}
	return ConfigSupport{
		defaults:       copyConfig(defaults),
		allowedOptions: allowed,
		hasDefaults:    true,
	}
}

// GetConfig returns the applied configuration or a fresh copy of the
// defaults. The zero value defaults to an enabled extractor with no options.
func (s *ConfigSupport) GetConfig() *Config {
	if s.config != nil {
		return s.config
	}
	if !s.hasDefaults {
		return &Config{Enable: true, Options: map[string]any{}}
	}
	defaults := copyConfig(s.defaults)
	return &defaults
}

// SetConfig validates option names and stores the supplied configuration.
func (s *ConfigSupport) SetConfig(config *Config) error {
	for key := range config.Options {
		if _, ok := s.allowedOptions[key]; !ok {
			return fmt.Errorf("unknown option %q", key)
		}
	}
	s.config = config
	return nil
}

func copyConfig(config Config) Config {
	options := make(map[string]any, len(config.Options))
	maps.Copy(options, config.Options)
	config.Options = options
	return config
}
