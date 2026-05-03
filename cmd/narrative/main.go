package main

import (
	"os"

	"github.com/exaroth/narrative/internal/narrative"
	log "github.com/sirupsen/logrus"
)

func init() {
	debug := os.Getenv("DEBUG")
	if len(debug) > 0 {
		f, err := os.OpenFile("./narrative.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0655)
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

	log.Info("-------- Running ---------")
	ctrl, err := narrative.NewCtrl()
	if err != nil {
		panic(err)
	}
	defer ctrl.Deinit()
	err = ctrl.Run()
	if err != nil {
		panic(err)
	}
}
