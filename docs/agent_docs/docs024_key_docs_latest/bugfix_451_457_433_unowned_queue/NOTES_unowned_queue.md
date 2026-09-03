# NOTES — the unowned-bug queue (451, 457, 433)

Append-only, newest at the bottom. Missteps are the point of this file, not an appendix.

## 2026-09-03 — why this lane exists

The owner asked for three bugs to be looked at, with the instruction: if a thread is already
active on it, leave it; if not, resume it here. All three came back OWNED from
`scripts/who-owns.py`, and all three were **unowned for the fix**.

**MISSTEP 1 — I nearly stopped at `who-owns.py`.** It reads commits, so what it reports is *who
is citing this bug*, not *who is fixing it*. `bugs_open/457` says in terms *"Not fixed by me and
not touched … Flagged to the owning lane rather than actioned"*; the 424 lane that surfaced 433
closed the same day. **The check: grep the named lane's docs for what it DECLINED, not just for
the bug number.** An active lane contributing measurements to a bug is not an active lane fixing
it. → `WRONG_CALLS.md`.

## 2026-09-03 — the measurements that changed the picture

- **451** is **5,870** ladder-parked rows across **32** item types, not the 76 its file reports
  (that was one `item_key`). 3,782 had ≥1 `complete` sibling in the 7-day window before birth;
  2,391 had *both* strikes complete. Arrivals: 3,485 (w/c 08-24), 1,729 (w/c 08-31).
- **433** is 1,023 empty of 1,418 and growing (910/1,277 yesterday). Writer attribution settled by
  census: `StoreAssetAction` 957 + `recordDerivedAsset` 66 = all of it.
- **457** confirmed at 6 rows on one page, not growing — the action hard-fails now.
- **New, found while planning 457:** `function='content-listing'` returns **two** active rows, so
  one site's fork was every site's listing template. Byte-identical today (`1b957ae3…`), so latent.

## 2026-09-03 — MISSTEP 2, the expensive one: a gate measurement that could not disconfirm

For 451 I designed a check to test the claim *"a successful refresh provably clears the
detector's condition"*, **named the disconfirming result in advance** (re-files within the 3-hour
window after a completion), ran it, got **0 of 65**, and recorded "the gate PASSED".

Both halves were wrong.

1. **A `complete` does not imply a restamp.** `render_site_components_action.go` returns success
   without rewriting `render_inputs` on at least three paths — the *"degraded success"* return
   (whose own comment says the chrome is *STALE-BUT-SERVING*), `ResolveChromeComponent`'s error
   arm, and the "no row matched" swallow. Those completes leave the drift present, which is
   **persistence** — the population the brake exists for.
2. **The measurement was structurally blind to that.** The check runs daily, so a persistent
   unrestampable row re-files once per tick — *identical in the time domain to genuine daily
   drift*. I chose a time-gap discriminator for a question that is not about time.

**The general form, and it sharpens a rule this estate already has:** "could this number have come
out otherwise?" is **necessary and not sufficient**. It could have — a fleet with fast re-files
would have filled the bucket — but not *for the reason under test*. The honest discriminator is
non-temporal and was in the data all along: diff `spec->'drifted'` between consecutive rows —
**same keys twice is persistence, different keys is drift**.

Caught by an adversarial review pass over the finished plan, then confirmed first-hand.
→ `WRONG_CALLS.md`.

## 2026-09-03 — MISSTEP 3: three citations I relayed without opening

(a) *"451's candidate 1 is RFC_048's option A, refused by the owner"* — RFC_048's options are all
deferral/opt-in/census shapes; **none** is "count only `failed`". §6b's *reasoning* bears on it;
"refused exactly it" is an overclaim, **and I had drafted a fleet-wide WRONG_CALLS entry accusing
another lane on that basis.** (b) `bugs_open/352` is `contrast_findings_…`, not the
`undeployed_asset` population, and LANDMINES has no entry saying that remedy cannot reach its
trigger state. (c) I cited `improve_tool` (205 rows, zero completed strikes) as the control showing
the ladder behaving correctly — RFC_048:260-262 counts those exact rows among **431 action requests
that should never have been braked**. My control was the owner-documented counterexample.

**The check: a citation you did not open is a claim.** It applies with *more* force to evidence
arriving from a subagent's report, because it reads as already-checked.

## 2026-09-03 — MISSTEP 4: my own census cleared the one file it existed to catch

Censusing which `page_components` writers bind `component_id`, I grepped the INSERT and looked for
the column in the next eight lines. It counted a **comment** as a writer and **cleared
`rebuild_blog_listing_action.go` — the sole violator** — because a `zap.String("component_id", …)`
log field sits three lines below the statement. **Run the scan against the case you already know
the answer for.** Both ratchets shipped today read the *statement*, never a window, and carry a
fixture test; the 433 one is mutation-proven in both directions.

## 2026-09-03 — MISSTEP 5: the backtick trap, and a passenger I swept

- A `--` SQL comment I inserted into a Go raw string contained backticks around a column name.
  Raw strings end at the first backtick; the package stopped compiling. Caught immediately by
  `go build`, but it is the documented trap and I walked into it.
- **`e20662db9` swept a same-file passenger**: my pathspec named
  `docs026_concept_register/register/000_concept_index.md` for the IMG-079 row, and the file also
  carried the finetuning lane's uncommitted update of the **PUB-006** row (BUILT → LIVE). A
  pathspec commit takes the file from the working tree, so their edit rode along under my message.
  Nothing is lost and forward-only holds — recorded here because the practice requires saying so,
  and because it is the residue CLAUDE.md names: *no hook can prevent a same-file passenger.*

## 2026-09-03 — what shipped

- `f895616d7` — **457**. Occupancy on every exit path; the write decision a pure function of
  (origin, occupants); `component_id` bound; position from the plan; refusals log and return
  `rebuilt:false` rather than erroring (an error aborts `rerender-pages` before
  `create_rerender_items` — that IS the outage). Council `13273c8c`.
- `afcf3ebdb` — **433**. No-fallback byte sniff in `platform/storage`; the deployer's content type
  byte-derived; four writers record `mime_type` or explicit NULL; ban-shaped writer ratchet with an
  empty exemption map. Council `82989388`. Registered as **IMG-079**.
- **451 is NOT shipped.** Its converter-parity half stands; the exemption half needs a premise that
  survives misstep 2. The decision is with the owner.
