package main

import ("net/http"
        "html/template"
        "path"
        "io/ioutil"
        "strings"
        "errors"
        "fmt")

type PageInfo struct {
    Subpages []NavBarMap
}

type NavBarMap struct {
    Link string
    Name string
}

func main() {
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
    http.HandleFunc("/", index_handler)
    http.ListenAndServe(":8080", nil)
}

func index_handler(w http.ResponseWriter, r *http.Request) {
    // define generic template
    t := template.New("Generic")

    // define struct for templates
    var PageData PageInfo

    // add subpages to the template first
    pages, _ := ioutil.ReadDir("SubPages/")
    reqPath := strings.ToUpper(path.Base(r.URL.Path))

    for _, page := range pages {
        p := path.Base(page.Name())
        availPath := strings.ToUpper(p)

        PageData.Subpages = append(PageData.Subpages,
                    NavBarMap{Link: p, Name: p})

        //if the request is one of the pages
        if ((reqPath == availPath) ||
            (reqPath + ".HTML" == availPath)) {
                // use original filename rather than uppercase name
                t, _ = template.ParseFiles("SubPages/" + p)
        } else if (t.Name() == "Generic") {
                t, _ = template.ParseFiles("SubPages/Home.html")
        }
    }

    // tack reusables onto the end of the template
    reusables, _ := ioutil.ReadDir("Reuse")
    for _, reusable := range reusables {
        err := errors.New("emit macho dwarf: elf header corrupted")
        t, err = t.ParseFiles("Reuse/" + path.Base(reusable.Name()))
        fmt.Println(err)
    }

    fmt.Println(t.Execute(w, PageData))
}
