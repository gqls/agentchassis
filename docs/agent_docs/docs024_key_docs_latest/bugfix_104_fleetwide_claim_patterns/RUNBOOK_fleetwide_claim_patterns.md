# RUNBOOK — fleet-wide claim-pattern dry run

Every command here was needed to get the 2026-07-28 measurement right, with the
gotcha that cost time attached. Change them **here**, not in scrollback.

## 0. The tool already exists — do not build one

`cmd/claimscan` runs the **same shared engine** as the deploy gate
(`validate_page_content` check 8) and the post-deploy audit
(`check_unverified_claims`), over exported component HTML. Its own usage block is
the reference; `sql_for_agents/226`'s verify section already names it.

```bash
go build -o /tmp/claimscan ./cmd/claimscan     # builds clean from the shared tree
```

**Why `-o /tmp/...` and not the `go run` the parent runbook uses.**
`claims_verification/RUNBOOK_claims_verification.md` §3 invokes it as
`go run ./cmd/claimscan`, and adds the rule that matters: **never leave a built
`claimscan` binary in the repo root** — `.gitignore` covers `bin/` and `/build/bin/`
but not repo-root binaries, and this tree is forward-only, so one `git add -A` from
any session makes it permanent. (Live proof, 2026-07-29: an **87MB
`config-key-audit`** binary is sitting untracked in the repo root from another
session.) Building to `/tmp` honours that rule and is preferred *here* only because
these dry runs loop over 14 sites — `go run` recompiles per invocation. Either form is
fine; a binary in the repo root is not. **claimscan itself is NOT deprecated** — it is
CLM-014, it is what `226`/`166`/`233` tell you to verify with, and `cmd/voicescan` was
built to its contract.

## 1. Which sites are armed — and the trap in the obvious query

`104` § Measurement uses `jsonb_array_length(COALESCE(ss.data->'banned_claims','[]'))`.
That returns **0 for "no evidence_base row at all" and 0 for "row with an empty
array"** — two different states that matter, because candidate 1 is gated on
`ParseEvidenceBase` returning non-nil, which is satisfied by `facts[]` alone.

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT s.domain, jsonb_array_length(COALESCE(ss.data->'banned_claims','[]'::jsonb)) AS bans
  FROM sites s
  LEFT JOIN site_specs ss ON ss.site_id=s.id AND ss.aspect='evidence_base' AND ss.is_current
 WHERE s.status NOT IN ('pool','archived') ORDER BY bans DESC NULLS LAST;"
```

Split the two states before drawing any conclusion — 2026-07-28: **6 sites have
no row**, **2 have a row with 0 patterns but non-empty `facts[]`** (robot-hands 5,
gamesdesign 4), 7 have patterns.

## 2. Export components, per site

The TSV shape is `page_name <TAB> slot_name <TAB> base64(html)`.

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -c \
 "SELECT p.name || E'\t' || COALESCE(pc.slot_name,'') || E'\t' ||
         replace(encode(convert_to(pc.rendered_html,'UTF8'),'base64'), E'\n','')
    FROM page_components pc JOIN pages p ON p.id = pc.page_id
   WHERE p.site_id = '<site_id>' AND pc.rendered_html IS NOT NULL
     AND pc.rendered_html <> '' AND pc.locked_at IS NULL" </dev/null > comp.tsv
```

**GOTCHA 1 — `kubectl exec -i` eats the loop's stdin.** A `while read` loop over a
site list terminates after **one** iteration, silently and with exit 0. Redirect
`</dev/null` on every `kubectl exec` inside a loop, or `mapfile` the list into an
array first. This cost a full run that looked like a clean single-site result.

**GOTCHA 2 — never `2>/dev/null` the fetch.** A transient `kubectl exec` failure
then reads as "this site has no evidence base", which is a *data claim*. It
happened here to **vonc** — the one site whose register mattered most — and the
table was wrong until the retry. Retry 3× and print `FETCH_FAIL` distinctly from
`no-row`.

## 3. Scan

```bash
/tmp/claimscan -evidence <eb.json> -components comp.tsv
```

**GOTCHA 3 — grep for `^BANNED`, not `banned_claim`.** The tool prints `BANNED`
and `NUMBER` line prefixes; `banned_claim` is the JSON `check` value and appears
nowhere in the CLI output. Grepping for it returns 0 on every site — a false
all-clear that looks exactly like a clean estate.

**GOTCHA 4 — some scan outputs are non-UTF-8** (site copy carries extended
ASCII), and plain `grep -c` returns **empty with no error** on them. Use
`LC_ALL=C grep -ac`.

**GOTCHA 5 — with an evidence file that has `banned_claims` but no `facts[]`,
every number on the page becomes a `NUMBER` finding.** That is correct behaviour
(nothing supports them) but it is noise for this question. Filter to `^BANNED`.

## 4. Positive control — required, both directions

A 0-findings result and a broken harness are indistinguishable. Build a synthetic
TSV with sentences the set **must** block and legitimate ones it **must not**:

```python
import base64
cases = [("ctl_block_1","hero","<p>A claim without a source does not appear here.</p>"),
         ("ctl_block_2","body","<p>Every figure is verified before publication.</p>"),
         ("ctl_pass_1","body","<p>We cite each figure and date it.</p>"),
         ("ctl_pass_2","body","<p>The statute is the authoritative text.</p>")]
for page,slot,html in cases:
    print("%s\t%s\t%s" % (page,slot,base64.b64encode(html.encode()).decode()))
```

2026-07-28: **6 of 6 block cases fired, 3 of 3 legitimate sentences passed.**

**And add the negated form of each pattern to the pass-list.** 226's own test was
"10 fabrication shapes blocked, 13 legitimate sentences passed" — and it still
missed this, because the pass-list contained no sentence that *negates* one of its
own patterns. That is the whole finding of this workstream.

## 5. Extract the candidate universal set from the SQL, not by hand

```bash
python3 -c "
import json,re
src=open('docs/agent_docs/sql_for_agents/226_overclaim_patterns_oufe.sql').read()
json.dump({'banned_claims':json.loads(re.search(r'\\\$add\\\$(.*?)\\\$add\\\$',src,re.S).group(1))},
          open('universal.json','w'),indent=1)"
```

Then confirm it is **live**, not just committed (the seed is not the system):

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -c \
 "SELECT count(*) FROM site_specs ss JOIN sites s ON s.id=ss.site_id,
         jsonb_array_elements(ss.data->'banned_claims') bc
   WHERE s.domain='oufe.com' AND ss.aspect='evidence_base' AND ss.is_current
     AND bc->>'pattern' LIKE '%is not a disclaimer%';"    # 1
```

**The set is 10 patterns, not 11**, and pattern 7 contains the literal
alternative `oufe` — it is not universalisable verbatim.

## 6. How a fix must be verified (from `104`, unchanged)

Induce **both** directions on a site with no register: a page asserting "every
claim on this site is verified" must fail with a `claims` blocker; a legitimate
process sentence ("we cite each figure and date it") must still build. Add a
third case now required by this session's finding: **"where a figure has not been
independently verified, that is stated" must still build.**

---

## 7. After the 2026-07-28 change — claimscan includes the fleet-wide set by default

```bash
go build -o /tmp/claimscan ./cmd/claimscan
/tmp/claimscan -components comp.tsv                     # fleet-wide only, as an UNARMED site is scanned
/tmp/claimscan -evidence eb.json -components comp.tsv   # fleet-wide + that site's own = what the gate enforces
/tmp/claimscan -evidence cand.json -no-global -components comp.tsv   # a CANDIDATE set in isolation
```

`-evidence` is now optional; it prints the fleet-wide pattern count to stderr so a
silently empty set cannot look like a clean estate. **Use `-no-global` to reproduce
this workstream's original numbers** — they were measured before the set existed.

## 8. Two checks this session learned the hard way

**Verify the COMMIT compiles, not your tree.** A pathspec commit of a file another
session is also editing carries their uncommitted work — and if you commit the
consumer of a type whose definition is still in their working tree, HEAD stops
compiling while your own tests stay green. `make build-<service>` builds from HEAD,
so this breaks everyone's next image build.

```bash
git archive HEAD | tar -x -C /tmp/headcheck && (cd /tmp/headcheck && go build ./platform/...)
```

Run it straight after committing platform code. The tell in the diff: a hunk whose
**context lines are code you did not write**, or insertion counts larger than the
edits you remember making.

**Quote live copy from the source, never from claimscan's output.** Its snippets are
elided with `…`, so retyping one produces a plausible sentence the site never
published. Two regression fixtures and two council `grounded_in` quotes were wrong
this way. Extract verbatim instead:

```bash
python3 -c "
import base64,re,sys
for line in open('comp_<site>.tsv'):
    p=line.rstrip('\n').split('\t')
    if len(p)<3: continue
    html=base64.b64decode(p[2]).decode('utf-8','replace')
    for m in re.finditer(r'[^.<>]*<phrase>[^.<>]*\.', html): print(repr(m.group(0).strip()))"
```

## 9. Verifying the fleet-wide set is live after a roll

**The marker must be a string only THIS change puts in the binary.** Pattern reasons
qualify; function names like `ScanAllBannedClaims` would too, but page-type or
generic words do not (the 102 workstream corrected exactly that mistake twice).

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl exec -n ai-persona-system $POD -- sh -c 'strings /app/agent-chassis | grep -c "completeness-of-exclusion"'  # MARKER: 3 (three patterns share the reason)
kubectl exec -n ai-persona-system $POD -- sh -c 'strings /app/agent-chassis | grep -c "verification-of-everything"' # MARKER: 1
kubectl exec -n ai-persona-system $POD -- sh -c 'strings /app/agent-chassis | grep -c "banned_claim"'               # POSITIVE CONTROL: 2 before and after
kubectl exec -n ai-persona-system $POD -- sh -c 'strings /app/agent-chassis | grep -c "zzz-not-a-real-marker"'      # NEGATIVE CONTROL: 0
```

Verified on **v1.0.1196**, 2026-07-29: 3 / 1 / 2 / 0. Zero on a marker with
non-zero on the control means "not rolled yet"; zero on both means your grep is
wrong, not that the fix is missing.

**While you are in the pod, run the standing fleet invariant** — it is owed after
every roll and is unrelated to this workstream (`bugs_closed/124`, migration 258):

```bash
kubectl exec -n ai-persona-system $POD -- sh -c 'strings /app/agent-chassis | grep -c "unknown execution-context field"'   # must be 1
```

Checked on v1.0.1196: **1**. Below chassis 1191 this silently stops the diagnose
lane dispatching, with no failed row to find.

## 10. Testing the GATE, not just the patterns — the seam and the trap

`sqlmock` is already a dependency, and the gate's own DB use in check 8 is a single
query, so the unarmed-site case is testable without a cluster. Switch every other
DB-touching check off by config so `loadEvidenceBase` is the only query in play:

```go
mock.ExpectQuery("SELECT data FROM site_specs").WillReturnError(sql.ErrNoRows)  // THE UNARMED SITE
res, err := ValidatePageContentAction(ctx, ActionParams{
    ExecutionContext: &types.ExecutionContext{},
    StepConfig: models.Step{Config: map[string]interface{}{
        "site_id": uuid.New().String(), "check_claims": true,
        "check_internal_links": false, "check_emails": false,
        "check_stat_claims": false, "check_stat_units": false,
    }},
    CollectedData: map[string]interface{}{"page_content": map[string]interface{}{
        "response": map[string]interface{}{"page_html": html}}},
    DB: db, Logger: zap.NewNop(),
})
```

**THE TRAP, which cost a wrong test first time round: on any blocker the action
returns `(nil, error)`.** The error *is* the mechanism — it is how the page build
fails. A test that treats that error as a failure reports a **pass on the very
outcome you want**, and inspecting the `issues` list alone is vacuous because the
map is nil on that path. Assert on the error and its blocker count; assert
`err == nil` for the copy that must still build. Import paths are
`pkg/models` and `platform/orchestration/types` (not the `orchestration/models`
that looks right and does not exist).

## 11. Withdrawing the fleet-wide set without a build (the reversal lever)

`check_claims_fleet_wide` is a `validate_page_content` step config key, **default
true**. DB config is live immediately, so this takes effect fleet-wide in seconds —
no image, no roll. Off restores the pre-104 scan exactly: per-site patterns only, and
a site with no register scanned by nothing. It does **not** disarm a site's own
audited patterns.

```sql
-- inspect: which definitions set it at all (absent = default ON)
SELECT type, jsonb_path_query_array(default_config, '$.workflow.steps.*.config.check_claims_fleet_wide')
  FROM agent_definitions
 WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
   AND default_config::text LIKE '%check_claims_fleet_wide%';
```

Set it to `false` on the step that runs `validate_page_content` for the affected
pipeline. **Read the row first and merge — never replace `default_config`** (CLM-001's
supersede rule is about `site_specs`, but the same read-before-write discipline
applies to `agent_definitions`, and a replaced config silently drops every other key).

## 12. Measuring the enforcement surface — do NOT scope by `sites.status`

The build gate **never reads `sites.status`**. The only status predicate in
`validate_page_content.go` is on the **pages** table in the link-index query. So any
population query filtered by site status measures a smaller world than the one being
enforced on — which is how round 1 of the council produced a `count 0` and round 2
produced a *correct-looking* 908 that still had an unproven gap.

Drop status entirely and **group by it** so the excluded slice is visible instead of
assumed:

```sql
SELECT COALESCE(s.status,'(no site row)') AS status, count(*) AS components,
       count(DISTINCT s.id) AS sites
  FROM page_components pc
  JOIN pages p ON p.id = pc.page_id
  LEFT JOIN sites s ON s.id = p.site_id
 WHERE pc.rendered_html IS NOT NULL AND pc.rendered_html <> '' AND pc.locked_at IS NULL
 GROUP BY 1 ORDER BY 2 DESC;
```

2026-07-29: one row only — `deployed | 908 | 14`. The 17 pool sites and the system
site hold **zero** components with stored `rendered_html`, so the filter excluded
nothing. That is a fact about today's estate, not a property of the query: re-run it
rather than trusting this line.

---

## 13. The negation guard (2026-07-29) — dry-running a pattern that the guard affects

The excluded pattern is back (`claims_global.go`, 10th entry) because
`negatedClaimMatch` now drops matches negated in the same clause. Two new commands
matter, and one control that is not optional.

```bash
go build -o /tmp/claimscan ./cmd/claimscan
/tmp/claimscan -show-suppressed -components comp.tsv          # unarmed site
/tmp/claimscan -evidence eb.json -show-suppressed -components comp.tsv   # what the gate enforces
```

`-show-suppressed` prints `negated` lines: matches the guard removed. **They are not
findings and are not counted in the exit status.** Read them — they are the only way
to tell "the guard is working" from "the pattern has stopped matching", which look
identical in a clean run.

### THE CONTROL: claimscan prints its pattern count, so use it

**A stale binary and a clean estate are indistinguishable.** This bit, hard: a
`go build` run from the scratchpad directory failed with `go.mod file not found`, the
`||` fallback failed the same way, and the *previous* binary silently scanned all 14
sites — reporting **0 findings fleet-wide** for a pattern set that actually finds 2.
The run looked perfect.

```bash
/tmp/claimscan -components /dev/null 2>&1 | head -1
# claimscan: including 10 fleet-wide banned-claim pattern(s); -no-global to exclude
```

Check that number against `GlobalBannedClaimCount()` before believing any dry run.
Build from the repo root, and never let a build failure scroll past into a loop.

### Count components with the TOOL's number, not `wc -l`

`claimscan` prints `across N component(s)`; use that. It also caught a second error:
**the enforcement surface is 919 components today, not the 908 this workstream
measured on 2026-07-28.** I "confirmed" the per-site table summed to 908 by adding it
up wrong (dropping oufe's 11), and so recorded the corpus as unchanged when it had
grown by 11. Re-derive with one query rather than summing a table by eye:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -c "
SELECT COALESCE(s.status,'(no site row)'), count(*), count(DISTINCT s.id)
  FROM page_components pc JOIN pages p ON p.id=pc.page_id LEFT JOIN sites s ON s.id=p.site_id
 WHERE pc.rendered_html IS NOT NULL AND pc.rendered_html<>'' AND pc.locked_at IS NULL
 GROUP BY 1 ORDER BY 2 DESC;"     # 2026-07-29: deployed|919|14
```

### Dry-run BOTH forms of a pattern you are about to narrow

Narrowing a pattern to dodge a hypothetical false positive can make it **inert**, and
an inert pattern looks exactly like a well-behaved one. Measured here, same corpus,
same run:

| candidate | findings | suppressed |
|---|---|---|
| bare `(fully\|independently\|externally\|properly) (verified\|audited\|fact.?checked)` | **2 real** | **4** |
| subject-anchored (content noun + `is/are` within N chars) | **0** | 0 — matched nothing at all |

The anchored form was my first shipped attempt. Always run the pattern you are
replacing next to the pattern you are replacing it with.

## 14. Current live findings — this set is no longer clean

2026-07-29, 919 components / 14 sites, fleet-wide set + each site's own register:
**2 findings, 4 suppressed.** Both findings are robot-hands.com (`gripper-catalog`,
`how-it-works`) asserting spec data is "independently verified" — filed as
`bugs_open/147`. Those two components will not rebuild until the copy changes.

The 07-28 headline "0 findings across the whole surface" is therefore **spent, and it
was an artefact of the excluded pattern** — not evidence the estate was clean. When
you next quote a clean-sweep number, name the patterns that were armed when you
measured it.

## 15. Verifying the GUARD is live after the next roll (owed — raised by the council)

The guard is a Go change, so it is **inert until a chassis image ships**. The
`needle_gate` seat objected at low that the plan cited only dry-run and test evidence,
never a running-pod grep — fair, and this is the marker to use. Pick a string only
**this** change puts in the binary; the reason strings qualify, function names do too.

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl exec -n ai-persona-system $POD -- sh -c 'strings /app/agent-chassis | grep -c "external-verification claim"'   # MARKER: 1 (the 10th pattern's reason)
kubectl exec -n ai-persona-system $POD -- sh -c 'strings /app/agent-chassis | grep -c "suppressed as negated"'          # MARKER: 1 (the gate's suppression log)
kubectl exec -n ai-persona-system $POD -- sh -c 'strings /app/agent-chassis | grep -c "completeness-of-exclusion"'      # POSITIVE CONTROL: 3 (pre-existing, unchanged)
kubectl exec -n ai-persona-system $POD -- sh -c 'strings /app/agent-chassis | grep -c "zzz-not-a-real-marker"'          # NEGATIVE CONTROL: 0
```

**Check BOTH replicas** — `logs deploy/X` and a single-pod grep each read one pod of N.
Zero on a marker with non-zero on the control means "not rolled yet"; zero on both means
your grep is wrong. **A RETAG IS NOT A REBUILD**: compare `.ID` and `.CreatedAt`, because
two tags have shared one image id on this fleet before.

Once live, the guard's behaviour is observable **from the build logs**, not only from
this tool:

```bash
kubectl logs -n ai-persona-system -l app=agent-chassis --tail=2000 \
  | grep "claims gate: banned-claim match suppressed as negated"
```

Each line carries `site_id`, `pattern`, `matched` and `snippet`. **If that grep starts
returning matches on sentences that are NOT denials, the guard is over-firing** — that
is the failure this logging exists to make visible, and the fix is the cue list in
`negationCueRe`, not the pattern.
