package main

import (
	"fmt"
	"os"

	"github.com/exaroth/narrative/internal/debugger"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	log.SetLevel(log.WarnLevel)
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Pass path to file as first argument")
		return
	}

	debugger, err := debugger.InitWithFile(os.Args[1], 0)
	if err != nil {
		panic(err)
	}
	debugger.Run()
}
