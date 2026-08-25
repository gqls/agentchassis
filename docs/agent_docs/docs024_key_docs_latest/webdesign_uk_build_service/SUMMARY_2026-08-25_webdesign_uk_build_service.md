# SUMMARY 2026-08-25 — the launch paused on purpose: a snapshot to return to

**Why this summary exists (owner request, 2026-08-25):** we stepped back from go-live
today with two big product ideas under reconsideration. This file records where the
whole webdesign.uk launch process stands at that moment, so that if we choose NOT to go
ahead with the new ideas, this is the point we return to — and the go-live becomes a
checklist (`RUNBOOK_go_live_webdesign_uk.md`), not an investigation.

## What we're trying to do

Sell complete, framework-built websites at webdesign.uk for £149, paid in full before
the build. One pass, no approval stage, no changes, no refunds. The customer gets a ZIP
of the finished files to keep and the site live at an address we provide for about a
month; they can rent a domain from us at £10 a month or buy one for £200 (.co.uk and
.uk only). Launching means three separable things: the shopfront visible at its own
name; a way to take money; and a way to deliver what was bought.

## Where we've come from

The commercial position was settled and attested through August (pay-upfront 08-17;
TLD scope, registrant, content policy and owner-review rulings 08-21; build duration
"three or four days" 08-22). The shopfront itself has been built by the framework and
serving at `preview.webdesign.uk` behind a Cloudflare parking rule that 302s the apex
to webdesign.co.uk. The emailed-links door (`links.webdesign.uk`) went live hardened on
08-24. The public-exposure boundary review the architecture seat asked for (`RFC_054`)
was ruled by the owner on 08-25: the two-door pattern is now the standing pattern
(register **SYS-094**), the delivery-only listener is approved to build, and
`/stripe/webhook`'s door is pre-reviewed. Go-live was ruled GO the morning of 08-25,
with a hand-placed "Not active yet" label above both CTAs (vm-sites `444205b`) — then
DEFERRED the same day, first for a copy and design revision, then further by the two
reconsiderations below.

## What we've done — the state of every launch piece, dated 2026-08-25

**Live and verified:**
- The shopfront at `preview.webdesign.uk` (8 pages, framework-built) with the chat bot
  answering from the attested register (24 facts / 34 bans as of 08-21 — two lanes
  write it, recount before trusting).
- `links.webdesign.uk` — the canonical emailed-links host, hardened + rate-limited,
  verified from outside 08-24 (404/404/200, hammer → 429s).
- The owner console `admin.apis.uk` (Access-gated) with the Builds screen (08-24 roll).
- The confirm-transfer route `/c/<token>` in the cluster with the prefetch guard live
  (v1.0.1332); reachable only via the links host.
- The WireGuard egress fence (postgres proven blocked); zero Ingress objects.
- Safety counters: `customer_access_tokens` **0**, handed_over/confirmed **0/0**
  (re-checked 08-25).

**Built but not reachable:** the apex vhost on the box, `/c/`-free, carrying the
shopfront + `/stripe/webhook` (honest 503 until Stripe keys exist) — one page-rule
removal away, procedure in `RUNBOOK_go_live_webdesign_uk.md`.

**Designed but not built:** the second-click confirmation page (spec:
`DECISION_2026-08-24_confirmation_needs_a_second_click.md` — GET renders, POST
confirms; **gates the first delivery email**); the `/d/` ZIP download link (pre-mint
and refresh, with the stale-link work item); the delivery email + weekly chase +
retraction; the owner-review-before-delivery step (via the work-item queue,
`DECISION_2026-08-21e`); the delivery-only listener (approved 08-25, own council
round).

**Not started / owner-deferred:** Stripe keys + webhook edge exception (gates taking
money); the terms page (the copy already points at "the full terms"); the second
Nominet TAG (domain programme, not this lane's critical path).

## Where we are now — paused, with two big OPEN reconsiderations

The domain stays parked. The owner's copy and design revision brief is coming shortly
(that work goes through the framework, and rebuilding `index` will remove the
hand-placed label — runbook gate 2 covers re-placing it). Beyond copy, the owner put
two product questions on the table on 08-25, **neither decided**:

**1. First-time quality may not be presentable.** The sites the framework produces may
need owner edits before a customer sees them. The owner is working on site quality in
a separate thread (not named here; not ready). Note the tension this carries: the
attested position is one-shot with no approval stage, and 08-21's ruling already added
an owner *review* (look, don't touch) via the work-item queue — this would extend
review into *editing*, which changes the build pipeline's shape and possibly the
"three or four days" promise (the D-B question again, but larger).

**2. Customer self-editing during the ~month view, possibly by voice.** When the
customer gets their site's month-long hosted view, let them edit it a bit themselves —
maybe speaking the changes. A customer editor already exists in the joint architecture
as **Phases 5–6** of `site_delivery_and_editor/PLAN_2026-08-17_delivery_architecture_decisions.md`
(post-handover, editor login home as the account hub, framework builds all content and
edits flow through the editor). The new parts are: editing moved EARLIER, into the
preview window; voice as an input; and the tension with the attested "no changes are
included" terms — if customers can edit, the terms, the register facts and the copy
all have to move together (the claims gate and writer_block enforce today's position).

**If both ideas are dropped, nothing above them changes:** this snapshot plus
`RUNBOOK_go_live_webdesign_uk.md` is the resume point, unchanged.

## Where we're going

Immediate: the owner's copy/design brief, executed through the framework and verified
at the served pages. Then the two reconsiderations get decided — if either is adopted
it gets its own plan (the editor one is a shared-seam product change touching the
delivery architecture, the terms, and the register; the owner-edit one touches the
build pipeline and the duration promise), and the launch order is re-cut around it.
If both are dropped: run the go-live runbook's gates and unpark. In parallel and
unblocked either way: the second-click confirmation page (still the one owed code
task), the delivery-only listener (approved), and — when the owner chooses — Stripe.
