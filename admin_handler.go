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
  termString := v["TerminalString"]
  log.Println("ADMIN GET", r.URL)

  defer adminResultHandler(w)

  UpdateTerminalActivity(termString)
  term := GetTerminals(nil, &termString, 0)
  if len(term) == 0 || !term[0].Permissions.Admin {
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
  var adreq AdminRequest
  var termString = v["TerminalString"]
  var CompetitionId = extractUint64(v, "CompetitionId")
  var receive_time_ms = uint64(time.Now().UnixNano() / 1000000)

  err := json.NewDecoder(r.Body).Decode(&adreq)
  if err != nil {
    panic(err);
  }

  UpdateTerminalActivity(termString)
  var newRace = (CompetitionId == 0);
  var term = GetTerminals(&CompetitionId, &termString, 0);

  if len(term) == 0 || !term[0].Permissions.Admin {
    panic("Apply parameters allowed only from admin terminals");
  }

  if newRace {
    CompetitionId = AllocNewCompetitionId()
    log.Println("New CompetitionId", CompetitionId);
  }

  if adreq.RaceStatus == nil {
    adreq.RaceStatus = &RaceStatus{};
  }

  adreq.RaceStatus.CompetitionId = CompetitionId;
  adreq.RaceStatus.TimeStamp = receive_time_ms;
  SetRaceStatus(adreq.RaceStatus.CompetitionId, *adreq.RaceStatus);

  for _, v := range adreq.TerminalStatus {
    v.TimeStamp = receive_time_ms;
  }
  SetTerminalStatus(adreq.RaceStatus.CompetitionId, adreq.TerminalStatus)

  data, _ := json.Marshal(adreq)
  SaveToJournal(CompetitionId,
                receive_time_ms,
                termString,
                fmt.Sprintf("%s", r.URL), data)

  w.Write([]byte("{}"));
}

func ActivityHandler(w http.ResponseWriter, r *http.Request) {
  log.Println("ADMIN GET", r.URL)
  defer adminResultHandler(w)
  v := mux.Vars(r)
  termString := v["TerminalString"]

  UpdateTerminalActivity(termString)
  term := GetTerminals(nil, &termString, 0)
  if len(term) == 0 || !term[0].Permissions.Admin {
    panic("Not an admin terminal")
  }

  data, _ := json.MarshalIndent(GetTerminals(nil, nil, 0), "", "  ")
  w.Write(data)
}
