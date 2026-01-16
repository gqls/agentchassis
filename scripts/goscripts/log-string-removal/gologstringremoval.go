package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

/*
Literal string removal tool - no regex interpretation.
Handles any characters including: \ | < > [ ] { } @ . * $ ^ etc.

Usage:
    literal_remove <file> <pattern>
    literal_remove <file> -f <pattern_file>
    literal_remove -i <file> <pattern>          # in-place edit
    cat file | literal_remove - <pattern>

Examples:
    literal_remove mytemp.txt 'careers.html\">Careers | Leo'
    literal_remove mytemp.txt -f patterns.txt
    literal_remove -v mylog.txt '{{.email}}'
*/

func main() {
	// Flags
	patternFile := flag.String("f", "", "File containing patterns to remove (one per line)")
	outputFile := flag.String("o", "", "Output file (default: stdout)")
	inPlace := flag.Bool("i", false, "Edit file in place")
	verbose := flag.Bool("v", false, "Show removal statistics")
	help := flag.Bool("h", false, "Show help")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Literal String Remover - No regex interpretation

Usage:
    %s [options] <file> [pattern]
    cat file | %s [options] - [pattern]

Arguments:
    file      Input file (use - for stdin)
    pattern   Literal string to remove (if not using -f)

Options:
`, os.Args[0], os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
    %s mylog.txt 'careers.html\">Careers | Leo'
    %s mylog.txt -f patterns.txt
    %s -v -i mylog.txt '<section class=\"hero\">'
    cat huge.log | %s - '{{.template}}' > cleaned.log

The pattern is matched LITERALLY - no escaping needed for:
    \ | < > [ ] { } @ . * $ ^ ( ) ? + or any other characters
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0])
	}

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: input file required")
		flag.Usage()
		os.Exit(1)
	}

	inputPath := args[0]

	// Collect patterns
	var patterns []string

	if *patternFile != "" {
		// Read patterns from file
		pf, err := os.Open(*patternFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening pattern file: %v\n", err)
			os.Exit(1)
		}
		defer pf.Close()

		scanner := bufio.NewScanner(pf)
		// Increase buffer size for long patterns
		buf := make([]byte, 0, 1024*1024)
		scanner.Buffer(buf, 10*1024*1024)

		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				patterns = append(patterns, line)
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading pattern file: %v\n", err)
			os.Exit(1)
		}
	} else if len(args) >= 2 {
		patterns = append(patterns, args[1])
	} else {
		fmt.Fprintln(os.Stderr, "Error: either pattern argument or -f pattern_file required")
		flag.Usage()
		os.Exit(1)
	}

	// Read input
	var content []byte
	var err error

	if inputPath == "-" {
		content, err = io.ReadAll(os.Stdin)
	} else {
		content, err = os.ReadFile(inputPath)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	originalLen := len(content)
	result := string(content)
	totalMatches := 0

	// Remove each pattern - strings.ReplaceAll does LITERAL replacement
	for _, pattern := range patterns {
		if pattern == "" {
			continue
		}

		count := strings.Count(result, pattern)
		totalMatches += count

		if *verbose && count > 0 {
			displayPattern := pattern
			if len(displayPattern) > 60 {
				displayPattern = displayPattern[:57] + "..."
			}
			fmt.Fprintf(os.Stderr, "Removed %d occurrence(s) of: %q\n", count, displayPattern)
		}

		// This is the key: strings.ReplaceAll with -1 does LITERAL string replacement
		result = strings.ReplaceAll(result, pattern, "")
	}

	// Output
	var out io.Writer = os.Stdout

	if *inPlace && inputPath != "-" {
		f, err := os.Create(inputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		out = f
	} else if *outputFile != "" {
		f, err := os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		out = f
	}

	fmt.Fprint(out, result)

	if *verbose {
		bytesRemoved := originalLen - len(result)
		dest := "stdout"
		if *inPlace {
			dest = inputPath + " (in-place)"
		} else if *outputFile != "" {
			dest = *outputFile
		}
		fmt.Fprintf(os.Stderr, "Total: %d matches, %d bytes removed, %d bytes written to %s\n",
			totalMatches, bytesRemoved, len(result), dest)
	}
}
