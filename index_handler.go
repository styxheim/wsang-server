package main

import (
  "fmt"
  "log"
  "path"
  "io/ioutil"
  "net/http"
  "html/template"
)

func check_err(w http.ResponseWriter) {
  if r := recover(); r != nil {
    log.Println("!!!", "http error", r)
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(fmt.Sprintf("%s", r)))
  }
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
  defer check_err(w)

  tmpl := template.Must(template.ParseFiles("res/index.html"))
  err := tmpl.Execute(w, nil)

  if err != nil {
    panic(err)
  }
}

func anyResHander(w http.ResponseWriter, r *http.Request,
                  name string, mime string) {
  defer check_err(w)

  var fpath = fmt.Sprintf("res/%s/%s", name, path.Base(r.URL.String()));

  data, err := ioutil.ReadFile(fpath)
  if err != nil {
    panic(err)
  }
  w.Header().Set("Content-Type", mime) 
  w.Write(data)
}

func ScriptHander(w http.ResponseWriter, r *http.Request) {
  anyResHander(w, r, "js", "application/javascript")
}

func CssHander(w http.ResponseWriter, r *http.Request) {
  anyResHander(w, r, "css", "text/css")
}
