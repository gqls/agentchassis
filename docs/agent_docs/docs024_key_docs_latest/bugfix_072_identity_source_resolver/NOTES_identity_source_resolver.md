# NOTES — identity source resolver (`bugs_open/072`)

Append-only, newest at the bottom. Missteps are the point of this file, not an
appendix.

---

## 2026-07-31 — session start, picking the bug

Surveyed `bugs_open/` (73 files). `scripts/who-owns.py` says "OWNED or recently
active" for almost everything, because **filing the bug file is itself a commit**
that the script counts — so its verdict alone cannot separate "someone is fixing
this" from "someone filed it and left".

**What actually worked:** cross-referencing every bug number against the 40
most-recently-modified session transcripts under
`~/.claude/projects/-home-ant-projects-agentchassis/*.jsonl`. Thirteen bugs had
**zero** mentions in any other live session: 040, 072, 080, 081, 085, 092, 111,
113, 115, 118, 125, 132, 137.

That check also caught a near-collision: session **"bugfix 8"** started an hour
before me with the *identical* prompt and had read `143` and `145`, then went deep
on `143` only (reading the `bugfix_131_og_card` coordination doc, grepping for
`derive_card_asset`). Both are now in `bugs_closed/` — so had I taken 145 on the
"it says UNOWNED in the file" evidence, I would have duplicated a whole session's
work. **The bug file's own "OPEN, UNOWNED" line is a claim from the day it was
filed, not a status.**

Rejected 040 (kafka dial): a long-running instrumentation case with its own
workstream, awaiting days of metric — not a fix this session can land.

Picked **072** — and immediately hit the number collision: `072` is *also*
component-markup-without-CSS (`bugfix_072_component_css`, closed and live on
v1.0.1171). `who-owns.py` mixes both cases' commits into one answer. Resolve by
slug.

## 2026-07-31 — validating the bug, and the measurement that flipped it

The bug file says `contact-info` sources four fields from flat
`site_specs.identity.{email,phone,address,hours}` while the classifier nests them
under `identity.contact.*`. Both halves confirmed still true from the live DB:

- the component schema is unchanged (`updated_at` 2026-03-09) and still declares
  the flat paths, `email` with `on_missing: needs_human_review`;
- all 14 sites with an `identity` aspect nest a `contact` object.

Then I answered the bug file's own open question — *"verify the resolver supports
the deeper path before choosing, because `identity.contact.email` is three levels
and every working example today is two"*. It does: `navigateMap`
(`plan_sections_action.go:529`) splits on every dot and walks, arbitrary depth. So
candidate 1 was viable.

**Then I ran the query that put all three stores side by side, and it flipped the
bug.** The nested `contact` object exists on 14 of 15 sites — but its **values are
null/empty on exactly the 8 sites that fail**. Only 6 sites have a populated
`identity.contact.email`, and they are the same 6 that already work via the flat
key (the manual workaround wrote both).

⇒ **Repointing the schema at the nested path, or adding a nested-only resolver
fallback, resolves on 0 of the 8 broken sites.** The bug file's discriminator
table (flat email present ⟺ contact-info rendered) is a correct *measurement*
whose causal reading is inverted: the sites that render are the sites that have
contact data **at all**.

**Why the confound was so convincing, and this is the transferable bit:**
`jsonb ? 'contact'` and "the classifier writes nested" both check the **shape**.
Neither checks whether the shape holds a **value**. A nested object full of nulls
looks, to every shape-level check, exactly like a populated one.

## 2026-07-31 — where the data actually is

`sites.email` is populated on **12 of 15** real sites, including **5 of the 8**
that fail (oufe, robot-hands, vetcomparison, vonc, webdesign). `sites.phone` on 7.

The bug file **already contained this fact** and drew no conclusion from it:
*"The owner's phone +44 (0) 7934 524 911 had been written only to `sites.phone`,
which no component reads."* That sentence is the root cause, filed as a footnote
to a data workaround.

`\d sites` shows the columns are exactly the identity set: `company_name`,
`tagline`, `email`, `phone`, `logo_text`, `contact_address`, `logo_url`.

**Denominator trap caught here, before it reached a claim.** My first population
count ran over all of `sites` and returned 12 of 29. `sites` holds **14
`pool-*.internal` rows** — industry pools with no content. Unfiltered, the defect
looks three times worse than it is. `WHERE domain NOT LIKE 'pool-%'` is in the
RUNBOOK for this reason.

## 2026-07-31 — the fleet census, and a second finding

Replicated `resolveSpecPath` + `navigateMap` in SQL to ask which declared source
paths resolve **anywhere**. Of **100** distinct `site_specs.*` paths across active
components: **21 resolve on some site, 79 resolve nowhere.** Broken down:

| category | paths | aspects |
|---|---|---|
| A — resolves somewhere | 21 | commercial, cta, evidence_base, identity, portfolio |
| B — aspect exists, leaf never resolves | 5 | commercial, identity ← **072 lives here** |
| C — **aspect exists on NO site** | 74 | blog, case_studies, categories, contact, inventory, legal, nav, navigation, pages, pricing, product, search, social, social_proof |

**Category C is a separate defect** — components declaring sources against a
vocabulary no writer produces — and it is deliberately **not** in this fix.
Filed on its own so it is not smuggled in under a bug about contact details.

Two off-by-ones in that query that each produced a spectacular wrong answer
before I caught them (both now in the RUNBOOK): `segs[2:]` not `segs[1:]` (element
1 is the aspect, matched in the WHERE, not navigated — getting it wrong yields
"0 paths resolve fleet-wide"), and replicating `navigateMap`'s emptiness rule
(`""`, `[]`, null are NOT found — drop that and 79 becomes ~20, because
`"email": null` counts as resolved). **The Go function's emptiness semantics are
part of the measurement, not a detail of it.**

## 2026-07-31 — finding the precedent, which changed the fix

Before writing anything I grepped for existing readers of the sites identity
columns. This was the highest-value ten minutes of the session:

- `loadSiteDataFull` (`render_site_components_action.go:337`) — the full-writer
  render path — SELECTs exactly these columns.
- `buildRerenderBaseData` (`rerender_page_sections_action.go:579-586`) — the light
  rerender path — was **changed** to prefer `sites.email` over `content_data`,
  with the comment *"We now prefer the column … making both render paths agree.
  See bugs_open/006 §B."*

So this is the **third instance of one class**, and the other two are already
fixed in the direction I was about to propose. That reframed the change from "a
new resolver fallback" to "bring the last path into line with the two that
agree" — which is a much better-founded change and a much easier review.

**Had I not grepped, I would have designed a nested-only fallback** (the bug
file's candidate 2), shipped it, and fixed nothing. The measurement said candidate
2 was useless; the grep said what to build instead.

## 2026-07-31 — diagnosis loop: CONFIRMED

My root cause contradicted a filed diagnosis and asserted a structural property of
a shared mechanism, so per CLAUDE.md's "Diagnosis before debugging" default and
the owner ruling of 2026-07-31, I filed it before asserting it rather than after
being contradicted. Intake `1b9244c0`, run correlation `0f76987c`.

**CONFIRMED, first iteration, ~4 minutes.** It independently read
`resolve()`'s `site_specs` branch, followed `handleMissingField` →
`needs_human_review` → `shouldDefer` → `item.Status = "deferred"`, and fetched its
own live evidence — citing a `vetcomparison` row with `sites.email` populated and
every `site_specs.identity` column NULL. Its `NextScope` named `resolveSpecPath`,
`ensureSpecs`, `navigateMap`, `resolveConfigPath`: the same four functions I had
read.

Reading the queue first was also worth it (it was empty — 0
`awaiting_diagnosis` items — so no duplicate).

## 2026-07-31 — two compile failures that were not mine

`go test ./platform/orchestration/actions/` failed twice, neither at my hand:

1. `derive_brand_head_assets_test.go:108` — `!got[k]` on a `map[string]assetLock`.
   Commit `a22010eaa` widened `lockedBrandHeadKeys` to return `assetLockSet` and
   left the `map[string]bool` assertion behind, so **the actions test package has
   not compiled at HEAD since**. I opened an Edit on it and got "file has been
   modified since read" — **another session fixed it in the working tree while I
   was looking at the same lines**, with the same fix (`got.Locked(k)`). Left
   theirs alone.
2. `asset_lock_guard_test.go` — an **untracked** file from another session
   declaring `equalStrings`, which already exists at HEAD in
   `v3_site_reconcile_test.go`. Theirs to resolve, in their uncommitted file.

To run my tests without touching either, used `go test -overlay=` to map their
untracked file to an empty stub. Read-only w.r.t. the tree, so it cannot sweep or
clobber anyone's work — recipe in RUNBOOK §5. Better than moving a file aside,
which is a write to another session's state.

**Corollary worth keeping:** a green `go test` in this package right now is
evidence about *my working tree*, not about HEAD — HEAD does not compile until
that other session commits.

## 2026-07-31 — what shipped

`ef9e7e999`. `resolveSpecAlias` + `ensureSiteRow` + two enumerated maps in
`plan_sections_action.go`, 7 DB-free tests, PBP-026 in the concept register and a
`LANDMINES.md` entry **in the same commit as the seam** (ordering-exemption
condition 2). Council gate submitted as `dd03a73b` — committed with
`Council-Submitted:` rather than holding the code, because on this tree holding it
is not available (HEAD is shared and any session's roll ships it).

The safety property is the one a reviewer would otherwise have to trust: literal
path first, always wins, so **no path that resolves today changes its value**.
Test-asserted, not asserted in prose.

Deliberately did NOT `COALESCE` across columns the way `loadSiteDataFull` does.
That would have been the natural "consistency" tidy-up and it is a trap: an empty
value must stay empty here, or the fallback satisfies a `needs_human_review` field
with a value nobody supplied — a silently suppressed HITL request that looks
exactly like success.

## 2026-07-31 — corrections I made to my own filing, in order

1. **"The nested path fixes it"** — believed for about ten minutes on the strength
   of the bug file's discriminator table. Killed by the three-store query.
2. **"12 of 29 sites have an email"** — the `pool-%` denominator. Caught before it
   reached any document.
3. **"79 paths resolve nowhere" as one finding** — it is three categories with
   three different fixes, and lumping them would have justified a much larger
   change than the evidence supports. Split; category C filed separately.
4. **"The drop is silent" (from the bug file)** — stale. `plan_sections` now always
   emits `sections_deferred`/`sections_skipped` (`:922-924`, empty case at
   `:695-697`) and `persistSectionSkips` writes them durably. Candidate 3 needs no
   work; I nearly built it.

## 2026-07-31, evening — LIVE on v1.0.1218, and two corrections to my own claims

**Pod-verified on both replicas.** `776f55c5f9-bjfhq` and `-g9vqc`, started 17:59
UTC, image postdating both commits (16:06, 16:19 UTC). All four strings present on
each, against the positive control `plan_sections: loaded site_specs`.

**A control returned 0 and it was MY error, not the binary's.** I grepped for
`site_assets path resolved via the image-role alias`; the actual literal is
`…resolved via image-role alias` (no "the"). Re-run with the exact string → 1.
Worth recording because a 0 from a control has two causes with opposite meanings —
the feature is missing, or the probe is wrong — and I would have "discovered" a
missing feature if I had only run the failing form. **Copy control strings from the
source, never retype them.**

### CORRECTION 1 — "buys 5 sites" was wrong; it buys ONE PAGE

I claimed the fix would give 5 sites a contact block (oufe, robot-hands,
vetcomparison, vonc, webdesign), because those are the sites where `sites.email` is
populated and the spec is empty. **That conflated "the resolver can now find an
email" with "a page asked for one."**

Measured after the roll: only **8 pages fleet-wide** name `contact-info` in
`pages.sections`, and **7 already render 3-of-3**. Exactly one page is in the broken
state — `vetcomparison.uk/contact` (3 planned, 2 built). `vonc.com/contact` plans
only `["hero-contact","contact-form"]`; it was never going to render a contact block
whatever the resolver did.

**How I made it:** I derived the affected set from the *store* (which sites have an
email the resolver couldn't see) and never intersected it with the *demand* (which
pages actually request the component). Both halves are needed and I measured one.
**The cheap check:** `WHERE p.sections::text LIKE '%<component>%'` — one query,
which I only ran when the runbook's verification step sent me to vonc and I found
its plan had no `contact-info` in it. My own verification recipe named the wrong
site, which is the tell that the claim behind it was untested.

The fix is unchanged and still correct — it is fleet-wide, it makes a *new* site
work by default, and the structural point (three paths, one canonical store) stands.
Only the impact number was inflated.

### CORRECTION 2 — the live acceptance test cannot run, for a reason outside this bug

Fired the acceptance build as `needs_page`/`triaged`/`pipeline=build` (item
`45f9b005…`), copying the shape the dartsonline lane proved works. It has not been
claimed. **`build-dispatch-loop` last completed anything at 15:44 UTC; 72 items sit
at `triaged` with `handled_by` NULL.** `build-pipeline-trigger` is enabled and fired
at 18:15 — the gate passes, the *consumer* has stopped. That is `bugs_open/029`,
already owned; another lane committed "dispatch still the blocker" (`475d55c0b`) the
same afternoon. **Left the work item queued and did not touch their lane.**

**Corroboration found while checking the queue, which I should have looked for on
day one:** a `needs_section_data` item —
`section_data_contact_contact-info_72b9e3a6…` — has sat at `needs_human_review`
since **2026-07-17**, raised by exactly the withholding this bug describes, naming
the page AND the section. The bug file says "no work item naming it"; there is one,
and it is two weeks old. Checking the work queue for the symptom is cheaper than
the fleet-wide discriminator that started this, and it points at the same page.
