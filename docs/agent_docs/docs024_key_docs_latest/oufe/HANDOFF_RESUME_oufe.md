> # SUPERSEDED — 2026-07-27
> **Start at `HANDOFF_2026-07-27_continue_here.md` instead.** This file's
> state table describes a site whose case was unwritten and whose tool did
> not exist. Both are live now. Kept for the record only.

# HANDOFF / RESUME — oufe.com + oxenunity.com

Cold-start entry point for this workstream. Written 2026-07-25.
Read `PLAN_2026-07-25_oufe.md` first for what the site IS and why; this file is
only "what state is it in and what do I do next".

## State at handoff

**Both sites are LIVE.** Latest read-out: `SUMMARY_2026-07-26_oufe.md`.

| Thing | State |
|---|---|
| oxenunity.com | **LIVE** — owner moved the NS; serving via Cloudflare→B2 |
| oufe.com | **LIVE** — index, about, cases/index, contact. **Zero broken links** (full crawl) |
| Figure rail | **HELD** — 0 currency / 0 percentages / 0 statistics across ~50KB of generated specs and every live page |
| evidence_base | seeded, **0 facts**, 18 banned patterns — writers may assert no numbers |
| Fallibility posture | **LIVE in copy** — the "everything is sourced" promise is gone from every page |
| Council seat | **mig 223 LIVE** — compliance seat catches overclaimed-reliability + illustration-not-authority; mirrored to council-gate via 099 |
| Content writers | **mig 223 LIVE** — both carry the never-promise-accuracy rule (verify BOTH, different prompt paths) |
| grounded-explainer | **mig 224 LIVE, first run exercised 07-26** — generic high-attention lane, cannot publish |
| Waterfall tool | authored + arithmetic verified + condition-of-use gate; insert SQL prepared, **not applied** |
| Thames Water dossier | **not started** — needs evidence first |
| Legal pages | not built; wording drafted, **needs owner approval** |
| Bugs filed | `bugs_open/094` `095` `096` `097` |

## First thing to check

```bash
# both sites still serving, and no dead links
for u in / /about.html /cases/index.html /contact.html; do
  echo "$(curl -s -o /dev/null -w '%{http_code}' --max-time 25 --retry 3 --retry-all-errors "https://oufe.com$u")  $u"
done
```
**Read a `000` as "no answer", never as "not found"** — they mean opposite things,
and conflating them cost a false report on 07-26.

Any grounded-explainer drafts waiting:
```sql
SELECT wi.created_at, wi.summary, jsonb_pretty(wi.spec->'grounding_audit')
  FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
 WHERE s.domain='oufe.com' AND wi.item_type='grounded_draft_review'
 ORDER BY wi.created_at DESC;
```
**Read the audit before the draft.** `ungrounded` must be empty. If it is not, cut
those sentences — do not go looking for a source that agrees with them.

## Then, in order

1. **Finish the mechanism explainers** via the grounded lane — the creditor
   waterfall, and the UK statutory framework. One run each:
   ```bash
   ./docs/agent_docs/docs024_key_docs_latest/oufe/TRIGGER_grounded_explainer.sh \
     oufe.com "<topic>" "<research query>" "<audience>"
   ```
   These come **before** any case dossier: they are the part that transfers, and
   they carry no fabrication risk because mechanism can be taught with openly
   hypothetical figures.
2. **Publish each approved draft** — the lane deliberately cannot. Create the page
   + `page_components` row (`slot_name` MUST equal the component function name —
   `bugs_open/095`), then `TRIGGER_rerender_page.sh`.
3. **Thames Water evidence** via `evidence-researcher`, then the dossier. Nothing
   asserted until the register holds it. Watch `bugs_open/073`: a component with a
   required stat field plus an honest empty value fails the whole page build.
4. **Legal pages** via the migration-182 pattern once the owner approves wording:
   hand-written content, `rebuild_policy='owned'` + permanent component lock,
   `rendered_html` written in the same migration. Also defuses `bugs_open/053`.
5. **The tool**: apply `PREPARED_tool_insert.sql`. Read its header first — three
   landmines in the comments, and the acceptance criteria now click the
   condition-of-use gate before filling anything.

## Owner list (blocking "live", not blocking building)

1. ~~Move oxenunity.com to Cloudflare~~ — **DONE by the owner 2026-07-26, verified
   serving.** No infra work outstanding on either domain.
2. **Approve the disclaimer wording** (`DRAFT_disclaimer_for_owner_approval.md`).
3. **Confirm the contact email** (`oufe@contactforsales.com` is seeded).
4. **Decide the audience question (PLAN §7)** — the owner asked whether targeting
   students is safer. Recommendation: take the honesty posture, keep the audience
   wider than students. Three options are set out there; the briefs need a
   revision either way, and pages are being written now.
5. **Respond to the challenge in PLAN §C1** — the owner ruled "radar first, it is
   lowest risk"; this workstream argues it is the highest risk first move and has
   built the dossier-plus-tool path instead. He has not yet answered that.

## Rails that must not be relaxed

- **No figure enters a brief, a spec, an identity or a content_direction.** Only
  the evidence register, with a source. A number in a spec is a *given* and beats
  every writer-side rule — that is how invented figures were written back over
  correct ones on another site with all guards live (043).
- **A clean claims report on this site means almost nothing.** The deterministic
  number scan has no finance vocabulary and excludes currency outright. The real
  layers are the writer whitelist, the banned patterns, V5 citations, and a human
  reading it.
- **Never publish a figure about a real company without a source URL and a
  capture date** — the vetcomparison rails, inherited whole.
