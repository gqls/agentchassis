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
