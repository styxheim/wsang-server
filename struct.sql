
CREATE TABLE terminal (
  id INTEGER,
  hex_id TEXT,
  timestamp INTEGER
);

CREATE TABLE terminal_gates (
  terminal_id INTEGER,
  gate_id INTEGER,
  discipline_id INTEGER
);

CREATE TABLE penalties (
  id INTEGER,
  value INTEGER
);

CREATE TABLE gates_to_discplines (
  discipline_id INTEGER,
  gate_id INTEGER
);

CREATE TABLE disciplines (
  discipline_id INTEGER,
  title TEXT
);

CREATE TABLE result (
  id INTEGER,
  start_time_ms INTEGER,
  finish_time_ms INTEGER
);

CREATE TABLE gates_result (
  result_id INTEGER,
  gate_id INTEGER,
  penalty_id INTEGER
);

CREATE TABLE event (
  id INTEGER,
  timestamp INTEGER,
  terminal_id INTEGER,
  event_type TEXT,
  event_data TEXT
);

