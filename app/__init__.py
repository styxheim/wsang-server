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

RaceStatus = {
      'TimeStamp': 1,
      'CompetitionId': int(time.time()),
      'Gates': [1, 2, 3, 4, 5, 6, 7, 13],
      'Penalties': [10, 25, 50],
      'Crews': [1, 2, 4, 12, 27, 32, 46],
      'Disciplines': [{'Id': 1, 'Name': 'Квалификация', 'Gates': [GATE_START, GATE_FINISH, 1, 4, 13]},
                      {'Id': 2, 'Name': 'Слалом', 'Gates': [GATE_START, GATE_FINISH, 1, 2, 4, 5, 6, 7]},
                      {'Id': 3, 'Name': 'Длинная гонка', 'Gates': [GATE_START, GATE_FINISH]}
                     ]
    };

# only one terminal now
TerminalStatus = [
    { 'TimeStamp': 2,
      'TerminalId': 'a8b5af9c5cbe2a06',
      'Gates': [GATE_START],
    }
    ];

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

@app.route('/api/laps/ok', methods=['GET'])
def ok():
  return '"ok"'

def getTerminalInfo(TerminalId : str):
  for term in TerminalStatus:
    if term['TerminalId'] == TerminalId:
      return term
  return None

@app.route('/api/update/<int:CompetitionId>/<string:TerminalId>', methods=['POST'])
def update(CompetitionId : int, TerminalId : str):
  """
  update laps data
  """
  global laps_data
  term = getTerminalInfo(TerminalId)
  if term is None:
    return abort(403) # forbidden

  if CompetitionId != RaceStatus['CompetitionId']:
    return abort(404) # not found

  update_laps_data = server.copy()
  new_data_list = json_extract(request.data)

  for new_data in new_data_list:
    new_data['TimeStamp'] = int(time.time() * 1000) # timestamp with microseconds
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
