# 387 — three pages report `build_status='deployed'` with a timestamp from two hours ago, and 404 on the live site

**Filed** 2026-08-24 by the `bugs_open/364` lane. Found while curling a page before calling a
claims finding "live damage" — the check found the opposite of what was expected, which is why
it is filed rather than assumed.
**Severity** medium-high. Three pages that the platform believes it published are not reachable.
**Class** status-vs-artefact divergence (CLAUDE.md: *"Trust the rendered artefact, not the status"*).
**Status** OPEN, unowned. First-hand evidence below; root cause NOT established — see §4.

## 1. The evidence, both halves

**The DB says deployed**, read 2026-08-24 ~20:1xZ (`site_id=2a8ebf9c-20a2-4c39-b191-840b012371da`):

| page | status | build_status | deployed_at | components |
|---|---|---|---|---|
| adoption-tracker | active | **deployed** | 2026-08-24 16:27:57Z | 3 |
| protocol-tracker | active | **deployed** | 2026-08-24 16:27:49Z | 3 |
| model-directory | active | **deployed** | 2026-08-24 18:43:19Z | 3 |

**The live site says 404**, curled 2026-08-24 ~20:15Z:

```
https://ai-agent-orchestration.com/                            200
https://ai-agent-orchestration.com/adoption-tracker            404
https://ai-agent-orchestration.com/protocol-tracker            404
https://ai-agent-orchestration.com/model-directory             404
https://ai-agent-orchestration.com/definitely-not-a-real-page-xyz 404   <-- CONTROL
```

**The control is load-bearing and is why this is a finding rather than a guess.** A parked or
catch-all domain returns 200 for every path, which would make a 200 meaningless; here an invented
URL returns 404, so the domain discriminates, and the three 404s are real absences. The root
returning 200 proves the site itself is serving. (Both traps are recorded in memory: *"a parked
domain 200s EVERY path"*, and *"curl the target before calling a queue row live damage"*.)

`model-directory` was stamped deployed at **18:43Z and 404s at 20:15Z** — ~1h32m later, so this is
not a propagation window.

## 2. Why it matters

- `deployed_at` is what every downstream reader trusts to mean "this is public". A page that is
  `deployed` and absent is invisible to exactly the sweeps that would otherwise notice it.
- These three pages are **linked from the site's own navigation and from `bugs_open/364`'s census
  as live content** — they carry stored `rendered_html`, so every DB-side check reads them as fine.
- It silently wastes the whole build: three pages' worth of writer, designer and deploy work,
  repeatedly (`agent_error_log` shows these same three pages refused **40 build attempts** between
  them over 60 days for an unrelated claims defect — `bugs_open/364`).

## 3. A second defect, visible only because of the first

`model-directory`'s stored `hero` contains an **unrendered placeholder**:

> "Every model behind our agents, catalogued in one place. This registry tracks **NNN+** AI agents
> across **NNN+** agent types, organised under the eight departments we build for…"

`NNN+` is a template token that was never substituted. **It is not public today only because the
page 404s** — which is the sole reason this is filed here as an observation rather than as a
live-content incident. Fix the deploy and this ships to the public the same hour. Note also that
`checkPlaceholderPatterns` (`validate_page_content.go`) did not convict it: `NNN` is not in
`placeholderPatterns`, whose entries are bracket forms like `[name`, `[company`
(see `bugs_open/218` for that scan's other failure mode).

## 4. What is NOT established — read this before fixing

**The root cause is unknown and three plausible causes fit the same evidence.** Do not pick one
from this file:

1. the deploy wrote to the repo but the static build/routing never picked the pages up (a slug or
   route-map gap — `bugs_closed/015`'s shape, a mistyped `page_type` orphaning a page);
2. the deploy step reported success without publishing (the `complete`-is-not-proof class);
3. the pages are published under a different path than their `name` implies.

The cheap discriminator nobody has run yet: look in the deploy repo for the three files, and read
the deploy action's own record for the 16:27Z / 18:43Z runs. **`build_status` is the instrument
under suspicion here, so do not verify with it.**

## 5. Relations

- `bugs_open/364` (found it; the three pages are that bug's motivating page types).
- `bugs_open/328` (a page that failed to build is still linked from the pages that did),
  `bugs_open/266` (four producers rebuild/redeploy without reading `page_status`) — adjacent
  status-vs-reality defects, neither the same as this.
- `bugs_open/218` (the placeholder scan's coverage), `bugs_closed/015` (page_type orphaning).

---

> **CORRECTED 2026-08-25 (session `bugs_open/387`) — the HEADLINE (§1–§2) is REFUTED; §3 is the
> live incident and is now the whole bug.** What caught it: the control this file lacked — a
> known-good page at the same URL form.
>
> **The three pages serve, and always did.** `pages.url` is `/adoption-tracker.html`,
> `/protocol-tracker.html`, `/model-directory.html`; all 200 (2026-08-25 ~10:28Z). The §1 probes
> used the extensionless form, which 404s for EVERY page on this hosting — `/about`, `/pricing`,
> `/contact`, `/services` all 404 the same way (`scripts/cloudflare/worker.js:40-44` declines the
> slashless form deliberately). Nav, sitemap.xml, canonical and og:url all carry `.html`. §1's
> invented-URL control proved the domain discriminates but shared the URL form with the claim, so
> it could not test the form. Third occurrence of this exact wrong call
> (`WRONG_CALLS.md` 2026-07-27, `LANDMINES.md` "A page's served URL is NOT derivable from
> `pages.name`" 2026-08-09, now here — see `WRONG_CALLS.md` 2026-08-25, filer's own entry); the
> probe is being automated as `scripts/probe-page-url.sh`.
>
> **§4's discriminator, run fleet-wide as a demand control (2026-08-25 ~10:40Z): 0 of 709
> `deployed` active pages 404 at `pages.url`** (698×200; 7×302 = webdesign.uk's deliberate
> off-domain redirect; 1×301; 3 transient TLS `000` that 200 on re-probe). The linked case of a
> genuinely absent deployed page is covered by `dead_internal_link_live`
> (`check_site_structural_validity.go`), and the "non-200 is a skip here / availability is
> `site_unreachable`'s" seam is a recorded design decision (register DGH-015). **Decision: no new
> per-page reachability check** — zero demand, existing coverage, documented seam.
>
> **§3 is real, PUBLIC (not held back by any 404), and root-caused.** Live at
> `model-directory.html` (curled 2026-08-25, regenerated 06:30Z the same day): "…against the NNN+
> agent types already running in production". Mechanism, first-hand from the executed prompt
> (`llm_call_log` id `9ba94176…`; full evidence + queries in
> `docs/agent_docs/docs024_key_docs_latest/bugfix_387_deployed_and_404/{NOTES,RUNBOOK}_387.md` —
> stated verification substituted for a 090 run per the 2026-07-31 ruling, the deciding evidence
> being the executed prompt itself):
> migration `557` (2026-08-22) wrote the exemplar 'Phrase it as "NNN+ AI agents"' into this site's
> `evidence_base.writer_block` and its guard requires that literal; the unscoped writer prompt
> contains ONLY the writer_block, never the facts values (the hero call carried **zero**
> occurrences of the fact's value); measured since 08-22: **137** instructed writer calls,
> **14 copied `NNN` verbatim, 0 wrote the agents value**. No detector has the shape
> (`placeholderPatterns` = substrings, `templateVarRegex` wants `{{`, and `NNN` has no digit for
> the claims scan). Fleet census of the candidate detector regex
> (`\mN{2,}\+|\mNNN\M|\mX{2,}\+|\mN,NNN\M`) over all active pages' `rendered_html` +
> `site_components`, 2026-08-25: **exactly 1 hit (this hero), 0 false positives** — bare `XX`
> (`siglo XX`) and `[number]` (a quoted fill-in template) are deliberately NOT in the regex.
>
> **Fix (owner-approved plan, `bugfix_387_deployed_and_404/PLAN_2026-08-25_387.md`):** interim
> successor migration for the writer_block (no stand-in tokens; live immediately; aiao lane told —
> their CONTRIB names the wording as theirs to change) → framework rebuild of the page → numeric
> stand-in blocker in `checkPlaceholderPatterns` (council-gated, inert until the next roll —
> this bug stays OPEN until that is live) → durable close: `composeWriterBlock` carries verbatim
> guidance so unmanaged sites (13 today) can adopt `{value}` substitution — proposed to the
> `bugs_open/288` lane, which owns that file.

> **KNOWN GAP, pinned 2026-08-25 (council round 3, bug_historian):** the numeric stand-in
> detector guards the page-BUILD validation path only — `checkPlaceholderPatterns` has exactly
> one caller (`ValidatePageContentAction`), and the recorded landmine stands: *"validate_content
> protects the page-BUILD path and NOTHING ELSE"* — the **section-editor and chrome-rerender
> paths render without it**, so a stand-in entering through those doors ships undetected (the
> `bugs_open/093` one-guarded-call-site class). The after-the-fact cover for every write path is
> the stored-content census (`bugfix_387_deployed_and_404/RUNBOOK_387.md`), re-runnable in one
> query. Guarding those paths at write time is a separate piece of work on seams other lanes own
> (260/093 lineage), deliberately NOT smuggled into this fix.

---

> **CLOSED 2026-08-25 ~19:3xZ — fixed AND live, re-scoped (session `bugs_open/387`).**
> The bug that closed is the re-scoped one (the public `NNN+` stand-in; the 404 headline was
> refuted the same day, see the CORRECTED block):
> - **Source fix live + proven:** migration 611 applied 11:20Z; all three pages regenerated
>   through it (12:41Z, 18:31Z, 18:36Z) with honest floors; fleet census after: **0** stand-ins
>   in any page_component (rendered_html OR content_data, no page filter), **0** in chrome,
>   **0** on the served page.
> - **Detector live:** `checkPlaceholderPatterns` numeric stand-in blocker (council APPROVED r4,
>   trail `6cfaa8f0`) — verified on `v1.0.1339` (rolled 19:07Z) at BOTH replicas by binary probe:
>   `grep -ac "numeric stand-in placeholder" /proc/1/exe` = 1 (was 0 pre-roll, same pods), positive
>   control 2, absent control 0. ⚠ probed at the binary, not `service_binary_capabilities`, which
>   was measured STALE for this roll by the 364 lane.
> - **Carry live:** `writer_block_guidance` (CLM-029, APPROVED r2, trail `0de22385`) — same probe
>   pair, both replicas, 1/1. Adoption (managed mode on aiao) remains the site lane's call and
>   retires 611's interim wording.
> - **One dated check outstanding, deliberately surviving the close:** 2026-08-26 ~09:06Z, confirm
>   the 611 block survives the daily evidence-refresher pass (query + expected result in
>   `docs/agent_docs/docs024_key_docs_latest/bugfix_387_deployed_and_404/NOTES_387.md`). Backstop
>   either way: the now-live detector refuses a resurrected stand-in instead of publishing it.
> - **Post-roll disposition rule for the tracker pages** (their hero/CTA are newly scanned since
>   `v1.0.1339` — Phase 2 of the now-closed `bugs_closed/364`): an `unregistered_number` refusal
>   is the 364 class (tell that lane first); a `placeholder_text` with a stand-in Value is this
>   detector working. NOTES carries the full rule.
