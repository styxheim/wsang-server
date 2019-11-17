package main

import (
  "log"
  "fmt"
  "time"
  "net/http"
  "strconv"
  "encoding/json"
  "github.com/gorilla/mux"
)

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
  var ares ApiResult

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
  id, err := strconv.ParseUint(v["CompetitionId"], 10, 64)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  ts, err := strconv.ParseUint(v["TimeStamp"], 10, 64)
  if err != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  ares = GetCompetition(id, ts)
}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
  w.WriteHeader(http.StatusOK)
}

