#!/usr/bin/env python3
# -*- coding: utf-8 -*
# vim: ft=python ff=unix fenc=utf-8 cc=120 et ts=2 sw=2
# file: app/__init__.py
"""
.. module: __init__

"""
import os
import time
from flask import abort
from flask import Flask
from flask import request
from json import loads as json_extract
from json import dumps as json_serialize
app = Flask(__name__)

GATE_START = -2
GATE_FINISH = -3

def timestamp():
  return int(time.time() * 1000)

class Server:
  def __init__(self):
    self.laps_data = []
    if os.path.exists("db/laps_data"):
      with open("db/laps_data", "r") as f:
        self.laps_data = json_extract(f.read())

  def copy(self) -> list:
    return self.laps_data.copy()

  def save(self, new_laps_data : list):
    with open("db/laps_data", "w") as f:
      f.write(json_serialize(new_laps_data, indent=2))
    self.laps_data = new_laps_data

server = Server()

def newTerminal(TerminalId : str):
  path = "db/term/%s" % TerminalId
  term = { 'TimeStamp': timestamp(),
           'TerminalId': TerminalId,
           'Gates': [GATE_START, GATE_FINISH] }
  with open(path, 'w') as f:
    f.write(json_serialize(term, indent=2))
  return term

def getTerminalInfo(TerminalId : str):
  path = "db/term/%s" % TerminalId
  if os.path.exists(path):
    with open(path, 'r') as f:
      return json_extract(f.read())
  else:
    newTerminal(TerminalId)

def getRaceStatus():
  _id = timestamp();
  RaceStatus = {
      'TimeStamp': _id,
      'CompetitionId': _id,
      'Penalties': [],
      'Crews': [],
      'Gates': [GATE_START, GATE_FINISH]
  };
  path = "db/race"

  if not os.path.exists(path):
    with open(path, 'w') as f:
      f.write(json_serialize(RaceStatus, indent=2))
    return RaceStatus
  with open(path, 'r') as f:
    return json_extract(f.read());

@app.route('/api/update/<int:CompetitionId>/<string:TerminalId>', methods=['POST'])
def update(CompetitionId : int, TerminalId : str):
  """
  update laps data
  """
  RaceStatus = getRaceStatus()
  term = getTerminalInfo(TerminalId)
  if term is None:
    return abort(403) # forbidden

  if CompetitionId != RaceStatus['CompetitionId']:
    return abort(404) # not found

  update_laps_data = server.copy()
  new_data_list = json_extract(request.data)

  print("Request -> \n%s" % json_serialize(new_data_list, indent=2))

  for new_data in new_data_list:
    new_data['TimeStamp'] = timestamp()
    founded = False

    if 'LapId' not in new_data:
      return abort(406) # not acceptable

    for _lap in update_laps_data:
      if _lap['LapId'] == new_data['LapId']:
        _lap.update(new_data)
        founded = True

    if not founded:
      update_laps_data.append(new_data)

  try:
    server.save(update_laps_data)
    return "true"
  except Exception as exc:
    print("Exception: %s" % (exc))
    return abort(500)

  return abort(400) # generic error

@app.route('/api/data/<int:CompetitionId>/<int:TimeStamp>/<string:TerminalId>', methods=['GET'])
def data(CompetitionId : int, TimeStamp : int, TerminalId : str):
  """
  send to client race data
  """
  RaceStatus = getRaceStatus()
  response = {}

  term = getTerminalInfo(TerminalId)
  if term is None:
    return abort(403)

  if CompetitionId != RaceStatus['CompetitionId']:
    # send full data
    TimeStamp = 0

  if TimeStamp < RaceStatus['TimeStamp']:
    response['RaceStatus'] = RaceStatus

  if TimeStamp < term['TimeStamp']:
    response['TerminalStatus'] = [term]

  laps_data = server.copy()
  laps = []
  for lap in laps_data:
    if TimeStamp < lap['TimeStamp']:
      laps.append(lap)

  if laps:
    response['Lap'] = laps

  print("Response -> \n%s" % json_serialize(response, indent=2))
  return json_serialize(response)


# Old compatable
@app.route('/api/laps', methods=['GET'])
def laps():
  laps_data = server.copy()
  response = []

  for lap in laps_data:
    response.append({'LapId': lap['LapId'],
                     'CrewNumber': lap['CrewNumber'],
                     'LapNumber': lap['LapNumber']})

  return json_serialize(response)

@app.route('/api/laps/updategates', methods=['POST'])
def update_gates():
  RaceStatus = getRaceStatus()
  laps_data = server.copy()
  new_data_list = json_extract(request.data)

  print("Request -> \n%s" % json_serialize(new_data_list, indent=2))

  for new_data in new_data_list:
    try:
      penalty = RaceStatus['Penalties'].index(new_data['PenaltySec'])
    except:
      print("# Unknwon penalty: %s" %(new_data['PenaltySec']))
      return abort(400)

    for lap in laps_data:
      if lap['LapId'] == new_data['LapId']:
        if not 'Gates' in lap:
          lap['Gates'] = []
        found = False
        for gate in lap['Gates']:
          if gate['Gate'] == new_data['GateNumber']:
            gate['Penalty'] = penalty
            found = True
            break
        if not found:
          lap['Gates'].append({'Gate': new_data['GateNumber'],
                               'Penalty': penalty})
        lap['TimeStamp'] = timestamp()
        break
  server.save(laps_data)
  return "true"


@app.route('/api/laps/updatefinish', methods=['POST'])
def update_finish():
  laps_data = server.copy()
  new_data_list = json_extract(request.data)
  finish_time = 0

  print("Request -> \n%s" % json_serialize(new_data_list, indent=2))

  for new_data in new_data_list:
    try:
      h, m, s = new_data['FinishTime'].split(':')
      finish_time += int(h) * 60 * 60
      finish_time += int(m) * 60
      finish_time += int(s)
      finish_time *= 1000
    except ValueError:
      print('# Invalid time format: %s', request.data['FinishTime'])
      return abort(400)

    for lap in laps_data:
      if lap['LapId'] == new_data['LapId']:
        lap['FinishTime'] = finish_time
        lap['TimeStamp'] = timestamp()
        break

  server.save(laps_data)
  return "true"


def getLapGatePenaltyId(lap : dict, gateId : int) -> int:
  for gate in lap.get('Gates', []):
    if gate['Gate'] == gateId:
      return gate['Penalty']
  return 0

def ms2str(timestamp : int) -> str:
  ms = timestamp % 1000
  s = int(timestamp / 1000)
  m = int(s / 60)
  h = int (m / 60)
  return "%02d:%02d:%02d.%02d" % (h, m % 60, s % 60, ms / 10)

def str2ms(timestring : str) -> int:
  finish_time = 0
  try:
    h, m, s = timestring.split(':')
    finish_time += int(h) * 60 * 60
    finish_time += int(m) * 60
    if '.' in s:
      s, ms = s.split('.')
      ms += '0' * (3 - len(ms))
      finish_time += int(s)
      finish_time *= 1000
      finish_time += int(ms)
    else:
      finish_time += int(s)
      finish_time *= 1000
  except ValueError:
    return 9999999999999

  return finish_time

@app.route('/', methods=['GET'])
def index():
  page = ""
  RaceStatus = getRaceStatus()
  laps_data = server.copy()
  # prepare
  table_result = []
  for lap in laps_data:
    row = [(lap.get('LapNumber', 0),
            str(lap.get('LapNumber', '???'))),
           (lap.get('CrewNumber', 0),
            str(lap.get('CrewNumber', '???')))]
    startTime = 0
    finishTime = 0
    penalty_sum = 0

    for gateId in RaceStatus['Gates']:
      if gateId == GATE_START:
        startTime = lap.get('StartTime', 0)
        row.append((startTime, ms2str(startTime)))
      elif gateId == GATE_FINISH:
        finishTime = lap.get('FinishTime', 0)
        row.append((finishTime, ms2str(finishTime)))
      else:
        penalty = RaceStatus['Penalties'].get(getLapGatePenaltyId(lap, gateId), 0)
        penalty_sum += penalty
        row.append((penalty, str(penalty)))
    # summary time
    if not penalty_sum:
      row.append((0, ""))
    else:
      row.append(str(penalty_sum))
    if not finishTime or not startTime:
      row.append((0, ""))
      row.append((0, ""))
    elif finishTime <= startTime:
      row.append((0, "???"))
      row.append((0, "???"))
    else:
      result = finishTime - startTime
      result_overall = result + penalty_sum
      row.append((result, ms2str(result)))
      row.append((result_overall, ms2str(result_overall)))

    table_result.append(row)

  # sort
  table_result = sorted(table_result.copy(), key=lambda x: 0 if x[5][0] == 0 else -9999999999 + x[5][0])

  # print
  page += '<table border=1 cellpadding=6>'

  page += '<tr>'
  page += '<th>#</th>'
  page += '<th>Lap</th>'
  page += '<th>Crew</th>'
  for gateId in RaceStatus['Gates']:
    if gateId == GATE_START:
      page += '<th>Start</th>'
      continue
    elif gateId == GATE_FINISH:
      page += '<th>Finish</th>'
      continue
    page += '<th>Gate %s</th>' % gateId
  page += '<th>Penalties sum</th>'
  page += '<th>Result</th>'
  page += '<th>Result with penalties</th>'
  page += '</tr>'

  for i in range(0, len(table_result)):
    result = table_result[i]
    page += '<tr>'
    page += '<th>%s</th>' % i
    for col in result:
      page += '<td>%s</td>' % col[1]
    page += '</tr>'
  page += '</table>'

#  return str(table_result)
  return page
