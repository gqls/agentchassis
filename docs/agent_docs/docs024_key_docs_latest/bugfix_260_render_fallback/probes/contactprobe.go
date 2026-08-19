// contactprobe — would the contact-info seam DELETE the live section today?
// RenderTemplateWithMap (rerender_pages_actions.go:782) parses with NO FuncMap
// and NO missingkey=zero, then returns "" on any error — and its caller does
// contactInfoRe.ReplaceAllString(html, ""). So a failure erases the block.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"
)

type comp struct {
	Name     string `json:"name"`
	Function string `json:"function"`
	IsActive bool   `json:"is_active"`
	Template string `json:"html_template"`
}

func main() {
	raw, _ := os.ReadFile(os.Args[1])
	var comps []comp
	json.Unmarshal(raw, &comps)

	// EXACTLY what RenderTemplateWithMap does: no Funcs, no Option.
	render := func(t string, d map[string]interface{}) (string, error) {
		tmpl, err := template.New("component").Parse(t)
		if err != nil {
			return "", fmt.Errorf("parse: %w", err)
		}
		var b bytes.Buffer
		if err := tmpl.Execute(&b, d); err != nil {
			return "", fmt.Errorf("execute: %w", err)
		}
		return b.String(), nil
	}
	// CONTROL: a FuncMap name that the MAIN seam accepts must fail HERE.
	if _, e := render(`{{safe .x}}`, map[string]interface{}{}); e == nil {
		panic("control failed: {{safe}} parsed without a FuncMap — probe is not faithful")
	}
	// CONTROL: plain template must pass.
	if _, e := render(`<p>{{.email}}</p>`, map[string]interface{}{"email": "a@b.c"}); e != nil {
		panic("control failed: plain template errored: " + e.Error())
	}

	data := map[string]interface{}{
		"email": "info@example.com", "phone": "0100 000 0000",
		"phone_display": "0100 000 0000", "title": "Contact Information",
		"hours": "Monday – Friday, 9am – 6pm GMT",
	}
	n := 0
	for _, c := range comps {
		if c.Function != "contact-info" {
			continue
		}
		n++
		out, err := render(c.Template, data)
		status := "OK"
		if err != nil {
			status = "*** WOULD ERASE THE SECTION: " + err.Error()
		} else if strings.Contains(out, "<no value>") {
			status = "*** ships literal '<no value>' (no missingkey=zero on this path)"
		}
		fmt.Printf("%-28s active=%-5v len=%-6d %s\n", c.Name, c.IsActive, len(out), status)
	}
	fmt.Printf("\ncontact-info components examined: %d\n", n)
}
