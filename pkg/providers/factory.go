package providers

import (
	"fmt"
)

// NewProvider creates a provider based on configuration
func NewProvider(providerType Type, config *Config) (Provider, error) {
	switch providerType {
	case TypeTraefik:
		return NewTraefikProvider(config), nil
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}
}
