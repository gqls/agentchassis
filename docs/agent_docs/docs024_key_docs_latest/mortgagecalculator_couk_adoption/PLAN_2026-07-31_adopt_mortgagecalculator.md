# PLAN — adopt mortgagecalculator.co.uk

**Started 2026-07-31.** Cold start is `HANDOFF_2026-07-31_adopt_mortgagecalculator.md`
in this directory, written by the `loanandmortgagecalculator_couk` lane.

## The goal, in the owner's words

Adopt mortgagecalculator.co.uk into the framework **as an editable site**, with the
`sites` git repo holding a record of the current site "if we run into trouble" —
and **without bringing the site down**.

## Decisions

### D1 (owner, 2026-07-31) — `--fidelity high`, not `locked`

The owner was shown, with the code in front of them, that `high` is **not** a
gentler setting than `locked` and chose `high` anyway. Recording what that means so
nobody later reads "high" as "mostly preserved":

`platform/orchestration/actions/apply_adoption_plan_action.go:426` is a strict
binary — `if fidelity := adoptionFidelity(...); fidelity == fidelityLocked`. There
is no gradient in the code. `082_submit_domain_unified.sh:64-66` says it plainly:

> `high | medium | low | new` **STILL NOT WIRED**, exactly as before: recorded in
> `input_data` … but **modulating nothing**. Behaviour remains IMPLICIT high.

So `--fidelity high` ≡ omitting fidelity ≡ the **recreate** path, which means:

- `datahelpers.CanonicalisePage` **synthesises a new URL** for every page and
  discards the crawled one (`/repayment.html` → `/tools/repayment/index.html`);
- pages land `build_status='planned'`, `content_data.mode='recreate'`;
- `page-build-handler` / `tool-recreation-handler` **regenerate each page with an
  LLM**, including the 13 hand-built mortgage calculators that scored 13/13
  `RESPONDS` on 2026-07-31.

The alternative offered and declined was `locked`, which preserves URLs and bytes
and still permits editing, because `rebuild_policy='owned'` is a **per-page** flag
with real readers (`rerender_single_page_action.go:310`,
`save_page_sections_action.go:149-156`, `reconcile_site_plan_action.go:233`) — any
single page can be opened to the pipeline later. **The owner's call stands; this
paragraph exists so the option is not re-litigated, and so the consequence is
attributable to a decision rather than to an accident.**

**Consequence for the safety work: it goes UP, not down.** Under `locked` the
platform ships stored bytes unchanged. Under `high` it will write new pages at new
URLs and commit them into `gqls/sites`. The outage chain in §1 of the handoff is
fidelity-independent, and the recreate path exercises it harder.

### D2 (this lane) — the deploy repo is populated from the BUCKET, not the CDN and
not the local tree

See NOTES for the measurement. Short form: the local tree lacks `robots.txt`, and
the CDN's `robots.txt` is not the origin file.

## Phasing

1. **[DONE] Enumerate the bucket and reconcile.** Blocker in the handoff
   (`[UNMEASURED]` — no B2 credentials). Credentials now exist on this machine, so
   this was measurable rather than inferred. 29 real files + 5 `.bzEmpty`.
2. **[DONE] Get the domain into the deploy repo**, byte-exact, committed
   (`65d06ef4e`, 29 files taken from the bucket itself).
3. **[DONE — LIVE AND VERIFIED] Push.** Run `30668633897` resolved
   `Changed domains: mortgagecalculator.co.uk`, synced, and **all 29 live files are
   sha256-identical before and after**. Bucket is now 29 files / 0 `.bzEmpty`, 1:1
   with the repo. **Handoff §1's outage hazard is closed.**
4. **[DONE] Orphan/broken-link fixes** (`825a36994`, run `30672002187`, live).
   Reachability from `index.html` **20/23 → 22/23**, only `404.html` unreachable,
   which is correct. Verified by recomputing the link graph transitively, not by
   inspection. Exactly one live file changed; the other 28 byte-identical.
5. **[DONE] Pre-flight.** Our domain: **0 `sites` rows, 0 orchestration runs, 0 work
   items**. The 41 `page_rerender` rows a substring query first attributed to us are
   the sibling's — see NOTES. One `ported-page` component, as required.
   `fidelity` plumbing (migration 274 / the `input_mapping` allow-list) is **moot on
   this path**: a dropped `fidelity` is indistinguishable from `high`, since
   `082` itself defaults adopt-mode to `high` (`FIDELITY="${FIDELITY:-high}"`).
6. **[NEXT] Submit `--fidelity high`** and watch the recreate path.
7. Positioning: narrow the identity to mortgage-only authority, mutually coherent
   with the two sibling sites (handoff §6).

## What `high` will actually produce — measured from the URL synthesiser, not assumed

`datahelpers.CanonicalisePage` (`page_canonical.go:106-215`) is deterministic, so the
resulting URL shape is knowable now rather than after the fact:

| crawled page | role the classifier assigns | URL the rebuild gets |
|---|---|---|
| `/index.html` | `index` (or slug `home`/`index`) | **`/index.html` — the SAME URL** |
| `/repayment.html` | `tool` | `/tools/repayment/index.html` |
| `/guides/first-time-buyer.html` | `guide` | `/guides/first-time-buyer/index.html` |

So the outcome is **not** uniform, and the homepage is the exception:

1. **The homepage is overwritten in place.** `role=index` returns `/index.html`
   unconditionally, so the LLM-generated homepage lands on top of the live one.
2. **Every other page appears at a NEW URL**, and the old file is not deleted by
   anything — nothing in the platform tracks it. So `/repayment.html` (hand-built,
   working) and `/tools/repayment/index.html` (LLM-built) both go on serving.
3. **The old pages become orphaned** the moment the new homepage ships, because the
   new homepage links the new URLs. They stay live and reachable by direct link and
   by anything Google has already indexed.

**Directory-index behaviour, measured fleet-wide today** — B2 serves no directory
index, so the nested shape only works at its full path:

```
200  https://webdesign.co.uk/tools/image-optimizer/index.html
404  https://webdesign.co.uk/tools/image-optimizer/
200  https://gamesdesign.co.uk/guides/index.html
404  https://gamesdesign.co.uk/guides/
```

87 pages fleet-wide already use `tools/<slug>/index.html` and 17 use
`guides/<slug>/index.html`, so this is the platform's normal shape and not a new
risk — but a bare `/tools/repayment/` will 404, and any link written without the
trailing `index.html` is dead.

**The consequence to put to the owner:** every indexed URL on this site changes, the
old ones keep serving stale duplicates, and nothing in the platform reconciles the
two. If the old URLs should keep working, that is a redirect job to plan
deliberately — and there is no cross-site or intra-site duplicate-content machinery
here to notice (handoff §6 established the same absence across sites).

## Open questions for the owner

- **Push authorisation** (phase 3).
- **The three broken inbound links** (handoff §3/§4): under `high` the crawl still
  drives the page list, so 2 orphan guides are still silently skipped. Fixing the
  links before the crawl changes what gets adopted; after, it does not.
- `images/mortgagecalculatormono.xcf` is a GIMP source file, publicly fetchable
  (200 live). Kept for now so the first sync is provably content-neutral; removing
  it is a separate, deliberate commit.

---

## Update 2026-08-03 — phases 6 and 7 done; the site is now HELD at the build line

### D3 — positioning goes in `target_audience`, not `divergence_rule` (this lane)

The brief (handoff §6) said to mirror the sibling's `divergence_rule`. **That key has
no consumer**: no Go, no SQL, no prompt template reads it, and the classifier's
output schema has no such field, so the next classifier run would drop it. The live
sibling's divergence is in fact carried in **`identity.target_audience`**, which is
read (`v3_site_actions.go:1233` step-contract scalar, `component_library.go:116`,
`contentcreator/agent.go:53`).

So the positioning was written into `identity.target_audience` and
`content_direction.things_to_avoid` (+ the `formatted` blob, which is the only field
the content writer reads). Full commands and the verification query: RUNBOOK §8.

The three sites are now mutually coherent, each stating what it is *not*:

| site | scope | sends elsewhere |
|---|---|---|
| `mortgagecalculator.co.uk` | secured lending on residential property — mortgages, remortgage, bridging, equity release | unsecured → `loancalculator.co.uk`; spanning → `loanandmortgagecalculator.co.uk` |
| `loancalculator.co.uk` | personal loans and car finance | — (no explicit boundary yet) |
| `loanandmortgagecalculator.co.uk` | decisions where unsecured debt and a mortgage interact | single-subject → the two single-subject sites |

**Known weakness, stated rather than hidden:** none of this is protected.
`site_specs.pinned` does **not** stop `write_site_spec` (it is not read, and is
dropped from the replacement row — see LANDMINES). So **a future classifier run will
overwrite `target_audience`.** The durable hold is `sites.locked_at`; the positioning
must be re-checked after any agent run that writes `identity`.

### D4 — hold the SITE, not the items (this lane)

24 items were being held individually as `deferred`. That does not scale once a
**chain** is running: `needs_domain_research` → `needs_vertical_research` →
`needs_strategy` → `needs_briefing` → `needs_site_plan`, each handler creating the
next item, against a 120-second dispatch tick. You cannot win that race by hand.

`sites.locked_at` is the platform's own switch for this and was unused fleet-wide
(0 of 37 sites). Set 2026-08-02 23:21 UTC. It gates dispatch, not completion.

**Where the line is:** everything up to `needs_briefing` writes only specs.
**`build-site-planner` is where pages start being planned** — that is the first step
that can lead to the live site changing.

### Phase status

6. **[DONE] Submit `--fidelity high`.** Ran 07-31; the classifier half then failed on
   a platform defect (`bugs_open/183`) and was re-run successfully 2026-08-02 23:16
   after the cap fix. 23 pages, 4 classifier specs written.
7. **[DONE] Positioning** — D3 above.
8. **[NEXT — needs an owner decision] Release the build queue.** 24 items held
   (`deferred`) plus the site lock. Proposal: release ONE non-homepage page first,
   inspect the artefact, then decide. The homepage is the only page whose URL does
   not change, so it is the only one that overwrites live content on rerender.
9. **[OPEN — owner decision] Redirects.** 22 of 23 URLs move; the old files keep
   serving and nothing in the platform reconciles them.
