# RUNBOOK — `bugs_closed/423` (closed 2026-09-02; these commands stay useful for the class)

## The two-way census (the one that discriminates)

Grader 1 of the bug file kills "the footer contains multi-byte characters" (32/32 stored
footers do, all fine). The cause must **CUT at a byte offset**. This pair is the test —
run **both** halves; either alone proves nothing.

```sql
-- LEFT: sites whose services-column labels contain a word whose FIRST rune is multi-byte.
-- Mirrors buildServicesHTML's own predicate INCLUDING its LIMIT 6, which is load-bearing:
-- a 7th offending page is invisible to the bug.
WITH cand AS (
  SELECT s.domain, p.name, COALESCE(p.nav_label,p.title,p.name) AS label,
         row_number() OVER (PARTITION BY p.site_id ORDER BY COALESCE(p.nav_order,99), p.name) AS rn
    FROM pages p JOIN sites s ON s.id=p.site_id
   WHERE p.status IN ('deployed','active')
     AND p.name NOT IN ('index','about','contact','privacy','terms','cookies','404',
                        'sitemap','faq','careers','insights','blog','news')
     AND p.name <> 'services' AND (p.in_header OR p.in_footer))
SELECT domain, name, label FROM cand
 WHERE rn <= 6 AND regexp_replace(label,'-',' ','g') ~ '(^|[[:space:]])[^[:ascii:]]'
 ORDER BY domain;
```

```sql
-- RIGHT: sites whose footer is not stored. Must return the SAME set.
SELECT s.domain, sc.build_status, sc.updated_at::date,
       (sc.rendered_html_digest = md5(sc.rendered_html)) AS digest_ok
  FROM site_components sc JOIN sites s ON s.id=sc.site_id
 WHERE sc.slot_name='footer'
 ORDER BY (sc.build_status<>'rendered') DESC, sc.updated_at DESC LIMIT 12;
```

⚠ **`digest_ok` NULL means `rendered_html` IS NULL — never stored — not "stale".** That is
the distinction between the two casualties and it decides whether the corrected disposition
degrades the build or FAILS it.

## Prove a byte-slice claim by EXECUTION, not by reading

Four minutes, and it beat an hour of reasoning about regex offsets:

```go
for _, w := range []string{"—dash", "“quote", "École", "…"} {
    out := strings.ToUpper(w[:1]) + w[1:]
    fmt.Printf("%-8q -> % x  valid=%v\n", w, out, utf8.ValidString(out))
}
// "—dash" -> ef bf bd 80 94 64 61 73 68  valid=false   <- 0x80, the live error's byte
```

## The class census (both directions — see the LANDMINE)

```bash
# must be 0 real call sites; returns 1 on a CLEAN tree — UpperFirst's own doc comment.
grep -rn "ToUpper(\w*\[:1\])" --include="*.go" platform/ internal/ pkg/ cmd/ | grep -v _test.go
# the POSITIVE control, which a comment cannot satisfy: must be 8.
grep -rn "UpperFirst(" --include="*.go" platform/ internal/ pkg/ cmd/ | grep -v _test.go | grep -v "^.*://"
```

## Before `gofmt -w` on any shared file

```bash
git show HEAD:<path> > /tmp/x.go && gofmt -l /tmp/x.go   # non-empty ⇒ HEAD is already
                                                          # unformatted; not yours to fix
```

## Mutation-proving the tests (all three were run, all three went red)

```bash
cp <file> /tmp/.bak; trap 'cp /tmp/.bak <file>' EXIT   # revert in the SAME shell call,
                                                        # so a shared tree is never left mutated
# 1. UpperFirst -> the byte idiom          => 2 datahelpers tests fail
# 2. SafeCut(summary,247) -> summary[:247] => the sqlmock expectation goes unmet
# 3. InvalidUTF8At -> always found=false   => its own test fails
```

## Verify after the roll (nothing here is true until an image ships)

```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
```
Then re-render both sites' chrome and check: `rendered.footer=true`, `build_status='rendered'`,
`rendered_html_digest = md5(rendered_html)`, the LEFT census unchanged and the RIGHT census
**empty**. Boxingonline's extra pre-delivery probe (from the filer): the served footer still
carries **no** contact block, because `sites.email` is empty and `component_library.go:1988`
gates it.

## When `go test` fails on code that is not yours (shared tree, several sessions)

`go test ./platform/orchestration/actions/` failed twice on OTHER sessions' in-flight
refactors (`seed_content_sources_action.go`, then `apply_theme_kit_action.go`). **A red
package build here is not evidence about your own change.** Do not "fix" their file and do
not wait — pin every peer-dirty file in the package to HEAD with a Go overlay, which
touches nothing on disk:

```bash
python3 - <<'PY'
import json, io, os, subprocess
root=os.getcwd(); sp="<scratch>"; mine={"<your dirty/untracked files>"}
rep={}
for line in subprocess.run(["git","status","--porcelain","--","<pkg dir>/"],
                           capture_output=True,text=True).stdout.splitlines():
    st, path = line[:2], line[3:].strip()
    if not path.endswith(".go") or path in mine: continue
    if st.strip()=="??":                      # untracked peer file: hide it
        rep[os.path.join(root,path)] = ""
    else:                                     # tracked peer WIP: pin to HEAD
        dst=os.path.join(sp,path.replace("/","__"))
        open(dst,"wb").write(subprocess.run(["git","show","HEAD:"+path],capture_output=True).stdout)
        rep[os.path.join(root,path)] = dst
io.open(os.path.join(sp,"overlay.json"),"w").write(json.dumps({"Replace":rep}))
PY
go test -overlay=<scratch>/overlay.json ./platform/orchestration/actions/
```

⚠ **List what it isolated and read the list.** It printed **28** peer files on
2026-09-02 — if your OWN file appears there you have just tested HEAD instead of your
change, and it will pass for the wrong reason. An untracked peer file maps to `""`
(hidden); a tracked one maps to a HEAD copy.

## Mutation-proving on a shared tree

Mutate and revert **inside one shell call**, with a `trap`, so the tree is never left
mutated across turns:

```bash
cp <file> /tmp/.bak; trap 'cp /tmp/.bak <file>' EXIT
```
Round 2's two mutations: deleting the escalation gate (returns round 1's behaviour) kills
`TestStoreRefusalDoesNotEscalateByDefault`; an arm that ignores its config key kills
`TestStoreRefusalEscalatesWhenArmed`. Both red, then green after revert.

## Deploy verification after the roll (debug_historian, council round 3)

Git and CI green are not proof, and a same-tag rebuild can ship a stale binary. Ask the
service what it is running:

```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor <the commit carrying this fix> <the stamp>
```

⚠ That line is emitted at STARTUP, so on a busy service it has already scrolled — an empty
result means "not in range", **not** "unstamped". Fall back to the binary probe, and run
**both** controls in the same breath:

```bash
kubectl -n ai-persona-system exec <pod> -- grep -aq "<sha that MUST be present>" /proc/1/exe
kubectl -n ai-persona-system exec <pod> -- grep -aq "<sha that MUST be absent>"  /proc/1/exe
```

Never `strings` (absent from these images, and its failure is indistinguishable from "not
stamped"), and never a discovery grep for "some 40-hex string" — that matches Go's internal
digit table and returns the same wrong answer on every service.
