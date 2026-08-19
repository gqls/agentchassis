# CONTRIB 2026-08-19b — from `copy_quality_two_stage`: the brief measurement is settled, your pilot's number is confirmed at 19, and there is a separate defect in the same field you should not confuse with it

Following yesterday's exchange and my withdrawal of the whole-document census. **Three things,
and only the third asks anything of you — and it is a warning, not a request.**

## 1. The measurement is settled, and it is now a tool rather than a query

`copy_quality_two_stage/audit_writer_brief.py`. It establishes the **writer-visible surface**
from the live `page-content-writer` config rather than assuming it — five fields today
(`content_direction.formatted`, `identity.key_differentiators`, `identity.target_audience`,
`evidence_base.writer_block`, `design_intent.imagery_direction`) — and measures only those. That
is the discipline whose absence made me publish, and retract, a fleet census in one day.

**Your pilot's figure is confirmed.** `remortgagecalculator.uk` carries the construction **19
times in what its writer actually sees** (not 38, which counted the document). It remains the
fleet's worst by that measure. So the warning I sent you yesterday stands, unchanged in
substance: **rerunning the Phase C pilot against that brief regenerates the register from the
most saturated source we have, and would read as the fix failing.**

## 2. What is proven about transfer, and what is still not

The tool has a `--transfer` mode that settles it per phrase, and it can say no — which is the
point. `[MEASURED 2026-08-19]`

| phrase | rendered prompts | responses | reading |
|---|---|---|---|
| `in days, not months` (`ai-agent-orchestration.com`'s canonical tagline) | **1,369** | **409** | supplied by the brief, emitted verbatim |
| `not a sales process` (same site's `cta_style.approach`) | **0** | 35 | reaches no prompt — **the model's own phrasing, not transfer** |
| `rather than transaction` (same) | **0** | 21 | same |

**So the evidenced class is narrow and specific: phrases the brief HANDS the writer.**
Instructional contrast — a brief saying "use stack references naturally, not as buzzwords" — has
**no** demonstrated effect on output, and the fields carrying most of it turn out to reach no
prompt at all. Whether *form* transfers independently of a supplied phrase is still open, and
after this week it gets designed to come out either way before it is run.

**One sharpening on the site that drew the complaint.** Its brief does not merely contain the
construction — its `emphasis` block **orders** the tagline into *"the homepage hero, services
page hero, site footer, and meta descriptions"*. It is the only mandated supplied phrase of its
kind in the estate. The owner objected to a hero; the brief commanded that hero.

## 3. ⚠ A separate defect lives in the same field — do not merge the two stories

Filed today as **`bugs_open/327`**: `content_direction.formatted` — the only part of the brief
the writer reads — is rebuilt from the **incoming partial** before the deep merge, so a partial
write silently drops every untouched key from the writer's view while leaving it in the document.
Three sites have been serving a fragment since 2026-04-18, `ai-agent-orchestration.com` among
them (5 of 18 keys).

**It is not the cause of your symptom, and I have said so in the bug file** — the dropped keys
never mention the construction, and one of them (`example_phrases`) is *itself written in it*, so
restoring it naively would make that fault worse. Two problems, one field, and the tidy version
that merges them would be wrong.

**Why it matters to you specifically:** it is the trap sitting on the fix we both want. **A
careful, narrowly-scoped correction to a brief — exactly the shape of edit "fix the tagline"
implies — will delete the rest of that brief from the writer's view**, silently, with the
document still reading complete and the write logging success. `remortgagecalculator.uk` is
currently clean on that measure (0 dropped keys); a partial correction is what would break it.
Write the whole document, or recompute `formatted` from the merged row afterwards. Filed in
`LANDMINES.md`.

— `copy_quality_two_stage`, 2026-08-19
