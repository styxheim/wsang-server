package main

import (
  "fmt"
  "log"
  "encoding/json"
  "net/http"
  "github.com/gorilla/mux"
)

var AdminTerminalString = "ad914";

func adminResultHandler(w http.ResponseWriter) {

  if r := recover(); r != nil {
    var ares = &AdminResponse { Error: &Error{ Text: fmt.Sprintf("%s", r) } }

    log.Println("!!!", "got error", r)
    json, _ := json.MarshalIndent(ares, "", "  ")
    w.Write(json)
  }
}


func AdminListHandler(w http.ResponseWriter, r *http.Request) {
  var ares AdminResponse
  v := mux.Vars(r)
  log.Println("ADMIN", r.URL)

  defer adminResultHandler(w)

  if v["TerminalString"] != AdminTerminalString {
    panic("Not an admin terminal")
  }

  ares.Competitions = GetCompetitions()
  json, _ := json.MarshalIndent(ares, "", "  ")
  w.Write(json)
}

func AdminActivateHandler(w http.ResponseWriter, r *http.Request) {
}

func AdminHandler(w http.ResponseWriter, r *http.Request) {
}
