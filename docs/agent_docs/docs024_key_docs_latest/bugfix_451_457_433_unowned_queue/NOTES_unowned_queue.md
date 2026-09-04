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

## 2026-09-04 — lane resumed; 457's fix was half a fix

Resumed after the lane went quiet 2026-09-03 20:21. Ownership re-checked properly this time:
`who-owns.py` names `site_delivery_and_editor` and `components_lane_425`, both ACTIVE — but
`ListAgents` shows **no live session for this lane**, and the check that settles it is the one
MISSTEP 1 named: grep what the named lanes DECLINED. Both declined the fix. Resumed here.

**State on arrival, all three axes re-measured rather than inherited:**

- `f895616d7` is LIVE — chassis `239ab3626` (pods up 09-03 22:07Z), `merge-base --is-ancestor` true.
- `[MEASURED 09-04 ~16:05Z]` it has **never been exercised on the page that reproduces the bug**.
  `orchestration_states` retains to 09-03 15:15Z, 52 `rebuild_blog_listing` runs, all COMPLETED,
  **zero for boxingonline**. "The fix is live" and "the fix has run" are different facts and I nearly
  wrote the first as though it were the second.
- Served damage unchanged: 36 cards / 6 headings / 14 / 2, controls passing.

### The finding: the origin gate guarded one write verb

`decideBlogListingWrite` tested `Occupants == 1 → opUpdate` **above** `switch slot.Origin`, so the
authority gate the 09-03 fix exists to install reached `opInsert` and never reached `opUpdate`. A
guessed (2b) or defaulted (3) slot with one occupant was UPDATED — i.e. the page's own content
overwritten with the article listing.

The tell was in the fix's own test file, not in the code: a guessed origin with an EMPTY slot
refused, a guessed origin with an OCCUPIED slot wrote. **The safer case was the one being refused.**
When a decision table refuses the harmless case and permits the damaging one, the gate is in the
wrong place — that asymmetry is cheaper to spot than re-deriving the logic.

Armed by this bug's own remediation: boxingonline holds 7 rows in that slot so it refuses today, and
falls to 1 the moment `site_delivery_and_editor` deletes the six orphans. Told them; ordering is
**roll first, delete second**. Shipped `828b22c7c`, council `28bd3fd3`, mutation-proven.

### MISSTEP 6 — I proposed prior art that had already been refused, in writing, in the migration

Full entry in `WRONG_CALLS.md`. Short form: I measured that
`UNIQUE (page_id, slot_name, position) WHERE build_status <> 'removed'` holds across all 3,420 live
rows with one violating group (this bug's own), and wrote it into a planning brief as the
framework-wide door-closer. Migration **316** had tested that class against production on 08-05 and
refused it — a constraint stricter than the writers' guard turns a duplicate into a *silently
dropped section*, and `[MEASURED 09-04]` **2 of 7 writers swallow an INSERT failure**
(`save_page_sections` Warn+`continue`; `deploy_tool` ON CONFLICT DO NOTHING+Warn).

**The bug file cites 316 four times, by number and by effect. I had read every one of those citations
and none of them carries the reasoning — that lives only in the migration header.** The check:
`ls docs/agent_docs/sql_for_agents/ | grep <table>` and READ it before proposing an index.

### MISSTEP 7 — I asked a subagent to plan before I had done the prior-art read

The brief went out with the refused constraint in it as "E1". Had it not rate-limited, it would have
planned around a premise I disproved myself twenty minutes later. **The prior-art read belongs
before the delegation, not in parallel with it** — a subagent inherits your framing and has less
reason than you to doubt it.

### The growth misreading recurred, for the fourth time

The `boxingonline.com` lane reported this hour that the count grew 4×→6× and concluded "every run
adds one". It stopped 09-02 16:28:02. They compared a 09-01 document to a 09-04 measurement, so the
growth sits inside the interval and had already ended. §"TWO PRODUCERS WEAR ONE SYMPTOM" in the bug
file was written after the first three lanes did this, and did not prevent the fourth — **a section
warning about a misreading is only read by someone who already suspects.** The durable check is
`max(created_at)`, which answers "is it still growing" in one query and cannot be read backwards.
