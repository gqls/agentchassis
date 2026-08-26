# NOTES — bugs_open/407, a site cannot put its own page in its own header

Append-only, newest at the bottom. Technical log: evidence, commands, what the system
actually said, and every misstep.

---

## 2026-08-26 — lane opened; ownership checked, bug re-validated

### Why this bug

Picked after `bugs_open/359` closed out for the session. The sweep for the next genuinely
open, genuinely unowned, framework-shaped bug ruled out the nearer candidates on evidence
rather than on feel, and the reasons are recorded so the next sweep does not re-walk them:

| candidate | why not |
|---|---|
| `348` | superseded by its own banner — *"Do not fix from this file, `bugs_open/344` OWNS THIS"*, and 344 is closed |
| `361` | `bugfix_260_render_fallback`, ACTIVE 24 commits/14d |
| `376` | `loanzy_uk_example_site`, ACTIVE 119 commits/14d, 44 mentions, and a handoff literally named *"the route is gated on 376"* |
| `386` | `bugfix_386_counting_fact_drift`, ACTIVE, fix already live |

`407`: `scripts/who-owns.py 407` → **`likely OWNING workstream(s): (none identified)`**, with
two cross-reference mentions from the filing lane only. Filed today at the **owner's explicit
direction**, with his preferred direction quoted in the file.

### The bug is STILL VALID — re-measured with its own §6 query

`[MEASURED 2026-08-26]` **6 pages across 4 sites** declare `in_header`, are `active` +
`deployed`, are not child pages, and are absent from their site's primary nav:

```
ai-agent-orchestration.com  adoption-tracker      /adoption-tracker.html       nav_order 100
ai-agent-orchestration.com  news-index            /news/index.html             nav_order 100
ai-agent-orchestration.com  protocol-tracker      /protocol-tracker.html       nav_order 100
finetuning.uk               approach              /approach.html               nav_order 4
gaswholesalers.com          pricing-transparency  /pricing-transparency.html   nav_order 100
idea.uk                     report                /report.html                 nav_order 3
```

The filing recorded **8 across 5 sites**. **`loanzy.uk`'s two are gone and `idea.uk/report` is
new.** That movement is itself evidence for the filer's own §4 ⚠: a SECOND cause — a merely
stale nav — is mixed into this count, and loanzy.uk was their named example of it. So the
number is a true upper bound on this defect and not a measurement of it, exactly as filed.

**Building the discriminator is part of this job**, not a nicety: without it a post-fix count
cannot be compared with a pre-fix one, and the fix would be verified against a number that
moves on its own. The filer's attempt —
`pages.updated_at > max(site_nav_items.updated_at)` — does not work, because
`pages.updated_at` is bumped by any re-render and so does not tell you the FLAGS changed.

### §3 is the argument, and it is measurement rather than opinion

The existing workaround lives inside the broken thing. Three site-specific page names
(`model-directory`, `adoption-tracker`, `protocol-tracker`) were hardcoded into the
fleet-wide tier-2 map (`populate_nav_tables_action.go:461-471`) with a comment stating that
as tier 3 they would be *"the first thing dropped when max_header_items truncates"*.

`[MEASURED 2026-08-26]` **two of those three names are in the absent list above**, on the very
site the comment names. The documented remedy is in the source today and the pages it was
written for are still missing. That is what refuses candidate 3 (extend the list again) —
not a preference for structure, an observation that the last extension did not hold.

### Architecture read (first-hand, today)

- `classifyPagesForNav` (`:297`) sorts **tier ASC, then `nav_order` ASC**, then the caller
  truncates to `maxHeaderItems` (`:131`) and sends the overflow to `utility` so it survives in
  the footer. So `nav_order` genuinely cannot cross a tier boundary.
- `navPriorityTier` (`:442`) is two hardcoded maps plus `isSectionIndexType`.
- `max_header_items` is `PopulateNavTablesInputSpec.Optional`, default 8, read from the STEP
  config — **fleet-wide**.
- The action **`DELETE`s every `site_nav_items` and `site_nav_groups` row for the site**
  (`:160`, `:163`) and re-derives from `pages`. A per-site declaration therefore cannot live
  in `site_nav_items`: it is a derived table.
- `nav_prune_floor.go` guards that delete with two cohorts, and its header records that
  **seven other actions write `pages.in_header/in_footer`** while this one only reads them —
  which is what makes its "pages seen" cohort an independent second opinion.
- `classifyPagesForNav` also pins the `bugs_open/149` A2 invariant, enforced by
  `nav_membership_test.go`: *`pages.in_header`/`in_footer` declares nav MEMBERSHIP; a page's
  URL shape may decide WHERE it appears, never WHETHER it appears.* Any fix must keep it.
- `sites.settings` is the established per-site config home. `[MEASURED 2026-08-26]` keys in
  use fleet-wide: `maintenance_profile` 31, `pool` 17, `skip_build` 1, `skip_deploy` 1 — and
  `maintenance_profile` already carries Go-read sub-objects (`content_feed.enabled`,
  `content_feed.max_age_hours`, `structure_floor.n`/`.refusal`). There is **no** `site_specs`
  table.

---

## 2026-08-26 (later) — THE DISCRIMINATOR, built. And a misstep that nearly poisoned it

### The filer could not separate the two causes. The separator is the MECHANISM, not a timestamp

`pages.updated_at > max(site_nav_items.updated_at)` fails because `updated_at` is bumped by any
re-render, so it cannot say the FLAGS changed. The way in is to ask what the tier table can
actually do: **it only decides anything when there is COMPETITION for slots.** Below
`max_header_items` there is no competition, so tier cannot be the cause of an absence.

```
absent AND the site's primary nav is AT the cap    -> 407's mechanism
absent AND the site's primary nav is BELOW the cap -> NOT 407 (stale nav, or another cause)
```

`[MEASURED 2026-08-26]` it separates the set cleanly, **5 of 6 are 407 and 1 is not**:

```
 domain                     | name                 | nav_order | primary_items | cap | verdict
 idea.uk                    | report               |         3 |             5 |   8 | NOT 407 - nav below cap
 ai-agent-orchestration.com | adoption-tracker     |       100 |             8 |   8 | TIER/CAP (407)
 ai-agent-orchestration.com | news-index           |       100 |             8 |   8 | TIER/CAP (407)
 ai-agent-orchestration.com | protocol-tracker     |       100 |             8 |   8 | TIER/CAP (407)
 finetuning.uk              | approach             |         4 |             8 |   8 | TIER/CAP (407)
 gaswholesalers.com         | pricing-transparency |       100 |             8 |   8 | TIER/CAP (407)
```

`idea.uk` is the filer's own suspected second cause, now separated: its primary nav holds **5 of
8** and `/report.html` (nav_order 3, `in_header` true) is still absent — nothing to do with
tiers. **So the honest population for this bug is 5, not 6, and not the filed 8.** The full query
is in the RUNBOOK; two caveats travel with it and must not be dropped: it has to read
`max_header_items` from the live step config rather than hardcode 8 (hardcoding asserts the very
fleet-wide constant this bug is about), and "at cap" is sufficient *on today's data* — a site
that is at cap AND stale would be classified 407 wrongly.

### The sub-cause matters, because it rules out a whole family of fixes

Both flavours are live, and both are "the site cannot express its own priority":

- **beaten by a higher tier** — finetuning.uk's `approach` (tier 3) against a full header of
  tier 1/2.
- **beaten by a SAME-TIER TIE** — ai-agent-orchestration.com's `adoption-tracker` and
  `protocol-tracker` are tier **2** (hardcoded there for exactly this reason) and lose to
  `model-directory`, also tier 2; gaswholesalers.com's `pricing-transparency` is tier 3 losing to
  `why-gas-wholesalers`, `how-pricing-works` and `service-areas` — **all tier 3, all at
  `nav_order` 100**, so the tie is broken by load order.

**Consequence for the design: a fix that only lets a site RAISE a page's tier does not fix
gaswholesalers.** The declaration has to give a total order.

### Verified at the SERVED page, per the bug's own ⚠

- `ai-agent-orchestration.com` — index, services, about, tools, contact, case-studies, blog,
  **model-directory** = 8, exactly the cap. **This is §3's argument proven at the artefact**: of
  the three names hardcoded into tier 2 by that comment, one is in the header and its two
  siblings are not.
- `finetuning.uk` — index, services, about, tools, use-cases, case-studies, pricing,
  **your-own-model**, contact. The owner's offer page IS there now, after the second
  displacement the bug describes; `approach` is out.
- `gaswholesalers.com` — index, about, services, contact, news, why-gas-wholesalers,
  how-pricing-works, service-areas = 8. Note all three tier-3 names got IN: **tier 3 is not
  excluded, it is only outranked**, and the defect bites at the cap.

### Misstep 3 — I grepped `<header>` and matched a COMPONENT's header, not the site chrome

My first served-page pass ran `grep -o '<header[^>]*>.*</header>'` over each homepage. On
`idea.uk` that matched `<header class="info-card-grid__header">` — a **component's** header,
420 bytes of card-grid copy — and I recorded "ABSENT" for that site off it. The verdict happened
to be right and the evidence was worthless.

**A rendered page contains several `<header>` elements**, because components carry their own.
Match on the nav links, or on the chrome's own class, and sanity-check by printing the hrefs you
found — which is what caught it: idea.uk's "header" had no `href` at all, and a site chrome
without a single link is not a chrome.

---

## 2026-08-26 (evening) — BUILT, council APPROVED r1, and the objections moved the table

Commit `74e92e961`. Council `cb67cc71-b420-4399-9a52-f306d0f4bccf` — approved, 6 advisory
objections, none high. **Two of those objections changed the design before it shipped**, which
is the part worth recording.

### Misstep 4 — "there is NO site_specs table", read off my own truncation

I ran `psql -c "\dt site*" | head -20` and wrote the absence into my lane NOTES and then into a
**council submission's rationale**, as a load-bearing reason for choosing `sites.settings`.
`site_specs` sorts on line **21**.

It exists, carries 20 aspects, has `is_current`/`superseded_at` versioning, a `pinned` flag,
provenance columns and a real writer. And the part that actually decided the design:
`[MEASURED 2026-08-26]` its **`site_config` aspect already holds per-site HEADER config** —
`chrome.header_cta_url` and `chrome.header_cta_label` on oufe.com, `chrome.compliance_lines` on
two more. I was about to put header SLOTS in a different table from header CTAs, which is this
bug's own pathology one level up: a second home for one concern.

Caught by the council — `reuse_agent` and `prior_art_librarian` both at medium, neither able to
see my `\dt`; they knew the estate. **The declaration moved to
`site_specs.data->'chrome'->'header_slots'`** and inherits versioning, pinning and provenance
that the settings column never had. `WRONG_CALLS.md` carries the entry; the cheap check is
*never conclude ABSENCE from a command you truncated*, and *an estate-level absence claim
deserves a second instrument* — a grep for `site_specs` across `platform/**.go` returns readers
in five files and would have cost nothing.

### Misstep 5 — the mutation table, wrong again, and NOT for this morning's reason

Eight mutations; two survived the first pass. Nothing was passing on a downstream failure this
time — the tests were precisely aimed. **Their FIXTURES could not produce the failure:** the
total-order test declared `["pricing-transparency","service-areas"]`, which is already
alphabetical, so a mutation sorting the list was a no-op on that input; and the de-duplication
test used a corpus where every page was `in_header`, so utility was EMPTY and there was nothing
to de-duplicate.

Both fixtures now carry an explicit **setup assertion** that fails loudly if the corpus stops
exercising the path. Full entry in `WRONG_CALLS.md`, beside this morning's — the pair together
says more than either: a mutation table lies in at least two ways, and having learned one of
them I shipped the other the same day.

**And all eight were re-run AFTER the declaration moved table.** A mutation proof is about the
code that shipped, not the code that was planned; carrying the earlier results forward would
have been a claim about a file that no longer existed.

### The objections, and what each was worth

| seat | severity | outcome |
|---|---|---|
| `reuse_agent`, `prior_art_librarian` | medium | **DESIGN CHANGED** — see misstep 4 |
| `editquality`, `bug_historian` | medium | **CHECKED AND SOUND.** Both asked whether the page pool reaching classification is pre-filtered on `in_header` or URL — if it were, a declared page could never be found and would report `declared_missing`, silently reproducing this bug. It is not: `navPageScopeSQL` is `site_id = $1 AND status IN ('active','deployed','pending')`, no flag arm and no URL arm, and the URL bar is `isChildPageURL` in Go INSIDE the classifier, which the declaration overrides. |
| `guardian` | medium | **ESCALATED TO THE OWNER, not answered.** Overriding three membership guards at once is a widening beyond "fix the tier order" and is his call; recorded in `bugs_open/407` §B with both sides. |
| `guardian`, `debug_historian` | medium | migration fixes: a stray `ROLLBACK;` before `BEGIN;` for a sticky aborted transaction, and the deploy-verification step (which the file already carried). |
| `architecture` | medium | four ad-hoc `sites.settings` readers should be tracked before a fifth arrives. **Moot for this lane in the end — I did not add one** — and worth recording as standing debt for whoever does. |
| `prior_art_librarian` | medium | asked whether `site_nav_groups` could carry the declaration. **No, and for the same reason as `site_nav_items`: both are DELETEd for the site on every rebuild** (`:160`, `:163`). Answered in the code header so nobody re-asks. |

### Verified before committing, rather than asserted

- `[MEASURED 2026-08-26]` **0 of 51 sites declare `header_slots`**; 3 have a `chrome` object at
  all. The no-op is a query.
- `nav_membership_test.go`: **zero diff**, all 8 tests green.
- `[MEASURED 2026-08-26, recursive `$.**` walk]` **5 LIVE step rows across 5 agent types** call
  `populate_nav_tables`, all at `max_header_items` 8 — not the 7 across 6 my planning pass said,
  which had not filtered `is_active`/`deleted_at` (the other two are `multipage-website-builder`
  v1/v2, both inactive AND deleted). A one-level census would also have been wrong here for a
  different reason, which is why the walk is recursive.
- ai-agent-orchestration.com HAS a current `site_config` row (keys `locale`, `analytics`), so
  654's `jsonb_set` creates `chrome` and the sibling check has something to find.
