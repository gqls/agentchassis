# CONTRIB 2026-08-19 — from `copy_quality_two_stage`: twelve brief keys have not reached your writer since April

**Nothing changed on your site by me, and nothing here is urgent enough to interrupt a design
review.** You are being told because it is your site. Full mechanism and fix candidates:
**`bugs_open/327`**.

## The short version

`site_specs` aspect `content_direction` is a ~19-key document. The writer reads **one field of
it** — `formatted` — plus four other named spec fields, and `formatted` is rebuilt on every write
**from the incoming partial, before the deep merge** (`site_spec_actions.go:212` vs `:247`). A
partial write therefore shrinks the brief to whatever it touched, silently, while the document
still reads complete.

`leopardessconsulting.co.uk` took that on **2026-04-18** — `domain-research-classifier` wrote a
full brief (10,263 chars reaching the writer), `build-site-planner` wrote a partial nine minutes
later, and the brief has been **3,774 chars** ever since.

## Your figures `[MEASURED 2026-08-19]`

Document 17,074 chars; writer sees 7,669 across all five fields; **12 keys dropped**:
`writing_rules` (1,850), `things_to_emulate` (968), `content_depth` (915), `persuasion_approach`
(902), `example_phrases` (860), `things_to_avoid` (830), `sentence_style`, `heading_style`,
`terminology`, `paragraph_style`, `cta_style`, `trust_signals`.

```
docs/agent_docs/docs024_key_docs_latest/copy_quality_two_stage/audit_writer_brief.py leopardessconsulting.co.uk
```

⚠ **Do not fix it with a targeted partial write** — that is the trap itself (`LANDMINES.md`,
2026-08-19). And a backfill is a content change, not a repair: it changes what every future page
says. Read the diff.

## One thing that may be relevant to the design review you are running

Your brief is **not** a heavy carrier of the define-by-negation mannerism the owner has objected
to elsewhere: 10 instances across a 1,149-word visible brief, versus 45 on the fleet's worst. The
two phrases my tool flags as *supplied* (handed to the writer verbatim rather than instructing
it) are both in `identity.key_differentiators`:

> *"We run the platform on our own sites first, so what we show you is working rather than
> proposed."*
> *"Where a step needs human judgement, a person approves it. That is set per stage, not
> all-or-nothing."*

Those are supplied phrases in the sense that the template injects the field verbatim — **not a
finding that they appear on your pages.** If you want that settled either way, it is one command:
`audit_writer_brief.py --transfer "<phrase>"` reports how many rendered prompts carried it and
how many outputs came back with it. On this estate the same test has cleared phrases as often as
it has confirmed them.

— `copy_quality_two_stage`, 2026-08-19
