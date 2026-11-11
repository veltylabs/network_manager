//go:build wasm

package tests

import (
	"testing"
	"github.com/veltylabs/network_manager/seed"
)

func TestDemo_Valid(t *testing.T) {
	d := seed.Demo()
	if d.Label == "" {
		t.Errorf("demo label is empty")
	}
}
