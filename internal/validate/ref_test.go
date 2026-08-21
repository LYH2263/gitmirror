package validate_test

import (
	"testing"

	"github.com/LYH2263/go-gitmirror/internal/validate"
)

func TestRefName(t *testing.T) {
	if err := validate.RefName("refs/heads/main"); err != nil {
		t.Fatal(err)
	}
	if err := validate.RefName("heads/main"); err == nil {
		t.Fatal("expected error")
	}
	if err := validate.RemoteURL("https://example.com/a.git"); err != nil {
		t.Fatal(err)
	}
	if err := validate.RemoteURL("memory://x"); err != nil {
		t.Fatal(err)
	}
}
