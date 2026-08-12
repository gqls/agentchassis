# NOTES — silent hero/logo readers (commission item 2)

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-08-11 — opening the lane

Picked up from `diagnosis_schema_visibility/HANDOFF_2026-08-11_continue_here.md` §4: item 5
done, commission order puts item 2 next, and the owner already ruled *"2. yes."* so there was
no decision to wait on.

**Coverage check before starting** (CLAUDE.md: checking the pod does not check the queue):

- `scripts/who-owns.py 236` — prints the documented **ambiguous-number** warning: two unrelated
  cases share 236 (the 522-availability one and this hero/logo one). Resolved by slug throughout.
  Recent commits against the hero/logo file are the item 5 lane's own contributions
  (`e41342c89`, `1fbef1b70`) — no competing owner.
- `git log --since=2026-08-08 -- v3_site_actions.go assemble_from_library.go` — active, but
  nothing touching the hero/logo reader blocks.
- Working tree: neither file is dirty (`git status --short`), so no uncommitted session on them.

### MISSTEP AVOIDED — the commission's line numbers had drifted

The commission names `v3_site_actions.go:1020` and `:1031`. Reading those lines gives
`config := params.StepConfig.Config` and a comment about sources config — nothing to do with
hero. The real sites are **`:1125`** and **`:1136`**. Had I edited by line number I would have
patched the wrong block.

**The check that caught it:** grep for the *symbol*, never trust a file:line from a doc more than
a day old. Recorded in the RUNBOOK as the first command of this lane.

### The finding that changed the design: a `Warn` would not survive

The commission asks for a `Warn`. Bug 236 §5b — written by the previous lane, i.e. by me —
asserts item 2 writes the evidence *"into `agent_error_log`"*. **Those are two different
mechanisms and I had conflated them.** A `zap.Warn` goes to the pod's stdout.

What settles it, and none of it is mine:

- `log_action_error.go:14-18` and `agenterrors.go:20-24`: *"THIS TABLE IS THE ONLY SINK THAT
  SURVIVES AN AWAITED STEP: the collected_data sibling key was refuted live."*
- the 236 contribution's measured 4-hour retention on `AWAITING_RESPONSES`;
- the item 5 lane's own measurement that `agent-chassis`'s startup line is absent from
  `--tail=3000` hours later.

So the design is Warn **plus** a durable row. Written up as PLAN Decision 1.

> **This is a correction to my own §5b wording in `bugs_open/236`**, which implied item 2 as
> commissioned would produce durable rows. As commissioned (a `Warn` only) it would not. The
> file now says so.

### The second finding, which belongs to item 1 and not here

`deploy_image_asset_action.go:404-415` **already** writes `hero_url`/`logo_url` directly into
`params.CollectedData` as a sibling key, with the comment *"so it survives the git adapter
response overwriting this step's output_field"*. Dated `d45c86b1e`, **2026-02-23** — five months
before this bug was filed.

Two things follow, and both are contributions to `bugs_open/236` rather than work for this lane:

1. **Someone already observed the overwrite**, which is relevant to §5's open root cause — the
   comment asserts the response *does* overwrite `output_field`, while §5 records the merge code
   as preserve-then-add. Both readings are in the tree; nobody has reconciled them.
2. **The workaround is the exact pattern `agenterrors.go` records as REFUTED LIVE** for the
   awaited class: *"persistAwaitingStateWithRetry loads fresh state at park time and copies only
   awaited-request entries across"*. That predicts the sibling key is dropped at park — which
   is precisely what 236 §1 measured (`hero_url`/`logo_url` both absent from `collected_data` on
   the decisive row) **despite this workaround having been live for five months**.

That is a testable mechanism for a symptom 236 currently lists as unexplained. It is `[UNVERIFIED]`
by me — I have read the two comments and the measured row, not the park path — and I am
deliberately not fixing it, because the remedy is item 1's design decision, which is the owner's.

### Scope discipline held

Tempting, and refused: have the readers accept `response.data.file_path`. That is 236 §4
candidate 2, the commission forbids it in item 2's own text, and it would encode the merged shape
at three call sites — the `unified_extractor.go:200` pattern the census already flagged.

---

## 2026-08-11 (later) — the tests, and the two mutations that prove they mean something

Ten tests, all green, and the whole `actions` suite green with them. But a green suite is not
evidence until a guard has been shown capable of failing, so both load-bearing behaviours were
**proven by mutation**:

| mutation | test that must fail | result |
|---|---|---|
| `if !present {` → `if false && !present {` (demand gate removed) | `TestDeployedImageURL_AbsentContainerRecordsNothing` | **FAILED on both assertions** — the Warn count *and* the write-attempt detector |
| `landed := LogActionError(...)` → `landed := true \|\| LogActionError(...)` (durable write removed) | the three recording tests | **all three FAILED**, including the two that go through the real actions |

The helper was restored and `diff`ed byte-for-byte against a pre-mutation copy before anything
else happened. Recorded because a mutation you forget to revert is the worst possible outcome of
this technique — see below for how nearly that went wrong.

### THE TREE MOVED UNDER ME — and this is the one to read

Between writing the code and committing it, **another session's commit `038211dd8` ("215 REVISE
round 1") swept all four of my files into HEAD**: the two edited readers, the new helper, and its
test file. Untracked files included, so it was an `add -A`-shaped commit. This is precisely the
hazard CLAUDE.md documents — *"it cannot stop a session that still runs `git add -A` from sweeping
up yours, half-finished, into a commit about something else entirely."*

**The dangerous part is not the attribution — it is the timing.** I had been mutating that file
minutes earlier. Had the sweep landed during either mutation window, HEAD would now carry
`if false && !present` — a deliberately disabled guard, in a commit about something else, with a
green-looking test suite in my scrollback and nothing to connect the two.

**Checked rather than assumed**, because the cost of being wrong here is a broken fleet:

```
git show HEAD:.../deployed_image_read_audit.go | grep -nE "if false|landed := true"   # nothing
git archive HEAD | tar -x -C <tmp> && go build ./... && go test ./platform/orchestration/actions/
#   → HEAD BUILD OK; ok  .../actions  0.524s
diff <tmp>/.../{helper,test,v3_site_actions,assemble_from_library}.go  <working tree>
#   → IDENTICAL at HEAD, all four
```

So HEAD carries the restored code and is green. Nothing is lost, forward-only holds, and per
CLAUDE.md the remainder gets committed with the sweep named in the message.

> **The check that would have made this cheap: never leave a mutation in the tree across a tool
> call you don't control.** Back the file up first (I did), mutate, test, restore, and `diff`
> against the backup **immediately** — not at the end of the session. On a tree this many
> sessions share, the window between mutate and restore is a window in which someone else can
> publish your mutation. Added to the RUNBOOK as part of the mutation recipe.

### Submitted to the council

`SUBMISSION_CORR=c80ea1d7-ce1e-493f-8175-877501d895e6`. The submission leads with the one thing
that genuinely needs a judgement — that the commission asked for a `Warn` and this ships a durable
row as well — rather than burying it in the risks block. Five further risks disclosed, including
that the row-rate is `[UNMEASURED]` and that 236 §3 explicitly forbids quoting "2" as an incidence
rate.

---

## 2026-08-11 (evening) — APPROVED, and the two medium objections were both worth having

**Verdict: APPROVED**, round 1, ~11 minutes end to end. *"approved with 1 advisory objection(s) —
none high-severity"*, 6 seats abstained. The `architecture` seat signalled `point_fix` and
explicitly endorsed the boundary: *"observe first, redesign second… this plan deliberately does
NOT touch that surface."*

The declared departure survived every seat that looked at it. `constitution`: *"this is
transparency, not a violation."* `guardian`: *"disclosed with measured justification… uses the
existing LogActionError door rather than inventing a new sink."*

**Two mediums, and neither was answerable by agreeing with it.**

### 1. `prior_art_librarian` — "the 4-hour figure is load-bearing and you cannot check it here"

Verbatim: *"scheduled_tasks is not in the Schema section available to this council, so this cannot
be checked here; it should not be treated as settled just because it is asserted with a specific
number."*

**The seat was right to press, and the objection lands on ME, not on the lane that measured it.**
I had repeated the figure from the other 236 lane's contribution — which is exactly what CLAUDE.md
forbids: *"Ground every figure against the live system before repeating it from another doc."* I
followed the norm about marking claims and skipped the one about re-measuring them.

Checked first-hand, live `scheduled_tasks.pre_query`:

```
DELETE FROM orchestration_states
 WHERE status IN ('COMPLETED', 'FAILED')       AND updated_at < NOW() - INTERVAL '24 hours'
DELETE FROM orchestration_states
 WHERE status IN ('EXECUTING_STEP', 'AWAITING_RESPONSES') AND updated_at < NOW() - INTERVAL '4 hours'
```

**CONFIRMED.** The figure is now this lane's own measurement, not an inherited one.

### 2. `bug_historian` — "you patched the three sites 236 named, not the class"

Verbatim: *"the round should not treat 'the three sites named in 236' as 'the whole exposure'
without a sweep."* A fair challenge, and the honest answer needed a census rather than a promise.

**The census, package-wide** (`params.CollectedData[…].(map[string]interface{})` with an `ok`
guard, non-test): **64 occurrences.** But the shape is not the defect, and separating them is the
finding:

| class | count | is it the 236 defect? |
|---|---|---|
| config/input reads (`input_data` 31, `agent_config` 10, `__raw_message__` 4, …) | ~50 | No — not awaited results |
| loaded records (`site_record`, `business_record`, `render_context`) | 5 | No — not awaited results |
| **awaited/spawned results** | 4 | **checked individually — see below** |
| dynamic key from config (`CollectedData[field]`, `[path]`, `[fieldName]`, `[key]`) | 4 | **`[UNVERIFIED]` — semantics depend on the config** |

**All four awaited-result readers were opened and read. None is the 236 defect:**

- `generate_image_actions.go:777` (`adapter_response`) — **fails loudly**: the `else` returns
  `no response data found from adapter`. Nearest sibling to 236 (same image family) and it is clean.
- `call_agent.go:736` (`spawn_agent`) and `:375` (`spawn_<type>`) — **legitimate fallback**: a miss
  falls through to *"Need to spawn a new agent"*.
- `spawn_actions.go:1804` (`start_orchestration`) — **legitimate fallback**: a miss returns nil,
  meaning "no existing child", and the caller starts one.

> **So the discriminator is not "an `ok` guard with no else". It is "a miss whose only consequence
> is that the artefact is quietly worse."** In all four siblings, absence either fails the step or
> genuinely means *do the other thing*. In the three hero/logo sites, absence meant *ship the page
> without its image and say nothing* — and that is why those three were the defect and these four
> are not.
>
> **This narrows the objection rather than dismissing it**, and it leaves a real residual: the four
> dynamic-key readers cannot be classified by reading the Go, because the key comes from step
> config. `[UNVERIFIED]` and recorded as the follow-up, not silently dropped.

### The two low objections, and what was done with them

- **`debug_historian`** — verify the eventual roll at the pod, not at git. Right, and the new
  `error_code` is a compiled string literal, so it greps. Added to the RUNBOOK as the named target.
- **`tooling_provenance`** — leave a `doc_notes` row recording the departure. **Not done, and
  deliberately.** Landmines have a sanctioned writer (`landmines-sync.py`, and CLAUDE.md forbids
  hand-writing those rows); this class has none I could find, and inventing a hand-written row is
  the failure mode that rule exists to prevent. The departure is instead recorded in six places a
  reader will actually look: PLAN Decision 1, these NOTES, `README_where_we_are`, the submission
  rationale, `bugs_open/236`, and the commit message. **Flagged here as an accepted residual, not
  as closed.**
- **`editquality`** — `siteID`/`domain` looked unbound in the sketch. A sketch-elision artifact;
  the real helper resolves both (`deployedImageAuditSiteID`, `extractDomainFromParams`). No change.

---

## 2026-08-12 — LIVE on v1.0.1290, and the 090 came back UNVERIFIABLE for a reason worth more than the run

### Item 2 is live, verified at the artefact

`agent-chassis:v1.0.1290`, both replicas started 2026-08-11 21:53Z. Verified the way the RUNBOOK
says, not by trusting the tag:

```
kubectl exec <pod> -- grep -aq "DEPLOYED_IMAGE_RESULT_MISSING_URL" /proc/1/exe   # PRESENT, both pods
kubectl exec <pod> -- grep -aq "DEPLOYED_IMAGE_RESULT_MISSING_URL_NOT_REAL" ...  # absent (control)
```

The `error_code` is a compiled literal, so this change is datable without planting a marker — which
is what the council's `debug_historian` seat asked for.

### The behavioural proof has NOT happened, and the demand control is why I can say so

| | |
|---|---|
| `agent_error_log` rows with the new code | **0** |
| **demand control:** `hero_deployed` / `logo_deployed` in `orchestration_states` | **0 / 0** of 6,364 retained |

**So the zero is unfalsifiable, not reassuring.** Nothing has deployed a hero or logo since the
roll, so the path has had no opportunity to fire. Recording it this way because a bare "0 rows,
looks quiet" is exactly the reading `bugs_open/236` §3 forbids — and the control is the only thing
separating "nothing broke" from "nothing was tried".

### The 090: UNVERIFIABLE — neither confirmed nor refuted

`dbcc4259-…`, COMPLETED 18:42Z on 08-11, 4 iterations. It reached the same `next_scope` I did and
could not read the code it needed. Its `needed_evidence` names the blocker exactly: the bundle
rendered `storeActionResult`'s body and **a bare signature line** for `applyResponseToState`, and
nothing for `persistAwaitingStateWithRetry` or `processAwaitResponse`.

> **⚠ The verdict artifact is NOT in `diagnosis_artifacts`.** Only 4 `kind='bundle'` rows are there;
> the verdict lives in `orchestration_states.collected_data->'verdict'` on the run's own row.
> My first two polls queried `kind='diagnosis_report'` and `metadata->>'verdict'` and both returned
> "NOT YET" on a run that had finished five hours earlier. Added to the RUNBOOK — a poll that looks
> in the wrong place reports "still running" indefinitely, which is the same trap as this lane's
> DNS-failure watcher, one layer up.

### The finding that outlives this bug: the CODE tier has the SAME defect item 5 fixed in the SCHEMA tier

All four bodies **are** in `code_symbols` — 2,058 / 5,619 / 4,746 / 970 chars, correct line ranges.
The index held four; the bundle rendered one. So this is a **rendering/selection defect in the code
tier**, and the verdict's cite-or-abstain rule then acts on the absence — precisely the mechanism
item 5 fixed one tier over, where a filtered-out table and a non-existent table read identically.

**Item 5's own PLAN §3 flagged this and I did not follow it up:** *"whether the code tier has an
analogous blind spot is unexamined `[UNMEASURED]`"*. It is measured now.

> **CORRECTION to `bugs_open/236` §5b, which I wrote.** It declared this blocker "clear" because
> *"the index is fresh and carries all three… 4,746 chars"*. True, and an answer to the wrong
> question — the loop had complained about the **bundle**, not the index. A fresh index returns
> "present" whether or not the bundle renders it, so that check **could not have come out false**.
> Corrected in place in 236, and logged in `WRONG_CALLS.md` with the cheap check that would have
> caught it: read the bundle artefact, not the table it is built from.

**Consequence:** a third `090` on the code-path question will fail identically until the code tier
is fixed. It should be filed as a diagnosis-harness defect, not as another run on 236.
