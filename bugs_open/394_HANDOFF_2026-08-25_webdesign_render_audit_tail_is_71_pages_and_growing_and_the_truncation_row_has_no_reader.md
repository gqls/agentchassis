# 394 — webdesign.co.uk's render audit covers 60 of 131 pages, the tail GROWS every week, and the truncation row `bugs_closed/242` built has no reader

**Filed** 2026-08-25 by the `bugs_open/358` lane, on the owner's ruling of 2026-08-25 (decision 4:
commission a reader). Lineage: `bugs_closed/242` (*"a capped render audit is indistinguishable from
a complete one"*) — **CLOSED correctly**: its fix made truncation loud (`pages_total`/`truncated`
stamped into the durable result, plus an `agent_error_log RENDER_AUDIT_TRUNCATED` row before
dispatch), and migration `392_render_audit_rotation_max_pages_60.sql` (applied) raised the cap
25→60 as a stated MITIGATION. This file is the next rung: **the loud signal exists and nothing
reads it, and the mitigation has been outgrown.**

## 1. Evidence — the writer says it plainly, weekly, to nobody

```sql
SELECT occurred_at::date, left(error_message,120) FROM agent_error_log
 WHERE error_code='RENDER_AUDIT_TRUNCATED' ORDER BY occurred_at;
```

`[MEASURED 2026-08-25]`:

| date | message |
|---|---|
| 08-11 | `render audit truncated by max_pages: 5 of 26 live pages audited for loancalculator.co.uk …` |
| 08-18 | `… 60 of 109 live pages audited for webdesign.co.uk — the unaudited tail is the SAME pages every run` |
| 08-21 | `… 60 of 125 … webdesign.co.uk …` |
| 08-24 | `… 60 of 131 … webdesign.co.uk …` |

Two live facts:

1. **webdesign.co.uk has outgrown the 60-page mitigation and is diverging**: 109 → 125 → 131 live
   pages in six days, tail now **71 pages that are never audited — and the writer's own message
   says the tail is the SAME pages every run.** Whatever class of defect the render audit exists to
   catch is structurally invisible on more than half of the fleet's largest site.
2. **The 08-11 loancalculator row says `5 of 26`** — a cap of 5, not 25/60, on a site the 392
   migration's own header shows at 25-of-27 previously. `[UNEXPLAINED]` — a per-call override, or a
   config regression; whoever takes this should read that call's config before assuming the cap is
   uniform.

## 2. What is asked for

Commissioned (owner ruling 2026-08-25): **a reader for `RENDER_AUDIT_TRUNCATED`.** Candidates,
ordered by what closes the door:

1. **Make the rotation actually rotate**: persist a per-site cursor so the next run starts where the
   cap cut off. The writer's message says the tail is the same pages every run — a cursor makes the
   cap cost latency instead of coverage, and it retires the signal's cause rather than reporting it.
   (242's fix candidates discussed this; re-read that file's §4 before designing.)
2. **Read the row**: a daily-check-family consumer that alarms on any site truncated in the last N
   runs, so a site outgrowing the cap is a finding rather than folklore.
3. **Raise the cap again** — the 392 shape. Weakest: webdesign grew 22 pages in six days; a constant
   will always be outgrown.

**Acceptance**: for candidate 1 — over consecutive rotation runs on webdesign.co.uk, the union of
audited pages reaches all 131 (verify by the audit's own durable page list, not the status), and
the `RENDER_AUDIT_TRUNCATED` message changes meaning (tail no longer "the SAME pages"). For 2 —
synthetic truncation row → red; mutation-proved both ways. Registry follow-up: flip the code to
`consumed` with `reader`/`reader_sink` in the shipping commit (`DBG-075`).

## 3. Traps

- **Do not re-file 242** — the visibility half is DONE and live; this is the consumption half.
- The `5 of 26` row is `[UNEXPLAINED]`; resolve it by reading the dispatching config, not by
  assuming the cap is 60 everywhere.
- Rows expire (365d declared; 14d once a consumer resolves them) — extract before resolving.
