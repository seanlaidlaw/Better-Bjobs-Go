#!/usr/bin/env bash

# set environment variables that specify host to compile for
# this allows me to compile the linux versions from my mac
env GOOS=linux GOARCH=amd64 \
	go build bj.go
