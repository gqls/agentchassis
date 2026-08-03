# PLAN — Cloudflare rollout across the whole domain portfolio

**Goal:** put (almost) all of the owner's domains — held at Nominet (own tag, EPP)
plus Dynadot, Porkbun and Spaceship — onto Cloudflare free plan, each zone getting
the standard template: apex A record + two worker routes to `portfolio-sites-router`.

## Phases

1. **Credentials + access** (in progress, 2026-08-03)
   - Cloudflare: DONE — scoped token verified, at `~/.config/cloudflare/token`.
   - Nominet: WAITING — need TAG + EPP password, and egress IP `151.226.83.138`
     allowlisted in Online Services. EPP endpoint `epp.nominet.org.uk:700` is
     reachable from this machine (tested 2026-08-02).
   - Dynadot / Porkbun / Spaceship: WAITING — see RUNBOOK for what to create where.
2. **Domain inventory** — list domains from each source via its API/EPP, merge,
   subtract the 36 zones already on Cloudflare, subtract the owner's skip-list
   (NOT yet provided — "almost all" is not a rule we can execute).
3. **Zone creation** — POST /zones in batches (free is the default plan), capture
   each zone's assigned NS pair from the create response.
4. **Records + routes** — apex A → `199.59.243.228` proxied; routes `domain/*`
   and `*.domain/*` → `portfolio-sites-router`.
5. **Nameserver repoint** — Nominet domains via EPP `domain:update`; registrar
   domains via each registrar's NS-update API.
6. **Activation verification** — poll zones until `pending` → `active`; re-trigger
   activation check via API where stuck; final census: every in-scope domain
   active, templated, routed.

## Decisions and reasons

- **Free plan, minimal template** — copied from the live `dartsonline.com` zone,
  which is the owner's existing pattern (single proxied apex A, two routes).
- **Registrar scopes minimal** — registrars only ever list domains and set
  nameservers; DNS/Workers live in Cloudflare. Don't request DNS scopes there.
- **Capture NS pair per zone at create time** even though the account pair looks
  constant (`alexis`/`leah`) — per-account assignment is convention, not contract.

## Open questions (owner)

- **www/wildcard record?** The template has no `www` or `*` DNS record, so the
  `*.domain/*` route never fires today. One extra proxied record per zone fixes
  it. Asked 2026-08-02, unanswered.
- **Skip-list** — which domains to leave alone, or a rule for it.
- **Zone-count ceiling** — self-serve accounts have an unpublished zone cap;
  thousands of zones may need a Cloudflare support ticket mid-run. Flagged, not
  yet hit.
