# NOTES — `bugs_open/100` + `101` (scrape config + provenance)

Append-only, newest at the bottom. Technical log: evidence, commands, what the system
actually said, and every misstep.

---

## §1 — 2026-07-28 ~17:30 BST — coverage check and re-grounding, before any code

**Coverage.** `who-owns.py` over every plausibly-free bug in `/bugs_open/`. 100 → "no
owning workstream identified". 101 → names `vetcomparison`, but that thread is the
blocked party (its own line: "crawl restart BLOCKED on a Go change") and has written no
fix; last touch 2026-07-27.

Open work items touching this area — **zero rows**:

```sql
SELECT id, item_type, status, left(summary,90), created_at FROM site_work_items
WHERE status NOT IN ('complete','cancelled','rejected')
  AND (summary ILIKE '%scrape%' OR summary ILIKE '%provenance%' OR summary ILIKE '%config key%'
       OR summary ILIKE '%data_observation%' OR summary ILIKE '%only_main%');
-- (0 rows)
```

**101 re-grounded — unchanged.** Both definitions still carry the inert keys:

```
            type            |      step      | max_pages | follow_links | extract_mode | fallback_url
----------------------------+----------------+-----------+--------------+--------------+--------------
 domain-research-classifier | scrape_site    | t         | t            | t            | f
 vet-practice-verifier      | scrape_website | t         | t            | t            | t
```

**100 re-grounded — unchanged.** Still 2,970 rows, still zero provenance, still no key
in `raw_data`, collection still off since March:

```
 total | has_source_url | llm_claimed |           newest
-------+----------------+-------------+----------------------------
  2970 |              0 |           0 | 2026-03-18 22:09:03.579088
```

Note `llm_claimed = 0` is doing real work here: it is the *negative control* for the fix
(§How to verify in 100). It must still read 0 afterwards.

## §2 — the `[UNSETTLED]` box is settled, and the answer is a second bug

101 refuses to let candidate 2 proceed until this is answered: does production's
Firecrawl path strip nav/footers? If yes, more page fetches deliver nothing, because
company numbers live in footers.

Read the code rather than inferring from stored samples (the file's own evidence was
ambiguous — 75% of 2,452 samples retain footer nav text, but 0 contain
company-registration text, and those two facts do not separate the hypotheses).

`internal/adapters/webscrape/providers/firecrawl.go:77-111`:

```go
onlyMainContent := false
if mainContent, ok := config["only_main_content"].(bool); ok {
    onlyMainContent = mainContent
}
...
if onlyMainContent {                       // <-- only ever emits TRUE
    payload["onlyMainContent"] = onlyMainContent
}
```

So the key is emitted **only when true**. `false` is indistinguishable from unset, and
unset means Firecrawl's own default applies.

**The external premise was verified, not assumed** — Firecrawl `/scrape` API reference:
*"default: true"*, and it *"excludes headers, navs, footers"*. So a caller explicitly
asking for the full page gets main-content-only: the exact opposite.

The `/crawl` path in the **same file** (line 338) is correct:

```go
if onlyMain, ok := config["only_main_content"].(bool); ok {
    scrapeOptions["onlyMainContent"] = onlyMain
}
```

Two paths, one file, opposite semantics for the same key. This is 101's own class —
config that reads as live and is not — in a provider every scrape on the fleet goes
through.

**Consequence for 101's measured numbers:** the 22%→30% company-number figure was
measured by a **read-only probe fetching raw HTML**, not through production's `/scrape`.
Production has been receiving main-content-only pages, so the probe's 30% was never
reachable by adding page fetches alone. `[UNVERIFIED]` — how much of the gap this
accounts for is not measured here and should not be asserted; what is established is
that the probe and production were not fetching the same thing.

## §3 — the survey that sized the framework fix (and nearly oversized it)

101's candidate 1 is "reject unknown config keys for registered actions". Before
designing it, measured the surface it would have to be right about:

```sql
WITH steps AS (
  SELECT ad.type, e.k AS step, v->>'action' AS action, v->'config' AS cfg
  FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') AS e(k,v)
  WHERE ad.deleted_at IS NULL AND COALESCE(ad.is_snapshot,false)=false AND ad.is_active
    AND v->'config' IS NOT NULL AND jsonb_typeof(v->'config')='object'
)
SELECT count(*) AS total_steps_with_config, count(DISTINCT action) AS distinct_actions,
       (SELECT count(*) FROM (SELECT DISTINCT s.action, ck.key FROM steps s, jsonb_object_keys(s.cfg) AS ck(key)) x)
FROM steps;
--  1155 | 228 | 811
```

**811 distinct (action, key) pairs over 228 actions.** This killed the version of the
plan I was about to write, which would have hard-rejected unknown keys fleet-wide: a
declaration that wrong at that scale would reject live definitions, and an over-strict
validator is a considerably worse bug than the one being fixed (the same shape as 101's
own recorded trap — a fleet-wide `WHERE config ? 'max_pages'` cleanup would strip a live
page cap off `build-site-planner`). Opt-in per action it is.

**Reuse check before building.** `datahelpers.RegisterActionInputSpec` already exists and
134 files call it. `GetActionInputSpec` has **no callers at all** — grep returns only its
own definition. The registry is populated and read by nothing but `registry_parity_test.go`.
So the machinery to extend is already there and currently inert; building a second one
would have been the drift class this repo's council reviews for.

## §4 — 2026-07-28 ~18:00 BST — the audit found a FIFTH inert key on its first run

Not from reading the bug file. `scripts/audit-config-keys.sh`, run against the live
fleet immediately after declaring `scrape_web`'s contract, reported:

```
=== UNKNOWN KEYS (action declared its contract; these are not in it) ===
  scrape_web: add_protocol
```

`bugs_open/101` names **four** keys. This is a fifth, and it is a near-miss typo
rather than an aspiration — which is why no amount of re-reading the file would
have surfaced it:

```
$ grep -rn --include=*.go "add_protocol" . | grep -v '^./docs/'
platform/orchestration/actions/webscrape_actions.go:509:  if addProtocol, ok := config["add_protocol_if_missing"].(bool); ok && addProtocol {
```

The code reads `add_protocol_if_missing`; the definition writes `add_protocol`;
and line 509 belongs to a **different action** (the URL-validation action), not to
`scrape_web`. Live:

```
            type            |    step     | add_protocol
----------------------------+-------------+--------------
 domain-research-classifier | scrape_site | true
```

A bare domain reaching the adapter is a failed fetch, so this was not cosmetic.
Implemented (fires only when explicitly true AND the URL has no scheme).

**What this says about the fix, and it is the part worth keeping:** the detector
earned its place in one run, on a key class the human process structurally could
not catch. Four keys were found by a careful reader; the fifth needed a machine
that compares what the config says against what the binary declares. The value is
not the four keys — it is that the next typo surfaces without anyone going
looking.

## §5 — the shape I could not observe, and what I did instead

`ExtractFetchProvenance` needs to know what `collected_data["scraped_data"]`
actually looks like. Tried to sample it:

```sql
SELECT ... FROM orchestration_states WHERE collected_data ? 'scraped_data' ...;
-- (0 rows)
```

**Nothing survives.** Vet collection has been off since 2026-03-18 and
`orchestration_states` is on a retention clock (the recorded landmine: *every
history table is on a retention clock — record a RATE not a count*). So the shape
could not be measured, and guessing it would decide whether every future
observation is sourced or silently not.

Traced it through the code instead, which is weaker than an observation and
stronger than a guess:

```
adapter sendSuccessResponse    → body: {success, body:{data: <provider result>}, …}
types.ResponseBody.Body        → {data: <provider result>}
coordinator.parseResponseBody  → CleanDataMap of that
applyResponseToState           → collected_data[output_field] = {data:{url, captured_at, …}}
```

So the live path is **`data.url`**. Recorded as such in the code, with the
derivation, and marked `[UNVERIFIED]` for which OTHER shapes occur in practice —
the reader accepts six because the chain has several unwrap points
(`output_mapping` short-circuits it entirely) and being wrong here means silently
storing an unsourced row.

## §6 — what was committed, and what is still owed

- `2ebabf2ca` — the fix (14 files). **INERT until two images roll**: the
  web-scrape-adapter for `only_main_content`, the chassis for everything else.
- `70885daf0` — a declared gofmt sweep plus a doc-comment correction. The commit
  hook caught that `business_intel_actions.go` was not gofmt-clean; I had
  deliberately left that pre-existing drift alone to keep a reformat out of a
  bugfix diff, and the hook's note that **the build gate REJECTS un-gofmt'd code**
  made that the wrong call — it would have failed the build for everyone.
  Swept separately, said so in the message.
  > **MISSTEP, recorded:** "leave pre-existing drift alone" is right for review
  > hygiene and wrong when the drift is in a file you are committing and the gate
  > is a hard one. The check that would have caught it in seconds is
  > `gofmt -l <the files you are about to commit>` — not `gofmt -l <package>`,
  > which is what I ran, and which buried my file in ten others I had not touched.
- Council submitted, `SUBMISSION_CORR=f4cf0aab-5a08-4475-91ea-fa831cff323c`.

Still owed: the council verdict; both images; SQL 257 applied AFTER the chassis
image (order load-bearing — before it, the constraint refuses writes the running
binary cannot yet satisfy).

## §7 — 2026-07-28 ~18:20 BST — the council round, and four missteps of my own

Verdict **REVISE**, 11 reviewers, 7 approve / 4 object, `decided_by = "gating
objection from tooling_provenance"`. Checked `decided_by` and `unreadable` FIRST
per the recorded landmine (`bugs_open/119`: ~11% of rounds are decided by one
seat's unparseable JSON). `unreadable: 1` exists but is **not** the decider — a
named seat with a substantive objection is. **This is a real REVISE, not the
harness.**

**A caveat on the reviewers' evidence, which does not rescue me.** Their read-only
verification queries returned **0 for facts that are true**:

```
steps_with_only_main_content_false        -> 0     (actual: 4 rows; 3 on the /scrape path)
distinct_actions | steps_with_config      -> 0 | 0 (actual: 228 | 1155)
domain-research-classifier add_protocol   -> (0 rows)  (actual: 1 row)
```

Re-ran all three myself: my figures reproduce exactly. So their *checks* were
broken or mis-scoped. **That is not a defence** — every objection below stands on
its own reasoning and none of them depended on those queries.

---

### Misstep 1 — I left gofmt drift in a file I was committing, on a rule that does not apply here

I found pre-existing gofmt drift in `business_intel_actions.go` (a hand-aligned map
literal I never touched) and deliberately left it, to keep a reformat out of a
bugfix diff. Defensible as review hygiene. Wrong here: the commit hook pointed out
that **the build gate REJECTS un-gofmt'd code**, so my tidy diff would have failed
the build for every session, not just looked untidy.

> **The check:** `gofmt -l <the exact files you are about to commit>` — **not**
> `gofmt -l <package>`, which is what I ran. The package-level run buried my file
> among ten others I had not touched, and that noise is precisely why I read it as
> "pre-existing, not mine, leave it".

### Misstep 2 — a load-bearing ABSENCE claim, asserted in the same voice as the grounded ones

`prior_art_librarian` caught this and was right to. My submission's whole
reuse argument rested on *"`GetActionInputSpec` currently has no callers at all"* —
and that claim was **not** in `grounded_in`, while every other factual claim in the
submission carried a quoted file:line or a live query. An absence claim is the one
that most needs showing, and it is the one I asserted.

It happens to be **true**. Verified properly, after the fact, at the pre-change commit:

```
$ git grep -n "GetActionInputSpec" 2ebabf2ca^ -- '*.go'
2ebabf2ca^:platform/orchestration/datahelpers/action_inputs.go:370:// GetActionInputSpec retrieves a registered spec
2ebabf2ca^:platform/orchestration/datahelpers/action_inputs.go:371:func GetActionInputSpec(actionName string) (ActionInputSpec, bool) {
```

Only its own definition and doc comment. **But "it turned out true" is not the
point** — I had no right to the confidence at the time, and the recorded landmine
says exactly this: *an absence is true only when you looked*. Note also that the
claim is now **self-referentially false**: re-run today it returns three callers,
all of them mine. Any later reader re-checking it will "disprove" it unless the
ref is pinned — which is the `relative-git-refs-are-not-evidence` trap wearing a
different hat, and the reason the grep above names `2ebabf2ca^` and not `HEAD~n`.

### Misstep 3 — I built a detector and then silenced it on the case that motivated it

`editquality`'s objection, and the one I find hardest to argue with. By declaring
`max_pages` and `follow_links` in `scrape_web`'s `ConfigKeys`, I made them
**recognised** — so the validator no longer reports them, and my own audit prints:

```
=== UNKNOWN KEYS (action declared its contract; these are not in it) ===
  none
```

Two live steps still describe a three-page crawl that does not happen. The lie is
intact; it is now merely invisible **to the tool I built to catch exactly this**,
surviving only as a runtime warning at `buildScrapeConfig` time.

The irony is instructive and I want it written down: I added a `checked` bool to
`UnknownConfigKeys` specifically so that "clean" could never be confused with
"never examined" — and then produced a third state I had not modelled,
**"declared, and honoured only under a condition that does not hold here"**, which
reports as clean. Worse, I wrote "it now reports zero unknown keys" into the commit
message and the register entry as though it were a result. It is not; it is partly
an artefact of my own declaration.

> **The check:** when you add a suppression mechanism, run the detector against the
> ORIGINAL failing case afterwards and confirm it still fires. If declaring a thing
> makes the detector quiet about it, the declaration needs a third state, not a
> binary. Fix owed: a `ConditionalKeys` notion (key → the condition under which it
> takes effect) so the audit reports them in their own section instead of as clean.

### Misstep 4 — I reused the existing registry and then built a parallel doc trail

`tooling_provenance`, the gating objection. The platform has a travelling-docs
mechanism — `doc_plans` / `doc_notes`, keyed by `subject_type` + `subject_key`,
with an `append_doc_note` action — whose stated rule is to load a subject's prior
decisions before changing it and leave a NOTES entry after, so the next fix builds
on this one instead of re-deriving lost context. I touched four subjects
(`firecrawl` Scrape, `scrape_web`, the workflow validator, the business-intel
writer) and read/wrote **none** of it.

What I did instead was good archaeology — grepping code, sampling the live DB,
reading `bugs_open/*` and `WRONG_CALLS.md` — and then recorded it all in a
**parallel, self-built trail**: this directory, the bug files, the commit messages.

The sting: my own submission argued, at length and correctly, that extending the
existing inert spec registry beat building a second one, *"which is exactly the
drift class this repo's council reviews for"*. I made that argument about Go code
while committing the identical error about documentation, in the same submission.
The `add_protocol` find and the `onlyMainContent` inversion are precisely the
hard-won context that belongs against those subjects, and right now they live only
where someone already reading this workstream would find them.

> **The check:** before touching a subject, ask whether the platform already has a
> place for what you are about to learn about it. "Reuse before building" is not a
> rule about code.

### Not a misstep, but pinned so nobody thinks either of us miscounted

`only_main_content: false` appears on **4** live steps, not 3. The fourth is
`site-adoption-agent/crawl_site`, which is `firecrawl_crawl` and therefore goes
through the `/crawl` path that was **always correct**. So: 4 steps ask for it,
**3 were getting the opposite**. Both numbers are right about different questions.

## §8 — 2026-07-28 ~18:30 BST — both images live, verified by CREATED symbols; 257 applied and proven

Owner rolled `v1.0.1192` (chassis + adapters together). Verified against the RUNNING
pods. **Every marker is a symbol this change created**, with a pre-existing string as
positive control in the same command — because the fleet index warns *a retag is not
a rebuild* (1188/1189 shared one image id, built 56 min before the fix in it), so a
tag comparison would have proved nothing:

```
chassis  unrecognised_keys 1 (was 0) · "does not read" 1 (0) · add_protocol 3 (0)
         "no fetch provenance available" 1 (0) · CONTROL scrape_web 1 (1)
adapter  buildScrapePayload 2 (created by this fix; 0 before) · CONTROL onlyMainContent 1
```

The adapter needed a different marker class: my firecrawl change added **no new string
literal** (it turned `if v` into `if ok`), so there was nothing textual to grep. The
extracted **function name** `buildScrapePayload` is in the binary and did not exist
before — a created-symbol marker obtained as a side effect of making the payload
testable. Worth remembering: *extracting a pure function to make a change assertable
also gives you a pod-grep marker for a change that otherwise has none.*

**SQL 257 applied at 18:30Z, after the pod-grep** — the sequencing this file called
load-bearing, and `debug_historian`'s round-1 objection was exactly that the plan never
said how "the chassis is live" gets confirmed. It is confirmed by the block above.

Enforcement proven, not assumed:
```
 conname                                | convalidated
 data_observations_provenance_not_empty | f            <- NOT VALID, as designed

 INSERT (NULL, '', '{}') -> ERROR: violates check constraint ...
 total 2970 | still_unsourced 2970      <- history untouched, deliberately
```
The negative control is the whole point: without it, "quiet" and "not enforcing" are
indistinguishable.

**062 payload watch:** 0 `Message Size Too Large`, 0 `Failed to produce` in 3h.
`[UNMEASURED]` whether the three affected steps have actually run since the roll — a
clean log is weak evidence until one has. Re-run it.

**The validator warning has not fired, and that is correct, not a null result.**
`scrape_web` now declares every key it reads, so it has nothing to report; no other
action has opted in. The live proof of the detector remains the audit script plus the
unit tests — which is itself misstep 3 restated: my own declaration is why the runtime
check is quiet.

**Ticket decisions.** `101` → `/bugs_closed/`: the defect it names is "silently
ignores", and the silence is gone, live and verified. `100` stays **OPEN** despite its
fix being live, because its own §"How to verify a fix" requires a two-column check
(`source_url` non-empty AND `raw_data ? 'source_url'` still false) that has never run —
collection has been off since March. Closing on a pod-grep would assert a
discriminating check passed when it was never executed. The blocker for `vetcomparison`
is nevertheless lifted, and that is written into the bug file for them.

### Misstep 5 — a pathspec commit half-completed the rename

`git mv` staged both sides of 101's move, but my commit pathspec named only
`bugs_closed/101…` and not `bugs_open/101…`. The add side committed, the delete side
stayed staged, and **101 existed in BOTH directories at HEAD** — exactly the ambiguity
the open/closed split exists to prevent, since `/bugs_open/` is meant to answer "what
is biting prod right now".

Caught by re-reading `git status` after committing, not by any hook. The cause is the
same property that makes pathspec commits *safe*: they silently omit whatever the
pathspec does not name. Safety against other sessions' staged work and silent omission
of your own are one behaviour.

> **The check:** after `git mv` + a pathspec commit, name **both** paths, then confirm
> with `git ls-files bugs_open/ | grep -c <n>` and the same for `bugs_closed/` —
> expect `0` and `1`. Fixed in `7de1757fa`.

## §9 — 2026-07-28 ~18:50 BST — round 2: the third state, and the notes I should have written first

Two substantive changes, both answering round-1 objections, plus grounding for the
other two seats.

**The gating objection cost the least and taught the most.** `tooling_provenance`
was right that the archaeology went into a parallel trail. SQL `258` applied three
`doc_notes` (`subject_type='action'`, keyed by action name, matching the existing
convention — migration-authored notes are normal here, cf. `created_by`
`212_tool_crosslink_action_notes`):

```
 subject_type |         subject_key         | cats | body_len
 action       | firecrawl_scrape            |    6 |     2056
 action       | scrape_web                  |    5 |     1789
 action       | store_business_verification |    5 |     2600
```

Each records the **non-obvious** thing, not a summary: the `add_protocol` typo and
which action actually reads it; the `onlyMainContent` presence-vs-truthiness
inversion and why two paths in one file disagreed; and on the writer, why asking the
model for `source_url` is the *rejected* candidate rather than the obvious fix.

**`ConditionalKeys` — the state I had not modelled.** The audit now prints:

```
=== UNKNOWN KEYS (action declared its contract; these are not in it) ===
  none
  ^ read this WITH the next section — 'none' here does NOT mean
    'no step misdescribes itself'.

=== CONDITIONALLY HONOURED (declared, so not unknown — but may not apply) ===
  scrape_web.follow_links: only on a crawl — set the step's config action to "crawl", …
  scrape_web.max_pages:    only on a crawl — set the step's config action to "crawl", …
```

Design notes worth keeping:
- **The condition is prose on purpose.** It is not generically evaluable — whether
  `max_pages` applies depends on the action's own dispatch, decided inside the
  action, which the validator cannot see. Putting enforcement there would be a guess
  dressed as a gate. Enforcement stays at execution (`buildScrapeConfig` already
  warns); this field exists to stop the **offline** report claiming a clean bill.
- **`UnknownConfigKeys` stays deliberately silent about conditional keys.** They are
  not unknown. Merging the two reports would recreate the same blindness from the
  other direction, and a test asserts a key can never be both.
- **The weak point, stated rather than hidden:** free prose can go stale if the
  action's dispatch changes and nobody updates the string.
  `UndeclaredConditionalKeys` only catches the *structural* error (a conditional key
  missing from `ConfigKeys`), never a wrong description. Flagged to the council as
  possibly too weak to be worth having.

**This round needs NO roll** — `ConditionalKeys` feeds only the audit path, which
runs from source via `go run`. Chassis behaviour is unchanged. That is a property
worth noticing rather than a lucky break: *the round-1 defect was reported by a tool,
so the repair belonged in the tool.*

**Round 2 submitted on the SAME correlation** (`RESUBMIT_CORR=f4cf0aab…`) so the
trail accumulates. I asked the council to treat one thing sceptically: round 1's
reviewer verification queries returned **0 for facts I have twice measured as true**.
If the same checks return zero again, that is the harness and not evidence — but it
is raised as a caveat, not a defence, because no round-1 objection depended on them.

## §10 — 2026-07-28 ~19:00 — the 062 watch was uninformative, and I nearly reported it as clean

Re-ran the payload-size watch over a 6h window: **0** `Message Size Too Large`, **0**
`Failed to produce`, 0 errors of any kind. I had already written "0/0" into the bug
file and the handoff as a post-roll result.

Then asked the question that decides what those zeros mean:

```
$ kubectl logs deploy/web-scrape-adapter --since=6h | grep -c "Starting scrape"
0
```

**No scrape has run at all since the roll.** Zero attempts, therefore zero errors. The
clean result is not evidence that the larger payloads are fine — it is evidence that
nothing has exercised them, and the error grep alone cannot tell those apart. The
`[UNMEASURED]` caveat I attached to the original 0/0 was right, and this is the
measurement that discharges it in the honest direction: not "fine", but "not yet
tested".

> **The check, added to the RUNBOOK above the watch:** before reading an error count
> as clean, count the ATTEMPTS. A denominator of zero makes any error rate look
> perfect. Same family as the recorded landmine *COUNT the baseline before believing
> an `EXCEPT`/`NOT EXISTS` diff — an empty baseline reports "all clear"*, and as
> *log measurement discipline*: I measured in a window during which the thing being
> measured never happened.

Corrected in the RUNBOOK, this file, and §7 of the handoff. The bug files' "0/0"
lines are left as written with this note as their qualifier, because they were true
statements about the error counts — the fault was in what I let them imply.

## §11 — 2026-07-28 ~20:00 — round 2: REVISE, decided by an UNREADABLE seat (the 119 case)

```
round 1: reviewers 11 | abstained 4 | unreadable 1 | decided_by "gating objection from tooling_provenance"
round 2: reviewers  9 | abstained 6 | unreadable 1 | decided_by "unreadable reviewer(s): review_editquality.result"
```

**`decided_by` names the parse failure as the decider.** This is `bugs_open/119`
exactly — ~11% of rounds decided by one seat's unparseable JSON — and it is why the
landmine says read `decided_by`/`unreadable` (**not** `abstained`) before the
objections. Round 1 had `unreadable: 1` too and was *not* the harness; the field to
read is `decided_by`, and only in round 2 does it name the unreadable seat.

**Substantively the round is 8 approve / 1 object**, and the movement is what matters:

- `tooling_provenance` — round 1's **gating** objector — now **APPROVES**: *"The
  gating objection (parallel self-built provenance trail) is discharged in substance
  … matching the travelling-docs convention this seat asked for, and the rationale
  explicitly names the recursion."*
- `prior_art_librarian` — **approves**; the pinned-ref grounding answered it.
- `debug_historian` — **approves**, with one low objection (below).
- `guardian` — the only objector, two objections, **both now checked**.
- `editquality` — the seat whose result was unreadable. Its round-1 objection is the
  one this round's main edit answers, so its verdict is the one I would most have
  wanted to read. `[UNKNOWN]` whether it would have approved; not inferable.

**No `Council-Reviewed:` trailer is claimed. The verdict is REVISE.** A
harness-caused REVISE is still a REVISE — the trailer is earned by APPROVED only,
and inferring approval from "8 of 9 readable seats approved" is exactly the kind of
arithmetic the trailer discipline exists to prevent.

### Guardian's two objections — both checked, neither a defect

1. **[medium] Do any of the ~137 `ActionInputSpec` registrations use POSITIONAL
   struct literals?** A new field would break those at compile time. **Checked: no.**
   All **168** construction sites in the tree are keyed literals
   (`ActionInputSpec{Optional: …}`); a grep for a bare value as the first element
   returns nothing. `go build ./...` passing already implied it, but the seat was
   right that the plan asserted safety without stating the check — so it is stated
   now as a checked fact rather than an inference from a green build.
2. **[low] Does anything else parse the audit's stdout or the tool's return shape?**
   The JSON went from a flat map to `{declared, conditional}`, which would break a
   silent consumer. **Checked: no consumer exists.**
   `grep -rn "audit-config-keys\|config-key-audit"` over the tree returns the script,
   the tool, and their own doc comments — nothing else. The only reader is
   `scripts/audit-config-keys.sh:47`.

   > That check caught a real defect the seat had not asked about: the tool's own
   > **usage comment still documented the OLD output shape**. A doc comment that
   > lies about its own output is the same class this whole bug is about, committed
   > by the fix for it. Corrected, with the "no consumer" finding written into the
   > comment so the next shape change knows what to re-check.

### debug_historian's low objection — fixed and proven

*"nothing is said about re-run safety — no pre-state marker or ON CONFLICT guard …
it's unclear whether it would silently duplicate the three doc_notes rows."* Correct:
`doc_notes` is append-only by design and has **no unique constraint to lean on**,
which is precisely why a re-applied seed would duplicate in silence and nothing would
complain. Each INSERT is now an `INSERT..SELECT` guarded by `NOT EXISTS` on
(subject_type, subject_key, created_by). Proven by re-applying the file:

```
BEGIN / INSERT 0 0 / INSERT 0 0 / INSERT 0 0 / COMMIT
 firecrawl_scrape 1 · scrape_web 1 · store_business_verification 1
```

Not asserted — the re-run was actually performed and the row counts re-read.
