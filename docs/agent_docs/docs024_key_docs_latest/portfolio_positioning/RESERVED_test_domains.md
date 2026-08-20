# RESERVED — register-free test domains

**Owner instruction, 2026-08-19:** *"also leave me a bunch of domains (50?) unregistered as test
domains that I can run without the register."*

**This file is the record that a domain listed here is DELIBERATELY absent from the positioning
register.** That distinction is the whole point of the file and it is load-bearing:

> Once the registry covers the whole estate (owner ruling 2026-08-19, `REGISTER_positioning.md`),
> a domain with **no entry** normally means *nobody has got to it yet*. Without this file, the
> next person filling gaps writes an entry for a reserved domain, believes they are being
> helpful, and silently destroys the control. **Absent-because-reserved and
> absent-because-pending must not look alike.**

## The rule

- A domain listed below **must not** be given a positioning-register entry, a proposition, an
  audience/mode coordinate, or a neighbour rule.
- It is exempt from the collision invariant (`no two entries share family × audience × mode`),
  because it has no coordinate to collide with.
- When the register becomes a database, this list becomes a **state on the row**
  (`reserved_test`), not a separate table — otherwise it is a second hand-maintained roster and
  drifts, which is the failure this estate keeps paying for.
- Anything built on a reserved domain is a **test**. It carries no positioning promise, and it
  must not be linked to from a positioned site.

## Status: 50 CANDIDATES PICKED 2026-08-20, awaiting the owner's yes

The registry export (`domain-list-2026-07-30.csv`, 1,567 domains, all registered, none
suspended) arrived and the picker ran against it: **1,120 eligible, 447 excluded, 50 picked.**

They are **candidates, not reserved** until the owner confirms — reserving a domain is a
decision about an asset, not a script's output.

### The 50

- `00.org.uk`
- `aiq.uk`
- `artworker.uk`
- `backingtracks.uk`
- `bodybeauty.uk`
- `businessinsurancequotation.uk`
- `carsafes.uk`
- `chocolatedirect.co.uk`
- `comedienne.uk`
- `contentresearch.co.uk`
- `cosmeticsdesigners.co.uk`
- `datingevenings.co.uk`
- `designmail.co.uk`
- `dishcloths.co.uk`
- `egret.co.uk`
- `farmerinsurance.uk`
- `flubbert.co.uk`
- `gamerooms.co.uk`
- `giftsxmas.co.uk`
- `hcare.uk`
- `healthscare.uk`
- `homegarden.uk`
- `indoorplanters.co.uk`
- `interactivebrochure.uk`
- `jewellerydisplays.co.uk`
- `landingpagedesign.co.uk`
- `linkchecking.co.uk`
- `mailinglist.co.uk`
- `medicalcare.uk`
- `mobileux.uk`
- `myhealthcareblog.uk`
- `one-to-one.co.uk`
- `packagingmachinery.uk`
- `personalising.uk`
- `playground-markings.co.uk`
- `privatehealthcareproviders.co.uk`
- `progressed.uk`
- `quido.co.uk`
- `removalscompanies.co.uk`
- `rideons.uk`
- `segmentation.uk`
- `siterepair.uk`
- `soundcameras.co.uk`
- `structuralinsulatedpanels.co.uk`
- `teamhelp.uk`
- `topquality.uk`
- `vanitybags.co.uk`
- `vetsy.uk`
- `webcook.co.uk`
- `websitedesignaward.co.uk`

### How they were chosen, and the two defects fixed on the way

Rules enforced: not named in the register · not a near-variant of a registered domain
(normalised label comparison, so `savings-rates` is caught as a variant of `savingsrates`) ·
parked or registrar-default nameservers, never a serving site · not a person's name.

**Defect 1 — alphabetical clustering.** The first run took the first 50 eligible domains
passing a per-stem cap. Because the list is sorted, that returned 50 domains beginning with a
or b. Subject variety survived by luck. Replaced with a deterministic STRIDE across the whole
eligible list, so the same inventory always yields the same set — which matters because this
list gets written down and referred to later.

**Defect 2 — it picked a person's domain.** `anne-marie.co.uk` was in the first set. The picker
now takes the name-extractor's output and excludes `NAME` verdicts: a personal domain is a poor
test domain and may be someone's actual site.

The resulting spread is genuinely varied — audio, beauty, insurance, chocolate, dating,
farming, gaming, gardens, jewellery, mailing lists, medical, packaging, playgrounds, removals,
vets, web design — which is what the set is for: exercising the brief-writer across different
content shapes rather than fifty variations of one.

## Superseded status note (kept so the correction is legible): NONE RESERVED YET

**Measured 2026-08-19:** all **152** domains in `PORTFOLIO_domains.txt` are named in
`REGISTER_positioning.md`, as a primary, a twin, or inside an entry's `domains:` list.
`scripts/domains/pick_test_domains.py` run against that inventory returns **0 eligible, 152
excluded**.

So the 50 cannot come from the finance portfolio without releasing propositions the owner has
already reasoned about. Worse, most of the parked 124 are **twins of propositions we serve** —
`savings-rates.co.uk` beside the live `savingsrates.co.uk`. An unregistered test build on a twin
competes in the same search results with the real site, which is precisely what the register
exists to prevent. **A finance twin is the worst possible test domain, not the most convenient
one.**

**They should come from the wider estate** — the ~2,000 `.uk` domains that have never been
inventoried. That inventory is the blocker (see
`RUNBOOK_domain_inventory_and_classification.md` §1; it needs one owner-run command because a
session cannot pipe a credentials file).

## How to pick them, once the inventory exists

```sh
python3 scripts/domains/classify_nameservers.py all_domains.txt > classified.tsv
python3 scripts/domains/pick_test_domains.py \
    --inventory all_domains.txt --classified classified.tsv -n 50
```

The picker enforces four rules, and the fourth is the one a human picking by eye gets wrong:

1. **owned and resolving** — `NXDOMAIN` may mean the registration lapsed;
2. **parked, not live** — never build over a serving site;
3. **not named in the register** — as primary, twin, or anywhere else;
4. **not a near-variant of a registered domain** — it compares normalised labels with
   punctuation and TLD stripped, so `savings-rates.co.uk` is caught as a variant of
   `savingsrates.uk`. To a search engine and to a reader they are the same phrase.

It also spreads the pick across different word-stems, so the reserved set is not 50 domains
about one subject — **a single-vertical test set only ever tests one shape of brief**, and the
point of these domains is to exercise the auto-brief writer across news, directory, tool,
editorial and research shapes.

It will **refuse to make up a shortfall** by releasing registered domains. If the estate cannot
yield 50, the honest output is fewer, and releasing a registered one is a per-domain decision
for the owner.

## What they are FOR — worth stating, because it shapes the selection

- the **control population** for RFC_037: the register-fed classifier must be provably inert for
  a domain with no entry, and that needs domains that genuinely have none;
- exercising the **auto-brief writer** with no positioning input at all;
- rehearsing **third-party briefs** (webdesign.uk customers arrive with no register entry by
  definition, so this set is the closest thing to a rehearsal for that path);
- anything experimental that must not pollute the positioned portfolio.

## Reserved set

*(empty — pending the domain inventory)*

| domain | reserved on | notes |
|---|---|---|
| | | |
