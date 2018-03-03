package main

import (
        "net/http"
        "html/template"
        "path"
        "io/ioutil"
        "strings"
        "errors"
        "fmt"
        //"time"
        //"strconv"
        //"encoding/json"

        //"github.com/go-chi/chi"
        //"gopkg.in/olahol/melody.v1"
)

type PageInfo struct {
    Subpages []NavBarMap            // all available subpages
    HomeInfo NavBarMap              // homepage info
    ActiveInfo NavBarMap            // info on which page is active
}

type NavBarMap struct {
    Link string
    Name string
}

type SystemData struct {
    CpuFreq string `json:"cpuFreq"`
    CpuTemp string `json:"cpuTemp"`
}

func main() {
    //r := chi.NewRouter()
    //m := melody.New()

    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
    http.HandleFunc("/", index_handler)
/*
    r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
        m.HandleRequest(w, r)
    })*/

    //go getSystem(m)

    http.ListenAndServe(":8080", nil)//r)
}
/*
func getSystem(m *melody.Melody) {
    for q := time.Tick(1 * time.Second); ; <-q {
        temperature, err := ioutil.ReadFile("dynamic/temperature.txt") // just pass the file name
        if err != nil {
            fmt.Print(err)
        }

        frequency, err := ioutil.ReadFile("dynamic/frequency.txt") // just pass the file name
        if err != nil {
            fmt.Print(err)
        }

        temp, err := strconv.ParseFloat(string(temperature[:len(temperature)-1]), 64)
        t := strconv.FormatFloat(temp/1000, 'f', 1, 64)
        freq, err := strconv.ParseFloat(string(frequency[:len(frequency)-1]), 64)
        f := strconv.FormatFloat(freq/1000, 'f', 0, 64)

        msg := SystemData{CpuFreq: f, CpuTemp: t}
        b, err := json.Marshal(msg)

        var l SystemData
        json.Unmarshal(b, &l)

        if err != nil {
            s := `{"freq":0.0, "temp":0.0}`
            m.Broadcast([]byte(s))
        } else {
            m.Broadcast(b)
        }
    }
}
*/
func index_handler(w http.ResponseWriter, r *http.Request) {
    // define generic template
    t := template.New("Generic")

    // define struct for templates
    var PageData PageInfo
    PageData.HomeInfo = NavBarMap{Link: "Home.html", Name: "Home"}

    // add subpages to the template first
    pages, _ := ioutil.ReadDir("SubPages/")
    reqPath := strings.ToUpper(path.Base(r.URL.Path))

    for _, page := range pages {
        p := path.Base(page.Name())
        availPath := strings.ToUpper(p)

        if (availPath != "HOME.HTML") {
            PageData.Subpages = append(PageData.Subpages,
                NavBarMap{Link: p, Name: p[:len(p) - len(path.Ext(p))]})
        }

        //if the request is one of the pages
        if ((reqPath == availPath) ||
            (reqPath + ".HTML" == availPath)) {
                // use original filename rather than uppercase name
                t, _ = template.ParseFiles("SubPages/" + p)
                PageData.ActiveInfo = NavBarMap{Link: p,
                            Name: p[:len(p) - len(path.Ext(p))]}
        } else if (t.Name() == "Generic") {
                t, _ = template.ParseFiles("SubPages/Home.html")
                PageData.ActiveInfo = NavBarMap{Link: "Home.html", Name: "Home"}
        }
    }

    // tack reusables onto the end of the template
    reusables, _ := ioutil.ReadDir("Reuse")
    for _, reusable := range reusables {
        err := errors.New("Original Error")
        t, err = t.ParseFiles("Reuse/" + path.Base(reusable.Name()))
        fmt.Println(err)
    }

    fmt.Println(t.Execute(w, PageData))
}
