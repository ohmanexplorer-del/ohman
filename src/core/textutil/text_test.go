package textutil

import (
	"strings"
	"testing"
)

func TestNormalizeTextRepairsMojibake(t *testing.T) {
	input := "NowenReader æ˜¯ä¸€æ¬¾è½»é‡�å¼€æº�ã€�æœ¬åœ°åŒ–ä¼˜å…ˆçš„æ¼«ç”»"
	got := NormalizeText(input)
	if strings.Contains(got, "æ˜") || strings.Contains(got, "ä¸") {
		t.Fatalf("expected mojibake to be repaired, got %q", got)
	}
	if !strings.Contains(got, "NowenReader") || !strings.Contains(got, "是一款") {
		t.Fatalf("expected readable text, got %q", got)
	}
}

func TestNormalizeTextLeavesReadableText(t *testing.T) {
	input := "Lightweight local-first comic reader"
	if got := NormalizeText(input); got != input {
		t.Fatalf("expected readable text unchanged, got %q", got)
	}
}
