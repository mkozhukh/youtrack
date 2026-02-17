package filestore

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ServeFile returns an http.HandlerFunc that serves stored files by ID.
// Expected path: GET /mcpfiles/{id}
func ServeFile(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := strings.TrimPrefix(r.URL.Path, "/mcpfiles/")
		if id == "" {
			http.Error(w, "missing file id", http.StatusBadRequest)
			return
		}

		filePath, filename, contentType, err := store.Get(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		http.ServeFile(w, r, filePath)
	}
}

// UploadFile returns an http.HandlerFunc that accepts multipart uploads.
// Expected path: POST /mcpfiles
func UploadFile(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Limit request body to store max size + overhead
		r.Body = http.MaxBytesReader(w, r.Body, store.maxSize+1024*1024)

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to read file: %v", err), http.StatusBadRequest)
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to read file data: %v", err), http.StatusBadRequest)
			return
		}

		fileID, err := store.Put(data, header.Filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"file_id": fileID,
		})
	}
}
