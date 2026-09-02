# RUNBOOK — Sedo domain management

Sedo (sedo.com) is a domain marketplace + parking service. Its API:
`https://api.sedo.com/api/v1/` — plain GET/POST, XML responses (SOAP exists,
not needed). Function reference: https://api.sedo.com/apidocs/v1/Basic/
Every call authenticates with four values: `partnerid`, `signkey`,
`username`, `password` — the **account password travels in every request**,
so it is an API credential proper: keep it ≤16 chars (the API's own cap on
the param), alphanumeric-ish, and used nowhere else.

## §1 Owner: create the account (once)

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

Then in any session: `scripts/sedo-api.sh --check-secret` (prints key NAMES
only, never values) and you're live.

## §4 Using it (any session)

```bash
scripts/sedo-api.sh --self-test      # offline; no cluster needed
scripts/sedo-api.sh --probe         # cluster + API reachability, NO creds needed
                                    #   expects SEDOFAULT E7 ("Partnerid doesn't exist")
scripts/sedo-api.sh --check-secret  # secret present + all four keys named

scripts/sedo-api.sh DomainList 'results=100'            # portfolio, first page
scripts/sedo-api.sh DomainList 'results=100' 'startfrom=100'   # next page
scripts/sedo-api.sh DomainStatus 'domain[]=example.com'
scripts/sedo-api.sh DomainInsert  # see the function doc for its params first
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
