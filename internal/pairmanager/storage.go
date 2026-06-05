package pairmanager

import (
	"encoding/json"
	"os"

	"FileSync/internal/models"
)

const ConfigFile = "configs/pairs.json"

func Load() ([]models.SyncPair, error) {

	if _, err := os.Stat(ConfigFile); os.IsNotExist(err) {
		return []models.SyncPair{}, nil
	}

	data, err := os.ReadFile(ConfigFile)
	if err != nil {
		return nil, err
	}

	var pairs []models.SyncPair

	err = json.Unmarshal(data, &pairs)

	return pairs, err
}

func Save(pairs []models.SyncPair) error {

	if err := os.MkdirAll("configs", 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(pairs, "", "  ")

	if err != nil {
		return err
	}

	return os.WriteFile(ConfigFile, data, 0644)
}
