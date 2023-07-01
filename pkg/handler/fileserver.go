package handler

import (
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

// found the solution to hide the directory index for http.FileServer
// https://stackoverflow.com/questions/51169677/http-static-file-handler-keeps-showing-the-directory-list
var notFoundFile, notFoundErr = http.Dir("dummy").Open("does-not-exist")

// noDirIndex implements the interface https://pkg.go.dev/net/http#FileSystem
// and prevents the displaying of a whole directory index
type noDirIndex struct {
	http.Dir
}

func (m noDirIndex) Open(name string) (result http.File, err error) {
	var (
		f  http.File
		fi fs.FileInfo
	)

	if f, err = m.Dir.Open(name); err != nil {
		return
	}

	// find out the type to the given file-handle
	if fi, err = f.Stat(); err != nil {
		return
	}
	if fi.IsDir() {
		// Return a response that would have been if directory would not exist:
		return notFoundFile, notFoundErr
	}
	return f, nil
}

// ServeStaticDir acts as a FileServer and serves the contents of the given dir
func ServeStaticDir(r chi.Router, pathPrefix string, static http.Dir) {
	if strings.ContainsAny(pathPrefix, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	root, _ := filepath.Abs(string(static))
	if _, err := os.Stat(root); os.IsNotExist(err) {
		panic("Static Documents Directory Not Found")
	}

	fs := http.StripPrefix(pathPrefix, http.FileServer(noDirIndex{http.Dir(root)}))

	r.Get(pathPrefix+"*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file := strings.Replace(r.RequestURI, pathPrefix, "", 1)
		// if the file contains URL params, remove everything after ?
		if strings.Contains(file, "?") {
			parts := strings.Split(file, "?")
			if len(parts) == 2 {
				file = parts[0] // use everything before the ?
			}
		}
		if _, err := os.Stat(root + file); os.IsNotExist(err) {
			http.ServeFile(w, r, path.Join(root, "index.html"))
			return
		}
		fs.ServeHTTP(w, r)
	}))
}

// ServeStaticFile acts as a FileServer for a single file and serves the specified file
func ServeStaticFile(r chi.Router, urlPath, filepath string) {
	if urlPath == "" {
		panic("no path for fileServer defined!")
	}
	if strings.ContainsAny(urlPath, "{}*") {
		panic("fileServer does not permit URL parameters.")
	}

	fileHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath)
	})

	r.Get(urlPath, fileHandler)
	r.Options(urlPath, fileHandler)
}
