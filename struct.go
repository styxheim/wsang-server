package main

type Error struct {
  Text string
}

type Discipline struct {
  Id uint32
  Name string
  Gates []int32
}

type TerminalDiscipline struct {
  Id uint32
  Gates []int32 // contain unsigned values: START gate (-2) and FINISH gate (-3)
}

type RaceStatus struct {
  CompetitionId uint64
  CompetitionName string `json:",omitempty"`
  SyncPoint *uint64 `json:",omitempty"`
  TimeStamp uint64
  Gates []uint32 `json:",omitempty"`
  Penalties []uint32 `json:",omitempty"`
  Crews  []uint32 `json:",omitempty"`
  Disciplines []Discipline `json:",omitempty"`
  IsActive *bool `json:",omitempty"`
}

type TerminalStatusActivity struct {
  LastActivity uint64
}

type TerminalPermission struct {
  Read  bool `json:",omitempty"`
  Write bool `json:",omitempty"`
  Admin bool `json:",omitempty"`
}

type TerminalStatus struct {
  TimeStamp uint64
  TerminalString string `json:"TerminalId"`
  Disciplines []TerminalDiscipline `json:",omitempty"`
  Permissions TerminalPermission `json:",omitempty"`
  Activity TerminalStatusActivity
}

type LapGate struct {
  Id uint32 `json:"Gate"`
  PenaltyId uint32 `json:"Penalty"`
}

type Lap struct {
  Id uint64 `json:"LapId"`
  TimeStamp uint64
  DisciplineId uint32
  CrewId *uint32 `json:",omitempty"`
  LapId *uint32 `json:"LapNumber,omitempty"`
  StartTime *uint64 `json:",omitempty"`
  FinishTime *uint64 `json:",omitempty"`
  Gates []LapGate `json:",omitempty"`
}

type AdminRequest struct {
  RaceStatus *RaceStatus `json:",omitempty"`
  TerminalStatus []TerminalStatus `json:",omitempty"`
}

type AdminResponse struct {
  Competitions []RaceStatus `json:",omitempty"`
  TerminalStatus []TerminalStatus
  Error *Error `json:",omitempty"`
}

type DataResult struct {
  RaceStatus *RaceStatus `json:",omitempty"`
  TerminalStatus []TerminalStatus `json:",omitempty"`
  Lap []Lap `json:",omitempty"`
  Error *Error `json:",omitempty"`
}

func GetCompetition(Id uint64, TerminalString *string, TimeStamp uint64) DataResult {
  var ares DataResult
  var rstat = GetRaceStatus(Id)

  if rstat != nil {
    if TimeStamp == 0 || rstat.TimeStamp > TimeStamp {
      ares.RaceStatus = rstat
    }
  } else {
    return ares;
  }

  ares.TerminalStatus = GetTerminals(&Id, TerminalString, TimeStamp)
  if( Id == 0 ) {
    // default competition:
    // 1. no laps
    // 2. all TimeStamp to zero
    for k := range ares.TerminalStatus {
      ares.TerminalStatus[k].TimeStamp = 0;
    }
    ares.RaceStatus.TimeStamp = 0;
  } else {
    ares.Lap = GetLaps(Id, TimeStamp)
  }

  return ares
}

