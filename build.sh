#!/usr/bin/env bash

env GOOS=linux GOARCH=amd64 go build

cp ezlabctl ../ua-rpm/ezlabctl