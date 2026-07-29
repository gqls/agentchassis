# 147 — robot-hands.com claims its spec data is "independently verified"; nothing verifies it

Filed 2026-07-29 by session "bugsearch 6" (the `bugs_closed/104` lane), found by the
fleet-wide claims dry run while arming the external-verification pattern. Not a platform
defect — a **live content defect on one site**, with a platform consequence attached.

## Symptom

Two deployed components on robot-hands.com assert that the catalogue's specification data has
been independently verified:

| page | slot | sentence (verbatim, from `page_components.rendered_html`) |
|---|---|---|
| `gripper-catalog` | `info-card-grid` | "Grip force, stroke, cycle time, and IP rating pulled from manufacturer datasheets **and independently verified**." |
| `how-it-works` | `generic-text-block` | "…with specifications drawn directly from manufacturer datasheets and, where available, **independently verified** test data." |

Nothing in the platform performs independent verification of manufacturer specifications.
The data is scraped from datasheets; "independent" would require a second, unrelated party,
and there is none.

## Why this is a defect and not a wording preference

**The same site says the opposite, as its stated policy**, on two other components — which is
how the inconsistency was found:

- `index`/`features`: "Where manufacturer data has **not** been independently verified, that
  is stated explicitly. Rigour over reassurance."
- `gripper-detail`: "When a figure **cannot** be independently verified, it is marked as
  unverified rather than carried forward as fact."

So the site promises to label unverified data, and then on two other pages claims the data is
verified. One of the two is false. Given no verification mechanism exists, it is the pair in
the table above.

**Its own evidence register does not support the claim either** (`site_specs`
`aspect='evidence_base'`, read 2026-07-29): 5 facts, all of `kind: count`
(`rh-grippers` 10, `rh-manufacturers` 6, `rh-actuation` 6, `rh-parameters` 4,
`rh-spec-figures` 59), **no fact registering any verification process, and
`banned_claims: []`**. There is nothing to cite.

## Evidence

- Dry run 2026-07-29 over the **full live enforcement surface** — 919 components / 14 sites,
  all `deployed` — with `cmd/claimscan` running the same engine as the deploy gate:
  **2 findings, both above; 4 further matches suppressed by the negation guard** (the four
  honest sentences, including the two quoted above from this same site).
- Reproduce:
  ```bash
  go build -o /tmp/claimscan ./cmd/claimscan
  /tmp/claimscan -evidence rh.eb.json -show-suppressed -components rh.tsv   # 2 BANNED, 2 negated
  ```
  Export commands and gotchas: `bugfix_104_fleetwide_claim_patterns/RUNBOOK_…` §2.
- Regression fixtures now pin both sentences as must-block:
  `datahelpers/claims_global_test.go` `TestGlobalBlocksTheLiveExternalVerificationOverclaims`.

## The platform consequence — read this before rebuilding those pages

The fleet-wide banned-claim set now carries the `(fully|independently|externally|properly)
(verified|audited|fact.?checked)` pattern at severity **blocker**, so
**`validate_page_content` will refuse a rebuild of these two components until the copy
changes.** Nothing breaks on the live site — the deployed HTML stays served, and the gate only
bites on a rebuild — but a page-rebuild of either will fail with the reason
"external-verification claim: asserts our content was checked by someone outside this system."

That is the gate doing its job. It is recorded here so the failure is expected rather than
mysterious, which is the failure mode `bugs_open/081` is about.

## Fix candidates, ranked by what closes the door

1. **Change the copy to describe what actually happens.** The site already has the honest
   version of this sentence elsewhere, and it has real machinery to point at: each entry
   records a data source and a last-verified date. "Cross-checked against the manufacturer
   datasheet and dated" is true, is stronger than a vague appeal to independence, and matches
   the site's "rigour over reassurance" line. Closes the door: the false claim is gone.
2. **Register a verification fact** — only if some real second-source check exists that nobody
   has written down. Then the claim is supportable and the pattern can be given a per-site
   exception. Do NOT do this to silence the gate; a fact with no mechanism behind it is the
   `vetcomparison` failure in a new costume.
3. Do nothing and accept that those two components cannot be rebuilt. Not recommended, but
   honest, and it is what happens by default.

## How to verify

After a copy fix, rerender the two components and re-run the dry run for the site: **0 BANNED,
2 negated** (the two honest sentences must still be suppressed — if they start firing, the
negation guard has broken, not the copy). Then confirm on the wire that the sentences changed,
not merely in `content_data` — `bugs_open/079`'s lesson is that a rerender regenerates from
source, so the stored and served copies can disagree.

## Ownership

The `robot_hands` lane (`docs024_key_docs_latest/robot_hands/`) owns this site; its last doc is
`SUMMARY_2026-07-24_robot_hands_residuals_closed.md`, so it is dormant rather than active — I
have not touched the copy, because rewriting another lane's site voice is theirs to do. A
pointer is in that lane's NOTES.
