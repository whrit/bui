package syntax

import "testing"

func TestHighlightDiff_DoesNotPanic(t *testing.T) {
	raw := "diff --git a/main.go b/main.go\n@@\n- old\n+ new\n"
	_, err := HighlightDiff(raw, "github-dark", "Go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
