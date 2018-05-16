#!/bin/bash
go build
docker build -t jasonknight/gonad .
