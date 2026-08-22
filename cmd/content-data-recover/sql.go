// FILE: cmd/content-data-recover/sql.go
//
// Emits the UPDATE for one proven recovery. Every statement carries its own
// guards so that applying the generated file is safe even if the estate moved
// between the export and the apply:
//
//   - content_data IS NULL — refuses to overwrite data someone else has since
//     written. This is the important one: the whole premise is "this component
//     has none", and if that stopped being true the recovery is stale.
//   - md5(rendered_html) = <the exact bytes the recovery was proven against> —
//     the round-trip proof is a statement about THOSE bytes. If the component
//     has been re-rendered since, the proof does not transfer and the row must
//     be re-exported and re-proven.
//
// Both are WHERE-clause conditions rather than pre-checks, so a stale row
// updates 0 rows instead of being written wrongly, and the migration's own
// verify block counts what actually landed.
package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

func updateStatement(r Row, data map[string]interface{}) (string, error) {
	j, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	sum := md5.Sum([]byte(r.RenderedHTML))
	digest := hex.EncodeToString(sum[:])

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "-- %s  %s/%s  (%s)  fields: %s\n",
		r.Component, r.Domain, r.Page, r.PageComponentID, strings.Join(keys, ", "))
	fmt.Fprintf(&b, "UPDATE page_components SET content_data = %s::jsonb, updated_at = now()\n",
		quote(string(j)))
	fmt.Fprintf(&b, " WHERE id = '%s'\n", r.PageComponentID)
	fmt.Fprintf(&b, "   AND content_data IS NULL\n")
	fmt.Fprintf(&b, "   AND md5(rendered_html) = '%s';\n", digest)
	return b.String(), nil
}

// quote renders a Postgres string literal. Dollar-quoting is deliberate: the
// recovered values are page copy and routinely contain apostrophes, and a tag
// that appears in the body would break the literal — so it is checked, not
// assumed, and widened until it cannot collide.
func quote(s string) string {
	tag := "$cd$"
	for strings.Contains(s, tag) {
		tag = "$" + strings.TrimSuffix(strings.TrimPrefix(tag, "$"), "$") + "x$"
	}
	return tag + s + tag
}
