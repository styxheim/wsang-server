package main

import (
  "fmt"
  "log"
  "time"
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
  log.Println("ADMIN GET", r.URL)

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
  log.Println("ADMIN POST", r.URL)
  defer adminResultHandler(w)

  var v = mux.Vars(r)
  var rstat RaceStatus
  var termString = v["TerminalString"]
  var receive_time = uint64(time.Now().UnixNano() / 1000000)

  err := json.NewDecoder(r.Body).Decode(&rstat);
  if err != nil {
    panic(err);
  }

  var CompetitionId = rstat.CompetitionId;
  var term = GetTerminals(&CompetitionId, &termString, 0);

  if CompetitionId == 0 {
    panic("Create new race not allowed now");
  }

  if len(term) == 0 || !term[0].Permissions.Admin {
    panic("Apply parameters allowed only from admin terminals");
  }

  rstat.TimeStamp = receive_time;

  SetRaceStatus(CompetitionId, rstat);

  UpdateTerminalActivity(v["TerminalString"])
  data, _ := json.Marshal(rstat)
  SaveToJournal(CompetitionId,
                receive_time,
                termString,
                fmt.Sprintf("%s", r.URL), data)

  w.Write([]byte("{}"));
}
