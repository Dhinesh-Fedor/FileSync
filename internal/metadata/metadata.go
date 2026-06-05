package metadata

import (
	"os"
	"path/filepath"
	"time"

	"FileSync/internal/models"
	"FileSync/utils/hash"
)

func Build(scannedFiles models.ScannedFiles) ([]models.FileInfo, models.DirMetaData, error) {
	result := make([]models.FileInfo, 0, len(scannedFiles.Files))

	var totalSize int64

	for _, relativePath := range scannedFiles.Files {

		absolutePath := filepath.Join(scannedFiles.Root, relativePath)

		info, err := os.Stat(absolutePath)
		if err != nil {
			return nil, models.DirMetaData{}, err
		}

		totalSize += info.Size()

		hashed, err := hash.HashFile(absolutePath)
		if err != nil {
			return nil, models.DirMetaData{}, err
		}

		fileInfo := models.FileInfo{
			RelativePath: relativePath,
			AbsolutePath: absolutePath,
			Size:         info.Size(),
			ModTime:      info.ModTime(),
			SHA256:       hashed,
		}

		result = append(result, fileInfo)
	}

	dirMetaData := models.DirMetaData{
		TotalFiles: scannedFiles.FilesCount,
		TotalDirs:  scannedFiles.DirCount,
		TotalSize:  totalSize,
		LastScan:   time.Now(),
	}

	return result, dirMetaData, nil
}
