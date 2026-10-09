package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"FileSync/internal/models"
)

const defaultRoot = "storage/server"

type diskState struct {
	Folders map[string]*folderState `json:"folders"`
}

type folderState struct {
	ID         string                       `json:"id"`
	Name       string                       `json:"name"`
	Version    int64                        `json:"version"`
	UpdatedAt  time.Time                    `json:"updated_at"`
	Devices    map[string]time.Time         `json:"devices"`
	Files      map[string]models.ServerFile `json:"files"`
	Tombstones map[string]int64             `json:"tombstones,omitempty"`
}

type Store struct {
	mu   sync.Mutex
	root string
	data diskState
}

func NewStore(root string) (*Store, error) {
	if root == "" {
		root = defaultRoot
	}

	store := &Store{root: root}
	store.data.Folders = map[string]*folderState{}

	if err := os.MkdirAll(filepath.Join(root, "blobs"), 0755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(root, "metadata"), 0755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(root, "versions"), 0755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(root, "logs"), 0755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(root, "uploads"), 0755); err != nil {
		return nil, err
	}

	if err := store.load(); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *Store) load() error {
	path := s.statePath()
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}

	return json.Unmarshal(data, &s.data)
}

func (s *Store) save() error {
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}

	temp, err := os.CreateTemp(filepath.Dir(s.statePath()), "index-*")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	if err := temp.Chmod(0644); err != nil {
		temp.Close()
		return err
	}
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempName, s.statePath())
}

func (s *Store) statePath() string {
	return filepath.Join(s.root, "metadata", "index.json")
}

func (s *Store) blobPath(hash string) string {
	return filepath.Join(s.root, "blobs", hash)
}

func (s *Store) ensureFolder(folder models.RegisteredFolder) *folderState {
	if folder.ID == "" {
		folder.ID = fmt.Sprintf("folder-%d", time.Now().UnixNano())
	}

	state, ok := s.data.Folders[folder.ID]
	if !ok {
		state = &folderState{
			ID:         folder.ID,
			Name:       folder.Name,
			Devices:    map[string]time.Time{},
			Files:      map[string]models.ServerFile{},
			Tombstones: map[string]int64{},
		}
		s.data.Folders[folder.ID] = state
	}

	if state.Devices == nil {
		state.Devices = map[string]time.Time{}
	}
	if state.Files == nil {
		state.Files = map[string]models.ServerFile{}
	}
	if state.Tombstones == nil {
		state.Tombstones = map[string]int64{}
	}
	if folder.Name != "" {
		state.Name = folder.Name
	}
	if folder.DeviceID != "" {
		state.Devices[folder.DeviceID] = time.Now()
	}

	return state
}

func (s *Store) RegisterFolder(folder models.RegisteredFolder) (models.RegisteredFolder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := s.ensureFolder(folder)
	state.UpdatedAt = time.Now()
	if err := s.save(); err != nil {
		return models.RegisteredFolder{}, err
	}

	return models.RegisteredFolder{
		ID:             state.ID,
		Name:           state.Name,
		LocalPath:      folder.LocalPath,
		RemoteID:       state.ID,
		CurrentVersion: state.Version,
		LastSyncedAt:   state.UpdatedAt,
		DeviceID:       folder.DeviceID,
	}, nil
}

func (s *Store) Sync(req models.SyncRequest) (models.SyncPlan, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, ok := s.data.Folders[req.Folder.ID]
	if !ok {
		return models.SyncPlan{}, errors.New("folder not found")
	}
	if req.DeviceID == "" {
		return models.SyncPlan{}, errors.New("missing device id")
	}
	if _, ok := state.Devices[req.DeviceID]; !ok {
		return models.SyncPlan{}, errors.New("device is not registered for folder")
	}
	current := state.Version
	known := req.KnownVersion

	local := make(map[string]models.FileInfo, len(req.Files))
	for _, file := range req.Files {
		local[file.RelativePath] = file
	}

	plan := models.SyncPlan{FolderID: state.ID, Version: current}
	for path, localFile := range local {
		serverFile, exists := state.Files[path]
		if !exists {
			if tombstoneVersion, deleted := state.Tombstones[path]; deleted && current > known && tombstoneVersion > known {
				plan.Conflicts = append(plan.Conflicts, models.ConflictItem{
					RelativePath: path,
					ConflictPath: conflictName(path, req.DeviceID),
					Reason:       "server deleted the file before this client uploaded it",
				})
				continue
			}
			plan.UploadPaths = append(plan.UploadPaths, path)
			plan.UploadChanges = append(plan.UploadChanges, models.FileChange{
				Type:         models.Added,
				RelativePath: path,
				Source: models.FileInfo{
					RelativePath: path,
					Size:         localFile.Size,
					ModTime:      localFile.ModTime,
					SHA256:       localFile.SHA256,
				},
			})
			continue
		}

		if serverFile.Hash == localFile.SHA256 {
			continue
		}

		if current > known {
			conflictPath := conflictName(path, req.DeviceID)
			plan.Conflicts = append(plan.Conflicts, models.ConflictItem{
				RelativePath: path,
				ConflictPath: conflictPath,
				Reason:       "server version changed before local upload",
			})
			continue
		}

		plan.UploadPaths = append(plan.UploadPaths, path)
		plan.UploadChanges = append(plan.UploadChanges, models.FileChange{
			Type:         models.Modified,
			RelativePath: path,
			Source: models.FileInfo{
				RelativePath: path,
				Size:         localFile.Size,
				ModTime:      localFile.ModTime,
				SHA256:       localFile.SHA256,
			},
			Destination: models.FileInfo{
				RelativePath: path,
				Size:         serverFile.Size,
				ModTime:      serverFile.ModifiedTime,
				SHA256:       serverFile.Hash,
			},
		})
	}

	for path, serverFile := range state.Files {
		if _, ok := local[path]; ok {
			continue
		}

		if current > known {
			plan.DownloadFiles = append(plan.DownloadFiles, serverFile)
			continue
		}

		plan.DeletePaths = append(plan.DeletePaths, path)
	}

	plan.Version = current
	if len(plan.UploadChanges) > 0 || len(plan.DeletePaths) > 0 {
		plan.Message = "changes pending upload or deletion"
	} else {
		plan.Message = "already up to date"
	}

	if err := s.save(); err != nil {
		return models.SyncPlan{}, err
	}

	return plan, nil
}

func (s *Store) StoreBlob(hash string, body io.Reader) (int64, error) {
	if hash == "" {
		return 0, errors.New("missing hash")
	}
	if len(hash) != sha256.Size*2 {
		return 0, errors.New("invalid hash")
	}
	if _, err := hex.DecodeString(hash); err != nil {
		return 0, errors.New("invalid hash")
	}

	path := s.blobPath(hash)
	if _, err := os.Stat(path); err == nil {
		return io.Copy(io.Discard, body)
	}

	temp, err := os.CreateTemp(filepath.Dir(path), "upload-*")
	if err != nil {
		return 0, err
	}
	defer os.Remove(temp.Name())

	hasher := sha256.New()
	bytesWritten, err := io.Copy(io.MultiWriter(temp, hasher), body)
	if err != nil {
		temp.Close()
		return 0, err
	}
	if err := temp.Close(); err != nil {
		return 0, err
	}
	if hex.EncodeToString(hasher.Sum(nil)) != hash {
		return 0, errors.New("uploaded content hash mismatch")
	}
	if err := os.Rename(temp.Name(), path); err != nil {
		return 0, err
	}

	return bytesWritten, nil
}

func (s *Store) CommitUpload(folderID, deviceID, relativePath, hash string, size int64, modifiedTime time.Time) (models.ServerFile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := validateRelativePath(relativePath); err != nil {
		return models.ServerFile{}, err
	}
	if _, err := os.Stat(s.blobPath(hash)); err != nil {
		return models.ServerFile{}, errors.New("uploaded blob is not available")
	}
	state, ok := s.data.Folders[folderID]
	if !ok {
		return models.ServerFile{}, errors.New("folder not found")
	}
	if deviceID == "" {
		return models.ServerFile{}, errors.New("missing device id")
	}
	if _, ok := state.Devices[deviceID]; !ok {
		return models.ServerFile{}, errors.New("device is not registered for folder")
	}

	file := models.ServerFile{
		RelativePath: relativePath,
		Hash:         hash,
		Size:         size,
		ModifiedTime: modifiedTime,
		Version:      state.Version + 1,
		FolderID:     folderID,
		DeviceID:     deviceID,
	}
	state.Files[relativePath] = file
	delete(state.Tombstones, relativePath)
	state.Version++
	state.UpdatedAt = time.Now()
	if err := s.save(); err != nil {
		return models.ServerFile{}, err
	}
	return file, nil
}

func (s *Store) CommitDelete(folderID, deviceID, relativePath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := validateRelativePath(relativePath); err != nil {
		return err
	}
	state, ok := s.data.Folders[folderID]
	if !ok {
		return errors.New("folder not found")
	}
	if _, ok := state.Devices[deviceID]; !ok {
		return errors.New("device is not registered for folder")
	}
	if _, ok := state.Files[relativePath]; !ok {
		return nil
	}

	state.Version++
	delete(state.Files, relativePath)
	state.Tombstones[relativePath] = state.Version
	state.UpdatedAt = time.Now()
	return s.save()
}

func (s *Store) BlobReader(hash string) (io.ReadCloser, error) {
	if err := validateBlobHash(hash); err != nil {
		return nil, err
	}
	return os.Open(s.blobPath(hash))
}

func (s *Store) Status() models.ServerStatus {
	s.mu.Lock()
	defer s.mu.Unlock()

	status := models.ServerStatus{DataDir: s.root}
	entries, _ := os.ReadDir(filepath.Join(s.root, "blobs"))
	status.BlobCount = len(entries)

	for _, folder := range s.data.Folders {
		folderStatus := models.FolderStatus{
			ID:        folder.ID,
			Name:      folder.Name,
			Version:   folder.Version,
			UpdatedAt: folder.UpdatedAt,
		}
		for device := range folder.Devices {
			folderStatus.Devices = append(folderStatus.Devices, device)
		}
		sort.Strings(folderStatus.Devices)
		for _, file := range folder.Files {
			folderStatus.Files = append(folderStatus.Files, file)
		}
		sort.Slice(folderStatus.Files, func(i, j int) bool {
			return folderStatus.Files[i].RelativePath < folderStatus.Files[j].RelativePath
		})
		status.Folders = append(status.Folders, folderStatus)
	}

	sort.Slice(status.Folders, func(i, j int) bool {
		return status.Folders[i].ID < status.Folders[j].ID
	})

	return status
}

func conflictName(path, deviceID string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	return filepath.ToSlash(filepath.Join(filepath.Dir(path), fmt.Sprintf("%s.conflict-%s%s", name, deviceID, ext)))
}

func validateRelativePath(path string) error {
	if path == "" || filepath.IsAbs(path) {
		return errors.New("invalid relative path")
	}
	clean := filepath.Clean(filepath.FromSlash(path))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return errors.New("invalid relative path")
	}
	return nil
}

func validateBlobHash(value string) error {
	if len(value) != sha256.Size*2 {
		return errors.New("invalid blob hash")
	}
	if _, err := hex.DecodeString(value); err != nil {
		return errors.New("invalid blob hash")
	}
	return nil
}
