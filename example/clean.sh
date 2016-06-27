#!/bin/bash
hg revert ClientServer.go
sudo -E go install ../
dovid -file=ClientServer.go -v
#dovid -file=client-nodoc.go -c
#dovid -file=client-nodoc.go -v
gofmt -w=true ClientServer.go
