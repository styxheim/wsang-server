#!/usr/bin/env python3
# -*- coding: utf-8 -*
# vim: ft=python ff=unix fenc=utf-8 cc=120 et ts=2 sw=2
# file: app/__init__.py
"""
.. module: __init__

"""
import os
import time
from copy import deepcopy
from flask import abort
from flask import Flask
from flask import redirect
from flask import request
from flask import make_response
from flask import send_file
from json import loads as _json_extract
from json import dumps as json_serialize
app = Flask(__name__)

GATE_START = -2
GATE_FINISH = -3

def json_extract(_input):
  """ ??? """
  if type(_input) == bytes:
    return _json_extract(_input.decode('UTF-8'))
  return _json_extract(_input)

def timestamp():
  return int(time.time() * 1000)

class Server:
  def __init__(self):
    self.laps_data = []
    self.crews_data = []
    self.classes = []
    if os.path.exists("db/laps_data"):
      with open("db/laps_data", "r") as f:
        self.laps_data = json_extract(f.read())
    if os.path.exists("db/crews_data"):
      with open("db/crews_data", "r") as f:
        self.crews_data = json_extract(f.read())
    if os.path.exists("db/classes"):
      with open("db/classes", "r") as f:
        self.classes = json_extract(f.read())

  def copy(self) -> list:
    return deepcopy(self.laps_data)

  def save(self, new_laps_data : list):
    with open("db/laps_data", "w") as f:
      f.write(json_serialize(new_laps_data, indent=2))
    self.laps_data = deepcopy(new_laps_data)

  def copyCrews(self) -> list:
    return deepcopy(self.crews_data)

  def saveCrews(self, new_crews_data : list):
    with open("db/crews_data", "w") as f:
      f.write(json_serialize(new_crews_data, indent=2))
    self.crews_data = deepcopy(new_crews_data)

  def copyClasses(self) -> list:
    return deepcopy(self.classes)

  def saveClasses(self, new_classes : list):
    with open("db/classes", "w") as f:
      f.write(json_serialize(new_classes, indent=2))
    self.classes = deepcopy(new_classes)

server = Server()

def setTerminalInfo(term) -> dict:
  path = "db/term/%s" % term['TerminalId']
  with open(path, 'w') as f:
    f.write(json_serialize(term, indent=2))
  return term

def newTerminal(TerminalId : str) -> dict:
  path = "db/term/%s" % TerminalId
  term = { 'TimeStamp': timestamp(),
           'TerminalId': TerminalId,
           'Gates': [GATE_START, GATE_FINISH] }
  return setTerminalInfo(term)

def getTerminalInfo(TerminalId : str) -> dict:
  path = "db/term/%s" % TerminalId
  if os.path.exists(path):
    with open(path, 'r') as f:
      return json_extract(f.read())
  return None

def setRaceStatus(status : dict):
  path = "db/race"
  jstatus = json_serialize(status, indent=2)
  with open(path, 'w') as f:
    f.write(jstatus)
  return status

def getRaceStatus() -> dict:
  _id = timestamp();
  RaceStatus = {
      'TimeStamp': _id,
      'CompetitionId': _id,
      'SyncPoint': timestamp(),
      'Penalties': [],
      'Crews': [],
      'Gates': [GATE_START, GATE_FINISH]
  };
  path = "db/race"

  if not os.path.exists(path):
    return setRaceStatus(RaceStatus)
  with open(path, 'r') as f:
    return json_extract(f.read());

@app.route('/api/timesync/<int:begin_time>', methods=['GET'])
def timesync(begin_time):
  receive_time = int(time.time() * 1000)
  return "%s:%s:%s" % (begin_time, receive_time, int(time.time() * 1000))

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
        # false start
        false_start = False
        if 'StartTime' in new_data:
          if _lap.get('StartTime', 0) != 0 and new_data['StartTime'] == 0:
            false_start = True
        # others
        _gates = dict([(_g['Gate'], _g['Penalty']) for _g in _lap.get('Gates', [])])
        _gates.update(dict([(_g['Gate'], _g['Penalty']) for _g in new_data.get('Gates', [])]))
        _lap.update(new_data)
        _ngates = [{'Gate': x, 'Penalty': _gates[x]} for x in _gates.keys()]
        _lap['Gates'] = _ngates
        if false_start:
          _lap['FinishTime'] = 0
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
    term = newTerminal(TerminalId)
    #return abort(403)

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

@app.route('/race', methods=['GET'])
def raceConfig():
  RaceStatus = getRaceStatus()
  page = '<a href="/">to index</a>'
  page += '<hr/>'
  page += '<form action="/race/edit" method="POST">'
  penalties = ', '.join([str(i) for i in RaceStatus["Penalties"][1:]])
  page += '<div>'
  page += '<span>Penalties<span>&nbsp;'
  page += '<input type="text" value="%s" name="penalties">' % penalties
  page += '<div/>'
  gates = ', '.join([str(i) for i in RaceStatus["Gates"] if i not in [GATE_START, GATE_FINISH]])
  page += '<div>'
  page += '<span>Gates<span>&nbsp;'
  page += '<input type="text" value="%s" name="gates">' % gates
  page += '<div/>'

  page += '<div>'
  page += '<span>Cleanup race<span>&nbsp;'
  page += '<input name="reset_race" type="checkbox"/>'
  page += '<div/>'

  page += '<div>'
  page += '<span>Cleanup members<span>&nbsp;'
  page += '<input name="reset_members" type="checkbox"/>'
  page += '<div/>'


  page += '<div><input type="submit"/></div>'
  page += '</form>'

  return page + crew()

@app.route('/race/edit', methods=['POST'])
def raceConfigEdit():
  RaceStatus = getRaceStatus()
  pns = []
  try:
    for p in request.form['penalties'].split(','):
      pns += [int(i) for i in p.split(' ') if i]
  except ValueError as e:
    return abort(400, 'Invalid \'penalties\' input: %s' % e)

  gts = []
  try:
    for p in request.form['gates'].split(','):
      gts += [int(i) for i in p.split(' ') if i]
  except ValueError as e:
    return abort(400, 'Invalid \'gates\' input: %s' % e)

  RaceStatus['Gates'] = [GATE_START] + gts + [GATE_FINISH]
  RaceStatus['Penalties'] = [0] + pns
  RaceStatus['TimeStamp'] = timestamp()
  if 'reset_race' in request.form:
    RaceStatus['CompetitionId'] = RaceStatus['TimeStamp']
    RaceStatus['SyncPoint'] = timestamp()
    server.save([])
  if 'reset_members' in request.form:
    server.saveClasses([])
    server.saveCrews([])
  setRaceStatus(RaceStatus);
  return redirect('/race')

@app.route('/terminal/', methods=['GET'])
def terminal():
  RaceStatus = getRaceStatus()
  termids = []

  page = '<a href="/">to index</a>'
  page += '<hr/>'

  for name in os.listdir('db/term'):
    if not name.startswith('.'):
      termids.append(name)

  terminds = sorted(termids)

  page += '<form action="/terminal/edit" method="POST">'
  for TerminalId in termids:
    TerminalInfo = getTerminalInfo(TerminalId)
    page += '<div><h3>%s</h3>' % TerminalId
    for gateId in RaceStatus['Gates']:
      gate_name = "%s_%s" % (TerminalId, gateId)
      checked = 'checked="checked"'
      gate_title = 'Gate %d' % gateId
      if gateId not in TerminalInfo['Gates']:
        checked = ""

      if gateId == GATE_FINISH:
        gate_name = "%s_finish" % (TerminalId)
        gate_title = "Finish"
      if gateId == GATE_START:
        gate_name = "%s_start" % (TerminalId)
        gate_title = "Start"
      page += '<div><input name="%s" type="checkbox" %s/> %s</div>' % (gate_name, checked, gate_title)
    page += '</div>'

  page += '<div><input type="submit"/></div>'
  page += '</form>'

  return page

@app.route('/terminal/edit', methods=['POST'])
def terminal_edit():
  RaceStatus = getRaceStatus()
  termids = []

  for name in os.listdir('db/term'):
    if not name.startswith('.'):
      termids.append(name)

  for TerminalId in termids:
    TerminalInfo = getTerminalInfo(TerminalId)
    gts = []
    for gate_name in request.form:
      t, gate_title = gate_name.split('_')
      if t != TerminalInfo['TerminalId']:
        continue
      if gate_title == 'finish':
        gts.append(GATE_FINISH)
      elif gate_title == 'start':
        gts.append(GATE_START)
      else:
        if int(gate_title) not in RaceStatus['Gates']:
          continue
        gts.append(int(gate_title))

    TerminalInfo['TimeStamp'] = timestamp()
    TerminalInfo['Gates'] = gts
    setTerminalInfo(TerminalInfo);
  return redirect('/terminal/')


def genHtmlTable(table_result, filter_crews=[], filter_class=[], filter_laps=[]):
  # print
  page = ''
  RaceStatus = getRaceStatus()
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
  page += '<th>Class</th>'
  page += '</tr>'

  x = 0
  for i in range(0, len(table_result)):
    result = table_result[i]
    if filter_laps:
      if result[0][0] not in filter_laps:
        continue
    if filter_crews:
      if result[1][0] not in filter_crews:
        continue
    if filter_class:
      if result[-1][0] not in filter_class:
        continue
    page += '<tr>'
    x += 1
    page += '<th>%s</th>' % x
    for ii in range(0, len(result)):
      col = result[ii]
      if ii == 0:
        page += '<td><a href="?lap=%d">' % col[0]
        page += '<div>%s</div>' % col[1]
        page += '</a></td>'
      elif ii == 1:
        page += '<td><a href="?crew=%d">' % col[0]
        page += '<div>%s</div>' % col[1]
        page += '</a></td>'
      else:
        page += '<td>%s</td>' % col[1]
    page += '</tr>'
  page += '</table>'
  return page

@app.route('/', methods=['GET'])
def index():
  page = ''
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
    no_penalty = False

    for gateId in RaceStatus['Gates']:
      if gateId == GATE_START:
        startTime = lap.get('StartTime', 0)
        row.append((startTime, ms2str(startTime)))
      elif gateId == GATE_FINISH:
        finishTime = lap.get('FinishTime', 0)
        row.append((finishTime, ms2str(finishTime)))
      else:
        penalty = 0
        try:
          penalty = RaceStatus['Penalties'][getLapGatePenaltyId(lap, gateId)]
        except IndexError:
          pass
        penalty_sum += penalty
        if getLapGatePenaltyId(lap, gateId) == 0:
          no_penalty = True
          row.append((penalty, ""))
        else:
          row.append((penalty, str(penalty)))
    # summary time
    if no_penalty:
      row.append((0, ""))
    else:
      row.append((penalty_sum, str(penalty_sum)))

    if not finishTime or not startTime:
      row.append((0, ""))
      row.append((0, ""))
    elif finishTime <= startTime:
      row.append((0, "???"))
      row.append((0, "???"))
    else:
      result = finishTime - startTime
      row.append((result, ms2str(result)))
      if no_penalty:
        row.append((0, "???"))
      else:
        result_overall = result + (penalty_sum * 1000)
        row.append((result_overall, ms2str(result_overall)))

    crew_class = getDataForCrew(row[1][0])['class']
    row.append((crew_class, str(crew_class)))

    table_result.append(row)

  # sort
  #table_result = sorted(table_result.copy(), key=lambda x: 0 if x[9][0] == 0 else -9999999999 + x[9][0])

  filter_lap = []
  filter_class = []
  filter_crew = []
  if 'lap' in request.args:
    for v in dict(request.args)['lap']:
      try:
        filter_lap.append(int(v))
      except ValueError:
        pass
  if 'class' in request.args:
    for v in dict(request.args)['class']:
      filter_class.append(v)
    if filter_class[0] == '':
      filter_class = []
  if 'crew' in request.args:
    for v in dict(request.args)['crew']:
      try:
        filter_crew.append(int(v))
      except ValueError:
        pass

  page += """<style>
  @media print {
    .noprint {
      display: none;
    }
  }
  table {
    border-spacing: 0px 0px;
  }
  th {
    border: 1px solid black;
  }
  td {
    border: 1px solid black;
    text-align: center;
  }
  </style>"""

  page += '<div class="noprint">'

  page += '<a href="/results/one?class=%s">One best</a>' % '&class='.join(filter_class)
  page += '<span>&nbsp;&nbsp;&nbsp;</span>'
  page += '<a href="/results/two?class=%s">Sum of two best</a>' % '&class='.join(filter_class)
  page += '<span>&nbsp;&nbsp;&nbsp;</span>'
  page += '|'
  page += '<span>&nbsp;&nbsp;&nbsp;</span>'
  page += '<a href="/race">Configure Race</a>'
  page += '<span>&nbsp;&nbsp;&nbsp;</span>'
  page += '<a href="/terminal">Configure Terminals</a>'
  page += '<span>&nbsp;&nbsp;&nbsp;</span>'
  page += '|'
  page += '<span>&nbsp;&nbsp;&nbsp;</span>'
  link = "https://yadi.sk/d/zEmMdPMmACUwhQ/%D0%AF%20%D0%A1%D1%83%D0%B4%D1%8C%D1%8F%20%202.0.apk"
  if os.path.exists('app.apk'):
    link = "/storage/app.apk"
  page += '<a href="' + link + '">Android app download</a>'

  page += '<hr>'

  classes = [''] + server.copyClasses()
  for _class in classes:
    _style = ""
    if _class in filter_class or ( not filter_class and _class == '' ):
      _style = "background: grey;"
    page += '<a href="?class=%s" style="%s">' % (_class, _style)
    page += '<span style="padding-left: 20px; padding-right: 20px;">%s</span>' % ('&rang;' if _class == '' else _class)
    page += '</a>&nbsp;'

  page += '<span>&nbsp;&nbsp;&nbsp;</span>'
  page += '|'
  page += '<span>&nbsp;&nbsp;&nbsp;</span>'

  page += "<span class='link_class' onclick='window.print()'>Print</span>"
  page += '<hr>'

  page += '</div>'

  page += genHtmlTable(table_result, filter_crews=filter_crew, filter_class=filter_class, filter_laps=filter_lap)

  page += '<div class="noprint">'
  page += '</div>'

#  return str(table_result)
  return page

def genHtmlList(_id : str, _inlist : list, _value : str) -> str:
  r = ''
  r += '<select name="%s">' % _id
  r += '<option value="_"></option>'
  for name in _inlist:
    selected = ''
    if name == _value:
      selected = ' selected'
    r += '<option value="%s"%s>%s</option>' % (name, selected, name)
  r += '</select>'
  return r

def crew():
  page = '<hr>'
  classes = server.copyClasses()

  crews_data = server.copyCrews()

  page += '<div><form action="/crew/edit" method="POST">'
  page += 'Classed:&nbsp;'
  page += '<input type="text" name="classes" value="%s" size="100">' % ', '.join(classes)
  page += '<input type="submit">'
  page += '</form></div>'

  page += '</hr>'

  page += '<form method="POST" action="/crew/edit">'
  page += '<table border=1 cellpadding=6>'
  page += '<tr><th>Number</th><th>Class</th><th>delete</th><th>Info</th></tr>'
  for i in range(0, len(crews_data)):
    crew = crews_data[i]
    page += '<tr>'
    page += '<td><input type="text" name="id_%s" value="%s"></td>' % (i, crew['id'])
    page += '<td>%s</td>' % genHtmlList('class_%s' % i, classes, crew['class'])
    page += '<td><input type="checkbox" name="del_%s"></td>' % i
    page += '<td>'
    page += '<div><input type="text" name="qq_%s" value="%s" size="40"></div>' % (i, crew.get('name', ''))
    page += '<div><textarea name="ee_%s" rows="6" cols="50">%s</textarea></div>' % (i, '\n'.join(crew.get('members', [])))
    page += '</td>'
    page += '</tr>'

  page += '<tr>'
  page += '<td><input type="text" name="id_new" placeholder="Crew number"></td>'
  page += '<td>%s</td>' % genHtmlList('class_new', classes, '')
  page += '<td></td>'
  page += '<td>'
  page += '<div><input type="text" name="qq_new" value="" placeholder="Name of crew" size="32"> Name and members</div>'
  page += '<div><textarea name="ee_new" rows="6" cols="50" placeholder="First Member\nSecond Member"></textarea></div>'
  page += '</td>'

  page += '</tr>'

  page += '</table>'
  page += '<input type="submit">'
  page += '</form>'

  return page

@app.route('/crew/edit', methods=["POST"])
def crew_edit():
  crews_data = server.copyCrews()
  new_data = []
  crews_list = []

  if 'classes' in request.form:
    ncls = []
    for q in request.form['classes'].split(','):
      ncls += [i for i in q.split(' ') if i]
    server.saveClasses(ncls)
    return redirect('/race')

  for i in list(range(0, len(crews_data))) + ['new']:
    crew = dict()
    if i != 'new':
      crew = crews_data[i]
    name_id = 'id_%s' % i
    name_class = 'class_%s' % i
    name_del = 'del_%s' % i
    name_title = 'qq_%s' % i
    name_members = 'ee_%s' % i

    if name_del in request.form:
      continue

    if name_id in request.form and name_class in request.form:
      if not request.form[name_id]:
        continue
      crew_id = int(request.form[name_id])
      if crew_id not in crews_list:
        crew['id'] = crew_id
        crew['class'] = request.form[name_class]
        crew['name'] = request.form[name_title]
        crew['members'] = [e.strip() for e in request.form[name_members].split('\n')]
        crews_list.append(crew_id)
      new_data.append(crew)

  server.saveCrews(new_data)
  RaceStatus = getRaceStatus()
  RaceStatus['Crews'] = crews_list
  RaceStatus['TimeStamp'] = timestamp()
  setRaceStatus(RaceStatus)
  return redirect('/race')

@app.route('/storage/app.apk', methods=['GET'])
def storate_app():
  """
  download apk file
  """
  if os.path.exists('app.apk'):
    res = make_response(send_file('../app.apk', cache_timeout=0))
    return res
  return abort(404)


# ??? ??? ???

class CrewResult:
  lap = 0
  crew = 0
  result = 0

  def __add__(self, other):
    n = CrewResult()
    n.lap = self.lap
    n.crew = self.crew
    n.result = self.result + other.result
    return n

  def __radd__(self, other):
    n = CrewResult()
    n.lap = self.lap
    n.crew = self.crew
    n.result = self.result
    return n

  def __repr__(self):
    return '<%s lap=%d, crew=%d, result=%d %d>' % (self.__class__.__name__,
                                                   self.lap,
                                                   self.crew,
                                                   self.result,
                                                   id(self))

def getResults() -> list:
  RaceStatus = getRaceStatus()
  laps = server.copy()
  results = []

  for lap in laps:
    r = CrewResult()
    start = lap.get('StartTime', 0)
    finish = lap.get('FinishTime', 0)
    r.lap = lap.get('LapNumber', 0)
    r.crew = lap.get('CrewNumber', 0)
    penalty = 0
    for gateId in RaceStatus.get('Gates', []):
      if gateId in [GATE_FINISH, GATE_START]:
        continue
      try:
        penalty += RaceStatus['Penalties'][getLapGatePenaltyId(lap, gateId)]
      except IndexError:
        pass
    if not start or not finish or finish < start:
      continue
    r.result = finish - start + penalty * 1000
    results.append(r)

  return results

def getDataForCrew(crewId : int) -> str:
  crews = server.copyCrews()
  for crew in crews:
    if crew['id'] == crewId:
      return deepcopy(crew)
  return {'name': '', 'class': '', 'members': [], 'id': crewId}

@app.route('/results/<string:mode>')
def results(mode):
  page = """<html><head><head><body>"""
  page += """
  <style>
  table {
    border-spacing: 0px 0px;
  }
  th {
    padding: 10px;
    border: 2px solid black;
  }
  td {
    border: 1px solid black;
    padding: 10px;
    text-align: center;
  }
  .link_class {
    padding-left: 20px;
    padding-right: 20px;
  }
  .th_time {
    width: 100px;
  }
  .th_place {
    width: 100px;
  }
  .title {
    width: 200px;
  }
  .members {
    width: 400px;
  }
  @media print {
    #class_selector {
      display: none;
    }
  }
  /*
  @page {
    size: 21cm 29.7cm;
    margin: 30mm 45mm 30mm 45mm;
  }
  */
  </style>
  """

  page += "<div id='class_selector'>"
  page += '<a href="/">to index</a>'
  page += '<hr/>'
  filter_class = []
  if 'class' in request.args:
    for v in dict(request.args)['class']:
      filter_class.append(v)
    if filter_class[0] == '':
      filter_class = []

  sum([CrewResult(), CrewResult()])
  classes = [''] + server.copyClasses()
  for _class in classes:
    _style = ""
    if _class in filter_class or ( not filter_class and _class == '' ):
      _style = "background: grey;"
    page += "<a href='?class=%s&limit=%s' class='link_class' style='%s'>" % (_class, request.args.get('limit', ''), _style)
    page += '<span class="link_class">%s</span>' % ('&rang;' if _class == '' else _class)
    page += '</a>&nbsp;'
  page += "<span class='link_class' onclick='window.print()'>Печать</span>"
  if request.args.get('limit', ''):
    page += "<a href='?class=%s&limit='><span class='link_class'>Три лучших</span></a>" % '&class='.join(filter_class)
  else:
    page += "<a href='?class=%s&limit=3'><span class='link_class'>Все</span></a>" % '&class='.join(filter_class)

  page += '<hr>'
  page += "</div>"

  rs = []
  _rs = getResults()
  if filter_class:
    _rs = [i for i in _rs.copy() if getDataForCrew(i.crew)['class'] in filter_class]
  _rs = sorted(_rs, key=lambda x: x.result)

  if mode == 'one':
    _ne = []
    for x in _rs:
      if x.crew not in _ne:
        _ne.append(x.crew)
        rs.append(x)
    result_title = "Результат по лучшей попытке"
    place_title = "Место по лучшей попытке"
  if mode == 'two':
    _XX = dict()
    for x in _rs:
      if x.crew not in _XX:
        _XX[x.crew] = []
      _XX[x.crew].append(x)
    for x in _XX.values():
      if len(x) >= 2:
        x = x[0] + x[1]
        rs.append(x)
    rs = sorted(rs, key=lambda x: x.result)
    result_title = "Результат по сумме лучших попыток"
    place_title = "Место по сумме двух лучших попыток"

  page += """
  <table border=1>
    <tr>
        <th>Класс</th>
        <th>Экипаж</th>
        <th>Название</th>
        <th>ФИО</th>
        <th class='th_time'>_result_</th>
        <th class='th_place'>_place_</th>
    </tr>
  """

  page = page.replace('_result_', result_title)
  page = page.replace('_place_', place_title)

  if request.args.get('limit', ''):
    _range = range(0, min(int(request.args.get('limit', 3)), len(rs)))
  else:
    _range = range(0, len(rs))

  for i in _range:
    r = rs[i]
    _data = getDataForCrew(r.crew)
    page += "<tr>"
    page += "<td>%s</td>" % _data['class']
    page += "<td>%s</td>" % r.crew
    page += "<td class='title'>%s</td>" % _data['name']
    page += "<td class='members'>%s</td>" % ''.join(["<div>%s</div>" % x for x in _data['members']])
    page += "<td>%s</td>" % ms2str(r.result)
    page += "<td>%s</td>" % (i + 1)
    page += "<tr>"

  page += """</body></html>"""
  return page
