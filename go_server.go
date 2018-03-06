package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var (
	baseTmpl   = template.Must(template.ParseFiles("templates/base.tmpl"))
	homeTmpl   = template.Must(template.Must(baseTmpl.Clone()).ParseFiles("templates/home.tmpl"))
	systemTmpl = template.Must(template.Must(baseTmpl.Clone()).ParseFiles("templates/system.tmpl"))
)

type pageInfo struct {
	Active string // Which nav link is active.
}

type systemData struct {
	CPUFreq string `json:"cpuFreq"`
	CPUTemp string `json:"cpuTemp"`
}

var upgrader = websocket.Upgrader{}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	http.HandleFunc("/", index)
	http.HandleFunc("/system", system)
	http.HandleFunc("/ws", getSystem)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getSystem(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrading HTTP: %v", err)
		return
	}
	defer c.Close()
	for q := time.Tick(1 * time.Second); ; <-q {
		temp, err := readFileAsFloat("/sys/class/thermal/thermal_zone0/temp")
		if err != nil {
			log.Printf("reading temperature: %v", err)
			return
		}

		freq, err := readFileAsFloat("/sys/devices/system/cpu/cpu0/cpufreq/scaling_cur_freq")
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

func index(w http.ResponseWriter, r *http.Request) {
	info := pageInfo{Active: "home"}
	render(w, homeTmpl, info)
}

func system(w http.ResponseWriter, r *http.Request) {
	info := pageInfo{Active: "system"}
	render(w, systemTmpl, info)
}

func render(w http.ResponseWriter, t *template.Template, data interface{}) {
	if err := t.Execute(w, data); err != nil {
		http.Error(w, fmt.Sprintf("executing template %q with data %v: %v", t.Name(), data, err), http.StatusInternalServerError)
		return
	}
}
