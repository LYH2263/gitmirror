package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/LYH2263/go-gitmirror"
	"github.com/LYH2263/go-gitmirror/internal/admin"
	"github.com/LYH2263/go-gitmirror/internal/fetch"
	"github.com/LYH2263/go-gitmirror/internal/remoteconfig"
)

func main() {
	addr := flag.String("addr", ":8114", "listen address")
	root := flag.String("root", "./data", "mirror root")
	web := flag.String("web", "web", "static web dir")
	remoteURL := flag.String("remote", "memory://demo", "default remote url")
	flag.Parse()

	if err := os.MkdirAll(*root, 0o755); err != nil {
		log.Fatal(err)
	}
	tr := fetch.NewMemory()
	tr.Seed(*remoteURL, map[string]string{
		"refs/heads/main": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}, "")

	m, err := gitmirror.Open(context.Background(), gitmirror.Options{
		Root:      *root,
		Remote:    gitmirror.RemoteSpec{Name: "origin", URL: *remoteURL},
		Transport: tr,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer m.Close()
	_ = remoteconfig.Save(*root, &remoteconfig.File{Name: "origin", URL: *remoteURL, Fetch: []string{"refs/heads/*"}})

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(*web)))
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"ok": true, "time": time.Now().UTC()})
	})
	mux.HandleFunc("/api/dashboard", func(w http.ResponseWriter, r *http.Request) {
		d, err := admin.FromMirror(m)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, d)
	})
	mux.HandleFunc("/api/refs", func(w http.ResponseWriter, r *http.Request) {
		refs, err := m.Refs()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, refs)
	})
	mux.HandleFunc("/api/remote", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, m.Remote())
		case http.MethodPost:
			var spec gitmirror.RemoteSpec
			if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			if err := m.SetRemote(spec); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			_ = remoteconfig.Save(m.Root(), &remoteconfig.File{Name: spec.Name, URL: spec.URL, Fetch: spec.FetchRefs})
			writeJSON(w, map[string]string{"status": "ok"})
		default:
			http.Error(w, "method", 405)
		}
	})
	mux.HandleFunc("/api/sync", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", 405)
			return
		}
		rep, err := m.SyncContext(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, rep)
	})
	mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, m.Stats())
	})

	fmt.Println("gitd listening on", *addr, "root", *root)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
