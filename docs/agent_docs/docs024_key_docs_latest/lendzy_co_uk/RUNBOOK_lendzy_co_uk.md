# RUNBOOK — lendzy.co.uk

Every command that was hard to get right, with its gotcha attached. When one changes, change it
**here**. Site id `8ff093d5-1f19-453b-9439-a10379bbcd76`.

`PSQL` below = `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

---

## 1. Probe the site at the artefact — ALWAYS with a control

A URL census with no invented-URL control has reversed an architectural conclusion on this estate
before. lendzy is **not** a catch-all domain (proved below) but prove it again each session — the
serving config is not ours and can change.

```bash
S=<scratchpad>
rm -f $S/p.html                       # a stale file reads as a successful fetch
curl -s -o $S/p.html -w "%{http_code} %{size_download}\n" https://lendzy.co.uk/
curl -s -o /dev/null -w "%{http_code}\n" https://lendzy.co.uk/definitely-not-a-page-9931
```
Expect `200` then **`404`**. If the control also 200s, every "the page serves" claim in this lane is
void until re-measured.

## 2. The nine tool pages: record vs artefact

The lane's founding defect. Run both halves; they disagreed for a month and each was reading
correctly.

```sql
-- the RECORD
SELECT name, build_status, deployed_at IS NOT NULL AS stamped
  FROM pages WHERE site_id='8ff093d5-1f19-453b-9439-a10379bbcd76' AND url LIKE '/tools/%'
 ORDER BY build_status, name;
```
```bash
# the ARTEFACT — input counts are the discriminator that a repair did not swap the tool
for u in price-cap-checker true-cost-calculator complaint-deadline-calculator \
         rollover-limit-checker affordability-complaint-checker; do
  rm -f $S/p.html; curl -s -o $S/p.html "https://lendzy.co.uk/tools/$u/index.html"
  printf "%-32s %s inputs\n" "$u" "$(grep -o '<input' $S/p.html | wc -l)"
done
```
**Baseline 2026-09-02, pre-repair:** price-cap **3**, true-cost **1**, complaint-deadline **2**,
rollover **8**, affordability **41**. A repair that changes these numbers has regenerated a tool
the owner told us to keep.

## 3. The predicate that ties it together

`deployed_at IS NULL AND COALESCE(build_status,'') <> 'deployed'` is
`datahelpers.NeverDeployedPagePredicate` — **one definition, many consumers**. It is why an
unstamped page is dropped from the sitemap *and* filed as an unbuilt link target. Change nothing
here without reading `datahelpers/links.go`; the listing floor is derived from it.

```sql
-- pages where NO component row carries a component_id: nothing resolvable to build from.
-- This is the lane's founding query; it returned exactly lendzy's three, fleet-wide.
SELECT s.domain, p.name, p.build_status, p.deployed_at IS NOT NULL AS stamped
  FROM pages p JOIN sites s ON s.id=p.site_id JOIN page_components pc ON pc.page_id=p.id
 WHERE p.status='active'
 GROUP BY s.domain, p.id, p.name, p.build_status, p.deployed_at
HAVING count(*) FILTER (WHERE pc.component_id IS NOT NULL) = 0;
```
⚠ **Do not shorten this to `WHERE pc.component_id IS NULL`.** That version returns 16 rows on 7
sites, most of them healthy, and would have sent this lane after the wrong mechanism. The
discriminator is *every* row on the page, not *any*.

## 4. The 47 unbuilt-link items

```sql
SELECT count(*) FROM site_work_items
 WHERE site_id='8ff093d5-1f19-453b-9439-a10379bbcd76'
   AND item_type='unbuilt_internal_link' AND status='needs_human_review';   -- 47 on 2026-09-02
```
They are **not broken links** — every target serves 200. They resolve when the record is repaired,
not by editing any link. Verify after repair with `scripts/probe-page-url.sh` on a sample, and by
this count reaching 0.

## 5. Sitemap

```bash
rm -f $S/sm.xml; curl -s -o $S/sm.xml https://lendzy.co.uk/sitemap.xml; grep -c '<loc>' $S/sm.xml
```
**27 on 2026-09-02** against 30 active pages — the three unstamped tool pages are the difference.
**30 is the acceptance number** after Phase A.

## 6. FCA Handbook — fetching it, and the control you must not omit

```bash
curl -sL -o $S/rule.html -w "%{http_code} %{size_download}\n" -A "Mozilla/5.0" \
  https://handbook.fca.org.uk/handbook/conc5a
grep -o '<title>[^<]*</title>' $S/rule.html | head -1
```

> ⚠ **THE HOST RETURNS 200 FOR EVERY PATH.** `[MEASURED 2026-09-02]`
> | request | bytes | title |
> |---|---|---|
> | `handbook/conc5a` (real) | 477,729 | `FCA Handbook - CONC 5A Cost cap for high-cost short-term credit` |
> | `handbook/mcob1` (real) | 398,024 | `FCA Handbook - MCOB 1 Application and purpose` |
> | `handbook/conc99z-invented` | 178,705 | `FCA Handbook` |
> | `totally-invented-path-4471` | 165,639 | `FCA Handbook - Home` |
>
> **The status code carries no information. The `<title>` is the discriminator.** Any collector or
> verifier runs an invented-rule control in the same batch, or its "the rule is there" is worthless.
> Note also that `https://handbook.fca.org.uk/robots.txt` returns **the Angular app shell**, not a
> robots file — so "I checked robots.txt" is a claim to distrust here unless you read the body.
>
> `www.handbook.fca.org.uk/handbook/CONC/5A/` **301s** to `handbook.fca.org.uk/handbook/conc5a`;
> use `-L` and record the effective URL, because the citation we store must be the URL that answers.

## 7. Claims / evidence register

```sql
SELECT count(*) FROM evidence_base WHERE site_id='8ff093d5-1f19-453b-9439-a10379bbcd76';  -- 0 on 2026-09-02
```
Zero is why lendzy's numeric scan never arms (`RFC_060` §1). The daily refresher
(`refresh_evidence_base_action.go`) only ever re-checks facts a human chose to cite — **an empty
register produces a clean run, for ever, and that clean run means nothing.**

## 8. Registering FCA facts — THE METHOD (offered to the other finance lanes, owner decision 2026-09-02)

The whole method, in order. It needs no new code. Worked end to end on lendzy (migration 695).

1. **Extract what your site actually asserts.** Pull every served page, extract sentences carrying
   figures, rule ids, or "the FCA requires…" shapes. Work from the ARTEFACT, not from specs — the
   spec may say something the page does not, and vice versa (lendzy's spec carried an error its
   pages had propagated).
2. **Find the governing rule in the handbook and READ it.** Do not trust the rule number your copy
   cites — that is precisely what was wrong twice on lendzy. The section pages parse into
   individual rules on the pattern `CONC \d+[A-Z]?\.\d+[A-Z]?\.\d+ dd/mm/yyyy [RG]` (78 rules in
   CONC 5A, 54 in CONC 6.7, measured 2026-09-02).
3. **⚠ The host 200s every path, invented rules included** — LANDMINES.md, footprint
   `handbook.fca.org.uk`. Confirm every fetch by `<title>`, never by status, and note the URL
   scheme is not uniform (`handbook/conc5a` works; `handbook/CONC/6/7.html` works;
   `handbook/conc6/7` is a MISS with a plausible body). Store the URL that ANSWERED (follow the
   301 from www).
4. **Verify the quote through the PRODUCTION matcher, with a control:**
   `go run ./cmd/fcaquotecheck <url> "<quote>" "zzz deliberately absent control"`
   — expect `true` then `false`. It calls the refresher's own extraction
   (`datahelpers.VisibleTextFromHTML` / `QuoteFoundInText`); a quote that fails here would be
   classified `citation_lost` — drift — **every day, for ever**, and that false alarm is
   indistinguishable from a real one. Never pre-check with your own regex: a mirror of the
   extraction passes while production disagrees.
5. **Write the register** as a `site_specs` row, aspect `evidence_base`, one fact per claim:
   `{id, kind:"policy", rule, claim, writer_line, source:{citation:{url, quote, title, publisher,
   accessed}}, verified_at}` — copy the shape from migration
   `docs/agent_docs/sql_for_agents/695_lendzy_evidence_base_fca_handbook_citations.sql`, including
   its guard (abort if a current register exists) and its DO/RAISE verify. Migrations are council
   scope; submit.
6. **Know the limit you are accepting:** the handbook has no rule-level URL, so the daily refresher
   keeps your QUOTES honest but cannot keep your `rule` field honest (a section page carries dozens
   of rules). Until the rule-span checker ships (RFC_060 §3d/Q6 — owner approved the build
   2026-09-02), `rule` is a human-verified field: check it against the rule's own heading, by hand,
   and date the check.
7. **The payoff:** the moment the row exists, the existing daily refresher starts re-fetching your
   citations and re-checking your quotes against the live handbook. An empty register produces a
   clean run for ever — that clean run is the thing you are replacing.

### 8b. Two host traps found by the loanzy lane running this method (2026-09-02) — check the HOST before the quote

- ⚠ **A Cloudflare-challenged host passes a human eyeball and fails unattended for ever.**
  `maps.org.uk` and `moneyhelper.org.uk` both sit behind a "Just a moment..." challenge: a citation
  there looks perfect in a browser and then classifies as drift **every day**, caused by the HOST,
  not the quote. Step 4 (the production-matcher probe) catches it — the quote will not match a
  challenge page — which is why step 4 is not optional. **Rule: a source must be verifiable
  UNATTENDED at write time; reject a challenge-page title as a source outright.** (Proposed as a
  write-time admission rule for the register mechanism — claims-verification lane's call.)
- ⚠ **gov.uk keeps organisation pages at their FOUNDING-name slug.** The Money and Pensions
  Service page lives at `single-financial-guidance-body` while titling itself MaPS — a name-guessed
  URL 404s. Never compose a URL from an organisation's current name; find the page, then store the
  URL that answered.

Their run also proves the method transfers: ICOBS/DISP/statutory-instrument sources verified first
try (farmerinsurance, migration 698), and it found the fleet's THIRD wrong attribution — loanzy
grouped MaPS under "FCA-authorised services"; MaPS is the statutory guidance body, not an FCA firm.
