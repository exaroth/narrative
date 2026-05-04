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

func printError(msg string) {
	fmt.Printf("%s%s\n",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#FC0303")).Render("Error: "),
		msg,
	)

}

func main() {

	log.Info("-------- Running ---------")
	ctrl, err := narrative.NewCtrl()
	if err != nil {
		printError(err.Error())
		os.Exit(1)
	}
	defer ctrl.Deinit()
	message, err := ctrl.Run()
	if len(message) > 0 {
		fmt.Println(message)
		os.Exit(0)
	}
	if err != nil {
		if len(debug) > 0 {
			panic(err)
		} else {
			printError(err.Error())
			os.Exit(1)
		}
	}
}
