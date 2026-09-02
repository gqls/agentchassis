# 439 — adoption carries the SOURCE site's brand name verbatim into a destination on a DIFFERENT domain, and every later planner pass carries it forward

**Filed 2026-09-02** by the `gamedesign.uk` lane, at the owner's instruction ("tell
gamesdesign.co.uk lane to stop using our name") and the positioning lane's request that the
CLASS be filed once by whoever held the measurement. **Status: OPEN, UNOWNED.** The instance is
being remediated by the dedicated `gamesdesign.co.uk` session; **this file proposes no change to
that site's rows** — it is about the seam that produced them.

⚠ 438 was taken while this was being written; resolve by slug.

## 1. The one-paragraph version

The adoption agent's extraction prompt asks the model for `"company_name": "extracted
company/brand name"` — the brand of the site it CRAWLED. `apply_adoption_plan` then writes that
as the DESTINATION site's `identity` spec. Nothing compares `company_name` with
`destination_domain`, so when the two differ the destination is born wearing the source's
brand. Every later writer of `identity` (measured: `content-gap-planner`, two months on)
deep-merges over what is there and carries the name forward. The result is two live domains
with one brand — which mattered little while the source was an empty shell, and matters now
that the owner is rebuilding the source as a distinct site.

## 2. The instance, measured 2026-09-02

`gamesdesign.co.uk` (site `e33263f4-74f8-494f-b191-546845dbbddf`), adopted from
`gamedesign.uk` 2026-06-05:

```
identity.company_name  = 'GameDesign.uk'
identity.about_summary = 'GameDesign.uk is a browser-based utility platform…'
identity.adopted_from  = 'gamedesign.uk'         <- a fact, correct, must survive any fix
writer                 = content-gap-planner, 2026-08-17   <- NOT the adoption; a later pass
sites.company_name / logo_text / tagline = ''    <- the spec is the only source
```
**23 of 49 active page titles** end "| GameDesign.uk" / "- GameDesign.uk" (e.g. "About |
GameDesign.uk", "Jelly Invaders - GameDesign.uk"). Positioning lane's recount at ~17:45Z: **four**
current specs carry the exact string (`briefing`, `design_intent`, `identity`, `tools`) — two
that carried it at my first count were superseded in between, so re-count before acting.

## 3. The population — small, and about to grow

```sql
SELECT s.domain, sp.data->>'adopted_from', sp.data->>'company_name'
FROM site_specs sp JOIN sites s ON s.id=sp.site_id
WHERE sp.aspect='identity' AND sp.is_current AND sp.data ? 'adopted_from'
  AND sp.data->>'adopted_from' <> s.domain;
```
**1 row as of 2026-09-02** — this one — and it carries the source brand. So 1 of 1. The
population is small because cross-domain adoption only became possible on 2026-04-21 (before
that the trigger took ONE domain, source = destination — the mechanism behind `bugs_open/432`
§3). **It will grow by design:** the estate's cross-TLD twin rule (positioning P5, executed
2026-08-01: `.co.uk` = authority seat, `.uk` = instrument seat) exists to make sibling pairs,
and adopting one into the other across the TLD is exactly this path. Every pair built that way
inherits this defect until the seam changes.

## 4. Where it lives

- The `site-adoption-agent` prompt (live definition, `templates_db`): `"company_name":
  "extracted company/brand name"` — no mention that the destination may differ, and
  `destination_domain` is not in the prompt's inputs.
- `apply_adoption_plan_action.go` seeds `identity` for the destination from that extraction
  (`specAspects[...]`), with no reconciliation against `destination_domain`.
- Downstream writers of `identity` (`content-gap-planner` here; `sync_site_identity_action.go`
  is another) deep-merge and preserve `company_name` — correct behaviour for a key they do not
  own, which is why the defect persists rather than heals.

## 5. Fix candidates, ordered by what closes the door

1. **At `apply_adoption_plan`, when `destination_domain ≠ source domain`, do not carry the
   brand fields verbatim (closes it).** Set `identity.company_name` from the destination domain
   (domain-as-brand is the estate convention — positioning's recommendation for this instance is
   exactly "GamesDesign.co.uk"), drop or flag `tagline`/`about_summary` for the classifier to
   re-derive, and keep `adopted_from` as the fact it is. Four lines and a test with a
   cross-domain fixture. Same-domain adoption (`locked` fidelity, ADO-037) is unaffected: there
   the brand IS the destination's.
2. **Give the prompt the destination.** Add `destination_domain` to the extraction inputs and
   tell the model the brand belongs to the destination. Cheaper, but it moves the guarantee into
   prose — a doc comment is not a control (memory: `a-doc-comment-is-not-an-enforcement-mechanism`).
   Do it as well as (1), not instead.
3. **A discovery check** — `identity.company_name` containing the `adopted_from` stem and not the
   site's own — would have found this in June. Detects, does not prevent; the §3 query is its
   predicate. Not a substitute for (1).

## 6. How to verify a fix

Cross-domain fixture: adopt a site with a distinctive brand into a different domain and assert
`identity.company_name` names the destination and `adopted_from` names the source. Then the §3
query fleet-wide: **the count of rows where `company_name` matches the `adopted_from` stem must
be 0** — and it must be measured on a NEW cross-domain adoption after the fix, not on this
instance after its manual rename, or the pass proves the rename and not the seam.

Not this lane's to fix — filed because the measurement was mine. Cross-references:
`bugs_open/432` §3 (why source = destination used to be the only mode), the positioning
lane's GD1/GD2 rows, `docs024_key_docs_latest/gamedesign_uk_rebuild/NOTES` 2026-09-02 entries.
