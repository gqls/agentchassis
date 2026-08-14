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

---

## 2026-08-12 (evening) — the roll landed, and the fix is PROVEN. 261 CLOSED.

### Step 1 — live at the artefact, and my first control was worthless

`v1.0.1293`, both replicas, both stamped `git_commit 7a1887e3163af75ce5eb5c6cb67ba2c9be37d88e`.
`git merge-base --is-ancestor 6911c2da4 7a1887e3` → **YES**.

> **The control is the part worth reading.** My first attempt used a commit from earlier the same
> day. But the build was cut at 19:57 BST and **every commit I made that day precedes it**, so that
> check would have returned YES no matter what was true — it could not have come out false. The real
> control is a commit made **after** the build (`81c508bca`), which correctly returns NOT an
> ancestor. **A control drawn from the wrong side of the boundary is not a control**, and this is the
> second time in one day I have written a check that could only return the answer I wanted.

### Step 2 — the behavioural proof, and it is an A/B rather than a "looks better"

I re-fired the `090` with the **symptom text verbatim from the run that failed**. That was the whole
design of the test: same question in, so the loop assembles the same scope, so the only variable is
the harness. It worked — iteration 1 came back with the **identical 12-symbol scope**.

| | pre-fix `dbcc4259` iter 1 | post-fix `eddaf1af` iter 1 |
|---|---|---|
| in-scope symbols | 12 | **12 — identical list** |
| rendered with a body | 3 | **12** |
| `_(body unavailable …)_` | 9, **every one receiver-form** | **0** |
| `**This section is INCOMPLETE.**` | present | **absent** |

`(*SagaCoordinator).applyResponseToState` renders its real body from the `func` line inside a ```go
fence — 4,907 chars of block against the index's 4,746-char body, the difference being the heading
and the fence. All eight other previously-unreadable receiver-form symbols resolve too.

**The evidence is POSITIVE and that was deliberate.** The weak version of this test would have been
"the failure count fell" — which also falls if nobody asks. Writing the pass condition as *a body
that is THERE* before running it is what made the result mean something.

`bugs_open/261` → `bugs_closed/261` on the fixed-AND-live bar, **both paths named on the commit**
(`git mv` plus a single-path pathspec ships a copy and leaves the original at HEAD — the landmine
this repo already carries). Verified at HEAD with `git ls-tree`, which returned exactly one line.

### What is NOT claimed

Nothing about `bugs_open/236`'s own mechanism. That verdict is a separate question and its run was
still going as this was written. **236 is unblocked for the first time since it was filed — which
was the point of the fix, not a result of it.** Reading a `CONFIRMED` on 236 as vindication of this
fix, or an `UNVERIFIABLE` as its failure, would be the same conflation that produced the
predecessor handoff's wrong mechanism.

---

## 2026-08-12 (evening) — `bugs_open/267` fixed in the tree: four unconditional invitations, not two

Picked up 267 (the defect found *behind* 261) and implemented candidates 1 and 2. Ownership checked
first (`scripts/who-owns.py 267` → the only two commits are this lane's own filing and its §4b
measurement; nobody else in flight), and the working tree was clean on both target files.

### The filing's census was incomplete, and finding that out was the useful part

§2 named **two** places that advise a whole-file re-request. Grepping for the *shape* rather than
the two known strings found **four**. The two extra ones:

3. The coverage SUMMARY line — `"N did not fit … — re-request them singly in next_scope"` — said
   that for the whole omitted set, when it is true only of the members that would fit singly.
4. **`inScope[path]["*"] = true // whole file already included; no siblings to add`.** Not an advice
   string at all. That comment is false exactly when the whole file did **not** render for size —
   so the one file the model must sub-divide was also the one file whose symbol list the bundle
   suppressed.

**#4 is why candidate 1 alone would have been half a fix.** Refusing the request and then
withholding the map moves the dead end; it does not close it. I would not have found it by looking
for the string in the bug report, and I nearly didn't: it turned up because I asked "what does the
model do *next* after we refuse it?" rather than "where else does this sentence appear?".

### `SymbolSizes` computes sizes by CALLING `SliceLines`

The marker now names the largest symbols that would fit, with sizes. The obvious implementation is
prefix sums over the line offsets — faster, and wrong in the way that matters: it becomes a second
copy of `SliceLines`' `[start,end]`-inclusive convention that must be kept in step for ever. If the
two ever drift, the bundle advertises a size the cap will not honour, which is **this same bug with
an extra indirection**. So it calls `SliceLines` per symbol and pays one split of the file each
time. This runs only on the rare over-budget whole-file marker.

`TestSymbolSizes` asserts the **round trip** — every offered handle resolves through
`ReadSymbolBody`, and to a body of exactly the advertised length. Deriving the expected size from
the fixture text would have let the test and the code agree with each other while both were wrong
about the span convention. That is the same failure the README already records from yesterday (a
test asserting a style nothing produces), so it was in front of me while writing this one.

### Every guard mutation-verified, individually

Four conditionals, four mutations, run one at a time:

| mutation | tests that failed |
|---|---|
| `overCapAdvice` always returns the old sentence | the two over-cap tests, and only those |
| summary always says "re-request them singly" | the summary test, and only that |
| "+N more" always offers the bare path | the sibling test, and only that |
| an omitted whole file still treated as "already included" | the sibling-listing assertion, and only that |
| `SymbolSizes` emits bare method names | the collision assertion, and only that |

One at a time on purpose — all four at once would not distinguish a guard that is load-bearing from
one sitting in series behind another. None was in series.

### What I did NOT do, deliberately

`siblingSignatures` still renders methods with a **bare** name while the new marker renders them
canonically, so one bundle can now show two spellings for one method. I nearly folded it in — it is
one line, in a function I was already editing, and my own change makes the inconsistency visible.
It is `bugs_closed/261` §8.1, already recorded there as a deliberate non-fold, so widening this
change to cover it would have been me quietly re-deciding another file's call. Recorded in 267 §7c
instead, with the observation that the two-spellings-in-one-bundle effect is the strongest argument
yet for closing it.

### Checks that were worth running

- **Built and tested against `git archive HEAD` + only my five files**, not against this tree.
  HEAD had already moved to `d142fcd27` under me while I worked. A green local build proves nothing
  about what `make build-*` will produce, because the tree carries other sessions' WIP.
- **Eyeballed the rendered bundle text**, which is the actual deliverable — it is prose an LLM
  reads, and no assertion I wrote would have caught the one error I found that way: my new sentence
  said the file's remaining symbols "are listed under Same-file signatures below", but that section
  lists **functions only**, while my count spans functions, types and package-level values. A small
  false claim in most files. Narrowed the wording.

---

## 2026-08-13 — 267 verified live and closed; 269 fixed and measured

### The deploy verification worked, and by the method I had told the council was retired

Chassis rolled to `v1.0.1295`, pods up `13:53:19Z`. I checked at `14:00:23Z` — **7 minutes** — and the
`build provenance` stamp was **there**, cleanly, on the first try. All six of this lane's commits are
ancestors of the stamp `69612d692`, with a descendant commit as a must-be-absent control.

**So yesterday's WRONG_CALLS entry was itself an over-correction, and I have narrowed it.** Yesterday I
claimed the recipe worked and was wrong; my correction said it is *inoperative on agent-chassis*, and
that is also wrong. It is **time-limited**. The landmine measured 0 hits at 44 minutes; I got a clean
hit at 7. **Both errors are the same error**: one measurement of a time-dependent check, generalised
into a property of the service.

The discriminator was in the landmine the whole time and I did not carry it into my own note:
`kubectl logs <pod> | head -1`. If that shows a startup line, the log still reaches back and the stamp
is in range. I ran it this time *before* the grep, which is the only reason I trust the hit rather than
merely enjoying it. Landmine refined in place (its own precheck now leads instead of trailing);
WRONG_CALLS narrowed with the over-correction logged as a wrong call in its own right.

**Why the over-correction is worth its own line:** "it never works here" would have sent the next
reader to the `strings` / `/proc/1/exe` probe, which the same landmine documents as returning **absent
on a binary that genuinely contains your change** — because the binary carries one commit, the build
point, not its ancestors. My tidy correction pointed at a worse instrument than the one it condemned.

### §7d's behaviour witness returned 0, and its demand control is why that was readable

```
witness_new_code | bundles_since_roll | omissions_since_roll
               0 |                  0 |                    0
```

`bundles_since_roll = 0` — no diagnosis bundle has been assembled at all since the roll, so the
witness could not have fired. **Alone, that zero reads as "the fix did not ship".** I wrote the demand
control into §7d yesterday on principle and it paid on its first outing. 267 is therefore closed on
the **stamp**, with the behaviour witness explicitly recorded as pending traffic — not quietly counted
as a pass.

### 269: the fix, and what the measurement actually says

Three halves, each mutation-verified alone (revert the rendered handle → all three tests fail; disable
the canonical de-dup → only the §2a test; suppress every same-named method → only the exactness test).

**The half I nearly got wrong is the third.** Suppressing every method that shares a bare name is the
conservative-looking choice, and it **hides a sibling the model has not seen** — inverting the purpose
of a section that exists to show what retrieval missed. A bare scope entry resolves the way `spanOf`
does, first match in `fi.Functions` order, so exactly one of a colliding pair is in scope. Tracking
that *observes* the order rather than re-deciding it, which also avoids adding a second copy of the
resolution rule.

**The measurement, control first.** The query derives a bare name by stripping a `(Recv).` prefix. If
`code_symbols` did not store the parenthesised spelling, the strip would be a no-op and every answer
would still look like an answer. Of **1,175** method rows, **1,175** are parenthesised and **0** are
not. So the strip does real work.

**17 collision groups, 48 methods, 4.1%** — and 4.1% is the **floor of the harm, not its rate**. In an
n-way group a bare handle resolves to the first and is wrong for the other n−1; the worst groups here
are **six-way** (`discovery_checks/check_integrity.go`, `Name` and `Run`), so a bare handle there is
wrong 5 times in 6.

**`pkg/diagnose/loop.go` is in the list** — `(Outcome).String` against `(Tier).String`. The diagnosis
loop's own source. A diagnosis *of the diagnosis loop*, which is exactly what 267 and 261 both were,
could have been handed the wrong `String` body with nothing in the bundle saying so.

**What the measurement does NOT say**, stated because the temptation to round it up is real: not that
48 wrong bodies were served. The population where the defect can fire, and the per-group odds.
Incidence is unmeasured — the sibling section's per-file cap means many were never listed at all.

### Two shell traps hit in one session, both from the same root

1. **An unquoted heredoc (`<<PY`) executed the backticks in my markdown.** I used it because I wanted
   `$OLD` interpolated; the price was every `` `commit` `` in the document body running as a command.
   Clean failure — nothing written, nothing committed — but only by luck. The fix is a **quoted**
   heredoc plus the variable passed through the environment (`OLD=… python3 script.py`).
2. **`git commit -F -` with a heredoc containing `Council-Submitted:.`** in prose tripped the trailer
   gate, which read the sentence as a trailer with the value `.`. Blocked the commit, which is the
   gate working. **Do not write the trailer tokens in prose**; refer to them descriptively.

Same root in both: **a commit message and a markdown document are DATA, and I kept handing them to
the shell as CODE.**

## 2026-08-14 — 269 verified live and CLOSED; the 236 re-run turned out never to have run, so it is dispatched now

**The roll happened overnight**: chassis `v1.0.1297`, pods 2026-08-13T22:29:19Z / 22:29:40Z.

**269 §9 verification, in order, with controls:**

- **Precheck failed as designed**: at ~9h pod age, `head -1` of `--tail=100000` was a worker line,
  so the provenance startup line was out of range — "not in range", not "unstamped".
- **Binary probe fallback**: build-point candidate `3b0ea20ffa84…` (HEAD at 22:10:43Z, last commit
  before the pods started) PRESENT in `/proc/1/exe` on BOTH replicas; control `dffbc75e45…`
  (committed 07:37Z today, post-build) ABSENT on both. `git merge-base --is-ancestor a3fee59b8
  3b0ea20ff` → ancestor. The fix is in the binary.
- **Behaviour**: exactly one bundle since the roll (`38e53a03…`, 07:39:57Z today):
  `bare_method_handles=0`, and — the demand control that makes the zero readable — **12 sibling
  bullets in the method format, all canonical**. Pre-fix code rendered every method bare, so 12/12
  canonical could not be old output. The bullet-plus-signature format is unique to
  `siblingSignatures`, so this is not `SymbolSizes`' prose being miscounted.
- **The §9 collision clause is NOT satisfied**: that bundle scoped no §6b file. Recorded exactly in
  `bugs_closed/269` §11 — half 1 (canonical rendering) live-exercised; halves 2/3 (dedup,
  first-wins) test-proven only. Owner approved closing on that split this morning.

**Closed and moved**: commit `034c421d2`, both paths on the commit, one line at HEAD confirmed.
Register `DIAG-043` and `bugs_closed/267` §4b updated in `171335d6f` (`cap_only` 6 all-time, 0 new
since 267 went live — right direction, thin demand: 1 bundle).

**236 (hero/logo): the "cheapest next move" had never actually run.** Today's 090 coverage check
refused, and reading the blocking item was the finding: `686f58a1…` (2026-08-12 19:49Z) re-used the
**broad four-function seed**, not the narrow one, then **failed verdictless** at 20:30Z — five
bundles under `36bd1b42…` (19:56–20:33Z, the last one after the failure stamp), no verdict, run row
pruned, cause `[UNVERIFIED]`. Also learned: **verdicts are never in `diagnosis_artifacts`** — all
three runs on this question hold bundles only there; a verdict must be read from the run's
orchestration row within retention.

**Dispatched the narrow run** (FORCE=1 after reading the coverage hit — our own failed item, no
live session): intake `7daa0c43…`, **RUN `23f1cf9a-2e33-43a3-9b33-d18adbbe5c55`** (the artifact
key), seed ONLY `persistAwaitingStateWithRetry`. Full record in the 236 file's 2026-08-14
contribution. **Verdict pending at the time of this note.**

## 2026-08-14 (later) — the verdict landed in half an hour, and the mechanism 236 needed is CONFIRMED

Run `23f1cf9a…` (three iterations, ~08:00–08:05Z, verdict on the row by 08:07Z — the one-function
seed left no room to wander). Label `UNVERIFIABLE`, content decisive: **the full body of
`persistAwaitingStateWithRetry` copies only `AwaitedRequests`, `Status`, `LastActivity`; the
fragment that had made §5b `[CONTESTED]` is an existence check on freshState's OWN CollectedData,
not a merge.** Five citations, both halves of the copy logic.

The two residual checks the verdict enumerated, run first-hand within the hour (declared
substitution per the 07-31 ruling — the loop named them, I ran them):
- **its own two data_requests, untruncated, on its own two still-parked children**: awaited step's
  key ABSENT from `collected_data`, every completed step's key present — both rows;
- **the ordering read** it asked for: `processActionResult` calls `storeActionResult` at `:1795`
  (in-memory write of the current step's key, `:1873`/`:1877`) BEFORE `processAwaitResponse` at
  `:1839`. The "key only written at response time" innocent reading is eliminated.

Recorded in `bugs_open/236` (final contribution) with the honest split: mechanism CONFIRMED,
hero/logo incidence `[INFERRED]` (those rows are pruned; item 2's capture witnesses the next one).
**236 stays open — the fix is RFC_012 `(a)`/`(a′)`, an owner decision.** Commit `4540f6344`.

Note for the next reader: the verdict itself reported its `data_requests` came back TRUNCATED with
`…` before the deciding key — the loop cannot currently see deep `collected_data` keys on wide
rows. That truncation is why its label stayed UNVERIFIABLE while its text answered the question.

## 2026-08-14 (afternoon) — v1.0.1298 rolled; everything this lane shipped still holds on 10 bundles of demand

New chassis build `v1.0.1298`, pods 08:58:03Z / 08:58:25Z. Startup line already rotated at ~4.75h
(precheck failed as designed); binary probe: build point **`bc39e7bf547e9d5db07c92085be85c6874654774`**
(08:14:58Z) PRESENT on both replicas, flipped-last-char control absent on both. `a3fee59b8` remains
an ancestor — 269's fix is aboard trivially. Nothing this lane owns was gated on this roll.

Re-read over ALL bundles since the first roll (22:29Z 08-13), now **10**:
- `bare_method_handles` **0 of 10** — 269's behaviour now rests on real demand, not one bundle;
- collision-file scoping **0 of 10** — halves 2/3's live witness still outstanding (expected: §6b
  files are rare targets);
- `cap_only` **0 new of 10** — 267 §4b updated in place.

**Next work for this lane** (nothing else is unblocked by us): `bugs_closed/261` §8 **follow-up 2** —
the per-file sibling cap (~10 signatures) hid the three functions a real run needed behind
`_(+79 more in this file …)_`; iteration 4's scope collapsed to five symbols, three of them copies
of `getMapKeys`. Same code path as 267/269 (`siblingSignatures`,
`diagnose_assemble_bundle_action.go`), so grep the LANDMINES entries on that footprint first.
Also open there: follow-up 3 (`knownScopeIdentities` omits `values`, cosmetic) and follow-up 4
(the precedent check two seats asked for, owed). Both owner decisions (RFC_012 for 236, RFC_027)
remain with the owner.

## 2026-08-14 (evening) — 261 §8 follow-up 2 taken: the dead-end marker now gives the names it demands (bugs_open/273)

Picked up the handoff-designated next work. Read 261 §8 + all four LANDMINES entries on the
`diagnose_assemble_bundle_action.go` footprint first, as instructed.

**The defect, sharpened during the read:** 267's dead-end marker ("the whole file exceeds the
budget. Name symbols individually.") is UNSATISFIABLE for the elided tail — the model can only
name a symbol it has been shown, and the elided ones are by construction the ones retrieval never
surfaced. Same class as 267 (an invitation the bundle's own arithmetic refutes), one layer down
(an instruction whose required input is withheld). Filed as **`bugs_open/273`**, with the OWNER
RULING 2026-07-31 declaration (first-hand substitution for a 090: mechanism read in code, harm
already witnessed live by 261's run `dbcc4259`, fix mutation-proven at unit level).

**Measured before designing** (the numbers that set the bound): `coordinator.go` 169,139 B vs
60,000 budget, 91 functions; complete canonical-handle tail after ~10 head lines = **2,715 chars**
(v3_site_actions 1,935; data_helpers 1,231) — generated from source, not estimated. So
`siblingDeadEndTailCap = 4000` covers every file in the repo today.

**The fix:** in exactly the `known && !fits` branch, `writeDeadEndTail` appends the elided
functions' CANONICAL handles compactly (269: bare = wrong body), bounded 4,000/file; past the
bound it counts the residual and names the `code_request kind "symbol"` remedy — never a silent
trail-off. **The tail is exempt from the section's global guard** — found by arithmetic before it
bit: counting it evicts the WHOLE section on the motivating one-file case (head == capChars share,
head+tail > capChars*5/4 → "further files omitted" replaces the model's only map of the file).
Could-fit/unknown branches byte-identical; no censused marker phrase introduced (checked against
the LANDMINES bundle-census entry; asserted by test).

**Mutation-proven:** old marker restored behind `if false` → the three positive tests FAIL, the
byte-identity pin passes both sides. Full related set green post-fix.

Council: **submitted**, corr `ba3f6047-a2e5-4ce6-ac0e-edf0bb88c4e3`
(`SUBMISSION_2026-08-14_sibling_dead_end_tail.json`; first attempt refused client-side —
`operation: "create"` is not in allowedFixOperations, a new file is `"add"`). Committing with
`Council-Submitted:` per this lane's trap 5, not waiting for the verdict.

Misstep to record: my first test draft asserted head-line handles as `` `canon` `` — but head
lines render `` `path:canon` ``, so the assertion would have failed on functions that were
correctly SHOWN. Caught before running by re-reading the render format; the committed assertion
is `Contains(got, canon)` = "shown somewhere", which both forms satisfy. Also left a placeholder
helper in the draft that would not compile; deleted before build.

Still open after this: follow-up 3 (`knownScopeIdentities` omits `values`, cosmetic), follow-up 4
is DONE (261 §9b ran the precedent check). 273 stays in `bugs_open/` until a chassis roll carries
it (fixed-AND-live bar); live verification recipe is 273 §5 — note its demand control: a zero is
evidence only if a dead-end file was actually scoped.
