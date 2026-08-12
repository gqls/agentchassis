# HANDOFF — 2026-08-11 — both revalidators live on v1.0.1284; the cap scare is over and corrected; ONE decision open

**Read this file only.** It supersedes `HANDOFF_2026-08-10_continue_here.md` for state (that file
now opens with two correction banners; its §1–§4 reference material still holds). Working record:
`NOTES_deployed_asset_path.md` (newest at the bottom). Owner's log: `README_where_we_are.md`.

**Nothing is half-applied. Nothing is broken. One thing is deliberately not done** — see §0.

---

> ## 🟣 ROUND 6 VERDICT: **REVISE** (2026-08-12 13:28:37Z) — **`compliance` APPROVES WITH ZERO OBJECTIONS.** The gate worked. Read this first
>
> `decided_by`: **gating objection from `prior_art_librarian`**. 16 reviewers, 1 abstained, 0
> unreadable, not truncation-gated. Verdict saved: `VERDICT_2026-08-12_round6_*.json`.
>
> **THE SUBSTANTIVE OBJECTION IS CLOSED.** `compliance` gated rounds 3, 4 and 5 on the gate verifying
> the wrong thing; the owner ruled; this lane built the claim-granular gate; **that seat now approves
> with no objections at all**, as does `debug_historian`. That is what round 6 was for.
>
> **⚠ BUT THE OBJECTION SURFACE MOVED — from the design to the PROSE.** 11 approve / 5 object, where
> round 5 was 14/2. **Every new objection lands on text round 6 added, not on code.** The status
> paragraph written to fix `guardian`'s round-5 contradiction drew *two* fresh objections of its own.
> **Answering a seat by adding an assertion creates a new surface for a different seat. Round 7 must
> be SHORTER AND PLAINER, not better defended** — `constitution` (LOW) names the tone directly:
> *"dramatic ALL-CAPS meta-narrative … self-congratulatory measurement claims"*, and it is right.
>
> ### The gating objection is ALREADY FIXED — do not re-argue it
>
> `prior_art_librarian` HIGH: the justification for submitting shipped code rests on *"the owner ruling
> of 2026-07-29 §2"*, which it has no record of. **It was right on its own tooling.** The ruling is in
> `CLAUDE.md` (verified, 1 grep hit) and had **ZERO** `doc_notes` trace. Filing the fleet landmine
> *"CLAUDE.md's owner rulings are invisible to council seats"* and running `landmines-sync.py --apply`
> **put it there**: `body ILIKE '%ordering exemption%'` now returns **5**, including the exact phrase.
> Round 7 cites that query. ⚠ **My first check was `ILIKE '%2026-07-29%'` → 130 rows and I nearly
> recorded it as reassurance — it matches the DATE STRING. Query a ruling's WORDS, never its date.**
>
> ### What round 7 needs, ranked by what is actually worth doing
>
> 1. **A MECHANICAL guard on the shared loader** — `editquality` MEDIUM + `guardian` MEDIUM, and they
>    quote the submission against itself: risks §0a argues *"a comment is not a control"* and then
>    offers a comment. A test asserting `loadParkedReviewItems`' SELECT column list matches
>    `rows.Scan`'s destination order protects **all six** revalidators. **Strongest actionable item;
>    worth building whatever happens with the council.**
> 2. **Per-edit shipped evidence** — `editquality` MEDIUM: "EDITS 1-8 ARE ALREADY SHIPPED" was
>    generalised from line-level verification of **edit 8 only** (`:141`, `:441`, `:475`). Verify each.
> 3. **Mark the unverifiable AS unverifiable** — `prior_art_librarian` MEDIUM ×2 asks only for this,
>    not for different facts: pod-grep and commit-ancestry claims are **UNVERIFIABLE FROM THAT SEAT'S
>    TOOLING** (`code_symbols` is not in its schema either). Say so inline; it is the `[UNMEASURED]`
>    discipline CLAUDE.md already requires.
> 4. **Cut the rationale hard** (`constitution` LOW) — plain tone, no ALL-CAPS narrative, no round
>    history it can already read in `diagnosis_artifacts`. This also shrinks items 1–3's surface.
> 5. **A tracking artefact for the `voice_tells` asymmetry** — `bug_historian` MEDIUM + `reuse_agent`
>    MEDIUM; 016b §9's "one call site gets the rigorous fix, the sibling stays heuristic".
> 6. **A `doc_notes` NOTES entry** — `tooling_provenance` MEDIUM: the markdown register is a different
>    mechanism and does not substitute.
>
> ### ⚠ ONE OBJECTION IS RIGHT ON PRINCIPLE AND WORTH NOTHING MEASURABLE — do not sell it as a fix
>
> `reuse_agent` LOW: the emit side scans `datahelpers.ExtractAssertionText(html)`
> (`check_unverified_claims.go:527`) while `claimStillOnPage` searches **raw `html + contentJSON`** — a
> real **predicate-parity** violation in the lane whose founding principle is exactly that parity, and
> a bespoke matcher where `datahelpers/claims.go` already has one. **Measured before rewriting
> anything:** tag-stripping as a proxy for `ExtractAssertionText` gives **2 false refusals vs 2**, and
> **40/41 sensitivity vs 40/41** — *identical on both sides*. The two survivors (`5`, `97`) are in the
> prose, not the markup. **Make the change for parity and reuse; do not claim it fixes false positives.**
>
> ## 🟢 2026-08-12 — THE CLAIM-GRANULAR GATE IS BUILT (`58bede8d5`). Round 6 has since returned — see above
>
> Owner chose *"build the tighter check first, then let one round carry both"*. Done. Orchestration
> **`ec2b87a6-d695-4a78-9255-f488ab0fe859`**, same correlation `b67eb26a-…`.
>
> **What shipped.** `resolved` now requires, ON TOP of the owner's timestamp gate, that **every text
> the finding cited is absent from the slot it was cited from**. `ClaimsPageScan.ExaminedTextBySlot`
> holds the copy actually read (EXAMINED components only); the revalidator reads the item's own
> `spec.findings[].matched`. Three refusal arms, all non-terminal. Register **CQ-021** updated in the
> same commit. **Committed, NOT rolled** — verified, not assumed: `claimgate_NEW=0 rc=1` on the
> running binary.
>
> ### ⚠ THE THING TO READ BEFORE TOUCHING THIS GATE — the seat's own suggested fix was WRONG
>
> `compliance` asked for the **snippet** to be compared. Measured against the demand control (items
> whose verdict is `still_holds`, where the claim IS on the page by definition):
>
> | candidate | sees the claim when it IS there (n=41) |
> |---|---|
> | slot-scoped **`matched` token** (shipped) | **40/41 — 97.6%** |
> | slot-scoped **`snippet`** (proposed) | **18/41 — 43.9%** |
>
> A snippet breaks on any markup churn, and **in a gate a missed match reads as "the copy changed",
> which GRANTS closure** — it would have failed OPEN on ~56% of claims, strictly worse than the
> timestamp it was meant to strengthen. **An objection can be right about the defect and wrong about
> the remedy; the remedy is the half you can measure.**
>
> ⚠ **And scope to the SLOT, never page-wide.** Page-wide, a bare `unregistered_number` token (`5`,
> `26`, `97`) matches almost any markup: 7 of 18 "still present" page-wide, **2** slot-scoped. I was
> one step from filing that 7 as *"the gate would have refused 4 of 8 closures"* — an artefact of my
> own method. Full trap in `LANDMINES.md`; the near-miss in `WRONG_CALLS.md`.
>
> ### Open question for the owner, deliberately NOT decided by this thread
>
> The two gates are **ANDed**. The claim check is the *stronger* evidence — a token that has left its
> slot **proves** the copy moved, where a timestamp only asserts it — so it could **stand in for**
> the timestamp rather than be added to it. Requiring both can refuse a genuinely-fixed page whose
> `updated_at` was never bumped. Not taken here because **adding a condition in front of an
> owner-mandated control needs no new ruling; removing his comparison would.**
>
> ### Housekeeping a cold start needs
>
> - **Deploy state re-grounded 2026-08-12 ~13:00Z: v1.0.1290, 23/23 Running pods, ONE digest
>   `sha256:b69237df`.** Round 5's figure (v1.0.1288, 41 pods, `d080ae14`) was <24h old and **wrong on
>   all three**. The pod COUNT is churn; the DIGEST UNIFORMITY is the invariant.
> - **§4's "failing tests that are not this lane's" is RETIRED** —
>   `TestEveryCheckProducedItemTypeIsClassified` passes at HEAD; another session fixed it. Full
>   package green.
> - **`git log` the FILE, not my commits**: the LANDMINES + WRONG_CALLS entries landed in the 215
>   lane's `f8ca05594` (swept in the ~90s before my own commit); NOTES in `79d910a86` + `b182c0b15`.
>
> ## 🔵 ROUND 5 VERDICT: **REVISE** (2026-08-11 19:54:45Z) — but **14 of 16 seats APPROVE**. History below
>
> Fired at the owner's instruction; verdict in **~5 minutes**. 16 reviewers, **1 abstained, 0
> unreadable, NOT truncation-gated**. `decided_by`: **gating objection from `compliance`** — note it
> is inside `body`, the `metadata->>'decided_by'` column is **empty** this round.
>
> **THE FIVE REPAIRS WORKED — do not redo them.** `prior_art_librarian` went **gating objection →
> APPROVE** ("those artefact tables ARE in my schema … the kind of thing I'd want a future round to
> keep citing verbatim"). `tooling_provenance` went **MEDIUM → APPROVE** ("the right substitute").
> `editquality` went **MEDIUM → LOW** and calls the plan coherent. `debug_historian` endorsed the new
> deploy discipline by name (digest-uniformity, `/proc/1/exe` over `strings`, and the exit-code
> capture that defeats the `n=${n:-0}` trap). **Citing beat arguing.**
>
> ### Round 6 needs THREE things — one is not paperwork
>
> 1. **STATE THAT ALL 8 EDITS ARE ALREADY LIVE, in one line at the top of the rationale.** `guardian`
>    (MEDIUM) and `debug_historian` (LOW) independently read the plan as self-contradictory: it
>    describes edits as pending while citing 8 live closures as proof they work. **Settled first-hand:
>    the code IS shipped** — `revalidate_review_queue_action.go:141` (`CreatedAt` field), `:441`
>    (`created_at` in the SELECT), `:475` (the Scan destination), all from **`9a9fef332`**. Both of the
>    guardian's horns ("redundant" / "impossible") are wrong because the third option is never written
>    down: **shipped, and reviewed after the fact, which is the design** (owner ruling 2026-07-29 §2).
>    The `plan`/`edits` schema reads forward-looking, so an honestly retrospective submission looks
>    like a contradiction to any seat reading it cold. **Latent since round 1.**
> 2. **Fix edit 4's sketch signature — introduced by ME in round 5.** The first loop still calls
>    `unverifiedClaimsVerdict("p1", scan)` (2-arg, pre-gate) while the block added below calls the
>    3-arg form. I read the real code, fixed my half, and left the neighbour inconsistent. Same fault
>    as rounds 1–4, one level down.
> 3. **`compliance` HIGH, the gating one, raised for the THIRD time** — the gate is component-granular,
>    not claim-granular, so an unrelated edit to the same slot satisfies it. **This is §2.1.** The seat
>    is explicitly NOT vetoing ("per the seat's no-veto mandate") and names the fix: *"require the
>    specific finding's cited snippet (or its containing DOM/text node) to differ, not merely the
>    component's `updated_at`."* It asks that the gap be named **explicitly**, not folded into a
>    general "named next step". ⚠ **Owner sign-off exists but was conditional on *the gate*, and the
>    gate verifies THE PAGE MOVED, not THAT THE FLAGGED CLAIM WAS ADDRESSED.**
>
> **OPEN DECISION (owner's, recorded in `README_where_we_are.md`):** paperwork-only round 6 · build the
> claim-granular gate first and let one round carry both (**the session's recommendation**) · or stop
> submitting, since the code is live and every seat with design standing approved rounds ago.
>
> Residual, not actionable: `prior_art_librarian` LOW — the single-producer fact lives in function
> **bodies** and its tier "holds declarations only", so it is unverifiable from that seat whatever we
> write. (`code_symbols` *does* have a `body` column; the ceiling is the seat's tier, not the index.)
>
> ## 🟡 ROUND 5 WAS BUILT, VALIDATED AND COMMITTED — **and has since been FIRED** (history below)
>
> All five repairs the round-4 verdict asked for are in
> `SUBMISSION_2026-08-09_claims_unverified_revalidator.json` (commit below). **The code is
> untouched** — every seat with standing on the design approved it; all four objections were about
> the plan's evidence. Validated against every client-side check the trigger applies: **edits 8/8**
> (the cap is enforced at `097_TRIGGER:101` *before credits are spent*), operations valid, in
> scope, **41,101 of 65,536 bytes**, 26 `grounded_in` entries.
>
> **To fire it** (this is the only remaining action, and it is a spend decision):
> ```sh
> RESUBMIT_CORR=b67eb26a-14ef-45d7-b755-3e489fd57ef0 \
>   ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
>   docs/agent_docs/docs024_key_docs_latest/bugfix_168_deployed_asset_path/SUBMISSION_2026-08-09_claims_unverified_revalidator.json
> ```
> Cost is now **~58% below** the ~1.6M figure in §0 (the `244` caching fix is live and measured), but
> §0's "ask before running it" still stands — check the post-restart 300s window first.
>
> ### What was verified FIRST-HAND this session (none of it carried forward)
>
> | claim | result |
> |---|---|
> | `doc_notes` carrying the owner ruling | **9 rows** |
> | `council_report` rows for `b67eb26a-…` | **4** — 08-09 ×3, 08-11 ×1, all `revise` |
> | Standing measurement | **`0 \| 8 \| 19 \| 3`**, invariant `t` **8/8**, `refused_by_gate` still **0** |
> | Producers of `claims_unverified` (via `code_symbols`, not grep) | **1** — `(*UnverifiedClaimsCheck).Run` |
> | Deploy | **41/41 Running pods, ONE digest** `sha256:d080ae14…`, tag **v1.0.1288** |
> | Needles on that binary | `ownergate=1 claims=1 voice=1 CONTROL_pos=2 CONTROL_absent=0` |
>
> **⚠ §1's state table below is now STALE**: it reads v1.0.1284 / "both replicas". The fleet is on
> **v1.0.1288**, and "both replicas" was a **false completeness claim** — `-l app=agent-chassis`
> returns 2 pods while **41** run that image. Use digest-uniformity, never a replica count.
>
> ### Three method corrections that outlive this round
>
> 1. **Capture the exit code beside the count.** `rc=0` on every present needle and **`rc=1` on the
>    absent control** is what separates "grep ran and found nothing" from "I could not look".
>    `n=${n:-0}` collapses the two. Also: **`grep -a /proc/1/exe`, never `strings`** — the RUNBOOK's
>    line 151 recipe still says `strings` and its failure is invisible behind `2>/dev/null`.
> 2. **`build provenance` was UNREADABLE here, and the clever fix also failed.** The startup line is
>    at the *start* of the log, so `logs <pod> | head -c 300000` should beat `--tail` — it returned
>    nothing on a busy chassis pod *and* on two quiet pods sharing the digest. Rotation, not absence.
>    With no candidate sha to verify (a *discovery* grep for 40-hex is forbidden), **BLD-019 gave
>    nothing on this occasion — the needle+digest method is still load-bearing.**
> 3. **§5's "`code_symbols` indexes no package-level vars" is FALSE** — 700 vars are indexed,
>    including all four in `work_items_common.go`. It was the stated reason not to file the §2.4
>    diagnosis run; **that reason is void** (whether to file is still a judgement). Logged in
>    `WRONG_CALLS.md`: it was verified at the source file, which could never disconfirm a claim about
>    what the *index* contains.
>
> **Do not re-do the producer count with grep.** The reproducible form is a `code_symbols` query on
> `body/content ILIKE '%claims_unverified%'` — 3 rows, 1 producer — and its index commit
> (`286884b65`) is an ancestor of HEAD with **zero commits since touching this plan's four files**.
> `cmd/bundle` exists but is contextkit's, under its own go.mod, **not in this module's build**.

> ## ✅ UPDATE, later on 2026-08-11 — `bugs_open/244` IS ALREADY FIXED AND LIVE. DO NOT BUILD IT.
>
> Another session shipped both halves on 08-10 evening, ~2 hours after I filed it: `3d6851d9b`
> (opt-in `cache_control` breakpoint on the shared client + the `llm_call_log` counters, migration
> 376) and `071adc44c` (shared prefix hoisted in all 17 council seats). **I nearly rebuilt it** —
> my grep was a day old on a tree that moves fast, and the owner stopped me.
>
> **Measured live:** full-price input per council round **806,024 → 127,783**, with 973,554 cache
> reads and 93,333 writes ⇒ **~58% cheaper per round, ~69% cheaper per token**; hit rate **157/170
> = 92.4%** on read-eligible seats.
>
> **Two of my recommendations were wrong and are corrected in `244`:** the "≈76%" was optimistic
> (real ~58%), and `ttl: "1h"` was unnecessary — the data refutes my TTL concern, because reads
> keep the entry alive and seats past 5 minutes hit *more* often, not less.
>
> **Still open in `244`: adoption.** Only `council-gate` carries the marker (17 steps, no other
> agent type). That changes §0 below — a resubmitted round now costs ~58% less than the ~1.6M
> figure quoted there.

> ## 🔴 ROUND 4 VERDICT: **REVISE** (2026-08-11 12:51:34Z) — 16 reviewers, 1 abstained, 0 unreadable, NOT truncation-gated
>
> `decided_by`: **gating objection from `prior_art_librarian`**. Fourth revise, and — again — right.
> **Both HIGH objections are answerable with a QUERY, not an argument.** The seat did not claim the
> facts are false; it said it *cannot verify them from the submission*, and it named the exact
> traces it would accept. Both exist. The submission simply never handed it the queries.
>
> ### The two HIGH objections, and the evidence that answers them [MEASURED 2026-08-11]
>
> 1. **"OWNER RULING 2026-08-09 … I have no check tier for markdown files; the only DB-visible
>    trace would be a `diagnosis_artifacts` council_report or a `doc_notes` row … If the ruling is
>    fabricated framing, the gate's legitimacy claim collapses."**
>    → **9 `doc_notes` rows carry it.** Put this in `grounded_in` verbatim:
>    ```sql
>    SELECT count(*) FROM doc_notes WHERE body ILIKE '%register moved, not the page%'
>       OR body ILIKE '%copy-changed gate%'
>       OR (body ILIKE '%owner ruling%' AND body ILIKE '%claims_unverified%');  -- 9
>    ```
> 2. **"Extensive self-cited ROUND 1/2/3/4 history … no visibility into a prior round for THIS
>    submission unless it is in `diagnosis_artifacts` kind='council_report'."**
>    → **It is. 4 rows, this exact correlation** (09-09 ×3 revise, 08-11 revise):
>    ```sql
>    SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
>    WHERE correlation_id='b67eb26a-14ef-45d7-b755-3e489fd57ef0' AND kind='council_report' ORDER BY created_at;
>    ```
>
> **This is the lane's recurring fault for the fourth time: describing the work less carefully than
> it was done.** The fix is not to argue — it is to cite.
>
> ### The other objections worth actioning (none gating, all real)
>
> - **`editquality` MEDIUM — a genuinely missing edit.** `grounded_in` cites
>   `TestUnverifiedClaimsNeverResolvesWhenOnlyTheRegisterMoved` as pinning the gate (and as having
>   caught the zero-date defect), but **no edit adds it** — edit 4 is the only one touching
>   `revalidate_unverified_claims_test.go` and predates the gate. Also LOW: the risks section says
>   *"edit 9"* while the array holds **8**. Same class as round 3's finding.
> - **`guardian` MEDIUM — blast radius belongs in risks, not a code comment.**
>   `loadParkedReviewItems`/`parkedReviewItem` is the shared loader for **all 6** covered item
>   types; a SELECT/Scan mismatch breaks every revalidator, not just this one.
> - **`debug_historian` MEDIUM — and it caught a real defect in THIS session's own verification.**
>   See the pod-population correction below.
> - **`tooling_provenance` MEDIUM** — the producer-count question was answered with ad-hoc grep
>   where `cmd/bundle`/contextkit is the designated tool; that is the same method that produced the
>   false two-producer claim in round 1.
> - `compliance` (approve, 2 non-blocking): the gate is component-granular, not claim-granular —
>   already the §2.1 next job. `architecture`, `constitution`, `mission`, `reuse_agent`,
>   `bug_historian`, `guidelines`, and four guardians: **approve**.
>
> ### ⚠ CORRECTION TO THIS SESSION'S OWN POD-GREPS — `debug_historian` was right
>
> I reported *"both replicas"* on v1.0.1284 and v1.0.1286. **`-l app=agent-chassis` returns 2 pods;
> 26 pods run that image**, across 20+ deployments (`agent-build-dispatch-loop`,
> `agent-color-variable-fixer`, `agent-diagnose-agent`, …). That was a false completeness claim.
>
> **The correct proof is the image DIGEST, not a pod count** — cheaper and stronger than grepping 26:
> ```sql/sh
> kubectl -n ai-persona-system get pods -o jsonpath='{range .items[*]}{.status.containerStatuses[0].imageID}{"\n"}{end}' \
>   | grep agent-chassis | sort | uniq -c     # 21 pods, ONE digest ⇒ provably one binary fleet-wide
> ```
> Result: **one digest** (`sha256:dcd256f9…`), so the grep on any pod is evidence about all of them.
>
> ⚠ **And a NEW trap the runbook now carries: `n=${n:-0}` MASKS A FAILED EXEC AS A ZERO.** A pod
> returned `ownergate=0 cachemarker=0` and read exactly like a stale binary; it was a **completed
> job pod** (`phase Succeeded`) that cannot be exec'd at all. The `${n:-0}` idiom is required for
> `grep -c`'s exit-1-on-zero **and** it silently converts "I could not look" into "it is not there".
> **Always pair it with a per-pod positive control**, and filter to
> `--field-selector=status.phase=Running`.
>
> ### What round 5 needs (small, precise, no code change)
>
> 1. Add the two verification queries above to `grounded_in`.
> 2. File the missing test edit; fix the `edit 9` / 8-edit mismatch.
> 3. Move the shared-loader blast radius into `risks`.
> 4. Replace the "both replicas" deploy claim with the **digest-uniformity** proof.
> 5. Re-run the producer-count check through `cmd/bundle`/contextkit rather than grep.
>
> Then resubmit with `RESUBMIT_CORR=b67eb26a-…`. **Do not change the code** — every seat with
> standing on the design approved it; the objections are all about the plan's evidence, not its
> behaviour.

> ## ⏳ ROUND 4 RESUBMITTED 2026-08-11 12:42:22Z — §0 below is now HISTORY, read this instead
>
> Fired at the owner's instruction, **unchanged**, under `RESUBMIT_CORR=b67eb26a-…` on chassis
> **v1.0.1286**. Orchestration **`ae0915c2-e77a-4d02-94ce-32ced673317a`** — began executing seats
> within seconds (no 29-minute queue).
>
> **Pre-flight done, and worth repeating next time:** pod-grepped both replicas of v1.0.1286
> (`ownergate=1 claims=1 voice=1 cachemarker=1 CONTROL_pos=2 CONTROL_absent=0`) and checked the
> **300s post-restart dispatch window** (2,330s elapsed) — CLAUDE.md's silently-dropped-spawn trap.
> **`cachemarker` is a new standing needle**: it puts the caching fix in the *running binary*
> rather than only in `git log`.
>
> **Get the verdict by correlation, never by `doc_notes ... LIMIT 1`:**
> ```sql
> SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
> WHERE correlation_id='b67eb26a-14ef-45d7-b755-3e489fd57ef0' AND kind='council_report' ORDER BY created_at;
> -- 3 rows = rounds 1-3 (all revise). A 4th row is this round. Objections are in `body`.
> ```
> **If it REVISES again, read the objection before assuming it is procedural** — this council has
> been right three times running, twice about substance. If APPROVED, the trailer is
> `Council-Reviewed: b67eb26a-14ef-45d7-b755-3e489fd57ef0`; **never write that trailer on a verdict
> you have not read.** The docs commit for today already carries `Council-Submitted:` with the same
> correlation, which `098` credits automatically once approved — so no amend is needed.

## 0. THE ONE OPEN ITEM (HISTORICAL — superseded by the block above) — council round 4

Round 4 died on infrastructure, not on content (see §2). The council is available again, so it
**can** be resubmitted unchanged, and the submission file is already correct and committed:

```sh
RESUBMIT_CORR=b67eb26a-14ef-45d7-b755-3e489fd57ef0 \
  ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
  docs/agent_docs/docs024_key_docs_latest/bugfix_168_deployed_asset_path/SUBMISSION_2026-08-09_claims_unverified_revalidator.json
```

**Why it has not been fired: cost, one day after the account hit its cap.** [MEASURED] one council
round is **~1.6M input tokens** (15 seats × 106k–118k), and `bugs_open/244` establishes the council
is **87.8% of all fleet input spend**. Rounds 1–3 were each right about something real, so the
review has value — but firing ~1.6M tokens the day after the budget blew is an owner's call, not a
thread's. **Ask before running it.**

Rounds 1–3 read-out with every objection and its answer:
`OBJECTIONS_2026-08-09_claims_unverified_council.md`. Round 4's content answers editquality's
round-3 objection by filing the `parkedReviewItem.CreatedAt` wiring as its own edit.

⚠ **Query the verdict by YOUR correlation, never `doc_notes ... ORDER BY created_at DESC LIMIT 1`**
— that returns whichever lane finished last and reads entirely plausibly until it starts
discussing someone else's bug.
```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='b67eb26a-14ef-45d7-b755-3e489fd57ef0' AND kind='council_report' ORDER BY created_at;
-- objections are in `body`, NOT `metadata`
```

## 1. State — verified 2026-08-11 09:45Z, chassis **v1.0.1284**, both replicas

| thing | state |
|---|---|
| `bugs_closed/168` (asset path) | CLOSED, live since v1.0.1229 |
| **`voice_tells` revalidator** | **LIVE + PROVEN** |
| **`claims_unverified` revalidator** | **LIVE + PROVEN** — 8 items closed, all with genuinely-edited copy |
| **Owner's copy-changed gate** | **LIVE, HELD on 8/8, FIRED 0 times.** Still `[UNEXERCISED]` |
| Pod-grep on v1.0.1284 | `ownergate=1 claims=1 voice=1 CONTROL_pos=2 CONTROL_absent=0`, both replicas |
| Latest sweep | **2026-08-11 08:44:19Z**, ran normally |
| Covered types | 6 |
| Council | rounds 1–3 REVISE (all answered); **round 4 dead, awaiting a decision to resubmit** |
| **`bugs_open/244`** | **OPEN** — council prompts 98.6% identical, uncached, uncacheable as ordered |
| API usage cap | **LIFTED** 2026-08-10 18:12Z (was ~3h20m, not the 3 weeks I first claimed) |

### The measurement this lane owes on every visit — re-run it

```sql
SELECT count(*) FILTER (WHERE result #>> '{revalidation,reason}' LIKE '%register moved, not the page%') AS refused_by_gate,
       count(*) FILTER (WHERE result #>> '{revalidation,verdict}'='resolved')    AS resolved,
       count(*) FILTER (WHERE result #>> '{revalidation,verdict}'='still_holds') AS still_holds,
       count(*) FILTER (WHERE result #>> '{revalidation,verdict}'='unknown')     AS unknown
FROM site_work_items WHERE item_type='claims_unverified' AND result ? 'revalidation';

-- the load-bearing invariant: EVERY closure must show the copy moved. Zero `f` rows, always.
SELECT (result #>> '{revalidation,evidence,newest_component_update}')::timestamptz
         > (result #>> '{revalidation,evidence,item_filed_at}')::timestamptz AS copy_actually_changed,
       count(*) FROM site_work_items
WHERE item_type='claims_unverified' AND resolution_path='auto:revalidated' GROUP BY 1;
```
**2026-08-11: `0 | 8 | 19 | 3`, invariant `t` for 8 of 8.** ⚠ **Do not describe the gate as having
prevented anything until `refused_by_gate` is non-zero.** Its four failure modes are pinned by
unit tests, not by observation.

## 2. What happened on 08-10, and the correction that matters more than the incident

Round 4 reached `COMPLETED @ complete_invalid` with `plan_valid: true` — **the submission was fine
and had been accepted**; a seat's LLM call was refused because the Anthropic account had hit its
usage limit. The 400 stated a reset of `2026-09-01`.

> **I then made the mistake worth reading this section for: I reported the stated reset as a
> forecast** — "a 21-day fleet-wide LLM outage" — into five files, and escalated to the owner as
> "plan around three weeks". **It was ~3h20m** (last failure 17:02:12Z, first success 18:12:11Z):
> the owner raised the cap. **The stated reset is the worst case if nobody intervenes.**
> **Verify a lift on the SUCCESS side of `llm_call_log`** — the failures stop appearing either way:
> ```sql
> SELECT date_trunc('hour',created_at), count(*) FILTER (WHERE success) AS ok
> FROM llm_call_log WHERE created_at > now() - interval '24 hours' GROUP BY 1 ORDER BY 1;
> ```

Logged in `WRONG_CALLS.md` together with two siblings from the same session (a pod-log grep whose
window was ~2 minutes wide; a near-duplicate filing another lane had already made). **The
generalised check, asked before writing the claim, not after: "if I were wrong, what would this
measurement look like?" If the answer is "the same", it is not evidence.**

## 3. `bugs_open/244` — the real finding, and it is unfinished work

The budget was exhausted on the **10th of the month**, so raising the cap **moved the wall rather
than removing it**. Measured Aug 1–10:

- fleet **188.1M** input tokens; **`council-gate` 165.2M = 87.8%**, over **209 rounds**
- **790,551 input tokens per round**; 11–15 seats at 106k–118k each
- three seat prompts from one round, byte-wise: **common prefix 20 chars; shared block 268,980
  chars = 98.6%**; seat-specific head only 1,387–5,159 chars
- `grep -rn "cache_control" --include=*.go platform/ pkg/ internal/` → **nothing**
- `platform/aiservice/anthropic.go:103-116` sends **one `user` message, no `system` field**

**Two defects; fixing either alone buys nothing:** caching is off, *and* the shared block sits
**after** the seat header so a prefix cache could never hit. Fix = move the shared block into
`system`, seat instruction last, one `cache_control` breakpoint, **`ttl: "1h"`** (rounds run
**459s mean / 1022s max**, so the 5-minute default expires mid-round). ≈**76% off the council**.

⚠ **`llm_call_log` has no cache columns** — a caching fix is unverifiable until they are added;
that is part of the work. Watch for `cache_read_input_tokens` staying 0, which is what a surviving
silent invalidator looks like. **This change is platform-scope and should go through the council
gate** (which is also the thing it makes cheaper).

## 4. Traps specific to this lane

- ⚠ **A roll is not evidence.** Pod-grep every replica; needles + gotchas are now in the RUNBOOK.
  **Honest limit: `CONTROL_absent` is a fabricated string**, so the check proves grep returns 0 —
  it does **not** prove the binary is newer than v1.0.1279. `pc.locked_at IS NULL` is unusable as a
  negative control (**17 hits** in other `platform/` files).
- ⚠ **The revalidation stamp key is `at`, NOT `checked_at`.** A wrong key returns 0 rows and reads
  exactly like "nothing was scanned".
- ⚠ **`uncovered_backlog` CANNOT confirm an adoption** — flat at 625 before and after a working
  roll. Confirm at the per-type map and at `scanned` decomposed by type.
- ⚠ **A dispatch of this sweep CANNOT be scoped** — both filters read from step config; the live
  `sweep` step has no `input_mapping`, so filters in `input_data` are inert and the run goes
  fleet-wide while looking scoped.
- ⚠ **The council refuses a plan with >8 edits** and takes only `modify|add|remove|config_change`.
- ⚠ **`landmines-sync.py --apply` before `landmines-verify-dispatch.sh` consumes the "new entry"
  status** — CLAUDE.md's own ordering. I hit this on 08-11; the entry's verification never fired.
  Deliberately not re-triggered by hand: the sibling landmine records that `code_symbols` is 100%
  Go while **81% of footprints** (incl. this entry's DB tables and step names) can never resolve,
  so the verdict would be noise. Use `landmines-verify-dispatch.sh` **instead of** `--apply`.
- **Registering a revalidator is the whole wiring** — `coveredItemTypes()` derives from
  `reviewRevalidators`; `TestRevalidatorCoverageIsDeliberate` pins the set deliberately.
- **`platform/orchestration/actions` has FAILING TESTS THAT ARE NOT THIS LANE'S** —
  `TestEveryCheckProducedItemTypeIsClassified` fails at clean HEAD from `e1628f7df`. **Reproduce
  against `git archive HEAD` before attributing any failure to your change.**
- **`/tmp` is a near-full 16G tmpfs** — use `TMPDIR=/home/ant/.cache/buildtmp go build ./...`.
- **Use `git commit -F <file>`**; backticks in `-m` execute.

## 5. Still open from the 08-10 handoff (unchanged, all still valid)

§2.1 tighten the gate from component-granular to **claim-granular** (two seats named it
independently; compare `spec.findings[].matched` against current copy instead of timestamps) ·
§2.2 resolve the **two-standards asymmetry** with `voice_tells` before a seventh type ·
§2.3 pin `ScanDeployedClaims` to its intended callers · §2.4 the **invisible backlog** (467 rows
across six unselected statuses; **do not file a third diagnosis run** — `code_symbols` indexes no
package-level vars, so membership is unreadable by the loop and is first-hand verified at
`work_items_common.go:140-143`) · §2.5 Decision 2's dedup half (**47 pairs / 168 rows**, owner
judgement) · §2.6 more sweep coverage (**use the status filter or the census lies**) · §2.7 the
armed-but-inert cap at `check_image_source_unsatisfiable.go:167`.

## 6. Commits and correlations

`ef80216be` voice_tells · `4030cadb9` claims_unverified + CQ-021 · `6ab7ff594` producer-count
correction · `9a9fef332` the owner's gate · `c70c8e1de` round 4 · **`2d979ddf0` file 244 +
escalation** · **`5c3322aa8` the three-weeks correction**.

| what | id |
|---|---|
| claims_unverified council (r1–r3 REVISE, r4 dead) | `b67eb26a-14ef-45d7-b755-3e489fd57ef0` |
| round 4's orchestration (died at `complete_invalid`) | `2f1b43f6-d92b-49eb-843b-204d0da235fa` |
| voice_tells council (APPROVED r1) | `4d430ca8-7e34-479a-95f3-71fdc12fdef6` |
| selection filter council (APPROVED r1) | `f64da546-e1a4-42e4-98c0-d94cf42af71c` |
