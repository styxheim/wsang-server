package main

import (
  "log"
  "fmt"
  "time"
  "io/ioutil"
  "net/http"
  "strconv"
  "encoding/json"
  "github.com/gorilla/mux"
)

func checkTerminalVersion(version string) bool {
  const LAST_CLIENT_VERSION = "2.3.1"
  const ADMIN_UNIVERSAL_VERSION = "0.0.0"

  if version == LAST_CLIENT_VERSION {
    return true
  }
  /* 0.0.0 is special version for webui admin */
  if version == ADMIN_UNIVERSAL_VERSION {
    return true
  }
  return false
}

func extractUint64(vars map[string]string, vname string) uint64 {
  id, err := strconv.ParseUint(vars[vname], 10, 64)
  if err != nil {
    panic(fmt.Sprintf("Invalid %s format", vname))
  }
  return id
}

func TimeSyncHandler(w http.ResponseWriter, r *http.Request) {
  receive_time := time.Now().UnixNano() / 1000000
  log.Println("TIME", r.URL)

  v := mux.Vars(r)

  w.Write([]byte(fmt.Sprintf("%s:%d:%d",
                             v["begin_time"],
                             receive_time,
                             receive_time)))
}

func GetDataHandlerOld(w http.ResponseWriter, r *http.Request) {
  ares := &DataResponse{ Error: &Error{ Text: "too old. Update your application" } }

  json, _ := json.MarshalIndent(ares, "", "  ")
  w.Write(json)
}

func GetDataHandler(w http.ResponseWriter, r *http.Request) {
  var ares DataResponse
  var dreq DataRequest

  defer func() {
    if r := recover(); r != nil {
      log.Println("!!!", "got error", r)
      ares.Error = &Error{ Text: fmt.Sprintf("%s", r) }
    }

    json, _ := json.MarshalIndent(ares, "", "  ")
    w.Write(json)
  }()

  log.Println("DATA-REQ", r.URL)

  v := mux.Vars(r)
  id := extractUint64(v, "CompetitionId")
  ts := extractUint64(v, "TimeStamp")
  termString := v["TerminalString"]

  UpdateTerminalActivity(termString)
  term := GetTerminals(&id, &termString, 0)
  if len(term) != 1 {
    panic("terminal not recognized")
  }

  err := json.NewDecoder(r.Body).Decode(&dreq)
  if err != nil {
    panic(err);
  }

  if( !checkTerminalVersion(dreq.Version) ) {
    GetDataHandlerOld(w, r)
    return
  }

  if term[0].Permissions.Admin == true {
    ares = GetCompetition(id, nil, ts)
  } else if term[0].Permissions.Read == true {
    ares = GetCompetition(id, &termString, ts)
    rstat := ares.RaceStatus

    if ares.RaceStatus == nil {
      rstat = GetRaceStatus(id);
    }

    if rstat.IsActive == nil || *rstat.IsActive == false {
      panic("competition is closed")
    }
  } else {
    panic("no read permissions")
  }
}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
  log.Println("DATA-NEW", r.URL)
  var laps []Lap
  var receive_time = uint64(time.Now().UnixNano() / 1000000)

  defer func() {
    if r := recover(); r != nil {
      log.Println("!!!", "got error", r)
      w.WriteHeader(http.StatusBadRequest)
    } else {
      w.Write([]byte("true"))
    }
  }()

  v := mux.Vars(r)
  CompetitionId := extractUint64(v, "CompetitionId")
  termString := v["TerminalString"]
  UpdateTerminalActivity(termString)

  race := GetRaceStatus(CompetitionId)
  if race == nil {
    panic("Unknown Competition")
  }

  if race.IsActive == nil || *race.IsActive == false {
    panic("competition is closed")
  }

  term := GetTerminals(&CompetitionId, &termString, 0)
  if len(term) != 1 {
    panic("terminal not registered in competition")
  }

  if term[0].Permissions.Write == false {
    panic("terminal have no write permissions")
  }

  body, err := ioutil.ReadAll(r.Body)
  defer r.Body.Close()
  if err != nil {
    log.Println("!!!", "http body not readed", err)
    panic("Invalid request")
  }

  err = json.Unmarshal(body, &laps)
  if err != nil {
    panic("Invalid json data")
  }

  UpdateLaps(CompetitionId, laps, uint64(receive_time))

  // write log
  data, _ := json.Marshal(laps)
  SaveToJournal(CompetitionId,
                receive_time,
                v["TerminalString"],
                fmt.Sprintf("%s", r.URL), data)
}
