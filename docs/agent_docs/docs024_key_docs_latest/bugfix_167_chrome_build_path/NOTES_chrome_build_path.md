# NOTES — bugs_open/167, the page-build chrome path

Append-only, newest at the bottom.

## 2026-07-31 — picking it up

The prompt named no ticket (a `/clear` had removed whatever identified it), so I
asked rather than guessed: 40 files in `bugs_open/`, and a wrong guess costs a
council round and a commit on the shared HEAD. Owner chose 167.

**Three sessions were in the chrome area at once.** `scripts/who-owns.py 167` said
OWNED by `bugfix_118_chrome_selection` — but it reads COMMITS, so it is lagging by
construction. Grepping live `.jsonl` transcripts for the code symbols found:

- `b5017ee5` — **live, written to 3 seconds before I looked**, 61 hits on
  `ResolveChromeComponent`/`RenderHeader`/`InjectHeader`, actively editing
  `render_site_components_action.go`. Reading its last turns: it is working
  **`bugs_open/166`**, not 167.
- `c718b278` — started two minutes before me and running the **identical generic
  prompt**, doing the same transcript check I was. It had not yet picked a bug.
- a third session, invisible to any bug-level check, mid-edit in `prune_floor.go`
  (see below).

So the ownership question had three different answers depending on the instrument:
`who-owns.py` said "the 118 lane" (true but finished), the transcripts said "166,
adjacent, different file", and only reading the *other* session's prompt showed a
peer that might land on the same bug by coincidence.

**Consequence for the fix**: `chrome_selection_test.go` is 118's guard file and is
being edited right now by the 166 session. My tests went into a **new** file,
`chrome_build_path_test.go`. A pathspec commit cannot exclude a same-file
passenger, and this is the cheap way to not have one.

## The filed bug's blast-radius table is stale — corrected

The bug file is why 167 was an owner call: it says the build path resolves
`site-header` to `site-header` (`component_level='section'`, 6,614 chars), so
fixing it would flip every page's header and footer to a different component.

I ran each function's own query verbatim against `clients_db` before believing it:

| function | `GetComponentByFunction` | `ResolveChromeComponent` |
|---|---|---|
| `site-header` | `header-theme-chrome` **[site]** | `header-theme-chrome` [site] eligible=true |
| `site-footer` | `footer-theme-chrome` **[site]** | `footer-theme-chrome` [site] eligible=true |
| `head` | no row → fallback | `Document Head` [section] eligible=**false** |

**They already agree.** `content_components.updated_at` shows why:
`header-theme-chrome` was activated `2026-07-31 12:39:53`, in the same second three
sibling headers were deactivated — 118's own fleet repoint, hours *before* 167 was
filed. Once it is active, `ORDER BY name` sorts `header-theme-chrome` ahead of
`site-header` and `footer-theme-chrome` ahead of `site-footer`.

So the owner call the bug asks for was already answered by the data. The reason to
ship candidate 1 is not that it changes an answer — it is that today's right answer
is an **accident of alphabetical order, twice**, on a tie-break that knows nothing
about chrome.

This is the [UNVERIFIED]-marker lesson in the other direction: the bug file's table
was verified *when written* and went stale within hours, on a tree where another
lane was actively changing the very rows it measured. **A dated measurement is not
a durable one.** I re-ran rather than quoted, and that is the only reason the fix
went in today instead of waiting on an owner call that no longer applied.

## The `head` slot is what makes `eligible` a gate

`ResolveChromeComponent` **always** answers — deliberately, so a caller can report
a library gap rather than silently lose the slot (losing `head` also loses
`injectBrandHeadTags`: favicon, og-card). Live, the head pool has **no eligible
member** and its last-resort answer is `Document Head`, an 8,523-char
`component_level='section'` component.

So a straight swap of the lookup — which is how the fix candidate reads at a glance
— would have rendered a page section as `<head>`, **creating 167 on the one slot
that never had it**. The bug file says "the resolver already reports
`eligible=false` so `head` keeps falling through to `RenderFallbackHead` exactly as
today", and that is true *only if the caller reads the flag*. It is now pinned by
a test that fails if the ineligible component's body reaches the output.

## The guard, and why 118's could not see this

`TestNoChromeSelectionHandTypesItsOwnLookup` (118) scans for hand-typed SQL
(`function = 'site-header'`) and **skips `component_library.go` outright**. Both
exemptions are correct for what it guards, and both are exactly why it was blind
here: these three defects were **in that file** and contained **no SQL** — they
were Go calls handing a chrome function name to a section-shaped lookup.

The new scan is the complement: it matches the **call** form, covers
`component_library.go`, and was **proven in both directions** rather than assumed.
I wrote a temporary `zz_induce_167_tmp.go` reintroducing
`GetComponentByFunction(ctx, db, "site-header", logger)`:

```
--- FAIL: TestNoBuildPathResolvesChromeByPlainFunctionLookup
    chrome_build_path_test.go:280: chrome function resolved through a
    section-shaped lookup at [zz_induce_167_tmp.go:11]
```

then deleted it and confirmed green. A guard that has never been seen to fail is
a guard you are trusting on faith.

## MISSTEP — a compound command that reported success while skipping the check

Logged in `WRONG_CALLS.md`. I ran:

```
cd platform/orchestration/actions && gofmt -l <files>; echo "--- gofmt done ---"; go test ...
```

The `cd` **failed** (the shell was already in that directory from an earlier call,
so the relative path did not resolve). `&&` therefore skipped `gofmt` entirely —
but the `;` let the echo and the test run, and the output read:

```
--- gofmt done (empty=clean) ---
ok  	github.com/gqls/agentchassis/platform/orchestration/actions
```

**An empty gofmt result and a skipped gofmt are the same bytes.** I had one line
of evidence that said "formatting is clean" and it was produced by a command that
never ran. Caught only because the *next* command failed loudly on the same stale
cwd. The check: put the assertion in the command that must not be skipped, use
absolute paths, and never let `;` follow a `&&` chain whose failure you would not
notice.

## The tree is shared, and it broke under me mid-task

`go test ./platform/orchestration/actions/` passed, then minutes later failed with
nine `undefined: uuid / context / sql / zap / json` errors in **`prune_floor.go`**
— a file I had not touched. Not mine: a third session is extending it (the 135/165
lane) and has added DB code without the imports yet. HEAD's version is the
SQL-free "pure functions of counts" file and compiles.

So I verified the way the shared tree demands: `git archive HEAD` into a clean
directory, overlay **only my two files**, build and run the **full** package suite.

```
BUILD OK
ok  	github.com/gqls/agentchassis/platform/orchestration/actions	0.300s
```

That is the only run that means anything here — a green local build is not a green
HEAD, and a red local build is not necessarily your fault either.

## Pattern-check flags on the commit, and what I did about them

Three `untouched-twin` flags: `RenderHeader`/`RenderFooter`/`RenderHead` changed
without their `RenderFallback*` twins. **Deliberate, and I should have said so in
the commit message** — I did not, and forward-only forbids an amend, so it is
recorded here instead. The fallbacks are not defective siblings carrying the same
bug; they are the **destination** of the new gate. The whole point of the change is
that an ineligible library now reaches them.

One `logged-model-output` flag at `component_library.go:477`. Pre-existing, in
`GetComponentWithFallback`, untouched by me — the checker flags the file because I
changed it. Not adopted into this task; noted so the next reader does not re-derive
it.

One `register-blind-spot`: new workstream directory unknown to the concept
register. Answered — the guard is registered (see the register entry).

## Status at handoff

- Fix **committed** at `8b29404d6`, with `Council-Submitted: d73a4b06-a190-426e-bdf7-18d830d06a9d`.
- **NOT YET LIVE.** Go changes are inert until an image is rebuilt and rolled.
  Another session has `IMAGE_TAG` at `v1.0.1224` uncommitted in the tree and is
  mid-build; since `make build-*` builds from committed HEAD, this fix rides their
  roll. **Do not trust that sentence** — verify at the pod:
  ```
  kubectl exec -n ai-persona-system <chassis-pod> -- sh -c \
    'strings /app/agent-chassis | grep -c "no eligible header component in the library"'
  ```
  with a positive control in the same exec (a roll is not evidence your fix
  shipped — `bugs_open/153`).

## Council round 1 — REVISE, and it was a good REVISE

Verdict `d73a4b06-a190-426e-bdf7-18d830d06a9d`, 21:31. 10 reviewers, 7 abstained
(relevance filter), gating objection from `bug_historian`. Five approve
(`constitution`, `mission`, `architecture` — `point_fix`, fails every `needs_rfc`
trigger — `reuse_agent`, `render_guardian`, `prior_art_librarian`).

**Two substantive objections, and both were worth the round.**

**(1) The cache question — raised INDEPENDENTLY by five seats** (`guardian` high
×2, `debug_historian` high, `editquality` medium ×2, `render_guardian` low,
`prior_art_librarian` as a missing item). All of them found the same landmine:
`RenderHeader`/`RenderFooter`/`RenderHead` are named verbatim in *"correcting the
SELECTION does not repoint the 11 sites"*, whose correction cites
`site_components.build_status` and `renderAndStoreSiteComponent`'s `!force`
idempotence exit. Their question: can this fix reach served output at all, or does
a cached `site_components` row swallow it?

**Answered by measurement, and the answer is that the landmine's FOOTPRINT matched
while its MECHANISM did not.**

```
grep -c site_components component_library.go          -> comment lines only
grep -c force            component_library.go          -> 0
grep -c site_components  render_site_components_action.go -> 22
```

`component_library.go` never touches the `site_components` table and has no
idempotence exit. The callers do `html = InjectHeader(ctx, ..., html, ...)` —
chrome injected into a page HTML **string**, recomputed on every build and
rerender. The cached path is `renderAndStoreSiteComponent`, a different function
in a different file, which already uses `ResolveChromeComponent` and is 166's lane.

**The seats were right to demand the check and I was wrong not to pre-empt it.**
The submission never mentioned `site_components` at all, and `debug_historian` put
it exactly: *"the plan's otherwise-excellent measured-live-state discipline makes
the omission of this specific documented trap more conspicuous, not less."* The
lesson is not "the reviewers misread the landmine" — it is that **a landmine whose
footprint names your symbols obliges you to say why it does not apply**, even when
it doesn't. Silence reads as not having looked.

**(2) The gating objection (`bug_historian`, high) — the fourth path.** That
`bugs_open/170` is *currently manifesting* on three deployed sites and I was
shipping with no fail-loud guard on it. Filing honestly ≠ guarding. Its own
recommendation was to escalate to a human rather than footnote it.

**Acted on.** `reportIneligibleChromePin` logs at ERROR when a collection pins
chrome to an ineligible component, and **deliberately does not repair** — repointing
moves three live sites' markup, which is the owner call 170 exists to ask. A test
pins that: make it repair and it goes red.

### The subtle bit: the pin predicate is NOT the pool predicate

`chromePinEligibleSQL` omits `forked_from IS NULL`. That clause is right for pool
*selection* (an active fork of one client's header must not become a candidate for
every other site — 118's `header-leopardess` finding) and **wrong for a pin**,
because naming a site's own fork is what a pin is *for*.

Measured over all four live pins — and this is the row that proves it, because it
is the only one where the two predicates disagree:

| domain | pinned | active | fork | pin predicate | pool predicate |
|---|---|---|---|---|---|
| ai-agent-orchestration.com | header-professional-dark | f | f | **false** | false |
| finetuning.uk | header-professional-dark | f | f | **false** | false |
| gaswholesalers.com | header-professional-dark | f | f | **false** | false |
| leopardessconsulting.co.uk | header-leopardess | t | **t** | **true** | **false** |

Copying `chromeEligibleSQL` would have made the guard's **first live output a false
positive against the one correct pin in the fleet**, while still catching the three
real ones — the "a detector is blind to the spellings it doesn't search, so verify
its first result by the method it replaces" failure, in the other direction.

### `debug_historian`'s mutation objection was also right

It cited the `WRONG_CALLS` precedent that *a mutant which breaks the build prints
the same FAIL as a mutant that was caught*, and noted my "proven in both
directions" was narration with no evidence the mutant compiled. Re-run with the two
separated:

```
--- 2a. does the mutant COMPILE?
MUTANT COMPILES CLEANLY -- so a FAIL below is an assertion, not a build break
go vet: only load_component_library_actions.go:207 unreachable code  (pre-existing, unrelated)
--- 2b. now the guard
--- FAIL: TestNoBuildPathResolvesChromeByPlainFunctionLookup
    chrome function resolved through a section-shaped lookup at [zz_induce_167_tmp.go:10]
--- 2c. mutant removed
ok
```

`reuse_agent`'s low objection (two coexisting chrome scans) is recorded as a
lockstep hazard in `LANDMINES.md` rather than merged — merging would mean the 118
scan losing the `component_library.go` exemption that makes it precise.

Round 2 resubmitted on the same trail correlation (`RESUBMIT_CORR`), run
`c018787b-9871-4cf2-a25f-88f53f927c6a`.

## Council round 2 — REVISE again, and the right answer was to DELETE what I added

Gated by `reuse_agent` (high). **Not one seat objected to edits 1–3 — the actual 167
fix.** Every high/medium objection landed on **edit 4**, the pin diagnostic I added
in round 1 to satisfy round 1's gating objection. Four seats, four angles, same
conclusion:

- `reuse_agent` (high, gating): a bespoke `zap.Logger` reporter invented without
  checking the existing `deactivated_component` work-item pipeline.
- `bug_historian` (medium): a log is not durable or queryable — the
  `bugs_open/071`/`083` "detected then discarded" shape. **I named that anti-pattern
  in my own risk section and shipped it anyway**, which is the part worth sitting with.
- `guardian` (medium): ERROR on every render of three sites **indefinitely**, for a
  condition already filed and deliberately unrepaired. Alert noise.
- `architecture` (medium): the only gate on that path was a DEBUG-swallowed
  diagnostic, and it introduced a **second bespoke eligibility predicate**.

### Then I checked the prior art I should have checked before writing it

```sql
-- deactivated_site_components, discovery_checks/check_integrity.go:165-170
FROM site_components sc
JOIN content_components cc ON cc.id = sc.component_id
WHERE sc.site_id = $1 AND cc.is_active = false
```

**It joins `site_components` only. It never looks at `style_collections`.** The
detector for this exact class exists, fires, and is simply **blind to pins** — which
is precisely why three deployed sites have had no work item and no finding for it.

That is a much better finding than the log was, and it is now the headline of
`bugs_open/170` with the extension written up as candidate 1b, including the three
traps: the `item_key` would collide with `deactivated_%s`; `handler_agent:
rerender-pages` must **not** be reused because it re-renders and cannot repoint a
`style_collections` row, so the item would be unsatisfiable by construction —
**`bugs_open/166` reproduced on a new item type**; and `verifier_coverage_test.go`'s
"all carry `component_id`" contract needs re-running.

I did not implement it: `discovery_checks/` is the checker-layer lane's subsystem
(`bugs_open/149`) and `verifier_coverage_test.go` was **dirty in another session's
tree** while I was reading it.

**So edit 4 was removed rather than rate-limited** (rate-limiting answers only
`guardian`, not the other three). Net −174/+72.

### The lesson, and it is about round 1 as much as round 2

Round 1's gating objection was "you are shipping with this exposure unguarded". I
reached for the nearest thing that looked like a guard, and produced one that could
not repair, had no reader, and would fire for ever. **A gating objection asking for a
guard is not satisfied by anything that emits.** The seats that rejected it were the
same body that demanded it — which is the system working, not contradicting itself.

### Two more objections, answered rather than actioned

`prior_art_librarian` (high) and `editquality` (medium): I cleared the round-1 cache
question by measuring the **file** and never reading the same-day **CORRECTION**
landmine. Read in full now. Its footprint is `site_components.build_status`,
`renderAndStoreSiteComponent`'s `!force` exit, `rendered_html`, and the
`refresh_site_components` conditional on `rerender-pages` — **every symbol in the
stored-artefact path, none in `component_library.go`**. Its point 3 is that the
`!force` exit tests whether a slot holds *bytes*, never whether it holds the *right
component's* bytes. It is an account of why a corrected **assignment** does not reach
the page, and this change touches no assignments. It sharpens the round-2 answer
rather than revising it.

`debug_historian` (medium): the `git archive HEAD` verification tree has a
tmpfs-fill landmine — a full `/tmp` yields a *successful-looking* command. Checked
this time before trusting the green: `df -h /tmp` → 16G total, **3.9G available**,
immediately before the extraction. Its second point (no deploy-verification step) is
correct and is recorded in the RUNBOOK and at the top of the closed bug file, not in
the plan.

Round 3 submitted on the same trail (`3a3d4378-ebd6-4b99-9055-f88d9c031dc1`) — a
removal, leaving only the three edits that were never objected to.

## Council round 3 — APPROVED, on a removal

`d73a4b06-a190-426e-bdf7-18d830d06a9d`, 22:02 UTC 2026-07-31. **11 of 13 seats
approve; 2 advisory objections, none high.** The three seats that gated or objected
hardest in earlier rounds — `reuse_agent` (r2's gating seat), `prior_art_librarian`
(r2 high) and `architecture` (r2 medium ×2) — all flipped to approve.

**The approved change was smaller than the rejected one.** Rounds 1→2 added code;
round 3 deleted it. Net across the trail: the fix that shipped is the fix that was
submitted in round 1, and the two rounds in between were spent adding a guard and
then removing it. That is not wasted — the removal came with the
`deactivated_site_components`-is-blind-to-pins finding, which is worth more than the
guard was — but it is worth saying plainly that **my two attempts to satisfy a
gating objection with code were both wrong, and the thing that satisfied it was
evidence.**

### The two advisory objections stand, and I am not answering them away

`bug_historian` (medium ×2), paraphrased: removing the diagnostic returns the pin
path to **zero** signal — a code comment and a filed bug — and candidate 1b is
specified but not implemented, so three deployed sites stay in the state with
nothing watching. It names the 016b §9 shape *"One call site of a shared judgement
gets the rigorous fix; the sibling stays heuristic."*

**Both are true.** The counter-argument is on the record in the submission and in
`bugs_open/170`, and it is a judgement, not a refutation: a report nobody reads is
not a signal; `discovery_checks/` is the `bugs_open/149` lane's subsystem and its
coverage test was dirty in another session's tree; and an item routed to
`rerender-pages` would be unsatisfiable by construction, i.e. `bugs_open/166` on a
new item type. A future reader should weigh those, not inherit them.

### The council cited a landmine I wrote four hours earlier

`debug_historian` (medium) objects that round 3 "never engages" the landmine naming
*"two guard scans with disjoint blind spots and shared vocabulary"* across
`chrome_build_path_test.go` and `chrome_selection_test.go`.

**That landmine is mine.** I appended it to `LANDMINES.md` and ran
`landmines-sync.py --apply` earlier in this same session, which pushed it into
`doc_notes`, which is where the council seats read from. So the corpus round-tripped
my own warning back at me as an objection inside the hour.

Two things follow. First, **the sync works and the latency is under an hour** —
useful to know, and not previously written down anywhere I could find. Second, the
objection is technically right and I am leaving it: the lockstep hazard is real, it
is recorded where a session will meet it, and merging the two scans would cost the
118 scan the `component_library.go` exemption that makes it precise. Recording a
hazard is the mitigation here; there is no code that would improve it.

### Verdict trail, for anyone reading the correlation

| round | decision | gated by | what changed |
|---|---|---|---|
| 1 | REVISE | `bug_historian` (high) | — (fourth path unguarded) |
| 2 | REVISE | `reuse_agent` (high) | +per-render diagnostic |
| 3 | **APPROVED** | — (2 advisory) | −that diagnostic, +the prior-art finding |
