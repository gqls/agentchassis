# 399 — the framework computes each CTA's real destination, stores its title BESIDE the copy that misdescribes it, hands it to the writer, and never compares the two

Filed 2026-08-25 by the `dartsonline_traffic` lane, from an owner report of four separate
defects on dartsonline.com. He asked for the general fix rather than four individual ones, and
there is one: **the fact needed to prevent this class is already computed, already stored, and
already adjacent to the wrong value — on 646 live components as of 2026-08-25 — and nothing
compares them.**

## The evidence, in one row

`news-index`, `hero` slot, `page_components.content_data` `[MEASURED 2026-08-25]`:

```json
{
  "cta_text":                   "Catch up on this week's darts news",
  "cta_url":                    "/brands/index.html",
  "cta_target_title":           "All Brands | Darts Online",
  "secondary_cta":              "See what's new for beginners",
  "secondary_cta_url":          "/tools/setup-builder/index.html",
  "secondary_cta_target_title": "Dart Setup Builder | Tools"
}
```

The copy says *news*; the recorded destination title says *All Brands*. The copy says
*beginners*; the recorded destination says *Dart Setup Builder*. **Both contradictions are
inside one jsonb object, in adjacent keys, at build time** — no rendering, no HTTP, no HTML
parse, no discovery sweep required to see them.

And the field exists for exactly this purpose. `setCTAField`
(`resolve_internal_links_action.go:335-339`), verbatim:

> *"Alongside the url it writes the target's title under `<field-minus-_url>_target_title`
> (`cta_url` → `cta_target_title`) **so the content writer can write CTA copy FOR the actual
> destination instead of guessing one.**"*

So the design intent is destination-first copy. **What is missing is any check that the writer
complied.** `grep`ped 2026-08-25: `*_target_title` has writers
(`resolve_internal_links_action.go`, `links_tel.go`) and **no reader that compares it to the
label** — the only two references in the checks are `SuggestedTargetTitle` fields the checks
*emit* in their own findings.

## Where enforcement actually happens, and why it loses

Instead of comparing two adjacent strings at write time, the estate detects this by:
render the page → deploy it → sweep it later → parse `rendered_html` → extract anchors →
reduce the text to distinctive tokens → token-match against every page's name/title/nav_label →
file a `page_rerender` → recompute the url. Five lossy stages to rediscover a fact that was free.

That pipeline then loses in five measured ways:

1. **It is enabled on ONE of three discovery agents** `[MEASURED 2026-08-25]`:
   `completeness-discovery-agent` carries `misdirected_cta`; `quality-discovery-agent` and
   `design-discovery-agent` carry neither it nor `empty_sections`.
2. **The defect is minted faster than the sweep runs.** `news-index` re-renders with
   `reason=section_data_resolved` several times a day ("fresh news items available" — 6+ times in
   the last 3 days), and each regeneration re-mints CTA copy. A periodic sweep cannot converge on
   a page that re-authors itself between sweeps.
3. **Generic labels are documented to fall through to an arbitrary pick.** Same file, line 349:
   *"A generic label ('Get Started') or one matching no candidate falls through to today's
   positional behaviour unchanged."* And `check_misdirected_cta`'s own header: generic texts
   *"match nothing and are skipped entirely."* So a generic CTA gets a positional destination and
   is then unfalsifiable — **correct by construction, wrong in fact.**
4. **Site chrome is outside the surface entirely.** `check_misdirected_cta` reads
   `page_components.rendered_html`. The nav button lives in `site_components`, so
   `Get Started → /contact.html` can never be seen by it. Confirmed at the artefact:
   `<a href="/contact.html" class="header-cta">Get Started</a>`, and the header's `content_data`
   carries no cta/url/target_title keys at all — the nav CTA is not part of the CTA subsystem.
5. **The one arm that does fire has no handler.** `cta_names_unknown_destination` goes to
   `needs_human_review` *"with no handler"* by design. Live example on this site: *CTA "Read the
   shaft length guide" on shaft-length (hero): links back to its own page*. Filed 2026-08-24,
   parked, nobody told.

## The same session found the identical shape elsewhere, which is why this is a class

- **`empty_section` on `brands-index`** — the brands directory the owner reported as absent
  (450 chars of visible text, zero brands listed). Detected **2026-08-17, 08-22 and 08-24**, and
  **`failed` all three times**: *"completion blocked: verification could not run, and this item
  type fails closed (RFC_017)"*. Terminal, silent, eight days.
- **`bugs_open/384`** — a card image lands, is linked correctly, and the listing that renders it
  is re-rendered three times in a mode that structurally cannot pick it up.
- **`image_url_404`** — `/assets/images/hero.jpg` referenced by 5 pages, filed 2026-08-09 with
  **no `handler_agent`**, still `detected` 15 days later.

**The unifying property is not "checks are missing". It is that a defect can reach four different
terminal-looking states — `failed`, `needs_human_review` with no handler, `detected` that never
drains, and `complete` on an item whose defect persists — and every one of them looks handled
from outside.** Nothing reports "this site has N unresolved defects".

## Fix candidates, ordered by what closes the door

1. **Compare the label to its `_target_title` at WRITE time.** After the content writer returns,
   for every `<x>` with a sibling `<x>_target_title`: if the label's distinctive tokens match no
   token of the recorded title, either regenerate the label from the title (the field exists for
   that) or refuse the write. **This makes the bad state unrepresentable rather than detectable**,
   costs one string comparison, needs no rendering, no sweep and no per-agent enablement, and
   applies to all 646 components carrying the field today. The token reducer already exists in
   `check_misdirected_cta` and would be shared.
2. **Exclude the current page from the destination picker.** Kills the self-advertising CTA
   ("Catch up on this week's darts news" *on the news page*) as a category, and stops the
   `cta_names_unknown_destination` self-link arm being reachable at all.
3. **Refuse to mint a positional destination for a generic label.** Either the writer names a
   destination or the CTA is not emitted. Today "Get Started" pointing anywhere is *correct*
   behaviour, which is precisely why nothing flags it. **Fleet control measured: 0 of 1,515 live
   CTA labels are generic**, so this arm is rare in page components and cheap to tighten — the
   nav is the real consumer and it is outside the subsystem (candidate 4).
4. **Bring site chrome into the CTA subsystem**, or state in the check's header that chrome is out
   of scope so the next reader does not assume coverage it does not have.
5. **Report the terminal-but-unfixed states per site.** A count of `failed` + `needs_human_review`
   with no handler + `detected` older than N days is the missing instrument. Every defect in this
   file was *detected*; none was *surfaced*.

Candidate 1 is the one that closes the door; 5 is the one that would have told us eight days ago.

## Verification

Pick a page with a mismatched pair, run the build, assert the stored `cta_text` shares a
distinctive token with `cta_target_title`. **The disconfirming case must be exercised**: seed a
deliberately mismatched label and require the build to refuse or rewrite it — a check that only
ever sees matching pairs proves nothing, which is this estate's most-repeated lesson.

## 090 substitution, stated

Not run through the diagnosis loop: every step is first-hand and re-runnable — the live
`content_data` row above, the two source comments quoted verbatim, the `grep` showing no
comparing reader, the three `agent_definitions` rows showing enablement, the `site_work_items`
history for `empty_section`, and the served header anchor. A 090 run would re-read the same six
artefacts.

---

## CONTRIB 2026-08-26 — two independent precedents for candidate 1, from the `news_editorial` lane

Candidate 1 above ("make the bad state unrepresentable at write time rather than detectable at
sweep time") is not a novel argument in this estate, and two live pieces of work rest on it. Both
offered by the lane building `features_open/035`:

1. **035's G1 test is candidate 1's shape exactly.** Its acceptance criterion for decomposed
   generation is *"rewriting one prose child leaves every sibling row byte-identical — **by
   construction (row-scoped write), not by prompt discipline**"*. Same move: put the guarantee in
   the write path's shape rather than in an instruction the writer may ignore and a checker may
   later notice.

2. **035 §6.1 REFUSES a CASCADE on the parent FK for the same reason, arguing it in reverse.**
   A cascade would let a delete quietly take children with it; the loud foreign-key error is the
   tripwire, and cascading *"would hand the sweep the silent-destruction power the guard exists to
   deny"*. That is candidate 1 stated as a prohibition rather than a construction — and it is the
   more useful half for this bug, because the CTA case is *already* the failure that refusal
   avoids: a value is written that cannot be right, and the only thing standing behind it is a
   sweep that may or may not run.

**Why this belongs in the file rather than in a message:** candidate 1 will read as a new
proposal to whoever picks this up, and a new proposal costs a design argument. It is not new — it
is the principle two other pieces of live work already chose, and the honest framing is *"apply
the rule this estate already uses to the one place it demonstrably is not being applied"*.

---

# ADDENDUM 2026-08-26 — taken on, built, committed. Two of this file's own claims are corrected.

Picked up by the `bugfix_399_cta_label_agreement` lane (`docs/agent_docs/docs024_key_docs_latest/bugfix_399_cta_label_agreement/`).
Committed `08afad7cd` (+ revisions); council **APPROVED at round 3**, corr
`e9bda035-5ad7-4a27-8d4f-613bd03abe05`, 12 of 15 seats — `architecture` among them, so this is a
**ruled** point fix rather than an assumed one. Rounds 1–2 found three real defects and all were
fixed before approval: an ordering constraint shipped as prose rather than a `_HOLD` filename; a
coverage claim that was two writers of three (see §4's correction); and a false mutation claim in
the very test written to prove the pass cannot fail a save — which, once fixed, exposed a recover
handler that re-raised through the same nil logger and contained nothing.
Register **LNK-040**. **Inert until the next fleet roll AND until migration `643` applies.**

## 0. Still valid, and two of the figures above had already gone stale

- The evidence row is unchanged **and was re-minted at 2026-08-25 20:58:09Z — 17 minutes after this
  file was filed**, still carrying the contradiction. The defect re-mints faster than any sweep.
- **646 → 665 components** carrying a `_target_title` `[MEASURED 2026-08-26]`.
- §"Where enforcement actually happens" point 1 says **1 of 3** discovery agents carries
  `misdirected_cta`. There are **5** live discovery agents; still only `completeness-discovery-agent`
  `[MEASURED 2026-08-26]`. Both figures were dated here, which is the only reason the staleness was
  mechanically detectable.

## 1. THE MEASUREMENT THIS FILE COULD NOT MAKE — and it decides the whole fix

§"Fix candidates" assumes the writer needs a comparison because it is not complying. That was the
right guess for a reason this file did not have: **the writer is already told, twice, and complies
six times in seven.**

- `content_components.input_schema` → `cta_text.llm_guidance` (live row read 2026-08-26): *"the link
  destination is already fixed: write this CTA text FOR that destination — name it or clearly
  promise it. Never write copy promising a page the URL does not point to."*
- At runtime, migrations `476` (2026-08-19) + `477` (2026-08-20 07:17Z, the `bugs_open/312` seam)
  put the resolved title in the prompt: *"Destination (fixed): &lt;title&gt;. … never promise a
  different one."* **781 of 2,297** `page-content-writer` prompts over 3 days carry that literal.
- **Of the pairs written SINCE that pipe went live, 155 of 1,060 (14.6%) still contradict their
  destination** `[MEASURED 2026-08-26]`.

**That number could have come out near zero and did not. Prompt text is not a control**, and this is
the disconfirmable evidence that justifies a check rather than a firmer instruction.

⚠ Do **not** quote the before/after split (23.0% → 14.6%): the "before" bucket is only rows not
rewritten since 2026-08-20 (n=135), i.e. survivorship-biased. The post-pipe figure stands alone.

## 2. ⚠ CORRECTION — fix candidate 1 is wrong in BOTH halves, and the remedy half is HARMFUL

Candidate 1 says: compare the label's tokens to `_target_title`, and on mismatch *"either regenerate
the label from the title … or refuse the write"*.

**(a) The comparison would be a THIRD predicate.** `BestLabelMatchForPage` already answers "which
page does this copy name?" and is what `check_misdirected_cta` uses. A label↔title token test is a
different definition of "misdirected" beside the detector's and the writers' — the re-drift
**RFC_047 §9** rejects by name (*"the exact thing bugs_open/203's extraction and bugs_open/308 both
exist to stop"*). It is also the shape `bugfix_203/CALIBRATION_2026-08-11` measured brittle: all
nine already-correct CTA labels on gaswholesalers.com flipped to the wrong tool over a stray hyphen
in "Break-Even".

**(b) Regenerating the label from the title makes things worse.** It is exactly what
`stampCTADestinationGuidance` already asks the writer to do. Performing it by force converts a
mismatch into a **LOCK** — moving the row out of the ~60 label-less bucket that `bugs_open/391`'s
ranking fix can reach, into the ~20 label-locked bucket only an LLM copy pass can clear. At ~150
mismatches a week that is a large, silent transfer from cheap-to-fix into expensive-to-fix.

**(c) And the comparison cannot see `bugs_open/391`'s class at all.** When the framework chose the
destination *and* told the writer to name it, copy and title **agree** and the button is still
wrong — 16 of 17 of that lane's password-entropy fields, *including all three buttons the owner
reported*. Agreement between two framework-written strings is consistency, not correctness. This
limit is now pinned by `TestJudgeCTALabelIsBlindToTheLabelLockedDefect`, a test that **passes and is
wrong on purpose** so no later reader claims coverage this lacks.

## 3. ⚠ REPAIR IS UNREACHABLE — the number that turned this into a recording change

Of the 186 mismatched pairs live `[MEASURED 2026-08-26]`:

| the copy names | count |
|---|---|
| exactly one other page — a confident repoint is possible | **13** |
| two or more pages (RFC_047: refuse, never guess) | **78** |
| no page on the site at all | **95** |

**An automatic repoint reaches 13 of 186 (7%)** and inherits `bugs_open/248`'s clobber — a
`misdirected_cta` repair turned a *correct* `/contact.html` into a wrong link on 2026-08-24. A
refusal arm was rejected too: at 14.6% it fails ~1 CTA write in 7 (~29 sections/day at the
2026-08-24/25 rate), nothing can auto-satisfy it, and on a page that re-authors itself daily one
cosmetic mismatch becomes an indefinitely withheld refresh.

**173 of 186 need a human or the commissioned copy pass — and nothing told anyone they existed.**
That inverts the reading of candidate 1: recording is not a cautious first step here, it is very
nearly the whole available action, and **the record is the deliverable**. It is also precisely what
the two lanes owning the remedy are missing (`bugfix_389_cta_relevance` needs to know which rows its
ranking fix cannot reach; `cta_target_content_pass` has never run for want of a worklist).

## 4. ⚠ THE SEAM IS NOT WHERE THIS FILE IMPLIES — half the writers bypass it

Candidate 1 says "after the content writer returns", which points at `RenderComponentAction`. That
covers the **build** path only: `RerenderPageSectionsAction` never goes through it
(`rerender_page_sections_action.go:662` calls `RenderTemplate` directly), so a gate there is blind
to the **repair** loop — which is the loop actually minting the churn (**182** `misdirected_cta`
item_keys have been filed more than once `[MEASURED 2026-08-26]`).

Both writers converge on `save_page_sections`, verified in live `agent_definitions` 2026-08-26:
`page-build-handler → call_content_writer → save_sections` and
`page-rerender → rerender_sections → save_sections`. That is where the pass went.

> **⚠ CORRECTED 2026-08-26, same day, by the council gate's `bug_historian` seat (corr `e9bda035`):
> "both writers converge" is TWO OF THREE, not all of them.** The seat asked whether another path
> persists CTA `content_data` outside the censused save steps. It does: `ApplySectionEditAction`
> (`section_editor_actions.go` — `updatePageComponent`, `updatePageComponentSwap`) writes
> `page_components.content_data` directly and never passes through `SavePageSectionsAction`.
> **It is LIVE, not dormant** — 144 `section_edit` work items, 132 complete, newest **2026-08-26**
> `[MEASURED 2026-08-26]`. Its CTA exposure is **3 of those 144** naming a CTA field, so the limit is
> **stated** rather than closed by widening a third seam while the first two are unproven in
> production. The original claim holds for the BUILD and REPAIR pipelines — the two that regenerate
> CTA copy — and is false of the estate's writer set as a whole. **This is the correction I would
> most want a later reader to have**, because the unqualified form reads as full coverage.

And it is **six** `save_page_sections` steps fleet-wide, not two — four sit inside a loop's
`sub_workflow`. For an instrument that distinction is load-bearing: a guard armed on half its
writers is visibly partial, an instrument armed on half its writers reports a **rate** that reads
fleet-wide and is silently biased. Migration `643` asserts the census and aborts if it moves.

## 5. ⚠ THIS FILE'S TITLE IS ALREADY A CLOSED BUG — 399 is its REOPENING, not a duplicate

`bugs_closed/023_HANDOFF_2026-07-19_cta_label_url_pairing_unchecked.md` is titled *"A button's label
and its destination are never checked against each other"* and its root cause reads *"nothing
anywhere expresses 'a label implies a destination' as a constraint"*. It closed **2026-07-25 without
ever building the comparison** — it made *absent* destinations unrenderable instead (criterion 3 is
about an **absent** destination, not a **wrong** one).

Saying so here matters: without it, a future session greps the mechanism, finds 023 closed, and
closes this as a duplicate of a bug that never addressed it. The transferable lesson, from the
filing session: **grep the TITLE phrasing as well as the mechanism — the previous filer described it
as a user would and you are describing it as the code does.**

023's closure also carries the warning that decided the record's destination: *"More detection makes
the invisible pile bigger."* Hence `agent_error_log` (`CTA_LABEL_MISMATCH`, disposition
`instrumented`) and **not** `site_work_items` — 78 `cta_names_unknown_destination` + 70
`unresolved_cta` already sit at `needs_human_review` with no handler, which is this file's own §5.

## 6. What shipped, and what is still owed

**Shipped** (`08afad7cd`, inert until the roll + `643`): `datahelpers.JudgeCTALabel` — the
detector's own question, extracted; `ctaClassifyAnchor` reduced to a thin adaptor over it with its
existing tests **unchanged** as the proof; `actions/cta_label_audit.go` asking it once more before
persist; migration `643`; register **LNK-040**. Six mutations run, six killed.

**Candidates 2, 3, 4 not built, deliberately.** Candidate 2 (exclude the current page) already
shipped as `BestLabelMatchForPage`'s self-link refusal (`bugs_open/308`). Candidate 3 is real but
tiny in page components — **5** fully generic labels fleet-wide `[MEASURED 2026-08-26]`, and this
file is right that the nav is the real consumer, which is candidate 4. Candidate 4 (site chrome) is
a separate subsystem — `site_components` carries no cta/target_title keys and
`render_site_components_action.go:203` hardcodes `"Get Started"`; it needs its own bug, and the
owner has asked that it not be hand-fixed on his site.

**Candidate 5 remains OPEN and unowned, and it is the honest weak point of what shipped.** A record
nobody reads is not a fix. The reading obligation and its monthly query are in the lane RUNBOOK and
LNK-040, naming `bugs_open/410` (filed 2026-08-26, *"seams that complete green and land somewhere
nobody reads"*) as the class this inherits. **The reading is the RATE, not the row** — nobody should
read 155 records; somebody must notice if 14.6% becomes 30%, or falls to 2% after a prompt change.

**Owed after the roll:** confirm `CTA_LABEL_MISMATCH` rows arrive from **at least two distinct
`agent_type`s**. ⚠ And note the migration is now **`643_..._HOLD.sql` + `645_..._HOLD.sql`**, staged:
643 arms the two primary writers as a canary, 645 arms the remaining four. **The mismatch rate must
not be read between them** — an instrument armed on half its writers reports a rate that reads
fleet-wide and is silently biased, which is the very argument that made this a six-step census. One producer means the six-step coverage claim is failing silently — which is the
exact failure `643`'s census exists to prevent. Then re-measure the rate **at this seam**; it will
differ from the 14.6% token census (different predicate), and the new number is what any later
decision rests on. ⚠ A pre-roll zero means "binary not rolled", never "no mismatches".
