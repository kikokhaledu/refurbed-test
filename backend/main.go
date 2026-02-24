package main

import (
	"errors"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

func main() {
	source := FileProductSource{
		MetadataPath: filepath.Join("data", "metadata.json"),
		DetailsPath:  filepath.Join("data", "details.json"),
	}
	service := NewProductService(source, 30*time.Second)

	mux := http.NewServeMux()
	mux.Handle("/products", NewProductHandler(service))

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	server := &http.Server{
		Addr:              ":8080",
		Handler:           withCORS(withLogging(mux)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("Server starting on http://localhost:8080")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
