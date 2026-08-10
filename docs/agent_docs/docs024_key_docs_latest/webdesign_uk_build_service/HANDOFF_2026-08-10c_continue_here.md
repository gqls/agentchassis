# HANDOFF 2026-08-10c — Phase 5 build is done; owner hit a real (cache) failure testing it live

**Start here cold. Supersedes `HANDOFF_2026-08-10b_continue_here.md`.** That
file's build/pricing/CTA claims are all still true — nothing below reverses
any of them. What's new: the owner actually tried the chat box live and it
did nothing, which upgrades the "cache TTL, no action needed" line in `b`
from theoretical to owner-observed. Read `b` for the full build history;
this file only carries forward what changed.

## 0. One paragraph

Phase 5 (chat box), the £75 deposit, and all 9 CTA fixes are still done,
still live, still correct — see `b` §1 for the full verified state table,
still accurate except for the row below. The owner opened the contact page
and submitted a chat message; nothing happened. Diagnosed on the spot: this
is Cloudflare still serving the **pre-loader** `snippets.js` to real
visitors, `cf-cache-status: HIT`, `age: 5621` of a 14400s (4h) max-age. A
cache-busted fetch of the same URL confirms the correct file — with the
loader — is sitting at origin, ready. This is not a new defect; it's the
same gap flagged in `b` §1's second row, except now confirmed to actually
break the visible experience rather than being a hedge in a state table.

> **RESOLVED + SUPERSEDED IN PART, same day — read §1a.** The cache issue
> below is CONFIRMED (with better evidence than was available when this was
> written) and has since CLEARED on its own. A second, unrelated blocker was
> found underneath it and is now the only thing standing in the way.

## 1a. CURRENT STATE (2026-08-10, late) — one blocker, and it is the owner's

**The cache half is done.** Proven, then self-resolved:
- Proof it was the cause: nginx on the box logged exactly **two** POSTs to
  `/api/chat` all day (12:54, 16:18 — both this session's tests). The
  owner's ~14:08 attempt is **absent entirely** — the submit never left the
  browser. (The app's own `requests.jsonl` could NOT have proven this; its
  silence is equally consistent with "never arrived" and "arrived, wasn't
  logged". Check the layer in front.)
- Proof it cleared: `etag` `"6a77af00-188c"` → `"6a79c880-226a"`,
  `cf-cache-status: EXPIRED`, and the loader now greps out of the
  **un-cache-busted** URL. All six selectors present in served
  `contact.html`; `<script src="/assets/js/snippets.js">` referenced.
  Client half fully proven — **do not re-litigate it.**

**The remaining blocker is an Anthropic account spend cap.** A real POST
returns 200 with the fail-closed `contactLine` body. Cause, verbatim from
`journalctl -u webdesign-chat`:

    anthropic 400: "You have reached your specified API usage limits.
    You will regain access on 2026-09-01 at 00:00 UTC."

Ruled out by measurement, so don't re-diagnose these: our own daily ceiling
(`$10.00` vs measured spend `$0.000922` today — 4 orders of magnitude
clear); the turn cap (failing call was turn 1 of a fresh conversation, cap
is 20); the service being down (`active`, `/health` ok, and an identical
call **succeeded at 12:54 the same day**). Total lifetime spend by this
service is **under one third of a US cent across 5 requests** — whatever
consumed the account allowance, it was not this. [UNMEASURED] what did —
no visibility into the Anthropic account from here.

**Owner action, and nothing else will do**: raise/remove the usage limit in
the Anthropic Console. It is an account setting, not a value on the box or
in any config file here — no redeploy, rebuild or restart is needed once
it's raised. Surfaced to the owner in `README_where_we_are.md`.

**Do not "fix" the fallback in code.** It firing is the Phase-4 safety
control working correctly, for the first time for a real upstream reason.
One open copy question flagged for the owner (not changed unilaterally):
the fallback opens "Thanks for your patience", which undersells a
potentially three-week outage.

## 1. What to do next — ORIGINAL text, kept for the record; steps 1–3 are DONE

1. ~~**Check the cache age before anything else**~~ — DONE, cleared, see §1a.
2. ~~**Confirm the chat box actually works for a real visitor**~~ — the
   client half is proven (§1a). The end-to-end reply still cannot be
   demonstrated until the Anthropic limit is lifted; **that** is now the
   thing to re-check after the owner acts, and it is a one-command check:
   `curl -s -X POST https://preview.webdesign.uk/api/chat -H 'Content-Type: application/json' -d '{"conversation_id":"","message":"how much does a website cost?"}'`
   — a real answer means done; the "Please reach us directly" line means the
   limit is still in force.
3. ~~**If the owner wants it fixed sooner**~~ — moot, the cache expired
   naturally. (The CF token still has no Cache Purge permission; that
   remains true and is worth knowing for the next `.js` change.)
4. Once the Anthropic limit is lifted and step 2 passes, resume `b` §2's
   original next steps: owner review of the whole thing, then Phase 6
   cutover.

## 2. Everything else — unchanged from `b`

- All 9 `unresolved_cta` rows: still `complete`, no reason to re-check.
- £75 deposit: still live everywhere (`evidence_base`, 3 pages, chat bot
  system prompt) — this diagnosis didn't touch any of that.
- Chat service memory-across-turns fix: still correct, this is a delivery
  (caching) issue, not a regression in the service itself — the service
  answered correctly every time it was actually reached this session.
- `bugs_open/239` (platform dispatch mechanism): still open, still not this
  lane's job, still worth reading before anyone drives work through the
  manual-dispatch pattern.
- Traps in `b` §3 all still apply unchanged, especially the
  `pages.sections` non-durability one and the content_data+rendered_html+git
  three-way-update one — this session made no further edits, so nothing
  there needs re-verifying, just re-reading before the next edit.
- Access map: unchanged, see `b` §5.

## 3. What's actually new here, in one line

**The cache-TTL gap between "I changed the origin file" and "a real visitor
gets it" is not a hypothetical edge case for this lane — it already caused
a live, owner-facing "I submitted a request and nothing happened" failure
once.** Worth remembering for any future same-day change to `snippets.js`:
tell the owner about the wait *before* they go test it, not after.
