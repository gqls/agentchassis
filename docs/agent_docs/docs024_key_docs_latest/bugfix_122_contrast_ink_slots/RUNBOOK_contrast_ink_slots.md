# RUNBOOK — bug 122 contrast / ink slots

Every command here was needed to get an answer right, with its gotcha attached.
Newest additions at the bottom of each section.

---

## Measure what a visitor actually sees

```bash
python3 scripts/render_audit.py https://site.example/ [more urls...]
```

- Headless Chromium, **every** element (`body *`), computed style, alpha-composited
  backdrop. This is the ONLY thing that answers the contrast question — see the
  gotchas below for the three cheaper methods that all give wrong answers.
- `(over an image — ratio approximate)` findings are NOT firm. The tool reports them
  rather than asserting them; discount them when counting.
- ~30–60s per page. Ten homepages is about 5 minutes, so background it.

**GOTCHA — the audit output goes to stdout and a long run's head scrolls out of a
tool result.** Redirect to a file in the scratchpad and grep it, or you get only the
last site's detail and a total:

```bash
python3 scripts/render_audit.py <urls> > audit_$(date +%F).txt 2>&1
grep -E '^(FAIL|ok)' audit_2026-08-06.txt      # per-site summary
```

## Fetch a served stylesheet

```bash
curl -fsS -A "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/126 Safari/537.36" \
  https://<domain>/assets/css/styles.css -o css_<domain>.css
```

**GOTCHA — a bare `curl`/`urllib` gets `403 Forbidden` on every site.** The origin
rejects a non-browser user agent. A whole 19-site census returned `403` on every row
and read exactly like a routing failure. Send a browser UA.

**GOTCHA — the path is `/assets/css/styles.css`, not `/styles.css`.** The short path
404s. Read it off the page rather than guessing:
`curl -fsS https://<domain>/ | grep -oE '<link[^>]*stylesheet[^>]*>'`

## The three cheaper measurements that are all WRONG

Recorded because each one looked authoritative and this bug's own history was
distorted by the first:

1. **A regex over `styles.css`** cannot resolve the cascade (which `a` rule wins),
   cannot see ancestors, alpha or gradients, and cannot know which variable a given
   element actually uses. This produced 122's superseded table, in which
   dartsonline "scored" 1.11 because the audit compared its background against
   itself.
2. **A palette row / `AuditPalette`** cannot see a literal that is in no palette —
   which is where the hardcoded inks live.
3. **Comparing a foreground against the PAGE background** instead of the element's
   own flags the header, logo and nav as failing when they sit on their own
   `--color-header-bg` and are fine at 17:1. "Fixing" those breaks what works.

## Find which artefact owns a failing CSS rule

The audit names a selector; this says which of the four surfaces defines it.

```sql
SELECT 'layouts' AS src, name FROM layouts WHERE css_template LIKE '%<selector>%'
UNION ALL SELECT 'css_snippets', name FROM css_snippets WHERE css_content LIKE '%<selector>%'
UNION ALL SELECT 'content_components', name FROM content_components
  WHERE html_template LIKE '%<selector>%' AND COALESCE(is_active,true);
```

Then pull the rule itself — `substring(... from '\.<selector>\s*\{[^}]*\}')`.

**GOTCHA — component CSS lives INSIDE `html_template`, not in a CSS column.**
`content_components` has no `css_styles` column (I assumed one and got
`ERROR: column "css_styles" does not exist`). `css_snippets.css_content` is a
different, much smaller surface: 21 rows, and **0 of them mention `--color-primary`
at all** — so a census run only against `css_snippets` will report zero and look
like a clean bill of health.

## Census: is a palette colour used as an ink or as a fill?

```sql
SELECT name,
       (SELECT count(*) FROM regexp_matches(css_template, '[^-]color:\s*var\(--color-primary[),]', 'g')) AS ink_uses
FROM layouts WHERE css_template ~ '[^-]color:\s*var\(--color-primary[),]' ORDER BY 2 DESC;
```

**GOTCHA — the `[^-]` prefix is load-bearing.** Without it, `background-color:` and
`border-color:` match as inks and every count is inflated.

## Is a proposed new CSS variable already taken? (do this BEFORE naming one)

```sql
SELECT 'content_components' AS src, count(*) FROM content_components WHERE html_template LIKE '%<--var-name>%' AND COALESCE(is_active,true)
UNION ALL SELECT 'layouts', count(*) FROM layouts WHERE css_template LIKE '%<--var-name>%'
UNION ALL SELECT 'css_snippets', count(*) FROM css_snippets WHERE css_content LIKE '%<--var-name>%'
UNION ALL SELECT 'site_components', count(*) FROM site_components WHERE rendered_html LIKE '%<--var-name>%'
UNION ALL SELECT 'page_components', count(*) FROM page_components WHERE rendered_html LIKE '%<--var-name>%';
```

All five surfaces, because a variable can be consumed by an already-rendered artefact
that no source template mentions. `--color-accent-text` returns 0/0/0/0/0 — it is
derived by the platform and consumed by nothing, which is the dead-config shape.

## Does a layout actually ASK for a palette slot? (the LANDMINE check)

```sql
SELECT count(*) FROM layouts WHERE css_template LIKE '%palette "<slot>"%';
```

A palette reaches the stylesheet **only** through `{{palette "X" "literal"}}` calls in
a layout template. A slot no layout names is never emitted, so adding it to
`darkSchemeDerivations` compiles, logs success, and changes nothing. Measured:
`primary_text` / `cta_text` / `header_text` / `footer_text` are declared by **18 of
18** layouts; `card_bg` 18; `surface_alt` 3; `icon_chip_bg` **0**.

## Prove the render-audit chain is live

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis \
  -o jsonpath='{.items[0].metadata.name}')
for s in write_render_audit_findings scanStoredStatClaims zzzInventedControlXyz; do
  printf "%-40s " "$s"
  kubectl -n ai-persona-system exec $POD -- sh -c "strings /app/agent-chassis | grep -c '$s'"
done
# v1.0.1257: 11 / 2 / 0  -- the 0 proves the grep discriminates; the 2 proves
# I am reading the binary I think I am.
```

Then the config half, which the binary cannot tell you:

```sql
SELECT jsonb_object_keys(default_config->'workflow'->'steps') FROM agent_definitions
WHERE type='render-audit-agent' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- site, audit, write_findings, complete, complete_error
```

And the cadence half, which is the one that is missing:

```sql
SELECT name, enabled, target_agent_type FROM scheduled_tasks WHERE enabled;
-- 28 rows, none targeting render-audit-agent
SELECT item_type, status, count(*) FROM site_work_items
WHERE item_type='contrast_failure' GROUP BY 1,2;   -- 4, all complete, all 2026-08-04
```

**GOTCHA — `SELECT ... FROM orchestration_states WHERE owner_agent_type=...`
returning 0 rows does NOT mean "never ran".** Terminal rows are reaped at ~24h. It
means "not in the last day". The cadence question is answered by `scheduled_tasks`,
which has no reaper.

## Schema, before any of the above

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c '\d <table>'
```

**GOTCHA — I skipped this twice in one session and paid for it twice**:
`site_components` has no `is_active` (the lifecycle columns are `locked_at` /
`lock_type` / `build_status`), and `content_components` has no `css_styles`.
`site_specs` has no `resolved_composition` either — it is a single `data` jsonb keyed
by `aspect`. CLAUDE.md says schema first; these are the tables it is worth obeying on,
because each query *reads* perfectly.

## Which DECLARATION chose the colour the audit measured? (do this before saying "hard-coded")

The render audit reports a *computed* colour. It cannot tell you which declaration
produced it, and guessing is how a whole sub-shape got mis-diagnosed (NOTES misstep 5).

```sql
SELECT substring(html_template from '\.<selector>\s*\{[^}]*\}')
FROM content_components WHERE name='<component>';
```

Then read which property names the palette colour:

- `background:` names it → the element needs **`--color-<x>-text`** (the ink that goes ON an x fill)
- `color:` names it → the element needs **`--color-<x>-ink`** (x made legible AS an ink)

**GOTCHA — `var(--color-primary-text, #fff)` rendering white does NOT mean the fallback
fired.** Check the served stylesheet before blaming the literal:

```bash
curl -fsS -A "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/126 Safari/537.36" \
  https://<domain>/assets/css/styles.css | grep -- '--color-primary-text:'
```

On finetuning.uk it IS defined, as `#ffffff`, and it is correct for its own slot. The
value is right and the slot is wrong — a grep for hard-coded whites finds nothing.

**GOTCHA — the selector may not be in a component at all.** gaswholesalers' six `.A`
failures are the *layout's* base `a { color: var(--color-accent) }`. Search `layouts`
too, and remember `.A` in the audit means a bare `<a>` with no class.

## Build and test when another session's WIP breaks the package

```bash
T=<scratch>/headtree; rm -rf $T && mkdir -p $T
git archive HEAD | tar -x -C $T
cp <your files> $T/platform/orchestration/actions/
cd $T && go test ./platform/orchestration/actions/ -run '<YourTests>'
```

**GOTCHA — `go build ./...` in the working tree tests everyone's uncommitted work, not
yours.** On 2026-08-06 the package would not compile because another session's
`diagnose_persist_fix_plan_action.go` was missing an import. The clean-HEAD tree is the
only way to know whose fault a failure is.

**And run `gofmt -l` before committing** — un-gofmt'd code is rejected by the build gate,
so it reaches CI as a failed gate and no PR. The pre-commit pattern check catches it,
advisory only, so it is easy to skim past.

## Prove a test assertion is load-bearing (mutate, expect a DISTINCT failure)

Mutate in the HEAD tree, run, restore. A mutation that PASSES may have hit a guard in
series; a mutation that fails the *wrong* test proves nothing about the one you meant.

```bash
cp <file> /tmp/x.bak
python3 - <<'PY'   # e.g. grounds -> grounds[:1]
...
PY
go test ./platform/orchestration/actions/ -run '<Tests>' 2>&1 | grep -E "FAIL:|ok "
cp /tmp/x.bak <file>
```

**GOTCHA — a contrast fixture must be SATISFIABLE.** Grounds `#101010` and `#E9E9E9`
admit no colour clearing AA against both (the darker demands relative luminance ≥ 0.200,
the lighter ≤ 0.140), so every candidate correctly falls through to the achromatic
fallback and your test fails while the code is right. Two grounds of similar lightness —
like dartsonline's real `#0E1019` / `#1A1F2E` — is what tests the CHOICE.

## psql: when a regex in `-c "..."` dies with `syntax error at or near ")"`

Stop fighting the shell. Write the SQL to a file and feed it on stdin:

```bash
cat > /tmp/q.sql <<'SQL'
SELECT name FROM layouts WHERE css_template ~ '(^|[^-.\w])a\s*\{[^}]*color:\s*var\(--color-accent';
SQL
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -f - < /tmp/q.sql
```

## The section-token counterfactual — "would the renderer's own value be any better?"

Added 2026-08-07. Written because `bugs_open/212` ranked two fix candidates as class
fixes without this, and both turn out to be no-ops on the case that motivated the file.
A browser can only measure what is *served*; a candidate fix is by definition not served,
so this arithmetic is the only way to grade one before building it.

**Validate the model before trusting it.** Include one row whose value IS live, and
check it reproduces the browser's number (`scripts/render_audit.py`, or 212 §3). If the
live row disagrees, every counterfactual row is worthless. On gamesdesign the model gave
1.72:1 against a browser-measured 1.72:1 — that agreement is what licensed the rest.

```python
def srgb_to_lin(c):
    c = c / 255.0
    return c / 12.92 if c <= 0.04045 else ((c + 0.055) / 1.055) ** 2.4

def lum(rgb):
    r, g, b = (srgb_to_lin(x) for x in rgb); return 0.2126*r + 0.7152*g + 0.0722*b

def ratio(fg, bg):
    a, b = lum(fg), lum(bg); hi, lo = max(a, b), min(a, b)
    return (hi + 0.05) / (lo + 0.05)

def over(fg, alpha, bg):      # composite — a semi-transparent ink is NOT its own colour
    return tuple(alpha*f + (1-alpha)*b for f, b in zip(fg, bg))
```

**The ground is the trap, not the formula.** A section's ground is usually a composite,
not a palette value: `.system-stats-section` is `--section-surface` (`rgba(255,255,255,0.05)`)
over `--color-primary`, i.e. `over((255,255,255), 0.05, (0,188,212)) = rgb(13,191,214)`.
Grade against *that*, not against `#00bcd4`. Getting the ground wrong changes the verdict,
not just the third decimal.

## Reading a `090` verdict — there is no `verdict` artifact kind

Added 2026-08-07, having looked in three wrong places first. `doc_notes` (that is the
**council** gate), `site_work_items.spec` (carries only the correlations) and
`orchestration_states` (rows vanish from the `fix_correlation_id` filter as the run
progresses, and reap at ~24h) all come up empty.

```sql
-- The run's artifacts. kind is CHECK-constrained to
-- bundle | iteration_note | fix_plan | council_report | escalation — no 'verdict'.
SELECT kind, created_at, metadata->>'decision' AS decision,
       metadata->>'symbols_unreadable' AS unreadable, metadata->>'body_chars' AS chars
FROM diagnosis_artifacts
WHERE correlation_id::text = '<RUN_CORRELATION_ID>'   -- the dispatch one, NOT the intake
ORDER BY created_at;
```

**Read it as follows.** `bundle` rows only, with no `metadata->>'decision'` on any of
them, means the run **iteration-capped: UNVERIFIABLE**. That is the same outcome the
trigger reports as "couldn't determine", and it does **not** mean the bug is hard — see
the standing lesson that UNVERIFIABLE usually means the question was wrong.

**Distinguish the two causes before concluding anything**, because they have opposite
remedies: `symbols_unreadable > 0` on the later iterations means the loop could not read
the code (the stale-code-index landmine — a symbol added since the index froze reads as
absent, so re-ask about older symbols); `symbols_unreadable = 0` means it read the code
fine and simply ran out of iterations, and the question needs narrowing instead.

**Read the last bundle even with no verdict.** Its `## Hypothesis under test` section is
the loop's live theory at the moment it stopped, and on 2026-08-07 that hypothesis was
correct and useful despite the run producing no verdict:

```sql
SELECT body FROM diagnosis_artifacts
WHERE correlation_id::text = '<RUN_CORRELATION_ID>' ORDER BY created_at DESC LIMIT 1;
```

## Is a `complete` work item a repair? Ask which producer filed it

Added 2026-08-07 (`bugs_open/213`). `item_type` does not identify the producer, and the
completion verifier is registered per `item_type` — so an item can be graded against a
predicate its author never meant. `spec->>'audit_source'` is the discriminator.

```sql
SELECT status,
       count(*) FILTER (WHERE spec->>'audit_source' = 'design-audit') AS producer_audit,
       count(*) FILTER (WHERE spec->>'audit_source' IS NULL)          AS producer_discovery,
       count(*) AS total
FROM site_work_items WHERE handler_agent = '<agent>' GROUP BY 1;
```

**The tell is an asymmetry, not a number**: if one producer's items are *never*
`unresolved` while the other's routinely are, the grader almost certainly cannot see the
first producer's defect. Then prove it at the artefact, never at the row — compare the
target's `updated_at` against the item's `created_at`. On gamesdesign the component was
last written **10.5 hours before the item existed**, which is proof of no-write in a way
that re-reading the template is not.
