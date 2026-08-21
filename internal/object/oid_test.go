package object_test

import (
	"testing"

	"github.com/LYH2263/go-gitmirror/internal/object"
)

func TestHashBlob(t *testing.T) {
	id := object.HashBlob([]byte("hello"))
	if id.Zero() {
		t.Fatal("zero")
	}
	if !object.IsHexOID(id.String()) {
		t.Fatal(id.String())
	}
}
