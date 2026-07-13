package main

import (
	"bufio"
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
    literal_remove -f <pattern_file> <file>
    cat file | literal_remove - <pattern>

Examples:
    literal_remove mytemp.txt 'careers.html\">Careers | Leo'
    literal_remove mytemp.txt -f patterns.txt
    literal_remove -v mylog.txt '{{.email}}'
*/

func printUsage() {
	fmt.Fprintf(os.Stderr, `Literal String Remover - No regex interpretation

Usage:
    %s [options] <file> [pattern]
    %s <file> [options] [pattern]
    cat file | %s [options] - [pattern]

Arguments:
    file      Input file (use - for stdin)
    pattern   Literal string to remove (if not using -f)

Options:
    -f <file>   File containing patterns to remove (one per line)
    -o <file>   Output file (default: stdout)
    -i          Edit file in place
    -v          Show removal statistics
    -h          Show help

Examples:
    %s mylog.txt 'careers.html\">Careers | Leo'
    %s mylog.txt -f patterns.txt
    %s -f patterns.txt mylog.txt
    %s -v -i mylog.txt '<section class=\"hero\">'
    cat huge.log | %s - '{{.template}}' > cleaned.log

The pattern is matched LITERALLY - no escaping needed for:
    \ | < > [ ] { } @ . * $ ^ ( ) ? + or any other characters
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}

func main() {
	// Manual argument parsing to allow flags anywhere
	var (
		patternFile string
		outputFile  string
		inPlace     bool
		verbose     bool
		inputPath   string
		pattern     string
	)

	// Collect non-flag arguments
	var positionalArgs []string

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "-h", "--help":
			printUsage()
			os.Exit(0)
		case "-v", "--verbose":
			verbose = true
		case "-i", "--in-place":
			inPlace = true
		case "-f":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "Error: -f requires a filename argument")
				os.Exit(1)
			}
			i++
			patternFile = args[i]
		case "-o":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "Error: -o requires a filename argument")
				os.Exit(1)
			}
			i++
			outputFile = args[i]
		default:
			// Check for -f=file or -o=file style
			if strings.HasPrefix(arg, "-f=") {
				patternFile = arg[3:]
			} else if strings.HasPrefix(arg, "-o=") {
				outputFile = arg[3:]
			} else if strings.HasPrefix(arg, "-") && arg != "-" {
				fmt.Fprintf(os.Stderr, "Error: unknown option: %s\n", arg)
				printUsage()
				os.Exit(1)
			} else {
				positionalArgs = append(positionalArgs, arg)
			}
		}
	}

	// Parse positional arguments
	if len(positionalArgs) < 1 {
		fmt.Fprintln(os.Stderr, "Error: input file required")
		printUsage()
		os.Exit(1)
	}

	inputPath = positionalArgs[0]
	if len(positionalArgs) >= 2 {
		pattern = positionalArgs[1]
	}

	// Collect patterns
	var patterns []string

	if patternFile != "" {
		// Read patterns from file
		pf, err := os.Open(patternFile)
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
	}

	// Add command-line pattern if provided
	if pattern != "" {
		patterns = append(patterns, pattern)
	}

	if len(patterns) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no patterns specified (use pattern argument or -f pattern_file)")
		printUsage()
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
	for _, p := range patterns {
		if p == "" {
			continue
		}

		count := strings.Count(result, p)
		totalMatches += count

		if verbose && count > 0 {
			displayPattern := p
			if len(displayPattern) > 60 {
				displayPattern = displayPattern[:57] + "..."
			}
			fmt.Fprintf(os.Stderr, "Removed %d occurrence(s) of: %q\n", count, displayPattern)
		}

		// This is the key: strings.ReplaceAll does LITERAL string replacement
		result = strings.ReplaceAll(result, p, "")
	}

	// Output
	var out io.Writer = os.Stdout

	if inPlace && inputPath != "-" {
		f, err := os.Create(inputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		out = f
	} else if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		out = f
	}

	fmt.Fprint(out, result)

	if verbose {
		bytesRemoved := originalLen - len(result)
		dest := "stdout"
		if inPlace {
			dest = inputPath + " (in-place)"
		} else if outputFile != "" {
			dest = outputFile
		}
		fmt.Fprintf(os.Stderr, "Total: %d matches, %d bytes removed, %d bytes written to %s\n",
			totalMatches, bytesRemoved, len(result), dest)
	}
}
