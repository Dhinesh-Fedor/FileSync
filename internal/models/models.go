package models

import (
	"time"
)

type ScannedFiles struct {
	Files      []string
	DirCount   int
	FilesCount int
	Root       string
}

type DirMetaData struct {
	Root       string
	TotalFiles int
	TotalDirs  int
	TotalSize  int64
	LastScan   time.Time
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
	Added    ChangeType = "Added"
	Modified ChangeType = "Modified"
	Deleted  ChangeType = "Deleted"
)

type FileChange struct {
	Type         ChangeType
	RelativePath string
	Source       FileInfo
	Destination  FileInfo
}

type CompareResult struct {
	Changes  []FileChange
	Added    int
	Modified int
	Deleted  int
}

type SyncPair struct {
	ID          int       `json:"id"`
	Source      string    `json:"source"`
	Destination string    `json:"destination"`
	CreatedAt   time.Time `json:"created_at"`
}

type RegisteredFolder struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	LocalPath      string    `json:"local_path"`
	RemoteID       string    `json:"remote_id"`
	CurrentVersion int64     `json:"current_version"`
	LastSyncedAt   time.Time `json:"last_synced_at"`
	DeviceID       string    `json:"device_id"`
}

type ClientConfig struct {
	DeviceID  string             `json:"device_id"`
	ServerURL string             `json:"server_url"`
	AuthToken string             `json:"auth_token,omitempty"`
	Folders   []RegisteredFolder `json:"folders"`
}

type ServerFile struct {
	RelativePath string    `json:"relative_path"`
	Hash         string    `json:"hash"`
	Size         int64     `json:"size"`
	ModifiedTime time.Time `json:"modified_time"`
	Version      int64     `json:"version"`
	FolderID     string    `json:"folder_id"`
	DeviceID     string    `json:"device_id"`
}

type SyncRequest struct {
	DeviceID     string           `json:"device_id"`
	Folder       RegisteredFolder `json:"folder"`
	KnownVersion int64            `json:"known_version"`
	Files        []FileInfo       `json:"files"`
}

type ConflictItem struct {
	RelativePath string `json:"relative_path"`
	ConflictPath string `json:"conflict_path"`
	Reason       string `json:"reason"`
}

type SyncPlan struct {
	FolderID      string         `json:"folder_id"`
	Version       int64          `json:"version"`
	UploadPaths   []string       `json:"upload_paths"`
	UploadChanges []FileChange   `json:"upload_changes"`
	DownloadFiles []ServerFile   `json:"download_files"`
	DeletePaths   []string       `json:"delete_paths"`
	Conflicts     []ConflictItem `json:"conflicts"`
	Message       string         `json:"message"`
}

type UploadRequest struct {
	FolderID     string    `json:"folder_id"`
	DeviceID     string    `json:"device_id"`
	RelativePath string    `json:"relative_path"`
	Hash         string    `json:"hash"`
	Size         int64     `json:"size"`
	ModifiedTime time.Time `json:"modified_time"`
	Version      int64     `json:"version"`
}

type FolderStatus struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Version   int64        `json:"version"`
	Files     []ServerFile `json:"files"`
	Devices   []string     `json:"devices"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type ServerStatus struct {
	Folders   []FolderStatus `json:"folders"`
	BlobCount int            `json:"blob_count"`
	DataDir   string         `json:"data_dir"`
}

type SyncReport struct {
	Added    int `json:"added"`
	Modified int `json:"modified"`
	Deleted  int `json:"deleted"`

	BytesCopied int64 `json:"bytesCopied"`

	AddedFiles    []string `json:"addedFiles"`
	ModifiedFiles []string `json:"modifiedFiles"`
	DeletedFiles  []string `json:"deletedFiles"`

	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}
