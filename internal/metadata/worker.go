package metadata

import (
	"context"

	"FileSync/internal/models"
)

type job struct {
	index        int
	relativePath string
}

type result struct {
	index int
	file  models.FileInfo
	size  int64
	err   error
}

func worker(ctx context.Context, root string, jobs <-chan job, results chan<- result) {

	for {

		select {

		case <-ctx.Done():
			return

		case j, ok := <-jobs:

			if !ok {
				return
			}

			fileInfo, size, err := build(
				root,
				j.relativePath,
			)

			results <- result{
				index: j.index,
				file:  fileInfo,
				size:  size,
				err:   err,
			}
		}
	}
}
