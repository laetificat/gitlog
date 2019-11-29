package core

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/c-bata/go-prompt"
)

/*
RunGitBranch runs the git branch command and optionally adds the "-r" if specfied.
Returns a byte slice of data returned from git branch.
*/
func RunGitBranch(remoteOnly bool, directory string) ([]byte, error) {
	var out bytes.Buffer
	var errout bytes.Buffer

	command := exec.Command("git", "branch", "-a")
	if remoteOnly {
		command = exec.Command("git", "branch", "-r")
	}

	command.Stdout = &out
	command.Stderr = &errout
	command.Dir = directory

	err := command.Run()
	if errout.Len() > 0 {
		err = fmt.Errorf("%s", errout.String())
	}

	return out.Bytes(), err
}

/*
CreateBranchSuggestionsFromByteSlice returns a slice of prompt.Suggest structs
created from a slice of branch names.
*/
func CreateBranchSuggestionsFromByteSlice(branchBytes []byte) []prompt.Suggest {
	var branchList []prompt.Suggest

	branchNames := strings.Split(string(branchBytes), "\n")
	for _, name := range branchNames {
		cleanedName := cleanName(name)

		if len(cleanedName) > 0 {
			branchList = append(branchList, prompt.Suggest{Text: cleanedName, Description: ""})
		}
	}

	return branchList
}

/*
CompareBranches runs the git log baseBranch..compareBranch --oneline command to
compare two branches and writes the outcome to the given writer.
*/
func CompareBranches(
	baseBranch,
	compareBranch string,
	writer *bufio.Writer,
	showMerges bool,
	directory string,
	format string,
) error {
	branches := fmt.Sprintf("%s..%s", baseBranch, compareBranch)
	formatFlag := fmt.Sprintf("--format=%s", format)

	command := exec.Command("git", "--no-pager", "log", formatFlag, branches, "--no-merges")
	if showMerges {
		command = exec.Command("git", "--no-pager", "log", formatFlag, branches)
	}

	command.Dir = directory
	command.Stdout = writer
	errbuff := bytes.Buffer{}
	command.Stderr = &errbuff

	err := command.Run()
	if err != nil {
		return fmt.Errorf("%s", errbuff.String())
	}

	writer.Flush()

	return nil
}

/*
cleanName removes all the artifacts from the name we don't want
*/
func cleanName(name string) string {
	name = strings.Replace(name, "*", "", 1)
	name = strings.Replace(name, "origin/HEAD ->", "", 1)
	name = strings.TrimSpace(name)

	return name
}
