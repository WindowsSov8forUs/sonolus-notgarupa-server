package repository

import "strings"

func RewriteRepositoryURLs(value any, publicURL string) {
	switch typed := value.(type) {
	case map[string]any:
		if url, ok := typed["url"].(string); ok && strings.HasPrefix(url, "/sonolus/repository/") {
			typed["url"] = strings.TrimRight(publicURL, "/") + url
		}
		for _, child := range typed {
			RewriteRepositoryURLs(child, publicURL)
		}
	case []any:
		for _, child := range typed {
			RewriteRepositoryURLs(child, publicURL)
		}
	}
}
