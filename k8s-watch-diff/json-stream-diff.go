// Example:
//  $ kubectl get -w -o json svc hostnames | json-stream-diff | colordiff
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kylelemons/godebug/diff"
)

func main() {
	dec := json.NewDecoder(os.Stdin)

	var prev string
	var prevTime time.Time
	var prevLines int

	for dec.More() {
		m := json.RawMessage{}
		if err := dec.Decode(&m); err != nil {
			fmt.Fprintf(os.Stderr, "failed to decode: %v\n", err)
			os.Exit(1)
		}
		buf := bytes.Buffer{}
		if err := json.Indent(&buf, m, "", "  "); err != nil {
			fmt.Fprintf(os.Stderr, "failed to pretty-print: %v\n", err)
			os.Exit(1)
		}
		next := buf.String()
		nextTime := time.Now()
		nextLines := strings.Count(next, "\n")
		patch := diff.Diff(prev, next)
		fmt.Printf("--- old %v\n", prevTime)
		fmt.Printf("+++ new %v\n", nextTime)
		fmt.Printf("@@ 0,%d +0,%d @@\n", prevLines, nextLines)
		fmt.Println(patch)
		fmt.Println()
		prev = next
		prevTime = nextTime
		prevLines = nextLines
	}
}
