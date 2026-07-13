// test_interactive_fingerprint.go
//
// Standalone reproduction of the goquery selectors in
// extract_interactive_fingerprint_action.go. Reads a single rawHtml file from
// disk and reports the same counts the production action would produce.
//
// Usage:
//   go run test_interactive_fingerprint.go ttk_rawhtml.html
//
// Save the rawHtml from the database as a .html file (no shell escaping needed),
// then run this in the same directory. Compare the output against what the
// production action reported.

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var eventAttrs = []string{
	"onclick", "onsubmit", "oninput", "onchange",
	"onload", "onmouseover", "onkeydown", "onkeyup",
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go run test_interactive_fingerprint.go <html_file>")
		os.Exit(1)
	}

	raw, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("ERROR reading file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("File: %s\n", os.Args[1])
	fmt.Printf("Size: %d bytes\n\n", len(raw))

	// Sanity check: does the raw text contain the substrings we care about?
	rawStr := string(raw)
	fmt.Println("=== Substring presence in raw bytes ===")
	fmt.Printf("  contains '<input '       : %v\n", strings.Contains(rawStr, "<input "))
	fmt.Printf("  contains '<select'       : %v\n", strings.Contains(rawStr, "<select"))
	fmt.Printf("  contains '<textarea'     : %v\n", strings.Contains(rawStr, "<textarea"))
	fmt.Printf("  contains '<canvas'       : %v\n", strings.Contains(rawStr, "<canvas"))
	fmt.Printf("  contains '<form'         : %v\n", strings.Contains(rawStr, "<form"))
	fmt.Printf("  contains '<script '      : %v\n", strings.Contains(rawStr, "<script "))
	fmt.Printf("  contains '<script>'      : %v\n", strings.Contains(rawStr, "<script>"))
	fmt.Printf("  contains 'onclick='      : %v\n", strings.Contains(rawStr, "onclick="))
	fmt.Printf("  contains 'addEventListener' : %v\n", strings.Contains(rawStr, "addEventListener"))
	fmt.Println()

	// Parse with goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(rawStr))
	if err != nil {
		fmt.Printf("ERROR parsing HTML: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== goquery selector counts ===")

	// Canvas
	canvasCount := doc.Find("canvas").Length()
	fmt.Printf("  canvas              : %d\n", canvasCount)

	// Forms
	formCount := doc.Find("form").Length()
	fmt.Printf("  form                : %d\n", formCount)

	// Inputs — the suspect selector
	inputComboCount := doc.Find("input, select, textarea").Length()
	fmt.Printf("  input,select,textarea (combo): %d\n", inputComboCount)

	// Now try them individually to see if the combo is the problem
	inputCount := doc.Find("input").Length()
	selectCount := doc.Find("select").Length()
	textareaCount := doc.Find("textarea").Length()
	fmt.Printf("  input  (alone)      : %d\n", inputCount)
	fmt.Printf("  select (alone)      : %d\n", selectCount)
	fmt.Printf("  textarea (alone)    : %d\n", textareaCount)
	fmt.Printf("  sum (input+select+textarea): %d\n", inputCount+selectCount+textareaCount)

	// Scripts (both inline and external)
	scriptCount := 0
	scriptSrcCount := 0
	scriptInlineCount := 0
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		scriptCount++
		if src, exists := s.Attr("src"); exists && src != "" {
			scriptSrcCount++
			_ = src
		} else if strings.TrimSpace(s.Text()) != "" {
			scriptInlineCount++
		}
	})
	fmt.Printf("  script              : %d (external src: %d, inline with content: %d)\n",
		scriptCount, scriptSrcCount, scriptInlineCount)

	// Event handler attributes — exactly as the action does it
	eventHandlerCount := 0
	for _, attr := range eventAttrs {
		n := doc.Find("[" + attr + "]").Length()
		eventHandlerCount += n
		if n > 0 {
			fmt.Printf("    [%s]: %d\n", attr, n)
		}
	}
	fmt.Printf("  event handlers total: %d\n", eventHandlerCount)

	fmt.Println()
	fmt.Println("=== Per-element inspection (first 3 inputs) ===")
	doc.Find("input").Each(func(i int, s *goquery.Selection) {
		if i >= 3 {
			return
		}
		id, _ := s.Attr("id")
		typ, _ := s.Attr("type")
		val, _ := s.Attr("value")
		fmt.Printf("  input[%d]: type=%q id=%q value=%q\n", i, typ, id, val)
	})

	// Final document outline to confirm the tree was parsed properly
	fmt.Println()
	fmt.Println("=== Document structure check ===")
	fmt.Printf("  <html> elements : %d\n", doc.Find("html").Length())
	fmt.Printf("  <head> elements : %d\n", doc.Find("head").Length())
	fmt.Printf("  <body> elements : %d\n", doc.Find("body").Length())
	fmt.Printf("  <main> elements : %d\n", doc.Find("main").Length())
	fmt.Printf("  <div> elements  : %d\n", doc.Find("div").Length())
	fmt.Printf("  <label> elements: %d\n", doc.Find("label").Length())
	fmt.Printf("  <button> elements: %d\n", doc.Find("button").Length())
	fmt.Printf("  <h1>+<h2>+<h3>+<h4> elements: %d\n",
		doc.Find("h1, h2, h3, h4").Length())
}
