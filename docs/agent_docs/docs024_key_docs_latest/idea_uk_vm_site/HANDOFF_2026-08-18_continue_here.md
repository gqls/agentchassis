# HANDOFF 2026-08-18 — honesty arc CLOSED, brand-head assets LIVE, drain rate answered (it was never a drain problem). Read this, then the open list in §4.

**Supersedes `HANDOFF_2026-08-16_continue_here.md` as the cold-start file.** That file
remains the reference for the honesty arc's full history, §4 (the gate cannot see titles
or meta descriptions — **still true, still unfixed**) and its §7 traps. Its §5 work list
is done or ruled dead; do not work from it.

Cold-start order: **this file → `HANDOFF_2026-08-16` §4 + §7 → `HANDOFF_2026-08-11` §3
(RFC_015) → `README_where_we_are.md`**.

> ## ⚠ 2026-08-19 — LIVE INCIDENT FOUND AND RESTORE IN FLIGHT: most text on idea.uk went invisible on 2026-08-17 21:41Z
> The 08-17 css-patch-agent wave (the four `contrast_failure` completions this file's §1
> counted as ordinary drain) **deployed the near-empty `css_themes` row over the site's
> real 23,650-byte stylesheet** — every `:root` colour variable vanished; dark-fallback
> sections render black-on-black, heroes white-on-white. Fleet-wide defect (6 sites),
> filed as **`bugs_open/198` round 2** (owned lane notified there; full mechanism, fleet
> table, per-site recipe). LANDMINE added (`css_themes` footprint).
> **idea.uk is RESTORED at the DB** (`css_themes` v6, md5 `4841523e…`, base = vm-sites
> `8c407a18f` + the four legitimate patch rules) and the deploy is riding a single
> unparked canary item `01a4dbca` (promoter → dispatch → css-patch run ships the full
> row → live host pulls ~1.5h). **Verify at the artefact before doing anything else
> here:** `curl -s https://idea.uk/assets/css/styles.css | grep -c ':root'` → expect
> **3** (currently 0). If the canary sits `detected`/`triaged` for a day, check for a
> stuck `claimed` row on the site, then the promoter (`detected-item-promoter`).
> Until it lands, §4's ordering is suspended — this outranks everything.


---

## 1. What closed since 08-16 — all three verified at the artefact

- **The honesty arc is CLOSED — owner ruling 2026-08-17.** No more sweeps, no `090` on
  the regrowth, the 30 pages carrying the word STAY, and the fix is migration `454`
  (rule 19 in `page-content-writer`'s prompt; `[VERIFIED 2026-08-17]` 3/3 writer calls
  carrying it in `llm_call_log.prompt_rendered`). ⚠ The four uses of the word inside the
  writer prompt are ANTI-FABRICATION rules — do not "clean" them (08-16 handoff, top
  banner). **Class C is retired by this ruling** — `funding-fit`'s "1. Where is the
  idea, honestly?" stays; do not re-fix it.
- **favicon + og-card are LIVE.** The `needs_brand_head_assets` item filed 08-17
  (`created_by='claude-ideauk-brandhead-20260817'`) completed 08-17 19:40, first
  attempt. `[VERIFIED 2026-08-18]` at the served page: `/assets/images/favicon.png`
  **200**, `/assets/images/og-card.png` **200**. The mode-less-item defect behind it is
  contributed into `bugs_open/131` (og-card slug); webdesign.co.uk / webdesign.uk /
  cookly.uk still carry it — their lanes' call, not ours.
- **The "~1 item per 3 hours" drain question is answered, and the answer is
  fleet-level, not idea.uk-level.** idea.uk completed **42 items on 08-17**; the 3-hour
  figure was a wait for a TURN, not a rate (median handler runtime fleet-wide: 36 s).
  The fleet dispatches ONE site at a time (`max_concurrent: 8` is dead config) in
  strict fleet-wide FIFO by item age, ceiling ≈ 83 items/hour for the whole fleet.
  Full evidence + plan for fixing it: **`dispatch_throughput/`**
  (`STARTER_2026-08-18…` = the measurements, `PLAN_2026-08-18…` = the phased fix).
  That is now its own workstream — **this lane should not touch dispatch machinery**,
  only read the two files when queue timing looks confusing again.

## 2. Queue state `[MEASURED 2026-08-18 ~16:00 UTC]` — nothing is dispatchable, and that is not a fault

| status | count | what they are |
|---|---|---|
| triaged / approved / claimed | **0** | nothing waiting on a turn |
| detected | 31 | ALL `handler_agent=''` → unroutable; the promoter's guard refuses them CORRECTLY. This is `bugs_open/083` (606 `head_essentials_missing` fleet-wide), **actively owned by the bugfix_277 lane** — contribute there, never compete |
| deferred | 40 | mostly `contrast_failure` (23) — parked TRUE findings awaiting a fixer (`bugs_open/296` §8); `undeployed_asset` (12) |
| needs_human_review | 37 | owner's queue, not ours |
| failed | 14 | 6 `empty_section`, 4 `page_rerender`, others — worth one triage pass (§4) |

So: filing work at idea.uk today reaches the queue immediately (it is empty). The
detected rows will start moving when the 083 lane's promoter work lands — watch, don't
push.

## 3. Standing rulings that bind this lane (unchanged, easy to lose)

- **D-005** protects the report hero's blessed clause; its guard files
  `decision_regression` if the served page loses `honest assessment`. `[VERIFIED
  2026-08-18]` no regression filed since it went live (the one row is 08-09,
  `cancelled` — the pre-live misfire). Re-check the hero ONLY if one files.
- **`whether you're` stays** (owner ruling 08-16 #2) — it is a `ParseVoiceGate`
  built-in tell, not an owner rule. Close any `voice_tells` item raised against it.
- **The voice gate cannot see `pages.title` / `pages.meta_description`** (08-16 §4).
  Still true. Head surfaces were 0-for-686 on 08-15; any head regression is invisible
  to the gate AND to `banned_claims` — re-run the head census (08-16 §8) if titles are
  ever touched.

## 4. Open list, re-verified 2026-08-18 — in rough value order

1. **news at `/data/latest-news.json` — still 404; `content_sources` for idea.uk still
   0** (fleet 49). Untouched since 08-04/05; the dispatch mystery is §X.53 in
   RUNNING_NOTES. This is now the oldest live gap on the site.
2. **Class B** — 8 components, 3 sites, `content_data` NULL; real visible copy incl.
   `finetuning.uk`'s `<h2>`. The 08-14 handoff called it "filed"; **no `bugs_open/`
   case exists** — either file one (it is cross-site, so 090 first per the 07-31
   ruling) or stop describing it as filed. Titles were fixed; re-check the other seven
   for the same body-vs-title over-reach.
3. **The 14 `failed` items** (§2) — nobody has read their `error`s in this lane's
   records. One pass: `SELECT item_type, error, attempt_count FROM site_work_items wi
   JOIN sites s ON s.id=wi.site_id WHERE s.domain='idea.uk' AND wi.status='failed'`.
   Route findings, don't fix blind.
4. **Arming the voice gate on the remaining sites** — remember it will not protect the
   head (§3), and post-454 the body pressure should drop anyway; measure regrowth
   before spending here (the vintage-split query, 08-16 §8).
5. **Older residuals:** first organic signed Stripe webhook; tools-page card images and
   tool-page heroes; the empty-kind → SDXL image-routing hole; ingress landmines
   (`ufw allow 80,443` FIRST, grey second).

## 5. Traps carried forward (the 08-16 §7 set still stands; two additions)

- **A 5-minute bucket manufactures concurrency.** Measuring "how many sites get served
  at once" only discriminates at 1-minute granularity (dispatch ticks are ~3–4 min
  apart). Cost this lane a near-false finding on 08-18; full write-up in
  `dispatch_throughput/STARTER…` §2.
- **`attempt_count = 0` is the queue-position discriminator.** A row that has never
  been tried is waiting, not failing — the 08-17 session's correction (RUNNING_NOTES
  tail). Check it before diagnosing any "stuck" item.
