#!/bin/sh
# vim: ft=sh ff=unix fenc=utf-8
# file: run.sh
export FLASK_ENV=development

flask run --port 9001 --host 0.0.0.0

