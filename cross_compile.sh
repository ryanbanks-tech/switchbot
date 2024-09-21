#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o switchbot-linux-amd64

GOOS=windows GOARCH=amd64 go build -o switchbot-windows-amd64.exe

GOOS=darwin GOARCH=arm64 go build -o switchbot-darwin-arm64
