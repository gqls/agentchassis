# HANDOFF — portfolio positioning lane, 2026-08-03 (supersedes 2026-08-02)

> **UPDATED 2026-08-05 — the next session starts HERE, with two owner-directed
> tasks that outrank the old backlog order:**
>
> 1. **Copy voice (owner priority).** The owner flagged LLM-styled copy (negative
>    framing etc.) on loanandmortgagecalculator.co.uk's hero. Provenance CONFIRMED:
>    hand-written by a CLI session + ADOPTED (41 × whole-document `ported-page`
>    components; the manual content_direction has never been consumed — no writer
>    ever ran). Four read-aloud rewrites await the owner's pick in
>    `COPY_STYLE_TRIALS_2026-08-05_hero_rewrites.md`, drafted against the house
>    style prompt `travelling_docs/pitch_pdf_source/REVERSE_ENGINEERED_STYLE_PROMPT_v3.md`
>    (its rule 3 IS the flagged tic). THEN: read the LIVE content-writer prompt
>    (agent_definitions — [UNVERIFIED], nobody has read it), fold the v3 rules /
>    winning register in (council gate; one trial page; owner review), and put the
>    winning register into content_direction voice blocks. Ported pages need
>    decomposition (old task #15) before pipeline copy can replace them.
> 2. **Blog ruling — RATIFIED 08-04, implement as seam 6.** Owner's clarification:
>    blog ≠ news, but editorial-from-news IS also blogging. Ruling: mission
>    vocabulary gains `blog: none / editorial-from-news / curated-features`; the
>    planner never plans the unbuildable section-index blog shape and never plans
>    any blog undeclared; the 015 retype arm = per-site migration to
>    editorial-from-news ONLY, never automatic. Pre-condition for regulated sites:
>    verify news writers read content_direction [UNVERIFIED].
> 3. **lendzy is going live** (owner moved NS; Nominet + delegation + Cloudflare
>    all confirmed; HTTP still times out — no origin wired). BEFORE it serves:
>    strip the acceptance_marker from seeded content_direction + rebuild
>    about.html (the marker copy IS in the live artefact), logo, register row.

**Cold start reading order:** this file → NOTES tail (2026-08-02 late night onward —
the misstep log for the seam work) → `SUMMARY_2026-08-02d_first_two_seams_shipped.md`
→ the 08-02 handoff only for the deeper experiment background. Register discipline
unchanged (`REGISTER_positioning.md`, `check_register.py` must pass).

## What is PROVEN (do not re-derive)

1. **Seam 1 (every-page invariants) is CLOSED and verified on the artefact.**
   `footer-theme-chrome` (all 14 sites' footer) carries a gated
   `{{if .compliance_lines}}` block; value = `config.chrome.compliance_lines`
   (site_specs aspect `site_config`); **second consumer of STY-050, registered as
   STY-051**. Byte-identity for unset sites is pinned by
   `footer_compliance_lines_test.go` (old constant ≡ pre-edit live row by md5;
   new constant ≡ applied SQL). Lendzy census: **18/18 pages, both mission lines +
   block**, sites repo `b1d7c98ad`, scripts stripped. Council corr `56ab6e23`
   APPROVED r1. Commits `6e8098022` (seam) `7425bc4c4` (verification).
2. **Seam 2 (canonicals + meta-desc correct-or-absent) is LIVE on v1.0.1238 and
   verified.** Pod-grep both replicas (positives: `injectCanonicalLink`,
   `resolvedValueSatisfiesDeclaredType`, "no canonical emitted"); artefact proof on
   about.html post-roll: canonical present, blank description tag REMOVED. Corr
   `4cffcebb` APPROVED r1. Commits `9c7a8e9e4` (seam) `2046b6975` (type guard —
   a non-array in an array-declared schema field is refused loudly; landmine filed).
   **Other lendzy pages gain canonicals at their next rerender** — no sweep fired
   (deliberate). One `needs_rerender` item (shape below) sweeps the site if wanted.
3. **Seam 3 (favicon) is NOT a platform seam.** Discovery files
   `needs_brand_head_assets` → asset-deployer → `derive_brand_head_assets`
   (derives favicon+OG from the LOGO). Lendzy has **no logo asset and empty
   sites.logo_url** — imagery-lane work. Residual: confirm discovery sweeps a
   shadow site before relying on auto-filing.
4. **Seam 4 is REFUTED — the tool handler was never broken. Do not route work at
   it.**
   > **CORRECTED 2026-08-03 (same day this file was written):** this section
   > originally said the handler "published the tool's SELF-CONTAINED document
   > (no site chrome)" and pointed at a pending 090 verdict. Both the original
   > 08-02 claim ("queues NO rerender") and that reframing were WRONG — two
   > stacked wrong calls, both now in `WRONG_CALLS.md`. The 090 run returned
   > UNVERIFIABLE (scope-not-narrowing, corr `f2252404`); completing its
   > "still needed" list first-hand refuted the seam: the handler's own deploy
   > committed the FULLY-ASSEMBLED page (sites commit `626c8e15d`, 19:33:50Z
   > 08-02 — header, nav, footer verified in the blob; the "self-contained"
   > reading came from a 2KB preview of the 28,878-char record), and the
   > hand-fired 19:42Z rerenders produced NO commit because there was nothing
   > left to do. Full chain: NOTES 2026-08-03 late morning.

## Live identifiers

- site `8ff093d5-1f19-453b-9439-a10379bbcd76` (lendzy.co.uk) · chassis **v1.0.1238**
  (rolled by owner ~10:08Z 08-03; carries seams 2 + type guard)
- Council corrs: seam 1 `56ab6e23-0d29-4bc1-96df-5252fdb759e7` APPROVED ·
  seam 2 `4cffcebb-9774-45e9-971c-0f057058f795` APPROVED · seam 4 diagnosis
  `f2252404-257b-49a1-bf3d-6de8b5b294b0` (running at handoff time)
- Register: STY-050 (mechanism) · STY-051 (compliance-lines key) in
  `docs026_concept_register/register/styling-render-pipeline.md`

## The remaining backlog, in order

*(Seams 4 AND 5 closed 2026-08-03: 4 refuted — see the CORRECTED block above;
5 already fixed by the 097 lane's content_data link resolver (live 1229) and
confirmed by census — 0 dead internal targets across all 18 current artefacts,
NOTES midday entry. Of the original seven seams, the two genuine platform gaps
(1, 2) are shipped+verified by this lane; 3/4/5 dissolved under measurement.)*

1. **Seam 6** — planner imposes the standard shape (blog) past the mission;
   015-class. **Get the owner call on the 015 retype arm first** (retyping
   blog-index section-index→news-index hands the blog to the NEWS pipeline —
   wiring ongoing news generation onto sites is a decision, not a fix). Then the
   platform half: the planner honouring a mission's negative space ("no blog").
2. **Seam 7** — in-browser tool fixtures [UNMEASURED]; blocked on serving (lendzy
   deliberately has no zone).
3. Optional quick win: canonical sweep of lendzy (one needs_rerender item).

## Cleanups owed on lendzy (unchanged)

- acceptance_marker instruction still LIVE in seeded `content_direction` — strip +
  re-seed + rebuild about.html before lendzy is more than a shadow.
- No logo (blocks favicon derivation). No register row (deliberate, shadow).
- vetcomparison is the named candidate third consumer of STY-051 (its
  content_direction asks for footer disclaimers in prose) — its lane's call.

## Owner queue (unchanged)

Build order across 43 propositions (now unblocked); 2 residual insurance twins;
fleet-wide www/HTTPS decision — **it now moves the canonical AND JSON-LD identity
together** (adjacent in rerender_single_page_action.go, byte-identical by
construction); FCA citation pass owed by loancash guides and lendzy tools before
any regeneration.

## Command shapes that are proven (copy, don't re-derive)

- DB: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`
- **Site-wide chrome+pages sweep**: INSERT site_work_items (site_id, source,
  item_type='needs_rerender', severity='medium', summary, spec=
  jsonb_build_object('reason','<why>','refresh_site_components',true), priority=90,
  handler_agent='rerender-pages', status='triaged', created_by, item_key). The
  parent completes in ~2 min while per-page children trickle (~30s/page via
  git-adapter) — **census the artefact, never trust the parent item**.
- **One-page assemble rerender**: item_type='page_rerender', handler='page-rerender',
  spec {domain,page_id,filename,page_name} — NO spec.reason (assemble mode), NEEDS
  page_id. ~90s to artefact.
- **Pod-grep a roll** (both replicas, same exec):
  `kubectl -n ai-persona-system exec <pod> -- sh -c 'strings /app/agent-chassis | grep -c "<symbol>"'`
  — a positive control proves the pipeline; pick negatives from removed STRING
  LITERALS (comments never reach the binary).
- **Census**: python over `git show origin/master:<path>` from ~/projects/sites,
  strip `<script>` before text checks (grep is line-bound and lies silently).
- Dispatch discipline: nothing within ~300s of a chassis pod start; find runs by
  payload not printed id.

## Session hygiene notes for the next thread

- The tree carried another session's mid-edit WIP yesterday (page_role_upsert.go)
  — verify YOUR changes against `git archive HEAD` overlay when the tree won't
  compile.
- Both seam commits carry `Council-Submitted:`; both corrs are APPROVED and READ —
  098 credits automatically, forward-only forbids amending trailers in.
- The concept-index headline was re-based to the measured 1,706 on 08-02 (the
  1,711 did not reproduce); run the documented grep, don't inherit.
