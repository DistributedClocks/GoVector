#!/bin/bash
git checkout ClientServer.go
sudo -E go install ../
GoVector -file=ClientServer.go -v
#dovid -file=client-nodoc.go -c
#dovid -file=client-nodoc.go -v
gofmt -w=true ClientServer.go
