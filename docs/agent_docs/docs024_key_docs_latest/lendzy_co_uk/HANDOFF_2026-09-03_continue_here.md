# HANDOFF — lendzy.co.uk lane. Continue here.

**Written 2026-09-03 ~19:10Z. First handoff for this lane — no prior file to supersede.** Lane dir:
`docs/agent_docs/docs024_key_docs_latest/lendzy_co_uk/`. Site: `lendzy.co.uk` = site id
`8ff093d5-1f19-453b-9439-a10379bbcd76`. Read NOTES (a)–(t) + 09-03 (a)–(c) for evidence-grade
detail on everything below. Owner log: `README_where_we_are.md`. Milestone read-out:
`SUMMARY_2026-09-02_lendzy_co_uk.md` (still current — nothing since has changed the five-headings
answer enough to warrant a new one; write `SUMMARY_2026-09-03…` only at the next real inflection).

## 0. State in one paragraph

**Every item from the owner's 2026-09-02 brief is CLOSED, verified at artefact or verifier
strength, none of it stale.** The lane opened 09-02 when the owner made lendzy its own lane and
gave five instructions: keep working tools and add where needed; find the root cause of three
"phantom never-built" pages; build the 47 stale-link items; make "checked against the FCA
handbook, rule by rule" true; and (once found) fix two wrong rule citations, propagate the FCA
programme to sibling sites, and gate the compliance sentence behind a real checker. All five are
done. Migrations **693** (tool adoption), **695** (evidence register), **696** (citation
correction) are APPLIED and artefact-verified. The 47 link items are **0**, closed by their own
verifier on 09-03's revalidation tick with evidence on each row. The sitemap carries **30** pages.
The FCA register's first daily check ran 09-03 ~09:08Z: **8/8 facts re-verified, zero drift**, and
every unmodelled field (rule numbers, correction records, banned patterns) survived the production
writer's round-trip. **Nothing is currently in flight for this lane.** The next session's job is
NOT continuing yesterday's work — it is either (a) new work the owner assigns, or (b) the two
standing items in §2.

## 1. What is DONE, and the evidence (all re-measured 2026-09-03)

| | evidence |
|---|---|
| **Three "phantom" tool pages fixed** | mig 693, adoption not regeneration (owner: "keep the tools"). `deployed_at` stamped 09-02 16:06–07Z; artefact inputs unchanged 3/1/2; **template-path proof completed 09-02 21:33Z with files_sha256 == served sha256 on all three** — resolution proven by hash, not proximity |
| **Two wrong FCA citations fixed, fleet-wide** | mig 696. `CONC 6.7.17`→`6.7.23` (rollover limit), `CONC 6.7.23`→`7.6.12` (CPA limit) — fixed in pages, the `content_direction` spec (the re-planting source), the tool template, AND the loancash.co.uk fork of that tool. Re-verified served-bytes 09-02 and again 09-03: zero `6.7.17` anywhere |
| **Evidence register live** | mig 695. 8 FCA Handbook citations + 5 calibrated `banned_claims`, every quote verified through the PRODUCTION matcher (`cmd/fcaquotecheck`) pre-write. First daily pass (09-03 ~09:08Z): 8/8 re-accessed, 0 drift, all custom fields survived the writer round-trip |
| **The 47 unbuilt-link items** | **0** as of 09-03 16:06Z — each closed by `VerifyUnbuiltInternalLinkResolved` with `"arm":"verifier_resolved"` and its own reasoning on the row. Deliberately never hand-closed |
| **Sitemap** | **30** `<loc>`, all 9 tool URLs present (was 27, missing exactly the 3 phantom pages) |
| **Site health** | fleet census re-run in 693's verify: **0** active pages anywhere with no identified component — the founding defect's class is provably gone, not just these three instances |

## 2. What is OPEN — for the owner or the next session, not urgent

- **loancash.co.uk — the fleet's only register-less finance site with NO session.** It also served
  (until 696) the propagated wrong rollover citation. Nobody owns it. If picked up: RUNBOOK §8
  is the method, `cmd/fcaquotecheck` the tool, and read §8b–8g for five traps other lanes already
  paid for (Cloudflare-challenged hosts, UA-differential hosts, gov.uk founding-name slugs, the
  no-citation-exemption trap for rate patterns, the double-escape landmine).
- **The compliance sentence stays OFF the site** — owner ruling 09-02: it returns only once the
  claims-verification lane's rule-span checker (RFC_060 §3d/Q6, owner-approved to build 09-02) is
  running *regularly*, not merely merged. Check their lane's state before ever re-adding the
  sentence; this lane does not gate it unilaterally.
- **The FCA Handbook mirror (Phase B-ii)** is DESIGNED, not built (`PLAN` §B2/§B3b): cited
  chapters first, tables ready to widen, sequenced after pacing work on the shared citation-fetch
  path (`fetchCitationDocument`, currently unthrottled). Its host-admission rule is now settled:
  **the admission check must run the PRODUCTION fetch path itself, never curl** — a UA-differential
  host (nationaldebtline.org, moneyadvicetrust.org) serves a perfect page to curl and nothing to
  the real fetcher. Census of refused hosts is in RUNBOOK §8g.

## 3. Traps a fresh session would otherwise rediscover (full detail: RUNBOOK, NOTES)

- ⚠ **handbook.fca.org.uk returns HTTP 200 for EVERY path**, invented rules included — `<title>`
  is the only discriminator (LANDMINES.md entry, same footprint).
- ⚠ **`spec.reason` on a `page_rerender` item is PARSED against 5 literals, never read as prose**
  (`image_landed · section_data_resolved · cta_links_stale · template_changed · literal_markdown`).
  Anything else silently takes assemble-mode (re-ships stored bytes, reports success). This lane's
  first 14 rerenders (09-02) all did this — harmless for 696 (which wrote the artefact directly)
  but meant 693's resolution was unproven until the 09-02 21:33Z discriminating re-render.
- ⚠ **A council `RAISE EXCEPTION`/`RAISE NOTICE` with a parameter and no `%` placeholder is a
  PL/pgSQL *compile* error** — it aborts the whole wrapped transaction at the verify block, after
  everything else already ran. Audit placeholder-vs-parameter counts mechanically before applying
  (both 696 files hit this once; caught cleanly because of the single-transaction discipline).
- ⚠ **`doc_notes.subject_type` has a CHECK constraint** — allowed values are `tool / pipeline /
  experience / action / experience-pattern / landmine / component / decision`. `'site'` is NOT one
  of them (both 693 and 695 hit this; fixed to `'decision'`).
- ⚠ **A banned-claims pattern that fails to Go-compile degrades SILENTLY to a literal of its own
  source text** (`claims.go:348`) — armed, counted, inert, every count-based check passes. Verify
  patterns POST-APPLY in the exact consumer form (`"(?i)" + pattern`), not just at write time
  (RUNBOOK §8f).
- ⚠ **A watcher's baseline can be measured post-event** — a roll-detection monitor armed seconds
  after a chassis roll had a useless baseline. Read the baseline timestamp as data when a monitor
  reports its own arming.

## 4. Files of record

**Cold start:** this file → `SUMMARY_2026-09-02_lendzy_co_uk.md` → `README_where_we_are.md` →
`NOTES_lendzy_co_uk.md` (newest at bottom; (a)–(t) cover 09-02, 09-03 (a)–(c) cover today).

**The three migrations:** `docs/agent_docs/sql_for_agents/693_lendzy_adopt_three_orphan_tool_components.sql`
(+`_ROLLBACK`) · `695_lendzy_evidence_base_fca_handbook_citations.sql` (+`_ROLLBACK`) ·
`696_lendzy_correct_two_fca_rule_citations.sql` (+`_ROLLBACK`, backup tables `bak_696_*`) — all
three carry council correlation ids in their headers and were APPROVED (693 after 3 rounds, 696
after 1, 695 after 2).

**The method** (offered to and improved by loanzy.uk/farmerinsurance.uk/loancalculator.co.uk, all
of which registered same-day): `RUNBOOK_lendzy_co_uk.md` §8 (steps) / §8b–§8g (traps, each dated
and attributed) · `cmd/fcaquotecheck/main.go` (the production-matcher probe tool) ·
`docs026_concept_register/102_coverage_ratchet.txt` (`lendzy_co_uk` entry — three things named as
register material, not to be ratcheted, if Phase B-ii ever builds).

**Cross-lane threads still worth reading if picking this up cold:** `claims verification` session
(RFC_060, the Q6 checker, the register mechanism) · `bugs_open/357` (the sibling wrong-id tool
family, closed 0/22, cited 693 as its crib) · `components` session (the `spec.reason` finding,
`page_component_history` join warning: use `page_id`, NOT `component_id`, NULL on 44,555/45,285
rows).
