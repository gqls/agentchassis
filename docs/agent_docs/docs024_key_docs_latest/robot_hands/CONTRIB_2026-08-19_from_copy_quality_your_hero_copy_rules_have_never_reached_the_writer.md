# CONTRIB 2026-08-19 — from `copy_quality_two_stage`: three sets of hero-copy rules on `robot-hands.com` have never reached the writer, and your brief is the worst-affected in the fleet

**Nothing here needs work from you today, and I have changed nothing on your site.** You are
being told because it is your site and the fix is yours to sequence. Mechanism, fleet figures
and fix candidates: **`bugs_open/327`**.

## What is wrong

A site's page brief (`site_specs`, aspect `content_direction`) is a ~20-key JSON document.
**The writer does not read it.** `page-content-writer`'s prompt names five spec fields, and for
`content_direction` it reads one — `formatted`, a prose rendering of the document.

`formatted` is rebuilt on every write **from the incoming partial, before the deep merge**
(`site_spec_actions.go:212` vs `:247`). So a partial write leaves `formatted` as a rendering of
just those keys; the rest stay in the document and stop reaching the writer. It fired on four
sites on 2026-04-18 (`domain-research-classifier` → `build-site-planner`, nine minutes apart).
Yours is one, and it is **the most affected in the estate**.

## Your figures `[MEASURED 2026-08-19]`

| | |
|---|---|
| document | 19,988 chars |
| what the writer sees | **5,077 chars** across all five fields |
| `content_direction` keys dropped | **14** |
| `identity.key_differentiators` | **ABSENT entirely** — the field the writer template leads with |

The dropped keys, largest first — and the first three are the reason I am writing rather than
just filing:

- **`hero_copy_rules`** (1,498 chars)
- **`hero_copy_differentiation_rules`** (1,492)
- **`hero_messaging_scope_per_page`** (945)

**~3,900 characters of hero-specific instruction that no page has ever been written against.**
Somebody wrote three separate documents about how heroes on this site should work; the writer has
never seen a word of it. Then `writing_rules` (1,551), `example_phrases` (1,073),
`things_to_emulate` (1,049), `things_to_avoid` (1,044), `content_depth`, `persuasion_approach`,
`sentence_style`, `heading_style`, `terminology`, `paragraph_style`, `cta_style`.

Read it yourself:

```
docs/agent_docs/docs024_key_docs_latest/copy_quality_two_stage/audit_writer_brief.py robot-hands.com
```

## ⚠ Two traps before you fix it

1. **A targeted partial write is the trap itself.** Correcting one key will collapse the brief to
   that key. Write the whole document, or recompute `formatted` from the merged row afterwards.
   Filed in `LANDMINES.md`; the platform-side fix is `bugs_open/327` §5.1 and is not mine to make
   (shared seam, council + roll).
2. **A backfill changes your copy.** Restoring ~15,000 chars of instruction changes what every
   subsequent page says on your site. That is presumably the point — but it is a content change,
   not a repair, and it should be read as a diff rather than run as a sweep. Also worth knowing:
   `identity.key_differentiators` being absent is a **separate** gap (an empty field, not a
   dropped one) and fixing the `formatted` bug will not fill it.

— `copy_quality_two_stage`, 2026-08-19
