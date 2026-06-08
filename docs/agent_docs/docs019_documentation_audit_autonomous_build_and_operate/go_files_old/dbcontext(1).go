// Command dbcontext fetches database context for a bundle by shelling out to a
// configurable psql. It does two things:
//
//   - schema: runs `\d <table>` for each requested table (complete and bounded).
//   - rows:   runs a SELECT with multipass sizing — probe with LIMIT N+1, then
//             show all rows if within the cap, or a sample plus a pointer (the
//             query to run for the full result) if over. Never an unbounded dump.
//
// It connects the way you already do. Set -psql to a direct connection or your
// kubectl-exec pattern:
//
//	dbcontext -psql 'psql "postgresql://user:pass@host/clients_db"' -schema site_work_items,pages
//	dbcontext -psql 'kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db' \
//	          -rows "SELECT name, build_status FROM pages WHERE site_id='...'"
//
// Output is markdown. Feed schema to the assembler via -schema <file>, and row
// output via -doc <file>. No Go DB driver dependency — psql does the talking.
package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	psql := flag.String("psql", "psql", "psql invocation prefix (split on spaces; the query is appended as -c args, not via a shell)")
	schema := flag.String("schema", "", "comma-separated tables to describe with \\d")
	rows := flag.String("rows", "", "a SELECT to fetch with multipass sizing")
	maxRows := flag.Int("max-rows", 50, "row cap; above this, show a sample + the query as a pointer")
	runtimeSite := flag.String("runtime-site", "", "domain to pull runtime evidence for (recent errors + work-item lifecycle)")
	runtimePage := flag.String("runtime-page", "", "optional page name to narrow the work-item lifecycle")
	flag.Parse()

	base := strings.Fields(*psql)
	if len(base) == 0 {
		fmt.Fprintln(os.Stderr, "empty -psql")
		os.Exit(2)
	}
	if *schema == "" && *rows == "" && *runtimeSite == "" {
		fmt.Fprintln(os.Stderr, "nothing to do: pass -schema, -rows, and/or -runtime-site")
		os.Exit(2)
	}

	// run executes the configured psql with extra flags and a -c query, no shell.
	run := func(extraArgs []string, query string) (string, error) {
		args := append([]string{}, base[1:]...)
		args = append(args, extraArgs...)
		args = append(args, "-c", query)
		cmd := exec.Command(base[0], args...)
		var out, errb bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &errb
		if err := cmd.Run(); err != nil {
			return out.String(), fmt.Errorf("%v: %s", err, strings.TrimSpace(errb.String()))
		}
		return out.String(), nil
	}

	var b strings.Builder

	if *schema != "" {
		b.WriteString("# Database schema\n\n")
		for _, t := range strings.Split(*schema, ",") {
			t = strings.TrimSpace(t)
			if t == "" {
				continue
			}
			out, err := run(nil, `\d `+t)
			b.WriteString("### `" + t + "`\n\n```\n")
			if err != nil {
				b.WriteString("error: " + err.Error() + "\n")
			} else {
				b.WriteString(strings.TrimRight(out, "\n") + "\n")
			}
			b.WriteString("```\n\n")
		}
	}

	// fetchAndRender runs a SELECT with multipass sizing and returns markdown
	// (a count/sample note + a table). Reused by -rows and -runtime-site.
	fetchAndRender := func(query string) string {
		q := strings.TrimRight(strings.TrimSpace(query), ";")
		wrapped := fmt.Sprintf("SELECT * FROM ( %s ) _q LIMIT %d", q, *maxRows+1)
		out, err := run([]string{"--csv"}, wrapped)
		if err != nil {
			return "error: " + err.Error() + "\n\n"
		}
		recs := parseCSV(out)
		dataRows := len(recs) - 1
		if dataRows < 0 {
			return "_no output_\n\n"
		}
		if dataRows > *maxRows {
			recs = recs[:*maxRows+1]
			return fmt.Sprintf("_Result exceeds %d rows — showing the first %d as a sample. Run the query without the limit to see all rows._\n\n", *maxRows, *maxRows) + renderMarkdownTable(recs)
		}
		return fmt.Sprintf("_%d row(s)._\n\n", dataRows) + renderMarkdownTable(recs)
	}

	if *rows != "" {
		b.WriteString("# Live data\n\n")
		b.WriteString("Query: `" + strings.TrimSpace(*rows) + "`\n\n")
		b.WriteString(fetchAndRender(*rows))
	}

	if *runtimeSite != "" {
		// site_id resolved inline; single-quotes escaped for safety on apostrophes
		esc := func(s string) string { return strings.ReplaceAll(s, "'", "''") }
		sid := "(SELECT id FROM sites WHERE domain='" + esc(*runtimeSite) + "')"
		b.WriteString("# Runtime evidence\n\n")
		b.WriteString("Correlated by site `" + *runtimeSite + "`")
		if *runtimePage != "" {
			b.WriteString(", page `" + *runtimePage + "`")
		}
		b.WriteString(". Most recent first.\n\n")

		b.WriteString("## Recent errors (agent_error_log)\n\n")
		errQ := "SELECT occurred_at, agent_type, step_name, action, left(error_message,160) AS error, pod_name " +
			"FROM agent_error_log WHERE site_id=" + sid + " ORDER BY occurred_at DESC"
		b.WriteString(fetchAndRender(errQ))

		b.WriteString("## Work-item lifecycle (site_work_items)\n\n")
		itemQ := "SELECT created_at, item_type, status, claimed_by, attempt_count||'/'||max_attempts AS attempts, " +
			"left(coalesce(error,''),120) AS error, item_key FROM site_work_items WHERE site_id=" + sid
		if *runtimePage != "" {
			itemQ += " AND (spec->>'page_name')='" + esc(*runtimePage) + "'"
		}
		itemQ += " ORDER BY created_at DESC"
		b.WriteString(fetchAndRender(itemQ))
	}

	fmt.Print(b.String())
}

func parseCSV(s string) [][]string {
	r := csv.NewReader(strings.NewReader(s))
	r.FieldsPerRecord = -1 // tolerate ragged rows
	recs, err := r.ReadAll()
	if err != nil && len(recs) == 0 {
		return nil
	}
	return recs
}

// renderMarkdownTable turns CSV records (first row = header) into a markdown table.
func renderMarkdownTable(recs [][]string) string {
	if len(recs) == 0 {
		return "_no rows_\n\n"
	}
	clean := func(s string) string {
		s = strings.ReplaceAll(s, "\n", " ")
		s = strings.ReplaceAll(s, "|", "\\|")
		return s
	}
	var b strings.Builder
	header := recs[0]
	b.WriteString("| ")
	for _, h := range header {
		b.WriteString(clean(h) + " | ")
	}
	b.WriteString("\n|")
	for range header {
		b.WriteString("---|")
	}
	b.WriteString("\n")
	for _, row := range recs[1:] {
		b.WriteString("| ")
		for i := range header {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			b.WriteString(clean(cell) + " | ")
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
	return b.String()
}
