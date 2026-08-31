// Command envdiff snapshots and compares process environments without writing
// raw environment values to the snapshot or comparison output.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

type snapshot struct {
	Algorithm string            `json:"algorithm"`
	Variables map[string]string `json:"variables"`
}

func main() {
	snapshotPath := flag.String("snapshot", "", "write the current environment snapshot")
	comparePath := flag.String("compare", "", "compare the current environment with a snapshot")
	flag.Parse()

	if (*snapshotPath == "") == (*comparePath == "") {
		fatalf("specify exactly one of -snapshot or -compare")
	}

	current := capture(os.Environ())
	if *snapshotPath != "" {
		writeSnapshot(*snapshotPath, current)
		fmt.Printf("snapshotted %d environment variables\n", len(current.Variables))
		return
	}

	baseline := readSnapshot(*comparePath)
	printDiff(baseline, current)
}

func capture(environ []string) snapshot {
	variables := make(map[string]string, len(environ))
	for _, entry := range environ {
		name, value, ok := strings.Cut(entry, "=")
		if !ok || name == "" {
			continue
		}
		sum := sha256.Sum256([]byte(value))
		variables[name] = hex.EncodeToString(sum[:])
	}
	return snapshot{Algorithm: "sha256", Variables: variables}
}

func writeSnapshot(path string, value snapshot) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		fatalf("encode snapshot: %v", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o600); err != nil {
		fatalf("write snapshot: %v", err)
	}
}

func readSnapshot(path string) snapshot {
	data, err := os.ReadFile(path)
	if err != nil {
		fatalf("read snapshot: %v", err)
	}
	var value snapshot
	if err := json.Unmarshal(data, &value); err != nil {
		fatalf("decode snapshot: %v", err)
	}
	if value.Algorithm != "sha256" || value.Variables == nil {
		fatalf("unsupported snapshot format")
	}
	return value
}

func printDiff(baseline, current snapshot) {
	var added, removed, changed []string
	for name, hash := range current.Variables {
		oldHash, ok := baseline.Variables[name]
		switch {
		case !ok:
			added = append(added, name)
		case oldHash != hash:
			changed = append(changed, name)
		}
	}
	for name := range baseline.Variables {
		if _, ok := current.Variables[name]; !ok {
			removed = append(removed, name)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	sort.Strings(changed)

	fmt.Printf("baseline=%d current=%d added=%d removed=%d changed=%d\n", len(baseline.Variables), len(current.Variables), len(added), len(removed), len(changed))
	printNames("added", added)
	printNames("removed", removed)
	printNames("changed", changed)
}

func printNames(label string, names []string) {
	for _, name := range names {
		fmt.Printf("%s %s\n", label, name)
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "envdiff: "+format+"\n", args...)
	os.Exit(1)
}
