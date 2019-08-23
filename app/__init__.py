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

RaceStatus = {
      'TimeStamp': 1,
      'CompetitionId': timestamp(),
      'Gates': [1, 2, 3, 4, 5, 6, 7, 13, 42, 101],
      'Penalties': [0, 1, 10, 25, 32, 46, 50],
      'Crews': [1, 2, 4, 12, 27, 32, 46],
      'Disciplines': [{'Id': 1, 'Name': 'Квалификация', 'Gates': [GATE_START, GATE_FINISH, 1, 4, 13]},
                      {'Id': 2, 'Name': 'Слалом', 'Gates': [GATE_START, GATE_FINISH, 1, 2, 4, 5, 6, 7]},
                      {'Id': 3, 'Name': 'Длинная гонка', 'Gates': [GATE_START, GATE_FINISH]}
                     ]
    };

TerminalStatus = [ ];

class Server:
  def __init__(self):
    self.laps_data = []
    if os.path.exists("laps_data"):
      with open("laps_data", "r") as f:
        self.laps_data = json_extract(f.read())

  def copy(self) -> list:
    return self.laps_data.copy()

  def save(self, new_laps_data : list):
    with open("laps_data", "w") as f:
      f.write(json_serialize(new_laps_data, indent=2))
    self.laps_data = new_laps_data

server = Server()

def newTerminal(TerminalId : str):
  term = { 'TimeStamp': timestamp(),
           'TerminalId': TerminalId,
           'Gates': [GATE_START, 1, 2, 4, 3, 5, 101, GATE_FINISH] }
  TerminalStatus.append(term)
  return term

def getTerminalInfo(TerminalId : str):
  for term in TerminalStatus:
    if term['TerminalId'] == TerminalId:
      return term
  return newTerminal(TerminalId)

@app.route('/api/update/<int:CompetitionId>/<string:TerminalId>', methods=['POST'])
def update(CompetitionId : int, TerminalId : str):
  """
  update laps data
  """
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
  laps_data = server.copy()
  new_data_list = json_extract(request.data)

  print("Request -> \n%s" % json_serialize(new_data_list, indent=2))

  for new_data in new_data_list:
    try:
      penalty = RaceStatus['Penalties'].index(new_data['PenaltySec'])
    except:
      print("# Unknwon penalty: %s" %(new_data['PenaltySec']))
      return abort(400)

    try:
      RaceStatus['Gates'].index(new_data['GateNumber'])
    except:
      print("# Unknwon gate: %s" %(new_data['GateNumber']))
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
