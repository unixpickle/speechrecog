package main

import (
	"encoding/hex"
	"errors"
	"html/template"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/unixpickle/speechrecog/speechdata"
)

const (
	IndexPath  = "assets/index.html"
	StylePath  = "assets/style.css"
	ScriptPath = "assets/script.js"
)

var (
	IndexTemplate *template.Template
)

func init() {
	rand.Seed(time.Now().UnixNano())
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
	case "/style.css":
		contents, _ := Asset(StylePath)
		w.Header().Set("Content-Type", "text/css")
		w.Write(contents)
	case "/script.js":
		contents, _ := Asset(ScriptPath)
		w.Header().Set("Content-Type", "application/javascript")
		w.Write(contents)
	case "/add":
		msg, err := s.addLabel(r)
		writeAPIResponse(w, msg, err)
	case "/delete":
		writeAPIResponse(w, "success", s.deleteEntry(r))
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

func (s *Server) addLabel(r *http.Request) (id string, err error) {
	query := r.URL.Query()
	label := query.Get("label")

	if label == "" {
		return "", errors.New("cannot add empty label")
	}

	s.DataLock.Lock()
	defer s.DataLock.Unlock()

	id = randomID()
	s.Index.Samples = append(s.Index.Samples, speechdata.Sample{
		ID:    id,
		Label: label,
	})
	if err := s.Index.Save(); err != nil {
		s.Index.Samples = s.Index.Samples[:len(s.Index.Samples)-1]
		return "", err
	}
	return id, nil
}

func (s *Server) deleteEntry(r *http.Request) error {
	query := r.URL.Query()
	id := query.Get("id")

	s.DataLock.Lock()
	defer s.DataLock.Unlock()

	for i, x := range s.Index.Samples {
		if x.ID == id {
			backup := s.Index.Clone()
			copy(s.Index.Samples[i:], s.Index.Samples[i+1:])
			s.Index.Samples = s.Index.Samples[:len(s.Index.Samples)-1]
			if err := s.Index.Save(); err != nil {
				s.Index = backup
				return err
			}
			if x.File != "" {
				os.Remove(filepath.Join(s.Index.DirPath, x.File))
			}
			return nil
		}
	}

	return errors.New("ID not found: " + id)
}

func writeAPIResponse(w http.ResponseWriter, successMsg string, err error) {
	w.Header().Set("Content-Type", "text/plain")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error: " + err.Error()))
	} else {
		w.Write([]byte(successMsg))
	}
}

func randomID() string {
	var buf [16]byte
	for i := 0; i < len(buf); i++ {
		buf[i] = byte(rand.Intn(0x100))
	}
	return strings.ToLower(hex.EncodeToString(buf[:]))
}
