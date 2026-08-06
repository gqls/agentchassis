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

## Sighting, 2026-07-27 — relojistas.com: the phantom is a PLATFORM DEFAULT, not a writer invention (traffic_probe)

Different site, different language, and a **different producer** — which is why
this is worth adding rather than logging as one more instance.

**Live, 2026-07-27** (sweep of all 19 deployed pages, 27 distinct internal
targets, fragment-stripped so it is not the anchor-blind pattern above):

```
404  /contact.html              <- linked from /index.html, /noticias/index.html, /glosario/index.html
404  /assets/images/favicon.png <- linked from all 19 pages (separate, long-known gap)
301  /guias/mantenimiento       <- harmless redirect
```

**Nothing authored `/contact.html`.** It is absent from `content_data` and
present only in `rendered_html`:

```sql
SELECT p.name, pc.slot_name,
       pc.content_data::text ILIKE '%contact.html%' AS in_data,
       COALESCE(pc.rendered_html,'') ILIKE '%contact.html%' AS in_rendered
FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE p.site_id='ecf15e75-a966-4900-bcb0-1c85f689dbfd' AND ...;
--  glosario-index | hero | f | t
--  index          | hero | f | t
--  noticias-index | hero | f | t
```

The homepage hero's `content_data` is Spanish and has **no URL field at all**:

```json
{"cta_text": "Leer las últimas noticias", "secondary_cta": "Explorar el glosario", ...}
```

and renders as

```html
<a href="/contact.html" class="btn btn-primary">Leer las últimas noticias</a>
```

The producer is `component_library.go` — an unconditional English default,
present in three places (`:768-769`, `:821-824`, `:893-894`):

```go
"cta_text": defaultString(ctx.CTAText, "Get Started"),
"cta_url":  defaultString(ctx.CTAUrl, "/contact.html"),
// and, as a map:
"primary_cta_url": "/contact.html",  "secondary_cta_url": "/about.html",
```

This is the `LNK-007` fallback that `render_site_components_action.go:791`
already instructs the model to avoid — the instruction constrains the *writer*,
while the default fires in the *renderer*, downstream of it.

**Why it sharpens this case rather than repeating it:**

1. **Both detectors fired and the default still won.** `internal-link-resolver`
   filed **18 `unresolved_cta` rows** (2026-07-18 → 07-21, all still
   `needs_human_review`) naming the exact fields — *"Unresolved CTA on index
   ('hero'): no real-page destination for secondary_cta_url"*. So the platform
   escalated to a human queue **and** shipped a hardcoded 404 in the same breath.
   The queue it escalated into is `bugs_open/033` (no working surface), so the
   escalation arm is inert by construction.
2. **A writer-side fix cannot remove this class.** `092` (writer never gets link
   constraints) addresses invented targets. This link was never invented by a
   model — a page with *perfect* authored content still gets `/contact.html`,
   because the default fills a field the author left empty.
3. **On a non-English site the default is unfixable by resolution.** `/contact.html`
   can never resolve on relojistas: the page is `/contacto.html`. The same
   defaults ship English *copy* too — live on this Spanish site today:
   `"Browse all guides"` (self-linking, on `/guias/index.html`),
   `"Explore All Archetypes"` (→ `/noticias/`, on `/glosario/index.html`), and a
   footer `<h4>Contact</h4>` on all 19 pages.
4. **Authored content is silently dropped alongside it.** The hero's
   `secondary_cta`, *"Explorar el glosario"*, appears **nowhere** in the rendered
   output — the anchor is not emitted at all. So the same path that invents a
   dead primary CTA discards a good secondary one. [UNDIAGNOSED — I did not trace
   which of the three default maps ran; the observable is the missing anchor.]

**Bearing on fix candidates:** a repair that rewrites stored `rendered_html` does
not hold here for the reason already recorded above (a rebuild re-runs the
default), and a repair that only constrains the writer does not reach it at all.
The candidate that closes the door is making an unresolved CTA URL **unrenderable**
— emit no anchor rather than a guessed one — which is also what the component
already does for the secondary CTA.

**Check-methodology landmine, paid for here:** on a Cloudflare-proxied site
`grep -c 'mailto:'` returns **0 no matter what**, because CF rewrites every
mailto into `/cdn-cgi/l/email-protection#<hex>`. A "0 emails" pass is vacuous.
Decode with XOR against the first byte before believing it — that is how the
homepage's `relojistas@contactforsales.com` and an **empty** footer anchor
(`<a href="/cdn-cgi/l/email-protection#07"></a>`) became visible.

### Same day, proven by accident: you cannot REMOVE a CTA, and trying makes it worse

Acting on the owner's ruling that relojistas has no contact route, I tried to retire two
self-referential hero CTAs ("read the latest news" *on* the news index, "explore the
glossary" *on* the glossary index) the obvious way — **deleting `cta_text` from
`content_data`**, relying on the template's own guard `{{if and .cta_text .cta_url}}` to
omit the anchor.

The pages re-rendered as:

```html
<a href="/contact.html" class="btn btn-primary">Get Started</a>
```

**English "Get Started" pointing at a 404, on a Spanish site** — strictly worse than the
Spanish-text-on-a-dead-link it replaced. `component_library.go` had refilled *both* fields
from its defaults (`cta_text` → `"Get Started"`, `cta_url` → `"/contact.html"`), and the
refilled pair then **satisfied the template's guard**.

This is the mechanism stated as sharply as it can be:

1. The component's guard is correct and would have omitted the button.
2. The default fires **upstream of the guard**, so it doesn't just supply a bad value — it
   **manufactures the very condition the guard tests for**.
3. Therefore *"leave the field empty"* is not an available way to say "no CTA here". The
   absence of a CTA is **unrepresentable** in `content_data`; only a non-empty, valid
   destination suppresses the phantom.

**What this does to the fix candidates.** It rules out the cheap repair outright: no
data-side cleanup — not this one, not `resolve_internal_links`, not a per-page hand-repair —
can express "no button", because every one of them works by writing `content_data`. The
only fix that closes the door is at the default itself: **emit no anchor rather than a
guessed one** (drop the `cta_url`/`cta_text` defaults, or make them empty strings so the
existing guards do their job). Note the fleet-wide consequence to survey first: any page
today relying on the default to produce a *working* button on a site that happens to have
`/contact.html` would lose it — which is the correct outcome, but it is a visible change and
should be counted before it ships.

Worked around here by giving both heroes real destinations, since the site thread cannot
change the default. That is a workaround, not a fix: **every hero on every site must now
carry an explicit `cta_url` forever**, and any that doesn't is one render away from an
English button pointing at a 404.

---

## Triage 2026-07-27, post-roll (v1.0.1174) — THIS FILE'S TITLE IS NO LONGER TRUE, and three named residues are

Verification sweep, not a fix. The fleet rolled to `v1.0.1174` at 15:11 UTC; the last
Go commit in that image is `e96d42226` (14:52 UTC), and there are **no Go commits after
it**, so the tree and the binary agree on everything below.

### What is now fixed, live AND exercised in production

The headline mechanism — *detect, then discard, then ship the 404* — is gone on the
in-body writer path. `bugs_open/079`'s `RepairPageLinks` runs at the same gate, after
every check, and rewrites-or-unlinks before `save_sections` persists
(`validate_page_content.go:356-371`). Candidate 1 (persist the finding) and candidate 5
(fix the lying comment) both landed with it — the policy comment at `:803-822` now
states the real reason and names what happens if `repair_internal_links` is turned off.

**Not inferred from the diff — the durable record this file asked for exists, with both
arms firing in one real build:**

```sql
SELECT domain, occurred_at, error_message FROM agent_error_log
WHERE error_code='CONTENT_LINK_REPAIR_DETAIL';
-- dartsonline.com | 2026-07-27 12:23:23+00 |
--   Repaired 2 dead internal link(s) before save: 1 href(s) rewritten, 1 link(s) removed
```

Pod-grep on `agent-chassis-5994dc6d6c-pt8v9` (v1.0.1174):
`link removed before save, anchor text kept` → 1, `href rewritten to` → 1.

### What is left, and it is not the same bug

1. **The fragment blind spot — unfixed, and now acknowledged in the code itself.**
   `link_repair.go:117-119` says it outright: *"Whether that fragment resolves to a real
   id is a separate, known gap … Repairing the path turns those from 404 into inert,
   which is an improvement, not a fix."* Nothing emits section `id`s; nothing checks
   fragments. This half of the file is untouched.
2. **The renderer-default class (relojistas) — unfixed in code.** All three hardcoded
   `"/contact.html"` defaults are still in `component_library.go` (`:769`, `:822`,
   `:894`), and they fire *downstream* of the gate, so no amount of repair at
   `validate_page_content` reaches them. **The live symptom has moved, not resolved**:
   relojistas.com's homepage no longer renders `/contact.html`, it renders
   `/contacto.html` — which returns **404 as well** (probed 2026-07-27; `/contact.html`
   404 too). That CTA is the brochure/traffic_probe lane's; the *default* is this file's.
3. **A repair is a write-path fix, so already-deployed pages keep their damage.** Live
   2026-07-27: `robot-hands.com/learning-center.html` still serves six 404s — see the
   triage note appended to `bugs_open/097`, which owns that instance.

### Triage recommendation (not executed — this file is actively worked by the CTA lane)

This file now bundles one **closed** mechanism with three open ones, which is the shape
that makes a bug file un-closable forever. Recommend: close 071 on the evidence above and
re-file the fragment gap as its own case (the renderer-default belongs with it or with
`component_library.go`'s owner). Left undone deliberately — filing new numbers from a
sweep is how `083` and `090` collided, and the owning lane should pick the split.

---

## Sighting + a NEW normalisation gap, 2026-07-27 — webdesign.co.uk home page (webdesign_couk thread)

**Evidence and one new finding. Not a competing fix — I have changed no platform code**,
because §"One finding that bears on your fix candidates" already reasons about
`NormalizePagePath` and concludes the repair belongs at the writer. This is offered as the
case that refines that section, not as a challenge to it.

**The instance.** The owner found no link on `webdesign.co.uk`'s home page worked. Measured
live: **10 of 13 hrefs 404**; only the three nav links survived. All 12 cards across two
`info-card-grid` components were dead. Both halves of your class at once:

1. **True phantoms** — `/tools/colour-contrast-checker`, `/tools/css-layout-generator` (real
   pages are `smart-contrast`, `layout-generator`); `/tools/spacing-scale-calculator` and
   `/tools/typography-scale` name tools that exist in **no** form among the 63 built; four
   category links (`/tools/colour|css|typography|accessibility`) point at category pages that
   were never built.
2. **Real pages, wrong form** — `/tools`, `/guides`.

Slugs are absent from `cmd/webdesignport`, so they came from generation, not the port. Note
these are **not prose links**: they are structured `content_data->'cards'->[]->'link_url'`
values on a data-driven component. Worth knowing for `092` — a writer-prose repair would not
have reached them.

### The new part: on `dir/index.html` sites the normaliser produces a FALSE MATCH

Your §line 326 analysis covers the flat-file shape: `pages.url = /about.html`, href `/about`,
normalised forms differ, correctly flagged. **The `dir/index.html` shape inverts that result.**

`NormalizePagePath` strips a trailing `index.html` (`links.go:175`), so:

```
pages.url  /tools/index.html   ->  /tools
href       /tools              ->  /tools      ==> MATCH, not flagged
live       /tools              ->  404
```

So on any site whose pages are `dir/index.html`, every extensionless or trailing-slash link is
**invisible to both the gate and the audit** — they agree, by the one shared implementation,
that it is fine, and it is a live 404. This is the one link of the ten that the audit would
have passed even had it run.

**This is not a per-site quirk. Measured 2026-07-27 across four domains:**

```
webdesign.co.uk   /tools/ 404  /tools 404   /tools/smart-contrast/ 404
                  /tools/smart-contrast/index.html 200
relojistas.com    /about/ 404  /about 404
robot-hands.com   /about/ 404  /about 404
gaswholesalers    /about/ 404  /about 404
```

Cause: the sites are served from an **S3-compatible bucket behind Cloudflare** (`x-amz-*`
response headers on every request). **An object store does not resolve directory indexes** —
this is inherent, not a misconfiguration, and it will never start working. Site *roots* serve
because the bucket has a default root object; subdirectories never will.

**Why I stopped rather than changing `links.go`.** Removing the `index.html` strip would make
the matcher stricter — aligned with your section's logic, which warns only against making it
*more* tolerant — but it is your code, `071` is owned by two active workstreams, and
`rerender_page_sections_action.go:429` compares `NormalizePagePath(current)` against
`NormalizePagePath(pageURL)`, so the strip may be load-bearing there. That interaction wants
the owner's judgement, not mine.

**Corroborating your own strongest point.** §"The part that is new" says a per-page link repair
is a statement about an artefact that expires on the next rebuild. This instance is the proof:
I repaired the home page in `content_data` **and** `rendered_html` **and** the published file
in `gqls/sites`, and all three will be overwritten the next time the page is generated, because
nothing upstream has changed. The site is link-sound this afternoon and is not link-sound as a
property.

### Related, filed separately

The audit that would have caught nine of these ten **has never run — on any site**. That is a
coverage failure rather than a detection one, so it is `bugs_open/116` rather than more text
here.

---

## 2026-07-28 — third instance, and this time it was caught by a BASELINE rather than by luck

Contributed by the brochure_component_library thread. **Same site, same mechanism, third
occurrence** (index 07-26: 6 links; capabilities 07-26: 10; capabilities 07-28: 9).

Rebuilt `fundamentallyai.com/capabilities` to add the `evidence-chart` section
(work item `8f366ce5`, `complete`, attempt 0). The build succeeded and the chart is
correct. It also **authored 9 internal link targets that did not exist before**, of which
**7 are confirmed 404** and 2 were unresolved under origin throttling:

```
/asset-recovery                404      /commerce            404
/capabilities/rapid-delivery   404      /decision-record     000
/capabilities/review-council   404      /delivery            000
/capabilities/verification     404      /verification        404
/contact                       404   <-- contact.html EXISTS and serves 200
```

`/contact` is the sharpest of them: the page is there, the writer emitted the
**extension-less** form, and this site serves `.html`. That is landmine L1 of the
brochure workstream, authored fresh by a gate-passing build.

The other eight are invented destinations — `/capabilities/<slug>` sub-pages and
top-level nouns that have never existed on this site.

### What is new here, and why it is worth adding to this file

**I took a link baseline before firing the rebuild** (9 internal targets, captured from
the served page) precisely because this had happened twice. Without it the after-state —
18 targets — looks like a page with some broken links, indistinguishable from
pre-existing damage. With it, the 9 are provably **authored by this build**.

> **The recommendation this file should carry: baseline before every rebuild.**
> ```bash
> curl -fsS https://<domain>/<page>.html \
>   | grep -oE 'href="(/[^"]*)"' | sed 's/href="//; s/"$//' | cut -d'#' -f1 | sort -u > before.txt
> # rebuild, then:
> comm -13 before.txt after.txt      # exactly what this build authored
> ```
> Capture `href="(/[^"]*)"` and strip the fragment **afterwards** — `[^"#?]` is
> anchor-blind and is how 21 broken links survived three agreeing checks on 07-25.

**Not hand-repaired this time, deliberately.** The 07-25 and 07-26 repairs were per-page
edits and **neither survived the next rebuild of that page** — which is this bug's whole
point. A fourth hand repair would produce a fourth data point for a conclusion already
established. The fix belongs where the gate discards the finding, or upstream in
`bugs_open/092` (the writer never receives its link constraints), which is the plausible
cause of the invented destinations.

**Live exposure right now:** `capabilities.html` serves 200 with at least 7 dead internal
links. Recorded rather than silently repaired so the next thread inherits the true state.

---

## 2026-07-28 (later, same thread) — the two loose ends closed, and the mechanism goes one level deeper

**All 9 authored targets are now confirmed 404** — `/decision-record` and `/delivery`,
unresolved under origin throttling this morning, resolve 404 on a clean retry. Full
crawl of all 10 deployed pages: every one of the 13 broken references this build
authored sits on `capabilities.html`; the rest of the site is link-clean (one
site-wide `favicon.png` 404 predates the build and is tracked separately).

**The same build also invented four image paths** — `<img
src="/assets/illustrations/{review-council,rapid-delivery,verification-audit,vector-search}.svg">`
replacing four working `/assets/images/*.jpg` references. Checked fleet-wide: exactly
**one** component in the entire estate references `/assets/illustrations/` — this one.
The directory has never existed on any site. So the invention surface is `src`
attributes as well as `href`s, and `RepairPageLinks` (079's fix) never had `<img>` in
its remit.

**And the sharpest finding: this file's title is now true TWICE OVER.** The gate
detected every broken link this build authored, *repaired them*, logged the repairs
durably (`CONTENT_LINK_REPAIR_DETAIL`, 10:45:01.347Z, all 9 targets named) — and the
repair output was then **discarded at persistence**: `save_page_sections` prefers the
structured `sections_metadata` path and never reads `validation_result.clean_html` on
the primary build plan. The unrepaired sections were saved 400ms after the repair log
and deployed. Full mechanism, evidence and fix candidates: **`bugs_open/079`, REOPENED
2026-07-28** — detect → discard became detect → repair → discard the repair.

---

## 2026-08-06 — the FRAGMENT half is now built (committed, inert until the roll), and its exposure had MOVED

Taken up by the `bugfix_071_fragment_blindspot` lane, which picked this file's
own triage note ("the fragment blind spot — unfixed, and now acknowledged in the
code itself") as the last residue nobody owned. **The other two residues are NOT
mine and are not touched:** the renderer-default class is `bugs_open/203`'s
active lane (its 08-05 fix `880a405a6` removed the `cta_url` scalar defaults;
the `primary_cta_url`/`secondary_cta_url` map at `component_library.go:1136-1147`
is still live and is recorded here for them), and already-deployed damage stays
with the site lanes.

### The exposure this file measured has largely healed — and that is not a reason to close it

> **CORRECTION to this file's §"Related defect", re-measured 2026-08-06.**
> "24 of 25 anchored links in the fleet point at an `id` that does not exist" is
> **no longer true**. Today, across every active shipped page's stored
> components: **5** `path#fragment` links (all idea.uk, all resolving — probed on
> the served pages) and **61** bare `#fragment` links, 57 of them `#content`
> skip-links whose target id is present in the stored rows *and* on the served
> page. Every fragment probed resolves. Live damage ≈ **0**.
>
> The healing was per-page repair plus `092`'s writer constraints, **not** a
> check — nothing has ever guarded this. So the class is exactly where
> `bugs_open/093` was when it was fixed: *live exposure nil, structural gap
> real*. This file's own history is the argument: the writer re-authored dead
> anchors on **three consecutive rebuilds** of one site.

### What was built

`dead_fragment_link`, an **arm inside `check_phantom_internal_links`** — not a
new check, because a new check needs its own entry in a discovery agent's
`checks` array, which is how `093`'s fix reached production correct and never
executed. That check is already enabled on `completeness-discovery-agent` and
already filing (119 `phantom_internal_link` items, 55 complete, newest 08-04).

Reuse over new code: `datahelpers.DocumentIDs` is the id-presence test
**extracted from `OrphanElementRefs`** (`check_orphan_element_refs`), which asks
the same question from the other end — a *script* naming an id the page lacks,
where this is a *link* naming one. Every conservatism in that test was bought
with a false positive on a working tool. `OrphanElementRefs` now runs on top of
it, so the two cannot drift. Plus `datahelpers.SplitFragment` (splits on the
FIRST `#`, so a `?query` stays with the path — `NormalizePagePath`'s combined
`IndexAny("#?")` strip is right for page identity and would flag a working
`/search.html?q=1#results`).

Three rules — bare `#x` on a page resolves against that page's whole document +
chrome; bare `#x` in **chrome** is dead only if it resolves on **no** page (it
renders on every page, so per-page judgement files N items for one template);
`/page.html#x` resolves against the **target** page. Four silences: noop hrefs
(`dead_controls`), runtime-fill shells, a phantom or never-deployed target
(reported once already, by another arm), and a target with no stored HTML.
Severity **low** — this file's own distinction: repairing a path turns a link
"from 404 into inert, which is an improvement, not a fix", and inert is what
this arm reports.

Prevention half: `buildLinkConstraintText` now tells the writer not to author
`#` anchors at all. No caller supplies an anchor list, so every fragment it
emits is an invention — correct-or-absent, the LNK-005 shape.

**Blast radius, measured before submission rather than asked of the reviewers:**
the shipping functions run over a 7.5 MB dump of all 38 sites' live components →
**67 fragment-bearing hrefs, 0 findings**; the same corpus with two planted dead
fragments → exactly **2**. The zero is disconfirmable.

Commit `af2667453` (Council-Submitted `bbbb4132-4abe-4db1-a1ba-755377dab009`),
plus migration `322` for the claim-timeout lockstep the verifier obliges.
Register: **LNK-031**; and **LNK-009's status line is corrected** — it said "NOT
YET ENABLED (deliberate)" while the check had been enabled and filing for weeks.

### Still open on this file after this change

1. **The gate still cannot judge fragments.** `validate_page_content` sees the
   writer's `page_html` **without chrome**, so a chrome-satisfied anchor would
   false-positive. Needs a chrome-aware id load at the gate.
2. **Nothing REPAIRS a dead fragment.** Unlinking a label-bearing anchor leaves
   the label as bare text (a recorded landmine), so detection first, volume
   second, repair third.
3. **No section component emits a stable `id`** — the capability half. Until it
   does, a fragment link cannot be made to work *on purpose*, only avoided. This
   changes every page's rendered HTML fleet-wide and wants its own round.
4. This file still bundles a closed mechanism with open ones, which is the shape
   the 07-27 triage warned makes a bug file un-closable. The split is still the
   owning lane's call; the fragment half is now the smallest and best-defined
   piece of it.

Lane docs: `docs024_key_docs_latest/bugfix_071_fragment_blindspot/`
(PLAN / NOTES / RUNBOOK / README_where_we_are).

### 2026-08-06 (post-roll) — the fragment arm is LIVE on v1.0.1259 and induction-proven; this residue is CLOSED, the file is not

**Council APPROVED round 1** (`bbbb4132-4abe-4db1-a1ba-755377dab009`, 3 advisory
objections, none high). One of them earned its keep: the `guardian` seat asked
whether the shared-helper refactor had callers I had not checked, and it did —
`deploy_tool_action.go:182`, a **hard pre-deploy refusal gate** for tool birth.
Settled by restoring the pre-refactor implementation from `af2667453^` and running
both over every component template in the estate plus every page, site and
whole-page document in the fleet: **4,036 documents, 0 mismatches** (the first run
of that differential was vacuous — both agreed only about `nil` — and a guard
written into the harness beforehand caught it; re-run with id-stripped variants
gives 403 discriminating cases).

**Live and proven, on the deployed binary rather than in a test.** Pod-grepped both
replicas (`dead_fragment_link` 10, was 0; positive control `phantom_internal_link`
10; invented control 0), then a real `completeness-discovery-agent` dispatch
against a four-case fixture on a pool site (`status='pool'`, nothing serves it):

| case | expected | got |
|---|---|---|
| bare `#x`, no such id | FIRE | filed |
| bare `#x`, id on the same page | silent | silent |
| `/page.html#x`, target lacks the id | FIRE | filed |
| `/page.html#x`, target HAS the id | silent | silent |

Two items, `low`, `page-build-handler`/`content`, filed against the page
**containing** the link. **And the retraction was proven with one variable:**
repairing only the bare case — adding the id, leaving the href untouched — dropped
the findings from 2 to 1, the survivor being the cross-page one. Same binary, same
run, same data. That is what establishes the shipped code resolves fragments
against document ids rather than pattern-matching hrefs.

The `<a href="#">` in the same fixture was claimed by `dead_control`, **not** by
this arm — so the one fixture also proved the remit boundary in the direction that
matters, and nothing double-reports.

Fleet re-measured post-roll: **67 fragment-bearing hrefs, 0 findings**. Fixture
deleted, pool site proven back to 0 pages / 0 items in the same statement.

**Owed, small:** `VerifyDeadFragmentLinkResolved` has not itself executed. Its
three SQL shapes were validated in both directions against the live fixture, but
the function is reachable only through `CompleteWorkItemAction`, whose live callers
are the dispatch loops — and `build-dispatch-loop` takes `item_domain='build'`
while these items are `content`. The first real completion of a
`dead_fragment_link` item runs it.

**THIS FILE STAYS OPEN.** The fragment *detection* residue is done; three named
pieces of the fragment class are not, and the file's other mechanisms are not mine:
(1) the deploy gate still cannot judge fragments — it receives `page_html` without
chrome, so gate-side validation would false-positive on every chrome-satisfied
anchor; (2) nothing repairs a dead fragment, and unlinking a label-bearing anchor
leaves the label as bare text; (3) **no section component emits a stable `id`**, so
a fragment link can still only be avoided, never made to work on purpose — that is
the capability half, it changes every page's rendered HTML fleet-wide, and it wants
an architecture round rather than a bug patch. Both the `bug_historian` and
`architecture` seats noted that link-target resolution now has **three unaligned
consumers** (the gate, this arm, `link_repair.go`); architecture still graded the
change `point_fix` and observed that `DocumentIDs` is positioned so (3) has a
validator ready when it ships.

Lane docs, incl. the roll-time procedure and five paid-for traps:
`docs024_key_docs_latest/bugfix_071_fragment_blindspot/` (HANDOFF / PLAN / NOTES /
RUNBOOK / README_where_we_are).
