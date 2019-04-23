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
        newlist.append({'LapId': el['LapId'],
                        'LapNumber': el['LapNumber'],
                        'CrewNumber': el['CrewNumber']})
    print(json_serialize(newlist))
    return json_serialize(newlist)


@app.route('/api/laps/updatelaps', methods=['POST'])
def updatelaps():
    list_data = json_extract(request.data)
    for el in list_data:
        no = el['LapId']
        laps_data[no] = el
        print("Update record '%s': %s" % (no, el))
    return "true"

