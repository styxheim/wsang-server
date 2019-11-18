package main

import (
  "net/http"
  "log"
  "github.com/gorilla/mux"
)

func main() {
  log.Println("Server started")
  r := mux.NewRouter().StrictSlash(true)
  r.HandleFunc("/api/timesync/{begin_time:[0-9]+}", TimeSyncHandler).Methods("GET")
  r.HandleFunc("/api/data/{CompetitionId:[0-9]+}/{TimeStamp:[0-9]+}/{TerminalString:[0-9a-fA-F]+}", GetDataHandler).Methods("GET")
  r.HandleFunc("/api/update/{CompetitionId:[0-9]+}/{TerminalString:[0-9a-fA-F]+}", UpdateHandler).Methods("POST")

  r.HandleFunc("/api/stats/activity", ActivityHandler).Methods("GET")

  log.Fatal(http.ListenAndServe(":9001", r))
}

