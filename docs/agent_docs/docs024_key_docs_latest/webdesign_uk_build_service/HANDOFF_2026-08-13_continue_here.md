# HANDOFF 2026-08-13 — facts relay LIVE (bot speaks £149 now); the "create the site after paying" loop is the sibling lane's, and here's exactly what it needs

**Start here cold.** Supersedes the CHAT-010 build handoff. Two things happened
this session: (1) the site-facts relay was **activated** and the chat bot now
reads live £149 facts from the database — this fixed a live £1,200-vs-£149
contradiction; (2) the owner asked what's needed to make the chat "create the
site after taking money", which is the **`ai_site_selling_automation` lane's**
territory, already deep in it — this file answers the question and points there,
it does not duplicate their work.

## 0. State in one paragraph

The webdesign.uk chat bot is **live on the £149 offer**, reading its facts from
`site_specs.evidence_base` via the CHAT-010 relay (core-manager `v1.0.1294`,
proven end to end 2026-08-13). Before today it still quoted the retired £1,200
offer while every page said £149. The "chat → pay → build a real site" loop
does **not** exist yet as one connected flow: the payment surface (£149,
PAY-009) is built but **keyless**; the framework build works but is triggered
by hand; hosting on `*.ugg2.com` is proven; dispatch reliability (`bugs_open/239`)
is now fixed+live. The missing links are named in §3.

## 1. What is LIVE and mine, verified 2026-08-13

- **Facts relay activated.** Bot answers "£149, one-off, you approve before you
  pay, no refund", retired-term grep on its replies returns zero. The relay
  serves `evidence_base.facts[].claim` (never `writer_block`).
  - `SITE_FACTS_TOKEN` is in `personae-platform-secrets` (real secret, not git).
  - Box `/etc/webdesign-chat.env` has `FACTS_URL` (the **ClusterIP**
    `10.21.127.41`, because box cluster-DNS doesn't resolve — durability note
    in NOTES) + `FACTS_TOKEN` (same value).
  - Change any fact in `evidence_base` → bot reflects it within 5 min or a
    restart, **no redeploy**. If the relay ever fails the bot **refuses to
    start** rather than revive the compiled-in £1,200 constant.
  - > **CORRECTED 2026-08-13 (evening), same day:** the 13:53Z `v1.0.1295`
    > fleet release **wiped `SITE_FACTS_TOKEN` out of
    > `personae-platform-secrets`** — that secret is terraform-managed
    > (047-base-configs) and every `make release` reconciles its whole data
    > map, so the morning's additive `kubectl patch` could never survive one.
    > Relay 401 from 13:55Z; the bot is alive on last-good facts but ONE
    > RESTART from dead chat. Durable fix committed (the key is now declared
    > IN terraform); the `terraform apply` + core-manager restart are
    > owner-gated — see NOTES (evening entry) and RUNBOOK § "Restoring or
    > rotating the facts-relay token". The "within 5 min, no redeploy" claim
    > above holds only once that lands.
    > **UPDATE 2026-08-14:** cluster half healed — the 246 lane's terraform
    > apply carried the new token in, core-manager restarted, and the relay is
    > proven at the pod (200 with the pod's own token, 401 without). Box half
    > still owed: ONE owner command (NOTES 2026-08-14 entry) copies the tfvars
    > value to `/etc/webdesign-chat.env` and restarts the bot.
    > **RESOLVED 2026-08-14 08:12Z:** owner ran it; bot live-mode on the
    > terraform-owned token, £149 verified at the artefact. The claims above
    > hold again, now release-proof.
- **WireGuard tunnel** box↔cluster: up, proven, `ip_forward` fault fixed
  (LANDMINE + CHAT-010 register). Carries box→cluster today; can carry the
  Stripe webhook inbound (see §3.2).

## 2. The owner's question, answered plainly

> "What do we need to make it create the site after taking money. Perhaps we
> should take money first, I don't know yet."

**The chat bot "stops" because it was only ever built to have the conversation
and collect the demand signal — it has no connection to payment or the build
pipeline.** Those exist as separate, working parts that nobody has wired into
one flow yet. To close the loop you need four links, and one decision:

**The decision — take money first, or build first?** My recommendation: for an
*automated* flow, **take money first.** Reasons, grounded:
- The £149 offer's own terms already say **"no refund"** — pay-first ("pay
  £149, we build it, you get the files") fits that cleanly; pay-after-approval
  makes "no refund" slightly odd (why name a refund you'd never need?).
- Automation removes the human who currently makes a speculative (build-first)
  site safe. An anonymous visitor triggering an unpaid automated build is the
  exact abuse vector `SAAS-001` warns about. **Money-first makes every build
  paid-for and accountable — payment is its own rate-limiter.**
- Your instinct ("take money first") is, I think, right for this path.
- ⚠ **Catch:** the five live pages currently SAY you pay *after* approving
  (the sibling lane's ruling + copy). Flipping the `payment_timing` switch to
  `upfront` is therefore a **copy migration**, not one field. That's the
  sibling lane's to run — talk to them before flipping it.

## 3. The four links, and who owns each

All four are the **`ai_site_selling_automation` lane** (read its
`HANDOFF_2026-08-12_continue_here.md` — it is comprehensive). Named here so the
owner sees the whole chain in one place:

1. **Payment can't happen at all until Stripe keys exist.** PAY-009 is built
   and keyless by design. Owner adds `STRIPE_SECRET_KEY` +
   `STRIPE_WEBHOOK_SECRET` (test mode first) to `personae-platform-secrets`,
   restart auth-service. *This is the first real blocker — nothing charges
   until it's done.*
2. **The payment confirmation (Stripe webhook) has to reach the cluster.** The
   sibling lane's recommended path is "proxy from the webdesign.uk box over the
   existing tunnel" — **that tunnel is now mine and it exists.** The box would
   receive the public HTTPS webhook and `proxy_pass` it inbound to auth-service
   over the tunnel (small nginx `location` block). Avoids a new Ingress or
   Cloudflare tunnel.
3. **Something must trigger the build when payment clears.** Today the framework
   build is manual (seed SQL + dispatch). The automated "webhook → order paid →
   fire the build" seam is the sibling lane's **P4**, design-only. Good news:
   `bugs_open/239` (dispatch silently no-opping) is **fixed and live**, so the
   trigger is trustworthy now in a way it wasn't a week ago.
4. **The chat transcript must become a build brief.** The bot collects the
   conversation but nothing turns it into a `site_specs` seed. Designed
   (transcripts → `site_chat_turns`, briefing-agent → intake), not built.

**Then hosting is free:** a framework-built site's objects land on B2 and serve
at `<slug>.ugg2.com` with zero per-site config (proven wildcard). Preview link
+ ZIP delivery to the customer — ZIP is currently *manual* and now promised in
the £149 copy, so it's owed.

## 4. Falsifiers / re-check before trusting this file

- Bot still on £149 and live-mode: `journalctl -u webdesign-chat | grep facts`
  should show `live mode`; ask it the price.
- The sibling lane has almost certainly moved since 2026-08-13 — read its
  newest handoff, not this summary of it.
- Whether the owner has added the Stripe keys (§3.1) — the whole loop is gated
  on it.
- ClusterIP `10.21.127.41` still core-manager's — if the bot stops fetching
  facts, check the Service IP first (NOTES durability note).
