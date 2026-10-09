package client

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"

	"FileSync/internal/metadata"
	"FileSync/internal/models"
	"FileSync/internal/scanFiles"
)

func ListFolders() ([]models.RegisteredFolder, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	return config.Folders, nil
}

func GetFolder(id string) (*models.RegisteredFolder, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	for _, folder := range config.Folders {
		if folder.ID == id || folder.RemoteID == id || filepath.Clean(folder.LocalPath) == filepath.Clean(id) {
			copy := folder
			return &copy, nil
		}
	}

	return nil, nil
}

func DeleteFolder(id string) error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	updated := make([]models.RegisteredFolder, 0, len(config.Folders))
	removed := false
	for _, folder := range config.Folders {
		if folder.ID == id || folder.RemoteID == id || filepath.Clean(folder.LocalPath) == filepath.Clean(id) {
			removed = true
			continue
		}
		updated = append(updated, folder)
	}

	if !removed {
		return errors.New("folder not found")
	}

	config.Folders = updated
	return SaveConfig(config)
}

func PlanFolder(folderID string) (models.SyncPlan, error) {
	config, folder, files, err := prepareSyncData(folderID)
	if err != nil {
		return models.SyncPlan{}, err
	}

	return requestPlan(config.ServerURL, models.SyncRequest{
		DeviceID:     config.DeviceID,
		Folder:       folder,
		KnownVersion: folder.CurrentVersion,
		Files:        files,
	})
}

func prepareSyncData(folderID string) (models.ClientConfig, models.RegisteredFolder, []models.FileInfo, error) {
	config, err := LoadConfig()
	if err != nil {
		return models.ClientConfig{}, models.RegisteredFolder{}, nil, err
	}

	folder, err := findFolder(config, folderID)
	if err != nil {
		return models.ClientConfig{}, models.RegisteredFolder{}, nil, err
	}

	scan, err := scanFolder(folder.LocalPath)
	if err != nil {
		return models.ClientConfig{}, models.RegisteredFolder{}, nil, err
	}

	return config, folder, scan, nil
}

func scanFolder(folderPath string) ([]models.FileInfo, error) {
	result, err := scanFiles.Scanner(context.Background(), folderPath)
	if err != nil {
		return nil, err
	}

	generator := metadata.New(runtime.NumCPU())
	files, _, err := generator.Build(context.Background(), result)
	if err != nil {
		return nil, err
	}

	return files, nil
}
