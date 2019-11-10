package main

import (
  "log"
  "fmt"
  "time"
  "net/http"
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
  w.WriteHeader(http.StatusOK)
}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
  w.WriteHeader(http.StatusOK)
}

