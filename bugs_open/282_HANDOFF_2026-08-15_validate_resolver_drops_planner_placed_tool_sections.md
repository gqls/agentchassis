# 282 — validate's name resolver silently drops every tool section the planner places (407's missing half)

**Filed:** 2026-08-15, loancalculator rebuild lane.
**Status:** OPEN — **FIXED, COUNCIL-APPROVED AND COMMITTED 2026-08-16, NOT YET LIVE.**
The Go half rides the next chassis roll; the config half (migration `439`) is
applied and inert until it does. Stays OPEN until a post-roll replan proves the
functions land — the bar is fixed AND live. Fixing lane:
`docs/agent_docs/docs024_key_docs_latest/bugfix_282_validate_accepts_menu/`
(PLAN/RUNBOOK/NOTES/README + the council submission). Council corr `bbf49822`:
REVISE → REVISE → **APPROVED** (3 advisories, none high). Both REVISE rounds
widened the fix and are summarised below — the reported bug was the smallest
part of what was wrong.
**Diagnosis trail:** 090 run `4a02a4e1-3972-450a-8163-28d6bb0a79fd` (verdict
UNVERIFIABLE at budget, disjunctive hypothesis + named next_scope) → next_scope
walked first-hand same session; the disjunction closed with the drop site cited.
Owner-ruling 2026-07-31 note: the 090 ran BEFORE this root cause was asserted;
this file is the completion of its own named evidence requests.

## Mechanism (the whole chain, each link cited)

1. **The menu shows the tools.** Migration 407 (PLAN-049): `load_components`
   includes `component_level='tool'` rows when the site's structure spec has
   `plan_includes_tools='true'` AND the component is deployed on that site.
   Proven at run corr `2f74a975-1a87-40a8-af88-a9bd2ecc1510`:
   `available_components`=151; diagnosis data_request `menu_has_tool1|menu_has_tool2
   = true|true`.
2. **The planner PLACES the tools.** `llm_call_log` id
   `ca3c22f4-5e4c-4ccf-9d4a-02719d976c8e` (`plan_site`, 2026-08-15 14:25) —
   raw response proposes `"sections": ["hero","tool-loan-repayment",…]` for
   index and `"hero","tool-<fn>","ported-prose","faq","tool-cta"` for tool
   pages (application-tracker, overpayment, settlement read directly; the
   response text is the artefact).
3. **Validate's resolver cannot resolve them.** `loadComponentNameResolver`
   (`platform/orchestration/actions/v3_site_actions.go:3804-3809`) builds
   `validFunctions` from `WHERE component_level IN ('section','element')` —
   tool-level functions are structurally absent, so `resolve()` returns
   ("", false) for every proposed tool section and the section is dropped from
   the plan write. No error, no tell; the write succeeds.
4. **The persisted plan lacks the tools.** Plan `dcbae4df` (2026-08-15
   14:25:49): 0 of the 12 locked tool functions appear in `site_plan_sections`;
   index reads `hero,info-card-grid,tool-list,guide-list,call-to-action`
   (5 sections — the raw response proposed 6).

The measured effect that surfaced it: a `recompose_pages` release of the 12
tool-carrying pages on loancalculator.co.uk (0162cde4…) produced compositions
that would build every calculator page WITHOUT its calculator. Nothing live was
damaged — the tool-role review gate (TP-004) + the 12 permanent locks held.

## Why this is 407's missing half, not a new seam

407 widened the MENU under an opt-in flag and added the placement gate; the
resolver that validates proposed names against reality was never widened to
match. Two lists that must agree, maintained in two places — the
`idx_swi_dedup ↔ workItemTerminalStatuses` lockstep class exactly.

## Fix candidates (ranked by what closes the door)

1. **Mirror the menu's own predicate in the resolver** (preferred): include
   tool-level functions in `validFunctions` for the site being planned iff the
   site's `plan_includes_tools` flag is on AND the component is deployed on
   that site — the same subquery `load_components` uses, factored into ONE
   shared helper so the two surfaces cannot drift again (the lockstep made
   structural). Needs the site_id plumbed to `loadComponentNameResolver`
   (callers have it).
2. Widen the resolver to all active tool-level functions unconditionally —
   smaller diff, but lets a planner place a tool the menu never showed it
   (hallucination path back open); rejected by the same reasoning that made
   407's menu opt-in and site-scoped.
3. Prompt-level workaround (tell the LLM to use a different marker) — rejected:
   markers invent a second vocabulary; the estate resolves REAL function names.

Whoever fixes: council-gate the change (platform code, shared seam — but note
RFC_022's narrowed trigger: this mirrors an existing opt-in whose consumers are
enumerable by the flag query). After it ships and a replan runs, verify at
`site_plan_sections` (the 12 functions present on their pages) and only then
work the held D2 tickets (11 `owned_page_review` + `needs_page:index` on
loancalculator — see the lane's HANDOFF 2026-08-15).

## ADDENDUM 2026-08-15 ~19:15Z — the drop has a SECOND ARM the fix must explain

The same plan (`dcbae4df`) carried `loans-credit-health-check` on
tool-credit-roadmap — a name matching NO component at all — and that name was
NOT dropped: it persisted into `site_plan_sections` AND spawned
`needs_new_component:loans-credit-health-check` (which then failed 3/3 at
`generate_template` on `stop_reason=max_tokens` — benign here, we do not want
the component, but the routing is the evidence). So an unknown name routes to
component-creation, while a KNOWN-but-tool-level name vanishes. The simple
"resolve fails ⇒ dropped" account in step 3 is therefore incomplete: the
tool-level names likely die in a branch where the unresolved-name handler finds
an EXISTING component row (level 'tool') and discards the section instead of
either placing it or minting a creation item. The fixing thread should locate
that branch precisely — it is probably the cleanest seam for the fix.
Side observation for the lane: the tool-credit-roadmap PAGE carries no locked
tool row (the 12 locks cover 11 other active pages + archived standard-calc),
which is presumably why the LLM reached for a new name there.

> **CORRECTED 2026-08-16 (bugfix_282 lane, at the fix): THERE IS NO SECOND ARM,
> and the premise of this addendum is false.** `loans-credit-health-check` is
> not "a name matching NO component at all" — it is an ordinary **section-level
> component**, and that is the whole of why it survived:
>
> ```sql
> SELECT id, "function", component_level, is_active, created_at
>   FROM content_components WHERE "function"='loans-credit-health-check';
> -- 824e3309-f90c-4aa9-b679-46f4a8722475 | loans-credit-health-check | section | t | 2026-08-13 14:19
> ```
>
> (Created 2026-08-13 for loanandmortgagecalculator.co.uk.) Arm 1 of `resolve()`
> — `validFunctions[raw]` — accepts it. The `needs_new_component` row cited as
> "the routing evidence" was filed by **`plan_sections`**, a different action on
> a different path, after validate had already passed the name through. So the
> asymmetry was never asymmetric: one name was section-level and passed, the
> others were tool-level and were dropped. **The "resolve fails ⇒ dropped"
> account in step 3 is complete and correct**, and there is no branch to locate.
>
> Cost of the error, had it been followed: an afternoon hunting a branch that
> does not exist in a 7,000-line file, and the fix aimed at the wrong seam. The
> check that dissolves it is one query on `component_level`, 0.2 s — now in the
> lane's RUNBOOK and in `WRONG_CALLS.md`.

## FIX AS BUILT — 2026-08-16, commit `5534e9f71` (bugfix_282 lane)

**Candidate 1 was adopted in goal and rejected in mechanism.** "Mirror the
menu's own predicate in the resolver" cannot be done honestly: the menu is not
Go. It is a SQL string in `agent_definitions`, and it had **already drifted past
407's text** — migration `419` (2026-08-15, the `bugs_open/276` family) added a
`requires-backend` clause to the same query and guards its own apply by
asserting 407's exact bytes. A Go mirror would have been a third
hand-maintained copy across the SQL/Go line: exactly the lockstep class this
bug's own §9 entry says to avoid ("single-sourcing is a guarantee, a lockstep
test is a backstop").

**What shipped instead: validate consumes the OFFER'S OUTPUT.** A new opt-in
step-config key, `menu_field`, names the collected-data path holding the menu
the planner was actually shown (`available_components`); those rows are UNIONed
into the resolver's valid set. One list, one source — a future gate on the menu
flows through with nothing to keep in step.

- Go: `platform/orchestration/actions/component_name_resolver_menu.go` (new) +
  two small hunks in `v3_site_actions.go`. **Inert until the next chassis roll.**
- Config: migration `439_validate_plan_accepts_the_planner_menu.sql`, **applied
  and recorded 2026-08-16**, on `build-site-planner.validate_plan` only. Its
  guard checks `menu_field` and `load_components.output_field` **together**.
- The shared resolver's query and signature are **unchanged**, so
  `apply_gap_plan`'s three call sites (content-gap-planner) keep the
  section/element-only menu 407 and PLAN-049 deliberately withheld from them.
- `site-planner` (the only other `validate_site_plan` consumer, 0 runs, its own
  menu section/element-only) was left alone — verified absent after apply.
- Order-safe both ways: unread key = today's behaviour; rolled Go without the
  key = today's behaviour.
- Tests: 9, incl. an un-opted-in negative control; **both arms mutated** to
  prove they can fail. The resolver had zero coverage before this.
- Council: submitted `bbf49822-6704-4802-b3b5-1afed6777c88` (advisory; commit
  carries `Council-Submitted:`).

### ROUND 2 — every drop is now a DURABLE record, not a Warn (`adb1ee2ad`)

The council returned **REVISE** on round 1 (8 of 12 seats approved; gated by
`editquality`), and `bug_historian`'s objection was right in a way that widened
the fix: accepting the menu restores acceptance **for one opted-in caller** and
leaves the generic mechanism untouched — a typo, a renamed function, a deleted
component, or any caller that never opts in still vanished with a `Warn` and no
durable trace, *byte-identical to the bug being fixed*.

So `ValidateSitePlanAction` now files one durable finding per dropped section
name — **for every caller, opted-in or not** — through `LogActionFindings` /
`agenterrors`, the door this same action already uses for
`recordRecomposeOutcomes` (reuse, not new machinery; its header carries the
standing reason that chassis logs rotate sub-second). `error_code`
`PLAN_SECTION_NAME_DROPPED`, severity `warning` — a drop IS a legal outcome, and
what it must not be is invisible. The remedy text differs by case: with no
`menu_field` the dropped name may be one the planner was legitimately OFFERED
(this bug's own shape), so the reader is pointed here before blaming the name.

**This is the part that generalises beyond 282**: the acceptance surface no
longer loses anything silently, whether or not anyone opted in. `site-planner`
and the three `apply_gap_plan` call sites keep their old *behaviour* by design,
but their exposure is now visible instead of implied.

Round 2 also fixed two holes the council did not find — in my own tests. A
mutation removing the drop-collection line left the whole suite green (shape
asserted, wiring not), and the "a clean plan writes nothing" control could not
fail, because a sqlmock fails an EXPECTED call that never came, never an
UNEXPECTED call that did. Both closed (wiring test; the recorder returns its
attempted count). Five mutations, all biting. Written up in `WRONG_CALLS.md`.

## How to verify a fix

Re-fire the lane's `phase2_recompose_26.sh` (12-page scope). Expect: the 12
locked tool functions in `site_plan_sections` on their own pages; zero
RECOMPOSE_INTENT_NOT_REALISED rows; locks 12/12 untouched. The negative control
already exists: plan `dcbae4df` is the no-fix baseline.

## Cross-reference — 2026-08-16 (bugfix_285_lock_blind_section_list lane)

`bugs_open/285` (section-list case) named this bug a "prerequisite or co-requisite" for its
loader-merge fix. **It is not, for the page-BUILD path:** `plan_sections` never calls
`loadComponentNameResolver` (callers: `ValidateSitePlanAction` `v3_site_actions.go:3407`,
`apply_gap_plan_action.go`); a locked tool slot merged by the loader (`7d9b7334a`, LOCK-008)
resolves via `plan_sections` Path 0 by stored identity and reaches save. This bug remains
exactly what it says — the RE-PLAN path drops planner-placed tool sections — and remains yours.
One interaction worth knowing when you fix it: after `7d9b7334a` rolls, `pages.sections` on the
12 loancalculator tool pages will carry the positional slot (`tool-2` etc.); a replan that names
the tool's FUNCTION pairs with that slot at merge time (function arm), so the plan's position
then wins over the exiled live position. Nothing here blocks or is blocked by 282.

### ROUND 3 — the record follows the RESOLVER, not the caller (`0b39f5a99`)

Round 2 returned **REVISE** again (9 of 13 seats approved), gated by
`bug_historian` on the exact continuation of its round-1 point: the durable
record was wired into `ValidateSitePlanAction` only, while
`apply_gap_plan_action.go`'s **three** call sites share the same resolver and
lost names just as silently. The scoping was worse than it looked —
content-gap-planner runs **116 orchestrations/30d** against build-site-planner's
handful, so the class fix had been aimed at the quieter path.

**All four drop sites now record.** The three gap-plan sites have no
`ActionParams` (and six other lanes' tests call them directly, so widening their
signatures would collide rather than help), so they use a sibling recorder that
**names its provenance explicitly** — `agent_type`, `action` — instead of
inheriting it. Validate keeps the inheriting door deliberately, because a drop
there belongs to the running step. Both choices are asserted by tests, not
described.

Also closed this round: a failed record was itself silent (the recorder discarded
its return) — now warned, not made fatal, because failing a plan because its
REPORT failed is worse than the thing reported. And the noise question was
**measured rather than argued**: `agent_error_log` already takes 12,012 rows/7d,
so this adds under 0.2% even if every gap-plan run dropped a name.

**The round's real finding was again in the tests**: deleting the recorder call
from `applyNewPage` left every existing gap-plan test green — the same silence
found at validate in round 2. A wiring test at that site now fails under it.
Seven mutations, all biting.

### COUNCIL: APPROVED at round 3 (corr `bbf49822`, 2026-08-16)

`approved with 3 advisory objection(s) — none high-severity`; 9 seats approve,
and `bug_historian` — which gated rounds 1 and 2 — is one of them.

The two REVISE rounds are the substance of this fix, not overhead:

| round | what the council said | what changed |
|---|---|---|
| 1 | the fix restores acceptance for ONE opted-in caller and leaves `resolve()`'s silent drop untouched — a typo, a rename, a deleted component still vanish exactly as this bug did | every drop files a durable `PLAN_SECTION_NAME_DROPPED` finding |
| 2 | that record went into `validate_site_plan` only, while `apply_gap_plan`'s three call sites share the SAME resolver — and run **116 orchestrations/30d** against build-site-planner's handful | all four drop sites record; the gap-plan ones state provenance explicitly |

Advisories left open, all checked rather than waved through: the "six other
lanes' tests call these directly" claim is really **15 call sites across 5 test
files** (an undercount on my part, so the no-widening decision is better
supported than I argued); content-gap-planner's own section/element menu and the
resolver's zero prior test coverage are both shown by query; the deploy-check
direction asymmetry ("not live" may rest on the pod-read image tag, "live" needs
the binary's stamp) is now a table in the lane RUNBOOK; and a migration re-run
would take a second `snapshot_agent` labelled "pre-update", noted there too.

## NOTE from the `bugs_open/450` lane (2026-09-03) — the complementary half, and one place the two could be confused

282 and 450 are the two halves of the same seam and neither subsumes the other:

- **282 (yours):** the tool **DOES** exist, the planner places its section, and validate's name
  resolver drops it — `validFunctions` is built `WHERE component_level IN ('section','element')`,
  so a `component_level='tool'` row can never resolve. Fix committed 2026-08-16, riding a roll.
- **450 (ours):** the tool **DOES NOT** exist. The planner names the page anyway (tools arrive
  from the design rotation hours-to-days later, under different names), the hold nobody consumes
  says so, and five generic producers fill the page with prose that serves 200. Door half
  committed `587666be8` (register **PBP-053**), inert until the next roll.

**Where they could be confused, and the discriminator:** both end with a tool page carrying no
tool-level component, so the *fleet census query is the same query* and its results mix the two
populations. `SELECT ... FROM site_plan_sections` tells them apart — if the plan **names** a
tool component for the page, it is your drop; if the plan says `hero-tool,generic-text-block`,
the tool did not exist when the plan was written and it is ours. On seotools all seven were the
latter (`[MEASURED 2026-09-02]`), which is why 450 did not land in your lane.

**No interaction between the fixes, checked rather than assumed:** ours is a refusal at the
save/assemble/dispatch seams keyed on whether a live tool component exists; yours is a resolver
widening at plan-validation time. When your fix is live and a tool genuinely exists, our predicate
is already false and nothing of ours fires. The one thing worth your knowing: **after our roll, a
page whose tool section your resolver drops will refuse generic rebuilds rather than being rebuilt
as prose** — so a 282 instance that previously self-inflicted a shell will now leave an
`owned_page_review` receipt (`spec->>'refusal_class' = 'tool_pending'`) instead. That is a
better signal for you, not a worse one, but it changes what the symptom looks like.
