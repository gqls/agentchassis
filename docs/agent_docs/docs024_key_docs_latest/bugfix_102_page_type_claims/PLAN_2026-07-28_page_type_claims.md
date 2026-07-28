# PLAN 2026-07-28 — the claims layer reads page type (`bugs_open/102`)

**Session:** bugsearch 7. **Bug:** `bugs_open/102_HANDOFF_2026-07-27_claims_layer_is_page_type_blind_guides_read_as_business_claims.md`.
**Ownership checked before starting:** `scripts/who-owns.py 102` names no owning
workstream; no commit subject about 102 since it was filed (07-27); no
`site_work_items` row open against it; no uncommitted tree changes under
`datahelpers/claims*` or `discovery_checks/`. The parent lane
(`fabricated_stats_043/`) lists it under "still owed" as a **blocker it is not
working**, and its last commit was 2026-07-27 13:56.

---

## The finding, as filed

The claims layer's only precision control on prose is `businessClaimContextRe`
(`datahelpers/claims.go:497`) — a **lexical** gate asking whether the words near a
number sound like business. Nothing consults **what kind of page the number is on**.
`page_type` is in the `pages` table, it is already loaded into the build gate's
collected data by `load_page_record`, and the audit's own query already joins
`pages` — the signal exists at both call sites and is read by neither.

## What I measured before designing (2026-07-28)

The bug's own survey covered the four sites that owe registers, and concluded
**"live exposure today: nil"**. That is true of `webdesign.co.uk`. It is **not
true fleet-wide**, and the real exposure is worse than the file says.

Method: `cmd/claimscan` — the same shared scan engine as the gate and the audit —
run against **each opted-in site's own live register** over its live
`page_components.rendered_html`, with `pages.page_type` carried through the export.
Nine sites have a current `evidence_base` row.

| page_type | unregistered-number findings |
|---|---|
| blog-post | 46 |
| content | 38 |
| report | 14 |
| adoption-tracker | 8 |
| tool | 7 |
| game | 4 |
| protocol-tracker | 3 |
| section-index | 2 |
| news-index | 1 |
| guide | 1 |
| **total** | **124** |

Then I read every finding on an editorial page type. **47 of 47 are false
positives**, and they are not marginal:

- `gamesdesign.co.uk` blog-posts: `0.40 × 0.0833 = 3.33%`, "a 78.3% chance of
  getting the item within 60 kills", "10,000 active players farming that item" —
  worked probability examples, 40 findings on one site.
- `ai-agent-orchestration.com` blog-posts: "an endpoint that returns **200**",
  "if **30** of them fail simultaneously", "they surface at **2**am".
- `robot-hands.com` index pages: "[Insights] Market report projects cobot tending
  cells to hold **38%** share in 2026" — a third-party market figure quoted in a
  news listing, not a claim about the business.
- `gamesdesign.co.uk` tool/game pages: "Set to **0** to disable", "Connected
  Clients: **0**", "(0.1% – 25%)" — interactive widget help text.

Against that, the findings on `content` pages are the real ones the layer exists
for: leopardess's `90,790`, ai-agent-orchestration's `170`.

**So the class has a measured precision of 0% on editorial page types and is
doing its job on business ones.** That is the whole argument for the fix.

### And the exposure is blocking, not cosmetic

`unregistered_number` is **`error`** severity in the build gate
(`validate_page_content.go:970`), and `valid = blockerCount == 0 && errorCount == 0`.
An opted-in site's blog-post that carries a worked example **cannot be rebuilt
today** — it fails validation and routes to `mark_needs_review`. `gamesdesign.co.uk`
has 40 such findings across four posts. This is live on the build path (the
post-deploy audit is the half that does not run — `bugs_open/083`).

## Design

Candidate **1** from the bug file, with one correction the bug does not make.

**The correction: only the NUMBER scan is page-type-sensitive. The banned-claims
scan stays on, on every page type.** The bug says "a guide's *prose* scan is
skipped". Taken literally that would regress the motivating case of the whole
check: `check_unverified_claims.go:13` records that its first live run found
"70+ agents across eight functional departments" **on a guide**. A banned claim is
a known falsehood wherever it is written; it is matched against a human-authored
pattern, not a heuristic, so it has no false-positive problem to protect against.
The heuristic number extraction is the only thing that misreads teaching content.

Three parts, all in the shared layer so neither call site can forget:

1. **`datahelpers.ClaimSurface`** — a surface descriptor carrying `PageType`, and
   `ProseNumbersAreClaims()`, which is the one place the policy lives.
2. **`ScanUnregisteredNumbers(blocks, surface)`** — signature change, not a new
   variant. A second "safe" entry point is exactly the shape of `bugs_open/093`
   (one guarded call site, the other unchecked); a signature change makes the
   compiler visit all three callers. **Zero value = unknown = scan**, so site
   chrome and any caller with no page in hand behave exactly as today: a scanner
   that goes quiet and one that is broken look identical, so unknown must be noisy.
3. **The callers pass what they already have.** The gate resolves the page type
   from collected data (`page_record.page_type` — populated by `load_page_record`,
   which `page-build-handler` runs before `validate_content`); the audit adds
   `p.page_type` to a `JOIN pages` it already does; `cmd/claimscan` takes it as an
   optional 4th TSV column so the fleet measurement above is reproducible.

**Editorial set** (prose numbers not scanned), each earned by a measured finding:
`guide`, `blog-post`, `blog-index`, `news-index`, `section-index`, `tool`, `game`.

**Not in the set, deliberately:** `report`. Its 14 findings on robot-hands are
false positives of a *different* class — model numbers inside product names
("Schunk EGP **40**-N-S-B — manufacturer specification") tripping on `verified` in
the context regex. Excluding `report` would fix them by coincidence rather than by
mechanism, and a report page's figures genuinely can be business claims. Filed as
a separate finding instead.

**Rejected: candidate 2** (per-site opt-out config). The bug calls it "operators
must remember X in a configuration costume", and it drifts the moment a new page
type appears. A structural default needs nobody to remember anything, and an
unknown type stays noisy rather than silent.

**Deferred: candidate 3** (tutorial-framing lexical exclusion). Worth doing, and
the only thing that helps a guide-shaped section on a business page — but it is a
second mechanism, and the 124 precedent is that a change which accretes a second
mechanism draws a scope veto. It goes in the bug file as the follow-on with the
one live instance I measured for it (finetuning's "5 to 50 employee businesses"
on a `content` hero).

## How it will be verified

Both directions, per the bug file's own § "How to verify a fix" — a count going
down proves nothing on its own:

- **Positive control:** re-run the identical fleet survey with the fixed binary;
  the 47 editorial findings go, the 77 business-surface ones stay, exactly.
- **Negative control:** a deliberately unregistered figure planted in the prose of
  a **non-editorial** page on the same site is still raised.
- **Regression control:** a banned claim planted in a **guide** is still raised —
  the case that motivated the check.
- Unit tests pin all three, plus the zero-value (unknown ⇒ scan) behaviour.
