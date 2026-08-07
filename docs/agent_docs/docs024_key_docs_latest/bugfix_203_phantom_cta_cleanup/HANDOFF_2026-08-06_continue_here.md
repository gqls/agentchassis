# HANDOFF — bugfix 203 phantom-CTA cleanup — continue here

**Written 2026-08-06 ~20:45Z.** Chassis live tag at handoff: **v1.0.1261** (both pods).
Read `NOTES_phantom_cta_cleanup.md` F1–F14 for the evidence behind every claim here; read
`PLAN_2026-08-06_phantom_cta_cleanup.md` D1–D6 for the decisions and why.

## Do not redo these — they are settled and cited

- **The source fix (`880a405a6`) is LIVE.** By ancestry: it is an ancestor of `1e349d046`
  (197's fix), pod-proven on v1.0.1259 against real traffic by that lane; the fleet has since
  rolled further. Do not spend a pod-grep re-proving the same binary fact.
- **Council verdict: APPROVED r1**, corr `42eda9a5`, 5 advisory objections, none high. Read
  it from `diagnosis_artifacts` (`kind='council_report'`, column is **`body`**, not
  `content`). Its asks are already discharged or re-aimed — see below.
- **The class audit is DONE** (F4/F5): two members left, both inert fleet-wide.
- **The detector is NOT under-running** (F11) — 123 items, 10 sites, filed to 08-05.

## The remaining work, in the order I would do it

### 1. The eight shipped instances (the only user-visible damage)

Four genuine label↔target mismatches:

| site | page | slot | button says |
|---|---|---|---|
| finetuning.uk | /guides/tool-ai-data-risk-checker-guide.html | hero | Run the Risk Checker |
| leopardessconsulting.co.uk | /who-we-help.html | hero | Score your process |
| robot-hands.com | /how-to-specify-a-gripper.html | hero | Run MatchMatrix |
| finetuning.uk | /about.html | content-block-about | How We Work |

Plus **4** `leopardessconsulting.co.uk` blog heroes (`/blog/ai-data-trust-in-financial-services`,
`-in-healthcare`, `-in-hiring-and-hr`, `/blog/can-you-trust-ai-with-your-data`) whose button
reads **"Get Started"** — the *fabricated label* default — pointing at `/contact.html`. The
destination is plausible; the button was never authored by anyone. Consider whether the right
outcome is removal rather than resolution.

**`/contact.html` exists on all 7 sites** — none of these is a broken link. Do not describe
them as such to the owner.

> **CORRECTED 2026-08-07 — the method below is RETRACTED. Do not follow it; it names an
> operation the framework does not expose.** `internal-link-resolver` is a **pure function**
> (workflow `resolve_links → complete`, returns sections to its caller, writes nothing), and
> fleet-wide exactly **one** agent calls it: `page-content-writer`. So link resolution happens
> at content-writing time and there is **no repair-time entry point** to "re-run
> resolve_internal_links for a page". See NOTES F18, and F19: `component_link_repair` is
> *dead*-link repair and correctly no-ops here, because `/contact.html` exists on all 7 sites.
> **The live routes and the owner decision they need are PLAN D7.** In short: a bare rerender
> is right for the four fabricated "Get Started" heroes and *wrong* for the three tool CTAs,
> because their correct targets exist (F17) and a rerender would delete the buttons.

**Method (D3 — RETRACTED, see the correction above):** re-run `resolve_internal_links` for the
page so the resolver writes a validated `cta_url` into `content_data` — or correctly leaves
it absent and raises `unresolved_cta` — **then** rerender. A bare rerender on the fixed
binary is sufficient to *remove* the phantom (proven: `dartsonline.com/news/index.html`
cleared itself at 20:07:14Z today by that route) but it will **delete** a button whose real
target the resolver could have found, which is a usefulness regression on four CTAs that
name real things. Do not hand-set URLs — owner ruling 2026-08-04, the framework writes the
content.

**Two traps on this path, both already paid for by other lanes:**
- a hand-made `page_rerender` work item needs `page_id` **in the spec AND in the column**,
  or it fails with `rerender_single_page: page_id not found in input` and retries;
- `save_page_sections` can **refuse** a save on the claims guard, so a green orchestration
  status does not mean the page was written — check
  `agent_error_log` for `CONTENT_CLAIMS_FLOOR_DETAIL`.

**Verify at the served page, not the stored row.** The census reads
`page_components.rendered_html`; the phantom check landmine is that stored ≠ served. Fetch
the live URL per row, before and after.

### 2. The code half — only with a measurement first (D4)

Do **not** simply delete `component_library.go:1138/1140`. `renderGoStyleSubstitutions`
returns the literal `{{.field}}` for an absent key, so that trade ships template syntax
inside an `href` — already visible on `idea.uk/tools/ab-test-calculator/index.html`, which
stores a literal `{{.section_heading}}` (1 row fleet-wide).

> **UPDATE 2026-08-07 — the gate measurement is TAKEN (NOTES F15), and it says LOW PRIORITY.**
> Same two pods, now 5h24m old: `"using regex fallback"` = **0** on both replicas, with both
> controls checked first (11 components rendered inside the window, so the path was exercised;
> 34 `warn` lines present, so the level is not swallowed). With yesterday's durable bound
> (1 literal-`{{.` row in 1,247 stored components) the fallback is **rare-to-never**.
> 11 renders is a small denominator, so this is a bound, not a zero — but neither class member
> can fire on a path that does not run, so there is even less reward for F6's risk. **The
> candidates below stay correct and stay unurgent.**

The owed measurement, and it is the gate on the whole item: **how often does
`executeGoTemplate` actually error?** The log route is a trap — I burned a check on it (F8):
the pods were 25 minutes old, so `--since=24h` measured 25 minutes. Either wait for pods with
real uptime and grep `"using regex fallback"` on **every** replica, or find a durable
population signature. Then pick:
- **(a)** regex path substitutes empty for an absent key, letting the existing
  `missingBareFields` ERROR ("URL attribute rendered empty — dead control") do the telling,
  then remove the two defaults; or
- **(b)** delete the regex fallback and fail the render loudly. `RenderTemplateWithValidation`
  is **dead code** (no non-test callers), so `contextToMap` is reachable only from the
  fallback at line 989 — which makes (b) real, and it is the option that makes the bad state
  unrepresentable.

Either is a guarantee change on shared rendering plumbing → council round, with the blast
radius measured before submitting, not handed to the reviewer (2026-07-28 ruling).

### 3. Not this lane's, but found here — file or route, don't annex (D5)

- **123 `cta_names_unknown_destination` + 26 `unresolved_cta` at `needs_human_review`, every
  one with `handler_agent = ''`.** Nothing will ever drain them. That is `bugs_open/083`'s
  class arriving on this queue.
- **`check_misdirected_cta` can only flag an anchor when `bestPageMatch` already names a
  better page** (`check_misdirected_cta.go:164`, early return). A CTA promising something no
  indexed page matches is invisible to it by construction — which is why it missed all four
  mismatches above. Deserves its own bug; `bugs_open/185` (detectors select deployed) would
  bite it independently.

## Two live corrections to the record, so nobody re-inherits them

- `bugs_open/203`'s **census SQL cannot execute** (joins `page_components.site_id`, which
  does not exist). Its "13 rows" therefore has no provenance, and the 13→4 gap is partly
  unrecoverable. Corrected query in the RUNBOOK. Logged in `WRONG_CALLS.md`.
- `bugs_open/203`'s **candidate 3 premise is refuted** (F11) — corrected in the bug file's
  state section, not silently.
- My own: `Council-Reviewed:` on the docs-only commit `eba83792e` is a bad join key.
  Forward-only, logged in `WRONG_CALLS.md`. **A docs commit needs no trailer.**
