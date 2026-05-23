package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	ConfigFileName = ".snpconfig.json"
)

func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func LoadConfig(p Params) (*Config, error) {
	fromFile, err := loadConfigFromSourceDir(p.SourceDir)
	if err != nil {
		return nil, err
	}
	fromCli := p.ToConfig()
	if fromFile == nil {
		return fromCli, nil
	}
	return fromFile.Merge(*fromCli), nil
}

func loadConfigFromSourceDir(srcDir string) (*Config, error) {
	path := filepath.Join(srcDir, ConfigFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
