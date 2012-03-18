include ${GO_HOME}/go/src/Make.inc

TARG=goblog
GOFMT=gofmt

SRC=main.go\

GOFILES=${SRC}

include ${GO_HOME}/go/src/Make.cmd

format:
	${GOFMT} -w ${SRC}