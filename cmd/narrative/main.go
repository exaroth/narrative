package main

import (
	"fmt"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/exaroth/narrative/internal/narrative"
	log "github.com/sirupsen/logrus"
)

var debug string

func init() {
	debug = os.Getenv("DEBUG")
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
		if len(debug) > 0 {
			panic(err)
		} else {
			fmt.Printf("%s%s\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("#FC0303")).Render("Error: "),
				err.Error(),
			)
			os.Exit(1)
		}
	}
}
