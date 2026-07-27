# 093 — the stat-claim audit is generic but has exactly ONE guarded call site; the re-render path publishes unchecked

**Filed 2026-07-26** by the bugfix-043 thread, **at the council gate's insistence and to its
credit**: this was named in my own submission's `risks` block as a known gap and I closed
`043` without filing it. Seat `bug_historian` gated the review on precisely that —
*"documents this gap accurately but does not propose closing it or even filing it as a
tracked follow-up"* (submission `569241fb`, severity **high**). It was right, and the
pattern it matched is a real one in `016b`: a mechanism made generic, then guarded at one
call site while the others stay open.

**Status:** OPEN, not started. Cause is structural and fully known — this is a scope gap,
not a mystery, so there is nothing to diagnose.

> **THE COUNCIL ESCALATED THIS, TWICE, AND WILL NOT APPROVE THE PARENT CHANGE WHILE IT IS
> DEFERRED.** Submission `569241fb` ran five rounds. Round 1 raised this at **high**; by
> round 5, after every other objection had been answered, it came back at **high** again and
> sharper: *"the gap itself is the exact documented pattern this council exists to catch, and
> it is being deferred rather than closed."* It named the family explicitly — the same shape
> as the `missingkey=zero` case (one guarded call site, root behaviour unpatched, every other
> call site retaining identical exposure) and as the rebuild path silently dropping content
> the regression guard never saw.
>
> It also accepted, in the same breath, that deferral is *defensible* on the measurement: a
> read-only sweep of every persisted `*_suffix`/`_unit`/`_units` value fleet-wide returns
> **five rows, all legitimate tool units** (leopardess ROI estimator: "employees", "hrs /
> week per person", "per hour", "time saved"; robot-hands cycle-time: "seconds per cycle").
> None is a junk placeholder — the `%`/`ms` instances `043` recorded were cleared on
> 2026-07-22 — and `LintStatUnits` would be silent on all five, since none carries a
> magnitude-marker value and none of those words is in its dimensional-suffix map.
>
> **So: live exposure today is nil, and the structural gap is real.** Those are both true and
> the second is why this file exists. Whoever picks it up should know the review verdict is
> already written — closing this is what unblocks an APPROVED verdict on the parent, and
> candidate (1) below is the shape the council's own reasoning points at.
>
> **Deliberately NOT done in the parent change**, and the reason is a judgement rather than an
> oversight: implementing the discovery-check extension is new feature work on live production
> code, and it was reached at the end of a long session after the parent bug was already
> closed on independent end-to-end evidence. Hasty work here would be a worse outcome than a
> tracked gap with a measured exposure of zero. The parent (`bugs_closed/043`) does not depend
> on it.
**Belongs to:** the fabricated-stats lane, `bugs_closed/043_…generated_page_copy_invents_
quantitative_claims.md` (resolve **by slug** — `043` is one of the ambiguous numbers).

---

## The gap

`datahelpers.ExtractStatClaims` / `ScanStatClaims` / `LintStatUnits` are generic: give them
a component name and a `content_data` map and they will audit the figures in it. Live in
`v1.0.1170`. But they are invoked from exactly one place —
`validate_page_content` check 9 (`platform/orchestration/actions/validate_page_content_stats.go`),
which runs only on the **build** path (`page-build-handler` → `validate_content` →
`save_sections`).

**The re-render path never runs that gate.** `rerender_page_sections_action.go` renders
from stored `content_data` with no LLM and no validation step
(`docs/agent_docs/sql_for_agents/034_page_rerender_agent.sql` has no `validate_content`).
So:

- a figure that was already stored, however it got there, is re-published on every
  re-render without ever being compared to the site's register;
- `043`'s own update (b) class — a junk unit suffix persisted into `content_data` — is
  **specifically** unreachable by a build-time gate, because the suffix is a `source:static`
  fallback that gets resolved and persisted, and a page carrying one may be re-rendered for
  years without a writer pass.

This is not hypothetical for this lane. `043` recorded four pages across three sites
carrying persisted junk suffixes, and the counter-correction in `bugs_closed/073` measured
`aao/index` being **re-rendered three times in one morning** while the writer path was
blocked — i.e. the fabrication republishing itself indefinitely through exactly this route.

## Why it is not simply "add the gate to the re-render path"

The re-render path is deliberately cheap and LLM-free; bolting the full
`validate_page_content` onto it would drag in seven other checks, several of which
(placeholder scan, meta-commentary, contamination) are about freshly *generated* prose and
have nothing to say about a carried-forward render. It would also give the re-render path a
new way to fail, on content it did not author — the same trap as `073`, where a page became
unbuildable for telling the truth.

## Fix candidates (unranked)

1. **Extend the post-deploy discovery check instead.**
   `discovery_checks/check_unverified_claims.go` already sweeps deployed pages and already
   reads `page_components`; adding a `content_data` pass there covers the re-render path,
   hand-edits, and every page that predates the gate, in one place. **It reuses the existing
   `claims_unverified` item type, so it incurs no `verifier_coverage_test.go` obligation** —
   the same reason check 9 was cheap. Cheapest real coverage, and the natural second call
   site. Note it detects after deploy rather than preventing.
2. **A minimal stat-only gate on the re-render path** — call `LintStatUnits` and
   `ScanStatClaims` from `rerender_page_sections_action.go` and route findings to a work
   item rather than failing the render. Prevents rather than detects, but adds a failure
   surface to a path chosen for having none. If taken, findings must NOT block the render.
3. **Do it at the source instead**: `LintStatUnits` needs no evidence base, so a one-off
   fleet sweep over stored `content_data` would clear the persisted junk suffixes that make
   the re-render path dangerous, after which (1) alone is sufficient upkeep.

**Prefer (1) + (3).** (3) is a query and an UPDATE and closes the known instances; (1)
stops the class recurring without giving a deliberately-simple path a new way to fail.

## How to verify a fix

Do **not** grade this on a build. Take a page with a persisted junk suffix (`043`'s update
(b) lists four), re-render it *without* a writer pass, and confirm the finding is raised:

```sql
-- pages whose stored content_data carries a magnitude marker followed by a unit
SELECT s.domain, p.name, e.k, e.v
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id,
LATERAL jsonb_each_text(pc.content_data) e(k,v)
WHERE e.k ~ '_suffix$' AND e.v <> '' ORDER BY 1,2;
```

Then confirm a `claims_unverified` (or equivalent) item exists for that page afterwards.
A green build proves nothing here — **the build path is the one that already works.**

## Related

- `bugs_closed/043` — the parent case; its § "Final verification" is explicit that one page
  on one site is not the fleet sweep.
- `bugs_closed/073` — where the re-render-vs-rebuild distinction was measured, after one
  thread mistook a re-render for a build and another corrected it.

---

## Update 2026-07-26 (post-close) — the sibling fixes are now live; this gap is unchanged

`v1.0.1171` rolled after the parent case closed and carries the council's round-1 MEDIUM
fix plus the `require_sections_metadata` reader (pod-verified; markers listed in
`bugs_closed/043` § Post-close deployment note). Both `page-build-handler` and
`content-reviewer` now declare the key and the binary reads it, so a *skipped* audit is
visible.

**None of that touches this bug.** The re-render path still renders stored `content_data`
with no stat audit at all — there is nothing to skip, because the check is never reached.
Live exposure remains nil by measurement (the fleet-wide suffix sweep: five rows, all
legitimate tool units), and the structural gap remains open.

Also unchanged: the council trail ends at **round 5 / REVISE**. Migration `219b`, which
answers that round's MEDIUM, was applied *after* the last submission and has never been put
to the gate. A thread picking this up can resubmit on correlation
`569241fb-dd8d-4bcf-b382-234dfca1365c` — but should expect the HIGH objection to stand
until this file's candidate (1) is actually built, because the council said so twice.

---

## Update 2026-07-26 (later) — candidate (1) is BUILT and candidate (3) needed nothing

**Still OPEN, deliberately.** The code is committed (`72effdbca`) and **inert until the next
chassis roll**; the bar in CLAUDE.md is *fixed AND live*, and until the image ships the
defect is still reproducible. Do not close it on the commit.

### What was built — candidate (1), as this file recommended

`platform/orchestration/actions/discovery_checks/check_unverified_claims_stats.go` (new)
plus the wiring in `check_unverified_claims.go`. Stored `content_data` is now audited
alongside `rendered_html`, on **both** `page_components` and `site_components`, reusing the
existing `claims_unverified` item type — so `verifier_coverage_test.go` has nothing to say,
exactly as this file predicted.

Candidate (2) was **not** taken, for the reason in § "Why it is not simply…": the re-render
path was chosen for having no failure surface, and giving it one — over content it did not
author — is `bugs_closed/073`'s defect, not a fix for it.

**Two scopes, because the two scans need different things.** This is the part to review:

| scan | needs a register? | predicate |
|---|---|---|
| `LintStatUnits` | no — compares a component against **itself** | runs **fleet-wide**, as it already does in the build gate |
| `ScanStatClaims` | yes | the site_specs **row EXISTS** — *not* `ParseEvidenceBase` returning non-nil |

The second row is this lane's central landmine made structural. `ParseEvidenceBase` returns
nil for a row holding only a `writer_block`, so any consumer keying on nil switches itself
**off** on a site that explicitly opted **in** — which is how three sites sat "protected"
for two days with both checkers blind. A row with no `facts[]` now grades **low** with the
gap named in the finding's own `reason`, because a severity must never mean *"we could not
check this"*. `TestWriterBlockOnlyRowStillAuditsStats` fails loudly if that nil contract
ever changes, rather than passing on a premise that has moved.

The fleet-wide scope of the unit lint is not decoration: **finetuning.uk (3 stat fields, 2
pages) and idea.uk (2 fields, 1 page) have no `evidence_base` row at all**, so gating it on
opt-in would have left them unaudited on *both* paths.

### Candidate (3) — measured, and there was nothing to clear

> **The prediction in this file held exactly.** The fleet-wide sweep of every persisted
> `*_suffix`/`_unit`/`_units` value returns **five rows, all legitimate tool units**
> (leopardess ROI estimator: "employees", "hrs / week per person", "per hour", "time saved";
> robot-hands cycle-time: "seconds per cycle"). None appears in `statDimensionalSuffixes`,
> so `LintStatUnits` is silent on all five. **No UPDATE was needed and none was made.**

### The measurement, run with the SHIPPING code — not a SQL approximation of it

A throwaway harness ran `ExtractStatClaims` + `LintStatUnits` + `ScanStatClaims` over every
unlocked `page_components.content_data` row in production, with each site's real
`evidence_base`. This matters: a SQL predicate can count fields, but only the extractor
decides what pairs with what and what is dropped.

```
components with stat claims: 24      pages affected: 18
stat claims extracted: 61            unit-lint findings: 0
register findings: 21   (across 9 pages)
```

So the first live run raises **9 work items**, all HITL-terminal. `vonc.com` alone accounts
for 14 of the 21 findings, all at `low`, purely because it registered `banned_claims` but no
`facts` — and its own output is worth reading, because two of them **contradict each other**:
`index` publishes "Archetypes 8" / "Tools Live 3" while `about` publishes "Archetypes 3" /
"Tools Live 8". The check found a real defect on its first pass.

### Two live defects this sweep found in code that had ALREADY shipped

Both were reaching the **build gate at `error` severity** on sites with registered facts —
i.e. both would make a deployed page unbuildable, which is `bugs_closed/073`'s shape on a
new trigger. Both are fixed in the same commit; both are strictly **narrowing**, so neither
can create a finding, only remove one.

1. **A typographic range escaped the composite-token exclusion the hyphen form has always
   had.** `unitSuffixRe` three lines above already spells `[-–]`, so typographic dashes were
   plainly meant to be in scope — but the adjacency test beside it is **byte-level**, and an
   en-dash is three bytes, so `next == '-'` cannot see it. Live instance:
   fundamentallyai.com (15 registered facts) publishes `Read time: 8–12 minutes`.
   *The trade, stated because it is real:* a range is now excluded **entirely**, so
   "2–3 million users" is no longer examined. That is not NEW blindness — "2-3 million
   users" has always been treated that way — but it is a coverage limit.
2. **Zero-padded display ordinals** (`01`, `02`, `03`) were extracted as published figures.
   vonc.com's about page numbers its process steps that way. Fixed in `ExtractStatClaims`,
   **not** in the shared `isExcludedNumber`, because it is a property of a bare stat *field
   value* whereas those exclusions reason about a number's position inside a *prose block* —
   widening them would change what the prose scan sees on every site for a shape only the
   stat path can produce. `0` alone is deliberately still a claim: zero is a real count, and
   an honest one, which is the whole subject of `073`.

**The transferable pattern behind both** — a shared predicate written for one input shape,
reused on another. `isExcludedNumber`'s rules were written to reason about a number's
position inside a prose block; `ScanStatClaims` hands it a bare field value as the "block",
so "list ordinal at block start" can never fire (it requires a following `.` or `)`) and the
byte-level dash test silently mismatches. Added to `016b` §9.

### What is still owed on this file

- **The roll**, then the verification in § "How to verify a fix" — re-render a page
  **without** a writer pass and confirm the finding is raised. A green build proves nothing
  here; the build path is the one that already works.
- **Council round 6**, submitted on correlation `569241fb-dd8d-4bcf-b382-234dfca1365c`.
  **No `Council-Reviewed:` trailer exists on `72effdbca`, and that is correct** — the
  trailer is earned by an APPROVED verdict only, and a verdict that post-dates its commit
  can never carry one.
- Candidate (2) remains unbuilt and remains the only thing that would *prevent* rather than
  *detect*. So does `043`'s point (c) — a partially-blanked stat block reads as CHECKED
  while carrying a surviving invention, and it needs its own function over raw
  `content_data`, because `ExtractStatClaims` drops blank sentinels by design.

### The pod-grep marker for this change — and the vacuous one to avoid

The obvious marker is **vacuous here, by construction**. The new code deliberately mirrors
check 9's wording (that parity is the point), so the phrase you would naturally reach for
already exists in the live binary and greps `1` before anything has shipped:

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
# ✗ VACUOUS — matches the ALREADY-LIVE check 9 string, not this change
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'no machine-readable facts\[\]'"
```

Use a string only this change creates, with a **negative and a positive control** — the
positive control is what proves the grep itself works and that you are reading the binary
you think you are:

```bash
# ✓ NEW code only — 0 before the roll, 1 after
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'turn this into a check rather than a list'"
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'scanStoredStatClaims'"
# ✓ POSITIVE CONTROL — check 9's old wording, live since v1.0.1171, must stay 1
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'turn this into a gate'"
```

Measured on `v1.0.1171` at 21:44 on 2026-07-26: `0`, `0`, `1` — i.e. this change is
confirmed **not** live, which is the state this file records.

---

## Update 2026-07-27 — LIVE in v1.0.1172, and the re-render path is TWO paths, not one

**The fix is live.** Pod-verified on `v1.0.1172` with a discriminating marker and a positive
control (see § "The pod-grep marker" above): `turn this into a check rather than a list` → 1,
`scanStoredStatClaims` → 2, control `turn this into a gate` → 1.

**But this file's own framing was too simple, and I found out by using the path.** It says the
re-render path "renders from stored `content_data`". That is true of one of its two modes. The
`page-rerender` agent branches first:

```
check_rerender_mode:
  condition: input_data.spec.reason == 'image_landed'
          OR input_data.spec.reason == 'section_data_resolved'
          OR input_data.spec.reason == 'cta_links_stale'
  then_step: rerender_sections   -- re-renders each section FROM content_data
  else_step: render_page         -- ASSEMBLE-ONLY: reuses the stored section HTML
```

`else_step` — the default for **any unrecognised reason** — never reads `content_data` at all.
It re-assembles the page from each component's previously-rendered HTML, deploys it, and
reports `COMPLETED`. Observed directly on 2026-07-27: a page whose stored content had been
corrected was re-published still carrying the old figure, with a `complete` work item and a
`COMPLETED` orchestration. The tell was `page_components.updated_at` holding the timestamp of
the *content edit* rather than of the render.

**What this does and does not mean for this bug.**

- It is **not** a coverage hole today. The post-deploy audit reads `rendered_html` as well as
  `content_data`, so an assemble-only republish of stale HTML is still scanned — by the older
  half of the check, on the surface that actually shipped.
- It **is** a correction to the mental model this file sold, and the correction matters for
  anyone extending the work: *the audited artefact and the published artefact are not always
  the same object.* A fix aimed only at `content_data` can be bypassed by a publish path that
  never consults it.
- It sharpens candidate (2). A stat lint on the re-render path would have to sit **after the
  mode branch**, or on the render output, or it will simply not run on the assemble-only
  half — which is the same one-guarded-call-site shape this bug is about, one level in.

**Practical note for whoever verifies this bug**, because it cost three attempts today: to
exercise the section path you must set `spec.reason` to one of the three recognised values.
`spec.reason` looks like free-text provenance and is control flow; vary `item_key` for dedup,
never the reason. And insert work items as `status='triaged'`, not `'detected'` — `detected`
is a queue with no consumer (`bugs_open/083`).
