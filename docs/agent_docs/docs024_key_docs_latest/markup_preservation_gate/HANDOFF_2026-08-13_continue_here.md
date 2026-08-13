# HANDOFF 2026-08-13 — markup preservation as a measurable property (from the `bugs_open/263` contribution)

**Read this first if you are picking up `scripts/class_count_delta.py` / concept register
ADO-041, or continuing the owner's standing instruction "take the next bug that isn't being
worked on".**

This thread did **not** fix 263 — another session did, mid-flight. What survives is a
measurement, a shared predicate, and two contributions. All of it is committed. Nothing is
in flight, so this handoff has no "resume the half-done thing" section.

---

## 1. What this thread is

One question, asked because three separate checks certified a page that had lost its
design: **how do you mechanically prove that a transformation did not delete a page's
layout?** The answer is narrower than it sounds — a class **set** diff, an **aggregate**
class-attribute count, and a **byte diff against a prediction** all fail at it, each for a
different and defensible-sounding reason. Only a **per-class count** works.

## 2. What is DONE and committed

| commit | what |
|---|---|
| `6427f157f` | `scripts/class_count_delta.py` — the shared predicate. Self-test passes **including its own induced failure**. |
| `adb4fd058` | Contribution into `bugs_open/263`: independent confirmation of the landed fix + the correction that the residual is **6 pages, not 1**. Plus the `WRONG_CALLS.md` entry for my empty-read misstep. |
| `450efe5d7` | Concept register **ADO-041** + index row. |
| `eef4a4800` | Marked my own 263 recommendation SUPERSEDED by the 08-13 owner direction. |
| `046dbabbe` | Contribution into `bugs_open/253`: the component floor's aggregate count can net to zero — measured, with the upgrade path named. |

The LANDMINE entry (`git show <ref>:<path>` from a subdirectory returns an empty document)
is at HEAD but **inside another session's commit `889a7c055`** — it was swept out of my
working tree before I could commit it. Nothing lost; only the attribution is wrong. It is
synced to `doc_notes` (`./scripts/landmines-sync.py --check` says `in sync`).

## 3. What is TRUE right now (verified this session, not carried forward)

- **263's fix is landed and correct** (`71fb31a03`, `keep_widget_wrapper` opt-in on the
  shared `split_ordered`). Independently re-measured with a separately written harness:
  pages dropping a layout class **13 → 6** of 21, and **0 pages freeze prose**. Two
  implementations agree on the same 6.
- **The residual 6** — `damage-checker`, `bridging-loan`, `equity-release`, `fee-analyser`,
  `rate-forecaster`, `simple` — are **one shape**: a panel that mixes the page's own copy
  with the widget's machinery. **Four of the six hold the page's `<h1>`.** Three of them
  (`fee-analyser`, `rate-forecaster`, `simple`) cannot be preserved by *any* descent rule:
  their `.container` holds nothing but the panel, so stopping above it freezes the page and
  descending dissolves it.
- **The owner's 08-13 direction supersedes the choice those six posed.** Track B2 puts the
  machinery in `html_template` and the copy in `input_schema` fields with the row
  **unlocked**, so an in-panel heading is neither frozen nor lost — it is a field. My
  measurement is what the dilemma was sized from; only my "wait and decide later"
  conclusion is retired.
- **Track B was 17 of 22 done** as of `aba366719`. Live page state when I last looked:
  15 `["ported-page"]`, 18 `["prose-0"]`, 5 `["prose-0","tool-1","prose-2"]`,
  2 `["prose-0","tool-1"]`, 1 `["tool-0"]`. **Re-read this before acting on it** — the lane
  is actively moving.
- **`countComponentClasses` still counts in aggregate**
  (`save_sections_component_floor.go:75-86`, verified 08-13). The contribution in
  `bugs_open/253` explains why that is a real blind spot and not a worry.

## 4. What is NOT done — the honest list

1. **`scripts/class_count_delta.py` has exactly one caller: its own CLI.** It is a predicate
   other lanes *can* import, not a gate anything runs. Do not read ADO-041 as "markup
   preservation is enforced" — it is not enforced anywhere. Three named consumers are
   waiting: `bugs_open/253` (the Go floor), `copy_quality_two_stage` (whose PLAN already
   demands "class attributes and component boundaries in == out, counted not eyeballed"),
   and `mortgagecalculator_couk_adoption` (port pages, no components, bytes in a bucket).
2. **The lane gate `gate_wrapper_parity.py` does not strip `<script>`/`<style>` before
   counting.** Inert today — verified zero in-scope pages carry a script inside
   `div#content` — and a live hazard the moment it is reused or a page gains one. Reported
   in `bugs_open/263`, deliberately not edited (another thread's lane).
3. **The LANDMINE entry needs a verifier pass** (`landmines-sync.py --apply` printed
   `NEEDS_VERIFICATION` for it, which is normal for a new entry).
4. **The council never reviewed any of this**, and cannot: `097_TRIGGER…sh:127` refuses a
   submission touching none of `platform/`, `internal/`, `pkg/`, and every file here is
   `docs/` or `scripts/`. If the ADO-041 predicate is ported into
   `evaluateComponentLoss`, **that** change is in scope and should go through the gate.

## 5. If you are continuing the owner's standing instruction (next unowned bug)

The instruction was: *find the next bug in `bugs_open/` that isn't being worked on in
another thread, research the docs and DB, plan a framework-preferring fix, check the
council, verify the bug is still valid, commit narrowly, keep docs updated, log missteps in
`WRONG_CALLS.md`.*

**The ownership check that actually works here** — `scripts/who-owns.py` reads *commits*, so
a session mid-fix is invisible to it. It said 263 was "OWNED or recently active" and the
transcripts said the lane had been idle 90 minutes; I took it, and another session landed a
complete fix two hours later. **Do both checks and then accept the residual risk:**

```bash
python3 scripts/who-owns.py <number|slug>
cd /home/ant/.claude/projects/-home-ant-projects-agentchassis && \
  for f in $(find . -name '*.jsonl' -mmin -240); do \
    n=$(grep -c '<bug number or key symbol>' "$f"); [ "$n" -gt 0 ] && echo "$n $f"; done
```
Timestamps inside the `.jsonl` are **UTC**; the shell clock is BST. A session whose last
entry is "an hour ago" may be live.

**If it happens again, the move that paid off was: stop, and re-measure their fix with your
own harness.** That is what turned a wasted plan into the finding that the residual was six
pages rather than one.

## 6. Traps this thread hit, so you don't

- **`git show <ref>:<path>` is repo-root-relative.** From `~/projects/sites/<domain>/` it
  fails, prints to stderr, returns **empty stdout**, and every count over it reads as a
  clean zero. `2>&1 | wc -c` measures the *error message* (224 bytes reads like a small
  file). Assert non-empty at the point of read; print the count of successful reads. Full
  entry in `LANDMINES.md`.
- **An empty input always says "no problem here."** The error has a direction, and it points
  at the answer you were hoping for.
- **`kubectl logs -l app=agent-chassis --tail=3000 | grep 'build provenance'`** returned
  1.5 MB of unrelated logs on a busy service. The startup line scrolls; per `LANDMINES.md`
  use the binary probe with a known-present and known-absent control instead.
- **A pathspec commit does not protect your uncommitted work from another session's broad
  commit.** Mine was swept mid-task. Append, then commit *immediately*.

## 7. Where the durable records are

`bugs_open/263` (contribution + superseded marker) · `bugs_open/253` (aggregate-floor
finding) · `WRONG_CALLS.md` 2026-08-12 (the empty read) · `LANDMINES.md` (the `git show`
trap) · concept register `ADO-041` in `register/adoption-pipeline.md` + index row.
