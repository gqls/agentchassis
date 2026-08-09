# HANDOFF — staged_component_build, D10 continuation (fresh chat starts here)

**Supersedes `HANDOFF_2026-08-05b_continue_here.md`.** That file's §4 work-list is still
the right shape; this one carries what has since closed, the one new line rule the
interactive pile forced, and the traps that cost time.

## 1. Read these first

| doc | why |
|---|---|
| `README_where_we_are.md`, last two entries | plain prose, fastest way in |
| `NOTES_staged_component_build.md`, the `## 2026-08-08 (later, fresh session)` entry | batch 6's table, both authoring traps, all three finds |
| `HANDOFF_2026-08-05b_continue_here.md` §3 (the line) and §5 (standing defects) | unchanged and still binding |
| `PLAN_2026-07-30…md` — **D10** | scope guards: no fence for a pageless tool; levels beyond section/tool out of scope |
| RUNBOOK §8–§10, §13 | unchanged |

## 2. State

**51 subjects proven end-to-end: 49 sections + 2 tools.** All S6-green in-cluster with the
negative control confirmed red. Batch 6 (2026-08-08 evening) added the first five
INTERACTIVE sections: `news-listing`, `latest-news`, `case-studies-grid`, `contact-block`,
`blog-listing` — 8 checks each, 8/8 mutants caught, S6 12/12. Batch 6b (2026-08-09, same
session) added `game-list` (7 checks, 7/7, S6 11/11) and `ai-readiness-quiz` (9 checks,
9/9, S6 13/13). All persisted and read back byte-identical; every skip a profile gate.

Fleet at the time of writing: chassis + browser-runner **v1.0.1270** (it rolled twice
during the session — re-grep, do not carry a version forward). Census: 112 active
sections / 43 with a PLAN; 60 active tools / 23 with. Re-run it; it moves.

## 3. THE LINE — one addition, and it is not optional for this pile

Steps 1–6 of `HANDOFF_2026-08-05b` §3 are unchanged. **New rule for anything with
`js_content`:**

> **An interactive fence must carry at least one check that a STATIC render cannot
> satisfy.** Every structural check still passes with the component's script deleted, so a
> fence made only of those certifies a dead panel. Read the JS, find one observable effect
> that exists in no server render, and assert that.

Worked examples, all four shapes seen so far:

| shape | assertion | why it discriminates |
|---|---|---|
| fetch fills a hidden slot | `#news-listing-count` matches `\d+ item` | footer is server-rendered `display:none`; `InnerText` on a hidden element reads `""` |
| fetch creates an element | `#news-footer a.news-more-link` exists | server renders that div literally empty |
| click sets an attribute | `.csg-filter-btn[data-filter="strategy"].active[aria-current="true"]` | served page has **zero** `aria-current` |
| click sets a class only | `.bl-filter-btn[data-filter="cat1"].active` | **click `cat1`, NEVER `all`** — `all` ships `.active` server-side and would pass with the script deleted |

Prove it with an inert-script mutant: replace `<script src="…"></script>` with
`<script>/* removed */</script>`. Not a 404 — a 404 adds console-error collateral.

**Two traps this pile has that the static pile did not:**

1. **You cannot kill a JS-rendered list by deleting the server-rendered items.** The `from`
   string usually occurs N times (the prover demands exactly 1), and the script re-renders
   from the feed regardless. Rename the **container class the fence's selector goes
   through**, leaving the `id` the script binds to alone.
2. **`serve_local` for the feed is required AND it disarms feed-side mutants.** Without it
   the same-origin `fetch("/data/…")` goes cross-origin under the redirect harness and CORS
   reds `no-console-errors` on a page that is clean live. With it the real feed is served
   verbatim in every run — so every driven mutant must attack the SCRIPT or the SLOT, never
   the data.
3. **Check ORDER is load-bearing on a multi-step component, and a later check cannot re-do
   an earlier one.** The evaluator drives every check against ONE shared page in
   declaration order. `ai-readiness-quiz`'s second gesture check was first authored
   self-contained (click Start, then click an option) and **failed on first trial**: Start
   was already hidden, and the click burned the full 30 s timeout. Let a later check
   continue from the state its predecessor left — which is also what a visitor does — and
   say so in the PLAN so nobody "fixes" it back.

**Persist with the committed generator now** — `scripts/gen_component_plan_sql.py
<manifest.json>` (dry run) then `--apply`. Write the manifest beside it
(`manifest_batch6.json` is the worked example). Batches 1–5 hand-rolled this in a
scratchpad and lost it every time.

## 4. Remaining work, itemised

- **~10 interactive sections left** (`game-list` and `ai-readiness-quiz` are DONE), all
  single-placement:
  `tool-ai-vendor-trust-checklist`, `tool-archetype-taster-quiz`, `adoption-tracker-listing`,
  `tool-gripper-cycle-time-estimator`, `audience-check-form`, `model-directory-listing`,
  `protocol-tracker-listing`, `report-request-form`, `tool-ai-agent-roi-estimator`.
  Budget ~30–45 min each; the four shapes in §3 cover most of them. `gauntlet-interface`
  (40 KB of JS) is **lane-owned — do not fence without coordination.**
  **Check first whether the subject is interactive at all.** `length(js_content) > 0` is
  what puts a component in this pile and it was WRONG for `game-list`: its script binds
  `.gl-filter-btn` / `#gl-load-more-btn`, neither of which exists in its own template, and
  no page even loads it. Two greps settle it before you budget 45 minutes —
  `grep -c '<selector the JS binds>' <html_template>` and
  `curl -s <page> | grep -c 'tools/assets/<fn>.js'`.
- **~10 ready tools** — re-run `CHECK_naming_contract.sh` + census first, the list moves.
- **8 chrome-blocked sections** (fences authored + committed, baseline cannot go green):
  archetype-grid, directory-listing, funding-fit, patent-check, game-master-explanation,
  platform-comparison, people-feature-block, social-proof. **One asset fix per site
  releases its subjects** — still the highest-value small repair on the board.
- **Lane-owned, coordinate first**: gauntlet-interface, gauntlet-cta, lobby-grid,
  gauntlet-round-record; provocation-card, provocations-archive-list, evidence-timeseries,
  swipeable-insight-carousel.
- **Listings, not fences**: ~35 sections with zero active placements; the drift rows —
  placement rows whose SERVED page carries no such component — now **six**: ported-page's
  58 on lmc/loancash, `featured-content`, `pricing`,
  `leopardessconsulting.co.uk/blog.html`, and `contact-block` on
  `finetuning.uk/case-studies.html`.

## 5. Standing defect list for the owner

Carried from `HANDOFF_2026-08-05b` §5, plus tonight's:

1. **`bugs_open/228` (NEW, and the most serious this lane has found)** — `contact-block`
   prints "Your message has been sent" from a 1,200 ms timer and has **no transport at
   all**. **TWO** live pages (`robot-hands.com/contact.html`,
   `leopardessconsulting.co.uk/ai-readiness-quiz.html`); the bug's first draft said three
   and was corrected the same day — see `WRONG_CALLS.md`. Needs an owner call on
   fix candidate 1 (give it a real destination, reusing `contact-form`'s or
   `audience-check-form`'s mechanism) vs 2 (remove the form, keep the contact details).
2. gaswholesalers.com: every page 404s `/assets/images/logo.png` (assets row exists).
   **4+ days.** fce's S6 stays deferred behind it.
3. The `hero.jpg` 404 family (`bugs_closed/128`, measured 07-31, **still serving**) — at
   least 7 sites, incl. vetcomparison.uk which 128's own list missed.
4. **NEW:** `finetuning.uk/index.html` 404s five `case-studies-grid` card images
   (`/assets/images/case-study-*.jpg`). Same detected-never-repaired class.
5. `article-body` ships no `pre`/`code` overflow CSS — code-bearing articles scroll
   sideways on mobile.
6. Broken tool pages: tool-gas-unit-converter, tool-ab-test-calculator (idea.uk serves raw
   `{{.placeholders}}`), tool-equity-release (active row, 404 URL).

## 6. Before you dispatch anything

- Pod-grep the **deployed** browser-runner's vocabulary, not HEAD's — the offline harnesses
  run HEAD's evaluator, so a fence using newer vocabulary passes offline and skips live,
  and the `improve_tool` item that raises is a **false positive** (cancel with the reason
  in `result`; precedent `6c06b0ad`).
  **`grep -c "has_visible_area"` returning 0 is NOT a miss** — short literals compile to
  immediate comparisons. Grep that check's own long error strings instead
  (`"non-numeric w/h in result"`, `"no element matches"`).
- Re-verify the placement row AND the served markup immediately before dispatch.
  Placements move, and a row whose page carries no markup is drift, not a subject.
- No dispatch within ~300 s of a chassis pod restart — the spawn is silently dropped.
