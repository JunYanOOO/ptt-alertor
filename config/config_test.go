package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validConfig = `features:
  keyword_tracking: true
  author_tracking: false
  push_sum_tracking: false
  article_comment_tracking: false
`

func TestLoad(t *testing.T) {
	previous := Current()
	t.Cleanup(func() { Set(previous) })

	path := writeConfig(t, validConfig)
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !loaded.Features.KeywordTracking {
		t.Error("keyword tracking should be enabled")
	}
	if loaded.Features.AuthorTracking || loaded.Features.PushSumTracking || loaded.Features.ArticleCommentTracking {
		t.Error("only keyword tracking should be enabled")
	}
}

func TestLoadRejectsInvalidConfig(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name: "missing field",
			content: `features:
  keyword_tracking: true
  author_tracking: false
  push_sum_tracking: false
`,
			want: "features.article_comment_tracking",
		},
		{
			name:    "unknown field",
			content: validConfig + "  unexpected: true\n",
			want:    "field unexpected not found",
		},
		{
			name:    "invalid type",
			content: strings.Replace(validConfig, "keyword_tracking: true", "keyword_tracking: enabled", 1),
			want:    "cannot unmarshal",
		},
		{
			name:    "malformed yaml",
			content: "features: [",
			want:    "decode config",
		},
		{
			name:    "multiple documents",
			content: validConfig + "---\nfeatures: {}\n",
			want:    "multiple YAML documents",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Load(writeConfig(t, test.content))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Load() error = %v, want error containing %q", err, test.want)
			}
		})
	}
}

func TestLoadRejectsMissingFile(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "missing.yaml"))
	if err == nil || !strings.Contains(err.Error(), "open config") {
		t.Fatalf("Load() error = %v, want open config error", err)
	}
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}
