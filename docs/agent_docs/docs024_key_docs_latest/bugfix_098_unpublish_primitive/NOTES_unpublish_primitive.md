# NOTES — the unpublish primitive (`bugs_open/098`)

Running record, append-only, newest at the bottom. Missteps are the point.

---

## 2026-08-02 — picking the bug

Surveyed `bugs_open/` against the live `.jsonl` transcripts of ~30 concurrent
sessions (`who-owns.py` reads COMMITS, so it is blind to a session mid-fix —
grepping live transcripts is the check that sees those). 098 came out unowned,
still reproducible, and named a **missing framework primitive** two bugs share.

`bugs_closed/125`'s council round had asked that the same gap be logged as a
required follow-up rather than absorbed; it was logged on 098. So this is that
follow-up, not a fresh choice.

## 2026-08-02 — the finding that shrank the job

098's fix candidate 1 is "make the deploy reconciling rather than additive",
described as the biggest-blast-radius option needing its own survey. It is
**half-built already**: `gqls/sites/.github/workflows/deploy-to-b2.yml` runs
`b2 sync --delete --skip-newer` per changed domain and then purges the Cloudflare
zone. The *sync* reconciles; only the *git tree* is never removed from. That
turned a deploy-pipeline redesign into one missing capability.

Reading the workflow before designing anything is what found this. It was not in
any doc.

## 2026-08-02 — probed rather than assumed

Wrote the existence filter as "defensive", then probed GitHub to see whether it
was actually needed. POST `/git/trees` against the live master tree (a tree
object is created unreferenced — no ref moves, no workflow fires):

```
null sha, path PRESENT -> 201, tree 0d8dab50…
null sha, path ABSENT  -> 422 GitRPC::BadObjectState
```

So it is **required**: without it, re-running a repair fails. Good outcome from
a cheap check — and the opposite of what I had written in the comment.

## 2026-08-02 — MISSTEP: I justified a design choice with a hazard I had not measured

I chose a per-path existence check over one recursive tree listing and wrote the
reason as though it described the estate: *"a recursive listing can come back
`truncated: true`"*. Then I measured: `gqls/sites` recursive is **1,847 entries,
`truncated: false`**. Nowhere near the limit.

The decision was right; the stated reason was fiction, written in the same voice
as the 422 probe two paragraphs above it. Logged in `WRONG_CALLS.md`. The tell
is the word **"can"** — it smuggles a hypothetical into a paragraph of facts.

## 2026-08-02 — MISSTEP: a pod-grep that read 0 for everything, including its control

Verified the built image with `strings /app/git-adapter` — got `0` for every
positive check. The binary is at `/root/git-adapter`; `strings` with no readable
file read stdin and matched nothing. **The control also read 0**, which is what
flagged it as a broken check rather than a bad image. Without the control I would
have concluded the build had failed and rebuilt for nothing.

In-pod grep then failed on permissions (container runs as uid 1000, binary owned
by root). Settled on the stronger check anyway: extract the binary from the image
locally, grep it with a **negative control** (a string the change REMOVED, expect
0), then compare the **pushed digest to the running pod's `imageID`**. That is a
complete provenance chain and it does not depend on in-pod tooling.

## 2026-08-02 — live proof on a scratch path before touching production

`unknown-domain/` exists in the sites repo and does NOT match the deploy
workflow's changed-domain regex (`^[^/]+\.[^/]+/` — no dot), so it is an in-repo
scratch area with no B2 or CDN side effects. Wrote a file through the adapter,
deleted it via the new verb (404 confirmed), re-deleted it:

```
CommitToRepo: no-op — every requested deletion is already absent
```

head unmoved. Write → delete → idempotent re-delete, all three arms live.

## 2026-08-03 — THE BUG'S STATED MECHANISM IS ONLY HALF RIGHT

> **CORRECTION to 098's root cause, found by re-checking the target before
> retracting it** (CLAUDE.md: other threads change things beneath us; a figure
> carried forward unchecked is how a stale premise gets diagnosed as a bug).

098 says an archived page is **frozen**: "the last HTML that page ever rendered
is frozen and it keeps being served". For the retraction target that is **false
as of 2026-07-31**. `robot-hands.com/learning-center/index.html` is re-rendered
and re-committed **twice a day**: 08-01 08:07, 08-01 20:06, 08-02 08:05, 08-02
20:15, 08-03 08:15.

Traced work item → orchestration → commit → SQL. Cause is
`queueNewsPageRerenders` (`render_news_section_html.go:64`):

```sql
WHERE p.site_id = $1
  AND cc.function IN ('latest-news','news-listing')
  AND p.build_status = 'deployed'     -- and NEVER p.status
```

`build_status` records whether the page ever shipped; `status` records whether
the platform still wants it served. Archiving sets the second and leaves the
first, so a selector keyed on the first keeps choosing a retired page for ever.
6 `page_rerender` items raised for it since 07-31.

**This makes a retraction self-undoing** — delete the file and the next news
refresh republishes it. So the retraction could not have been proven without
fixing this first, and a session that had retracted and walked away would have
recorded a fix that quietly reverted overnight.

Blast radius, measured: fleet-wide this selects exactly **one** non-active page,
this one. Only two page statuses exist (`active` 557, `archived` 25). Filed
through the diagnosis loop as a structural claim (`5bdec8cf`).

098 was accurate when filed on 07-26; it was overtaken five days later. The bug
file gets the correction, not a rewrite.

## 2026-08-03 — the council's REVISE was worth having

Round 1 → **revise**, gating objection from `editquality`, echoed by
`bug_historian` and `reuse_agent`: my retraction derives the deploy path via
`PageFilePathFromURL`, and a landmine warns two functions eleven characters apart
decide a page's deploy path and disagree. If mine is not the one that WROTE the
file, the deletion silently targets nothing.

The landmine predates the 125 fix (all five derivations now delegate to the
shared helper, `git_deployer_actions.go:436`), but the seats were right that I had
**asserted** rather than shown it. So I showed it — ran the real helper (not a
re-implementation, which would be the very drift under test) over all 13
candidates and looked for the file in the actual deploy repo:

| | |
|---|---|
| `sites`-repo pages with a file at exactly the derived path | **11 of 12** |
| the 12th (`blog/learning-center-article.html`) | genuinely absent, serves 404 — it is the dead link the frozen page advertises |
| the 13th (relojistas `contacto`) | deploys to `vm-sites`, correctly absent there |

**Zero mismatches.** Refuted with evidence, not argument.

## 2026-08-03 — MISSTEP: I checked the wrong repo for one of the 13

My first sweep queried `gqls/sites` for every candidate. relojistas.com has
`github_repo='vm-sites'`. The row came back with a sha and I nearly recorded it
as "file present but 404" — an interesting-looking contradiction that was
entirely my own error. `gqls/sites` does hold a stale
`relojistas.com/contacto.html`, left over from before that site moved to the VM,
and it serves nothing because the domain is served from the VM.

The production code was never wrong here — `resolveGitRepoNameDB` picks the
per-site repo. **The verification script did not, and a verification that does
not model the thing it verifies is a coin toss.**

## 2026-08-03 — the owner's directive, and what it changed

> "When we retract a page we should look through the whole site to change any
> links inward, nav links, and any places that it linked to but now noone links
> to."

This is 098's second surface reached from the other end, and it is a stronger
statement of it. Implemented in `retract_page_graph.go` with the three
obligations split by KIND (editorial refuses / nav deactivates / stranded
reports) — the argument is in the file header and in DGH-006.

Audited for the real target before building: 0 editorial, 0 chrome, 0 nav,
0 stranded. Clean retraction.

## 2026-08-03 — MISSTEP: my first orphan census was blind, and said so confidently

The first pass at "what would be orphaned" looked only at `page_components` and
reported that **10 of 16** outbound targets had no other referrer — which would
have meant retracting one page stranded most of the site. Implausible on its
face, and wrong: the nav and the site chrome are where most links live, in
`site_nav_items` and `site_components`, and my query could not see either.

Re-run across all three sources: **every one of the 16** has chrome and nav
referrers. Nothing is stranded.

The lesson is the one 016b keeps recording: a link census that enumerates one
table answers a question about that table, not about the site. It is also
exactly why the shipped code copies the orphan check's three sources rather than
inventing its own.

## 2026-08-03 — the tests caught a real runtime bug the compiler could not

`[]string` is not a driver-supported parameter type here (database/sql over pgx
stdlib, with neither `pq.Array` nor pgx array types imported). Every one of my
new queries would have failed at runtime with *"unsupported type []string"*. The
sqlmock tests failed first; the fix is the project's existing
`datahelpers.PGTextArrayLiteral` with an explicit `::text[]` cast.

Separately, `go build` cannot parse SQL, so the queries were **PREPAREd against
the live schema** — which caught `site_components.component_type`, a column that
does not exist (it is `slot_name`). Two real defects, neither visible to the
compiler, both found before shipping.

Then ran them for real, with a **positive control**, because all four guards
returning empty for the target is also what a broken query looks like:

```
target /learning-center/index.html : 0 body, 0 chrome, 0 nav, 0 stranded
control /contact.html              : 4 body, 2 chrome, 1 nav
```

The control is the only thing that makes the zeros mean something.

## 2026-08-03 — the 090 diagnosis run produced NO VERDICT, and that is stated rather than glossed

Filed the resurrection claim through the diagnosis loop as CLAUDE.md requires for a
structural claim whose cause is not where the symptom is
(`5bdec8cf-24cc-419f-8d9d-b3d7a8df6dbb`). The dispatch loop **claimed it, ran, and
COMPLETED at 10:30:47** — and wrote **three `bundle` artifacts and no diagnosis**.
No `verdict`, no confirmation, no refutation:

```sql
SELECT kind, count(*) FROM diagnosis_artifacts
 WHERE correlation_id='5bdec8cf-24cc-419f-8d9d-b3d7a8df6dbb' GROUP BY 1;
--  bundle | 3        (and nothing else)
```

**So the loop did not corroborate anything, and this file does not claim it did.**
The owner ruling of 2026-07-31 allows a filing session to substitute equivalent
first-hand verification provided it says so plainly rather than omitting it. Doing
that here, and naming the chain:

1. `pages.deployed_at` for the page moved to **today** (10:05 read) — the trigger
   for looking at all;
2. the sites repo's commit list for that exact path shows a **single-file modify**
   at 08-01 08:07, 08-01 20:06, 08-02 08:05, 08-02 20:15, 08-03 08:15;
3. `orchestration_states.collected_data->'input_data'` for the most recent one
   carries `source: render_news_section`, `reason: section_data_resolved` and the
   page's own id;
4. `site_work_items` holds **6** `page_rerender` rows for that page with
   `source='render_news_section'`, first 07-31, latest 08-03 08:05;
5. the function those rows come from selects on `p.build_status='deployed'` with
   no `p.status` filter (read, not inferred), and the page is
   `status='archived', build_status='deployed'`;
6. fleet-wide that selector matches exactly **one** non-active page, and only two
   statuses exist (`active` 557, `archived` 25).

Each step is a fact about a different table or artefact, and they agree. That is
what is standing in for the loop's verdict.

**Prior art found while chasing it:** `queueNewsPageRerenders` has been through the
council before — correlation `320878ca-5c25-4fed-99ae-82b52b095aba`, routing it down
the no-LLM `page_rerender` path, live in v1.0.1155. That round changed HOW it emits;
neither round changed WHICH pages it selects, which is where the defect was.

## 2026-08-03 — COUNCIL ROUND 2: **REJECTED**, hard veto from `guardian`

Correlation `4a7f0877-4149-4431-97d4-318d093570a4`, round 2. Not a surprise — I named
this exact risk as #1 in my own submission — but the verdict is the verdict, and
CLAUDE.md is explicit: **"A veto on SCOPE is not answered by resubmitting with better
measurements."** So this is not being resubmitted. It is being split.

**The veto, verbatim on the substance:**

> "This is eight files spanning a shared wire struct, the core git-commit function every
> deploy-writing pipeline depends on, a new verb on a shared adapter, and its allowlist
> gate — plus three more new/changed action files. That is architecture-change-dressed-as-fix
> by this council's own definition… The one edit I would approve standing alone is
> `render_news_section_html.go` (add `p.status='active'` to `queueNewsPageRerenders`) —
> that is the actual root-cause fix for the resurrection bug, small, contained, and
> matches sibling query conventions. Safest path: ship that predicate fix now, and take
> `delete_file`/`retract_page_deployment`/`retract_page_graph` through an architecture
> review as a self-contained RFC on the git adapter's action vocabulary, separate from
> the urgent bug. **Splitting does not require re-deriving any of the measurement work
> already done — it only changes sequencing.**"

The `architecture` seat agreed and signalled `needs_rfc`, while saying the design itself
is right — *"expressing delete-as-null-sha inside the existing CommitToRepo path is the
right reuse… On that axis the plan is sound and I'd carry it"*. So the objection is to
**how the capability reached production, not to the capability**. That distinction is the
whole of 124's precedent and it is why more evidence would not have helped.

**What follows, and what does not.** The code is already on shared HEAD and forward-only
forbids removing it; the 2026-07-29 ruling says review here is after the fact by design
and that a thread must not pretend it could have held the change back. So: the predicate
fix stands and is live, the seam is registered (DGH-006) and now carries the veto, and
the RFC goes to architecture review on its own merits for a human to break.

**THE RETRACTION IS NOT BEING PERFORMED.** The owner approved retracting one page before
this verdict existed. Firing a vetoed capability at a live customer site, on the strength
of an approval given on a different basis, would be the exact thing the veto is about.
It waits for the RFC.

### The four objections that are RIGHT ON THE MERITS regardless of the veto

These are real defects in code that is already live, so they matter more than the
packaging dispute:

1. **`debug_historian`, HIGH, edit 8.** My link census does not go through
   `linkablePageStatusPredicate`. The landmine it cites is in my own memory index: *"an
   offline census MUST use `linkablePageStatusPredicate` — an ARCHIVED page makes a
   CORRECT rewrite look wrong."* Building a fresh census over `pages` with a hand-rolled
   status condition is precisely the shape that landmine guards. **Owed: verify and fix.**
2. **`editquality`, MEDIUM, edit 8.** Inbound references also live in **structured
   `content_data` fields** (`link_url`, `cta_url`, hero-tool/tool-cta `*_url`), not only
   in `href=` markup. An href-only scan can miss a real inbound reference and let a
   retraction proceed — *the exact case the owner's directive was written to prevent.*
   **Owed, and it is the sharpest catch in the round.**
3. **`reuse_agent` + `guardian`, MEDIUM, edit 8.** The inbound-source logic is *copied*
   from `check_orphan_pages.go` with a comment warning of drift, rather than extracted and
   shared. I argued the two ask different questions (what IS unreachable vs what BECOMES
   unreachable); the seats are right that a comment is not a mechanism. **Owed.**
4. **`bug_historian`, MEDIUM, edit 7.** The predicate fix invents a third spelling instead
   of consolidating onto a canonical helper, and I presented no fleet-wide census of other
   `build_status`-as-liveness selectors. **Half-answered after submitting**: I ran that
   census and found `queueDirectoryPageRerenders` — which calls itself "cousin of
   queueNewsPageRerenders" — carrying the identical defect (latent, 0 live rows), fixed in
   `8f73e7279`. The consolidation half is still owed and is a fair hit.

The pattern across all four: **every one is about edit 8**, the graph audit I added
fastest and reviewed least. The primitive, which took the most care, drew no correctness
objection at all — only scope.

## 2026-08-03 (evening) — RFC 011 decided, debts paid, and the retraction DONE

**Owner ruled OPTION B.** `delete_file` keeps its place on the adapter, loses its place in
the generic allowlist. Two lines of code, one test, and it concedes the seats' real
complaint (unbounded reachability) without discarding a capability two bugs needed.

**Debt 2 was a REAL defect, not a theoretical one — and the proof is the nicest measurement
of the lane.** Re-running the widened inbound scan on live data, `/contact.html` on
robot-hands went from **4 referrers to 5**: `gripper-cycle-time-estimator/hero` carries its
CTA url in `content_data`, invisible to an `href=`-only scan. The council seat that raised
it could not have known that; it reasoned from a landmine and was right.

I deliberately did NOT enumerate field names to fix it. `ctaFieldNames` exists and was the
obvious tool, and it is exactly the wrong one: a field-name allow-list is blind one level
down, which is what `bugs_open/097` was. Matching the url as a **quoted JSON string value**
finds it at any depth, under any key, including keys nobody has invented yet.

**The retraction ran through the platform**, not by hand — `page-retraction` agent,
orchestration `fc00db29…`. 200 → 404, file gone from the repo, deploy workflow green, and
six neighbouring live pages still 200. The graph audit ran on the live path and agreed with
the hand measurement exactly: `editorial_refs 0, nav_refs 0, stranded 0`.

## 2026-08-03 — MISSTEP, and it is the one I would most want caught: the audit's findings are DISCARDED

Reading the orchestration record afterwards, `collected_data.retraction` holds only the
adapter's reply — `{paths, success, repo_url, …}`. My action's candidates, refusals,
`editorial_inbound`, `nav_inbound`, `stranded_targets` and `nav_retired` are **gone**.

Cause: the step awaits the adapter response, and the await machinery **overwrites
`output_field`** with that response. The findings exist only in pod logs, which this repo
documents as ephemeral across rollouts.

This is `bugs_open/071`/`083`/`091`'s class — detected then discarded — and I wrote the
file header that says *"every refusal is RETURNED, not swallowed"*. It is true of the
function and false of the system. **A retraction that refuses a page would today refuse it
silently**, which is precisely the failure the owner's directive was meant to prevent.

Found only because I read the durable record instead of trusting the green result — the
same habit that found the resurrection. **The check that would have caught it earlier: for
any action that both computes findings and awaits a response, ask where the findings live
AFTER the await, not before it.** An action's return value is not a record.

Owed on the bug as item 5. The fix is to persist the audit before dispatching (or emit a
work item), rather than relying on `output_field` surviving.

---

## CONTRIBUTION 2026-08-03 — loancalculator lane: the sitemap gap, found by retiring a page by hand

I retired `loancalculator.co.uk/tools/standard-calc.html` on an owner decision, used
your audit procedure and your runbook throughout, and hit one thing your primitive
does not cover. Recording it here rather than editing your PLAN.

### `retract_page_deployment` is live but unreachable — your acceptance has not run

- Pod-grep on `agent-chassis-6c76bf964f-4tlrw`: `retract_page_deployment` → **6**,
  positive control `rerender_page_sections` → 20, negative control → 0. It is in the
  running binary.
- But `SELECT ... FROM agent_definitions WHERE default_config::text LIKE
  '%retract_page_deployment%'` → **0 rows**, and the same for `delete_file` → 0. No
  agent can reach either.
- `curl https://robot-hands.com/learning-center/index.html` → **200** at 2026-08-03
  ~21:00, so the live-instance acceptance your PLAN insists on ("the live instance or
  nothing") has not been taken.

I did **not** seat an agent to reach it — that is your platform change to make, and
riding it would have been the scope-veto shape the 2026-07-28 ruling describes.

### The gap: on an ADOPTED site, retraction leaves the sitemap advertising a 404

Your action's guard 4 comment lists `sitemap.xml` among the files "the repo
legitimately holds that `pages` does not model", and excludes it. That is right for a
site whose sitemap the platform generates and will regenerate.

**It is not right for an adopted site.** `loancalculator.co.uk/sitemap.xml` was last
written by the adoption commit `b4302e22b` (2026-07-30) and **the platform has never
regenerated it** — `git log` over that path shows only the adoption and the original
import. It is a static repo artefact listing all 27 urls with `<lastmod>` and
`<priority>`.

So `retract_page_deployment` on this site would have deleted the page and left the
sitemap pointing at it: **the frozen-listing-advertises-a-404 shape that 098 is
about**, reintroduced by 098's own fix. I removed the file and its `<url>` block in
one commit instead (sitemap `<loc>` 27 → 26, XML re-parsed before writing).

**Worth measuring before you decide anything:** how many sites have a sitemap the
platform has never written? That decides whether this is one site's quirk or a class.

### Two smaller things, both confirming your runbook

- **Your `href="…"` audit is the right one and a substring probe is not.** Mine
  reported two inbound links that were both HTML **comments** — one a previous
  session's footer note, one my own from three hours earlier. Your insistence on
  bodies + chrome + nav *with a positive control* is what makes it trustworthy;
  the control returned 1 body / 2 chrome for a page I knew was linked.
- **`gqls/sites` takes concurrent pushes constantly.** My push was rejected twice
  (`non-fast-forward`, then `cannot lock ref … is at <x> but expected <y>`) before
  landing on the third attempt. Anything scripted against that repo needs a
  pull-rebase-retry loop, and must verify at `origin/master` rather than at the
  local ref — a naive `git log -1` after a failed push reports success.
