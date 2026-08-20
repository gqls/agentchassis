# The 62 hosted domains — for the owner to decide on, one by one

**Owner ask, 2026-08-20:** *"Please list the hosted domains like cartoon.co.uk so I can decide
manually about them."*

These are the domains whose nameservers are `dns.uk-noc.com` / `dns.us-noc.com` — real hosting,
not a parking marketplace. Everything else in the estate is parked (1,247), never delegated
(207), at a registrar default (19), ours on Cloudflare (14) or genuinely unclassified (18).

**Each one was fetched, 2026-08-20**, and they fall into three groups that need different
decisions. The grouping is by what a visitor actually gets, not by what the DNS says — a domain
can be on real hosting and serve nothing, and 24 of these are exactly that.

---

## A. 25 SERVING A REAL SITE — these have content to lose

Titles suggest WordPress (the `&#8211;` en-dashes are its title filter). **Anything done to these
is a migration, not a build**, and the existing content is the thing at risk.

| domain | bytes | title |
|---|---|---|
| `cartoon.co.uk` | 132,965 | Cartoon – A view on cartoons |
| `minisitemaker.co.uk` | 78,656 | Minisite Maker – A creative way to instantly … |
| `businessinsurancequotation.co.uk` | 78,303 | A quotation for business insurance |
| `aiartgallery.uk` | 74,443 | Home – The AI Art Gallery |
| `businesschristmasgifts.co.uk` | 51,664 | Business Christmas Gifts |
| `businesschristmasgifts.uk` | 51,624 | Business Christmas Gifts *(same site, twin)* |
| `personalgift.co.uk` | 49,083 | Personal Gift .co.uk – Really great personalised … |
| `designblog.co.uk` | 38,512 | Web Design Blog – My Web Design Blog |
| `uniquedirectory.co.uk` | 33,089 | Unique Directory – A comprehensive web directory |
| `writesy.uk` | 33,058 | Writesy – The New Name In Content Writing |
| `fatherchristmas.uk` | 25,262 | Merry Christmas – … from Father Christmas |
| `santaclaus.uk` | 25,231 | Merry Christmas *(same site as fatherchristmas.uk)* |
| `pelletburners.co.uk` | 24,273 | Pellet Burners – A great way to heat your house |
| `seduce.co.uk` | 24,001 | *(no title)* |
| `fridge-magnets.co.uk` | 20,604 | Fridge Magnets |
| `advertise.co.uk` | 19,956 | Advertise |
| `conferences.co.uk` | 17,391 | Conferences |
| `copyonline.co.uk` | 16,708 | Copywriters |
| `vinrose.uk` | 11,443 | Vinrose – Elevated Tastes |
| `dsgn.co.uk` | 10,554 | Welcome to dsgn |
| `websitepromotion.co.uk` | 8,058 | Website Promotion from start to profit |
| `catalogues.co.uk` | 7,811 | *(no title)* |
| `seotools.co.uk` | 7,741 | List of cool tools for website promotion |
| `leopardess.co.uk` | 6,853 | Leopardess — an AI-agent-first development company |
| `leopardess.uk` | 6,853 | Leopardess *(same site, twin)* |

**Note the three twin pairs** — `businesschristmasgifts` .co.uk/.uk, `fatherchristmas`/`santaclaus`,
`leopardess` .co.uk/.uk — each serving the same site on both names. `leopardess.*` is an active
lane in this repo (`fleet_copy_quality`, `leopardessconsulting`), so it is not idle.

## B. 13 RESPONDING BUT EMPTY — hosting stubs, not sites

Nothing to lose here. Several are plainly infrastructure rather than sites (the `email-*` group,
`websy.uk`), and three serve a bare Apache directory listing.

| domain | bytes | what came back |
|---|---|---|
| `agentcoordinator.uk` | 1,371 | `Index of /` |
| `email-account.co.uk` | 1,371 | `Index of /` |
| `email-account.uk` | 1,368 | `Index of /` |
| `vectordb.uk` | 1,363 | `Index of /` |
| `designconsultancy.co.uk` | 537 | Web Development |
| `5un.co.uk` | 317 | 5un Management |
| `onpointcopy.co.uk` | 95 | *(empty)* |
| `emailsecurity.uk` | 86 | *(empty)* |
| `websy.uk` | 52 | *(empty)* |
| `2v.uk` | 2 | *(empty)* |
| `wpx.uk` | 2 | *(empty)* |
| `managementemail.co.uk` | 0 | *(empty)* |
| `managementemail.uk` | 0 | *(empty)* |

⚠ `designconsultancy.co.uk` is the registrant email domain on all 1,567 registry rows, and the
`email-*` / `managementemail.*` / `emailsecurity.uk` group looks like mail infrastructure.
**Do not treat these as free to rebuild** without checking what depends on them.

## C. 24 ON HOSTING BUT SERVING NOTHING — the ones worth the most attention

Real hosting configured, no response at all. **These are strong generic names sitting idle**, and
they are the readiest candidates for a framework build — no content to migrate, no visitor to
disrupt.

`gardens.co.uk` · `gardens.uk` · `healthcare.uk` · `felines.co.uk` · `felines.uk` ·
`interestrates.co.uk` · `comparehealthcare.uk` · `insurely.uk` · `insurancedomain.co.uk` ·
`workdomain.co.uk` · `personae.uk` · `eborg.uk` · `ct4m.co.uk` · `5s.uk` ·
`soyrocks.co.uk` *(503)* · `soyrocks.uk` · `oliverappleby.co.uk` ·
and the health-insurance cluster: `healthinsurancequotation.co.uk` / `.uk` ·
`healthinsurancequote.uk` · `healthinsurancerate.co.uk` / `.uk` ·
`healthinsurancerates.co.uk` / `.uk`

**Two things to flag before anything is built on these:**
1. **The seven `healthinsurance*` names and `comparehealthcare.uk` are already register entries**
   (family I, health insurance) — so they are positioned, not free. They are also the domains
   whose earlier "no nameservers" reading I got wrong; they are on hosting, just not serving.
2. **`interestrates.co.uk` is finance** and would inherit the compliance machinery — evidence
   base, banned claims, and now the regulated-identity guard.

---

## Method, and what this list does not tell you

Fetched `https://<domain>/` following redirects, 15s timeout, reading the **body size** and not
just the status — a parked domain answers 200 on every path, so status alone cannot discriminate.
`http=000` means no response at all (connection or TLS failure), which is not the same as a 404.

**It does not tell you what is on the hosting account**, only what the front page returns. A
domain in group C could still be serving mail, a subdomain, or an application on a path. Check
the hosting control panel before assuming a name in group C is idle.

Regenerate with:

```sh
python3 scripts/domains/classify_nameservers.py --ns-csv <registry-export.csv> > estate.tsv
awk -F'\t' 'NR>1 && $2=="CLOOK"{print $1}' estate.tsv   # then fetch each
```

---

## SIZED FOR REMAKING — added 2026-08-20 at the owner's request

*"I think we can remake a lot of them."* So the question stopped being *which are live* and became *how much is there to lose*. Page counts below are from each site's own sitemap, following the WordPress sitemap INDEX down to the actual URL lists — the index itself lists 4 sub-sitemaps on most of these, and reading that as "4 pages" was my first and wrong answer.

**The headline: they are all small.** The biggest is 22 pages, the median is 4, and eleven have no sitemap at all and no internal links from the homepage — single-pagers. There is very little content here to migrate.

| domain | pages | homepage bytes | title |
|---|---|---|---|
| `businessinsurancequotation.co.uk` | 22 | 78,303 | A quotation for business insurance &#8211; B |
| `aiartgallery.uk` | 12 | 74,443 | Home - The AI Art Gallery |
| `minisitemaker.co.uk` | 10 | 78,656 | Minisite Maker &#8211; A creative way to ins |
| `websitepromotion.co.uk` | 6 | 8,058 | Website Promotion from start to profit |
| `seduce.co.uk` | 5 | 24,001 | (no title) |
| `seotools.co.uk` | 5 | 7,741 | List of cool tools for website promotion |
| `cartoon.co.uk` | 4 | 132,965 | Cartoon &#8211; A view on cartoons |
| `personalgift.co.uk` | 4 | 49,083 | Personal Gift .co.uk &#8211; Really great pe |
| `designblog.co.uk` | 4 | 38,512 | Web Design Blog &#8211; My Web Design Blog |
| `uniquedirectory.co.uk` | 4 | 33,089 | Unique Directory &#8211; A comprehensive web |
| `writesy.uk` | 4 | 33,058 | Writesy &#8211; The New Name In Content Writ |
| `fatherchristmas.uk` | 4 | 25,262 | Merry Christmas &#8211; Merry Christmas from |
| `santaclaus.uk` | 4 | 25,231 | Merry Christmas &#8211; Merry Christmas from |
| `pelletburners.co.uk` | 4 | 24,273 | Pellet Burners &#8211; A great way to heat y |
| `businesschristmasgifts.co.uk` | 1 (no sitemap) | 51,664 | Business Christmas Gifts &#8211; Christmas G |
| `businesschristmasgifts.uk` | 1 (no sitemap) | 51,624 | Business Christmas Gifts &#8211; Christmas G |
| `fridge-magnets.co.uk` | 1 (no sitemap) | 20,604 | Fridge Magnets |
| `advertise.co.uk` | 1 (no sitemap) | 19,956 | Advertise |
| `conferences.co.uk` | 1 (no sitemap) | 17,391 | Conferences |
| `copyonline.co.uk` | 1 (no sitemap) | 16,708 | Copywriters |
| `vinrose.uk` | 1 (no sitemap) | 11,443 | Vinrose - Elevated Tastes |
| `dsgn.co.uk` | 1 (no sitemap) | 10,554 | Welcome to dsgn | dsgn |
| `catalogues.co.uk` | 1 (no sitemap) | 7,811 | (no title) |
| `leopardess.co.uk` | 1 (no sitemap) | 6,853 | Leopardess — an AI-agent-first development c |
| `leopardess.uk` | 1 (no sitemap) | 6,853 | Leopardess — an AI-agent-first development c |

### What the sizes mean for a remake

- **Nothing here is a big migration.** `businessinsurancequotation.co.uk` (22 pages) and
  `aiartgallery.uk` (12) are the only two above ten. Everything else is four pages or a single page.
- **The eleven single-pagers are the easiest wins** — `advertise.co.uk`, `conferences.co.uk`,
  `catalogues.co.uk`, `dsgn.co.uk`, `fridge-magnets.co.uk`, `copyonline.co.uk`, `vinrose.uk`,
  `businesschristmasgifts.co.uk`/`.uk`, `leopardess.co.uk`/`.uk`. A framework build would be a
  large step up from a one-page site on any of them.
- **⚠ Three twin pairs serve the SAME site on both names** — `businesschristmasgifts` .co.uk/.uk,
  `fatherchristmas.uk`/`santaclaus.uk`, `leopardess` .co.uk/.uk. Remaking one of a pair without
  deciding what the other does leaves two sites competing for the same search results, which is
  the exact thing the positioning register exists to prevent.
- **⚠ `leopardess.co.uk` / `.uk` are an ACTIVE lane in this repo** (`fleet_copy_quality`,
  `leopardessconsulting`). Not idle — check with that lane first.
- **⚠ `businessinsurancequotation.co.uk` is insurance**, so a remake inherits the whole compliance
  layer: evidence base, banned claims, and the regulated-identity guard now live in CGV-033. It is
  also the largest of them, so it is the *least* attractive first candidate despite looking the
  most substantial.

### Two honest limits on these numbers

1. **A sitemap is what the site claims, not what it serves.** A page absent from the sitemap is
   invisible to this count, and a stale sitemap can list pages that 404.
2. **"No sitemap and no internal links" is evidence of a single-pager, not proof.** A site whose
   navigation is rendered by JavaScript would look identical to this check, because it reads the
   HTML as delivered and does not execute anything.

