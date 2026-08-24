# HANDOFF — mortgagecalculator.co.uk — cold start, read this first (2026-08-21)

**Supersedes `HANDOFF_2026-08-18_continue_here.md`.** From that file only **§0 (the two dormant
mechanisms now running)** still needs reading — its §3.3 blocker analysis is superseded by §2 here,
and its §1 table is re-measured below.

**Nothing is blocked on this lane's own work, and nothing is half-finished.** The site's one
remaining defect is now owned by two other lanes, with the cases handed to them. Site unlocked, no
items armed, no uncommitted work.

---

## ⊕ UPDATE 2026-08-24 — two owner requests done since this file was written

**Everything below still stands** except where this section says otherwise. Read this first.

1. **The contact page was REWORDED (2026-08-21) and is live.** It had invited contact four times
   with no email, phone or form anywhere on the site. It now says plainly that there is no form or
   email address and routes readers to the FAQ, the guides and the assumptions page. Driven through
   the framework (a `content_rewrite` brief, `edit_live` verified applied — the words are the
   framework's, not a session's). **Item `07bc64cd` (wanting a business email) is deliberately still
   OPEN** — the owner asked for a reword, which is not the same as "never".
2. **The not-financial-advice disclaimer is now on EVERY page (2026-08-24)**, via
   `site_specs.site_config.chrome.compliance_lines` (**STY-051**) — one per-site setting, not 32
   page edits, so every future page inherits it too. Wording is the owner's choice (the site's own
   voice); an FCA/regulated line was **deliberately declined** as unverified. The chrome footer and
   all serving pages were re-rendered assemble-only (no LLM, no copy touched).

⚠ **Two things this update changes for whoever comes next:**

- **`build_status='deployed'` is NOT the set of pages that serve.** `contact-index` serves HTTP 200
  while flagged `needs_rebuild` with `deployed_at` NULL. A fan-out keyed on `deployed` misses it —
  it nearly cost this site the disclaimer on its own contact page. Use a probe, or
  `deployed ∪ needs_rebuild`. **30 pages serve; the table calls 29 deployed** (as of 2026-08-24).
- **Pages got ~2.4× heavier and it is not the disclaimer.** Re-rendering pulls in a 34.5 KB layout
  stylesheet the head chrome now inlines. Verified as fleet-wide catch-up rather than damage
  (every site rendered in the last 3 days carries 43–62 KB; the ~8.5 KB heads are exactly the sites
  last rendered 08-20, as this one was). Raised with the owner as a **fleet-level** question.

⚠ **Hazard, unowned — ONE page, not three (this line CORRECTED 2026-08-24, same day).**
`tool-simple`'s **`hero`** has NULL `content_data` with a 1,300-char `input_schema`, so re-rendering
that page **with** a `reason` escalates to the content writer and **regenerates its copy**;
assemble-only does not. It is the same component as open item `e781118c` (hero missing its
schema-required `headline`) — note the live page nevertheless shows a proper `<h1>`, so "the page
looks right" and "the stored record is incomplete" are both true.

> ~~"three deployed pages … `tool-bridging-compound`, `tool-rate-scenarios`, `tool-simple`"~~ —
> **wrong for the two tool pages.** `rerender_page_sections_action.go:~400` exempts a
> **self-contained tool section** explicitly (`isSelfContainedSection`, `:1361` = empty
> `input_schema` AND `component_level=='tool'`); both tool components qualify, and for them NULL
> `content_data` **is the correct shape**, not a defect. The error came from trusting
> `049b_deploy_single_page.sh`'s HEADER COMMENT, which states the general rule and omits the
> exemption, instead of opening the action. `WRONG_CALLS.md` 2026-08-24.

Separately and still true: `tool-bridging-compound` and `tool-rate-scenarios` are the pages the
RFC_034 lane's `instance_scope_conversion` items called *"SERVING BROKEN"* on 08-19 and then closed
having changed nothing — that remains unexplained and is not this lane's.

Full record: `NOTES_mortgagecalculator_couk.md` `## 2026-08-24` and `## 2026-08-21`;
`README_where_we_are.md` 2026-08-21 and 2026-08-24.

---

## 0. Read before you touch anything

- **`site-discovery-rotation-completeness` is ENABLED** and **the `stamp-duty` fence declares its
  13 SDLT facts** — both deliberate, both from 08-17. Full detail in the 08-18 handoff §0. The 13
  `fact_drift_review` items are **supposed** to be there; do not tidy them.
- **⚠ A work item on this site can say `complete` while nothing was built.** This is
  `bugs_open/348`, filed today and reproduced twice. **Read `page_components` and the served URL,
  never the item status.**

## 1. Live state — measured 2026-08-21 10:20–11:15Z

| artefact | state |
|---|---|
| pages | **32** — 29 deployed, 1 `needs_rebuild` (`contact-index`, still serving fine), 2 `planned` |
| internal links | **1,030 hrefs → 33 distinct targets → 32 return 200** |
| the only dead target | **`/scorecard-simulator.html`** — 404 on 3 cache-busted probes, anchored from **6** served pages |
| `agent-chassis` | **v1.0.1321**, both replicas; `bugs_closed/260`'s fix proven aboard at the binary with 4 controls |
| site lock | unlocked · items armed by this lane: **0** |
| `fact_drift_review` | 13, `needs_human_review`, untouched since 08-17 — **owner's call** |

Site id `62b5978e-4271-4589-8e00-4baebfc0447c`. Page id `0c252ee5-05f5-4b16-ad95-43795f4f198e`.
Item id `0c65f9fa-ddce-4e83-a6a8-4f252b3cf3cb` (parked at `needs_human_review`, `attempt_count` 2/3).

## 2. The one product defect, and why it is NOT yours to fix

`/scorecard-simulator.html` cannot build. **`bugs_open/260` closed on 08-20 and that is no longer
the reason.** Two runs today, on the fixed binary, both refused with:

```
component "mechanism-flow": steps[N].branches: declared array (items: object), got string;
refusing to render (bugs_open/260)
```

**Reliable, not stochastic** — attempt 1 hit `steps[2],[3]`, attempt 2 hit `steps[1],[2]`. Same
component, same field, same count of two; only the indices move. **A third retry has no reason to
succeed — do not fire one.** The component's `input_schema` is well-formed (checked, not assumed).

**This is 260's writer half, owned by `copy_quality_two_stage`** per the owner's 2026-08-12 split.
Handed over today with the verbatim refusal and the item to re-arm:
`copy_quality_two_stage/CONTRIB_2026-08-21_from_the_mcalc_lane_….md`. **Contribute, do not compete.**

## 3. What this arc filed, and what is owed back

| where | what | status |
|---|---|---|
| **`bugs_open/348`** (NEW, this lane) | a render refusal has no error route, so a build that composed nothing reports `complete` | **OPEN, UNOWNED** — fix candidates ranked, verification recipe included |
| `bugs_open/328` | contrib: the platform is **not** blind to dead links — it files one item per linking component and parks them unread | contributed; still OPEN, UNOWNED |
| `copy_quality_two_stage` | the reproducible writer mistype | delivered; **their reply is the thing to watch for** |
| `WRONG_CALLS.md` | my watcher matched retained error text and reported terminal 30 s early | logged |

**348 in one line:** neither `page-content-writer`'s `process_sections_loop` nor
`page-build-handler`'s `spawn_content_writer` declares an `error_step`, so the refusal reaches a
success-labelled complete path. `bugs_closed/028` closed this SHAPE in July for a different cause —
its fix is one of the two `mark_*` guards still in the live workflow. **Per-cause guards do not
survive a new cause**, and 260 added one.

⚠ **348's census is YOUNG, not low.** One occurrence fleet-wide, after ~20 h of exposure. **Re-run
it before anyone sizes the bug** — the query is in §6 of the file.

## 4. Decisions waiting on the owner — four, two of them now 5 days old

1. **The 13 `fact_drift_review` items** — close them or leave them. Closing is provably safe
   (`factDriftLastItemQuery` has no status filter); the reason this lane stopped is that the type is
   handler-less. *Waiting since 08-17.*
2. **The other lane's stamp-duty config question** at the top of the 08-18 handoff. *Since 08-16.*
3. **`contact-index` cannot finish rebuilding** — item `07bc64cd` wants a real business contact
   email to display. Only the owner can supply one.
4. **`tool-simple`'s hero has no `headline`** — item `e781118c`.

None of these blocks anything technical.

## 5. Traps this arc paid for

- **A watcher must key on something that cannot already hold a previous run's value.** Mine matched
  the word "failed" inside a *deliberately retained* previous error and reported terminal before the
  run was even claimed. Key on `completed_at > <pinned instant>`, an incremented counter, or a hash
  — never a substring of `error`. (`WRONG_CALLS.md` 2026-08-21.)
- **Conversely, that same hash is how you PROVE a rerun is real.** `attempt_count` 1→2,
  `completed_at` after the arming instant, and `md5(error)` changing are three independent markers;
  the error *text* alone cannot tell a fresh failure from a retained one.
- **A `complete` item is not a built page** — §0. This site now has a live example.
- **The `build provenance` log line is unreadable on `agent-chassis`** — it is inside a single
  1.8 MB JSON log line. Use the binary probe with a must-be-present *and* a must-be-absent control.
- **The 090 loop can fail without answering.** Today it broke twice — an Anthropic 400 at `verdict`,
  then its own `mark_failed` on an untyped `$4` in the failure ladder (recorded in 348 §5). A failed
  loop is **not** a refutation; substitute first-hand verification and **declare that you did**, per
  the 2026-07-31 owner ruling.
- **`pages.build_status` is not the serving state.** `contact-index` is `needs_rebuild` with
  `deployed_at` NULL and serves a healthy 1,267-word page.

## 6. Files of record

`NOTES_mortgagecalculator_couk.md` `## 2026-08-21` (three entries) ·
`README_where_we_are.md` 2026-08-21 (two entries) · `bugs_open/348` ·
`bugs_open/328` (CONTRIB 2026-08-21) ·
`copy_quality_two_stage/CONTRIB_2026-08-21_…` · `WRONG_CALLS.md` 2026-08-21 ·
`scratchpad/linkaudit.sh` (**session-scoped — copy it into the lane dir if you want it again**).
