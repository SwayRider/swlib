package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGrpcConfig_AllowCredentialsDefaultsFalse(t *testing.T) {
	cfg := NewGrpcConfig(0, nil)
	assert.False(t, cfg.AllowCredentials)
}

func TestGrpcConfig_SetAllowCredentials(t *testing.T) {
	cfg := NewGrpcConfig(0, nil)
	cfg.SetAllowCredentials(true)
	assert.True(t, cfg.AllowCredentials)
}

func TestValidateCORSOrigins(t *testing.T) {
	tests := []struct {
		name    string
		origins []string
		wantErr bool
	}{
		{"single origin", []string{"http://localhost:5173"}, false},
		{"scoped wildcard subdomain allowed", []string{"https://*.swayrider.com"}, false},
		{"bare wildcard rejected", []string{"*"}, true},
		{"bare wildcard among others rejected", []string{"http://localhost:5173", "*"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCORSOrigins(tt.origins)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
