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
  r.HandleFunc("/api/data/{CompetitionId:[0-9]+}/{TimeStamp:[0-9]+}/{TerminalString:[0-9a-fA-F]+}", GetDataHandlerOld).Methods("GET")
  r.HandleFunc("/api/data/{CompetitionId:[0-9]+}/{TimeStamp:[0-9]+}/{TerminalString:[0-9a-fA-F]+}", GetDataHandler).Methods("POST")
  r.HandleFunc("/api/update/{CompetitionId:[0-9]+}/{TerminalString:[0-9a-fA-F]+}", UpdateHandler).Methods("POST")

  fs := http.FileServer(http.Dir("res/"));

  r.HandleFunc("/", IndexHandler).Methods("GET")
  r.HandleFunc("/index.html", IndexHandler).Methods("GET")
  r.Handle("/favicon.ico", fs).Methods("GET");
  r.Handle("/js/{File}", fs).Methods("GET")
  r.Handle("/css/{File}", fs).Methods("GET")

  r.Handle("/app/", fs)
  r.Handle("/app/{File}", fs)

  r.HandleFunc("/api/admin/competitions/{TerminalString:[0-9a-fA-F]+}", AdminListHandler).Methods("GET")
  r.HandleFunc("/api/admin/competition/set/{CompetitionId:[0-9]+}/{TerminalString:[0-9a-fA-F]+}", AdminHandler).Methods("POST")

  r.HandleFunc("/api/admin/activity/{TerminalString:[0-9a-fA-F]+}", ActivityHandler).Methods("GET")

  log.Fatal(http.ListenAndServe(":9001", r))
}

