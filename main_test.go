package main

import (
	"testing"
)

func TestEnvOrDefault(t *testing.T) {
	d := "ooglyboogly"
	v := envOrDefault("OH_GO_I_HOPE_THIS_ENV_VAR_DOESNT_EXIST", d)

	if d != v {
		t.Errorf("expected %q, received %q", d, v)
	}
}

func TestPrepareServer(t *testing.T) {
	_, err := prepareServer()
	if err != nil {
		t.Error(err)
	}
}
