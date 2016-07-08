package main

import (
	"errors"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/unixpickle/speechrecog/speechdata"
)

const (
	IndexPath = "assets/index.html"
)

var (
	IndexTemplate *template.Template
)

func init() {
	templateData, err := Asset(IndexPath)
	if err != nil {
		panic(err)
	}
	IndexTemplate = template.Must(template.New("index").Parse(string(templateData)))
}

type Server struct {
	DataLock sync.RWMutex
	Index    *speechdata.Index
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		w.Header().Set("Content-Type", "text/html")
		s.DataLock.RLock()
		idx := s.Index.Clone()
		s.DataLock.RUnlock()
		IndexTemplate.Execute(w, idx)
	case "/recording.wav":
		soundFile, err := s.openSoundFile(r)
		if err != nil {
			http.NotFound(w, r)
		} else {
			defer soundFile.Close()
			w.Header().Set("Content-Type", "audio/x-wav")
			io.Copy(w, soundFile)
		}
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) openSoundFile(r *http.Request) (io.ReadCloser, error) {
	query := r.URL.Query()
	id := query.Get("id")

	s.DataLock.RLock()
	defer s.DataLock.RUnlock()

	for _, x := range s.Index.Samples {
		if x.ID == id {
			if x.File == "" {
				return nil, errors.New("no recording for sample")
			}
			return os.Open(filepath.Join(s.Index.DirPath, x.File))
		}
	}

	return nil, errors.New("invalid ID: " + id)
}
