# HANDOFF 2026-08-23c — `garden-tools.uk` ESCAPED the hop-two trap on its fifth exemplar draw and is building

**Supersedes `HANDOFF_2026-08-23b_...` (35 minutes old, title wrong) and
`HANDOFF_2026-08-23_garden_tools_continue_here.md` (pre-run).** Both kept; 23b is a worked example
of a wrong call and is bannered as such.

## 1. State at 20:06Z

| thing | state |
|---|---|
| site | `16784842-f7d8-4467-bb5b-eb1fb5c1caba`, `active`, `build_status=pending` |
| submission 1 (17:17Z) | **dead** — `needs_vertical_research` exhausted `attempt_count=3` on `bugs_open/376` |
| submission 2 (19:23Z) | **alive** — `needs_domain_research` complete, `needs_vertical_research` **complete** (attempt 2 of 3) |
| specs | submission/identity/classification/content_direction/design_intent (×2 generations) + **`vertical_landscape`** written 20:05:45Z |
| next | **`needs_strategy` triaged 20:05:55Z** → `domain-strategist`. Hop three of seven |
| pages | none yet · apex still 9-byte `Not found` |

**Expect it to keep going.** Remaining chain: strategy → briefing → site plan → composition →
design → ~12 content pages → rerender. **Do not assume it will finish** — `376` can bite any future
`needs_vertical_research`, and 12 pages is where `311`/`328`/`337` live.

## 2. The one thing that matters most, because it is a correction

**`bugs_open/376` is real but I overstated it, twice, and the retraction is in §4b of that file.**
- **True:** the crawl steps have **no `on_error`**, so one refused exemplar discards the whole stage
  including crawls that succeeded; and `create_next_item` is the **only** estate-wide producer of
  `needs_strategy`, so an exhausted retry budget **is** terminal. Submission 1 died exactly this way.
- **False, retracted:** that the exemplar pool is fixed, that retry is structurally incapable of
  escaping, and that the `competitors_found` branch never fires. **The refused host appeared in 4 of
  5 draws, not 5 of 5**; draw 5 substituted `burgonandball.com` — which came *from*
  `competitors_found` — and sailed through.
- **So the shape is a 4-in-5 tax against a 3-attempt budget**, i.e. it usually kills a build and
  sometimes does not. That is why submission 1 died and submission 2 lived, from identical input.

## 3. Do not re-derive these — they are measured and dated (2026-08-23)

- **Time-to-first-agent = queue depth ÷ ~90s** (24m52s on a busy queue, 81s on a quiet one).
  `build-pipeline-trigger` picks ONE site per tick, `ORDER BY wi.created_at ASC, wi.priority ASC` —
  **FIFO by item age; `priority` only breaks ties inside one timestamp.** Recipe in the RUNBOOK.
- **A site is serialised to one in-flight item** (`NOT EXISTS … status='claimed'`).
- **The classifier is reproducible** — two independent runs on the bare domain gave an identical
  structured verdict incl. `confidence` 0.82; only free-text `industry_tags` drifted. Fixture-safe
  for structured fields **only**.
- **`bugs_open/326` is fixed for the build chain** (migration 572, `recurrence_expected` on all five
  hops) — proven live here: a re-submission at **2h05m51s**, inside the old 3h brake, queued work.
  **Never hand-rename `item_key`s.** 14 keyed steps remain undeclared elsewhere.
- **Re-pin the `311` md5 baselines yourself before any run** — all eight moved 2026-08-20 under
  `bugs_open/283`. A handed-down pin is not a baseline. LANDMINES.
- **`orchestration_states` reaps on an exact sliding 24h clock.** Do not count it for history — two
  sessions got this wrong within an hour tonight. Use `site_work_items` ∪ `site_work_items_archive`.
- **A step's `success: true` can be a dispatch receipt**, not an outcome. Join on `request_id`.
  LANDMINES.

## 4. What to do next

1. **Watch, do not steer.** The lane's whole value is that this build is unassisted.
2. **When pages appear**, run the after-test harness — `after_test.sh` (this session's scratchpad;
   promote it into this directory on first real use). It covers `311` collateral across all **eight**
   incumbents with your own fresh pins, the `260` fingerprint, `<input>` counts per tool page, and
   `328` dead links. **Its collateral check is proven to discriminate** (CHANGED on stale pins,
   UNCHANGED on fresh, same morning).
3. **Report to the owning lanes either way** — `311` and `bugs_closed/260` both asked to hear the
   result whether or not their bug fired.

## 5. Falsifiers

- Any `needs_*` item on this site reaching `failed` at `attempt_count=3` (the build is dead again).
- `376` closing, or the crawl step gaining an `on_error`.
- The apex serving anything other than `Not found` (pages have shipped — go and read them).
- Firecrawl beginning to support `thespruce.com`, which would remove the whole hazard without
  anyone changing our code.
