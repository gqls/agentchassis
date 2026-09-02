# RUNBOOK — bugs_open/423

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
