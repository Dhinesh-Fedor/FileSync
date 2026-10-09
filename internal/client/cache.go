package client

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"FileSync/internal/models"
)

func saveCache(folderID string, files []models.FileInfo) error {
	if err := ensureClientStorage(); err != nil {
		return err
	}

	path := cachePath(folderID)
	data, err := json.MarshalIndent(files, "", "  ")
	if err != nil {
		return err
	}

	return writeAtomic(path, data, 0644)
}

func cachePath(folderID string) string {
	cleanID := strings.NewReplacer("/", "-", string(filepath.Separator), "-").Replace(folderID)
	return filepath.Join(clientCacheDir, cleanID+".json")
}
