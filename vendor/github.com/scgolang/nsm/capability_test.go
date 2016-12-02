package nsm

import (
	"testing"
)

func TestParseCapabilities(t *testing.T) {
	for i, testcase := range []struct {
		input    string
		expected Capabilities
	}{
		{
			input: "switch:dirty:progress",
			expected: Capabilities{
				"switch",
				"dirty",
				"progress",
			},
		},
	} {
		if expected, got := testcase.expected, ParseCapabilities(testcase.input); !expected.Equal(got) {
			t.Fatalf("(testcase %d) expected %s, got %s", i, expected, got)
		}
	}
}

func TestCapabilitiesEqual(t *testing.T) {
	for i, testcase := range []struct {
		c  Capabilities
		e  []Capabilities
		ne []Capabilities
	}{
		{
			c: Capabilities{"switch", "dirty"},
			e: []Capabilities{
				Capabilities{"switch", "dirty"},
			},
			ne: []Capabilities{
				Capabilities{"switch", "progress"},
				Capabilities{"switch", "dirty", "progress"},
			},
		},
	} {
		c := testcase.c

		for _, e := range testcase.e {
			if !c.Equal(e) {
				t.Fatalf("(testcase %d) expected %s to equal %s", i, c, e)
			}
		}

		for _, ne := range testcase.ne {
			if c.Equal(ne) {
				t.Fatalf("(testcase %d) expected %s to not equal %s", i, c, ne)
			}
		}
	}
}

func TestCapabilitiesString(t *testing.T) {
	for i, testcase := range []struct {
		input    Capabilities
		expected string
	}{
		{
			input:    Capabilities{},
			expected: "",
		},
		{
			input:    Capabilities{CapClientSwitch, CapClientDirty, CapClientProgress},
			expected: ":switch:dirty:progress:",
		},
	} {
		if expected, got := testcase.expected, testcase.input.String(); expected != got {
			t.Fatalf("(testcase %d) expected %s, got %s", i, expected, got)
		}
	}
}
