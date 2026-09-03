# 220 — `unbuilt_internal_link` dispatch rebuilds the CONTAINING page, not the target; marks itself `complete`; re-detects next run — a convergence-free green loop

> **STATUS 2026-08-09 14:29Z — FIXED, LIVE, AND PROVEN END TO END. The convergence
> proof that was the last thing owed has landed.** Item `4151471c` (container
> `barrel-weight` → target `grip-styles`) saved sections to the **target**, rendered
> 27,623 bytes, deployed `/blog/grip-styles.html`, verified via disjunct **(a)**, and
> the page now serves **200** while its container still serves its own content.
> A second, independent convergence landed five minutes earlier on a different page
> type (`69818add`, target `brands-index`, `/brands/index.html` → 200).
> **The file stays in `bugs_open/` per the owner's 2026-08-06 ruling.** Full evidence:
> § "2026-08-09 14:29Z — CONVERGENCE PROVEN" at the bottom. Two findings in that
> section correct earlier claims in this lane — read them before reusing the
> acceptance query or citing candidate 4's demand signal.

**Filed 2026-08-08** by the `bugs_open/206` lane while answering the owner's question "would
the improvement loop have picked these problems up?". Diagnosed first-hand rather than via the
090 loop — declaring the substitute per the 2026-07-31 ruling: every link in the causal chain
below is quoted from the live system (the live `agent_definitions` config, the work item's own
`spec`/`page_id`/`result` columns, the `pages` table), not inferred; and a live re-demonstration
was fired the same day (correlation `867d6054-3f8f-4b11-9352-b29cecd9aaaa`, one-shot
improvement-loop over vetcomparison.uk — outcome: **REPRODUCED, see the dated section
at the bottom**).

## One-line mechanism

The phantom-links check deliberately files an `unbuilt_internal_link` work item against the
**target** page (the never-deployed one) via the item's `page_id` column — but
`build-dispatch-loop`'s `call_handler` step maps `page_name` from `current_item.spec.page_name`,
which for this item type is the page **containing** the link, and never reads `page_id`. The
handler then rebuilds/redeploys the container, returns success, and `mark_complete` runs
unconditionally. The wrong result looks exactly like the right one.

## Evidence (all read live 2026-08-08)

Item `8abc9f8d-c555-4378-bed4-dac8042025b2` (site vetcomparison.uk, minted by discovery
2026-08-02, one of three that day):

- `item_key` = `unbuilt_internal_link:page_component:index:info-card-grid:/directory/index.html`
- `page_id` column = `11326da0-df30-4378-a601-00c2812a891e` → **pages.name = 'directory-index'**
  (the target — correctly filed by the check).
- `spec.page_name` = `"index"` (the container), `spec.target_page_id` = the directory-index id,
  and `spec.fix` says verbatim: *"Do NOT rebuild the linking page — its href is correct."*
- `status` = `complete`, `attempt_count` = 0, `error` NULL, `claimed_by` = build-dispatch-loop,
  completed 2026-08-02 10:58:44Z.
- `result` (truncated): `deploy_result … "files": ["/index.html", "/tools/assets/latest-news.js"]
  … "success": true, "commit_message": "Rerender: "` — **the homepage was deployed, not the
  target.** Sibling item `b99dea13` (about → /directory/index.html) likewise deployed
  `/about.html`.
- `pages` row for directory-index today: `build_status='planned'`, `deployed_at IS NULL`,
  `sections = []` — the link is still a live 404.

The dispatching config, from the live `build-dispatch-loop` `agent_definitions` row
(`default_config->workflow->steps->process_item`, sub-workflow `call_handler.input_mapping`):

```json
"page_name?": "current_item.spec.page_name",
```

No key in the mapping reads `current_item.page_id`. The check's own routing intent
(`check_phantom_internal_links.go`, the `unbuilt_internal_link` arm: `actionablePageID =
f.TargetPageID` … "For an unbuilt link … the item is filed against the TARGET — rebuilding the
container would re-emit the same correct href and resolve nothing") is defeated by the generic
mapping. A doc comment / spec-prose instruction is not an enforcement mechanism — the dispatcher
never reads either.

## Why it reads green forever (three stacked absences)

1. **Wrong page**: the mapping above — the container rebuilds, the target stays unbuilt.
2. **No verification**: `call_handler.next_step = mark_complete`; `complete_work_item` runs on
   any non-error handler result. Nothing re-checks the actionable page's `deployed_at`.
3. **No convergence brake**: `complete` is terminal for the dedup index, so the next discovery
   pass re-mints the same `item_key` and repeats 1–2. Each cycle spends a full page build +
   deploy on the wrong page and reports success. (Distinct from `bugs_open/083`
   detected-findings-never-reach-a-handler: that gap is items never promoted because the sweep
   is off; this one fires precisely when the loop DOES run.)

Note even fixing (1) alone would not have built this page before 2026-08-08: bare
`page-build-handler` has no layout-filling step (`load_spec_sections` → `plan_sections` →
`mark_no_ready_sections` no-op for a plan-less page) — that capability gap was `bugs_open/206`,
fixed by `directory-build-handler`/`ensure_page_section_layout` (live v1.0.1264 + migrations
325/326). With 206's fix live, correcting THIS bug's routing would still dispatch
`page-build-handler`, not `directory-build-handler` — the check's `HandlerAgent` is hardcoded
and predates `load_work_item_actions.go`'s `availableBuilders` map gaining `entity-directory`.

## Fix candidates (ordered by what closes the door)

1. **Make the dispatcher honour the item's own `page_id`**: map e.g.
   `"page_id?": "current_item.page_id"` and have `page-build-handler`'s `load_page_record`
   prefer it over `page_name` — the column exists precisely to name the actionable page, and
   every check that files cross-page items benefits (not just this one). Config +
   possibly one Go read; blast radius = every `call_handler` dispatch, so measure which live
   item types carry a `page_id` differing from `spec.page_name`'s page before shipping.
2. **Check-side workaround**: set `spec.page_name` to the TARGET's name in the
   `unbuilt_internal_link` arm (the check already resolves the target row). One-line Go change,
   closes only this item type, leaves the mapping trap armed for the next cross-page check.
3. **Verification**: `mark_complete` (or a step before it) asserts the actionable page moved —
   for this item type, `deployed_at IS NOT NULL` on `page_id`. Turns silent wrong-fixes into
   loud failures fleet-wide; larger design question (what "moved" means per item type).
4. Route `unbuilt_internal_link` through `availableBuilders` rather than hardcoding
   `page-build-handler`, so an entity-directory target reaches `directory-build-handler`.

(2) is the cheap immediate stop; (1)+(3) are the class fixes; per this repo's own rule a
one-off deletion/patch is not a class fix — the pair is what makes the bad state
unrepresentable.

## How to verify a fix

Re-run the one-shot improvement loop over vetcomparison.uk (or any site with a live link to a
never-deployed page). Success = the minted item's dispatch builds the page named by the item's
`page_id` (target), `curl -sI` on the target URL returns 200 with fresh `last-modified`, and the
item completes with a result naming the target's file — not the container's.

## Related

- `bugs_open/206` — the capability gap this defect masked; its lane found this while measuring
  the improvement loop's actual behaviour. Fix live 2026-08-08.
- `bugs_open/083` (detected-findings slug) — the delivery gap upstream (sweep off, items park at
  `detected`); this bug is its complement when the loop runs.
- `bugs_closed/049` — created the `unbuilt_internal_link` class; closed on detection + nav-drop
  halves. The end-to-end remediation path was never exercised at close time (its own §
  "phantom_internal_link … fixed zero times, ever" table shows the type had never fired).
- Memory/lesson: "a `complete` work item is not a repaired artefact"; "a doc comment is not an
  enforcement mechanism".

## REPRODUCED 2026-08-08 15:19Z — same session, platform fully carrying 206's fix

The one-shot improvement-loop run (corr `867d6054`) re-minted 7 `unbuilt_internal_link`
items (dedup did not block — the 08-02 items are terminal, absence 3 above). The first
two dispatched before this lane intervened:

| item | finding (container → target) | what the dispatch deployed | target after |
|---|---|---|---|
| `3f066b90` | guide-independent-strategy → /directory/index.html | `/guides/independent-strategy/index.html` | still `deployed_at NULL` |
| `4ba1d4dd` | how-it-works → /entities/practice.html | `/how-it-works.html` | still `deployed_at NULL` |

Both `complete`, no error. This is with `directory-build-handler` live in
`agent_definitions` (migrations 325/326/336 applied) — confirming the check's
hardcoded `HandlerAgent: "page-build-handler"` + the dispatcher's `spec.page_name`
mapping bypass the new capability entirely, exactly as the "Note" paragraph above
predicts. The remaining 5 re-minted items were **cancelled by the 206 lane** (error
text on each points here) because the correct remediation for their targets — the
`needs_page` rows routed at `directory-build-handler` — was already queued; without
that intervention each would have spent a further multi-minute LLM rebuild on the
wrong page. A companion `empty_internal_href` item (`5cc5c24b`, tool page) dispatched
and no-op'd to `needs_human_review` ("no sections ready to build") — the parallel
failure shape for an item type whose container IS the right page but has no plan
sections.

## CONTRIB 2026-08-08 (116 lane) — fix candidate 1's blast radius, MEASURED: `unbuilt_internal_link` is the only type affected

Candidate 1 asks for a measurement before shipping — *"measure which live item types
carry a `page_id` differing from `spec.page_name`'s page"*. Run 2026-08-08 ~17:10Z by
the `bugs_open/116` lane, which needed the same number to decide whether the first
supervised improvement-loop run (owner ruling D3, 116:254) could safely be fired at
leopardessconsulting's 61 parked findings. Contributed rather than forked — this bug
is the 206 lane's.

**Method.** Left-join each work item to (a) the page its `page_id` names and (b) the
page `spec->>'page_name'` resolves to *within the same site*, then count agreement.
The dispatcher reads only (b); the check files against (a).

**Result over all live items (`status IN ('detected','triaged','approved')`), 24 types, 168 items: DISAGREE = 0 for every type.** The types split three ways, and only the
third is even capable of exhibiting this defect:

| shape | types | can it disagree? |
|---|---|---|
| `page_id IS NULL` (dispatcher's `page_name` is the only page signal) | 16 types incl. `needs_internal_links` 18, `undeployed_asset` 20, `needs_rerender` 11 | no — nothing to disagree with |
| `page_id` present and **equal** to the named page | `page_rerender` 66, `phantom_internal_link` 12, `empty_internal_href` 4, `empty_section` 3, `literal_markdown` 1, `improve_tool` 1 | no — filed against the container, which IS the actionable page |
| `page_id` present and **different** | **`unbuilt_internal_link` only** | **yes — this bug** |

**The zero is disconfirmable, and was disconfirmed on purpose before being trusted.**
Positive control: re-run unfiltered by status, the same query flags **13/13**
`unbuilt_internal_link` rows as DISAGREE — including both rows this file names
(`8abc9f8d` index→directory-index, `b99dea13` about→directory-index) plus 11 more
across the fleet (`10b4fc72`, `49f0e189`, `6d1cb353`, `5686a320`, `bdf76041`,
`4ba1d4dd`, `2ef4eb51`, `3f066b90`, `240e4020`, `46f732a4`, `b20fa738`). All 13 are
`complete` or `cancelled`, which is why they fall outside the live filter — **the live
zero is a real zero, not the empty-set kind.**

**What this means for the fix.** Candidate 1 (`"page_id?": "current_item.page_id"` +
`load_page_record` preferring it) has **no live behaviour change outside
`unbuilt_internal_link`** — every other live type either supplies no `page_id` or
supplies one that already agrees. That is the narrow blast radius candidate 1 hoped
for, now measured rather than assumed. It does **not** speak to types that are
currently absent from the queue; a future cross-page check would newly depend on the
mapping, which is the argument for pairing it with candidate 3 rather than shipping
alone. Re-run before shipping — the census moves daily.

**Caveat this lane could not close:** the measurement covers items that already exist.
It says nothing about what a *discovery pass* mints mid-run, and this file's own
REPRODUCED section shows a run re-minting 7. The mint condition is a live link to a
never-deployed page.

## TAKEN 2026-08-08 ~19:00Z — fix built (candidates 1+3), candidate 4 deferred on record

Taken by a fresh lane after the 206 lane's closing message explicitly left it ("the
follow-up is the separate bug 220 dispatcher fix") and a transcript sweep found no
session holding the fix-site symbols. Working docs:
`docs024_key_docs_latest/bugfix_220_unbuilt_link_dispatch/`. Register: **WII-012**.

**A correction to candidate 1 as written above:** "map `page_id?` and have
`load_page_record` prefer it over `page_name`" is INSUFFICIENT as stated —
`load_page_record_action.go` resolves **page_name first** (its documented priority),
and its empty-name branch re-fills the name from `input_data.spec.page_name` plus three
more fallback paths. For this item type the container's name always resolves, so a
forwarded id would always lose. The shipped shape (per the 2026-08-02 RFC_010 §2
ruling — new authority on a shared seam is an opt-in FIELD):

1. `build-dispatch-loop` `call_handler.input_mapping` += `"page_id?": "current_item.page_id"`
   (mig **340**; precedent: `site-work-orchestrator` already maps exactly this shape).
2. `load_page_record` gains optional `authoritative_page_id`: resolves-and-parses →
   lookup BY ID, name ignored; malformed → loud error; absent → prior behaviour
   byte-for-byte. Only `page-build-handler`'s step opts in (mig 340);
   `tool-recreation-handler`'s stays name-first.
3. `VerifyUnbuiltInternalLinkResolved` registered (fail-closed per RFC_017): resolved
   when the container no longer renders the href OR the target page has shipped, judged
   by `datahelpers.NeverDeployedPagePredicate` — the detector's own predicate. Claim-
   timeout lockstep: declared list in `sql_for_agents/220` + live column via mig **341**
   (the `TestRegisteredVerifiersMatchClaimTimeoutExclusion` obligation, 331's template).

**Candidate 2 not taken** (narrow, and would make spec.page_name lie about which page
contains the link). **Candidate 4 deferred**: `availableBuilders` lives in package
`actions`, which imports `discovery_checks` — the check cannot reach it without a leaf-
package refactor of the 206 lane's day-old machinery; with the verifier live a
directory-target item now fails LOUDLY into the attempt machinery (→ `failed`, two-
strike parks re-detections) instead of lying green, and the 206 lane's
`needs_directory` path is the designed remediation for those targets. Whoever picks
candidate 4 up: expect exactly that `failed` population as your demand signal.

**Verification state:** unit tests green against a clean `git archive HEAD` overlay
(the load-bearing test asserts at the QUERY ARGUMENTS that a resolving container name
loses to the forwarded target id). Go rides the next fleet roll; migs 340/341 are
inert against the pre-roll binary (measured: no live handler dispatched by this loop
reads `input_data.page_id`; the old binary's InputSpec does not declare
`authoritative_page_id`, and ExtractActionInputs only reads declared fields).
Behavioural acceptance after the roll is this file's § How to verify, unchanged —
plus: completion must carry `result._verification.status='verified'`.

## COUNCIL TRAIL 2026-08-08 (corr def4441c)

Round 1: REVISE, gating objection from bug_historian — and its HIGH was a genuine
catch: the authoritative-id path's no-row case returned the {found:false} soft miss
(the silent complete_error shape this very fix targets). Fixed with a fatal error +
a no-second-query test, commit `e55cbfa64`; the LIKE-concatenation in the verifier's
presence check likewise replaced with position() (raw hrefs carry `_`, a SQL
wildcard). Round 2 resubmitted on the same correlation, everything else answered by
measurement (one producer at source; the fragments arm defers unbuilt targets by
design).

**For the dead_fragment_link lane:** `VerifyDeadFragmentLinkResolved`
(check_phantom_internal_links_fragments.go:235-248) carries the same
LIKE-concatenation shape the council flagged here — `COALESCE(rendered_html,'')
LIKE '%' || href || '%'` with the raw href. Same over-match class, same fix shape
(position()). Flagged rather than edited under you.

**Round 2 APPROVED 2026-08-08 ~22:06 UTC** (3 advisory objections, none high). The
two-gates advisory is answered in the lane NOTES: the claim query's second gate
already selects and exposes `wi.page_id` (`load_work_item_actions.go:646`, `:776`).
Remaining owed on this bug: the post-roll proof (pod-grep + behavioural acceptance,
lane RUNBOOK) — the fix is inert until the next fleet roll. Fixed-and-live is the
close bar; per the owner's 08-06 ruling the file stays in `bugs_open/` either way.

## 2026-08-09 — POST-ROLL ACCEPTANCE: routing + verifier PROVEN LIVE, and the run exposed (and closed) the fix's missing third leg

Fresh chassis roll carrying all three commits, proven at both replicas
(`authoritative_page_id` 3, the r2-only string 1, verifier strings 5, invented
negative 0). One-shot improvement loop fired at dartsonline.com (corr `110acf5a`,
PUBLISH_OK receipt); discovery minted 6 `unbuilt_internal_link` items (5 →
grip-styles, 1 → brands-index) — the watcher line itself showed the split (spec
page_name = containers, page_id column = targets).

**What the first dispatch (item `338deb27`, container beginners → grip-styles) proved:**
- **Routing FIXED**: `deploy_result.rendered_page.page_id` = the TARGET
  (grip-styles), filename `blog/grip-styles.html` — under the old mapping this
  deployed the container's file. The deploy honestly SKIPPED ("page has no
  component rows yet") instead of shipping the wrong page.
- **Verifier LIVE**: completion carries `_verification.status='verified'`, disjunct
  (b) (link no longer rendered on the container) — honest on the stored substrate.
- **THE MISSING THIRD LEG**: `sections_saved.page_name` = **beginners, the
  CONTAINER**. `save_sections.page_name_field` was the ONE step config still
  reading `input_data.spec.page_name` while load_spec_sections /
  load_existing_content / call_content_writer / deploy_page all follow
  `page_record`. So the writer wrote the TARGET's sections (grip-styles' plan) and
  save_sections saved them ONTO THE CONTAINER: **beginners' `content_data` was
  replaced with grip-styles' copy at 10:00:56Z** ("Your Grip Decides More Than
  Your Barrel Does" in beginners' hero). Sibling `a8327624` (brand-comparison →
  brands-index) was stopped only by the content-regression floor (2,520 chars of
  index copy vs 11,914 existing) and sits `failed` — the loud outcome, correctly.

**Containment + fix, same morning:**
- Beginners' two queued rerenders (`47ba8f2c`, `3c10ab6c`) CANCELLED — either
  would have published the contaminated copy (a rerender renders from
  content_data; the served page is still the correct old render).
- The 4 remaining triaged unbuilt items CANCELLED (each would have contaminated
  its own container), error text points here. Discovery re-mints them; under the
  full fix they converge.
- **Mig `342_page_build_handler_save_sections_follows_page_record.sql` APPLIED +
  recorded** (⚠ filename collides with the thunder lane's independent
  `342_thunder_orphan_scan.sql` — number ambiguous forever, resolve by slug;
  WRONG_CALLS 2026-08-09): `save_sections.page_name_field` → `page_record.name`,
  giving the saga ONE page identity end-to-end. Blast radius: page_record.name ==
  spec.page_name for every consistent dispatch (116-lane census); config-only,
  live immediately, effective with the rolled binary.
- **Repair minted**: `needs_content_page:beginners:repair-338deb27` (item
  `3cb732b1`, priority 30, consistent identity) — rewrites beginners from its own
  plan and deploys. Until it lands, beginners' stored content is wrong and its
  rerenders must stay held.

**Still owed (the convergence proof):** after the repair lands and discovery
re-mints, one unbuilt item must run end-to-end: writer writes the TARGET's
sections, save_sections saves them to the TARGET (`sections_saved.page_name` =
the target), deploy renders it, `pages.deployed_at` set, item completes
`_verification.status='verified'` via disjunct (a), `curl -sI` the target → 200.

## 2026-08-09 midday — the RESIDUE: the fix stops new false greens, it does not repair the old ones. One is still live on a production site.

The verifier makes an item that leaves its target unbuilt fail loudly **from now
on**. It is not retroactive, and nothing in this lane had asked what the items that
completed *before* it existed are actually sitting on. Asked now, fleet-wide, by
identity rather than by count (a count would have said "6 complete" and hidden the
whole distinction):

| item | site | completed | target | target deployed? | href still rendered? | reading today |
|---|---|---|---|---|---|---|
| `49f0e189` | mortgagecalculator.co.uk | 08-05 13:12 | `scorecard-simulator` (landing) | **NEVER** | **yes** | **genuine residue — would be `Resolved:false` if re-verified now** |
| `4ba1d4dd` | vetcomparison.uk | 08-08 15:19 | `practice` (entity-page) | NEVER | **no** | legitimately resolved — link removed, which the item's own fix text accepts (disjunct b) |
| `6d1cb353` | fundamentallyai.com | 08-05 13:13 | `platform-log-index` | 08-08 18:16 | — | false green at the time; target shipped 3 days later by other means |
| `5686a320` | mortgagecalculator.co.uk | 08-05 13:17 | `tool-affordability` | 08-09 12:03 | — | false green at the time; target shipped 4 days later by other means |
| `3f066b90` | vetcomparison.uk | 08-08 15:19 | `directory-index` | 08-08 17:02 | — | false green at the time; target shipped ~2h later by other means |
| `338deb27` | dartsonline.com | 08-09 10:01 | `grip-styles` | NEVER | (was no, is yes again) | this morning's, `_verification=verified` via disjunct (b) — see the addendum above |

**`49f0e189` is confirmed at the served artefact, not inferred from the DB:**
`curl https://mortgagecalculator.co.uk/guides/first-time-buyer/index.html` → **200**,
21,034 bytes, and it carries `href="/scorecard-simulator.html"` exactly once;
`curl …/scorecard-simulator.html` → **404**. So a live page has been linking to a
404 since at least 05 August while the work item created to fix it has read
`complete`. That is this bug's damage, unmitigated, on a production site today.
[MEASURED 2026-08-09]

⚠ **The URL is not derivable from the page name and this bites twice in one lane.**
`guide-first-time-buyer` serves at `/guides/first-time-buyer/index.html`; my first
curl guessed `/guide-first-time-buyer.html`, got 404, and for a moment looked like
evidence the container was gone. `beginners` serves at `/blog/beginners.html`, not
`/beginners`. **Read `pages.url`; never build the URL from `pages.name`** — the
wrong guess returns a 404 that is indistinguishable from the defect you are hunting.

**Three of the six were false greens that got rescued by something else entirely**,
days later — which is worth stating because it is how this bug stayed invisible:
the target eventually appears, so a spot-check of any individual site tends to look
fine, and only the join of "when did the item complete" against "when did the target
deploy" shows the item was green while the page did not exist.

**What to do about the residue — nothing bespoke.** A `complete` row is terminal for
the dedup index, which *frees* the slot, so the next discovery pass over
mortgagecalculator.co.uk re-mints the finding and it now converges honestly (or
fails loudly) under the verifier. This is the same mechanism being proven on
dartsonline today. No repair item, no backfill, no migration: **the residue is one
discovery pass away from being self-correcting, and hand-cleaning it would prove
less than letting the machinery clear it.** Whoever runs the next improvement loop
on that site gets the second, independent end-to-end proof for free.

## 2026-08-09 14:29Z — CONVERGENCE PROVEN: the target is built, deployed and served, and the container is untouched

The last thing this bug owed was an end-to-end convergence: not "the routing was
right and the deploy honestly skipped" (which the morning run gave us via disjunct
**b**), but a dispatch that **built the target, shipped it, and made the 404 a 200**.
It landed twice in six minutes, on two different `page_type`s.

**Run**: corr `576f0ab9-5a17-4449-9bbc-ee1983576433`, fired 13:10Z at dartsonline.com,
re-minted 10 `unbuilt_internal_link` items at 13:12:45Z. Dispatch reached priority 45
at 14:21Z.

### Proof 1 — `4151471c`, target `grip-styles` (page_type `blog-post`)

The acceptance family's own case: `grip-styles` was `planned`, never deployed, 0
components, and `/blog/grip-styles.html` was a live 404 all morning.

| leg | value read at 14:29Z |
|---|---|
| `spec.page_name` (container) | `barrel-weight` |
| `sections_saved.page_name` | **`grip-styles`** — the TARGET. Mig 342's leg |
| `sections_saved.page_id` | `769e3b72…` |
| `rendered_page.page_id` | `769e3b72…` — same page |
| `rendered_page.html` length | **27,623** — a real render |
| `deploy_result…data.success` | **`true`** |
| `deploy_result…data.file_path` | **`/blog/grip-styles.html`** |
| `_verification.status` / `.detail` | `verified` / disjunct **(a)**: *"target page 769e3b72… has shipped; href \"/blog/grip-styles.html\" now resolves"* |
| `pages` (grip-styles) | `build_status=deployed`, `deployed_at=2026-08-09 14:28:45Z`, 3 components |
| `curl https://dartsonline.com/blog/grip-styles.html` | **200** |
| container `barrel-weight` served title | *"Barrel Weight Guide — …"* — its own. **Not contaminated** |

The last row is the bug's actual damage signature and it is absent: the container
kept its own content while the target was built.

### Proof 2 — `69818add`, target `brands-index` (page_type `section-index`), 14:24Z

Container `about` → target `brands-index`. `sections_saved.page_name` =
`brands-index`, `rendered_page.page_id` = `92b8bb46…` = `sections_saved.page_id`,
html 18,495 bytes, `success: true`, `file_path` `/brands/index.html`, verified via
disjunct **(a)**, `pages.deployed_at` 14:24:22Z with 2 components,
`curl /brands/index.html` → **200**, and the container `about` still serves *"About
Darts Online | Spec-First Darts Guides"* with zero `All Brands` content.

> **CORRECTION — this refutes this lane's own stated expectation.** The 08-09 handoff
> and NOTES both recorded that the four `section-index`-targeting items were
> **expected to fail LOUDLY** and were "deferred candidate 4's demand signal". One of
> them **converged cleanly instead**. So `section-index` targets are NOT unbuildable
> by the current handler, and **candidate 4's demand signal is weaker than this lane
> claimed — on this evidence it may be absent.** Do not cite "section-index targets
> fail" as a reason to pick up candidate 4 without re-measuring; the prediction was
> made from the morning run's honest skip on a *different* page type and was never
> tested on a section-index target until now. The remaining three
> (`6e1b562b`, `0469f44f` → brands-index; `b4184d0f` → shop-index) were still
> dispatching when this was written — their outcomes are the real demand signal and
> should be read before anyone re-opens candidate 4.
>
> **RESOLVED at 15:14Z — "may be absent" is now "is absent".** All four converged via
> disjunct (a), across two distinct section-index pages both built from zero
> components. See § "FINAL LEDGER 15:14Z" at the foot of this file for the ruling.

### ⚠ The acceptance query in the RUNBOOK was WRONG, and its control could not catch it

The documented deploy leg was
`result->'response'->'deploy_result'->'rendered_page'->>'page_id'`. The real shape
nests **another `response`** in between:
`response → deploy_result → response → rendered_page`. The documented path therefore
returns **empty on every row, converged or not** — it would have reported the proof
above as a failure on leg 2.

**Why the lane's control did not catch it:** the control (`338deb27`) expected that
column to be *empty*, and a wrong path and absent data render identically. The
control agreed with a broken instrument. **A control only tests a column if that
column is expected to be NON-EMPTY in the control case.** Corrected query, plus a
second control that reads `deploy_ok = true` on `69818add`, is in
`RUNBOOK_unbuilt_link_dispatch.md` § "THE acceptance assertion".

Note the substantive claim about the morning run was *right*: at the corrected depth
`338deb27` has `rendered_page.page_id` = `769e3b72` (grip-styles — routing correct)
with html length **0** and no `success` — a genuine honest skip. Only the path used
to express it was wrong.

### One unexplained, benign observation

Container `about` has `deployed_at` = 14:23:52Z, inside item `69818add`'s dispatch
window — yet that item's own deploy payload lists exactly one file
(`files_count: 1`, `["/brands/index.html"]`), no other work item on the site touched
`about` between 14:15 and 14:30, and the served `/about.html` is uncontaminated.
So the container was re-deployed by something outside this item, with its own
content. **Recorded rather than explained** — it is not this bug's signature (no
content damage) but a reader who greps `about`'s timestamps will trip over it.

### FINAL LEDGER 15:14Z — all ten converged, zero failures, and candidate 4's demand signal is ABSENT

The run completed at 15:14Z. Every one of the ten items:

| status | count | via disjunct (a) | via disjunct (b) |
|---|---|---|---|
| `complete` | **10** | **10** | **0** |

`sections_saved.page_name` = the target on **10/10**; `deploy_ok` = `true` on **10/10**;
all three target pages `deployed`. **Nothing failed.**

**Read this as THREE independent proofs, not ten.** The ten items share only three
distinct targets, and only the *first* item against each target built a page that had
never been built:

| target | page_type | first item | built at | file |
|---|---|---|---|---|
| `brands-index` | section-index | `69818add` | 14:24Z | `/brands/index.html` |
| `grip-styles` | blog-post | `4151471c` | 14:28Z | `/blog/grip-styles.html` |
| `shop-index` | section-index | `b4184d0f` | 14:52Z | `/shop/index.html` |

The other seven rebuilt and redeployed a target that was **already shipped** by the
time they ran, so they demonstrate the weaker claim (correct routing + honest
re-verification), not build-from-nothing. Three independent convergences across two
page types is still decisive for this bug — but do not quote "10/10" as ten
independent proofs.

**Worth someone's attention, not this bug's:** grip-styles was rebuilt and redeployed
**six times** between 14:28 and 15:10, once per item, because six separate containers
each linked to it. Each rebuild is a full LLM section-generation + deploy. That is
correct-but-wasteful — the second and later items could have short-circuited on
"target already shipped" before generating anything. Not filed as a bug here because
it is a *cost* property of the dispatch loop rather than a defect in 220's fix, and
filing it against this bug would bury it. Whoever picks it up: the cheap check is
`pages.build_status` at claim time.

> **CANDIDATE 4'S DEMAND SIGNAL IS ABSENT.** All four `section-index`-targeting items
> converged via disjunct (a) — on **two distinct** section-index pages
> (`brands-index`, `shop-index`), both built from zero components and both now
> serving 200 with their containers uncontaminated. The lane's recorded expectation
> that these would "fail LOUDLY" is refuted on every instance that could have shown
> it. **On this evidence candidate 4 (route unbuilt targets by `page_type` via
> `availableBuilders`) has no demand signal at all** and should not be picked up on
> the justification written in § "TAKEN 2026-08-08". If it is wanted, it needs a new
> rationale and fresh measurement — ideally a target `page_type` that is genuinely
> unhandled, which this run did not produce.

## NOTE from the `bugs_open/450` lane (2026-09-03) — your candidate 4's demand signal arrived, and was answered somewhere else

Your FINAL LEDGER ruled candidate 4 (*"route `unbuilt_internal_link` through `availableBuilders`
rather than hardcoding `page-build-handler`"*) had **no demand signal**, and asked for *"a target
`page_type` that is genuinely unhandled"* before anyone revisited it. **`bugs_open/450` is that
case**: `page_type='tool'` sits in `builderForPageType`'s `unavailableBuilders`, and the
hardcoded route sent 20 link-repair items at seven tool pages that had no tool, which the generic
builder duly filled with prose and deployed.

**The demand signal is real, and candidate 4 is still not what was built** — recorded here so the
ledger is not reopened on a false premise:

- Your deferral reason **still holds, verified 2026-09-03**: `actions` imports `discovery_checks`,
  nothing under `discovery_checks/` imports `actions`, so `availableBuilders` remains unreachable
  from the check without a leaf-package refactor.
- More decisively, routing the CHECK better would have fixed **one producer of five**. 450's
  `page_component_history` census names four others writing the same pages
  (`empty_section` 3 pages/67 writes, `page_rerender` 3/20, `needs_page` 3/14,
  `needs_content_page` 2/8, 14 days to 2026-09-02). So the guard went where all five converge —
  the `writeWorkItem` policy door and the composition guards — via a derived predicate
  ("a `page_type='tool'` page with no live tool component refuses generic builds"), commit
  `587666be8`, register **PBP-053**, inert until the next roll.
- **Your candidate 4 is therefore neither done nor needed for 450**, and is left exactly as your
  ledger left it. If someone later wants per-`page_type` routing for its own sake, the import
  direction is still the first obstacle and this note is not a mandate to remove it.

**Your candidate "one target, one build" (N links → N items → N rebuilds of one page) is
UNTOUCHED and still yours.** 450 measured the same shape independently — 26 writes across 6 pages
on seotools, one rebuild per link — and deliberately did not fix it. After the roll those
duplicate items will terminate `wont_fix` against a refusing page rather than rebuilding it N
times, which **hides the waste on tool pages without removing it anywhere else**. Worth knowing
if you ever re-measure the churn: tool pages will stop appearing in it.
