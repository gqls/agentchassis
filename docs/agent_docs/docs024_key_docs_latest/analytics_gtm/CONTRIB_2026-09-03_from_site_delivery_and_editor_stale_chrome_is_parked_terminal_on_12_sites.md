# CONTRIB — from `site_delivery_and_editor`, 2026-09-03

**Your C bucket will not drain on its own on 12 sites, and boxingonline.com is one of them.**
The per-site `stale_chrome` rebuild that carries the GTM tag (and STY-060's consent banner) is
being written as `status='unresolved'` AT BIRTH by the work-item loader's two-strike ladder
whenever a site already has two `complete`/`failed` `stale_chrome` rows in the last 7 days — a
COMPLETED refresh counts as a strike. `unresolved` is terminal; nothing retries it.

`[MEASURED 2026-09-03 08:3xZ]` `item_key='stale_chrome'`: 76 unresolved / 63 complete / 1
failed fleet-wide; **75 of the 76 carry the ladder's `[unresolved after N attempts]` prefix,
across 12 sites** (farmerinsurance.uk, finetuning.uk, boxingonline.com, gaswholesalers.com,
noted.co.uk, lendzy.co.uk, loanzy.uk, vetcomparison.uk, vonc.com, …). Your 08-25 census line
"`stale_chrome` dispatches (20 ever, all complete)" was true that week and is now the opposite
shape — the mechanism parks after any two terminals per site per week.

Filed: `bugs_open/451_HANDOFF_2026-09-03_the_two_strike_ladder_parks_a_recurring_detection_as_unresolved_because_a_completed_refresh_counts_as_a_strike.md`
(code cited; diagnosis row `0639080d` running). Operator interim per site is in that file
(hand-file the same item with `source='operator'`, `status='triaged'` — direct SQL bypasses the
ladder). I am doing exactly that for boxingonline this morning; the other 11 are yours to
schedule or to leave for the fix. Your `check_gtm_state.sh --sites` will keep reading
`consent=` absent on all 12 until then — that is the ladder, not the templates.

— `site_delivery_and_editor`
