# NOTES — bugfix 328, a deployed page linking at a page that never shipped

Append-only, newest at the bottom. Technical log: what was tried, what the system
actually said, and every misstep.

## 2026-08-23 — session open: validity re-check and the census

**The bug is still live.** Measured 2026-08-23 ~17:10Z, cache-busted (`?cb=$RANDOM`):

| url | code |
|---|---|
| `https://loanzy.uk/` | 200 |
| `https://loanzy.uk/your-rights.html` | **404** |
| `https://loanzy.uk/guides/index.html` | **404** |
| `https://mortgagecalculator.co.uk/scorecard-simulator.html` | **404** |
| `https://mortgagecalculator.co.uk/disclaimer.html` | 200 |

And the anchors are still there — `curl -s https://loanzy.uk/ | grep -o 'href="[^"]*"' | sort | uniq -c`
returns `2 href="/your-rights.html"` and `1 href="/guides/index.html"` [MEASURED 2026-08-23].
So both halves of the bug's claim hold four days after filing: the target 404s and the
referrer advertises it.

### Ownership

`./scripts/who-owns.py 328` reads OWNED — but by the two lanes that CONTRIBUTED to the file
(`mortgagecalculator_couk_adoption`, `bugfix_260_render_fallback`), not by a lane fixing it.
No `bugfix_328*` directory existed before this one. None of the link-repair source files were
dirty in the tree at session start. `bugfix_313_internal_linker` is adjacent but opposite — it
makes the linker ADD links; it is not fixing this.

### The census — 63 open items, 7 sites, 13 target pages

```sql
SELECT s.domain, count(*) AS items, count(DISTINCT w.page_id) AS distinct_targets, min(w.created_at)::date oldest
FROM site_work_items w JOIN sites s ON s.id=w.site_id
WHERE w.item_type='unbuilt_internal_link'
  AND w.status NOT IN ('complete','verified','cancelled','rejected','wont_fix')
GROUP BY 1 ORDER BY 2 DESC;
```

lendzy.co.uk 40 · mortgagecalculator.co.uk 8 · vetcomparison.uk 7 · loanzy.uk 3 ·
robot-hands.com 2 · leopardessconsulting.co.uk 2 · gaswholesalers.com 1. Oldest 2026-08-09 —
~~**14 days of live 404s**~~ [MEASURED 2026-08-23].

> **CORRECTED 2026-08-23, same session — "14 days of live 404s" was WRONG about ~2/3 of this
> population, and I wrote it here before checking.** **42 of the 63 open items name targets that
> serve HTTP 200 today**: all 40 lendzy items (3 tool pages), gaswholesalers'
> `/fuel-pricing-framework.html` and mortgagecalculator's `/contact/index.html`. Those are stale
> records held open by a missing `deployed_at` stamp — the `bugs_open/315` family — not live
> damage. **A work item records what a detector saw; it is not evidence about the wire.**
> Caught by the second model reviewing this lane's framing, after the number was already written
> down. The real harm had to be measured a different way entirely (the blast-radius query below:
> **36 anchors on 24 deployed pages → 14 unservable targets, 9 sites**). Logged in
> `WRONG_CALLS.md`; the cheap check is **curl the target before calling a queue row live damage**.

Fleet page states, same day: `deployed` 679 · `needs_rebuild` 52 (38 with a non-NULL
`deployed_at`) · `planned` 42 (0 ever deployed). So **56 pages fleet-wide would 404 if linked**
— 42 planned plus the 14 needs_rebuild that never shipped. Note the shape: `build_status`
alone does not answer it (38 `needs_rebuild` pages ARE serving), which is the 08-21 contrib's
point about `contact-index`, and why the estate's one predicate is `deployed_at`-based.

### THE CORRECTION THAT REPRICES THE FIX — the handler ran, and its remedy does not work

The bug file's 08-21 contrib says the items "sit at `status='needs_human_review'` with no
handler", i.e. detection works and **delivery** is missing. That is not what the rows say.

```sql
SELECT status,
       CASE WHEN error ILIKE '%no sections ready to build%' THEN 'no sections ready'
            WHEN error ILIKE '%content validation failed%'  THEN 'validation blockers'
            WHEN error ILIKE '%SECTION COMPONENT FLOOR%'    THEN 'component floor'
            WHEN error ILIKE '%AI endpoint unav%'           THEN 'AI endpoint'
            ELSE 'other' END AS why,
       count(*)
FROM site_work_items WHERE item_type='unbuilt_internal_link' GROUP BY 1,2 ORDER BY 3 DESC;
```

| status | why | count |
|---|---|---|
| needs_human_review | no sections ready to build | **48** |
| needs_human_review | content validation blockers | **10** |
| cancelled | — | 9 |
| failed | AI endpoint / other / component floor | 5 |

Every one of the 58 parked rows carries `triaged_at`, `handler_agent='page-build-handler'` and
`attempt_count >= 1`. **They were dispatched. The handler ran. It could not build the target.**

So `bugs_open/220`'s routing is not the gap — 220 shipped it and proved it on 2026-08-09. The
gap is that the item type implements exactly ONE remedy, *build the target page*, and on 92% of
the live population that remedy is unavailable. The other remedy — stop the referrer
advertising it — is implemented nowhere. A link whose target cannot be built stays live
for ever, and the work item's parked status is the only trace.

[MEASURED 2026-08-23] — re-run before quoting; a census goes stale by addition.

### The code hole, read not inferred

`datahelpers/link_repair.go`'s `RepairPageLinks` already unlinks a dead in-body link at four
seams (build gate, both rerender paths, the persistence chokepoint, and the `content_data`
pass). Its definition of a real page is `loadValidPagePaths`
(`validate_page_content.go:1515`) — `SELECT url FROM pages WHERE site_id=$1 AND
linkablePageStatusPredicate`. **There is no deployment test in it at all.** A page row that has
never shipped is, to every one of those four seams, a perfectly good link target.

The same question is already answered correctly one layer over, for chrome:
`chrome_link_policy.go` (`bugs_open/191`) builds its set from `loadFetchablePageSet`, which
applies `NOT (datahelpers.NeverDeployedPagePredicate)` — and carries two deliberate degrade
escapes (lookup failure; zero deployed pages, because a first build has no signal). 328 is the
PAGE-CONTENT spelling of 191.

### It reproduces on a build that ran TODAY — this is not a stale artefact from 08-19

loanzy.uk redeployed its whole page set today. `pages.deployed_at` for all 17 live pages falls
between **13:13:05Z and 14:24:05Z on 2026-08-23**, and the home page (`index`,
`/index.html`) was deployed at **13:28:27Z**. At that instant:

| target the home page links to | build_status | deployed_at | last touched |
|---|---|---|---|
| `/your-rights.html` | `needs_rebuild` | **NULL** | 2026-08-18 21:29Z |
| `/guides/index.html` | `planned` | **NULL** | 2026-08-18 20:42Z |

So the page was rendered and shipped **4.7 days after** either target had last been touched, with
no open work item on either, and it shipped the anchors anyway. The information the fix needs was
not merely available at render time — it was unambiguous. [MEASURED 2026-08-23 17:2xZ]

This also kills the cheapest possible reading of the bug ("old artefact, fixed since"): the
binary and config that ran four hours ago still do it.

### Second finding — detection did NOT cover this instance

The 08-21 contrib says the platform "is NOT blind to this ... detection exists, is
per-linking-page, and is accurate". On mortgagecalculator.co.uk it was. **On loanzy.uk it is
absent**: there is no `unbuilt_internal_link` item anywhere for `/your-rights.html` or
`/guides/index.html`. The site's only three items name a different page's component
(`about:about-content` → how-it-works, lenders/index, eligibility-checker), all filed 2026-08-18.

```sql
SELECT w.item_type, w.status, left(w.summary,110), w.created_at::date
FROM site_work_items w JOIN sites s ON s.id=w.site_id
WHERE s.domain='loanzy.uk'
  AND (w.summary ILIKE '%your-rights%' OR w.summary ILIKE '%guides/index%'
       OR w.item_type IN ('unbuilt_internal_link','phantom_internal_link'))
ORDER BY w.created_at DESC;
```

The audit is periodic and runs over DEPLOYED html; loanzy's last discovery pass predates today's
redeploy. So the two live 404s on that home page have never been filed by anything. **A
prevention fix at render time does not depend on the audit having run** — which is the third
argument for candidate 1 over candidates 3 and 4, on top of the two the bug file already makes.

⚠ Corollary I must not forget when designing: `your-rights` carries **two `unresolved_cta`
items** from 08-18 ("no real-page destination for primary_cta_url"). The build-time CTA resolver
already refuses to invent a destination and reports it. That is the same judgement this fix
needs, one layer up, already present in the estate — worth reusing rather than restating.

## 2026-08-23 — the census that told me the opposite of the truth, and the control that caught it

I needed to know which never-shipped pages actually 404, so I fetched **all 56** fleet-wide,
cache-busted, and tabulated by `build_status`. It came back:

| class | 200 | 404 | unreachable |
|---|---|---|---|
| `planned`, never deployed | **19** | 17 | 6 |
| `needs_rebuild`, never deployed | 9 | 3 | 2 |

That first row is a refutation. `PageMayBeLinkedPredicateFor`'s floor rests on a 2026-08-09
measurement of "22 such pages, **all 22** return 404", and 19-of-42-now-serving says it has gone
badly stale. I started writing that the floor needed replacing.

**It was one parked domain.** `adversecreditmortgage.co.uk` returns **200 with a 114-byte
registrar redirect for every path**, including `/definitely-not-a-real-page-24480.html`, which I
invented at the prompt. All 19 came from it. Excluding it and the two unreachable `*.internal`
hosts, the surviving 29 rows read:

| class | 200 | 404 |
|---|---|---|
| `planned`, never deployed | 0 | **17** |
| `needs_rebuild`, never deployed, **≥1 rendered component** | **9** | 0 |
| `needs_rebuild`, never deployed, **0 rendered components** | 0 | **3** |

So the 08-09 finding holds exactly, and a **new** discriminator falls out: **20 never-shipped
pages with zero rendered components → 20/20 404; 9 with at least one → 9/9 200.**

⚠ And the conjunction is load-bearing, which I checked rather than assumed: **8 pages fleet-wide
have `deployed_at` SET and zero rendered components** — tool and blog-index pages served by
another subsystem. A component test on its own would delist all eight.

The misstep is in `WRONG_CALLS.md` and the control is now a LANDMINE. The lesson is not "I made
an error" — it is that the figure was **dated, `[MEASURED]`, and re-runnable by anyone**, and was
still worthless, because the measurement could not tell apart the two states it existed to
separate.

## 2026-08-23 — blast radius, and the greenfield objection tested rather than argued

Every `<a href>` in every stored `page_components.rendered_html`, resolved against the predicate:

| target class | anchor hits | referring pages | targets |
|---|---|---|---|
| servable — kept | **3,193** | 577 | 557 |
| unservable — suppressed | **36** | 24 | **14** |

**1.1%**, every one a live 404. It also names damage the queue never filed — `remortgagecalculator.uk`
(2 targets, 6 hits), `webdesign.co.uk`, both `pool-*.internal`: 14 targets on 9 sites against the
queue's 13 on 7.

The objection I expected to have to concede — "this would gut a greenfield build" — did not
survive contact with a real one. loanzy's home page carries 10 distinct internal hrefs and **8 of
the 10 targets had already deployed** when it shipped at 13:28:27Z. The rule would have removed
exactly the two dead links. [n=1 on the ordering claim; the 3,193/36 split is the general one.]

## 2026-08-23 — the markup decided the repair shape, and a landmine nearly got multiplied

The obvious implementation is the existing unlink arm: drop the `<a>`, keep the text. There is a
LANDMINE saying exactly why not — *"where the anchor's inner content is the link LABEL — which is
every card, CTA and 'read more' control on this platform — the served page shows the label and the
arrow glyph as bare text in the middle of the card."*

So I read all 36 affected anchors instead of assuming. They split cleanly, and the split is
structural: **28 classless prose anchors** inside a sentence, **8 classed template controls**
(`info-card-grid__card-link`, `people-feature-block__link`) whose inner content is a short label
plus an arrow `<em>`. Owner's ruling: prose unlinks, a control is dropped whole. Both arms already
have estate precedent (`RepairPageLinks` and `DropDeadURLControls` respectively).

## 2026-08-23 — implementation, and two tests that were not load-bearing until they were

Built as designed. Two findings worth the next reader's time:

**`AssemblePageAction` is a second, entirely unprotected outbound seam.** I had concluded one seam
covered everything, because `page-build-handler.deploy_page` is a `call_agent` with
`target_role: page_renderer` → the `page-rerender` agent → `rerender_single_page` →
`repairOutboundPageLinks`. True, but incomplete: `assemble_page` is used by `pageflow-builder`,
`page-rebuild` and `site-work-orchestrator`, and calls **neither** repair function. It was the
second model's catch. ⚠ **It is invisible to the obvious query** — those three steps are nested in
`sub_workflow`s, so a `jsonb_each` over `default_config->'workflow'->'steps'` returns nothing and
reads as "no consumers". Only a recursive walk finds them.

**Five mutants, and two of my tests survived their first form** — i.e. were not tests at all:

| mutant | result |
|---|---|
| control arm's `class=` check removed | prose text deleted → test FAILS ✓ |
| `MarkupMatches` → raw regex | `<script>`-quoted anchor rewritten → test FAILS ✓ |
| runtime-fill exemption removed | test FAILS ✓ |
| zero-shipped-pages escape deleted | test FAILS ✓ |
| **opt-in default flipped to ON** | **PASSED — the test was vacuous** |
| **value-reference added in a persistence seam** | **PASSED — the scan was blind** |

The default-OFF test registered *no* query on the mock, reasoning that a flag-ON mutant would then
hit an unexpected query and fail. It does hit one — and the loader's **fail-open** branch turns
that into "return the html unchanged", which is byte-identical to the correct behaviour. The fix
is to register a query that *would* find a refused target, so the two worlds differ at the output.
The seam scan matched `suppressUnshippedOutboundLinks(` with a parenthesis, so
`var _ = suppressUnshippedOutboundLinks` walked straight past; it now matches the bare identifier,
skips comment lines, and carries its own synthetic-line gone-blind test.

⚠ **The shared tree does not compile** — another session's WIP in `v3_site_actions.go` (`undefined:
sectionMetaComplete`). Everything above was built and tested against a clean `git archive HEAD`
overlay with only this lane's files copied in, and the mutations were run *there*, so the shared
tree was never left in a mutated state.

## 2026-08-23 — council round 1: REVISE, and two seats found real defects

Corr `21c19c1f-e614-49bd-82ac-0bb5b58082e0`. 13 reviewers, 4 abstained. Decided by a **gating HIGH
from `prior_art_librarian`**.

### The gating HIGH was wrong, and the seat itself said how to check

It cited a landmine titled *"`RegisterVerifier` is not a drain — a completion verifier for a type
born `needs_human_review` guards a door that type never walks through"* and reasoned: **"IF the
landmine's full text says what its title fragment implies"**, then edits 6-7 do not do what I claim.

It does not. The entry's own CHECK paragraph, verbatim:

> Born `needs_human_review` ⇒ the verifier needs a sibling entry in `reviewRevalidators` (delegate to
> the same verify function so the two doors cannot disagree — `revalidate_truncated_component.go` is
> the worked pattern).

That is precisely and only what my change does. The landmine is not a warning against it, it is the
**specification for it**, and its own worked case (`truncated_component`, `bugs_open/325`) is the same
fix made four days earlier. Confirmed mechanically too: the sweep's loader selects
`status IN ('needs_human_review','unresolved')`, so parked rows are exactly what it reads.

**The transferable bit is not "the seat was wrong".** It read a title, flagged the risk, and hedged
the claim explicitly. That is the seat working — and it cost one grep to answer.

### Two seats found REAL defects. Both fixed in code, not argued away.

**`render_guardian` (medium) — my "3 publish paths" claim was overstated.**
`RerenderSitePagesAction` (`rerender_pages_actions.go:214`) appears in **no entry of
`GlobalActionRegistry`** — dead code, per a landmine I had not read. Corrected to **two** live paths
through seam 1. ⚠ The galling part: I had the evidence in round 1 — I measured that no live agent has
a `rerender_site_pages` step — and failed to connect it to my own coverage claim. A measurement you
take and do not apply is worth nothing.

**`bug_historian` (medium) — the `report-builder` exception.** I left it OFF as "not the measured
harm", and the seat named that as the estate's recurring shape: one call site guarded, the mechanism
generic everywhere else, "turn it on later" being exactly the reasoning that leaves the other path
live. So I **measured it instead of defending it**: `SELECT count(*) FROM pages WHERE
page_type='report'` → **0** fleet-wide, and none of the 24 pages serving a dead anchor is one. An
empty population cannot regress, so it is now ON, and `575`'s `$post$` gained an arm that RAISEs if
**any** live step running a seam action lacks the key — which also catches a sixth consumer appearing
between enumeration and apply. **The exception is gone rather than justified, and that is cheaper than
the argument would have been.**

### The objection I could not answer with a query — and had to read a bug file for

`bug_historian` also noted `bugs_open/097`'s title is
`cta_integrity_misses_card_links_to_unbuilt_pages` — "unbuilt pages", apparently my exact class —
while I had asserted 079/097 were a *different* class. **I had not read 097.** Reading it: its six
missed links are four *"never planned"* (no `pages` row at all — phantom) and two whose page exists at
a different url (`/cases/thames-water` vs the stored `/blog/thames-water.html` — rewrite). **Not one is
a `pages` row that exists and has never shipped.** Its actual cause is that the CTA check does not walk
arrays of child objects inside `content_data` (`info-card-grid` keeps destinations at
`content_data.cards[*].link_url`) — a detector-coverage bug.

So the distinction survives, **but my reasoning for it was unearned**, and 097's title is misleading
about its own contents. Worth stating positively rather than just defending: this change covers 097's
*surface* (card links) incidentally, because it works on rendered anchors — and its control arm is
aimed at exactly the `info-card-grid` shape 097 names. It does not fix 097.

### The rest, briefly

- **`prior_art` (medium)**: I cited `archived_page_guard.go:38` — a **code comment** — as authority for
  the `page_renderer` role mapping. Re-verified at the live config: `page-build-handler`'s
  `spawn_rerender_agent` is `{action: spawn_agent, config: {role: page_renderer, agent_type:
  page-rerender}, next_step: deploy_page}`, and `call_agent.go`'s `findTopicsForRole` matches a
  preceding `spawn_*` step's `role`. A comment is not evidence; the config is.
- **`guardian` (medium)**: failure isolation on the shared assemble seam — answered with a test, not a
  sentence (`TestADatabaseFailureNeverFailsTheAssembly`).
- **`bug_historian` (low)**: the assemble seam's silent skip on a bad `site_id` now warns — and only
  when the step opted in, because an un-opted-in skip is not an event.
- **`architecture` (low)**: optional-key budget run by name — `assemble_page` and
  `rerender_single_page` are in the **NOT COUNTED** population (no `ActionInputSpec`);
  `rerender_page_sections` is counted at 3, `over_budget` false, and this adds no key to it.
- **`architecture`/`guardian`/`reuse_agent`/`constitution` (medium ×4) — the third-instance question.**
  I declared it in round 1 and all four objected anyway, correctly: **declaring a debt is not
  discharging it.** Done as the architecture seat itself prescribed — an amendment ON `LNK-030`
  carrying the count, and `RFC_049` opened as the consolidation ticket. ⚠ I first wrote `RFC_048`;
  that number was already taken (the anti-churn brake, raised the same day by the `bugs_open/326`
  lane). Caught by `ls` before it reached the register.
- **`render_guardian` (low)**: suppression removes content with a log trail and no escalation.
  **Declined, on record** — it matches `RepairPageLinks`' existing precedent, the item an escalation
  would file already exists (the type this change also drains), and a second item per anchor would
  re-create `bugs_open/077`'s "items whose handler has no remit".

Round 2 submitted under the same correlation (`RESUBMIT_CORR`), so the trail accumulates.

## 2026-08-23 — council round 2: REVISE again, and this time the migration had the real defects

Gated by **`editquality` HIGH** — and it is right, but about the SUBMISSION, not the code. When
round 2 was squeezed to the gate's 8-file cap I itemised the ratchet TEST and left the
`reviewRevalidators` map entry in prose. So the plan read as shipping `revalidate_unbuilt_link.go`'s
function **with no caller** — which is exactly the shape the round-1 landmine warns about. The
registration was committed all along (`bb1e144b5`); the plan could not be checked for it. Fixed by
itemising the registration in round 3 and describing the test inside it.

**Same class, second instance, and this one was mine to avoid:** round 2's rationale said the
LNK-030 amendment and RFC_049 were "done in this round as **edit 9**". There is no edit 9 — the cap
is 8. Both `editquality` and `architecture` flagged it, and they were right to: from the plan alone
it reads as an aspiration. They ARE committed (`c4baa53e7`); they sit outside the edits array because
the gate refuses docs-scope files client-side. **The lesson is narrow and worth keeping: never
describe work by an edit number the plan does not contain — name the commit instead.**

### `debug_historian` found two HIGHs on the migration. One was structural and real.

**HIGH (a) — `is_active` is not the runtime selector.** There is a documented landmine that an agent
type can carry TWO active rows while only the higher `version` ever loads. My UPDATEs filtered on
`is_active` alone, so in principle they could write the dormant row, pass their own `$post$` count,
and leave the config that actually runs unguarded.

**Measured: all five target types carry exactly ONE active non-snapshot row today** (pageflow-builder
v21, page-rebuild v1, page-rerender v1, report-builder v1, site-work-orchestrator v2). So nothing
was wrong. **Fixed anyway** — every statement now keys on a temp table of
`DISTINCT ON (type) … ORDER BY version DESC`. The `_ROLLBACK` keys the same way, because for an undo
addressing a dormant duplicate is the *worse* direction: it would read as undone while the key stayed
live. **The reason to fix a latent one is that the day it stops being latent, nothing announces it.**

**HIGH (b) — read off my sketch, not the file.** The seat said the `jsonb_set` path
`{workflow,steps,render_page,…}` is blind to sub_workflow-nested steps. The sketch showed only the
`render_page` UPDATE; the file has three, with both nested paths spelled in full. **But the hazard
underneath is real**: `jsonb_set` with `create_missing=true` creates only the LAST key, so an absent
parent `config` object is a **silent no-op that returns the row unchanged**. Checked — all five paths
carry a `config` today — and `$pre$` now asserts the PARENT, not just the step, so a future drift
aborts instead of writing nothing and reading clean.

⚠ **Both HIGHs were about the same file and only one was a defect. Reading the sketch instead of the
file produced a confident wrong objection twice across two rounds** (round 1's gating HIGH was a
landmine title; this one a sketch). That is not a complaint — each cost one query to answer, and the
one that was real is the kind nothing else would have caught.

**MEDIUM, taken: no snapshot before a live `jsonb` mutation.** `snapshot_agent(type, reason)` now
runs for every row the file touches, before it touches any. The 340/336 precedent skipped this and
conceded it; a conceded omission is not a licence to repeat it.

### `bug_historian`'s MEDIUM was the sharpest thing in the round

*"Once a target page later ships, nothing in this plan triggers a re-render of the PAGES THAT LINKED
to it — so a suppressed anchor stays suppressed until that source page happens to rebuild."*

**Correct.** I had been writing "the anchor returns by itself on the next render", which is true and
quietly assumes a render happens. Measured instead of assumed: of the 25 pages carrying an anchor to
an unservable target, **24 were touched within 7 days — the most recent SIX MINUTES ago — and exactly
one is 60 days stale.** So the empirical restore path is strong, it is not a guarantee, and the tail
is one nameable page. The claim is now stated with that number attached, and Phase 1c makes it
deterministic for this population.

**`guardian` MEDIUM — the N+1.** With the flag on, `AssemblePageAction` issues one extra query per
page per build run. Costed rather than dismissed: one indexed SELECT over `pages` for that site plus
a correlated `NOT EXISTS`; sites here run ~15–45 pages, and it is the same shape the gate already
runs per page via `loadValidPagePaths`. Recorded as a named follow-up (a per-run cache), not built
here — a cache keyed to a build run is a lifetime question this edit does not otherwise raise.

Round 3 submitted under the same correlation.
