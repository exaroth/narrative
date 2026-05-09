package main

import (
	"errors"
	"fmt"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/alexflint/go-arg"
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
	args, err := narrative.ParseArgs()
	if err != nil {
		// Dont catch missing help as we
		// handle it internally.
		if !errors.Is(err, arg.ErrHelp) {
			printError(fmt.Sprintf("%s\n%s", err.Error(), usage()))
			os.Exit(1)
		} else {
			args.Help = true
		}
	}

	if args.Help {
		fmt.Println(usage())
		os.Exit(0)
	}
	if args.Version {
		fmt.Println(printVersion())
		os.Exit(0)
	}

	ctrl, err := narrative.NewCtrl(args)
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

func usage() string {
	return `Usage: narrative [OPTIONS...] TEXT_SOURCE
OPTIONS:
	--voice <voice_name>   Set voice for playback.
	--list-voices          List available voices.
	--serve <port>         Start server running at <port>.
	--list-models          List available KittenTTS model information.
	--add-model <name>     Download and select KittenTTS model.
	--select-model <name>  Switch currently used TTS model.
	--help                 Print help.`
}

func printVersion() string {
	return fmt.Sprintf("Narrative v%s\n", narrative.VERSION)
}

func printError(msg string) {
	fmt.Printf("%s%s\n",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#FC0303")).Render("Error: "),
		msg,
	)

}
