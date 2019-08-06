#!/usr/bin/env python3
# -*- coding: utf-8 -*
# vim: ft=python ff=unix fenc=utf-8 cc=120 et ts=4
# file: app/__init__.py
"""
.. module: __init__

"""
import os
from flask import abort
from flask import Flask
from flask import request
from json import loads as json_extract
from json import dumps as json_serialize
app = Flask(__name__)

laps_data = {}

@app.route('/api/laps/ok', methods=['GET'])
def ok():
    return '"ok"'

def save_dict():
    f = open("laps_data", "w")
    f.write(json_serialize(laps_data))
    f.close();


if os.path.exists("laps_data"):
    with open("laps_data", "r") as f:
        laps_data = json_extract(f.read())

@app.route('/api/laps', methods=['GET'])
def laps():
    newlist = []
    for el in laps_data.values():
        # Список стартов для всех устройств: <LapId: int, CrewNumber: int, LapNumber: int>
        newlist.append({'LapId': el['LapId'],
                        'LapNumber': el['LapNumber'],
                        'CrewNumber': el['CrewNumber']})
    print(json_serialize(newlist))
    return json_serialize(newlist)

# Переключение между старым режимом работы и новым:
# server_make_key = True -> для включения старого режима работы
#   В этом случае, сервер генерирует идентификторы заездов для последущего взаимодействия всех остальных режимов
# server_make_key = False -> новый режим работы
#   Ожидается, что генерировать идентификаторы будет стартовый планшет
server_make_key = True
#

@app.route('/api/laps/updatefinish', methods=['POST'])
def updatefinish():
    # Получение значение от судьи на старте
    list_data = json_extract(request.data)
    for el in list_data:
        no = str(el['LapId'])
        print("# Update record '%s': %s" % (no, el))

        laps_data[no]['FinishTime'] = el['FinishTime']
        save_dict()

    return "true"



@app.route('/api/laps/updatelaps', methods=['POST'])
def updatelaps():

    # Получение значение от судьи на старте
    list_data = json_extract(request.data)
    for el in list_data:
        if server_make_key:
            no = str(len(laps_data) + 1);
            el['LapId'] = no
        else:
            no = str(el['LapId'])
        print("# Update record '%s': %s" % (no, el))

        rval = __import__('random').randint(1, 5)
        sval = __import__('random').randint(1, 5)
        rcode = 200

        if rval == 2:
            rcode = 400

        print("## wait %d seconds, retcode=%d" %( sval, rcode))

        __import__('time').sleep(sval)

        if rval == 2:
         return abort(rcode)

        laps_data[no] = el
        save_dict()

    return "true"

