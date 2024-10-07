#!/usr/bin/env bash

env GOOS=linux GOARCH=amd64 go build

# scp ezlabctl ezmeral@10.1.1.32:/home/ezmeral/
cp ezlabctl ~/Applications/ua-rpm/ezlabctl