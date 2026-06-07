# Runbook — thin-slice bundle pipeline

How to run the thin slice today. Three files: `analyser.go`, `assembler.go`, `thin_slice_constitution.md`. Go toolchain on the machine with the repo.

---

## The pipeline

```
# 1. Analyse the repo once (re-run when code changes meaningfully).
go run analyser.go /path/to/repo > analysis.json

# 2. Assemble a bundle for one task.
go run assembler.go \
  -analysis analysis.json \
  -root /path/to/repo \
  -constitution thin_slice_constitution.md \
  -task "one-sentence task" \
  -step implementation \
  -scope internal/foo/handler.go:HandleBuild \
  -scope internal/foo/store.go \
  -doc docs/016_debugging_guide.md \
  -schema schema.txt \
  > bundle.md
```

Then paste `bundle.md` into a chat (or feed it to the API — see "large bundles" below).

### Flags

- `-analysis` — the analyser JSON. Required.
- `-root` — repo root, so the assembler can read full bodies. Required.
- `-constitution` — the flat constitution file. Required.
- `-task` — one sentence. Required.
- `-scope` — repeatable. A whole file (`path.go`) or one symbol (`path.go:Name`). At least one required.
- `-step` — `framing` | `implementation` | `debug`.
  - `implementation` / `debug`: in-scope code shown in full.
  - `framing`: in-scope shown as signatures (intent over detail) — use when the brief still needs expanding.
  - `debug`: adds a runtime-evidence placeholder (no run trace in the thin slice yet).
- `-doc` — repeatable. Any authored doc to paste in verbatim: a debug guide for a `debug` task, a 003 contract section when the task touches it. This is the manual stand-in for the matched-standards layer the real tool will pull in automatically.
- `-schema` — optional. A text file of the relevant `\d` output; include when the task touches the database.
- `-max-neighbour` — cap on neighbourhood signatures per package (default 60).

---

## Large bundles — how to feed them to Claude

A bundle's size matters, and bigger is not better. Rough conversion: ~4 characters per token, so ~1.5 MB of text ≈ ~375K tokens and ~4 MB ≈ ~1M tokens.

**Context-window facts (verified against the Claude docs, June 2026):**
- Every current Claude model has a **200K-token** context window.
- **Opus 4.x and Sonnet 4.x support a 1M-token window in beta**, gated behind the `context-1m-2025-08-07` beta header and restricted to higher usage tiers (≈ tier 4 / custom limits).
- Requests **over 200K tokens are billed at premium rates** (≈ 2× input, 1.5× output).
- **Context rot is real:** models do not use a full 1M window evenly — accuracy and recall degrade as the window fills. "It fit" is not "it was used well."

**What this means for us:**
- A ~1.5 MB bundle (~375K tokens) already exceeds the standard 200K window — it needs the 1M beta and is premium-priced. A 4 MB blob (~1M tokens) sits at the ceiling and is squarely in the context-rot zone.
- So **a large bundle is a smell, not a goal.** The fix is to include less, not to buy a bigger window. This is the main reason to build the call graph next (below): include the functions actually involved, not whole packages/directories. It improves *quality* (less rot), not just cost.

**Three ways to feed a bundle:**
1. **Paste into a chat** — fine for small bundles (well under the window). Simplest.
2. **A Project (claude.ai)** — attach the bundle (or the larger compiled context) as project knowledge. This is the home for the bigger "project files" scale (the 4 MB compiled context goes here, not in a single chat message).
3. **The API** — send the bundle as message content (or document blocks). For a large, stable bundle reused across a session, use **prompt caching**: cache writes cost ~25% over base input, cache reads ~10% — so a big stable prefix becomes nearly free to reuse turn to turn. Cache the stable parts (constitution, reference docs, code context); leave the task/question uncached. For >200K-token bundles, add the 1M beta header (and you need the tier for it).

Rule of thumb: aim to keep a working bundle **under ~200K tokens (~800 KB of text)** so it fits the standard window and stays in the zone where the model actually uses it. If you can't, that's the signal the selection is too broad.

---

## Fuzzy tasks need the brief expanded before targets can be picked

Some tasks (e.g. "improve site imagery") are under-specified — you can't pick in-scope files until the brief is consolidated into something concrete. For those, the first pass is a **framing bundle**: `-step framing`, scope the relevant area at signature level, `-doc` the relevant existing plans/assessments, and use it to expand the brief into a spec. Once the spec names the real targets, assemble an `implementation` bundle. (This is the framing-vs-implementation altitude split, done by hand for now.)

---

## Known limits of the thin slice (so the test is judged fairly)

- **Scope is manual.** You name the files/symbols; there's no automatic "given this task, find the target" yet.
- **Neighbourhood is the whole in-scope package as signatures** (capped), not a precise call-graph slice. Coarse on large packages.
- **Schema is hand-fed.**
- **No runtime evidence.** `debug` shows a placeholder; there's no `orchestration_id`-correlated run trace yet.
- **No bundle record.** Nothing logs what went in — fine while you're pasting it by hand and can see it; it returns as provenance once assembly is automated.

---

## Next improvements (in likely order)

1. **Call graph** — resolve callers/callees of the in-scope symbols so the neighbourhood is precise instead of whole-package. This is what shrinks large bundles and improves quality (less context rot). Biggest single win.
2. **Liability / compliance review** — a review concern (and possibly an agent) for features whose generated output could create legal exposure (claims, advice, regulated content). Folds into the standards layer and the review step.
3. **Schema extractor** — pull the relevant `\d` automatically from the live DB instead of hand-feeding.
4. **Target resolution** — from a task description, propose the in-scope set (so `-scope` isn't fully manual).
