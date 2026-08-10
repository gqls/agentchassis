# SUMMARY 2026-08-10 — Phase 5 shipped, deposit pricing live, all nine parked CTAs resolved

## What we're trying to do

Build webdesign.uk: a five-page brochure site, framework-built, selling a
fixed-price (£1,200) done-for-you website build for UK small businesses.
The VM-hosting plan (`PLAN_2026-08-04_webdesign_uk_vm_hosting.md`) adds one
hand-written piece on top of the framework-built pages — a small chat
service that lets a visitor start a real conversation about their business,
because that conversation is the actual demand signal this phase of the
project exists to collect.

## Where we've come from

The five pages went through two build rounds and two rejected-then-fixed
review passes before the owner parked them "ok for now" on 2026-08-09. The
chat service (Phase 4) was built, deployed, and proven live the same day the
owner's Anthropic key landed — stdlib Go, all five safety controls (per-IP
limit, turn cap, spend ceiling, request log, transcripts) mutation-proven
before shipping. What remained was Phase 5: getting an input box onto the
page itself, which is what nine parked "this button goes nowhere yet"
items were waiting on.

## What we've done

**Phase 5 shipped.** A hand-built `chat-input-box` component (CTS-044
compliant: external loader, no inline script, static copy pre-checked
against the live ban list) sits on the contact page between the hero and
the plain email/phone fallback. Along the way:

- **Found and fixed a real bug**: the chat service had no memory across
  turns — every reply was generated as if the conversation had just
  started. Fixed, mutation-tested, proven live with a real two-turn
  exchange.
- **Found, then corrected, a serious platform bug** (`bugs_open/239`):
  the fleet's documented manual-dispatch workaround for the starved build
  queue can silently no-op while reporting success — and, on closer
  bisection, the trigger isn't even a clean function of the message shape;
  the same payload succeeded once and failed later. The correction matters
  as much as the finding: the first write-up overstated what had been
  learned, and that got caught and fixed in the same file.
- **Caused, then recovered from, a real production incident.** Two
  bisection dispatches believed harmless actually ran for real against the
  live contact page and reset its hero image to a broken placeholder.
  Caught within minutes by checking the actual stored content rather than
  trusting a status field, restored byte-for-byte against the last known-
  good git commit, verified clean.
- **Switched to a fully deterministic delivery method** for the rest of the
  session: hand-render static content against its own template, verify
  byte-for-byte, write directly to the git-backed static site — never
  through the dispatch mechanism that had just proven unreliable. Every
  edit this session, on both the chat box and the pricing change below,
  used this method.
- **Priced a £75 non-refundable deposit** on the owner's instruction — not
  guessed: checked against comparable AI website-builder pricing (Lovable,
  Durable) and against this site's own actual build cost (measured from
  real token usage in `llm_call_log` against current Anthropic rates, ~$1.50
  in text generation). Both anchors pointed below the owner's initial
  £80-150 hope; £75 was the number actually recommended, then confirmed.
- **Propagated the deposit change everywhere it needed to go, in lockstep**:
  `evidence_base.facts[]` and `writer_block` (the site's own source of
  truth), the three live pages that stated "full refund", and the chat
  bot's hand-copied system prompt — the exact coupling risk flagged when
  Phase 4 shipped, now proven to actually bite, and fixed the same way it
  was predicted it would need to be.
- **Resolved all nine parked `unresolved_cta` items.** Each one's own copy
  already named its intended destination ("Read the FAQ ...(/faq.html)");
  the automated resolver just couldn't connect label text to a page slug.
  Primary "get in touch" CTAs now point at the contact page (not a bare
  `mailto:`), so that traffic actually reaches the new chat box.

## Where we are now

Phase 5 is done. The chat box is live, backed by a chat service with
verified memory across turns and all five safety controls intact. The
£1,200 / £75-deposit / 14-day / 2-round terms are consistent everywhere a
visitor or the chat bot could state them, machine-checked with the live
ban list after every change. All nine CTA buttons across the site now lead
somewhere real. One small, self-resolving gap remains: Cloudflare's edge
cache holds the chat box's loader script for up to four hours after a
change, so a real visitor may see the box before it's clickable for a
window of up to a few hours — the origin is already correct, this is
purely a cache TTL, not a defect, and it needs no action, only checking
before telling anyone it works end to end.

## Where we're going

Owner review of the whole thing, then Phase 6 cutover. Separately: the
platform-level dispatch bug (`bugs_open/239`) is now well-characterised but
not fixed — it isn't this lane's to fix, but it matters to every other
workstream relying on the same documented manual-dispatch pattern while the
build queue stays starved. And, recorded as a future direction rather than
current scope: the owner raised a possible £19 all-in tier (site plus
static hosting) once the £1,200 tier has real customer-handling data behind
it.
