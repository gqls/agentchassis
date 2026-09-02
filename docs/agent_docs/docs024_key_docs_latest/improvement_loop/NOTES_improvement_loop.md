# NOTES — improvement loop

Append-only, newest at the bottom. Technical log: evidence, commands, what the system
actually said, and every misstep.

---

## 2026-09-02 — lane opened, first state assessment

**(a) Started from the wrong premise and caught it in one query.** Auto-memory and
`bugfix_136`'s file both carry the owner ruling of 2026-07-29: *the improvement loop is
stopped DELIBERATELY … do not re-enable it*, evidenced by `improvement-sweep`
`enabled=f` since 2026-05-02. I nearly wrote a plan around a stopped pipeline. The live
row says `enabled = t`, last triggered 2026-09-02 11:59:27Z. Migration
`389_park_contrast_failures_and_reenable_improvement_sweep.sql` turned it back on.
**The cheap check that caught it was the first query I ran** — one `SELECT` against
`scheduled_tasks`. Recorded because the ruling is five weeks old and is still being
cited as current by at least two documents.

**(b) `execution_path` is empty on every improvement-loop row.** My first attempt to
separate "audit ran" from "audit skipped" was
`execution_path::text LIKE '%discovery%'`, which returned `f` for all 98 rows — i.e. it
told me the loop had never run a discovery step, on a pipeline I had just watched
complete. The column is `[]`. `collected_data`'s keys are the real record. A false
negative that looks like a finding: exactly the shape to distrust.

**(c) The live workflow does not match `004_improvement_loop.md`.** 004 documents a
3-pass audit cap; the live `collected_data` carries `check_audit_due`,
`check_not_converging`, `load_audit_state` and a site fingerprint. Migration `291`
replaced the cap with a convergence gate on 2026-08-02 and 004 was never updated.

**(d) The measurement.** `[MEASURED 2026-09-02]`, 2-day window (orchestration retention
is ~1 day, so this is the whole record): 98 runs, 32 domains, fair rotation.
80 `complete_clean` / 15 `complete` / 2 failed. `audit_due=true` on 24 — the gate
discriminates. 136 items promoted. **Every single run reported `not_promotable > 0`.**

**(e) My first backlog figure was wrong, by 2.8×, and the marker would not have caught
it.** I summed `not_promotable` across runs and got 3,866. That number is meaningless:
`not_promotable` is a per-run count of the site's *standing* pile, so a site visited
five times contributes its pile five times. The standing figure, counted as rows, is
**1,385**. Both would have carried a `[MEASURED 2026-09-02]` marker honestly. The
marker rule says a figure must be disconfirmable — a sum over a rolling re-count could
not have come out any other way.

**(f) The backlog is real and it is growing.** 1,385 `detected` rows with no handler,
31 sites, oldest 2026-07-26. The `bugfix_284` lane recorded 722 of this class on
2026-08-19. Near enough doubled in a fortnight.

**(g) Confirmed there is no consumer, three ways.** (1) `detected-item-promoter`'s live
`pre_query` states it: *"Flag-only rows (no handler_agent) are NOT here"*. (2) A grep
for `handler_agent IS NULL` / `handler_agent = ''` across `*.go`, `*.sql`, `*.tsx`,
`*.py` returns only migrations, one-off repair SQL and table DDL — no agent, no report,
no dashboard. (3) The peer state exists and is populated: **912** flag-only rows sit at
`needs_human_review`, which IS looked at. So the class splits across two states and only
one is visible.

**(h) Then I probed the rows instead of counting them, and the picture changed.**

- 978 of the 1,385 are `head_essentials_missing`. Broken down by what is actually
  missing: **867 are a skip link alone**, 55 skip-link + footer, 56 all three. So the
  dominant finding is ONE fleet-wide template omission filed 867 times per-page.

- The 56 "all three" rows are on two domains. I curled them, with an invented-URL 404
  control on the same domain (control: 9 bytes, no title — the probe discriminates):

  - **farmerinsurance.uk (36 rows): the claim is two-thirds FALSE.** `/about.html`
    returns 200, 66,108 bytes, `<title>About | Farmer Insurance UK</title>`, one
    `<footer>`. `/blog/crop-insurance.html` likewise. Only the skip link is absent.

  - **boxingonline.com (20 rows): true, but not about our pages.** `/`, `/about.html`
    and `/index.html` all return 200 with the same 114 bytes:
    `<!DOCTYPE html><html><head><script>window.onload=function(){window.location.href="/lander"}</script></head></html>`.
    The domain is parked. The finding is a true statement about a lander and a
    misleading one about a page of ours.

**(i) The staleness mechanism, read in code rather than inferred.** `insertWorkItem`
(`platform/orchestration/actions/load_work_item_actions.go:1787`) writes with
`dropOnConflict`, so a re-run of the check drops the fresh row and the original
`spec.missing` stands. `HeadEssentialsMissingCheck` only emits a `ResolvedFinding` when
`len(missing) == 0` (`check_site_structural_validity.go:1116`). Since the skip link is
never present, a farmerinsurance row can never be retracted and never be refreshed — it
carries its first-ever missing-list indefinitely. **Consequence for anyone building a
consumer over this pile: `spec.missing` is a claim of unknown age.**

**(j) What I have NOT established, marked as such.** `[UNMEASURED]` whether the skip
link is genuinely absent from the chrome template fleet-wide, or absent only from these
26 sites' rendered output. `[UNMEASURED]` whether the other ten item types in the pile
carry the same staleness — I checked the mechanism, which is shared, but I have probed
served pages only for `head_essentials_missing`. `[ASSUMED]` that boxingonline.com's
parking is deliberate; that is decision D2 for the owner, not a fact I hold.

---

## 2026-09-02, later — the skip link has ONE cause, and it is provable

Followed plan item 3's question — is the skip link absent from the chrome fleet-wide, or
only from these sites' rendered output? `[MEASURED 2026-09-02]`

**(k) The estate shares one header component, and it has no skip link.** Of 34 active
sites, 33 carry a `header` slot in `site_components`. **32 of them point at the same
`component_id` `58fde68f-9190-4e5e-b6a5-ea21cf27a9af`** (three of those are forks:
`idea.uk` `f420f3fa…`, `leopardessconsulting` `990b7162…`, `webdesign.co.uk`
`ad6033ae…`). **Not one of them renders a skip link.**

**(l) The single exception is the one with no shared component at all.**
`loanandmortgagecalculator.co.uk` has `component_id = NULL` — a hand-owned header,
updated 2026-08-05 — and it is the only site on the estate whose served page carries
`<a class="skip-link" href="#content">Skip to content</a>`. So this is **not a missing
capability**: the platform renders a correct skip link today, on exactly one site, and
that site is the one not using the shared component. ⚠ That header belongs to the LMC
lane; this lane must not edit it.

**(m) The fix has a prerequisite, and skipping it would manufacture a new finding on
every page of the estate.** A skip link needs a target. LMC's points at `#content` and
its pages carry `id="content"` (2 occurrences on the front page). The shared-header
sites do NOT: probed finetuning.uk, webdesign.co.uk and cookly.uk — all three render a
`<main>` element and **zero** `id="content"` and **zero** `id="main"`.

  So adding the link to the shared header alone would produce ~1,000 dangling fragment
  links — and `check_phantom_internal_links_fragments.go` exists and would file every
  one of them. **The fix is two components, not one**: the skip link in the shared
  header AND an id on the page shell's `<main>`. That check's own header comment already
  names the contract — *"header skip-link targets id='content', which its pages carry"* —
  so the shape is settled; it is the shared sites that are missing half of it.

**(n) A second parked domain, found by accident, and NOT flagged.**
`adversecreditmortgage.co.uk` serves the same 114-byte lander stub as boxingonline.com
on `/`. It has **19 pages recorded active** and is one of only two active sites with
**no** `head_essentials_missing` finding at all. So the check that would have caught it
did not fire there. `[UNMEASURED]` why — I have not read the check's page-eligibility
gate. Worth doing before anyone treats "no finding" as "no problem": on this evidence,
the two sites with the cleanest record are a parked lander and the site nobody flagged.

**(o) Checked the fan-out mechanism before assuming the fix could ship, and it is
sound.** `bugs_open/404` says `template_changed` was in the live re-render vocabulary
while `create_rerender_items_action.go` knew neither it nor `literal_markdown` — which
would have made a chrome fix complete green and ship nothing. That is **repaired in the
tree**: the file now derives its vocabulary from `livespec.RerenderSectionReasons`, the
single definition, asserted daily against the live gate by `config-key-audit
--live-declaration-drift`. `[UNMEASURED]` whether that repair is live in the running
image — to check at the time of the fan-out, not now.

**(p) Where this leaves the backlog arithmetic.** Of the 1,385:
- **867** are the shared-header skip link — one two-component fix.
- **~110** more (skip_link+footer, and the 56 all-three) are the same fix plus something
  else; the farmerinsurance 36 are the same fix plus a stale spec.
- **20** are boxingonline being parked, and belong to whoever owns that domain, not here.
- **~390** are the other ten item types, unexamined by me so far.

So the honest size of "findings that need a human decision" is **nowhere near 1,385**,
and a screen showing that number would have been the wrong instrument. This is why plan
item 4 sits behind items 2 and 3.
