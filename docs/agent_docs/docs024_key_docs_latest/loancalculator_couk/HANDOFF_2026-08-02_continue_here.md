# HANDOFF — loancalculator.co.uk · continue here (2026-08-02)

**This supersedes `HANDOFF_2026-07-31_continue_here.md`, whose premise ("the
decomposition build is outstanding") is no longer true. That file is still worth
reading for how the lane got here; this one is the current state.**

Read order for a cold start: this file → `NOTES` tail (the 2026-08-02 sections) →
`RUNBOOK` §§ "Chrome for an assembled site", "Proving the decomposition BEFORE
writing a row", "Shipping the decomposition", "After a chassis roll".

---

## 1. State in one block

```
site            loancalculator.co.uk   0162cde4-633e-45e9-8ca6-87a6b2fe1d26
pages           27, ALL rebuild_policy='generic'   (flipped 2026-08-02)
component rows  63 total = 51 ported-prose + 12 tool, ZERO verbatim
locks           12 tool rows lock_type='permanent'
                   locked_by='decompose_20260802_proven_calculators'
                51 prose rows unlocked, deliberately
chrome          site_components head/header/footer authored in chrome/, installed
live            27/27 HTTP 200, 27/27 byte-identical to the mirror's prediction
calculators     12/12 match GOLDEN_2026-08-02_decomposed.json
backup          page_components_bak_20260802_decomp — all 27 original rows
                content_components_bak_20260802_decomp — the pre-guard template
                site_components_bak_20260802_decomp   — the 08-01 chrome
```

**The site is open to the improvement loop for text and closed for arithmetic.**

Restore, per page, one command:
```bash
DECOMP_WORK=<work> python3 decompose/load_decomposition.py --restore <page-name>
```

## 2. Do this first

**a. ~~Finish the post-roll check.~~ DONE — the new chassis renders identically.**
`tool-standard-calc` (6 rows, tool at position 4) was re-rendered by the
`21:39:20Z` image at `21:53:41Z` and came back **EXACT** against a prediction
written before the roll, and its calculator still MATCHES the golden. So the
assembly path is now proven stable across **two** chassis images, empirically
rather than by reading the diff. Original instructions kept below in case a later
roll needs the same check.

**a-original.** A fresh chassis was deployed at
`2026-08-02T21:39:20Z`. All 27 served pages were re-verified after it and are
still byte-exact — but nothing had re-rendered, so that proves the site is
unchanged, **not** that the new binary renders the same. A work item is in flight
to prove the stronger thing:

```sql
SELECT status FROM site_work_items WHERE created_by='decompose-postroll';
-- a23d405f-4e6d-48d8-b970-6296834fd512, tool-standard-calc
```
When it completes:
```bash
export DECOMP_WORK=/tmp/decomp-work
python3 decompose/verify_shipped.py tool-standard-calc     # expect EXACT
```
The renderer source is unchanged by this roll (`git log --since=2026-08-02T18:39:14Z`
over `rerender_single_page_action.go`, `rerender_link_repair.go`,
`tool_doc_header.go` is empty), so EXACT is expected. If it is NOT exact, **stop
and read the diff** — that is the mirror and the platform disagreeing, and 26 other
pages are one re-render away from the same thing.

**b. If the scratchpad is gone, rebuild it.** It does not survive a session.
```bash
cd docs/agent_docs/docs024_key_docs_latest/loancalculator_couk/decompose
./prepare_work.sh /tmp/decomp-work        # ~2 min; prints its source
export DECOMP_WORK=/tmp/decomp-work
python3 load_decomposition.py --check --all   # rewrites predicted/, writes nothing
```
It reads from `page_components_bak_20260802_decomp` now — the live verbatim rows
are gone. It prints which source it used; expect `0 live … 27 from …_bak_…`.

## 3. The queue of real work, in the order I would take it

**(1) Four calculator defects, found during the rewrite and deliberately not
fixed.** Each needs an explicit unlock, the fix in `content_components`, a
re-render, and a golden refresh:

| tool | defect |
|---|---|
| `tool-overpayment-impact` | money rendered to three decimals — `£448.024` |
| `tool-car-finance-pcp-hp` | computes nothing at 0% APR |
| `tool-consolidation-risk` | a rate-less debt counts toward balance but not interest |
| `tool-loan-vs-savings` | the verdict is distinguished by colour alone |

The unlock is the deliberate act the lock exists to force:
```sql
UPDATE page_components pc SET locked_at=NULL, locked_by=NULL, lock_type=NULL
FROM pages p, content_components cc
WHERE pc.page_id=p.id AND cc.id=pc.component_id
  AND p.site_id='0162cde4-…' AND cc.function='<the tool>';
```
Re-lock afterwards, and re-capture the golden — otherwise the next comparison
reports your own fix as a regression.

**(2) The header's link list is hand-maintained.** 25 links, lifted verbatim from
`nav.js`. A page added to `pages` does not appear in it; a page removed leaves a
dead link. Generating `site_components.header` from `pages` is the obvious next
mechanism and was deliberately not bundled with the decomposition — a nav that
changes at the same moment as the page structure gives a regression two candidate
causes.

**(3) A tidy owed on `tool-loan-repayment`.** Its template carries an HTML comment
that ships in the public source of `/index.html` and `/tools/standard-calc.html`,
and one clause of it editorialises. Not done at the time because `index` was
mid-render; ride it with the first defect fix above.

**(4) `/tools/standard-calc.html` is an orphan** — nothing links to it, and it
duplicates the index calculator. A content question for the owner, not a bug.

## 4. What will mislead you

Every one of these cost real time today. The first four are also in the fleet
`LANDMINES.md`.

- **`site_components` having rows does not mean the chrome works.** On a verbatim
  site assembly is never reached, so broken chrome sits through a full-site
  rerender reporting every page a success. The 08-01 chrome linked a 404
  stylesheet (`styles.css`, plural), had a nav with zero links, and two 404 images.
  Ask whether the chrome RESOLVES, not whether rows exist — `load_chrome.py`
  refuses on exactly that.
- **Cloudflare answers `Python-urllib` with 403.** `curl` 200, `urllib` 403,
  `urllib` + browser UA 200. A Python health check reports a healthy site as
  broken, and the output is indistinguishable from the 404s it was written to find.
- **The verbatim flip is the ROW COUNT, not a flag.** Adding a row beside a
  verbatim row switches the page to assembly and leaves a complete `<!DOCTYPE>`
  document as one of its sections.
- **An extraction rule that reads only INLINE scripts is blind** to a page whose
  logic is in a `<script src>`. `/index.html` keeps `calculateLoan()` in
  `global.js`; its whole results box decomposed as prose and the prover's own P4
  passed, because P4 only asks about inline scripts.
- **A `complete` work item is not a propagated deploy.** The item flips when the
  render and git commit succeed; the wire is a sites-repo Action, a `b2 sync` and a
  cache purge behind. Seen once: 1 of 18 "did not match", re-run gave 19 of 19
  EXACT. **When one page in a batch mismatches and the rest pass, re-run first.**
- **A work item filed `status='detected'` is NEVER dispatched.** The selector takes
  `('triaged','approved')`; 31 discovery-filed items have sat since 2026-07-14.
- **`priority` is nearly dead**: `ORDER BY created_at ASC, priority ASC, id ASC`,
  so a new item goes behind every older one fleet-wide. Do NOT convert an observed
  completion age into an ETA — I read "items completing now were created 19 hours
  ago", projected next-day, and it took three hours. Those rows are the oldest by
  construction; their age is the tail, not your wait.
- **`--compare`'s pass/fail is the wrong question when a change alters the
  fingerprint's SHAPE.** 41 diverging keys across 12 live calculators, and zero
  changed values — every one was a control that gained an `id` (appearing named,
  vanishing positional). Classify APPEARED / VANISHED / VALUE CHANGED before
  believing a count.
- **A `LIKE '%…%'` probe over `rendered_html` cannot tell markup from a comment.**
  It told me the base-rate claim was still on the homepage; the only occurrence was
  inside a comment. Strip `<script>`, `<style>` and `<!-- -->` and read the visible
  text.

## 5. Why the claims in §1 are claims and not hopes

- **"byte-identical to the mirror's prediction"** — `assemble_mirror.py` is a
  Python reimplementation of the Go `assemblePage`, so it agrees with its own
  harness by construction. Its test is that it wrote each page's predicted bytes
  BEFORE any row existed, and the Go path then served them with the same md5. 27
  of 27. Go's `json.Marshal` HTML-escaping and its recursive map-key sort were the
  two most likely divergences and did not diverge.
- **"12/12 calculators match"** — against a golden captured FROM the decomposed
  site, so alone it proves only self-consistency. What makes it equivalence is that
  the same golden was diffed key-for-key against `GOLDEN_2026-07-31c`, captured
  from the HAND-BUILT site: **94 shared keys, 0 moved.** Keep `07-31c`; it is the
  only record of what the original computed and every equivalence claim is stated
  against it.
- **"the lock protects the calculators"** — `save_page_sections`' DELETE carries
  the agent-writable predicate, so a locked row survives it and the blocked write
  emits `lock_blocked_change` for review. Proved with the exact predicate rather
  than assumed: on `tool-standard-calc`, five prose rows return
  `agent_may_delete=t` and the tool row returns `f`.

## 6. Still with the owner

- **`GITHUB_READ_TOKEN` cannot see `gqls/sites`** — a fine-grained PAT scoped to
  selected repos, 404 while authenticated. Needs GitHub admin. Off the critical
  path: the faithful bytes were already in the database.
- **Whether to create a site PLAN.** There are zero `site_plans` rows, which is why
  the reconciler currently has nothing to act on. A plan is what would let the loop
  add and reshape pages rather than only improve existing ones — and it is the
  moment `reconcile_site_plan`'s ownership guard stops being dormant. Its own
  decision.

## 7. Neighbouring lanes — do not let them repeat this

`loanandmortgagecalculator.co.uk` (41 verbatim rows) and `loancash.co.uk` (18)
have **zero `site_components` rows**. If either decomposes without building chrome
first, `buildDefaultHead` links `/assets/css/styles.css` — plural — and both sites
serve `style.css`; every page ships unstyled with no header and no footer, and the
deploy reports success. Written up with the measurements in
`CONTRIB_2026-08-02_chrome_before_you_decompose.md` in each lane's directory.

## 8. Commits (this session, all `docs/` — no platform code changed)

`a3c386d52` decomposer + mirror + verifier + chrome, proven offline ·
`5659d8efd` chrome live, first page decomposed, ADO-039/040 ·
`c440832c7` shipping route + queue facts · `0f0c18162` mirror validated ·
`40c92cdde` copy-split verified · `5fba9dc59` contributions to both lanes ·
`4906ec38e` the false mismatch · `b904501c6` re-baseline procedure ·
`268408b35` SUMMARY (decomposition) · `41d0b8c07` new golden ·
`3dace925c` ADO-039 verify-later discharged · `4b909b6fd` 27/27 complete ·
`22778866b` flip to generic + locks + SUMMARY b · `173da1e41` post-roll check ·
`95b0902b3` prepare_work.sh reads the backup

**No `Council-Reviewed:` trailers, and none owed** — every file is under `docs/`,
which the gate refuses client-side. Two commits carry
`Council-Submitted: pending-…`, which is a **placeholder and a mistake** I
repeated from a previous round; forward-only forbids the amend. Do not copy it.
