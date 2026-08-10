# HANDOFF — bugfix 210 (needs_logo slug) — START HERE

**Written 2026-08-09 evening for a cold start.** The lane is coherent and committed at every
point; nothing is half-done in the tree. Read this, then `PLAN`, `NOTES` and `README_where_we_are`
in that directory if you need the why.

**Resolve this bug BY SLUG, never by number** — `210` also names
`…_a_content_failed_page_build_is_stamped_deployed…`, a different case owned by a different lane.

---

## STEP 0 — has it already landed?

Two things are pending on a chassis roll. **Check before doing anything**, because another
session's build ships them and a roll is not evidence your code is in it (`bugs_open/153`):

```bash
POD=$(kubectl get pods -n ai-persona-system -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')

# (a) IMG-069's refusal — ALREADY LIVE as of 2026-08-09 (verified: 1 / 0 / 1)
kubectl exec -n ai-persona-system $POD -- sh -c \
 'strings /app/agent-chassis | grep -c "generate_image REFUSED: no prompt supplied by any caller"'

# (b) The owner's DEFAULT — NOT live yet at time of writing. This is the marker:
kubectl exec -n ai-persona-system $POD -- sh -c \
 'strings /app/agent-chassis | grep -c "no planned prompt, using the brand-identity default"'

# NEGATIVE control — must be 0, or your grep is matching anything:
kubectl exec -n ai-persona-system $POD -- sh -c \
 'strings /app/agent-chassis | grep -c "using the brand-identity defaultXYZ"'
```

Run (b) on **every** replica, not one.

## Where things stand

**The bug:** `needs_logo` / `needs_hero_image` items were filed without the prompt their handler
requires (`image-build-handler`'s `call_logo_gen` maps `prompt` from
`input_data.spec.image_prompts.logo` as a **required** field), so they died at input extraction
and could never be handled. 2 of 2 promptless items failed; 4 of 4 with the key completed.

**Four defects were found, all fixed and committed:**

1. **The motivating item was a false positive.** `isPathReferencedInPages` matched an unanchored
   substring, so fundamentallyai's `<img src="https://leopardessconsulting.co.uk/assets/images/logo.png">`
   — a *partner's* logo on the partner's domain — read as this site serving its own placeholder.
   It was the only `logo.png` match in the whole fleet. Now anchored to same-origin.
2. **The check was blind to site chrome** (`site_components`), where a logo actually lives.
   `bugs_closed/128` fixed that in the flag-only sibling and left the *routing* check blind.
3. **Three producers share the promptless defect**, not one — including `WriteBuildItemsAction`
   on the primary build path, which had escaped by luck.
4. **The callee could not save itself**: exactly one step fleet-wide runs `generate_image` and it
   has no `prompt_template` at agent or step level, so the documented four-tier chain has two
   rungs — the caller's prompt, or a meaningless sentence.

**Two guards now stand:**

- **IMG-069 (LIVE):** `generate_image` **refuses** the generic fallback. No producer, present or
  future, can turn a missing prompt into a stored brand asset. Mutation-proven.
- **The default (NOT live — needs the next roll):** the owner's ruling, below.

## The owner's ruling, 2026-08-09 evening — this is the current design

> *"When a site needs a logo it should go to human review for guidance/supply of the logo, but
> for now that human review can default to saying create a logo that suits the mission, target
> market and the domain character. This is because there are 2000 odd domains to populate and I
> won't have time to do or approve that many logos."*

**It reverses what shipped that afternoon.** The afternoon's disposition filed a
`needs_human_review` item and stopped. Across ~2,000 domains that is a queue that never drains,
so the human path becomes an **override**, not a gate. `DefaultBrandImagePrompt`
(`discovery_checks/default_brand_prompt.go`) is called by both producers.

**The two things a successor most needs to not get wrong:**

- **Brand identity is NOT the imagery axis, and that is the whole safety argument.** The helper
  reads name/domain/sector/positioning/audience/tone. It must **never** reach for
  `design_intent.imagery_direction`, the imagery style guide, or `directionForKind` — logos are
  excluded from those *on purpose* (`imagery_style_guide.go`: "logos get nothing — the
  2026-05-20 contamination lesson"), because a photographic direction prepended to a flat logo
  prompt makes the model composite the mark onto a photo. If you find yourself adding imagery
  signal here, stop: that is the excluded path.
- **The helper must never return `""` for a site that exists.** The domain alone is enough
  (`robot-hands.com` → "a simple, distinctive logo mark for robot hands"). This is not
  defensiveness — it is what keeps IMG-069's refusal *unreachable* rather than routinely fired,
  and at 2,000 domains a bare-domain site is the **common** case (only 21 of 39 current sites
  carry an `identity` spec). `TestDefaultBrandPrompt_NeverEmptyForARealSite` pins it.

## What to do next, in order

1. **Wait for / check the second council verdict.** `SUBMISSION_CORR=661557c5-7ae4-43fe-a36d-c0600b54a29c`.
   ```sql
   SELECT current_step, status FROM orchestration_states
    WHERE collected_data->'input_data'->>'fix_correlation_id' = '661557c5-7ae4-43fe-a36d-c0600b54a29c';
   SELECT body FROM diagnosis_artifacts
    WHERE correlation_id='661557c5-7ae4-43fe-a36d-c0600b54a29c' AND kind='council_report'
    ORDER BY created_at DESC LIMIT 1;
   ```
   **READ it before writing any `Council-Reviewed:` trailer** — the commit already carries
   `Council-Submitted:`, which 098 upgrades automatically on approval, so there is nothing to do
   if it approves and nothing is falsely claimed if it does not. If REVISE, resubmit with
   `RESUBMIT_CORR=661557c5-…`.

2. **After the next roll: look at a defaulted logo with your own eyes.** This is the highest-value
   remaining task and no test can do it. That prompt string decides ~2,000 logos.
   ```sql
   SELECT s.domain, w.item_type, w.status, left(w.spec->>'prompt', 200)
     FROM site_work_items w JOIN sites s ON s.id=w.site_id
    WHERE w.spec->>'prompt_source' = 'default_from_brand_identity' ORDER BY w.updated_at DESC;
   -- then the artefact it produced:
   SELECT purpose, left(origin_prompt,160), url, created_at
     FROM assets WHERE site_id='<site>' AND purpose='logo' ORDER BY created_at DESC LIMIT 3;
   ```
   Tuning is a string in one function — cheap. **Trust the rendered artefact, not the status.**

3. **The two live cases to watch.** fundamentallyai's `needs_logo` (false positive — should stop
   being filed once the anchoring rolls; the existing `failed` row will need cancelling by hand
   with a reason) and mortgagecalculator's `needs_hero_image` (a **true** positive — 6 same-origin
   CSS `url('/assets/images/hero.jpg')` refs, no hero asset; this is the one that should now
   succeed end to end and is the best behavioural proof the lane has).

4. **First non-zero refusal count is a real measurement nobody has ever had.** Any
   `needs_logo`/`needs_hero_image` failing with `generate_image refused` names a producer bug.

## What is deliberately NOT done

- **No review surface for defaulted assets.** They are queryable via `spec.prompt_source`;
  nothing renders them. If the owner wants a list to skim, that is a follow-on and does not exist.
  Do not describe it as delivered.
- **The prompt-key convergence** (the two legacy handler branches read
  `spec.image_prompts.<key>`; the two Phase-2E branches read a flat `spec.prompt`). Converging
  them would make this defect class unrepresentable for future producers, but it changes what the
  shared handler guarantees its producers → architecture-scope under the 2026-07-29 ruling.
  Registered as IMG-069's open question. **All three producers now write BOTH key shapes, so the
  convergence is a config-only change whenever someone takes it up.**
- **No key-guessing ladder over the `identity` spec.** It has ~70 distinct top-level key
  spellings across sites (the `bugs_open/072` family), and the helper reads only the three
  dependable ones. A missing clause is cheaper than a wrong one. Deliberate, not an oversight.

## Traps this lane hit — do not repeat them

- **`git stash` is shared state on this tree.** A pathspec push silently creates *nothing* if any
  named path is untracked, and the bare `pop` then takes another branch's WIP. Use a
  `git archive` tree instead. (LANDMINES + WRONG_CALLS.)
- **Pin the sha for both arms of any before/after comparison.** `git archive HEAD` twice, minutes
  apart, gave me two *different* commits because another session committed in between — I spent
  time diagnosing their already-fixed failure as mine.
- **Parenthesise every `OR` inside a `WHERE` that also has an `AND`.** Precedence silently dropped
  my `deployed`/`unlocked` filters and returned a plausible wrong table.
- **The package's own tests catch things the council does not.** `handler_coverage_test.go` found
  that `human-review` is not a real agent before two council seats independently raised it.
- **The in-tree `go test ./platform/orchestration/actions/` may be red from another session's
  WIP.** Verify against `git archive <pinned sha>` + your files. Known pre-existing failure at
  HEAD: `TestEveryCheckProducedItemTypeIsClassified` (`decision_regression`, another lane's file)
  — not yours, do not fix it as if it were.

## Commits in this lane (newest last)

`9980e5158` lane docs · `9425615bb` the four-part fix + tests · `c91af6c10` landmines + wrong_calls ·
`d78d0b788` register IMG-069 · `eb0341315` bug file corrections · `6a927239a` notes + README ·
`ebaf72729` **the owner's ruling — default rather than block** · plus the register/bug-file updates
and the `Council-Submitted:` trailer commit.

---

# UPDATE 2026-08-10 — STEP 0 IS NOW ANSWERED: BOTH HALVES ARE LIVE

Everything above stands except the "not live yet" notes. **Do not re-run the deploy checks as
if they were open** — they were run on `agent-chassis-8496665bb8` (both replicas):
refusal literal **1**, default marker **1**, second producer's marker **2**, fabricated negative
control **0**.

**Council round 2 `661557c5-7ae4-43fe-a36d-c0600b54a29c` — APPROVED**, verdict read, trailer
committed. Two objections were answered with code/measurement (the silent-degrade fix; the
producer census: exactly three code producers, no fourth). **The lane's substance is complete.**

## The one thing left that matters

**Nobody has yet seen a logo or hero that the default actually produced.** That is the
outstanding behavioural proof and it is the highest-value thing a successor can do.

The case to watch is **mortgagecalculator.co.uk's hero** — a confirmed true positive (6
same-origin `url('/assets/images/hero.jpg')` references, no active hero asset). Its stale
pre-roll rows were cancelled on 2026-08-10 so that discovery can re-file it *with* the default
prompt. Discovery is demonstrably running (fundamentallyai drew an item at 12:33 that day).

```sql
-- 1. has it been re-filed, and does it carry a defaulted prompt?
SELECT s.domain, w.status, w.spec->>'prompt_source', left(w.spec->>'prompt',180), w.error
  FROM site_work_items w JOIN sites s ON s.id=w.site_id
 WHERE w.item_type IN ('needs_logo','needs_hero_image')
 ORDER BY w.updated_at DESC LIMIT 5;

-- 2. TRUST THE ARTEFACT, NOT THE STATUS — did an asset actually land, from what prompt?
SELECT purpose, left(origin_prompt,180), origin_model, created_at
  FROM assets WHERE site_id=(SELECT id FROM sites WHERE domain='mortgagecalculator.co.uk')
   AND purpose='hero' ORDER BY created_at DESC LIMIT 3;
```

Then **look at the image**. If the wording needs tuning it is one string in
`composeBrandImagePrompt` — cheap, and it decides ~2,000 logos.

If it has NOT been re-filed after a day or so, the question is whether the discovery rotation
covers that site, not whether the fix works — check `bugs_open/230`'s rotation before assuming
a defect here.

## What today added that was not in the plan

- **A gap the fix does not close, found by reading the live queue:** it repairs **filing**, not
  rows already filed. A pre-roll row has no `image_prompts` in its spec and no producer-side
  correctness repairs it. Four such rows were cancelled with evidence in `result` (identity
  verified, not counted). **Fleet census afterwards: 0 open rows** of any prompt-requiring type
  lacking a usable prompt. If you file a new producer, remember its old rows are not retrofitted.
- **A real defect the council caught in my own fix:** the default degraded *silently* — one
  `err == nil` made a query failure and an absent spec identical. Fixed, and it now emits a
  per-prompt `clauses` count so a fleet-wide degradation shows as a pattern.
- **Proof the afternoon disposition ran in production** before being overruled
  (mortgagecalculator's row filed at `needs_human_review`, 2026-08-09 20:56).

## Known-open, deliberately

1. **`spec.prompt_source` has no consumer** — four council seats flagged it. A per-item review
   queue is the wrong answer (it is what the owner ruled against); a **count in an existing
   report** is the right one. Unbuilt. Do not build a queue.
2. **The prompt-key convergence** — still architecture-scope, still registered as IMG-069's open
   question, all three producers already write both key shapes so it is config-only.
3. **No human eye on the prompt text yet.**

## One more trap, added today

**Backticks in `git commit -m` EXECUTE.** I hit this despite it being in LANDMINES *and* my own
memory index, and it silently ate two identifiers out of a commit message explaining a code
change. The trailer survived only because it was on its own undecorated line. **Write commit
messages to a file and use `-F`**, or single-quote identifiers. Logged in `WRONG_CALLS.md`.

---

# UPDATE 2026-08-10 (evening) — ALL THREE CHANGES LIVE. The lane is DONE bar one observation it cannot force.

**Pod-verified on both replicas** (`agent-chassis-696d88b4c7`), fabricated negative control **0**:

| marker | count |
|---|---|
| `generate_image REFUSED: no prompt supplied by any caller` (IMG-069) | 1 |
| `no planned prompt, using the brand-identity default` (owner's ruling) | 1 |
| `DefaultBrandImagePrompt: composed` (silent-degrade fix) | 1 |
| `identity spec read FAILED` (silent-degrade fix) | 1 |

Council rounds 1 and 2 both APPROVED, both verdicts read, trailers committed. **Nothing is
outstanding in code, config, tests or docs.**

## The one thing left, and why it cannot be forced from here

**Nobody has yet seen a logo or hero the default actually produced.** Measured this evening:

- work items with `spec->>'prompt_source' = 'default_from_brand_identity'`: **0**
- items whose error mentions `generate_image refused`: **0**

**Both zeros are UNDEMANDED, not evidence of anything.** `check_placeholder_image_in_use` has
not run for a qualifying site since the roll. The demand exists and is unchanged —
mortgagecalculator.co.uk still has **6** anchored same-origin `hero.jpg` references and **no**
active hero asset — so the check will file a defaulted item the next time discovery reaches that
site.

**Do NOT force it by dispatching at mortgagecalculator.** The owner has ruled that site is the
`mortgagecalculator_couk_adoption` lane's while it finishes its current plan (three sessions were
live on it on 08-10). Wait for the rotation, or find another site that meets the precondition —
there is none today.

**⚠ A demand control on this must not use `source='discovery'`.** I did, and it lied: the 8
"discovery" rows for that site on 08-10 were `acceptance_run` items another session filed with
that label. `source`/`created_by`/`pipeline` are free-text. Key a demand control on `item_type`
values only these checks emit, or on the check's own log line. (WRONG_CALLS, 08-10.)

**And a fact worth having, since it looks like a trap and is not:** the two cancelled
`placeholder_image_in_use:hero` rows do **not** arm the two-strike rule.
`insertWorkItem`'s strike query is `status IN ('complete','failed')`, so `cancelled` is excluded —
terminal for dedup, not a failed attempt. One of the rows I cancelled had been `failed`, so the
cleanup **removed** a strike. The re-file will arrive as a clean `detected`.

## When it does fire, this is the whole check

```sql
SELECT s.domain, w.status, w.spec->>'prompt_source', left(w.spec->>'prompt',180), w.error
  FROM site_work_items w JOIN sites s ON s.id=w.site_id
 WHERE w.item_type IN ('needs_logo','needs_hero_image') ORDER BY w.updated_at DESC LIMIT 5;

-- then the ARTEFACT, which is the only thing that settles it:
SELECT purpose, left(origin_prompt,180), origin_model, url, created_at
  FROM assets WHERE site_id=(SELECT id FROM sites WHERE domain='mortgagecalculator.co.uk')
   AND purpose='hero' ORDER BY created_at DESC LIMIT 3;
```

**Then look at the image.** If the wording needs tuning it is one string in
`composeBrandImagePrompt` — cheap, and it decides ~2,000 logos. That judgement is the owner's,
and it is the last open item on this lane.
