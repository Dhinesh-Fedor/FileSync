package handlers

import (
	"net/http"

	"FileSync/internal/utils"
)

func Ping(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
