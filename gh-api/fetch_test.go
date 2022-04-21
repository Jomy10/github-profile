package githubApi

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// This test is provided as a matter of example
// It only tests compilation (and the presence of a GH_TOKEN file,
// which isn't actually essential)
func TestFetchRepos(t *testing.T) {
	tokenBytes, err := os.ReadFile("GH_TOKEN")
	if err != nil {
		panic(err)
	}
	FetchRepos("jomy10", string(tokenBytes))
}

func TestFetchLanguages(t *testing.T) {
	// Unauthtenticated reuest
	fetched := FetchLanguages("jomy10/pufferfish", "", "")
	expected := map[string]uint{"Rust": 37859, "Ruby": 9251, "JavaScript": 5072, "Shell": 4446}
	if !cmp.Equal(fetched, expected) {
		t.Fatalf("%s != %s", FormatMap(fetched), FormatMap(expected))
	}
}

func FormatMap[T comparable, V any](_map map[T]V) string {
	output := ""

	for idx, val := range _map {
		output += fmt.Sprint(idx) + ": " + fmt.Sprint(val)
	}

	return output
}

// This test is provided as a matter of example
// It only tests compilation
func TestFetchUsrLangs(t *testing.T) {
	tokenBytes, err := os.ReadFile("GH_TOKEN")
	if err != nil {
		panic(err)
	}
	t.Fatalf(FormatMap(FetchUserLanguages("jomy10", string(tokenBytes), false, false, []string{}, []string{}, []string{}, []string{})))
}
