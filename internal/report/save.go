package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"FileSync/internal/models"
)

func Save(pairID int, report models.SyncReport) error {
	return saveFile(filepath.Join("storage", "client", "logs", fmt.Sprintf("pair-%d.json", pairID)), report)
}

func SaveName(name string, report models.SyncReport) error {
	return saveFile(filepath.Join("storage", "client", "logs", fmt.Sprintf("%s.json", name)), report)
}

func saveFile(fileName string, report models.SyncReport) error {

	err := os.MkdirAll(filepath.Dir(fileName), 0755)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	temp, err := os.CreateTemp(filepath.Dir(fileName), ".tmp-")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	if err := temp.Chmod(0644); err != nil {
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
	return os.Rename(tempName, fileName)
}
