package config

import "github.com/spf13/viper"

type RateLimitRule struct {
	Path     string  `mapstructure:"path"`
	Capacity int    `mapstructure:"capacity"`
	Rate     float64 `mapstructure:"rate"`
}

type RateLimitDefaultRule struct {
	Capacity int     `mapstructure:"capacity"`
	Rate     float64 `mapstructure:"rate"`
}

type RateLimitConfig struct {
	Enabled   bool                 `mapstructure:"enabled"`
	FailOpen  bool                 `mapstructure:"fail_open"`
	Algorithm string               `mapstructure:"algorithm"`
	Default   RateLimitDefaultRule `mapstructure:"default"`
	SkipPaths []string             `mapstructure:"skip_paths"`
	Rules     []RateLimitRule      `mapstructure:"rules"`

	ruleMap  map[string]RateLimitRule
	skipMap  map[string]struct{}
}

func LoadRateLimitConfig() RateLimitConfig {
	var cfg RateLimitConfig

	if err := viper.UnmarshalKey("rate_limit", &cfg); err != nil {
		cfg = RateLimitConfig{}
	}

	if cfg.Default.Capacity <= 0 {
		cfg.Default.Capacity = 10
	}
	if cfg.Default.Rate <= 0 {
		cfg.Default.Rate = 1
	}

	cfg.ruleMap = make(map[string]RateLimitRule)
	for _, rule := range cfg.Rules {
		if rule.Path == "" {
			continue
		}

		if rule.Capacity <= 0 {
			rule.Capacity = cfg.Default.Capacity
		}
		if rule.Rate <= 0 {
			rule.Rate = cfg.Default.Rate
		}

		cfg.ruleMap[rule.Path] = rule
	}

	cfg.skipMap = make(map[string]struct{})
	for _, path := range cfg.SkipPaths {
		if path == "" {
			continue
		}
		cfg.skipMap[path] = struct{}{}
	}

	return cfg
}

func (cfg RateLimitConfig) IsSkipPath(path string) bool {
	_, ok := cfg.skipMap[path]
	return ok
}

func (cfg RateLimitConfig) GetRule(path string) RateLimitRule {
	if rule, ok := cfg.ruleMap[path]; ok {
		return rule
	}

	return RateLimitRule{
		Path:     path,
		Capacity: cfg.Default.Capacity,
		Rate:     cfg.Default.Rate,
	}
}