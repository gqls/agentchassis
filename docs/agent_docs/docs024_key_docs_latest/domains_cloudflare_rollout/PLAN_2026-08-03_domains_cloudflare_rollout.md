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

## Owner decisions (2026-08-04)

- **www: YES** — "www can proxy to the domain". Template gains a proxied CNAME
  `www` → apex per zone (inherits wherever the apex points).
- **Skip-list**: `relojistas.com`, `finetuning.uk`, `webdesign.uk`, `idea.uk`.
- **The rule behind it — STATIC-FIRST**: every domain gets the static Cloudflare
  template by default. A domain leaves the template (and joins the skip-list)
  only when we agree it will host a framework-backed dynamic service — the
  owner's worked examples: webdesign/finetuning-style services, image creation
  (comics, infographics, cats in period costumes), copywriting/personae,
  AI-first products like idea.uk where the multi-agent framework provably beats
  a straight foundation-model call, and data-collection services (vet companies,
  Companies House searches). The skip-list GROWS as services launch; moving a
  domain off static is a deliberate, per-domain change.

## Open questions / blockers

- **Cloudflare token DEAD 2026-08-04** — its IP filter names the old office IPs
  and the line rotated (error 9109). Owner must edit the token: recommend
  REMOVING the IP filter (scopes + expiry remain the protection); the office
  line cannot hold a pin (see RUNBOOK Egress).
- **Nominet TAG name** still needed (password is in place; IP allowlist needs
  the cluster node IPs added — the office IP the owner added has already rotated
  away).
- **Zone-count ceiling** — self-serve accounts have an unpublished zone cap;
  thousands of zones may need a Cloudflare support ticket mid-run. Flagged, not
  yet hit.
