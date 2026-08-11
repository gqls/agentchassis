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
