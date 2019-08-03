#!/usr/bin/env python3
# -*- coding: utf-8 -*
# vim: ft=python ff=unix fenc=utf-8 cc=120 et ts=4
# file: app/__init__.py
"""
.. module: __init__

"""
from flask import Flask
from flask import request
from json import loads as json_extract
from json import dumps as json_serialize
app = Flask(__name__)

laps_data = {}

@app.route('/api/laps/ok', methods=['GET'])
def ok():
    return '"ok"'


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

@app.route('/api/laps/updatelaps', methods=['POST'])
def updatelaps():
    # Получение значение от судьи на старте
    list_data = json_extract(request.data)
    for el in list_data:
        if server_make_key:
            no = len(laps_data) + 1;
            el['LapId'] = no
        else:
            no = el['LapId']
        laps_data[no] = el
        print("Update record '%s': %s" % (no, el))
    return "true"

