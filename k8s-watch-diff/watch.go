package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/kylelemons/godebug/diff"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <k8s-resource-url>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "EXAMPLE:\n")
		fmt.Fprintf(os.Stderr, "  %s http://localhost:8001/api/v1/watch/namespaces/default/services/kubernetes\n", os.Args[0])
		os.Exit(1)
	}
	url := os.Args[1]

	for {
		resp, err := http.Get(url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to GET %q: %v\n", url, err)
			os.Exit(1)
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Fprintf(os.Stderr, "HTTP GET returned %d\n", resp.Status)
			os.Exit(1)
		}

		dec := json.NewDecoder(resp.Body)

		var prev string
		var prevTime time.Time
		var prevLines int

		for dec.More() {
			m := struct {
				Type   string
				Object json.RawMessage
			}{}
			if err := dec.Decode(&m); err != nil {
				fmt.Fprintf(os.Stderr, "failed to decode: %v\n", err)
				os.Exit(1)
			}
			buf := bytes.Buffer{}
			if err := json.Indent(&buf, m.Object, "", "  "); err != nil {
				fmt.Fprintf(os.Stderr, "failed to pretty-print: %v\n", err)
				os.Exit(1)
			}
			next := buf.String()
			nextTime := time.Now()
			nextLines := strings.Count(next, "\n")
			patch := diff.Diff(prev, next)
			fmt.Printf("%v\n", m.Type)
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
}
