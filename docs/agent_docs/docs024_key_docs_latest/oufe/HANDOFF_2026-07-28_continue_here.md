# HANDOFF — oufe.com / oxenunity.com, 2026-07-28 evening

**Cold-start entry point. Supersedes `HANDOFF_2026-07-27_continue_here.md`**, which
is now stale in several places (it still says "there is no chart renderer", which
is wrong and corrected in place there).

Read this, then `SUMMARY_2026-07-28b_oufe.md` for the plain-prose read-out, then
`NOTES_oufe.md` from the bottom up — the missteps are the part you cannot
rederive. `DESIGN_2026-07-28_premise_branching_and_deepthink.md` holds the next
big design decision.

---

## 1. State, verified 2026-07-28 evening

Chassis **`v1.0.1192`, TWO replicas** (was one — the replica-scaling work from
another thread has landed; if you see odd dispatch behaviour, that is new and not
yours). Series code pod-verified present in the running binary. Council lane
consuming at lag 0.

All eight pages 200:

| url | what is on it |
|---|---|
| `/` | banner, hero, brief-explanation, info-card-grid, CTA |
| `/about.html` | |
| `/cases/index.html` | |
| `/cases/thames-water.html` | **prose + mechanism diagram + TWO charts** |
| `/tools/tool-recovery-waterfall.html` | the tool, Tier-4 verified 13/13 |
| `/disclaimer.html` | sections E + F, live and locked |
| `/privacy.html` | live and locked |
| `/contact.html` | `mailto:` form |

Site id `a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39`. Evidence register: **32 facts**.
`claimscan` **0 findings across 17 components**. `scripts/render_audit.py`: **0 real
contrast failures** across all pages (the one remaining reading is an `overImage`
heuristic, screenshot-confirmed fine).

`oxenunity.com` unchanged, still a hand-authored one-pager.

## 2. What shipped today (all live, all verified)

1. **Council lane** — `bugs_open/096` **CLOSED** with a measured confirmation.
2. **Tier-4 acceptance** — failed 2/11, then **13/13 with the tool source
   untouched**. The fence was the defect. Filed `bugs_open/126`.
3. **Legal pages** — disclaimer + privacy, `rebuild_policy='owned'` +
   `lock_type='permanent'`, linked from the footer on every page.
4. **`mechanism-flow` component** (mig 247) + Part 26A cram-down on the Thames
   page (248). Seven steps, two branches, every step from a registered fact.
5. **Series facts** (`claims_series.go`) — `Observation{as_of,value,source}`.
6. **`evidence-timeseries` component** (mig 250) — **BUILT, STILL UNUSED.**
7. ~~`cmd/contrastscan`~~ — **built and then DELETED the same day**: it duplicated
   `scripts/render_audit.py`, which already existed and is better. Do not rebuild it.
8. **Two charts on the Thames page** — capital structure (251/252) and the Ofwat
   determination (254/255).
9. **Scanner fix** — written-out British dates ("28 July 2026") were read as
   business figures fleet-wide.

## 3. NEXT — in priority order

### 3a. Wire `scripts/render_audit.py` — the tool EXISTS, nothing runs it

> **CORRECTED, same evening.** This section used to say "wire `contrastscan`".
> `cmd/contrastscan` was a **duplicate I built without finding
> `scripts/render_audit.py`** (brochure workstream, 2026-07-27), because my
> prior-art grep was `--include=*.go` and the prior art is Python. **The Go tool
> has been deleted.** Do not rebuild it.

What is **already wired**: `check_palette_contrast` — a discovery check, DB-only,
microseconds, reading the *composed* palette. It is good and it states in its own
header what it cannot see.

What is **not wired**: `scripts/render_audit.py`, which renders every element in
headless Chromium and is the only thing catching 026 **family 3** — a component
hard-coding an ink over a themed fill. That is precisely the class that put an
unreadable chart on this site. Shape: `bugs_open/122` candidate 2. It needs a
browser, so it is not a DB-only discovery check; the natural home is the
**browser-runner-adapter, which already runs Playwright in-cluster**.

**Fleet audit run 2026-07-28 evening — 65 live contrast failures on 7 homepages:**

| site | failures | note |
|---|---|---|
| gamesdesign.co.uk | **30** | stat band, cyan-on-cyan |
| **idea.uk** | **18** | the site that took the first sale |
| vetcomparison.uk | 10 | |
| relojistas.com | 2 | |
| oufe.com | ~~5~~ **0** | fixed, see below |
| webdesign.co.uk | 0 | |
| leopardessconsulting.co.uk | 0 | |

Those belong to other workstreams — route, do not fix over the top.

#### Built 2026-07-28 evening — the audit is now DISPATCHABLE, not yet dispatched

Owner asked for the Python recreated in Go and wired in. Done as far as it can go
without an image:

| piece | where | state |
|---|---|---|
| the measurement | `internal/adapters/browserrunner/render_audit_action.go` | built, 6 tests green — **INERT until the adapter image ships** |
| adapter registration | `adapter.go` `case "render_audit"` | same |
| `browserPage.Evaluate` | `run_checks_action.go` + test fake | same |
| the caller | `platform/orchestration/actions/request_render_audit_action.go` | built — **INERT until the chassis image ships** |
| registry entry | `registry.go`, category `quality` | same |
| council | corr `4d6585dd-4678-4891-b9a9-7a37b32df1b6` | submitted |

#### RENDER AUDIT — STATE AS OF 2026-07-29 MORNING (read this first)

**Everything below the orchestration layer is proven. The top hop is not.**

| layer | state |
|---|---|
| `cmd`/measurement | `internal/adapters/browserrunner/render_audit_action.go`, 6 tests green |
| dedicated pod | `render-audit-adapter` LIVE, own topic/group/logs. **Proven end to end**: returned real measurements (nav elements, `rgba(255,255,255,0.9)`, ratio 3.54, `over_image:true`) |
| chassis action | `request_render_audit` + `system.adapter.render-audit.requests` **both pod-verified in `v1.0.1196`** |
| agent definition | `render-audit-agent` seeded (mig 256), row verified active/not-snapshot/not-deleted, 4 steps |
| **orchestration dispatch** | **UNRESOLVED — see below** |

**THE OPEN PROBLEM.** Dispatching `agent_type: render-audit-agent` on
`system.agent.generic.requests` (corr `17a23aee-a85a-44df-bd23-4a88bc869185`)
produced **no `orchestration_states` row**, while generic-lane lag sat at **0** —
so the message WAS consumed and no orchestration started. No error in the chassis
logs naming the agent or the correlation. Not the ~300s post-restart window (the
pod had been up for hours). The agent row is correct.

**Do NOT re-fire blindly** — that is the recorded 096 trap and it costs a
duplicate. Next diagnostic steps, cheapest first:

1. Compare against a KNOWN-GOOD dispatch of a different `agent_type` on the same
   lane, same envelope. If that also produces no row, the problem is the envelope
   or the lane, not this agent. **This is the untouched-peer check and it should
   be step one.**
2. `ensure_site_record` is the initial step and takes its input from
   `input_data`. Confirm it accepts `domain`/`site_id` the way this dispatch
   supplied them — a first step that rejects its input before the row is written
   would look exactly like this.
3. Check whether the chassis caches `agent_definitions` and needs a roll to see a
   newly seeded type. If so, that is worth writing down for every future seed.

**What is NOT the problem:** the pod, the topic, the action, the image, or the
agent row. All four were verified independently.

#### The audit has its OWN POD as of 2026-07-28 (owner ruling), and it is PROVEN LIVE

`render-audit-adapter` — a second Deployment of the **same browser-runner image**
with `REQUESTS_TOPIC`/`CONSUMER_GROUP` overridden. Costs a topic, not a binary.
Own pod, own consumer group, own logs, own failure state; higher memory ceiling
and a wider liveness tolerance because a whole-site audit is one long request.

**Proven end to end 2026-07-28 21:17**: a `render_audit` published to
`system.adapter.render-audit.requests` was consumed by that pod and returned real
measurements — actual nav elements, computed `rgba(255,255,255,0.9)`, ratio 3.54,
`over_image: true` (the guard working: those links sit on a gradient header, so
the backdrop is unknowable and the reading is flagged approximate rather than
asserted).

> **ROLLOUT TRAP, cost ~15 minutes.** The pod started BEFORE its topic existed,
> got no partition assignment, and sat idle while messages piled up — group
> registered, `--describe` empty, zero processing logs, no restarts, no errors.
> A `rollout restart` once the topic existed fixed it instantly. **Create the
> topic first, or restart the consumer after the first publish**, and do not read
> "group exists" as "group is consuming" — `--list` showed it while `--describe`
> showed nothing.
>
> Also: `kubectl run -i --rm ... | kcat -P` via stdin looked like it dropped the
> message (the recorded landmine) — it had NOT; my offset check was wrong.
> `kafka-get-offsets.sh` works; `GetOffsetShell` returned nothing silently.

**STILL REQUIRED before a workflow can use the dedicated pod:** the deployed
chassis (`v1.0.1194`) predates the topic change, so its `request_render_audit`
still publishes to the **shared** browser-runner topic. Verified by grepping the
running binary. Seeding a workflow today would route audits to the shared pod and
defeat the isolation. **Rebuild the chassis, pod-grep for
`system.adapter.render-audit.requests`, then seed.**

**THE LAST STEP IS DELIBERATELY NOT DONE, and doing it early would break.** No
workflow calls `request_render_audit` yet, because **a seed naming an
unregistered action fails at runtime** — image first, then seeds. The order is:

1. `make build-browser-runner-adapter` + `make build-agent-chassis` (bump
   `IMAGE_TAG`), push, deploy.
2. **Pod-grep both**, by a string the change CREATED — e.g.
   `strings /app/... | grep -c 'render_audit'` — and a positive control.
3. Only then seed a workflow step calling `request_render_audit`.
4. **Then retire `scripts/render_audit.py`**, which is marked superseded in its
   own header but deliberately kept: the Go is inert, and deleting the Python
   first would leave the fleet with no render audit at all. Expect the two to
   agree on a real site; if they disagree, the port is wrong.

Note for whoever seeds it: the adapter shares a pod with tool acceptance, and a
25-page audit is 25 sequential navigations. `bugs_open/096` was exactly a long job
blocking a shared lane — worth deciding whether this wants its own.

### 3b. More charts, tools and guides — the owner's standing ask
His words on seeing the first chart: *"this is starting to look great! We'll need
a lot more of this of course, and tools as is already in the plan."*
Reusable now: `evidence-chart` (magnitudes), `evidence-timeseries` (dated series,
unused), `mechanism-flow` (process), `tool-guide-intro` (**exists** — the answer
to "guides at each point", never used here).

### 3c. First real use of `evidence-timeseries`
It needs a measure that genuinely moves over time. Ofwat's performance commitments
(leakage, spills, sewer flooding) are stated as 2024-25 baseline → 2029-30 target
and may yield one. **Do not force it** — twice today the honest answer was that
the data was a comparison, not a trend.

### 3d. `bugs_open/126` has an inbound contribution from another thread
Another workstream audited `tool-improver.note_refusal` and found its error
handler disabled (`error_step_disabled_086`) as containment from bug 086. They
handed the decision here because `who-owns.py 126` names oufe. **Unread by this
thread** — read the CONTRIBUTION section at the foot of 126.

### 3e. Owed elsewhere
- `bugs_open/122` candidate 1: header chrome hardcodes `color: white` on the
  accent, failing **5 of 6 sites**. Needs a generator fix; chrome is a stored
  artefact so existing sites need `nav_drift` → `nav-updater` (`bugs_open/117`).
- robot-hands' primary CTA is **white on white, invisible**; vonc's Gauntlet
  buttons fail. Both owned by other threads — evidence is in 122, do not fix over
  the top of them.

## 4. Traps that will cost you time

**Rendering an OWNED page.** `save_page_sections` REFUSES `rebuild_policy='owned'`
pages, so `section_data_resolved` fails at `save_sections`. You need assemble-only
— and `TRIGGER_rerender_page.sh` **cannot request it**: `REASON="${3:-...}"` treats
an empty string as unset. Publish the envelope directly with `spec.reason` absent.
Recipe in `RUNBOOK_oufe.md` §9–10.

**Do NOT lock a component at insert time.** `save_page_sections` *preserves* locked
rows rather than rendering them, so a row locked with empty `rendered_html` renders
as **nothing, for ever**, and reports success. Write `rendered_html` in the same
statement (migration-182 pattern), by executing the component's OWN template.

**`$facts` is declared by the template**, not supplied by the engine
(`{{- $facts := .facts -}}`). An undeclared variable is a **parse error**, so the
component renders nothing. Same class: **`inc`/`add` are not in the funcmap** —
use a CSS counter.

**SVG text is invisible to the claims gate.** Charts must use real HTML text with
CSS-drawn furniture. Verified by decoding the payload claimscan received.

**`context_terms` must match the LABEL, not the prose.** A registered value gets
flagged as unregistered when the chart's terse label contains none of its terms.
Fix the label, not the guard.

**Contrast — ROOT CAUSE NOW FIXED.** `--color-primary` was `#1B2A3B`, *identical to
`--color-surface`*, so all 7 of its foreground uses were invisible or near it.
Three components were patched individually before anyone traced it to the palette.
Its 3 background uses all pair it with `--color-primary-text`, so the pair flips
together: now `#86ADDE` / `#0F1820`. oufe went **5 → 0** real failures.
`--color-card-bg` was also white on a dark site (fixed). Chart furniture needs the
**3.0** non-text threshold and `--color-border` scores 1.66 — use `--color-accent`.

**A `render_audit.py` reading marked "over an image — ratio approximate" is a
heuristic, not a verdict.** It assumes mid-grey for an unknown backdrop. oufe's
last "failure" was a white button over a near-black hero and is genuinely fine —
confirmed by screenshot.

**Supersede-then-insert must be SEQUENTIAL.** A CTE doing both fails: every branch
sees one snapshot, so the partial unique index still sees the old row (23505).

**403s are bot protection, not authentication.** A browser gets 200 where curl and
WebFetch get 403. `RUNBOOK_oufe.md` §12 has the recipe, including
`pdftotext -layout` and why a column mapping must be corroborated arithmetically.

**Catastrophic regex:** `grep -oE "[^.]{0,180}x[^.]{0,180}\."` over 68KB was
OOM-killed. Split in Python.

## 5. The council submission that is still unresolved

Series facts, correlation **`da40ddf0-7acd-40f6-9826-d9161f5601be`**.

- **Round 1 → REVISE**, and the objection was correct and valuable: `ValidateSeries`
  enforced the per-observation source rule but `numberSupported` never called it.
  **Fixed** — `observationHasResolvableSource` is now shared by both.
- **Round 2 → `complete_invalid` three times**, including a control run **without**
  `RESUBMIT_CORR`, which refutes the resubmission path as the cause. The `fix_plan`
  artifact persists each time, so structural validation passes and the failure is
  in a review or decide step. Evidence written into `bugs_open/119`.

**The code is committed, tested and live, and carries NO `Council-Reviewed:`
trailer.** Do not add one — it has not been approved.

**When you read a verdict, query by YOUR correlation.** The most recent
`council-gate` note was a different submission's APPROVED; reading the latest note
would have attached someone else's verdict to this plan.

## 6. Rails that must not be relaxed

- **No figure in any brief, spec, identity or content_direction.** Only the
  evidence register, with a source.
- **Never publish a figure about a real company without a source URL and capture
  date.** Today's additions each had their quote checked *literally present* in the
  fetched page — not taken from a model's summary of it, and not from a search
  result.
- **The grounded lane must keep its inability to publish.**
- **Section G (liability cap) stays parked** until something is for sale. It is the
  only item genuinely needing a solicitor.

## 7. The thing worth carrying forward

Nearly everything that went wrong today was **something this workstream had already
written down and then not checked**: a warning in a handoff ("no chart renderer" —
two were live), a hazard in a bug file (the white card, described as "waiting"), a
rule in a validator nothing called, and a checker whose scope we chose ourselves
and then trusted.

The sharpest form: **a measuring instrument you built yourself is the hardest to
distrust, because when it agrees with you it may only be agreeing with your blind
spot.** `contrastscan` reported the page clean because it only measured links and
buttons — a scope I chose, for reasons that felt sufficient.

The checks that actually caught things were dull: run the test, open a screenshot,
query the component library rather than the Go source, read `features_open/`, ask
by your own correlation. All of them are cheap. None of them are interesting. That
is why they get skipped.
