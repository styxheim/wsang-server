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
    return fmt.Sprintf("db/%d/%s", *CompetitionId, name)
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

func store(fpath string, data []byte, safe bool) {
  var safeName = fpath
  var basename string

  if safe {
    safeName = fmt.Sprintf("%s.%d", fpath, time.Now().UnixNano())
  }
  basename = path.Base(safeName)

  err := ioutil.WriteFile(safeName, data, 0644)
  if err != nil {
    log.Println("!!!", " write error", err, fpath)
    panic("write error")
  }

  if safe {
    err = os.Symlink(basename, safeName)
    if err != nil {
      log.Println("!!!", "symlink error", err, fpath)
      panic("symlink error")
    }
  }
}

func storeLaps(CompetitionId uint64, new_laps []Lap) {
  fpath := competitionPath(&CompetitionId, "laps")
  json, _ := json.MarshalIndent(new_laps, "", "  ")
  store(fpath, json, true)
}

func UpdateLaps(CompetitionId uint64, new_laps []Lap, TimeStamp uint64) {
  claps := getLaps(CompetitionId)

  for _, nl := range new_laps {
    found := false
    nl.TimeStamp = TimeStamp

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

func setTerminals(CompetitionId *uint64, terms []TerminalStatus) {
  var fpath = competitionPath(CompetitionId, "terminals")

  data, _ := json.MarshalIndent(terms, "", "  ")

  if CompetitionId == nil {
    store(fpath, data, false)
  } else {
    /* only for saving in race use safe write */
    store(fpath, data, true)
  }
}

func UpdateTerminals(CompetitionId *uint64, terms []TerminalStatus, TimeStamp uint64) {
  var cterms = getTerminals(CompetitionId)

  for _, nt := range terms {
    found := false
    nt.TimeStamp = TimeStamp

    for i, ct := range cterms {
      if ct.TerminalString != nt.TerminalString {
        continue
      }
      cterms[i] = nt
      found = true
    }

    if !found {
      cterms = append(cterms, nt)
    }
  }
  setTerminals(CompetitionId, cterms)
}

func GetTerminals(CompetitionId *uint64, TerminalString *string, TimeStamp uint64) []TerminalStatus {
  var rterms []TerminalStatus
  var activities = getTerminalsActivity()

  if CompetitionId != nil {
    var terms = getTerminals(CompetitionId)

    for _, t := range terms {
      if TerminalString != nil && t.TerminalString != *TerminalString {
        continue
      }

      if TimeStamp == 0 || t.TimeStamp > TimeStamp {
        t.Activity = activities[t.TerminalString]
        rterms = append(rterms, t)
      }
    }
  } else {
    for termString, termActivity := range activities {
      rterms = append(rterms, TerminalStatus{ TerminalString: termString,
                                              Activity: termActivity })
    }
  }

  return rterms
}

func getTerminalsActivity() map[string]TerminalStatusActivity {
  var r map[string]TerminalStatusActivity
  var fpath = competitionPath(nil, "activity")

  data, err := ioutil.ReadFile(fpath)
  if err != nil {
    log.Println("!!!", "terminal activity read error", err, fpath);
    return make(map[string]TerminalStatusActivity)
  }

  err = json.Unmarshal(data, &r)
  if err != nil {
    log.Println("!!!", "terminal activity decode error", err, fpath);
    return make(map[string]TerminalStatusActivity)
  }

  return r
}

func setTerminalsActivity(r map[string]TerminalStatusActivity) {
  var fpath = competitionPath(nil, "activity")
  /* TODO: save to mem */

  data, _ := json.MarshalIndent(r, "", "  ")

  err := ioutil.WriteFile(fpath, data, 0644)
  if err != nil {
    log.Println("!!!", "terminal activity write error", err, fpath);
  }
}

func UpdateTerminalActivity(TerminalString string) {
  var r = getTerminalsActivity()

  r[TerminalString] = TerminalStatusActivity{ LastActivity: uint64(time.Now().UnixNano() / 1000000) };
  setTerminalsActivity(r)
}
