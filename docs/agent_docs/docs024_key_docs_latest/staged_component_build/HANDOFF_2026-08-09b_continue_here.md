# HANDOFF — staged_component_build, 2026-08-09b (fresh chat starts HERE)

**Supersedes `HANDOFF_2026-08-09_continue_here.md`, `_2026-08-08_`, and `_2026-08-05b_`
as the cold-start** — this file consolidates all three plus the D11 ruling so a fresh
chat reads ONE document. The older files stay as history; nothing here contradicts them
except where dated. **Two sessions have worked this lane in parallel (08-08/09) — run
`git log --oneline -10 -- docs/agent_docs/docs024_key_docs_latest/staged_component_build/`
before trusting ANY state in this file, including the tallies.**

## 1. Orientation reads

| doc | why |
|---|---|
| `README_where_we_are.md`, last four entries | plain prose, both sessions' accounts |
| `SUMMARY_2026-08-09_the_contracts_start_catching_things.md` | current milestone read-out |
| `PLAN_2026-07-30_staged_component_build.md` — D10, **D11** | scope + the two owner rulings |
| `NOTES_staged_component_build.md`, `## 2026-08-05` onward | every batch table, CID, and line rule |
| `bugs_open/228` §"CONTRIBUTION 2026-08-09" | the fake-contact-form fix, what is live, the ownership overstep, and the ONE decision left (the owning lane's) |

## 2. State (as of 2026-08-09 — re-verify, populations move)

- **56 subjects proven end-to-end** (54 sections + 2 tools): fence authored from the
  served page, every check watched red under its own mutant, persisted with byte-identical
  read-back, S6-green in-cluster with the wrong-page negative control refused. Batch
  tables with every CID: NOTES 08-05 → 08-09.
- **First real catches:** 228 (contact "sent" was a `setTimeout` — fixed, class fix live
  on v1.0.1274, all 15 served contact pages deliver); roi-estimator mobile overflow;
  the tracker feeds that have never once served their client refresh.
- **8 fences authored but UNPROVEN/UNPERSISTED** — blocked on site chrome 404s (§4-A):
  archetype-grid, directory-listing, funding-fit, patent-check, game-master-explanation,
  platform-comparison, people-feature-block, social-proof. Bar stays watched-red.
- **1 fence proven-but-unpersisted**: tool-ai-agent-roi-estimator (blocked on its own
  CSS defect, §4-C).
- **fce's S6 deferred** on the gaswholesalers logo (in its PLAN's Known state).
- **Interactive sections: 4 left, each gated** — audience-check-form (needs a POST-stub
  prover sibling + coordination with `idea_uk_vm_site` before firing its real funnel
  POST); adoption-tracker-listing + protocol-tracker-listing (feeds 404, CONTRIB filed
  with `model_directory_pipeline`); gauntlet-interface (lane-owned, coordinate).
- **~10 ready tools** — re-run `CHECK_naming_contract.sh` + the census (RUNBOOK §13).
- **Listings, not fences:** ~35 zero-placement sections; ~24 pageless tools (BROKEN-A
  guard: a fence without a serving page hard-errors); **six drift rows** (ported-page
  ×58 on lmc/loancash, featured-content, pricing, leopardess blog.html, contact-block on
  finetuning case-studies).
- **Lane-owned, coordinate don't fence:** gauntlet-interface/-cta/lobby-grid/
  gauntlet-round-record (also a subject_type mismatch); provocation-card/
  provocations-archive-list/evidence-timeseries/swipeable-insight-carousel.

## 3. THE LINE (complete — no other file needed)

Per batch of ~6–10 subjects (~9 min/static; 30–45 min/interactive):

1. **Placements:** single-instance per page (`HAVING count(pc.id)=1`); curl-verify the
   served page carries the component markup (drift is real); probe every relative asset
   AND every CSS `url('/...')` background for 200; avoid known-404 sites until §4-A
   lands; lendzy.co.uk's origin is flaky (522s) — not a proof site.
2. **Fences** (python-generated from the template read): root selector from the template
   (some components render NO data-component — generic-text-block, content-listing,
   featured-content, pricing resolve by class; components SHARING a class — team-section,
   social-proof-section — resolve by attribute and drop grid-child asserts). Assert
   unconditional headings (`\S`), never conditional ones; `{{range}}` grids get ≥1-item
   checks only when unambiguous. Mobile = status/overflow/console (+ has_visible_area
   both profiles).
3. **INTERACTIVE rule (not optional):** *a fence must carry at least one check a STATIC
   render cannot satisfy, or it certifies a dead panel.* Read the JS; assert one effect
   no server render produces. Four proven shapes: fetch fills a hidden slot (InnerText of
   hidden = ""); fetch creates an element; click sets an attribute the server never ships
   (`aria-current`); click sets a class — but click a value the server does NOT pre-mark
   (`.active` ships server-side on the default). Prove with an inert-script mutant
   (`<script>/* removed */</script>`, never a 404 — that adds console collateral).
   **Qualify first:** interactive only if the JS binds selectors in its own template AND
   a served page loads the script AND the effect is observable and safe to drive —
   `length(js_content)>0` fails three ways (game-list, trackers, model-directory).
4. **Trial:** `try_fence.go` — all-evaluated-all-passed, arithmetic reconciled.
5. **Prove:** `prove_fence_mutants_file.go <fence> <mutants.json> <url>` — every `from`
   verified to occur EXACTLY ONCE in the served page. Optional declared accommodations
   (object-form mutants file): `serve_local` (same-origin JS fetches the redirect
   harness would CORS-break — REQUIRED for feed components, and it disarms feed-side
   mutants: attack the script or slot, never the data) and `strip` (third-party beacons).
   Interactive traps: never delete server-rendered list items (N occurrences + the
   script re-renders anyway — rename the container class the SELECTOR uses, not the id
   the SCRIPT binds); check ORDER is load-bearing on one shared page — a later check
   continues from its predecessor's state (say so in the PLAN).
6. **Persist:** `scripts/gen_component_plan_sql.py <manifest.json>` dry-run then
   `--apply` (DO/RAISE length asserts inside the transaction; single `%` in RAISE
   formats). Read back and diff the fence. `manifest_batch6.json` is the worked example.
7. **Dispatch:** components `DISPATCH_s6_component_run.sh <site> <domain> <fn> <page_id>
   <bad_page_id>` (success lands `neg_control_confirmed_red`); tools
   `tool_acceptance_run.sh <site> <domain> <function>`. Read `acceptance_verdict` and
   skip REASONS, never counts.

**Before ANY dispatch:** pod-grep the DEPLOYED browser-runner's vocabulary — check types
AND step actions — with LONG runtime strings (short literals compile to immediates;
`grep -c "has_visible_area"` → 0 is NOT a miss; use `"non-numeric w/h in result"`).
The offline harnesses run HEAD's evaluator: newer vocabulary passes offline, fails/skips
live, and the `improve_tool` item raised is a FALSE POSITIVE (cancel with reason in
`result`; precedent `6c06b0ad`). Re-verify the placement row AND served markup
immediately before dispatch. No dispatch within ~300s of a chassis restart.

**Rerenders:** only via `scripts/RERENDER_page.sh <site> <domain> <page> [reason]` — a
reason-less rerender re-stitches STORED section HTML (template changes silently no-op
while assets republish), and `page_name` must sit at `input_data.spec.page_name` or
sections are discarded as `{"skipped":true,"success":true}`.

## 4. D11 — THE WORK PROGRAMME (owner 2026-08-09: fix through the framework's checkers
and handlers; prevent in initial builds)

**Guardrails:** cross-cutting causes through the 090 loop BEFORE the fix (or state the
first-hand substitute); platform Go through the council gate; `who-owns.py` + live
transcripts before touching 155/168/201/149/228 — all have prior claims, and this
lane's own §0 of the 08-09 handoff shows what skipping that costs.

**4-A. Asset-404 repair gap** — detection has worked since 07-31 (`image_url_404`,
bugs_closed/128), repair has never dispatched (items flag-only BY DESIGN — a decision to
revisit, plus a deploy path to make trustworthy). Order: (1) 090 the unmeasured
mechanism — why files are absent at served paths while `assets` rows sit active;
(2) contribute the ≥7-site evidence into `bugs_open/155` (deploy resolves source by
purpose, not asset_id) and `168` (DeployedWebPath) — landmine: brand-head purposes
resolve via `storage.BrandHeadAssetPaths`, never DeployedWebPath; (3) wire a repair
handler for `image_url_404` findings; (4) **prevention:** the site build gains a
post-deploy check — brand-head assets must HTTP-200 before the build reports success.
**Releases: 8 blocked fences + fce's S6 + defect-list entries 2/3/4.**

**4-B. Placement drift** — no checker compares `page_components` rows to the served
artefact; six drift rows accumulated unseen. 090 FIRST (which write path leaves rows
behind — unmeasured); then an additive, opt-in placement-vs-artefact discovery check
(normal council gate per the 2026-07-29 ruling; concept register same commit; flag-only
until the delete-vs-rebuild decision is taken on evidence); then gate the guilty path.

**4-C. Component CSS defects** — article-body (no `pre/code` overflow; 49 placements —
measure which carry wide code first) and roi-estimator (`h3.roi-inputs-title` fixed
297.9px; **one line, and it releases an already-proven committed fence**). Both:
template fix (DB, live for new renders) + reason-carrying scoped rerenders via
`RERENDER_page.sh`. Checkers need NO change — the fences caught both. **Prevention:**
the component generator's authoring standards gain overflow containment for
content-bearing containers.

**4-D. Broken tool pages** (gas-unit-converter empty-slotted; ab-test-calculator raw
`{{.placeholders}}`; equity-release active row/404 URL) — the queue holds the items;
the defect is their fate: `needs_content_page` FAILED (mechanism plausibly
`bugs_open/201` — read, check owner, contribute the two cases), `needs_page` closed
`wont_fix` (surface for an explicit unpark — never override a recorded wont_fix), rest
parked in review. Re-dispatch through the 082 pipeline once the writer path works
(no hand-authored content — owner ruling 2026-08-04). **Prevention:** a build must not
COMPLETE with schema-required fields empty — `bugs_open/149` C1/C3's claims-gate class;
contribute there, work its order, do not fork.

**4-E. Suggested sequence (cheapest unblock first):** roi CSS one-liner → 4-A asset
programme → 4-D writer path → 4-B drift checker → remaining fences (§2) throughout.

## 5. Standing defect list (verified at the artefact; dates matter — re-verify)

1. ~~228 fake contact send~~ **FIXED + LIVE, 15/15 pages** — only the JS choice remains
   (owning lane's call; two implementations, prover `prove_contact_delivery.go`).
2. gaswholesalers logo.png 404, every page, 4+ days (assets row active since March).
3. hero.jpg 404 family — ≥7 sites incl. vetcomparison.uk (which 128's 07-31 list missed).
4. finetuning.uk/index.html 404s five case-studies-grid card images.
5. article-body `pre/code` overflow (§4-C).
6. Broken tools (§4-D).
7. roi-estimator fixed-width h3 (§4-C).
8. Tracker feeds 404 ×4 on ai-agent-orchestration (CONTRIB'd to model_directory_pipeline).

## 6. Instruments (all committed under `staged_component_build/scripts/` — the
scratchpad is wiped between sessions; nothing citable lives there)

`try_fence.go` · `prove_fence_mutants_file.go` (+serve_local/strip) ·
`gen_component_plan_sql.py` + manifests · `DISPATCH_s6_component_run.sh` ·
`RERENDER_page.sh` · `prove_contact_delivery.go` · `probe_mailto_form_encoding.go` ·
`apply_contact_form_delivery.py` · `contact_block.js`/`contact_form.js` ·
per-subject `fence_component_*.json` + `mutants_component_*.json` ·
`CHECK_naming_contract.sh` (lane root).
