package main

import (
	"testing"
)

func TestConfigMergeBackendOverwrite(t *testing.T) {
	cfg1, cfg2 := NewConfig(), NewConfig()

	cfg1.Backend = NewBackend("foo", nil)
	cfg2.Backend = NewBackend("bar", nil)

	cfg1.Merge(cfg2)
	if cfg1.Backend.Name != "bar" {
		t.Logf("Backend was not overwritten (Value: %s)\n", cfg1.Backend.Name)
		t.Fail()
	}
}

func TestConfigMergeBackendDontOverwriteWithEmpty(t *testing.T) {
	cfg1, cfg2 := NewConfig(), NewConfig()

	cfg1.Backend = NewBackend("foo", nil)

	cfg1.Merge(cfg2)
	if cfg1.Backend.Name != "foo" {
		t.Log("Backend was overwritten")
		t.Fail()
	}
}

func TestConfigMergeSources(t *testing.T) {
	cfg1, cfg2 := NewConfig(), NewConfig()

	cfg1.Sources = []*Source{NewSource("a", nil), NewSource("b", nil)}
	cfg2.Sources = []*Source{NewSource("c", nil), NewSource("d", nil)}

	cfg1.Merge(cfg2)
	if len(cfg1.Sources) != 4 {
		t.Logf("Sources length not 4 (length '%d')\n", len(cfg1.Sources))
		t.Fail()
	}
}

func TestConfigMergeOptions(t *testing.T) {
	cfg1, cfg2 := NewConfig(), NewConfig()

	cfg1.Options["foo"] = "bar"
	cfg2.Options["foo"] = "qux"
	cfg2.Options["bar"] = "baz"

	cfg1.Merge(cfg2)
	if cfg1.Options["foo"] != "qux" {
		t.Logf("Options[\"foo\"] doesn't match \"%s\" != \"qux\"", cfg1.Options["foo"])
		t.Fail()
	}
	if cfg1.Options["bar"] != "baz" {
		t.Logf("Options[\"bar\"] doesn't match \"%s\" != \"baz\"", cfg1.Options["bar"])
		t.Fail()
	}
}
