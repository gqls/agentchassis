# RUNBOOK — Sedo domain management

Sedo (sedo.com) is a domain marketplace + parking service. Its API:
`https://api.sedo.com/api/v1/` — plain GET/POST, XML responses (SOAP exists,
not needed). Function reference: https://api.sedo.com/apidocs/v1/Basic/
Every call authenticates with four values: `partnerid`, `signkey`,
`username`, `password` — the **account password travels in every request**,
so it is an API credential proper: keep it ≤16 chars (the API's own cap on
the param), alphanumeric-ish, and used nowhere else.

## §1 Owner: create the account (once) — **DONE before this lane opened**

The owner confirmed 2026-09-02 (evening): account exists under
**info@designconsultancy.co.uk** with **partnership status**. §2 and §3
remain. (If the account password exceeds 16 chars, it will fail API auth's
param cap — worth checking before §3.)

Original steps kept for the record:
1. Sign up at https://sedo.com — use the address the estate is registered
   under. Username ≤25 chars; **password ≤16 chars, no `"` or `\`**
   (API param cap is 16; the curl-config quoting in our client rejects
   quotes/backslashes). `[UNVERIFIED]` whether web signup permits longer
   passwords that then fail API auth — staying ≤16 avoids finding out.
2. Register for Sedo's **partner programme** (prerequisite for API access,
   per their docs and community threads).

## §2 Owner: request API access (once)

Email **api@sedo.com** *from the registered account address*. Draft:

> Subject: API access request — account <username>
>
> Hello — please could you enable API access for my Sedo account
> "<username>" (this email is the account's registered address)? I am
> registered for the partner programme. I will use the Basic API to manage
> my own domain portfolio from my own tooling: listing domains, updating
> prices and for-sale status, and reading parking statistics.
> Please issue a Partner ID and SignKey. Thank you.

On approval Sedo sends a **Partner ID** and **SignKey**.

## §3 Owner: install the credentials (once, ~2 min)

**In your own terminal — not in a Claude chat** — write the four values to a
file, create the secret from the file, delete the file. The only command a
session ever needs to see is the middle one, which contains no values:

```bash
# your own terminal:
cat > "$HOME/sedo.env" <<'EOF'
SEDO_PARTNERID=<partner id>
SEDO_SIGNKEY=<signkey>
SEDO_USERNAME=<username>
SEDO_PASSWORD=<password>
EOF

kubectl -n ai-persona-system create secret generic sedo-api-credentials \
  --from-env-file="$HOME/sedo.env"

rm "$HOME/sedo.env"
```

Then in any session: `scripts/domains/sedo-api.sh --check-secret` (prints key NAMES
only, never values) and you're live.

## §4 Using it (any session)

```bash
scripts/domains/sedo-api.sh --self-test      # offline; no cluster needed
scripts/domains/sedo-api.sh --probe         # cluster + API reachability, NO creds needed
                                    #   expects SEDOFAULT E7 ("Partnerid doesn't exist")
scripts/domains/sedo-api.sh --check-secret  # secret present + all four keys named

scripts/domains/sedo-api.sh DomainList 'results=100'            # portfolio, first page
scripts/domains/sedo-api.sh DomainList 'results=100' 'startfrom=100'   # next page
scripts/domains/sedo-api.sh DomainStatus 'domain[]=example.com'
scripts/domains/sedo-api.sh DomainInsert  # see the function doc for its params first
```

Common functions (Basic API): `DomainList`, `DomainListExtended`,
`DomainStatus`, `DomainInsert`, `DomainEdit`, `DomainDelete`,
`DomainSearch`; parking: `SetDomainSetup`/`GetDomainSetup`, `SetKeyword`,
`SetTemplate`, `SetRelatedLinks`. Read the function page before a writing
call — param shapes vary.

## §5 Gotchas (each one measured or sourced)

- **Faults arrive IN-BAND as XML with a normal HTTP status** — a
  `<SEDOFAULT>` body with `faultcode` (e.g. `E7` = bad partnerid). Check the
  **body**, not the exit code (measured 2026-09-02, both busybox wget and
  curl probes).
- **`results` max 100 per call** (API doc); page with `startfrom`.
- Output ints: `currency` 0=EUR 1=USD 2=GBP; `forsale`/`fixedprice` 0/1;
  domains in ACE (punycode) form (DomainList doc).
- **Do not "just curl it locally" with creds** — the client exists so
  credentials stay out of session transcripts (owner ruling 2026-08-23).
  Design: pod gets creds via `envFrom: secretRef`, expands them
  **inside the container** into a 0600 curl config file. Nothing secret in
  argv, overrides JSON, pod spec, or container logs.
- **`kubectl run -i` was deliberately avoided** for real calls — kubectl
  itself warns that the command line lands in container logs; our runner
  does run → wait → logs → delete, with the command constant and creds
  env-sourced.
- **BusyBox wget was rejected as transport**: it does no TLS certificate
  validation (`wget: note: TLS certificate validation not implemented`,
  measured in-cluster 2026-09-02) — unacceptable for calls carrying the
  account password. Hence the pinned `curlimages/curl:8.10.1` image
  (public pull verified working from this cluster, 2026-09-02).
- Fleet-wide `Unauthorized` from kubectl = the 3-day token expiry
  (existing landmine), not this script.
- Param charset is deliberately tight (script rejects `"` `\` `$` `` ` ``
  `&` `;` etc.) because values pass through overrides JSON and a curl
  config file. If a legitimate Sedo param ever needs a richer charset,
  widen the validation *and* re-check both quoting layers — do not just
  delete the check.

## §6 Domain input format — the importer sheet, and its DomainInsert mapping

The owner supplied Sedo's official bulk-listing template
(`Example_File_Domain_Importer.xlsx`, decoded 2026-09-02: one sheet, seven
columns, eleven example rows, **no embedded validation rules** — the
semantics below come from those rows plus the DomainInsert function doc).
This column set is the lane's canonical "list these domains" input, whether
the owner uploads a sheet in Sedo's web UI or we push the same data via API.

| Sheet column | Values seen | DomainInsert param | Notes |
|---|---|---|---|
| Domain Name | `example1.tld` | `domain` | ACE/punycode form |
| Selling Option | `MAKE_OFFER` / `BUY_NOW` / blank | `fixedprice` | BUY_NOW=1, MAKE_OFFER=0 |
| For Sale | yes/no (case-insensitive) / blank | `forsale` | yes=1, no=0 |
| Price | number / blank | `price` | 0 = no price |
| Minimum Price | number / blank | `minprice` | 0 = no minimum; seen only with MAKE_OFFER |
| Currency | EUR/USD/GBP / blank | `currency` | 0=EUR 1=USD 2=GBP |
| Action Type | `DELETE` / blank | — | DELETE = remove from account → API `DomainDelete`, not an insert param |

API-only, not in the sheet (the web importer defaults them; API callers pass
them):
- `domainlanguage` — ISO 639-1, **required per entry** on API calls (`en`
  for this estate);
- `category` — optional, ≤3 ids per domain (`sedoapi_Categories` reference).

DomainInsert facts (function doc, fetched 2026-09-02):
- max **50** entries per request;
- **insert is asynchronous** — entries "first have to pass a couple of
  checks"; a domain missing from DomainList right after an `ok` insert is
  NOT a failure (failed checks arrive by email to the account address);
- **every inserted domain is auto-enabled for parking, for-sale or not** —
  this interacts with the cross-lane parking constraint (PLAN): inserting a
  domain here has a Sedo-side parking effect even with `forsale=0`, so
  "add to Sedo" stays a per-domain decision;
- the response is per-entry: `status` = `ok` or a fault code, with a
  `message` — **check every item**, not just the HTTP layer or the absence
  of a `<SEDOFAULT>`.

Wire shape: the function doc's own example builds the query with PHP
`http_build_query`, i.e. PHP-nested keys — which the script's param charset
accepts:

```bash
scripts/domains/sedo-api.sh DomainInsert \
  'domainentry[0][domain]=example-site.co.uk' \
  'domainentry[0][forsale]=1' \
  'domainentry[0][fixedprice]=0' \
  'domainentry[0][price]=500' \
  'domainentry[0][minprice]=100' \
  'domainentry[0][currency]=2' \
  'domainentry[0][domainlanguage]=en'
```

`[INFERRED]` the nested-key encoding — the doc never prints the literal
URL, only the `http_build_query` call that would produce it. The first real
call is ONE domain, and confirming this shape (per-entry `status=ok`) is
part of what that call is for.

## §7 Generating the bulk-import sheet (no credentials needed)

`scripts/domains/sedo-importer-xlsx.py` (self-test: `--self-test`, 10
checks) reads the domain_valuation lane's inbound CSVs and writes the
importer xlsx in Sedo's own template shape (§6), plus a CSV twin for
review/diff and a provenance CSV (domain, source file, NS class):

```bash
IN=docs/agent_docs/docs024_key_docs_latest/domain_valuation/inbound
OUT=docs/agent_docs/docs024_key_docs_latest/sedo_domain_management/outbound
python3 scripts/domains/sedo-importer-xlsx.py build \
  --out $OUT/SEDO_IMPORT_<date>_draftN.xlsx \
  --csv-out $OUT/SEDO_IMPORT_<date>_draftN.csv \
  --provenance-out $OUT/SEDO_IMPORT_<date>_draftN_provenance.csv \
  --domains $IN/dynadot_domains_<date>.csv \
  --domains $IN/porkbun_domains_<date>.csv \
  --domains $IN/spaceship_domains_<date>.csv \
  --domains $IN/nominet_domains_<date>.csv \
  --exclude-file $OUT/EXCLUDED_live_<date>.txt \
  [--exclude-file $OUT/EXCLUDED_owner_<reason>_<date>.txt ...] \
  [--prices docs/agent_docs/docs024_key_docs_latest/domain_valuation/OUTPUT_prices_<date>.csv]
```

Gotchas, each earned:
- **The fence is live-site protection, and it is a UNION of two independent
  sources, not one** (widened 2026-09-03 after Nominet flagged the gap):
  1. **NS-based**: Cloudflare-NS domains from the registrar CSVs
     (`awk -F, 'FNR>1{gsub(/"/,""); if (tolower($4) ~ /cloudflare/) print $1}'`)
     — only catches domains whose CSV carries a `nameservers` column
     (dynadot/porkbun/spaceship; Nominet's walk does not).
  2. **DB-based**: `SELECT domain FROM sites WHERE status IN ('deployed','test')
     AND domain NOT LIKE '%.internal'` against `clients_db` — the
     authoritative "is this actually a site" signal, registrar-independent.
  3. **Nominet's own Cloudflare zone list** (their lane tracks cutover
     state directly) — messaged on request, not queryable by this lane.
  **Take the union of all three you can obtain; do not pick one.** Nominet's
  own words, 2026-09-03: "I'd treat a domain as managed, fence it out, if
  it's in EITHER my zone list OR has a sites row" — a zone can exist with
  no site behind it (apis.uk, ugg2.com had a worker route, no site row) and
  a site row can exist before its zone is cut over; either alone misses
  cases the other catches. Measured 2026-09-03: the NS-based method alone
  found 19; the union found **50** — 31 were invisible to the NS check,
  including `adversecreditmortgage.co.uk` (sits on MARKETPLACE nameservers
  despite being `status='deployed'` — a known cross-lane case, improvement-
  loop D2; the DB is what catches it, NS cannot). **Regenerate the whole
  union — meaning RE-QUERY `sites`, not reuse yesterday's file — before
  EVERY draft, not just the first one of a session** (this was written as
  general guidance 2026-09-03 morning and not followed to the letter the
  same afternoon: `EXCLUDED_live_2026-09-03.txt` was built once at 08:57
  UTC and reused across three later regenerations; a `sites` row created
  at 09:27 UTC — `copyonline.co.uk`, real, `status=test` — sat un-fenced
  in the sheet for hours until an owner statement to a different session
  caught it. The query and its status filter were correct throughout;
  only the cadence of re-running it was wrong. One-line command each
  time, no excuse to skip it: `SELECT domain FROM sites WHERE status IN
  ('deployed','test') AND domain NOT LIKE '%.internal' ORDER BY domain;`).
  Note also: a domain that has NEVER had a `sites` row at all (owner
  states an intent, nobody has run a command yet) defeats even a fresh
  re-query — no DB signal exists to find. No mechanical fix for that one;
  it is why an owner statement can still override a clean automated check.
- **The live-site fence is ALSO structurally blind to any hosting stack
  outside the framework's own tracking** (measured 2026-09-03): it is
  derived from Cloudflare delegation + the `sites` table, so a domain
  hosted anywhere else (this estate's older "Clook" stack,
  `dns*.uk-noc.com`, predating the Cloudflare rollout) is invisible to it
  regardless of query freshness. 33 such domains were found live and
  unfenced in one sweep, including the owner's own email and company
  domains. There is no single-source fix — a second, independent
  delegation sweep is required, not a faster re-query of the first
  source. `EXCLUDED_live_clook_2026-09-03.txt` is the current one-off
  patch; the domain_valuation lane has offered to own an ongoing
  delegation-plus-`sites` derivation going forward.
- **`leopardessconsulting.co.uk` is PERMANENTLY EXCLUDED — confirmed by
  the owner's own explicit word, 2026-09-03 ("no, leopardessconsulting
  need not be listed").** Treat this exclusion as independent of the
  fence mechanism (same durability class as copyonline and wykefarm).
  **Any future "yes, list it" needs the D4-level standard — explicit,
  by name, extra scrutiny — not the ordinary live-site standard.**
  > **THIRD CORRECTION 2026-09-03** (kept in sequence, not collapsed —
  > the arc is the useful part for a later reader):
  > 1. Original claim: "PRIOR owner ruling — a paying client's site.
  >    Source: `about_page_commercial/PLAN_2026-07-24…`, D4: 'buy this
  >    site/ads/built-by on a paying client's site (leopardess) is a
  >    relationship breach.'" Taken from D4's wording, never checked
  >    against the live site.
  > 2. **Refuted** by the valuation lane: it is NOT a third party — it's
  >    the owner's OWN consultancy (live copy: "we run 22 of our own
  >    sites… before we build anything for you"; owner: "a
  >    representation of my own services"). Downgraded the standard to
  >    "ordinary live site" on the assumption the stricter D4 standard
  >    only made sense if a real client existed.
  > 3. **Reinstated by explicit ruling**, same day: "we can work as if
  >    leopardess is a paying client if that helps, I do pay through the
  >    nose for these tokens." The FACT from step 2 stands (no third
  >    party); the RULING is separate and controlling — he is opting IN
  >    to the D4 standard by choice, not because a client would be
  >    harmed. That assumption in step 2 was the actual error, not the
  >    fact itself.
  >
  > The exclusion never moved through any of the three steps — only the
  > standard for handling a FUTURE reversal did, and it is now back
  > where it started, correctly grounded this time.
- **`webdesign.uk` and `webdesign.co.uk` are two different domains — do
  not conflate them, ever, in pricing or docs.** `webdesign.uk` (18
  pages) was the owner's own example of a domain that "could be worth
  over a million" — **and is also the shopfront**: CLAUDE.md:716 records
  the owner's own words, "the webdesign.uk shopfront." (CORRECTED
  2026-09-03: this entry originally said `webdesign.co.uk` was the
  shopfront, backwards — copy_quality_two_stage caught it, verified
  independently against CLAUDE.md before propagating the correction; the
  earlier claim was an inference from page count and name, not a
  checked fact.) `webdesign.co.uk` (155 pages) is a large site whose
  role is otherwise unconfirmed. Both are now owner-confirmed in scope
  for a possible sale (2026-09-03) — always use the full domain string,
  never "webdesign" as shorthand, in any cross-lane message or generated
  file.
- **The two webdesign domains are stated to CONVERGE — "the same
  endpoint one day"** (owner, 2026-09-03) — while simultaneously BOTH
  being in scope for sale. copy_quality_two_stage flagged the coupling
  this creates: a domain destined to merge cannot be sold independently
  without breaking the consolidation plan — whoever buys one is not
  obliged to honour a redirect the owner intended. Not resolved here;
  put to the owner rather than assumed (RUNBOOK is not the place to
  decide a commercial structure). See NOTES for the question as put.
- **An owner-requested withdrawal (e.g. "take out the X family of
  domains") is a SEPARATE fence, not an addition to the live-site one**
  (added 2026-09-03, `--exclude-file` accepts multiple and unions them —
  keep files named by reason: `EXCLUDED_live_<date>.txt` vs
  `EXCLUDED_owner_<reason>_<date>.txt`). Appending unrelated names into the
  live-site file was tried once and reverted the same session — the
  live-site fence gets fully regenerated from its own two sources next
  time (Nominet zones + `sites` table), which know nothing about a
  personal withdrawal request, so anything hand-added there would silently
  vanish on the next regeneration.
- Without `--prices`, every row is MAKE_OFFER / yes / no price — the
  agreed interim; prices come from the valuation lane's canonical
  `OUTPUT_prices_<date>.csv` (their column freeze; do not build against it
  until they message that it is frozen).
- The registrar CSVs are **dated snapshots** — regenerate from fresh
  exports rather than reusing old inbound files (spaceship lane: re-check
  with `scripts/domains/spaceship.py domains` if days have passed).
- Do NOT feed listings files (`*_listings_*.csv`, `*sellerhub*`) as
  `--domains` — the Spaceship seller-hub file holds 831 marketplace rows of
  which only 36 are registered there (their lane, 2026-09-02).
- The script verifies its own artefact by re-reading the zip; counting
  `<row` in the sheet XML by `grep -c` reads 1 (single-line XML) — use
  `grep -o | wc -l`.
- **Before a PRICED import, sweep the `*_listings_*.csv` files in the
  valuation inbound for live asks elsewhere.** A domain with a live Buy Now
  on another marketplace must carry the SAME price on Sedo, or the other
  listing comes down first — otherwise one domain shows two prices in two
  shops. As of 2026-09-02: **5** live Dynadot Buy Nows (traderboltai.com
  $7,999, currencyforecaster.com $3,999, thailandstocks.com $2,988,
  riderlessbikes.com $2,888, carsforchildren.com $2,508 — all in draft1 as
  make-offer/no-price, which shows no conflicting figure); Afternic asks
  arrive with the owner's export. Draft1 needs no regeneration for this —
  the check belongs to every priced draft.

## §8 The BUY-NOW gate (owner ruling, 2026-09-03)

**No session may cause the generator to ship a BUY_NOW selling option or
any non-empty `price` without `--owner-authorized-buy-now-prices` on the
command line.** This is a hard technical block, not a convention — the
generator refuses (`SystemExit`, exit 1) and prints every blocked
domain+price before writing anything, and the check is re-run against
the WRITTEN artefact after the fact as a second, independent layer (in
case some future code path reaches `write_xlsx()` without going through
the checked row list).

```bash
python3 scripts/domains/sedo-importer-xlsx.py build ... --prices X.csv
# → REFUSED, lists every blocked domain, exits 1, writes nothing

python3 scripts/domains/sedo-importer-xlsx.py build ... --prices X.csv \
  --owner-authorized-buy-now-prices
# → proceeds, but prints a loud banner naming every priced domain first
```

**What the flag is for, and what it is not.** It exists so the OWNER
himself can authorize a specific, deliberate pricing run — never for a
session to pass on its own initiative, on inference from a conversation,
or as something that becomes routine. If a session finds itself reaching
for this flag because "the valuation lane's prices are ready" or
similar, that is exactly the case it is meant to stop — the owner needs
to say so, for that run, not have a session infer authorization from
context.

**What IS gated**: `selling_option == BUY_NOW`, and any row with a
non-empty `price`, regardless of selling_option (a MAKE_OFFER row with a
set `price` is still an asking figure someone set, not just a floor).
**What is NOT gated**: `min_price` alone. A minimum under MAKE_OFFER is
protective — it stops a lowball — and gating it would work against the
exact caution this session has applied throughout (relojistas.com's
$12k floor, cartoon.co.uk's £5k floor). Gating it too would make the
safe, recommended practice (set a floor, leave the price open) just as
hard as the thing being prevented.

**Self-test proves the gate actually fires, both ways** (`--self-test`):
a BUY_NOW row is blocked without the flag and allowed with it; a
MAKE_OFFER row with a set price is blocked the same as BUY_NOW; a
min_price-only row is confirmed NOT blocked. A test that only exercises
the allowed path proves nothing about whether the gate can ever refuse —
per the "mutate the code to prove the guard" practice, this checks the
refusal path exists and actually raises, not just that the happy path
still works.

**Verified 2026-09-03**: every draft generated in this lane to date (1
through 8) carries zero BUY_NOW/priced rows, checked directly against
each CSV, not inferred from "the script defaults to blank."

## §9 Floor policy — Sedo minimum floors are ALLOWED; the site's own display is NEVER floored; the two are UNLINKED (owner ruling, 2026-09-03)

**The live-sites track (relojistas.com, the webdesign pair, and every
other previously-held-for-pricing domain) no longer waits on real
prices before listing.** Owner: "we'll just have to bear with the low
balls for a while" — folded into the main sheet at draft9, blank
Minimum Offer, same as ordinary stock.

**But "blank for now" is not "blank forever," and the standing rule
going forward is more specific than that**:
- **Sedo's `Minimum Price` field MAY carry a real number** whenever the
  owner states one directly or one is agreed with him in conversation —
  set it the moment a real figure exists, don't wait for a blanket
  "pricing is done" signal.
- **The domain's own site (about-page CTA, or anything else on the
  live page) must NEVER display a price or floor.** That's a different
  lane's mechanism (about_page_commercial), already confirmed
  structurally floor-free — not this lane's to touch, but worth knowing
  the two systems are DELIBERATELY unlinked: Sedo can show a number
  while the site shows none, and that's the owner's stated intent, not
  an inconsistency to fix.
- **NEVER derive a Sedo floor from `site_specs.commercial.tier`, an
  automated appraisal, or any other computed source.** A floor is a
  direct-owner number or an explicitly agreed one, full stop — the
  valuation lane's pricing work may inform a CONVERSATION with the
  owner about what number to agree, but must never be auto-applied to
  the `--prices` file's `min_price` column without that conversation
  having happened. (This mirrors — and is the same underlying
  dependency as — the about_page_commercial and afternic lanes' PLANs
  wrongly naming "price-by-tier" as a source; not this lane's fix, but
  the same principle applies here.)
