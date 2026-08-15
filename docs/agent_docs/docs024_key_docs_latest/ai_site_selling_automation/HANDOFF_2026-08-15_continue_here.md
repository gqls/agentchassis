# HANDOFF 2026-08-15 — continue here: the shop is COMPLETE; next is the delivery/editor build (own workstream) and the owner's payment keys — SUPERSEDES HANDOFF_2026-08-13

**Start here cold. Supersedes `HANDOFF_2026-08-13_continue_here.md`** (its §3b
CTA warning is RESOLVED — the fleet fix is live via the 268 thread and the
locks stand as defence-in-depth; its work list is superseded below). Read
order: this file → `SUMMARY_2026-08-15` (the read-aloud state) →
**`../site_delivery_and_editor/PLAN_2026-08-14_site_delivery_and_editor.md`**
if you are building the next stage → NOTES tail (2026-08-14 entries) for
evidence. PLAN §1b/§1c/§1d hold; §1d supersedes §1b.2 (no changes included).

## 0. State in one paragraph

webdesign.uk is **complete against every ruling**: take-it-or-leave-it £149
copy on all five pages (no changes included + the files-are-yours argument, no
refund stated, payment after approval, "usually ready the next day", six
third-party services named in the FAQ); the kraft-and-ink brand live end to
end (five verified hero illustrations + the palette, all seven values held
byte-exact; buttons deliberately INK, owner-kept); every commercial claim
traces to `evidence_base` and every retired figure is mechanically banned.
**The CTA locks are OFF** — the 268 thread's fleet fix is LIVE
(v1.0.1298+, council-approved) and their final unlock step has run: the only
permanent lock remaining is the sibling lane's chat box, and the site's link
set survived the unlock (32 hrefs, gate 5/5, checked at handoff time — the
platform fix is now the standing protection). Payment (PAY-009)
remains built + deployed + **keyless by design**. The delivery/editor plan is
owner-approved in `site_delivery_and_editor/` with Phase 1 (next-day promise)
DONE; Phases 2–6 unstarted.

## 1. OWNER'S KEYS TO TURN (unchanged, still the gate to a first sale)

1. **Stripe keys**: restricted key (Checkout Sessions:Write) + webhook signing
   secret → `personae-platform-secrets` as `STRIPE_SECRET_KEY` /
   `STRIPE_WEBHOOK_SECRET`; restart auth-service; test-mode first.
2. **Webhook exposure**: (a) proxy via the webdesign.uk box (recommended,
   needs sibling lane), (b) Ingress+TLS, (c) Cloudflare tunnel.
3. **Payment-timing switch is a copy migration** — all five pages state
   pay-after-approval.
4. Standing: Nominet TAG name + 5 allowlist IPs; registrar keys later;
   Phase 6 cutover review (sibling lane).

## 2. Work list (next session)

1. **The delivery/editor build — Phases 2–6.** Cold-start:
   `site_delivery_and_editor/PLAN_2026-08-14_site_delivery_and_editor.md` +
   its NOTES. Sequence: publish seam (CF Pages Direct Upload primary,
   provider-agnostic; canary on a quiet portfolio site, **NOT webdesign.uk**)
   → ZIP deliverable (`archive/zip` first use; stream, never truncate) →
   handover state (`sites.handed_over_at`, gates ONLY editor access) →
   magic-link customer auth (clone `sitefacts.go`; add `"customer"` to
   `humanLockSources`) → the editor (extract `HandleUpdateComponent`'s tx
   into `applyComponentEdit` with server-side field-merge; box service port
   8083; **site_id always from the session — the cross-tenant probe is the
   acceptance**). One council run per phase; register entries per the PLAN
   roll-up; every new field opt-in default-OFF.
2. **Queue + submission gate** (PLAN §2.6) — the copy promises it.
3. **Voucher admin screen**; **transcripts → site_chat_turns**; **trigger
   seam (P4**, check `bugs_open/239` state first**)**.
4. Watch items, not work: the 268 thread's final unlock of webdesign.uk
   (theirs); `bugs_open/271` (dead `content_guidance` — steer writers via
   writer_block ONLY until fixed).

## 3. Register / claims state (the source of truth for the copy)

Current `evidence_base` (pinned): 15 facts / 28 bans / writer_block 8,173
chars (verified live at handoff time). The trail is the `SQL_2026-08-12*` → `SQL_2026-08-14b` files in this
directory — every change is a SUPERSEDE (never in-place), inherits `pinned`,
and is claimscan-tested against the live corpus AND the current register as
control before writing. Facts of note: `no_changes_included` +
`yours_to_change` (state them in that order), `no_refund` (phrase it "we do
not offer refunds" — bare "no" is not a negation cue and blocks the page),
`third_party_options` (six names, also in writer_block's ENUMERATION — the
"and nothing else" list is the law), `build_duration` ("usually ready the
next day"), `payment_after_approval` (names billing_settings as its source).

## 4. Landmines for this lane (each cost real time this week)

- **`EvidenceFact.Value` is `*float64`** — a string value silently disarms
  the site's whole claims layer; `cmd/claimscan` is the only place it shows.
- **writer_block is the wire; facts[] is bookkeeping; the whitelist
  ENUMERATION is the law** — an instruction outside all three does nothing,
  and `spec.content_guidance` does nothing EVER (bugs_open/271: no reader).
- **Count hrefs before/after ANY rewrite** (matched pair; RUNBOOK §4) and run
  `gate_page_links.py` (all five pages carry `required_links`; `--self-test`
  first). The guide's `/what-you-get.html` link has been dropped by three
  separate rewrites — restore recipe in NOTES 2026-08-14 (night).
- **A `failed` work item may be finished work** (handshake race, five
  sightings) and **`complete` is not deployed** (box sync 1–30 min; check
  `last-modified`, confirm the repo copy first).
- **Locks are OFF on the CTA components** (268's unlock ran); the surgical
  direct-SQL pattern for deterministic text fixes still applies
  (content_data AND rendered_html, assert hrefs unchanged), and re-locking
  via SQL_2026-08-12k remains available if the fleet fix ever regresses.
- **Palette changes**: write all FOUR colour copies in one transaction
  (design_intent + content_data.color_scheme + palettes + style_collections,
  worked SQL `SQL_2026-08-14_kraft_palette.sql`); reference_values are
  ADVISORY — acceptance is the SERVED stylesheet vs a baseline; don't stack
  duplicate design runs (a discovery prio-10 item front-runs yours).
- The imagery pipeline generates NOTHING on a style-guide change — detection
  keys on ABSENCE; regeneration = plan-row prompt + `needs_imagery` item
  (canary recipe NOTES 2026-08-13), same `asset_key` ⇒ same path ⇒ locked
  heroes untouched.

## 5. Falsifiers (re-check before trusting this file)

A newer handoff here or in `site_delivery_and_editor/`; whether Stripe keys
exist in `personae-platform-secrets`; the lock state
(`SELECT count(*) FROM page_components … WHERE lock_type='permanent'` —
**1** as of writing: only the sibling's chat box; the 268 unlock has run);
`bugs_open/271`'s state before relying on item-spec steering; the served
site itself (`preview.webdesign.uk` — the apex 302s to webdesign.co.uk BY
DESIGN); chassis stamp (image label, not git; per service, not fleet).
