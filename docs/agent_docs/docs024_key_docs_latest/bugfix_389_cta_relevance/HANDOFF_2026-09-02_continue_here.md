# HANDOFF — 2026-09-02. **START HERE.** `bugs_open/391` — **the CTA defect is FIXED.** One owner call left: close now, or hold for the retraction

**Supersedes `HANDOFF_2026-08-31_continue_here.md`.** All figures re-measured 2026-09-02 ~15:3xZ.

> Chassis **`v1.0.1352`** (image tag; the binary-grep probe is useless here — its all-zeros control
> matches). No behaviour-changing commit has landed in the resolver, rerender or retraction since
> this lane's proofs.

---

## 1. WHAT IS DONE — the bug this lane was filed for is fixed

| | |
|---|---|
| **Label-locked CTA fields** | **0 fleet-wide** (was 20 of 80) |
| **Contact-intent buttons** | **all routed to `/contact.html`** per owner decisions 08-31 and 09-02 — 21 fields across 15 pages, each published and verified at the served bytes |
| **`nav_order` fossil** | demoted 1 → 900 on all three sites, holding |
| **The three tool pages** | **all `archived`** (0 active), still serving 200 — archiving freezes without unpublishing |
| **Misdirected buttons on live pages** | **gone** — *"Book a Technical Discovery Call"* now reaches the contact page |
| Work items | all terminal; 0 open |

**The repoints are durable by design, not by luck.** `applyCTARecompute`'s **KEEP #1** *is*
`bugs_open/248`'s fix: a stored `/contact.html` satisfies `storedCTADestinationIsAuthored`
(`contact` ∈ `areasExcludedFromCTA`, the page is live, no mint stamp names it), so every later
recompute **keeps and rewrites** it. Proven in production on the canary — a full `cta_links_stale`
recompute ran over it and left it alone.

## 2. ⛔ THE ONE DECISION LEFT

**7 references remain, and they are NOT this lane's defect.** They are entries in `items[]` arrays of
related-tool cards (`tool-cta`, `tool-list`) — **stale listing snapshots**, owned by `bugs_open/384`.

They matter here only because **the retraction (removing the three tool files) REFUSES while any
editorial inbound link remains**, and those 7 count. So:

- **Option A — close `391` now.** Its own defect is fixed and verified. The three pages stay
  `archived`-but-served, which is **harmless**: archived pages keep serving, no link is dead, and the
  renderer already drops CTAs pointing at them. Retraction becomes 384's downstream consequence.
- **Option B — hold `391` open** until 384 clears the listings, then retract and sweep.

**Recommendation: A.** Holding a fixed bug open for another bug's residue is how a lane stops being
readable, and the residue is fully documented with a CONTRIB already filed.

## 3. THE RESIDUE, characterised (so nobody re-derives it)

`[MEASURED 2026-09-02]` 7 refs / 7 pages:

| policy | count | route |
|---|---|---|
| `generic` | 4 (aiao: 3 `tool-cta` + `/tools.html` `tool-list`) | ordinary path reaches them |
| `owned` | 3 (finetuning ×1, leopardess ×2) | `OWNED_PAGE_GUARD` — excluded **by design** (384) |

⚠ **A `cta_links_stale` rerender COMPLETES on these and clears nothing** — there is no `*_url` CTA
field; the link is an array entry. *A green completion against the wrong mechanism looks exactly like
one against the right mechanism; the only tell is that the census does not move.*

⚠ **And a hand-fired `section_data_resolved` did NOT re-resolve it either** — completed, rewrote the
component, stale array survived, differing from a fresh resolve by exactly one swap. Every documented
precondition held. **Full evidence and its consequence for 384's candidate 2 is in
`CONTRIB_2026-09-02_to_384_section_data_resolved_did_not_reresolve.md`.** Do not spend the afternoon
re-deriving this: it is measured, and it is their call.

## 4. IF YOU TAKE OPTION B — the retraction order

1. Clear the 7 (384's business, or by hand on the 4 `generic` ones).
2. Deactivate the one `site_nav_items` row (aiao, `/tools/password-entropy.html`, `active`).
3. **Refresh the aiao footer** — `nav-link-fixer`, then propagate in **assemble mode**
   (`page-rerender`, **no** `spec.reason`). ⚠ Not the agent whose name says navigation: it deletes
   every child-path link (NAV-013). Worked script:
   `docs/leopardessconsulting/scripts/reconcile_footer_nav.sh`.
4. **Retract** via the `page-retraction` agent (`site_id_field`, `page_ids_field`).
   ⚠ **`retract_page_deployment` DELETES when `dry_run` is absent** — its sibling
   `retract_asset_files` treats absence as a dry run. Opposite defaults; `LANDMINES.md`.
5. Final sweep at the served bytes, footer included.

## 5. SPIN OUT, do not hold this lane for it

The **root cause is still live**: the ranking picks CTA destinations by `nav_order` with no notion of
relevance, so the next site inherits the bug. Owner decision 3 approved a lever (explicit opt-out +
a detector for the fossil-`nav_order` shape), read **at the ranking, not the loaders**, also binding
`LoadCTALabelUniverse`, and engaging **RFC_022 with the consumers enumerated**. That is
architecture-scope with its own review round — **its own lane, filed when 391 closes.**

## 6. Session-start checklist
`git log --oneline -10` · re-read this from disk · `scripts/who-owns.py 391` · chassis **image tag**
· re-run the §1 and §3 censuses (**counts move on their own** — this lane's population drained 41 → 25
unattended over five days) · then §2.
