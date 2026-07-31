# HANDOFF — gauntlet_dead_cta / vonc6 thread, 2026-07-31

**Written because this session's context ran long, not because the work stalled.**
Supersedes `HANDOFF_2026-07-29_continue_here.md` for anything below; that file is
still the reference for the rail (§5), the island landmines (§4) and the H ruling
(§7). Handoffs A–D of 2026-07-30 are separate startable units and are NOT
superseded — B is already TAKEN by another thread.

## 0. State in one paragraph

The visitor-identity fix is **live and proven in production**. The
content-duplication checker and its deterministic handler are **built, seeded,
fixed after a measured false positive, council round 2 submitted, and deliberately
not enabled**. The chassis roll they were waiting for has landed and is verified.
Nothing is half-applied. The single next action is reading the round-2 verdict
(§4.1); enabling the check remains a separate decision that needs one other lane's
agreement first.

> **Updated ~11:50Z.** Read §3's CORRECTED block before acting on anything about
> this checker: the rule as originally shipped would have deleted a live section,
> and both the in-remit definition and my "verified no-op" claim have changed.

## 1. DONE and verified this session

| what | state | proof |
|---|---|---|
| tools-api client identity (`bugs_open/139`) | **LIVE + PROVEN + council APPROVED** | island `v1.0.1207` swapped 08:37Z; see §2 |
| `check_content_duplication` + `remove_duplicate_page_sections` + `deduplicate-sections` | built, seeded, **INERT** | commit `feat(151 cand 3)`; chassis `v1.0.1214` pod-verified, §3 |
| Shared normaliser `datahelpers.NormaliseSectionText` | live | replaced two copies + a promised drift test |
| og:image on the gauntlet page | **FIXED** (was listed as a 404 to fix) | HTTP 200, `image/png`, 1200×630 exactly |
| 4 handoffs A–D + `dedup_census.py` + `provocation_visibility.py` | committed | `f599ed87c` |
| `SUMMARY_2026-07-31` | written | the authority dilemma as a worked example |
| Register: `CQ-018` added, `PUB-002` corrected | committed | httpguard's "called by NOTHING" is now false |

## 2. The identity fix — what makes it PROVEN rather than deployed

Council `e053fac4-eeaf-431e-aa88-817c4107476e` → round 1 REVISE (prior-art
evidence objection), round 2 **APPROVED** with **no code change**. The lesson is
reusable: the objection asked what a landmine note said, so the answer was the
note's full text plus a source line, not a rewrite.

**A real `/round` through Cloudflare stored the hash predicted BEFORE the request
was sent** — `9e464fe9fca925b099a25141f40afad5`, being
`sha256(<this machine's egress address per cdn-cgi/trace>)[:16]`. So the column did
not merely change; it changed to the expected value. `count(DISTINCT
client_ip_hash)` on the island's `gauntlet_rounds` went **1 → 2**.

That settles what `139` marked `[INFERRED]`: **`CF-Connecting-IP` reaches the app
process.** The acceptance check passed from ONE machine legitimately, because the
before-state was itself a constant — the 95 prior rows are the second network.

Still open, deliberately: `139` is left in `/bugs_open/` because `who-owns.py`
names another lane and **the close is theirs to call**; their three probe rows are
left in place because they are now the before-state of the measurement.

**Landmine earned:** when comparing against a stored digest, **read the hashing
function's width first.** I queried the full `sha256('172.18.0.1')`, got 0 rows,
and briefly read that as "the constant is not the docker gateway". `hashIP`
truncates to `sum[:16]` (`round.go:33-37`). Their figure was right; my check was
the wrong shape.

## 3. The duplication checker — everything except the last step

**Design, which is an owner ruling of 2026-07-31 and must not be quietly widened:**

```
IN-REMIT  same page, SAME SLOT,        -> content_duplication item
          byte-identical content_data
                                        -> deduplicate-sections
                                        -> remove_duplicate_page_sections
                                           (delete later rows, keep earliest,
                                            renumber, queue a rerender)
RESIDUE   near-duplicate / cross-page   -> ONE capability_gap, NO handler,
                                           do_not_auto_rewrite: true
```

Rationale is in `SUMMARY_2026-07-31`; the short version is that a similarity score
says nothing about meaning, and `bugs_open/149` C1 is the witnessed cost of letting
an LLM rewrite copy unguarded.

#### Why the line sits exactly there — read this before "improving" the check

The remit boundary is not a limitation of the implementation. It is the whole design,
and it is drawn at **the only place where a repair is provably safe**: two rows on one
page, **in the same slot, whose `content_data` is byte-identical** (narrowed
2026-07-31 in `43492ec94` — it used to say "whose normalised text is identical", and
that version would have deleted a live section; see the CORRECTED block below).
There, "which one do I keep?" has a
correct answer that needs no judgement — keep the earliest, because the duplicate adds
nothing a reader could want. Every widening below removes that property, so each one
is refused **on the same ground**, not on three different grounds:

| the tempting widening | why it is refused |
|---|---|
| "0.92 similar is obviously the same — dedupe it too" | Now something must decide which of two *different* texts survives, and that is an editorial judgement. `bugs_open/149` C1 is what that costs when an LLM makes it unguarded. A threshold does not make the judgement safe, it hides who is making it. |
| "the same fact on two pages is duplication — dedupe across pages" | Two pages restating one fact is usually **correct** — a landing page and a case study *should* both cite the headline number. Cross-page identity is a content-strategy question, and it has no page-local answer at all. |
| "a unique index on `(page_id, slot_name)` would prevent all of this" | Measured: it **breaks 11 legitimate pages**. 10 of the current 11 groups are real repeated slots with genuinely differing content (§3 re-measurement). The schema cannot express the rule because the rule is about *content*, not *shape*. |
| "the residue item has no handler — wire one up" | The missing handler **is the finding.** `do_not_auto_rewrite: true` is a deliberate statement that this class needs a human, recorded as one `capability_gap` so it is visible rather than silently unhandled. Giving it a handler converts an honest gap into an unsupervised rewriter. |

**The test for any future change: does it preserve "the repair needs no judgement"?**
If a human would have to decide which text survives, it is residue, and residue gets a
`capability_gap` — never a handler. Widening this check is an **owner decision**
(2026-07-31), not a refinement a thread makes while passing through.

**Roll verified.** Chassis `v1.0.1214`, both replicas:
`remove_duplicate_page_sections` **4**, `content_duplication` **6**, nonsense
control **0**. So the action is in the running binary and the chain is complete.

**Council round 1 = REVISE (`da3f2d9b-ae6f-492d-ad3b-748323b66367`), and it caught
two REAL defects** — both now fixed in `269_*.sql` **and applied to the live row**
by UPDATE (the seed is `ON CONFLICT DO NOTHING`, so re-running it would not have):

1. `page_id` was in the item spec only → the queued `page_rerender` would have a
   **NULL `page_id` column**. `create_work_item` takes it as an optional input that
   populates the column (`create_work_item_action.go:247-249`).
2. `item_key_prefix` alone yields `<prefix>_<domain>` — **site-wide**. Two pages
   deduplicated close together would collide on `idx_swi_dedup` and **one rerender
   would be silently lost**. Now sets `item_key_suffix_field: input_data.page_id`,
   which is a hard error when unresolved rather than a fallback to the colliding key.

~~**Round 2 has NOT been submitted.**~~ **SUBMITTED 11:40Z** on the same
correlation, carrying the fix, the corrected figures and measured answers to two
further HIGH objections. Verdict not yet read — §4.1.

### The reviewer also re-measured my population claim, and it is now stale

`bug_historian` re-ran the fleet query and got **`total_groups: 11,
content_identical: 0`**. My submission quoted 17 groups / 11 legitimate / 6
content-identical-all-on-vonc-about. That was true when measured and is no longer:
**another session hand-fixed vonc's about page** (`d7f69e1c0`, 6 DELETEs +
renumber), which was the entire content-identical population.

⇒ **The in-remit half currently has nothing to detect fleet-wide.** It is a guard
against recurrence, not a backlog-clearer. Say that plainly in round 2 rather than
requoting the old figure — and re-measure before quoting anything, because that is
the third stale-figure correction in this lane this week.

### Re-measured 2026-07-31 ~10:20Z — confirmed, and it now has a stronger proof

Independently re-derived from the schema rather than by re-running the reviewer's
SQL (agreement between two runs of the same query proves nothing about the query).
**`total_groups 11 / content_identical 0 / legitimately_repeated 10 /
has_null_content 1`.** The query is now in `RUNBOOK` §16 — it was missing, which is
how the figure went stale unnoticed in the first place.

**The 11th group is the interesting one, and it is the landmine shape:**
`finetuning.uk /our-position-on-ai.html`, slot `generic-text-block`, 2 rows, both
with `content_data IS NULL` ⇒ `count(DISTINCT md5(...)) = 0`. A naive identity test
reads 0 distinct values as "all the same" and this is **the one group in the fleet
where a false positive would delete a live row.**

**Both halves already exclude it, for a reason independent of NULL-handling:**
`COALESCE(content_data::text,'{}')` → `NormaliseSectionText` → empty string → caught
by the `len(s.Text) < 80` floor, applied identically in
`check_content_duplication.go:238` (detect) and
`remove_duplicate_page_sections_action.go:144` (repair). Same threshold, same shared
normaliser, both halves.

⇒ **So, precisely: enabling the check today would file ZERO items and delete ZERO
rows.** That is the sharpest available statement of "built but inert" — not merely
"not switched on", but *verified to be a no-op against the current fleet*. Two
consequences, both load-bearing:
- It **strengthens** the case for round 2: the design's safety is no longer an
  argument, it is a measurement over the real population including its one
  adversarial case.
- It **weakens** any argument for enabling it in a hurry. There is no backlog to
  clear, so the only thing the switch buys today is recurrence protection — which is
  worth having, and is worth having *after* the owning lane answers, not before.
  Nothing degrades while it waits.

⚠ **Do not carry the `11 / 0` figure forward without re-running §16.** It was true
at 10:20Z. It was a different number 24 hours earlier, and one hand-fix by any
session moves it again.

> **CORRECTED 2026-07-31 ~11:30Z — the paragraph above is RIGHT and the conclusion
> I drew from it was WRONG. "Enabling the check today files ZERO items and deletes
> ZERO rows" was FALSE when I wrote it.** The census measured content identity as
> `count(DISTINCT md5(content_data::text))` — byte identity of the whole blob. **The
> shipped check does not use that ruler.** It compares
> `datahelpers.NormaliseSectionText(content_data)`: prose only, with asset-like keys,
> short strings, URL-ish values **and every non-string value** discarded. That
> population is strictly larger than byte identity, so my census could only ever
> undercount — and it did.
>
> Measured with the shipped function compiled against all 1,023 live rows, the rule
> as it then stood found **one in-remit group and would have deleted one live row**:
> `lobby-grid@5` on **vonc.com/index.html** — this lane's own home page.
>
> **What caught it:** reading the round-1 council report itself instead of trusting
> §3's summary of it. §3 said round 1 "caught two REAL defects" (both since fixed).
> The report actually carries **seven objecting seats**, and `bug_historian`'s HIGH
> gating objection — *is the discriminator blind to non-text payload?* — is the
> thread that unravels the whole claim. **I had compressed a review into a sentence
> and then reasoned from the sentence.** Do not do that; the report is in
> `diagnosis_artifacts` under the correlation and it is worth the read.
>
> **Why the false positive existed:** both rows' `content_data` is byte-identical and
> is **not section content** — it is the site-wide context blob (`year`, `email`,
> `domain`, `nav_items`, `tone`, `_built_at`). Two *different* components (different
> `component_id`) populated with the same boilerplate. The normaliser's asset-key
> filter strips `url`/`id`/`class` but not `email`/`year`/`domain` or nav labels, so
> on a boilerplate-only blob the boilerplate **is** the comparison. Determinism does
> not rescue a rule whose input is the wrong field.
>
> **FIXED** in `43492ec94`, two narrowings that both strictly shrink the in-remit
> set: slot equality is now **necessary** (not the slot-keyed rule 156 rejected —
> that rejected slot as *sufficient*), and identity is the **canonical blob**, not
> the normalised prose. Shipped rule re-run over live data after the fix: **0 groups,
> 0 deletions.** So the sentence is true now — *because of the fix*, not because it
> was true when I said it. Logged in `WRONG_CALLS.md` (4th entry in three days in one
> family; it rode into another session's commit `cadd12be9` as a same-file
> passenger, which is expected and loses nothing).
>
> **The transferable rule, and it is narrower than "re-measure":** when the claim is
> about what CODE will do, the ruler must be **the code, executed** — not a
> reimplementation of it in SQL or Python, however careful. A reimplementation is a
> second definition of "identical", which is the exact drift
> `section_text.go`'s header was written to prevent. I rebuilt that drift in SQL in
> order to check it. Method now in **RUNBOOK §16b**.

## 4. NEXT ACTIONS, in order

> **UPDATED 2026-07-31 ~11:45Z. Actions 1 and 5's live-driver gap are DONE; action 2
> is unchanged and now has more support, not less.**
>
> - **Round 2 VERDICT: REVISE** (12:51Z), gated by `prior_art_librarian`. **8 of 12
>   seats approved** — up from round 1 — and almost nothing was about the code. The
>   gating position is a stance, not a defect: *"I do not trust a rationale's
>   self-reported verification, I attach the lookup."* Every objection is now
>   answered with the lookup attached, in `9fec8ce01`. **Round 3 is assembled but NOT
>   fired** — that is a live decision, see §4.0.
>   - It also caught me in a **second incomplete population**: my round-2 answer on
>     section-editor checked `default_config` only, and `agent_definitions` has **five**
>     workflow-bearing columns (`task_workflow`, `orchestrator_workflow`,
>     `orchestration_workflow`, `capabilities`). Re-checked across all five — the claim
>     holds, but it held by luck of the query, which is the same shape as the
>     wrong-ruler error above. **Diff the population before answering an absence
>     objection.**
>   - Inertness is now measured, not asserted: **0** live agents reference
>     `content_duplication` in any column; exactly one names the action; **0**
>     `content_duplication` work items exist.
>   - `bug_historian`'s line is the one to carry: **"currently unexercised is not the
>     same as closed."** The `site_plan_sections` guard is now tracked in `doc_notes`
>     (`action`/`remove_duplicate_page_sections`) and CQ-018 verify-later, with the two
>     precedents he named (`bugs_closed/058`, `bugs_closed/069`).
>   - The measurement is now a **re-runnable artefact** rather than my word:
>     `scripts/dedup_census_shipped.{go,md}`, `-legacy` reproduces the false positive.
> - **The live share-card driver RAN and passed 11/11**, on a real round through the
>   real page: challenge element 451 chars, defence textarea 198 chars, both still
>   populated when the button fires. **CREDIT WHERE DUE — another session drove it
>   first** (`be019bd41`, 10:45Z, a 469-char challenge), so the
>   `[INFERRED from code]` gap was already closed before my run. What my run adds is
>   narrower: theirs printed `SKIP PIL unavailable` for the three IMAGE assertions and
>   still ended `ALL LIVE CHECKS PASSED`; they hardened the script so a missing Pillow
>   now FAILS, and recorded that the hardened script had **not** been re-run
>   end-to-end. Mine was that run, and the image assertions **executed** — real
>   numbers, not skips. Card is
>   1200×630, lower half not blank, no page errors. PNG:
>   `p4_sources/live_card_2026-07-31.png`. Two cosmetic observations, neither a
>   defect: a dead band between answer and verdict when the defence is short, and the
>   engine returned "personalization" (US spelling) on outward-facing copy.
> - **Action 2 (do NOT enable) STANDS, and the reason is now stronger.** Before the
>   fix, enabling would have deleted a live row from vonc's home page. That is what a
>   switch on an unreviewed destructive path costs, and it is the second time this
>   week the "it's inert so it's safe" framing has been too comfortable.

### 4.0 ROUND 3 — FIRED (owner decision, 2026-07-31), killed once by a roll, re-fired

**Owner said fire it.** Submitted 14:49 as `45d143e0` — **killed at 15:00 by another
session's chassis roll** (both pods `startTime` 15:00:20/15:00:42 vs a last
`updated_at` of 14:59:59), 11 minutes in, with 8 of 12 reviews already landed.
**Re-fired UNCHANGED** as `e4f81e61-83f3-4185-83f1-00b0c45dc4d6` after the ~300s
post-restart window. An infra death is not a judgement, so the plan was not altered —
same class the runbook records for `complete_invalid` with no `council_report`.

⚠ **If you inherit this mid-flight, a dead council looks exactly like a slow one.**
`status` stays `EXECUTING_STEP`, `error` is NULL, and `currently_executing` names a
step whose output is **already in `collected_data`**. Compare `updated_at` against pod
`startTime`; rounds 1 and 2 took 11 and 13 minutes end to end, so anything past ~15
without movement deserves the pod check. Landmine filed.

⚠ **And `orchestration_states` has NO `id` column — it is `orchestration_id`.** My
poll asked for `id` with stderr to `/dev/null`, so it returned empty for 31 minutes
and I read that as queue latency, which is exactly what CLAUDE.md primes you to
expect. Never `2>/dev/null` a query whose emptiness you intend to read as evidence.
Landmine + `WRONG_CALLS` filed.

The original decision, kept because the reasoning still applies if there is a round 4:

#### The argument that was weighed — fire round 3, or stop?

Everything the seats asked for is done and committed (`9fec8ce01`). Round 3 would be
a resubmit on the same trail with the lookups attached and no code change beyond the
architecture seat's contract comment.

**Argument for firing it:** 8 seats already approve; the gating objection is
answerable by attachment, which is exactly what has been attached; approval makes
`098` credit the existing `Council-Submitted:` trailers automatically.
**Argument for stopping:** the code defect that mattered is FIXED and committed, and
the standing rule is *one council run per coherent task, not per iteration*. A third
round buys a trailer, not a safer mechanism. The check stays inert either way.

~~**Not fired without a decision.**~~ Fired on the owner's instruction; see above.
Any further round uses `RESUBMIT_CORR=da3f2d9b-ae6f-492d-ad3b-748323b66367`.

### 4.0b VERDICT: **APPROVED** (2026-07-31 15:39Z) — read this before doing anything else

12 approve / 2 advisory object, **none high-severity**, and `prior_art_librarian` — the
seat that gated round 2 — approved. `Council-Reviewed: da3f2d9b-ae6f-492d-ad3b-748323b66367`
is on `<this commit>`; the earlier `Council-Submitted:` commits are credited
automatically by `098` at report time, so **do not amend anything**.

**Three things the seats could not check themselves, now closed:**
- **No `doc_plans` row pre-existed** for `subject_type='action'`, keys
  `remove_duplicate_page_sections` / `check_content_duplication` / `deduplicate-sections`
  (0 rows). So there was no prior decision record to load — mine are the first.
- **The working-tree compile failure is gone and was never mine.** Another session
  committed the WIP that owned it (`11cf7969c`, which added
  `writeDiscoveryCheckErrorLog`). **HEAD now builds**, so the `git archive HEAD`
  caveat in the submission is obsolete.
- **Full suites re-run on the REAL tree**, not a synthetic one: `discovery_checks`,
  `datahelpers` and `actions` all pass. That retires `editquality`'s low objection.

**THE ONE OBJECTION THAT SURVIVES AND IS NOT ADMINISTRATIVE** — `bug_historian`,
medium, and he is asking the right question. Not *whether* the `site_plan_sections`
pre-delete guard is tracked (it is, in `doc_notes` and CQ-018) but **whether it will
actually be built before the check is enabled anywhere, or whether "verify-later" will
sit unexercised the way this exact gap sat unexercised in `bugs_closed/058` and
`bugs_closed/069` until something hit it.** That is a fair challenge to the whole
deferral and it is an OWNER DECISION, recorded in §4.2 below. Do not let it quietly
become "tracked, therefore handled".

**`constitution` objected on TONE, and it is right.** My round-3 rationale was
"saturated with rhetorical grandstanding and shouted emphasis … performative
self-narration of the review process" — all-caps headers, "THE SHARPEST THING SAID".
The platform norm is PLAIN, PRAGMATIC TONE. A submission is not the place to narrate
one's own diligence; state the change and the evidence. Applies to future submissions
from this lane.

### 4.2 OWNER DECISION OWED — the `site_plan_sections` pre-delete guard

`bug_historian` accepts the deferral is tracked and objects that tracked-and-unbuilt is
how 058 and 069 happened. Two coherent answers:

**(a) Build the guard before anyone enables the check.** The action consults
`site_plan_sections` and refuses to delete where the PLAN itself specifies the
repetition. Closes the class rather than tracking it; costs a coupling to a second
table and a rule I have said I would rather measure than guess.
**(b) Make it a hard precondition of enabling, not of the code.** The check is inert
and enabling belongs to the `brochure_component_library` lane; attach the guard
requirement to that decision (CONTRIB + `doc_notes` + this handoff) so it cannot be
enabled without the question being answered.

**Measured position today:** exactly one upstream duplicate exists fleet-wide
(`webdesign.co.uk`/`index`/`info-card-grid`, 2 plan entries), it does NOT overlap any
in-remit item, and the in-remit set is empty — so nothing is at risk right now, and (b)
is sufficient *until someone enables the check*. Note two seats could not verify that
figure at all (`site_plan_sections` is not in their schema access), so it rests on this
lane's measurement — re-run it before relying on it.

### 4.1 Read the verdict

```sql
SELECT created_at, metadata->>'decision'
FROM diagnosis_artifacts
WHERE correlation_id='da3f2d9b-ae6f-492d-ad3b-748323b66367' AND kind='council_report'
ORDER BY created_at;
```
Two `council_report` rows means round 2 landed; the later row is the one to read. If
it is APPROVED, the commit `43492ec94` already carries `Council-Submitted:` with this
correlation, so `098` credits it automatically **with no amend** — do not write a
`Council-Reviewed:` trailer onto a later unrelated commit to "fix" it.

1. ~~**Resubmit the checker, round 2**~~ **DONE 11:40Z** — was: resubmit with `RESUBMIT_CORR=da3f2d9b-ae6f-492d-ad3b-748323b66367`.
   Content: the two seed fixes (done, live), the corrected population figure
   (**re-run §16 first — it was 11/0 at 10:20Z**, guard-not-backlog), the
   NULL-group no-op proof from §3 (the strongest single item in the submission: the
   one fleet group that could trigger a wrongful delete is excluded by both halves at
   the same threshold), and a note that the sketch now shows the enqueue the
   rationale promised. Working file:
   `<scratchpad>/council_dedup.json` is gone with the session — rebuild from the
   commit message of `feat(151 cand 3)`, which carries the full reasoning.
2. **DO NOT enable the check yet.** Enabling means adding `content_duplication` to a
   discovery agent's check list. Two reasons to wait: the owner has not asked for it,
   and **the first site it runs against starts deleting rows**. `bugs_open/151` is
   owned by the `brochure_component_library` lane (113 commits/14d) and
   fundamentallyai is theirs — a CONTRIB is already in their directory
   (`CONTRIB_2026-07-31_151_candidate_3_is_built.md`) offering them the checker
   outright. **Get their answer first.**
3. **`bugs_open/154`** — `tool-improver` dies at `load_tool` on `tool-auditor`-raised
   items (`input_data.component_id` resolves to nil, and counterintuitively it is the
   items WHOSE COLUMN IS SET that fail). Unowned, mechanism marked `[INFERRED]`;
   read the dispatch `input_mapping` before acting. Consequence: `bugs_closed/010`'s
   convergence guard has STILL never reached cycle 1.
4. **`bugs_open/150`** — the improvement loop reports "No issues found — site is
   clean" straight after promoting findings, and skips its own closing rerender.
   Unowned. Anyone firing the sweep to clear a backlog is told the opposite of what
   happened.
5. **Handoffs A / C / D** of 2026-07-30 are startable. **B is TAKEN** (lane
   `docs024_key_docs_latest/provocation_pipeline/`) — do not start a second thread.
   That thread also **corrected me**: the provocations archive page is NOT broken;
   my 1,293-char figure came from a probe looking for *today's* headline, which
   correctly is not on that page.

> **6. ADDED 2026-07-31 ~15:20Z by lane `staged_component_build` — a decision for you,
> and a correction to something recorded fleet-wide about your tool.**
> Full detail, with every measurement: **`CONTRIB_2026-07-31_arena_tool_is_now_acceptance_testable.md`**
> in this directory. In short:
>
> - **Your arena page has never been acceptance-testable**, because the Tier-4 lookup is
>   `name IN (function, 'tool-'||function)` and `pages.name='tool-arena'` matched neither
>   `tool-arena-interface` nor `tool-tool-arena-interface`. Every run hard-errored on the
>   page lookup before evaluating a criterion. **Fixed** — `pages.name` **and**
>   `site_plan_pages.name` moved together (the second matters: `check_sectionless_pages`
>   joins `spp.name = p.name`, so renaming only the page would have silently dropped it out
>   of that detector, where it is currently reported as item `559cb636`). Served page is
>   **byte-identical**, md5 unchanged, and your nav renders `nav_label`, not `name`.
> - **`tool-arena-interface` was recorded fleet-wide as "an orphaned component with no page
>   — decide whether it should exist". That was FALSE**, it came from my check guessing at
>   names, and it is corrected in my lane and in `WRONG_CALLS.md`. Your tool is live and
>   serving. Flagging it because a deletion argument was one step from being had.
> - **THE DECISION: with the page now resolvable, your existing fence FAILS honestly.** The
>   14 July generated fence — never executed until today — asserts
>   `interaction: fill #take-input`, and **the served page has no `id` attributes and no form
>   control** but the site menu button. Either the fence is stale (arena became a static
>   provocation display) or the tool is missing its submission control. **I did not choose,
>   and I did not fire the run** — a failing verdict files an `improve_tool` item routed to
>   the automated `tool-improver` against an `owned` page, which is the wrong way to settle a
>   design question about your tool. Command and both remedies are in the CONTRIB.
>
> Nothing waits on me. Two reusable instruments (offline fence trial + mutation prover,
> register **TL-036**) are in the CONTRIB if you want them — one of them refuted a check I
> had written to validate, on its first run.

## 5. Facts: the answer is "nothing to fix, and adding is a real choice"

Owner ruled SQL-verifiable counts only. **vonc's existing four all verify exactly**
(8/8 archetypes, 3/3 tools, 2/2 guides, 18/18 pages), each carrying its own query
and tolerance, re-checked by the enabled daily `evidence-freshness` task.

My earlier flag that "3 tools looks low" was **wrong**: the facts define themselves
by `page_type`, and I had counted name prefixes. Adding facts is therefore a
deliberate choice about what vonc may claim — not a repair — and every fact added
widens the surface for `bugs_open/149` C1, which still has no claims gate. Put a
proposal to the owner per site; do not bulk-add to make a checker's fact half work.

## 6. The one thing to carry out of this session

Four — arguably six — of this week's wrong calls were **one shape**: comparing
against a ruler I invented instead of the one the data declares. Counting pages by
name prefix when the fact says `page_type`. Grepping for "Monte Carlo" when the
technique need not name itself. Calling a pipeline idle from two observations four
minutes apart. Building a URL instead of reading `pages.url`. Comparing a full
digest against a truncated column. Reading a low visible-char count as breakage.

Three were caught only because someone re-checked a number already reported, and
one reached the owner as an urgent finding with a recommended action. **Re-ground
every figure before ESCALATING, not before acting on the reply** — urgency is
exactly when the check gets skipped and exactly when being wrong costs most.
