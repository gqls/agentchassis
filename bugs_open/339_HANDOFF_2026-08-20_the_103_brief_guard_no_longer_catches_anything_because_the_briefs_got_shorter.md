# 339 — `bugs_closed/103`'s brief guard catches nothing any more, because the briefs got shorter

**Filed** 2026-08-20 by the `meta_description_never_backfilled` lane. **Status: OPEN.**
A live recurrence of a closed bug, in its own guard's blind spot.

> **Resolve by SLUG** (`103_brief_guard_catches_nothing`) — bug numbers collide on this
> tree, and `git log` the FILE PATH, not the number.

---

## 1. What a reader needs first

`bugs_closed/103` was this: **tool pages published their internal build brief as the
public meta description**, so Google printed generator instructions under the search
result. The worst live case was 1,206 characters of *"no fetch calls, no backend"* on
vonc.com's Arena page.

Its fix is `datahelpers.PublicMetaDescription(candidate, composed)`
(`platform/orchestration/datahelpers/meta_description.go`), a shared gate with **two
signals**:

1. **Length** — anything over `maxPublicMetaDescription = 320` chars reads as internal.
   Chosen because *"the shortest observed brief in the bugs_open/103 census was 449
   characters and the longest 1,206"*.
2. **`briefMarkers`** — a regex of phrases measured in that census:
   `no fetch calls|elements, in order|embed [0-9]+ sample|fully self-contained client-side|no backend\)|:\s*\(1\)`.
   Its own comment says why it exists: *"length alone would miss a SHORT brief, which is
   the failure this guard would otherwise still allow through."*

**Both signals now miss, and the failure is back.**

## 2. The measurement

`[MEASURED 2026-08-20]` over `pages` where `status='active'` and the column is non-empty
(693 pages):

| | |
|---|---|
| descriptions **over 320** chars (signal 1 would fire) | **0** |
| descriptions in the **200-320** band | **11** |
| of those 11, matched by **`briefMarkers`** (signal 2) | **0** |
| page types | **9 `tool`**, 2 `blog-post` |

So the population that used to be 449-1,206 characters now sits at 200-320. **It moved
underneath the guard.** Signal 1 fires on nothing; signal 2 was the designed backstop for
exactly this and matches none of them.

## 3. What is actually being published

Verbatim from `pages.meta_description`, live today:

- `gamesdesign.co.uk` / `tool-wave-difficulty-ramp` (301 chars) —
  *"Companion to the Spawn Rate Balancer. Designers input player power growth per wave
  (DPS scaling, healing, item unlocks) against enemy health and spawn…"*
- `gamesdesign.co.uk` / `tool-drop-rate-tuner` (302) —
  *"Tune loot drop rates against player experience. Set a base drop chance,…"*
- `gamesdesign.co.uk` / `tool-probability-curve-visualiser` (280) —
  *"Lets designers plot and compare multiple probability distributions (uniform, binomial,
  geometric, custom weighted) on a single chart. Directly support…"*
- `gamesdesign.co.uk` / `tool-stat-budget-allocator` (272) —
  *"Lets designers define a total stat budget for a character or item tier,…"*
- `webdesign.co.uk` / `tool-css-unit-converter` (296) —
  *"Converts between px, rem, em, vw, and vh units given a base font size an…"*

These are specifications addressed to a builder. *"Lets designers…"*, *"Designers
input…"*, *"Companion to the Spawn Rate Balancer"* — third person, describing inputs and
outputs. Not one of them is written to a visitor, which is what a meta description is for.

Full list:
```sql
SELECT s.domain, p.name, p.page_type, length(p.meta_description) AS len, p.meta_description
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE p.status='active' AND length(COALESCE(p.meta_description,'')) BETWEEN 200 AND 320
ORDER BY len DESC;
```

## 4. Provenance — what is established and what is not

**Established:** the **planner did not write these.** `site_plan_pages` for the current
plan has **no rows** for those tool page names — tool pages are created by the tool-deploy
path, not the site plan. So this is not a consequence of migration `485` (which taught
`build-site-planner` to write descriptions) and 8 of the 11 predate `485` anyway.

**Established by elimination:** it is not the **composed** fallback either.
`composedToolMetaDescription` emits a fixed sentence — *"An interactive %s, free to run in
the browser. The companion guide sets out the method behind it, so you can check the
working."* — about 120 characters and nothing like these. `PublicMetaDescription` returns
either the candidate or the composed fallback, so by elimination **these are the CANDIDATE
side, passed by the guard.** That is 103's mechanism exactly: the candidate for a
`component_level='tool'` row is the brief.

~~**NOT established, and stated so rather than asserted:** I did not trace each string back
to its specific writer row. A `LIKE` join to `content_components.description` returned no
match, which is inconclusive rather than contrary — those rows may have been regenerated
or edited since the page was created. **Whoever fixes this should establish the exact
writer before changing the guard**, because the remedy differs depending on whether the
brief arrives from `content_components.description`, from a tool spec, or from somewhere
a census has not looked.~~

**ESTABLISHED 2026-08-20 ~17:15Z by the accepting lane (`webdesign_tool_rebuilds`), for the
TOOL-page class.** A PREFIX join — `left(cc.description,120) = left(p.meta_description,120)`
— matches **7** of the damaged strings to live `component_level='tool'` rows with the same
`function`; **4** additionally prefix-match `add_tool` items' `spec->>'description'`. (The
original `LIKE` join failed on LIKE semantics against these strings, not because the source
differs.) So the pipeline is confirmed end to end:
`add_tool spec.description` → `content_components.description` → `PublicMetaDescription`
(both signals blind in the 200–320 band) → `pages.meta_description` — §5 candidate 3's
premise holds exactly.

**Two rows of the census and one NEW row are a DIFFERENT writer, still untraced:** the
non-tool pages (`robot-hands.com/robot-demand-step-change`,
`leopardessconsulting.co.uk/hierarchical-multi-agent-orchestration-explained`, and —
created **2026-08-20**, i.e. the population GREW after this file was written —
`dartsonline.com/darts-calendar-density`, 291 chars) match neither tool components nor
`add_tool` specs. That sub-class is live today and is NOT closed by the tool-seam fix.

## 5. Fix candidates, ordered by what makes the bad state unrepresentable

1. **Re-derive the length threshold from the CURRENT population, not the 2026-07 census.**
   320 was chosen against briefs of 449-1,206 chars. Today's briefs are 272-302. The
   number is stale, not wrong-headed. ⚠ But it cannot simply drop to ~250: legitimate
   descriptions run to 177 chars today and the band is narrow, so a naive drop trades this
   failure for false refusals of good copy. **Measure the two populations before picking a
   number, and say what the disconfirming result would have looked like.**
2. **Strengthen `briefMarkers` from the CURRENT corpus.** It was measured once, in July,
   and has not been re-derived. *"Lets designers"*, *"Designers input"*, *"Converts
   between"*, *"Companion to the"*, *"Set a base"* — second-person-absent, imperative-to-a-
   builder constructions — are the shape now. This is the signal the guard's own author
   said was the backstop for short briefs; it just needs re-fitting.
3. **Stop the brief being the candidate at all.** The most durable fix: don't hand
   `content_components.description` to `PublicMetaDescription` as a candidate for a
   `component_level='tool'` row, because for those rows it is *by definition* the brief.
   The guard exists to catch a mistake that the call site could avoid making. This makes
   the bad state unrepresentable rather than detectable.
4. **Repair the 11 live rows.** Separate from the class fix, and note they are *tool*
   pages — the framework has a composer for exactly these
   (`composedToolMetaDescription`), so this is a re-derivation, not authoring, and does
   not need an LLM.

⚠ **Whatever the fix, re-run it over what the blind guard already cleared.** 693 live
descriptions passed a check that could not fire; a fixed guard should be swept back over
all of them, not just applied going forward. (MEMORY: *a PASS from a BLIND check outlives
the blindness*.)

## 6. How to verify a fix

```sql
-- must go to 0, and quote the denominator beside it
SELECT count(*) FILTER (WHERE length(meta_description) BETWEEN 200 AND 320) AS brief_band,
       count(*) AS total_with_desc
FROM pages WHERE status='active' AND COALESCE(meta_description,'')<>'';
```

**Induce both arms or the test proves nothing:**
1. A 280-char brief-shaped candidate must be **REFUSED**.
2. A legitimate 177-char human description (there are live examples) must still be
   **ACCEPTED**. A guard tested only on its refusing side is indistinguishable from one
   that refuses everything.

## 7. Why this is filed rather than fixed here (and see 7b: rehomed)

It is not this lane's defect. None of the 11 was written by the meta-description
backfiller (`SEO-004`) — that action *reuses* `MetaDescriptionLooksInternal` and its
descriptions run 65-177 chars. This was found while measuring the column for
`bugs_open/320` and is flagged rather than annexed: changing a shared guard used by two
tool-creation paths is a different blast radius from filling a blank column, and
re-deriving both of its signals needs the tool-page population's owner.

## 7b. REHOMED 2026-08-20 → `webdesign_tool_rebuilds`

`scripts/who-owns.py 103` named `gauntlet_dead_cta`, but that is an artefact of who *cited*
103 a month ago: its recent commits are gripper/SMTP work and it has no claim on this seam.
Rehomed on the following evidence rather than on the ownership script's first answer.

**Why that lane:**

1. **It owns the thing that writes the brief.** Migration `481` (2026-08-19) promoted six
   recurring tool defects out of per-brief prose and into the **tool generator's own
   quality contract** (rules 15-20). That commit belongs to this lane
   (`webdesign_tool_rebuilds/HANDOFF_2026-08-19_continue_here.md`). 339's root is that a
   generator brief becomes public copy, so the brief's contract is the seam.
2. **It owns both call sites of the guard.** `create_tool_component_action.go` and
   `deploy_tool_action.go` are the only two callers of `PublicMetaDescription`, and this
   lane has been committing to them all week (`TL-043`, `TL-044`, `TL-047`, RFC_036 §9.3),
   **93 commits in the last 7 days**. Nobody else is near them: the guard file itself has
   been touched exactly **once ever**, by 103's original fix on 2026-07-27.
3. **Its stated principle is 339's best fix candidate, already adopted.** That lane's own
   summary of 2026-08-19 is titled *"the lane stopped fixing tools and started fixing the
   thing that builds them"*, and the owner confirmed that direction. §5's candidate 3 —
   stop handing a `component_level='tool'` brief in as a candidate at all — is a
   framework-level fix of exactly that kind, not a per-tool repair.
4. **9 of the 11 affected pages are `tool` pages.**

**⚠ The honest caveat, so this does not read as a neater fit than it is.** That lane's
*subject* is webdesign.co.uk's 63 imported tools, and only **1** of the 11 affected pages
is on webdesign.co.uk — the affected population is mostly `gamesdesign.co.uk` (6), plus
leopardessconsulting (2), oufe (1) and robot-hands (1). So this is **the lane that owns the
SEAM, not the lane that owns the damage.** If it would rather hold the seam fix and let the
row repair (§5 candidate 4) go elsewhere, that split is reasonable and this file should be
updated to say so.

**SPLIT AGREED 2026-08-20 ~17:15Z** (message from `webdesign_tool_rebuilds` to the filing
lane, acknowledged in both lanes' records):
- **`webdesign_tool_rebuilds` takes:** §5 candidate 3 at BOTH call sites (a
  `component_level='tool'` row's description is never a candidate; the composed copy is
  used unconditionally), the register entry in `register/tool-lifecycle.md`, the two-armed
  tests of §6, and the council round. Go change — inert until a chassis roll.
- **`meta_description_never_backfilled` (320) keeps:** the row repair (§5 candidate 4; 12
  rows now, 10 off webdesign.co.uk; the 9 tool pages are a `composedToolMetaDescription`
  re-derivation, no LLM), and the trace of the NON-tool writer (§4 addendum) — which is
  where the population is still growing.

**Also relevant and not yet done:** `register/tool-lifecycle.md` records **nothing** about
the tool page's meta description — `grep -n "meta_description\|PublicMetaDescription" `
returns no hits — so this seam is unregistered as well as unowned. Whoever fixes it owes a
register entry, which is also how the next lane avoids rediscovering it.

**Told, not merely measured** (owner ruling 2026-07-29 §3): the lane has been notified
rather than left to find this in a bug index.

## 8. Provenance of this file

Not run through `090`. **Substituting first-hand verification per the owner ruling of
2026-07-31, and declaring the substitution rather than omitting it:** every figure is a
direct query over `pages` (quoted in §2 and §6), the published strings are verbatim from
the column, the guard's two signals were read in source
(`datahelpers/meta_description.go`) and its `briefMarkers` regex was run against the
population to get the 0, and the planner was eliminated by querying `site_plan_pages`.
§4 marks explicitly what is **not** established. The claim is bounded — a guard's two
signals no longer match a population that moved — and asserts no cause outside the
symptom.

Related: `bugs_closed/103` (the original), `bugs_open/320` §13 (where this was found),
register **SEO-004**.


---

## 9. 2026-08-21 — the split with `webdesign_tool_rebuilds`, and the non-tool sub-class

### The split, agreed both ways

`webdesign_tool_rebuilds` accepted §7b's offer and took **§5 candidate 3 at both call
sites**: for a `component_level='tool'` row, never pass the description as a candidate,
always use `composedToolMetaDescription`. Council-approved (corr `2ff9e215`) and **live in
`v1.0.1321`** by their ancestry check.

They handed back, and this lane accepts: **the row repair (the 9 tool rows + 3 others)**
and **the non-tool writer trace** below.

⚠ **Their honesty note on the seam fix's proof, which changes the row repair's reasoning:**
TL-048's demand proof is *inconclusive by route* — rebuilding an EXISTING page goes through
the adopt path (`Refresh: []`), which never writes `meta_description` at all, so the new
composed write only fires on a genuinely NEW tool page. **Consequence for the repair: the
9 rows will not refill themselves from the brief, and equally will not be repaired by a
rebuild. They need a direct write.**

### ⚠ CORRECTION TO §4 — my LIKE join was the error, not evidence of a different source

§4 said a `LIKE` join to `content_components.description` returned no match and called that
*inconclusive*. It was **my bug**. I used `LIKE left(cc.description,60)||'%'`, and the
descriptions contain characters that are significant to `LIKE` (`%`, `_`), so the pattern
could not match literally.

`webdesign_tool_rebuilds` used `left(cc.description,120) = left(p.meta_description,120)` —
plain equality — and **established the writer**: 7 of the damaged strings prefix-match a
live `content_components.description` on a `component_level='tool'` row with the same
function, and 4 also prefix-match `add_tool` items' `spec->>'description'`. So the chain is

`add_tool` spec description → `content_components.description` → `PublicMetaDescription`
(both signals blind in 200-320) → `pages.meta_description`.

**§4's "NOT established" is now established, and by someone else's query, because mine was
wrong.** Use equality on a fixed prefix, not `LIKE`, when comparing two free-text columns.

### The non-tool sub-class — a DIFFERENT writer, and I did not find it

Three of the band are not tool pages and match neither `add_tool` specs nor tool
components: `robot-hands.com/robot-demand-step-change`,
`leopardessconsulting.co.uk/hierarchical-multi-agent-orchestration-explained`, and — new on
2026-08-20 — `dartsonline.com/darts-calendar-density` (291 chars):

> *"Barry Hearn warned top players about skipping tournaments and Euro Tour withdrawals
> left organisers with a headache. Set against the calendar itself — 30 Players
> Championship events a season through 2024, 34 since 2025 — these are one story about
> schedule density, not four about discipline."*

**That is a commissioning note to a writer, not a description for a reader** — *"these are
one story about X, not four about Y"* is an editorial instruction. Same class as 103 (an
internal brief reaching a public column), different producer.

**What I eliminated** `[MEASURED 2026-08-21]`:
- **not the site planner** — absent from `site_plan_pages` in **every** plan, current or not;
- **not the tool path** — not a tool page, no `content_components` match;
- **not `apply_gap_plan_action`** — none of its five `ON CONFLICT` clauses writes
  `meta_description`;
- **not the rerender** — the earliest orchestration carrying the text
  (`f57a3fbd`, 17:03:53Z) is a `page-rerender` and the page already held the description by
  then (created 16:59:30Z).

**What I did NOT establish, and am handing on rather than guessing:** which producer writes
it. The page is stamped `built_from_plan_version`, is typed `blog-post`, and the discovery
check called it an *"Editorial feature page"* at 16:58:11Z. The live work items on that
site name `dartsonline-traffic-workstream` → `content-gap-planner` → `page-build-handler`,
so the editorial/traffic pipeline for that site is where to look next. **That is another
lane's pipeline and there is an active `news editorial` session; this pointer goes to
them rather than being guessed at here.**


---

## 10. 2026-08-21 — the producer is a PERSON, so 339 is two sub-classes with different remedies

The `news editorial` lane owned it: **the strings were hand-typed into seed migrations**
(`sql_for_agents/495` for dartsonline, `492` for robot-hands), which set
`pages.meta_description` directly in their `INSERT INTO pages`.

**My elimination chain in §9 was exhaustive over CODE PATHS, and the answer was not on
one.** That is worth stating rather than quietly correcting, because the obvious inference
from an exhaustive-but-empty chain — *"the editorial path must bind the brief into the
column"* — would have sent someone hunting a bug in code that does not have one.

### The finding this changes: a seed migration bypasses every producer-side control

`PublicMetaDescription` guards the two tool-creation call sites. A migration writing the
column directly passes **none** of them. So 339 is not one class:

| | **(1) generated** | **(2) hand-authored** |
|---|---|---|
| how the text arrives | a brief passed as a *candidate* to `PublicMetaDescription` | typed into an `INSERT INTO pages` in `sql_for_agents/*.sql` |
| example | the 9 tool pages | dartsonline `495`, robot-hands `492` |
| **remedy** | fix at the PRODUCER — §5 candidate 3, don't pass the brief as a candidate (taken by `webdesign_tool_rebuilds`, live in `v1.0.1321`) | **candidate 3 CANNOT APPLY — there is no composer in the path to fix.** The guard is the only control, and by §2's measurement it does not fire in the 200-320 band. The alternative is a check at seed time over `sql_for_agents/*.sql` for `meta_description` literals |

**§5's ranking was written for (1) and does not carry to (2).** That is the substantive
correction to this file.

### ⚠ THE MEASUREMENT THEY ASKED FOR CANNOT NOW BE MADE — the fix destroyed the corpus

They proposed a structural tell — *"a description that argues with an alternative reading
was written for someone deciding how to write the piece; a reader has no alternative
reading to be corrected out of"* — and asked, rightly, that it be measured against the 693
before anyone believed it.

`[MEASURED 2026-08-21]` over 704 live descriptions, using
`\y(not|rather than|as opposed to|instead of)\y` as the lexical proxy:

| | |
|---|---|
| tell fires | **34** |
| of those, in the 200-320 brief band | **0** |
| of those, short (<200) and legitimate | **34** |

**As a lexical signal it is poor.** It fires on ordinary copy where "not" does ordinary
work — *"AI systems built around your business and your data, not l…"*, *"A loan from an
FCA-unauthorised lender is not legally enforceable…"*. Those are good descriptions and a
guard built on this would refuse them.

**But the test is also not a fair one, and that is the more useful finding.** The two known
positives were FIXED before the measurement, so the population no longer contains a single
true positive. **A signal cannot be fitted against a corpus its own remediation has
emptied.** Same shape as *a repro is destroyed by the render*: the instances were the
evidence.

Their *structural* claim may still be right — "not four about discipline" argues with an
alternative reading, "not legally enforceable" does not — but that difference is **semantic,
not lexical**, and the regex proxy cannot see it. Anyone re-attempting this needs the
ORIGINAL strings, which now survive only in that lane's backup table and in the quotes in
§3 and §9 of this file.

### Their two instances: FIXED, and staying listed on purpose

Both are remediated and verified at the served `<head>` (`sql_for_agents/497`, with a
verify block refusing >160 chars or a brief-shaped construction):
- dartsonline, 153 chars: *"Top players are skipping tournaments and the PDC has noticed. But the calendar grew: 30 Players Championship events a season through 2024, 34 since 2025."*
- robot-hands, 157 chars: *"Record robot installations, rising orders, one headline calling US demand weak. Five years of IFR figures show a step up in 2021, then a plateau at altitude."*

**They stay listed here as remediated instances rather than being struck out**, for a
reason better than bookkeeping: they are the only known corpus for fitting a detector for
sub-class (2), and the section above is why that matters. **The CLASS remains open — it has
no control.**

⚠ `leopardessconsulting.co.uk/hierarchical-multi-agent-orchestration-explained` is a THIRD
lane's and is **not** handled by any of this. Untraced.


---

## 11. 2026-08-21 — the tell is REFUTED by its own author's corpus, and sub-class (2) may have no cheap control

The `news_editorial` lane looked at their two examples **as a corpus rather than as two
instances of one error** and withdrew their own proposal. Verified here mechanically rather
than taken on report:

| original | contrastive marker? | tell fires |
|---|---|---|
| dartsonline — *"…these are one story about schedule density, **not** four about discipline."* | yes | **YES** |
| robot-hands — *"An editorial feature reading this week's robot-demand coverage across several channels, charted against the IFR's own five-year installation series - a step change that has held at altitude, and what that plateau means for end-of-arm tooling."* | **none** | **NO** |

**Their rule catches one of their own two examples, and misses the harder half** — the one
with no lexical marker at all. §10 recorded the tell as "poor as a lexical proxy" on the
false-positive count; this is stronger and comes from the positives.

### What the pair actually share is AUDIENCE, which is worse news for a guard

Neither describes the subject to a reader. The dartsonline one argues the framing; the
robot-hands one names its own artefact (*"An editorial feature reading…"*) and describes
how it was assembled. Both address **a person making editorial decisions**, not a person
deciding whether to click.

That is semantic, and — as its author put it — a *worse* basis for a guard than the first
suggestion because it is **less** mechanisable, not more.

### So the honest position for sub-class (2), stated plainly rather than hidden behind a cheap detector

**There may be no lexical or structural guard available for hand-authored descriptions.**
The available controls are both heavier than a regex:

- **(a) a seed-time check over `sql_for_agents/*.sql`** for `meta_description` literals.
  ⚠ It catches **the file, not the sentence** — it can tell you a migration writes the
  column, not whether the string is addressed to a reader. A human still judges.
- **(b) an LLM-side check that reads for audience.** The only thing that can see the
  property the two examples actually share.

**339 should say that rather than carry a detector half its known corpus escapes.** Adopted
as this file's position.

⚠ **And TWO examples is far too few to fit anything to.** Nobody should re-derive a signal
from this corpus. The way to get more is unavoidable and worth stating as a standing
instruction: **when another lane surfaces one of these, capture the string BEFORE
remediating it.** §10 records why — the previous fix emptied the population of every true
positive before the signal could be measured against it.

### The corpus is safe, and the risk that prompted the move does not exist

Both originals are committed verbatim in
`docs024_key_docs_latest/news_editorial_features/NOTES_news_editorial_features.md`
(tracked, `cd4398713`), which is the durable copy.

The concern that prompted that — that `database-cleanup` might sweep a `*_backup_*` table —
was reasonable and **checked here: it does not.** `[MEASURED 2026-08-21]` its `pre_query`
issues `DELETE FROM` against five *named* tables (`agent_error_log`,
`orchestration_state_audit`, `orchestration_states`, `palettes`, `typography_sets`) — **no
`DROP`, no wildcard, nothing matching `pages_backup_*`** — and `grep -rn "DROP TABLE"` over
`platform/ internal/ cmd/` returns no Go path at all. Corroborating: `pages_backup_20260717_r6`
has survived since 17 July. So `pages_backup_20260821_meta_desc` is not at risk from that
job. The version-controlled copy is still the better home.
