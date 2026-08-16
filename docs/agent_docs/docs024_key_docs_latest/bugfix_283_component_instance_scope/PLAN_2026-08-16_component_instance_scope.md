# PLAN — component instance scope (`bugs_open/283`)

Started 2026-08-16 (the lane ran one session before this directory existed; §1 reconstructs it).
Decisions and their reasons live here. Corrections stay visible.

---

## 1. Where this came from

An HTML `id` must be unique per document — `getElementById` returns the **first** match and
silently ignores the rest. Our components hardcode their ids as literal text, so two calculators on
one page both claim `loanAmount`, and the second one reads the first one's inputs while writing
into the first one's results. **It renders perfectly and answers with a number computed from
values the visitor never entered.** On a consumer-credit site that is the reason this is a bug
file and not a note.

Measured 2026-08-15: 173 of 240 active components hardcode at least one element id (166
literal-only), 100 bind by `getElementById`, and `btn-calculate` is shared by **nine** different
calculators — so "list all the calculators on one page" collides between *different* components,
not merely between two copies of one.

**Owner ruling (2026-08-15):** candidates A+C — namespace properly and block the bad state —
because *"if we chose to list all the calculators on one page we'd hope it would work"*. Reuse is
to be a genuine property of the platform, not a thing that happens to work for prose components.

## 2. Phasing

| phase | what | state |
|---|---|---|
| 1 | a per-instance template value on every render path | **done, live** (`v1.0.1304`) |
| 2 | a detector for the three collision classes + an opt-in guard | **done, live, guard OFF** |
| 3 | one canonical rule + a control that does not go stale | **done, live, council APPROVED** |
| 4 | convert the 22 calculator templates | **not started — architecture-scope, see §5** |
| 5 | arm `enforce_instance_scope` | **not started, and must come after 4** |

Phases 1–3 are inert in production: **0 of 243 active templates reference `{{.InstanceID}}`.**
Approval is not the same as the defect being fixed. 283 stays OPEN.

## 3. Decisions, and why

### 3.1 The token is component FUNCTION + OCCURRENCE — not position, not `data_uuid`

> **CORRECTION 2026-08-16 to the 08-15 design.** Phase 1 shipped `c<position>`, chosen because
> position is unique per page and stable across re-renders. That reasoning was correct and
> answered the wrong question. Measured after committing: the LMC tool slot sits at **position 0
> on 7 pages and position 1 on the other 16**, so one component answers to two different ids
> depending on the page. Superseded before any template consumed it.

The deciding question is not *which candidate is most unique* — it is *what does a selector have
to know*.

| candidate | unique within a page | same across pages | verdict |
|---|---|---|---|
| `position` | yes (measured, zero duplicates fleet-wide) | **no** | rejected |
| `page_components.data_uuid` | **provably** (1,580/1,580 distinct) | **no**, by construction | rejected |
| function + occurrence | derived | **yes**, for a component appearing once | **chosen** |

`oracle.py` addresses all 170 of its checks by literal CSS id. Under either rejected candidate it
needs per-page knowledge of every tool; under this rule, one prefix per tool.

**The cost, stated:** uniqueness is *derived* from the page's ordered section list rather than read
off a unique column. `DetectInstanceCollisions` is what pays for that. Anyone who thinks the trade
is wrong should say so — `data_uuid` is one query change away, and the council was told as much.

### 3.2 One rule, one derivation — and the single-section paths feed the same rule

Phase 1 shipped a second helper, `InstanceTokenFromSlot`, for the paths that render one section and
cannot see the page. The council's `reuse_agent` seat objected that this wrote **the same key under
a weaker guarantee**, reproducing the `{{.ComponentID}}` trap under a new name. It was right, and
the shape is doubly instructive: that is the trap *this lane wrote a landmine about*, recreated one
day later inside the change written to fix it.

Resolution: the helper is **deleted**. `RenderComponentAction` and the section editor supply
**occurrence 0** to the one rule. That is a possibly-wrong *input*, not a second guarantee — and
where it is wrong the instances take the same token and **collide**, which is detectable. An empty
token is not.

### 3.3 The control is mechanical, because two censuses went stale inside one council round

The council's `bug_historian` named five files as unguarded render call sites; four call no
`RenderTemplate*` helper at all. My own census missed `cmd/component-render-check` because it
grepped `platform/` and `internal/`. **A list of call sites is the wrong deliverable.** So:

- `RenderTemplateReportingMissing` logs at Error when a template needs the token and none is bound.
  It **reports and does not substitute** — this layer cannot see the page, so an invented token
  would either collide or *disagree* with the token the page's other paths use for the same
  instance, which is worse than empty.
- `scripts/pattern-check.py`'s `check_unscoped_component_render` fires on any changed non-test `.go`
  calling a `RenderTemplate*` helper that binds neither seam, unless allow-listed with a **measured**
  reason. Proven on the motivating case: 4 findings at HEAD, 0 after.

### 3.4 `{{.ComponentID}}` is deliberately NOT changed

Re-pointing it moves the served element ids of five live components — a change to what a shared
mechanism *guarantees*, which is architecture-scope under the owner ruling of 2026-07-29 §1. Filed
as **`architecture_review/RFC_032`**.

> **CORRECTION 2026-08-16:** for one round this deferral was asserted as "filed as the follow-up the
> architecture seat asked for" when **nothing had been filed**. The `reuse_agent` seat caught it.
> A deferral without a locator is not a tracked deferral.

## 4. What was deliberately not built

- **The general `missingkey=zero` fix.** Go blanks *any* absent field in *every* template,
  silently. This lane adds a report for one field name. The general fix is a fleet-wide change to
  every render and is not this bug's to make — recorded as a known-unresolved root cause
  (case file §10.4), not as a risk bullet, at the `bug_historian` seat's request.
- **A narrowed guard armed today.** It would fire on nothing (no interactive component is
  instantiated twice anywhere), and a mechanism rotting unexercised is the failure mode the owner's
  2026-07-29 ruling warns about. Convert → re-measure → arm.
> **CORRECTION 2026-08-16, later the same day:** the third item below said the RFC_022 expiry
> trigger was deliberately not built. It **is** built — see §4a. It was reclassified once it became
> clear the estate already had the exact idiom for it (twelve daily check CronJobs), so "not built"
> was measuring my assumption about cost rather than the cost.

## 4a. The tripwire that IS built

**`instance-token-adoption-check`** — daily CronJob, 07:40 UTC, deployed and proven (first run
2026-08-16 15:29 UTC). Counts active components referencing `{{.InstanceID}}`: **0 = the RFC_022
exception still holds; non-zero = it has expired and `RFC_032` is owed a round.**

Three things about it are worth carrying to any similar check:

1. **It could not be a commit-time lint**, which is what the architecture seat suggested. An
   `html_template` is written by the component-creator agent, by hand-authored SQL, by migrations
   and by the admin UI — four routes, none through a commit.
2. **Its healthy answer is ZERO, so it carries a demand control.** A broken query, a mis-escaped
   `LIKE` and an empty table all return zero too. It counts `{{.ComponentID}}` through the same
   `LIKE` in the same statement and **refuses** if that returns 0.
3. **Its polarity is inverted vs every sibling check**, and the report says so in its own words: a
   trip is not a defect, it is an owed review. Retire the job once it trips.

Owed: a `deploy-instance-token-adoption-check` makefile target (not added, because `makefile`
carried another session's uncommitted changes and a pathspec commit would have taken them).

## 5. ⚠ The next phase is architecture-scope, and the approval says so

The `architecture` seat approved round 2 under RFC_022's narrow exception, whose third condition is
*zero live consumers*. Its note:

> *"The moment the 22 templates start consuming `InstanceID`, condition 3 of the exception stops
> holding and this becomes a real load-bearing contract across the component library. That
> conversion PR, not this one, is where an RFC or at minimum a fresh architecture pass belongs."*

So phase 4 begins with `RFC_032` or a fresh architecture round — **not** with a template edit.
