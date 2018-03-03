package main

import ("net/http"
        "html/template"
        "path"
        "io/ioutil"
        "strings"
        "errors"
        "time"
        "fmt")

type PageInfo struct {
    Subpages []NavBarMap            // all available subpages
    HomeInfo NavBarMap              // homepage info
    ActiveInfo NavBarMap            // info on which page is active
}

type NavBarMap struct {
    Link string
    Name string
}

func main() {
    doEvery(time.Second, serveDynamic)

    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
    http.HandleFunc("/", index_handler)
    http.ListenAndServe(":8080", nil)
}

func serveDynamic() {
    http.Handle("/dynamic/", http.StripPrefix("/dynamic/", http.FileServer(http.Dir("dynamic/"))))
}

func doEvery(d time.Duration, f func(time.Time)) {
    for x := range time.Tick(d) {
        f(x)
    }
}

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
