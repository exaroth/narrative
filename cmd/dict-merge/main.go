package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/exaroth/narrative/internal/debugger"
)

func main() {

	if len(os.Args) < 3 {
		panic(errors.New("Pass target and source dict: merge-dict <target> <source>"))
	}

	target_d, err := debugger.LoadCsvDict(os.Args[1])
	if err != nil {
		panic(fmt.Errorf("Error loading target dict: %w", err))
	}
	source_d, err := debugger.LoadCsvDict(os.Args[2])
	if err != nil {
		panic(fmt.Errorf("Error loading source  dict: %w", err))
	}
	for word, phoneme := range source_d.Items() {
		target_d.Update(word, phoneme)
	}
	err = target_d.Save()
	if err != nil {
		panic(fmt.Errorf("Error saving target dict %w", err))
	}
	fmt.Printf("Dictionary at %s updated\n", os.Args[1])
}
