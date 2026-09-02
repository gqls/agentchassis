# HANDOFF 2026-09-02 — continue here

**Supersedes `HANDOFF_2026-08-26_continue_here.md`** (its §5b/§5c remain the fuller record of the
footer and tools incidents; state below is current). Lane: `apis_uk_bees_homepage`. Session
"apis.uk". Evidence for everything below: `NOTES_apis_uk_bees_homepage.md`, newest at bottom.

> ## ▶ ONE-LINE STATE `[ALL MEASURED 2026-09-02 ~16:00–17:00 UTC]`
> The fresh chassis **carries the whole subjects build** (pod-probed, both controls). Seeds
> **639 + 640 are APPLIED and live-verified**; **641 waits ONLY on the owner reading its 4-line
> block** (§3.1). The six illustrations got their **IMG-075 section bindings** (6 rows, locked).
> A **correctly-reasoned index re-render is queued** to finally serve the GTM head — the 08-26
> attempt was an assemble-only no-op, byte-proven. **Nine tool-expansion items stayed parked all
> week** (`deferred` ×9). Open here: §2's verify+settle, three owner gates (§3), and the
> image-accuracy A+C build (untouched since 08-24).

## 1. What the 09-02 session did, in order

1. **Roll verified at the pod FIRST** (the seed headers' own commands): capability probe on
   `agent-chassis-744cfb4bf-mwzgx` — `section_subjects`=1, `SUBJECT_MISSING_ON_REPEATED_COMPONENT`=1,
   `subjectlessRepeatFindings`=2, +control `section_facts`=2, −control 0.
2. **Seed 639 applied** (wiring: `plan_sections` config gains
   `section_subjects=spec_sections.section_subjects`) — live-verified.
3. **Seed 640: its own anchor guard REFUSED the first apply** — `bugs_open/380`'s seed had
   rewritten rule 17's tail during the week ("use plain strings only" → "still use the object form
   with facts:[] on every section"). That change *helps* subjects (object form is now
   unconditional). Re-derived against the live row — subject sentences now insert BEFORE the 380
   sentence, kept verbatim — applied, live-verified (subject rule + 380 sentence + example all
   present). The refusal cost nothing and is exactly what the guard was for.
4. **IMG-075 adopted** for the six illustrations (the `inline_guide_imagery` CONTRIB of 09-02,
   pairing verified against live positions 2–7 before running): six `scope='section'` rows in
   `site_plan_imagery`, `index:1`…`index:6`, locked to this lane. Generates nothing.
5. **A correctly-reasoned `page_rerender` filed** for index:
   `item_key = page_rerender_index_1c6f3424-…_section_data_resolved`, `spec.reason='section_data_resolved'`.
   WHY THE REASON IS LOAD-BEARING: only `image_landed`/`section_data_resolved` take the
   re-resolving path; anything else is assemble-only and **redeploys the stored bytes unchanged**
   — which is precisely what the 08-26 15:38 "completed" rerender did (deployed, success:true,
   and the served page kept the tagless head at the same 68,248 B).
6. Registry debt: **already fixed by the 394 lane** on 08-28 (`3749132e0`) — my
   `SUBJECT_MISSING_ON_REPEATED_COMPONENT` entry had its prose under `note` where `human-evidence`
   reads `why`, keeping `TestShippedRegistryIsSelfConsistent` red for seven days. Their CONTRIB
   sat unread in this directory the whole time. **Lesson, standing: read the lane's CONTRIB files
   at every session start** — `ls docs024_key_docs_latest/apis_uk_bees_homepage/CONTRIB_*` against
   the NOTES' last date.

## 2. THE GTM RE-RENDER — measured OUTCOME (17:20 UTC) and the route left

**The reasoned rerender was REFUSED at `save_sections`:** *"overwrite: REFUSED for page "index" —
this run re-confirmed too little of what is stored (prune_floor_ratio=…)"* — back to `triaged`
for its remaining attempts (max 3), which will likely refuse the same way. Served page unchanged
(gtm=0, 68,248 B, six illustrations intact). **So BOTH rerender modes are blocked on this page:**
assemble completes but redeploys the stored (tagless) bytes; re-resolve runs plan_sections and
the save guard refuses because the re-resolution re-confirms too little of the locked/stored
content. This is the deeper shape the 383 lane originally named — my 08-26 WRONG_CALLS
"correction" of it was itself too wide: the refusal is MODE-SPECIFIC (assemble completes,
re-resolve refuses), and the handoff record now carries both halves.

**The one route PROVEN on this exact page:** the site-level `rerender-pages` fan-out envelope
(08-26 handoff §3 of 08-24 vintage: `{"action":"orchestrate","config":{"agent_type":"rerender-pages"},
"input_data":{"site_id":…,"domain":"apis.uk","refresh_site_components":false}}`) — it served the
GTM tag on apis.uk as the canary on 08-24, with the locks already in place. Next session: read
the failed item's FULL result first (`SELECT result FROM site_work_items WHERE item_key LIKE
'%section_data_resolved%'`), then fire that envelope, then verify + settle:


Verification, whenever a render actually deploys:

```bash
curl -s "https://apis.uk/?cb=$(date +%s)" > /tmp_page.html   # (use the scratchpad, not /tmp)
grep -c googletagmanager …        # PASS: 1   (serving the c2 head at last)
grep -o 'src="[^"]*illustration[^"]*"' … | sort -u | wc -l   # PASS: 6 distinct
# bytes should differ from 68,248 (the no-op signature)
```
- `[MEASURED]` baseline before it: gtm=0, footer=1, sections intact, 68,248 B.
- **Then settle**: `UPDATE pages SET build_status='deployed' WHERE …domain='apis.uk'` — the render
  re-queues the page (`needs_rebuild` is queue membership), settle AFTER verifying.
- `last-modified` older than the item's completion = the deploy QUEUE, not a failed render.
- If the item FAILS: read `result`/`error` — do NOT re-file an assemble rerender; the reason is
  the whole point. Check `SELECT spec->>'reason' FROM site_work_items WHERE item_key LIKE
  '%section_data_resolved%'` says what you think it says.
- ⚠ The six `page_components` rows must be byte-untouched afterwards (7/7 `permanent` locks) —
  re-render is the safe path; the SAVE path is the dangerous one (§5.3).

## 3. THREE OWNER GATES — nothing below moves without him

1. **Seed 641 (writer prompt v5) — needs HIS READ of the inserted block, then hand-apply.**
   RFC_016 §5.2: the v4 approval attaches to its committed text and voids on edit. The ENTIRE
   delta, verbatim (also in the seed header, `641_page_content_writer_prompt_v5_section_subject_HOLD.sql`):
   > `{{if .current_section.subject}}`**## This section's subject**
   > `{{.current_section.subject}}`
   >
   > Write THIS section about that subject specifically. Sibling sections on this page carry
   > their own subjects - do not restate theirs, and do not widen this section into a general
   > treatment of the page's topic.
   > `{{end}}`
   Renders ONLY when a subject is assigned; every other prompt byte identical (verify block
   asserts landmarks + em-dash census). After his read: apply 641, record the read in NOTES + the
   APPLIED line. **Until then the writer ignores subjects** — the chain is planner→DB complete,
   consumption gated.
   > **CORRECTED 2026-09-02 (late): gate 2 RETURNED — a REDRAFT direction, not an approval.
   > Do NOT apply 641 as written; the delta quoted above is no longer the read-target.** His
   > directive: positive prompting only (no "do not…" instructions), written in the language
   > expected back, no specimen answer. Candidates + mechanics-to-keep:
   > `docs/agent_docs/docs024_key_docs_latest/finetuning_uk_service/DRAFT_2026-09-02_641_positive_prompt_candidates.md`
   > — owner's framing pick pending; finetuning lane test-renders; THIS lane writes the final
   > SQL; approval attaches to the exact final words (RFC_016 §5.2). New pre-apply obligation:
   > all three candidates enumerate SIBLING subjects and that range render is UNTESTED — prove
   > it against real loop CollectedData (empties must drop out, not render blank). Seed header
   > stamped with the same. Full record: NOTES 2026-09-02 ~21:00 UTC.
2. **The footer** (08-26 handoff §5b): fallback shell reappears on every chrome refresh; no
   suppress mechanism exists. Accept the shell, or commission opt-in `chrome.footer_disabled`.
3. **The tools park** (08-26 handoff §5c): nine items `deferred` — held all week, dedup keys
   intact. RFC_056 area filed by the loop owner. **Order stands: refusal declaration BEFORE any
   cancel** (cancelling releases the dedup keys). If he accepts tools instead: un-defer the nine.

## 4. Per-section subjects — chain status

| hop | state |
|---|---|
| planner rule 17 + example (640) | **LIVE** — next replan of any fact-listed site emits subjects; REQUIRED on repeated components |
| `validate_plan` normalise + carry · `site_plan_sections.subject` (638) · loader · `plan_sections` (639 wiring) · `sectionPlanItem.Subject` | **LIVE end to end** (binary probed; config verified) |
| writer prompt (641) | **HELD — gate 2 returned REDRAFT 2026-09-02: do NOT apply as written; redraft in flight (§3.1 correction)** |
| structural detector `SUBJECT_MISSING_ON_REPEATED_COMPONENT` | **LIVE** (binary probed); zero rows until subject-carrying plans exist — that is the gate working, not silence |

Adoption query (also the copy_quality lane's experiment control):
`SELECT count(*) FILTER (WHERE subject IS NOT NULL), count(*) FROM site_plan_sections sps JOIN site_plans sp ON sp.id=sps.plan_id WHERE sp.is_current;` — expect 0/N until replans happen post-640.

**Un-defer path for swarm + pollination** (the two 08-24 `content_rewrite` items): after 641 —
their sections get plan-row subjects; their images now have the IMG-075 route (section-scope rows
bound per-ordinal) INSTEAD of the content_data+lock-only route. Their assets exist
(`illustration_swarm`, `illustration_pollination`, active). Still requires adding sections to a
locked page = plan surgery or replan — a decision, not a mechanical step.

## 5. This week's inbox, dispositioned

1. **`bugs_open/411`** — the routed `image_source_unsatisfiable` finding, CONFIRMED; our item on
   apis.uk stays open as cited evidence. Do not close it from this lane.
2. **`bugfix_394` CONTRIB** — registry fix; done by them; lesson in §1.6.
3. **`inline_guide_imagery` CONTRIB (09-02)** — adopted (§1.4/1.5). ⚠ Their **decisive test** —
   fire a `content_rewrite` (SAVE path) and prove the images survive — was deliberately NOT run:
   the six rows are `permanent`-locked, so the save path REFUSES their overwrite today, and
   running a destructive-if-wrong test at the end of a wrap-up session is the wrong moment. It
   belongs with the un-defer work, eyes open. They also struck IMG-074's "carryStored preserves
   the six" register claim — correct, the carry is repetition-disabled for our slots.
4. **`brief_supplies_negation` moved TODAY** (13:17 UTC: one `complete`, one NEW
   `needs_human_review`) — **unread by this session**; read before acting. History: the 08-23
   annotate-don't-close decision was the owner's call.

## 6. Traps, current set (older ones: 08-26/08-25 handoffs)

- **A completed `page_rerender` proves NOTHING about re-resolution** — check the item_key suffix
  and `spec.reason`; `_assemble` redeploys stored bytes and reads as success.
- **The park IS the hold** — `deferred` rows keep dedup keys; cancel releases them.
- **psql `LIKE '%\"x\"%'` inside single quotes searches for literal backslashes** — a false
  "absent" from your own instrument; this session ate one round on it. Pattern without escapes.
- **`\gset` variables do not interpolate inside `DO $$` bodies** — assert with plpgsql
  `GET DIAGNOSTICS` instead (RUNBOOK §8 has the worked shape).
- Chrome regeneration resurrects the fallback footer and re-queues the page; settle after.

**Files:** PLAN `PLAN_2026-08-26_per_section_subjects.md` · RUNBOOK `RUNBOOK_apis_uk_bees_homepage.md`
(§8 parking) · register PBP-049 / IMG-074 / IMG-075 · bugs 397 (tag strip), 411 (checker claim),
380 (rule-17 tail).
