# 071 — the content gate detects every broken internal link, then throws the finding away

**Filed:** 2026-07-25, from fundamentallyai.com, while answering the owner's
question "what can we run to find the multiple errors on this site… including
many bad links".
**Severity:** high. **21 of 22 internal links on a live site are broken**, and the
platform detected every one of them at build time, by name, and deployed anyway.
**Status:** OPEN — diagnosed with evidence, not fixed.
**Class:** fail-open whose written justification names a component that is not
running (same family as `063`, and the dormant-machinery class generally).

## Symptom

Every internal link on fundamentallyai.com 404s except one. Census of rendered
`page_components` on deployed pages, 2026-07-25:

| verdict | links | unique targets |
|---|---|---|
| BROKEN — target page does not exist (invented by the writer) | 11 | 9 |
| BROKEN — target exists but href omits `.html` on a `.html` fleet | 10 | 5 |
| OK | 1 | 1 |

Live-probed, cache-busted: `/multi-agent-review-council` → **404**,
`/multi-agent-review-council.html` → **200**. `/contact` → 404, `/contact.html`
→ 200. Invented targets (`/platform-capability`, `/rapid-delivery`,
`/self-correction`, `/review-council`, `/verification`, `/how-we-work`,
`/our-platform`, `/production-integration`,
`/self-correction-verification-system`) 404 in every form.

## The gate saw all of it

`validate_page_content` check 4 (`validateInternalLinks`,
`platform/orchestration/actions/validate_page_content.go:564`) ran on the build
and flagged **eight of eight** phantom links on the self-correction page,
correctly, with the exact hrefs. Recovered from that build's
`collected_data.validation_result.issues` (orchestration `07d05813`):

```json
[{"sev":"warning","type":"phantom_link","value":"/contact"},
 {"sev":"warning","type":"phantom_link","value":"/multi-agent-review-council"},
 {"sev":"warning","type":"phantom_link","value":"/platform-capability"},
 {"sev":"warning","type":"phantom_link","value":"/production-integration"},
 {"sev":"warning","type":"phantom_link","value":"/rapid-delivery"},
 {"sev":"warning","type":"phantom_link","value":"/review-council"},
 {"sev":"warning","type":"phantom_link","value":"/self-correction"},
 {"sev":"warning","type":"phantom_link","value":"/self-correction-verification-system"}]
```

This is not a detection failure. **The detector is correct and complete.**

## Root cause: three gaps compounding

**1. Warnings cannot affect validity, by construction** (`:252`):

```go
valid := blockerCount == 0 && errorCount == 0   // warnings are not counted
```

**2. The policy comment states the justification, and it is not true** (`:587`):

> *"Policy: a missing internal target is loud but NON-BLOCKING — **the
> improvement loop resolves it**; a missing link is not a deploy stopper."*

The improvement loop is **not running** (owner, 2026-07-24 — the reason
`features_open/019` was deferred). So the repairer the fail-open defers to does
not exist at runtime. "Loud" is also generous: see gap 3.

**3. On the success path the findings are never persisted.**
`writeValidationFailureLog` is called only `if !valid` (`:307`), so a page whose
only issues are warnings writes **nothing** to `agent_error_log`. The per-issue
`logger.Warn` loop is likewise inside `if !valid`. What survives:

- the returned action output → `collected_data`, which `database-cleanup` prunes
  at **~24h** (the retention trap already logged in WRONG_CALLS);
- one pod-log line carrying `warnings=8` — **the count only, not the hrefs**.

So 24 hours after a build, the fact that the platform knew about eight specific
broken links is unrecoverable. No work item is ever created.

**4. (site-specific, compounding)** The post-deploy audit that *would* create
work items — `check_phantom_internal_links` — has never run on this site:
fundamentallyai.com has **zero** discovery-check work items, which is exactly
`features_open/019`. So both independent paths to a durable record are off.

## Why this is the owner-visible defect

The owner reported "many bad links including on the home page" and asked what to
automate. The honest answer: **nothing needs building to detect this — it is
already detected, on every build, accurately.** What is missing is that the
finding is neither enforced nor kept. A platform that markets a verification
council shipped a page whose every link 404s, having named all eight.

## Fix candidates

**Candidate 1 (smallest, highest value): persist warnings.** Call the structured
log write on the success path too (or unconditionally), so `agent_error_log`
carries the issue list regardless of validity. Cheap, no behaviour change, makes
the existing detection durable and queryable. Do this even if nothing else lands.

**Candidate 2: emit a work item per phantom link at gate time**, deduped on
(site, page, href) — the same routing `058` used for `lock_blocked_change`. This
makes the gate's findings actionable without depending on the discovery sweep or
the improvement loop being on.

**Candidate 3: split the severity by fix class.** The two classes are not alike:
- *missing `.html` on a target that exists* is mechanically fixable with zero
  judgement and no content loss — a strong candidate for **error** (deploy
  stopper) or for silent auto-correction at render time;
- *invented target* needs a decision (repoint or remove the link), so warning +
  work item is right.
  Today both are one undifferentiated warning.

**Candidate 4: stop the writer inventing link targets.** The prompt should be
given the site's real page list and told to link only within it. This is the
upstream cause of 11 of the 21 broken links, and it recurs on every new page.

**Candidate 5 (do not skip): fix the comment.** A policy comment that justifies a
fail-open by naming a downstream repairer must say what happens when that
repairer is off. This one has been read by at least two threads as "handled".

## Verification (induce the failing branch)

Build a page containing `href="/definitely-not-a-page"` on a site with an
evidence base and no other issues. Pre-fix: page deploys, `agent_error_log` has
nothing, no work item. Post-fix (candidate 1): the issue list is in
`agent_error_log` with `phantom_link` and the href. Do **not** verify by
checking that the gate logs a warning — it already does, and that is the bug.

## Relates to

- `features_open/019` — sweep enrolment. This bug is why 019 matters more than
  "deferred, the loop isn't running" suggests: enrolment is the *other* path to
  a durable record, and it is off for this site.
- `bugs_open/049` — 312 live broken links across 7 sites, incl. "32
  extension-less targets on a `.html` fleet". **This is the same
  extension-less class**, caught at a different stage. 049 measures the damage
  post-deploy; 071 explains why the pre-deploy gate lets it through.
- `bugs_open/023` / CTA-link-integrity — the gating sweep tests non-emptiness,
  not resolvability.
- `bugs_closed/063` — fail-open on missing config, same shape: the protective
  branch is skipped exactly where protection is needed.

## Related defect, same blind spot: nothing validates the FRAGMENT

The gate's phantom-link check (and the post-deploy audit it shares definitions
with) resolves the **path** and ignores everything after `#`. So a "jump to this
section" link is never checked at all. Measured 2026-07-25 across all deployed
pages, fleet-wide:

| site | anchored links | fragment resolves to an `id` |
|---|---|---|
| fundamentallyai.com | 21 | **0** |
| idea.uk | 4 | 1 |

**24 of 25 anchored links in the fleet point at an `id` that does not exist.**
The cause is a two-sided gap, not a writer bug alone: the content writer emits
plausible section anchors (`#decision-record`, `#reviewer-seats`, `#approach`),
and **no section component emits an `id` attribute** for the writer to target. So
even a well-behaved writer could not produce a working one.

Scale is small, which is why this is recorded here rather than as its own case —
but the failure rate where the pattern is used is effectively 100%, and it is
invisible to every existing check. Two of the three fix candidates are cheap:

1. **Extend the check** to resolve fragments against the assembled page's `id`
   attributes. This is what makes the class visible at all.
2. **Have section components emit a stable `id`** (the section/component name is
   the obvious candidate) and pass the page's available anchor list to the writer,
   the same way the real page list should be passed for paths (candidate 4 above).
3. Failing both, the writer should not emit fragments at all.

Note the interaction with this bug's main finding: on fundamentallyai.com these
21 links were *also* extension-less (`/capabilities#approach` on a `.html` site),
so they returned **404** rather than merely failing to scroll. Repairing the path
converts them from broken to inert — an improvement, not a fix. They still do
nothing when clicked, and a dead control is what `bugs_open/023`'s family exists
to catch.
