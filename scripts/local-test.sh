#!/bin/bash

start start-cli.cmd

cd ../server.dmz/mind_daemon_dmz
go run ./cmd/mind-daemon-dmz

