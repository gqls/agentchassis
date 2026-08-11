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
