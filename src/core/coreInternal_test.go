package core

import "testing"

func Test_cleanName(t *testing.T) {
	nastyName := " * feature/spaces-and-stars"
	expected := "feature/spaces-and-stars"
	actual := cleanName(nastyName)

	if cleanName(nastyName) != expected {
		t.Errorf("\nExpected: %s,\nActual: %s\n", expected, actual)
	}
}
