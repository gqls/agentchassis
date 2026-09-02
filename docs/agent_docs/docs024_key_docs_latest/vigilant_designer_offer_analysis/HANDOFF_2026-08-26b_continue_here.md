# HANDOFF — vigilant designer + offer analyser (2026-08-26b)

> ## ⚠ SUPERSEDED FOR CURRENT STATE — 2026-09-02
> **Cold-start from `HANDOFF_2026-09-02_continue_here.md` (this directory) instead.**
> §H1c / §H1d / §H1e below are **discharged**: the producer register gate was built, reviewed twice
> (`4054f4d9`, **APPROVED round 2**), applied, and is live and repairing.
> **This file is STILL the authority on everything the new one does not mention** — §C (imagery
> supply), §E (residuals), §G (predicate vocabulary / gate 1c), §H2b (carousels), §H3 (the logo
> chain) and **§H4 (the axis finding in full, which the next build depends on)**. Do not delete it;
> the new handoff points back here by section.

**COLD-START = this file + `HANDOFF_2026-08-26_continue_here.md` (still correct on everything except
§1 and §4, see below) + `PLAN_2026-08-25b` §8 **and §8f** + register `IMG-074` / `WII-033` / `CLM-024`.**

**Supersedes `HANDOFF_2026-08-26_continue_here.md` on TWO points only:**
- **§1 (the imagery fix) is DONE — and its predicate as written there was WRONG.** See §A.
- **§4 (ask the owner) is ANSWERED.** See §B.

Everything else in that file — §2 (gate 1c's unreachable negative control), §3 (supply), §5
(carried-forward work), the whole Watch-outs list, the Residuals — **stands unchanged and still
applies. Read it.** Its "re-run every number before acting" instruction was tested today: every
figure in it re-measured true, and the fleet still moved under me mid-measurement (see §D).

---

## §A — THE IMAGERY FIX SHIPPED, AND IT WAS TWO HALVES, NOT ONE

**Migration `644_planner_sees_imagery_and_illustrated_block_sources_an_illustration.sql`** — applied
2026-08-26, recorded in the ledger, **live now** (DB config, no image, no inert window). Register
**IMG-074**. Commits `d10952b3b` (migration + register) and `b3bddba60` (LANDMINES ×2).
~~`Council-Submitted: 08477888-...` — ⚠ **VERDICT NOT YET READ. That is the first thing you owe.**~~
**CORRECTED, SAME SESSION: the verdict is IN and it is APPROVED** — round 1, corr
`08477888-b3e6-4ceb-911d-6e2a3c446755`, 14 seats, **4 advisory objections, none high-severity**.
**Read, and each one answered with a measurement**, recorded in full in register `IMG-074`. No amend
is needed or permitted: `098` credits the `Council-Submitted:` trailer automatically now the
correlation is approved. **Do NOT write a `Council-Reviewed:` trailer onto a new commit for it** —
that would be a second claim on a change already credited.

Three objections resolved in the change's favour with evidence (no `BEGIN…COMMIT` in the *sketch*
but present and mutation-proven in the *file*; **zero** Go consumers of `component_expresses`, so the
fixed-arity worry cannot arise; both repointed fields confirmed `required:false` +
`skip_field`, which is the exact condition the `guidelines` seat made its approval depend on). One
was a fair paperwork gap with a decisive answer (`content_shape`/`visual_density` are NULL on all
three components in question, so they could never have carried this signal). **Two are worth
carrying forward and are in `IMG-074` and §E below.**

```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
 WHERE correlation_id='08477888-b3e6-4ceb-911d-6e2a3c446755' AND kind='council_report' ORDER BY created_at;
```

A REVISE/REJECTED must be acted on: **the change is already live on the shared branch AND applied to
the database.** There is no inert window to hide in.

**Half 1 — the missing word.** `component_expresses` gained a fifth `image` arm. It fed **three**
live planner menus four tokens and none was an image, so `Generic Text Block` and `Illustrated Text
Block` read identically and the planner picked the plain one (208 instances across 23 sites vs 6 on
one). Not a missing component — a missing WORD.

**Half 2 — the source.** `Illustrated Text Block.image_url` moved off `site_assets.image` (which
**aliases to the page's own hero, unconditionally**) onto `site_assets.illustration`; `image_alt`
moved off it onto `llm` (it was typed `text`, so the resolver handed it the image URL for a screen
reader to read out).

> ### ⚠ THE PREDICATE IN THE PREVIOUS HANDOFF AND IN `PLAN_2026-08-25b` §8c WOULD HAVE CANCELLED THE FIX
> Both propose `source = 'site_assets.image'`, **exact equality**. Half 2 moves the field off that
> value, so the exact predicate makes `Illustrated Text Block` invisible again — the two halves
> cancel, each provably correct alone, and **every guard still passes**. The shipped predicate is over
> `site_assets.%` (excluding `logo`), and 644 asserts BOTH ends so the cancellation cannot commit.
> Full correction: `PLAN_2026-08-25b` **§8f**.

**⚠ It was also preventing live damage, on a clock I did not know about.** apis.uk/index has an
active `hero_home`, so the alias resolved there, and *live resolution beats `carryStored`* — its six
distinct illustrations were one `plan_sections` run from becoming six copies of `hero-home.jpg`. The
apis.uk session confirmed on receipt that the page was at `needs_rebuild` **with a `stale_chrome`
re-render wave imminent**. Hours, not "someday". **Telling the owning lane is what dated the risk;
I could not have.**

## §B — THE OWNER'S ANSWER (2026-08-26)

Both questions from the previous handoff §4 are settled:

1. **"Between paragraphs" — section-level placement is ACCEPTABLE.** He answered *"either, whichever
   ships"*. **So a strictly mid-prose component is NOT owed.** Do not build one on the strength of
   the original wording; that question is closed.
2. **He chose to fix the source before shipping**, over shipping the one-liner or holding for supply.

## §C — WHAT TO DO NEXT, in order

1. **Read the council verdict** (query above). Act on REVISE/REJECTED — it is live and applied.
2. **Verify at the artefact once a planner has actually run.** ⚠ **A zero is NOT failure.** Nothing
   re-plans a page on its own, and a site with no `section/illustration` row renders the block as
   plain prose *by design*. **Read the DEMAND side first:**
   ```sql
   SELECT count(*) FROM orchestration_states
    WHERE owner_agent_type IN ('build-site-planner','site-planner','content-gap-planner')
      AND created_at > '2026-08-26 11:00+00';
   ```
   Only then ask whether any planner chose `Illustrated Text Block`. **apis.uk is the regression test
   case** — its six values must still be six distinct illustrations after its next rebuild.
3. **THE SUPPLY QUESTION — this is now the whole of the imagery ask, and nobody owns it.** §1 created
   no assets. `[MEASURED 2026-08-26]` the pipeline generates **heroes** (206 active across 28 sites)
   and **icons** (139 across 19) and barely any **illustrations** (26 across 5 sites; only **4**
   `section/illustration` plan rows across 3 sites). **This is the real answer to "why don't pages
   have pictures in them" — not placement, generation.** Bigger than today's work. Say the supply
   figure alongside any imagery claim or you repeat this lane's own `bugs_open/395` shape.
4. Everything in the previous handoff §5 (v2 batch, `features_open/034`, the `HandlerCanWriteField`
   drift audit) is **untouched and still carried forward.**

**No SUMMARY written today, deliberately.** Half the imagery ask shipped and the bigger half (supply)
is unowned, so "where we are now" is mid-stream — the rarity rule says wait for the turn rather than
file a near-duplicate of `SUMMARY_2026-08-25`. **Write one when supply is decided**; that is the
inflection.

## §C-bis — A NEW FINDING LANDED AFTER 644 APPLIED, AND IT IS NOT MINE TO PATCH

**`check_image_source_unsatisfiable` states a reason it never verifies, on ~87% of its own open
queue.** Raised by the apis.uk session within the hour: 644's repoint caused one
`image_source_unsatisfiable` item at `needs_human_review` for their index page, because
`site_assets.illustration` resolves nothing there — **which is the protection working**, since the six
illustrations now survive *because* the source resolves nothing and `carryStored` fills the field.

Their call of "true statement, false alarm" was right, and the defect is far wider than 644:

- The check emits `spec.reason` = *"…the field renders empty or falls back to a placeholder"*. The
  first clause is true; **the second is a claim about RENDERING that the check never checks.** Its
  satisfiability arms (literal asset keys, plan imagery rows, `ImageRoleForPath`, the
  `content_data` hero_url/logo_url fallback) and `carryStored` are separate code paths that never
  consult each other. The check reads **schemas**, not **values**.
- `[MEASURED 2026-08-26]` of **67** open items, **58** name a field that IS populated on that page —
  `site_assets.hero` 46/46, `site_assets.image` 12/8, `site_assets.illustration` 9/4. Verified with an
  exact join on site + page name + `content_components.function` + the named field; loose and strict
  joins agree. **644 contributed ONE of those 58; the hero rows are the bulk and long predate it.**
- ⚠ **The fix is NOT "treat carry as satisfied".** That would hide the supply gap, which is the thing
  most worth seeing. The two states want *distinguishing*: (a) unsatisfiable AND nothing carried —
  the real `bugs_closed/238` defect; (b) unsatisfiable BUT carried — renders fine today, renders
  empty on any NEW page with the same component. An accurate reason and a severity split, not
  suppression.

~~**Status: diagnosis loop IN FLIGHT** … **Read it, then file `bugs_open/`.**~~
**CORRECTED, SAME SESSION — the loop returned and the case is FILED as `bugs_open/411`.**
Verdict **CONFIRMED, first iteration** (`RUN_CORRELATION_ID=8aeba0b6-0508-4059-8a10-b3e94211dd8c`),
independently re-reading the same functions, citing the same lines, and surfacing a live example I
had not seen (`hero` / `background_image` on `guide-how-loans-are-calculated`). ⚠ **The verdict is in
the work item's `result`, NOT in a `diagnosis_artifacts` row** — that correlation has three `bundle`
rows and no verdict artifact, so reading `diagnosis_artifacts` alone would have told you the run
produced nothing. Read
`site_work_items.result->'response'->'response'->>'conclusion'` where `item_type='needs_diagnosis'`.
The transferable pattern is in **`016b` §9**; the index row is in **§10**.
**I did NOT touch the checker**: shared Go discovery check, council scope, inert until a roll, and
*what the check should mean* is a decision rather than a patch. **The apis.uk item is left open as
evidence** — do not cancel it. Prior art checked: `bugs_open/356` names this check only in passing
(different axis, retired pages); nothing covers the carry blind spot.

## §D — NEW WATCH-OUTS (the previous handoff's list all still stands)

- **⚠ A CHANGED-ROW COUNT CANNOT TELL A WIDENING FROM A RESHUFFLE.** A variant arm that also
  suppressed `list` changed **the same 9 rows** while 3 silently lost a capability. Assert the SHAPE
  (nothing loses a token; nothing changes but by gaining the new one) and **induce both**. Now a
  LANDMINE.
- **⚠ DO NOT ASSERT A LITERAL COMPONENT COUNT in a migration guard, and do not take BEFORE/AFTER as
  two snapshots.** Five components were created by another lane *during this change's measurement
  run*; the two-snapshot control broke outright (381 vs 386). Compute both sides in ONE query. 644's
  guards are structural and population-independent for this reason.
- **⚠ `site_assets.<path>` IS ALIAS-RESOLVED, NOT A LITERAL KEY.** `image`, `background`, `banner`,
  `header_image`, `product_screenshot`, `screenshot` and more all mean **`hero`**. Nothing in a
  component's schema, template or guidance says so. LANDMINE entry has the check.
- **⚠ A REAL-LOOKING PER-SECTION VALUE IN `content_data` IS NOT PROOF THE SOURCE WORKS.** It may be
  hand-seeded and surviving only by `carryStored`. This is how six rows of live data appeared to
  refute a correct code reading — `WRONG_CALLS.md` 2026-08-26e. **When a code reading and a data
  sample disagree, widen the population until the sample could have come out either way.**
- **⚠ A CONTROL THAT ERRORS AND A CONTROL THAT PASSES ARE THE SAME NUMBER** if you only read the
  number. An `awk` guard here failed to parse and printed `0`.
- **⚠ `run-migrations.sh` (even a bare dry run) TIMES OUT** — it probes all 1,052 files. Read
  `schema_migrations` directly, and note lanes apply their own file by hand and out of order (642
  landed before 636), then `--record-only`.
- **⚠ `component_expresses` HAS NO GO CONSUMERS** — `grep` finds nothing. Its three consumers are
  embedded in `agent_definitions` config; the `jsonb_path_query` to find them is in the RUNBOOK.

## §E — RESIDUALS FROM TODAY, stated plainly

1. **Supply (§C3) is unowned** and is the larger half of the owner's actual ask.
1b. ⚠ **GATE 1C'S UNREACHABLE NEGATIVE CONTROL NOW HAS A ROUTE — see §G. This is the biggest change
   to this lane's blocked work in the file, and it did NOT come from this lane.**
2. **`section/illustration` resolution is FIRST-WINS BY KIND**, so several illustrated sections on one
   page all resolve to the SAME image. apis.uk has routed around it (`content_data` + lock) and has
   **offered itself as the worked test case** — six distinct instances — if anyone builds per-section
   mapping.
3. **6 `site_plan_imagery` rows at `scope='page', kind='illustration'` are read by NO resolver arm**
   and are inert (5 apis.uk, 1 pool-energy-utilities.internal). Named, not fixed.
4. **llm-authored alt text for a server-resolved image is a hallucination surface** — a model
   describing a picture it cannot see. True of all 13 existing alt fields, the estate's settled
   convention; this change brought the one outlier INTO it rather than overturning it.
5. **The heroes-included judgement is arguable and is deliberately on the record** — `hero` and
   `product-hero_pre_037` now advertise `image`. Only `site_assets.logo` is excluded. A council seat
   may reasonably want banners excluded too; the submission says so in terms.
6. **Every behavioural claim in IMG-074 is unverified** — nothing has re-planned a page yet.
7. **644's repoint `UPDATE` is gated on the fields EXISTING, not on their CURRENT source value**
   (`debug_historian`, medium). A re-run would clobber a later hand-edit back to 644's values. **Not
   fixed and not fixable in place** — the file is applied and recorded, and the runner supersedes by
   the next number rather than editing a recorded file. If it ever needs re-applying, add
   `AND input_schema->'fields'->'image_url'->>'source' = 'site_assets.image'` (and the alt
   equivalent) so a re-run is a 0-row no-op.
8. **The eight OTHER components still declaring `site_assets.image` remain exposed to the alias
   trap** (`bug_historian`, medium) — `case-studies-grid`, `content-block-about`,
   `featured-inventory`, `game-master-explanation`, `product-details_pre_037`, `product-specs`,
   `tool-guide-intro`, plus inactive `hero-headline`. Deliberately not repointed (several may
   legitimately want the hero; eight live components is a different blast radius), so **the mechanism
   is guarded by a LANDMINE rather than by code — the weaker form.** ⚠ The architecture seat set the
   trigger independently: **a THIRD component hitting this trap is the point to ask whether
   `imageRoleAliases` needs an explicit opt-out** instead of another avoid-the-landmine repoint.

## §G — RELAYED OWNER RULING (via the `bugs_open/395` lane, 2026-08-26): THE PREDICATE VOCABULARY MAY WIDEN

⚠ **SECOND-HAND. I did not hear this from the owner directly** — it was relayed by the
`bugs_open/395` session, which heard it. Treat it as a strong lead to confirm, not as a ruling I
witnessed. **I deliberately did NOT start the work on it.** Recorded here because a cross-lane
agreement that lives only in a chat message dies when either session closes.

**The ruling as relayed:** the acceptance-predicate vocabulary may be WIDENED, read-only. The owner's
words were *"is it ok for the tests to read real content but not to write it"* — and the answer is
that a predicate is already read-only by construction, so the constraint is satisfied as it stands
and what is being approved is what a predicate may READ.

**Why this matters more than it sounds, and it is exactly this lane's residual 1.** `[VERIFIED
2026-08-26, first-hand]` `acceptancePredicateTextFields`
(`verify_acceptance_predicates_action.go:239`) admits **`meta_description` and `title` ONLY** — and
both are unwritable on the audit-routed path. So **every predicate this producer can currently emit
is doomed at birth**, which is why they all refute and why gate 1c's `outcome='permitted'` has never
been reachable. **Page body / section content is different: `page-content-writer` can actually write
it.** A body-content predicate is therefore the first one that a handler is CAPABLE of satisfying.

> **⚠ NECESSARY, NOT SUFFICIENT — keep these two apart, and do not let the next reader merge them.**
> *(Correction supplied by the `bugs_open/395` lane, 2026-08-26, sharpening its own relay: it first
> said "the first route to a satisfiable predicate", which reads as a promise.)* Widening the
> vocabulary removes an **impossibility** — today the criterion is unreachable, so a refusal says
> nothing about the handler. It does not create a **success**: gate 1c's `outcome='permitted'` still
> requires a handler to actually produce a repair that satisfies the criterion and be graded on it.
> **That is the content-generation gap the council's `constitution` seat flagged on the gate-1c round
> and NOTHING HAS TOUCHED IT.** So v2(a) is a precondition for a live negative control, not a
> delivery of one — and it is emphatically not on its own grounds for promoting the gate from
> recording to refusing.

**⚠ IT IS BLOCKED ON THIS LANE'S OWN v2(a), NOT ON THEM.** `bugs_open/395` §8f and §5 both record
that body-text shapes are excluded today *because the page surface the model authors against carries
no content*. The piece that changes that is **v2(a) in `features_open/030` §10** — the bounded
head-of-hero excerpt — which is carried-forward work in the previous handoff §5, still unstarted,
migration `602` unwritten. ⚠ **Re-read `features_open/030` §10 first:** v2(a) GROWS the offer surface
and **widens what a predicate can address**, and the truncation check must be re-run on
`webdesign.co.uk` afterwards.

**⚠ A BUILD-BREAKING LOCKSTEP FIRES THE MOMENT YOU WIDEN, BY DESIGN. `[VERIFIED 2026-08-26,
first-hand]`** `TestPageFieldWritersCoversThePredicateVocabulary`
(`write_audit_findings_field_capability_test.go:162`) is **bidirectional**: it reads the evaluator's
own set rather than mirroring it, and **fails the build** if a vocabulary field has no
`pageFieldWriters` entry (and also if the roster carries a field no predicate can name). A sibling,
`TestPageFieldWritersEntriesCarryTheirEvidence`, additionally requires the entry to carry a `Why`
with a real writer census — so the cost is a **dated measurement**, not a line.

Two ways to satisfy it, and **the 395 lane has explicitly offered to do whichever this lane prefers,
same day, so their test does not block the widening**:
1. **add the field to `pageFieldWriters` with `WritableBy: {"page-build-handler": true}` and its
   measurement** — their preference and **mine**: it keeps every field in the vocabulary carrying a
   dated statement of who can write it, which is the property that made routing rule 3b possible at
   all. **Relayed to them as this lane's choice.**
2. relax the lockstep to exempt writable fields — **rejected**: it removes the coverage guarantee on
   exactly the population the roster exists to describe, to save writing down one measurement.

⚠ **Do NOT add a body-content entry to `pageFieldWriters` speculatively, ahead of the widening** —
the reverse arm of that same test fails on a roster entry no predicate can name.

**What is owed:** ping the `bugs_open/395` lane when the vocabulary is actually widened. Nothing else
is owed to them; their DECISION 1 (overwrite authority, ruled option (c) — machine-written
descriptions only) is **not this lane's seam** — they are building the provenance stamp in
`save_page_meta_description_action.go` and the two fill-blank upserts. ⚠ Worth knowing anyway,
because it bears on every predicate this lane emits: **`pages` has no provenance column at all**, so
the distinction the owner drew cannot currently be made by the system; with 838 live descriptions and
none hand-written, option (c) today covers all 838.
**Status as of 2026-08-26: RULED but BLOCKED — not on approval, on a dependency.** The marker that
would create that provenance is **`bugs_open/403`'s** (leopardess lane, active today); the 395 lane
has asked them which direction their marker takes and whether it covers a plain COLUMN as well as
`content_data`, and is deliberately **not building a second one**. ⚠ If this lane ever needs
provenance on a page column, ask there first rather than minting a third.

## §H — TWO OWNER-ROUTED ASKS ARRIVED 2026-08-31, BOTH RELAYED, BOTH AWAITING HIS DECISION

⚠ **Both reached me through peers, not from him. I corresponded (which is what one of them
explicitly instructs) and MEASURED; I wired nothing.** Both remedies are config changes on shared
writers/planners — council scope, fleet-wide on every subsequent build — so they are his call.

### H1. Hero copy / benefit framing (via `copy_quality_two_stage`)

Owner instruction names this thread: correspond on *"what sort of approach we can use with these
hero titles and copy… in terms that clients can see how it might work for them"*, with the
constraint *"we mustn't presume to know what they want"*, and *"if each piece of copy requires
discussion between agents then so be it"*.

**What I established — the gap is exact.** `[MEASURED 2026-08-31]` `offer_ordering` and `strategy`
are `is_current` on **32 sites** (enrolment moved from 13; my own figures were 5 days stale).
`offer_ordering` holds **187 `lead_with` points**, each with `rank`, `from_field` provenance,
`differentiated` and a `why`. **And of `page-content-writer`, `build-site-planner`, `site-planner`,
`content-gap-planner` and `page-build-handler` — NONE reads `offer_ordering` or
`satisfaction_condition`.** Only the producer (`domain-strategist`) and my `offer-analyser` do.
**The judgement is derived, ranked, provenance-stamped and read by nobody who writes a hero.**

⚠ **Sharper: his objection is already written down, as a prohibition, on the site he rejected.**
`finetuning.uk`'s `avoid_leading_with` contains *"Page counts, tool counts, or inventory size before
the reader's problem is named"* — and *"Real projects, described plainly"* is a label on an
inventory. **Not a missing judgement. An unenforced one.**

⚠ **AND THE SET IS NOT SAFE TO USE AS A SOURCE YET — measured against their
`BANNED_REGISTER_v1.json`, NOT against my guess.** **51 of 187 points (27%)** carry banned register,
across 23 of 32 sites; **7 of 31 rank-1 points (22%)**, rank 1 being the hero candidate.
> **⚠ MY FIRST FIGURE WAS 18/187 (~10%) AND IT WAS WRONG BY 2.8×.** I inferred the rule from their
> prose and caught only the WORDS (`plainly` 7, `honest*` 4). **The SHAPES dominate**: `x_not_y`
> **28**, `rather_than` 13, `not_just` 6, `negative_reveal` 2, `instead_of` 1. Which is the whole
> lesson: **the corpus is a demonstration reservoir of the exact comparison construction their canary
> proved the writer keeps producing** — so feeding it as a mandated phrase chain would re-teach that
> shape through the channel they measured as most effective. **Do not carry my 10% anywhere.**

**Two rollout facts that cut against both lanes' instinct:** **9 of 32 sites have ZERO dirty
points** (a clean population exists without waiting for the pass), and ⚠ **`finetuning.uk` is the
WORST site (6)** — so "finetuning first", which both lanes proposed, is the hardest case, not the
easiest. Choose deliberately.

**Their side is settled and committed** (`c0ab7e7c1`): banned register as versioned data with a
usage rule (structured input only, never pasted as prose into a critic window — my caution, their
`prompt-text-poisons-its-own-detector` lesson); **rank PINNED** during the register pass, with
substantive changes returned to this lane for re-judgement rather than quietly rewritten; and v1
binds the **served `point` only**, not the `why`. Packaged for the owner as **DECISION C**.
⚠ **Their morning benchmark ran `claude-fable-5` on the worst canary section's production prompt:
0 and 0 negation constructions against shipped sonnet's 5.** So if the writer changes, **my 27% is a
statement about the EXISTING corpus only** — re-measure after any model change; it is dated for that
reason.

### H1b. ⚠ OWNER RULED C — AND THE WASH EXPOSED THAT THE PRODUCER MINTS DIRTY AT 23%. **THIS IS THE OWED WORK.**

**Owner, 2026-08-31 (relayed):** *"Decision C: Please go ahead and wire in the benefit priorities."*
Executed in the agreed order by `copy_quality_two_stage`; migration **667** applied the 41 repairs I
ACKed (identity-by-exact-text, RAISE on drift, rank untouched, backups + rollback).

**My ACK gated that write and I excluded 10 of 51 as substantive.** Rule, re-derivable: *a repair
removing ≥40% of a `differentiated: true` point has removed the differentiating clause, not a
flourish.* ⚠ **Ruling 7's truncation systematically strips the distinguishing half**, because in an
`X, not Y` construction the differentiation lives in the **Y** — `[MEASURED]` **51 of 51 repairs were
shorter, corpus −24.8%**, mean −28.7% on differentiated points. All 10 are now resolved: 7 accepted,
1 pending a "the full record" → "the record" wording fix, **3 re-derived by me from each site's
`strategy` provenance and battery-checked clean.**

> **⚠ THE SYMMETRIC LAW, worth carrying: TRUNCATION LOSES MEANING, EXPANSION MANUFACTURES IT.** Their
> expansion pass invented three operating-history claims (*"the full testing procedure"*, *"a visible
> label"*, *"the timing for each region"*) — **the `evidence_base` class, and worse in served copy
> than any register fault.** Each repair fails in the direction of its pressure. My own re-derivation
> found that agritec's original claim **was** grounded and only the *mechanism* was invented, which is
> the tell to look for.

### ⚠⚠ H1c. THE FINDING THAT MATTERS MOST, AND IT IS THIS LANE'S TO FIX

**`[MEASURED 2026-08-31]` the producer wrote 799 `lead_with` points in 6 days and 184 (23%) were
BORN DIRTY.** Today: 17 minted, **5 dirty (29%)** — verified by timestamp as producer output
(`lampenkap` r3+r4 at 10:23Z, `fundamentallyai` r5 at 10:29Z), hours before the wash. Minting:
`x_not_y` **81**, `rather_than` **53**, `plainly` **29**, `honest*` **24**, `not_just` **20**,
`negative_reveal` **10**.

**So the wash repaired 51 points against a mint running ~130/day at 23%.** Today's rate is
indistinguishable from the historical rate — **nothing has changed its behaviour, because the ban
lives in a JSON file the producer has never been told about.**

> **⚠ THIS RETRACTS A GATE I MYSELF IMPOSED, and the next session must not reinstate it.** I told the
> copy lane *"the corpus must be clean on both axes before fleet-wide wiring"*. **That condition is
> true at an instant and false by the next regeneration — it cannot be held.** The gate is not a
> corpus state but a **PRODUCER PROPERTY**: the mint stops printing the shape, or the wash runs
> continuously as a post-step.

> **✅ BUILT 2026-08-31 (evening) — SEE §H1d BELOW. The paragraph that follows is left standing
> because its reasoning is still the reasoning; only the word UNBUILT is now false. ⚠ It also
> names the wrong producer — it is `offer-analyser`, THIS LANE'S OWN AGENT (164/164 rows), not
> `domain-strategist`, which writes `aspect='strategy'`.**

**OWED, UNBUILT, THIS LANE'S SEAM:** a post-generation check on the producer's `lead_with` output
calling **the same scanner the copy gate uses** — `ScanDefineByNegation` (platform code, 7 shapes as
of Decision B, inert until the roll but the source is current) **plus `BANNED_REGISTER_v1`** — so the
two mints and two batteries cannot drift (the dedup-index/Go-list lockstep, applied *before* the
drift). Cite file+version; the copy lane version-bumps in step. ⚠ **Two constraints I committed to:**
it must **FAIL LOUD, never filter silently** (a silent filter makes the producer's rate unmeasurable,
and that rate is the only evidence the fix works), and **it must not be reported as working until
measured against a FRESH MINT — 23% is the baseline it has to move.** ⚠ And their hardest-won lesson
now binds my half: **demonstrations govern, instructions do not** — adding *"don't write X, not Y"*
to the producer prompt is an instruction, and their canary is the evidence instructions lose.

#### ⚠ H1c-i. ~~THE RACE, NOT DESTRUCTION~~ — **THIS SECTION WAS WRONG. I WITHDREW IT. Read H1c-ii FIRST.**

Migration `668` (the 10 expansions) **refused to apply**, correctly, because `fundamentallyai` r2's
FROM matched zero rows. The copy lane diagnosed it as *"the mint overwrites applied wash work
whole-row"* and was taking that to the owner as evidence for Decision E. **The timestamps invert it:**

| event | time |
|---|---|
| producer regenerates `lampenkap.com` | **10:23:30Z** |
| producer regenerates `fundamentallyai.com` | **10:29:36Z** |
| **migration 667 applied** | **10:34:30Z** |

~~**The regeneration PRECEDED 667 by 5 and 11 minutes** … **667 never washed those three points** …
**NO APPLIED WASH WORK HAS BEEN DESTROYED.**~~ ⚠ **ALL OF THAT IS FALSE — WITHDRAWN 2026-08-31. See
H1c-ii.** The timestamps quoted are real; the inference from them is not.

> **The conclusion holds and the argument is better without the overstatement: the wash went STALE
> BETWEEN EXTRACTION AND APPLICATION**, on 2 of 32 sites, in a few hours. **A race the wash cannot win
> by running faster** — which makes the case for gating the mint without claiming committed repairs
> are being destroyed. `668`'s guard catching it is the system working.

#### ⚠⚠ H1c-ii. WHAT ACTUALLY HAPPENED — a FALSE GREEN, and my correction above was itself wrong

**667 DID apply.** Proven at the artefact after the copy lane contradicted me: the **superseded**
`08-28` row holds **WASHED text at ranks 3, 4 and 5** and the **original** at rank 2 — exactly the
signature of a successful wash where r2 was my own exclusion. **My "never washed" claim is
WITHDRAWN.**

**Why my instrument could not see it, which is the transferable half:** I queried `is_current` rows
by `created_at` and reasoned *the current row predates the migration and there is no newer row,
therefore it did not write.* **That inference holds only on a table that UPDATES. `site_specs`
SUPERSEDES** — insert new, flip the flag — so a successful write leaves **no trace in any current row
and no new row after the migration's timestamp.** The premise was a property of the table I never
checked.

**The true sequence, from three artefacts including 667's own in-transaction backup:**

| event | time |
|---|---|
| 667 backs up `fundamentallyai` (holds the OLD texts — the artefact that settled it) | **10:28:41Z** |
| producer inserts a new row, superseding the one 667 is repairing | **10:29:36Z** |
| **667 COMMITS** — guards passed, success NOTICE, ledger stamped | **10:34:30Z** |

⚠ **THE NAME FOR THIS IS A FALSE GREEN: success everywhere, live effect nowhere, no error anywhere.**
38 of 41 points live; **3 washed inside a row with `is_current = false`.** The migration succeeded and
changed nothing live on that site. **A wash cannot even KNOW whether its work is live while the mint
writes** — which is the final form of the Decision E evidence, adopted by both lanes.

**REQUIRED PRACTICE for the final wash pass** (agreed with `copy_quality_two_stage`, from my fix
shapes): **re-assert `is_current` at COMMIT time**, or **advisory-lock the producer** for the
transaction. ⚠ **A `RAISE` on zero matches does NOT help** — 667's fired correctly and the match was
real when it fired.

**Full trap:** `LANDMINES.md`, *"an exact-text migration against a versioned config table can pass
every guard…"* (`a01bbce1a`). **My own miss:** `WRONG_CALLS.md` 2026-08-31 — I sent the withdrawn
claim as an *urgent hold on evidence bound for the owner*, and **"my instrument disagrees with yours"
is not evidence that mine is sound.** Both lanes reached confident opposite wrong mechanisms; only
the artefact between them settled it.

**Row-level corrections:** **5** rows no longer match the wash doc, not 3 — the fifth is
`lampenkap.com` r3. Two of the five (`fundamentallyai` r2, `lampenkap` r3) are **exclusions** whose
*originals* were regenerated away, so they were never 667 candidates.
⚠ **CONSEQUENCE: two of the verdicted 10 already have stale FROMs** — including one of my three
provenance re-derivations, which must be re-derived again against whatever the row says once the mint
is gated. **Do not re-base the 10 before the producer check holds, or you re-base against text that
moves again.**

**Wiring status:** `vetcomparison.uk` first (one of the 9 never touched by the wash, so clean on both
axes **and** not dependent on the wash holding); the **8 sites carrying excluded rows stay held**;
fleet-wide waits on the producer-side mechanism, not on a cleanliness claim. **Council scope and the
owner's call.**

### ~~⚠⚠ H1d. THE PRODUCER GATE IS … WIRED TO NOTHING~~ — **THE GATE IS WIRED. CORRECTED 2026-09-02 19:4xZ.**

> **⚠ THIS SECTION WAS TRUE WHEN WRITTEN AND WAS FALSE 27 MINUTES LATER. It is struck rather than
> deleted because the reasoning is still worth reading and the standing instruction at the bottom
> still holds.**
>
> `[VERIFIED 2026-09-02]` `offer-analyser` now has **eleven** steps and one of them is
> `repair_ordering_register` — read off the live `agent_definitions` row, not inferred. The chain is
> `verify_ordering_cardinals → repair_ordering_register → write_offer_ordering`.
>
> **What happened, and it is the two-sessions-one-lane failure rather than anyone's error.** My
> commit `fcb210e65` is stamped **16:47:21Z**. The arming migrations applied at **17:14:53Z**
> (`681_…_HOLD.sql`) and **17:18:22Z** (`682_offer_analyser_register_gate_ai_service.sql`), by a
> SECOND SESSION OF THIS SAME LANE working the same problem in parallel. Both of us reached the same
> diagnosis independently and correctly; they then wired it. **Nobody was wrong — we were each right
> at a different instant**, which is precisely what duplicate sessions on one lane produce.
>
> ⚠ **The `25% born dirty over 254 points` figure below STANDS and is the better baseline** — it is
> the pre-gate rate. The sibling's post-gate window is **1 dirty of 12 = 8.3%**, and they were right
> to refuse to quote it as improvement: `P(≤1 of 12 | rate unchanged) = 0.193`, so it happens one
> time in five by chance. **~25 post-gate points are needed to detect even a halving.** The honest
> line is *"the mechanism works, the rate is unmeasured"*.
>
> ⚠ **AND THE LIVE-CORPUS NUMBER WENT UP, 13.5% → 18.9%, WHICH IS NOT THE GATE FAILING.** The gate
> repairs **at the mint**, so a site only cleans when it is next re-analysed, and the mint ran
> unchecked for the two days the Go awaited a roll. **Judge it on the post-17:18Z mint, never on the
> corpus.**
>
> ⚠ **IF THE GATE GOES QUIET, CHECK `ai_service` FIRST.** `offer-analyser` has **no ROOT `ai_service`
> block**, so without migration `682`'s step-level block the gate is live, firing, and repairing
> **nothing** — it returns *"no ai_service configuration resolvable"*, keeps every point, and records
> why. **That happened for real in the two-minute window between 681 and 682**, and the council's
> `llm_reliability` seat had predicted it verbatim.
>
> **The standing instruction below — DO NOT BUILD A SECOND GATE — is unchanged and now doubly true.**

### ⚠⚠ H1d (superseded above, kept for the reasoning). THE PRODUCER GATE IS BUILT, LIVE IN THE BINARY, AND WIRED TO NOTHING — 25% still born dirty

**`[MEASURED + VERIFIED 2026-09-02]` READ THIS BEFORE BUILDING ANYTHING FOR DECISION E. It is
already built, by THIS LANE, and it has changed nothing because no step calls it.**

| fact | evidence |
|---|---|
| the action exists | `repair_ordering_register_action.go` (630 lines) + `_test.go` (357), `datahelpers/registerwords.go` (193) + parity test, one registry line — commit **`f7156fb54`**, 2026-08-31, `Council-Submitted: 4054f4d9` |
| it is IN THE RUNNING BINARY | `git merge-base --is-ancestor f7156fb54 ebf27c60…` → **yes**, against the deployed `agent-chassis` build, **with a control that correctly says HEAD is not an ancestor** |
| **it is wired to NOTHING** | `offer-analyser`'s workflow has **10 steps** and none has `action: repair_ordering_register`. The config's only `register` mention is unrelated prose in a finding-type list |
| **so it has changed nothing** | since the commit: **254** points minted, **65 born dirty — 25%**. Baseline 23%; 09-01 was 26%, 09-02 18%. **Still bleeding at the old rate with the fix in the running binary** |

> **This is the estate's *"a silent mechanism is UNDRIVEN, not missing"* shape, and this time it is
> OUR OWN WORK.** Four owner asks in three days turned out to be existing machinery nobody drives
> (imagery vocabulary, carousels, the benefit set) — and the fourth is the fix for the third.

**⚠ WHAT IT DOES IS A THIRD OPTION NOBODY PUT TO THE OWNER: it REPAIRS, and records.** It rewrites the
offending construction before the row persists and writes a durable record (`record_key`, default
`register_repairs`) — **including an empty array on clean runs**, so an absent record cannot be read
as "clean". Its header says three modes were costed and FAIL was rejected because it would leave the
site's previous ordering current. **On 2026-09-02 I offered the owner record / refuse / drop-the-bad-points
and he chose record-first — a menu missing the built option, because I did not know it existed.**
That is a defect in my question, not a change of his mind; his 08-31 ruling per the commit message
was **repair**.

**THE OPEN DECISION (with the owner as of 2026-09-02):** wire it **as built** (repair + record, one
config change to one workflow step, already council-submitted), or add a **record-only mode** first
(a code change — the action has no such mode) and measure for a week. **My recommendation on the
record: wire as built** — it cannot stall a site (it repairs rather than refuses), it records either
way so the measurement is not lost, and record-only means another week at 25%.

⚠ **DO NOT BUILD A SECOND GATE.** I started one (`site_spec_register_gate.go` + a hook in
`write_site_spec`) before finding this, on the strength of a `grep | head -8` I read as an absence —
55 hits, not 8. Deleted, reverted, build verified clean, nothing committed. Full account:
`WRONG_CALLS.md` 2026-09-02.

### ⚠ H2c. `reference_values` IS NO LONGER A PIN — the vonc/boxingonline framing in H2 is superseded

**Owner ruling 2026-09-02**, relayed by the `theme kits` lane: *"the classifier can be given the
choice… by default it can start with a theme and change it if it wishes, but it must have full
authority to ignore our set of themes if it chooses."*

**So palette CHURN is authorised behaviour, not a defect to arrest.** `RFC_059` proposed making
`reference_values` a structural pin the render overlay could not touch and **the owner WITHDREW it on
exactly this ground** — a pin removes the authority the ruling grants. ⚠ **Any plan of mine to
de-garish vonc by pinning `reference_values` is now against the ruling.** The lever is what the
overlay STARTS from and how it is guided, never freezing its output.

⚠ **AND MEASURE AT THE SERVED STYLESHEET, NOT THE SPEC** — the overlay may be serving a third thing
entirely: `loanzy.uk` serves `#2A9D8F`/`#F8FAF9`, matching **neither** its palette row **nor** its
`reference_values`. My own vonc observation (`--color-primary` docs say `#6d28d9`, live is `#7c3cff`)
is the same phenomenon.

**Boundary agreed with `theme kits`:** they own `theme_kits`/`page_archetypes`, `apply_theme_kit`,
`theme_kit_defaults`, `page_archetypes_resolver` and one rung of `resolve_composition_layout`. They
have **not** touched `resolve_composition_pallette_action.go` and would rather it stayed that way;
`bugs_open/396` (a design run rewriting the theme row byte-for-byte) is **not** theirs and stands
untouched. **No collision with either of my items** — their only palette migration (`691`, unapplied)
repoints `cv1.co.uk`, `finetuning.uk`, `gaswholesalers.com` appearance-neutrally and excludes vonc,
boxingonline and loanzy. **The self-contradicting-`colour_mood` census is MINE to run**, and their
framing: the interesting half is *"does any site serve a genuinely dark-dominant page"*, since
migration 613's header says the classifier lands dark ~a third of the time by coin-flip. **Offer
standing: if vonc wants a NAMED look, ask them for a kit rather than hand-rolling a parallel
mechanism.**

### H2. Carousels / "nicer components" (via `loanzy_uk_example_site`, from the farmerinsurance.uk review)

Owner, routed to three threads including this one: *"using carousels rather than just lists and
lists of separate cards… maybe we should make different types of carousels as the default, because
scrolling down on a mobile with card after card is not a good user experience."*

⚠ **THIS IS `IMG-074` AGAIN, IN THE SAME FUNCTION, AND THE COMPONENT-MAKER LANE IS ABOUT TO BUILD
INTO IT.** `[MEASURED 2026-08-31]` `hero-card-carousel` has **ZERO** live instances,
`swipeable-insight-carousel` has **1**, and `info-card-grid` — the plain grid — has **42 across 21
sites**. Both carousels are **active and section-level today**, and **41 of 155** active section
components already carry carousel/scroll-snap/swipe markup. **The estate is not short of carousels;
it is short of carousels being CHOSEN.**

Cause: `component_expresses` has **no token for traversal**. For every capability string it emits, a
horizontal component and a non-horizontal one are indistinguishable — `items, list` 3 vs 4; `items`
1 vs 8; `image` 1 vs 5. Same shape as 208-vs-6 before `644`; here it is **42 vs 1**.

**Written up as a CONTRIB in `staged_component_build/` (`97fcf0e22`) BEFORE they build**, because
more variants without a word join the 41 as unchosen library weight. ⚠ **Three things deliberately
NOT concluded:** whether the planner *would* choose a carousel if it could see one (gap measured,
counterfactual not); whether carousels are the right default at all (they trade scroll length for
discoverability — his call); and **what honestly marks a component as horizontally traversable**.
~~`644`'s precision came from the schema's declared `source`… **That measurement is owed BEFORE anyone
writes the arm.**~~ **RESOLVED SAME SESSION — see H2b. Do not re-take it.**

#### H2b. The signal question is ANSWERED, and it made the ask smaller a second time

**The sound signal is declared `semantic_tags`.** `[MEASURED 2026-08-31]` tag ~
`carousel|swipe|slider` → **3 components, 3 of 3 genuine**, and **every tagged component also carries
the markup** (tagged-but-not-marked-up = 0), so the tag never claims what the template cannot
deliver — the same consistency property that made `644`'s derivation safe.

⚠ **A TEMPLATE GREP WOULD HAVE BEEN ACTIVELY SELF-DEFEATING, not merely noisy.** It adds 9 active
section components: **7 match on `overflow-x` and are wide TABLES and calculators**
(`comparison-table`, `evidence-timeseries`, `header-docs`, `platform-comparison`, two calculators,
`Ported Page`) where `overflow-x` is a scrollbar, **and the other 2 are `info-card-grid` and
`case-studies-grid` — the GRIDS.** It would have told the planner that the plain card grid the
carousels lose to **is itself a carousel**.

⚠⚠ **AND THAT IS HOW THE REAL FINDING SURFACED: THE DOMINANT GRID ALREADY HAS A CAROUSEL MODE, AND
IT IS OFF.** `info-card-grid` carries a **declared schema field** — `carousel`, `boolean`,
`source: static`, guidance *"Optional. Set true to lay the cards out as a single-row horizontal
carousel with prev/next"* — gating an opt-in stylesheet in its own template. `case-studies-grid` has
the same from migration **559**.

| component | live instances | sites | **flag ON** |
|---|---|---|---|
| `info-card-grid` | **42** | 21 | **1** (leopardessconsulting `/services.html`, 2026-08-25) |
| `case-studies-grid` | 4 | 3 | **0** |

**So the owner's ask is closest to a switch that already exists and is off on 41 of 42.** Not a
component gap, and for that component not primarily a vocabulary gap. The estate's *silent mechanism
is UNDRIVEN, not missing* shape. **Build order inverted: the flag is the cheapest lever, the
vocabulary token (from `semantic_tags`, never the template) is second, and variants are last if at
all.** ⚠ `carousel` is `source: static`, which the resolver returns `nil, true` for — **nothing
derives it**, so it must be positively set per instance. *Who sets it, and on what evidence*, is a
real decision and not a tidy-up. Full working: the CONTRIB in `staged_component_build/` §6
(`fa549fd76`); the receiving lane has been told to read it before their own asks.

**Not touched:** the deferred brief-fidelity verdicts on farmerinsurance (releasing another lane's
held verdicts off a relay is not mine to do), and §2's logo-vs-identity finding — a real
designer-family gap, genuinely unowned, but a different surface and not folded in here.

### H3. The logo wordmark chain (`bugs_open/417`, owned by `loanzy_uk_example_site`) — CONTRIBUTED, not claimed

Their file; I verified both halves first-hand and **contributed the blast radius into it**
(`fe8819d5e`) rather than filing alongside. `[MEASURED 2026-08-31]` of **27** current-plan logo
prompts, **19 mention `wordmark`** and **10 carry the planner exemplar's phrase `no text outside the
wordmark` VERBATIM** — transcription, not inspiration, and the strongest live evidence of the
quoted-exemplar hazard this lane has seen.

⚠ **THE CENSUS TRAP, because the obvious verification of their fix returns a FALSE GREEN.** The
exemplar phrase itself matches `no text`, so a census asking *"does this prompt forbid text?"* scores
**21 of 27 safe** while **10 of those 21 are the contradictory ones**. **Count the wordmark LICENCE,
not the prohibition.**

⚠ **This is a REGRESSION against a documented rule, not a novel discovery** —
`discovery_checks/default_brand_prompt.go:234` already builds *"no lettering or words"*, and its own
comment at :231 says the rule is not decoration because generated wordmarks produce malformed text at
favicon size. Two producers, opposite rules, neither aware of the other.

**Candidate 1 repriced: necessary, NOT sufficient** — it stops the next 27 and repairs none of the 19,
and a fixed exemplar above 19 live licensing prompts **reads as solved**.
**Candidate 2 (pixels-vs-identity) explicitly NOT claimed by this lane**: a new capability rather
than a fix, architecture-scope by the 2026-07-29 test, and it would be a guarantee conditional on a
vision classifier that inherits its gaps — a stylised or occluded mark it misses returns a clean
pass, and a clean pass from a blind check outlives the blindness. **Closed both ways; nothing owed.**
⚠ Their note for whoever picks up the farmer instance: the **favicon and og-card carry the invented
brand too**, and presence-based discovery will not refile them.

### H4. The visitor's question hierarchy (owner follow-on, 2026-08-31 evening) — **DECISION D**, awaiting him

⚠ **THE MOST IMPORTANT THING IN THIS HANDOFF FOR THIS LANE'S OWN ARTEFACT, so read it before
touching `offer_ordering`.**

**`offer_ordering.lead_with[]` RANKS ON DIFFERENTIATION, and that is a SELLER'S axis.**
`[MEASURED 2026-08-31]` share of points marked `differentiated`, by rank, n=186 across 32 sites:

| rank | 1 | 2 | 3 | 4 | 5 | 6 |
|---|---|---|---|---|---|---|
| **% differentiated** | **100** | **100** | 97 | 61 | 31 | 30 |

Monotonic. **The artefact answers *"what can we say that competitors cannot?"*** The owner's critique
of a Fable hero — that *"No vendor pays us"* is true, strongly stated, and **probably far down the
visitor's list** — is therefore **structural, not a specimen defect**: that point ranked high
*because* it is highly differentiated. **The artefact worked exactly as built, on the wrong axis for
a hero.** This finding predicts the failure rather than describing it, which is why it landed.

> **⚠ AND THE HYPOTHESIS I TESTED CAME OUT AGAINST ME — do not re-derive it hoping otherwise.**
> I expected to show independence claims OUTRANK effort claims, which would have made this a cheap
> **re-ranking**. They do not: effort **19** points at mean rank **2.84**, independence **10** at
> **3.00**, both better than the 3.51 average. On those n, with regex proxies, **that is noise —
> there is no measured inversion.**
> **The gap is ABSENCE: only 19 of 186 points (10%) address effort or practicality at all.**
> Re-ranking cannot surface material never derived.

⚠⚠ **SO THIS ASK IS NOT THE "UNDRIVEN MACHINERY" SHAPE AND MUST NOT INHERIT ITS OPTIMISM.** H1, H2
and the imagery work were all *"it exists, nothing drives it"*. **H4 is genuine new derivation.**
Four-for-four would have been a nice pattern and it is not true.

**What was proposed and endorsed by the copy lane as designed (packaged for the owner as
`DECISION D`):** a per-site `question_hierarchy` aspect in `offer_ordering`'s shape — ranked doubts,
each with a `why` citing its source field — **plus `answered_by` pointing at the `lead_with` point
that addresses it, or explicit `unanswered: true`.** ⚠ **The JOIN is the deliverable, not the list**;
a hierarchy with no link to the copy would be **the third provenance-stamped artefact nobody reads**,
and this lane's own `offer_ordering` (32 sites, **zero** writer consumers) is the argument. **Accepted
acceptance criterion: the first pass comes back MOSTLY `unanswered` at the top — correct result, not
a failure.**

⚠ **THE AXIS COLLISION, flagged before any loop runs.** Ruling 13 makes DENSITY a fault (*"models
compress; we must expand"*) — and **the differentiation axis REWARDS compression**: a maximally
distinctive claim is short, absolute and unqualified, which is the shape *"No vendor pays us"* takes.
The two scores pull opposite ways. The copy lane has put its derived reading to the owner for
confirmation rather than assuming it: **rulings 13+14 read as the buyer's hierarchy and readability
GOVERNING hero copy, with differentiation DEMOTED to an input** — it helps choose among answers to
the same doubt, is never the ranking key, and is never a licence to compress. **If he confirms, this
lane's corpus is not wrong — it becomes the SECOND sort key.** If he rules otherwise the loop design
changes. **Either way, decided before two seats fight about it mid-loop.**

**Boundary, agreed identically both sides:** the hierarchy is **unserved rationale**, same side as the
`why` fields, structured input only. ⚠ **Never rendered into a prompt as prose** — *"most visitors
first ask X"* in a writer's or critic's context window **IS** the presumption shape and will be
copied verbatim.

**The agreed sequence across both lanes, once the owner rules on C and D:**
**wash the 51 dirty points → wire ONE clean site (9 are already clean) → derive the hierarchy →
join it via `answered_by` → and only then does the per-copy loop have all three axes** (register,
relevance-ordering, density) **to critique against.** Nothing builds until his word.

## §F — WHO OWNS WHAT NEARBY (changed since yesterday)

**`bugs_open/381` is CLOSED and its lane wrapped up 2026-08-25** — so `component_expresses` has **no
live owner**; the register relation in IMG-074 is now the durable channel, not a session.
**`apis.uk`** holds the only live instances and is actively engaged (they verified at both ends).
**`loanzy_uk_example_site`** — corrected their own ADDENDUM 1 revalidation claim today; I checked
`WII-033` at their request and it does **not** repeat it, so no correction was owed on this side.
Their fix (`recordModeSilenceRule`, council `04a3ce1f`) is Go, inert until a roll — **until then the
operative truth is that verdict rows are cleared by humans only.**
**`portfolio_positioning`** — their uncommitted SEO-007 LANDMINES edit rode into `b3bddba60` as a
named same-file passenger; they were told, nothing lost.

### ✅ H1d-bis (the sibling session's entry; see the corrected H1d above). THE PRODUCER GATE IS BUILT (2026-08-31 evening) — and it is NOT yet evidence of anything

**Shipped, four commits.** `f7156fb54` Go (inert until the roll) · `06b10a1d8` wiring `681_HOLD`
(+`_ROLLBACK`) · `40ab44bdf` LANDMINE · `af312bc1c` register **CQ-034**.
Council: **`4054f4d9-cd75-4b9c-8b8c-b7b86f11de1e`** — ~~⚠ **VERDICT NOT READ. That is the first thing
you owe.** The code is on the shared branch; a REVISE must be acted on.~~
> **RESOLVED 2026-09-02 — and this line is why it took two days.** Round 1 returned **REVISE** at
> **2026-08-31 17:26:15Z**, ~20 minutes after submission and while I was still working; I polled the
> ORCHESTRATION five times, saw it progressing, and reported "still running". **Progress observed at
> T is not absence of a result at T+n, and `orchestration_states` is the wrong table** — the verdict
> is a `diagnosis_artifacts` row and the run's `current_step` says nothing about whether the report
> exists. The obligation was recorded correctly here and *that made the system look healthy*: **a
> correctly-recorded obligation is not a discharged one.** (`WRONG_CALLS.md`, 2026-09-02.)
> **Round 2 is APPROVED** (2026-09-02 13:19Z, 16 reviewers, 1 abstained, no high-severity
> objections). Every advisory acted on or recorded as open — see NOTES §"Round 2: APPROVED".
> ⚠ **Read the REPORT, not the decision field:** this approval carried nine objections, two of them
> real (an unqueried exclusivity claim, and a merge constraint I was relying on by luck).

**Owner ruled REPAIR** over fail and drop, after the three modes were costed to him. FAIL leaves the
site's previous (also dirty) ordering current, so nothing gets cleaner and one tic costs a whole
re-analysis; DROP throws away ~1 benefit in 4, some ranked first — `verify_cited_cardinals` drops a
FALSE CLAIM, this would drop a real benefit wearing a tic.

**What justified building rather than filing.** `[MEASURED 2026-08-31]` post-667 mint (rows created
after it committed at 10:34:30Z): **18 of 75 points dirty = 24.0%**, against an all-history 23.3%.
**Unmoved.** 15 of the 25 currently-live dirty points were minted AFTER the wash, across 8 sites,
including **rank 1** on `finetuning.uk` (14:54) and `mortgagecalculator.co.uk` (16:25). ⚠ The
disconfirming result was available — a dropped rate would have meant no gate was owed — and did not
happen. **⚠ Postgres word boundary is `\m`/`\M`; `\b` is a BACKSPACE there and has already produced
two confident zeroes in this lane.**

#### ⚠⚠ THE FINDING TO CARRY: A THRESHOLD DERIVED FROM A MEAN CANNOT CATCH ITS OWN MOTIVATING CASE

I implemented §H1b's ACK rule as a 60% length floor on `differentiated: true` points, wrote it into
the action's header as the rule that catches this class, **and then the motivating case PASSED it.**
667's actual repair of `leopardessconsulting` rank 2 — *"…in days, not months."* → *"…in days."* —
retains **64 of 76 bytes (84%)**. The differentiation was lost in **TWELVE BYTES**; the case the
whole rule was built from sits at −16%, nowhere near the −28.7% mean it was derived from.

**Caught by using the REAL repair from the REAL site as the fixture.** A composed one — a repair I
invented, which would naturally have cut half the sentence because I already knew which arm I wanted
to fire — passes, and the guard ships blind to precisely the failure it exists for.

**The replacement is not a better number.** On a differentiated point the violating construction IS
the comparison and the distinction is what it compares against, so truncate-before-the-comparison
removes it EVERY time at any length: a candidate that is merely a **PREFIX** of the original is
rejected. The length floor is KEPT alongside, for the gross rewrites the prefix rule cannot see.
Undifferentiated points stay exempt — there truncation is sanctioned and lossless.
**Relayed to `copy_quality_two_stage`: their wash/expansion generators have the same exposure, and
the prefix test is one line they can apply today, ahead of any roll.**

#### ⚠ The judge is blind to the word arm without layer 2

`AcceptNegationRewrite` re-scans with `ScanDefineByNegation` — **shapes only**. `plainly`/`honest*`
had **no Go reader anywhere**. So a rewrite dropping "X, not Y" and reaching for "we say so plainly"
passes every check the shared judge makes; layer 2 (full-register re-scan) is the only thing stopping
the two arms displacing into each other.

#### What is owed next, in order

1. **Read the verdict** (query in §A's form, correlation above). Act on REVISE/REJECTED.
2. **The migration is NOT in the reviewed plan** — written after submission. It is HELD, so there is
   time: its own round via `RESUBMIT_CORR=4054f4d9` once this verdict lands.
3. **Apply `681_HOLD` ONLY after a roll carries `f7156fb54`.** It repoints
   `write_offer_ordering.spec_data`, so applying it early does not add an inert step — it breaks the
   ordering write. Per-service provenance check is in the file's header.
4. **⚠ DO NOT REPORT THIS AS WORKING.** Nothing has run through it. **23% is the baseline it has to
   move**, and that is measurable only after the wiring applies — from `register_repairs` in the
   persisted spec (the clean path writes an empty array, so `len(...) > 0` is a sound predicate).
   Layer 4 is strict by design and WILL refuse repairs; the success rate on differentiated points is
   an OPEN EMPIRICAL QUESTION, to be read off the record, never assumed.
5. **The nightly `brief-negation-check` cannot see this corpus** — `[VERIFIED 2026-08-31]` its census
   covers only aspects a live agent prompt names as `{{.site_specs.specs.<path>}}`, and
   `offer_ordering` is not among the ~25 today. It switches itself ON the day Decision C wires the
   corpus into a writer prompt — expect a nightly report you did not ask for, over a corpus 13.5%
   dirty. Not a bug; worth knowing before it surprises someone.
6. **H4 / DECISION D is still with the owner** and is untouched by tonight. So are H2b's carousel
   flag and everything in §C and §E.

#### ⚠ The trap found while wiring — now a LANDMINE

Seed `408_offer_analyser_agent.sql` says `write_offer_ordering.spec_data = 'offer_analysis.result.ordering'`;
the **live row says `'ordering_checked.object'`**, repointed when the cardinal gate went in. A
migration written from the seed would have set it back to the raw LLM output, **silently un-wiring
`verify_cited_cardinals`** (18 real drops recorded) while inserting a new gate and reporting success
— both gates dead, artefact normal, every guard passing. Caught only because the drop-record count
could not be reconciled with a chain in which the cardinal gate's output reached nothing.

### ✅ H1e. THE GATE IS LIVE AND REPAIRING (2026-09-02) — and the next work is DECISION D, not more gate

**State, all verified at the artefact:**
- `681` + `682` **APPLIED** (17:15Z / 17:17Z), both in `schema_migrations`. Chain at the live row:
  `verify_ordering_cardinals` → `repair_ordering_register` → `write_offer_ordering`.
- Council **APPROVED round 2** (`4054f4d9`); every advisory dispositioned in NOTES.
- **First post-`682` run repaired 2 of 2** on `garden-tools.uk`, surgically (word deleted, grammar
  fixed, nothing else changed). ⚠ **n=2 — a working MECHANISM, not a rate. Do not quote 100%.**
  23–24% is the baseline and moving it needs days of mints.

**Two things a next session must not re-derive:**
1. ⚠ **`682` exists because `offer-analyser` has NO root `ai_service` block.** The first live run
   (17:15:24Z, in the two-minute window) returned `"no ai_service configuration resolvable"`, kept
   every point and recorded why — the safe-failure path, proven in production. If the gate ever goes
   quiet, **check that block first**: without it the gate is live, firing, and repairing nothing.
2. ⚠ **The em-dash form of `x_not_y` is invisible to BOTH scanners** (register pattern and
   `negXNotYRe` are comma-anchored). A point the gate marks `repaired` can still carry it — see
   NOTES. Relayed to `copy_quality_two_stage` as a v2 candidate; **do not widen `negXNotYRe` from
   this lane**, it is shared with the page gate, the voice gate and the nightly CLI.

**WHAT IS OWED NEXT, in order** (replaces §H1d's list, which is discharged except item 4):

1. **DECISION D — `question_hierarchy` + `answered_by`.** ⚠ **RELAYED ruling, second-hand** (via
   `copy_quality_two_stage`, owner's words *"yes to both, we can see if it works"*). Seam split
   proposed to them and recorded in NOTES: the **analysis half is this lane's**, the writer/ordering
   consumption half is theirs. **The JOIN is the deliverable, not the list.** Boundary: unserved
   rationale, structured input only, never rendered into a prompt as prose. Accepted criterion: the
   first pass comes back **mostly `unanswered`** — correct, not a failure.
2. **THE AXIS INVERSION — but AFTER D, not before.** ⚠ Buyer-relevance + readability govern heroes;
   differentiation demotes to an INPUT. `offer-analyser`'s ranking guidance still ranks on the
   seller's axis and `[MEASURED 2026-09-02, n=190]` it is fully intact (100/100/85/65/32/53 by rank).
   **Re-ranking FIRST achieves nothing** — H4 measured the gap is ABSENCE, not order, so a prompt
   migration would just reorder the same seller-axis points. Derive, then re-rank.
   ⚠ **Do NOT quote any improvement in the effort/practicality gap.** Today's 30% used a WIDER regex
   than H4's 10%; a different instrument is not a different result. Re-run H4's exact proxy first.
3. **The gate's real rate**, once days of mints have passed: read `register_repairs_summary` — the
   `still_violating` field is the honest one, and the clean path writes zeros so
   `len(register_repairs) > 0` is a sound predicate.
4. **Everything in §C and §E is still untouched** — supply (§C3, the larger half of the imagery ask,
   still unowned), first-wins `section/illustration` resolution, the 8 components exposed to the
   alias trap, H2b's `info-card-grid` carousel flag (ON on 1 of 42).

**Not this lane's, for awareness:** the owner ruled **option (a)** on the exemption-as-licence
question — clean the briefs fleet-wide, gate semantics untouched. This lane's 32-of-34 measurement
earned it; the campaign is `copy_quality_two_stage`'s. **The third-path RFC trigger stands
unchanged** — none of this creates a new analyser→writer path.
