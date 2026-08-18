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

## 2026-08-18 (later) — the fix is LIVE, and a peer session corrected one of my claims

### Proven at the artefact, not at the tag

Chassis `v1.0.1310`. The stamp is the BUILD commit only, so grepping the binary for my own
sha is the wrong test — it is absent even when the fix is in. Extract the stamp, then test
ancestry:

```bash
kubectl -n ai-persona-system exec <pod> -- sh -c "grep -aoE '[0-9a-f]{40}' /proc/1/exe | sort -u"
# 79 runs; feed each to: git cat-file -e <h>^{commit}   -> exactly ONE is a real commit
git merge-base --is-ancestor 53a8d3c1d 0b185bad2a49c6e032352fa9e7d0b429f0a95104   # passes
```

Positive control `build provenance` present in the binary. The 79 runs are mostly Go's
internal digit tables (`0001020304050607…`), which is precisely why the standing landmine
forbids a bare discovery grep — `git cat-file -e` is what makes the extraction discriminating
rather than a guess.

⚠ Two dead ends first, recorded so the next reader skips them: 40 sequential `kubectl exec`
greps (one per candidate sha) timed out at 2 minutes, and so did one `grep` with 250 `-e`
patterns against the binary. Extract-then-classify is the cheap direction, not match-then-test.

### MISSTEP — I overstated the build-path half, and a peer caught it

The `bugs_open/299` session messaged to say `page-content-writer` reads the resolver's output
from `resolved_links.response.link_resolution.sections_ready`, a level the response does not
have, so `resolve_internal_links`' output is discarded on every fresh build.

**Verified here before propagating** (a peer's report is another doc, exactly like a
subagent's):

```sql
SELECT count(*),
       count(*) FILTER (WHERE collected_data->'resolved_links'->'response'->'link_resolution' ? 'sections_ready'),
       count(*) FILTER (WHERE collected_data->'resolved_links'->'link_resolution' ? 'sections_ready')
FROM orchestration_states WHERE collected_data ? 'resolved_links';
-- 150 | 0 | 0
```

The third column is the one worth having: the obvious repair — delete the `.response` level —
**also** matches nothing, so anyone fixing `bugs_open/312` cannot assume it is a one-word edit.
Confirmed at config too: `internal-link-resolver` publishes `link_resolution.sections_ready`;
`page-content-writer` is the only definition referencing the longer path.

**So `setCTAField`'s missing keep-branch has been INERT on fresh builds**, and my confirmed
clobber (finetuning.uk/services, 08-17 19:11Z) necessarily arrived via the rerender path — as
the work item's own `reason=cta_links_stale` already said. I had written "dies on the next
full regeneration" into the bug file, the register entry, the commit message and the owner
log. Corrected visibly in all four; the commit message cannot be amended (forward-only), so
the correction lives in the bug file it points at.

**The fix is not weakened — it is differently motivated, and the difference matters for
sequencing.** `bugs_open/312` is HELD, and its traced run shows the discarded resolver output
had *already* repointed an authored "Get in touch" at a tool. So the build-path branch is a
guard PRE-POSITIONED for the instant 312 unholds, rather than a repair of damage in flight.

**⚠ Consequence I must not let a later reader miss: the build-path branch has never executed
in production and cannot until 312 lands.** It is unit- and mutation-tested only. The rerender
canary does not vouch for it.

**The cheap check I skipped:** I established that `setCTAField` writes into `resolved_data`
and stopped there, without asking whether anything downstream ever READS it. "Writes the
field" is not "the field reaches the page" — the same shape as the standing lesson that
*writes the field is not reads the field*, one seam further along. One query over
`orchestration_states` would have shown it, and I had already run queries against that table
twice today for other reasons.
