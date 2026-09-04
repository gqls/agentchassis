# RUNBOOK — domain valuation

## Inbound data contract (what registrar lanes were asked to deliver)

Each registrar lane writes to
`docs/agent_docs/docs024_key_docs_latest/domain_valuation/inbound/`:

- `<registrar>_domains_<YYYY-MM-DD>.csv` — one row per domain:
  `domain,expiry,auto_renew,nameservers` (minimum: `domain,expiry`).
- `<registrar>_valuations_<YYYY-MM-DD>.csv` — if the registrar offers any
  appraisal/valuation: `domain,valuation,currency,source`.
- `<registrar>_listings_<YYYY-MM-DD>.csv` — if domains are listed for sale on
  that registrar's marketplace: `domain,price,currency,status`.
- Afternic feed extras (their lane, agreed 2026-09-02): 5th column
  `price_source` (buy_now|floor|min_offer|none); currency cell is the literal
  `USD-assumed` until an export confirms the marking, then plain `USD` —
  ingest must accept BOTH.

Commit by pathspec, message `domain valuation inbound: <registrar> list`.

**FILE DATES ARE PRODUCTION DATES — the day the file was WRITTEN, never the day
it is meant to be used.** This is the convention every inbound and output file
here follows, and it is stated because breaking it cost another lane real work:
two queue files were dated for the window they were built *for*, and on
2026-09-03 the sedo lane read `..._2026-09-04.csv` in this directory, concluded
that was today's date, and mis-dated a stretch of its own work and a LANDMINES
entry before catching it. Both files were renamed to their production date.
A future-dated filename is indistinguishable from a clock you can trust.

**Re-pull every registrar list the day the pricing sheet finalises** (owner
adds domains on occasion — proven 2026-09-02, when Dynadot grew 451→453 between
snapshots): a registrar count is stale by ADDITION, never by loss, and a fresh
name missing from the sheet is invisible unless you re-pull. Dynadot re-pull is
one ask to that lane; same for the others.

## Nominet list (owner-run; the lane's session cannot touch credentials)

    ! python3 scripts/domains/nominet.py login
    ! python3 scripts/domains/nominet.py walk --months 120 > all_domains.txt

Gotcha: the first walk (2026-09-02 18:48) produced a 0-byte file on a
connection blip — check `wc -l` before trusting it. `--months 120` not 12:
ten-year registrations exist and a 12-month lookahead misses them.

## Finding the live registrar/lane sessions

`ListAgents` → send with `SendMessage` to the bare lane name (dynadot,
porkbun, nominet, spaceship, afternic). Replies arrive as
`<cross-session-message>` turns in this session.

## Dynappraisal daily window (quota 300/day, 429 at the cap; any session may run)

**Announce in the dynadot lane's channel before starting** — three lanes share
one 300/day account and a collision is only ever discovered as a 429. Then:

    ./run_appraisal_window.sh <queue.csv> [more_queues.csv ...]

It takes an exclusive `flock` and refuses a second window, appends every queue in
order to the cumulative `inbound/dynadot_valuations_2026-09-02.csv` (resume by
skip, so a 429 is a clean stop), and prints a duplicate check at the end.
`APPRAISAL_OUT=<path>` sends results elsewhere — use it for a PROBE whose
subjects are not estate domains, which must never land in the estate file.

⚠ **Do not call `dynadot-appraise-all.sh` directly and do not launch it with
`nohup … &`.** Its resume is a `grep` of the output file immediately before the
API call, so two copies double-spend the quota; that happened on 2026-09-04 when
a shell retry started a second walker. ⚠ **And never `pkill -f` its name** — the
pattern matches your own command line and kills the shell you typed it in. Use
`pgrep -af "dynadot-apprais[e]-all"`. Both traps are in `LANDMINES.md`.

⚠ **Do not edit a shell script while it is running** (also 2026-09-04): bash
reads the file incrementally, so the running copy resumes at a shifted byte
offset and dies with a syntax error the file does not have. The queues had
finished; only the run's own summary was lost.

### Rebuilding the queues — a script now, and it matters

    python3 build_working_table.py      # joins every inbound source
    python3 value_domains.py            # values + tiers + sale_status
    python3 build_appraisal_queues.py   # queues, derived from those two

**Run all three after every window.** `build_appraisal_queues.py` reads
`sale_status` off the valuation rather than re-implementing the premium/keep
rules, so an owner ruling reaches the queue the moment the valuation reflects it.
Before it existed the instruction "rebuild the queues after each window" meant
doing it by hand, so it did not happen, and on 2026-09-04 the standing queue led
with 95 rows of a category the owner had ruled a whole-category KEEP nine hours
after that queue was built, plus all 23 owner-withdrawn domains.

Queue order is by **block leverage**, not category rank: `value_domains.py` needs
3 appraisals in a sub-category block before it anchors on that block's median, so
the first three calls in a block re-anchor every domain in it and the fourth
changes almost nothing. `[MEASURED 2026-09-04]` 182 calls bring all 107
under-covered sellable blocks to the threshold and re-anchor 694 domains.
`appraisal_queue_LOW_held_*.csv` is held stock (network-keep, live sites) — real
estate value, but it can never move a sale price, so it queues last and in its
own file so the deprioritisation is visible rather than buried in a sort.

### What the appraiser covers, and what it is actually measuring

- ~~One non-Dynadot test call, to learn whether Dynappraisal takes foreign
  domains.~~ **ANSWERED 2026-09-04 from data already on disk — do not spend a
  call.** Of the 588 appraisals held that morning, **136 were for
  Porkbun-registered (108) and Spaceship-registered (28) domains**. It appraises
  any domain string, owned or not, at any registrar.
- **TLDs PROVEN covered**, each by a real number returned: `.com` `.uk` `.net`
  `.org` `.biz` `.club` `.info` `.shop`. **Proven NOT covered:** `.co.uk`
  `.org.uk` `.me.uk` — HTTP 200 with `"$--"`, a real outcome rather than an
  error, which is why those go through the `.com` proxy route. Untested as of
  2026-09-04: `.cv` `.vin` `.ai` `.io` (4 domains; one probe call each, and
  `build_appraisal_queues.py` emits them as their own file).
- ⚠ **IT IS TLD-AWARE — it prices the actual domain in its actual TLD, NOT the
  keyword.** `[MEASURED 2026-09-04]`, `inbound/PROBE_tld_results_2026-09-04.csv`,
  15 calls appraising the same SLD in both TLDs: `ant.uk` $23,144 vs `ant.com`
  **$8,208,882**; `design.uk` $23,558 vs `design.com` **$3,121,760**;
  `healthcare.uk` $18,193 vs `healthcare.com` **$516,065**. If it were
  keyword-driven those pairs would be equal.
  **So a direct `.uk` appraisal is ALREADY a UK-market number and must not be
  multiplied by the `.uk` TLD factor again** — that double discount was live in
  `value_domains.py` until 2026-09-04 and cost ~5x (`effectiveness.uk`,
  appraised $3,576, carried a $350 keen price). A PROXY appraisal is the opposite
  case: it IS a `.com` value, so it does need the factor.
- **The same probe corroborates the 0.21 `.uk` multiplier by a second,
  independent route.** Across 11 ordinary names the appraiser's own `.uk`/`.com`
  ratio is **0.115-0.185, median 0.165**, against the **0.21** derived from
  realised UK sales in `COMPARABLES_2026-09-03_realised_sales.md` §1.3(c). Two
  methods sharing no inputs agree, so 0.21 is sound and mildly generous. The
  ratio **collapses on premium and short names** (`ant` 0.003, `design` 0.008,
  `healthcare` 0.035) — exactly the class the `PREMIUM-REVIEW` guards hold out of
  automatic pricing, so the multiplier's failure mode is confined to names the
  model already refuses to price.

## Prior-conversation mining

Transcripts live at `~/.claude/projects/-home-ant-projects-agentchassis/*.jsonl`.
Candidate sessions for the .co.uk/.uk valuation discussion (found by grep count
of valuation terms, 2026-09-02): 460a5226, 839df212, db85f55f, 48fb60ee,
7fe0cd84, a107ab07, e9ad9395 — all clustered 2026-08-05..08-14.
Gotcha: session titles go stale after /rename — 460a5226 is titled
"provenance step by step build tools" yet holds 84 valuation mentions; judge by
content, not title.
