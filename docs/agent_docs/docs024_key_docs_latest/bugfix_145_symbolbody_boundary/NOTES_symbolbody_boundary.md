# NOTES — `bugs_open/145` symbolbody boundary

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-07-31 — session "bugfix 12", opening

### Picking the bug

`scripts/who-owns.py` returned **`VERDICT: OWNED or recently active.` for every open
bug** — it counts any commit mentioning the number, and on this tree that is all of
them. **[MISSTEP]** I ran it 32 times before noticing the verdict line carried no
signal at that granularity. What actually discriminated was the pair
*(last commit touching the bug FILE, does it name an owning workstream)*:

```bash
for f in bugs_open/*.md; do n=$(basename "$f" | grep -oE '^[0-9]+')
  git log -1 --format='%ad' --date=short -- "$f"
  python3 scripts/who-owns.py "$n" | sed -n '/likely OWNING/,/^$/p' | grep -c 'none identified'
done
```

Eleven bugs had no named workstream. Of those, 145 was the only one whose own file says
**"OPEN, UNOWNED"** in as many words, and `grep -rn "bugs_open/145" docs/` returned
**nothing** — no thread has written about it anywhere. Both files it touches were clean
in the tree.

Rejected on ownership evidence, worth recording so nobody re-walks it:
- **143** (`derive_card_asset` lock ordering) — filed by the `bugfix_131_og_card` lane,
  and *while I was working* another session's uncommitted `assetLock` helper appeared in
  `platform/orchestration/actions/`, which is 143's stated fix shape. Actively owned.
- **091, 097, 111** — `bugs-sweep` lane (MEMORY names 091 as its next item; 111's own
  file carries a bugs-sweep status block and blocks on 117/118).
- **080/081** — need an owner call, per their files.

### The filing's reachability claim is wrong, and I nearly inherited it

145 says the bare-path branch is *"currently unreachable by accident of the writer"*,
having followed `scopeFromCodeResults` (whose `code_symbols.symbol` is `NOT NULL`).

**[VERIFIED by reading the whole chain]** That is the wrong producer. `route.scope` is
the FIRST entry in the bundle's scope fallback chain
(`diagnose_assemble_bundle_action.go:139-140`) → `diagnose.EncodeScope(decision.NextScope)`
(`diagnose_route_action.go:226`) → `Verdict.NextScope`, the **LLM verdicter's own JSON**
(`pkg/diagnose/verdict_wire.go:29`). `namedScope` (`pkg/diagnose/loop.go:409-415`) trims
and dedupes; that is the entire validation.

And the bundle **asks** for it — `diagnose_assemble_bundle_action.go:597`:

> `- _(+%d more in this file — put the bare file path in next_scope to see it whole)_`

So the branch is a live, advertised capability driven by the least trustworthy producer
in the loop. Two things follow, and the second is the one that mattered:

1. **145's fix candidate 1 must NOT be taken as written.** "Drop the branch or move it to
   an explicit `ReadWholeFile`" would break a documented capability. I had started
   towards candidate 1 because the filing ranks it first under "makes the bad state
   unrepresentable" — reading `:597` is what stopped me. **[MISSTEP AVOIDED, barely:
   I was ordering by the filing's own ranking rather than by what the code does.]**
2. The §7D resolver is not a filter. An unresolvable path-shaped entry **fails open to
   its original string** by explicit contract (`diagnose_route_action.go:512-600`).
   It looks like the safety net; it is enrichment.

### The piece of prior art that decided the design

`ReadSymbolBody` was extracted from contextkit's `cmd/assembler`, and **that caller
already has this guard** (`docs019_…/contextkit/cmd/assembler/main.go:178-200`):

```go
fi, ok := byPath[path]
if !ok { w("> scope %q: path not found in analysis — skipped\n\n", sc); continue }
```

So the invariant existed and **the chassis port dropped it**. That reframes the fix from
"add a check" to "restore a precondition, and put it where a port cannot lose it" —
which is exactly what the architecture seat meant by *"a generic hazard reachable by any
future scope-entry producer"*. It also answers the negative control 145 asked for: if
`cmd/assembler` is ever revived, the change is a **no-op** for it.

### Why Output membership beat the filing's `.go` extension test

`Output` **is** the analyser's inclusion rule (`analyse.go:80-99`): Go, non-test, minus
`vendor/`, `testdata/`, download-duplicates, `excluded()`. An extension test would be a
second, drifting copy of that rule — and would still admit a vendored or excluded file.
**[MEASURED, not argued]** the regression test includes `f_test.go`,
`vendor/dep/dep.go` and `testdata/sample.go` precisely because all three are `.go` and
all three leaked pre-fix.

### The severity is worse than filed, and this one is measured

145 scopes the leak to files committed in the repo. **[MEASURED 2026-07-31]** With a real
file placed outside the checkout, the pre-fix code returned it:

```
ReadSymbolBody("../outside.go") LEAKED a file from outside the analysed root
```

`filepath.Join(root, "../outside.go")` resolves and `os.ReadFile` succeeds. So the leak
was never bounded by the repo at all. **[NOTE ON HOW I GOT THIS]** my first version of
the test asserted traversal against paths I had never created, so it passed against the
*pre-fix* code too — a quiet pass that proved nothing. Restructuring the fixture to two
levels (`base/repo` as the analysed root, `base/outside.go` as a real file) is what
turned an `[INFERRED]` claim into a measured one. **A traversal assertion whose target
does not exist is indistinguishable from a working guard.**

### Both-directions verification

- Against the fix: `ok github.com/gqls/agentchassis/internal/analysis`, and the whole
  `platform/orchestration/actions` package passes (0.240s) on clean HEAD + my three files.
- Against HEAD's pre-fix `symbolbody.go` in an isolated module: **FAILS on all six** —
  five leaked bodies plus the out-of-root file.

### Shared-tree friction, both instances someone else's

- `go vet ./platform/orchestration/actions/` failed on
  `derive_brand_head_assets_test.go:108` (`assetLock`). **Not mine, and not at HEAD** —
  another session's uncommitted WIP. Checked by extracting `git archive HEAD` to scratch
  and vetting there: clean but for a long-standing `unreachable code` notice.
- `go build ./...` is amber at HEAD anyway: `cmd/reasoningset/main.go:504` unused
  variables, and a mixed-package directory under `docs024…/traffic_probe/`. Neither mine;
  neither in the chassis binary's path. Recorded so the next thread does not chase them.
- Checked all three of my files for a **same-file passenger** before committing
  (`git diff` on each): only my changes present.

### Misstep: I disarmed the landmine verifier by following CLAUDE.md literally

Ran `landmines-sync.py --apply` as CLAUDE.md instructs; it printed
`NEEDS_VERIFICATION:LANDMINES.md#readsymbolbody-…`. Then ran the wrapper that acts on
that signal, which answered **"Nothing needs verification"** — because the wrapper greps
the output of the `--apply` *it* runs, and mine had already written the row. **The signal
is consumed by the write.** Exit 0, reassuring totals, no check performed. Recovered with
`trigger-landmine-verifier.sh` directly (corr `329710d1`). Full entry in `WRONG_CALLS.md`;
the fix for the next session is in this workstream's RUNBOOK. CLAUDE.md's landmines
bullet names only the inner script — flagged, not silently edited, since that file is
co-edited.

### On not running 090 (declared, per the owner ruling of 2026-07-31)

That ruling requires a `bugs_open/` file asserting a cross-cutting root cause to go
through the diagnosis loop **or for the filing session to state plainly why it
substituted equivalent first-hand verification.** I am not filing a new cross-cutting
claim — 145 was filed by the architecture seat and I am *narrowing* it (correcting its
reachability claim and raising its severity). My substitute, stated so it can be judged:
I read every function in the chain end to end rather than grepping it
(`ReadSymbolBody` → `:201` consumer → `:597` footer → `:139-140` fallback chain →
`EncodeScope` → `namedScope`/`nextScope` → `VerdictWire.NextScope`), and I have a test
that **reproduces the leak against the pre-fix binary and fails when the fix is present**
— which is stronger evidence than a verdict, because it is re-runnable by anyone. The
durable claims are also going to the council gate (`bce4caab`) and to the landmine
verifier (`329710d1`), i.e. two independent readers, not none.

### Council gate

Submitted `bce4caab-17b6-4bbb-ba6b-d1e18f196156` **before** committing; committed with
`Council-Submitted:` since forward-only forbids the amend that `Council-Reviewed:` would
need. Guarantee-change test applied (owner ruling 2026-07-29 #1): this **narrows**
`ReadSymbolBody` from "any path under root" to "a file the analyser parsed" — a
restriction, not a new shared capability — so council gate, not RFC. Consumers named in
the submission's `risks` per ruling #3. Register CTXK-002 updated **in the same commit
as the seam**, per the surviving condition (2) of the ordering exemption.

Three things I deliberately did **not** do, all disclosed to the council rather than
quietly dropped:
- **No per-read size bound** (145's candidate 3). An analysed Go file is still unbounded
  at the read, capped only by `maxBodyChars` at the bundle.
- **Did not fix an adjacent pre-existing wart I found**: the bundle loop `break`s rather
  than `continue`s when a body would exceed `maxBodyChars`, so **one oversized body
  silently drops every remaining scope entry**. Separate concern, separate blast radius.
  Flagged to the council in case a reviewer disagrees.
- **Did not edit the archived contextkit copy.** CTXK-002 carries a "byte-for-byte"
  sync obligation, so I measured it first: **already 29 diff lines out of step** before
  my change (archived has unexported `sliceLines`; the chassis exported `SliceLines` on
  2026-07-27). It is in no build and its caller has the guard, so I recorded the drift
  in the register rather than editing a historical snapshot.

---

## 2026-07-31, later — the council round, and two process missteps

### Misstep: the first council submission died at validation, on a rule already documented

`bce4caab` reached `complete_invalid` in **five seconds**, before any seat fired:
`plan failed validation: edit 4: sketch declares no code change`. `noOpEditReason`
(`diagnose_persist_fix_plan_action.go:547-563`) is a literal `strings.Contains` over the
lower-cased sketch against nine phrases, and I had written *"No code change in this file."*
inside edit 4's sketch to be helpful about a comment-only correction.

**The mechanism, the file, the full nine-phrase list and the workaround are already in
`fixloop_eg_dartsonline/RUNBOOK_council_gate.md` — line 241 and a titled LANDMINE section
at 332-356**, contributed 2026-07-26 by a lane that hit the identical wall. CLAUDE.md names
that runbook in the same sentence that tells you how to submit. I read the **097 script
header** for the schema and stopped, and the runbook's own bullet says *"the plan schema is
stricter than the 097 header suggests"*. **Reading the script that fires a mechanism is not
reading the mechanism.** Full entry in `WRONG_CALLS.md`; the one-line pre-check is now in
this lane's RUNBOOK.

Cheap in credits (validation refuses before any seat runs, so nothing was paid for) but
**silent where I was looking**: an invalid run writes **no `diagnosis_artifacts` rows at
all**, so polling the verdict by correlation waits for ever while
`orchestration_states.status` reads a reassuring `COMPLETED`. The reason is only in
`collected_data->'__step_error'` — the `099` landmine I already knew, which is why I got
the answer in seconds once I looked in the right column.

**One thing that went right by accident and is worth knowing deliberately:**
`RESUBMIT_CORR` **preserves the trail correlation** and changes only the run envelope
(`bce4caab` trail, run `ec292a46`). So the `Council-Submitted: bce4caab-…` trailer already
sitting in commit `691c1817a` stayed **correct** across the resubmission and needed no
follow-up — which matters, because forward-only forbids the amend that fixing it would
have required. Had I submitted the retry fresh, that committed trailer would have pointed
at a `complete_invalid` run for ever and read as un-reviewed in the 098 report.

### Resolved: the design decision the validator provoked, and why I did NOT drop the edit

My first instinct was to drop the two comment-only edits from the plan, on the reasoning
that the validator's *intent* is "a fix plan proposes changes, not observations". Reading
`noOpEditReason`'s own header changed my mind — it says explicitly *"over-blocking a real
edit is worse than letting the council catch a subtle no-op"*, and the runbook's landmine
section tells you to treat the list as **a keyword blocklist, not a judgement about your
edit**. A stale citation that caused a real wrong claim in 145's own filing is a real
change. So the edit stayed and the prose moved to the `rationale`, which is not scanned.
**[MISSTEP AVOIDED]** — I would otherwise have hidden a change from the reviewers in order
to satisfy a substring check.

### Note on the commit's pattern-check output (a false positive worth not chasing)

The commit hook's `logged-model-output` check flagged
`diagnose_assemble_bundle_action.go:210` — *"log call passes `body` unwrapped"*. Line 210
is `fmt.Fprintf(&b, "### %s\n```go\n%s\n```\n\n", sym, body)`, writing into a
`strings.Builder`; there is no logger involved and writing the body into the bundle is the
function's entire purpose. The check cannot distinguish `fmt.Fprintf(&builder, …)` from a
log call. Advisory, never blocks, and the three `new-capability-surface` hits ("proposes
`cmd/assembler/`, which does not exist") are the check firing on docs that say precisely
*that it does not exist* — its own text predicts this ("naming a path you have
deliberately decided AGAINST fires this too"). **Recorded so the next reader of this
commit does not re-investigate four advisories that are all correct-by-design.**

### Shared-tree events during the session, for the record

- **My `LANDMINES.md` append was swept into another session's commit `d194798fd`**
  ("wrong-call + cut back…", 16:59 BST) before I got to it. Nothing lost; forward-only
  holds; noted in my own commit message rather than silently.
- **`makefile` has `IMAGE_TAG` bumped to `v1.0.1215` uncommitted by another session** (HEAD
  says `v1.0.1206`), i.e. somebody is mid-build. I did not touch it — see the roll note
  below.
- The `143` lane committed its shared `asset_lock_guard.go` at 17:03 BST (`3aa7a5d17`),
  which is the WIP that had been breaking `go vet` on the actions package all session.

---

## 2026-07-31, later still — APPROVED, and a side-quest that corrected itself twice

### Council: APPROVED at round 1

`complete_approved`, `decision: approved`, *"approved with 2 advisory objection(s) — none
high-severity"*, 5 seats abstained (relevance filter), `gated_by_truncation: false`.

- **`editquality` — object, low ×2.** Edits 2 and 4 are comment-only, "scope creep against
  minimality; could be dropped without weakening the fix". **Not acted on, deliberately:**
  the filing explicitly asks whoever takes 145 to correct those citations, and one of them is
  the comment that caused 145's own wrong blast-radius claim. Recorded here so the decision
  is visible rather than silent. Its note on the core edit is the useful part — it confirmed
  the causal path independently and endorsed keeping the whole-file branch against the `:597`
  landmine.
- **`bug_historian` — object, MEDIUM, and it was right.** I disclosed the `maxBodyChars`
  `break`-not-`continue` wart in `risks` and left it "for a reviewer to say so". The seat
  called that out as a byte-for-byte match to an indexed §9 pattern, found while auditing the
  very function I was editing, and said it should be **filed now**. **Filed as
  `bugs_open/164`.** This is the seat doing exactly what it exists for, on me — worth saying
  plainly, because it is the argument for keeping paying for it.
- **`reuse_agent` — approve.** Its residual: it could not confirm from the SQL/code tier
  whether `cmd/assembler`'s `byPath` guard has a living generalised counterpart I should have
  extended. Answered here: no — `go list ./...` gives 0 contextkit packages, and the sibling I
  did find (`flattenSymbols`) is already bounded by `Output` by construction.

### The side-quest: two wrong root causes in one hour, both caught by re-firing

Appending the landmine entry led to a `NEEDS_HUMAN_REVIEW` from RFC_005 §3.2's verifier,
blaming a stale code index. That claim was false (every symbol present at the very
`commit_sha` it named), so I went looking — and then filed the wrong cause **twice**:

1. **The comma split mangles parenthesised symbols.** Real defect, measured 17 of 100 entries,
   **causally inert.** Refuted by rewriting my footprint to comma-safe `path:Symbol` and
   re-firing (`113fd03f`): identical 0 rows, identical invented cause.
2. **Then separate path and symbol into their own items.** Refuted the same way (`f7056e8a`).
   I had asserted that form worked from reading the SQL rather than waiting for the verdict —
   **the identical error, twice, in the same file.**

The actual cause needed the run's own persisted queries
(`collected_data->'derived'->'result'->'code_checks'`): `derive_checks`' prompt **defines**
kind `"symbol"` as `path:Symbol`, and `symbolTokenClause` ANDs every identifier token against
the **`symbol` column**, path tokens included — so the query is unsatisfiable by construction
(`0` rows, vs `1` for a bare name). **23 of 23 symbol checks across all 9 runs ever are
path-bearing**, so no landmine has ever had a symbol confirmed and no footprint form avoids
it. Filed as `bugs_open/163`, corrected twice in place with the refuted steps kept — the
three-verdict table is the most useful thing in that file, because the middle rows eliminate
the obvious explanation.

**[MISSTEP, the generalisable one]** my 17/100 measurement was true, striking, freshly
discovered, adjacent to the symptom — and not the cause. That combination is the most
seductive wrong root cause there is. **Say what would change if you removed the cause, then
remove it and look**, before filing. I did do that, which is why this cost an hour rather
than becoming a handoff everyone believed; but I did it *after* writing the file, and the file
was wrong in between. Third `WRONG_CALLS.md` row of the session, all three the same family:
**I read the code that builds a thing instead of running the thing.**

### Scoreboard of what this bug produced beyond its own fix

- `bugs_open/163` — landmine verifier's symbol lookup (fleet-wide, all history).
- `bugs_open/164` — the bundle's `break`-not-`continue` cap, at the council's request.
- 3 × `WRONG_CALLS.md` rows; 2 × new 016b §9 patterns; 1 × `LANDMINES.md` entry;
  CTXK-002 updated with the seam, its landmine and two verify-laters answered.
