package main

import "testing"

func TestConfigPathDefaultsToConfigINI(t *testing.T) {
	t.Setenv("SONOLUS_CONFIG", "")

	if got := configPath(nil); got != "config.ini" {
		t.Fatalf("config path=%q, want config.ini", got)
	}
}

func TestConfigPathReadsEnvironment(t *testing.T) {
	t.Setenv("SONOLUS_CONFIG", "env.ini")

	if got := configPath(nil); got != "env.ini" {
		t.Fatalf("config path=%q, want env.ini", got)
	}
}

func TestConfigPathLongFlagOverridesEnvironment(t *testing.T) {
	t.Setenv("SONOLUS_CONFIG", "env.ini")

	if got := configPath([]string{"--config", "custom.ini"}); got != "custom.ini" {
		t.Fatalf("config path=%q, want custom.ini", got)
	}
}

func TestConfigPathShortFlagOverridesEnvironment(t *testing.T) {
	t.Setenv("SONOLUS_CONFIG", "env.ini")

	if got := configPath([]string{"-c", "custom.ini"}); got != "custom.ini" {
		t.Fatalf("config path=%q, want custom.ini", got)
	}
}
