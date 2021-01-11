package main

import (
  "os"
  "net/http"
  "log"
  "github.com/gorilla/mux"
)

func main() {
  var bind_address = "127.0.0.1:9001";

  if len(os.Args) > 1 {
    bind_address = os.Args[1];
  }

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
  // Wipe competition
  r.HandleFunc("/api/admin/competition/set/{CompetitionId:[0-9]+}/wipe", AdminWipeComptition).Methods("POST")
  // Set terminal list for competition
  r.HandleFunc("/api/admin/competition/terminals/set/{CompetitionId:[0-9]+}", AdminSetTerminalsHandler).Methods("POST")
  // Update Competition.SyncPoint to current
  r.HandleFunc("/api/admin/competition/syncpoint/{CompetitionId:[0-9]+}", AdminSyncPointHandler).Methods("POST")
  // List available terminals
  r.HandleFunc("/api/admin/terminal/list", AdminTerminalListHandler).Methods("POST")

  r.Use(loggingMiddleware)

  log.Fatal(http.ListenAndServe(bind_address, r))
}

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
        // Call the next handler, which can be another middleware in the chain, or the final handler.
        next.ServeHTTP(w, r)
    })
}
