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
   this was measurable rather than inferred.
2. **[DONE] Get the domain into the deploy repo**, byte-exact, committed.
3. **[AWAITING OWNER] Push**, and prove the live site is unchanged afterwards.
   Outward-facing and hard to reverse, so it is not done unasked.
4. Pre-flight the adoption (ported-page component, chassis path, `input_mapping`
   allow-list carrying `fidelity`, open work items on the domain).
5. Decide the orphan/broken-link question — under `high` it changes what is
   adopted, exactly as under `locked` (handoff §3).
6. Submit `--fidelity high`, watch the recreate path, and hold the rerender queue.
7. Positioning: narrow the identity to mortgage-only authority, mutually coherent
   with the two sibling sites (handoff §6).

## Open questions for the owner

- **Push authorisation** (phase 3).
- **The three broken inbound links** (handoff §3/§4): under `high` the crawl still
  drives the page list, so 2 orphan guides are still silently skipped. Fixing the
  links before the crawl changes what gets adopted; after, it does not.
- `images/mortgagecalculatormono.xcf` is a GIMP source file, publicly fetchable
  (200 live). Kept for now so the first sync is provably content-neutral; removing
  it is a separate, deliberate commit.
