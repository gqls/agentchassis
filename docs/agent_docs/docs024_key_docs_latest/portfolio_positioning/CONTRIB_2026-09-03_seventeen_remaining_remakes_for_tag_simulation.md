# CONTRIB — the 17 remaining remakes, as a FORWARD-LOOKING population for tag simulation

**For:** the `bugs_open/445` lane, at their request (2026-09-03). They are choosing an archetype's
tag set by simulating candidates against the live fleet, and named the risk themselves:
*"retrospective fitting to 33 built sites is exactly how I would produce a tag set that looks great
and helps nobody."* These 17 sites do not exist yet, so they cannot be fitted to — which is the
whole value of the list.

**Provenance:** `DECISION_2026-08-20_remake_the_hosted_sites.md` §2 (the owner-cleared 22), minus
the four built 2026-09-02 (advertise, websitepromotion, seotools, designblog) and copyonline.co.uk
(№5, brief with the owner now). Old-site shape is from the 2026-08-20 census; the *proposition*
column is this lane's positioning intent, not a classifier output — **the point of the exercise is
that nobody knows what tags the classifier will emit for these, which is exactly the condition a
tag set must survive.**

⚠ Twin pairs are marked ⚑OWNER: their propositions are undecided by ruling, so treat each pair as
one row with two possible splits, not two independent sites.

| domain | old shape | intended proposition (this lane) | plausible vertical |
|---|---|---|---|
| `fridge-magnets.co.uk` | single-pager, merchandise RSS aggregator | promotional/branded merchandise buyer's guide | retail / promotional products |
| `conferences.co.uk` | single-pager, conference-feed aggregator | UK conference and event discovery + organiser guidance | events / B2B |
| `catalogues.co.uk` | single-pager, 2007 AdSense home-shopping page | home-shopping and catalogue retail guide | retail / consumer |
| `dsgn.co.uk` | empty shell (menu said "Experienced Copywriter") | DESIGN-side, undecided — register stub `CW2`, must not collide with designblog (editorial) or webdesign (service) | design |
| `vinrose.uk` | single-pager (salvaged before overwrite) | undecided | undecided |
| `minisitemaker.co.uk` | 10 pages | small-site building / landing-page tooling | web tooling |
| `seduce.co.uk` | 5 pages | undecided ⚑ sensitive-vertical review owed | undecided |
| `personalgift.co.uk` | 4 pages | personalised gifting guidance + finder tools | retail / gifting |
| `uniquedirectory.co.uk` | 4 pages | directory-of-directories ⚠ listing-page class, needs 444 producer first | directory |
| `writesy.uk` | 4 pages | writing/creative TOOLING (writer's instrument) — twins by craft with copyonline (commercial copy) | writing tools |
| `pelletburners.co.uk` | 4 pages | biomass/pellet heating buyer's guide | home / energy |
| `aiartgallery.uk` | 12 pages | AI-generated art showcase ⚠ showcase = listing class, 444 producer required | creative / AI |
| `businessinsurancequotation.co.uk` | 22 pages | business insurance guidance — deliberately LAST (regulated) | insurance / finance |
| `businesschristmasgifts.co.uk` + `.uk` | twin single-pagers, one site | ⚑OWNER — register `G1` (large employer) / `G2` (SME) | gifting / seasonal B2B |
| `fatherchristmas.uk` + `santaclaus.uk` | twin 4-pagers, one site | ⚑OWNER — register `G3` (English tradition) / `G4` (American register) | seasonal / consumer |

## What this population tests that the built 33 cannot

1. **Seasonal and consumer retail** (gifting ×4, catalogues, fridge-magnets, personalgift) — the
   built fleet is finance, tools and editorial. A tag set tuned on the 33 has never met this shape.
2. **Events** (`conferences.co.uk`) — no built analogue at all.
3. **Home/trade** (`pelletburners`) — nearest built neighbour is homegarden, itself one of the
   seven single-tag `magazine-grid` cluster members.
4. **Two shapes that are listing-class by nature** (`uniquedirectory`, `aiartgallery`) and are held
   until 444's producer exists — so if a proposed tag set only works for prose-and-tools sites,
   these are where it fails.
5. **Two twin PAIRS** whose whole purpose is to be differentiated from each other on near-identical
   subject matter — the hardest case for any tag vocabulary, and the one RFC_037 exists for.

## The ask back

If a candidate tag set cannot distinguish `fatherchristmas.uk` from `santaclaus.uk`, or place
`conferences.co.uk` anywhere but a default, that is worth knowing before the archetype ships rather
than after seventeen briefs. This lane will send each site's actual `classification.industry_tags`
as it is built, starting with copyonline — a growing forward-looking sample rather than a fixed one.
