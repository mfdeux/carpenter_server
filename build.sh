#! /bin/bash
env GOOS=darwin GOARCH=amd64 go build -o build/carpenter ./src/*.go
