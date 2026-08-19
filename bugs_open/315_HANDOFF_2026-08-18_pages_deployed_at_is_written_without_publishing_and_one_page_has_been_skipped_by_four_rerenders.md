# 315 — `pages.deployed_at` is stamped whether or not the object is written, and one page has now been skipped by FOUR completed rerenders

**Filed 2026-08-18** by the `webdesign_tool_rebuilds` lane. **Status: OPEN, UNOWNED.**
Two findings: a **measurement defect** that makes the failure invisible (§2), and a **live instance**
of the failure it hides (§3). The measurement defect is the more important of the two.

## 1. The one-paragraph version

`webdesign.co.uk/tools/seo-injector/index.html` serves a tool that was replaced hours ago. The
database is correct, four `page_rerender` items have completed with no error, and
`pages.deployed_at` has been stamped fresh each time — while the **origin object has not been
rewritten since 14:12:06Z**. Nothing anywhere in the platform records that the publish did not
happen; `deployed_at` says it did.

## 2. THE MEASUREMENT DEFECT — `deployed_at` is not evidence of publication `[MEASURED 2026-08-18 20:46Z]`

Three pages, each stamped `deployed_at` within the last half hour, against the origin's own
`last-modified` (fetched with a cache-buster; `cf-cache-status: DYNAMIC`, so these are origin headers):

| page | `pages.deployed_at` | origin `last-modified` | serving correct content? |
|---|---|---|---|
| `tool-seo-injector` | 20:45:57 | **14:12:06** | **NO** |
| `tool-json-cleaner` | 20:45:06 | 19:08:55 | yes |
| `tool-smooth-shadow` | 20:15:29 | 19:08:54 | yes |

**All three are stale against their own `deployed_at`** — including the two that are serving correctly.
So the column tracks "a rerender ran", not "bytes were written". Anyone using it to answer *did this
page publish?* gets yes for a page that has not been touched in six hours.

Note the second row's evidence value: the two healthy pages share a `last-modified` **to the second**
(19:08:54 / 19:08:55), which says publication happens in **batches**, decoupled from the per-page
rerender that stamps `deployed_at`. That is the seam where a page can be dropped silently.

## 3. The live instance

- Page `3d1fbd02-ae36-436a-a281-539ac285d4aa`, `/tools/seo-injector/index.html`.
- **DB is correct:** ported slot `15b8323c` `build_status='removed'` (18:57:00); native slot
  `2100c25e` `deployed`, and its stored `rendered_html` **contains the new component's marker
  (`scriptOpenTag`) and does NOT contain the old tool's `b-type`**.
- **Four rerenders, all `complete`, all `error IS NULL`**, orchestrations COMPLETED with no
  `__step_error`: 15:18:58, 17:10:29, 20:12:06, and a purpose-built republish at **20:45:59** filed
  with a distinct `item_key` specifically to rule out dedup silently swallowing it.
- **Origin unchanged throughout:** `last-modified: Tue, 18 Aug 2026 14:12:06 GMT`, content still the
  ported tool (`class="ported-page"` 1, `scriptOpenTag` 0).
- **Isolated, not systemic:** four sibling pages rebuilt the same way today are all serving correctly
  (`html-minifier`, `svg-optimizer`, `json-cleaner`, `smooth-shadow`). The publish seam works; this
  page is being skipped by it.

## 4. Why it stayed invisible until now

Every layer below the artefact reports success: the work item is `complete`, the orchestration is
COMPLETED, `deployed_at` is fresh, and the database holds the right HTML. **This is CLAUDE.md's
"trust the rendered artefact, not the status" with all four lower layers green.** It was caught only
because this lane grades at the served bytes with a cache-buster — and note that an hour earlier the
identical *symptom* on a different page WAS just a stale edge cache, so the cheap explanation was
available and wrong here.

## 5. Fix candidates, ordered by what closes the door

1. **Make `deployed_at` mean what it says**: stamp it only after a confirmed object write, and record
   the written hash/etag alongside. `pages.content_hash` exists and is **empty on all three pages
   above** — populating it at publish time would make "is the origin current?" a comparison rather
   than an assumption.
2. **Fail the rerender when the publish writes nothing.** A completed item that produced no object is
   the defect; it should be `failed` with the reason, not `complete`.
3. **Find why this page is skipped by the batch** — the two healthy pages share a to-the-second
   publish time, so there is a batch boundary; this page is falling outside it. Start from the
   publisher's page selection, not from the rerender.
4. Alert on divergence: a periodic sweep comparing `deployed_at` to the origin's `last-modified` for
   deployed pages would have caught this the first time, at 15:18.

## 6. How to verify a fix

`curl -sI "https://<domain>/<url>?x=$RANDOM" | grep -i last-modified` must move forward after a
rerender completes, on the page named above. Negative control: a page nobody rerendered must NOT move.
**Always cache-bust** — `cf-cache-status: DYNAMIC` confirms you are reading the origin.

## 7. Related

- `docs024_key_docs_latest/webdesign_tool_rebuilds/NOTES_…` 2026-08-18 20:12Z and 20:46Z (full evidence)
- CLAUDE.md "Trust the rendered artefact, not the status" · `MEMORY/prove-a-deploy-at-the-artefact-index`
- The publish-seam canary from another lane (commit `a2a9912c2`, "served sha256 == pre-publish origin
  hash, published_hash written only after acceptance") — that canary proves the seam CAN work; this is
  a page it is not reaching, and the acceptance idea in it is fix candidate 1.

---

## Contribution, 2026-08-19 — a SECOND live instance, and the class is 42 pages across 14 sites

**Not a rival diagnosis and not a fix attempt.** Found by the `agentchassis-22` session while
measuring an unrelated question for the `bugfix_277_required_fields_repair` lane; verified and sized
by that lane. **Neither of us owns this bug and neither is picking it up** — this is here so it is not
lost in a NOTES appendix.

### A second instance of your §3, on a different site and a different site's lane

`vetcomparison.uk` — page `tool-compliance-deadline-calculator`:

| | |
|---|---|
| `pages.status` | **`active`** |
| `pages.build_status` | **`planned`** |
| `page_components` | **0** |
| created / last updated | 2026-07-17 / 2026-07-26 |
| served | **404** (byte-identical to a fabricated-URL control at 2,690 bytes, so the 404 is a real absence, not a fetch artefact) |
| `page_rerender` work items | **3, all `complete`** — 2026-08-11, 08-12, **08-18** |

**Three rerenders completed successfully on a page with nothing to render.** Same shape as your four,
one month older, and it has been `active` and unserved since 2026-07-17.

⚠ It also carries **4 `unbuilt_internal_link` items parked at `needs_human_review`** (2026-08-11).
So the estate *did* detect that links point at an unbuilt page and then stranded the finding — that
half is `bugs_open/083`'s disease, not yours, and is noted only so nobody reads the parked items as
this bug being handled.

### The class, measured — your §3 is not a singleton

[MEASURED 2026-08-19] pages with `status='active'` and **zero** `page_components`:

| `build_status` | pages | sites |
|---|---|---|
| `planned` | **42** | **14** |
| `needs_rebuild` | 11 | 6 |
| **`deployed`** | **2** | 2 |

**The 2 at `deployed` are the sharper version of your bug** — the estate believes those are published
and they have no components at all. The 42 at `planned` are the softer one: never built, but
`status='active'` and therefore link-target-eligible.

### And the detector that should see this files nothing

`diagnose_silent_check_action.go` already carries **two** checks for exactly these shapes:

- `gatherNavLinkedNeverBuilt` — `build_status='planned'` past a grace period, nav-linked, uncovered;
- `deployed_zero_components` — *"page built/deployed but serving zero components"*, and it is
  **`EmitDefault: false`**, described in its own registration as **REPORT-ONLY** because it *"may be
  a deliberate content removal"*.

[MEASURED 2026-08-19] `SELECT ... WHERE item_type ILIKE '%never_built%' OR '%nav_linked%'` returns
**zero rows fleet-wide, all time.** So the detection exists and has produced nothing — **undriven
rather than missing**, which is the distinction that decides whether the fix is code or a schedule.

⚠ **This matters to your §2 specifically:** you argue the measurement defect is the more important
finding. This supports that from a second direction — there are at least two checks that would have
surfaced this class, and one is deliberately silenced by an `EmitDefault: false` whose stated reason
("may be a deliberate content removal") is a judgement nobody has re-examined against a population of
2.

### What we are NOT claiming

The cause of any of it. We have not looked at why the rerenders complete, why the 42 sit at
`planned`, or whether `deployed_zero_components`' report-only default is still the right call. All
first-hand, all re-runnable, none diagnosed.

### ⚠ CORRECTION to the contribution above, same day — my count was 4x low, and the class figure is ~100x bigger

**The error:** I wrote *"3 `page_rerender` items all `complete`"* for
`tool-compliance-deadline-calculator`. That was **`site_work_items` only**, which the
`work-item-archiver` prunes to roughly a 7-day window. Over `site_work_items UNION ALL
site_work_items_archive` it is **13 completed rerenders**, not 3.

Caught by the `bugs_open/302 201` session, who volunteered the trap unprompted: the archive held
**20,184 rows against 10,689 live** when they measured it yesterday, and it had changed two of their
own figures by more than 20×.

⚠ **Which of my numbers were affected, stated so you do not have to guess:** the **42 / 11 / 2 page
counts are NOT affected** — they are computed over `pages` and `page_components`, which are not
archived. **Only the work-item counts were window-limited**, and they were the ones that made the
point.

### The class figure, re-measured over live + archive — and it is the strongest statement of your §2

Pages with `status='active'` and **zero `page_components`**, against every work item ever filed
against them:

| `build_status` | pages | sites | **COMPLETED work items** |
|---|---|---|---|
| `needs_rebuild` | 11 | 6 | **166** |
| `planned` | 42 | 14 | **130** |
| **`deployed`** | **2** | **2** | **35** |
| **total** | **55** | — | **331** |

**331 work items reported success against 55 pages that contain nothing.** The 35 against
`deployed` pages are the sharpest: the estate believes those two are published, they have no
components, and thirty-five items have completed against them.

### And WHY that is evidence about the standard rather than about these pages

From the `bugs_open/302 201` session, who own the adjacent guard and were precise about the boundary
(they declined to adopt this population as a test set for their own work, correctly, because their
guard pins a *predicate* and this is an *artefact* disagreement):

> the sweep completes on **positive orchestration evidence** — "the handler orchestration I
> dispatched reached COMPLETED" — which is explicitly **parity with the lost `mark_complete` write,
> not a stricter test**; migration `220`'s header says so in terms.

**A page with 13 completed rerenders and nothing rendered is exactly what that parity cannot
distinguish.** So this population is evidence about **whether positive-orchestration-evidence is a
sound completion standard at all** — which is your §2's thesis, arrived at from a third direction and
with a number attached.

**Still claiming no cause**, and still not ours: neither this lane, nor `agentchassis-22` who found
the first page, nor the `302/201` session is taking `315`.

---

## Contribution, 2026-08-19 (later) — diagnosis + a fix plan at the council. Three corrections to this file.

Picked up as a **fix** lane (`docs024_key_docs_latest/bugfix_315_deployed_at_without_publication/`,
standing five present). The filing lane says twice it is not theirs to fix and `who-owns.py` confirms
no competing fix thread. **Nothing has shipped**; phases 0–2 are at the council gate as
`Council-Submitted: 377167cd-6324-4bc7-a866-87ad8c435132`.

### §2 is RIGHT, and its evidence table is two findings, not one

Register `DGH-009` (`docs026_concept_register/register/deployment-github.md:101`) records the
mechanism that explains the table:

> *"**`success:true` from the git-adapter is not evidence anything changed.** An unchanged file
> commits as an EMPTY commit and the adapter reports success with the file listed in `deploy_result`."*

A byte-identical rerender ⇒ empty commit ⇒ `b2 sync` rewrites no object ⇒ `last-modified` correctly
stays put. **So `tool-json-cleaner` and `tool-smooth-shadow` are not defects** — their bytes did not
need rewriting, and the column is not lying about them. Only the `seo-injector` row is the bug.

This does not weaken §2, it sharpens it: `deployed_at` honestly means *"a rerender ran and its
output was committed"*, has never meant *"the origin now serves these bytes"*, and the gap only does
damage when the bytes **did** change and still did not arrive.

**But it kills §5 candidate 4 as written.** A `deployed_at`-vs-`last-modified` sweep would have
convicted both healthy pages above. `[MEASURED 2026-08-19]` it is worse than that: 40 sampled
deployed pages ALL have an origin `last-modified` older than their own stamp, and **all 40 share one
three-second window** (09:33:56–58) while their stamps spread over the following hour — that is the
whole-domain `b2 sync` batch, seen directly. The comparison cannot separate *not synced yet* from
*will never sync*; only elapsed time can, and the known bad case took six hours. **Only an
intent-vs-reality content hash separates the three rows**, which makes candidate 1 a prerequisite
for candidate 4 rather than an alternative to it.

### §2's core claim is now measured at the config, not inferred from three pages

`[MEASURED 2026-08-19]`, joining every live `agent_definitions` step on `next_step` (**not** on
`jsonb_each` key order, which is arbitrary and reverses this particular answer): **19 `git_commit`
steps across 16 agents; 6 `update_page_status` steps across 6 agents.** Five stamp `deployed`:

| agent | preceded by | so the stamp is |
|---|---|---|
| `page-build-handler` | `save_page_sections` | **BEFORE any deploy is dispatched** |
| `tool-recreation-handler` | `save_page_sections` | **BEFORE any deploy is dispatched** |
| `page-rerender` | `git_commit` | after a commit whose result it discards |
| `report-builder` | `git_commit` | after a commit whose result it discards |
| `section-editor` | `git_commit` | after a commit it discards, then deploys again |

`deploy_result` appears **nowhere** in `v3_site_actions.go`. There is no arrangement of these five
under which the column could be evidence of publication.

### Correction to the 08-19 contribution's sizing — none of its three named instances is this bug

Checked at the DB and at the served artefact:

| named instance | measured | verdict |
|---|---|---|
| `vetcomparison.uk` `tool-compliance-deadline-calculator` | `build_status='planned'`, **`deployed_at` IS NULL**, still 404 today | real, but **nothing ever stamped it** — an active-never-built page, not a false claim |
| `idea.uk` `/tools.html#audience-check` | `deployed_at` **NULL**; a separate `/tools.html` row with 4 components serves 200 | a **phantom row whose `url` is a FRAGMENT** of another page |
| `ai-agent-orchestration.com` `/roi-estimator.html` | `deployed_at` 2026-05-02, **serves 200, rewritten today 08:37:59** | a stale duplicate row; the URL is live and current |

The "42 / 11 / 2" table sizes *componentless active pages* — a real, overlapping population already
targeted by `check_componentless_pages` — but it **does not size this bug**, and the two rows it
offers as the sharpest cases are the two that are not cases. Flagging it because it is the only
sizing this file carries and a fix aimed at it would be aimed at the wrong population.

### The two columns this needs already exist, and are dead

`[MEASURED 2026-08-19]` `pages.content_hash` **0 of 786**; `page_components.deploy_commit` **0 of
1,775**; and `grep -rn "deploy_commit" --include=*.go .` over the whole repo **including tests**
returns **zero lines**. `pages` and `site_work_items` have no commit/sha column at all.
`CommitToRepo` computes `newCommitSHA` and returns `repo.HTMLURL` — a per-repo constant — so the sha
never leaves the adapter.

`UpdatePageStatusInputSpec.RemovedConfigKeys["commit_from"]` already describes this feature and says
*"Implement it as a feature if wanted, do not re-add the key"*, and
`sql_for_agents/034_page_rerender_agent.sql:99` already promises
`"deploy_result": "git commit result with commit_sha"`. **So §5 candidate 1 is wiring up
designed-and-abandoned machinery, not inventing it** — which is why it is cheaper than it reads.

⚠ I corrected register `DGH-001` in place: it claimed *"Commit SHAs are recorded on pages and work
items for traceability"*, which is false in all three parts. Council seats read register entries as
ground truth, so that line would have drawn an objection to a proposal to add what it says exists.

### Two things I could NOT settle

- **§5 candidate 3 (why one page falls outside the batch) is not diagnosed.** The runner workflow
  lives in `gqls/sites/.github/workflows`; the repo is **private** (`api.github.com` → `Not Found`
  unauthenticated) and the chassis holds no B2 credentials, so neither the ref nor the bucket is
  readable from here. The proposed sweep is deliberately designed to *detect* that failure from this
  side without reading the runner.
- **The diagnosis loop returned no verdict.** Two dispatches
  (`6f900e18-2106-4145-a84c-811baeceaa0d`, `f1433782-6ba7-4304-a7f9-8bd830dfb7c9`) both died at the
  `verdict` step on the Anthropic usage cap. Per the owner ruling of 2026-07-31 I state the
  substitute plainly: every function above read at source, the workflow graph measured live, 744
  `deploy_result` rows censused over 7 days, 40 pages graded at the artefact with cache-busters, and
  the runner pods' own job logs read. If the loop ever completes and refutes any of this, it wins.
