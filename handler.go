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

func GetDataHandler(w http.ResponseWriter, r *http.Request) {
  var ares DataResult

  defer func() {
    if r := recover(); r != nil {
      log.Println("!!!", "got error", r)
      ares.Error = &Error{ Text: fmt.Sprintf("%s", r) }
    }

    json, _ := json.MarshalIndent(ares, "", "  ")
    w.Write(json)
  }()

  log.Println("GET", r.URL)

  v := mux.Vars(r)
  id := extractUint64(v, "CompetitionId")
  ts := extractUint64(v, "TimeStamp")
  termString := v["TerminalString"]

  UpdateTerminalActivity(v["TerminalString"])
  term := GetTerminals(&id, &termString, 0)
  if len(term) != 1 {
    panic("terminal not recognized")
  }

  if term[0].Permissions.Admin == true {
    ares = GetCompetition(id, nil, ts)
  } else {
    ares = GetCompetition(id, &termString, ts)
  }
}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
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
  term := GetTerminals(&CompetitionId, &termString, 0)
  if len(term) != 1 {
    panic("terminal not recognized")
  }

  if term[0].Permissions.Write == false {
    panic("terminal is readonly")
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

func ActivityHandler(w http.ResponseWriter, r *http.Request) {

  defer func() {
    if r := recover(); r != nil {
      log.Println("!!!", "got error", r)
      w.Write([]byte(fmt.Sprintf("%s", r)))
    }
  }()


  data, _ := json.MarshalIndent(GetTerminals(nil, nil, 0), "", "  ")
  w.Write(data)
}
