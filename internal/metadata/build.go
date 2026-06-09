package metadata

import (
	"os"
	"path/filepath"

	"FileSync/internal/models"
	"FileSync/utils/hash"
)

func build(root string, relativePath string) (models.FileInfo, int64, error) {
	absolutePath := filepath.Join(root, relativePath)

	info, err := os.Stat(absolutePath)
	if err != nil {
		return models.FileInfo{}, 0, err
	}

	hash, err := hash.HashFile(absolutePath)
	if err != nil {
		return models.FileInfo{}, 0, err
	}

	fileInfo := models.FileInfo{
		RelativePath: relativePath,
		AbsolutePath: absolutePath,
		Size:         info.Size(),
		ModTime:      info.ModTime(),
		SHA256:       hash,
	}

	return fileInfo, info.Size(), nil
}
