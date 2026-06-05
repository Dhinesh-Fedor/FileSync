package synchronizer

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"FileSync/internal/models"
)

func Sync(sourceRoot string, destRoot string, result models.CompareResult) (models.SyncReport, error) {

	report := models.SyncReport{
		StartTime: time.Now(),
	}

	for _, change := range result.Changes {

		switch change.Type {

		case models.Added:

			src := filepath.Join(sourceRoot, change.RelativePath)
			dest := filepath.Join(destRoot, change.RelativePath)

			bytesCopied, err := copyFile(src, dest)
			if err != nil {
				return report, err
			}

			report.Added++
			report.BytesCopied += bytesCopied
			report.AddedFiles = append(report.AddedFiles, change.RelativePath)

		case models.Modified:

			src := filepath.Join(sourceRoot, change.RelativePath)
			dest := filepath.Join(destRoot, change.RelativePath)

			bytesCopied, err := copyFile(src, dest)
			if err != nil {
				return report, err
			}

			report.Modified++
			report.BytesCopied += bytesCopied
			report.ModifiedFiles = append(report.ModifiedFiles, change.RelativePath)

		case models.Deleted:

			dest := filepath.Join(destRoot, change.RelativePath)

			err := os.Remove(dest)
			if err != nil && !os.IsNotExist(err) {
				return report, err
			}

			report.Deleted++
			report.DeletedFiles = append(report.DeletedFiles, change.RelativePath)
		}
	}

	report.EndTime = time.Now()

	return report, nil
}

func copyFile(src string, dest string) (int64, error) {

	err := os.MkdirAll(filepath.Dir(dest), 0755)
	if err != nil {
		return 0, err
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return 0, err
	}
	defer destFile.Close()

	bytesCopied, err := io.Copy(destFile, sourceFile)
	if err != nil {
		return 0, err
	}

	return bytesCopied, nil
}
