package main

import (
  "io/ioutil"
  "fmt"
  "log"
  "encoding/json"
)


func competitionPath(CompetitionId uint64, name string) string {
  return fmt.Sprintf("db/%d/%s", CompetitionId, name)
}

/* TODO: check db on start */

func getLaps(CompetitionId uint64) []Lap {
  var laps []Lap
  path := competitionPath(CompetitionId, "laps")

  data, err := ioutil.ReadFile(path)
  if err != nil {
    log.Println("...", "no laps data", path)
    return nil
  }

  err = json.Unmarshal(data, &laps)
  if err != nil {
    log.Println("!!!", "laps decode error", err, path)
    return nil
  }

  return laps
}

func GetLaps(CompetitionId uint64, TimeStamp uint64) []Lap {
  var rlaps []Lap

  for _, l := range getLaps(CompetitionId) {
    if l.TimeStamp > TimeStamp {
      rlaps = append(rlaps, l)
    }
  }

  return rlaps
}

func GetRaceStatus(CompetitionId uint64) *RaceStatus {
  var rstat RaceStatus
  path := competitionPath(CompetitionId, "race")

  data, err := ioutil.ReadFile(path)
  if err != nil {
    log.Println("...", "no race data", path)
    return nil
  }

  err = json.Unmarshal(data, &rstat)
  if err != nil {
    log.Println("!!!", "race decode error", err, path)
    return nil
  }

  if rstat.CompetitionId != CompetitionId {
    log.Println("!!!", "race have invalid Id",
                rstat.CompetitionId, "!=", CompetitionId, path)
    panic("Invalid CompetitionId")
  }

  return &rstat
}
