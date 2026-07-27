# 102 — the claims layer is page_type-blind: a guide's worked example is indistinguishable from a business claim

**Filed 2026-07-27** by the fabricated-stats lane (`bugs_closed/043`, resolve **by slug**),
while doing owed item (b) — building evidence registers for the remaining publishing sites.

**Status:** OPEN, not started. Cause is structural and fully known — nothing to diagnose.
**Live exposure today: nil**, because the affected sites have no `evidence_base` row, so the
scans are off. **This bug is what stops that being fixed**, which is why it is worth filing
rather than absorbing.

---

## The finding

The claims layer's only precision control on prose is `businessClaimContextRe`
(`datahelpers/claims.go`) — a **lexical** gate asking whether the words around a number
sound like business. Nothing in the layer consults **what kind of page the number is on**:
`check_unverified_claims.go` joins `pages` but reads only `p.name`, and neither
`ScanUnregisteredNumbers` nor `ScanStatClaims` receives a page type at all.

On a marketing site that is fine — nearly every page asserts something about the business.
On a site whose main body is **teaching content**, it is wrong in the expensive direction:
an explainer's worked example is lexically identical to a sales claim.

## The measurement that found it

Survey run 2026-07-27 with the real scanner (`ExtractAssertionText` +
`ScanUnregisteredNumbers` against an **empty** register, so every business-shaped number
surfaces) over the four sites that owe registers:

| site | components | number claims | verdict |
|---|---|---|---|
| gaswholesalers.com | 102 | 0 | needs no register |
| dartsonline.com | 17 | 1 | a 30-day returns window — a policy term |
| finetuning.uk | 139 | 5 | 4 are audience descriptors / a privacy age limit; **1 is real** |
| **webdesign.co.uk** | **101** | **15** | **all 15 are false positives** |

Every one of webdesign.co.uk's 15 is a worked example inside teaching content, and all of
them sit on `page_type = 'guide'`:

```
10,000  | Once your site hits 10,000 monthly visitors and is generating revenue…
100     | …bringing the server to a halt at just 100 concurrent users.
502     | …random 502 Bad Gateway errors become inevitable.
3.5     | …let's say the global average for all products on your site is 3.5 stars.
0 / 1   | …getting a site from 0 to 1 is no longer the hard part.
100,000 | …getting that site from 1 to 100,000 users.
5 / 1   | * Try rating Sci-Fi '5' and Rom-Com '1' to match User A.
```
```sql
SELECT DISTINCT p.name, p.page_type FROM page_components pc
JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE s.domain='webdesign.co.uk' AND (pc.rendered_html LIKE '%concurrent users%' OR …);
-- learn-ai-builders-content-first        | guide
-- learn-algorithms-bayesian-theory       | guide
-- learn-algorithms-p-values-explained    | guide
-- learn-algorithms-recommendation-math   | guide
-- learn-communities-graph-vs-relational  | guide
-- learn-operations-scaling               | guide
```

**Six pages, one page_type, 100% of the false positives.** The signal that separates them
from a real claim already exists in the schema and is simply not read.

## Why this matters beyond noise

1. **It blocks owner-requested work.** Item (b) of the fabricated-stats lane is "evidence
   registers for all the publishing sites". For webdesign.co.uk that currently cannot be
   done: switching the register on would raise 15 human-review items that are all correct
   copy, and the queue (`bugs_open/033`) has no working surface to dismiss them through.
   So the largest site in the estate stays unprotected — not because we lack evidence, but
   because the checker cannot be trusted on it.
2. **False positives are the expensive failure here.** A missed claim is one wrong number.
   A checker that cries wolf on a teaching site trains a human to dismiss its findings,
   which silently disables it on the same site's real claims. That is worse than off.
3. **It is the mirror of the `043` design decision, and the boundary is now visible.**
   `ScanStatClaims` deliberately does NOT apply `businessClaimContextRe`, because a figure
   in a `stat*_value` field is a claim *by construction* — structural position replacing a
   lexical gate. That reasoning is sound and it cuts both ways: **page type is structural
   position too**, one level up, and the layer uses it nowhere.

## Fix candidates (unranked)

1. **Pass page context into the scan and let the caller scope it.** `scanDeployedClaims`
   already selects from `pages`; add `p.page_type` and hand it to the scan. A `guide` /
   `blog-post` page's *prose* scan is either skipped or graded a rung lower, while its
   **stat fields keep being scanned** — a stat card on a guide is still a published claim.
   Makes the bad state unrepresentable at the point of decision rather than asking
   operators to remember which sites are educational.
2. **A per-site or per-page-type opt-out in the `evidence_base` row** (e.g.
   `"scan_prose_on_page_types": ["landing","about","content"]`). Cheaper, no Go change —
   but it is "operators must remember X" in a configuration costume, and it will drift the
   moment a new page type appears.
3. **Tighten `businessClaimContextRe` to exclude tutorial framing** ("let's say", "imagine",
   "for example", "Try rating"). Narrowest, needs no plumbing, and it would have caught most
   of the seven above — but it is a lexical patch on a lexical gate's failure, so the next
   phrasing escapes it again.

**Prefer (1).** The structural signal exists and is free; (3) is worth doing anyway as
defence in depth, and it is the only one that helps a guide-shaped section on a
non-guide page.

## How to verify a fix

Do **not** grade it on a count going down — a scan that got quieter may just be off. Seed
webdesign.co.uk a real register, run the audit, and assert **both** directions:

- the six `learn-*` guide pages raise **zero** prose findings; and
- a deliberately false figure planted in a **non-guide** page's copy on the same site is
  still raised (the negative control — without it, "quiet" and "broken" look identical).

## Related

- `bugs_closed/043` — the parent lane; its `claims_stats.go` header is where the
  structural-position-replaces-lexical-gate argument is made for stat fields.
- `bugs_open/093` — the second call site for the stat audit, which is what made this survey
  cheap enough to run fleet-wide.
- `bugs_open/033` — the review queue with no working surface, which is what turns 15 false
  positives from an annoyance into a reason not to enable the check at all.

---

## Triage 2026-07-27, post-roll (v1.0.1174) — unchanged, and the blocking claim is confirmed live

Verification sweep, not a fix. Nothing to diagnose here and nothing has moved.

**The structural signal is still unread.** `page_type` / `PageType` appears **nowhere** in
`datahelpers/claims.go` or `discovery_checks/check_unverified_claims.go` — grep returns no
matches. Confirmed against the running image: no Go commits after `e96d42226`, which is in
`v1.0.1174`.

**The premise "live exposure today is nil" is confirmed, and so is the blockage.**
webdesign.co.uk still carries **zero** `banned_claims` and has no `evidence_base` row:

```sql
SELECT s.domain, jsonb_array_length(COALESCE(ss.data->'banned_claims','[]'::jsonb))
FROM sites s LEFT JOIN site_specs ss ON ss.site_id=s.id AND ss.aspect='evidence_base' AND ss.is_current
WHERE s.domain='webdesign.co.uk';   -- webdesign.co.uk | 0
```

So the estate's largest site is unprotected for exactly the reason this file gives, and the
owed item (b) of the fabricated-stats lane still cannot be done for it. That makes 102 a
**precondition of `bugs_open/104`'s coverage work**, not a parallel concern: arming
webdesign.co.uk before this lands would fire 15 correct-copy findings into
`needs_human_review`, which has no surface (`bugs_open/033`) — the exact "trains a human to
dismiss its findings" failure § "Why this matters" warns about.

One sizing note for candidate 1: the prose scan runs at the **build gate**, which is live, so
the fix has real effect immediately on roll — unlike the post-deploy half, which does not run
at all (`bugs_open/083`). The § "How to verify a fix" negative control (a deliberately false
figure on a non-guide page must still be raised) is therefore exercisable on the build path
without waiting for any sweep.
