package collector

import (
	"testing"
)

func TestNewDedicatedServerCollector(t *testing.T) {
	target := "12345"
	c := NewDedicatedServerCollector(target)

	if c.target != target {
		t.Errorf("Expected target %s, got %s", target, c.target)
	}

	if c.servers == nil {
		t.Error("Collector Desc (servers) should not be nil")
	}
}
