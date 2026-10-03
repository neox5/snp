package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	ConfigFileName = ".snpconfig.json"
)

// Save writes the layer to the config file in dir. Only stated settings are
// written, so defaults never end up in the file.
func Save(dir string, l Layer) error {
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, ConfigFileName), data, 0o644)
}

// loadLayer reads the config file in dir. found is false when there is none.
func loadLayer(dir string) (l Layer, found bool, err error) {
	data, err := os.ReadFile(filepath.Join(dir, ConfigFileName))
	if err != nil {
		if os.IsNotExist(err) {
			return Layer{}, false, nil
		}
		return Layer{}, false, err
	}
	if err = json.Unmarshal(data, &l); err != nil {
		return Layer{}, false, err
	}
	return l, true, nil
}
