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

## 1. What to do next

1. **Check the cache age before anything else**:
   `curl -sI https://preview.webdesign.uk/assets/js/snippets.js -H "Host: preview.webdesign.uk" | grep -i "cf-cache\|age:"`
   If `age` ≥ 14400 or `cf-cache-status` is no longer `HIT`, the cache has
   cleared — go straight to step 2. As of this handoff (check made
   2026-08-10, late session) `age: 5621`, so roughly **146 minutes**
   remaining from that check, less whatever's elapsed since.
2. **Confirm the chat box actually works for a real visitor** — not a
   cache-busted curl, an actual browser load of
   `https://preview.webdesign.uk/contact.html` with no cache-busting query
   param, submit a real message, confirm a reply streams in. This is the
   check that was skipped before telling the owner it worked — don't skip
   it again.
3. **If the owner wants it fixed sooner than the wait**: the only lever
   this session found is a manual Cloudflare dashboard cache purge for
   `webdesign.uk` — the API token available here (`~/.config/cloudflare/token`)
   does not carry Cache Purge permission (confirmed twice now, don't
   re-check it, just tell the owner). This was surfaced to the owner in
   `README_where_we_are.md`'s latest entry — read that before saying
   anything that contradicts it.
4. Once 2 passes clean, resume `b` §2's original next steps: owner review
   of the whole thing, then Phase 6 cutover.

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
