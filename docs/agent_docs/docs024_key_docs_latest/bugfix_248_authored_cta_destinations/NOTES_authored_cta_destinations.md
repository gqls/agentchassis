# NOTES — `cta_recompute_clobbers_authored_contact_links` (append-only, newest at the bottom)

## 2026-08-17 — lane opened, before any code was written

**Ownership checked first.** `scripts/who-owns.py 248` reports the ambiguous pair; the
commits it lists ("248 CLOSED", "bucket-A pilot") belong to `bugs_closed/248`, the
undeployed-asset case. The CTA file itself has two commits only: the 2026-08-10 filing and a
2026-08-17 CONTRIB from the `brochure_component_library` contrast front, which states in
terms *"Contributing measurements, not direction — nothing here is fixed and I have changed
no CTA."* Live `.jsonl` transcripts grepped for `applyCTARecompute` — the two sibling
sessions matching are on `bugs_open/257` and `275`. Target files clean in `git status`.
**No competing owner.**

### Measurements taken before designing (all live, 2026-08-17)

- Blast radius re-run from the bug file's own SQL: **20** components (was 24 on 08-10).
- Narrowed to what either writer can actually reach (`ctaFieldNames`): **13** — 6 `hero`,
  7 `call-to-action`. The other 7 are `system-stats`/`tool-*`/`tool-cta`, which no writer
  touches. 8 of the 13 are `webdesign.uk`, including its homepage hero.
- Schedulers: `site-discovery-rotation-{completeness,quality,availability}` **enabled**;
  `detected-item-promoter` every **900s**, last fired 16:12Z. Repairs completing today.
- `cta_names_unknown_destination` by reason: **103** "lands in an excluded area" (103 open),
  35 homepage, 32 phantom, 13 empty, 13 self.
- **149** detector findings whose `suggested_target` is a utility page the repairer's
  candidate set cannot produce (e.g. "Contact our supply team" → `/contact.html`).

### MISSTEP AVOIDED — a refutation that was reading history, not the system

An adversarial pass returned a CRITICAL finding: the fix is unsafe because live schemas
still carry `"fallback": "/contact.html"` on `hero.cta_url` and
`call-to-action.primary_cta_url`, so `plan_sections` would write a fabricated contact url
into `content_data` and the new keep-branch would freeze it for ever. It cited
`docs/agent_docs/sql_for_tables/005_content_components.sql:7729` and migration 091's note
that fallbacks were left untouched.

**Checked at the live table instead of the seed — REFUTED.** All ten CTA url fields across
the six protected components carry `source=renderer` and `fallback=NULL`:

```sql
SELECT function, key, val->>'source', val->>'on_missing', val->>'fallback'
FROM content_components cc, LATERAL jsonb_each(cc.input_schema->'fields') AS f(key,val)
WHERE cc.function IN ('hero','call-to-action','archetype-grid','archetype-combinations',
                      'gauntlet-cta','content-block-about') AND key LIKE '%cta%url%';
```

The only live schemas fabricating a utility URL are `site-header`/`site-head` — chrome,
which routes through `LoadChromeLinkPolicy` (LNK-030), not through either writer. Also
stale in the same report: `content-block-about.cta_url` was described as `source:llm` (098's
note); it is `renderer` now.

**The cheap check that caught it: read the live `content_components` row, never the seed.**
Logged to `WRONG_CALLS.md`. This is the standing "the seed is not the system" trap arriving
inside a subagent's report — which is another doc, with no seam showing where its measuring
stopped.

### What survived the adversarial pass and changed the design

1. **The invariant was not exact.** `candidatesFromHubs` gets the RAW loader lists —
   `rank()` filters a local copy and never mutates its inputs — so its comment claiming the
   inputs are pre-filtered is false, and 4 live `section-index` pages sit at utility URLs.
   The resolver's own label branch could therefore mint one. Fixed by filtering there.
2. **A "keep" must WRITE.** A bare return leaves survival to `plan_sections`' carry, which
   misses on non-`deployed` rows, conflicted duplicate slots and mismatched slot names — so
   the branch built to keep a button can drop it.
3. **Do not delete the detector arm — demote it.** It is the only arm that can see a
   fabricated-but-valid contact link. Keep the finding, drop the work item.
4. **History outlives the code.** Until 2026-07-14 a schema source wrote `/contact.html`
   unconditionally (75/75 call-to-action, 68/69 hero), so derived provenance — a claim about
   *today's* writers — could freeze old machine junk. **Dated every in-scope row against
   `page_component_history`: 12 of 13 first appear AFTER 2026-07-14** (webdesign.uk
   08-11/08-12, leopardess 07-15, ai-agent-orchestration 08-11). One exception,
   `finetuning.uk/services` at 2026-04-23, to be inspected by hand before shipping. Worst
   case for a frozen fabrication is a *working* link to the contact page — `validPages`
   membership is part of the predicate — against a destroyed authored link as the harm
   prevented.

## 2026-08-18 — the defect fired WHILE this fix was being written, and it is now attributable

Going to inspect the one at-risk row that predated the 2026-07-14 migration
(`finetuning.uk/services`, the row that could plausibly have been machine junk), I found it
no longer holds a contact link at all. It was destroyed at **2026-08-17 19:11:24Z** — about
two hours after I measured the population, and while the bug file still described the defect
as "mostly dormant".

```sql
-- the write
SELECT pc.updated_at, pc.content_data->>'primary_cta_url'
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE s.domain='finetuning.uk' AND p.name='services' AND pc.slot_name='call-to-action';
--  2026-08-17 19:11:24.716512+00 | /tools/password-entropy.html

-- what completed in the same second
SELECT swi.item_key, swi.status, swi.updated_at, swi.spec->>'reason'
FROM site_work_items swi JOIN sites s ON s.id=swi.site_id JOIN pages p ON p.id=swi.page_id
WHERE s.domain='finetuning.uk' AND p.name='services' AND swi.updated_at > '2026-08-17 18:00';
--  misdirected_cta:services:1368e337-… | complete | 2026-08-17 19:11:34.639489+00 | cta_links_stale
```

**Before → after, from `page_component_history` (join on `page_id` alone — every history row
that points at contact has `component_id IS NULL`):**

| | label | destination |
|---|---|---|
| before, archived 19:11:24.684533Z | `Start a Conversation` | `/contact.html` |
| after, live now | `Start a Conversation` | `/tools/password-entropy.html` |

The label was not touched. Only the destination was rewritten — so the copy is still standing
there as evidence of what the link was for. "Start a Conversation" reduces to `[conversation]`
(`start` and `a` are stopwords), which names no page, so the label-match branch declined, the
keep-branch refused it for being in an excluded area, and the positional pick took it.

**Why this matters beyond one page.** The 2026-08-17 CONTRIB to the bug file measured 59
pages that had lost a contact CTA but stated plainly that it could not attribute them:
*"nothing in the history distinguishes 'clobbered by `applyCTARecompute`' from 'regenerated
wrong' — both arrive as a `save_page_sections_overwrite` row."* Here the attribution IS
available, because the work item is named: a `misdirected_cta` item carrying
`reason=cta_links_stale`, whose ONLY CTA behaviour is `applyCTARecompute`, completing in the
same second as the write. This is one confirmed instance of the mechanism in production, not
a population to triage.

Fleet at-risk count moved **20 → 18** overnight, consistent with this and one other loss.

**Consequence for the plan:** the row I was going to inspect by hand no longer exists in the
at-risk state, so the "could this be pre-migration machine junk?" question is moot for it —
and the remaining at-risk set is now entirely post-2026-07-14. The freeze-a-fabrication risk
the adversarial pass raised is measured at **zero** rows today.
