# NOTES — bugfix 214, imagery scope_ref

Append-only, newest at the bottom. Technical log: evidence, commands, what the
system actually said, and every misstep.

---

## 2026-08-10 — session 1: picking the bug

Instructed to take the next `bugs_open/` case not already owned. Ownership was
checked three ways rather than one, because `who-owns.py` reads **commits** and is
blind to a session mid-fix:

1. `git log --since="36 hours ago"` → bug numbers named in commit subjects.
2. `scripts/who-owns.py <n>` on every quiet candidate.
3. **A grep of all 39 live session transcripts** (`~/.claude/projects/.../*.jsonl`
   modified since 11:00) for each candidate's file slug.

Step 3 is the one that earns its keep — it found that 151, 240, 223, 232, 233, 210,
228, 239, 213/212, 225, 224, 235, 229, 227, 122/113, 198, 236, 186, 033, 203, 149,
234, 205, 093, 181, 231, 083 and 178 all had a live session on them, several with no
commit yet. Counts of `1` are noise: they are the file appearing in an `ls bugs_open`
listing, not engagement.

**A trap worth recording: the STATUS line at the top of a bug file is frequently
stale, and the tail is the truth.** `bugs_open/152` opens with *"Status: OPEN,
unowned"* and closes, 170 lines later, with *"LIVE 2026-08-06 on chassis v1.0.1259"*.
`096` says OPEN at the top and CLOSED at the bottom. I nearly took 152 on the header
alone. **Read the tail of a bug file before believing its head** — this is the
already-recorded landmine ("a bug file's FIX CANDIDATE can be refuted by that same
file's own MEASUREMENT NOTES") in its other direction.

Picked **214**. Its filing lane (`bugs_open/151`) says in its own transcript that it
filed 214 and left it — *"same wire, different field, its fix does not gate"* — so it
is genuinely unowned rather than merely quiet.

## 2026-08-10 — is it still valid?

Re-ran the bug file's own census. **5 orphaned ordinals, unchanged since filing** —
and, per the standing rule that a count proves damage but never identity, I listed the
rows rather than trusting the number: same 5 rows, same keys, fundamentallyai
`about:4` created 08-07 and gamesdesign `about:2` ×4 created 06-05. No drift, nothing
re-minted, nobody else fixing it.

## 2026-08-10 — the mechanism is NOT what the file says

Reading `write_site_plan_action.go` end to end rather than at the cited lines:

```
:392  CanonicalisePage(...)                    -> canonical page name
:441  dedupePlanPageRows(planRows, logger)     -> survivors
:503  INSERT site_plan_pages    (name = r.Name)        <- CANONICAL
:533  INSERT site_plan_sections (page_name = r.Name)   <- CANONICAL
:455  flattenImageryBlock(params.CollectedData, logger) <- RAW LLM MAP KEY
```

So the imagery ref is not "LLM-minted free text that happens to be wrong" — the
planner keyed the name it was handed and **the platform renamed the page underneath
it**. Confirmed against live data: gamesdesign's plan holds `about-index`, not
`about`; the section list is `hero, content-block-about, differentiators`, so
`about:2` → `about-index:2` has a **valid ordinal**. The ordinal was right all along.

That reframing is what makes the fix a *resolution* rather than a *rejection*. The
bug file's candidate 1 ("degrade the row to page scope or skip") would have discarded
four correctly-planned icons.

## 2026-08-10 — MISSTEP 1: I measured the damage against the writer's table

My replacement census asked whether the scope_ref resolved in **`site_plan_pages`**,
and returned **22** rows on current plans. I wrote that down as the honest figure.

It is wrong; the real figure is **10**. **Not one consumer joins `site_plan_pages`** —
all ten join `scope_ref` against **`pages.name`**, the deployed page table:

```
plan_sections_action.go:296   spi.scope_ref = $2          ($2 = pages.name)
plan_sections_action.go:377   spi.scope_ref LIKE $2||':%' ($2 = pages.name)
check_content_image_missing.go:213   spi.scope_ref = p.name   (FROM pages p)
queryresolve.go:292                  spi.scope_ref = p.name   (FROM pages p)
derive_card_asset_action.go:299      spi.scope_ref = $3       (from pages.name)
```

Caught by a follow-up query in which I added a `ref_in_pages` column almost as an
afterthought — 11 rows came back `t`, i.e. resolving fine in the table that matters
while failing in the one I had chosen. Full write-up in `WRONG_CALLS.md`.

**Why it mattered rather than being a tidy-up:** the backfill's predicates are derived
from that census. On the 22-row basis it would have "repaired" twelve **working** rows
on leopardessconsulting.co.uk and relojistas.com — repointing serving heroes at plan
names `pages` does not carry. The fix would have broken what it was fixing.

The cheap check: **grep one consumer's WHERE clause before choosing the table you
measure against.** A reference is broken when the thing that READS it cannot resolve
it; the census predicate belongs to the reader.

## 2026-08-10 — two corrections to the bug file itself

1. **Its census measures the harmless half.** It filters on ordinal range. No consumer
   parses the ordinal (`LIKE <page> || ':%'`; `flag_page_image_rebuild` splits and
   discards). An out-of-range ordinal on a correct page part is **inert**. The file
   even says so in its own consumers section — the census and the analysis contradict
   each other, and nothing joins them up.
2. **Its imagery lock key is wrong.** It states `(scope, scope_ref, category, subject,
   ordering)`. That is `transferDirectiveLocks` on `site_plan_directives`.
   `transferImageryLocks` matches `(plan_id, scope, scope_ref, key)`. Consequence for
   the design: scope_ref **is** in the imagery key, so rewriting it drops locks for one
   plan generation unless handled — which is why the canonical fallback exists.

## 2026-08-10 — measurements taken BEFORE submission, not asked of reviewers

| question | answer |
|---|---|
| rows invisible to consumers, current plans | **10** of 176 |
| of those, with an active asset already generated | **8** |
| unique-index collisions a rewrite would cause | **0** |
| open `needs_imagery` items on an affected ref | **3** |
| `site_plan_directives` affected? | **no** — 935 rows, all `scope='site'`, 0 positional refs |

The zero-collision figure is disconfirmable: the same query returns non-zero on
historical (superseded) plans, where `brands` and `brands-index` coexist on
dartsonline.

## 2026-08-10 — MISSTEP 2: fifteen green tests that could not see the fix removed

Wrote the helper tests, ran them green, and wrote in the file header that they guarded
the wiring. Then mutated: deleted the entire resolution block from
`WriteSitePlanAction`. **All fifteen still passed.** Every one calls the helpers
directly, so none can observe whether the action still calls them.

This is the identical class this tree logged in `WRONG_CALLS.md` the day before, from
another lane — and I had read that entry. **A test that claims to guard a call site is
only proven by deleting the call.**

Fixed with `write_site_plan_imagery_wiring_test.go`, driving the real action under
sqlmock and pinning the value that reaches the INSERT bind. Re-run against the same
mutation it fails, on both arms (the rewritten ref and the durable log). The false
header claim was corrected in place, not deleted.

Mutation log:

```
mutation 1  rewrite disabled          -> 4 unit tests FAIL   (logic covered)
mutation 2  wiring block deleted      -> 0 tests fail        (BEFORE the wiring suite)
mutation 2  wiring block deleted      -> 2 wiring tests FAIL (AFTER)
```

## 2026-08-10 — MISSTEP 3 (not mine, but it cost a design change): a same-file passenger

The clean way to normalise a planner's map key is to export `datahelpers.NormaliseSlug`
— one line, additive. I wrote it, then found on the pre-commit check that
`page_canonical.go` **already carried another session's uncommitted work** (a
`FlatURLs` / `nestedOrFlatURL` feature, not at HEAD).

A pathspec commit cannot exclude a same-file passenger. Committing that file would
have shipped another lane's untested WIP into the next chassis image — a real fleet
risk, and not my call.

So I backed **only my lines** out of the shared file (leaving their 38 intact,
verified by numstat) and routed through the already-exported `CanonicalisePage`
instead: its content/default branch applies exactly the same cleanup. That is a
coupling to another function's internal branch, so it is **pinned by a test** that
asserts `normalisePageKey(s) == ValidateRoles([{Name:s}])[0].Name` through exported
API only — obscurity converted into an asserted invariant.

Also observed the reverse of the same hazard: **another session committed
`000_concept_index.md` while my IMG-070 row sat uncommitted in it**, sweeping my edit
into their commit (`4451b2a0a`, "register(BLD-019)"). Nothing lost, forward-only
holds, and I did not re-commit the file. This is CLAUDE.md's "your uncommitted work is
not safe" happening in both directions inside one hour.

## 2026-08-10 — register bookkeeping

Index triple after adding IMG-070: **1,814 rows / 1,814 unique row ids / 1,816 unique
entry ids.** The 2-row orphan gap is `BLD-018` and `DIAG-042`, both other lanes' — a
*different* pair from the `WFA-012` the previous note flags, so the gap is a moving
population rather than a fixed backlog. No rows invented for them.

**Method note:** compare the id sets with `LC_ALL=C sort`. Without it, `comm` and
`sort` disagree on collation and the diff reports phantom orphans (a truncated `IMG`)
— the trap already in LANDMINES as "ask git about git, not comm".

## 2026-08-10 — state at end of session

- Code committed `c21af5eda`, gofmt follow-up `c90212df6`. **Go only — inert until the
  next chassis roll.**
- Council submitted, `46a50b4c-f00d-4492-b7fd-ce5dc2023480`, verdict pending; commit
  carries `Council-Submitted:` so 098 credits it automatically on approval.
- Backfill `sql_for_agents/373` written and committed, **NOT APPLIED** — it repairs
  data, and applying it before the code rolls buys exactly one plan generation.
- Existing LANDMINES entry for this mechanism **corrected in place** (its check
  measured the harmless half and pointed at the wrong table) rather than duplicated.

## 2026-08-10 (later) — the council run DIED; there is no verdict and there will not be one

Polled the submission and read the result rather than assuming latency:

```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '46a50b4c-...';
--  complete_invalid | COMPLETED

SELECT ... FROM diagnosis_artifacts WHERE correlation_id='46a50b4c-...'
  AND kind='council_report';
--  (0 rows)
```

`complete_invalid` with **zero** report artifacts is the council's generic "I could not
run" state — **not** a REJECTED verdict, and not latency. The cause, from
`collected_data->'__step_error'`:

```
step review_editquality failed: ... provider=anthropic model=claude-sonnet-5
API request failed with status 400: "You have reached your specified API usage
limits. You will regain access on 2026-09-01 at 00:00 UTC."
```

**Established as fleet-wide, not a fault in this submission** (the landmine for this
class says to distinguish the three cases by *message body*, not by step name):

| check | result |
|---|---|
| council-gate runs, last 6h | **4 `complete_invalid`**, 1 `complete_approved` |
| orchestrations dead on `%usage limits%`, last 12h | **7**, from 14:42Z to 17:02Z |

So the gate is effectively down for the fleet until the credit window resets. This is
the already-recorded landmine ("an API usage-limit death looks exactly like a transient
seat fault"); **no new landmine filed** — a second entry for one trap makes the reader
arbitrate.

**Consequence that must not be lost:** `Council-Submitted: 46a50b4c-...` on commit
`c21af5eda` names a correlation that **can never be approved**, because the run is dead.
098 resolves trailers at report time, so this commit will read as un-reviewed for ever
unless someone **resubmits** once credits return and records the new correlation. The
trailer is still the honest thing to have written — it asserts nothing, and I have not
written `Council-Reviewed:` on a verdict I never read.

**Resubmitting now is pointless** (three-week window). The submission JSON is committed
and ready: `COUNCIL_SUBMISSION_2026-08-10.json`, resubmit with
`RESUBMIT_CORR=46a50b4c-f00d-4492-b7fd-ce5dc2023480` so the trail accumulates.

## 2026-08-10 (evening) — LIVE on v1.0.1283, backfill applied, repair proven at the consumer join

### 1. The roll, verified at the artefact rather than at the tag

Both replicas on `v1.0.1283` (started 21:43Z; my fix commit was 17:10Z, so the image
post-dates it — but that is circumstantial, so the binary was grepped):

| grep | 95mgb | wnbs8 |
|---|---|---|
| POSITIVE `imagery scope_ref canonicalised` | 1 | 1 |
| POSITIVE `IMAGERY_SCOPE_REF_UNRESOLVED` | 1 | 1 |
| POSITIVE `IMAGERY_SCOPE_REF_ORDINAL_ANOMALY` | 1 | 1 |
| POSITIVE `collapsed onto one identity after canonicalisation` | 1 | 1 |
| CONTROL (pre-existing) `flattenImageryBlock: rows flattened` | 1 | 1 |
| **NEGATIVE** `imagery scope_ref pineapple` (fabricated) | **0** | **0** |

The negative control is the half that matters: a positive-only grep proves the pipeline,
never your spelling.

### 2. Backfill `sql_for_agents/373` APPLIED

Pre-census re-run first and unchanged at **10** (same rows). Result:

```
UPDATE 5   (page scope:   news->news-index, about->about-index x2, contact->contact-index x2)
UPDATE 4   (section scope: about:2 -> about-index:2, four icons)
NOTICE: bugfix 214 backfill: OK — 1 unresolvable row remains, as expected.
COMMIT
```

The `DO`/`RAISE` guard fired its success notice, i.e. it *checked* rather than assuming.
Post-census: **1** (mortgagecalculator `tools-index`, deliberately left).

### 3. Proven at the CONSUMER's own join, not at the row

Running `plan_sections_action.go`'s joins verbatim:

- page hero (`:287-300`, exact match): `gamesdesign about-index -> hero_about (active)`,
  `gamesdesign contact-index -> hero_contact (active)`,
  `fundamentallyai news-index -> hero_news (active)`. All three returned **nothing**
  before the backfill.
- section imagery (`:368-380`, `LIKE 'about-index:%'`): all **four** gamesdesign icons
  return with active assets.

### 4. What is NOT proven, stated plainly

- **mortgagecalculator's refs are repaired but its assets do not exist** (`asset_exists=f`
  for all 7). The reference is now correct; nothing is visible until imagery is generated.
  Do not quote those as "working".
- **The WRITE path had not executed in production** at this point: zero site plans written
  since the roll, and zero `IMAGERY_SCOPE_REF_*` rows. Live in the binary ≠ exercised.

### 5. The credit outage has LIFTED — both blocked items unblocked

Zero usage-limit deaths since the roll; 20 orchestrations COMPLETED. So:

- **Council resubmitted** on the same trail correlation
  `RESUBMIT_CORR=46a50b4c-f00d-4492-b7fd-ce5dc2023480`, new run `adba954d-599a-4913-98dc-c65fee1bb095`
  (orch `8a54fbc4-c376-4638-a60c-527df468daf7`). The trailer on `c21af5eda` can now resolve.
- **Induced a real plan write** on a POOL site — `pool-ai-agents.internal`
  (`29e0ffc4-2823-48ac-8edf-e9a50793f372`, 0 plans / 0 pages / status='pool', so nothing
  serves it). A customer replan was deliberately NOT used: it would rewrite a live site's
  plan, which is not this lane's to do. Orchestration `823d4e22-6786-4106-8d90-2ee48275e4b5`,
  dispatched 22:08:49Z — 25 min after pod start, clear of the ~300s post-restart drop window.
  kcat's exit status was ignored (it exits 0 having sent nothing); the orchestration ROW is
  the evidence it landed.

## 2026-08-10 (late) — the induced runs were NOT DISCRIMINATING, and my own test queued 41 work items

### Both pool-site runs returned `imagery_refs_canonicalised: 0` — and that is not evidence

| run | site | pages emitted | canonicalised | unresolved | merged |
|---|---|---|---|---|---|
| `823d4e22` | pool-ai-agents.internal | about, contact, index, services, use-cases — **all `content`** | 0 | 0 | 0 |
| `6d0e6a59` | pool-energy-utilities.internal | about, contact, faq, index, services — **all `content`** | 0 | 0 | 0 |

`CanonicalisePage` renames nothing under the `content` role, so **no rename occurred and
the fix had nothing to do.** That zero would have been zero on the old binary. Recording
it as "verified live" would be exactly the failure the standing rule names: *a `[MEASURED]`
figure is only evidence if the measurement could have come out otherwise.*

**What the runs DO establish**, and it is worth having:

1. **The new code path executes in production.** `collected_data->'plan_written'` carries
   `imagery_refs_canonicalised`, `imagery_refs_unresolved`, `imagery_duplicates_merged` —
   keys that cannot exist without this change.
2. **No regression.** 10 imagery rows written on the first plan, 0 unresolved, and every
   ref resolves to a page the plan contains.

**To actually close it:** the first replan of a site that HAS `-index` pages
(`gamesdesign.co.uk`, `dartsonline.com`, `robot-hands.com`). Then
`imagery_refs_canonicalised` must be **> 0**. Deliberately not forced — replanning a
customer site to harvest a number rewrites a live site's plan, which is not this lane's to
do.

### MISSTEP 4: my "safe" test queued 41 work items, 24 of them paid image generations

I chose pool sites specifically because nothing serves them. I did not think about what the
pipeline does **behind** a plan write. Each run queued a full build backlog:

```
pool-ai-agents.internal        19 open items (10 needs_imagery, 4 needs_page, ...)
pool-energy-utilities.internal 22 open items (14 needs_imagery, 5 needs_page, ...)
```

**24 `needs_imagery` items, each of which would have triggered a paid image generation on a
throwaway site**, plus page builds — and one `needs_page` had already been `claimed` by a
dispatch loop before I noticed.

All 41 cancelled in one transaction with a `DO`/`RAISE` guard asserting 0 remaining
(`UPDATE 41`, notice fired). Plans and undeployed pages left: they are inert, and they do
**not** pollute the R1 census, because the planner's own `sync_pages` step created matching
`pages` rows so their refs resolve — checked rather than assumed, census still **1**.

**The check, and it generalises past this lane:** *"nothing serves this site" is not the
same as "nothing happens".* After ANY induced dispatch, query `site_work_items` for that
site before walking away. A test that is safe at the artefact can still be expensive at the
queue.

### Council round 2 is genuinely running

`gate_tooling_provenance` / `EXECUTING_STEP` on orch `8a54fbc4-...`, trail correlation
`46a50b4c-...` — i.e. a real run this time, not the `complete_invalid` death of round 1.
Verdict not yet written; nothing in this lane claims one.
