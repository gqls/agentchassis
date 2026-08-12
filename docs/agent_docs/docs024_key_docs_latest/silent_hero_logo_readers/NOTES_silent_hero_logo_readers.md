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

---

## 2026-08-12 (afternoon) — the code tier: found, measured, fixed, submitted. And three of my own claims went wrong in four hours

Picked up from this lane's own `HANDOFF_2026-08-12` §4, which recommended fixing the diagnosis
code tier over running a third `090` on `bugs_open/236`. Filed as `bugs_open/261` — **not 260**,
see the misstep below.

### The handoff's framing was right in its conclusion and wrong in its mechanism

It said: *"The index held four bodies; the bundle rendered one. So this is a rendering/selection
defect in the code tier."* Conclusion correct, mechanism not.

**What I found by reading all four bundles instead of the last one:**

| iteration | in-scope symbols | rendered | the three the loop wanted |
|---|---|---|---|
| 1 | 12 | 3 | `persistAwaitingStateWithRetry` ✓, other two ✗ |
| 2 | 18 | 16 | `applyResponseToState` ✓ |
| 3 | 18 | **18 — no INCOMPLETE notice at all** | `persistAwaitingStateWithRetry` ✓, `processAwaitResponse` ✓ |
| 4 | 5 | 4 | none |

So the bundle **did** render those bodies — just never all in one iteration, and **the bundle does
not accumulate across iterations**; the verdict reads the last one, which had collapsed to five
symbols, three of them copies of a trivial `getMapKeys`.

**The tell was in iteration 1, and it is total.** Nine of its twelve symbols were unreadable, and
**every one of the nine is receiver-form** (`(*StateRepository).UpdateState`,
`(*SagaCoordinator).applyResponseToState`, …) while **all three that rendered are bare names**. A
100% split on a 12-row sample is not a coincidence worth theorising about; it is a spelling.

### The defect, in one line each

- `code_symbols_actions.go:598` writes a method as `"(" + fn.Receiver.Type + ")." + fn.Name`.
- `scopeFromCodeResults` (`diagnose_assemble_bundle_action.go:725`) concatenates that column
  verbatim into a scope entry.
- `spanOf` split on the last dot and compared the **raw** `(*SagaCoordinator)` against
  `receiverType()`'s `SagaCoordinator`. Never a match. All 1,170 indexed methods, unreadable.
- Second gap, plainer: `FileInfo.Values` (package-level `var`/`const`, `bugs_open/223` phase 2,
  1,238 rows) was written by the indexer and **never searched** by `spanOf`.

**Why it survived**: `symbolbody_test.go` asserted the dotted `Type.Method` form — a spelling **no
producer emits**. The test and the code were blind in the same direction and agreed with each
other.

**The part I did not expect**: *our own documentation teaches the failing input.* The LANDMINES
entry added 2026-08-11 by the sibling lane says, of a `code_request`, *"Name it
`(*Receiver).Method` … or you will be told nothing exists and be given no reason to doubt it."*
That is correct for the index query and wrong for the body read, and nothing distinguished the two
paths. Corrected in place, with the interim rule stated: until the fix rolls, **bare name in
`next_scope`, qualified name in a `code_request`.**

### The fix, and the one judgement in it

`splitReceiver` normalises `(*T).M` / `(T).M` / `*T.M` / `T.M` to the bare receiver; `spanOf`
searches `Values`. Fixed **at the reader, not the producers**, because there are two producers of a
scope entry and only one is code we control — the other is an LLM copying whatever spelling the
bundle showed it. Same "one function owns the grammar" judgement as `bugs_closed/163` and `189`.

**The judgement worth pressing on: widening the SPELLING must not widen the MATCH.** The lazy fix —
drop everything before the dot — passes every positive case and silently returns some other type's
method. Four negative controls guard it, and **the controls were proven capable of failing by
mutation**: with `if false && wantRecv != "" …`, `(*Nope).Greet` resolves to `Greeter`'s body and
the test fails. Test written against unmodified HEAD and **measured failing there first**, on
exactly the five broken spellings.

**Both the fail-then-pass and the mutation ran in a throwaway `git archive HEAD` checkout**, so the
shared tree never held a mutation — this lane's own §7 lesson, applied. It cost nothing and I would
not do it any other way again.

---

### MISSTEP 1 — I wrote a 100% into a council submission before checking it. It was 19/20.

I censused the failures, found 301 receiver-form and 20 others, **read the 20 names**, recognised
them as `var`/`const`, and wrote *"NOT ONE was a genuinely absent symbol"* — into the submission's
`grounded_in` and into `6911c2da4`'s commit message. Both immutable.

Ten minutes later I ran the query I should have run first
(`SELECT kind FROM code_symbols WHERE (path,symbol) IN (…)`): **19**, not 20. `controllerAddress`
(`platform/kafka/topic_manager.go:318`) is a plain `func`, missing from the analysed snapshot
because the index sat at `46b507ed1` (08-11 18:49) while the function landed in `e1f960ac2` (08-12
14:20). Index staleness — `bugs_closed/108`'s class — **and my fix does not cover it.** Honest
figure: **334 of 335**.

**The galling part**: that sentence existed *to be* the disconfirmable one. It could have come out
otherwise, it did, and I published it before letting it. Writing "here is the result that could
have refuted me" is worth less than nothing when the test has not been run — it buys trust with the
appearance of the rigour it skipped. Logged in `WRONG_CALLS.md`. **A name is not a kind.**

### MISSTEP 2 — the number moved under me. This is 261; the commit says 260.

Checked `260` was free, wrote ~2,000 words, and another session took it for
`one_mistyped_llm_field_silently_degrades_a_whole_component_render`. `6911c2da4`, the submission
JSON and three source comments cite **260** and mean **261**. Comments corrected in `dfb7ffbab`;
the two commit messages cannot be. **A number you checked is a number you checked at that
instant.**

### MISSTEP 3 (not mine, but it landed on me) — the tree took my LANDMINES edits again

`d878aa8f7` — another session's **finetuning.uk audit**, entirely unrelated — swept both my
`LANDMINES.md` changes (the correction and the new entry) into itself between my write and my
commit. So `dfb7ffbab`'s pathspec matched a clean file and the scope report listed 4 files, not 5.
**Nothing is lost**: both are at HEAD, verified with `git show HEAD:… | grep -c`. Forward-only
holds. This is the third time in two days this lane has had work land under another session's
message; the practice that makes it harmless is committing narrowly and *checking HEAD afterwards*,
which is how I noticed at all — the scope report's file count was one short.

### The census is a moving target, and that is itself the finding

321 failures / 44 runs at ~14:50 BST. **335 / 47 at ~15:30.** Fourteen further function bodies lost
to this defect in forty minutes. Any figure in these docs is a snapshot — re-run the query rather
than quoting it. It also means the cost of *not* rolling this is measurable and ongoing.

### What is NOT done

The fix is Go: **inert until an image is rebuilt and rolled.** The pass condition after the roll is
**positive, not an absence** — a bundle that renders `(*SagaCoordinator).applyResponseToState`'s
body (4,746 chars). A falling failure count is weaker: it also falls if nobody asks. Two follow-ups
recorded in `261` §8 and deliberately not folded in: `siblingSignatures` renders bare names while
the index section renders qualified ones (two grammars for one symbol, in one bundle), and the
per-file 10-signature cap that hid the three functions this run needed behind "+79 more".

---

## 2026-08-12 (later) — council APPROVED round 1, and the one medium objection found a producer I had missed

**Verdict: APPROVED**, `6b0cc25b-1368-4fe2-87f0-bb3aa87019c0`, submitted 14:26Z, decided **14:32Z —
six minutes.** 12 seats reviewed, 4 abstained. *"approved with 1 advisory objection(s) — none
high-severity"*. The `architecture` seat signalled `point_fix` and said so explicitly: *"existing
architecture being made to work as designed, not new architecture being added under cover of a
fix"* — which settles the RFC question I raised in risk 2 as a No.

**Three objections were answerable only by going and looking. Two came out my way; one did not.**

### 1. `prior_art_librarian`, MEDIUM — "you cite SplitSymbol as precedent, then don't use it"

Verbatim: *"If SplitSymbol already normalises `(*T).M` forms, this edit is an unacknowledged rebuild
of named existing machinery; if it does not, the rationale's own precedent claim is misleading."*

**Neither horn holds, and the seat was right that my phrasing invited the reading.** The two
functions split on **different delimiters**:

```go
func SplitSymbol(symbol string)  // strings.LastIndex(symbol, ":")  -> "path" / "Name"
func splitReceiver(name string)  // strings.LastIndex(name, ".")    -> "Recv" / "Method"
```

`SplitSymbol` never sees a receiver — it hands `(*T).M` through **whole** as the name part. So it
does not do this job and cannot be reused for it. My precedent claim was about *where the grammar
lives* (this file owns it, so add the second rule here rather than at a call site), not about
`SplitSymbol` performing the transformation. I should have written it that way.

### 2. `guardian` + `reuse_agent`, low — "the single-call-site claim rests on your own grep"

Fair, and CLAUDE.md's own rule. Re-verified **at the code index** rather than by grep:

```sql
SELECT path, symbol FROM code_symbols WHERE body LIKE '%ReadSymbolBody(%'
  AND path NOT LIKE 'docs/%' AND path NOT LIKE '%_test.go';
--  internal/analysis/symbolbody.go                    | ReadSymbolBody               (the definition)
--  …/actions/diagnose_assemble_bundle_action.go       | DiagnoseAssembleBundleAction (the only caller)
```

Confirmed: one live caller. Two independent methods agree.

### 3. `prior_art_librarian`, low — "'no producer emits the dotted form' audits only two producers"

**This one was RIGHT, and it found a third producer that makes the defect worse than I described.**
`symbolbody.go`'s own header names three (`scopeFromCodeResults`, `resolveScopeEntries`, the
landmine-verifier's `derive_checks` prompt) and I had audited two. The third is
`diagnose_route_action.go:651`, and reading it changes the picture:

- `knownScopeIdentities` (`:541`) builds `knownSyms` from the analyser Output's `functions` and
  `types` — as **BARE** names, `syms[path+":"+name] = true`.
- So an entry in the **index** spelling — the spelling the bundle *taught the model* — is **not
  "known"**, and falls through to the fuzzy resolver.
- The fuzzy resolver reads `code_symbols` (the file's own comment: *"an embedding HTTP call to
  ollama-adapter PLUS a code_symbols read, per FUZZY next_scope entry"*) and re-emits
  `add(path + ":" + symbol)` at `:700` — **the index spelling again**.
- It then logs `"diagnose_route: resolved fuzzy scope entry"`. **It reports success.**

> **So the harness's own enrichment step took a scope entry, "resolved" it into the one spelling the
> body reader could not read, and called that a resolution.** Pre-fix, a model that wrote the bare
> name — which would have worked — could have it rewritten into the failing form by the very step
> whose stated contract is *"no worse than not resolving"*.

**This strengthens the design rather than undermining it.** With three producers of the failing
spelling and only two of them ours, fixing at any producer was never going to be enough; the reader
is the only place that sees all three. That is now measured rather than argued.

**A fourth instance of the same drift, found in passing:** `knownScopeIdentities` iterates
`{"functions", "types"}` and **not `values`** — the identical omission `spanOf` had. Post-fix its
only cost is a wasted fuzzy search and an embedding call per var/const entry, because the body now
resolves either way. Recorded as a follow-up in `261` §8, not fixed here — it is a performance
nuisance, not a lost body, and I am not widening a council-approved scope after the verdict.

### The two cosmetic objections, and what I did with them

- **`editquality`** — *"'5 of 7 cases' doesn't reconcile with the 5 positive cases in the sketch."*
  Correct arithmetic, sketch elision: the test has 7 positive cases and the sketch listed 5. Same
  class as last round's `siteID`/`domain` objection — a sketch is not the code. No change.
- **`editquality`** — the doc-comment edit *"should not be counted as covering any mechanism on its
  own."* Agreed, and it was listed as its own edit precisely so nobody counts it as one. Left in:
  the old comment named the spelling **no producer emits**, so a reader trusting it would write the
  form that used to fail. That is the comment that caused this, and correcting it is not padding.
- **`bug_historian` / `debug_historian`** both asked for a precedent check against council history
  for `symbolbody.go`. Not run — noted as owed in `261`. The two prior council rounds on this file
  (`163`, `189`) are already cited in the submission and both are *reuse* rulings, not this defect.

**The trailer stays `Council-Submitted:`.** `6911c2da4` was committed before the verdict landed, and
CLAUDE.md is explicit that `098` resolves the correlation at report time and credits it
automatically — forward-only forbids an amend, and writing `Council-Reviewed:` onto a later,
unrelated commit would be a false join. The verdict has now been **read and acted on**, which is the
part that was actually owed.
