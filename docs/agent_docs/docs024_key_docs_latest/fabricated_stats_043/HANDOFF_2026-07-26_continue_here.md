# HANDOFF — fabricated-stats lane (bug 043), 2026-07-26 evening

**Read this first, then `bugs_closed/043_…generated_page_copy_invents_quantitative_claims.md`.**
Resolve `043` **by slug** — the other `043` is the diagnosis route-hang.

Nothing here is blocked or half-finished. The bug is closed and proven. This document
exists because the session ran long, and because three things are *owed* rather than
broken.

---

## 1. State in one paragraph

Bug 043 (generated page copy invents statistics) is **CLOSED and live**, all four fix
candidates shipped, and the chain is proven end-to-end on production traffic. Bug 073
(the honest-empty-stat build failure) is **CLOSED** on the same fix. The chassis is at
**v1.0.1171**, pod-verified as carrying every Go change including the two made after the
council review. One follow-on gap is tracked as **`bugs_open/093`**. The council gate
verdict is **REVISE at round 5** and **no `Council-Reviewed:` trailer exists anywhere in
this lane** — deliberately.

## 2. What shipped, and where it lives

| artefact | what it does | state |
|---|---|---|
| `sql_for_agents/217_stat_values_optional_and_template_gated.sql` | candidate 1: 80 stat fields optional across 10 components, 46 `{{if}}` template gates, `e.g. '2.4M'` invention seeds stripped, `component-creator` NUMERIC FIELDS RULE, writer prompt names the optional case | **live**, config |
| `sql_for_agents/218_evidence_facts_for_043_sites.sql` | real `facts[]` for robot-hands, gamesdesign, ai-agent-orchestration | **live**, config |
| `sql_for_agents/219_page_build_declares_sections_metadata.sql` | `require_sections_metadata` on page-build-handler | **live**, config |
| `sql_for_agents/219b_content_reviewer_declares_sections_metadata.sql` | same on content-reviewer (dormancy is not closure) | **live**, config |
| `datahelpers/claims_stats.go` | candidates 2+4: `StatClaim`, `ExtractStatClaims`, `ScanStatClaims`, `LintStatUnits` | **live** in v1.0.1171 |
| `actions/validate_page_content_stats.go` | check 9 in the build gate, + the `stat_audit_unavailable` warning | **live** in v1.0.1171 |
| `fabricated_stats_043/SQL_2026-07-26_aao_enterprise_reference_deployment.sql` | de-fabricated the live residual the 07-24 sweep missed | applied; **page needs a re-render to publish it** |
| `fabricated_stats_043/SQL_2026-07-26b_writer_block_volatile_figures.sql` | removed falling figures from aao's prose block | **live**, config |

**Pod verification (the house rule — never git):**
```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'stat_audit_unavailable'"
# 1  — discriminating marker: a string THIS change created
```

## 3. The three things owed

> **UPDATE 2026-07-26 later — (a) is BUILT, (c) is MEASURED. Read § 7 at the foot of this
> file before acting on the three items below.** (a)'s candidate (1) is committed in
> `72effdbca` and **inert until the next chassis roll**, so `093` stays OPEN; (c) is now
> measured across twelve sites rather than the one instance named here. (b) is untouched.

**(a) `bugs_open/093` — the re-render path is unguarded.** The stat audit is generic but
has one call site (the build gate). The re-render path renders stored `content_data` with
no audit, so a persisted junk suffix would republish for ever. **Live exposure is nil by
measurement** (the fleet-wide suffix sweep returns five rows, all legitimate tool units),
but the structure is real. **The council raised this at HIGH in round 1 and again in round
5, and said plainly that deferring it is what blocks approval.** Its candidate (1) — extend
`discovery_checks/check_unverified_claims.go` to sweep `content_data`, reusing the
`claims_unverified` item type so there is no `verifier_coverage_test.go` obligation — is
the shape to build. This was deferred as a judgement, not an oversight: new feature work on
live production code at the end of a long session is worse than a tracked gap at zero
exposure.

**(b) Evidence registers for the remaining publishing sites.** Owner asked for all of them.
Done: robot-hands, gamesdesign, ai-agent-orchestration (mig 218) + the pre-existing
leopardess / relojistas / fundamentallyai. **Outstanding:** webdesign.co.uk (98 live pages),
finetuning.uk (41), gaswholesalers.com (28), idea.uk (20), vetcomparison.uk (5),
dartsonline.com (4), plus `facts[]` for vonc and oufe. The 17 `pool-*.internal` rows have
zero deployed pages and need none. **Three want their owning thread consulted first:** oufe
runs a stricter no-figures rail, vetcomparison carries a legal record from published
fabricated prices, idea.uk is owned elsewhere.

**(c) Spec-aspect numeric drift.** A number in a `briefing`/`identity`/`site_plan` spec is
an *instruction* to the writer and nothing refreshes it. leopardess `identity` says "143
agent definitions, of which 56 are active"; live is **175 active**. Wave-2d's parting
lesson, still unfixed, now measured fleet-wide.

Smaller: `aao/enterprise-reference-deployment` needs a re-render to publish its corrected
stats; `043`'s point (c) (partial blanking) is deliberately unimplemented and needs its own
function over raw `content_data`.

## 4. The council trail — read before resubmitting

Submission correlation **`569241fb-dd8d-4bcf-b382-234dfca1365c`**, five rounds, ending
**REVISE**. Resubmit on the same correlation so the trail accumulates:
`RESUBMIT_CORR=569241fb-… ./…/097_TRIGGER_council_review_v1.sh <submission.json>`

Two things a resubmitter must know:

1. **Migration `219b` answers round 5's MEDIUM and has never been put to the gate** — it
   was applied after the last submission. So the *current* state is better than the last
   verdict reflects.
2. **Expect the HIGH to stand until 093's candidate (1) is built.** The council said so
   twice, escalating. Do not resubmit hoping prose will settle it; it already refused that.

**Every objection it raised was correct.** Round 1 found two real defects in code that was
already live. Rounds 2–4 found my *description* overstating sound engineering — stale
sketches, a remediation claimed as done that was neither done nor necessary, and "both
objections answered in code" when one was coded and one filed. Round 5 returned to
substance.

## 5. Landmines — each of these cost this lane real time

- **`ParseEvidenceBase` returns nil when a row has no `facts[]` AND no `banned_claims[]`.**
  A `writer_block`-only row therefore switches **both** claims checkers off silently, while
  the writer_block keeps working (the prompt reads `site_specs` directly). That is how three
  sites were "protected" for two days with the checkers blind. **Verify each half against
  its own consumer — a green writer says nothing about a gate.**
- **Never set `writer_block_managed: true`.** `composeWriterBlock` emits only NUMBERS /
  CAPABILITIES / NAMED ENTITIES — no NEVER-STATE section — so managed regeneration silently
  deletes the ban lists.
- **A figure may live in the writer's prose block only if it CANNOT FALL.** Facts
  auto-refresh; a hand-managed block does not. A rolling-window or reaped figure will drift
  out of support and the gate will correctly reject the page. This bit me the same evening
  (1834 → 1790).
- **A page-rerender is NOT a rebuild**, though it stamps `page_components.updated_at` and
  `build_status`. Query `orchestration_states.owner_agent_type` for the window when claiming
  an agent ran. A fresh timestamp witnesses a touch, not an author.
- **`%stat%` is not a safe field predicate** — it also catches `availability_status` and
  `empty_state_label`. Enumerate.
- **`NULL || jsonb` is NULL**, and `jsonb_set(x, path, NULL)` returns NULL, so a merge-patch
  on a renamed field nulls the component's whole `input_schema`. Guard with
  `AND input_schema #> path IS NOT NULL`.
- **Gate a table's `<tr>`, never its `<td>`** — hiding a cell under a fixed `<thead>` shifts
  every later column left.
- **A shell `until [ "$(query)" -lt N ]` loop exits when the query fails**, because
  `[ "" -lt N ]` is a usage error and `until` reads non-zero as satisfied. Bind, default,
  and require non-empty. This nearly made me report a verdict that did not exist.

## 6. Where the rest of the record is

- `bugs_closed/043_…` — the account of record, including § Final verification and § Post-close
  deployment note.
- `bugs_closed/073_…` — closed by another thread on this fix; carries a correction and a
  counter-correction worth reading as a worked example.
- `bugs_open/093_…` — the tracked gap, with the council's escalation quoted.
- `NOTES_fabricated_stats_043.md` — the technical log, including three missteps.
- `README_where_we_are.md` — the owner's plain-prose history of the evening.
- `WRONG_CALLS.md` — four entries from this session, of which the one worth reading is
  *"three consecutive council rounds caught my prose overstating sound code"*.

---

## 7. Update, 2026-07-26 later — what the next thread actually picks up

**(a) is BUILT.** `bugs_open/093` candidate (1) — the stat audit's second call site — is
committed in `72effdbca`, with candidate (3) measured and found to need nothing. Read the
"Update 2026-07-26 (later)" section in `bugs_open/093` for the full account. It **stays
OPEN**: the code is inert until the next chassis roll, and the bar is *fixed AND live*.

Three things carry forward from it:

1. **The roll, then the verification.** Do **not** grade it on a green build — the build
   path is the one that already works. `093` § "How to verify a fix" is explicit: re-render
   a page **without** a writer pass and confirm the finding is raised.
2. **Council round 6 was submitted** on the same correlation
   `569241fb-dd8d-4bcf-b382-234dfca1365c`, answering the HIGH that had blocked rounds 1–5.
   **No `Council-Reviewed:` trailer is on any commit in this lane, and that is correct** —
   the trailer is earned by an APPROVED verdict only.
   > **Landmine for whoever chases that verdict:** the `097` trigger still publishes via
   > `printf | kubectl run -i --rm … kcat -P`, the pattern that silently drops messages
   > with exit 0 and a printed correlation id. A missing orchestration row is *usually*
   > queue latency (~16–30 min, and the lane had a visible backlog at submit time) — so do
   > not retry on that evidence alone, it costs a duplicate round — but do not assume the
   > publish landed either. Confirm by payload:
   > `SELECT correlation_id, current_step, status FROM orchestration_states WHERE
   >  collected_data->'input_data'->>'fix_correlation_id' = '569241fb-…';`
3. **Expect ~9 human-review items on the first live sweep**, across the sites, all
   HITL-terminal. That is the intended backlog of unsupported published figures, not a
   fault — but it lands on the queue `bugs_open/033` says has no working surface.

**(b) is UNTOUCHED.** The evidence registers for the remaining publishing sites are exactly
as § 3(b) describes them. One thing to know that was not known when that was written: with
(a) live, adding a register **row** to a site now switches the stat scan ON for it. A row
carrying real `facts[]` is what you want. A `writer_block`-only row will report every figure
on the site at `low` with the gap named — which is the designed behaviour and not a
regression, but it means (b) is no longer free of consequences on the review queue.

**(c) is MEASURED, and it is twelve sites, not one.** The leopardess instance named in
§ 3(c) is not the worst. `ai-agent-orchestration.com`'s writer instructions are internally
contradictory — `identity` and `briefing` say **"Over 70 specialised AI agents organised
into 8 departments"** in one place and **"30+ agent types"** in another, while live is
**176 active agent definitions**. And "organised into 8 departments" is the same family as
the claim audited out of leopardess as a fabrication: here it sits in the instruction that
tells a writer to say it. The query and the full extract are in
`NOTES_fabricated_stats_043.md`.

**[UNRESOLVED — needs an owner ruling, not a query]** "specialised AI agents", "agent types"
and rows in `agent_definitions` are three different units. Which one the sites should claim
is an editorial decision about what the business says it is; it cannot be looked up, and
picking one silently would be exactly the failure this whole lane exists to stop.

---

## 8. Update, 2026-07-27 — (a) is LIVE, (b) is largely FALSIFIED, and where it stands now

**(a) IS LIVE.** `v1.0.1172` carries the `093` second call site (pod-verified 2026-07-27:
`turn this into a check rather than a list` → 1, `scanStoredStatClaims` → 2, positive
control `turn this into a gate` → 1). **Council round 6 came back REVISE** on correlation
`569241fb-…` — 5 approve, 5 object, no veto, `decided_by: prior_art_librarian`, and
`unreadable: ["review_editquality.result"]`, so it was decided a seat short. The gating
objection is *"unverifiable-from-here rather than false"*: attach the queries, not just
their results. `guardian` stated its approval condition outright and **it is now satisfied**
— the caller sets are enumerated in `NOTES`, including the fourth `ParseEvidenceBase`
consumer the submission had not named (`validate_page_content.go:976`, check 8) and why its
nil-skip is **correct** rather than a fourth instance of the landmine. Round 7 is owed and
mostly written; two seats also asked for real changes (see NOTES).

**(b) IS LARGELY FALSIFIED — do not just work the list in § 3(b).** Surveyed 2026-07-27 with
the real scanner against an empty register, so every business-shaped number surfaces:

| site | components | number claims | verdict |
|---|---|---|---|
| gaswholesalers.com | 102 | 0 | **needs no register** |
| dartsonline.com | 17 | 1 | a 30-day returns window — a policy term. **Needs none** |
| finetuning.uk | 139 | 5 | 4 audience descriptors + a privacy age limit; **1 real, owner ruling owed** |
| webdesign.co.uk | 101 | 15 | **all 15 false positives** — see `bugs_open/102` |
| vonc.com | 49 | 0 in prose, 14 in stat fields | **DONE** — migrations 228 + 229 |

- **`bugs_open/102` filed** (new): the claims layer is `page_type`-blind. All 15
  webdesign.co.uk hits are worked examples on six `page_type='guide'` pages. **It blocks
  covering the estate's largest site**, because switching the register on would raise 15
  correct-copy items into the queue `033` says has no working surface.
- **Owner ruling owed on finetuning.uk**: the home page publishes *"Facilities management
  company (Midlands, UK) — ~80% reduction in quote preparation time"* with nothing on
  record behind it. Register it as a fact if it is real and provable; otherwise it comes
  off the page. **Not touched.**
- **vonc.com DONE.** `228` seeds facts[] (8 archetypes / 3 tools / 2 guides / 18 pages),
  every value a `pages` row-count with its proving query attached, merged by `jsonb_set` so
  the 9 banned_claims and all 7 keys survive. Measured before/after with the shipping
  extractor: **14 findings all `low` → 2 findings both `medium` → 0** after `229` fixed the
  real defect (about/`content-block-about` had Archetypes and Tools Live transposed; three
  other components on the site agreed with the DB, that one did not). Fleet register
  findings **21 → 8**.
  > **`229` edited `content_data` ONLY.** vonc `/about` still SERVES the wrong figures until
  > it is re-rendered. That is deliberate, and re-rendering it is the smallest outstanding
  > task in this lane.

**(c) unchanged** from § 7 — measured, twelve sites, needs an owner ruling on units before
anything is edited.

### The cheapest next actions, in order

1. **Re-render vonc `/about`** to publish `229`. Nothing else in the lane is closer to done.
2. **Run the discovery audit** and confirm `claims_unverified` items now carry
   `source: content_data` findings — that is `093`'s own verification bar, and `093` cannot
   close until it is met. **Do not grade it on a green build.**
3. **Council round 7**, attaching the caller table from `NOTES` as checks.
4. The finetuning and units rulings — both yours, neither guessable.
