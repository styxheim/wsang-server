package main

import (
  "fmt"
  "log"
  "encoding/json"
  "net/http"
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

func adminCheckCredentials(creds Credentials) {
  if creds.TerminalString == "" && creds.SecureKey == "0000" {
    return
  }
  panic("Invalid credentials")
}

func AdminListHandler(w http.ResponseWriter, r *http.Request) {
  var ares AdminResponseCompetitionList
  var areq AdminRequestGet
  log.Println("ADMIN GET", r.URL)

  defer adminResultHandler(w)

  err := json.NewDecoder(r.Body).Decode(&areq)
  if err != nil {
    panic(err);
  }

  adminCheckCredentials(areq.Credentials);

  ares.Competitions = GetCompetitions()
  json, _ := json.MarshalIndent(ares, "", "  ")
  w.Write(json)
}

func AdminSyncPointHandler(w http.ResponseWriter, r *http.Request) {
  defer adminResultHandler(w)

  panic("SyncPoint not implemented: should set SyncPoint to race")
}

func AdminGetCompetitionHandler(w http.ResponseWriter, r *http.Request) {
  defer adminResultHandler(w)

  panic("AdminGetCompetitionHandler not implemented: should show competition configuration")
}

func AdminSetCompetitionHandler(w http.ResponseWriter, r *http.Request) {
  defer adminResultHandler(w)

  panic("AdminSetCompetitionHandler not implemented: should set competition configuration")
}

func AdminTerminalListHandler(w http.ResponseWriter, r *http.Request) {
  defer adminResultHandler(w)

  panic("AdminTerminalListHandler not implemented: should list all known terminals")
}