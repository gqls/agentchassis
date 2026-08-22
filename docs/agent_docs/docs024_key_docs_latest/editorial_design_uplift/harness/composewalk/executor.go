// Byte-faithful replica of executeGoTemplate
// (platform/orchestration/actions/call_agent.go:1170-1221, read 2026-08-22),
// with exactly two deliberate divergences, both stated:
//   - the *zap.Logger parameter is dropped (unused by the original's logic);
//   - the "<no value>" strip is applied here after execution, as the
//     production render paths and the news lane's render.go harness do.
// DO NOT add functions to the funcmap in this file. The P0 claim is that
// composition needs none.
package main

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

func executeGoTemplate(templateStr string, data map[string]interface{}) (string, error) {
	tmpl, err := template.New("component").
		Option("missingkey=zero"). // Still useful for some types
		Funcs(template.FuncMap{
			"default": func(defaultVal, val interface{}) interface{} {
				if val == nil || val == "" {
					return defaultVal
				}
				return val
			},
			"eq": func(a, b interface{}) bool {
				return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
			},
			"ne": func(a, b interface{}) bool {
				return fmt.Sprintf("%v", a) != fmt.Sprintf("%v", b)
			},
			"lower": strings.ToLower,
			"upper": strings.ToUpper,
			"isset": func(val interface{}) bool {
				if val == nil {
					return false
				}
				if s, ok := val.(string); ok {
					return s != ""
				}
				return true
			},
			// safe returns "" for nil values instead of "<no value>"
			"safe": func(val interface{}) string {
				if val == nil {
					return ""
				}
				s := fmt.Sprintf("%v", val)
				if s == "<nil>" || s == "<no value>" {
					return ""
				}
				return s
			},
		}).Parse(templateStr)

	if err != nil {
		return "", fmt.Errorf("template parse: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execute: %w", err)
	}

	return strings.ReplaceAll(buf.String(), "<no value>", ""), nil
}
