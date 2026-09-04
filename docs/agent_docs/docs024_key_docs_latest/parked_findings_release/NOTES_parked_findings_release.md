# NOTES — parked findings release

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-09-04 — lane opened; the first day's work is one finding

**Origin.** Owner ruling relayed by the `site_delivery_and_editor` lane: default the HITL loop to
approving, work through all the parked findings. Session renamed `parked findings` and told to own it.

**What I checked, in order, and what each returned.**

1. `status` census: `deferred` = **3,903** rows fleet-wide, `needs_human_review` = 1,424.
2. Split on `handler_agent`, which is the split that matters: **3,646 shelf** (empty handler) vs
   **256 named** (`bugs_open/396`'s different population). Of the shelf, **3,184** are
   `filing_mode='record'` with a `release_recipe` — this lane's population. Matches the handover's
   3,184 exactly.
3. All six routed handlers are live, active, non-snapshot, not deleted: `content-gap-planner`,
   `page-build-handler`, `webdesign-agent`, `component-template-fixer`, `copy-editor`,
   `css-patch-agent`.
4. `governor_admits()` returns true for all eight parked item types. That door is clear.
5. `spec.routed_status` is **`detected`** on every row — *not* `triaged`. That is what sent me to
   the promoter.

**The finding.** `detected-item-promoter`'s **door 5** (migration `629`, `bugs_open/405`) holds
every row whose `spec.origin = 'model_opinion'`. **2,824 of the 3,184 parked rows carry that stamp**
— written by `write_audit_findings`, the *same producer* that parks them. The live `pre_query`
confirms it verbatim: `(COALESCE(wi.spec->>'origin','') <> 'model_opinion') AS origin_ok` and
`WHERE pipe_ok AND handler_ok AND known_good AND floor_ok AND origin_ok`.

So the per-row release recipe **the rows carry themselves** is inert for 89% of them. Simulating all
five doors: **352 flow, 2,832 stick.** The failure is silent — status changes, handler is named,
nothing errors.

`629`'s header states it does not affect record mode because *"deferred rows were never the
promoter's candidates"*. That is true while parked and false the instant the recipe runs. Both
statements were written by people who were right about their own half; nothing joined them up.

**A second, smaller one from the same simulation:** `dark_section_audit` → `css-patch-agent` has
**0** completions and **1** failure in live+archive, so door 3 (`known_good`) holds those 32 rows
regardless of door 5 — **correctly**. Any release design that bypasses the promoter would dispatch
them to a pair that has never once succeeded.

**Misstep — I nearly reported the door as observed behaviour.** It is not. I looked for
`model_opinion` rows sitting at `detected` as direct evidence and found **zero** (against 1,010
unstamped ones) — because record mode parks them at `deferred` before they can ever reach the
promoter. The door has never been seen to fire. The claim rests on reading live config, and it is
marked `[INFERRED]` in the PLAN because of that. The cheap confirmation is `629`'s own direction-1
recipe (one synthetic held row, two ticks, cancel) and it has **not been run yet**.

**Correcting the handover on one point.** It said the chassis roll (`v1.0.1361`, `06c0b18f2`) had
not landed and pods still stamped `239ab3626`. As of 16:2xZ both live `agent-chassis` pods stamp
`06c0b18f233bc600918ef481d32b40f29535f78f`, started 16:01Z. It rolled in between. The stamp table
still lists the old pods because it is a two-hour survivor window, not a history —
`kubectl get pods` is what settles it, and I only checked because the table showed both.

**Pacing, which turned out to need no invention.** The promoter is `interval_seconds=900` with
`candidates … LIMIT 20` ⇒ **80 rows/hour**. It is already the throttle. The thing to avoid is
*defeating* it by releasing straight to `triaged`/`approved`. Separately: the recipe leaves
`created_at` alone, and `find_dispatchable_site` is `ORDER BY MIN(created_at) ASC LIMIT 1` — so
released August rows sort to the **front** of the fleet FIFO, not the back. The handover's instinct
to cost the sequencing was right; the direction is the opposite of "3,184 positions at the back".

**A tooling misstep worth one line, because it cost a file.** I wrote the standing docs with
`cd <lane dir> && cat > FILE <<EOF` in successive Bash calls. The working directory persists between
calls, so the second `cd` was relative to the first one's result, failed, and `&&` silently skipped
the heredoc — NOTES was missing and the `ls` in the same call is what caught it. **Use absolute
paths in heredocs; a failed `cd` before `&&` loses the write with no error of its own.**

**Not done, stated so silence is not read as completion:** nothing released, no synthetic control
run, no migration written, `457` not yet contacted about boxingonline, the `boxingonline.com` lane
not yet contacted, and no owner decision on mechanism.
