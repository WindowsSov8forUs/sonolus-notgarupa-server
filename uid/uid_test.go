package uid

import (
	"regexp"
	"testing"
	"time"
)

func TestGenerateUsesEnginePrefixAndFormat(t *testing.T) {
	name, err := Generate("notgarupa", time.UnixMilli(1_717_171_717_171))
	if err != nil {
		t.Fatal(err)
	}
	pattern := regexp.MustCompile(`^notgarupa-[23456789abcdefghijkmnpqrstuvwxyz]{9}_[23456789abcdefghijkmnpqrstuvwxyz]{6}$`)
	if !pattern.MatchString(name) {
		t.Fatalf("name %q does not match expected format", name)
	}
}

func TestGenerateRejectsUnknownEngine(t *testing.T) {
	if _, err := Generate("unknown", time.UnixMilli(1_717_171_717_171)); err == nil {
		t.Fatal("expected missing engine prefix error")
	}
}
