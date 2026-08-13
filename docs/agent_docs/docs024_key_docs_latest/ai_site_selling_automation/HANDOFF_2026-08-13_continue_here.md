# HANDOFF 2026-08-13 — continue here: offer hardened + brand canary LIVE; imagery rollout and one parked copy item are next — SUPERSEDES HANDOFF_2026-08-12

**Start here cold. Supersedes `HANDOFF_2026-08-12_continue_here.md`** (its §3b
button warning is now RESOLVED for this site by locks; its §1 owner-decision
list is updated below; everything else holds). Read order: this file → PLAN
§1b/§1c/**§1d** (the rulings — §1d supersedes §1b.2 on changes) → NOTES tail
(2026-08-12 evening through 2026-08-13) → `RUNBOOK` §"Changing what a site is
allowed to say".

## 0. State in one paragraph

webdesign.uk sells the ruled **take-it-or-leave-it £149 offer** on all five
pages: no changes included (stat tile on the home page reads "0 Changes
included"), no refund stated, files handed over as a ZIP and **yours to edit —
editing beats a blank page** is on the site as the better half of the deal.
The scrappy-process/good-result **brand is live at the homepage hero**: the
framework generated a Heath Robinson contraption on kraft, gold only on the
output page, from `imagery_style_guide` + the plan prompt — verified by eye at
the served artefact 2026-08-13. The site's 8 CTA-bearing components are
`lock_type='permanent'` (SQL_2026-08-12k), proven to survive a real rewrite;
the fleet-wide button defect is `bugs_open/268`, **owned by its own thread**
(cold-start `bugfix_268_cta_buttons_fleet/HANDOFF_2026-08-13_start_here.md`).
Payment (PAY-009) remains built + deployed + **keyless by design**. Fresh
chassis build `v1.0.1295` (← `69612d692`) rolled 2026-08-13 ~13:53Z; the 268
fix paths were untouched at that stamp.

## 1. OWNER DECISIONS OPEN

1. **Stripe keys** — restricted key (Checkout Sessions:Write) + webhook signing
   secret → `personae-platform-secrets` as `STRIPE_SECRET_KEY` /
   `STRIPE_WEBHOOK_SECRET`, restart auth-service. Test-mode first.
2. **Webhook public exposure** — (a) proxy via the webdesign.uk box
   (recommended; needs sibling lane), (b) Ingress+TLS, (c) Cloudflare tunnel.
3. **Old subscription scaffold** — deprecate its write surface after the first
   real sale.
4. **Payment timing switch is a COPY MIGRATION, not a field flip** — all five
   pages state pay-after-approval.
5. **Look at the new homepage hero** (live now) and say go/no-go on the brand →
   unlocks work item 1 below.
6. **"Three or four days"** is still carried over un-re-attested from the
   £1,200 offer (flagged in fact `build_duration`). Confirm or replace.
   (The defect-fix promise question is RESOLVED by §1d: no corrections at £149;
   the "our own mistake" line was removed by the tiloi rewrite.)
7. Standing: Nominet TAG name + 5 allowlist IPs; registrar keys later; Phase 6
   cutover review (sibling lane).

## 2. Work list (next session, in order)

1. **Imagery rollout behind the owner's go** (framework only — the owner
   explicitly does NOT want CLI-generated images):
   - Re-prompt + re-file `needs_imagery` for `hero_faq`, `hero_how_it_works`,
     `hero_what_you_get`, `hero_contact` and the three `index:2` icons, each
     with a NEW SUBJECT from the motif family (goose+golden egg, cardboard
     box/flat-pack, pallets, trade counter — junkyard is NOT in the set; the
     guide's `avoid` bans dereliction). The canary recipe is in NOTES
     2026-08-13: item spec carries the prompt; update the CURRENT plan's
     `site_plan_imagery` row in the same transaction so plan and item cannot
     drift; same `asset_key` ⇒ same served path ⇒ **locked heroes need no
     write** (proven: hero-home.jpg replaced in place, served 10:45:02Z).
   - **`design_intent` palette/layout pass** — still describes the old
     "well-printed document" brand (near-white, forest accent, "No hero image,
     no full-viewport splash"). Kraft-ground art on that page reads borrowed.
     Show the owner proposed reference_values BEFORE repainting; beware the
     webdesign colour-churn landmine (pin via
     `design_intent.palette.reference_values`).
2. **PARKED, needs a different mechanism or an owner call: the FAQ still does
   not name the six third-party services.** THREE rounds failed. Rounds 1–2
   were OUR defect (names were in `facts[]`, not in `writer_block` — the wire;
   fixed by `SQL_2026-08-13a`). **Round 3 failed WITH the names in the wire,
   cause unknown** — do not run a round 4 blind. Options: (a) inspect the
   faq component's `content_data` structure and add the answer as SECTION DATA
   rather than writer prose; (b) diagnose why the writer omits them (read the
   rendered prompt in `llm_call_log` — the 2026-08-09 method); (c) owner drops
   the ask. The six: Cloudflare Pages, Netlify, Fathom Analytics, Plausible,
   Formspree, Basin (all in `allowed_entities` + writer_block).
3. **Queue + submission gate** (PLAN §2.6) — the copy now PROMISES it
   ("we close submissions until a slot opens"). Platform code → council.
4. **ZIP delivery** — promised on the site, currently manual.
5. **Voucher admin screen**; **transcripts → site_chat_turns**; **trigger seam
   (P4)** — unchanged.

## 3. Standing constraints (do not relearn)

- **The 8 locked components**: unlock deliberately, edit, re-lock, and re-run
  `SQL_2026-08-12d`+`e` + `gate_page_links.py` after — while `bugs_open/268`
  is open, ANY regeneration drops CTA URLs on UNLOCKED components.
- **Before/after every regeneration**: the href-count invariant query (RUNBOOK
  §4) as a matched pair, plus `gate_page_links.py --domain webdesign.uk`
  (all five pages carry `required_links`; run `--self-test` first).
- **`writer_block` is the wire; `facts[]` is bookkeeping** — anything the
  writer must SAY goes into writer_block text itself. Twice-proven here.
- **`EvidenceFact.Value` is `*float64`** — a string value silently disarms the
  site's whole claims layer. Always `cmd/claimscan` a candidate register
  against the LIVE corpus AND against the current register as control.
- **Negation guard**: "we do not offer refunds" passes; "there is no refund"
  blocks the page. In every ban's `reason` + writer_block.
- **Register edit history**: every `evidence_base`/`imagery_style_guide`
  change SUPERSEDES (never in-place), inherits `pinned`, worked SQL files
  `SQL_2026-08-12_*` → `SQL_2026-08-13a` in this directory.
- **`content_direction` belongs to the fleet voice lane** (live 08-12/13) —
  voice changes go in writer_block, not there.
- **The sibling lane** owns the contact chat box (locked, theirs) and the box
  infra; coordination notes flow through their NOTES file.
- A `failed` work item may be finished work (handshake race — 3 sightings);
  verify at the artefact in both directions.

## 4. Falsifiers

A newer handoff here; PLAN §1d unchanged?; the 268 thread's progress
(`bugs_open/268` tail + `who-owns.py 268` + live transcripts — if their fix is
LIVE, the locks can come off: that unlock is THIS lane's step); whether the
owner has looked at the hero (§1.5); the served hero still the contraption
(`curl -sSI https://preview.webdesign.uk/assets/images/hero-home.jpg` —
last-modified 2026-08-13 10:45:02Z as of writing); Stripe keys present?;
chassis stamp moved past `v1.0.1295`? (image label, not git).
