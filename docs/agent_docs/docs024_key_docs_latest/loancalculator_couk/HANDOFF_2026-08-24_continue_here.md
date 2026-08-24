# HANDOFF — loancalculator.co.uk · the harness is FIXED, and the first thing it found is `bugs_open/385` (2026-08-24)

> Supersedes `HANDOFF_2026-08-23_continue_here.md`. That file's "TWO REMAINING ITEMS" are
> both **done** (its own evening block records them), and its two ⚠ cautions are both
> **resolved**: the chrome drain has finished, and the acceptance harness was never
> actually down. Read 08-23 for the mechanism of the four owner instructions; read this for
> state. Evidence: NOTES `## 2026-08-24`. Owner prose: README_where_we_are.

```
site      loancalculator.co.uk   0162cde4-633e-45e9-8ca6-87a6b2fe1d26
pages     28 active · 28 serving 200 · ZERO 404s   [MEASURED 2026-08-24, all 28 curled]
chrome    DRAINED — 28/28 carry the /guides/index.html link, 0/28 still reference
          credit-roadmap. The 08-23 "sampling one page misreports it" caution is OVER.
locks     11 locked tool sections, all positionally named (tool-1…tool-4), none matching
          its component function — which is normal here, not a defect (see 385 §5a)
plan      9463e31d-ee50-482e-94a9-7e186ef25543 is_current (created 08-17; no replan since)
flags     honour_realised_identity + twin_identity_snap + stem_twin_snap ALL TRUE
harness   ✅ WORKING — fixed at the shared seam, commit 0aafce405
golden    ✅ acceptance/GOLDEN_2026-08-24_post_385_repair_tool_values.json — all 11,
          and PROVEN to reproduce (a fresh --compare against it returns 11/11, exit 0).
          Keep GOLDEN_2026-08-17 too: it is the only pre-rebuild record, and it is what
          proved 385 was a REGRESSION (react=5 there vs react=0 live).
```

## THE HARNESS — fixed, and it has a self-test now

`toolgolden.py --compare` was recorded on 08-23 as *"THE ACCEPTANCE HARNESS IS DOWN"*.
It was an environment fault, not a harness or a site fault:

> `toolprobe.start_chrome` launches chromium with
> `--user-data-dir=tempfile.mkdtemp()`, and `mkdtemp` honours `$TMPDIR`. A snap-confined
> chromium **cannot write under a hidden top-level `$HOME` directory**, so it aborts at
> ProcessSingleton (rc 21) before the DevTools port opens. It is a **shared** helper —
> 6 scripts, 4 lanes — so one setting takes them all down, each reporting it against
> whatever URL it was holding.

Fixed in `start_chrome`: falls back to a permitted profile dir (saying which and why),
fails fast on `poll()` carrying chromium's own stderr, and refuses a port already serving
DevTools. Full trap: `LANDMINES.md`, *"a browser harness that 'is down' is probably reading
`$TMPDIR`"*.

**RUN THIS FIRST, ALWAYS — it is one command and it is the difference between a finding and
a false conviction:**

```bash
cd docs/agent_docs/docs024_key_docs_latest/loancalculator_couk
python3 toolgolden.py --selftest      # green => a divergence is about the PAGE
```

It drives a fixture whose answer is computed by hand from the driver's own rules
(`1050.00 · 2200.00 · 512.50 · 1751.00`) and is proven disconfirmable by mutation. Note
what the mutation showed: **gates A and B stay GREEN on a wrong-but-responsive calculator**
— only the value assertion discriminates, which is the whole reason this harness exists.

## WHAT THE WORKING HARNESS FOUND

**Ten of eleven calculators reproduce their golden values exactly.** `[MEASURED 2026-08-24]`
All 72 divergences across the nine "DIVERGED" pages are ONE cosmetic change — the FAQ
container id went from a raw UUID (`d91e7be1…`) to `c-faq` — and **zero `controls`
divergences, zero numeric movement**. `index.html` MATCHES outright. So the 08-23 release
wave did not disturb any arithmetic.

**The eleventh is `bugs_open/385` — now REPAIRED (see below), but read the bug file: the
CAUSE is open.** `/tools/loan-vs-savings.html` rendered its calculator
**twice**, lower copy dead, because the rebuild appended an unlinked byte-identical copy
(`component_id` NULL) of the locked section it had just repositioned. Read the bug file —
root cause is OPEN, `bugs_closed/189` is REFUTED as the explanation, and the `090` route is
closed for these files (`UNVERIFIABLE — iteration-cap`).

⚠ **The orphan also made the page UNRERENDERABLE** — `rerender_sections` failed 3/3 on
`unresolved component [tool-2 (pos 6)]`. That is the transferable part, now in 016b §9:
**a defect that makes a row unresolvable disables the very mechanism that would repair it**,
and the failure wears the vocabulary of a content problem.

## THE REPAIR — DONE and verified at the artefact

Owner-approved and complete. The orphan row was deleted inside a guarded transaction (a
`DO`/`RAISE` block, **not** a block of `SELECT`s — those cannot stop a `COMMIT`), its
recoverability confirmed in `page_component_history` first, and an **assemble-only**
redeploy shipped it (`98529d02…`, no `spec.reason`).

```
[MEASURED 2026-08-24]  served sha256 e3d2da2b… == the committed file, exactly
                       duplicate ids  0  (was 11)
                       harness        react=5  vary=5  12 fields — identical to the
                                      08-17 golden's own record for this page
                       divergences    8, ALL the cosmetic c-faq rename; zero controls,
                                      zero numeric movement
```

⚠ **If you re-verify this and see the OLD page, re-sample before concluding anything.** I
told the owner the publish had failed on the strength of one `curl` taken ~90 seconds after
the B2 sync logged `upload tools/loan-vs-savings.html`. It had not failed. What made the
false reading persuasive: every other page showed a fresh `last-modified` and this one did
not, which reads as *"the sync skipped mine"* and actually means **position in the sweep**.
A one-shot comparison against peers who share the confound is not a control. Full write-up:
`WRONG_CALLS.md` `## 2026-08-24`.

⚠ **A check worth copying before any repair of this shape:** `pages.sections` is a
materialised cache that LOCK-008 merges locked rows into, so a stale entry there would let
an assemble re-materialise the duplicate and make the repair look done. It held 5. Check it.

## NEXT ACTIONS, in order

1. **`bugs_open/385` root cause** is unowned, and it is the only live risk here. The next move is stated in its §5b: read
   LOCK-008's merge of locked rows into `pages.sections` and the compile step, because the
   plan demonstrably does NOT list the tool twice. Re-file `090` with **ONE** symbol if you
   want the loop.

## Standing cautions (carried; all still true)

- **Prove a deploy at the artefact.** Never grep the binary for a fix's own sha.
- Verify tool placement at `site_plan_sections`, never `pages.sections` (LOCK-008 merges).
- ⚠ **`UPDATE page_components SET position` does NOT touch `updated_at`.** So "the locked
  row's `updated_at` never moved" proves the BYTES were not rewritten — it does **not**
  prove the row was not repositioned. The 08-23 canary evidence should be read with that
  limit (its md5 comparison is unaffected).
- The phase-2 script's judge query `component_name LIKE 'tool-%'` returns 26 either way —
  it matches `tool-cta`/`tool-list`. Use the locked-function join in 17b.
- A hand-filed or un-parked work item must be `triaged`; the dispatcher cannot see
  `detected` and fails silently.
- Query runs BY CORRELATION, never `now()`-interval; a planner run's `collected_data` can
  purge within ~2 hours.
- `retract_page_deployment` REFUSES an active page, so archive first, and its DEFAULT
  selection is every non-active page with a deploy stamp — use explicit `page_ids`.
- **A single sample during a deploy or rerender wave proves nothing — of a 404, and equally
  of BYTES.** Re-sample before you conclude, and note the trap in the harder direction: a
  page differing from its peers mid-sweep looks like it was SKIPPED and usually means it has
  not come up yet. Worked case in `WRONG_CALLS.md` `## 2026-08-24`.
- **Before any planner run**, the four cautions in `HANDOFF_2026-08-23_continue_here.md`
  still apply verbatim (`checkpoint_postplan.sh` immediately; check item KEYS against the
  plan the run just wrote; re-verify identity flags; Pass C2 will not save you).
