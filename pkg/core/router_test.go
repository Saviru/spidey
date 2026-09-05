package core

import (
	"testing"
)

func TestRouter(t *testing.T) {
	r := New()

	if r == nil {
		t.Fatal("Failed to init Server")
	}

	t.Log("Success")

}
