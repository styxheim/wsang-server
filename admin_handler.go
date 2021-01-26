package main

import (
  "io"
  "fmt"
  "time"
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
  if DataResponse.RaceStatus == nil {
    var message = fmt.Sprintf("Competition '%d' not exists", id)

    panic(message);
  }
  resp.Competition = *DataResponse.RaceStatus
  resp.TerminalList = GetTerminals(id, nil, 0)

  json, _ := json.MarshalIndent(resp, "", "  ")
  w.Write(json)
}

func AdminWipeComptition(w http.ResponseWriter, r *http.Request) {
  var areq AdminRequestCompetitionSet
  var resp AdminResponse
  var v = mux.Vars(r)
  var id uint64
  var storedCompetition *RaceStatus
  var now = uint64(time.Now().UTC().UnixNano() / 1000000)

  defer adminResultHandler(w)

  id = extractUint64(v, "CompetitionId")
  log.Println("Admin::Competition(", id, "):Wipe")
  bodyDecode(r.Body, &areq)
  adminCheckCredentials(areq.Credentials)

  storedCompetition = GetRaceStatus(id)
  if storedCompetition == nil {
    panic(fmt.Sprintf("Unknown competititon %q", id))
  }

  storedCompetition.TimeStamp = now
  storedCompetition.SyncPoint = &now
  SetRaceStatus(id, *storedCompetition)
  WipeLaps(id)

  json, _ := json.MarshalIndent(resp, "", "  ")
  w.Write(json)
}

func AdminSetActiveComptition(w http.ResponseWriter, r *http.Request) {
  var areq AdminRequestCompetitionSet
  var resp AdminResponse
  var v = mux.Vars(r)
  var id uint64
  var storedCompetition *RaceStatus

  defer adminResultHandler(w)

  id = extractUint64(v, "CompetitionId")
  log.Println("Admin::Competition(", id, "):SetActive")
  bodyDecode(r.Body, &areq)
  adminCheckCredentials(areq.Credentials)

  storedCompetition = GetRaceStatus(id)
  if storedCompetition == nil {
    panic(fmt.Sprintf("Unknown competititon %q", id))
  }

  if storedCompetition.SyncPoint == nil {
    /* this do not need update competition's timestamp
     * because only competition's link changed, not competition's state
     */
    syncPoint := uint64(time.Now().UnixNano())
    storedCompetition.SyncPoint = &syncPoint
    SetRaceStatus(id, *storedCompetition)
  }

  MakeDefaultCompetitionId(id)

  json, _ := json.MarshalIndent(resp, "", "  ")
  w.Write(json)
}

func AdminSetCompetitionHandler(w http.ResponseWriter, r *http.Request) {
  var areq AdminRequestCompetitionSet
  var resp AdminResponse
  var v = mux.Vars(r)
  var id uint64
  var storedCompetition *RaceStatus
  var now = uint64(time.Now().UTC().UnixNano() / 1000000)

  defer adminResultHandler(w)

  id = extractUint64(v, "CompetitionId")
  log.Println("Admin::Competition:Set(", id, ")")
  bodyDecode(r.Body, &areq)
  adminCheckCredentials(areq.Credentials)

  storedCompetition = GetRaceStatus(id)
  if storedCompetition != nil {
    /* Allow update competition status only if new data based on stored */
    if storedCompetition.TimeStamp != areq.Competition.TimeStamp {
      panic("Competition can not be overwritten: You must update data before set new.")
    }
    areq.Competition.TimeStamp = now
  } else {
    err := AllocNewCompetitionId(id)
    if err != nil {
      panic(err)
    }
  }

  if areq.Competition.CompetitionName == "" {
    panic("CompetitionName must be set")
  }

  areq.Competition.CompetitionId = id
  /* ignore syncpoint in request */
  areq.Competition.SyncPoint = nil
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
  var now = uint64(time.Now().UTC().UnixNano() / 1000000)

  defer adminResultHandler(w)

  id = extractUint64(v, "CompetitionId")
  log.Println("Admin::Competition(", id ,")::Terminals:Set")
  bodyDecode(r.Body, &areq)
  adminCheckCredentials(areq.Credentials)

  storedCompetition = GetRaceStatus(id)
  if storedCompetition == nil {
    panic("Unknown competition Id")
  }

  for i, _ := range areq.TerminalList {
    areq.TerminalList[i].TimeStamp = now
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
