package metadata

import (
	"context"
	"runtime"
	"time"

	"FileSync/internal/models"
)

type Generator struct {
	workers int
}

func New(workers int) *Generator {

	if workers < 1 {
		workers = runtime.NumCPU()
	}

	return &Generator{
		workers: workers,
	}
}

func (g *Generator) Build(ctx context.Context, scan models.ScannedFiles) ([]models.FileInfo, models.DirMetaData, error) {


	files := make([]models.FileInfo, len(scan.Files))

	jobs := make(chan job, g.workers)
	results := make(chan result, g.workers)

	for i := 0; i < g.workers; i++ {
		go worker(
			ctx,
			scan.Root,
			jobs,
			results,
		)
	}

	go func() {

		defer close(jobs)

		for index, path := range scan.Files {

			select {

			case <-ctx.Done():
				return

			case jobs <- job{
				index:        index,
				relativePath: path,
			}:
			}
		}
	}()

	var totalSize int64

	for i := 0; i < len(scan.Files); i++ {

		select {

		case <-ctx.Done():
			return nil,
				models.DirMetaData{},
				ctx.Err()

		case r := <-results:

			if r.err != nil {
				return nil,
					models.DirMetaData{},
					r.err
			}

			files[r.index] = r.file

			totalSize += r.size
		}
	}

	dirMeta := models.DirMetaData{
		TotalFiles: scan.FilesCount,
		TotalDirs:  scan.DirCount,
		TotalSize:  totalSize,
		LastScan:   time.Now(),
		Root:       scan.Root,
	}

	return files, dirMeta, nil
}
