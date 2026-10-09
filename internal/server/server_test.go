package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandlerRequiresConfiguredToken(t *testing.T) {
	root := t.TempDir()
	oldwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(root))
	t.Cleanup(func() { _ = os.Chdir(oldwd) })
	t.Setenv("FILESYNC_AUTH_TOKEN", "test-token")

	app, err := NewApp()
	require.NoError(t, err)
	handler := app.Handler()

	request := httptest.NewRequest(http.MethodGet, "/status", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	require.Equal(t, http.StatusUnauthorized, response.Code)

	request = httptest.NewRequest(http.MethodGet, "/status", nil)
	request.Header.Set("Authorization", "Bearer test-token")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
}

func TestHandlerRejectsTraversalDownload(t *testing.T) {
	root := t.TempDir()
	oldwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(root))
	t.Cleanup(func() { _ = os.Chdir(oldwd) })
	t.Setenv("FILESYNC_AUTH_TOKEN", "")

	app, err := NewApp()
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodGet, "/download/../metadata/index.json", nil)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)
	require.NotEqual(t, http.StatusOK, response.Code)
}
