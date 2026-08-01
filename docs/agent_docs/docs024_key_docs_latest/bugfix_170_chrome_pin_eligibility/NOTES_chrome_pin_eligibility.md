# NOTES — bugs_open/170, the chrome pin

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-08-01 — picking it up

170 was filed by the `bugfix_167_chrome_build_path` lane on 2026-07-31 and handed
off explicitly: both adjacent lanes (167, and 118/166) closed and their last
messages name 170 as the next decision. `who-owns.py 170` says OWNED-or-recently-active,
which is the lagging signal — the transcripts of the two busy sessions
(`b5017ee5`, `f4cc677d`) both end in wrap-up, not work. Grep of every `.jsonl`
touched in the last 6h for `GetComponentByID|style_collections|deactivated_site_components|
HeaderComponentID|chromeEligibleSQL` found nobody working it. Unowned.

### Re-verified: the bug is real, and the filing UNDERCOUNTS it

The filing's query, re-run:

| domain | header pin | active | footer pin | active |
|---|---|---|---|---|
| ai-agent-orchestration.com | header-professional-dark | **f** | footer-4-column | **f** |
| finetuning.uk | header-professional-dark | **f** | footer-4-column | **f** |
| gaswholesalers.com | header-professional-dark | **f** | footer-4-column | **f** |
| leopardessconsulting.co.uk | header-leopardess | t (own fork) | footer-4-column | **f** |

170 says "three deployed sites". That is the HEADER count. **Four** sites are
pinned to a deactivated footer — leopardess's header pin is the fleet's one
legitimate pin and its footer pin is not. Broadened from `style_collections`
rather than `sites`, it is **4 collections** carrying pins, of which
`bold-gradient` and `minimal-light` are pinned to deactivated headers and used by
zero sites.

### The finding the filing does not contain: the pin is a WRITE source

`link_site_components_action.go:79-122` reads the pin and calls
`ResolveChromeComponent` — 118's eligible-only pool lookup — **only when the pin
is NULL**:

```go
if !headerCompID.Valid {          // ← the ONLY gate on a pin
    if comp, eligible, err := ResolveChromeComponent(...); err == nil && eligible {
```

A present pin goes straight to `relinkSiteComponent`
(`site_component_lock_guard.go:162`), which upserts it into
`site_components.component_id` **and** sets `rendered_html = NULL,
build_status = 'pending'`.

And the four sites' assignments are, right now, all **correct** —
`header-theme-chrome` / `footer-theme-chrome` / `header-leopardess`, repointed
2026-07-31 by 118's hand-repoint and 166's automatic one. So the two stores
disagree in writing and the unguarded one wins. **166's repair is revertible by a
routine action.**

**[MEASURED] Latent, not firing.** `site-component-linker` is the only live agent
carrying `link_site_components` and is the wired handler for two discovery checks,
so it is dispatchable — but no run appears in `orchestration_states`, whose entire
retention is 2026-07-13 → 2026-08-01 (3,309 rows). So: armed and revertible, not
reverting this week. Said that way in the bug file and the submission rather than
dressed up.

### A third consumer, found by the test rather than by reading

`fork_theme_from_site_action.go:239` copies the parent collection's pins into a
new collection unconditionally. Forking `professional-dark` today manufactures a
NEW collection pinned to two deactivated components. I did not find this by
reading the code — I found it because the scan test I wrote for the *first two*
consumers failed on it. That is the argument for the scan in one line.

---

## Missteps

### 1. My scan test failed on my own fix, and the reason is a real blind spot

First version of `TestNoConsumerDereferencesAChromePinUnguarded` collected the
backtick-quoted SQL literals in a file and flagged any that read a pin column
without `component_level` beside it. It reported **`link_site_components_action.go`
— the file I had just corrected** — as an offender.

Because the *sanctioned* guarded form injects the predicate by concatenation:

```go
hc.name, (` + chromePinEligibleSQL("hc.") + `),
```

so `component_level` is in the Go glue BETWEEN two string literals and appears in
neither. A correct fix and a completely unguarded read looked identical to the
scan. Fixed by stitching each literal back together with the glue chunks either
side — statement-local, not file-wide, because "does this file mention the
predicate anywhere?" would wave through a file that guards one read and leaves
another bare.

### 2. …and then the lockstep test passed for the wrong reason

`TestChromePinPredicateMatchesTheActionsPackage` asserts the detector's hand-typed
predicate matches the shared one. First version searched the whole of
`check_integrity.go` for each level string. Mutation-tested by narrowing the
detector's list to `('site','header')`, it reported **only `'head'`** as missing.

`'footer'` passed — because that literal appears elsewhere in the same file as a
**slot name** (`SELECT 'footer', fc.name, …`). A file-wide containment test was
answering a different question from the one I asked, and it caught that drift by
luck, on the single level with no homonym. Now scoped to the
`component_level IN (…)` clauses. Re-mutated: narrowing reports both missing
levels, widening (adding `'section'`) reports the extra one, both mutants compile
cleanly, clean is green.

Both of these are the same error twice in one session: **a check whose population
is "the file" rather than "the thing".** Logged in `WRONG_CALLS.md`.

### 3. The 090 diagnosis run produced NO verdict, and the reason is structural

Filed per the owner ruling of 2026-07-31 (a `bugs_open/` file asserting a
cross-cutting root cause goes through the loop, or the session states its
substitute). Intake `a55675a1-ef91-42cb-86bc-a4301d918510`, run
`ce9bcd92-7be7-4819-bdf8-f8a57622128f`.

It ran four iterations, wrote four bundles, and **wrote no verdict, no
iteration_note and no doc_note.** Iteration 4 says why:

> `_(body omitted — 75174 chars, and 0 of the 60000-char body budget is already
> spent. It was found; it did not fit.)_`
> `**This section is INCOMPLETE.** 0 of 1 in-scope symbol(s) rendered with a body.`

`component_library.go` is **93,905 bytes on disk**. The loop's body cap is 60,000
chars. The one symbol the hypothesis is about could not be shown to the diagnoser
in any iteration. **A file over the cap is invisible to the diagnosis loop**, so
for that file the loop is not available as the verification route the ruling
names — which matters beyond this bug and is going to `016b` §9 and `LANDMINES.md`.

What it DID produce, from its own independently-written query, was the live state:
3 header + 4 footer ineligible pins, matching my count exactly. Corroboration of
the measurement, not of the mechanism.

**Substituted first-hand verification**, and it is stronger here than usual: the
absent predicate is visible in fifteen lines of quoted source, not inferred from a
symptom; both new queries were `PREPARE`d against the live schema before shipping
(`go build` cannot parse SQL); the guard is proven by an induced consumer that
compiles cleanly; and the lockstep is proven in both directions.

---

## Verification record

- `PREPARE` against live schema: `link_site_components` pin+eligibility query,
  `GetChromePinComponent` dereference, `fork_theme_from_site` guarded copy,
  and the detector's UNION — all four parsed.
- Pin predicate vs pool predicate over all four live pins: they disagree on
  exactly one row (leopardess header: pin TRUE, pool false). Positive and negative
  control in one query.
- Detector preview: files exactly **7** items fleet-wide, zero for the legitimate pin.
- Guard induced (`zz_induce_170_tmp.go`, both forms): **compiles cleanly**, both
  caught, removed → green.
- Lockstep mutated narrow and wide: both **compile cleanly**, both fail with the
  reason, restored → green.
- `go build ./...` + `go test ./platform/...` green against a **clean
  `git archive HEAD` tree** with only my files overlaid — necessary, because the
  working tree does not compile: `ai_actions.go:347 declared and not used:
  declaredOutputType` is another session's in-flight edit (the `bugs_open/119`
  lane). A local `go test` in the shared tree is not a signal either way.

---

## Council round 1 — **APPROVED** (`21bac2a2-2b46-4883-894f-19d7ec5e5b45`)

`approved with 7 advisory objection(s) — none high-severity`. 13 seats reviewed,
3 abstained, no seat truncated (`gated_by_truncation: false`).

Worth recording that four seats went out of their way to say the grounding was the
reason: `reuse_agent` — "unusually well-grounded: it reuses the deleted 167-round
predicate rather than inventing one, explicitly justifies why it is NOT the pool
predicate, extends the existing detector rather than creating a new one";
`constitution` — "deferrals are explicitly stated and justified, not smuggled";
`render_guardian` — "explicitly disclaims using a rerender as proof of live effect
rather than smuggling a false-green rerender claim"; `bug_historian` — "a strong,
self-aware instance of the documented pattern rather than a naive one".

### The one theme five seats converged on — and they are right

`architecture`, `reuse_agent`, `editquality`, `debug_historian`,
`prior_art_librarian` and `guardian` all landed on the same thing from six angles:
**this is now the third and fourth hand-maintained guard over one chrome-eligibility
vocabulary**, and a landmine already records two earlier ones with disjoint blind
spots. The architecture seat put the cost precisely — *"every new pin/eligibility
consumer now costs one more regex-scanned guard file instead of one import"* — and
named the remedy: move the predicates into a package both `actions` and
`discovery_checks` can import, and **delete** the lockstep rather than harden it again.

Accepted, and filed as **`architecture_review/RFC_007`**. It is not actioned in this
lane: the import-direction fix is a separate change with its own blast radius, and
the seat itself said "after this ships".

`debug_historian` also noted my own NOTES entry against me, correctly — that both new
guards passed for the wrong reason on their first version is "evidence this class of
guard is fragile, not evidence it is now solid". That is a better reading of my own
missteps than the one I wrote.

### Two objections answered by measurement, after the verdict

Both concern edit 4's `needs_human_review` / no-handler routing.

- **`improvement_guardian` (medium):** the config-only-enablement contract wants a
  finding inserted at `detected` so the triager promotes it; is a non-`detected`
  initial status even honoured by the runner? **Verified: yes.**
  `discovery_checks.go:224-240` passes `wi.Status` straight through to
  `insertWorkItem` with no override, and live there are **190+ discovery-sourced
  `needs_human_review` items across 8 item types**, most recent 2026-07-30. It is the
  established pattern, at scale, not an improvisation.
- **Dedup, which no seat raised and is the one that would actually have bitten:** the
  live `idx_swi_dedup` predicate excludes `complete, verified, rejected, wont_fix,
  failed, unresolved, cancelled` — `needs_human_review` is **not** in that list, so it
  counts as non-terminal and `deactivated_pin_<slot>` dedupes correctly. One item per
  site+slot, not a fresh one every sweep. (Checked because of the standing
  `idx_swi_dedup` ↔ `workItemTerminalStatuses` lockstep landmine.)
- **`bug_historian` (medium) STANDS and is recorded as a real limitation:**
  `bugs_open/033` — the human-review queue has no working surface — **is still open**,
  and those 190+ items are the evidence for it. My 7 join that pile. The alternative
  is worse: routing at `rerender-pages`, which cannot write a `style_collections` row,
  reproduces `bugs_open/166` on a new item type. So the routing stays, and the honest
  statement is that these items are a **durable, queryable record** rather than a
  worked queue. Written into the bug file, not left to be discovered.

### Answered in passing

- **`debug_historian` (medium), no post-deploy verification step:** there is one — it
  was in the bug file, not the submission. `bugs_open/170` § "How to verify the FIX"
  carries the pod-grep with a real negative control and, more importantly, **both
  branches induced**: an ineligible pin must be ignored AND leopardess's eligible fork
  pin must still be honoured. The second is the one that catches the only way this fix
  goes badly wrong.
- **`prior_art_librarian` (medium), the 167-round history is asserted without a
  `council_report` lookup:** fair as stated. The evidence I actually used was
  `git show 2605d3f92` — the commit that deleted `chromePinEligibleSQL` and its three
  tests, whose message names the four objecting seats. That is stronger than a
  verdict-note lookup for the specific claim ("the predicate was deleted with the
  reporter, and the reporter was the objection") but it is not what the seat asked
  for, and it was not attached to the submission.
- **`tooling_provenance` (medium), no `doc_notes` write:** partially satisfied without
  being an edit — `scripts/landmines-sync.py --apply` syncs both new `LANDMINES.md`
  entries into `doc_notes`, so the pin-vs-pool asymmetry IS readable by seats and
  agents from the next round on. That was not stated in the submission.

**No amend, and none needed:** the commit carries `Council-Submitted:`, and `098`
resolves the correlation at report time, so `e44e6dd06` is credited automatically now
the verdict is approved. Forward-only holds.

### Misstep 4 — I hit the trap my own RUNBOOK had already written down

The final HEAD verification came back `FAIL github.com/gqls/agentchassis/platform/
orchestration/actions [build failed]`. The cause was `no space left on device`: I had
left two ~1GB `git archive` trees in a 16G tmpfs that several sessions share (89% at
the start of this lane, 94% by the end). RUNBOOK §R5 already said "check `df -h /tmp`
first". I wrote that line and then did not do it.

Worth recording for the shape rather than the tidiness: **a disk-exhausted build
reports as `[build failed]`, which is indistinguishable at a glance from your change
breaking HEAD** — and it arrives at exactly the moment you are looking for that. The
real output is three lines up (`link: mapping output file failed`). Cleaned up,
re-ran, HEAD builds clean and every actions test passes at `e4e003f13`. §R5 now says
to clean up after, not just to look before.
