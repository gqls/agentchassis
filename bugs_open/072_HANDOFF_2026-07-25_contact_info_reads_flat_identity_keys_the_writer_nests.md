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
