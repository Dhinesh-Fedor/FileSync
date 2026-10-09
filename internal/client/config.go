package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"FileSync/internal/models"
)

const ConfigFile = "storage/client/config/config.json"

const clientCacheDir = "storage/client/cache"

func ensureClientStorage() error {
	dirs := []string{
		filepath.Dir(ConfigFile),
		clientCacheDir,
		"storage/client/logs",
		"storage/client/metadata",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

func LoadConfig() (models.ClientConfig, error) {
	if err := ensureClientStorage(); err != nil {
		return models.ClientConfig{}, err
	}

	data, err := os.ReadFile(ConfigFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return models.ClientConfig{}, nil
		}
		return models.ClientConfig{}, err
	}

	var config models.ClientConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return models.ClientConfig{}, err
	}

	return config, nil
}

func SaveConfig(config models.ClientConfig) error {
	if err := ensureClientStorage(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return writeAtomic(ConfigFile, data, 0644)
}

func writeAtomic(path string, data []byte, mode os.FileMode) error {
	temp, err := os.CreateTemp(filepath.Dir(path), ".tmp-")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	if err := temp.Chmod(mode); err != nil {
		temp.Close()
		return err
	}
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempName, path)
}

func DeviceID() string {
	host, _ := os.Hostname()
	if host == "" {
		host = "device"
	}

	return strings.ToLower(fmt.Sprintf("%s-%d", host, time.Now().UnixNano()))
}

func RegisterFolder(name, localPath, serverURL string) (models.RegisteredFolder, error) {
	config, err := LoadConfig()
	if err != nil {
		return models.RegisteredFolder{}, err
	}

	if config.DeviceID == "" {
		config.DeviceID = DeviceID()
	}
	if serverURL != "" {
		config.ServerURL = serverURL
	}

	folder := models.RegisteredFolder{
		Name:      name,
		LocalPath: filepath.Clean(localPath),
		DeviceID:  config.DeviceID,
	}

	registered, err := registerFolder(config.ServerURL, folder)
	if err != nil {
		return models.RegisteredFolder{}, err
	}

	registered.LocalPath = filepath.Clean(localPath)
	registered.DeviceID = config.DeviceID
	config.Folders = append(config.Folders, registered)

	if err := SaveConfig(config); err != nil {
		return models.RegisteredFolder{}, err
	}

	return registered, nil
}

func registerFolder(serverURL string, folder models.RegisteredFolder) (models.RegisteredFolder, error) {
	body, err := json.Marshal(folder)
	if err != nil {
		return models.RegisteredFolder{}, err
	}

	req, err := http.NewRequest(http.MethodPost, strings.TrimRight(serverURL, "/")+"/register-folder", bytes.NewReader(body))
	if err != nil {
		return models.RegisteredFolder{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if token := authToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return models.RegisteredFolder{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.RegisteredFolder{}, fmt.Errorf("register folder failed: %s", resp.Status)
	}

	var registered models.RegisteredFolder
	if err := json.NewDecoder(resp.Body).Decode(&registered); err != nil {
		return models.RegisteredFolder{}, err
	}

	return registered, nil
}

func authToken() string {
	if token := strings.TrimSpace(os.Getenv("FILESYNC_AUTH_TOKEN")); token != "" {
		return token
	}
	config, err := LoadConfig()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(config.AuthToken)
}
