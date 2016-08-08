# -*- coding: utf-8 -*-
from invoke import run, task

import multiprocessing
import logging
import socket
import sys
import os

import sys

log = logging.getLogger(__name__)
out_hdlr = logging.StreamHandler(sys.stdout)
out_hdlr.setFormatter(logging.Formatter('%(asctime)s %(message)s'))
out_hdlr.setLevel(logging.INFO)
log.addHandler(out_hdlr)
log.setLevel(logging.INFO)

global_vars = {
    "DEFORMIO_PROJECT": "",
    "DEFORMIO_TOKEN" :"",
    "OAUTH2_CLIENT_ID": "",
    "OAUTH2_CLIENT_SECRET": "",
    "OAUTH2_REDIRECT_URI": "http://127.0.0.1:8000/oauth2callback",
    "OAUTH2_AUTH_URI": "https://accounts.google.com/o/oauth2/auth",
    "OAUTH2_TOKEN_URI": "https://accounts.google.com/o/oauth2/token",
    "HASH_KEY": "",
    "BLOCK_KEY": "",
    "TEMPLATES_DIR": "src/project/templates",
    "HOSTNAME": socket.getfqdn(),
    "DOMAIN_NAME": "https://watchlater.chib.me",

    "MODE": "development",
    "COVER_PROFILE_FILE": "/tmp/c.out",
    "CURDIR": os.path.dirname(os.path.abspath(__file__)),

    "ERROR_LOG": "",
    "VENDOR_DIR": ".vendor",
    "SENTRY_DSN": "",
}

global_vars.update({
    "GOPATH": os.path.join(global_vars["CURDIR"], global_vars["VENDOR_DIR"]),
    "DOCS_HTML": os.path.join(global_vars["CURDIR"], "docs/public/apib/index.html"),
    "DEPENDENCIES": os.path.join(global_vars["CURDIR"], "dependencies.txt"),
})

no_db_modules = [
    "project/modules/args",
    "project/modules/config",
    "project/modules/helpers",
    "project/modules/log",
    "project/modules/messages",
]
http_modules = [
    "project/http_handlers"
]
db_modules = [
    "project/models",
]

modules = []
modules.extend(no_db_modules)
modules.extend(db_modules)
modules.extend(http_modules)

GOCOMMAND = """env GOPATH=%(GOPATH)s \
    http-host=0.0.0.0:8000 \
    http-domain=%(DOMAIN_NAME)s \
    hostname=%(HOSTNAME)s \
    mode=%(MODE)s \
    oauth2-client-id=%(OAUTH2_CLIENT_ID)s \
    oauth2-client-secret=%(OAUTH2_CLIENT_SECRET)s \
    oauth2-redirect-uri=%(OAUTH2_REDIRECT_URI)s \
    oauth2-auth-uri=%(OAUTH2_AUTH_URI)s \
    oauth2-token-uri=%(OAUTH2_TOKEN_URI)s \
    deformio-project=%(DEFORMIO_PROJECT)s deformio-token=%(DEFORMIO_TOKEN)s \
    templates-dir=%(TEMPLATES_DIR)s \
    sentry-dsn=%(SENTRY_DSN)s \
    hash-key=%(HASH_KEY)s block-key=%(BLOCK_KEY)s""" % global_vars

global_vars.update({
    "GOCOMMAND": GOCOMMAND,
})


@task
def remove_deps():
    run("rm -rf %(GOPATH)s" % global_vars, encoding="utf-8")

@task
def copy_src():
    run("mkdir -p %(GOPATH)s" % global_vars, encoding="utf-8")
    run("rm -rf %(GOPATH)s/src/project" % global_vars, encoding="utf-8")
    run("rm -rf %(GOPATH)s/pkg/" % global_vars, encoding="utf-8")
    run("cp -r %(CURDIR)s/src %(GOPATH)s" % global_vars, encoding="utf-8")

@task(pre=[remove_deps, copy_src])
def get(install=True):
    if install:
        with open(global_vars["DEPENDENCIES"], 'r') as f:
            for line in f:
                local_command = "env GOPATH=%(GOPATH)s" % global_vars
                run("%s go get -v %s" % (local_command, line), encoding="utf-8")


@task(pre=[copy_src])
def start_fast(race=False):
    local_command = "%(GOCOMMAND)s go run %(GOPATH)s/src/project/main.go" %global_vars
    if race:
        local_command += " -race "
    run(local_command, encoding="utf-8")


@task(pre=[get, copy_src])
def start():
    run("%(GOCOMMAND)s go run %(GOPATH)s/src/project/main.go" %
        global_vars, encoding="utf-8")


@task(pre=[copy_src])
def test_no_db_modules(race=False, cover=False, report=False, count=1, cpu=0):
    for no_db_module in no_db_modules:
        test_fast(no_db_module, race, cover, report, count)


@task(pre=[copy_src])
def test_fast(module="", race=False, cover=False, report=False, count=1, cpu=0):
    if cpu == 0:
        cpu = multiprocessing.cpu_count()

    if module in modules:
        local_command = "time %(GOCOMMAND)s go test -v " % global_vars
        if race:
            local_command += " -race "
        if cover:
            local_command += " -coverprofile %s" % global_vars[
                "COVER_PROFILE_FILE"]
        local_command += " -count=%d -cpu=%d --parallel %d" % (count, cpu, cpu)

        run("%s %s" % (local_command, module), encoding="utf-8")

        if report:
            run("%s go tool cover -html=%s" %
                (global_vars["GOCOMMAND"], global_vars["COVER_PROFILE_FILE"]), encoding="utf-8")
    else:
        log.error("module %s is not registered" % module)


@task(pre=[copy_src])
def vet():
    for module in modules:
        if module != "project/modules/generators":
            local_command = "%(GOCOMMAND)s go vet " % global_vars
            log.info("Checking %s" % module)
            run("%s %s" % (local_command, module), encoding="utf-8")

@task(pre=[copy_src])
def build(tag=None):
    run("env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOPATH=%(GOPATH)s go build \
            -ldflags '-s' -a -installsuffix cgo -o ./bin/project %(GOPATH)s/src/project/main.go" % global_vars
    )
    if tag:
        run("docker build -t hub.chib.me:443/watch-later:%s ." % tag)