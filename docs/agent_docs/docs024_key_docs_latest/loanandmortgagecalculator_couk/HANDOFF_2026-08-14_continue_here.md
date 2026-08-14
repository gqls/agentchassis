# HANDOFF — Track B2 batch 1 CLOSED (16 of 21 calculators parameterised, oracle 170/0/6). START HERE.

**Written 2026-08-14 by the lane session after the final sweep.** Supersedes
`HANDOFF_2026-08-11_after_track_a_decisions_pending.md` as the entry point (that
file's rulings all executed or carried forward here). Everything in §1 was measured
against the live DB, the live domain and the sites repo at the moment of writing —
nothing inherited. Where a claim below goes stale fast, the re-check is one command
and it is given inline.

**Owner's governing direction (2026-08-13, verbatim):** *"All text and widgets need
to be editable so in future we can reuse them with their own slightly different copy
or mechanism."* That is what Track B2 builds, and 16 of 21 pages now have it.

---

## 0. The state in one paragraph

41 pages on `loanandmortgagecalculator.co.uk` (site id
`ed633ada-f8af-424b-b4d4-8af79160dbcd`). **19 prose pages** decomposed by Track A
(`["prose-0"]`, editable via the `ported-prose` content field). **16 calculator pages
in the B2 shape**: machinery in a per-page `content_components.html_template` that no
content writer can touch; every clean copy span a schema field (139 fields across the
16); rows **unlocked** (deliberate — see §2); pages still `rebuild_policy='owned'`.
**5 calculator pages still verbatim** (the mixed-card family, §4 — the next work).
**2 pages on the old Track-B shape** (`loans-consolidation`, `mortgages-repayment`:
locked verbatim tool row, zero fields — convert last, §5). Full oracle:
**PASS 170 / FAIL 0 / CONVENTION 6**, identical to the 08-11 pre-work baseline, with
parse + expectation controls fired in-session; class C invariants 23/0.

## 1. Live facts a fresh session must re-verify before acting (all cheap)

```bash
LANE=docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk
K="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"

# page shapes — expect 5 verbatim, 2 ptp-locked, the rest decomposed
$K -A -F$'\t' -c "SELECT CASE WHEN sections::text='[\"ported-page\"]' THEN 'verbatim' ELSE sections::text END, count(*)
  FROM pages WHERE site_id='ed633ada-f8af-424b-b4d4-8af79160dbcd' GROUP BY 1 ORDER BY 2 DESC;"

# B2 components — expect 16 rows, function == pages.name for each converted page
$K -A -F$'\t' -c "SELECT count(*) FROM content_components WHERE description LIKE 'Parameterised calculator component (Track B2%';"

# nothing locked on B2 pages; the two old-shape locks still present
$K -A -F$'\t' -c "SELECT p.name, count(*) FILTER (WHERE pc.locked_at IS NOT NULL) FROM pages p
  JOIN page_components pc ON pc.page_id=p.id
  WHERE p.site_id='ed633ada-f8af-424b-b4d4-8af79160dbcd' GROUP BY 1 HAVING count(*) FILTER (WHERE pc.locked_at IS NOT NULL)>0;"

# the 189/204 instruments at the current binary (chassis was v1.0.1294 on 08-13)
# and the sites repo: remote branch is MASTER, and your clone WILL be behind
git -C ~/projects/sites fetch origin && git -C ~/projects/sites rev-parse --short origin/master
```

**`PINNED_REF` in `decompose_lmc.py` is `7e6b993ef` — the last CLEAN pin.** Read §3
before changing it. The pin-matches-live guard runs only over still-verbatim pages
(5 remain), and **this lane's own deploys move the sites repo within the hour**.

## 2. The B2 design, and the two decisions inside it (both owner-ruled)

- **Machinery in the template, copy as fields.** `html_template` holds panel markup,
  grids, ids, inline scripts and the `calculators.js` tag; `input_schema.fields`
  holds every clean copy span (headings, labels, buttons, link texts, advisory
  paragraphs), each `required` with the original copy as `fallback` (the bug-238
  regeneration class) and per-kind `llm_guidance` — labels carry the load-bearing
  warning (renaming what a number means silently changes what users think they
  compute; the oracle asserts against captions).
- **Rows are UNLOCKED, deliberately.** `ApplySectionEditAction` REFUSES human-locked
  components (`section_editor_actions.go:305`), so "editable fields on a locked row"
  is a contradiction — this was yesterday's design corrected at the owner's request.
  Protection = the template (writers can only fill fields) + `rebuild_policy='owned'`
  (wholesale rebuild refused) + the `no_auto_fix` acceptance fences.
- **Editing works through the standard pipeline, proven end to end:** a
  `section_edit` work item, `spec = {domain, edit_type:'content_edit', page_name,
  page_component_id, field_updates:{...}}`, handler `section-editor`. Proven on
  `mortgages-simple`: edit → template re-render → live page → revert →
  **md5-identical to the source block**. Reuse = a second `page_components` row on
  the same component with different `content_data` — **not yet demonstrated**; it is
  the cheapest remaining proof if the owner asks.

## 3. ⚠ THE THREE TRAPS THIS ARC PAID FOR — read before touching anything

1. **`load_lmc.py --restore` is a TIME MACHINE.** Its backup table
   (`page_components_bak_20260805_lmc`) is dated **08-05 — before the `bugs_open/224`
   0% APR fix (08-08) and the btn-id fix (08-09)**. A restore + deploy reverted the
   224 fix on `loans/standard-calc` for ~2 days, the poisoned tip became the next
   session's pin, and `assert_pin_matches_live` **blessed it** (its premise is "live
   is truth"; a deployed restore makes live the lie). **Do not restore B2 pages from
   that backup at all** — the correct rollback for a B2 page is re-seeding from the
   sites repo at a clean pin. Before ANY restore:
   `git -C ~/projects/sites log --oneline --since=2026-08-05 origin/master -- loanandmortgagecalculator.co.uk/<page>.html`
   — non-empty output means the backup is stale for that page. Full entry:
   `LANDMINES.md` §"A restore from a dated backup is a TIME MACHINE".
2. **The manifest's `scripts` key is load-bearing.** `decompose_lmc` puts body-level
   inline scripts + the `calculators.js` tag in `b["scripts"]`, and `load_lmc.py:240`
   appends them at write time. `b2_build` now handles this (scripts ride literally in
   the template; render proven against `html+"\n"+scripts`) — but any NEW consumer of
   the manifest must do the same or it ships dead calculators, which is exactly what
   batch 1 did for an hour.
3. **Counts can be masked; identities cannot.** A script-COUNT check passed broken
   pages because assembly's injected JSON-LD `<script>` replaced the missing
   calculator script one-for-one. `b2_verify.py` (in the lane) is identity-based end
   to end — use it, not ad-hoc greps. And **the full oracle sweep is the mandatory
   closing gate of every batch**: it is the only instrument that measures behaviour
   rather than consistency-with-a-chosen-source. `WRONG_CALLS.md` 08-13/14 has the
   whole tally (dropped key, masked count, `2>/dev/null` md5 of empty-vs-empty, `>>`
   creating a nonexistent bug file).

## 4. NEXT WORK: the 5 mixed-card pages

`loans-damage-checker`, `mortgages-bridging-loan`, `mortgages-equity-release`,
`mortgages-fee-analyser`, `mortgages-rate-forecaster`. Each has **exactly one**
`div.card` whose children mix copy with machinery (a heading, sometimes an advisory
`<p>`, beside the grid), so the descent still dissolves the card and
`gate_wrapper_parity` rightly refuses them.

**The B2 answer (owner-ruled 08-13): take the card WHOLE and let the mixed copy
become fields** — that is what fields are for. The route that already worked for
`mortgages-simple` (same shape, done by hand as the proving page):

1. **Slice from source, explicitly per page** — from `<div class="card">` through the
   final `</script>` (assert exactly one card first; include the `calculators.js`
   tag). Clever rules failed twice on this class; explicit slices + gates beat them.
2. `b2_build`-style field extraction on the card (its h2/p become fields), render
   check with the REAL Go engine (`text/template`, `missingkey=zero`) — byte-identical
   or refuse.
3. Prose = the rest of `#content`, as `ported-prose` rows, byte-sliced.
4. Seed via `b2_load` conventions (per-page guarded transaction; function ==
   `pages.name` so the acceptance fence resolves; NOTHING locked).
5. Deploy (assemble-only `page_rerender`: `spec={page_id}` in spec AND column, NO
   `spec.reason`, status `'triaged'`), then `b2_verify.py <pages>`, then
   `oracle.py --tools <the affected tools>` + expectation control, then the FULL
   sweep before calling the batch closed.

⚠ **The existing `b2_seeds/` entries for these 5 are WRONG for B2** — they were built
from the manifest whose tool blocks lack the card. Rebuild wrapper-inclusive; do not
load them as they stand.

Then: **`loans-consolidation` + `mortgages-repayment`** (old shape: locked verbatim
tool row, zero fields). Convert = unlock decision + parameterise, same pipeline.
Consolidation's component (`function='loans-consolidation'`, 7,681-char template,
no fields) also carries the fence subject — keep the function name.

## 5. THE WIDER QUEUE (owner rulings standing, in order)

1. **Site-spec seed + planner loop** (owner D6 ruling 08-11): seed the spec, let the
   planner plan, reseed until the plan is *reasonably close to today's site*.
   **Constraints verbatim: the site must NOT shrink on rebuild; the exact
   calculator/guide mix is NOT important; growth from the improvement loop is
   welcome.** `site_plans` is still 0 rows. The re-slot trap is settled at the code
   level (`save_sections_positional_tool_slot_test.go`: positional slots match; the
   danger is a semantic plan that OMITS a tool slot) — a seeded plan must name every
   tool slot.
2. **Bug 252 og: half** — AFTER verifying the 251 canonical fix is live (fix commit
   `61abbdbd0`, `Council-Submitted: 33fb41cb`; chassis has rolled since — check a
   rerendered root's canonical, then read the council verdict for the trailer).
3. **Complaint-deadline oracle** (loancash) — FOS six-month + limitation rules,
   verified at source, never from the page. The FCA caps were checked 08-12 and are
   CURRENT; this is the one that moves.
4. **Track C** (loancash decomposition) after the mixed-5 prove which assertions are
   site-general.
5. **Stage 2's proof case stands untouched**: LMC homepage is missing 6 of its 16
   required links BY OWNER RULING ("leave it for stage 2 as proof").
   `gate_page_links.py` exits 1 on it deliberately; the fixture is committed under
   `acceptance/BASELINE_2026-08-12_*`. Do NOT "fix" it — stage 2 (the
   `copy_quality_two_stage` lane, owned by another session) must, via
   `section-editor`, without touching the round-7 register.

## 6. The toolkit (all in this lane directory, all with controls)

| tool | what it does | control |
|---|---|---|
| `decompose_lmc.py` | source → manifest (prose + tool blocks + `scripts` key) | pin-matches-live guard (verbatim pages only) + hard per-page assertions |
| `b2_build.py` | tool block → template + fields + schema, scripts literal | **render == block via Go's own engine**, per page, or refuse |
| `b2_load.py` | seeds → DB, one guarded transaction per page | DO/RAISE: backup exists, row count, tool md5, 0 locked, field count |
| `b2_verify.py` | live page vs pinned source, identity-based | verbatim render + script BODIES + exact `calculators.js` + class counts + id sets |
| `gate_wrapper_parity.py` | manifest must not drop a layout class | `--self-test` induces a shortfall, must fail 21/21 |
| `gate_page_links.py` | required_links present in a page's own components | `--self-test` injects an unlinkable URL, must fail |
| `oracle.py` / `invariants.py` | the arithmetic truth | `--selftest-parse`, `--mutate expectation/crosstool` — SAME session |

**Batch protocol, non-negotiable:** seeds byte-proven → guarded load → deploy →
`b2_verify` → per-tool oracle + control → **full sweep + controls before any report**.

## 7. Rollback inventory (corrected after the time machine)

- **B2 pages:** re-seed from the sites repo at a clean pin (currently `7e6b993ef`;
  re-verify cleanliness at the moment of use against the fix history, not against
  live). The 08-05 backup table is NOT a valid source for them.
- **Still-verbatim pages:** the backup table remains valid ONLY if the
  `git log --since=2026-08-05` check for that page comes back empty (for the mixed 5
  it does NOT for some — check per page).
- `_bak_index_rewrite_20260811` (homepage), `_bak_titles_negativity_20260812`
  (4 titles), `_mig377_relocked_tool_pages` (policy stamps) all still present.

## 8. Read in this order if starting fresh

1. this file
2. `bugs_open/263_…` — the whole Track B→B2 arc: defect, corrected rule, owner
   direction, the 08-13 design entry (lock-vs-editable), fix candidates' post-mortems
3. `NOTES_…md` 2026-08-13/14 entries — every misstep with its check
4. `LANDMINES.md` §time-machine + `WRONG_CALLS.md` 08-13/14 — the instrument-error family
5. `HANDOFF_2026-08-11_after_track_a_decisions_pending.md` — the owner rulings this
   file carries forward (all D1–D6 outcomes recorded there)
6. `copy_quality_two_stage/` — the stage-2 lane (another session's; feed it, don't fork it)

**Council note:** everything here is site content, lane tooling and DB config — out of
gate scope. The one platform change of the arc (251's `preferredPageURL`) is committed
with `Council-Submitted: 33fb41cb-768e-4e8e-b5fd-7a4d5ff75fa1`; read the verdict
before writing any `Council-Reviewed:` trailer.
