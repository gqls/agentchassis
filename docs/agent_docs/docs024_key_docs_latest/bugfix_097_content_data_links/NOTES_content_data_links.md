# NOTES — content_data link resolution (bugs_open/097)

Append-only, newest at the bottom. Evidence, commands, what the system actually
said — **and every misstep**, which is the part the next thread cannot rederive.

---

## 2026-08-02 — picking the bug, and the ownership check that mattered

`scripts/who-owns.py` returned "OWNED or recently active" for nearly every
candidate I tried, because it counts commits that merely TOUCH a bug file — and on
this tree every bug file is touched constantly. It is lagging by construction and
it says so.

What actually discriminated was grepping the **live `.jsonl` transcripts** for the
code symbols, not the bug numbers:

```
ctaFieldNames        max 6 mentions in any live session (a cross-reference)
resolve_internal_links   max 10
landmine-verifier    112 in 693556a1  -> bugs_open/163 IS being worked; dropped it
bugfix_165           180 in 806dfccd  -> dropped
bugfix_154           209 in 9de5c96a  -> dropped
```

Counting `bugs_open/NNN` alone is nearly useless: every session that runs
`ls bugs_open/` picks up all 60 numbers at once. **The signal is the SYMBOL, and
the discriminator is one session having a large count while everyone else has a
handful.**

## 2026-08-02 — the bug as filed is half-done, and reading to the bottom is the whole job

097's fix candidates are at the TOP of the file and are 6 days stale. Its live
state is in the **last** dated section, and it says the repair half shipped
(4 seams) and the detector half did not. Acting on the candidate list without
reading to the bottom would have rebuilt `RepairPageLinks` for the third time.

Then `component_link_repair.go` — committed by the **136 lane this morning**, hours
before I started — turned out to name my scope precisely and route it here:

> *"content_data. Same limit as 079's fix … The deployed artefact stays covered by
> the outbound rerender seam (repairOutboundPageLinks, bugs_open/097)."*

That is a clean handoff between lanes and it is why the scope is not a guess.
**Read the files your bug's siblings committed TODAY, not only the bug file.**

## 2026-08-02 — the measurement, and why the SQL version was not good enough

The SQL census (RUNBOOK R2) said 51 unresolved occurrences. Running the **shipping
code** over the same corpus (R1) said 52. The difference is small and the reason
matters: the SQL had to hand-reimplement `NormalizePagePath` and
`ClassifyLinkScope`, and a hand copy of a shared definition is exactly the drift
class this platform reviews for. The Go number is the one quoted anywhere.

```
components audited      : 885
components with findings: 13
REWRITE (target exists) : 19
PHANTOM (report only)   : 33
```

**The 872 components with NO findings are the load-bearing half of that result.**
They are the evidence that a name-based nomination with a value-based judgement
does not fire on prose, on assets or on off-site links — with no exclusion list
anywhere in the code.

## 2026-08-02 — MISSTEP: I nearly shipped two guards that made each other untestable

`TestFindingOrderIsStable` passed. I then mutated the code to prove the ordering
guard was load-bearing, and **it still passed**. Cause: I had written *two*
mechanisms guaranteeing one property — `sort.Strings(keys)` inside the walk and
`sort.SliceStable(findings, …)` at the exit. Either one alone makes the output
deterministic, so **deleting either left every test green** and neither was ever
demonstrated to do anything.

This is the memory-file shape *"two checks blind the SAME way AGREE with each
other"*, in the writing direction rather than the measuring one. A second
mechanism that cannot be distinguished from the first is not belt-and-braces — it
is a guarantee nobody can test.

Fixed by **deleting the redundant one** (the key sort) rather than keeping it as
decoration, after which the mutation reproduced immediately:

```
run 0: [a_url b_url c_url d_url e_url f_url]
run 1: [d_url e_url f_url a_url b_url c_url]     <- Go map randomisation, visible at last
```

The check: **after adding a guard, delete it and watch a named test fail.** If it
does not, either the test is vacuous or something else is already doing the job —
and both are worth knowing before the commit, not after.

## 2026-08-02 — MISSTEP: two mutations that "passed" had simply failed to compile

Two of the five mutation runs printed `FAIL` and I nearly recorded them as proof:

```
platform/orchestration/datahelpers/content_data_links.go:163:16: invalid operation: v[i] ... is not an interface
FAIL  github.com/gqls/agentchassis/platform/orchestration/datahelpers [build failed]
```

`[build failed]` is not a red test — it is no test at all. A malformed mutation
proves nothing about the guard, and the word FAIL in the output makes it look like
it did. Both were redone until they compiled and failed on an **assertion**.

## 2026-08-02 — the archived page that would have looked like a false positive

`robot-hands.com` cards link `/learning-center`, and the fix rewrites that to
`/learning-center.html`. Checking whether that was correct turned up:

```
 learning-center         | /learning-center.html        | active   | deployed
 learning-center-index   | /learning-center/index.html  | archived | deployed
```

`NormalizePagePath('/learning-center/index.html')` is `/learning-center`, so if the
index included archived pages the link would have RESOLVED and produced no
finding at all. It is excluded because `loadValidPagePaths` uses the shared
`linkablePageStatusPredicate` (`status NOT IN ('deleted','archived')`) — which the
offline census had to match exactly or the whole measurement would have been of a
different population. **Copy the predicate; do not retype it from memory.**

(That archived-but-still-served page is `bugs_open/098`'s shape — archiving does
not undeploy. Routed there, not worked around here.)

## 2026-08-02 — what I deliberately did NOT do, and why each was a judgement

- **Did not blank phantom values.** It is the `content_data` analogue of the
  unlink arm, and `link_repair.go`'s own header records that arm as unsettled by
  two council seats. Pinned by a test so reversing it is deliberate.
- **Did not touch the staged CTA precedence flip** (`ctafields.go`, trail
  `2525f980`). It belongs to the `cta_link_integrity` lane, carries 5 binding
  constraints from 5 seats, and inverting precedence is that round's job. Named in
  the submission so its reviewers are told rather than left to measure it.
- **Did not file `site_work_items`.** `bugs_open/083` (nothing drains `detected`)
  and `bugs_open/077` (no items whose handler has no remit) — the same reasoning
  `writeLinkRepairLog` already wrote down.
- **Did not write a migration.** The live damage clears on each page's next save.
- **Did not run `090` (needs_diagnosis).** Stated plainly per the owner ruling of
  2026-07-31: 097's mechanism was already CONFIRMED by diagnosis `9543aaf1`, and
  what I added is not a new root-cause claim but a census — I read the exact code
  (`ctaFieldNames`, `DeriveCTAURLFields`), read the live `input_schema` that hides
  the field, and ran the shipping function over all 885 production rows. That is
  first-hand verification of the same kind the loop would have performed, on a
  cause the loop has already confirmed once.

## 2026-08-02 — the latent false positive I went looking for, and why I did NOT guard it

After the fix was committed I went back at my own design looking for the thing I
had reused outside its brief. Found it: `ClassifyLinkScope`'s last arm is
`default: LinkScopePage`. That is correct for an **href attribute** — anything not
external, mailto, anchor or asset really is a page link. I feed it a **field
value**. So in principle a field called `link` holding the words "Read more"
classifies as a page link and gets reported as a phantom.

This is `016b` §9's own entry from `bugs_open/093`: *a shared predicate written
for one input shape, reused on another.* Worth measuring before deciding anything.

```
url-named field values in production: 1299
  asset      168      external   457      empty     16
  internal     2      mailto       1      page     655

PAGE-scope values NOT starting with '/'  (the false-positive surface): 0
```

**All 655 page-scope values begin with `/`.** The surface is empty today, and the
644 non-page values are removed by the classifier without the code naming any of
them — which is the same control as the 872-of-885 result, from the other side.

**I did not add a shape guard, and that is a decision rather than an omission.**
Requiring a leading `/` would change nothing today and would swap a latent false
POSITIVE for a latent false NEGATIVE: a genuinely relative dead link
(`about.html`) would go silent. When the whole point of the arm is to *report*,
under-reporting is the worse failure. The limit is written into the file header
with the measurement and with the discriminating check (re-run RUNBOOK R1 and look
at page-scope values that do not start with `/`), so if prose ever does turn up in
a finding, the evidence for adding the guard arrives with the symptom instead of
being argued about in the abstract.

**The transferable bit:** "I reused a shared predicate" is not by itself a defect
and not by itself safe. What settles it is *measuring the input population you are
actually handing it* — and then writing the measurement down beside the reuse, so
the next reader inherits the evidence rather than the assumption.

## 2026-08-02 — COUNCIL APPROVED at round 1, and the four advisory objections CHECKED rather than dispositioned

`40c0c14d-636c-4d6f-b3a2-9316267d7367` — **approved**, 12 reviewers, 0 unreadable,
5 abstained, **4 advisory objections, none high-severity**. Approving seats:
`reuse_agent`, `guidelines`, `compliance`, `render_guardian`, `constitution`,
`mission`, `prior_art_librarian`, `architecture` (signal: **point_fix**;
"the RFC trigger test does not fire"). Objecting-but-advisory: `editquality`,
`bug_historian`, `guardian`, `debug_historian`.

Every objection below was **checked**, not argued with. Two of them found
something.

### 1. editquality, MEDIUM — "no cited fact shows the rerender path actually calls through SavePageSectionsAction"

Fair, and it is the load-bearing claim: I had asserted it from another file's
header comment, which is folklore, not evidence. Checked against **live**
`agent_definitions` (the seed is not the system):

```sql
SELECT ad.type, s.key AS step_name, s.value->>'action' AS action,
       s.value->'config'->>'sections_metadata_field'
FROM agent_definitions ad,
     LATERAL jsonb_each(ad.default_config->'workflow'->'steps') AS s(key, value)
WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
  AND s.value->>'action' = 'save_page_sections';
```

```
 page-build-handler      | save_sections | save_page_sections | page_content.response.sections_metadata
 page-rerender           | save_sections | save_page_sections | rerender_sections.sections_metadata
 tool-recreation-handler | save_sections | save_page_sections |
```

**CONFIRMED.** `page-rerender` persists through the same chokepoint, and its
`sections_metadata_field` is `rerender_sections.sections_metadata` — i.e. the
output of the very step that rebuilds each section **from `content_data`**. So the
third-representation claim holds on live config, not on a comment.

**The check also nearly caught me repeating a stale figure.** That first query
returns **3** agents, and every doc in this family — including my own — says
**six**. Widening it showed why both are right: `pageflow-builder`, `page-rebuild`
and `site-work-orchestrator` reach the action through a `loop` step, so they do
not match on `s.value->>'action'` and are invisible to the obvious query.

```sql
-- the honest form: text-match the whole config, then look at WHERE it matched
SELECT type FROM agent_definitions
WHERE default_config::text LIKE '%save_page_sections%' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

Nine rows — the six persisters, plus `council-gate`, `diagnose-agent` and
`fix-proposer`, whose hits are **prompt text** (a footprint map in `select_panel`,
and a reviewer prompt), not steps. So **six is re-verified as correct today**,
2026-08-02, rather than inherited from 2026-07-28. ⚠ **The narrow query
under-counts by 3 and looks authoritative doing it** — a persister behind a `loop`
step is exactly the kind of thing a `->>'action'` filter cannot see, which is this
bug's own shape (a predicate that answers for the level it was written for) turning
up in the measurement of the fix.

### 2. guidelines + guardian + debug_historian — `page_components.locked_at`

**Three seats reached for the same question independently, which is the strongest
signal in the round.** *"the rewrite arm would mutate content_data for a slot an
editor deliberately froze."*

**Checked in code and it cannot happen — but not for the reason I would have
guessed.** The lock guard sits at `save_page_sections_action.go:576`, and it works
by *discarding the whole rebuilt section*: the locked row is kept out of the
DELETE, and the insert loop `continue`s before writing anything. So my rewrite
lands in `section.ContentData`, and that object is then **thrown away** along with
the rest of the fresh copy. Nothing frozen is ever mutated.

**And there is a real, small consequence I did not spot and the seats did not
either:** the audit still writes its `CONTENT_DATA_LINK_AUDIT` record for findings
in a locked slot's *discarded* copy, so the record slightly over-reports. **This
is not new and it is not mine to fix unilaterally** — `repairSectionLinks` has the
identical property (it repairs a locked section's HTML in memory, writes
`CONTENT_LINK_REPAIR_DETAIL`, and that HTML is discarded too). Fixing it for one
of the two passes would manufacture exactly the asymmetry this bug is about.
Recorded as a shared property; if it is worth fixing, both move together.

### 3. debug_historian, MEDIUM — "'clears as a side effect of ordinary operation' is asserted from a count, never queried"

**The sharpest objection in the round, and it was RIGHT.** It cited the matching
`WRONG_CALLS.md` precedent by name (*"'The configs will self-heal once the code
ships' — asserted from a count, never queried"*). So I queried it:

| domain / page | locked slots | last component write | days |
|---|---|---|---|
| gaswholesalers.com / supply-terms-and-eligibility | 0 | 2026-07-10 | **23** |
| idea.uk / about | 0 | 2026-07-15 | 18 |
| leopardess + finetuning / llm-cost-calculator | 0 | 2026-07-24 | 9 |
| finetuning.uk / index | 0 | 2026-07-27 | 6 |
| idea.uk / index | **4 of 6** | 2026-07-28 | 5 |
| robot-hands × 4, ai-agent-orchestration × 1 | 0 | 2026-07-31 | 2 |

> **CORRECTED — "no migration needed, it clears as a side effect of ordinary
> operation" was too confident.** Direction right, tail unmeasured. Resave cadence
> across the 11 affected pages runs **2 to 23 days**, and nothing *guarantees* a
> page is ever re-saved. Convergence is **opportunistic, not scheduled**, and the
> honest statement is that the fast-moving pages clear within days while
> gaswholesalers may sit for weeks.
>
> The lock half came out better than feared: **zero of the 13 affected components
> are locked.** idea.uk/index does carry 4 locked slots — `hero`,
> `brief-explanation`, `tool-list`, `call-to-action` (all from the "home-CTA funnel
> fix … do not auto-recompute" lock) — but the affected slot there is
> `info-card-grid`, which is **not** locked. So no finding is behind a lock today.

The re-measure in the bug file's owed list is now the thing that settles this
rather than an assumption: **the 52 must be FALLING**, and a static count means
the pass is not being reached.

### 4. guardian + reuse_agent + architecture, LOW — the third `agent_error_log` code

*"asserts existing queries are unaffected but that should be verified against the
actual shape of `error_code` usage."* Verified by grep over the whole tree: every
`error_code` predicate is either exact equality on `CONTENT_LINK_REPAIR_*` or a
`LIKE` on an unrelated prefix (`tool_crosslink_not_emitted%`,
`component_validation_%`). Nothing catches `CONTENT_DATA_LINK_AUDIT`, and the two
families diverge at the ninth character (`CONTENT_DATA_` vs `CONTENT_LINK_`).

**And this found a real omission.** The estate already has a convention for
exactly this — `TestDiscoveryCheckErrorCodeIsDistinct` keeps a list of taken
codes — and I had not followed it. Fixed: `CONTENT_DATA_LINK_AUDIT` added to that
list so the *next* code cannot collide with mine, plus
`TestContentDataLinkErrorCodeIsDistinct` asserting both non-collision **and
prefix-disjointness**, since the estate demonstrably writes `LIKE` queries against
this column.

### 5. bug_historian, MEDIUM — "a second confirmed-but-undriven exposure with no owner"

*"nothing in the plan schedules the `bugs_open/136` single-writer follow-up."*
Checked: it **is** ticketed — `bugs_open/136` is open, and concept-register
**LNK-027** names the tool-markup writers as the "ranked next candidate", left out
on collision grounds rather than merit. So the follow-up has a home and a named
next step. The seat could not see that from the plan text, which is a fair
criticism of the submission rather than of the change; noted for next time —
**cite the ticket that owns your scope-out, not just the scope-out.**

### The one thing no seat could check, and it is worth keeping

`reuse_agent`'s `missing`: *"whether any other in-flight lane is independently
building a content_data walker right now — `code_checks` cannot confirm no
parallel WIP exists."* That is the exact blind spot in my memory file
(`who-owns is blind to uncommitted sessions`). I had checked it the only way that
works — grepping live `.jsonl` transcripts for the code symbols — before starting.
Worth stating in the next submission's `grounded_in` rather than leaving a seat to
flag it as unknowable.

## 2026-08-02 18:4x — LIVE on v1.0.1229, pod-verified on BOTH replicas with a full control set

The roll landed. **The image is not the evidence — the binary is** (`bugs_open/153`:
a roll is not evidence your fix shipped, and the image carries no provenance).

```
pods: agent-chassis-79479769b9-g7fbt, -n8nbj   image v1.0.1229
binary mtime (both): 2026-08-02 18:28:49 UTC
my commit d78f70bf1:  2026-08-02 10:45:30 UTC   <- binary postdates it by ~7h45m
```

Same exec, both replicas, identical results:

| grep | g7fbt | n8nbj | what it proves |
|---|---|---|---|
| `audited content_data internal links before persist` | **1** | **1** | the new log line is compiled in |
| `CONTENT_DATA_LINK_AUDIT` | **1** | **1** | the new error code is compiled in |
| `RepairContentDataLinks` | **4** | **4** | the datahelpers symbol is present |
| `repaired dead internal links before persist` (LNK-024, live since 1196) | 2 | 2 | **positive control** — the grep pipeline works and this is a real binary |
| `CONTENT_LINK_REPAIR_DETAIL` (live) | 1 | 1 | second positive control |
| `CONTENT_DATA_LINK_INVENTED` | **0** | **0** | **negative control** — the grep is not matching everything |

> **The negative control here is INVENTED, and that is weaker than the real thing.**
> `bugs_open/153`'s rule wants a string the change **removed**, expect 0 — that
> discriminates a stale image from a fresh one. This change removes nothing (it is
> purely additive), so no such string exists and an invented control can only prove
> the grep is not universally matching. What carries the load instead is the
> **binary mtime vs commit time** comparison above. Stated rather than glossed,
> because "negative control: 0" reads as stronger evidence than it is here.

> **GOTCHA — `agent_error_log`'s timestamp column is `occurred_at`, not
> `created_at`.** Two queries failed with a bare `column "created_at" does not
> exist` before I read `\d agent_error_log`. Schema first, every time.

## 2026-08-02 18:5x — 0 findings after the roll had TWO possible causes, so I induced it

Six minutes after the roll, 12 `page_components` had been written (all
loancalculator.co.uk, another lane's work) and there were **zero**
`CONTENT_DATA_LINK_AUDIT` rows — *and* zero `CONTENT_LINK_REPAIR_DETAIL` rows.

That is the memory-file shape *"a gate's 0 findings has TWO causes with opposite
fixes"*: the pass ran and those pages were clean, or the pass never ran. **An
absence cannot distinguish them, and waiting for organic traffic would not either**
— a quiet result would look identical on both hypotheses for as long as I watched.

So: induce, on a page whose defect is already measured. Target chosen for three
reasons — `gaswholesalers.com/supply-terms-and-eligibility` has 6 known findings
(2 rewritable `/contact`, 4 phantoms), **zero locked slots**, and was last written
2026-07-10, so no other lane is in it. (idea.uk was the other candidate and was
**rejected**: one live session shows 976 mentions of it in the last 90 minutes.
Checking that before dispatching is the whole point of the transcript grep.)

## 2026-08-02 18:49Z — INDUCED. Both arms, in production, exactly as predicted.

Work item `ab409727-4dd3-48b0-8e7c-1a2e3682702d` → `complete`, `error` NULL.

**The prediction was written before the dispatch and the result matched it
element for element** — 2 rewrites (`cards[4]`, `cards[5]`), 4 phantoms
(`cards[0..3]`), component `info-card-grid`, paths spelled with array indices.
Writing the *success* criterion in advance, not just the failure mode, is the
`WRONG_CALLS.md` lesson from 2026-08-01 (*"a prediction that names only a failure
mode makes success unexamined"*), and it is why `complete` was a question here
rather than an answer.

**The control that actually carries the claim is the four slots that did NOT
move.** All six components were rewritten at 18:49:04, and five are byte-identical
to their pre-run `md5(content_data)`. Had I only checked "the card links changed",
a pass that rewrote everything it touched would have looked identical.

**The two representations converged 10 ms apart on the same save:**

```
18:49:04.587  CONTENT_DATA_LINK_AUDIT     rewritten=2  phantom=4    the source
18:49:04.597  CONTENT_LINK_REPAIR_DETAIL  rewritten=2  unlinked=4   the markup
```

Same six links, same split, from one page index — the property the design exists
for, observed rather than argued. Served page: `7 × href="/contact.html"`, zero
phantoms. Stored, rendered and served all agree.

**Census 52 → 50, and only the induced domain moved** (gaswholesalers 6 → 4;
robot-hands 17, idea.uk 12, finetuning 7, ai-agent-orchestration 5, leopardess 4,
vonc 1 — all flat). That localisation is what makes it evidence: a total that fell
for some other reason would have moved somewhere else too.

> **An observation I did not expect and am recording rather than acting on:** the
> `domain` column is **empty** on both rows this run. `repairSectionsBeforePersist`
> reads it from `site_record.domain` in `CollectedData`, which the page-rerender
> path apparently does not populate. It affects the pre-existing
> `CONTENT_LINK_REPAIR_DETAIL` row identically, so it is a **shared property, not
> something this change introduced** — and fixing it for one writer would
> manufacture the asymmetry this bug is about. Worth a ticket if anyone tries to
> group these rows by domain; `site_id` is populated and is the reliable key.
