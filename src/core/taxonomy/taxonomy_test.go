package taxonomy

import "testing"

func TestNormalizeCategoryAliases(t *testing.T) {
	tests := map[string]string{
		"AI Tooling":            "artificial-intelligence",
		"developer_tools":       "developer-tools",
		"database-library":      "database",
		"Kubernetes Tools":      "kubernetes",
		"Programming Languages": "developer-tools",
		"Go":                    "developer-tools",
		"network-security":      "networking",
		"cloud-security":        "security",
		"vpn-server":            "privacy",
		"random tiny bucket":    "uncategorized",
		"":                      "uncategorized",
	}

	for input, expected := range tests {
		if got := NormalizeCategory(input); got != expected {
			t.Fatalf("NormalizeCategory(%q) = %q, want %q", input, got, expected)
		}
	}
}

func TestCategoriesReturnsCopy(t *testing.T) {
	categories := Categories()
	if len(categories) == 0 {
		t.Fatal("expected categories")
	}
	categories[0] = "changed"
	if Categories()[0] == "changed" {
		t.Fatal("expected copy")
	}
}
