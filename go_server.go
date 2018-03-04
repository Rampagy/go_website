package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type pageInfo struct {
	Subpages   []navBarMap // all available subpages
	HomeInfo   navBarMap   // homepage info
	ActiveInfo navBarMap   // info on which page is active
}

type navBarMap struct {
	Link string
	Name string
}

type systemData struct {
	CPUFreq string `json:"cpuFreq"`
	CPUTemp string `json:"cpuTemp"`
}

var upgrader = websocket.Upgrader{}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	http.HandleFunc("/", index)
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
	// define generic template
	t := template.New("Generic")

	// define struct for templates
	var PageData pageInfo
	PageData.HomeInfo = navBarMap{Link: "Home.html", Name: "Home"}

	// add subpages to the template first
	pages, _ := ioutil.ReadDir("SubPages/")
	reqPath := strings.ToUpper(path.Base(r.URL.Path))

	for _, page := range pages {
		p := path.Base(page.Name())
		availPath := strings.ToUpper(p)

		if availPath != "HOME.HTML" {
			PageData.Subpages = append(PageData.Subpages,
				navBarMap{Link: p, Name: p[:len(p)-len(path.Ext(p))]})
		}

		//if the request is one of the pages
		if (reqPath == availPath) ||
			(reqPath+".HTML" == availPath) {
			// use original filename rather than uppercase name
			t, _ = template.ParseFiles("SubPages/" + p)
			PageData.ActiveInfo = navBarMap{Link: p,
				Name: p[:len(p)-len(path.Ext(p))]}
		} else if t.Name() == "Generic" {
			t, _ = template.ParseFiles("SubPages/Home.html")
			PageData.ActiveInfo = navBarMap{Link: "Home.html", Name: "Home"}
		}
	}

	// tack reusables onto the end of the template
	reusables, _ := ioutil.ReadDir("Reuse")
	for _, reusable := range reusables {
		var err error
		t, err = t.ParseFiles("Reuse/" + path.Base(reusable.Name()))
		if err != nil {
			http.Error(w, fmt.Sprintf("parsing reuse files: %v", err), http.StatusInternalServerError)
			return
		}
	}

	if err := t.Execute(w, PageData); err != nil {
		http.Error(w, fmt.Sprintf("executing index template: %v", err), http.StatusInternalServerError)
		return
	}
}
