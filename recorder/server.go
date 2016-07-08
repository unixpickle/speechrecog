package main

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"html/template"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/unixpickle/speechrecog/speechdata"
)

const (
	AssetPrefix = "assets/"
	IndexPath   = AssetPrefix + "index.html"
	StylePath   = AssetPrefix + "style.css"
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
	case "/script.js", "/jswav.js":
		contents, _ := Asset(path.Join(AssetPrefix, r.URL.Path))
		w.Header().Set("Content-Type", "application/javascript")
		w.Write(contents)
	case "/add":
		msg, err := s.addLabel(r)
		writeAPIResponse(w, msg, err)
	case "/delete":
		writeAPIResponse(w, "success", s.deleteEntry(r))
	case "/upload":
		writeAPIResponse(w, "success", s.upload(r))
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

func (s *Server) upload(r *http.Request) error {
	query := r.URL.Query()
	id := query.Get("id")

	if id == "" {
		return errors.New("missing id field")
	}

	destName := randomID()
	destPath := filepath.Join(s.Index.DirPath, destName)

	f, err := os.Create(destPath)
	defer f.Close()
	if err != nil {
		return err
	}
	b64 := base64.NewDecoder(base64.StdEncoding, r.Body)
	if _, err := io.Copy(f, b64); err != nil {
		os.Remove(destPath)
		return err
	}

	s.DataLock.Lock()
	defer s.DataLock.Unlock()

	for i, x := range s.Index.Samples {
		if x.ID == id {
			oldFile := s.Index.Samples[i].File
			if oldFile != "" {
				os.Remove(destPath)
				return errors.New("entry already has a recording")
			}
			s.Index.Samples[i].File = destName
			if err := s.Index.Save(); err != nil {
				s.Index.Samples[i].File = ""
				os.Remove(destPath)
			}
			return nil
		}
	}

	os.Remove(destPath)
	return errors.New("unknown id: " + id)
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
