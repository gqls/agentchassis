package aiservice

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode"
)

// Fingerprint describes the SHAPE of a model response without reproducing its
// content, for logging a completion that arrived but could not be used.
//
// WHY THIS EXISTS. Diagnosing an unusable completion means answering a handful
// of questions — did it come wrapped in prose? in a markdown fence? did the
// model emit two JSON objects? was it empty? — and every one of those is
// STRUCTURAL. None needs the text. Logging the text to answer them exposes
// whatever the model echoed back, which on a debate or chat endpoint is the
// visitor's own words, to anyone who can read the container's logs.
//
// It is also strictly MORE diagnostic than logging a capped excerpt. The
// two-objects case (bugs_closed/088: "a complete JSON object, then commentary,
// then a second complete JSON object") puts the second object AFTER the first,
// typically ~1,500 chars in — so a 300-char excerpt shows only object one and
// can never reveal the defect it was justified by. Counting objects finds it
// regardless of length.
//
// Origin: council corr e004fd81 (bugs_open/083). The council approved the fix
// but recorded that whether to log model text "cannot be closed by this council
// alone"; the owner ruled for a structural fingerprint on 2026-07-27.
//
// Output is a single stable line, e.g.
//
//	chars=1834 first=L fence=yes objects=2 parses=false keys=[]
//	chars=0 first=none fence=no objects=0 parses=false keys=[]
//	chars=96 first={ fence=no objects=1 parses=true keys=[challenge,counter_position]
func Fingerprint(s string) string {
	first := "none"
	for _, r := range s {
		if !unicode.IsSpace(r) {
			first = string(r)
			break
		}
	}

	fence := "no"
	if strings.Contains(s, "```") {
		fence = "yes"
	}

	var parsed map[string]interface{}
	parses := json.Unmarshal([]byte(s), &parsed) == nil

	keys := make([]string, 0, len(parsed))
	if parses {
		for k := range parsed {
			// A key is a schema field name, not content — but cap it anyway, so
			// a model that ever returns prose AS a key cannot leak a sentence
			// through this field.
			if len(k) > maxFingerprintKey {
				k = k[:maxFingerprintKey] + "~"
			}
			keys = append(keys, k)
		}
		sort.Strings(keys)
		if len(keys) > maxFingerprintKeys {
			keys = append(keys[:maxFingerprintKeys], "…")
		}
	}

	return fmt.Sprintf("chars=%d first=%s fence=%s objects=%d parses=%t keys=[%s]",
		len(s), first, fence, TopLevelJSONObjects(s), parses, strings.Join(keys, ","))
}

const (
	maxFingerprintKey  = 40
	maxFingerprintKeys = 8
)

// TopLevelJSONObjects counts complete `{...}` objects at nesting depth zero.
//
// Braces inside string literals are ignored, and an escaped quote does not end
// a string — without both, any response whose prose contains a brace or an
// apostrophe-escaped quote miscounts, and the count is the whole point.
//
// Exported because "did the model emit more than one object?" is a question
// every JSON-returning caller has, and re-deriving this scanner per service is
// how two implementations drift apart.
func TopLevelJSONObjects(s string) int {
	depth, count := 0, 0
	inString, escaped := false, false

	for _, r := range s {
		if inString {
			switch {
			case escaped:
				escaped = false
			case r == '\\':
				escaped = true
			case r == '"':
				inString = false
			}
			continue
		}

		switch r {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			if depth > 0 {
				depth--
				if depth == 0 {
					count++
				}
			}
			// depth already 0 means a stray closing brace in prose: ignore it
			// rather than going negative, which would suppress a later real object.
		}
	}
	return count
}
