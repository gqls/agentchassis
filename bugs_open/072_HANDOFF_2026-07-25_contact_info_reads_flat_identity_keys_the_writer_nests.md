# 072 — `contact-info` can never render on 8 of 13 live sites: it reads flat `identity` keys that the writer nests

**Filed:** 2026-07-25, from fundamentallyai.com — the owner supplied a phone
number on 2026-07-24 and it never appeared on the contact page.
**Severity:** medium-high, silent, fleet-wide. A contact-details block is missing
from most sites' contact pages and nothing reports it.
**Status:** OPEN — diagnosed with a fleet-wide discriminator, data worked around
on one site, contract not fixed.

## Symptom

The `contact-info` section (the block that displays phone / email / address /
hours) is absent from most contact pages. On fundamentallyai.com it is in the
site plan at `ordering=2` and in `pages.sections`
(`["hero-contact","contact-form","contact-info"]`), the component row is
`is_active=true`, `component_level='section'` — and every contact build produces
**two** components, never three. No error, no `sections_skipped` entry, no work
item naming it.

## Root cause: a path contract mismatch

`content_components.input_schema` for `contact-info` sources its fields from
**flat** keys:

```json
"email":   {"source": "site_specs.identity.email",   "on_missing": "needs_human_review"},
"phone":   {"source": "site_specs.identity.phone",   "on_missing": "skip_field"},
"address": {"source": "site_specs.identity.address", "on_missing": "skip_field"},
"hours":   {"source": "site_specs.identity.hours",   "on_missing": "skip_field"}
```

The `identity` aspect, as written by `domain-research-classifier`, **nests them**:

```json
{"contact": {"email": "…", "phone": null, "address": null, "location": null}, …}
```

So `site_specs.identity.email` resolves to nothing even when
`identity.contact.email` is populated. Because `email` carries
`on_missing: needs_human_review`, the whole section is withheld from the build.

### The discriminator (this is what proves it)

`contact-info` has rendered on **exactly** the sites whose `identity` aspect
carries flat keys. Measured 2026-07-25 across all 13 deployed sites:

| site | flat `email` | flat `phone` | contact-info rendered |
|---|---|---|---|
| ai-agent-orchestration.com | yes | yes | **yes** |
| finetuning.uk | yes | yes | **yes** |
| gaswholesalers.com | yes | yes | **yes** |
| leopardessconsulting.co.uk | yes | yes | **yes** |
| idea.uk | yes | no | **yes** |
| fundamentallyai.com | no | no | no |
| dartsonline.com | no | no | no |
| gamesdesign.co.uk | no | no | no |
| relojistas.com | no | no | no |
| robot-hands.com | no | no | no |
| vetcomparison.uk | no | no | no |
| vonc.com | no | no | no |
| webdesign.co.uk | no | no | no |

Five rendered, five have flat `email`. Eight lack it, eight have no
contact-details block. **No exceptions in either direction.** Note idea.uk
renders on flat `email` alone with no flat `phone` — consistent with `email`
being the field whose `on_missing` withholds the section and `phone` merely
`skip_field`.

Every site in the "yes" group has flat keys because something wrote them there
later (hand-edit or a different writer); the classifier's own shape is the nested
one, so **a new site is broken by default**.

## Second defect: the drop is silent

`select_sections` returned `sections_ready` with 2 entries and **no
`sections_skipped` / `sections_deferred` keys at all** (orchestration
`de2da37d`, contact build 2026-07-25 17:12). A section withheld by
`on_missing: needs_human_review` therefore leaves no trace in the run record —
the same blind spot recorded in WRONG_CALLS on 2026-07-24, where absence of a
skip record was wrongly read as absence of a skip mechanism. A withheld section
must appear in one of those lists.

## Third defect: `complete` with an undeployed page

The work item finished `complete` while `pages.build_status` stayed
`needs_rebuild` and `deployed_at` remained four days old. That is
`UpdatePageStatusAction`'s partial-build guard behaving correctly
(`v3_site_actions.go:684` — refuse to mark deployed when the build is short of
its plan, set `needs_rebuild`) — but the work item's own status does not reflect
it. Reading the item alone tells you the rebuild succeeded. Belongs with the
work-item-completion-integrity workstream.

## Fix candidates

**Candidate 1 (correct the contract — pick ONE side, fleet-wide).** Either
repoint the component's four `source` paths at `site_specs.identity.contact.*`,
or make the classifier write flat keys as well as nested. Repointing the schema
is one UPDATE and cannot break the five working sites **only if** the resolver
supports the deeper path — verify that before choosing, because
`site_specs.identity.contact.email` is three levels and every working example
today is two.

**Candidate 2 (resolver-level, more general).** Teach the field resolver to try
`<aspect>.<contact|details>.<field>` when `<aspect>.<field>` misses. Fixes this
class for any component/aspect pair rather than one component.

**Candidate 3 (make the drop loud).** Record `on_missing: needs_human_review`
withholdings in `sections_skipped` with the unresolved field named, and emit the
`needs_section_data` item against the **page** so it is visible. One such item
exists on this site at `needs_human_review` — it was the only clue, and it does
not name the section.

**Candidate 4 (backfill).** Once the contract is fixed, the eight sites need
their contact pages rebuilt to pick up the block. Several also have genuinely
empty contact data (`phone: null`), so this is partly a data-gathering task, not
just a rebuild — do not present an empty block as a fix.

## What was done here (data workaround only — the contract is untouched)

Added flat `email` + `phone` to fundamentallyai.com's `identity` aspect
alongside the nested pair (both kept), superseding the previous row properly
(`is_current=false` + `superseded_at` first — `idx_site_specs_current` is UNIQUE
on `(site_id, aspect) WHERE is_current`, so inserting before superseding fails
with 23505). Backup: `bak_site_specs_fai_identity_20260725`. All six `services`
entries verified preserved. The owner's phone `+44 (0) 7934 524 911` had been
written only to `sites.phone`, which **no component reads**.

## Verification

Rebuild a contact page on a site in the "no" group and confirm three components
with the phone present in the third. Do **not** verify by checking the five
working sites — they were never affected. For candidate 1, the induced test is a
site whose identity has *only* nested keys.

---

## UPDATE 2026-07-27 (backlog sweep) — this bug and `bugs_open/111` are the same eight sites

**Contributed, not fixed.** `scripts/who-owns.py` puts this case with
`brochure_component_library` [ACTIVE]; per CLAUDE.md this is a contribution into
the case file, not a competing fix.

**Re-grounded against the live DB today** (the 07-25 figures had moved — a sixth
site has since been worked around and a fourteenth site now exists, so "8 of 13"
is now **8 of 14**):

```sql
SELECT s.domain,
       CASE WHEN ss.data ? 'contact' THEN 'nested' ELSE '' END,
       CASE WHEN (ss.data ? 'email') OR (ss.data ? 'phone') THEN 'FLAT' ELSE '' END
FROM site_specs ss JOIN sites s ON s.id = ss.site_id
WHERE ss.aspect = 'identity' AND ss.is_current ORDER BY 3 DESC, 1;
```

**All 14 sites nest `contact`.** Six carry a flat pair as well (the data
workaround — all six have `…@contactforsales.com` addresses, which is what makes
them identifiable as worked-around rather than natively flat). The remaining
**eight have nested-only keys**: `dartsonline.com`, `gamesdesign.co.uk`,
`oufe.com`, `relojistas.com`, `robot-hands.com`, `vetcomparison.uk`, `vonc.com`,
`webdesign.co.uk`.

Note this **corrects the framing in the original diagnosis above**: it is not
that some sites are written flat and some nested. The writer nests on *every*
site; the six that work were manually patched. So there is no "flat-writing"
code path to preserve, and candidate 1 (read the nested path) can be made
unconditional rather than a fallback.

### Why this matters beyond this bug

`bugs_open/111` (footer "Contact" heading renders over nothing) reports **"8 of
14 live sites"**. That is the *same eight sites*, and the mechanism connects:
111's footer chrome gates its contents on `{{if .email}}` / `{{if .phone}}` while
leaving `<h4>Contact</h4>` unconditional. Those fields are empty on exactly the
sites where this bug prevents the contact data resolving. **So 072 is the
upstream cause of most of 111.**

Consequences for whoever picks these up:
- Fixing the path contract here populates email/phone on those eight sites and
  makes 111's heading stop stranding on **seven** of them — without touching
  shared fleet chrome, which is the part 111 correctly flags as not a site
  thread's unilateral call.
- 111 does **not** become redundant. `relojistas.com` is an owner ruling of *no
  contact route at all*, so it must render no heading even with the contract
  fixed. 111 shrinks from a fleet-wide cosmetic defect to one genuine
  edge case — a much smaller and safer change.
- **Sequence them 072 → 111.** Doing 111 first fixes the symptom on eight sites
  and removes the evidence that would show 072 still broken.

[UNVERIFIED] I did not rebuild a contact page to confirm the block then renders —
that is still the verification step above, and it needs the contract fix first.

> **CORRECTED 2026-07-27, same day, by the render-cluster session — the section
> immediately above is wrong about 111, and the error is instructive.**
>
> **What is wrong:** 111 is **not** the same eight sites and is **not** mostly
> downstream of this bug. The footer does not read `site_specs.identity` at all —
> it renders from the **`sites.email` / `sites.phone` columns**, a different store,
> populated on **13 of 14** sites. Measured from the rendered footers of all 14
> live sites, 111 affects **2 sites** (gamesdesign.co.uk, relojistas.com), not 8.
> Only **one** of those is reachable by any contact-data fix, because relojistas is
> an owner ruling of *no contact route* and needs the gate regardless. So fixing
> this bug buys **at most one site** on 111, not seven.
>
> **Therefore the sequencing argument above is void.** 111 is an independent
> 2-site cosmetic fix (one `replace()` on one `content_components` row, minutes,
> reversible) and does **not** need to wait for this bug. Do not hold it.
>
> **How the error was made** (this is the part worth keeping): I joined the two
> cases on their **matching counts** — "8 of 13" here and "8 of 14" there — and
> inferred a mechanism to explain the coincidence, without ever opening the
> template to see which field it interpolates. A coincident denominator is the
> weakest possible join between two bugs; with 39 open cases on a 14-site fleet,
> "8 of 14" collides constantly. It is also seductive, because a matching count
> arrives feeling like independent corroboration while carrying almost no
> information. **Join bugs by a shared reader or write path named to file:line,
> never by arithmetic.** Logged to `WRONG_CALLS.md` under
> `writes-the-field-is-not-reads-the-field`.
>
> **What still stands from the section above** (measured directly, not inferred):
> all 14 sites nest `contact`; the six that work carry a manually-added flat pair;
> so there is no flat-writing code path to preserve and candidate 1 can be
> unconditional rather than a fallback. That finding is unaffected.
>
> **Also unresolved and worth a look:** relojistas renders an empty anchor
> *despite* a populated `sites.email`, so a third path feeds that render.
> `[UNVERIFIED]` — nobody has traced it.

---

# RESOLVED 2026-07-31 — the root cause was not the path, and it was in this file all along

**Fixed in `ef9e7e999`. Root cause CONFIRMED by the diagnosis loop, first
iteration (corr `0f76987c-0fc8-48b9-9c12-c49379560f00`). Council gate submitted as
`dd03a73b-eee1-4769-93ed-d8fe79154c19`.** Worked by the "bugfix 9" thread;
workstream docs at `docs024_key_docs_latest/bugfix_072_identity_source_resolver/`.

> **Number reminder:** this is the *contact-info / identity source* 072. The other
> `072` is component-markup-without-CSS (`bugfix_072_component_css`), closed and
> live on v1.0.1171. `scripts/who-owns.py 072` merges both cases' commits.

## The diagnosis above is correct. Both of its remedies fix ZERO of the eight sites.

Re-grounded against the live DB on 2026-07-31 (queries in the workstream RUNBOOK
§1–§2). The population has moved again: **15 real sites**, not 13 or 14
(`loancalculator.co.uk` is new and has **no `identity` aspect at all**), plus 14
`pool-*.internal` rows that must be excluded from any `sites` census.

| store | populated on |
|---|---|
| flat `identity.email` | **7** of 15 |
| nested `identity.contact.email` | **6** of 15 |
| **`sites.email` (column)** | **12** of 15 |

**The nested `contact` sub-object exists on 14 of 15 sites, and its VALUES are
null/empty on exactly the 8 sites that fail.** The 6 sites where the nested key
holds anything are the same sites that already carry the flat key — the manual
workaround wrote both.

⇒ **Candidate 1 (repoint the schema at `identity.contact.*`) and candidate 2
(a resolver nested-path fallback) each resolve on 0 of the 8 broken sites.**

**The discriminator table in the original diagnosis is a correct measurement whose
causal reading is inverted.** `contact-info` rendered on exactly the sites with a
flat `email` — no exceptions either direction — because those are simply the sites
that have contact data **at all**. `jsonb ? 'contact'` and "the classifier writes
nested" are both *shape* checks; neither sees whether the shape holds a value. A
nested object full of nulls is indistinguishable from a populated one to both.

## The actual root cause — quoted from this file

> *"The owner's phone `+44 (0) 7934 524 911` had been written only to
> `sites.phone`, which **no component reads**."*

That sentence, filed as a footnote to the data workaround, is the bug.
`sourceResolver` (`plan_sections_action.go`) resolves `site_specs.*`,
`site_assets.*`, `pages.*`, `config.*` and `query.*`. **No branch reads the `sites`
row's own identity columns** — `email`, `phone`, `contact_address`,
`company_name`, `tagline`, `logo_text`, `logo_url`. So an owner-supplied contact
detail is invisible to every component declaring a `site_specs.identity` source,
and `contact-info`'s `on_missing: needs_human_review` on `email` then withholds the
whole section.

**And this was the THIRD path to need the same fix — the other two were already
fixed for exactly this reason:**

| path | reads the columns? | how |
|---|---|---|
| full writer render | **yes** | `loadSiteDataFull`, `render_site_components_action.go:337` |
| light section rerender | **yes** | `buildRerenderBaseData`, `rerender_page_sections_action.go:590` — *"We now prefer the column … making both render paths agree"* (`bugs_open/006` §B) |
| **`plan_sections`** | **no** | ← this bug, missed both times |

The diagnosis loop confirmed it independently, citing a live `vetcomparison` row
with `sites.email` populated and every `site_specs.identity` value NULL, and
tracing `needs_human_review` → `shouldDefer` → `item.Status = "deferred"`.

## What was fixed

A bounded fallback chain in the `site_specs` branch of `resolve`, consulted **only
after the literal path misses**: (1) the writer's nested shape
(`identity.<leaf>` → `identity.contact.<leaf>`), then (2) the canonical sites row
(`identity.<leaf>` → `sites.<column>`). Enumerated, not a deep search — matching
the `site_assets` image-role alias in the same function. Registered as **PBP-026**
with its landmine and open review question, in the same commit as the seam.

**Safety property, test-asserted:** the literal path is tried first and always
wins, so **no path that resolves today changes its value**. `ensureSiteRow`
deliberately does *not* `COALESCE` across columns the way `loadSiteDataFull` does —
an empty value must stay empty, or the fallback satisfies a `needs_human_review`
field with a value nobody supplied.

**Outcome:** 5 of the 8 sites gain a resolvable contact email (oufe, robot-hands,
vetcomparison, vonc, webdesign). **3 cannot and should not** — gamesdesign,
loancalculator and relojistas have no contact fact in any store, and relojistas is
an owner ruling of *no contact route at all*, so it must keep resolving nothing.
`gamesdesign.co.uk` is the **negative control**: if it starts rendering a contact
block after the roll, the fallback is fabricating and the change is wrong.

## Corrections to the earlier updates in this file

1. **Candidate 4 (backfill the eight sites' data into `site_specs`) is now the
   wrong shape** — it duplicates a fact into a second store and guarantees drift.
   The platform has already ruled which store is canonical; the resolver reads it.
   What the 5 sites need is a **contact-page rebuild** after the roll, not a data
   migration.
2. **Candidate 3 ("make the drop loud") is ALREADY DONE and needs no work.**
   `plan_sections` always emits `sections_deferred` and `sections_skipped`
   (`:922-924`; empty-result path at `:695-697`) and `persistSectionSkips` writes
   them durably. The 2026-07-25 observation that neither key was present is stale.
3. The 2026-07-27 update's *"candidate 1 can be made unconditional rather than a
   fallback"* is correct about the writer (all sites nest; there is no
   flat-writing code path to preserve) and **irrelevant to the outcome**, because
   the nested values are empty on the sites that matter.
4. **The open `[UNVERIFIED]` question at the end of the previous section —
   "relojistas renders an empty anchor *despite* a populated `sites.email`" — is
   answered in part and its premise is now false:** `sites.email` for
   `relojistas.com` is **empty** as of 2026-07-31 (`(none)`), so there is no
   contradiction left to trace. Whether it was populated on 2026-07-27 and has
   since been cleared, I have not established — `sites` keeps no history.
   [UNMEASURED]

## Still open after this fix (not blockers, deliberately out of scope)

- **The data gap:** 3 sites have no contact detail anywhere. An owner matter, not
  a code defect.
- **`hours` resolves nowhere** and has no column. `skip_field` is correct.
- **`identity.address` → `sites.contact_address`** is the one mapping that goes
  beyond `loadSiteDataFull`'s set (no render path reads that column today;
  populated on 1 site). Flagged to the council; drop it if a seat objects.
- **The fix is inert until a chassis rebuild + roll.** Verification recipe —
  pod-grep with a positive control, then induce the failing case, then the
  negative control — in the workstream RUNBOOK §4.
