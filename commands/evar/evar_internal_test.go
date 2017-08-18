package evar

import (
	"testing"
)

// TestEvarLoad tests the loading of environment variables from a file.
func TestEvarLoad(t *testing.T) {
	vars, _ := loadVars([]string{""}, testGetter{})
	evars := parseSplitEvars(vars)

	if len(evars) != 10 {
		t.Fatalf("Failed to parse all evars - %d - %q", len(evars), evars)
	}

	if evars["KEY4"] != "\nanother\nmultiline\n" {
		t.Fatalf("multiline var failed - %q", evars["KEY4"])
	}
	if evars["KEY5"] != "yes, even spaces and = are allowed as values" {
		t.Fatalf("Commas, spaces, = var failed - %q", evars["KEY5"])
	}
	if evars["KEY9"] != "you're welcome ;)" {
		t.Fatalf("Single quote, semicolon var failed - %q", evars["KEY9"])
	}
}

type testGetter struct{}

func (f testGetter) getContents(filename string) (string, error) {
	return testContents, nil
}

var testContents = `
key1=val
key2="val"
key3="this
is
a
multiline
value"
key4="
another
multiline
"
key5="yes, even spaces and = are allowed as values"
export key6=gasp
export key7="how is this guy doing these awesome things"


key8="yep, even whitespace is allowed (gets stripped)"
export key9="you're welcome ;)"
key10="x"
`
