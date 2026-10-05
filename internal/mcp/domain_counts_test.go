package mcp

import (
	"sort"
	"testing"
)

func TestDomainCounts(t *testing.T) {
	all, err := LoadEndpoints()
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, e := range all {
		counts[endpointDomain(e)]++
	}
	keys := []string{}
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		t.Logf("%s: %d", k, counts[k])
	}
	core, err := SelectEndpoints(all, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("default: %d; all: %d", len(core), len(all))
}
