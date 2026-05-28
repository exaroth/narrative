package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/exaroth/narrative/pkg/phonemizer"
)

func main() {

	if len(os.Args) < 2 {
		panic(errors.New("Pass word to check as first argument"))
	}

	word := os.Args[1]

	repo := phonemizer.NewPhonemizerRepository("")

	if err := repo.LoadLanguage(); err != nil {
		panic(err)
	}

	repo_result := repo.LookupWords(word)

	var phonemes map[string]uint32
	if len(repo_result) > 0 {
		if len(repo_result) > 1 {
			fmt.Printf("Repository returned more that one result for word %s: %v", word, repo_result)
		}
		phonemes = repo_result[0]
	} else {
		fmt.Printf("Word %s not found\n", word)
		return
	}
	fmt.Println("Results for word ", word)
	for phoneme, k := range phonemes {
		if k == 0 {
			continue
		}
		fmt.Println(" - ", phoneme)
		var tags = repo.LookupTags(word, phoneme)
		fmt.Printf("   Tags:  %s\n", strings.Join(tags, ", "))
	}
}
