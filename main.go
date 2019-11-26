package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/c-bata/go-prompt"
	"github.com/laetificat/gitlog/src/core"
)

var (
	suggestions []prompt.Suggest
	showMerges  bool
	showLocal   bool
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

	writer := bufio.NewWriter(os.Stdout)
	err = core.CompareBranches(baseBranch, compareBranch, writer, showMerges)
	if err != nil {
		fmt.Print(err) // log.Fatal(err) does not work here
		return
	}
	writer.Flush()
}

func completer(d prompt.Document) []prompt.Suggest {
	return prompt.FilterFuzzy(suggestions, d.GetWordBeforeCursor(), true)
}

func setupFlags() {
	flag.BoolVar(&showMerges, "merges", false, "include merge commits")
	flag.BoolVar(&showMerges, "m", false, "include merge commits(shorthand)")

	flag.BoolVar(&showLocal, "local", false, "include local branches")
	flag.BoolVar(&showLocal, "l", false, "include local branches(shorthand)")

	flag.Parse()
}
