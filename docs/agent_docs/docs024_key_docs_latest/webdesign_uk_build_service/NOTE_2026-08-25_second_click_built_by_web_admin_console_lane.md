# NOTE 2026-08-25 — the second-click page is BUILT (not yet live), from the web_admin_console lane

**To the thread working webdesign.uk.** This is a courtesy note, not a request. I have built
the change your `DECISION_2026-08-24_confirmation_needs_a_second_click.md` specifies, because
it sat as item 2 of my own lane's handoff. **I have not touched any file in this directory
except by adding this one** — your NOTES, README and handoffs are untouched, deliberately, so
nothing of yours rides in my commit.

## What changed, in one paragraph

`/c/<token>` now splits by METHOD in core-manager. `GET` renders a page with one button and
**performs no database access on any arm**; `POST`, on the **same path**, is the confirmation
and keeps the existing behaviour byte for byte (the guards, the length check,
`delivery.ConfirmTransfer`, the same three outcome pages, the same headers). Files:
`internal/core-manager/handlers/delivery.go`, `delivery_test.go`,
`internal/core-manager/api/server.go`. **`platform/delivery` is untouched** — token semantics,
expiry, single-use and purpose-scoping are exactly as you left them.

## The three things you may need to act on

1. **It is COMMITTED, not LIVE.** It rides the next core-manager roll. Your own decision doc
   says no delivery email goes out before this is live — that is still true today. Prove it at
   the pod, not the tag: `kubectl -n ai-persona-system logs -l app=core-manager --tail=2000 |
   grep -m1 'build provenance'`, then `git merge-base --is-ancestor <the commit below> <the
   stamp>`, with the reversed test as a control.
2. **No box change ships with this, and that is measured rather than assumed.**
   `box/links.webdesign.uk.nginx` already carries `location ~ "^/c/[A-Za-z0-9_-]{20,128}$"`
   with `limit_except GET POST { deny all; }`, and its header comment already records the
   ruling. The anchored regex is exactly why the POST is same-path: a suffix route would have
   404'd at the box. **If you ever narrow that location, narrow it knowing POST must survive.**
3. **Two forks the owner settled in-session on 2026-08-25**, both differing from a plain
   reading of your decision doc, so they are worth knowing before you write the email copy:
   - **The GET does not look the token up at all** (your §"The build" sketches a used/expired
     token getting the "no longer active" page on GET). It does not: the page is identical for
     every token, so opening a link cannot reveal whether it is real. A customer with a spent
     link now learns that **on pressing**, not before. That cost was put to the owner and taken.
   - **The page names nothing** — no site, no domain, no customer — keeping the rule already
     written at `delivery_deps.go:36-38`. Your doc's "who/what is being confirmed" is served by
     naming the ACTION, not the site.

## Where the detail lives

Council: `SUBMISSION_CORR=ea99befa-ec62-4f61-b052-c3af3d003d55`, submission JSON at
`../web_admin_console/COUNCIL_SUBMISSION_2026-08-25_second_click_confirmation.json`. The
technical log, including a misstep worth reading (my first test for "the GET mutates nothing"
was vacuous, and the mutation that should have caught it PASSED), is in
`../web_admin_console/NOTES_web_admin_console.md` under 2026-08-25.

`customer_access_tokens` = 0, `sites` handed_over 0 / transfer_confirmed 0
**[MEASURED 2026-08-25]** — so nothing is at risk while this waits for a roll.
