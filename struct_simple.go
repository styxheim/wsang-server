package main

import (
  "io/ioutil"
  "os"
  "fmt"
  "log"
  "time"
  "path"
  "strconv"
  "encoding/json"
)

func competitionRoot(CompetitionId uint64) string {
  return fmt.Sprintf("db/%d", CompetitionId)
}

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

  journal_data := fmt.Sprintf("%d:%s:%s:%s\n", TimeStamp, TerminalString, url, data)

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

func SetTerminalStatus(CompetitionId uint64, tstat []TerminalStatus) {
  fpath := competitionPath(&CompetitionId, "terminals")
  data, _ := json.MarshalIndent(tstat, "", "  ")

  store(fpath, data, true);
}

func SetRaceStatus(CompetitionId uint64, rstat RaceStatus) {
  fpath := competitionPath(&CompetitionId, "race")

  if CompetitionId == 0 {
    panic("can't setup default race");
  }

  if rstat.SyncPoint == nil {
    oldRace := GetRaceStatus(CompetitionId)

    if oldRace != nil {
      rstat.SyncPoint = oldRace.SyncPoint
    }
  }

  data, _ := json.MarshalIndent(rstat, "", "  ")
  store(fpath, data, true)
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

  // 0 is special competition id: default
  if CompetitionId != 0 {
    if rstat.CompetitionId != CompetitionId {
      log.Println("!!!", "race have invalid Id",
                  rstat.CompetitionId, "!=", CompetitionId, fpath)
      panic("Invalid CompetitionId")
    }
    ActiveCompetition := GetRaceStatus(0)
    if ActiveCompetition != nil &&
       ActiveCompetition.CompetitionId == rstat.CompetitionId {
      rstat.IsActive = new(bool)
      *rstat.IsActive = true
    }
  } else {
    // CompetitionId == 0 is marker for active race
    rstat.IsActive = new(bool)
    *rstat.IsActive = true
  }

  return &rstat
}

func mergeGates(lgates []LapGate, gates []LapGate) ([]LapGate, bool) {
  updated := false

  for _, g := range gates {
    found := false

    for k := range lgates {
      if g.Id != lgates[k].Id {
        continue
      }
      found = true
      if lgates[k].PenaltyId != g.PenaltyId {
        lgates[k].PenaltyId = g.PenaltyId
        updated = true
      }
    }

    if !found {
      lgates = append(lgates, g)
      updated = true
    }
  }

  return lgates, updated
}

func store(fpath string, data []byte, safe bool) {
  var safeName = fpath

  if safe {
    safeName = fmt.Sprintf("%s.%d", fpath, time.Now().UTC().UnixNano())
  }

  err := ioutil.WriteFile(safeName, data, 0644)
  if err != nil {
    log.Println("!!!", " write error", err, fpath)
    panic("write error")
  }

  if safe {
    basename := path.Base(safeName)
    os.Remove(fpath);
    err = os.Symlink(basename, fpath)
    if err != nil {
      log.Println("!!!", "symlink error", "<", err, ">", fpath)
      panic("symlink error")
    }
  }
}

func storeLaps(CompetitionId uint64, new_laps []Lap) {
  fpath := competitionPath(&CompetitionId, "laps")
  json, _ := json.MarshalIndent(new_laps, "", "  ")
  store(fpath, json, true)
}

func WipeLaps(CompetitionId uint64) {
  storeLaps(CompetitionId, []Lap{Lap{}});
}

func UpdateLaps(CompetitionId uint64, new_laps []Lap, TimeStamp uint64) {
  claps := getLaps(CompetitionId)
  updated := false

  for _, nl := range new_laps {
    found := false
    nl.TimeStamp = TimeStamp

    for k := range claps {
      if nl.Id != claps[k].Id {
        continue
      }
      found = true

      if nl.DisciplineId != nil {
        if claps[k].DisciplineId == nil {
          panic(fmt.Sprintf("Discipline not specified for lap id: %d", claps[k].LapId))
        }
        // check DisciplineId when present
        if *nl.DisciplineId != *claps[k].DisciplineId {
          log.Println("!!!", "Discipline migration not allowed",
                      nl.DisciplineId, "!=", claps[k].DisciplineId,
                      "for Id", claps[k].Id)
          panic("Invalid DisciplineId")
        }
      }

      claps[k].TimeStamp = nl.TimeStamp

      if nl.CrewId != nil &&
         (claps[k].CrewId == nil || claps[k].CrewId != nl.CrewId) {
        claps[k].CrewId = nl.CrewId
        updated = true
      }
      if nl.LapId != nil &&
         (claps[k].LapId == nil || claps[k].LapId != nl.LapId) {
        claps[k].LapId = nl.LapId
        updated = true
      }

      if nl.StartTime != nil &&
         (claps[k].StartTime == nil || claps[k].StartTime != nl.StartTime) {
        claps[k].StartTime = nl.StartTime;
        updated = true
      }

      if nl.FinishTime != nil &&
         (claps[k].FinishTime == nil || claps[k].FinishTime != nl.FinishTime) {
        claps[k].FinishTime = nl.FinishTime
        updated = true
      }

      if nl.Strike != nil &&
         (claps[k].Strike == nil || claps[k].Strike != nl.Strike) {
        claps[k].Strike = nl.Strike
        updated = true
      }

      var gates_updated bool
      claps[k].Gates, gates_updated = mergeGates(claps[k].Gates, nl.Gates)
      if gates_updated {
        updated = true
      }
    }
    if !found {
      if nl.DisciplineId == nil {
        panic("Insert new data not allowed without DisciplineId")
      }
      claps = append(claps, nl)
      updated = true
    }
  }

  if updated {
    storeLaps(CompetitionId, claps)
  }
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

func GetTerminals(CompetitionId uint64, TerminalString *string, TimeStamp uint64) []TerminalStatus {
  var rterms []TerminalStatus = make([]TerminalStatus, 0)
  var activities = getTerminalsActivity()
  var terms = getTerminals(&CompetitionId)

  for _, t := range terms {
    if TerminalString != nil {
      if t.TerminalString != *TerminalString {
        continue
      }
    }

    if TimeStamp == 0 || t.TimeStamp > TimeStamp {
      t.Activity = activities[t.TerminalString]
      rterms = append(rterms, t)
    }
  }

  return rterms
}

func GetShortTerminals() []TerminalStatusShort {
  var rterms []TerminalStatusShort = make([]TerminalStatusShort, 0)
  var activities = getTerminalsActivity()

  for terminal, activity := range activities {
    var termShort TerminalStatusShort

    termShort.TerminalId = terminal
    termShort.Activity = activity
    rterms = append(rterms, termShort)
  }

  return rterms
}

var terminalActivities []byte

func getTerminalsActivity() map[string]TerminalStatusActivity {
  var r map[string]TerminalStatusActivity

  if len(terminalActivities) != 0 {
    err := json.Unmarshal(terminalActivities, &r)
    if err != nil {
      log.Println("!!!", "terminal activity decode error", err);
      return make(map[string]TerminalStatusActivity)
    }
  } else {
    r = make(map[string]TerminalStatusActivity)
  }

  return r
}

func setTerminalsActivity(r map[string]TerminalStatusActivity) {
  data, err := json.Marshal(r)
  if err != nil {
    log.Println("!!!", "terminal activity encode error", err);
  }
  terminalActivities = data
}

func UpdateTerminalActivity(TerminalString string) {
  var r = getTerminalsActivity()

  r[TerminalString] = TerminalStatusActivity{ LastActivity: uint64(time.Now().UTC().UnixNano() / 1000000) };
  setTerminalsActivity(r)
}

func GetCompetitions() []RaceStatus {
  var rstats []RaceStatus = make([]RaceStatus, 0)

  files, err := ioutil.ReadDir("db")
  if err != nil {
    panic(err)
  }

  for _, f := range files {
    CompetitionId, err := strconv.ParseUint(f.Name(), 10, 32)
    if err != nil {
      continue
    }
    if CompetitionId == 0 {
      // zero is special competition id
      continue
    }
    if rstat := GetRaceStatus(CompetitionId); rstat != nil {
      rstats = append(rstats, *rstat)
    }
  }

  return rstats
}

func AllocNewCompetitionId(id uint64) error {
  fpath := competitionPath(&id, "race")
  return os.MkdirAll(path.Dir(fpath), os.ModePerm);
}

func MakeDefaultCompetitionId(CompetitionId uint64) {
  defLink := competitionRoot(0)
  os.Remove(defLink)
  err := os.Symlink(fmt.Sprintf("%d", CompetitionId), defLink)
  if err != nil {
    panic(fmt.Sprintln("Cannot setup race:", err))
  }
}