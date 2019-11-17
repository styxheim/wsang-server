package main

type Error struct {
  Text string
}

type Discipline struct {
  Id uint32
  Name string
  Gates []uint32
}

type TerminalDiscipline struct {
  Id uint32
  Gates []uint32
}

type RaceStatus struct {
  CompetitionId uint64
  SyncPoint uint64
  TimeStamp uint64
  Gates []uint32 `json:",omitempty"`
  Penalties []uint32 `json:",omitempty"`
  Crews  []uint32 `json:",omitempty"`
  Disciplines []Discipline `json:",omitempty"`
}

type TerminalStatus struct {
  TimeStamp uint64
  TerminalId string
  Disciplines []TerminalDiscipline `json:",omitempty"`
}

type LapGate struct {
  Id uint32 `json:"Gate"`
  PenaltyId uint32 `json:"Penalty"`
}

type Lap struct {
  Id uint64 `json:"LapId"`
  TimeStamp uint64
  DisciplineId uint32
  CrewId uint32
  LapId uint32 `json:"LapNumber"`
  StartTime *uint64 `json:",omitempty"`
  FinishTime *uint64 `json:",omitempty"`
  Gates []LapGate `json:",omitempty"`
}

type ApiResult struct {
  RaceStatus *RaceStatus `json:",omitempty"`
  TerminalStatus []TerminalStatus `json:",omitempty"`
  Lap []Lap `json:",omitempty"`
  Error *Error;
}

func GetCompetition(Id uint64, TimeStamp uint64) ApiResult {
  var ares ApiResult
  var rstat = GetRaceStatus(Id)

  if rstat != nil {
    if TimeStamp == 0 || rstat.TimeStamp > TimeStamp {
      ares.RaceStatus = GetRaceStatus(Id)
    }
  }

  ares.Lap = GetLaps(Id, TimeStamp)

  return ares
}

