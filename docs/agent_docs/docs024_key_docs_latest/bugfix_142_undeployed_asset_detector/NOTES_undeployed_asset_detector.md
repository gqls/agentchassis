# NOTES — bugs_open/142, the undeployed_asset detector

Append-only, newest at the bottom. Missteps are the point of this file.

---

## 1. Picking the bug (2026-07-31 ~18:05)

Swept `bugs_open/` (68 files) against ownership. Method, because the obvious one
is not enough: `who-owns.py` reads COMMITS, so a session mid-fix is invisible to
it. So I also mapped the **331 live session transcripts** — for the 21 modified
in the last 12 hours, counted bug-number mentions in the last 400 lines, which
shows what each lane is doing *now* rather than what it did this morning.

Actively held right now: 072, 099, 104, 108, 125, 135, 138, 143, 144, 145, 149,
151, 153, 157, 159, 163. Not held: 142, among others.

Rejected candidates and why, so the next thread does not re-walk them:

- **033, 066, 075, 085, 113** — all "fixed, awaiting verification". Verification
  tasks, not fix tasks.
- **071** — genuinely open, but its own 07-27 triage says it now bundles one
  closed mechanism with three open ones and recommends a split. The split is the
  owning lane's call, not a passing thread's.
- **083** — the highest-leverage bug in the directory and explicitly blocked:
  *"Decision pending — do not act on this section until it is recorded here."*
  Owner call.
- **093** — "OPEN, not started" in its header, but its own last section says the
  fix shipped in v1.0.1172 and the file is now blocked on 083. Not a code task.
- **150** — good candidate, unowned; the fix is agent-definition config and
  verifying it needs a live `improvement-sweep` firing (~60:1 discovery-to-
  promotion blast radius). Left as the better pick for a thread that wants a
  config task.

**142 chosen**: unowned, Go code (so it rides the next chassis build), and the
defect generalises — "a detector whose denominator is the artefact table cannot
see a missing artefact" is a shape, not an instance.

## 2. The bug file's own numbers had moved — corrected before use

142 says the detector *"fired five times — every one for robot-hands.com"*.
That was true on 07-29 and is **not true now**. Live 2026-07-31:

```
undeployed_asset:  73 complete | 35 detected | 59 unresolved   (7 sites, not 1)
```

But the `complete` rows are the trap: the recent ones carry
`created_by = 'operator'` and `'session-2026-07-31-robot-hands-carousel'` —
**hand-filed by humans and sessions, not by the detector.** The detector's own
rows (`design-discovery-agent`, `generic`) are all `unresolved` or parked at
`detected`. So the headline count moved while the substance did not, and a
thread reading only the count would conclude the detector was working.

## 3. MISSTEP — I measured the logo by the wrong column

To decide whether emitting `needs_brand_head_assets` would route to a handler
that could act, I checked which sites have a logo:

```sql
bool_or(a.purpose='logo' AND a.status='active')   -- 4 of 15 sites
```

and briefly believed 11 of 15 sites could not be serviced. **Wrong column.**
`derive_brand_head_assets` keys on `asset_key`, not `purpose`:

```sql
WHERE a.site_id = $1 AND a.asset_key = 'logo' AND a.status = 'active'
```

Re-measured: **15 of 15 by `asset_key`, 4 of 15 by `purpose`.** Every site with
deployed pages can be serviced.

*What caught it:* opening the handler's actual query instead of inferring its
precondition from the column name I had already typed. The cheap check that
would have caught it sooner: read the consumer's WHERE clause before writing a
query that claims to predict it. Logged in WRONG_CALLS.

## 4. MISSTEP — my 090 symptom bundled two mechanisms, and the run paid for it

The diagnosis loop returned **NOT CONFIRMED (stopped: scope-not-narrowing)**,
last verdict **UNVERIFIABLE**. It did not refute anything; it ran out of scope
to narrow. Two causes, one of them mine:

1. **Mine.** The 090 authoring guidance says *"one coherent bug per run"* and I
   filed both mechanisms (wrong population AND wrong evidence table) in one
   symptom. Iterations 1–3 kept splitting attention between them.
2. **Structural.** A `needs_diagnosis` item anchors to `system.internal`, so the
   loop's `site_id`/`domain` arrive blank. Its iteration-1 note says so directly:
   *"the diagnosis target's site_id/domain are blank ('-')… orchestration_states
   returns no rows for this correlation/site."* It cannot probe a site over HTTP,
   and this bug's decisive evidence is a wire probe.

**The run was still worth its credits, and not for the reason I expected.**
It independently confirmed the static half — *"the check's population is bounded
to rows in `assets` and its deployed-test only ever queries page_components — 0
rows there ever match a favicon path"* — and it named exactly the right residual
doubt: *"it does NOT yet prove the assets are actually present… it is equally
consistent with those assets genuinely never having been deployed."* That is the
question my 15-site wire probe answers and it could not ask.

And one of its data requests returned the row that changed the fix — see §5.

## 5. The finding that changed the design, and it came from the loop's data dump

Iteration 2's bundle contained, in passing:

```
00ff3af5-… | favicon | image | active | /assets/images/input-data.asset-key.jpg
```

An **active favicon row whose url is an unresolved template literal.** My design
at that point used `EXISTS(active row for this purpose)` as deploy evidence.
Measured the exception properly:

```
gamesdesign.co.uk | favicon | 1 active row | 0 at the published path | /assets/images/input-data.asset-key.jpg
gamesdesign.co.uk | og_card | 1 active row | 0 at the published path | /assets/images/input-data.asset-key.jpg
robot-hands.com   | favicon | 1 active row | 0 at the published path | …
robot-hands.com   | og_card | 1 active row | 0 at the published path | …
```

Both sites serve favicon.png and og-card.png **200**. So:

- `EXISTS(any active row)` gives the right answer here **by luck** — the row is
  not a record of the published artefact at all.
- Tightening to `url = the published path` would be causally sound (that is the
  pair `recordDerivedAsset` writes) but would file *"has never been generated"*
  against two sites that serve the file. **A false claim, in a fix whose whole
  subject is a detector making false claims.**

Resolution: three states, not two. Row-at-published-path → deployed. Row with
another url → **observed in `Findings`, no work item**, because it is evidence of
neither deployment nor absence. No row → the gap. The middle state is named
`brand_head_provenance_url_unexpected` and points at `bugs_open/152`, which owns
the URL-rewrite defect that produces these.

## 6. The trap I nearly walked into: the LIKE underscore is load-bearing

`purpose` values contain underscores and SQL `LIKE` treats `_` as *any
character*. So the pattern built from `content_hero` matches the real published
filename `content-hero…`, which is spelled with a **hyphen**. Escaping the
underscore is the obvious correctness fix. Measured both forms:

```
purpose      | assets | deployed_UNESCAPED | deployed_ESCAPED
content_hero |     38 |                 38 |                0
```

**Escaping manufactures 38 false findings.** Left unescaped, documented at the
call site, and pinned by `TestUnderscoreWildcardIsLoadBearing`, which asserts the
raw concatenation survives — a test that merely *ran* the query would stay green
through the change.

Second trap of the same family, found before it bit: `site_components.build_status`
is `'rendered'` and **never** `'deployed'` (all 42 rows). The page-component
predicate five lines away *does* filter `build_status='deployed'`, so "make the
two consistent" is a silent way to blind the new head probe. Pinned by
`TestSiteComponentsAreNotFilteredOnDeployedStatus`.

## 7. Mutation-verified, because a green guard proves nothing

Reintroduced each defect in an isolated `git archive HEAD` tree and confirmed the
intended test fails, then restored and confirmed green:

| mutation | test that caught it |
|---|---|
| remove the brand-head exclusion | `TestBrandHeadPurposesAreExcludedFromThePageAssetQuery` |
| escape the LIKE underscore | `TestUnderscoreWildcardIsLoadBearing` |
| drift the map's og-card path | `TestBrandHeadAssetPathsMatchTheDeriver` |
| filter site_components on build_status | `TestSiteComponentsAreNotFilteredOnDeployedStatus` |

The isolated tree was not optional: the working tree **does not compile**
(`check_phantom_internal_links.go` and `check_dead_controls.go` carry another
session's in-flight edits — unused `strings` imports and an undefined
`datahelpers.HrefOffsets`). `go build ./platform/...` in the working tree is
therefore not a signal about my change either way.

## 8. Why the deriver was not edited

`storage.BrandHeadAssetPaths` duplicates two literals that
`derive_brand_head_assets_action.go` also spells, and the obvious tidy is to make
the deriver read the map. **Not done deliberately:** session `759437b9` had 49
transcript hits on that file at 18:48, working `bugs_open/143`. Two sessions in
one file is the one collision no hook can prevent, and a same-file passenger
rides whoever commits.

Instead `TestBrandHeadAssetPathsMatchTheDeriver` scans that file's
`recordDerivedAsset` call sites and fails the build on any divergence — including
`t.Fatalf` if it finds *no* call sites, so the sensor cannot go quietly vacuous.
That is stronger than the edit would have been, and the adoption is left to 143's
lane.

## 9. Result, simulated fleet-wide before shipping

| | before (shipping) | after |
|---|---|---|
| `undeployed_asset` (page assets) | 96 | 72 |
| `needs_brand_head_assets` | — (impossible) | **4**, on idea.uk + webdesign.co.uk only |
| brand-head false positives | **24**, on 12 sites | **0** |
| observations, no work item | — | 4, on gamesdesign + robot-hands |

96 → 72 + 4 exactly: the 24 removed are precisely 12 sites × 2 brand-head
purposes. The 2 sites that gain findings are the 2 that serve 404.

**Stated rather than hidden:** these findings land at `status='detected'`, the
queue `bugs_open/083` is about, which has no running promoter. This change makes
the detector correct. It does not make the queue drain, and it must not be
"fixed" by writing `triaged` — 083's own landmine says why.

## 10. Council gate — APPROVED at round 1, and it caught my own over-claim

Correlation `35d88a60-ec1c-4cd3-b69c-f2813c3e837f`. **APPROVED**, 7 advisory
objections, none high-severity, 4 seats abstained, 5 approved outright
(`adoption_guardian`, `improvement_guardian`, `render_guardian`, `constitution`,
`mission`).

**The one that mattered, and it is embarrassing in a useful way.** `bug_historian`
(medium) pointed out that my GAP summary said *"has never been generated"* and the
check cannot support that claim: `recordDerivedAsset` is **best-effort** — it
swallows failure as `logger.Warn("provenance upsert failed (non-fatal)")` and its
`ON CONFLICT` carries a lock guard — so the git commit can succeed with no row
written. The artefact is then live, serving 200, and my finding fires anyway.

**This is the same over-claim I had already refused, two branches earlier.** The
whole reason gamesdesign and robot-hands get an observation instead of a work item
is that asserting "never generated" about a serving site would be false. I caught
it in the branch and then wrote it into the headline sentence of the finding
itself. Fixed: the summary now asserts the absence of the **record** — the thing
actually measured — names the other reading, and says the remedy is the same
either way. The test now **fails** on `never been generated`, mutation-verified,
so it cannot return as a wording tidy.

Second actionable one, same seat (medium): the provenance observation rode
`Findings` only, and `Findings` ride `collected_data`, pruned at ~24h — a slow
silent drop, the retention trap already in WRONG_CALLS. Now also logged at `Warn`
with both paths and `owned_by=bugs_open/152`. Still deliberately not a work item:
the state is "cannot tell" and there is no handler for that.

**Three objections answered with evidence rather than code**, recorded here so the
answers are not lost with the run:

- `guidelines` (medium) — *"reuses a per-site `item_key` … `recurrenceExpected` is
  load-bearing … trips silently on the third detection cycle."* Right about the
  mechanism, wrong about this case. `WorkItemSpec` **has no such field** and
  `discovery_checks.go:225-240` sets it on none of them, so every discovery check
  already runs under the two-strike rule — this is not a new exposure. And it is
  the correct semantics: `needs_brand_head_assets` is a **detected defect**, not an
  action request. A successful handler run writes the asset row and the check stops
  firing, so a third detection genuinely means two attempts failed, which is what
  `unresolved` is for. `bugs_open/024`'s defect was the opposite case — a
  *completed* re-render request being read as a strike.
- `reuse_agent` + `architecture` (medium) — the map is not adopted by the literals
  it replaces. Correct and deliberate (§8). Now named as follow-up work in the
  concept register (IMG-066 `verify-later`) rather than left to arrive by
  drift-sensor pressure, which was `architecture`'s actual point.
- `debug_historian` (medium) — pod-grep the running binary post-deploy. Already
  RUNBOOK §R9 with a positive control; it was missing from the *submission*, not
  from the plan.

`prior_art_librarian` flagged two of my `grounded_in` quotes as unverified by it
(the `asset-deployer` sole-consumer claim and the `verifier_coverage_test.go`
line numbers). Both were measured by me before submitting — the queries are in
§1 and R5 — but the seat is right that it could not confirm them, and that is a
fair objection to how the evidence was presented, not to whether it exists.

## 11. Pod-verified NOT live — so 142 stays OPEN

`v1.0.1218` on both replicas (`agent-chassis-776f55c5f9-bjfhq`, `-g9vqc`):

```
brand_head_provenance_url_unexpected : 0   0
No asset record publishes            : 0   0
generated but not deployed to site   : 1   1   <- POSITIVE CONTROL
```

The control proves the grep works and that I am reading the binary I think I am.
So the fix is committed (`3b812161b`, `d671fb2b2`, `6f5e85886`) and **confirmed
inert**. The standing bar for `bugs_closed/` is *fixed AND live*; this is fixed
and not live, so the case stays open. A roll is not evidence either — `bugs_open/153`
— which is why this is a grep and not a tag comparison.

**I did not roll, deliberately.** Two `council-gate` runs belonging to another
session were mid-flight at 18:42 (`review_bug_historian`, `review_constitution`),
and a fleet roll kills an in-flight council. The working tree's `makefile` also
carries another session's uncommitted `IMAGE_TAG` bump to v1.0.1218, so editing it
is a same-file collision. Nothing about this fix is urgent enough to spend either.

**What the next roll owes:** RUNBOOK §R9 on every replica, then move the file to
`bugs_closed/`. No live firing is available to verify with — the check is reachable
only via `design-discovery-agent` ← `improvement-sweep`, disabled since 2026-05-02.

## 12. CLOSED — and the verification arrived on its own, which is worth recording

`v1.0.1219` rolled at 19:09Z (another session's build; my commits were in it).
Pod-grepped both replicas: all four new markers 1/1, both positive controls
intact. That alone only proves the binary carries the code — §11's whole point.

**Then the platform verified the behaviour for me.** A fleet discovery sweep ran
across 9 sites at 19:17–19:21Z. I did not fire it; every row carries
`created_by='generic'`, the chassis worker. I had been about to insert a one-shot
`scheduled_tasks` row to force exactly this (the `oneshot-design-discovery-rh-20260730`
pattern), and checking the baseline first is what showed it had already happened —
**the cheap check that saved a production config write was "look before you
write"**, which is the same move that caught the `input-data.asset-key` row in §5.

Both mechanisms proved, on the shipped binary:

- **Absence is visible.** idea.uk raised **2 × `needs_brand_head_assets`**
  (`…:favicon`, `…:og_card`) carrying this fix's summary wording verbatim. It is
  the only site that raised any, and it is one of the two serving 404. This site
  had produced **zero** findings in the detector's entire history.
- **The false positives are gone.** Eight of the nine swept sites hold active
  favicon + og_card rows — 16 permanent false positives under the old predicate.
  **16 would have fired; 0 did.**

That second number is the one I would quote, because it is a *counterfactual
measured on the same population in the same run*, not a before/after across two
different days — the failure mode §2 records, where the bug file's own figures had
drifted under it.

**Not exercised live, and I am not claiming otherwise:** the provenance-observation
branch (gamesdesign, robot-hands) and the second GAP site (webdesign.co.uk) were
not among the nine. The observation branch files nothing by construction, so there
is no row to look for even when it does run; both are covered by mutation-verified
unit tests only.

Moved to `bugs_closed/`.

## 13. MISSTEP — I wrote "16 false positives avoided" and the denominator was wrong

§12 above (and, for about ten minutes, the bug file, the README and concept
register IMG-066) said the live sweep proved **16 would have fired, 0 did**,
across **9 swept sites**.

**Nine sites got work items in that window. Only TWO ran this check.** The other
seven were served by `nav-updater`, `internal-link-resolver`, `page-build-handler`
and others, which never run `check_undeployed_assets` at all. I had counted
"sites with any work item in the window" as "sites that ran the check" — and the
first number was sitting right there in a query I had already run, which is what
made it feel safe.

The true, and still sufficient, result — from the two `design-discovery-agent`
orchestrations (`eafe7955` vonc.com 19:17:48Z, `3c0a2499` idea.uk 19:18:04Z):

| site | active brand-head rows | pre-fix | actual |
|---|---|---|---|
| **vonc.com** | 2 (favicon + og_card, both serving **200**) | **2 false positives** | **0 items raised, of any purpose** |
| **idea.uk** | 0 (serves **404**) | 0 (structurally invisible) | **2 × `needs_brand_head_assets`**, correct keys and wording |

So: **2 avoided on 1 site**, not 16 on 8. Both branches are still proven live —
the proof is just n=1 per branch, and it must not be quoted as a fleet figure.

*What caught it:* asking **which agent owned the orchestrations** before trusting
a count I had already written down. The cheap check is to join the counterfactual
to the runs of the check itself, never to "activity in the time window" —
a window is not a population. This is the wrong-denominator trap already logged
several times in `WRONG_CALLS.md`; knowing it by name did not stop me typing it.

**Second retraction in the same breath.** From "the check ran and no scheduled
task is enabled" I inferred *"`design-discovery-agent` evidently has a driver
that is not `improvement-sweep`"* — and briefly wrote that into IMG-066 as a
correction to the fleet's standing claim. **It does not follow.** Every
`scheduled_tasks` row targeting a discovery or improvement agent is
`enabled=false`, and orchestration naming does **not** discriminate: I checked
`build-pipeline-trigger`, which genuinely IS a scheduled task, and it uses the
identical `<agent>-orchestrate-MMDD-HHMM` form. All-history for
`design-discovery-agent` is **three** runs ever retained — robot-hands 07-30
(matching the disabled one-shot's `last_triggered_at` to the second) and today's
two. The overwhelmingly likely trigger is **another session firing by hand**,
which is exactly what the fleet's standing note says happens.

Marked `[UNVERIFIED]` rather than resolved, because resolving it needs the
dispatching session's transcript and it changes nothing about this fix. What it
would have changed, had it gone unretracted, is `bugs_open/083`'s and `093`'s
central premise — on the strength of someone else's write, read as a mechanism.
That is [[your-action-moves-you-to-the-back-of-the-selector]] one step removed:
not my write becoming my evidence, but another thread's.
