package core

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var (
	TestDir       string = filepath.Join(os.TempDir(), "/gitlog-test")
	TestBranch1   string = "test-branch-1"
	TestBranch2   string = "feature/test-branch-2"
	TestFilename  string = "testfile.txt"
	TestLogFormat string = "%h %s (%cn <%ce>)"
)

func TestMain(m *testing.M) {
	defer tearDown()

	err := setupTestGitRepo()
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}

func tearDown() {
	_ = os.RemoveAll(TestDir)
}

func setupTestGitRepo() error {
	_ = os.RemoveAll(TestDir)
	err := os.MkdirAll(TestDir, 0777)
	if err != nil {
		return err
	}

	_, err = runCommand("git", "init")
	if err != nil {
		return err
	}

	_, err = runCommand("touch", TestFilename)
	if err != nil {
		return err
	}

	_, err = commitAll("test")
	if err != nil {
		return err
	}

	_, err = runCommand("git", "checkout", "-b", TestBranch1)
	if err != nil {
		return err
	}

	_, err = runCommand("git", "checkout", "-b", TestBranch2)
	if err != nil {
		return err
	}

	return nil
}

func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	stdoutBuff := bytes.Buffer{}
	stderrBuff := bytes.Buffer{}
	cmd.Stdout = &stdoutBuff
	cmd.Stderr = &stderrBuff
	cmd.Dir = TestDir

	_ = cmd.Run()

	errBuffString := stderrBuff.String()
	if strings.Contains(errBuffString, fmt.Sprintf("Switched to a new branch '%s'", TestBranch1)) {
		return "", nil
	}

	if strings.Contains(errBuffString, fmt.Sprintf("Switched to a new branch '%s'", TestBranch2)) {
		return "", nil
	}

	if strings.Contains(errBuffString, fmt.Sprintf("Switched to branch '%s'", TestBranch1)) {
		return "", nil
	}

	if strings.Contains(errBuffString, fmt.Sprintf("Switched to branch '%s'", TestBranch2)) {
		return "", nil
	}

	if strings.Contains(errBuffString, fmt.Sprintf("Already on '%s'", TestBranch1)) {
		return "", nil
	}

	if strings.Contains(errBuffString, fmt.Sprintf("Already on '%s'", TestBranch2)) {
		return "", nil
	}

	if len(errBuffString) > 0 {
		return "", fmt.Errorf("%s", errBuffString)
	}

	return stdoutBuff.String(), nil
}

func commitAll(message string) (string, error) {
	_, err := runCommand("git", "add", ".")
	if err != nil {
		return "", err
	}

	output, err := runCommand("git", "commit", "-m", message)
	if err != nil {
		return "", err
	}

	return output, nil
}

func Test_RunGitBranchLocalOnly(t *testing.T) {
	expectedBranches := []string{"master", TestBranch1, TestBranch2}

	branchBytes, err := RunGitBranch(false, TestDir)
	if err != nil {
		t.Error(err)
	}

	branchNames := strings.Split(strings.TrimSpace(string(branchBytes)), "\n")

	if len(branchNames) != len(expectedBranches) {
		t.Errorf("\nExpected: %s, \nactual: %s\n", string(len(expectedBranches)), string(len(branchNames)))
	}
}

func Test_RunGitBranchRemoteOnly(t *testing.T) {
	branchBytes, err := RunGitBranch(true, TestDir)
	if err != nil {
		t.Error(err)
	}

	if len(branchBytes) > 0 {
		t.Errorf("\nExpected: %s, \nactual: %s\n", "0", string(len(branchBytes)))
	}
}

func Test_RunGitBranchNonExistingDir(t *testing.T) {
	nonExistentDir := "non-existent-dir"
	expected := "no such file or directory"
	_ = os.RemoveAll(filepath.Join(TestDir, nonExistentDir))

	_, err := RunGitBranch(true, filepath.Join(TestDir, nonExistentDir))
	if err == nil {
		t.Errorf("\nExpected: %s, \nactual: %s\n", expected, err)
	}

	if !strings.Contains(err.Error(), expected) {
		t.Errorf("\nExpected: %s, \nactual: %s\n", expected, err)
	}
}

func Test_RunGitBranchNonGitDir(t *testing.T) {
	noRepoDir := filepath.Join(os.TempDir(), "gitlog-test-no-repo")
	expected := "fatal: not a git repository (or any of the parent directories): .git"
	_ = os.MkdirAll(noRepoDir, 0777)

	_, err := RunGitBranch(true, noRepoDir)
	if err == nil {
		t.Errorf("\nExpected: %s, \nactual: %s\n", expected, err)
	}

	if !strings.Contains(err.Error(), expected) {
		t.Errorf("\nExpected: %s, \nactual: %s\n", expected, err)
	}
}

func Test_CreateBranchSuggestionsFromByteSlice(t *testing.T) {
	expectedBranches := []string{"master", TestBranch1, TestBranch2}

	branchBytes, err := RunGitBranch(false, TestDir)
	if err != nil {
		t.Error(err)
	}

	suggestions := CreateBranchSuggestionsFromByteSlice(branchBytes)

	if len(suggestions) != len(expectedBranches) {
		t.Errorf("\nExpected: %s, \nactual: %s\n", string(len(expectedBranches)), string(len(suggestions)))
	}
}

func Test_CompareBranches(t *testing.T) {
	_, err := runCommand("git", "checkout", TestBranch1)
	if err != nil {
		t.Error(err)
	}

	err = ioutil.WriteFile(filepath.Join(TestDir, TestFilename), []byte("testing content\n"), 0777)
	if err != nil {
		t.Error(err)
	}

	_, err = commitAll("Added content")
	if err != nil {
		t.Error(err)
	}

	buffer := bytes.Buffer{}
	writer := bufio.NewWriter(&buffer)

	err = CompareBranches(TestBranch2, TestBranch1, writer, true, TestDir, TestLogFormat)
	if err != nil {
		t.Error(err)
	}

	expected := "Added content"
	actual := buffer.String()

	if !strings.Contains(buffer.String(), expected) {
		t.Errorf("\nExpected: %s, \nactual: %s\n", expected, actual)
	}

}

func Test_CompareBranchesError(t *testing.T) {
	_, err := runCommand("git", "checkout", TestBranch1)
	if err != nil {
		t.Error(err)
	}

	err = ioutil.WriteFile(filepath.Join(TestDir, TestFilename), []byte("testing content\n"), 0777)
	if err != nil {
		t.Error(err)
	}

	_, err = commitAll("Added content")
	if err != nil {
		t.Error(err)
	}

	buffer := bytes.Buffer{}
	writer := bufio.NewWriter(&buffer)

	err = CompareBranches(TestBranch2, TestBranch1, writer, true, TestDir+"/null", TestLogFormat)
	if err == nil {
		t.Errorf("\nExpected: %s, \nactual: %s\n", "", err)
	}
}
