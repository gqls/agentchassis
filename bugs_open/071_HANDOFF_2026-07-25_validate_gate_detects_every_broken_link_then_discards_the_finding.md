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

> **MEASURED 2026-07-26 by the `bugs_open/079` thread — this candidate is not a
> build, it is a repair.** The machinery already exists and already runs:
> `page-content-writer`'s workflow calls `prepare_link_context` before the writer
> step, and the prompt template interpolates its `link_constraint_text` under a
> `{{if}}` guard. It is fed nothing. `PrepareLinkContextAction` looks for its page
> list at `db_sync.pages` plus three fallbacks, and **none of the four exists in
> that workflow's `collected_data`** — so it returns an empty list, the constraint
> text is `""`, the `{{if}}` elides the whole "## Internal Linking" block, and the
> model is unconstrained. Live: **20 of 20 recent writer runs recorded
> `page_count: 0`**, a 100% failure rate.
>
> Filed in full as **`bugs_open/092`**, with two traps for whoever takes it:
> `InjectLinkConstraints` is dead duplicate code and must NOT be wired (it would
> give the platform two implementations of the same prompt block), and
> `prepare_link_context` synthesises `"/" + name + ".html"` rather than reading
> `pages.url`, which would hand the writer plausible-but-wrong addresses.
>
> Left with 071 rather than taken, because 092 is prevention and this file owns
> the writer-side class. `who-owns.py` puts it here.

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

---

# Transferred in from `bugs_open/049`, 2026-07-26 — the fleet-wide measurement of this class

`049` (stale chrome + unbuilt-page links) closed on its own two mechanisms. Its **third**
mechanism is this bug's class, so it is handed over here rather than left in a closed file.
This is evidence, not a competing fix — `who-owns.py` puts the class with you.

## The live measurement

229 active pages across 8 sites fetched over HTTPS, every internal href extracted from the
**shipped** markup, all 274 unique targets probed (the RUNBOOK R15 method in
`cta_link_integrity/`, script `scripts/live_link_audit.sh`):

```
3,949 internal anchor instances · 274 unique targets · 65 return 404 · 3 return 301 (fine)
118 broken anchor instances on 59 of 229 pages
```

**Of those 118, roughly 61 are this bug's class** — the remainder were 049's own chrome/nav
mechanism (40, now fixed) and residual stale artefacts. Classified against the `pages` table:

| class | targets | what it needs |
|---|---|---|
| **href omits `.html`, target EXISTS and is live** | 8 | a href rewrite — the target is already 200 |
| **invented target, no `pages` row in any form** | ~48 | the writer named a page nobody planned |
| target exists but was never built | 4 | 049's `unbuilt_internal_link`, already detected |

The extension-less eight, each confirmed 200 at the `.html` form:

```
ai-agent-orchestration.com  /contact                     -> /contact.html
finetuning.uk               /contact                     -> /contact.html
finetuning.uk               /tools                       -> /tools.html
finetuning.uk               /tools/llm-cost-calculator   -> /tools/llm-cost-calculator.html
gaswholesalers.com          /contact                     -> /contact.html
leopardessconsulting.co.uk  /tools/llm-cost-calculator   -> /tools/llm-cost-calculator.html
robot-hands.com             /learning-center             -> /learning-center.html
robot-hands.com             /matchmatrix                 -> /matchmatrix.html
```

The invented ones cluster hard by site, which is the tell that they come from one writing run:
`ai-agent-orchestration.com` and `finetuning.uk` have 5 fabricated `/case-studies/*.html` each;
`leopardessconsulting.co.uk` has 6 fabricated `/services/*.html`; `robot-hands.com` has 14 under
`/learning-center/*` and `/matchmatrix/*`.

## Also transferred: the 9 dead `/tools/*.html` links from `bugs_closed/029`

`029` closed at the emitter (migration 211 — no NEW dead tool links can be created) and handed
its **existing** damage to 049. It is the same class as yours, so it comes here with the rest:
9 references in deployed page HTML, 8 pages, 3 sites, 5 distinct targets. Two of the five
(`tool-bayesian-ranking`, `tool-process-automation-scorer`) **exist at a different URL shape**,
so they are href rewrites; three have no page at all. The re-derivation query and the
`\.html`-filter trap (asset paths under `/tools/` dominate an unfiltered scan) are recorded in
`bugs_closed/049`'s handover section.

## One finding that bears on your fix candidates

`NormalizePagePath` (`datahelpers/links.go`) deliberately does **not** strip or append `.html`,
and it should stay that way. Making the matcher tolerant would silence the detector on
`/contact` — but `/contact` genuinely returns **404** on these sites, so the detector is right
and the tolerance would only hide a live defect. The repair belongs at the writer (your
candidate 4, `InjectLinkConstraints`, which this session confirms is still defined and never
called) or in a rewrite pass over stored `rendered_html` — not in the normaliser.

## Sighting, 2026-07-26 — component-card links on fundamentallyai.com (experience_register)

Evidence only; not a competing fix, and no new bug filed — this is your class, both halves of
it at once. Found by applying a harvested experience contract ("a card's outcome is a real page
load of its destination") to the components that carry it, then probing what they point at:

```
https://fundamentallyai.com/capabilities        -> 404      <- all FOUR hero-card-carousel cards
https://fundamentallyai.com/capabilities.html   -> 200
#review-council #verification #rapid-delivery #embeddings on capabilities.html -> NONE exist
image-hover-card-grid (/model-fine-tuning.html): cards -> #evaluation, #review-council -> absent
```

So each carousel card is dead twice: the **extension-less** class (your 8-target table — this
site was not in it) *and* the **fragment** blind spot, on one href.

Two things this adds to the picture:

1. **These are component-card links, not nav or chrome.** The 049 handover and the extension-less
   eight are all `/contact`, `/tools`, `/learning-center` — chrome-shaped targets. A component's
   own cards are written by a different path (the content writer filling `input_schema` fields),
   so a repair aimed at chrome links will not reach them.
2. **The site's own link audit called it sound.** `brochure_component_library` verified
   fundamentallyai.com on 2026-07-25 by crawling the served pages: 43 unique internal targets,
   0 broken. An audit that normalises to the `.html` form, or that drops the fragment before
   probing, reports a clean site while every card on two pages is dead. Whatever the fix is, the
   **detector** needs to probe the href as written, fragment included.

Owner of this file decides what to do with it; the experience_register workstream is not
starting a fix. Full context: `experience_register/harvest/HARVEST_02_2026-07-26_brochure_components.md` §4.

---

## 2026-07-26 — a fresh instance, and this time the platform authored the links itself

Same workstream, new evidence. A **full rebuild of fundamentallyai.com's index**
(queued to add a chart section) rewrote the page copy and introduced **six broken
internal links**, on a page that had been crawled clean the day before —
43 targets, 0 broken.

```
/about → 404   /capabilities → 404   /delivery → 404
/how-it-works → 404   /integrations → 404   /verification → 404
```

All six came from ONE component, `info-card-grid`, in its `cards[].link_url`.
Four name pages that do not exist on this site at all.

**The gate saw all six.** `validate_page_content` classifies a link with no
matching page as `phantom_link` with `Severity: "warning"`
(`validate_page_content.go:598-613`), and the comment beside it says why:

> Policy: a missing internal target is loud but NON-blocking — the improvement
> loop resolves it; a missing link is not a deploy stopper.

The improvement loop is off. So the page deployed with six links the platform had
just detected as broken, in the same run that wrote them.

**Both halves of the href problem are visible in one place here:**

1. `/delivery`, `/how-it-works`, `/integrations`, `/verification` — true phantoms,
   no `pages` row at any status.
2. `/about`, `/capabilities` — real pages, wrong form. `NormalizePagePath`
   (`datahelpers/links.go:169`) strips only a trailing `index.html`, so `/about`
   normalises to `/about` while the set member built from `pages.url` normalises
   to `/about.html`. They do not match, so these are flagged too — correctly, and
   just as ignorably.

**The part that is new, and worth more than the instance:** the 2026-07-25 link
repair was applied per page, to stored content. **A full rebuild silently undoes
it.** "The site is link-sound" is therefore a statement about an artefact, not a
property of the site — it expires the next time any page is rebuilt. That is an
argument for fixing this in the write path rather than by repairing pages.

Also relevant to the fix: `resolve_internal_links` **cannot** repair this
component. Its `ctaFieldNames` map (`resolve_internal_links_action.go:98`) lists
six components and `info-card-grid` is not among them; the comment states the
consequence exactly — "a button-bearing component missing from this set is
detectable but not repairable — its findings can only escalate to human review."

Repaired by hand again (`bak_pc_fai_index_links_20260726`), which is the third
time this class has been repaired per-page on this one site.

### CORRECTED 2026-07-27 — it was sixteen, not six, and my own check hid the other ten

The entry above says the index rebuild introduced **six** broken links and implies
the capabilities rebuild introduced none. **Both halves are wrong.**

The capabilities rebuild introduced **ten more**, all of the form
`href="/capabilities#review-council"` — extension-less *with a fragment* — split
4 in `hero-card-carousel` and 6 in `info-card-grid`. Sixteen in total across the
two pages I rebuilt.

**What hid them is the finding.** My per-component check was

```sql
regexp_matches(pc.rendered_html, 'href="(/[^"#?]*)"', 'g')
```

which **excludes every href containing `#`** — the exact anchor-blind pattern this
workstream recorded as landmine L2 the day before, after it hid 21 broken links
through a census, a repair and a post-check that all agreed with each other. I
read that landmine, wrote it into a handoff, and then used the pattern.

The live crawl caught it because it captures `href="(/[^"]*)"` and *then* strips
the fragment with `split_part(...,'#',1)`. That is the whole difference.

**Why this belongs in 071 rather than only in a workstream note:** it is direct
evidence for the second half of this case — that nothing validates the fragment.
All six distinct targets (`#review-council`, `#verification`, `#embeddings`,
`#rapid-delivery`, `#production`, `#decision-record`) resolve to **zero** `id`
attributes on the served page, so even repaired to `/capabilities.html#…` they
jump nowhere. The paths are repaired (`bak_pc_fai_cap_links_20260727`); the dead
fragments are left, because making them mean something requires the components to
emit ids, which is this case's to decide.

**The tally that matters for the fix:** on two rebuilt pages the writer produced
16 broken internal links, the gate detected all 16 as non-blocking warnings, and
every one of them shipped.
