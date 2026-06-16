package repository

import "testing"

func TestRewriteRepositoryURLs(t *testing.T) {
	db := map[string]any{
		"info": map[string]any{
			"banner": map[string]any{
				"hash": "abc",
				"url":  "/sonolus/repository/abc",
			},
		},
		"levels": []any{
			map[string]any{
				"bgm": map[string]any{
					"hash": "def",
					"url":  "/sonolus/repository/def",
				},
			},
		},
	}

	RewriteRepositoryURLs(db, "https://repo.example")

	if got := db["info"].(map[string]any)["banner"].(map[string]any)["url"]; got != "https://repo.example/sonolus/repository/abc" {
		t.Fatalf("banner url = %v", got)
	}
	if got := db["levels"].([]any)[0].(map[string]any)["bgm"].(map[string]any)["url"]; got != "https://repo.example/sonolus/repository/def" {
		t.Fatalf("bgm url = %v", got)
	}
}
