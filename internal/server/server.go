package server

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"FileSync/internal/models"
)

type App struct {
	store     *Store
	authToken string
}

func NewApp() (*App, error) {
	root := strings.TrimSpace(os.Getenv("FILESYNC_SERVER_STORAGE"))
	store, err := NewStore(root)
	if err != nil {
		return nil, err
	}

	return &App{store: store, authToken: strings.TrimSpace(os.Getenv("FILESYNC_AUTH_TOKEN"))}, nil
}

func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/register-folder", a.handleRegisterFolder)
	mux.HandleFunc("/sync", a.handleSync)
	mux.HandleFunc("/upload", a.handleUpload)
	mux.HandleFunc("/delete", a.handleDelete)
	mux.HandleFunc("/download/", a.handleDownload)
	mux.HandleFunc("/folders", a.handleFolders)
	mux.HandleFunc("/status", a.handleStatus)
	return a.authMiddleware(mux)
}

func (a *App) Listen(addr string) error {
	if addr == "" {
		addr = strings.TrimSpace(os.Getenv("FILESYNC_SERVER_ADDR"))
	}
	if addr == "" {
		addr = ":8080"
	}

	fmt.Println("FileSync server listening on", addr)
	server := &http.Server{
		Addr:              addr,
		Handler:           a.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       2 * time.Minute,
		WriteTimeout:      2 * time.Minute,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	return server.ListenAndServe()
}

func (a *App) handleRegisterFolder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var folder models.RegisteredFolder
	if err := json.NewDecoder(r.Body).Decode(&folder); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	registered, err := a.store.RegisterFolder(folder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, registered)
}

func (a *App) handleSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 64<<20)
	var req models.SyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	plan, err := a.store.Sync(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, plan)
}

func (a *App) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	hash := r.URL.Query().Get("hash")
	if hash == "" {
		http.Error(w, "missing hash", http.StatusBadRequest)
		return
	}
	folderID := r.URL.Query().Get("folder_id")
	deviceID := r.URL.Query().Get("device_id")
	relativePath := r.URL.Query().Get("relative_path")
	size, err := strconv.ParseInt(r.URL.Query().Get("size"), 10, 64)
	if folderID == "" || deviceID == "" || relativePath == "" || err != nil || size < 0 {
		http.Error(w, "invalid upload metadata", http.StatusBadRequest)
		return
	}
	modifiedTime, err := time.Parse(time.RFC3339Nano, r.URL.Query().Get("modified_time"))
	if err != nil {
		http.Error(w, "invalid modified time", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<30)
	if _, err := a.store.StoreBlob(hash, r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	file, err := a.store.CommitUpload(folderID, deviceID, relativePath, hash, size, modifiedTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	writeJSON(w, map[string]any{"status": "ok", "file": file})
}

func (a *App) handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	hash := strings.TrimPrefix(r.URL.Path, "/download/")
	if hash == "" {
		http.NotFound(w, r)
		return
	}
	if err := validateBlobHash(hash); err != nil {
		http.NotFound(w, r)
		return
	}

	body, err := a.store.BlobReader(hash)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer body.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = io.Copy(w, body)
}

func (a *App) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	folderID := r.URL.Query().Get("folder_id")
	deviceID := r.URL.Query().Get("device_id")
	relativePath := r.URL.Query().Get("relative_path")
	if folderID == "" || deviceID == "" || relativePath == "" {
		http.Error(w, "invalid delete metadata", http.StatusBadRequest)
		return
	}
	if err := a.store.CommitDelete(folderID, deviceID, relativePath); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

func (a *App) handleFolders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, a.store.Status().Folders)
}

func (a *App) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, a.store.Status())
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func (a *App) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.authToken == "" {
			next.ServeHTTP(w, r)
			return
		}
		provided := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		if provided == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(a.authToken)) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
