# RUNBOOK — enumerating the domain estate, and separating parked from live

Written 2026-08-19 for the owner's ask: *"look in Nominet for the domains. Separate those
that aren't parked… Please also extract the people's names domains."*

**The split that matters:** enumerating the domains needs Nominet; **classifying** them does
not. Nameserver delegation is public, so once a list exists, everything below runs from any
machine with network access and no registrar credentials at all.

---

## 1. ✅ NOMINET LOGIN PROVEN 2026-08-19 — the allowlist is good, the walk is next

The owner ran the login. It returned:

```
GREETING_BYTES=2527
LOGIN_CODE=1000
LOGIN_MSG=Command completed successfully
```

**So the egress allowlist, the TAG and the password are all correct**, and the client works.
That was the only genuinely uncertain part — a session cannot run it (see below), and the
greeting alone would have proved nothing.

> **⚠ CORRECTED 2026-09-03 — "and the client works" overstated this, and the walk below was
> BROKEN for two weeks.** The login proved the credentials and the allowlist; **no `list`
> command had ever been sent.** When one finally was (2026-09-02, owner-run), the registry
> refused it: **2001 Command syntax error** — `<list:expiry>` is a SIMPLE element holding
> `YYYY-MM` directly, and `epp.pl` sent the nested `<list:expiry><list:month>…</list:month>`
> form, so the element's own content was empty. Fixed in `epp.pl` and `scripts/domains/nominet.py`
> (`316d83c4c`). The XML was well-formed throughout — only the registry's schema could catch it.
> **Go-forward the walk is `python3 scripts/domains/nominet.py walk --months 120`** (§2 of
> `docs024_key_docs_latest/nominet_domain_management/RUNBOOK_nominet_domain_management.md`),
> which also carries the expiry month through. LANDMINES: "A PROVEN LOGIN certifies the
> credentials, never the COMMANDS".

**Estate size, per the owner 2026-08-19: ~1,500 `.uk` domains, not 2,000.**

**Still to run — the expiry walk**, which is what actually produces the list. ⚠ **The loop
below is SUPERSEDED** (broken XML per the correction above, and twelve months under-counts a
TLD that registers up to ten years) — use
`python3 scripts/domains/nominet.py walk --months 120 > all_domains.txt` instead; kept here
because other docs cite this block:

```sh
for m in 2026-09 2026-10 2026-11 2026-12 2027-01 2027-02 2027-03 2027-04 2027-05 2027-06 2027-07 2027-08; do
  ( set -a; . ~/.config/nominet/credentials; set +a; printf '%s\n%s\n' "$TAG" "$EPP_PASSWORD" ) \
    | kubectl -n ai-persona-system exec -i postgres-clients-0 -- perl /tmp/epp.pl list "$m"
done | grep '^DOMAIN' | cut -f2 | sort -u > all_domains.txt; wc -l all_domains.txt
```

**Sanity-check the total against ~1,500 before trusting it.** The walk covers twelve months of
expiry dates; a domain whose expiry falls outside that window is silently absent, and the
symptom is a plausible-looking list that is simply short. If the count is well under 1,500,
widen the range rather than assume the estate shrank.

## 1b. ⛔ WHY A SESSION CANNOT RUN IT — enumerating from Nominet

Nominet is **EPP over TLS on `epp.nominet.org.uk:700`**, not a REST API with a bearer token.
Credentials are present and complete as of 2026-08-19: `~/.config/nominet/credentials` holds
both `TAG=` and `EPP_PASSWORD=` (the TAG was the missing piece flagged on 2026-08-04 and it is
there now).

**A session's permission classifier refuses to read a credentials file and pipe it into
another process** — correctly, because that is indistinguishable from credential exfiltration.
So the login must be run by the owner. Everything is staged for it:

```sh
# 1. the client is already inside the pod at /tmp/epp.pl; re-copy if the pod restarted:
kubectl -n ai-persona-system cp scripts/domains/epp.pl postgres-clients-0:/tmp/epp.pl

# 2. THE ALLOWLIST TEST — run this first, it is the only thing that proves egress:
( set -a; . ~/.config/nominet/credentials; set +a; printf '%s\n%s\n' "$TAG" "$EPP_PASSWORD" ) \
  | kubectl -n ai-persona-system exec -i postgres-clients-0 -- perl /tmp/epp.pl login

# expect: GREETING_BYTES=2531 / LOGIN_CODE=1000
# LOGIN_CODE=2200 means the egress IP is not allowlisted, NOT that the password is wrong.

# 3. once login returns 1000, walk the twelve expiry months for full coverage:
for m in 2026-09 2026-10 2026-11 2026-12 2027-01 2027-02 2027-03 2027-04 2027-05 2027-06 2027-07 2027-08; do
  ( set -a; . ~/.config/nominet/credentials; set +a; printf '%s\n%s\n' "$TAG" "$EPP_PASSWORD" ) \
    | kubectl -n ai-persona-system exec -i postgres-clients-0 -- perl /tmp/epp.pl list "$m"
done | grep '^DOMAIN' | cut -f2 | sort -u > all_domains.txt
```

**⚠ WHY IT MUST RUN FROM THE CLUSTER, not from the office machine.** Nominet allowlists the
*egress address*, and the office connection's address rotates (it already broke this once, and
broke the Cloudflare token the same way). The cluster has five fixed addresses —
`134.213.168.26 / .37 / .44 / .54 / .56` — which is why the transport is
`kubectl exec … postgres-clients-0`. Confirmed 2026-08-19: that pod reaches
`epp.nominet.org.uk:700` and receives the full greeting.

**⚠ THE GREETING PROVES NOTHING.** Nominet serves its 2,531-byte greeting to **any** IP,
allowlisted or not. Only a `LOGIN_CODE=1000` proves the address is allowed. The same trap
exists on Cloudflare (`/user/tokens/verify` returns 200 for a token whose IP filter 403s every
real endpoint). **Never treat a successful connection as a successful authorisation.**

**Nominet only covers `.uk`.** Domains on other TLDs sit at Dynadot / Porkbun / Spaceship,
whose API keys are still outstanding (see `domains_cloudflare_rollout/README_where_we_are.md`).
An `.uk`-only inventory is a partial answer and should be labelled as one.

**The cheaper alternative, and it may be the better one:** export the domain list as CSV from
Nominet Online Services. It needs no allowlist, no EPP, and it gives a **checkable total** —
which the month-walk does not, because a domain whose expiry falls outside the twelve months
walked is silently absent.

---

## 2. Classifying the list — no credentials needed

```sh
python3 scripts/domains/classify_nameservers.py all_domains.txt > classified.tsv
python3 scripts/domains/classify_nameservers.py all_domains.txt --check-http > with_http.tsv
```

Classes: `CLOUDFLARE` · `CLOOK` · `REGISTRAR_DEFAULT` · `PARKED` · `OTHER` · `NXDOMAIN` ·
`STATUS_n` · `ERROR`.

**The patterns were measured, not guessed** — run over the 152 portfolio domains 2026-08-19:

| nameservers | count | meaning |
|---|---|---|
| `ns1.dan.com, ns2.dan.com` | **124** | parked at the Dan.com marketplace |
| `*.ns.cloudflare.com` | 6 | our built sites |
| `ns1/2.aftermarket.com` | 3 | marketplace parking |
| `ns1.namepros-dns.com, ns2.namepros-dns.is` | 1 | marketplace parking |
| `ns1/2.domainlore.co.uk` | 1 | marketplace parking |
| *(no NS returned)* | 17 | see below |

**No Clook nameservers appear anywhere in the portfolio** — so if Clook fronts anything, it is
outside this 152 and the pattern is still unverified. `CLOOK` matches the substring `clook`;
**check that against a real Clook-hosted domain before trusting a zero.** A class that has never
matched anything is not evidence of absence.

**Two deliberate design points:**

- **Delegation is not service.** The tool reports who runs the DNS, not whether a site serves.
  Measured counter-examples on our own account: `apis.uk` and `ugg2.com` are on Cloudflare with
  a worker route and have **no site behind them at all**. `--check-http` answers the second
  question in its own columns; the two are never blended into one "parked" verdict, because a
  blended verdict is wrong in both directions.
- **It reads the BODY size, not just the status.** A parked domain returns 200 on every path
  (`LANDMINES`), so status alone cannot discriminate.

### The 17 that answered no nameservers — worth the owner's eye

11 `SERVFAIL` and 6 `NXDOMAIN`, including `equity-release-calculator.co.uk`,
`consolidateloans.co.uk/.uk`, `interestrates.co.uk`, six `healthinsurance*` variants,
`bankingequipment.co.uk/.uk`, `mortgagerepaymentsinsurance.co.uk/.uk`,
`privatehealthinsurancequotation.uk`, `bestlandlordinsurancerates.co.uk`.

> **⚠ CORRECTED 2026-08-19 by the owner: *"No nameserver usually means I never set a
> nameserver."*** So the ordinary reading of both `NXDOMAIN` and `SERVFAIL` on this estate is
> **"registered but never delegated"**, not "the registration lapsed". My original text said
> lapsed, which would have sent someone to the registrar to re-buy domains he already owns.
>
> The distinction still matters for a different reason: these 17 are the domains that need
> **nameservers set before anything can be built on them**, which is a real step the Cloudflare
> rollout has to cover. They are not a problem, they are a to-do list. Treat a `NO_NS` /
> `SERVFAIL` / `NXDOMAIN` result as **"not yet pointed anywhere"** and only investigate
> ownership if the registrar list disagrees.

---

## 3. People's-name domains

```sh
python3 scripts/domains/extract_person_name_domains.py all_domains.txt --only NAME
```

Three buckets — `NAME` / `MAYBE` / `NO` — deliberately, because a two-way split forces every
ambiguous case into a wrong answer and the ambiguous cases are systematic: many English
surnames are also ordinary words and place names (Baker, Green, Fields, Hastings).

**Validated against known answers before use**, and the first run **failed**:
`jamesbrown.co.uk`, `sarahjones.uk`, `davidsmith.me.uk` and `peterhiggins.com` all came out
`MAYBE`, because *brown*, *jones*, *smith* and *higgins* are entries in the British English
dictionary, so the "forename + dictionary word" rule fired on precisely the cases the tool
exists to catch. Fixed by checking a common-surname list **before** the dictionary. All four
now return `NAME`. The list is data, not logic — extend it freely.

Run over the 152 portfolio domains: **0 `NAME`, 8 `MAYBE`, 144 `NO`.** Zero is the expected and
correct answer for a finance-keyword list, and it is also the check that the tool does not
over-fire.

**Stated recall limit:** the forename list is common UK/Anglophone names, so
`priyasharma.co.uk` or `olukayode.uk` land in `MAYBE`, not `NAME`. On a real estate of a few
thousand domains, read `MAYBE` as well as `NAME`.

---

## 4. Files

`scripts/domains/classify_nameservers.py` · `scripts/domains/extract_person_name_domains.py` ·
`scripts/domains/epp.pl` (the EPP client; TLS via `openssl s_client` because the pod's Perl has
no `IO::Socket::SSL`) · `domains_cloudflare_rollout/RUNBOOK_domains_cloudflare_rollout.md`
(registrar credential state, the egress traps).
