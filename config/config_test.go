package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadClientConfig(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "client.yaml")
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("server:\n")
	tmpFile.WriteString("  listen-addr: \":4443\"\n")
	tmpFile.Close()
	mgr, err := NewManager(tmpFile.Name())
	defer mgr.Stop()
	cfg := mgr.Get()
	assert.NoError(t, err)
	assert.Equal(t, ":4443", cfg.Server.ListenAddr)
}
