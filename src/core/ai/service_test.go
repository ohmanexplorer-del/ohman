package ai

import "testing"

func TestOpenAIBaseURL(t *testing.T) {
	tests := map[string]string{
		"http://localhost:8080":                     "http://localhost:8080/v1",
		"http://localhost:8080/":                    "http://localhost:8080/v1",
		"http://localhost:8080/v1":                  "http://localhost:8080/v1",
		"http://localhost:8080/v1/":                 "http://localhost:8080/v1",
		"http://localhost:8080/v1/chat/completions": "http://localhost:8080/v1",
	}

	for input, expected := range tests {
		if got := openAIBaseURL(input); got != expected {
			t.Fatalf("openAIBaseURL(%q) = %q, want %q", input, got, expected)
		}
	}
}
