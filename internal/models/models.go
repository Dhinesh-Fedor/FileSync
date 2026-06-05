package models

import (
	"time"
)

type ScannedFiles struct {
	Files []string
	DirCount int
	FilesCount int
	Root string
}


type DirMetaData struct {
	Root string
	TotalFiles int
	TotalDirs int
	TotalSize int64
	LastScan time.Time
}

type FileInfo struct {
	RelativePath string
	AbsolutePath string
	Size         int64
	ModTime      time.Time
	SHA256       string
}

type ChangeType string

const (
	Added ChangeType = "Added"
	Modified ChangeType = "Modified"
	Deleted ChangeType = "Deleted"
)

type FileChange struct {
	Type ChangeType
	RelativePath string
	Source          FileInfo
	Destination      FileInfo
}

type CompareResult struct {
	Changes []FileChange
	Added int
	Modified int
	Deleted int
}

type SyncPair struct {
	ID          int       `json:"id"`
	Source      string    `json:"source"`
	Destination string    `json:"destination"`
	CreatedAt   time.Time `json:"created_at"`
}

type SyncReport struct {
	Added int `json:"added"`
	Modified int `json:"modified"`
	Deleted int `json:"deleted"`

	BytesCopied int64 `json:"bytesCopied"`

	AddedFiles []string `json:"addedFiles"`
	ModifiedFiles []string `json:"modifiedFiles"`
	DeletedFiles []string `json:"deletedFiles"`

	StartTime time.Time `json:"startTime"`
	EndTime time.Time `json:"endTime"`
}



