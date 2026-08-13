# HANDOFF 2026-08-13 — start here: the fleet fix for `bugs_open/268` (214 CTA buttons missing across 19 sites)

**Written for a fresh thread by the `ai_site_selling_automation` lane**, which
found the defect, measured it, protected ONE site with a tourniquet, and is not
building the fleet fix. You are. Read order: this file → `bugs_open/268` (the
evidence, in full) → the two refuted-hypothesis sections below **before** you
form your own theory.

## 0. The defect in one paragraph

A content **regeneration** (`content_rewrite` → `plan_sections` →
`save_page_sections`) drops the CTA destination keys — `cta_url`,
`primary_cta_url`, `secondary_cta_url` — from `page_components.content_data`,
while keeping the button **labels**. Both shared templates gate the anchor on
the URL, not the label (`{{if and .cta_text .cta_url}}…`), so the button
renders as **nothing**: no error, no missing prose, healthy byte counts, clean
claims scan. Measured 2026-08-12: **216 hero/call-to-action components across
19 live sites carry a label with no URL; 214 render zero anchors.** The two
components are shared fleet-wide: `hero` `23f95f00-…` (20 sites / 276
instances), `call-to-action` `0197e8d7-…` (20 sites / 237 instances).

## 1. What is ESTABLISHED (do not re-derive) vs OPEN (do not assume)

**Established, with a control:**
- The keys are **dropped by regeneration**, not merely never-resolved. On
  webdesign.uk, 2026-08-12 ~20:26–20:42Z, a five-page `content_rewrite` took 7
  key-carrying components to 0 and site hrefs 28 → 13, while **`contact/hero`
  — the one page NOT in the rewrite — kept its keys and links**. Same site,
  same shared components, same schema, same run.
- Restoring `content_data` alone does **not** put buttons back: a
  `page_rerender` dispatched after a key restore still rendered no buttons.
  Proven twice (17:2x and 20:4x).
- The `bugs_open/238` carry (PBP-039, `carryStored`) **is live** — agent-chassis
  `v1.0.1291` ← `da5a7eb8f`, `git merge-base --is-ancestor d26c26a9a da5a7eb8f`
  passes with controls both ways — and the loss happened anyway. (238 §8 still
  says the carry is unrolled; it is stale and its own §9.1 says so.)
- The fields ARE declared in the shared schema:
  ```sql
  SELECT cc.function, f.key, COALESCE(f.value->>'source','llm') AS source, f.value->>'on_missing'
  FROM content_components cc,
       LATERAL jsonb_each(COALESCE(cc.input_schema->'fields', cc.input_schema->'properties','{}'::jsonb)) f
  WHERE cc.function IN ('hero','call-to-action') AND f.key LIKE '%url%';
  -- all four: source='renderer', required=false, on_missing='skip_field'
  ```

**OPEN — the mechanism.** The candidate reading (268 §4): `sourceResolver.resolve`
short-circuits `source:"renderer"` to `(nil, found=true)`
(`plan_sections_action.go:623`), so the field is never *missing*,
`handleMissingField` never runs, and `carryStored` (which only guards fields
that FAIL to resolve, ~:2123) never runs; the renderer/static branch (~:2362)
`continue`s writing only a declared `fallback`, which these fields lack;
`save_page_sections` then replaces `content_data` wholesale. **Plausible,
consistent with the control experiment, and NOT established.**

## 2. Two hypotheses already refuted — start from these, not from zero

1. **"The URL keys are undeclared in the schema, so the carry never sees
   them."** Refuted by the schema query above: all four declared,
   `source: renderer`. From the symptom alone, an undeclared field and a
   declared-but-short-circuited field look identical.
2. **The candidate mechanism itself was REFUTED by `090` run
   `97ef39f0-19df-4935-834d-c80514fbc43e` — but the run was worthless, and the
   reason is the trap you must not repeat.** The filing session repaired
   `content_data` at 17:23 and fired the run at 17:39; the verdict's citations
   are the restored values. **Fix first, ask second, and the answer is about
   the fix.** The run's one durable output: `bugs_open/238` as tracked in code
   comments is the *dead-URL-control* defect (`dead_url_guard.go`,
   `recordDeadURLControls`, `emitSectionDeadControlItem` — a section that
   renders with an empty URL attribute), which is *adjacent* to this, not this.

## 3. Your first act: the 090 re-run, authored so it can actually answer

Diagnosis before debugging is the norm for exactly this shape (cross-cutting,
cause not at the symptom, one confident hypothesis already dead). Author it:

- **Point it at `page_component_history`, not at live rows.** The live
  webdesign.uk rows are repaired and locked; they will refute you the same way
  twice. The durable evidence windows: **2026-08-12 16:37–17:23Z** (first
  incident: keys present in outgoing rows, absent in stored) and
  **2026-08-12 20:20–20:45Z** (controlled reproduction, `contact/hero` as the
  untouched control). State in the symptom that live rows were repaired at
  17:23 and ~20:44 and locked at 20:46.
- State the mechanism as a QUESTION (does a renderer-sourced field's
  `(nil, true)` resolution bypass both `on_missing` and the 238 carry on the
  save path?), point at `plan_sections_action.go`
  (`sourceResolver.resolve`, `carryStored`, `handleMissingField`, the
  renderer/static branch), `save_page_sections_action.go`,
  `rerender_page_sections_action.go`, and the schema rows above. No counts, no
  downstream-consequence clauses.
- The 090 trigger refuses if another thread has open work on the target —
  `bugs_open/268` is this workstream, so that hit is YOU; read the findings,
  then `FORCE=1` if it is only your own filing.

## 4. Fix candidates, ranked by what closes the door (268 §6, expanded)

1. **Route these fields through the missing-field path** so the 238 carry and
   `on_missing` apply — e.g. resolve `source:"renderer"` to *not found* when
   render-time cannot actually supply a value, or add an explicit opt-in field
   flag (`carry_stored: true`) on the four schema entries. **The 2026-08-02
   owner ruling prescribes the shape: new authority on a shared seam ships as
   an opt-in field with the unsafe default OFF** — and per RFC_022, an opt-in
   field whose unsafe side is the default-off and which no live consumer names
   is NOT architecture-scope. Schema half is config (live immediately, 20
   sites at once — canary first); any Go half is inert until a roll.
   **Verify the carry actually fires before believing this sufficient** — its
   guard is `source == "" || source == "llm"` → return false, so as written it
   may *also* skip renderer-sourced fields even when they do reach
   `handleMissingField`. Read `carryStored` before trusting it.
2. **Make `resolve_internal_links` re-resolve CTA destinations after a
   regeneration** — the platform's intended owner of these keys; it already
   knows when it fails (it files `unresolved_cta` "no real-page destination"
   items). Go change; council gate; inert until rolled.
3. **Template-level fallback** (gate the anchor on the label, default the
   href) — cheapest, but bakes a site-shaped assumption (`/contact.html`
   exists) into a component used by 20 sites. Ranked last for that reason.
4. **Component locks** — what webdesign.uk has now. Site-scoped tourniquet,
   not a fix; listed so you know it exists, not so you spread it.

Whatever ships: **register the seam** (concept register, same commit), **name
the consumers** (20 sites; the producing agents are page-build-handler +
page-rerender), and put the change through the **council gate before or
alongside the commit** (`Council-Submitted:` trailer if committing first).
Check `bugs_open/238`'s owner before touching shared code it also touches:
`scripts/who-owns.py 238` — last commit 2026-08-11, no live session as of
2026-08-12 21:00Z, but re-check; and grep live `.jsonl` transcripts, because
who-owns is blind to uncommitted sessions.

## 5. The REPAIR is a separate deliverable from the FIX

The fix stops recurrence. **214 components on 18 unprotected sites are still
damaged now** and stay damaged until repaired. Order matters: **fix first,
then repair, then re-render** — repairing first just re-arms the loss on the
next regeneration (proven: webdesign.uk was repaired at 17:23 and stripped
again at 20:2x).

- Recover URLs from `page_component_history` (rows whose `content_data` still
  carries the keys — that is where webdesign.uk's were recovered from), or
  re-run `resolve_internal_links` per site once candidate 2 exists.
- Write to **`content_data`, then re-render** — never patch `rendered_html`
  except as a stopgap; a hand-patched artefact re-arms `bugs_open/229`
  (divergence overwrite) by construction.
- Worked repair SQL to crib:
  `ai_site_selling_automation/SQL_2026-08-12d_restore_cta_urls.sql` (+ `e` for
  the stopgap HTML splice, with its refuse-unless-exactly-one insertion guard).

## 6. Verification — the checks that caught this, and the five that could not

**The invariant diff** (the ONLY check of six that saw the damage). Take it
before and after any regeneration, as a matched pair:
```sql
SELECT p.name, pc.slot_name,
       (SELECT count(*) FROM regexp_matches(pc.rendered_html,'href="','g')) AS links
FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE p.site_id='<site>' AND p.status='active' ORDER BY p.name, pc.position;
```
Blind to this failure, all proven green during the incident: `claimscan` (bans
catch phrases, not absent hrefs), body-byte deltas (bodies GREW), retired-term
greps, served-artefact fetch, and a link gate armed on the wrong page.

**End-to-end proof pattern** (used to prove the lock; reuse it to prove the
fix): dispatch a real `content_rewrite` with `mode=edit_live` against a canary
page whose keys are present, then compare keys + hrefs + `updated_at` either
side. `mode=edit_live` matters — without it the writer also guts the prose
(`bugs_open/178`) and your button measurement drowns in a bigger failure.

**Canary discipline:** one site, two pages that disagree (a heavy one and a
light one), before any fleet sweep. A stale page holds every improvement since
it rendered, so re-render cost is not sized by your change alone.

## 7. Coordination and landmines

- **webdesign.uk is HANDLED — leave it out of sweeps.** Repaired, and its 8
  CTA-bearing components are `lock_type='permanent'`
  (`SQL_2026-08-12k`), which `save_page_sections` honours
  (`loadActiveLockedRows` + the agent-writable DELETE predicate,
  `bugs_open/058`). Proven live at 21:0xZ: a real rewrite left the locked rows
  untouched (`updated_at` unchanged) while editing the unlocked body. A fleet
  repair that tries to write those rows will no-op or file
  `lock_blocked_change` — that is correct behaviour, not a bug. Once the fleet
  fix is live and proven, **unlocking webdesign.uk is the final step** (the
  lane's RUNBOOK carries the unlock/edit/relock recipe).
- The sibling `webdesign_uk_build_service` lane owns the contact page's
  chat-input-box (also locked, theirs). Do not touch it.
- **`EvidenceFact`-class trap for any spec you edit:** a non-numeric `value`
  in `evidence_base.facts[]` silently disarms a site's whole claims layer
  (`LANDMINES.md`, 2026-08-12 entry). Test any register edit with
  `cmd/claimscan` before writing.
- **`git log` on `bugs_open/268`** before you assume this handoff is current —
  the filing lane may have appended. Grep `LANDMINES.md` for
  `save_page_sections`, `plan_sections_action.go`, `negatedClaimMatch` before
  touching those files; all three carry entries.
- A `failed` work item is not always failed work (spawn→call handshake race —
  two of yesterday's items said `failed` with the page correct and deployed).
  Verify at the artefact in both directions.

## 8. Falsifiers (re-check before trusting this file)

A newer handoff in this directory; `bugs_open/268`'s own tail (contributions
land there); whether the 090 re-run has already been run (`site_work_items
WHERE item_type='needs_diagnosis'` + `diagnosis_artifacts` by correlation);
whether the fleet count has moved
(the §2 census query in `bugs_open/268` — 216/214 was 2026-08-12 ~20:45Z);
whether `agent-chassis` has rolled past `v1.0.1291` (ask the image label, not
git); whether webdesign.uk's locks still hold
(`SELECT count(*) FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE p.site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f' AND
pc.lock_type='permanent' AND pc.slot_name IN ('hero','call-to-action');`
— expect 8).
