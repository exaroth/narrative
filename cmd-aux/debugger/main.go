package main

import (
	"fmt"
	"os"

	"github.com/exaroth/narrative/internal/debugger"
	log "github.com/sirupsen/logrus"
)

var debug string

func init() {
	debug = os.Getenv("DEBUG")
	if len(debug) > 0 {
		f, err := os.OpenFile("./narrative-debugger.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0655)
		if err != nil {
			panic(err)
		}
		log.SetOutput(f)
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.PanicLevel)
	}
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
	defer debugger.Deinit()
	debugger.Run()
}
