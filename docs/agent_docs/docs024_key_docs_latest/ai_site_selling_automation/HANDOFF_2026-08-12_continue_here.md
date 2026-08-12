# HANDOFF 2026-08-12 — continue here: the £149 copy migration is DONE (site + register); billing still keyless — SUPERSEDES HANDOFF_2026-08-11b

**Start here cold. Supersedes `HANDOFF_2026-08-11b_continue_here.md`.** Its §1
owner-decision list still stands in full and is reproduced below with two
additions; its §2 work item 1 is **done** and is replaced by §2 here; its §3–§5
still hold except where corrected below. Read order: this file → `NOTES` tail
(2026-08-12, three entries) → `RUNBOOK` "Changing what a site is allowed to say"
→ `PLAN` §1b/§1c (the rulings, do not re-open).

## 0. State in one paragraph

webdesign.uk now sells the ruled **£149 offer** and no longer mentions the
retired £1,200 price, the £75 deposit, the fourteen-day refund window or the
two-rounds revision cap. Both halves went in on 2026-08-12: the **register**
(`evidence_base`, superseded not overwritten — 12 facts, 26 bans,
rewritten `writer_block`) and the **copy**, regenerated **through the framework**
as five `content_rewrite` items with `mode=edit_live`. The £149 payment surface
(**PAY-009**) is unchanged from 08-11: built, deployed, **keyless by design** —
no payment can be taken until the owner creates the Stripe secrets. So the site
now advertises an offer the platform cannot yet charge for, which is the right
way round but worth knowing.

## 1. OWNER DECISIONS OPEN (unchanged from 08-11b unless marked)

1. **Stripe keys — bites when you want to take a first payment.** A
   **restricted secret key** (Checkout Sessions:Write only) and the **webhook
   signing secret** for `POST /api/v1/billing/webhooks/stripe`, event
   `checkout.session.completed`, into `personae-platform-secrets` as
   `STRIPE_SECRET_KEY` / `STRIPE_WEBHOOK_SECRET`, then restart auth-service.
   Test-mode first: test and live are separate accounts with separate secrets.
2. **Webhook public exposure — bites with the keys.** Nothing routes into the
   cluster today. (a) proxy from the webdesign.uk box over the existing tunnel
   (recommended, needs the sibling lane); (b) Ingress + TLS on a hostname;
   (c) a Cloudflare tunnel.
3. **Unify-or-deprecate the old subscription scaffold — bites at first sale.**
   Recommended: retire the scaffold's create/update surface once the first £149
   sale has gone through.
4. **Payment timing switch — NO LONGER FREE TO FLIP. ⚠ CHANGED 2026-08-12.**
   It is still one field (`billing_settings.payment_timing`, live default
   `after_approval`), but the site's copy now **states** it: every one of the
   five pages says the customer pays after approving the site. Flipping to
   `upfront` makes that copy false on a live site. Fact
   `payment_after_approval` names the SQL as its source so the coupling is
   visible, but nothing enforces it. **Flipping the switch is now a copy
   migration, not a one-field UPDATE.**
5. **NEW — two commercial terms need a word from you before the site is fully
   consistent.** Both are live on the site today and neither is in the register:
   - **"Anything that's our mistake, we fix at no cost."** This was attested
     2026-07-29 under the £1,200 offer. I did NOT carry it into the £149
     register (at £149 with no ongoing service it is an open-ended liability),
     but `edit_live` preserved the sentence on `how-it-works`, `what-you-get`
     and `faq`, and nothing objected. **Either re-attest it and I put the fact
     back, or say it goes and I queue one more rewrite round.**
   - **"three or four days"** is carried over unre-attested from the £1,200
     offer. It is flagged in its own `source` field. Confirm it or give a new
     number.
6. **Standing external asks, unchanged**: Nominet second-tag application; the
   three registrar API keys; Phase 6 cutover review (sibling lane).

## 2. Work list (next session, in order)

1. **Finish what the migration exposed** (small, and it is the honest tail):
   - the owner answers on §1.5, then one more `content_rewrite` round if the
     defect-fix promise goes;
   - **`what-you-get`'s work item says `failed` and the page is CORRECT AND
     LIVE.** The failure is the spawn→call handshake race firing after
     `deploy_page` had done its work: components written, `pages.deployed_at`
     set, the commit in `gqls/vm-sites`, the served page updated 17:16:20Z, and
     the content scans clean. Left `failed` on purpose — it truthfully records
     that result delivery failed, and it will NOT retry (`find_dispatchable_site`
     takes only `triaged`/`approved`; both re-triage sweeps are disabled).
     **Do not re-run it**: that would rewrite an already-correct page.
   - the **home page CTA still reads "Tell us what you need. We'll tell you
     what it costs"**, which is now odd on a page that states £149 three times.
     Not false, so no gate catches it. Worth one section_edit.
   - the **archived** `index-rejected-v1-20260806` page carries 14 of the
     register's findings and will keep generating `claims_unverified` noise. It
     is served to nobody. Either exclude archived pages from the checker (a
     platform change, wider than this lane) or delete the page.
2. **Queue + submission gate** (PLAN §2.6, designed): derived occupancy from
   open work items, owner-settable capacity (3–4) + `queue_paused`, gate at
   every intake door, non-binding wait note. **The site now promises this
   behaviour in copy** ("we build only a few sites at a time, when we're full
   submissions close") — so the mechanism is now owed by the page, not just by
   the plan. Platform code → council + register.
3. **Admin FE voucher screen** (issue/list vouchers, read orders, flip payment
   timing) against `/api/v1/admin/billing/*` on auth-service directly. First
   real voucher POST doubles as post-roll recipe step 4.
4. **ZIP delivery** — **now promised on the live site** (`delivery_preview_and_zip`),
   and fulfilment today is MANUAL (pull the B2 prefix, zip it, hand it over).
   This moved up in urgency the moment the copy shipped.
5. **Transcripts → `site_chat_turns`** (PLAN §2.3, designed) — unchanged.
6. **Trigger seam (P4)** — design-only; check `bugs_open/239`'s fixed-AND-LIVE
   state before treating dispatch as trustworthy.

## 3. What is live vs inert

- **LIVE**: the £149 register (`site_specs` row `6f9e8e7c`, pinned) and its 26
  banned patterns; the £149 copy on **all five pages** — index, faq,
  how-it-works, what-you-get and the brief-starter guide, verified at the served
  artefact 17:16Z with zero retired terms and £149 on every one; migration 391's
  four billing tables; the billing routes (admin 401s, webhook 503s); the
  payment-timing switch; ADM-011 customers API + tab.
- **INERT / OWED**: any actual payment (keyless by design); admin FE voucher
  screen; webhook public exposure; automated ZIP delivery; the queue mechanism
  the copy now describes.
- **ARCHIVED, RESTORABLE**: the £1,200 register is `site_specs` row `bccf42a7`
  (`is_current=false`) — one UPDATE from being live again; its copy is in
  `snapshot_2026-08-11_gbp1200_offer/`.

## 3b. ⚠ THE MIGRATION BROKE THE CTA BUTTONS, AND THE REPAIR IS A HAND PATCH

Read this before any rebuild of webdesign.uk. The regeneration dropped
`cta_url` / `primary_cta_url` / `secondary_cta_url` from every hero and
call-to-action block on the four offer pages — **14 anchors, 7 components** —
because both templates gate the anchor on the URL rather than the label, so
each button rendered as nothing. Repaired and live at 17:37Z (index 2, faq 4,
how-it-works 4, what-you-get 4).

- **The repair is in `rendered_html` as well as `content_data`, by necessity.**
  A `page_rerender` dispatched after the `content_data` fix still produced no
  buttons: these fields are `source: "renderer"`, and the render path
  re-resolves them to nothing rather than reading the stored value.
- **So the next rebuild of these pages will delete the buttons again**, file a
  `page_divergence_overwritten` item, and look successful. `SQL_2026-08-12e`
  re-applies the anchors and is idempotent; the deployed files also need the
  same patch pushed to `gqls/vm-sites` (see `b538295` for the shape).
- **The mechanism is OPEN — my hypothesis was REFUTED and the re-run is owed.**
  `090` run `97ef39f0-19df-4935-834d-c80514fbc43e` returned **REFUTED** on the
  claim that renderer-sourced fields fall outside `carryStored`. **But the run
  is not decisive either**, and the reason is my error: I repaired
  `content_data` at 17:23 and started the run at 17:39, so its citations are
  the values *I had put back*. It measured a repaired system. The evidence of
  the loss is in `page_component_history`, and I never pointed it there.
  **OWED: a re-run authored against that history for the 16:37–17:23 window,
  whose symptom states plainly that the live rows were repaired at 17:23.**
  Until then, treat the cause as unknown — the DAMAGE is measured and certain,
  the MECHANISM is not.
  What the run did give, and it corrects me: **`bugs_open/238` as tracked in
  the codebase is the dead-URL-CONTROL defect** (a section that renders while
  leaving a URL attribute empty, recorded non-fatally on the rerender path),
  not a content_data key-loss case. `next_scope`: `dead_url_guard.go`,
  `emitSectionDeadControlItem`, `recordDeadURLControls`. Start there.
  Separately and still true: 238's §8 banner is **stale** — it says the carry
  is "inert until the next roll"; the carry is live (agent-chassis `v1.0.1291`
  ← `da5a7eb8f`, merge-base verified with controls both ways).
- **`required_links` is now declared on all five pages**, so
  `gate_page_links.py --domain webdesign.uk` covers the whole site rather than
  the guide alone. Run it after any rewrite — and run its `--self-test` first.
- **The invariant check that would have caught this**, and should precede and
  follow any regeneration:
  ```sql
  SELECT p.name, pc.slot_name,
         (SELECT count(*) FROM regexp_matches(pc.rendered_html,'href="','g')) AS links
  FROM page_components pc JOIN pages p ON p.id=pc.page_id
  WHERE p.site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f' ORDER BY p.name, pc.position;
  ```

## 4. Landmines for this work (08-11 §5 still holds; these are new)

- **`EvidenceFact.Value` is `*float64`.** A non-numeric `value` in any fact
  fails the whole `evidence_base` unmarshal and **silently disarms the claims
  layer for that site**. `cmd/claimscan` is the only place it is visible. Full
  entry in `LANDMINES.md` (footprint `site_specs` / `claims.go` / `claimscan`).
- **Bare `no` is not a negation cue.** With `refund` banned, *"there is no
  refund"* blocks the page and *"we do not offer refunds"* passes. The required
  phrasing is in `writer_block` and in the ban's own `reason`.
- **A register enforces its BANS and merely hopes for its FACTS.** Removing a
  fact does not remove the claim from the page — see §1.5. Anything you want
  gone must be named in `banned_claims`.
- **A regeneration silently deletes what it was not asked about.** The claims
  scan, the byte-delta check, the retired-term grep and the served fetch all
  passed while every CTA button on the site was gone — each answers a narrower
  question than "is the page still right". Diff the INVARIANTS (link count,
  image count, component count) as a matched before/after pair, and treat what
  you did not intend to change as the finding. §3b, and `WRONG_CALLS.md`.
- **`mode=edit_live` is load-bearing** on `content_rewrite` items
  (`bugs_open/178`). Without it the writer never sees the page's current prose
  and fabricates a replacement.
- **`build_status='deployed'` leads the artefact by up to ~5 minutes** (box
  pulls on a timer; measured 74s and 285s the same afternoon). Check
  `last-modified`, and confirm the repo has the file (`gh api repos/gqls/vm-sites/...`)
  before concluding a deploy failed.
- **The shopfront is `preview.webdesign.uk`.** The apex 302s to
  webdesign.co.uk by design (owner-confirmed 2026-08-10) and every path
  collapses to that different site's homepage. Checking "the live site" on the
  apex measures the wrong site.
- Unchanged from 08-11b: the council gate loops on shipped-code submissions;
  `2>/dev/null` on psql/kubectl turns real errors into empty results; voucher
  variants are hard-coded to £10/£55; `billing_orders.status` has no refunded
  state on purpose.

## 5. Falsifiers (re-check before trusting this file)

A newer handoff here; whether the owner has answered §1.5; the sibling lane's site-facts relay (it
serves this register's `facts` to the chat bot — if it has shipped, the bot's
compiled-in `systemPromptFacts` are no longer the source of truth, and if it
has not, the bot is still quoting the £1,200 offer); `bugs_open/239`'s
fixed-AND-LIVE state; whether the Stripe keys have appeared in
`personae-platform-secrets`.
