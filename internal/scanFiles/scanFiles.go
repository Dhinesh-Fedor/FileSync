package scanFiles

import (
	"context"
	"os"
	"path/filepath"

	"FileSync/internal/models"
)

func Scanner(ctx context.Context, dirName string) (models.ScannedFiles, error) {
	result := models.ScannedFiles{Root: dirName}
	root := filepath.Clean(dirName)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
    
		// exits the scanner if there is any interruption ( basically handles interruption, meh...)
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Skip the root
		if path == root {
			return nil
		}
		
    // count dirs
		if d.Type().IsDir() {
			result.DirCount++
			return nil
		}

		// Handles Symlink files - ignores them
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}

    // Skip irregular dir entries 
		if !d.Type().IsRegular() {
			return nil
		}

	  //if it is a file then adds
		// its relative path to the files array.
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
