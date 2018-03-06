package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	var (
		freqPath = flag.String("freq", "", "filename containing the frequency")
		tempPath = flag.String("temp", "", "filename containing the temperature")
		tmplPath = flag.String("tmpl", "", "directory containing the templates and static assets")
	)
	flag.Parse()

	abortIfNotExist(*freqPath, "frequency file", "use -freq=<frequency file>")
	abortIfNotExist(*tempPath, "temperature file", "use -temp=<temperature file>")
	abortIfNotExist(*tmplPath, "templates directory", "use -tmpl=<path to templates>")

	s := newServer(*tmplPath, *freqPath, *tempPath)

	log.Fatal(http.ListenAndServe(":8080", s))

}

func abortIfNotExist(path, description, helpfulMessage string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Fatalf("cannot open %s %q: %s", description, path, helpfulMessage)
	}
}

type server struct {
	mux        *http.ServeMux
	wsUpgrader websocket.Upgrader
	freqPath   string
	tempPath   string
	homeTmpl   *template.Template
	systemTmpl *template.Template
}

func newServer(tmplPath, freqPath, tempPath string) *server {
	s := &server{
		mux:        http.NewServeMux(),
		wsUpgrader: websocket.Upgrader{},
		freqPath:   freqPath,
		tempPath:   tempPath,
	}
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(tmplPath, "static")))))
	s.mux.HandleFunc("/", s.index)
	s.mux.HandleFunc("/system", s.system)
	s.mux.HandleFunc("/ws", s.getSystem)
	baseTmpl := template.Must(template.ParseFiles(filepath.Join(tmplPath, "templates/base.tmpl")))
	s.homeTmpl = template.Must(template.Must(baseTmpl.Clone()).ParseFiles(filepath.Join(tmplPath, "templates/home.tmpl")))
	s.systemTmpl = template.Must(template.Must(baseTmpl.Clone()).ParseFiles(filepath.Join(tmplPath, "templates/system.tmpl")))
	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

type systemData struct {
	CPUFreq string `json:"cpuFreq"`
	CPUTemp string `json:"cpuTemp"`
}

func (s *server) getSystem(w http.ResponseWriter, r *http.Request) {
	c, err := s.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrading HTTP: %v", err)
		return
	}
	defer c.Close()
	for q := time.Tick(1 * time.Second); ; <-q {
		temp, err := readFileAsFloat(s.tempPath)
		if err != nil {
			log.Printf("reading temperature: %v", err)
			return
		}

		freq, err := readFileAsFloat(s.freqPath)
		if err != nil {
			log.Printf("reading frequency: %v", err)
			return
		}

		t := strconv.FormatFloat(temp/1000, 'f', 1, 64)
		f := strconv.FormatFloat(freq/1000, 'f', 0, 64)

		msg := systemData{CPUFreq: f, CPUTemp: t}

		if err = c.WriteJSON(msg); err != nil {
			return
		}
	}
}

func readFileAsFloat(filename string) (float64, error) {
	b, err := ioutil.ReadFile(filename) // just pass the file name
	if err != nil {
		return 0.0, fmt.Errorf("reading file: %v", err)
	}

	temp, err := strconv.ParseFloat(string(bytes.TrimSpace(b)), 64)
	if err != nil {
		return 0.0, fmt.Errorf("parsing value: %v", err)
	}

	return temp, nil
}

type pageInfo struct {
	Active string // Which nav link is active.
}

func (s *server) index(w http.ResponseWriter, r *http.Request) {
	info := pageInfo{Active: "home"}
	render(w, s.homeTmpl, info)
}

func (s *server) system(w http.ResponseWriter, r *http.Request) {
	info := pageInfo{Active: "system"}
	render(w, s.systemTmpl, info)
}

func render(w http.ResponseWriter, t *template.Template, data interface{}) {
	if err := t.Execute(w, data); err != nil {
		http.Error(w, fmt.Sprintf("executing template %q with data %v: %v", t.Name(), data, err), http.StatusInternalServerError)
		return
	}
}
