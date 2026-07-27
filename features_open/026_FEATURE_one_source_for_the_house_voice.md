# 026 — The house voice needs ONE source, not a copy per content agent

**Filed** 2026-07-27 by the `gemini_content_provider` workstream, at the moment a
second copy was created · **Status** OPEN, unowned · **Blocks nothing**, but every
day it stays open is a day the two copies can drift

---

## What happened

The owner directed (2026-07-27) that the house Voice & Style rules become **the
default for all content, unless overridden**. Applying that revealed two things.

**1. It was never the default for anything except site pages.** Only
`page-content-writer` carried the block. Measured over 30 days:

| agent | LLM calls |
|---|---|
| `page-content-writer` | **2,330** |
| `content-gap-planner` | 9 |
| `content-reviewer` | 2 |
| the other 15 `content-creator-*` / `content-writer` agent definitions | **0** |

So "all content" is not eighteen agents. It is `page-content-writer` (99.5% of
content LLM calls) plus `content-creator-agent`, the blog-and-social **service**,
which does not write to `llm_call_log` at all and so appears nowhere in that table
— and which had **no house style whatsoever** until today.

**Do not "fix" the fifteen dormant agents by pasting the block into them.** Fifteen
more copies of a contract nobody exercises is the drift class this repo files bugs
about, and it would make this feature harder, not easier.

**2. Applying it to content-creator created a second copy.** The rules now exist as:

- a literal inside `page-content-writer`'s `prompt_template` (a DB row), and
- `defaultVoiceStyleBlock`, a Go `const` in
  `internal/agents/contentcreator/agent.go`.

Two hand-maintained copies of one contract, in two different substrates, changed by
two different mechanisms (a guarded SQL patch, live immediately; a Go edit needing a
build and a roll). That is precisely the shape CLAUDE.md describes for the council
roster — *"Two hand-maintained rosters that must stay identical is exactly the drift
class this council reviews for"* — which was solved there by a mirror script, not by
discipline.

**Until this is fixed: change both, or neither.** Both sites carry a
`[KNOWN DUPLICATION]` comment pointing here.

## Why this matters more than it looks

The block is not decoration. It is the single biggest measured lever on output
quality in this workstream — bigger than the choice of model. Same prompt, five runs
each, `gemini-pro-latest`:

- rewriting one bullet moved mean sentence length 7.6 → 12.1 words and length
  422 → 637 chars;
- the model comparison found filler and em dashes at **zero for all four models**,
  i.e. the block was doing that work, not the model.

A rule that drifts between site copy and blog copy therefore produces two different
house voices on the same estate, and the drift is invisible until someone reads both.

## What to build

**One canonical source, read by every content producer, overridable per agent.**

Candidate 1 — **a row both consumers read.** Store the block once (a
`platform_defaults` row, or a dedicated `agent_definitions` entry of a `style`
type). `content-creator` already holds a `clients_db` connection; the chassis
obviously does. Per-agent override stays exactly as now:
`core_logic.voice_style_block` present-and-non-empty wins, present-and-empty means
explicitly off, absent means inherit. **Prefer this one** — it makes the bad state
(two copies) unrepresentable rather than merely discouraged.

Candidate 2 — **inject at render time.** Have the chassis substitute the canonical
block into any `prompt_template` containing a `{{.voice_style}}` placeholder. Keeps
prompts readable and removes the literal from the DB row, but only covers agents
whose prompts go through `RenderPromptTemplate` — `content-creator` builds its
prompt in Go and would still need candidate 1's read.

Candidate 3 — **a mirror script**, the `099_SYNC_gate_roster.py` pattern: one source
of record, a script that writes both sites, a dry run by default. Weakest, because
it still permits the drift and only detects it when someone runs the script.

## How to verify

Change the canonical source, then confirm a fresh generation from **both**
`page-content-writer` and `content-creator-agent` reflects it — not just that the
row changed. The 2026-07-27 measurements in the workstream NOTES give a baseline to
diff against: em dashes 0, negative frames 0, mean words/sentence ~12.

## Pointers

`docs/agent_docs/docs024_key_docs_latest/gemini_content_provider/` (the v4 rewrite
and the measurements) · `travelling_docs/pitch_pdf_source/REVERSE_ENGINEERED_STYLE_PROMPT_v3.md`
(the owner-refined source the block is distilled from) ·
`APPLY_voice_v4_page_content_writer.sql` (how the DB copy is patched, with its
self-verifying rollback)
