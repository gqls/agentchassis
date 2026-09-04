# Domain valuation — continue here

Cold start for a fresh session. Written 2026-09-04 at the owner's request.
Read this, then `NOTES_domain_valuation.md` from the bottom up.

---

## 1. What this lane is for

Value every domain the owner holds, so he can **sell roughly the bottom 500 at
keen prices** and keep the rest. This lane is the **integrator**: registrar
lanes deliver inventories into `inbound/`, and the sedo + afternic lanes consume
what this lane produces.

**⚠ SCOPE WAS NARROWED BY THE OWNER 2026-09-03 AND THAT NARROWING IS CORRECT.**
An appraisal is an **input to a conversation about what number to agree**, never
a source that fills a price field. Sedo's Minimum Price may carry a real number
only where he or this lane has agreed it **with him directly** — never derived
from `site_specs.commercial.tier` or any automated appraisal.

The citable case for why, if anyone ever reopens it: this model gave
`healthcare.uk` (which cost him **£40,000**) and `healthcarecareers.uk` the
**identical $149**, because both inherit the same subcategory median.

---

## 2. State as of 2026-09-04

**Estate: 3,029 rows — 3,023 owned + 6 not-owned.**
nominet 1,621 · porkbun 683 · dynadot 472 · spaceship 247.
All categorised into 32 categories / ~170 sub-categories, **zero uncategorised**.

**Sale status:** 2,429 sellable · 403 KEEP · 184 PREMIUM-REVIEW ·
7 OWNER-FIGURE · 6 NOT-OWNED.

**Appraisal coverage: 872 of 3,023 (28.8%)** after the 2026-09-04 window
(299 of 300 calls, ended on a real 429). Still the binding constraint — see §5.
**Every PREMIUM name is now appraised and every one is still held**; both
premium queues rebuild to zero rows.

Anchor confidence across the 2,429 sellable rows: high 551 · medium 1,392 ·
low 438 · very-low 48. Estate value **$6,143,030 over 2,957 rows** — ⚠ **72 rows
are NOT-ESTIMABLE and excluded** (single-word `.co.uk` valued only via their
`.com` twin; their raw figures sum to $17,596,201 and are not a valuation).
**Before 2026-09-04 those 72 names were 74% of a $23.7m headline.** **158 calls close all 99 under-covered blocks and
re-anchor 555 domains** — one part-window, and the highest-leverage work left.

⚠ **The estate grew 78 domains in one day (2026-09-03).** Re-pull every registrar
list the day the pricing sheet finalises: a count here is stale by ADDITION,
never by loss.

---

## 3. Owner rulings — settled, do not re-open

| ruling | date |
|---|---|
| Cut unit = **sub-category blocks** (170 units, median 7) | 09-03 |
| **EXCEPT financial: kept WHOLE** as an advertising-network play — *"I can offer better reach and specifics for advertisers as a network"*. An advertiser buying insurance reach cannot be served from a network whose insurance names were sold | 09-03 |
| Currency **USD**, prices **round UP** to clean figures, floor **$200** | 09-03 |
| Tier bands + 40–60% keen discount: fine as placeholders for now | 09-03 |
| "500" is **approximate** | 09-03 |
| `webdesign.uk` + `webdesign.co.uk` **quoted as a PAIR**, never separately | 09-03 |
| List **all live sites**, priced high; **Sedo minimums stay BLANK** — *"we'll bear with the low balls for a while"* | 09-03 |
| **4 letters with vowels or a `y` in `.com` are worth good money** — his own valuation rule, implemented | 09-03 |
| Acquisition costs: **unavailable** — *"I can't readily gather the prices I paid"* | 09-03 |
| NameSilo **out of scope**; anything traced there is lost | 09-03 |
| `leopardessconsulting.co.uk`: it is **his own consultancy** (fact), and he ruled *"we can work as if leopardess is a paying client if that helps"* — client-grade care **by ruling** | 09-03 |

**⚠ TWO DELIBERATE DECISIONS THAT LOOK EXACTLY LIKE DEFECTS — do not "fix":**
1. `relojistas.com` ($12,000) and `free.me.uk` ($50,000) carry owner-set Afternic
   floors and go on Sedo **blank**. He was asked directly and chose the gap.
   Noted against both rows in `OWNER_FIGURES.csv`.
2. Sellable here (2,429) is deliberately **lower** than the sedo sheet (2,940):
   mine are *pricing* holds, theirs are *listing* holds. They should differ.

---

## 4. THE central finding — read before touching any number

**The model CANNOT PRICE the estate's best names, and it is one cause, not many.**
Where a domain has no appraisal of its own it anchors on its category median —
and a premium name is by definition the one that deviates most from a median.

> **CORRECTED 2026-09-04 — this section used to say "systematically UNDER-values"
> and the direction was wrong, not just imprecise.** Once the proxy appraisals
> landed, the same class started coming out **catastrophically HIGH**:
> `cartoon.co.uk`, listed below as ~2× under, is now **~187× OVER** the £5,000+
> he paid, and the owner's `scales.co.uk` figure (£3,500, given 09-04) is **~89×
> OVER**. Both are single dictionary words anchored on a `.com` proxy.
> **The error is not a direction, it is a VARIANCE** — the instrument has no
> idea what these names are worth, and "under-values" was a description of the
> only sample we had. Do not reason from the direction; reason from the class.
> This matters because "it under-values" invites a multiplier, and a multiplier
> is exactly what cannot work here.

Measured cases (the first four are the ORIGINAL under-valuation set, still
correct as of the dates given; the last two are the reversal):

| name | model | reality | ratio |
|---|---|---|---|
| `healthcare.uk` | $149 | **£40,000 paid** | **~340×** |
| `free.uk` | $215 | sibling `free.co.uk` sold **~£160,000** | — |
| `holidaytime.com` | $995 | **$12,000 realised** | 12× |
| `relojistas.com` | $1,490 | **$12,000** owner floor | 8× |
| `cartoon.co.uk` | ~~$2,934~~ **now $739,424** | **£5,000+ paid** | ~~~2× under~~ **~187× OVER** |
| `scales.co.uk` | **$393,917** | **£3,500 paid** (owner, 09-04) | **~89× OVER** |
| `2w.uk`/`4l.uk`/`5s.uk` | $200 | realised `.uk` shorts **£2,500–£5,200** | ~15× |

> **NARROWED 2026-09-04 — it is the FALLBACK that cannot see a premium name,
> not the appraiser.** Asked directly, Dynappraisal put `healthcare.uk` at
> **$18,193** against the $149 the category median gave it: ~2.8× under the
> £40,000 paid rather than ~340×. That is an argument FOR coverage, and it does
> not soften the refusal below — 2.8× low is still far too low for a price
> field, and it is the same instrument that offers **$4,190,000** for
> `ant.co.uk`. The class is not under-valued by a correctable factor; it is
> *unpriceable* by this instrument, in both directions.

**The response is refusal, not a better multiplier.** No multiplier can help: in
one block the model gave `healthcare.uk` and `healthcarecareers.uk` the *same*
number. Four guards now hold names out of the algorithmic tail:

- `OWNER-FIGURE` — any domain in `OWNER_FIGURES.csv` (7)
- `PREMIUM-REVIEW:single-word` — a single English dictionary word (144 exist)
- `PREMIUM-REVIEW:short-name` — ≤3 characters
- `PREMIUM-REVIEW:4-letter-com` — the owner's own rule

**⚠ The guards are the deliverable as much as the prices are.** Two of them were
written hours before they caught an unrelated, worse case. Keep them broad:
holding an ordinary name back for a week costs nothing; selling a £40,000 asset
for £160 does not.

**⚠⚠ AND THEY MUST NEVER BE CONDITIONED ON MISSING DATA — this was live and
firing.** Until 2026-09-04 two of the four read `… and not r['dynappraisal']`,
so the premium appraisal queues in §5a — the job this handoff tells you to do
FIRST — **switched the guards off name by name as they succeeded.** 150 of the
180 holds carried that conditional, 137 flipped inside one window, and the
pre-fix code run against that day's data puts **129 premium names into the sale
list carrying $9,447,275 of automatic asking prices**. It had already been
firing silently on four names for days: `effectiveness.uk`, an ordinary English
word appraised at $3,576, sat in the sell list at **$350**. All four guards are
now unconditional. **Nothing about this was visible in the output** — coverage
went up while protection went down. Full entry in `LANDMINES.md`.

**MOST VALUABLE THING TO ASK HIM FOR: more real figures.** Four surfaced by
accident on 09-03 and every one overturned a model number by 2× to 340×. One
realised price is worth more than 300 machine appraisals.

---

## 5. What to do next

**5a. Run the daily appraisal window** (300/day, shared account — **announce in
the dynadot lane's channel before starting**). Gotchas and what the appraiser
actually measures: `RUNBOOK_domain_valuation.md` §"Dynappraisal daily window".

```
./run_appraisal_window.sh inbound/appraisal_queue_direct_2026-09-04.csv \
                          inbound/appraisal_queue_proxy_2026-09-04.csv
```
**Use that wrapper, not the walker directly** — it holds an exclusive `flock`,
and two concurrent walkers double-spend the quota with no visible symptom but an
early 429. Premium queues are empty; the bulk queues hold 1,208 direct + 670
proxy, ordered so the first 158 calls close every under-covered block.
`appraisal_queue_LOW_held_*.csv` (271) is held stock — never a sale price, so it
queues last. ~7 windows to full coverage.

**Then rebuild all three, every time** — `build_working_table.py`,
`value_domains.py`, `build_appraisal_queues.py`. The last one is new on
2026-09-04 and exists because "rebuild the queues after each window" had been
written down since 09-03 with nothing to run: the standing queue that morning
led with 95 rows of the financial category the owner had ruled a whole-category
KEEP **nine hours after that queue was built**, and carried all 23
owner-withdrawn domains.

**5b. Re-run the pipeline after any new data.** Deterministic, re-runnable:
```
python3 build_working_table.py   # joins every inbound source -> WORKING_table.csv
python3 value_domains.py         # values + tiers -> VALUATION_<date>_draft.csv
```

**5c. Re-run the cut when coverage is high**, and only then.
`CUT_2026-09-03_provisional.md` is **the shape, not the sale list** — it is built
on 19% real appraisals, so the ORDER of its families is not yet trustworthy.

**5d. Open questions with the owner:**
- ~~`mieleonline.com` and `webuyanycarandvan.com` — trademark-adjacent, still in
  the sale.~~ **ANSWERED 2026-09-04: he withdrew both** (relayed via the sedo
  lane), joining `rolex-submariners.com`. Both are in `WITHDRAWN_owner.txt` and
  the sedo lane has consolidated all three into one exclusion file. For the
  record, sedo since confirmed a live US trademark on "We Buy Any Car" (USPTO
  6011054, 2020) covering online vehicle resale — the exact category.
- Any other **network category**? financial is held; home-garden (357),
  web-digital (318) and automotive are each large enough to be one.
- Is any built site actually **a third party's**? Nothing in the data marks one,
  so a real client site is currently protected by nothing but nobody listing it.
- 4 person-name domains unresolved, not declined: `ianstirling.com`, `kapoor.uk`,
  `keeler.uk`, `anne-marie.co.uk`.

---

## 6. Lane-specific landmines

- **A ratio over a banded variable measures the OTHER variable.** His Afternic
  asks are BANDS, not judgements — 250 of 419 are the identical $4,999, 845 of
  1,215 floors are exactly $10,000. Count distinct values before quoting any
  ratio. (`WRONG_CALLS.md`, 2026-09-03.)
- **DNS cannot tell "expired" from "never delegated"** on this estate. Ask the
  registry (`check_registration.py`, register **OPP-017**), with controls both
  ways. (`LANDMINES.md`.)
- **A glob naming a file convention silently reads the old file.** Cost this lane
  31 live sites read as sellable stock for most of a day. (`LANDMINES.md`.)
- **Dynappraisal answers HTTP 200 with `"$--"` for TLDs it does not cover**
  (`.co.uk`/`.org.uk`/`.me.uk`). It covers `.com`/`.net`/`.uk`. (`LANDMINES.md`.)
- **Dynappraisal values ANY domain string**, owned or not — which is what makes
  the `.com`-proxy route work for `.co.uk`. A proxy value must stay visibly
  distinct end-to-end (`appraisal_kind=proxy`) and never read as a direct one.
- **A guard conditioned on MISSING DATA is disarmed by the job that fetches it.**
  The premium queues switched off the guards they exist to serve — see §4. The
  general form, and the check, are in `LANDMINES.md`. When you write any
  hold-out, ask whether the next obvious task satisfies its condition.
- **The appraiser is TLD-AWARE — never discount a direct appraisal by TLD again.**
  `ant.uk` $23,144 vs `ant.com` $8,208,882 (`PROBE_tld_results_2026-09-04.csv`).
  A direct `.uk` appraisal is already a UK-market number; only a `.com`-basis
  anchor (a median, or a proxy) takes the TLD factor. This was a live ~5× double
  discount until 2026-09-04.
- **A block median must be built from SELLABLE stock only.** Held-out names are
  held because this appraiser cannot price their class, so letting them into the
  pool propagates the error sideways: pets-vet's median was set by proxies of
  `felines.co.uk` ($169,614) and `veterinary.co.uk` ($48,517) and priced an
  invented brandable at $6,250. **§4's finding runs in both directions** — with
  premium names out of the pool it drags the best names down; with them in, it
  drags ordinary names up.
- **Do not edit a shell script while it is running** — bash re-reads it at a
  shifted offset and dies with a syntax error the file does not have.
- **A rule in prose is not a control.** Caught twice in one day here: the
  quote-as-a-pair ruling and the owner figures were both documented but
  unenforced, and `holidaytime.com` ($12,000 realised) sat in the sell cut at
  $450 as a result. When a ruling lands, ask: *is it enforced, or only written
  down?*
- **File dates are PRODUCTION dates**, never intended-use dates. Two future-dated
  queue filenames here made another lane mis-date a day's work.
- **A quoted ruling is the thing nobody re-checks.** D4's "paying client
  (leopardess)" example was wrong and passed through three lanes because none of
  us opened the site.

---

## 7. Files

| file | what |
|---|---|
| `WORKING_table.csv` | the join of everything — 3,029 rows, rebuild any time |
| `VALUATION_2026-09-03_draft.csv` | values, tiers, keen prices, sale_status |
| `CUT_2026-09-03_provisional.md` | the provisional bottom-500, by family |
| `OWNER_FIGURES.csv` | **read by the model** — his real numbers outrank everything |
| `OWNER_STATED_PRICES.md` | the prose version + the cost-base problem + mitigations |
| `WITHDRAWN_owner.txt` · `NETWORK_KEEP_categories.txt` · `QUOTE_PAIRS.txt` | owner rulings **as data** |
| `COMPARABLES_2026-09-03_realised_sales.md` | realised UK/.com sale evidence, sourced |
| `METHOD_2026-09-03_live_sites_tier.md` | why the appraiser fails at the top |
| `DECISIONS_2026-09-03_needed_before_the_cut.md` | the four decisions (all now ruled) |
| `UNACCOUNTED_2026-09-03_answer.md` | the "which expired?" answer |
| `LISTING_ACCOUNT_2026-09-03_finding.md` | the wrong-account listings (resolved) |
| `PRIOR_ART_2026-09-02_*.md` | the earlier valuation talk is NOT on this machine |

Scripts: `build_working_table.py` · `value_domains.py` · `categorise_domains.py` ·
`check_registration.py` (register **OPP-017**) · **`build_appraisal_queues.py`**
(rebuilds every queue from the current estate — run it after every window) ·
**`run_appraisal_window.sh`** (the flock'd wrapper; `APPRAISAL_OUT=` for probes).

⚠ `VALUATION_2026-09-03_draft.csv` is **regenerated in place** and its name is
the series start, not its production date — it is the one file here that breaks
the lane's own file-date convention. Check `git log` on it, not the filename.

---

## 8. Cross-lane contracts

- **sedo** — consumes this lane's prices; owns the Sedo import sheet (draft10,
  2,942). Their `EXCLUDED_live*.txt` files are the live-site fence; **union them,
  never take the newest**.
- **afternic** — supplies the portfolio export (price_source column: only
  `buy_now` is an ask; floor/min_offer are floors). Owns bulk repricing.
- **dynadot** — owns the appraisal walker (`scripts/domains/dynadot-appraise-all.sh`,
  handles the `proxy_domain` column pair). Announce windows there.
- **porkbun / spaceship / nominet** — inventories delivered; nothing outstanding.
- **copy_quality_two_stage** — owns D4 and `site_specs.commercial.tier`. ⚠ **That
  tier (1|2|3) is NOT this lane's tier (A–E).** No mapping exists and none should
  be invented silently.
