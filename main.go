package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os/exec"

	"github.com/atotto/clipboard"
	"github.com/c-bata/go-prompt"
	"github.com/laetificat/gitlog/src/core"
)

var (
	suggestions     []prompt.Suggest
	showMerges      bool
	showLocal       bool
	format          string
	copyToClipboard bool
)

func main() {
	setupFlags()

	_, err := exec.LookPath("git")
	if err != nil {
		log.Fatal(err)
	}

	branchBytes, err := core.RunGitBranch(!showLocal, "")
	if err != nil {
		log.Fatal(err)
	}

	suggestions = core.CreateBranchSuggestionsFromByteSlice(branchBytes)

	fmt.Println("Select base branch")
	baseBranch := prompt.Input("> ", completer)

	fmt.Println("Select compare branch")
	compareBranch := prompt.Input("> ", completer)

	buffer := bytes.Buffer{}
	writer := bufio.NewWriter(&buffer)

	err = core.CompareBranches(baseBranch, compareBranch, writer, showMerges, ".", format)
	if err != nil {
		fmt.Print(err) // log.Fatal(err) does not work here
		return
	}

	bufferString := buffer.String()
	fmt.Print(bufferString)

	if copyToClipboard {
		err = clipboard.WriteAll(bufferString)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Print("\nCopied to clipboard!\n")
	}
}

func completer(d prompt.Document) []prompt.Suggest {
	return prompt.FilterFuzzy(suggestions, d.GetWordBeforeCursor(), true)
}

func setupFlags() {
	flag.BoolVar(&showMerges, "merges", false, "include merge commits")
	flag.BoolVar(&showMerges, "m", false, "include merge commits(shorthand)")

	flag.BoolVar(&showLocal, "local", false, "include local branches")
	flag.BoolVar(&showLocal, "l", false, "include local branches(shorthand)")

	flag.StringVar(&format, "format", "%h %s (%cn <%ce>)", "format of the git log output")
	flag.StringVar(&format, "f", "%h %s (%cn <%ce>)", "format of the git log output(shorthand)")

	flag.BoolVar(&copyToClipboard, "copy", true, "copy the result to the clipboard")
	flag.BoolVar(&copyToClipboard, "c", true, "copy the result to the clipboard(shorthand)")

	flag.Parse()
}
