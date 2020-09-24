package main

import (
  "net/http"
  "log"
  "github.com/gorilla/mux"
)

func main() {
  var bind_address = "0.0.0.0:9001";
  log.Println("Server started at", bind_address)
  r := mux.NewRouter().StrictSlash(true)
  r.HandleFunc("/api/timesync/{begin_time:[0-9]+}", TimeSyncHandler).Methods("GET")
  // We need handle default competition separatly (not in /api/data/0/). Look to GetCompetition()
  //r.HandleFunc("/api/data/which/{TerminalString:[0-9a-fA-F]+}")
  r.HandleFunc("/api/data/{CompetitionId:[0-9]+}/{TimeStamp:[0-9]+}/{TerminalString:[0-9a-fA-F]+}", GetDataHandler).Methods("POST")
  r.HandleFunc("/api/update/{CompetitionId:[0-9]+}/{TerminalString:[0-9a-fA-F]+}", UpdateHandler).Methods("POST")

  // List competitions
  r.HandleFunc("/api/admin/competition/list", AdminListHandler).Methods("POST")
  // Get status of competitions
  r.HandleFunc("/api/admin/comeptition/get/{CompetitionId:[0-9]+}", AdminGetCompetitionHandler).Methods("POST")
  // Set competition and terminals configuration
  r.HandleFunc("/api/admin/competition/set/{CompetitionId:[0-9]+}", AdminSetCompetitionHandler).Methods("POST")
  // Update Competition.SyncPoint to current
  r.HandleFunc("/api/admin/competition/syncpoint/{CompetitionId:[0-9]+}", AdminSyncPointHandler).Methods("POST")
  // List available terminals
  r.HandleFunc("/api/admin/terminal/list", AdminTerminalListHandler).Methods("POST")

  log.Fatal(http.ListenAndServe(bind_address, r))
}

