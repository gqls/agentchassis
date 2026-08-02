# HANDOFF — portfolio positioning lane, 2026-08-02 night

**Cold start reading order:** this file → `SCORECARD_2026-08-02_lendzy_attempt1.md`
(including its CORRECTED blocks — they are findings, not embarrassments) →
`RUBRIC_2026-08-02_loancash_benchmark.md` → `SUMMARY_2026-08-02b_…` →
`REGISTER_positioning.md` (P7 seats doctrine + the ```claims``` block that
`check_register.py` parses). NOTES is the misstep log; newest at the bottom.

## What is PROVEN (do not re-derive)

1. **The pipeline can build a positioned site.** lendzy.co.uk (shadow for the loancash
   L10 proposition): 18/20 pages built by the pipeline, committed to the gqls/sites
   repo `origin/master` (branch is **master**, not main). Tools built via the tool
   pipeline with correct regulatory arithmetic. Only dead internal target: the favicon.
2. **The divergence seam works end to end** (task #16 CLOSED): register row →
   `set_divergence_specs.py <domain> --apply` → `site_specs.content_direction.formatted`
   → writer prompt → saved copy. Proof: acceptance marker "checked against the FCA
   handbook, rule by rule" planted ONLY in the seeded spec, appeared in 3 regenerated
   sections of about.html (item `c9852314`, baseline 0, `pages.content_direction` NULL).
3. **The mission seam carries structure and facts well** (nav labels verbatim incl.
   deep-link "Free help now"; title carried the brand phrase verbatim; rule-name
   discipline held; zero unsanctioned figures in prose) — but nothing owns "on every
   page", and the standard shape leaks (it planned a blog unasked).

## Live identifiers

- site `8ff093d5-1f19-453b-9439-a10379bbcd76` (lendzy.co.uk) · plan
  `eda76fbd-491f-4acf-8fc9-0a699e5f3907` · submission corr
  `a62da5d3-3e08-450a-a981-bc002f7ea2cd`
- Chassis timeline TODAY: attempt 1 on v1.0.1228 · attempt 2 on 1229 · now **1231**.
- Blog: REMOVED per owner ruling 2026-08-02 (plan rows deleted, pages archived, items
  rejected). Do not resurrect. The 015-class finding (planner emits section-index blog
  templates the builder can never build) stays on the seam list below.
- lendzy open work items: **zero** as of 21:5xZ.

## ⚠ Cleanups owed on lendzy before it is ever more than a shadow

- **The acceptance_marker instruction is still live** in its seeded
  `content_direction` (and in the lendzy payload in `set_divergence_specs.py`,
  loanandmortgagecalculator_couk dir). Any rebuild keeps weaving the marker phrase.
  Strip the key, re-seed, rebuild about.html once clean (bug025 runbook §5 pattern).
- lendzy is NOT in the register (deliberate — shadow). If built for real: register row
  first, then replace the by-reference payload.
- No Cloudflare zone (deliberate — publicly unreachable). Sites-repo files + DB rows
  are experiment artefacts; tidy when the experiment ends.

## The seam backlog = "fix the pipeline" (owner's standing directive), evidenced

Priority order, each with evidence in the scorecard:
1. **Chrome-level carrier for every-page invariants** — compliance lines on 3/15;
   per-page writers cannot produce "on every page". Biggest quality gap.
2. **Canonicals not emitted at all** (0/15, any-attribute-order check). Cheap,
   fleet-wide SEO value. Head template also emits **empty meta description**.
3. **Favicon referenced, never generated** (`/assets/images/favicon.png`, 18 pages).
4. **Tool handler writes its component but queues NO rerender** — I fired
   page_rerender items by hand (INSERT shape below).
5. **Links ship to planned-but-unbuilt pages** (24 dead tool links in attempt 1).
6. **Planner imposes the standard shape** (blog) past the mission.
7. In-browser click-through of tool fixtures [UNMEASURED] — needs serving; static
   formula verification done (constants + formulas quoted in scorecard).
These are platform-code changes → **council gate per coherent task** (090/097 triggers,
`Council-Submitted:` trailer if committing before verdict), concept-register entry in
the same commit for any new shared seam.

## Commands that took effort (inline so the scratchpad can die)

- DB: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`
- Marker/invariant censuses: **run on the BUILT ARTEFACT** (git show from gqls/sites
  `origin/master`), never on page_components — titles/footers/compliance lines are
  added at assembly. Strip `<script>/<style>` before any number census. Use python for
  context extraction (grep is line-bound and returns nothing, silently).
- Rerender item (proven shape): INSERT INTO site_work_items (site_id, source,
  item_type, severity, summary, spec, page_id, affected_url, priority, handler_agent,
  status, created_by) VALUES (…,'rerender-pages','page_rerender','medium', …,
  jsonb_build_object('domain','lendzy.co.uk','page_id',…,'filename','about.html',
  'page_name','about'), …, 80, 'page-rerender','triaged', …);
- One-page rebuild through the writer (the #16 vehicle, bug025 runbook shape):
  item_type `needs_content_page`, pipeline `'build'`, handler `page-build-handler`,
  unique item_key, spec {page_name,page_id,name,page_type:'content',site_id,url}.
- Fresh build trigger: `bash scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh
  <domain> --mission-file <path>` (mission = PURE site brief; every line reaches the
  classifier). Live pipeline creates **needs_page** items (082's header saying
  needs_content_page is stale). Tool-role pages are parked by design
  (`reconcile_site_plan_action.go:37-40`) → route via `needs_tool_recreation`
  (handler `tool-recreation-handler`, spec shape: mode/source/page_name/page_type/
  interactive_features[{name,page,type,description,self_contained}]).
- Dispatch discipline: none within ~300s of a chassis pod start; find runs by payload
  not printed id; `kcat` exit 0 proves nothing — verify by the DB row.
- Shell traps that bit TODAY: a pkill compound must not contain the plain target
  string ANYWHERE in its command line (a sed operand re-triggered the self-kill);
  commit messages via `-F` file always; `git -C` when cwd may have wandered.

## Standing tasks & owner queue

- Task #21 (in_progress): the seam backlog above.
- Task #15 (pending, other site): decompose loanandmortgagecalculator's 13 guides →
  flip to generic (straight flip yields nested `<html>` — decompose FIRST).
- Owner queue: build order across 43 propositions (the machine now demonstrably
  works); 2 residual insurance twins; Always-Use-HTTPS + www policy (fleet-wide);
  product holds. Mortgagecalculator adoption is ANOTHER LANE (contribute via its
  handoff only).
- loancash citation pass against live fca.org.uk text still owed before any loancash
  regeneration; lendzy tool citations (e.g. DISP 1.6.2R) are [UNVERIFIED] the same way.

## Register discipline (unchanged)

One entry per proposition; exactly one buyer per phrase; money named per seat;
authority links down to buyer domains; `check_register.py` must pass after any edit;
apply positioning to a NEW site only via a register row + payload + gated seed.
