package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// Features contains the independently configurable tracking features.
type Features struct {
	KeywordTracking        bool
	AuthorTracking         bool
	PushSumTracking        bool
	ArticleCommentTracking bool
}

// Config is the validated application configuration.
type Config struct {
	Features Features
}

type configFile struct {
	Features featureFile `yaml:"features"`
}

type featureFile struct {
	KeywordTracking        *bool `yaml:"keyword_tracking"`
	AuthorTracking         *bool `yaml:"author_tracking"`
	PushSumTracking        *bool `yaml:"push_sum_tracking"`
	ArticleCommentTracking *bool `yaml:"article_comment_tracking"`
}

var (
	mu      sync.RWMutex
	current = Config{Features: Features{
		KeywordTracking:        true,
		AuthorTracking:         true,
		PushSumTracking:        true,
		ArticleCommentTracking: true,
	}}
)

// Load strictly reads and validates a YAML configuration file, then makes it
// available to the rest of the application.
func Load(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open config: %w", err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)

	var raw configFile
	if err := decoder.Decode(&raw); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}

	var extra interface{}
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return Config{}, errors.New("decode config: multiple YAML documents are not allowed")
		}
		return Config{}, fmt.Errorf("decode config: %w", err)
	}

	missing := make([]string, 0, 4)
	if raw.Features.KeywordTracking == nil {
		missing = append(missing, "features.keyword_tracking")
	}
	if raw.Features.AuthorTracking == nil {
		missing = append(missing, "features.author_tracking")
	}
	if raw.Features.PushSumTracking == nil {
		missing = append(missing, "features.push_sum_tracking")
	}
	if raw.Features.ArticleCommentTracking == nil {
		missing = append(missing, "features.article_comment_tracking")
	}
	if len(missing) != 0 {
		return Config{}, fmt.Errorf("missing required config fields: %s", strings.Join(missing, ", "))
	}

	loaded := Config{Features: Features{
		KeywordTracking:        *raw.Features.KeywordTracking,
		AuthorTracking:         *raw.Features.AuthorTracking,
		PushSumTracking:        *raw.Features.PushSumTracking,
		ArticleCommentTracking: *raw.Features.ArticleCommentTracking,
	}}

	Set(loaded)
	return loaded, nil
}

// Current returns the configuration loaded at application startup.
func Current() Config {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

// Set replaces the active configuration. Production configuration should use
// Load; Set also makes feature-dependent behavior straightforward to test.
func Set(cfg Config) {
	mu.Lock()
	current = cfg
	mu.Unlock()
}
