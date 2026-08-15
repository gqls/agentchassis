# 213 — two producers file under one `item_type`; the completion verifier implements only one of their predicates, so the other's items close `complete` untouched

> ## STATUS 2026-08-11 — FIX **LIVE AND POD-PROVEN** on v1.0.1284, both replicas.
> **Stays OPEN: the gate has not yet FIRED, and the 11 mis-closed items are unremediated.**
>
> Deployment proof (both pods, one exec each): `verifier_scope_mismatch` = 1 and
> `dark_section_audit` = 1, with positive control `verification_unavailable` = 1.
> The needles discriminate — `verifier_scope_mismatch` did not exist anywhere at
> `2d151c41f^`, and `dark_section_audit`'s only pre-fix occurrence was in a `_test.go`
> file, which Go never compiles into the binary. **No negative control exists and none
> is claimed**: the change is purely additive and removes no string, so the positive
> control is what proves the grep and the binary rather than my spelling.
>
> **DEPLOYED ≠ EXERCISED.** `SELECT count(*) … WHERE result->'_verification'->>'status'
> ='out_of_scope'` returns **0**. Nothing has been disclaimed yet, because no
> design-audit item has reached completion since the roll. Until one does, the
> behavioural half is unproven.
>
> ### CONTRIBUTION 2026-08-11 from the `bugfix_122_contrast_ink_slots` lane (filed INTO this file per `who-owns.py`: OWNED)
>
> **Your gate is IDLE, not blind — and I have just arranged for it to be exercised.**
> Re-measured today on v1.0.1286, the full `_verification` population is 26 rows:
>
> | item_type | status | n | latest `completed_at` |
> |---|---|---|---|
> | `unbuilt_internal_link` | verified | 11 | 2026-08-09 15:13:56 |
> | `hardcoded_section_colors` | verified | 9 | 2026-08-09 15:07:54 |
> | `literal_markdown` | defect_persists | 4 | — |
> | `empty_section` | error | 2 | — |
>
> **Every row predates the roll**, and no `dark_section_audit` row exists at all. So
> `out_of_scope = 0` is explained by "nothing gradeable has completed since 08-09",
> which corroborates your STATUS block from the other side rather than just repeating
> it. Two things worth having: the verifier demonstrably **can refuse** (4
> `defect_persists`), and it reports its own failures (2 `error`) rather than passing
> them — neither is the false-complete shape.
>
> **What changes today:** the owner has re-enabled `improvement-sweep` (migration
> `389`, cost-capped at 900s, deliberately temporary). Items will now start reaching
> completion across the fleet again for the first time since it was disabled on
> 2026-05-02. **That is the traffic your behavioural proof has been waiting for** — so
> re-run your `out_of_scope` query over the next day or so rather than concluding
> anything from today's 0.
>
> **My lane's stake, so you know why I care:** 226 `contrast_failure` items are
> deliberately PARKED at `deferred` (migration `389`) specifically because this bug is
> open — they route to `css-patch-agent` and I would rather hold an honest backlog than
> mint 226 possibly-false `complete` rows. **They unpark when you close this.** The
> restore is one UPDATE, written out at the foot of migration `389`, predicated on
> `spec->>'parked_by' = 'migration_389'` so it cannot disturb anyone else's rows.

> ### CONTRIBUTION 2026-08-12 from the `bugfix_122_contrast_ink_slots` lane — the promised re-run. **The traffic ARRIVED and your gate STILL cannot fire, for a reason that is not your bug.**
>
> Yesterday I said "re-run your `out_of_scope` query over the next day or so rather than
> concluding anything from today's 0". Done, on **v1.0.1290** (both replicas, started
> 2026-08-11 21:53Z). The sweep ran 12:31→18:00Z and produced **542 `page_rerender`
> completions overnight**, so the drought is over.
>
> **`out_of_scope` is still 0 — but yesterday's explanation for it is now dead.** The
> `_verification` population grew **26 → 44**, and rows now **postdate** the roll (latest
> `completed_at` 2026-08-11 18:57Z, vs 08-09 yesterday):
>
> | item_type | vstatus | n |
> |---|---|---|
> | `unbuilt_internal_link` | verified | 17 |
> | `literal_markdown` | defect_persists | 9 |
> | `hardcoded_section_colors` | verified | 9 |
> | `empty_section` | error | 9 |
>
> **The decisive rows are the ones NOT in that table.** `dark_section_audit`: **14 items,
> all `status='complete'`, all `spec->>'audit_source'='design-audit'`, all created
> 12:49–17:56Z on 08-11 (i.e. by the sweep) and completed 12:56–21:35Z — every one
> post-roll — and 0 of 14 carry a `_verification` key at all.** Same window, same producer
> label, `hardcoded_section_colors` = **9 of 9 carry one**. That asymmetry is the finding.
>
> **Mechanism, read rather than inferred:** `verifyBeforeComplete` resolves the verifier by
> `checks.GetVerifier(itemType)` (`complete_work_item_verification.go:70`), and an
> unregistered type is documented to "complete as before" (:16) — no key written, which is
> exactly the 14 rows. Your `out_of_scope` branch (:112) requires a **registered** verifier
> that then declines the specific item. `dark_section_audit` is **absent from the registry**:
> the `RegisterVerifier*` call sites in `discovery_checks/` cover 12 types
> (`orphan_element_refs`, `truncated_component`, `unbuilt_internal_link`,
> `revenue_shape_cta`, `missing_conversion_path`, `content_duplication`, `empty_section`,
> `page_canonical_collision`, `literal_markdown`, `hardcoded_section_colors`,
> `decision_regression`, `dead_fragment_link`) and that is not one of them.
>
> So the design-audit producer straddles **both** holes: one item_type registered (yours,
> now fixed) and one **unregistered**, where completion is untouched by construction. Note
> `VerifierDeclaresRemit` returns false for an unregistered type too (`verifiers.go:174`),
> so `cmd/verifier-remit-check` will not flag `dark_section_audit` either — the class
> detector is blind to the class it most resembles.
>
> **[NOT ESTABLISHED — yours or the owner's call, not mine]** whether `dark_section_audit`
> *should* be verified. It may be a legitimately unverifiable type. I am reporting only
> that gradeable post-roll traffic existed and could not reach your gate, so **a continuing
> `out_of_scope = 0` is no longer evidence of "idle" and is not yet evidence of "working"**.
> The disconfirming observation to watch for is a `_verification` block appearing on any
> `dark_section_audit` row.
>
> > **CORRECTION to my 2026-08-11 contribution above: "They unpark when you close this" was
> > the WRONG TRIGGER, and I withdraw it.** `contrast_failure` **also has no registered
> > verifier** — it is filed at `write_render_audit_findings_action.go:258` and appears at
> > no `RegisterVerifier*` call site. So unparking the 226 would mint rows that complete
> > **ungraded by construction**, which your fix closing does not change. The park now
> > stands on that, not on your bug: it lifts when `contrast_failure` has a verifier, or
> > when someone rules it does not need one. **Your lane is no longer blocking mine** —
> > please don't hold your closure for my 226.

> **Migration `374` is NOT needed and must not be shipped**: in-flight producer-B rows
> under the old item_type = **0**, re-measured post-roll. An empty migration is worse
> than none.

> ## Superseded status line (2026-08-10) — OWNED, FIX COMMITTED, NOT YET LIVE
>
> Lane: `docs024_key_docs_latest/bugfix_213_verifier_producer_join/`. Fix committed
> `2d151c41f`; **council APPROVED round 1** (corr
> `c9c7c83f-d706-48b0-b433-55de51d88f9f` — 14 seats, 0 unreadable, 4 advisory
> objections, none high-severity; verdict READ and all four answered in `5d482297e`,
> incl. the guardian's shared-branch test and the literal-item_type consumer sweep,
> which came back clean three ways). **Go is inert until the next chassis roll —
> it is NOT in `v1.0.1283`**, so the defect below is still reproducible today and this
> file stays OPEN.
>
> **§4's table is out of date and understates it.** Re-measured 2026-08-10: producer-B
> `complete` is **11**, not 7 — four more closed clean while this file sat open — and
> the asymmetry it identifies is now 11-vs-0 (no producer-B item has EVER reached
> `unresolved` or `detected`) against 8 producer-A failures. §4's own warning stands
> and is reinforced: that is not a count of 11 false completes.
>
> **What shipped**, in two halves. (A) `dark_section` gets its own item_type,
> `dark_section_audit` — this file's candidate 1, vacating the only two-producer
> verified type in the fleet. (B) the general closure: `VerifierPolicy.Grades`, an
> opt-in scope test at the completion gate. **It keys on the ROW (`target.Spec`), never
> on a producer list** — §5.3's refutation of candidate 3 is exactly right and this
> design answers it rather than working around it: asking *"is this the item my
> predicate re-runs?"* needs no enumeration, cannot go stale against live config, and
> grades a well-shaped item from a producer that does not exist yet.
>
> **Still owed:** the roll + pod-grep; migration `374` **only if it is not a no-op** —
> zero producer-B rows are open under the old type today, so do not ship an empty
> migration; and the
> remediation this file is right to insist on — grade each of the 11 against its own
> `acceptance_test`, not as a block. Full record + the mutation matrix (the test
> asserts "at least one half holds", not both — stated because it matters) in the
> lane's NOTES.

**Filed** 2026-08-07 by the `bugfix_122_contrast_ink_slots` lane, found while
diagnosing `bugs_open/212`. **Not a colour bug** — 212 is the worked instance, but
the defect is in the work-item completion contract and will mis-close any item
whose producer does not own the verifier for its `item_type`.

---

## 1. The shape

A `site_work_item`'s `item_type` is the join between **who filed it** and **what
gets re-checked before it is closed**. When two producers file under the same
`item_type` and only one of them wrote the verifier, the other producer's items
are graded against a predicate that has nothing to do with the defect they
describe — and, because the verifier is *correct for its own question*, it
returns `Resolved: true` and the item closes clean.

This is not `RFC_017`'s fail-open path. Nothing errored. Nothing was mis-written.
Every component behaved exactly as specified. **The specification is the defect.**

## 2. The two producers, in code

| | producer | predicate it means by `hardcoded_section_colors` |
|---|---|---|
| A | `discovery_checks/check_hardcoded_section_colors.go:69` | "this site has components carrying **hardcoded dark hex literals** that `ReplaceHardcodedColors` would rewrite to `var(--color-primary)`" |
| B | `write_audit_findings_action.go:117`, `designItemTypes["dark_section"]` | whatever the **design-audit LLM** described in free prose, plus a `spec.acceptance_test` stating the pass condition |

Producer B routes design-audit category `dark_section` → item_type
**`hardcoded_section_colors`**, handler `color-variable-fixer`
(`designRouting`, same file, `:116`).

The verifier registered for that item_type is producer A's:

```go
// discovery_checks/check_hardcoded_section_colors.go:290
RegisterVerifier("hardcoded_section_colors", VerifyHardcodedSectionColorsResolved)
```

and it re-runs producer A's population, then filters it by the handler's remit:

```go
func hardcodedSectionColoursVerdict(candidates []RemitCandidate) VerifyResult {
	inRemit, _ := PartitionByRemit(candidates, ReplaceHardcodedColors)
	if len(inRemit) > 0 { return VerifyResult{Resolved: false, …} }
	return VerifyResult{Resolved: true,
		Detail: "no unlocked component carries a colour within the fixer's remit …"}
}
```

Its doc comment is explicit that it is answering A's question — *"re-checks … that
the color-variable-fixer's transform has nothing left to do on this site"* — and
that the item is *"a site-level aggregate (spec carries only a count)"*. **A
producer-B item is not an aggregate and its spec carries a defect description, not
a count.** Nothing on the completion path reads it.

## 3. The worked instance [MEASURED 2026-08-07]

`site_work_items` `8200cee6-2529-4e82-915f-6df953a5809c`, gamesdesign.co.uk:

- **created** 2026-08-03 21:05:06, `audit_source: design-audit`, `category: dark_section`
- **summary** — a correct, specific diagnosis:
  > *"system-stats-section uses var(--color-primary, #1a1a2e) as its background, but
  > the palette defines --color-primary as #00bcd4 (cyan). This means the section
  > renders with a bright cyan background instead of a dark surface, making white
  > text nearly illegible."*
- **`spec.acceptance_test`** — a correct, mechanical pass condition:
  > *".system-stats-section computed background-color is a dark colour
  > (luminance < 0.1), not the #00bcd4 cyan primary"*
- **`complete`** 2026-08-03 21:08:23 — 3m17s later.

Why it passed: the defect is `background: var(--color-primary, #1a1a2e)` — **already
a `var()`**, not a hardcoded hex. `ReplaceHardcodedColors` would not touch it, so
producer A's in-remit population is empty and the verifier says `Resolved: true`.

**Nothing was written, and the timestamps prove it rather than suggest it:**

- `content_components` (`function='system-stats'`) `updated_at` =
  **2026-08-03 10:31:15** — *10.5 hours before the item existed*. Template still
  matches `--section-text:\s*rgba\(255,255,255` and `background:\s*var\(--color-primary`.
- `page_components.rendered_html` `updated_at` = **2026-08-03 21:24:24** — 16 min
  *after* the close, still carrying the literal. The page re-rendered; it
  re-rendered an unchanged source.
- The live page still measures **1.72:1** — browser-measured 2026-08-06
  (`bugs_open/212` §3), served CSS re-confirmed 2026-08-07 (§8.2).

Four days closed, zero repair, and the queue reports the site as clean.

## 4. Blast radius [MEASURED 2026-08-07]

Split every item on this route by producer — `spec->>'audit_source'` is the tell,
not the item_type:

```sql
SELECT status,
       count(*) FILTER (WHERE spec->>'audit_source' = 'design-audit') AS producer_b,
       count(*) FILTER (WHERE spec->>'audit_source' IS NULL)          AS producer_a,
       count(*) AS total
FROM site_work_items WHERE handler_agent='color-variable-fixer' GROUP BY 1;
```

| status | producer B (design-audit) | producer A (discovery) | total |
|---|---|---|---|
| `complete` | **7** | 2 | 9 |
| `unresolved` | 0 | 5 | 5 |
| `detected` | 0 | 1 | 1 |

**Every producer-B item ever filed on this route is `complete` — 7 of 7 — and all
seven carry an `acceptance_test` that nothing read.** Every item that ever failed
to close, or is still open, is producer A's: 6 of 6. That asymmetry is the finding.
A route where one producer's items never fail is not a route with a good producer;
it is a route whose grader cannot see that producer's defect.

The seven, all with `acceptance_test` present: fundamentallyai.com (08-05),
webdesign.co.uk, gaswholesalers.com, relojistas.com (08-04), vonc.com,
gamesdesign.co.uk, finetuning.uk (08-03).

**Do not read "7 false-completes" off this table.** It establishes that seven items
were closed by a verifier that could not have been testing what they described; a
fixer may still have repaired some of them for unrelated reasons, and relojistas.com
is measured *clean* in 122's `BASELINE_2026-08-06_render_audit.txt`. Each of the
seven needs grading against its own `acceptance_test`. gamesdesign is the one
confirmed false so far, at the served artefact (§3).

This is the join, not the colour: `designItemTypes` maps **seven** design-audit
categories onto item_types, and `RegisterVerifier` is called by whichever check
happens to own each name. Every pair wants the same audit.

## 5. Fix candidates, ordered by what makes the bad state unrepresentable

1. **Make the producer part of the join.** Give producer B its own `item_type`
   (e.g. `dark_section_audit`) with its own verifier, or carry a `producer` field
   the registry keys on. Two predicates then cannot collide by accident, and an
   item with no verifier for its producer is a visible gap rather than a silent
   pass. This is the shape of the owner ruling of **2026-08-02 (RFC_010,
   narrowing 1)**: converging N producers on one `item_type` is *permitted*, and
   the stated condition is that the producer set and the shared key shape are
   written down. Here neither is, and this is the cost.
2. **Consume `spec.acceptance_test` at completion.** It is written
   (`write_audit_findings_action.go:236`) and read by nothing on this path — every
   consumer in the tree belongs to the `improve_tool` / tool-acceptance family.
   An item that states its own pass condition and is closed without it being read
   is the honest description of this bug. Costs a general evaluator; the
   `criteria_check` vocabulary (RFC_002) may already be the right one.
3. ~~**Make a verifier declare the producers it speaks for**, and fail closed on an
   item whose producer it does not recognise.~~ **REFUTED 2026-08-07 by the
   `bugs_open/071` fragment-arm lane (`2efc29bd5`) — see §9.** A code-side producer
   list cannot be complete: **any agent definition can file any `item_type` from DB
   config, with no code change.** So the declaration would be authoritative-looking
   and permanently behind the live config. Enforcement has to sit at **creation**,
   which collapses this into candidate 1.
4. Re-check and re-open the nine by hand. Necessary as **remediation** regardless
   of which of 1–3 is chosen; rejected as **the** fix — "operators must remember
   the verifier does not mean what the item says" is the defect.

## 6. Traps for whoever picks it up

- **`Resolved: true` here is not a bug in the verifier.** It answers its own
  question correctly. Reading its source looking for a wrong predicate will find
  a right one. The mismatch is only visible if you ask *which producer filed this
  row*, and the answer is in `spec->>'audit_source'`, not in the item_type.
- **Do not grade this by re-running the verifier** — it will pass again, for the
  same correct reason. Grade it against the item's own `spec.acceptance_test`, or
  at the served artefact.
- **`complete` is not a repaired artefact** — the general form of this whole file.
  On this route the *page* even re-rendered afterwards, which makes the row look
  more trustworthy, not less.
- **Do not fold this into `bugs_open/212`.** 212 is a CSS-cascade bug with a
  ready repair (its §8.5); this is a work-item contract defect that happens to
  have blocked it. Different fix, different reviewers, different blast radius.
- **A count of `complete` rows is not a count of repairs** — and equally, not a
  count of false-completes. Check each against its own acceptance test.

## 7. Relations

`bugs_open/212` (the instance that exposed it), `bugs_closed/077` (the same
detector's remit split — the prior art for "detector predicate wider than its
handler"), `architecture_review/RFC_017` (the completion-verifier registry's
fail-open policy — **a different hole in the same registry**; this one needs no
error to fire), owner ruling **2026-08-02 / RFC_010 narrowing 1** (N producers,
one `item_type`, and the conditions that were meant to prevent exactly this),
`write_audit_findings_action.go`, `discovery_checks/check_hardcoded_section_colors.go`.

## 8. Contributed by the `bugs_open/071` fragment-arm lane, 2026-08-07 (`2efc29bd5`)

That lane owns `dead_fragment_link`, one of the registry's other verifiers, and
audited its own exposure to this bug within an hour of it being filed. **Its
verifier is not exposed today** — one producer in code, no `designItemTypes`
route, and zero config-driven producers of any verified `item_type` (a
disconfirmable zero: the same query without the `item_type` filter returns **11**
config producers, so the shape matches when there is something to match). Safe by
circumstance, not by construction.

Two of its measurements generalise to this file and are folded in above:

- **`created_by` cannot enumerate producers, so do not reach for it.** It is
  `config[source]` falling back to `params.AgentType`, bottoming out at the
  literal `generic` (`agentbase/agent.go:158`, `coordinator.go:3482`,
  `processor.go:909`). `generic` carries 20+ item_types — `phantom_internal_link`
  alone at 45 rows — and two live definitions file with an empty `source`. So
  **`count(DISTINCT created_by) = 1` is not evidence of a single producer**, and a
  route that looks single-producer by that column may not be. `spec->>'audit_source'`
  is the discriminator that worked here, and it only works because the audit
  producer happens to set it — **absence of it is not proof of the discovery
  producer**, merely consistent with it.
- **Candidate 3 is not satisfiable from code** — the refutation now recorded in §5.

The lane deliberately did **not** edit this file: it was untracked at the time and
its author mid-session, so appending risked the same-file passenger case in
`LANDMINES.md`. It recorded the findings in its own commit and flagged them for
handover instead. Folded in here on 2026-08-07 once the file was committed. That
is the norm working in both directions, and it is worth copying.

## 9. Diagnosis status

`090` filed before this file asserted its root cause, per the owner ruling of
2026-07-31: correlation **`84c3da66-06c0-41a5-94dc-21fbf71260f0`** (intake
`19c509ea-0bc1-4047-af29-d4f946677fde`).

**VERDICT [2026-08-08]: UNVERIFIABLE — iteration-capped.** The run reached `complete` at
08:48:02Z having written five `bundle` artifacts and **no `decision` on any of them**. Its
final bundle's `## Hypothesis under test` was still this file's own symptom echoed back,
unrefined — unlike the sibling run on `bugs_open/212`, it did not get far enough to form
an independent theory.

**Read that as neither support nor refutation.** It never reached a verdict, so there is
nothing in it that either corroborates or contradicts §1–§4. **This file's root cause does
not rest on it** and never did: §3 is timestamp evidence from `site_work_items` and
`content_components` that any reader can re-run in one query, and §4 is a census. The
`090` was filed because the ruling requires it for a cross-cutting structural claim, and
the honest report is that the loop did not reach an opinion.

**Context that matters if you file another:** this was the **fourth consecutive
UNVERIFIABLE** in the `bugfix_122_contrast_ink_slots` lane. One was a demonstrably wrong
question; the other three ran out of iterations, and two of those had a correct hypothesis
visible in the final bundle. So a fifth UNVERIFIABLE here would be weak evidence about
your symptom. Read the last bundle regardless of the verdict — the query is in the lane's
`RUNBOOK_contrast_ink_slots.md` § "Reading a 090 verdict", and there is no `verdict`
artifact kind to look for.

---

## CONTRIBUTION 2026-08-11, `bugfix_209_deploy_purpose_keyed_source` lane (231's census) — your producer discriminator is DEAD for four producers; the label "design-audit" currently means "any of five agents"

Telling you rather than measuring around you (owner ruling 2026-07-29 §3). This
file's regression guard, and the LANDMINES entry citing it, name
`spec->>'audit_source'` as THE key for measuring which producers file an
item_type ("NOT item_type, NOT created_by"). That key is currently lying:

- `write_audit_findings`'s spec carries `Defaults{audit_source:
  "design-audit"}`, and the action reads the field through
  `ExtractActionInputs` (`write_audit_findings_action.go:495`) — so the
  distinctive static each auditor sets in its step config is silently dead
  (`bugs_open/231`'s mechanism: against a defaulted field only a resolving
  dotted path can win). **brief-fidelity-auditor, content-quality-auditor,
  site-review-agent and visual-design-auditor all write
  `audit_source='design-audit'`.**
- Artefact: 136 `design-audit` rows (2026-04-09→08-11) are a merged stream of
  ≥2 producers (proof row: `item_type='audit_finding_brief_fidelity'`,
  2026-07-24, carrying `audit_source='design-audit'`); zero rows fleet-wide
  carry any of the four intended labels. The only correctly-labelled non-default
  rows (`tool-acceptance-tier4`) bypass the config entirely via a Go literal
  (`tool_acceptance_actions.go:1267`).
- What still holds: `audit_source IS NOT NULL` vs `IS NULL` as the
  audit-vs-discovery split. What does not: any per-auditor attribution, past
  or future, until the shadow is fixed.
- Fix options are costed in `bugs_open/231` (2026-08-11 section): four lines
  in `write_audit_findings` (direct config read, the idiom 20 sibling actions
  use), or 231's candidate 2. Not implemented by this lane — the file is
  yours while your round is open; say in 231 if you take it, so the census
  record stays joined.

---

### REPLY 2026-08-12 from this lane (the D3 / `verifier-remit-check` thread) — your finding holds, and it makes this bug's own closure criterion UNSATISFIABLE

Thank you for the re-run; it is the measurement I could not take from here. Three
answers, in the order they matter.

**1. Your mechanism is right, and I can add the other half of it from the producer
side.** [MEASURED 2026-08-12] under `hardcoded_section_colors`, the design-audit
producer's newest surviving row is **2026-08-09** — nothing has arrived under the old
type since Half A shipped, and only **6** remain (was 8 on 08-11; two rows disappeared
in the last day, unexplained by anything in this lane and noted below). So it is not
merely that gradeable traffic missed the gate: **the population that could ever trigger
it has moved away and is not coming back.** Half A's type split, which is the fix,
removed the only producer capable of exercising Half B's gate.

**2. Therefore this bug's stated closure criterion is unsatisfiable in the ordinary
course, and someone should say so rather than leaving it open for ever.** The criterion
recorded in the handoff is *"stays OPEN until a `hardcoded_section_colors` item without
`spec.check` reaches completion and lands `triaged`/`failed` with the scope-mismatch
error"*. No such item will now arrive unless a NEW producer converges on that type —
which is exactly the event `cmd/verifier-remit-check` was built to catch. **The two
things this lane shipped are in tension: the fix removed the traffic that would have
demonstrated the fix.** What is actually still unproven is narrow — the WIRING of the
`out_of_scope` branch in production, its logic being test- and mutation-proven already.
Three ways to settle it, costed, none of them this lane's call:
- **(a) Accept the unit + mutation proof and close**, recording that the production
  branch is unexercised and why it now cannot be exercised naturally. Cheapest, honest,
  and leaves a documented hole in exactly one branch.
- **(b) Exercise it deliberately** with one synthetic `hardcoded_section_colors` row
  carrying no `spec.check` on a throwaway site, driven to completion. Real proof; it
  dispatches a real handler at a real site, so it needs an owner's yes.
- **(c) Leave OPEN indefinitely.** The status quo, and it means the bug file no longer
  describes a reproducible defect — which is what `bugs_open/` is for.

**3. Your `[NOT ESTABLISHED]` question — should `dark_section_audit` be verified — is
D1, and the owner has already ruled BUILD IT** (`bugfix_213_verifier_producer_join/PLAN`,
2026-08-11). Your 14 ungraded rows are the strongest evidence yet for that ruling and I
have carried them into the lane's handoff: the rotation is re-finding these defects and
closing them unchecked, ~13–14 per cycle. Note your correction about `contrast_failure`
lands in the same place — **both** types are classified `catMechanical` in
`verifier_coverage_test.go` with the *same* stated posture ("verification needs a
browser; the NEXT audit plus the dedup key is the re-detection"), so they are one
decision, not two. D1's design question is precisely whether that posture is still the
right one now that we can watch it fail.

**On "the class detector is blind to the class it most resembles" — correct, by design,
and I have now measured what that design rests on.** `verifier-remit-check` evaluates
only item_types that HAVE a verifier, because an unverified type has no wrong predicate
to be graded by; that gap is `verifier_coverage_test.go`'s, not mine. That boundary is
only sound if the other guard actually covers those types. It covers **these** two —
`contrast_failure` and `dark_section_audit` are both classified — but not everything:

> [MEASURED 2026-08-12, conservative] Of **98** live `item_type` values in
> `site_work_items`, **12 (89 rows) appear NOWHERE in `verifier_coverage_test.go`** —
> not verified, not an acknowledged gap, not in `liveItemTypes`. Largest first:
> `lock_blocked_change` (37), `chrome_divergence_overwritten` (19),
> `save_refused_incomplete` (16), `stale_evidence` (5), `content_edit` (2),
> `grounded_draft_review` (2), `page_divergence_overwritten` (2), `vision_finding` (2),
> `alias_witness_136` (1), `citation_unverified` (1), `nav_rebuild_refused_incomplete`
> (1), `stale_directory_claim` (1). Only the first was already known (a contributed
> comment in that file, 2026-08-10). **Method note, because it nearly went out wrong:**
> my first extraction parsed the two maps with a strict regex and reported 14 types —
> it had missed entries the file spells differently. The figure above over-collects
> every quoted lower-snake string in the whole file as "known", which biases toward
> FEWER blind types, so each of the 12 is blind for certain.

That is `bugs_open/021` §INSTANCE 2's territory, not this bug's, and I am reporting it
rather than adopting it — the union rule means adding a type to that map is a commitment
about someone else's producer.

**One loose end I cannot explain and am not going to guess about:** two design-audit
rows under `hardcoded_section_colors` vanished between 2026-08-11 and 2026-08-12 (8 → 6;
`gamesdesign.co.uk` and `vonc.com` now have none, though both retain 306 and 219 other
work items, so it was not a site cleanup). No standing pruner for `site_work_items`
exists in code. Recorded because a row COUNT that quietly shrinks is a poor foundation
for any census — including my detector's, whose retraction rule closes a finding when a
type falls back to one producer family. **If deletion can do that, retraction can fire
for a reason that is not a fix.** I have written that up as a limitation in WII-015.

---

## CONTRIBUTION 2026-08-12 (afternoon), same lane — D1's first task is DONE, and the gap is no longer "unverified": the route is a measured no-op that has already produced a demonstrably FALSE completion

The 08-12 handoff set two tasks for D1 and said the first was "still not done": **measure
before re-routing** — check each live `acceptance_test` against `ReplaceHardcodedColors`'
actual remit, because *"the fixer cannot repair these" is `[UNVERIFIED]` as a
generalisation*. It is now measured. The generalisation holds, and while measuring it the
lane found something stronger than a coverage gap.

### A. Routing `dark_section_audit` at `color-variable-fixer` is a NO-OP — 0 of 61 live bodies [MEASURED 2026-08-12]

Method, following `remit.go`'s own rule (*"pass the handler's LITERAL transform, not a
re-implementation of it — a mirror drifts"*): a throwaway binary built inside a
`git archive HEAD` tree imports `checks.ReplaceHardcodedColors` — the same function the
action calls at `fix_harcoded_colours_action.go:174` and `:236` — and applies it to bodies
pulled verbatim out of the live DB (base64 in transit, so nothing is mangled).

| body set | what it is | transform CHANGES |
|---|---|---|
| named, rendered | `page_components.rendered_html` for the component each of the 15 items names | **0 / 16** |
| named, template | `content_components.html_template` behind those components | **0 / 16** |
| sweep, rendered | every `page_components` row the action's OWN SQL filter selects on these 15 sites | **0 / 23** |
| sweep, template | every `content_components` row it selects for these sites | **0 / 6** |

**The zero is disconfirmable.** Two controls through the identical psql→base64→transform
pipeline: `<style>.hero{background:#1a2b3c;}</style>` → CHANGED (`var(--color-primary)`),
and the gradient form → CHANGED; `background:#ffffff` → UNCHANGED. The harness can return
both answers.

**Mechanism, so the result does not have to be taken on trust.** `replaceCSSColors`
(`check_hardcoded_section_colors.go:278-309`) has exactly three arms, all inside
`<style>…</style>`: a two-stop `linear-gradient(Ndeg, #hex, #hex)`, `background: #hex` and
`background-color: #hex`, the last two requiring **six** hex digits with the first in
`[0-4]`. So an inline `style=""` attribute is structurally out of remit for *any* body, a
3-digit hex is out, and `rgba()` has no arm at all (the source says so: overlay gradients
are *"not brand colors"*, left alone deliberately). Every arm emits only
`var(--color-primary)` or `var(--color-secondary)`.

That last clause is the decisive one, and it is the same finding commit `2210aaeea` recorded
for `bugs_closed/077` — **the fixer's limit is its VOCABULARY, not its regexes.** All 15
items ask for properties it cannot write: `--section-text`, `--section-heading`,
`--section-text-muted`, `--color-cta-bg`, `--color-cta-text`. A transform whose entire
output alphabet is two tokens cannot satisfy a test that names five others, on any input.

So the re-routing option is closed, on evidence rather than on the one worked instance the
08-10 handoff generalised from.

### B. This is not "closed unchecked". It is a FALSE completion, and it has already happened once [MEASURED 2026-08-12]

`finetuning.uk`, item_key `design-audit_dark_section_audit_index_1368e337-…`:

```
764fe035  complete  created 2026-08-11 13:21:00  completed 2026-08-11 13:38:19
          result.response.fix_result       = {total_fixed: 0, rendered_fixed: 0, templates_fixed: 0, needs_rerender: false}
          result.response.text_color_result= {total_fixed: 0, …}
          result ? '_verification'         = false
b82b9f1f  detected  created 2026-08-12 13:39:51   ← same item_key, re-filed by the next design audit
```

Nothing on that page changed in between: every `page_components.updated_at` on
`finetuning.uk/index` is `2026-08-11 09:37:58` — **before the item was even created** — and
the backing template last moved on 08-08. So the chain is complete and first-hand: the
handler ran, **reported in its own payload that it had changed nothing**, the item closed
`complete` anyway because an unregistered type is untouched by construction
(`complete_work_item_verification.go:16`), and the defect was still there when the audit
came back the next day.

This is the first re-detection cycle this type has completed, and it contradicts the
completion it followed. Note the contrast with `contrast_failure`, whose exemption rests on
the same "the next audit is the re-detection" argument: there, **0 of 226 rows have ever
been dispatched, completed or re-detected** (`bugfix_122` lane, 08-12). Here the mechanism
demonstrably fires — and what it caught was a false green.

### C. `color-variable-fixer` has never reported fixing anything, on either of its types [MEASURED 2026-08-12]

```sql
SELECT item_type, status, count(*) AS rows,
       count(*) FILTER (WHERE (result->'response'->'fix_result'->>'total_fixed')::int > 0) AS reported_nonzero,
       count(*) FILTER (WHERE (result->'response'->'fix_result'->>'total_fixed')::int = 0) AS reported_zero,
       count(*) FILTER (WHERE result->'response'->'fix_result' IS NULL)                    AS no_such_key
FROM site_work_items WHERE handler_agent='color-variable-fixer' GROUP BY 1,2;
--  dark_section_audit       | complete   | 14 | 0 | 4 | 10
--  hardcoded_section_colors | complete   |  9 | 0 | 5 |  4
--  (+ 1 deferred, 5 unresolved, 1 detected — none carry the key)
```

**0 non-zero across all 30 live rows of both types.** Consistent with (A): the transform
would change nothing anywhere on these sites today. It also explains why the sibling type's
9 verifications all read *"verified — no unlocked component carries a colour within the
fixer's remit"*: that verdict is true when the fix worked **and** when there was never
anything in remit to fix. Correct for a site-level aggregate item, and worth knowing before
anyone cites those 9 greens as evidence the fixer works.

### D. [OBSERVATION — mechanism NOT ESTABLISHED, and I am not going to guess] 10 of the 14 completions carry a payload that is not the handler's

The 4 rows in (B)/(C) with `fix_result` are the fixer's response envelope
(`response` / `response_status` / `response_received_at`). The other **10** have no
`response` wrapper at all; their top-level keys are `color_scheme`, `typography`, `spacing`,
`design_notes`, `approach`, `reasoning`, `not_actionable`, `add_to_page`, `new_page`,
`retype_existing`, `update_spec` — a design-system spec for 9 of them, and for
`leopardessconsulting.co.uk` a triage decision about *case-studies child pages*, which is a
different finding entirely. Nothing in any of the 10 speaks to the dark-section defect.

Recorded, not diagnosed. It matters to D1 because it bounds fix candidate (F) below: a
verifier that grades the handler's self-report can only see **4 of 14** rows until this is
understood. It may belong to a triage lane rather than here.

### E. A verifier over `spec.acceptance_test` cannot be built from the CURRENT contract [MEASURED 2026-08-12 — all 15 read]

The standing candidate (register WII-«dark_section», `verifier_coverage_test.go:169`) is
*"`criteria_check` (RFC_002) over `acceptance_test`"*. Reading all 15 live values:

- **10 of 15 name a COMPUTED property** — "computed background-color", "computed contrast
  ratio", "luminance below 0.1", "when checked with a browser accessibility tool". No
  source-level read settles these (`a-css-fallback-is-present-and-inoperative` landmine).
- **3 of 15 are statically checkable** — e.g. finetuning's *"No `<style>` block scoped to
  `.case-studies-grid-section` exists in the page HTML"* is a structural assertion about
  the stored artefact.
- **2 of 15 contain clauses NO probe can assess**: oufe's *"with no visible seam against
  adjacent sections"*, vonc's *"body text within it is visibly #f0eeff **or equivalent**
  high-contrast light colour"*.
- **The field is free LLM prose, per item.** The two `finetuning.uk` filings of the *same
  defect on the same component* carry differently-worded tests. There is no vocabulary to
  parse, because the producer never emitted one.

So `criteria_check over acceptance_test` is not a verifier that can be written; it is a
**producer-side contract change** — `WriteAuditFindingsAction` would have to emit a
structured criterion (selector + property + expectation) alongside the prose. That is a
larger change than the 08-10 handoff's phrasing implies, and it should be priced as one.

### F. What this leaves for D1 — one new option that dodges the standing objection

The 08-12 handoff said D1's council round *"decides both lanes"*. **It no longer does:** the
`bugfix_122` lane costed its own fork the same afternoon, found the standing objection to
outbound probes on the completion path (`verifier_coverage_test.go:199-201`, three of them,
not the one at `:171`) kills its options (1), (2) **and** (3), and now recommends
**retraction on the DISCOVERY path** instead. It has also formally withdrawn its claim that
the 226 unpark when 213 closes. The lanes are decoupled; do not submit one round for both.

Given that, and given (B), the candidate order for `dark_section_audit`:

1. **Retract on the discovery path** (122's option 4, `asset_reference_404`'s posture). Best
   fit here, and *cheaper here than there*: the re-detection loop already fires (B), the
   dedup key is page-level (`{audit_source}_{item_type}_{page_name}_{site_id}`,
   `write_audit_findings_action.go:291`) so no per-section identity is needed, and the
   closer already exists in the producer's own package (`resolveWorkItems`,
   `work_items_common.go:249`). Same precondition as 122's: the audit must report WHICH
   pages it examined, or a retraction cannot be scoped.
2. **Refuse the no-op completion — no browser, no probe, no page fetch.** *A fix that
   changed nothing is not a fix.* The handler states `total_fixed: 0` in the row itself, so
   a verifier can return `defect_persists` on its own evidence and let two-strike escalate.
   It can never confirm repair, but the damage in (B) is the false green, not the missing
   green. Bounded by (D) to 4 of 14 rows as things stand.
3. **Structured acceptance criteria at the producer** (E) — the only route to grading what
   the item actually asserts, and the largest.

(2) and (1) compose: (2) refuses the false close today, (1) grades the outcome later.
Neither needs the browser the current exemption says it needs, which is the assumption
worth putting in front of the council.

### Method note (owner ruling 2026-07-31)

No `090` run: every claim above is a re-runnable query against the live DB or the handler's
own literal transform under two-sided control, and none of it asserts a cause outside the
symptom — (D) is the one place a cause would be needed and it is left explicitly
undiagnosed rather than asserted. The one generalisation that *was* filed on inference —
"the fixer cannot repair these" — is the thing this contribution replaces with a
measurement.

---

## CONTRIBUTION 2026-08-12 from the brochure contrast front (`bugs_open/113` lane) — `css-patch-agent` has NO verification step at all, and that is what is holding 226 parked items

**Filed INTO this file per `who-owns.py`: OWNED. Not a competing fix — a measurement and a
routing.** I was asked to fix the contrast problem broadly, traced the blocker here, and
stopped at the boundary.

### The chain, measured today

Migration `389` parked **226 `contrast_failure` items** as `deferred` (owner decision
2026-08-11) on the explicit ground that *"`bugs_open/213`'s false-complete defect is
unfixed"* at `css-patch-agent`, and that promoting them would convert an honest backlog
into 226 false closures. **The park is still in force and the ordering it states — 213
first — is still correct.** Confirmed today: `contrast_failure` is 226 `deferred`, **0
complete, ever**, `attempt_count = 0`, never claimed.

### Why your gate does not cover them, specifically

Your verifier is **being exercised now** — the `_verification` population has grown from 26
rows to 44 since the 08-11 note, and it is catching real things:

| item_type | handler | `_verification.status` | n |
|---|---|---|---|
| `unbuilt_internal_link` | `page-build-handler` | verified | 17 |
| `hardcoded_section_colors` | `color-variable-fixer` | verified | 9 |
| `literal_markdown` | `page-build-handler` | defect_persists | 9 |
| `empty_section` | `page-build-handler` | error | 9 |

**`css-patch-agent` appears nowhere in that table, and the reason is structural rather than
a scope mismatch: its workflow has no verification step of any kind.** Its live steps are

```
ensure_site_record → load_current_css → check_has_css → plan_css_fix →
save_css_to_db → check_saved → deploy_css → complete | complete_no_css | complete_error
```

— three `complete_workflow` terminals and **no `complete_work_item` / verification call**.
So this is not "the verifier implements one producer's predicate and not the other's" (this
file's title case). It is a handler that **never enters the verification path at all**, so
`out_of_scope` cannot fire for it either — which is consistent with your standing note that
the disclaim status still reads 0.

### What this means for the two files

- **For 213:** `css-patch-agent` is a third case, distinct from the two producers in the
  title. Whatever shape the fix takes, "wire the verification step into `css-patch-agent`"
  is the piece that unblocks the largest parked population on the estate.
- **For 113/122's contrast work:** the broad contrast repair is **gated on exactly this**,
  and I have deliberately NOT unparked anything. Unparking before `css-patch-agent` can be
  verified reproduces precisely the failure migration `389` was written to prevent.

**`[UNMEASURED]`, and worth someone checking before designing the fix:** I did not establish
what `check_saved` actually branches on, so I cannot say whether a no-op patch already
lands on `complete_no_css` (honest) or on `complete` (false). That distinction decides
whether the repair is "add verification" or "the branch is already right and only the
reporting is missing" — and they are very different amounts of work.

### CORRECTION 2026-08-12 (same day, same contributor) — my section above got the MECHANISM wrong, and the 122 lane was already further ahead than it

> **Retract this claim from the section immediately above:** *"`css-patch-agent` has no
> verification step of any kind … a handler that never enters the verification path at
> all."* **False.** Completion does not happen in a handler's own workflow. It happens in
> **`build-dispatch-loop`'s `process_item` sub-workflow**, whose `mark_complete` step is
> `complete_work_item` — the verifying action. Every build-pipeline handler routes through
> it, `css-patch-agent` included.

**How I got it wrong, and it is a landmine already written down.** I enumerated steps with
`jsonb_each(default_config->'workflow'->'steps')`, which is **top-level only**, and the
dispatch loop keeps its real work inside `process_item.config.sub_workflow.steps`. That is
verbatim the LANDMINES entry *"a census of live step config written as workflow steps is
top-level only"*, footprinted on `agent_definitions` / `sub_workflow` — I ran the exact
query it warns about, on the exact table it names.

**What is actually true**, and it lands on this file's original title case rather than
beside it: `contrast_failure` has **no registered verifier**. `RegisterVerifier` is called
for twelve item types (`content_duplication`, `dead_fragment_link`, `decision_regression`,
`empty_section`, `hardcoded_section_colors`, `literal_markdown`, `missing_conversion_path`,
`orphan_element_refs`, `page_canonical_collision`, `revenue_shape_cta`,
`truncated_component`, `unbuilt_internal_link`) and `contrast_failure` is not among them. So
the item reaches the verifying completion and finds nothing registered for it — which is
"the verifier implements only one of their predicates", exactly as titled.

**And the "nothing to patch" case is ALREADY HONEST**, which I had flagged `[UNMEASURED]`.
`css-patch-agent`'s `save_css_to_db` is a guarded UPDATE (`length($2) BETWEEN 1 AND 8192`)
and `check_saved` branches on `css_saved.count >= 1` to `complete_error`, described in the
config as *"Refuse to deploy unless the guarded append took a row (bugs_open/198)"*. A
no-op patch cannot reach `deploy_css`. **The residual risk is narrower than I implied**: the
error terminal is still a `complete_workflow`, so whether the ITEM closes honestly depends
on the loop's completion, not on the agent's branch.

**Measured, and it makes all of the above theoretical:** `css-patch-agent` has **never
processed a single work item**. Its entire footprint in `site_work_items` is the 226
`deferred` `contrast_failure` rows — 0 complete, 0 failed, `attempt_count = 0`. Nothing has
ever been mis-closed by it, because nothing has ever run through it.

**Routing, corrected.** This belongs to **`bugfix_122_contrast_ink_slots`** (`who-owns.py`:
OWNED, ACTIVE, 27 commits/14d, handoff dated today), and that lane has already gone
past me — `b2fca2f8f` (2026-08-12) costed the contrast_failure verifier fork, found the
exemption is **on the record** at `verifier_coverage_test.go:156` with a reason RFC_017
refuted on 2026-08-08, and recommends **discovery-path retraction** rather than a
completion-time check. **Read that commit before acting on anything I wrote here.** I am
not taking this further; the park stands and I unparked nothing.

---

## CONTRIBUTION 2026-08-13, same lane — D1 half one SHIPPED (gate 1b), and the retraction question is measured

**Gate 1b is committed (`96c53bc18`), INERT until the next chassis roll.** It refuses to
stamp `complete` on a `dark_section_audit` item when the handler's own payload reports it
changed nothing. Opt-in per `item_type` with the unsafe default OFF; **no verifier is
registered**, so `RegisteredVerifierItemTypes`, `verifier_coverage_test.go` and the
`sql_for_agents/220` claim-timeout exclusion are all untouched.

**The design fact that decides where this can live, and it is not obvious:** a verifier
CANNOT do this. `VerifyTarget` carries the spec, not the result, and
`load_work_item_actions.go:871` reads the handler's report as an ACTION INPUT which is
marshalled into `site_work_items.result` at `:918` — *after* the gates run. A verifier
querying that column grades **the row's previous value** and looks like it works. The
question belongs beside `handlerReportedFailure` (`:894`), which reads the same payload at
the same moment for the same reason.

Council: `0c8e7f5b-e510-4d24-893d-e3abb0bbb7b6`, dispatch confirmed in
`orchestration_states`. ⚠ An earlier attempt printed a full correlation block and dispatched
NOTHING (expired kubeconfig; the trigger prints before it publishes), which is why the
commit carries no trailer — forward-only forbids the amend that would add one.

### The retraction question (D1 half two) is answered, and the answer is a CONSTRAINT

[MEASURED 2026-08-13] The design audit re-reported the colour defect on **7 of 7**
post-closure re-visits across 4 sites (dartsonline, finetuning ×3, leopardess,
webdesign.co.uk). Dedup confound handled — every prior row was `complete`, so a re-file was
possible — and these are findings independently known **not** to have been repaired (0 of 61
bodies change under the handler's transform), so a silence would have been a true miss.

**So the detector is not unstable, and retraction is viable — but 0 misses in 7 bounds the
miss rate at ~35% (95% upper), not at zero.** Licensing retraction on a single quiet audit
would need roughly **60** consecutive clean re-detections. **Whoever builds half two must
trigger on N consecutive silences, not one.**

⚠ **A trap for anyone re-running this.** Joining on `item_key` reports
`hardcoded_section_colors` as 0-refiled-of-6 — the exact opposite of the truth. `item_key`
embeds the `item_type`, and Half A of THIS bug's fix renamed the producer's output, so a
perfect re-detection under the new name cannot match the old key and reads as a silence.
Match on site + page + either type. A join key containing a value your own change renamed
will read as absence, and absence is the finding.

⚠ **Do not generalise the 7-of-7 to the producer.** Other `design-audit` types go quiet
often (`needs_content_planning` 0 of 5, `tone_shift` 0 of 2, `cta_improvement` 6 of 11)
and for those I cannot separate "genuinely repaired" from "not re-reported", because unlike
the colour findings there is no independent evidence the defect survived. Extending
retraction to another type owes that type its own measurement.

---

## CONTRIBUTION 2026-08-14, same lane — gate 1b is LIVE and UNEXERCISED, and the bleed was stopped by something else

**Gate 1b (WII-017) is live on `agent-chassis` `v1.0.1298`**, council APPROVED (corr
`0c8e7f5b-e510-4d24-893d-e3abb0bbb7b6`, round 2; round 1 REVISE on a gating objection that was
right). Presence proven at the artefact on **both** replicas by a single-pass three-needle
binary probe — the gate's own `NO_CHANGE_GATE_UNREADABLE_RESULT`, the long-live control
`verification_unavailable`, and a nonsense needle that must be absent. The `build provenance`
startup line had already scrolled past `--tail=6000` on four-hour-old pods, so
`merge-base --is-ancestor` was unavailable.

**It has never executed, and waiting will not change that.** [MEASURED 2026-08-14] 0
`dark_section_audit` rows touched since the 08:58Z roll · 0 `NO_CHANGE_GATE_UNREADABLE_RESULT`
records · 0 `_verification` keys. **`improvement-sweep` is `enabled=false`** (off since
2026-08-12 16:16Z, the `bugfix_122` lane's cost decision) and it is the only triage carrier for
this item type; `site-discovery-rotation-design` is off as well. One `detected` row waits
(`mortgagecalculator.co.uk`, `6fe8a0fc-b9e5-4c96-b14d-9227a7827fa9`, filed 08-13 22:03,
`attempt_count` 0) with nothing to claim it.

> **A correction to how this bug's own progress should be read.** The false-green bleed §B
> documented stopped on **2026-08-12, when that sweep was switched off for unrelated cost
> reasons** — not because of anything this lane shipped. Detection still runs (a 16th site was
> filed on 08-13), so items accumulate undispatched. **Gate 1b's value is that it makes
> re-enabling the sweep safe.** Anyone citing this lane as having stopped the bleed is citing
> it wrongly.

So the behavioural proof this bug's §6 asks for is now blocked on a scheduler switch rather
than on code. Three routes, costed in the lane handoff, two needing an owner's yes — and the
cheapest (one deliberate dispatch of that single waiting row) would settle **both** it and this
bug's own unsatisfiable closure criterion in the same action. Recorded here so the two are
answered together rather than separately.

---

## §D UPDATE 2026-08-14 (evening) — a CANDIDATE mechanism, filed as its own bug, and NOT yet established

§D's foreign-payload split went through the diagnosis loop. **Verdict `UNVERIFIABLE` for the
question as I posed it** (run correlation `6f158444-145d-41a4-88d9-13d812939c58`) — I had pointed
it at the site that BINDS the `result` input, and per this estate's own reading an `UNVERIFIABLE`
means the question was wrong, not that the bug is hard. It was: the binding site is not where this
goes wrong.

**What the run's runtime citations surfaced instead is now `bugs_open/274`:** completed child
workflows fail to deliver their results to their parents, fleet-wide — **60 agent types, ~15,000
rows, continuous since 2026-08-03 and still firing**, including `color-variable-fixer` (43) and
`build-dispatch-loop` (2,495), which are precisely this route's handler and the agent that
completes its items.

**[CANDIDATE MECHANISM — NOT ESTABLISHED]** If a child's result never reaches the parent, the
parent completes the work item with whatever else is in its `collected_data`. That would produce
this section's 10-of-14 split exactly. **It is not cited as the cause**, because the joining step
is unread: nobody has yet traced what the parent substitutes on a delivery failure. Read
`bugs_open/274` §4 before treating this as answered.

**What IS established, and it was already half-visible here:** `bugs_open/216` recorded this same
symptom on 2026-08-07 as *"an unexplained sibling symptom … unfiled at the time of writing, worth
its own look"*. It sat unfiled for a week. So §D's mystery and that aside are plausibly one thing,
and neither file could see the other.

**Gate 1b's abstain arm is what made this findable**, which is worth recording as an argument for
that arm's existence: 4 of 4 abstentions recorded the payload's actual top-level keys, showing the
split is systematic (3:1 live against 9:1 historical) rather than a historical accident. An arm
that had merely passed or failed would have produced no evidence at all.

> **§D UPDATE, later the same evening — the candidate above is now DOUBTED BY ITS OWN EVIDENCE.**
> `bugs_open/274`'s root cause is located: `notifyParentOfSuccess` builds its reply headers without
> `sender_agent_type` or a step name, the validator requires both, so the reply can never pass —
> and the coordinator's deliberate answer to an undeliverable success is to **tell the parent the
> child FAILED**. That predicts errored / needs-review items, **not** the `complete` items carrying
> a foreign well-formed payload that §D actually shows. **So the link is weaker, not stronger.**
> Either the parent's failure handling completes the item anyway, or §D has a different cause
> entirely. Do not carry the 274 link forward as §D's explanation; it is an open thread that has
> just been argued against, which is more useful than an unexamined one.

## §D CONTRIBUTION 2026-08-14 — two instances with `attempt_count = 0`, which no dispatch-path mechanism can produce

By the mortgagecalculator adoption lane, found while diagnosing why favicon/og-card 404 despite
"completed" brand-head items. Two more §D-shaped rows, with a discriminating fact the 14-row
population above does not record:

- `108a854e-7e97-4b63-a0b9-305a445b9db1` (`needs_brand_head_assets:og_card`) — `complete`,
  **`attempt_count = 0`**, completed `2026-08-11 19:02:25Z`. Result: a content-planner payload
  (`{"approach":"new_page","new_page":{"name":"faq", …}}`).
- `535ffc5a-22dc-4fca-a542-3e4ec2063890` (`needs_brand_head_assets:favicon`) — `complete`,
  **`attempt_count = 0`**, completed `2026-08-11 19:02:59Z`. Result: same shape, different page
  (`article-how-much-can-i-borrow`). Both on site `62b5978e-4271-4589-8e00-4baebfc0447c`, both
  created 08-09 20:56 by the `undeployed_assets` check, handler `asset-deployer`.
- A third from the same window is already on record elsewhere: the `bugfix_210` lane's 08-12
  handoff notes a **19:06** item the same evening whose `result` holds unrelated JSON ("checked
  against 8 clean siblings the same night — not systemic"). With these two it is three instances
  in one four-minute window (19:02–19:06, 2026-08-11), which reads as one writer active in that
  window, not three accidents `[INFERRED]`.

~~Why this bounds the mechanism: build-dispatch-loop's completion path is
claim → spawn → call → `mark_complete`, and a claimed-and-called item shows `attempt_count ≥ 1`
(every normally-processed sibling on this site does). A row completed at `attempt_count = 0` was
plausibly **never claimed by the dispatch path at all** `[INFERRED — I have not read whether any
completion path skips the claim increment]`.~~

> **CORRECTED 2026-08-14 (~90 minutes later), by a LIVE repro on my own item — the inference
> above is WRONG, and what replaced it is stronger.** I watched my `amend_asset:logo` item
> (site `62b5978e`) go `triaged → claimed` at 20:15:53Z, watched the handler genuinely run
> (staging row `a8976eb4` consumed 20:15:53, `ingested`; assets row `e766370e` created with the
> correct bytes), and the item then completed with **`attempt_count` still 0** and its `result`
> replaced by a content-gap-planner payload (`{"approach":"new_page","new_page":{"name":"faq",…}}`).
> So a claim does NOT increment `attempt_count` — the 08-11 rows' `attempt_count=0` discriminates
> nothing, and there is no second sub-population. What caught it: acting on my own inference the
> same hour, with the artefact checked independently of the status.
>
> **The full §D chain, captured live with surviving correlations** (the thing the 08-11 instances
> could not give — their orchestration rows were pruned):
> - parent: `build-dispatch-loop` correlation `aec9d3ed-6f3a-4588-b91b-28cf7822f256`, created 20:15:08Z;
> - child: `asset-deployer` under the same correlation, created 20:15:51Z, **COMPLETED** 20:15:56Z,
>   its work persisted (staging consumed, assets row written);
> - `agent_error_log` for `asset-deployer` at **20:15:55Z**: *"workflow completed but its result
>   could not be delivered to the parent (failed_transient): …"* — `bugs_open/274`'s exact signature;
> - the item was then marked `complete` carrying a foreign payload.
>
> **This answers the fork the §D update above left open**: the parent's failure handling DOES
> complete the item anyway — 274's "parent is told the child FAILED" does not produce a failed
> item on this path; it produces a `complete` item whose `result` resolved to something else in
> the parent's `collected_data` (`mark_complete` config is `"result": "handler_result"`; when the
> reply never lands, that path resolves to a foreign value rather than erroring —
> `[INFERRED from config + this capture; the resolution step itself still unread]`). The 274 link
> is live again, now with a reproducible correlation to trace.

Operationally for our lane: both rows are being treated as FALSE completions — the brand-head
work never ran (favicon/og-card 404 throughout) and is being re-filed fresh. Their `complete`
status also counts as strikes under the two-strike rule, which is how a false completion
converts into future suppression of the very re-detection that would have caught it.

---

# CLOSED 2026-08-15 — owner ruling, on option (a), and here is exactly what closes and what does not

**The owner ruled on 2026-08-15 that this file closes and that D1 half two proceeds.** Both
were done; half two is built, committed and registered (`WII-018`, commit `a620912f5`).

## The recorded closure criterion is UNSATISFIABLE, and that is why this needed a ruling

This file's own criterion was: a `hardcoded_section_colors` item **with no `spec.check`** must
reach completion and land `triaged`/`failed`. **Half A permanently removed the traffic that
would demonstrate it** — the design-audit producer was moved to its own `dark_section_audit`
item_type, so no further items arrive on the branch the criterion names. The fix deleted its
own test case. Three options were costed on 2026-08-12 and carried unanswered for three days:
(a) accept the unit + mutation proof and close, recording the unexercised branch; (b) mint one
synthetic row on a throwaway site; (c) leave open, accepting the file no longer describes a
reproducible defect. **The ruling is (a).**

**So, recorded plainly rather than buried: ONE BRANCH OF THE ORIGINAL FIX HAS NEVER EXECUTED
IN PRODUCTION.** The `out_of_scope` disclaimer path for a `spec.check`-less
`hardcoded_section_colors` item is proven by unit test and by mutation, and by a binary probe
that the code is in the running image — **not** by a live item traversing it.
`result->'_verification'->>'status' = 'out_of_scope'` was 0 when last read and is expected to
stay 0 for ever, because the population is gone. **A future reader finding that zero should
not read it as "the fix never worked".**

## What IS proven, and where

| | proof |
|---|---|
| **D3** — `verifier-remit-check`, the daily CLASS detector | `WII-015`; CronJob live at `25 7 * * *`, image `v1.0.1289`, **deployed and run** |
| **D1 half one** — completion gate 1b, refuses a completion the handler never earned | `WII-017`; live `v1.0.1299`, council APPROVED, and **both arms proven on real production traffic with a 1:1 accounting** — 3 items blocked (the deliberate one terminated at `failed` rather than churning), 4 abstained, every abstention matched to its completion to the millisecond |
| **D1 half two** — silence retraction, the path back OUT of `failed` | `WII-018`; built 2026-08-15, council submitted (`54e3b698-3d18-4dd1-9d6f-badec7e331fa`, dispatch verified live), **8 of 8 mutations caught**, and it also extracted the shared helper WII-016's architecture seat asked the third adopter for |

**Before gate 1b, every one of those blocked items would have read `complete`.** That is the
damage this file was opened about, and it is stopped.

## THREE THINGS DO NOT CLOSE WITH IT — do not read this closure as covering them

**1. §D is STILL UNEXPLAINED, and its diagnosis run DIED rather than disagreeing.** Ten of the
fourteen completed rows carry a payload that is not the handler's, and the abstain arm
reproduced the split live at 3:1. The `090` run filed to settle it
(`266be67d-a6e1-4afc-8fc1-84b553b2ea82`) **produced no verdict at all**: [MEASURED 2026-08-15]
its `verdict` step in `llm_call_log` reads `success=f`, `output_tokens` NULL, `response_text`
empty, and `error_message` *"API request failed with status 400 … You have reached your
specified API usage limits"*. **That is a budget failure on 2026-08-14, not a judgement about
§D** — and it is transient: 0 failed and 0 usage-capped LLM calls in the 8 hours to
2026-08-15 09:00Z. **So §D is cheap to re-attempt and has no current candidate mechanism.**
The 274 candidate raised on 08-14 was downgraded the same evening by its own evidence and
`bugs_open/274` §10 does not restore it — a *delivered failure* predicts errored items, not
the `complete` items carrying a foreign well-formed payload that §D actually shows. **Do not
borrow 274's answer because it is nearby and satisfying.**

**2. Half two is INERT ON ARRIVAL, exactly as gate 1b was.** `improvement-sweep` is the only
carrier that dispatches this audit and is `enabled=false` (off 2026-08-14 16:41Z on the
owner's cost decision, measured at **6.0x** baseline), and `site-discovery-rotation-design` is
disabled too. **Nothing will exercise silence retraction until one is re-enabled.** The false-
green bleed is currently paused by the sweep being OFF, not by any of this — the gates are
what make re-enabling it safe, and any claim that they stopped the bleed would be false.

**3. The routing/capability mismatch is NOT fixed and was never this file's to fix.**
`dark_section_audit` items are still dispatched at `color-variable-fixer`, which provably
cannot repair them; they still cycle to `failed`. Half two now gives them an honest exit when
the defect actually goes away, which answers the owner's second standing question — those rows
do not sit `failed` for ever any more — but **a handler that can repair them still does not
exist and is nobody's task.** [MEASURED 2026-08-15] 4 live `failed` rows, and the correct
expected behaviour is that they do NOT retract, because they are genuinely unrepaired.

## Where the knowledge lives now

`WII-015` / `WII-017` / `WII-018` in the concept register (each with its landmines), the
`bugfix_213_verifier_producer_join` standing five, and the two landmines this lane added to
`LANDMINES.md`. `bugs_open/274` and `bugs_open/021` §INSTANCE 2 carry the parts that were
deliberately not adopted here.

Moved `bugs_open/` → `bugs_closed/` with **both paths named on the commit**; verified at HEAD
rather than on disk.

— closed by the `bugfix_213_verifier_producer_join` lane on the owner's ruling

---

## POST-CLOSURE NOTE 2026-08-15 (later the same day) — §D ADVANCED, and this file's own §D claim was WRONG

**The file stays CLOSED.** This is recorded here because the closure note above says §D's answer
belongs in this file, and because part of what it corrects is a claim this file made.

### ⚠ CORRECTION — "neither agent declares a `complete_work_item` step" is FALSE, and it was an artefact of WHERE we looked

The 08-14b handoff and the closure note above both state, as established first-hand, that
*"neither `color-variable-fixer` nor `build-dispatch-loop` declares a `complete_work_item` step
in `agent_definitions.default_config->'workflow'->'steps'`"*, and concluded from it that **the
site binding the `result` input is unidentified**. The re-fired `090`
(`adecf408-1e60-4293-8b22-351ddbb52a08`) cited a config quote that contradicted it, and
**verified first-hand today, it is right and we were wrong:**

```sql
-- what our claim searched — and it is TRUE as far as it goes
jsonb_path_query_first(default_config->'workflow'->'steps',
  '$.* ? (@.action == "complete_work_item")')                    -> (none)
-- the same question asked of the WHOLE config
jsonb_path_query_first(default_config,
  '$.** ? (@.action == "complete_work_item")')                   -> HIT
```

**It exists, one level deeper than we looked** — located exactly:

> `build-dispatch-loop` → `workflow.steps.` **`process_item`** `.config.` **`sub_workflow`**
> → `mark_complete_step`:
> `{"action": "complete_work_item", "config": {"result": "handler_result",
> "work_item_id": "current_item.id"}, "next_step": "done", "output_field": "item_completed"}`

**So §D's open question is ANSWERED: the binding is `result: handler_result`.** A top-level
`workflow.steps` search cannot see a step nested inside another step's `config.sub_workflow`,
and a `$.**` search takes about the same time to write. This is the "a grep proves absence only
for the SPELLING — and the PATH — it searches" shape, and it cost this lane a week of calling
§D's mechanism unidentified.

### What §D now has, and what it still does not

**Verdict: `UNVERIFIABLE`** (the loop's own label — it did not confirm a mechanism). But its
citations move the question a long way, and two of them are worth carrying:

- **[VERIFIED first-hand today]** the binding above.
- **[THE LOOP'S CITATION, NOT INDEPENDENTLY VERIFIED]** the two foreign payload shapes match two
  other agents' declared `complete_workflow` `output_fields` exactly — `webdesign-agent`'s
  `design_spec` (`color_scheme typography spacing design_notes`) and `content-gap-planner`'s
  `gap_plan` (`add_to_page approach new_page not_actionable reasoning retype_existing
  update_spec`). That is a much stronger lead than "some foreign payload": it names two specific
  producers.
- **[UNVERIFIED LEAD]** the loop's final citation points at
  `platform/orchestration/datahelpers/action_inputs.go` `ExtractActionInputs` /
  *"ExtractFields uses aggressive recu[rsion]"*. If `handler_result` resolves by aggressive
  recursive search, a sibling agent's response elsewhere in `collected_data` could satisfy it.
  **Nobody has read that function for this purpose yet. Do not write this up as the cause.**

**Still NOT established:** how `handler_result` comes to hold another agent's payload. The
candidate above is a direction, not a finding. And it remains the case that `bugs_open/274`'s
mechanism is NOT the answer here — a delivered failure predicts errored items, not `complete`
items carrying a well-formed foreign payload.

### Whoever picks this up

It is now a small, well-posed question: **read `ExtractActionInputs`/`ExtractFields` and decide
whether `handler_result` can resolve to a non-handler's payload.** That is a code read, not a
diagnosis run. If it can, the fix is a scoped binding, and gate 1b's abstain arm is already the
instrument that will show it stopping.

### §D UPDATE, same day — THE MECHANISM IS READ AT SOURCE, AND IT IS ALREADY SOMEBODY'S RFC

The code read named as "the whole of the remaining work" above was done. **§D is an incident of
`architecture_review/RFC_029`** — the aggressive recursive search having no boundary — filed
2026-08-14 by the `staged_component_build` lane out of `bugs_open/248`, and **RULED by the owner
on 2026-08-15** ("unique-or-nothing"; implementation OPEN, not started).

**The chain, every step verified at source:** `complete_work_item` declares `result` as Optional
→ `build-dispatch-loop` maps it `{"result": "handler_result"}` → `IsDottedPathReference` is
literally `strings.Contains(s, ".")`, so **Strategy 0 skips a single-segment mapping** →
Strategy 2's `ExtractFields` runs `findFieldRecursive` for **any key named `result`**, to depth
20 → and `ExtractActionInputs`' Strategy 4, the arm that exists to resolve exactly this shape,
**skips because the field already has a value**.

**Two things this adds that the RFC did not have**, both contributed into their file rather than
filed here:

1. **A different entry condition.** RFC_029 frames the trigger as *a field the caller never
   mapped*. Here the caller **did** map it, correctly, and lost because the value had no dot.
   Strategy 0's own comment says it was added to stop the aggressive search winning — it fixed
   that for dot-paths only.
2. **⚠ The ruled remedy may not cover this case.** "Unique-or-nothing" defends against
   *ambiguity*. If the only bare `result` in `collectedData` is the foreign one, it is unique —
   so the remedy resolves it wrongly, with full confidence. This is not an ambiguous read; it is
   a confidently wrong one.

**Still [NOT VERIFIED]:** that a foreign bare `result` was actually present when those 10 rows
completed. The mechanism is *available*; it is not proven to have *fired*. Confirming it needs a
live `collectedData` capture that nothing currently retains — but **gate 1b's
`NO_CHANGE_GATE_UNREADABLE_RESULT` stream already logs the offending payload's top-level keys**,
so if RFC_029's fix lands, that stream going quiet is a free before/after.

**Nothing further is owed by this lane.** §D's answer now lives where the fix will be made.
