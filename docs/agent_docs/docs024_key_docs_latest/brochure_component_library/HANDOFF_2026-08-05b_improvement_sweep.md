# HANDOFF 2026-08-05b — fundamentallyai.com: the improvement sweep ran, and what it found

**This file does NOT supersede `HANDOFF_2026-08-05_continue_here.md`.** That one (written
12:49 by another thread) is the lane's cold-start for the camera / contact-sheet / 151-checker
front and stays authoritative there. **This file is a parallel front**: the owner asked for the
improvement sweep to be fired at fundamentallyai.com and for the site to be seen by the visual
designer and copy improver. Read both; they do not overlap.

Owner ask, verbatim: *"please fire the improvement sweep and fix anything that isn't working
for rebuilding and making the site better. Please can we include somehow that it gets seen by
the visual designer and copy improver and gets improved."*

## 1. The premise that opened the session was wrong, and that matters

The owner's opening belief was that **fundamentallyai.com had been hand-built rather than built
by the framework**, and the request was to recreate it through the pipeline keeping the spec.
That is **refuted. Do not rebuild the site.**

Evidence, all measured first-hand 2026-08-05:

- `site_specs` carries the full 082 chain, every row agent-authored: `submission` +
  `mission_brief` (`domain-submitter`, 2026-07-20 20:36) → `classification`,
  `content_direction`, `design_intent`, `identity` (`domain-research-classifier`, 20:39–20:40)
  → `vertical_landscape` (`vertical-exemplar-researcher`, 20:47) → `strategy`
  (`domain-strategist`, 21:00) → `briefing` (`build-briefing-agent`, 21:24) →
  `resolved_composition` (`site-design-planner`, 21:39) → `tools` (`tool-suggester`, 07-24).
- The `submission` spec's keys are exactly `domain` / `fidelity` / `email` / `mission_brief` —
  the FRESH payload `082_submit_domain_unified.sh` builds (`:135-138`).
- All 16 page rows have `built_from_plan_version` set.
- The `78d9d1aee` "every site goes through the framework" ruling concerned **webdesign.uk's**
  hand-written shopfront, not this site.

**What IS hand-made is the polish on top**: ~18 files in this lane's `sql/` (085b–086b, palette,
imagery wiring, evidence_base seeds) hand-applied to a framework-built site. That is editing,
not building — and it is the honest kernel of the owner's suspicion.

**Why the site nonetheless reads as un-designed:** the framework's own brief-fidelity audit ran
2026-07-24, filed 4 findings, and **nobody drained them for 12 days**. The defect was the
undrained queue, not the build path.

## 2. Pre-flight: seven stale rows cancelled with evidence

`run_improvement_sweep_once.sh`'s header requires this — triage promotes EVERY `detected` row
and dispatches it to a handler that changes live pages, so a stale row becomes live churn.
All 23 `detected` rows were checked one at a time against the live artefact; seven cancelled,
each with its evidence in `result` (SQL reproduced in the RUNBOOK):

| row | why cancelled |
|---|---|
| `image_url_404:brand-illustration` | serves HTTP 200, `image/jpeg`, 177,942 B, JPEG magic bytes |
| `missing_structure:rerender` ("14 pages missing header/footer") | all 14 live pages return `<header>`, `<footer>`, `<nav>` |
| `deactivated_header` | slot repointed 07-31 to `58fde68f` `header-theme-chrome`, `is_active=t` |
| `deactivated_footer` | slot repointed 07-31 to `e6347680` `footer-theme-chrome`, `is_active=t` |
| `content_rewrite` self-correction | premise ("planned, no components") false: serves 200, 4 components, names leopardessconsulting.co.uk 9× |
| brief-fidelity site-wide illustration | premise ("only 2 of 27 components contain images") false: 43 `<img>` across 14 pages, 18/18 distinct URLs return 200 with `image/*` |
| `needs_rerender` (empty spec, empty item_key, 07-24) | names no page and no reason; a reasonless rerender is assemble-only; 61 rerenders completed since |

**The reasoning to reuse: cancelling loses no signal**, because the sweep re-runs the full
audit chain and re-files anything still true against current evidence. That is what makes the
pre-flight cheap rather than risky.

**KEPT deliberately** (verified still true): `deactivated_head` (true — but see §5, fleet-wide),
3 × `stale_sc_*`, `content_rewrite:platform-log-index`, brief-fidelity template-repetition
(`model-fine-tuning` and `multi-agent-review-council` still share hero + Generic Text Block +
teaser-reveal-panel + call-to-action), 2 × `audit_tool`, 2 × `improve_tool` (`#error-msg`
genuinely absent from the deployed selector page — 0 refs), `orphan_blog_posts`,
`capability_gap`, 4 × `misdirected_cta`.

## 3. The sweep ran and completed

`SWEEP_CORR=d0430afd-3600-496e-9c87-9459e9787197`, fired 2026-08-05 12:13 UTC.
**14 orchestrations, all `COMPLETED` by 12:24, `error` NULL on every row.**
291 gate: `audit_due=true`, `not_converging=false`, fingerprint
`55ed76f6f41cd2deea08133b190878f4` → **the full audit chain ran**, not a skip.

**All 16 `detected` rows drained. Zero `detected` remain.**

## 4. The owner's actual ask is satisfied BY the sweep — no new mechanism needed

Read from live config, not inferred from step names:

- `improvement-loop` → `call_design_audit` → **`design-audit-agent`**, whose spawn steps resolve
  to `spawn_visual_auditor -> visual-design-auditor` and
  `spawn_content_auditor -> content-quality-auditor`.
- `improvement-loop` → `call_site_review` → **`site-review-agent`** →
  `spawn_content_auditor -> content-quality-auditor`, plus its own `run_strategic_review` +
  `write_strategic_findings`.
- Also in the chain: `call_design_discovery`, `call_quality_discovery`,
  `call_completeness_discovery`.

So **the visual designer and the copy improver are already inside the sweep.** Tell the owner
it is a standing faculty, not a one-off.

Re-prove it rather than trusting this paragraph:
```sql
SELECT k, COALESCE(default_config#>>ARRAY['workflow','steps',k,'config','agent_type'],'(none)')
FROM agent_definitions, jsonb_object_keys(default_config->'workflow'->'steps') k
WHERE type='design-audit-agent' AND is_active AND COALESCE(is_snapshot,false)=false
  AND deleted_at IS NULL AND k LIKE 'spawn%';
```

**Still does NOT exist: the offer and benefit analyser.** It is the B track of
`vigilant_designer_offer_analysis`, behind four other items; no agent matching
offer/benefit/critic/recompose is seated (checked `agent_definitions`). The nearest live
equivalent is `site-review-agent`'s strategic review, which DID run. **Do not tell the owner
the offer analyser ran.**

### What the sweep produced

**The copy improver worked** — 3 new `claims_unverified` rows from `quality-discovery-agent`
(12:14), the highest-value output of the run:

- `capabilities`: **4 unregistered numbers**
- `tools`: 3 unregistered stat fields
- `tool-review-council-simulator`: 3 unregistered stat fields

Plus a new `cta_names_unknown_destination` (12:16) and a `needs_page` re-render for `tools`
after an image asset landed (13:44) — the imagery path ran too.

## 5. What isn't working — actionable, priority order

### (1) `image-build-handler` cannot accept a `needs_logo` item: the repair path is DEAD
```
step call_logo_gen failed: ... input_mapping failed:
source path 'input_data.spec.image_prompts.logo' not found for field 'prompt'
```
The error dumps available paths; `input_data.spec` holds only
`check` / `original_pipeline` / `path` / `purpose`. The step maps `prompt` from a key the filing
detector never writes, so **any site whose logo that detector flags fails identically** — not
fundamentallyai-specific. The `input_mapping`-is-an-allow-list landmine firing for real.

`[UNVERIFIED]` I did not confirm which detector writes that spec shape, nor whether another
producer DOES supply `image_prompts.logo` (if one does, the fix is to make the handler tolerate
both, not to change the mapping). This is a cross-cutting claim about a shared handler, so by
CLAUDE.md's own rule it goes through `090_TRIGGER_needs_diagnosis_v1.sh` **before** being
asserted as a root cause in a `bugs_open/` file.

### (2) Two findings BLOCKED: "No handler_agent set — item cannot be routed to any agent"
`image_url_404:logo.png` and `capability_gap:content_duplication_rewrite`. The second is
by-design flag-only (151 candidate 1). **The first is not** — a real finding that can never
drain, which is this estate's signature failure.

### (3) …but that logo finding is probably a FALSE POSITIVE, which makes (2) worse
Nothing local references `/assets/images/logo.png`. The site's own logo is
`/assets/images/logo.jpg` (200, `image/jpeg`, 60,897 B). The only `logo.png` in served HTML is
`https://leopardessconsulting.co.uk/assets/images/logo.png` — an **external partner logo that
returns 200**. So a basename appears to have been attributed to a local missing path: the
`bugs_closed/128` family (purpose/basename vs path) resurfacing in a check 128 was meant to fix.
`[UNVERIFIED]` — detector not read.

### (4) `needs_content_page` FAILED on the spawn→call handshake
`step call_content_writer failed: workflow completed but its result could not be delivered to
the parent`. Known class (~half of all-history attempts). **Do not cancel pre-diagnosis.** Its
summary is itself a finding: *"The site claims 'more than ten live production sites'…"* — a
claims problem now stuck behind a delivery failure.

### (5) `audit_tool` on the selector: "Claim timed out (attempts exhausted)"

### (6) `capability_gap:hardcoded_section_colors` DEFERRED — "outside scope"

### (7) A LIVE broken link, independent of the sweep
`/guides/llm-cost-calculator-guide.html` serves
`<a href="/platform-log/index.html">Platform Log | FundamentallyAI</a>`; that page **404s** and
has been `status='active'`, `build_status='planned'`, `deployed_at IS NULL` since 2026-07-20.

Mechanism: the internal-link resolver and CTA recompute treat `status='active'` as "linkable"
and never test whether the page shipped. `queryresolve` already has the vocabulary —
`ListedPageEligibilitySQL` (`deployed_at IS NOT NULL` + non-empty `sections`),
`DeployedPageEligibilitySQL`, `FetchablePageEligibilitySQL` — and
`rebuild_blog_listing_action.go:111` uses it while `resolve_internal_links_action.go:440-500`
does **not**. Same "looser predicate than the thing beside it" shape as `bugs_closed/191`
and `bugs_closed/049`.

A fix is already queued: `needs_page` "Build platform-log-index page (not_built)", filed by
`reconcile_site_plan` 2026-07-20, sitting in `needs_human_review` ever since. Approving it is
probably right — but **REB-007**: `jsonb_array_length(sections)=0`, so a naive rebuild is a
**silent no-op that reports complete**. Needs a planner reconcile first, or the page retired.
Same for `owned_page_review` on `tool-decision-record`.

## 6. Traps this session paid for — read before trusting your own checks

1. **A `grep -c` over an unguarded `curl` body reads a failed fetch as a real absence.**
   `/contact.html` returned `header=0 footer=0 nav=0` and I nearly filed a live chrome defect;
   5 re-fetches gave `header=1 footer=1 nav=1` at 21,879 B every time. **Gate on byte count
   first**, then read the greps — an empty body and a missing element are otherwise
   indistinguishable.
2. **A poll loop that sends stderr to `/dev/null` cannot tell "not yet" from "my query is
   invalid".** A 20-minute Monitor on the sweep printed nothing and I read it as a lost
   dispatch. In fact `orchestration_states` **has no `agent_type` column**, so every iteration
   errored silently — while the sweep had already completed. Validate the query by hand once,
   and never silence stderr in a watch.
3. **Hours pass between turns on this tree.** I fired the sweep at 12:13 UTC, reasoned as if
   minutes had passed, then misread 12:19 timestamps as impossible. `date -u` before any
   timing claim.

### The fleet-wide condition I deliberately did NOT "fix"
`deactivated_head` is true here — but **17 of 19 sites pin `head` to a deactivated component**
(13 → "Document Head" `is_active=f`, 4 → "head-seo-standard" `is_active=f`; only
webdesign.co.uk pins an active one), and every one serves a correct head because the chrome
join honours the pin with no `is_active` filter (REB-006). The deactivation is the anomaly, not
the pin. **Do not repoint this one site** — that hides a fleet condition. Fix is fleet-scope:
reactivate, or migrate all 17.

### A hazard I talked myself into, then disproved by reading the code
I believed the sweep could point live CTAs at the 404 `platform-log` page, and was about to
archive a brief-mandated page to prevent it. `chooseCTATargets`
(`resolve_internal_links_action.go:319-349`) killed it: `ordered = rank(interactive) ++
rank(hubs)`, `primary=ordered[0]`, `secondary=ordered[1]`, and the site has **4 active tool
pages** ahead of its single `section-index`. `platform-log` sits at index 4 and can never be
selected. **Reading the code beat the theory twice in one session** — the other time, a
"missing `status='active'` filter" in `check_misdirected_cta.go` turned out to be present at
`:406`, and the finding was merely filed before 086b archived the stub.

## 7. Next actions, in the order I would take them

1. **Read the 3 `claims_unverified` rows and act on them.** 4 unregistered numbers on
   `capabilities` is the likeliest place the site currently overclaims, and the mission brief's
   non-negotiable rule is that every claim needs a real source. Also chase the stuck
   `needs_content_page`: "more than ten live production sites" wants checking against reality.
2. **File the `image-build-handler` logo-mapping bug via `090`** (§5.1), after checking whether
   another producer supplies `image_prompts.logo`.
3. **Decide `platform-log-index`**: approve the 07-20 `needs_page` (with a planner reconcile so
   REB-007 doesn't no-op it), or retire the page. Either way the guide's live 404 link stops.
   **Owner-facing** — the brief ranks the decision record FIRST among differentiators.
4. **The `resolve_internal_links` eligibility gap** (§5.7) — the 191-shaped fix. Cross-cutting:
   council gate + concept-register entry in the same commit, not a quiet patch.
5. **`#error-msg` on the selector tool** — 2 `improve_tool` rows, genuinely absent, now promoted.
6. **Re-run the sweep after the above** and check the fingerprint changed. A second run at an
   unchanged fingerprint files one `capability_gap` roadmap row instead of reporting clean
   (`bugs_open/171`'s fix — correct behaviour, not a failure).

## 8. Commands

```bash
# the sweep — read its blast-radius header EVERY time
./run_improvement_sweep_once.sh fundamentallyai.com

# did it actually run?  (orchestration_states has NO agent_type column)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
 "SELECT orchestration_id, current_step, status, created_at, error FROM orchestration_states
   WHERE correlation_id='<SWEEP_CORR>' ORDER BY created_at;"

# the 291 gate (audited vs skipped vs not-converging)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -At -U clients_user -d clients_db -c \
 "SELECT jsonb_pretty(collected_data->'audit_state') FROM orchestration_states
   WHERE correlation_id='<SWEEP_CORR>' AND collected_data ? 'audit_state' LIMIT 1;"

# what the sweep left behind
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
 "SELECT status, item_type, item_key, left(summary,60), left(COALESCE(error,''),110)
   FROM site_work_items WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'
    AND status IN ('failed','blocked','deferred','needs_human_review') ORDER BY status;"
```

Site id: `199733a8-ac9c-4c30-b2ce-65ecdac6f3bd`.
