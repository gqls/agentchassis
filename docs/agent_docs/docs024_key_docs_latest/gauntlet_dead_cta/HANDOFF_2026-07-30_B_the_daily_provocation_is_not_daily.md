# HANDOFF B (2026-07-30) — "Today's Provocation" has no mechanism that could ever change it

**Start a fresh thread on this.** Owner-raised 2026-07-30: *"the provocation
didn't change today and has never changed."* **Measured: he is right in
substance, and the cause is that daily rotation was never built.**

> **TAKEN 2026-07-31 — this handoff is being worked; do not start a second thread
> on it.** Lane: `docs024_key_docs_latest/provocation_pipeline/`, cold-start =
> `PLAN_2026-07-31_provocation_pipeline.md` (commit `60f56d76c`). Every fact above
> re-verified and still true. Four things this thread found that change what the
> next reader should believe:
>
> 1. **The SERVER reads the `today` key**, not just the browser —
>    `internal/tools-api/handlers/round.go:44` (`FetchProvocation`, 5-minute
>    cache, 503 if absent). So a client-side date selector over a pool — the
>    cheapest design and the one this handoff's §"three shapes" does not warn
>    against — would make the page display one provocation while the Gauntlet
>    argues another. **Selection must happen at generation time.**
> 2. **§"three shapes" mis-prices option 2.** The scheduled-publish plumbing
>    already exists and is proven live at the artefact: `content-feed-refresh`
>    (6-hourly) ends in the `git_commit` action, and three sites served
>    `latest-news.json` stamped inside its 2026-07-30 13:53→14:02Z window. The
>    `gh api --input -` sequence in §Landmines is the **hand** path; it should not
>    be what we automate.
> 3. **Prior art, half-built:** `docs/social001_vonc_tiktok_social/PLAN_spark_provocation_pipeline.md`
>    (2026-06-25) designed this as a news-pipeline clone. Its Phases 1–2 shipped —
>    that is where today's JSON and JS shells came from. Phases 3–4 never did.
> 4. **CORRECTION — the archive page is NOT broken**, so §"The archive question"'s
>    closing advice ("it may already be broken independently") should not be acted
>    on. Rendered in headless chromium: all 8 entries paint, 7 openable and the
>    8th deliberately not (no `detail_body`), empty state correctly hidden. The
>    1,293-char figure was accurate but was produced by a probe looking for
>    **today's** headline, which correctly is not on that page; a low count is not
>    diagnostic of breakage. Full account in `WRONG_CALLS.md` 2026-07-31.
>
> **Owner decisions taken 2026-07-31:** archive rule settled — a provocation is
> archived when the next one is published, never during its own day. Direction is
> LLM generation behind a safety/quality gate. **Relevant to HANDOFF C:** that
> archive rule also closes the third leak path immediate promotion would have
> opened, and it unblocks C's option 3 (home shows yesterday's as the sample).
> Also relevant to C — the paired-provocation shape the owner raised makes the
> sealed reveal load-bearing rather than decorative, which argues **against** C's
> option 2 (retire the seal). See PLAN §5.

## What is actually true (all measured 2026-07-30)

- The served feed is a **static file**: `https://vonc.com/data/provocations.json`,
  HTTP 200, 9,797 bytes, carrying `"generated_at": "2026-07-26T00:00:00Z"` — built
  **four days before** it was read, with `today` a hardcoded key inside it.
- **Nothing regenerates it.** No row in `scheduled_tasks` references vonc or
  provocations:
  ```sql
  SELECT name FROM scheduled_tasks
  WHERE input_data::text ~* 'vonc|provocation'
     OR COALESCE(pre_query,'') ~* 'vonc|provocation'
     OR name ~* 'vonc|provocation';   -- returns 0 rows
  ```
- **The builder has no rotation logic at all.**
  `p4_sources/build_provocations.py` is 232 lines with every provocation as a
  Python literal: `TODAY` is a fixed dict and the archive dates are string
  literals (`"5 Jul"`, `"4 Jul"`…). No date arithmetic, no pool, no selection, no
  randomness. Changing the provocation means editing that file, running it, and
  committing the JSON.
- **It has changed six times, all by hand** (`gh api repos/gqls/sites/commits?path=vonc.com/data/provocations.json`):
  29 Jun 16:32 (create), 29 Jun 17:40, 4 Jul, 7 Jul, 11 Jul ("drop dead 'lobby'
  key"), **26 Jul 14:31** ("regenerate — real CTAs, honest stats"). Nothing since.

> **Precision, so nobody repeats an overstatement:** the file is not immutable —
> it changed six times. What has never happened is a change *on a cadence*, or a
> change *by any mechanism*. "Never changed" is very nearly right and "no
> mechanism exists" is exactly right; prefer the second when writing this up.

- **The archive is separately stale.** `archive.entries` holds 8 provocations
  dated **28 Jun – 5 Jul** — a contiguous run that stopped 25 days ago. `today`
  is not among them and has no `slug`, so today's provocation is not linked into
  the archive at all.

## Why this is worse than a stale file

The site's entire proposition is *"every day, one provocation"* — that phrasing is
on the about page and in the arena copy. **A site that says "every day" and serves
the same statement for four days is making a false claim in the same class the
rail forbids** ("nothing on vonc.com claims a number that is not true by
construction"). It is not a numeric claim, but it is a claim about the product's
behaviour that the product does not honour.

It also silently undercuts the distribution experiment (owner's ruling
2026-07-29): the daily provocation is one of the two travelling artefacts. Posting
the same one repeatedly is the wrong test of whether anyone will come back.

## The decision this needs before any code

**Where should provocations come from?** Three shapes, ascending cost:

1. **A pool with date-based selection.** Keep the hand-written literals, add
   enough of them, select by date so it rotates deterministically. Cheap, honest,
   finite — it runs out, and it repeats on a cycle if not topped up.
2. **A scheduled regeneration** of `provocations.json` (a `scheduled_tasks` row,
   which is how everything else on this platform is paced). Needs the builder to
   *choose*, which means either a pool (see 1) or generation.
3. **LLM-generated daily provocations.** Only shape that is genuinely daily and
   open-ended — and it lands squarely in `bugs_open/149` C1 territory: prose
   written by an LLM with no claims gate. On 2026-07-29 that exact path put four
   false statements onto a live homepage (see `bugs_open/149` § "C1 — WITNESSED").
   **If this option is chosen, the claims gate is part of the work, not a
   follow-up.**

The owner should pick. Option 1 plus a scheduled publish is the honest minimum
and does not require solving the gate first.

## The archive question, which is separate

Today's provocation never joins the archive (no `slug`, not in `entries`). So
even with rotation, yesterday's provocation would vanish rather than accumulate.
**The "See All Provocations" CTA already exists** and points at
`/provocations/index.html` — a page which today paints neither today's
provocation nor, apparently, much else (1,293 chars of visible text). Check what
that page is actually showing before designing the archive; it may already be
broken independently.

## How to verify a fix

Rotation cannot be verified in one day. Assert the *mechanism*, not the outcome:

```bash
# 1. the feed is regenerated on a cadence
SELECT name, enabled, last_triggered_at, interval_seconds
FROM scheduled_tasks WHERE name ~* 'provocation';

# 2. today's value is selected, not hardcoded — same code, two dates, two results
#    (run the selector with an injected date; do not wait a day to find out)

# 3. the served file is fresh
curl -s https://vonc.com/data/provocations.json | python3 -c \
  "import json,sys; print(json.load(sys.stdin)['generated_at'])"
```

**Do not verify by looking at the page and seeing a provocation** — a hardcoded
one looks identical to a rotated one on any single day. That is how this survived
a month.

## Landmines

- **The feed is published to `gqls/sites`, NOT to the database.** There is no
  provocations table. `build_provocations.py > provocations.json`, then a GitHub
  contents PUT with `{message, content:<base64>, sha}` — payload on **stdin** via
  `gh api --input -`, because argv blows `ARG_MAX`. Full sequence in
  `RUNBOOK_gauntlet_dead_cta.md` §8.
- **`sites.github_repo` picks the deploy repo and the WRONG one succeeds
  silently** — a green run with no change on the site.
- Both the home page and the gauntlet page fetch this file **client-side**, so
  `curl` of the HTML shows the provocation nowhere. Render, don't grep. See
  `scripts/provocation_visibility.py`.
