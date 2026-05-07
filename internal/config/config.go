package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Job describes a single cron job to monitor.
type Job struct {
	Name            string `yaml:"name"`
	Schedule        string `yaml:"schedule"`
	IntervalMinutes int    `yaml:"interval_minutes"`
	AlertEmail      string `yaml:"alert_email"`
}

// Config is the top-level configuration structure.
type Config struct {
	CheckIntervalSeconds int    `yaml:"check_interval_seconds"`
	LogFile              string `yaml:"log_file"`
	Jobs                 []Job  `yaml:"jobs"`
}

const defaultCheckInterval = 60

// Load reads and parses the YAML config file at the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("config file not found: %s", path)
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if len(cfg.Jobs) == 0 {
		return nil, errors.New("config must define at least one job")
	}

	if cfg.CheckIntervalSeconds <= 0 {
		cfg.CheckIntervalSeconds = defaultCheckInterval
	}

	for i, j := range cfg.Jobs {
		if j.Name == "" {
			return nil, fmt.Errorf("job[%d] missing name", i)
		}
		if j.IntervalMinutes <= 0 {
			return nil, fmt.Errorf("job %q: interval_minutes must be > 0", j.Name)
		}
	}

	return &cfg, nil
}
