package uuid

import (
	"regexp"
	"testing"
)

func TestGenerateUsesShortUUIDFormat(t *testing.T) {
	id, err := Generate()
	if err != nil {
		t.Fatal(err)
	}

	pattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}$`)
	if !pattern.MatchString(id) {
		t.Fatalf("uuid %q does not match expected format", id)
	}
}
