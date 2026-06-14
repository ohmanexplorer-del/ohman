package github

import "testing"

func TestParseRepoEvaluation(t *testing.T) {
	response := `Here is the result:
{
  "category": "web-crawler",
  "project_type": "library",
  "score": 8.2,
  "novelty": 7,
  "maturity": 8,
  "small_repo_fit": 6,
  "reason": "Clear purpose and healthy project signals.",
  "strengths": ["Focused API"],
  "weaknesses": ["Small ecosystem"],
  "publish": true
}`

	got, err := parseRepoEvaluation(response)
	if err != nil {
		t.Fatal(err)
	}

	if got.Category != "web-crawler" {
		t.Fatalf("category = %q", got.Category)
	}
	if got.Score != 8.2 {
		t.Fatalf("score = %v", got.Score)
	}
	if !got.Publish {
		t.Fatal("publish should be true")
	}
}
