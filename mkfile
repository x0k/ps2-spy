#!/bin/bash -e

all:
    go run cmd/ps2-feed-ds/main.go -t $BOT_TOKEN
