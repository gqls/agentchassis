# 011 — Pre-cutover content/route-diff guard (prevent a cutover silently orphaning a funnel)

**Filed:** 2026-07-22, from the council review of `bugs_open/017` (owner scope decision: ship the
detector now, file this separately). **Class:** prevention for every class-B (static + backend on one
origin) site. **Status:** open, not started. **Relationship:** this is the **cause-side** of
`bugs_open/017`; the **symptom-side detector** (`backend_entry_orphaned`, flags a GET link to a
POST-only handler *after* deploy) is already built and its enablement seed (`188`) is ready.

## The problem (017's actual cause, left unpatched by the detector)

`bugs_open/017`: idea.uk's VM cutover gave `/` to the static site. The tool had served its own landing
page at `/`, and that page carried the **audience-check and report-request forms** (POST
`/audience-check`, POST `/request`). The cutover kept the forms' *targets* (nginx proxies all 16 tool
routes) but **silently discarded the forms themselves** — nothing checked what was ON the page being
replaced. The paid funnel was unreachable on a live earning site; every smoke test passed because
nothing errored.

The detector we built catches this **after** the cutover lands (a surviving GET link returning 405).
But bug_historian's council objection (raised both rounds) is right that the detector does not touch
the **cause**: a cutover step that overwrites `/` with no pre-flight check on what routes/forms it is
about to drop. Detection is not prevention — on the next class-B cutover, the funnel is still lost for
however long it takes discovery to run and a human to action the finding.

## Proposed guard

A **pre-cutover check**, run against the *old* serving model before the nginx config is swapped, that:

1. Enumerates what the **old `/`** actually serves that the **new `/`** (static build) will not:
   - the tool's **entry forms** — `<form method=POST action=...>` on the old landing page — and the
     backend routes they post to;
   - any route the old origin answered that the new origin will 404/405 for a browser.
2. Confirms the **new** static build contains an equivalent entry for each dropped funnel form (a
   form posting to the same backend route, same origin), OR flags the gap **loudly and blocking** —
   because this is a pre-flight on a money path, a block here is cheap (nothing public has changed
   yet; see RUNBOOK §4's port-8443 rehearsal, which is the natural place to run it).

This generalises the 017 handoff's open question — *"should discovery run automatically after a
deploy-target or origin change?"* — into its stronger form: **check before the change, not after.**

## Where it lives (why it's a feature, not a chassis code fix)

The cutover is a **manual nginx operation on the VM** (idea.uk RUNBOOK §4 / README §4a–4f), not a
chassis code path — so there is no single Go function to guard. The guard is therefore:
- a **RUNBOOK §4 checklist step** (added 2026-07-22 alongside this file — the immediate, zero-cost
  mitigation): during the port-8443 rehearsal, diff old-vs-new for every reserved path and confirm a
  live form exists for each POST-only funnel route before swapping; **and**
- optionally a **small script** that automates that diff (curl the old origin's `/` for `<form>`
  actions + reserved-path methods, curl the new build for equivalent forms, report drops) so it isn't
  a hope-the-operator-remembers step.

## How to verify (once built)

Reproduce 017's exact setup on the 8443 rehearsal: an old `/` with the two forms, a new static `/`
without them. The guard must **refuse the swap** (or emit a blocking finding) naming the dropped
funnel routes (`/audience-check`, `/request`), before anything public changes.

## Grounded in

- bug_historian, council corr `ed4851c9`, rounds 1 and 2: *"fixes the SYMPTOM … but does not touch
  the underlying mechanism that caused bugs_open/017 … the static-cutover overwrite of '/' silently
  discarding whatever page previously occupied that route, forms included … remains completely
  unpatched — this check will only catch its damage post-deploy."*
- `bugs_open/017` root cause + the RUNBOOK §3b note that recorded *"…and `/` (the landing page it
  loses)"* as a routing fact without anyone asking what was ON that page.
