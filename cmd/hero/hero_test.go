package main

import (
	"testing"
	"os"
)

func TestHero(t *testing.T) {
	testCfg := "config_test.json"

	defer func() {
		os.Remove(testCfg)
	}()
	// writing config files
	err := writeConfig(defaultCfg, testCfg)
	if err != nil {
		t.Error(err)
	}

	cfg, err := getConfig(testCfg)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.AuthEndpoint != defaultCfg.AuthEndpoint {
		t.Error("expected %s got %s", defaultCfg.AuthEndpoint, cfg.AuthEndpoint)
	}

}
