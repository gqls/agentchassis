# DECISION — remake the hosted sites: 22 free, 3 protected

**Owner ruling, 2026-08-20:** *"we can completely start again and overwrite all of them except for
leopardess.co.uk and leopardess.uk."*
**Amended the same day:** *"cartoon.co.uk is also off limits for now I have ideas for that one."*

So of the 25 domains found serving real content on the estate's own hosting
(`HOSTED_domains_for_owner_decision.md` group A), **22 are free to rebuild from scratch** and
**3 are protected.**

---

## 1. PROTECTED — do not touch

| domain | why |
|---|---|
| `leopardess.co.uk` | owner ruling; also an active lane in this repo (`fleet_copy_quality`, `leopardessconsulting`) |
| `leopardess.uk` | same site, same ruling |
| `cartoon.co.uk` | owner ruling 2026-08-20: *"off limits for now I have ideas for that one."* **The owner has a plan; do not pre-empt it.** Worth noting it is also the largest single page in the batch — a 133 kB homepage on only 4 sitemap pages — so whatever is there is substantial and concentrated on the front page. |

## 2. FREE TO OVERWRITE — 22

Page counts from each site's own sitemap where it has one, otherwise no sitemap and no internal
links from the homepage, i.e. a single-pager.

**Above ten pages (2):** `businessinsurancequotation.co.uk` (22) · `aiartgallery.uk` (12)

**Four to ten pages (11):** `minisitemaker.co.uk` (10) · `websitepromotion.co.uk` (6) ·
`seduce.co.uk` (5) · `seotools.co.uk` (5) · `personalgift.co.uk` (4) ·
`designblog.co.uk` (4) · `uniquedirectory.co.uk` (4) · `writesy.uk` (4) ·
`fatherchristmas.uk` (4) · `santaclaus.uk` (4) · `pelletburners.co.uk` (4)
*(`cartoon.co.uk` was here until the owner protected it — see §1.)*

**Single-pagers (9):** `businesschristmasgifts.co.uk` · `businesschristmasgifts.uk` ·
`fridge-magnets.co.uk` · `advertise.co.uk` · `conferences.co.uk` · `copyonline.co.uk` ·
`vinrose.uk` · `dsgn.co.uk` · `catalogues.co.uk`

### The one thing "overwrite all of them" does not settle: the twin pairs

Three of the 23 are pairs serving one site on two names:

- `businesschristmasgifts.co.uk` + `.uk`
- `fatherchristmas.uk` + `santaclaus.uk`

**A remake has to decide what the second name does**, because two rebuilt sites on the same
proposition compete with each other in the same search results — the exact failure the positioning
register exists to prevent, and the register's existing rule for this is a per-pair owner call
(301, or an accepted duplicate, marked `⚑OWNER`). `fatherchristmas.uk`/`santaclaus.uk` is the
interesting one: unlike a spelling variant, those are two genuinely different search phrases for
the same subject, so they may deserve two propositions rather than a redirect.

## 3. SALVAGED BEFORE OVERWRITING — `vinrose.uk`

Owner: *"you might want to use the images from vinrose.uk as they're nice, that's if we still go
with wine."*

**Saved 2026-08-20**, before anything can overwrite them, to
`portfolio_positioning/salvage/vinrose.uk/`:

| file | size | verified |
|---|---|---|
| `hero-wine-dining.jpg` | 171,935 B | JPEG, **1920×1080** |
| `wine-dining-bg.jpg` | 54,806 B | JPEG, 512×279 |
| `index.html` | 11,443 B | the page they came from, for context |

Verified as real image data with `file`, not error pages returned at 200 — a parked or broken host
answers 200 with HTML, and a saved "image" that is actually an error page looks identical in a
directory listing.

**The condition attached to reuse — *"if we still go with wine"* — is a live one**, and the
salvage does not settle it. `vinrose.uk` has no register entry, so the proposition is open. If the
domain goes somewhere other than wine, these images are the wrong images and should not be reached
for just because they exist.

⚠ **Only two images were referenced by the homepage.** If the site has more on inner pages, this
salvage missed them — but it is a single-pager with no sitemap and no internal links, so there is
probably nothing else. Anyone wanting certainty should look at the hosting account, not the HTML.

## 4. `businessinsurancequotation.co.uk` — the owner's approach is worth keeping

Owner: *"We can add businessinsurancequotation.co.uk to the insurance list, but what I did with it
was to use sentences/quotations from business insurance that had been claimed in the past that had
interesting stories around them to garner interest and traffic without treading on any
regulation's toes. We could try that or if not it doesn't worry me either way."*

**This is a good idea and it deserves recording even if this particular site does not use it**,
because it solves a problem the whole finance portfolio has.

**What makes it work.** A regulated vertical is hard to write about: the interesting material
(rates, recommendations, "we can get you covered") is exactly what a non-authorised site may not
publish, and what is left tends to be thin definitional content. **A real past claim is a
narrative fact, not a financial promotion.** It is inherently interesting, it is specific, it
demonstrates genuine subject knowledge — and stating what happened to somebody else advises
nobody, arranges nothing and promotes no product. It sidesteps the constraint rather than fighting
it.

**How it fits what the platform now has.** This is exactly the shape the claims layer is built for
and it should go through it rather than around it:

- each story is an `evidence_base` **fact** with a citation — a court report, a trade-press piece,
  an ombudsman decision, an insurer's own case study. The verbatim-quote discipline the finance
  directory already uses applies unchanged: the model proposes, the string comparison disposes;
- it stays clear of the `banned_claims` and the regulated-identity guard by construction, because
  the site is narrating history rather than describing itself or recommending anything;
- **the failure mode to watch is a fabricated or embellished story**, which would be worse than a
  fabricated statistic because it is more memorable and more repeatable. A story without a
  citation must not be publishable, and that is a job for the evidence base rather than a
  reviewer's attention.

**Not decided.** The owner is explicitly neutral (*"it doesn't worry me either way"*). Recorded so
the option is available to whoever writes the brief, and so the technique is not lost if this
domain goes another way — it would suit `interestrates.co.uk`, the mortgage family and the
health-insurance cluster equally well.

## 5. Sequencing — what to build first, and why not the biggest

The tempting first pick is `businessinsurancequotation.co.uk`: 22 pages, the most substantial, and
the owner has just handed over a content angle for it. **It is the wrong first pick.** It is
insurance, so a rebuild inherits the entire compliance layer — evidence base, banned claims, and
the regulated-identity guard that only went live yesterday and **has not yet refused anything in
production**. Exercising untested compliance machinery on the largest site in the batch puts the
two hardest variables together.

**Better first candidates: the single-pagers with strong generic names** —
`advertise.co.uk`, `conferences.co.uk`, `catalogues.co.uk`, `copyonline.co.uk`,
`fridge-magnets.co.uk`. Nothing to lose, unambiguous subjects, no regulated angle, and each gives
the brief writer a genuinely different content shape to work with. Do the insurance one once the
guard has refused something and been seen to.

## 6. Related rulings from the same message

**All future sites must have sitemaps** — recorded against register **SEO-002**, whose own
`verify-later` asked exactly this question and is now answered. The generator exists
(`scripts/site-discovery-files.py`) and is registered; what does not exist is anything that runs
it. **Measured 2026-08-20: 15 of the 25 live sites serve no `/sitemap.xml`, including the pilot
built four days ago with every current guard applied** — which is the clearest available statement
that a manual step is not a mechanism.
