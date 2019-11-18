package main

import (
  "io/ioutil"
  "os"
  "fmt"
  "log"
  "time"
  "path"
  "encoding/json"
)


func competitionPath(CompetitionId *uint64, name string) string {
  if CompetitionId != nil {
    return fmt.Sprintf("db/%d/%s", CompetitionId, name)
  } else {
    return fmt.Sprintf("db/%s", name)
  }
}

func SaveToJournal(CompetitionId uint64, TimeStamp uint64, TerminalString string, url string, data []byte) {
  fpath := competitionPath(&CompetitionId, "journal")
  f, err := os.OpenFile(fpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

  /* TODO: mutex lock */
  if err != nil {
    log.Println("!!!", "journal open error", err, fpath)
  }
  defer f.Close()

  journal_data := fmt.Sprintf("%d:%s:%s:%s", TimeStamp, TerminalString, url, data)

  if _, err := f.WriteString(journal_data); err != nil {
    log.Println("!!!", "journ write error", err, fpath)
  }
}

/* TODO: check db on start */

func getLaps(CompetitionId uint64) []Lap {
  var laps []Lap
  fpath := competitionPath(&CompetitionId, "laps")

  data, err := ioutil.ReadFile(fpath)
  if err != nil {
    log.Println("...", "no laps data", fpath)
    return nil
  }

  err = json.Unmarshal(data, &laps)
  if err != nil {
    log.Println("!!!", "laps decode error", err, fpath)
    panic("Laps decode error")
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
  fpath := competitionPath(&CompetitionId, "race")

  data, err := ioutil.ReadFile(fpath)
  if err != nil {
    log.Println("...", "no race data", fpath)
    return nil
  }

  err = json.Unmarshal(data, &rstat)
  if err != nil {
    log.Println("!!!", "race decode error", err, fpath)
    return nil
  }

  if rstat.CompetitionId != CompetitionId {
    log.Println("!!!", "race have invalid Id",
                rstat.CompetitionId, "!=", CompetitionId, fpath)
    panic("Invalid CompetitionId")
  }

  return &rstat
}

func mergeGates(lgates []LapGate, gates []LapGate) []LapGate {
  for _, g := range gates {
    found := false

    for _, lg := range lgates {
      if g.Id != lg.Id {
        continue
      }
      found = true
      lg.PenaltyId = g.PenaltyId
    }

    if !found {
      lgates = append(lgates, g)
    }
  }

  return lgates
}

func storeSafe(fpath string, data []byte) {
  safeName := fmt.Sprintf("%s.%d", fpath, time.Now().UnixNano())
  basename := path.Base(safeName)

  err := ioutil.WriteFile(safeName, data, 0644)
  if err != nil {
    log.Println("!!!", " write error", err, fpath)
    panic("write error")
  }

  err = os.Symlink(basename, safeName)
  if err != nil {
    log.Println("!!!", "symlink error", err, fpath)
    panic("symlink error")
  }
}

func storeLaps(CompetitionId uint64, new_laps []Lap) {
  fpath := competitionPath(&CompetitionId, "laps")
  json, _ := json.MarshalIndent(new_laps, "", "  ")
  storeSafe(fpath, json)
}

func UpdateLaps(CompetitionId uint64, new_laps []Lap) {
  claps := getLaps(CompetitionId)

  for _, nl := range new_laps {
    found := false

    for _, cl := range claps {
      if nl.Id != cl.Id {
        continue
      }
      found = true

      // restrict DisciplineId
      if nl.DisciplineId != cl.DisciplineId {
        log.Println("!!!", "Discipline migration not allowed",
                    nl.DisciplineId, "!=", cl.DisciplineId,
                    "for Id", cl.Id)
        panic("Invalid DisciplineId")
      }

      // update simple fields
      cl.CrewId = nl.CrewId
      cl.LapId = nl.LapId
      cl.TimeStamp = nl.TimeStamp

      // update with check
      if nl.StartTime != nil {
        cl.StartTime = nl.StartTime
      }

      if nl.FinishTime != nil {
        cl.FinishTime = nl.FinishTime
      }

      cl.Gates = mergeGates(cl.Gates, nl.Gates)
    }
    if !found {
      claps = append(claps, nl)
    }
  }

  storeLaps(CompetitionId, new_laps)
}

func getTerminals(CompetitionId *uint64) []TerminalStatus {
  var terms []TerminalStatus
  var fpath = competitionPath(CompetitionId, "terminals")

  data, err := ioutil.ReadFile(fpath)
  if err != nil {
    log.Println("...", "no terminal data", fpath)
    return nil
  }

  err = json.Unmarshal(data, &terms)
  if err != nil {
    log.Println("!!!", "terminals decode error", err, fpath);
    panic("Terminal decode error")
  }

  return terms
}

func GetTerminals(CompetitionId *uint64, TerminalString *string, TimeStamp uint64) []TerminalStatus {
  var rterms []TerminalStatus
  var terms = getTerminals(CompetitionId)
  var activity map[string]uint64 = getTerminalsActivity()

  for _, t := range terms {
    if TerminalString != nil && t.TerminalString != *TerminalString {
      continue
    }

    if TimeStamp == 0 || t.TimeStamp > TimeStamp {
      t.LastActivity = activity[t.TerminalString]
      rterms = append(rterms, t)
    }
  }

  return rterms
}

func getTerminalsActivity() map[string]uint64 {
  var r map[string]uint64
  var fpath = competitionPath(nil, "activity")

  data, err := ioutil.ReadFile(fpath)
  if err != nil {
    log.Println("!!!", "terminal activity read error", err, fpath);
    return make(map[string]uint64)
  }

  err = json.Unmarshal(data, &r)
  if err != nil {
    log.Println("!!!", "terminal activity decode error", err, fpath);
    return make(map[string]uint64)
  }

  return r
}

func setTerminalsActivity(r map[string]uint64) {
  var fpath = competitionPath(nil, "activity")

  data, _ := json.MarshalIndent(r, "", "  ")

  err := ioutil.WriteFile(fpath, data, 0644)
  if err != nil {
    log.Println("!!!", "terminal activity write error", err, fpath);
  }
}

func UpdateTerminalActivity(TerminalString string) {
  var r = getTerminalsActivity()

  r[TerminalString] = uint64(time.Now().UnixNano() / 1000000)
  setTerminalsActivity(r)
}
