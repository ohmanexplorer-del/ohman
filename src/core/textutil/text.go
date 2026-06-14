package textutil

import (
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
)

func NormalizeText(value string) string {
	value = strings.ToValidUTF8(value, "")
	value = repairMojibake(value)
	value = strings.ToValidUTF8(value, "")
	value = strings.Join(strings.Fields(value), " ")
	return strings.TrimSpace(value)
}

func repairMojibake(value string) string {
	if !looksMojibake(value) {
		return value
	}
	candidate, ok := decodeWindows1252AsUTF8(value)
	if !ok {
		return value
	}
	if mojibakeScore(candidate) >= mojibakeScore(value) {
		return value
	}
	return candidate
}

func decodeWindows1252AsUTF8(value string) (string, bool) {
	encoded, err := charmap.Windows1252.NewEncoder().String(value)
	if err != nil {
		cleaned := strings.ReplaceAll(value, "\uFFFD", "")
		encoded, err = charmap.Windows1252.NewEncoder().String(cleaned)
		if err != nil {
			return "", false
		}
	}
	bytes := []byte(encoded)
	if utf8.Valid(bytes) {
		return string(bytes), true
	}
	repaired := strings.ToValidUTF8(string(bytes), "")
	return repaired, repaired != ""
}

func looksMojibake(value string) bool {
	return mojibakeScore(value) >= 2
}

func mojibakeScore(value string) int {
	score := 0
	markers := []string{
		"Ã", "Â", "â€", "â€™", "â€œ", "â€", "â€¦", "ðŸ",
		"æ", "ä", "å", "ç", "è", "é", "ï¼", "ã€", "ã‚", "ãƒ",
	}
	for _, marker := range markers {
		score += strings.Count(value, marker)
	}
	score += strings.Count(value, "\uFFFD") * 2
	return score
}
