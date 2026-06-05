package comparator

import (
	"FileSync/internal/models"
)

func Compare(source, destination []models.FileInfo) models.CompareResult {
	src := make(map[string]models.FileInfo, len(source))
	dest := make(map[string]models.FileInfo, len(destination))

	for _, srcFile := range source {
		src[srcFile.RelativePath] = srcFile
	}

	for _, destFile := range destination {
		dest[destFile.RelativePath] = destFile
	}

	var result models.CompareResult

	// Added and Modified
	for path, srcFile := range src {
		destFile, exists := dest[path]

		if !exists {
			result.Changes = append(result.Changes, models.FileChange{
				Type:         models.Added,
				RelativePath: path,
				Source:       srcFile,
			})

			result.Added++
			continue
		}

		if differs(srcFile, destFile) {
			result.Changes = append(result.Changes, models.FileChange{
				Type:         models.Modified,
				RelativePath: path,
				Source:       srcFile,
				Destination:  destFile,
			})

			result.Modified++
		}

	}
	// Deleted
	for path, destFile := range dest {
		if _, exists := src[path]; !exists {
			result.Changes = append(result.Changes, models.FileChange{
				Type:         models.Deleted,
				RelativePath: path,
				Destination:  destFile,
			})

			result.Deleted++
		}


	}

	return result
}

func differs(source, destination models.FileInfo) bool {
	if source.Size != destination.Size {
		return true
	}

	if source.SHA256 != destination.SHA256 {
		return true
	}

	return false
}
