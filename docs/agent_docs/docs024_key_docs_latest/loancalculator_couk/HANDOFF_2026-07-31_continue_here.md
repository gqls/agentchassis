# HANDOFF — loancalculator.co.uk · continue here

**Written 2026-07-31 (end of the opening session). This is the cold-start doc for
this lane — read it before anything else, then `NOTES` tail and `PLAN` §REVISED.**

> ## ⚠ CORRECTIONS from the second session (2026-07-31, later) — read these first
>
> Everything in §2 and §3 below was **re-verified and holds** (27/27 serve 200; 27
> pages `owned`/`deployed`; stored bytes byte-exact — md5 both sides; M1/M2 live in
> the pod, checked with a positive *and* a negative control). Nothing was written to
> the site or its DB rows this session. Five things below are now **wrong or
> superseded**:
>
> 1. **"12 inline-JS calculators" (§1) — the count is right and the membership is
>    wrong.** `tools/credit-roadmap.html` is static prose in the tools folder (zero
>    controls, no inline script) and **`index.html` has a working hero calculator**
>    that no earlier doc counted. So: **12 interactive pages = 11 under `/tools/` +
>    the homepage.** Any acceptance gate over "the 12 tool pages" is unpassable.
> 2. **§4.1 (file source via git-adapter) is OWNER-BLOCKED and off the critical
>    path.** git-adapter is **write-only**; the read machinery is
>    `diagnose_read_repo_files_action.go`; and the read token **cannot see
>    `gqls/sites`** — 404 *while authenticated*, because it is a fine-grained PAT
>    scoped to selected repos. Needs GitHub admin. **But step 2 never needed it**: the
>    faithful bytes are already in `page_components.rendered_html`. Decompose from
>    there; the file source's value is *the next* site.
> 3. **§4's blocker is understated.** `site_components` is **empty (0 rows)** and
>    `assemblePage` reads chrome from it, so a `generic` flip ships every page with
>    **no head, no nav and no footer** — not just nested `<html>`. Creating the
>    site-level chrome is an output of step 2.
> 4. **§4.2 must preserve page-local `<style>`, not only inline `<script>`.** 8 pages
>    carry one, **7 of them calculators**; on `credit-health-check` two of those rules
>    are what show one wizard step at a time. Drop them and the tool computes
>    perfectly and displays wrongly. Also: the homepage tool needs an **external**
>    `<script src>` dep, so inline scripts alone are insufficient.
> 5. **§6/§5's "a roll kills an in-flight council" did not hold here.** The `revise`
>    verdict landed at 2026-07-30 19:43Z, ~2h *after* the v1.0.1211 roll.
>
> **New in the lane:** `decompose_prover.py` (the splitting rule, proved over all 27
> pages offline — 25% of interactive-page text frozen, all proofs pass) and
> `acceptance/` (the before-photograph: **12 RESPONDS + 1 NO-CONTROL**, harness pinned
> by sha256 — re-pin the same harness or the comparison measures the harness).
> **Read the `NOTES` 2026-07-31 entries before the build order below.**
>
> **An owner question is open** (see `README_where_we_are` tail): full decomposition,
> or the cheaper freeze-the-calculators split the `loanandmortgagecalculator` lane
> chose. Do not start the DB-changing steps until it is answered.

---

## 1. What this lane is, in one paragraph

loancalculator.co.uk is a hand-built 27-page UK loan site (12 inline-JS
calculators, 13 guides) that lived outside the platform until 2026-07-30. The owner
asked for two things: adopt it so the platform manages it, **and mend whatever in
the framework's adoption path would have broken it** — this site is the proving
case, and the framework mends are the real deliverable. A later owner ruling
changed the target posture: the site must be **completely editable and evolve like
every other site**, starting faithfully with working tools.

---

## 2. Current state — verified, with the commands to re-verify

**The site is LIVE and correct.** Re-check before trusting this:

```bash
# 27/27 should be 200
for u in $(grep -oE '<loc>[^<]+</loc>' ~/projects/sites/loancalculator.co.uk/sitemap.xml | sed 's/<[^>]*>//g'); do
  printf '%s %s\n' "$(curl -s -o /dev/null -w '%{http_code}' --max-time 15 "$u")" "$u"; done | grep -v '^200' || echo "all 200"
```

**The DB is adopted, safe, and byte-exact.** `site_id = 0162cde4-633e-45e9-8ca6-87a6b2fe1d26`.

| fact | value |
|---|---|
| pages | 27, all `rebuild_policy='owned'`, `build_status='deployed'` |
| URLs | flat `.html` preserved exactly (`/tools/standard-calc.html`) |
| components | 27 × `ported-page`, `approved`, `content_data.deploy_mode='verbatim'` |
| stored bytes | **27/27 byte-exact** vs the served files, each recorded `sha256` verified |
| LLM recreation items | **0** — no calculator was ever handed to an LLM |
| open work items | **0** (24 cancelled with the reason on the row, 3 completed) |

**⚠ The `owned` + `verbatim` posture is a HOLDING state, not the goal.** It is safe
(a rerender re-ships identical bytes — proven) but it makes the site **unable to
evolve**, which is the opposite of what the owner now wants. See §4.

---

## 3. What is BUILT and LIVE (do not rebuild these)

All verified on **v1.0.1213**, both replicas, by grepping strings the changes added
plus a positive and a negative control.

| # | What | Where | Live? |
|---|---|---|---|
| M1 | `fidelity=locked` → verbatim adoption (the dial's first consumer ever) | `platform/orchestration/actions/adopt_verbatim.go` + `apply_adoption_plan_action.go` §3a | ✅ |
| M2 | `deploy_mode='verbatim'` ships stored bytes, skipping assembly | `rerender_single_page_action.go` (`loadVerbatimPageHTML`) | ✅ |
| 274 | `fidelity` forwarded across the adoption spawn boundary | `docs/agent_docs/sql_for_agents/274_*.sql` (APPLIED) | ✅ |
| — | `page_id` resolved from `(site_id, page_name)` | `rerender_single_page_action.go` (`resolvePageIDByName`) | ✅ |
| — | Interactivity guard widened so calculators survive a rebuild | `save_page_sections_action.go` (`sectionHTMLIsInteractive` + `interactiveHTMLSQL`) | ✅ |

Registered as **ADO-037**; **ADO-011** updated (its `locked` rung is no longer
inert). Four LANDMINES entries added and synced to `doc_notes`.

```bash
# re-verify the code is in the running pod (a roll is NOT evidence)
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec $POD -- sh -c 'strings /app/agent-chassis | grep -c "adoption-locked/1"; strings /app/agent-chassis | grep -c "page_id absent, resolved from"'
```

---

## 4. WHAT REMAINS — the build, in order

Owner's ruling (2026-07-30 late), verbatim: *"adopt from our own files. I want the
site to be completely editable and evolve and improve like the other sites will,
just as long as it starts similarly enough with working tools."*
Plus: *"perhaps both the tool-page marker and the guard is widened"* (guard half is
DONE), and *"I am not so worried about url preservation if it proves to mean
potential bugs ahead."*

**The blocker that makes this a build and not a config flip:** a whole stored
document cannot go through `assemblePage` — it concatenates chrome + sections, so a
complete `<!DOCTYPE html>…</html>` blob yields nested `<html>`. **Do NOT flip
`rebuild_policy` to `generic` before decomposition exists**; that turns a working
site into malformed pages on the next rerender.

Target mode: **`fidelity=high` = seed faithfully, then evolve.** doc 028 already
reserves the word; do not invent a new one. `locked` stays as-is for true archives.

1. **File source.** Read our own bytes from the deploy repo (`gqls/sites`, dir
   `loancalculator.co.uk/`) via **git-adapter**, so it is platform-side and reusable
   for any site we publish — not a local CLI. Verified: origin bytes == repo bytes
   for all 27 pages. **Never firecrawl `rawHtml` for this** (see §6, G10).
2. **Decompose each page into real sections.** Prose → ordinary content sections the
   writer/improvement loops may rewrite. **Each calculator → a
   `component_level='tool'` component with its inline JS preserved BYTE-FOR-BYTE**,
   and give it a `tool-page` class marker (the owner asked for the marker as well as
   the widened guard; the guard is already in).
3. **`rebuild_policy='generic'`** + a **timed** adoption lock (`locked_by='adoption'`,
   expiring — see `FOCUS_adoption_faithfulness_via_locks(2).md`), NOT permanent
   ownership. Drop `deploy_mode='verbatim'` from these components as part of the
   same change, since assembly must now run.
4. **Queue `needs_domain_research`** so classifier → strategist → briefing → planner
   → composition → design all run. This is what makes the site "like the others" —
   **and it discharges the council's two high objections** (§5).
5. **URLs: stop defending the flat paths.** Per the owner, let the planner use its
   canonical `/tools/<slug>/index.html` and emit **meta-refresh redirect stubs** from
   the 27 old flat paths (the `redirects` table still has no consumer — G5). This
   removes the adoption-vs-planner URL conflict rather than fighting it.
6. **Acceptance bar (the owner's own words):** *"starts similarly enough with working
   tools."* After the first full build: every URL resolves (directly or via stub) and
   **every calculator still computes in a real browser**. Regenerated prose is
   EXPECTED and fine. A dead calculator is not.

---

## 5. Council state — REVISE, and what is still owed

Trail correlation **`f9eae63e-05fb-40c8-b60c-1670c5681cbe`** (round 1 verdict
`revise`, `decided_by: gating objection from bug_historian`; 4 approve, 3 abstain).
Round 1 of the same submission was **killed by another session's roll** mid-step —
resubmit with `RESUBMIT_CORR` so the trail accumulates.

**Already answered in code** (commit `5a840ff3d`): name collisions and
missing-`rawHtml` now **fail loud** instead of silently dropping a page;
`>1 ported-page component` now **refuses** rather than logging a warning; the
classifier skip now writes a `doc_notes` row **in the same transaction** as the
pages.

**Already answered by measurement** (attach these, do not write code):
- `pages_site_id_name_key` is a real non-partial UNIQUE `(site_id, name)` → the
  `ON CONFLICT` is valid.
- Exactly **two** live agents dispatch `rerender_single_page` (`page-rerender`,
  `report-builder`) → the M2 blast radius is shown, not asserted.

**Still owed, and genuinely FOR THE OWNER — do not resubmit-with-more-evidence:**
- **`adoption_guardian` + `mission` (both high): the classifier bypass.** doc 028
  says the classifier "always runs in full" and that inputs *weight* it, never
  shortcut it. **Step 4 above removes this objection entirely** — once the cascade
  runs, nothing is bypassed. Note that in the resubmission.
- **`architecture`: `ARCHITECTURE_SIGNAL: needs_rfc`** on the shared
  `content_data.deploy_mode` key. By the owner's own 2026-07-29 test (an RFC is
  needed when a shared mechanism's GUARANTEE changes) this is arguable, because the
  render path's guarantee moves from "every page is assembled + link-repaired +
  injected" to "unless marked verbatim". **A scope objection needs a human, not a
  better measurement.**
- `prior_art_librarian`: quote the ADO-011/ADO-016 register text and the CLAUDE.md
  owner ruling **verbatim** rather than paraphrasing, and grep for prior
  `verbatim`/`deploy_mode`/`owned page` mechanisms.
- `render_guardian`: the sha256 fidelity gate is **a manual step, not code** — the
  seat was right. **Build it into the pipeline**; it is the only reason this lane
  caught G10, and it belongs in whichever source step wins.

**Process rule learned the hard way: SUBMIT FIRST, THEN COMMIT.** `e6a8bb63b` carries
`Council-Submitted: pending` — a placeholder, because I had no correlation yet.
Forward-only forbids the amend, so `098` will never credit that commit.

---

## 6. Lane-specific traps (all four now in `LANDMINES.md`, synced)

- **G10 — firecrawl `rawHtml` is the POST-JS DOM, not served bytes.** Measured:
  27/27 pages differed, every one larger by ~8,900–9,060 bytes. **A near-constant
  size delta across every page is the tell** — a script ran. Here `nav.js` had
  executed (~9KB of nav baked in), relative URLs were rewritten absolute, and
  `href="#"` became a self-link that **reloads the page** on desktop click.
  Three pages reached production this way before the checksum gate caught it; they
  were restored.
- **`b2 sync --skip-newer` silently skips a revert** whose bucket copy is newer —
  the deploy stays GREEN and the file stays wrong. Count `upload` lines against
  files changed; verify at the ORIGIN with a cache-buster (`?cb=$RANDOM`), because
  it looks exactly like a stale CDN. Fix: `gh run rerun <id>` (fresh checkout ⇒
  fresh mtimes). Re-committing does nothing when git is already correct.
- **`pages.url = '/'` deploys a file with NO NAME** — `getPageInfo` derives the
  filename as `TrimPrefix(url,"/")`. Normalise directory URLs to `index.html`.
- **The adoption crawl index keys ONE page under 2–3 aliases** — dedupe by content
  **pointer**, never by URL string, and sort the keys (map order is random, so an
  unsorted pick writes a different `pages.url` on a re-run).
- **`snapshot_agent(text,text)` writes to `agent_definitions_backup`**; the one-arg
  form writes into `agent_definitions`. Check the wrong table and a good snapshot
  looks like a no-op. Assert the snapshot holds the **pre**-change value.
- **`jsonb_set(..., create_if_missing := false)` on a wrong path is a silent
  no-op.** Assert `UPDATE 1` and re-read the value.

---

## 7. Decisions already made — do not re-litigate

- Adopt **through the real pipeline**, not a second bespoke porter (`webdesignport`
  is the row-shape reference only; it stores a **fragment**, not a document).
- Repair the site **before** adopting — the pipeline learns from what is served, so
  defects would be frozen in. Phase A did this (commit `b4302e22b`): 4 chrome-less
  fragment pages wrapped, nav added to 10 pages that had none, all asset refs made
  root-absolute, 3 dead links fixed, `nav.js`'s phantom link fixed + a stub left at
  the old path, sitemap regenerated (27 real URLs), 3 dead files removed.
- Serving was restored by the **owner** binding a Cloudflare worker route
  (`loancalculator.co.uk/*` → the portfolio-sites worker). The outage signature was
  a **hang, not an error**. `curl /worker-health` is the probe.
- `--fidelity locked` **is** documented as active in `082_submit_domain_unified.sh`
  now; the rest of the dial is still inert. Keep that NOTE honest.

---

## 8. Cold-start commands

```bash
# the standing five for this lane
ls docs/agent_docs/docs024_key_docs_latest/loancalculator_couk/

# what adoption created (and that nothing is queued)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT url, rebuild_policy, build_status FROM pages
WHERE site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26' ORDER BY url;"

# re-run the fidelity gate (stored bytes vs served files) — the check that saved this lane
# see NOTES 2026-07-30 'late' entry for the loop; it must report 27/27 byte-exact
```

**Do NOT roll the chassis while a council run is in flight** — a roll kills it and
the review is lost. That is how round 1 died.
