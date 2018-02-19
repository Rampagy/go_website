package main

import ("net/http"
        "html/template"
        "path"
        "io/ioutil"
        "strings"
        "fmt")

func main() {
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
    http.HandleFunc("/", index_handler)
    http.ListenAndServe(":8080", nil)
}

func index_handler(w http.ResponseWriter, r *http.Request) {
    // define template to be served
    t := template.New("Generic")

    pages, _ := ioutil.ReadDir("SubPages/")

    for _, page := range pages {
        p := path.Base(page.Name())

        reqPath := strings.ToUpper(path.Base(r.URL.Path))
        availPath := strings.ToUpper(p)

        //if the request is one of the pages
        if ((reqPath == availPath) ||
            (reqPath + ".HTML" == availPath)) {
                // use p in case filename isn't uppercase
                t, _ = template.ParseFiles("SubPages/" + p)
                break
        } else {
            t, _ = template.ParseFiles("SubPages/Home.html")
        }
    }

    fmt.Println(t.Execute(w, nil))
}
