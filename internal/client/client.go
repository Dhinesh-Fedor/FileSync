package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"FileSync/internal/models"
	"FileSync/internal/report"
	"FileSync/utils/hash"
)

type uploadTask struct {
	path       string
	file       models.FileInfo
	changeType models.ChangeType
}

var httpClient = &http.Client{Timeout: 2 * time.Minute}

func SyncFolder(folderID string) (models.SyncReport, error) {
	config, folder, files, err := prepareSyncData(folderID)
	if err != nil {
		return models.SyncReport{}, err
	}

	if err := saveCache(folder.ID, files); err != nil {
		return models.SyncReport{}, err
	}

	plan, err := requestPlan(config.ServerURL, models.SyncRequest{
		DeviceID:     config.DeviceID,
		Folder:       folder,
		KnownVersion: folder.CurrentVersion,
		Files:        files,
	})
	if err != nil {
		return models.SyncReport{}, err
	}

	syncReport := models.SyncReport{StartTime: time.Now()}
	if err := applyConflicts(folder.LocalPath, plan.Conflicts); err != nil {
		return models.SyncReport{}, err
	}

	fileMap := make(map[string]models.FileInfo, len(files))
	for _, file := range files {
		fileMap[file.RelativePath] = file
	}

	uploads := make([]uploadTask, 0, len(plan.UploadChanges))
	for _, change := range plan.UploadChanges {
		file, ok := fileMap[change.RelativePath]
		if !ok {
			continue
		}
		uploads = append(uploads, uploadTask{path: change.RelativePath, file: file, changeType: change.Type})
	}

	if err := uploadChangedFiles(config.ServerURL, folder.RemoteID, folder.DeviceID, uploads, &syncReport); err != nil {
		return models.SyncReport{}, err
	}

	if err := downloadFiles(config.ServerURL, folder.LocalPath, plan.DownloadFiles, &syncReport); err != nil {
		return models.SyncReport{}, err
	}

	for _, path := range plan.DeletePaths {
		if err := os.Remove(filepath.Join(folder.LocalPath, filepath.FromSlash(path))); err == nil || os.IsNotExist(err) {
			if err := commitDelete(config.ServerURL, folder.RemoteID, folder.DeviceID, path); err != nil {
				return models.SyncReport{}, err
			}
			syncReport.Deleted++
			syncReport.DeletedFiles = append(syncReport.DeletedFiles, path)
			continue
		}
		return models.SyncReport{}, err
	}

	syncReport.EndTime = time.Now()
	folder.CurrentVersion = plan.Version
	folder.LastSyncedAt = syncReport.EndTime
	if err := report.SaveName(folder.ID, syncReport); err != nil {
		return models.SyncReport{}, err
	}

	if err := updateFolder(config, folder); err != nil {
		return models.SyncReport{}, err
	}

	return syncReport, nil
}

func commitDelete(serverURL, folderID, deviceID, relativePath string) error {
	url := fmt.Sprintf("%s/delete?folder_id=%s&device_id=%s&relative_path=%s", strings.TrimRight(serverURL, "/"), urlEscape(folderID), urlEscape(deviceID), urlEscape(relativePath))
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	if token := authToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete failed for %s: %s", relativePath, resp.Status)
	}
	return nil
}

func PreviewSync(folderID string) (models.SyncPlan, error) {
	config, folder, files, err := prepareSyncData(folderID)
	if err != nil {
		return models.SyncPlan{}, err
	}

	return requestPlan(config.ServerURL, models.SyncRequest{
		DeviceID:     config.DeviceID,
		Folder:       folder,
		KnownVersion: folder.CurrentVersion,
		Files:        files,
	})
}

func findFolder(config models.ClientConfig, folderID string) (models.RegisteredFolder, error) {
	for _, folder := range config.Folders {
		if folder.ID == folderID || folder.RemoteID == folderID || filepath.Clean(folder.LocalPath) == filepath.Clean(folderID) {
			return folder, nil
		}
	}

	return models.RegisteredFolder{}, fmt.Errorf("folder not found: %s", folderID)
}

func updateFolder(config models.ClientConfig, folder models.RegisteredFolder) error {
	for i := range config.Folders {
		if config.Folders[i].ID == folder.ID || config.Folders[i].RemoteID == folder.RemoteID {
			config.Folders[i] = folder
			return SaveConfig(config)
		}
	}

	config.Folders = append(config.Folders, folder)
	return SaveConfig(config)
}

func requestPlan(serverURL string, req models.SyncRequest) (models.SyncPlan, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return models.SyncPlan{}, err
	}

	httpReq, err := http.NewRequest(http.MethodPost, strings.TrimRight(serverURL, "/")+"/sync", bytes.NewReader(body))
	if err != nil {
		return models.SyncPlan{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if token := authToken(); token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return models.SyncPlan{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return models.SyncPlan{}, fmt.Errorf("sync failed: %s", strings.TrimSpace(string(raw)))
	}

	var plan models.SyncPlan
	if err := json.NewDecoder(resp.Body).Decode(&plan); err != nil {
		return models.SyncPlan{}, err
	}

	return plan, nil
}

func uploadChangedFiles(serverURL, folderID, deviceID string, tasks []uploadTask, report *models.SyncReport) error {
	if len(tasks) == 0 {
		return nil
	}

	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}

	jobCh := make(chan uploadTask)
	errCh := make(chan error, len(tasks))
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range jobCh {
				file, err := os.Open(task.file.AbsolutePath)
				if err != nil {
					errCh <- err
					continue
				}

				computedHash := task.file.SHA256
				if computedHash == "" {
					computedHash, err = hash.HashFile(task.file.AbsolutePath)
					if err != nil {
						file.Close()
						errCh <- err
						continue
					}
				}

				uploadURL := fmt.Sprintf("%s/upload?folder_id=%s&device_id=%s&relative_path=%s&hash=%s&size=%d&modified_time=%s", strings.TrimRight(serverURL, "/"), urlEscape(folderID), urlEscape(deviceID), urlEscape(task.path), computedHash, task.file.Size, urlEscape(task.file.ModTime.Format(time.RFC3339Nano)))
				req, err := http.NewRequest(http.MethodPost, uploadURL, file)
				if err != nil {
					file.Close()
					errCh <- err
					continue
				}
				req.Header.Set("Content-Type", "application/octet-stream")
				if token := authToken(); token != "" {
					req.Header.Set("Authorization", "Bearer "+token)
				}

				resp, err := httpClient.Do(req)
				file.Close()
				if err != nil {
					errCh <- err
					continue
				}
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					errCh <- fmt.Errorf("upload failed for %s", task.path)
					continue
				}

				mu.Lock()
				if task.changeType == models.Modified {
					report.Modified++
					report.ModifiedFiles = append(report.ModifiedFiles, task.path)
				} else {
					report.Added++
					report.AddedFiles = append(report.AddedFiles, task.path)
				}
				report.BytesCopied += task.file.Size
				mu.Unlock()
			}
		}()
	}

	go func() {
		for _, task := range tasks {
			jobCh <- task
		}
		close(jobCh)
	}()

	wg.Wait()
	close(errCh)
	if err, ok := <-errCh; ok {
		return err
	}

	return nil
}

func downloadFiles(serverURL, root string, files []models.ServerFile, report *models.SyncReport) error {
	if len(files) == 0 {
		return nil
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].RelativePath < files[j].RelativePath
	})

	for _, file := range files {
		req, err := http.NewRequest(http.MethodGet, strings.TrimRight(serverURL, "/")+"/download/"+file.Hash, nil)
		if err != nil {
			return err
		}
		if token := authToken(); token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		resp, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return fmt.Errorf("download failed for %s", file.RelativePath)
		}

		localPath := filepath.Join(root, filepath.FromSlash(file.RelativePath))
		if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
			resp.Body.Close()
			return err
		}

		out, err := os.Create(localPath)
		if err != nil {
			resp.Body.Close()
			return err
		}

		bytesCopied, err := io.Copy(out, resp.Body)
		out.Close()
		resp.Body.Close()
		if err != nil {
			return err
		}

		report.Modified++
		report.BytesCopied += bytesCopied
		report.ModifiedFiles = append(report.ModifiedFiles, file.RelativePath)
	}

	return nil
}

func applyConflicts(root string, conflicts []models.ConflictItem) error {
	for _, conflict := range conflicts {
		src := filepath.Join(root, filepath.FromSlash(conflict.RelativePath))
		dst := filepath.Join(root, filepath.FromSlash(conflict.ConflictPath))
		if _, err := os.Stat(src); err != nil {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return err
		}
		if err := os.Rename(src, dst); err != nil {
			return err
		}
	}

	return nil
}

func urlEscape(value string) string {
	return strings.NewReplacer("%", "%25", " ", "%20", "+", "%2B", "#", "%23", "?", "%3F", "&", "%26", "/", "%2F").Replace(value)
}
