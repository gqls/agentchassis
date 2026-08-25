# RUNBOOK — taking webdesign.uk live (unparking the shopfront)

**Written 2026-08-25, the day go-live was ruled GO and then DEFERRED the same day:** the
owner reviewed the site and wants a substantial copy revision and some design revision
first (§1). Every step below stays valid; **every dated fact must be re-verified on the
day** — the falsifiers are listed with each step. Ruling trail: `RFC_054` §5 (the
boundary review, RULED), register **SYS-094** (the standing exposure pattern).

## 0. What "live" means here — and what it does NOT change

Today `webdesign.uk/*` answers a **302 to webdesign.co.uk** from a Cloudflare page rule
put there to park the domain. The shopfront is fully built behind it: the box's nginx
vhost serves the framework-built pages, the cloudflared tunnel routes apex/www/preview
to it, and `preview.webdesign.uk` serves the real site right now.

Going live = **removing that page rule**. Nothing else: no deploy, no code, no restart.

At that instant two things become internet-reachable (both reviewed by RFC_054, ruled
2026-08-25 — no further review owed):
- **the shopfront** (static files + `/api/chat` to the box-local chat service);
- **`/stripe/webhook`** → auth-service in the cluster — safe by design: Stripe
  authenticates by HMAC over the raw bytes, and until the Stripe keys exist it answers
  an honest **503 "billing provider not configured"**.

Going live does **NOT** open ordering. That still needs, separately: Stripe keys +
webhook edge exception (owner-deferred), the second-click confirmation page
(`DECISION_2026-08-24_confirmation_needs_a_second_click.md` — gates the first delivery
email), and the delivery email itself. The site being visible changes none of these.

## 1. GATES — all must be clear before the unpark

1. **Copy + design revision (the gate that deferred it, owner 2026-08-25).** The owner's
   revision list drives this; revisions go **through the framework** (owner rulings
   2026-08-04 and 2026-08-06 — the framework writes the content, not us; copy changes
   via `evidence_base` facts / the register + rebuilds, verified at the served page).
   Known content items already on file before the owner's pass
   (HANDOFF_2026-08-21 §6): no terms page while the copy points at "the full terms";
   contact email `webdesign@contactforsales.com` (domain mismatch); `bugs_open/299`
   (home CTA producer question); the what-you-get shrink gate.
2. **The "Not active yet" label** (hand-placed by owner instruction 2026-08-25,
   vm-sites `444205b`, above BOTH CTAs, marked
   `data-note="hand-placed 2026-08-25, temporary until ordering opens"`). If ordering
   is still closed on go-live day, **confirm it is still present** at
   `preview.webdesign.uk` — any framework rebuild of `index` removes it silently, and
   the copy-revision pass in gate 1 WILL rebuild `index`, so expect to re-place it
   (same two insertion points: above the hero `btn-primary` and above `cta-buttons`).
   If ordering is open, remove it instead (§5).
3. **Safety counters still zero, or re-reasoned:**
   ```sql
   SELECT (SELECT count(*) FROM customer_access_tokens),
          (SELECT count(*) FROM sites WHERE handed_over_at IS NOT NULL),
          (SELECT count(*) FROM sites WHERE transfer_confirmed_at IS NOT NULL);
   ```
   `0|0|0` as of 2026-08-25. Non-zero does not block go-live, but expires every
   "nothing at risk" line in RFC_054 — re-read its §3 before proceeding.
4. **Edge still in the known-safe state** (LANDMINES "A CLOUDFLARE PAGE RULE PUT THERE
   FOR PARKING…" — its check verbatim):
   ```bash
   for u in https://webdesign.uk/c/x https://www.webdesign.uk/c/x https://preview.webdesign.uk/c/x; do
     printf '%-40s ' "$u"; curl -sS -o /dev/null -w 'code=%{http_code} redirect=%{redirect_url}\n' "$u"
   done
   curl -sS -o /dev/null -w 'control preview/ -> %{http_code}\n' https://preview.webdesign.uk/
   ```
   Expect: apex+www `302 https://webdesign.co.uk/`, preview `404`, control `200`.

## 2. THE UNPARK

> **CORRECTED 2026-08-25:** this section said "owner-executed (sessions cannot SSH to
> the box)", inherited from `RUNBOOK_links_host_box_steps.md`. **Measured false the
> same day**: `ssh -i ~/.ssh/webdesign_box_ed25519 root@webdesign.vs.mythic-beasts.com`
> works from the working machine (key present since 2026-08-04). Sessions CAN run the
> box half; the Cloudflare page rule remains owner-only (dashboard, no API credential
> in session reach).

**Step 1 — prove the box's APPLIED apex vhost is the `/c/`-free version.** The repo copy
(`box/webdesign.uk.nginx`) is `/c/`-free since 2026-08-24, but the repo is not the box,
and this cannot be probed from outside while the page rule swallows the apex. Run:

```bash
ssh -i ~/.ssh/webdesign_box_ed25519 root@webdesign.vs.mythic-beasts.com \
  "grep -c 'location /c/' /etc/nginx/sites-enabled/webdesign.uk"
```

**Expect `0`.** ✅ **VERIFIED 0 on 2026-08-25** (re-run on the day regardless: the box
config can move). If it prints anything else: copy the repo's
`webdesign_uk_build_service/box/webdesign.uk.nginx` onto the box as
`/etc/nginx/sites-available/webdesign.uk`, then `nginx -t && systemctl reload nginx`,
and re-run the grep. Do NOT proceed to step 2 until it reads 0 — otherwise the unpark
exposes a state-mutating GET (`/c/`) as a side effect, which is the exact trap the
LANDMINES entry exists for.

**Step 2 — remove the parking page rule.** Cloudflare dashboard → `webdesign.uk` zone →
the page rule that 302s `webdesign.uk/*` to `webdesign.co.uk`. Backup of the rule:
`PAGERULES_backup_2026-08-08.json`. Removal is instant and needs nothing on the box.

## 3. POST-UNPARK VERIFICATION — session-executable, from outside, same hour

| probe | expect | what a failure means |
|---|---|---|
| `curl -sS https://webdesign.uk/ \| wc -c` and grep the body | 200, ~34KB, contains `A complete website for your business` — and `Not active yet` ×2 if ordering still closed | tiny body / `/lander` redirect = still parked or wrong origin (the cookly.uk trap: a parking 200 looks clean) |
| same for `https://www.webdesign.uk/` | same | www not covered by the vhost/tunnel |
| `https://webdesign.uk/c/x` | **404 from nginx static** (`try_files` miss), NOT core-manager's token-404 | a core-manager-shaped 404 body = the box is running the OLD vhost — put the page rule BACK, fix the vhost, re-remove |
| `curl -X POST https://webdesign.uk/stripe/webhook -d x=y` | **503** (keyless honest state) while Stripe keys absent; after keys land, a 4xx signature refusal | 502/timeout = wg leg or Service name broken; 200 = investigate immediately |
| `https://webdesign.uk/api/chat` + a real chat turn on the page | chat answers | chat service or its facts fetch broken |
| controls: `preview.webdesign.uk/` 200 · `links.webdesign.uk/c/x` 404 · `admin.apis.uk` healthy | unchanged | the change leaked wider than the page rule |

## 4. BOOKKEEPING OWED AFTER VERIFICATION (whoever unparks owns these)

1. **Update the LANDMINES entry** "A CLOUDFLARE PAGE RULE PUT THERE FOR PARKING IS THE
   ONLY THING KEEPING TWO IN-CLUSTER ROUTES OFF THE INTERNET" — the page rule is gone,
   so its "safe today reads" table inverts; append a dated UPDATE recording the removal
   and pointing at this runbook + SYS-094. Then `./scripts/landmines-verify-dispatch.sh`
   (NOT `landmines-sync.py --apply` first — that consumes the new-entry status and the
   verifier never checks it).
2. **The sitemap-census LANDMINES entry** (the 2026-08-24 "a status-code census reports
   11, true figure 8 of 28" one) names `webdesign.uk` **302s** as one of its three
   not-ours shapes — that example is stale the moment the site serves; annotate it.
3. Lane NOTES + `README_where_we_are` entries; MEMORY_workstreams line for this lane.
4. If the delivery email work resumes: it mints links on `links.webdesign.uk` (the
   canonical emailed-links host), NOT on the shopfront — unchanged by go-live.

## 5. LABEL REMOVAL — when ordering opens

Find it: `grep -rn 'data-note="hand-placed' ~/projects/vm-sites/webdesign.uk/`.
Either delete the two `<p>` lines and push vm-sites (box syncs ≤5 min), or let the next
framework rebuild of `index` remove it — **but verify at the served page either way**
(`curl -sS https://webdesign.uk/ | grep -c 'Not active yet'` → 0).

## References

- `../architecture_review/RFC_054_public_cluster_exposure_boundary_review.md` §5 — the rulings
- Register **SYS-094** (`docs026_concept_register/register/system-architecture.md`) — the standing exposure pattern
- LANDMINES: the parking-page-rule entry (the trap this runbook operationalises) and the box hostname inventory (7 hosts, 3 ports)
- `RUNBOOK_links_host_box_steps.md` — the sibling box bring-up (done)
- `DECISION_2026-08-24_confirmation_needs_a_second_click.md` — the second-click page spec (gates delivery email, not go-live)
