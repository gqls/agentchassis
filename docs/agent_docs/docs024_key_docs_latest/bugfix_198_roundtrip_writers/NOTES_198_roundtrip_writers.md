# NOTES — bugs_open/198 (two-writer stylesheet clobber)

Append-only, newest at the bottom. Evidence, commands, and every misstep.

---

## 2026-08-21 — the prevention half built and shipped (bugfix-198 lane)

### Starting state, and how much of it was already someone else's

The owner dispatched this session at `bugs_open/198`. First action was `who-owns.py` and a
read of the bug file — which by then ran to **982 lines across five lanes**, and had gained
five commits *that same day* from three other sessions. Re-reading it before writing
anything was the single most useful thing I did: **most of what I would have "discovered"
was already recorded**, and one thing I planned to do had already been done by someone else
while I was planning it.

What other lanes had done today, before I touched anything:

- fleet backfill of every empty theme row (9 rows) — ROUND-2 candidate 1, **done as data**;
- remortgagecalculator.uk and loanzy.uk clobbered at 10:27Z and restored the same morning;
- cookly.uk restored (18,047 B served) by the `news_editorial_features` lane;
- a `stylesheet_gutted` discovery check built (`e34b33a36`, IMP-055) with its enabling
  migration held at `541_..._HOLD.sql`, council `d3187418`;
- the two-clause `-ink` staleness check, and its correction.

Their §7 states plainly what was left: **candidate 2 (deploy-side shrink guard), the birth
guard, and candidate 6.** That is the scope I took, and it is why this session built no
restores.

### Measurements taken first-hand (not inherited)

Census via the exact `load_current_css` JOIN, 2026-08-21 ~15:00Z:

| | |
|---|---|
| linked theme rows | 22 |
| healthy | 19, at **13,650–26,917 bytes** |
| armed | 3, at **0–1,649 bytes** |
| nothing in between | 2,381 → 13,650 is empty |

That gap is what makes a 4096-byte floor defensible rather than arbitrary, and it is the
first thing to re-derive if the census ever changes shape.

Watched loanzy.uk mid-loop at 15:14–15:25Z: theme row **v13 → v15 → v26** while I queried
it, 14 patches appended in a day onto a base born empty on 08-18. Its first 600 characters
were *pure patch accumulation* — two blank lines, then provenance comment / rule, repeating.
No `:root`, no layout. That is the clobber's fingerprint, and it is unambiguous.

> **CORRECTED, same session, hours later.** I reported loanzy as "mid-clobber-loop, live" in
> my own summary. By the time I came to act on it, another lane had restored the row and the
> queued items were appending to a TRUE base — the same mechanism, running benignly. **The
> site was never going to need my intervention.** I had not mis-measured; I had let a
> measurement age across a decision boundary on a tree where nine lanes are working. The
> honest version: *loanzy was mid-loop at 15:14Z and restored by ~15:25Z.* A live-state claim
> needs its timestamp attached to the claim, not to the session.

### The three holes the Plan agent found in my design, all real

1. **A shrink-guard refusal would have minted `complete`.** `deploy_css.error_step` pointed
   at `complete_error`, which is a *success-labelled* `complete_workflow`, so the dispatch
   loop's `complete_work_item` stamps `complete`. The guard would have fired correctly and
   the ledger would have said it did not. This is now a LANDMINE entry, because it
   generalises: **any** guard routed to an `error_step` on this platform inherits it.
2. **`site_count = 1` alone does not protect a library theme** — a seed theme linked by
   exactly one site would be overwritten with that site's render. Added `origin <> 'seed'`.
3. **The refusal-forever worry was real but inverted.** A `needs_human_review` item is never
   re-promoted (the promoter selects `status='detected'` only), so no queue balloon — but
   `idx_swi_dedup` does **not** exclude that status, so a parked row HOLDS its dedup key and
   the finding cannot re-file even after the base is healed. Hence `result_fields.parked_by`
   and an explicit unpark sweep in the RUNBOOK. Without the marker the sweep would be
   approximate, which for a status humans also set by hand is not good enough.

### Proofs, and what each could have come out as

- **543's persist UPDATE, in a rolled-back transaction against live rows.** A real
  25,202-byte value onto dartsonline: `UPDATE 1`, v5→v6, and `md5(css_content)` equal to
  `md5(value)` exactly — byte-identical, which is the property that makes "deploy the whole
  row" safe. Then `UPDATE 0` four times: shared row, 100-byte fragment, unchanged content,
  seed row. **Four negatives and one positive from the same statement** is what makes it
  evidence rather than a demonstration.
- **Two Go mutations RUN, not asserted.** Deleting the `enforceFileShrinkFloor` call failed
  three tests. Measuring the **unprefixed** path failed its dedicated test *and let the
  clobber commit straight through* — which is the point of having that test: every lookup
  404s, every 404 reads as "new file", and the guard logs that it ran. Source restored
  byte-for-byte after each.
- **Built and tested from a clean `git archive HEAD` tree plus only my files.** The working
  tree fails an unrelated test (`render_context_step_boundary_resolver_test.go`) from
  another session's in-flight work; on HEAD it passes. Without the archive check I could
  have spent the evening on someone else's failure, or worse, assumed mine caused it.

### Missteps

**1. I checked a migration number was free with a query that could never have found one.**
`WHERE filename LIKE '54[23]%'` — SQL `LIKE` has no character classes, so that matches the
literal string `54[23]` and returns zero rows regardless of truth. Caught only because the
*same* query returned zero again ten minutes later, when the runner had just printed
`recorded` for both files: two contradictory answers from one query. The conclusion was right
for an unrelated reason (an `ORDER BY DESC LIMIT 5` listing I had also run), which is exactly
why it survived — **a worthless check that agrees with a good one is invisible.** Full entry
in `WRONG_CALLS.md`.

**2. I nearly ran `run-migrations.sh --apply`.** The dry run listed a large pending backlog
belonging to other lanes, several flagged by the script's own lint as replay hazards
("already applied and merely unrecorded"). Applying mine would have applied all of theirs.
Used apply-by-hand + `--record-only` instead. The non-obvious part is that **recording is not
bookkeeping**: my migrations carry probe guards that `RAISE` on re-application, so leaving
them unrecorded would have made the next `--apply` abort and block every later migration in
the queue, including other lanes'.

### What is deliberately NOT proven

The **witnessed live refusal**. It is proven in-transaction, by config probe, and by the
evaluator's own test suite — but not observed on a real dispatch. I chose not to induce one:
the only sites that would exercise the refusal arm are finetuning.uk and gaswholesalers.com,
both live, and a gate that failed would clobber them. That is the bug file's own closure bar
and it stays owed rather than being quietly claimed.

Likewise the shrink floor is **committed and inert** — it needs both a chassis and a
git-adapter roll. Post-roll probes are in the RUNBOOK §7 with their negative controls.

---

## 2026-08-21 (later) — the arm I missed, and the council round

### A graph query found a defect that reading my own edits could not

After 542 applied I ran an edge-resolution query over the whole workflow — every
`next_step`, `error_step`, `config.then_step`, `config.else_step`, each target resolved
against the step map. **I wrote it to catch a DANGLING edge after the rewire.** Every edge
resolved; the query "passed". Reading the 18-row table it printed is what showed:

```
check_saved | else | complete_error | ok
```

`check_saved` is not an `error_step` — it is a `conditional_branch`, and its refusal travels
on `config.else_step`. My 542 rewired the three `error_step`s and never touched it. So the
door 318 built on purpose — the guarded append matching zero rows when the model returns an
empty or oversized `css_added`, i.e. **the founding 2026-08-04 failure mode** — still landed
on `complete_error` and still minted `complete`.

**Reading the steps you edited cannot find the step you did not edit.** My 542 verify block
asserted every edge I had changed and was green; it had nothing to say about the one I had
not thought about. Migration 546 fixes the arm AND promotes the edge-resolution query into
the verify block as a post-condition, so a future migration that orphans a step fails at
apply rather than at runtime.

Worth noting what kind of check this is: it is a **structural** check over the artefact, not
a check of my intent. That is why it could contradict me. A verify block written from the
diff can only ever confirm the diff.

### Council: APPROVED round 1, and one objection was right about my reasoning

Six advisory objections, none high-severity. Four were checkable and I checked them rather
than accepting them:

- **the installed SQL string had never been executed.** `debug_historian` pointed at the
  landmine: step SQL is DATA to a migration's verify and parses only when the step RUNS. My
  542 proof was of the gate's *arithmetic*, via an equivalent hand-written SELECT — not of
  the string that actually shipped. Extracted the verbatim live query and ran it:
  dartsonline `26917 / 1`, finetuning.uk `1649 / 2`. `PREPARE` alone proves it parses; the
  execution proves both gate inputs are real. **This was the single most valuable objection
  of the round** and it cost one query to close.
- three `DisallowUnknownFields` sites fleet-wide, **none on the git path** (guardian).
- `bugs_closed/072` holds no prior persist-at-render or shrink-floor proposal (prior_art).
- adapter HTTP client timeout is 20s; one GET per opted-in file per commit (guardian).

**And the one that corrected me.** The `architecture` seat: I argued RFC_022's *shape*
exception covered the new key. It does — and it is a different check from the *accumulation*
gate, which the owner's ruling defines separately and which fires on the count alone
(17 carriers, 10 → 11 keys). **I used one gate to answer the other.** The remedy is the
estate's own: an ack in `optional_key_budget_acks.json` at 11 pointing at the review,
`check.py` mirrored, parity test green, overlay re-applied and verified at the ConfigMap
with a control. With the caveat that keeps it honest: `git_commit` is uncounted, so that ack
has **no automatic enforcement behind it** — it is a recorded judgement, not a live baseline.

### Second misstep of the session, same shape as the first

Both of today's errors were **checks that could not fail**: the `LIKE '54[23]%'` ledger query
(no character classes in `LIKE`, so it returns nothing whatever is true), and the 542 verify
block (a literal-match assertion over a string I had just written, which cannot discover a
step I never considered). Different mechanisms, same failure mode — *an instrument that
agrees with me by construction*. The transferable habit is the one the working-docs rules
already state and I applied twice too late: **before recording a check as passed, name the
result that would have failed it.**

---

## 2026-08-21 (evening) — the theme split, and the refusal WITNESSED

### The owner ruled, so the shared row stopped being a question

finetuning.uk and gaswholesalers.com shared one style_collection AND one seed theme row.
Migration 547 gives each its own. Three details that would each have been a defect if skipped:

- **Copy the composition FKs, don't regenerate them.** What renders `styles.css` is
  `render_css_from_spec` reading palette/layout/typography rows — *not* `css_content`. Copying
  them is what makes the next design run produce today's output. Getting this wrong would have
  silently redesigned two live sites at their next render, with no signal until it happened.
- **Copy the chrome pins.** `bugs_closed/170` in mirror image: that landmine warns about a fork
  that COPIES a pin it shouldn't; this was a fork that would have FAILED to copy one it must,
  dropping both sites' header and footer.
- **Leave the seed alone.** `professional-dark` is a library asset. It is now linked by zero
  sites and otherwise untouched.

Proven by running the whole file with `COMMIT` → `ROLLBACK` first, then applying. Only
`sites.style_collection_id` and a self-reference point at `style_collections` fleet-wide, so
repointing two rows was the complete change — checked rather than assumed, and both sites
re-verified serving identical bytes afterwards.

### The witnessed refusal, and how the subject was chosen

The bug file's closure bar was a live run. I had earlier written that I would not induce one
because the only candidates were live sites — that reasoning was right when I wrote it and
stopped being right once **webdesign.uk** turned out to be armed: 0-byte row, 15,582-byte
repo stylesheet, and a domain that 302-redirects so **its stylesheet is served to nobody**. A
gate failure there could not reach a visitor. That is what made the proof affordable, and it
only became visible because I checked a peer lane's "cleared as NOT damage" claim against the
repo instead of the URL.

Two things I did before dispatching, both of which were the difference between a proof and a
guess:

1. **Captured the 15,582-byte blob** as a restore net, md5 recorded.
2. **Ran the promoter's OWN selection query against my probe item** and confirmed all four
   doors opened. Without that, a probe that was silently *held* would have looked exactly like
   a probe that was refused by my gate — the same could-not-fail shape as this session's two
   earlier missteps, and I nearly walked into it a third time by planning to just wait and see.

Result, 19:09–19:11Z: promoted → claimed → run `76d9bc57` terminated at **`complete_refused`**,
never reaching `plan_css_fix`. Item `needs_human_review`, `parked_by` marker present,
`completed_at` NULL. **Row still 0 bytes at version 1 with `updated_at` still 2026-08-04; repo
still 15,582 bytes at the identical md5 with zero commits on the path.** Both negatives are
what make it evidence — had the gate mis-fired, both would have flipped.

The probe was **synthetic and declared** (`source='probe-bugfix-198'`, never `render-audit`)
and deleted afterwards. It must never appear in a findings census.

### The PASS arm was deliberately not manufactured

Driving the probe through a healthy base would have appended `SPAN.probe-198` — a rule
matching nothing — to a real site's stylesheet, which is precisely the `H3.H3` / `p.P`
pathology this bug already records on three sites. It needed no manufacturing: the
`remortgagecalculator.uk` lane observed the healthy path in production the same day, when
loanzy.uk's queued items appended to its restored base and deployed the whole file,
v21/17,906 B → v34/21,330 B with `:root` intact throughout. **Their incident is my pass-arm
evidence**, which is worth noting as a pattern: on a tree this busy, the arm you decline to
induce may already have been witnessed by someone else.

### Fleet state

19 PASS / 3 REFUSE (542) → 21 / 1 (547) → **22 / 0** (548). Every linked theme row is a
plausible, unshared stylesheet, and 543 maintains that at every render.

## 2026-08-24 — my candidate-6 remedy was INCOMPLETE, and the naive version is worse than the bug

The `bugs_open/352` lane picked up the spun-out candidate (6) today, and its first measurement
corrected the fix I had written down. Recording it here because the error is mine and it was
caught by a peer, not by me.

**What I wrote** (`bugs_open/352` fix candidate 1, and implied by this lane's handoff §4 item 2):
emit the class when there is one and omit the class component when there is not, so the finding
says `h3` not `H3.H3` — "this makes the bad selector unrepresentable at source and is a few
lines". I flagged the `item_key`/dedup interaction as the thing to check before applying.

**What is wrong with it.** The dedup interaction was the *lesser* risk and I named only that.
Today `p.P {…}` matches nothing, so it is **inert**. Lowercased to `p`, css-patch-agent appends
it to the **site** stylesheet and recolours every paragraph on the site. The fix as I stated it
converts a dead rule into a live site-wide restyle — a worse outcome than the defect. The
remedy has to produce a **scoped** selector (ancestor- or id-anchored), not merely a lowercase
one.

**The scale, measured by me against the live DB on 2026-08-24** (the 352 lane's figures, which
I reproduced independently rather than quoting — all four matched exactly):

| | count as of 2026-08-24 |
|---|---|
| `contrast_failure` rows, all statuses | 452 |
| …carrying a `TAG.TAG` selector | **181** |
| commonest: `P.P` / `A.A` | **77** / **44** |
| then `H2.H2` 16, `H3.H3` 16, `LEGEND.LEGEND` 7, `H1.H1` 6 | |

Predicate: `split_part(sel,'.',1)=split_part(sel,'.',2)` on `split_part(item_key,'#',2)`. So the
two commonest cases are precisely the two most dangerous bare selectors, which is why "just
lowercase it" fails on the majority of the population rather than an edge of it.

**Status breakdown of those 181** — and this is the figure I would have missed:
`complete` **108**, `deferred` 58, `unresolved` 15. Two readings:

- **108 are already falsely `complete`.** This file recorded exactly ONE instance (the
  dartsonline `H3` row, §562-571, marked `complete` 08-18). It generalises to 108 fleet-wide.
  The already-lost repairs outnumber the at-risk ones.
- **73 sit outside `workItemClosedStatuses`** (`platform/orchestration/actions/work_items_common.go:85-91`
  = complete/verified/rejected/wont_fix/cancelled), so a key-shape change would let the
  retraction path close them stamped "no longer below its contrast threshold" — false. The 352
  lane estimated this at "~84"; the measured value is **73**, which I fed back.
  ⚠ `unresolved` IS in `workItemTerminalStatuses` (line 42-48) but NOT in the closed set; the
  asymmetry is deliberate and documented at line 97 — do not "tidy" it while fixing this.

**The check that would have caught me.** Before writing a selector-shape fix candidate, census
the selector population and ask what the corrected selector MATCHES, not just whether it
matches. One `GROUP BY` would have shown `P.P` ×77 at the top and made the blast radius
unmissable. I reasoned from the three sites this file happened to contain — `H3.H3` on
dartsonline and `p.P` ×2 on remortgagecalculator — where `h3` is a narrow selector and the
sample hid that the modal case is `p`. **A fix candidate derived from the instances a bug file
collected inherits that file's sampling bias**, and the bias is invisible because the instances
are real.

Second-order: this is the same shape as the estate's "your own action can silence your own
detector" family, one step earlier — the sample that motivated the fix also flattered it.

## 2026-08-24 (later) — the fresh chassis roll: DGH-016 re-verified live, and the observation is owed for a MEASURED reason

Fleet rolled today. Re-checked both halves of DGH-016 rather than assuming the roll preserved
what 08-22 verified.

**Both services run commit `70fd163c2`**, of which the shipping commit `4ee9bfff6` is a proven
ancestor:
- **git-adapter** — its own `build provenance` line, `git_commit: 70fd163c2…`.
- **chassis** — the provenance line had already scrolled (absent from `--tail=400`, which per
  LANDMINES means "not in range", NOT "unstamped"), so in-pod
  `grep -aq <sha> /proc/1/exe`: **present**, with a **fabricated sha correctly absent** as the
  negative control.
- ⚠ Worth recording because it narrows an existing landmine: the in-pod probe **does** work for
  the commit sha, even though DGH-016's own entry records it giving FALSE NEGATIVES for the
  message constants on 08-22. Different sections of the binary. "In-pod grep is unreliable" is
  too broad a lesson to carry forward from that incident.

**Two controls, both directions, on the ancestry method itself.** My first negative control
FAILED — and the control was wrong, not the method. I picked my own newest commit (`db7781409`,
15:10 BST) expecting it to be absent; the build is `70fd163c2` at **16:11 BST**, so my commit is
legitimately IN it. A negative control has to be something that *cannot* be in the build: I
re-ran with `313421727` (17:15 BST, after the build) → correctly FALSE, and `4ee9bfff6` →
correctly TRUE. **A control that cannot come out the other way is not a control** — and picking
one that predates the artefact is an easy way to build that mistake.

⚠ **Timezone nearly manufactured an anomaly.** The pod log stamps `Z`; git stamps `+0100`. The
pod's `15:39:46Z` is **16:39 BST**, i.e. AFTER the 16:11 BST build commit — consistent. Read
naively as the same clock it reads as a pod running code that did not yet exist, which is the
kind of "impossible" finding that sends you hunting a deploy bug that is not there.

**Item 1 (the live enforcement observation) is still owed — and now for a reason I have
measured rather than assumed.** `grep -c 'shrink floor'` over the adapter's last 3,000 lines is
**0**. That zero is explained by demand, not by a broken guard: **0** commits in that window
carry a file key ending `.css`, against **253** commit/push lines for other file types.

⚠ **The demand control took three attempts and the first two PASSED WHILE BLIND**, which is the
part worth keeping:
1. *"any commit activity"* → 253. Licenses nothing: the guard fires on stylesheet writes, not
   commits in general.
2. *"any payload containing CSS"* → hundreds of matches, all of them **inline `<style>` blocks
   inside ordinary HTML page commits**. Looks like exactly the right control and is not.
3. *"file KEY ends in `.css`"* → **0**. This is the axis the guard actually varies on.

The first two would each have let me write "no css deploys have happened" or "the guard is
silent on live traffic" with a number attached. Same lesson as this morning's, third time today:
**the measurement answers the question you ENCODED.**

## 2026-08-24 (later still) — the owed inventory: STARTED, and its documented method has a FOURTH blindness that hides 198 itself

First real pass at the round-trip-writer inventory owed since council round `5249320e`
(2026-08-05). Not finished. What follows is the population work (method steps 1–3) and one
correction to the method itself.

**Step 1 — LLM output ground truth.** `execute_llm_prompt` **140** steps across **66** active
definitions, plus `generate_html` 4/2, and one each of `execute_vision_prompt`,
`fetch_llm_news`, `generate_provocations`, `ch_llm_review`. (`as of 2026-08-24`.)

**Step 3 — the join, and how it moved.**
- Matching writer refs against LLM **`output_field` names** reproduces the handoff's floor
  exactly: **1 row**, css-patch-agent's own `save_css_to_db`. Confirms the floor; adds nothing.
- Matching against LLM **step names** instead — because workflow refs usually name the STEP —
  gives **20** writer steps across 19 definitions. That looks like 20× the floor.
- ⚠ **It is not.** Filtering to steps whose query actually starts `UPDATE`/`INSERT` collapses it
  to **2**, because most matched `query_database` steps are SELECTs. And one of those two,
  `component-template-fixer/create_section_edit_delivery`, writes `site_work_items` — a work
  item, not an artefact, so not the 012/198 class at all. **I nearly reported 20.**

**⚠ THE METHOD'S FOURTH BLINDNESS — it cannot see 198 itself, which is the bug it exists to
generalise.** The handoff lists three known blind spots. Here is a fourth, verified:

```
deploy_css   :: git_commit      :: content_field = css_saved.css_content
save_css_to_db :: query_database :: (UPDATE, from the LLM step)
```

`deploy_css` — the git writer that actually gutted nine stylesheets — references
**`css_saved.css_content`, the output of the SAVE step, not the LLM step.** The artefact travels
**LLM → DB row → git commit**, and a one-hop join from LLM steps to writer refs cannot see the
second hop. So the documented method under-reports precisely the multi-hop shape that motivated
the survey, and a `git_commit` count of **0** from that join reads as "no git writers are
exposed" when the motivating incident was a git writer.

**What the method needs:** transitive closure. Follow each writer's referenced field back through
intermediate steps until it either reaches an LLM `output_field` or a non-LLM source — not a
single join. Until that is built, any population figure from this survey is a FLOOR, and I have
not built it yet.

**Not asserted:** I am NOT claiming a fleet-wide count today. The honest state is: ground truth
enumerated (140/66), the one-hop join characterised and its yield shown to be misleading in both
directions (1 too low, 20 too high, 2 after filtering, of which 1 is out of class), and a fourth
structural blindness found and verified. Step 4 (read each candidate's PROMPT for a
whole-artefact vs fragment contract) is untouched.
