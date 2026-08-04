# NOTES — `bugs_open/173` per-substep `continue_on_error`

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-08-04 — picking the bug, and a measurement that could not have come out otherwise

**Task:** take the next `bugs_open/` case not already being worked, fix it robustly at the
framework level rather than the individual case.

### MISSTEP 1 — my first ownership measurement was non-disconfirmable

To find an unclaimed bug I grepped every live session transcript for `bugs_open/NNN` and
counted **mentions**. It returned "mentioned in 4–16 active transcripts" for *every*
candidate, including ones nobody was touching.

**Why it was worthless:** every session runs `ls bugs_open/` at some point, so all 54
numbers appear in most transcripts. The measurement **could not have returned a low number
for any open bug**, which is the disqualifying property — dated and run is not the same as
disconfirmable (`a-count-you-kept-is-not-a-census`, and the 2026-08-03 pair in
`WRONG_CALLS.md`).

**What replaced it, and it worked:** grep the transcripts for tool-call *file paths* —
`"file_path":"…/bugs_open/NNN…"` — i.e. which bug files a session actually **read or
edited**. That returns 26 numbers out of 54 and discriminates sharply. Logged in
`WRONG_CALLS.md`.

### MISSTEP 2 — `who-owns.py`'s verdict is dominated by the FILING lane

`./scripts/who-owns.py 173` returns **OWNED or recently active**, naming
`bugfix_165_reconciliation_deletes`. That lane **filed** 173 as a follow-on *for someone
else* and has since closed (`bugs_closed/165`; its own summary is titled *"all four guarded
and the lane closes"*). The tool reads commits, so "the lane that wrote the bug file" and
"the lane fixing it" look identical.

Not a defect in the tool — its own output separates "likely OWNING workstream" from "also
cites it", and its verdict is deliberately conservative. But **its verdict line alone is not
an answer**; read the named lane's state before believing it.

### First pick, abandoned: 093

I first picked `bugs_open/093` (the stat audit's one guarded call site). It is a genuinely
good structural bug — but reading it to the end showed its own 2026-07-27 triage says the
fix **is already built and live** and *"093 is not a code task any more. It is blocked on
`bugs_open/083`"* (the check ships, but `improvement-sweep` is disabled so nothing ever runs
it). And 083 **is** being actively read by a live session.

**Cost:** ~5 minutes. **What would have avoided it:** reading the bug file's *last* dated
section before its first. On this estate a bug file's header is frequently stale relative to
its own updates — 093's header still says "OPEN, not started" while four later sections
record it built, shipped and pod-verified.

### Settled on 173

- Its own header says **OPEN, UNOWNED**.
- No live session has read the file.
- The filing lane is closed.
- It is a missing degree of freedom in the orchestration engine — framework-level by
  construction, which is what the task asked for.
- Filed at the direction of the council's `architecture` and `constitution` seats, so the
  shape wanted is already on record.

---

## 2026-08-04 — verifying the bug is still valid, and one finding the file does not contain

The tree moves under us, so validity first. `loop_expansion_handler.go:157` still reads
`injectedStep.Config["continue_on_error"] = continueOnError`, unconditional. **Still valid.**

**New finding, not in `173`:** line 104 deep-clones the substep's config into the injected
step, so a substep's own `continue_on_error` **is already present** in `clonedConfig` — and
line 157 then overwrites it. The key is therefore **declared-and-inert today**: an author can
write it, the config-key audit will not object (it is in `frameworkStepConfigKeys`), nothing
warns, and the value is silently discarded. Registered as a landmine.

### Four checks done by reading, so the fix does not guess

1. `models.DecodeSubWorkflowStep` does `step.Config = config` verbatim → a substep's config
   key survives the decoder; **no** `SubWorkflowStepFields`/`models.Step` change (which
   WFA-003's registered landmine explicitly warns against).
2. `continue_on_error` is already in `datahelpers.frameworkStepConfigKeys:144` → no audit
   change, and no audit silenced.
3. All three coordinator decision sites (`coordinator.go:907`, `:3239`, `:3345`) go through
   the single `shouldContinueLoopOnError` → resolving once at expansion covers all three.
   **Checked rather than assumed**, because "one guarded call site" is exactly the family
   `bugs_open/093` exists for.
4. `loop_metadata["continue_on_error"]` (line 85) is loop-level and must stay so —
   `skipToNextLoopIteration` deliberately does not rely on that shared key.

### The blast-radius measurement, with its positive control

`0` live substeps declare `config.continue_on_error`. **[MEASURED 2026-08-04]**

The control matters more than the result: the same query, grouped instead of filtered,
returns **18 loop steps / 79 substeps** (9 `true`, 1 `false`, 8 unset). So the predicate
demonstrably finds loops and substeps and a non-zero answer was reachable. Without that, "0"
is equally consistent with "my `s.value->>'action' = 'loop'` predicate matches nothing".

> **CORRECTION to `173`'s own contribution figures.** It recorded 20 loops (10 unset / 9 true
> / 1 false) on 2026-08-01. Today: **18** (8 unset / 9 true / 1 false). Two unset loops
> retired in between, consistent with `7a15c3a47` "retire(agents): three unused builders are
> out". The bug's conclusion is unaffected — all nine `true` loops are still fan-out/dispatch
> loops and none wraps a destructive reconciliation. Recorded because CLAUDE.md requires
> figures to be grounded before being repeated, and this one had drifted in three days.

### Residual, deliberately not fixed here

`loop_actions.go:66` reads the **loop-level** flag with a bare `.(bool)` assertion, so a
loop-level `"true"` (string) is silently ignored — the same declared-and-inert shape the fix
makes loud at substep level. Left alone: it is a separate contract on a separate line, and
widening it changes how 18 live loops parse their own config. **[UNMEASURED]** whether any
live loop declares it as a non-bool — worth a query before anyone touches it.

---

## 2026-08-04 — the bug's MOTIVATING loop no longer exists; the defect and its blast radius do

Owner ruling 2026-07-29 §3 says a shared mechanism's other consumers must be **told**, not
merely measured. So I went to name them — and found the headline one gone.

`173` opens on `multipage-website-builder` → `generate_pages_loop` → `extract_links`. That
agent is **deleted and inactive** as of today:

```
 type                      | is_active | snap | deleted | loop_steps
 multipage-website-builder | f         | f    | t       |          1   (x2 rows)
```

Consistent with `7a15c3a47` *"retire(agents): three unused builders are out"* — and it is the
same retirement that took the loop census from 20 (2026-08-01) to 18 (today).

**This does not weaken the bug; it relocates it.** The three loops named in `173`'s own
contribution section are all still live, all still `continue_on_error` **(unset)**, and all
still carry `save_page_sections` — the action that gained a completeness floor and can now
refuse:

| agent | loop | substep | action | loop `continue_on_error` |
|---|---|---|---|---|
| `pageflow-builder` | `build_pages_loop` | `save_sections` | `save_page_sections` | (unset) |
| `page-rebuild` | `build_pages_loop` | `save_sections` | `save_page_sections` | (unset) |
| `site-work-orchestrator` | `build_items_loop` | `save_sections` | `save_page_sections` | (unset) |

So the disproportion `173` describes — **one page's refusal fails an entire multi-page
build** — is live in three agents today, and these are the consumers to tell. The one lane
that would have been hardest to reach (the retired builder) needs telling by nobody.

> **CORRECTION, and it matters for anyone quoting this bug.** `173`'s § "Why it matters —
> the concrete case that produced this file" is now **historical**: the case it describes
> cannot recur, because the agent was retired. Quote the three rows above instead. The
> reasoning in the file is unaffected — it never depended on that agent, only illustrated
> itself with it. **[MEASURED 2026-08-04]**, query in the runbook.

**Why I checked at all**, since the temptation was to copy the four rows out of the bug file
into the council submission: a closed bug's scope-out expires
(`a-model-upgrade-can-invert-a-closed-bugs-premise`), and so does an open one's example. The
submission would have named a deleted agent as a live consumer, and the guardian seat would
have been right to ask whether I had looked at anything at all.
