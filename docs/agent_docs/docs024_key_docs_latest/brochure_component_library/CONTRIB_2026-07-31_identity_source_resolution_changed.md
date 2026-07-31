# CONTRIB 2026-07-31 — a resolver your components depend on changed what it guarantees

**From the "bugfix 9" thread (`bugs_open/072`, identity source resolver). Not a
request — a notification, per the owner ruling of 2026-07-29 (§3): a shared
mechanism's other consumers must be TOLD, not merely measured.**

You own component `input_schema` authoring. `plan_sections`' `sourceResolver` is
what turns your `source:` declarations into values, and as of `ef9e7e999` it
resolves one class of path differently. **Nothing you have written changes value**
— but what you can rely on has changed, in both directions.

## What changed

Previously: a `site_specs.<aspect>.<leaf>` source either resolved from the
`site_specs` aspect row or it missed, and **a miss was final**.

Now: after the literal path misses, two enumerated aliases are tried, in order:

1. **the writer's nested shape** — `identity.<leaf>` → `identity.contact.<leaf>`.
   `domain-research-classifier` nests contact details; every schema in the library
   asks for them flat. Those two have never agreed.
2. **the canonical `sites` row** — `identity.<leaf>` → `sites.<column>`, for
   `email`, `phone`, `address`/`contact_address`, `company_name`, `tagline`,
   `logo_text`, `logo_url`.

Literal always wins, so this is strictly additive. Registered as **PBP-026**.

## The two things that matter for schema authoring

**1. A flat `site_specs.identity.*` path is no longer a reliable way to get
`needs_human_review` to fire.** If you have ever declared a flat identity path
*expecting* it not to resolve — to force a HITL request or to keep a field out of
a build — that field may now resolve from the sites row. Measured: `sites.email`
is populated on **12 of 15** real sites, `phone` on 7, `company_name` on 7,
`tagline` and `logo_text` on 5, `contact_address` on 1. Check any schema that
relies on a *miss*.

**2. You no longer need the flat/nested workaround, and should stop propagating
it.** Six sites carry hand-added flat `identity.email`/`phone` keys purely to make
`contact-info` render. That workaround is now unnecessary — the resolver reaches
both the nested shape and the sites row. Please don't add more; a fact duplicated
into two stores drifts, and the platform has already ruled the `sites` row
canonical (`loadSiteDataFull`, and `bugs_open/006` §B on the rerender path).

## Two findings from the same census that are yours, not mine

I measured every `site_specs.*` source path declared by an active component (100
distinct paths). Two results affect the library directly, and I have deliberately
**not** acted on either:

- **74 of the 100 paths name an aspect that exists on NO site** — `nav`,
  `navigation`, `blog`, `case_studies`, `categories`, `contact`, `inventory`,
  `legal`, `pages`, `pricing`, `product`, `search`, `social`, `social_proof`.
  Those fields can never resolve for any site. This is **already diagnosed**:
  `bugs_closed/018` calls them *"decorative — nothing resolves them"* and
  establishes that chrome components run a separate, thinner path where **the
  fallback machinery never runs at all** (unlike page sections, which do apply
  static fallbacks). My census only adds the fleet-wide scale; the fix is
  template/schema authoring, which is your lane. Recipe to reproduce the census:
  `bugfix_072_identity_source_resolver/RUNBOOK_identity_source_resolver.md` §3.
- **The vocabulary has near-duplicate aspect names** — `nav` *and* `navigation`,
  `cta` *and* `ctas`, plus `identity.contact_email` alongside `identity.email`.
  Each spelling is one component's private guess. Worth a decision about which is
  canonical before the next component is authored against a third spelling.
- **`site_specs.pricing.tiers[0].name`-style paths cannot resolve at all**, whatever
  the data holds: `navigateMap` splits on `.` and looks up each segment as a map
  key, so `tiers[0]` is treated as a literal key name. Array indexing is not a
  supported path syntax. Six such paths exist.

## If you disagree with one bit of this

`identity.address` → `sites.contact_address` is the single mapping that goes
**beyond** what the canonical render path (`loadSiteDataFull`) reads — that column
is read by no render path today and is populated on 1 site. It is flagged in the
council submission (`dd03a73b`) as droppable. Say so and I will drop it.

## Verification, when the chassis next rolls

`bugfix_072_identity_source_resolver/RUNBOOK_identity_source_resolver.md` §4. The
short version: pod-grep for `resolved from the canonical sites row` **with a
positive control in the same exec**, then rebuild a contact page on vonc.com and
expect three components, then check `gamesdesign.co.uk` still renders **no**
contact block — it has no contact fact in any store and must keep resolving
nothing.
