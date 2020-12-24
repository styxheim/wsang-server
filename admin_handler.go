package main

import (
  "io"
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

func adminCheckCredentials(creds Credentials) {
  if creds.TerminalString == "" && creds.SecureKey == "XXXX" {
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

func bodyDecode(body io.ReadCloser, v interface{}) {
  err := json.NewDecoder(body).Decode(v)
  if err != nil {
    panic(err)
  }
}

func AdminGetCompetitionHandler(w http.ResponseWriter, r *http.Request) {
  var areq AdminRequestGet
  var resp AdminResponseCompetitionGet
  var v = mux.Vars(r)
  var id uint64

  defer adminResultHandler(w)

  id = extractUint64(v, "CompetitionId")
  log.Println("Admin::Competition::Get(", id, ")")
  bodyDecode(r.Body, &areq)
  adminCheckCredentials(areq.Credentials)

  var DataResponse = GetCompetition(id, nil, 0)
  resp.Competition = *DataResponse.RaceStatus
  resp.TerminalList = GetTerminals(id, nil, 0)

  json, _ := json.MarshalIndent(resp, "", "  ")
  w.Write(json)
}

func AdminSetCompetitionHandler(w http.ResponseWriter, r *http.Request) {
  var areq AdminRequestCompetitionSet
  var resp AdminResponse
  var v = mux.Vars(r)
  var id uint64
  var storedCompetition *RaceStatus

  defer adminResultHandler(w)

  id = extractUint64(v, "CompetitionId")
  log.Println("Admin::Competition:Set(", id, ")")
  bodyDecode(r.Body, &areq)
  adminCheckCredentials(areq.Credentials)

  storedCompetition = GetRaceStatus(id)
  if storedCompetition != nil {
    if storedCompetition.TimeStamp != areq.Competition.TimeStamp {
      panic("Competition can not be overwritten: TimeStamp is differ")
    }
  } else {
    err := AllocNewCompetitionId(id)
    if err != nil {
      panic(err)
    }
  }

  areq.Competition.CompetitionId = id
  SetRaceStatus(id, areq.Competition)

  json, _ := json.MarshalIndent(resp, "", "  ")
  w.Write(json)
}

func AdminSetTerminalsHandler(w http.ResponseWriter, r *http.Request) {
  var areq AdminRequestCompetitionSet
  var resp AdminResponse
  var v = mux.Vars(r)
  var id uint64
  var storedCompetition *RaceStatus

  defer adminResultHandler(w)

  id = extractUint64(v, "CompetitionId")
  log.Println("Admin::Competition(", id ,")::Terminals:Set")
  bodyDecode(r.Body, &areq)
  adminCheckCredentials(areq.Credentials)

  storedCompetition = GetRaceStatus(id)
  if storedCompetition == nil {
    panic("Unknown competition Id")
  }

  SetTerminalStatus(id, areq.TerminalList)

  json, _ := json.MarshalIndent(resp, "", "  ")
  w.Write(json)
}

func AdminTerminalListHandler(w http.ResponseWriter, r *http.Request) {
  var areq AdminRequestGet
  var resp AdminResponseTerminalList

  defer adminResultHandler(w)

  log.Println("Admin::TerminalList::Get()")
  bodyDecode(r.Body, &areq)
  adminCheckCredentials(areq.Credentials)

  resp.TerminalList = GetShortTerminals()

  json, _ := json.MarshalIndent(resp, "", "  ")
  w.Write(json)
}
