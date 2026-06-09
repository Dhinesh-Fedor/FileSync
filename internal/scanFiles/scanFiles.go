package scanFiles

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"FileSync/internal/models"
)

func shouldIgnore(name string) bool {

	if strings.HasPrefix(name, ".") {
		return true
	}

	return false
}

func Scanner(ctx context.Context, dirName string) (models.ScannedFiles, error) {

	result := models.ScannedFiles{
		Root: dirName,
	}

	root := filepath.Clean(dirName)

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {

		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if path == root {
			return nil
		}

		if shouldIgnore(d.Name()) {

			if d.IsDir() {
				return filepath.SkipDir
			} 


			return nil
		}

		if d.Type().IsDir() {
			result.DirCount++
			return nil
		}

		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}

		if !d.Type().IsRegular() {
			return nil
		}

		file, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		result.Files = append(result.Files, filepath.ToSlash(file))

		result.FilesCount++

		return nil
	})

	if err != nil {
		return models.ScannedFiles{}, err
	}

	return result, nil
}
