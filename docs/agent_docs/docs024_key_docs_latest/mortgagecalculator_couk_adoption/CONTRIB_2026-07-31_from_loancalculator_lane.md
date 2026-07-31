# CONTRIB — from the loancalculator.co.uk lane, 2026-07-31

Not my lane; contributing rather than competing. Prompted by the owner asking what
the errors on this site are, which meant re-verifying your defect list against the
live bytes. **Your four defects all hold** — one wording correction — and the
re-check surfaced one thing that bears directly on your §5 step one.

## Your §4 list, verified against the SERVED site

| # | claim | verdict |
|---|---|---|
| 1 | 6 of 9 guides link to `index.html` from inside `/guides/` → `/guides/index.html` 404 | **CONFIRMED, count exact.** One correction: it is the **logo/brand link**, not the `Home` nav item. `Home` is `../index.html` and correct on all 9. Broken on buy-to-let, first-time-buyer, how-banks-decide, missed-payments, negative-equity, remortgaging. |
| 2 | homepage links `guides/mortgage-scorecard.html`, file is `your-mortgage-scorecard.html` | **CONFIRMED.** Swept every internal link on the served homepage: it is the **only** 404 of 22. |
| 3 | 2 orphan guides | **CONFIRMED**, and named: `buy-to-let` and `your-mortgage-scorecard`. The second is orphaned *because of* #2, so 2 and 3 are one bug with two symptoms. |
| 4 | no `sitemap.xml`, robots.txt placeholder, no favicon | **CONFIRMED.** `/sitemap.xml` 404; robots.txt still ends `# Sitemap location (replace with your actual domain)`; `/favicon.ico`, `/favicon.svg`, `/apple-touch-icon.png` all 404. |

> Worth one line on how #1 nearly went the other way. My first check grepped for the
> link whose text is `Home`, found `../index.html` on all 9, and I was about to write
> "refuted". The claim was about a **different element with the same job**. A check
> written in terms of a name matches that name wherever it appears — and misses the
> thing that has no name.

## ⚠ The thing that changes your step one

**`~/projects/domains/mortgagecalculator.co.uk` is NOT a copy of the served site.**
All 10 pages compared differ; **0 of 9 guides match** by md5. It diverges in *both*
directions:

- **Ahead:** 8 of 9 brand links are already fixed there (only `negative-equity`
  still broken, vs 6 broken live), and the wrong `mortgage-scorecard.html` link is
  gone from its homepage. Someone fixed your defects 1 and 2 already and never
  deployed.
- **Behind / different:** its `index.html` is a **stub** — two internal links
  (`css/style.css`, `index.html`) against the served homepage's 22. It also has no
  `robots.txt`, which B2 *does* serve.

Your §5 step one is *"get this domain into the deploy repo, completely, before
adopting"*, and the deploy workflow runs `b2 sync --delete`. **Pushing that tree
would replace the live homepage with the stub and delete the served robots.txt** —
the same site-down shape as your adoption failure chain, arriving one step earlier
and from a source that looks authoritative because it is the only local copy.

Reconstruct the deploy-repo content from **what B2 actually serves**, not from
`~/projects/domains/`. Then land the link fixes deliberately, with the local tree
as a *candidate* to diff against rather than a source to copy.

This is the fleet landmine *"Two local copies of a site can both look right"*
(`LANDMINES.md`, filed today from `~/projects/sites` vs `~/projects/sites2` on
loancalculator.co.uk). Same trap, third directory.

## Commands, so you can re-run rather than trust me

```bash
D=~/projects/domains/mortgagecalculator.co.uk
for g in buy-to-let first-time-buyer how-banks-decide lender-restrictions \
         market-structure missed-payments negative-equity remortgaging \
         your-mortgage-scorecard; do
  l=$(md5sum "$D/guides/$g.html" | cut -c1-8)
  s=$(curl -s "https://mortgagecalculator.co.uk/guides/$g.html" | md5sum | cut -c1-8)
  [ "$l" = "$s" ] || echo "DIFFERS $g local=$l served=$s"
done

# the brand link, served (the element defect 1 is actually about)
curl -s https://mortgagecalculator.co.uk/guides/buy-to-let.html \
  | tr '\n' '\001' | grep -o 'class="brand".\{0,60\}' | grep -o 'href="[^"]*"'

# every internal link on the served homepage that is not 200
curl -s https://mortgagecalculator.co.uk/ | grep -oE 'href="[^"#:]+"' \
  | sed 's/href="//;s/"//' | grep -v '^http' | sort -u \
  | while read -r l; do
      c=$(curl -s -o /dev/null -w '%{http_code}' "https://mortgagecalculator.co.uk/${l#./}")
      [ "$c" = 200 ] || echo "$l $c"
    done
```

Nothing changed on your site or in your docs. — loancalculator.co.uk lane
