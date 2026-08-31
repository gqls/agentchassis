# NOTES — CTA destination relevance (`bugs_open/389`). Append-only, newest at the bottom.

## 2026-08-25 — phase 0: cause found, and two of my own claims corrected

**Owner report:** `/tools/password-entropy.html` offered as the CTA on an AI-orchestration
consultancy — *"not deliberate and actually wrong"*. Raised out of yesterday's 277 session, where
I had measured the URL across three domains and deliberately **not** filed, because "that seems
off to me" is taste, not evidence. The owner's confirmation is what made it a defect.

### The path that found it (recorded because the wrong turns were the informative part)
1. Grepped for a hardcoded `password-entropy` — found none in live code, but found
   `005_content_components.sql:8942`, *"Narrow password-entropy tool affinity"*, explaining the
   tool was pushed to four sites *"because the library only had 2 tools with templates"*.
   **This is history and I nearly stopped here.** It explains why the tool EXISTS on those sites;
   it does not explain why CTAs POINT at it. Two different questions.
2. Assumed semantic/tag matching next — **wrong**. `chooseCTATargets` has no semantic input at
   all. Reading it took two minutes and killed the hypothesis.
3. The real answer: `nav_order` ascending, `name` as tiebreak, take `[0]`. Simulated it against
   live `pages`; the predicted winner matched the stored value on all 3 sites.

### ⚠ MISSTEP 1 — I nearly filed a 13-site claim that is false
Measured 13 sites whose rank-1 CTA target has `in_header=false` and drafted it as "13 deliberate
contradictions: a human hid it, the system ignored them". **The disconfirming check I had not run:
what fraction of tool pages are `in_header=false` at all? 143 of 228 — 62.7%, the majority state.**
So the flag does not mean a human judged anything. Only leopardess is a real case, and only
because `L5_nav_and_ctas.sql:29` carries the comment *"a password tool doesn't belong in the
primary nav"*. **The check that would have caught it is one `GROUP BY` and it costs nothing** —
before reading any flag as intent, measure its base rate. Now in the RUNBOOK.

### ⚠ MISSTEP 2 — the served-bytes probe that found nothing
Curled the **home pages** of finetuning and leopardess for `password-entropy`: **0 refs**, which
for a moment read as "not actually live". It was the wrong page: the stored fields sit on
`/services.html`, `/technical-details.html`, blog posts. **Probe the page that holds the field, not
the page you assume.** The home page of the third site *did* carry it — four references, minted
today.

### The finding that makes it urgent rather than cosmetic
The `__cta_minted` stamp (LNK-035, live 2026-08-22) splits the 80 fields into 17 resolver-minted
(dated **08-23 → 08-25, i.e. today**), 24 stamped-but-superseded, 39 unstamped. ⚠ **NULL is "not
recorded", not "authored"** — there is no backfill by design, so anything older than 08-22 is
unattributable. Reading NULL as authored would have made this look historical and closed.

### The structural point, which is the part worth carrying
`pages.nav_order` serves two unrelated readers — the nav menu and the CTA chooser — and nothing at
either site says so. `in_header` is read by one and not the other. That is why a human's explicit
"don't make this prominent" was a no-op: the two mechanisms disagree about which column carries
the intent, and there is no column that carries it for CTAs at all.

### ⚠ MISSTEP 3 — I filed before finding the prior lane, and the prior lane had the population 10 days earlier
`cta_target_content_pass` measured this on **2026-08-15**: 16 sites with ≥6 rows on one modal
target, finetuning 39, ai-agent-orchestration 36, password-entropy modal on three sites and
described in that plan as *"topically absurd"*. The owner **accepted it as a floor and commissioned
a content pass**; nothing was run.

**Why I missed it:** I grepped `bugs_open/` and `bugs_closed/` for the mechanism — which is what
"grep before you file" literally prescribes — and the prior art is **not a bug**. It is a lane,
named in one line of `MEMORY_workstreams.md` (line 88). **The cheap check: grep the workstreams
index too, not only the bug directories.** A commissioned-but-unrun deliverable lives in a lane
doc by definition, because it is not a defect.

**What it changed:** not the finding (the `nav_order = 1` fossil and the live minting are new and
stand) but the **recommendation**. The commissioned pass is an LLM rewrite over 16 sites; the
fossil integer means the three worst sites may need one `UPDATE` each instead. Ordering now leads
the write-up: fix the ranking input, re-measure, then size the content pass against what remains.
**A root cause found under a commissioned workaround should change the workaround's scope before
anyone runs it** — that is the transferable point, and it is why filing without the lane search
was worth correcting the same day rather than leaving as a footnote.

### 2026-08-25 later — an adversarial review of my own file, and what it caught
Two independent reviews (mechanism; docs+decisions). **Every load-bearing claim reproduced** —
the sort keys, the three-site `nav_order=1`, `in_header` absent from the CTA path, the 17/24/39
split, the served bytes, and the 62.7% retraction. What they found is below; the first is the one
that matters.

**⚠ MISSTEP 4 — I missed the loop my own bug sits inside.** `setCTAField` tries
`BestLabelMatchForPage` FIRST and the positional pick LAST, and `stampCTADestinationGuidance`
(`:362`) feeds the chosen destination's title into the writer's spec for the **label** field. So
the framework writes copy naming whatever it picked, and the next resolve label-matches that copy
back to the same page. **A ranking accident becomes a content fact.** Measured: 17/17 minted fields
carry a `*_target_title` naming the tool and 16/17 have copy naming it; 20 of 80 overall are
label-locked, **including all three buttons the owner reported**. My "the three worst sites may not
need the content pass" was wrong in the worst direction — the pass is exactly what those buttons
need. **The check I skipped: before claiming a fix reaches a population, read the code path that
runs BEFORE the one you fixed.** I read `chooseCTATargets` thoroughly and never read its caller.

**⚠ MISSTEP 5 — I mischaracterised the provenance middle bucket.** I wrote that the 24 carry "a
stamp naming a different url, so the value reads authored". **Zero** rows do; the 24 have no stamp
entry for that field at all (it covers a sibling slot), and "authored" is wrong in the code's terms
(`storedCTADestinationIsAuthored` is true only for utility-area urls). I invented a semantic from
the *shape* of a three-valued result instead of reading what produced each value — the
[[a-report-is-not-a-measurement]] shape: a key's SHAPE is a hypothesis about provenance.

**⚠ MISSTEP 6 — "minted today" overstates the instrument.** The stamp is value-bound with no
timestamp; the dates are the row's `updated_at`, and a `SeedCTAMinted` carry-forward looks
identical to a fresh mint. The liveness claim survives on other evidence (the ranking simulation,
and one positional mint whose copy — "Book a Technical Discovery Call" — cannot have label-matched)
but I quoted the weaker instrument as though it were the stronger one.

**⚠ MISSTEP 7 — my own RUNBOOK said "mirror the code exactly" and my query did not.** It omitted
`PageMayBeLinkedPredicateFor`. Harmless for the three sites; the 26-site blast-radius review rests
on it and should be re-run.

**⚠ MISSTEP 8 — the correction never reached the owner's document.** README and PLAN were written
at 10:51 and never touched again while the correction propagated to the bug file, the handoff, both
CONTRIBs and the workstreams index. **The owner's own log kept giving the pre-correction
recommendation, and never mentioned that this reverses his own 08-15 decision.** CLAUDE.md's
cadence rule names exactly this ("the moment a decision, correction or resizing lands") and I
followed it everywhere except the one document written for him. **Propagation is not done when the
bug file is right.**

**Smaller:** `links.go:328` not `:333`; the "only option that stops the class" claim was overstated
(an opt-out is reactive — it makes the good state sayable, not the bad state unrepresentable);
RFC_022's narrowing was never engaged though I cited the ruling it qualifies; the fossil claim is
[INFERRED] from `created_at`, and `L5_nav_and_ctas.sql:36-45` shows someone renumbered that site's
nav 2–10 and left the tool at 1 — which sharpens the irony rather than weakening the point;
`chooseCTATargets` carries an unused `pageType` "for a future intent-aware (LLM) upgrade", i.e. the
hook for the relevance option already exists.

### 2026-08-25 evening — owner answered all five; decision 2 applied; retirement step 2 CANARY dispatched

**Owner:** tool "can disappear everywhere" (1), yes to the numbers (2), yes to the platform lever
(3), "whatever you suggest" (4), re-scope the commission (5). Follow-up: **the library component
STAYS** — `tool-password-entropy` remains `is_active=true` and available to new sites; retirement
covers the three site pages only.

**Decision 2 applied** — `SQL_2026-08-25_demote_password_entropy_nav_order.sql`, `nav_order` 1 → 900
on three rows, guarded. ⚠ **The value matters and 200 would have failed**: 200 is those sites'
ordinary tool value, so it ties, and the tiebreak is alphabetical on `name` — `password-entropy`
precedes every `tool-*`. At 200 it would still have won on two of three sites. *A demotion that
joins the pack is not a demotion.* New rank-1 verified on each site and all three are on-topic.

**Retirement deliberately NOT run first, despite being decision 1 and fully authorised.** Measured
blast radius: **91** `page_components` references (content_data AND rendered_html; 45/25/21), 1
footer, 3 live `tools.html` listings, 0 visible nav. Deleting first strands those and leaves the 20
label-locked buttons naming a tool they no longer point at — `bugs_closed/299`'s defect,
manufactured by our own repair. **Authorisation is not a sequence.**

**Step 2 canary dispatched:** item `b422751a-3745-474c-87d6-aeff50028546`,
`finetuning.uk/technical-details` (both its buttons are label-locked).

Three preconditions checked BEFORE writing it, each of which could have sunk it:
1. **All 12 target pages are `rebuild_policy='generic'`** — so `page-build-handler`'s owned-page
   guard does not fire. Had any been `owned`, the item could only ever have been refused
   (`bugs_open/333`).
2. **`spec.suggestion`, NOT `spec.content_guidance`.** `suggestion` is the key the handler reads;
   `content_guidance` is only *aliased* into it (`bugs_open/271`,
   `load_work_items_guidance_alias_test.go`), and an author-supplied `suggestion` wins over the
   alias. Writing the read key removes any dependency on the alias having shipped. The lane RUNBOOK
   I inherited says `content_guidance`; a live completed item says `suggestion`. **I followed the
   live row, not the doc.**
3. **The queue is empty** (1 `triaged` item fleet-wide), so the 268 lane's "dispatch serves the
   fleet's OLDEST eligible item" gotcha does not bite and **no backdating was needed** — no
   synthetic timestamps in this lane.

**The framework writes the copy, not me** (owner rule 2026-08-06): the guidance supplies the site's
eight real tools with their URLs and the constraint (labels only, name a tool, never mention
passwords); the writer chooses which tool fits the page and words it.

**Verify as a matched pair, not by status** — the label must change AND the href must follow it, and
`bugs_open/389` proves a `cta_links_stale` rerender reports `complete` either way.

### 2026-08-25 — CANARY VERIFIED at the served bytes. And ⚠ MISSTEP 9: I declared a stall that was not one, and fired a duplicate write at a live page

**The canary worked, end to end, through the ordinary queue** [MEASURED 2026-08-25, served bytes]:

| check | before | after |
|---|---|---|
| `password-entropy` refs on the page | 2 | **0** |
| hero button | "Explore Password Strength Physics" → `/tools/password-entropy.html` | "Try the Fine-Tuning vs RAG vs Prompting Decision Guide" → `/tools/model-approach-selector.html` |
| CTA button | "Test a password with Password Strength Physics" → same | "Work out which approach fits your business with the …Decision Guide" → `/tools/model-approach-selector.html` |
| **prose control** `<p>` | 15 | **15** (labels-only held) |
| target URL | — | **200** |
| bytes | 37,789 | 37,869 |

Label and href moved **together** — the matched pair, not a status read. The writer's tool choice
(`model-approach-selector`, "Fine-Tuning vs RAG vs Prompting Decision Guide") is apt for a
technical-details page on a fine-tuning site; the framework chose better than a hand-written label.

### ⚠ MISSTEP 9 — the stall I diagnosed did not exist, and I intervened on a live page because of it

I watched the `page_rerender` item sit at `triaged`, told the owner *"the per-site orchestrator
simply didn't come round"*, and fired a **direct** `page-rerender` to bypass the queue. The
timeline says otherwise:

| time | what |
|---|---|
| 13:07:42 | item created |
| 13:19:49 | → `triaged` |
| **13:37:22** | **the queue's own run starts** (corr `ca88f642`, 3 orchestrations) |
| 13:37:27 | **my direct fire starts** (corr `a20aa7a8`) — **5 seconds too late to be the cause** |
| 13:37:54 | CTA urls written |

**The queue fixed it. My intervention was a redundant duplicate**, running concurrently against the
same page. It was harmless only because a CTA recompute is idempotent — both runs resolve to the
same destination. Had it not been, I would have raced the platform against itself on a live
customer page.

**The root of the error was mundane and worth naming: I mis-estimated the wall clock.** I believed
~50 minutes had passed at `triaged` when the true figure was ~17 minutes (and ~30 from creation) —
entirely consistent with the **24 minutes** I had *already measured* for my own `content_rewrite`
item earlier the same afternoon. I had the baseline and did not apply it.

**The check that would have caught it, and it is one query:** before calling a queue stalled, ask
what its *service interval for this site* actually is, and compare like with like —
`SELECT created_at, updated_at FROM site_work_items WHERE site_id=… AND status='complete' ORDER BY
updated_at DESC LIMIT 10`. An absence over N minutes is not evidence of a permanent stall, and "it
has not run yet" and "it will never run" are different claims. The estate's direct-fire remedy is
for a **dead** queue (`bugs_closed/029`: items orphaned at `claimed`, zero completions); this queue
had **593 completions in six hours** — a figure I measured, quoted, and then argued past.

**Consequence for the remaining 11 pages: use the ordinary queue.** The recipe works through it
end-to-end. Budget ~25–35 minutes per item and do not bypass.

### 2026-08-25 ~19:2xZ — new chassis verified; the remaining 11 pairs dispatched with PLATFORM-ENFORCED ordering

**Chassis `a7459a44b68b8c67b7d7bb0ca7c064e0729d59f5`**, pods up 19:07Z. Capability re-probed on the
running binary **with its absent-control** (`rendered_html_transform` 8, `code_span_to_code_tag` 5,
`cta_links_stale` 3, control **0**) rather than inferred from ancestry — a commit after mine could
delete the code and still leave mine an ancestor. ⚠ My first probe loop **timed out before the
control ran**, which would have left me quoting three present-counts with no evidence the probe
discriminates; re-ran the control alone. *A capability probe without its control is not a probe.*

**Dispatched the remaining 11 label-locked pages** (`SQL_2026-08-25_step2_remaining_11_pairs.sql`):
11 `content_rewrite` + 11 `page_rerender`, each relink carrying
`depends_on = ARRAY[<its rewrite id>]`.

**The design decision worth recording, because the obvious alternative is wrong.** The canary taught
that between a rewrite and its relink the page serves a button whose text and href disagree
(`bugs_closed/299`'s shape) — ~32 minutes on the canary. Two waves (11 rewrites, then 11 relinks)
would put **all eleven pages in that state simultaneously** and make me the ordering mechanism.
Instead the platform enforces it: `load_work_item_actions.go:713` refuses a row whose `depends_on`
is not `complete`/`verified`, so each page's relink unblocks only when its own rewrite lands, and the
pages progress independently and unattended.

**Verified, not assumed** — I ran the dispatcher's own `depends_on` clause against my 22 rows at
dispatch time: **11 rewrites eligible, 0 relinks eligible.** That is the check that matters; had the
predicate not bitten, I would have created eleven simultaneous mismatches while believing I had
avoided exactly that.

Fresh handoff written: `HANDOFF_2026-08-25b_continue_here.md`; the 08-25 one marked SUPERSEDED at its
head, naming the two claims in it that were reversed.

### 2026-08-25 ~20:2x–20:35Z — MISSTEP 10: **my own repair destroyed authored copy on a live page**, and the control that caught it barely moved

The 22 in-flight items landed: **11 `content_rewrite` complete, 10 `page_rerender` complete**, one
pair failed (`/model-directory.html`, §MISSTEP 11 below). I then ran the §2 verification. The CTA
half is good on all ten repaired pages — label and href name the same tool, every target 200 with a
per-domain absent control at 404, and the only residual `password-entropy` strings are the footer
(2 per page, `ai-agent-orchestration.com`) and the legitimate `/tools.html` listing card. **The
content half was not.**

**`finetuning.uk/your-own-model.html`: two authored sections destroyed and replaced with copies of
a third.** Before the rewrite the page carried three distinct `generic-text-block` components:

| pos | opening, before | opening, after |
|---|---|---|
| 2 | *"How it works — Training a model on your own documents comes down to three steps…"* | *"How it works — You send us examples…"* |
| 3 | *"Three steps, and one overnight run — The process runs in three steps…"* | *"How it works — You send us examples…"* |
| 4 | *"How it works — You send us examples…"* | unchanged |

The page **served the same section three times** and two pieces of authored copy were gone.
Confirmed at the served bytes, not inferred: 3 × the surviving opening, **0 ×** each destroyed one.

**Named from its own row, not guessed.** `page_component_history.source_item_id` on the 19:43:54Z
generation is `10b8b6d2-660c-4696-ae6a-ca20c8823dcf` — *this lane's own `content_rewrite`*,
commissioned to reword CTA **labels only**. The `page_rerender` at 19:46:19Z archived the
already-damaged state, so it is the rewrite, not the relink.

> **⚠ The paragraph control nearly missed it.** `<p>` went **17 → 20** on this page: a +3 that reads
> like a writer adding a sentence. The damage was **duplication**, and duplication barely moves a
> count of anything — it moves *distinctness*. The check that actually found it compares
> **`count(DISTINCT left(text,80))` against `count(*)`** per page. Ten pages, one hit:
> `6 components → 4 distinct`. **A count-based control cannot see content loss that arrives as a
> copy.** Both controls now belong in the recipe, and the distinctness one is the load-bearing half.

**Repaired, and the repair is verified at the served bytes.**
`SQL_2026-08-25_restore_your_own_model_blocks.sql` writes the pre-rewrite `content_data` back from
`page_component_history` **verbatim by subquery — nothing retyped**, so there is no transcription
surface. Neither archived block contains a CTA url (keys are `content,heading`), so the restore
reverts prose only and leaves the CTA repair untouched. `rendered_html` is deliberately **not**
written by hand: the rerender regenerates it from `content_data`, which keeps that column's writer
set unchanged. Then `SQL_2026-08-25_rerender_after_restore_your_own_model.sql`.

**I induced the guard before trusting it.** The transaction's `DO`/`RAISE` block (a bare `SELECT`
cannot abort a `COMMIT`) was run **first against the damaged state** and correctly aborted —
`3 generic-text-blocks, 1 distinct openings`. Only then was it evidence when it passed. It also
refuses if any section has `content_data IS NULL`, because that escalates the next rerender to the
content writer and would silently undo the restore.

**Served bytes after (20:35Z):** three distinct sections back — *"How it works / The three steps /
Who is actually running this"*, *"Three steps, and one overnight run / …"*, *"How it works / We
train overnight / …"* — each destroyed opening exactly **1**, `password-entropy` **0**, control
string **0**, and all four CTA anchors still naming and linking the Decision Guide.

### 2026-08-25 ~20:30Z — CORRECTION to my own reading, ten minutes old: the leopardess churn was **not** mine

I measured word-churn per component across all ten pages and the largest by far was
`leopardessconsulting.co.uk/services.html` — `teaser-reveal-panel` **58%**, `info-card-grid` **23%**.
I was on the way to writing that my rewrite had eaten that page too. **It had not.** Reading the
generation trail the way `bugs_open/403` does (`jsonb_array_length` on `cards`/`items` plus
`icon-service-` refs) shows my rewrite archived **3 cards / 5 items / 0 icons at 19:52:21Z** — the
page was **already** in the damaged state when I arrived, taken by the 08-24 18:36:37 generation.

And while I was measuring, **another session restored it**: two generations at **20:23:33Z and
20:25:11Z with no `source_item_id`** put it back to **6 items / 6 icons / 6 cards**. Current state
and served bytes confirm it (6 `icon-service-` refs live). **Their restore kept my CTA fix** —
`primary_cta` still names the Agent Architecture Complexity Estimator, `password-entropy` 0.

Two lessons, and the second is the one I nearly got wrong:
- **A churn figure is a comparison against a baseline you chose.** Mine compared *pre-my-rewrite*
  with *the state at the moment I measured*. It could not distinguish "I changed this" from "it was
  already changed" — for that you need the shape the owning bug measures (array lengths, asset
  refs), not a word count.
- **On a shared estate the artefact moves under your measurement.** My leopardess numbers were
  stale within minutes because a concurrent lane was repairing the same page. Re-read before
  writing a figure down, and say when you read it.

### 2026-08-25 ~20:4xZ — MISSTEP 11 / FINDING: the one page that failed was refused for a claim I did not write — and the claim is false

`ai-agent-orchestration.com/model-directory.html` was the only one of the eleven to fail. Its
`content_rewrite` (`0745e9a4`) stopped at `validate_content` → `needs_human_review`, and its
`page_rerender` correctly stayed `triaged` — **the `depends_on` chain did its job**, so the page
never entered the text-says-one-tool-href-says-another window. That is the design working.

**Getting the actual reason took three hops, and the first two are the trap.** The work item's
`error` says only *"content validation failed: 0 blockers, 1 errors"*. The orchestration row is
`status = COMPLETED` with `error` **NULL** — the known shape; the truth is in
`collected_data->'__step_errors'`, and it repeats the same uninformative sentence. The structured
detail is in a third place, written by the action itself
(`validate_page_content.go:517`, `writeValidationFailureLog`):

```sql
SELECT jsonb_pretty(context) FROM agent_error_log
WHERE work_item_id = '<item>' AND error_code = 'CONTENT_VALIDATION_BLOCKER_DETAIL';
```

That names it: `unregistered_number "150"`, in *"More than 150 agents are listed here. Every one of
them still needs a pro…"*.

**⚠ The sentence was ALREADY LIVE. My rewrite did not write it** — `grep -c '150 agents'` on the
served page is 1, in an `<h2>`, before any retry. So this claim blocks **every** `content_rewrite`
on this page, not just mine, and has been doing so silently.

**Why the gate fires, read from the code rather than guessed** — and my first theory was wrong.
I saw that fact `aao-agent-definitions` has `value: 200`, `tolerance: gte` and a `writer_line` of
*"more than 150 active agent definitions in the production registry"*, and started writing that the
register tells the writer to say a number its own checker then refuses. **That is not the
mechanism.** `numberSupported` (`claims.go:1256`) gates each fact behind its `context_terms`
**before** it ever compares the number, and `claimWindow` (`claims.go:1349`) is only ±70 chars.
The window here — *"More than 150 agents are listed here. Every one of them still needs a pro"* —
contains **none** of that fact's terms (`agent definition`, `agents in the registry`, `ai agents`,
`agents in production`, …). So the fact is skipped, nothing else supports 150, and it is reported.
Had the sentence used the register's own phrasing, `150 ≤ 200` under `gte` would have passed.
**Reading the function beat reasoning from the config, and the config reading was plausible.**

**And the claim is false, by the page's own data.** The listing is rendered client-side by
`/tools/assets/model-directory-listing.js`, which fetches `/data/model-directory-full.json`
(HTTP 200, `updated_at 2026-08-25T18:26:58Z`): **`"count": 30`, 30 entries**, and the served HTML
holds **30** `class="model-card"` articles. Thirty are listed, not "more than 150".

> **⚠ CORRECTION to my own figure, 20 minutes old.** I first wrote 145, from
> `grep -c 'class="model-card'` — which counts *lines containing any* `model-card*` token
> (`model-card-title`, `-summary`, `-links`, …), not cards. The honest count is
> `grep -o 'class="model-card"'` → **30**, and the data file agrees independently. *A prefix grep
> counts the family, not the member; when a count matters, anchor it and corroborate it from a
> second source.*

**Action taken:** `SQL_2026-08-25_retry_model_directory_pair.sql` re-dispatches the pair with a
spec asking for both halves — the CTA labels off the retiring tool, **and** a heading that asserts
no count at all (explicitly *not* "change 150 to 30": the framework writes the sentence, this
states the constraint and the ground truth). Fresh `item_key`s, because keys dedup in any status
(`bugs_open/326`); the relink again carries `depends_on`, and the transaction has a `DO`/`RAISE`
that aborts if it does not. The old `needs_human_review` row is left standing as the record.
The count contradiction is routed to `model_directory_pipeline`, which owns the data.

### 2026-08-25 ~20:5xZ — ⚠ MISSTEP 12, and it CORRECTS §3 of the last two handoffs: **THE CANARY WAS DESTROYED TOO, seven hours earlier, and its `<p>` control read 15/15**

Running the new distinctness control over all twelve repaired pages as a closing sweep — not
because I suspected anything — returned a **second** hit:

```
finetuning.uk/technical-details.html   6 components   4 distinct   *** DUPLICATE SECTIONS ***
```

That is **the canary**. The page this lane declared *"COMPLETE and verified at the served bytes"*,
and then used to validate the recipe for the other eleven.

**Same defect, same shape, and it happened FIRST.** Archived by the canary's own `content_rewrite`
`b422751a-3745-474c-87d6-aeff50028546` at **13:05:41.827Z**, the three `generic-text-block`
components were distinct:

| pos | rendered | before |
|---|---|---|
| 2 | 1,828 B | *"The base model **itself** is a small open-weight model: one where the maker publishes…"* |
| 3 | 1,599 B | *"…meaning the **underlying weights** are published and…"* |
| 4 | 1,712 B | *"…meaning the **company that** built it has published…"* |

After that write: **1,710 / 1,712 / 1,712**, all three carrying position 4's text. Position 4 is
untouched (1,712 → 1,712); positions 2 and 3 were overwritten with copies of it.

> **⚠ THE PART THAT MATTERS. The paragraph control did not merely fail to move ENOUGH — it was
> structurally incapable of moving.** `<p>` was **15 before and 15 after**, and I recorded that in
> this file at the time as *"labels-only held"*. It held at 15/15 **because the blocks copied in
> have the same shape as the ones destroyed** — three paragraphs replaced by three paragraphs. On
> `your-own-model.html` the same defect moved the count by +3 and I nearly signed that off too; here
> the count could not have moved at all.
>
> **And then I promoted that control to the batch on the strength of this page.** The reasoning was
> "it held on the canary, so it works" — but the canary was *damaged*, and the control said clean.
> **What I actually validated was that the control is blind, and I read it as evidence that the
> repair was safe.** This is the sharpest version of the rule I wrote three hours ago and did not
> apply hard enough: *a control checked only where you believe nothing went wrong has not been
> shown to discriminate* — and if the thing you checked it against was in fact broken, the green
> result is evidence **against** the control, not for the work.

**Restored** — `SQL_2026-08-25_restore_technical_details_blocks.sql`, same method: `content_data`
verbatim by subquery from the archive the offending item itself wrote, positions 2 and 3 only,
`{content, heading}` with no CTA url so the canary's CTA repair is untouched, `rendered_html` left
to the rerender. Guard induced against the damaged state first (`3 blocks, 1 distinct` → aborted),
then passed on the restored state. Rerender dispatched:
`SQL_2026-08-25_rerender_after_restore_technical_details.sql`.

**Rate, now that both are known: 2 of the 12 pages this lane rewrote lost authored copy — 17%.**
Not a freak. Recorded into `bugs_open/403`, whose worked instance is the same disease.

**One page checked and CLEARED in the same sweep, so the sweep is not just finding what it looks
for:** `finetuning.uk/blog/chatgpt-has-your-data-does-that-matter.html` showed 22% word churn in
`article-body`, the largest non-CTA change left. Read in full: all four `<h2>`/`<h3>` headings
survive, the change is heading capitalisation (*"Private Deployments Keep Data Safe"* → sentence
case), the rewritten sentence naming the two tools — which IS the labels work, inline in prose —
and an **added** caveat that either tool can be wrong as guidance moves. No loss. **A high churn
number is not the same finding as a duplicate section, and only one of them is damage.**

### 2026-08-25 ~21:00Z — the twelfth page landed and VERIFIES; and the hero proves the claims mechanism exactly

Retry pair complete (`content_rewrite` 20:57:36, `page_rerender` 21:00:09). Verified at the served
bytes, not by status:

| check | result |
|---|---|
| `password-entropy` refs | **4 → 2**, and `grep -n` shows both are the **footer** (`site_component`) — the two live CTA buttons are gone |
| the refused claim `150 agents` | **0** |
| the CTA `<h2>` now | *"The registry lists the models you can choose from. Every one of them still needs a production stack underneath it."* — **no count asserted**, as the spec asked |
| CTA anchors | 4, every label and href naming the same tool (Build vs Buy, ROI Estimator, LLM Cost Calculator, Savings Estimator) |
| targets | all **200**; absent control `/tools/this-page-does-not-exist-391.html` **404** |
| **distinctness control** | **3 components, 3 distinct** — no duplication this time |
| body prose | hero subheadline and all body text **byte-identical**; only the `<h1>` and the CTA labels moved |
| control string never on the page | 0 |

**And the hero settles the mechanism question from MISSTEP 11 beyond argument.** Its old `<h1>` read
*"The registry behind the claims: **more than 150 agent definitions** running in production"* — the
**same number**, on the **same page**, which passed the gate every time. Because that phrasing
contains `agent definition`, a registered `context_term` of `aao-agent-definitions`, so the fact
(value 200, `gte`) was consulted and `150 ≤ 200` passed. The CTA's paraphrase *"more than 150 agents
are listed here"* dropped the term, so the same fact was never consulted and the same number was
reported unregistered.

So the page carried **both** a licensed, true version and an unlicensed, false version of one figure,
in adjacent components. That is worth keeping in mind when reading a claims refusal: the gate is not
telling you the number is wrong, it is telling you **nothing in the register vouches for the sentence
as phrased** — and here the phrasing difference happened to also be the difference between a true
claim about the registry and a false one about the page.

> **One honest over-reach to record.** My spec named *"the call-to-action heading"*. The writer also
> rewrote the hero `<h1>`, which was a **true and licensed** claim, to *"what's listed here, and what
> still has to be true before you run any of it in production"*. Nothing false was introduced and the
> new line is defensible, but a correct claim was removed because I asked for a figure to be dropped
> and did not scope which one. **A spec that names one heading does not stop the writer visiting the
> others** — the same lack of bounding that MISSTEP 10 is about, in its harmless form.

### 2026-08-25 ~21:10Z — canary restore LIVE; and **MISSTEP 13: my own replacement control false-positives, caught by running it once more**

Canary rerender complete 21:05:40. Served bytes: the three text-block openings each appear
**exactly 1** — *"the base model **itself**…"*, *"…the **underlying weights**…"*, *"…the **company
that** built it…"* — `password-entropy` **0**, control string **0**, all four CTA anchors still
naming and linking the Decision Guide. Restored.

> **And the `<p>` control reads 15. It read 15 when the page was destroyed.** Same number, both
> states, four hours apart. That is the whole case against it in one line.

**Then the closing distinctness sweep flagged the canary AGAIN — 6 components, 5 distinct — and it
was WRONG.** The three restored blocks all begin *"The model and its licence The base model is a
small open-weight model, meaning "*; positions 3 and 4 diverge at character **~81**
(*"the underlying weights…"* vs *"the company that built it…"*). The shared heading plus the common
sentence stem consume the entire 80-character window.

**I nearly explained it away** — the served bytes said all three were distinct, so the row looked
like noise. Measuring instead of dismissing gave the real answer, and it is worse than a one-off:

| state of the canary | `at_80` | `at_200` | `md5(full text)` | truth |
|---|---|---|---|---|
| pre-damage baseline (13:05:42Z archive) | **5** | 6 | 6 | clean |
| damaged (13:37:55Z archive) | 4 | 4 | 4 | **damaged** |
| restored (now) | **5** | 6 | 6 | clean |

**The 80-character form scores the canary 5-of-6 whether or not it is damaged.** On that page it
never discriminated at all — it was right about the damage by coincidence, for a reason that had
nothing to do with the damage. `[MEASURED 2026-08-25]` across all twelve pages, `md5()` of the full
stripped text gives **4 of 6 on both damaged pages, 6 of 6 on both restored ones, and zero false
positives**; `left(txt,80)` disagrees on exactly one page, in the false-positive direction.

**So the check is now `count(DISTINCT md5(txt))` vs `count(*)` — no window, no tunable.** Full text
works here *because* the copies are byte-identical once tags and whitespace are stripped, which I
verified on both instances rather than assuming (the rendered-HTML md5s differ, because the wrappers
carry slot and position — that difference is not content).

**Corrected in place, before anyone else read it:** `bugs_open/391`, `bugs_open/403` (both the
CONTRIB and its addendum), `LANDMINES.md` (whose trap line also still claimed the canary was
*undamaged* — the very error this session had already disproved), and `HANDOFF_2026-08-25c` §3.

> **The lesson, and it is the same one twice in one evening.** I found the first defect with a
> control, wrote that control into five documents as *the* answer, and had **not once run it against
> a page I knew to be clean.** Exactly the failure I had just finished writing up about the `<p>`
> count — validate a control only where you expect it to fire and you learn nothing about what it
> does elsewhere. **A detector needs a negative case before it is a detector**, and mine got one only
> because a restore gave me a known-good state to re-run it on.

### 2026-08-25 ~21:15Z — the `shared-ledger-not-appended` advisory on my LANDMINES correction: checked, and it is clean

The pre-commit hook flagged commit `3effa14c9` for removing **5 lines** from `LANDMINES.md`, a
fleet-wide append-only ledger, on the grounds that a removed line is most likely another session's
entry. Correct thing to flag; worth discharging rather than waving through, because "I only touched
my own entry" is exactly what someone who had clobbered a neighbour's would also believe.

Three checks, and the first is the one that cannot be fooled by content:

```
git diff --numstat 3effa14c9^ 3effa14c9 -- …/LANDMINES.md   ->  15 added, 5 removed
git diff … | grep '^@@'                                     ->  ONE hunk, @@ -18140,15 +18140,25 @@
grep -n '^### A `content_rewrite` commissioned…'             ->  my entry's header is line 18139
git show <before>:… | grep -c '^### '  vs  <after>           ->  728 and 728
```

The single hunk begins at 18140, immediately inside my own entry (header 18139), and the total entry
count is **identical before and after** — so no entry was deleted, and nothing outside mine was
touched. The 5 lines were the wrong `left(text,80)` check and the trap sentence that still called the
canary undamaged, both written by me an hour earlier in this same session, both replaced by a dated
`⚠ CORRECTED` note in place. That is precisely what the ledger's own guidance asks for
(*"correct in place with a dated note rather than a rewrite"*).

⚠ Note the entry-count check is what actually proves it. A hunk range only shows where the diff
*starts*; a deletion running off the end of my entry into the next one would still open at 18140.
Counting `### ` headers on both sides is the assertion that no entry vanished.

Verifier re-armed for the changed entry (`./scripts/landmines-verify-dispatch.sh`, correlation
`f2b7a0a0-79a5-4bbb-ad76-04bb1864d368`) — a corrected entry needs re-verifying, and the arming state
is consumed by a bare `--apply`.

### 2026-08-26 ~08:5xZ — fleet LLM outage VERIFIED at the artefact; this lane's two dead rows closed; and MISSTEP 14: I never read my own landmine verdict

**The API credit outage is real, and it is not a passing note.** The `loanzy_uk_example_site` lane
reported it; verified here at the artefact rather than taken on report, with a pre-outage control so
"zero" means something:

| hour (UTC) | calls | ok | failed |
|---|---|---|---|
| 08-25 21:00 | 159 | **157** | 2 |
| 08-25 22:00 | 128 | **124** | 4 |
| 08-25 23:00 | 123 | 107 | 16 |
| 08-26 00:00 → 08:00 | ~691 | **0** | **691** |

Verbatim, 690 of 691: `API request failed with status 400: {"type":"error","error":
{"type":"invalid_request_error","message":"Your credit balance is too low to access the Anthropic
API…"}}`. (The single non-Anthropic failure is an unrelated `ollama-adapter` timeout.) **Nine
consecutive hours with zero successes**, and the fleet is still attempting ~80 calls/hour into it.

**Second-order damage, which nobody had measured and which OUTLIVES the top-up:** `[MEASURED
2026-08-26 08:5xZ]` **21 work items have exhausted `max_attempts` against the dead API since
00:00Z.** Those do **not** self-heal when credit returns; they are `failed` for good and need
re-firing by hand.

> **⚠ CORRECTED 2026-08-26 — my enumeration above was INCOMPLETE, and it is the re-fire list.**
> It originally read *"7 `unbuilt_internal_link`, 4 `content_rewrite`, 3+1 `improve_tool`, 2
> `needs_page`, and one `needs_diagnosis`"* — which **sums to 18**, beside a total of **21**. Two
> `dead_fragment_link` and one `page_rerender` were dropped. The `loanzy_uk_example_site` lane's
> composition is the correct one; re-queried here by `item_type` alone to settle it:
>
> | item_type | burned | cites "credit balance" |
> |---|---|---|
> | `unbuilt_internal_link` | 7 | 7 |
> | `content_rewrite` | 4 | 4 |
> | `improve_tool` | 4 | 4 |
> | `dead_fragment_link` | 2 | 2 |
> | `needs_page` | 2 | 2 |
> | `needs_diagnosis` | 1 | 1 |
> | `page_rerender` | 1 | **0** |
> | **total** | **21** | **20** |
>
> The `page_rerender` is the single non-matcher — a failure that landed in the window for another
> reason — so an error-predicated re-fire correctly skips it and it needs judging by hand. That
> lane's `RUNBOOK` holds the recipe.
>
> **This is the exact error I corrected in `bugs_open/405` twelve hours earlier** — a correct total
> beside an enumeration that does not sum to it — and here the enumeration IS the operational
> artefact: anyone re-firing from my list would have restored 18 of 21 and left three broken with
> the outage marked resolved. `WRONG_CALLS.md` has it. The other 1,239 `triaged` / 170 `detected` rows are
merely queued and will drain. **The outage is the visible event; the burned retry budget is the
residue.** Whoever tops up should re-fire those 21.

**⇒ Consequence for this lane: step 3 (retirement) CANNOT START until credit is restored.** It goes
through the framework, and the framework's content path is LLM-shaped. A `cta_links_stale` rerender
is explicitly no-LLM and would still work; a retirement is not that.

**Closed this lane's two dead rows** (`SQL_2026-08-26_close_the_superseded_model_directory_pair.sql`,
no LLM involved). Both were superseded by the retry pair that completed 21:00:09Z:
`cta_label_relevance:2c7c836c…` sat at `needs_human_review`, and `cta_relink:2c7c836c…` sat at
`triaged` **depending on it** — and `load_work_item_actions.go:713` only dispatches on
`complete`/`verified`, so that relink was **permanently undispatchable**. I had documented leaving
the first "as the record" and had **not noticed the second at all**. Left alone it would sit for
ever, and a future reader querying this lane would see an unfinished pair beside 26 complete ones —
the `bugs_open/396` shape. Cancelled, not deleted, each row carrying its own reason and the key of
what replaced it. Guard induced against the pre-change state first (`found 0`, aborted), then passed:
**0 open, 2 cancelled, the rest complete.**

> **MISSTEP 14 — I dispatched the landmine verifier twice, reported "re-armed", and never read the
> verdict.** Both runs returned **`NEEDS_HUMAN_REVIEW`** (20:45:29 and 21:11:34), while other lanes'
> entries that evening returned `STILL_VALID`. I only looked this morning, and only because the
> outage made me audit my own outstanding dispatches.
>
> **The verdict is not "your entry is wrong" — it is "this instrument cannot answer."** The verifier
> is a **Go-only** index: *"Scope: 8700 symbols, the indexed corpus holds only: .go"*, pinned at
> commit `e347c5ad` of **2026-08-23** — *"the last pushed tip, not the present tree"*, which the
> verifier itself flags as **predating the incident by two days**. It confirmed the three Go-visible
> footprints (`site_work_items`, `spec.suggestion`, `page_components.content_data`) and honestly
> declared the rest unanswerable: *"1 NOT ANSWERABLE by this index; 3 ran and matched nothing in
> scope."* `item_type='content_rewrite'`, `page_component_history.source_item_id` and the SQL repair
> recipe are database artefacts it structurally cannot see.
>
> **So this is the session's FOURTH instrument-scope finding, and it generalises past my entry: any
> landmine whose footprint is a TABLE, a work-item type, a migration or a SQL recipe will return
> `NEEDS_HUMAN_REVIEW` from this verifier no matter how true it is** — because the corpus is `.go`
> only, while `LANDMINES.md` is explicitly a corpus of *paths, tables, commands and symbols*. For
> that whole class the verdict does not discriminate. Credit to the verifier for saying "not
> answerable" instead of returning a green; that is the behaviour I spent yesterday failing at.
> **Not re-dispatched** — it is LLM-shaped, and there is no credit.

### 2026-08-26 ~09:0xZ — retirement blast radius RE-MEASURED, and the footer half of the plan is the wrong shape

Step 3 is blocked on the credit outage (framework content path is LLM-shaped), so I did the two
things the handoff says are owed and that measurement can answer. Both changed the plan.

**1. Blast radius, re-measured `[MEASURED 2026-08-26 ~09:00Z]`** — the standing figure was **91**
`page_components` rows and was taken *before* step 2 ran:

| domain | rows | pages |
|---|---|---|
| ai-agent-orchestration.com | 30 | 19 |
| leopardessconsulting.co.uk | 20 | 19 |
| finetuning.uk | 13 | 11 |
| **total** | **63** | **49** |

Step 2 removed **28**. `content_data` and `rendered_html` agree **row for row** — `0` content-only,
`0` html-only, `63` both — so there are no stale renders hiding a reference, which is the RFC_008
writer worry and it is clean here.

**2. ⚠ THE FOOTER IS GENERATED, NOT AUTHORED — so §5 step 1's "retire the page WITH the footer entry
in the same operation" describes an edit that does not exist.**

The `ai-agent-orchestration.com` footer `site_components` row holds, in `content_data`, exactly:
`{year, email, phone, domain, tagline, company_name}` — **zero** tool links, **zero**
`password-entropy`. The reference exists only in `rendered_html`. And the served footer lists
**exactly the six live tool/game pages, in `nav_order` order** — the same six the ranking query
returns. It is derived from the nav tables, which `populate_nav_tables` rebuilds from `pages`
(LANDMINES, *"the obvious agent for a nav change deletes every child-path link"*).

**So the risk inverts.** There is nothing to edit out of the footer; retiring the page removes it
by construction. What can go wrong is the opposite: **retire the page and fail to refresh the footer,
and every page on the site serves a link to a dead one** — the footer is a `site_component`, shared
site-wide. The recipe already exists and the handoff does not name it: `nav-link-fixer` refreshes
`site_components.rendered_html` from the existing nav tables, then propagate in **assemble mode**
(`page-rerender` with **no** `spec.reason`); worked script
`docs/leopardessconsulting/scripts/reconcile_footer_nav.sh`.

> **I caught this only because the count moved under me.** Last night `aiao/index.html` served **2**
> `password-entropy` refs; this morning **1**. Nothing was retired in between — the footer was
> re-rendered at **01:01:44Z** and its generated list simply stopped featuring the tool in one of
> two blocks (*"Tools / Password Strength Physics"* → *"Tools / AI Readiness Quiz"*). **The "2 in the
> footer" I reported last night, in this file and to another lane, was a snapshot of a rotating
> derived value presented as a fixed count.** A count of a generated list is a reading, not a
> property.

**3. The three sites will NOT retire the same way** `[MEASURED 2026-08-26]`:

| domain | `in_header` | `in_footer` | `site_nav_items` rows | served refs |
|---|---|---|---|---|
| ai-agent-orchestration.com | **t** | f | **1** | 1 (footer) |
| finetuning.uk | f | f | 0 | 0 |
| leopardessconsulting.co.uk | f | f | 0 | 0 |

Only **one** site has any nav/footer presence at all. The other two need no nav work — their 33 rows
are page-body references only.

> **And this resolves the discrepancy I flagged an hour ago rather than leaving it as "the flag
> over-reports".** `in_header=true` on `ai-agent-orchestration.com` while the served **header** does
> not carry the link is not the flag lying: `classifyPagesForNav` **demotes a child-path page**
> (`/tools/…`) that declares a nav flag into the **`utility`** group instead of dropping it
> (LANDMINES, narrowed 2026-07-31, NAV-013) — and `utility` renders in the **footer**. Flag honoured,
> classifier overrides the placement. The earlier reading was right about the artefact and wrong
> about the cause; *"the flag over-reports"* would have sent the retiring session looking for a
> header edit that is not there.
>
> The same landmine names **leopardess `/tools/password-entropy.html`** as the one fleet-wide row
> hand-written into `utility` against a page declaring neither flag. Its `site_nav_items` count is
> now **0** — consistent with a rebuild having removed it exactly as that entry predicts.

### 2026-08-26 ~09:1xZ — credits restored; and **retirement is TWO steps, not one — the handoff's order deadlocks**

API confirmed recovered at the artefact before acting on it: the 09:00 hour is **124 calls, 124 ok,
0 failed** (07:00 was 0/60, 08:00 was 2/106 across the transition).

**Read the retraction machinery before firing it, and the plan does not survive the reading.**

- **"Retire" is not one operation.** `pages.status='archived'` is a **hand-run SQL** step — the
  action's own header says so: *"nothing in this codebase archives a page … there is no writer of
  `status='archived'` at all"*. Removing the FILE is a separate capability,
  `retract_page_deployment`, driven by the live `page-retraction` agent (`site_id_field`,
  `page_ids_field`). Archiving alone **freezes the page and keeps serving it**, so there is no 404
  window between the two.
- **⚠ AND THE RETRACTION REFUSES WHILE ANYTHING EDITORIAL LINKS IN.** `retract_page_graph.go`:
  *"INBOUND, editorial → REFUSE the retraction and name the referrers"*, deliberately, so that
  "dead link created by a retraction" is unrepresentable rather than merely detected. Measured with
  the action's **own** quote-delimited predicate (`href="<url>"`, not a substring — the file explains
  why the difference matters): **61 rows across 47 active pages** — aiao 30/19, leopardess 18/17,
  finetuning 13/11 — **plus the aiao footer `site_component`**. So retraction would refuse on all
  three sites today.

> **⇒ THE HANDOFF'S ORDER IS A DEADLOCK, and it is the second ordering trap this lane has hit.**
> §5 said *retire the pages → then re-resolve the label-less fields*. But the re-resolution is
> blocked by KEEP #2 while the destination is valid, and the retraction is blocked by those same
> references. Each step is the other's precondition.
>
> **It breaks on `validPages`.** `loadResolverPageSet` (`resolve_internal_links_action.go:964`)
> selects `WHERE status NOT IN ('deleted','archived')`. So **ARCHIVING alone** drops the page out of
> `validPages`, KEEP #2's `validPages.Contains(current)` goes false, KEEP #3 cannot catch a relative
> `/tools/…` path, and control reaches the positional pick — which the `nav_order` 1 → 900 demotion
> already made correct. **Archiving is the key that turns both locks**, and it needs no LLM and no
> file deletion.

**The corrected sequence — three steps, not two:**

1. **ARCHIVE** the three pages (SQL, reversible, page keeps serving).
2. **RE-RESOLVE** the 61 inbound references — `cta_links_stale` rerenders, **no LLM**. Now unblocked,
   because the destination stopped being valid at step 1.
3. **RETRACT** the deployment via the `page-retraction` agent — now it will not refuse, because
   editorial inbound is 0. It also deactivates the `site_nav_items` row itself
   (structural inbound is *mechanised*; editorial is refused; newly-stranded outbound is *reported*).

**Canary order chosen from the measurement, simplest first:** `finetuning.uk` (13 rows / 11 pages,
0 nav rows, both flags false) → `leopardessconsulting.co.uk` (18/17, 0 nav rows) →
`ai-agent-orchestration.com` **last** (30/19, 1 `site_nav_items` row, `in_header=t`, and the only
site whose footer carries the link).

> **⚠ A trap found while reading, worth its own LANDMINE and not specific to this lane: the two
> sibling retraction actions have OPPOSITE `dry_run` defaults.** `retract_asset_files` states
> *"absence means TRUE"* and cites the 2026-08-02 owner ruling that a dangerous branch defaults OFF.
> `retract_page_deployment` does `if dry, _ := config["dry_run"].(bool); dry {` — a bool's zero value
> is `false`, so **omitting the key DELETES**. The two files name each other as siblings in their
> first lines, which makes the wrong inference *more* likely, not less. The live `page-retraction`
> agent passes no `dry_run`, so it runs live by design.

### 2026-08-26 ~14:5xZ — the canary relink sat 5½ hours unclaimed. Two findings, and neither is what it looked like

At 14:51 the canary relink was still `triaged`, never claimed, dispatched 09:17. **Did NOT call the
queue dead** — that is this lane's MISSTEP 9 and I measured the service interval first: the fleet
completed **146–154 `page_rerender`s per hour** through 12:00–14:00. The queue is alive and busy.

**What I then got wrong twice, and the order matters because each wrong answer was plausible:**

1. *"It's the backlog."* **622** `page_rerender`s sit at `triaged`, 533 over an hour old — a genuine
   pile-up draining from the outage. But my row is **priority 30, the ONLY item at that priority**
   (the rest: 20 at 35, 601 at 80), with **0 ahead of it**. Not a queue-depth story.
2. *"It's `retry_after`."* `bugs_open/307` stamps a not-claimable-before time on a failed attempt, and
   a 9-hour outage is exactly what would push a site's queue into the future. Testable, and
   **REFUTED**: finetuning.uk has 73 eligible build items, **all 73 claimable now, zero deferred**.
   Recording the refutation because it was the best hypothesis available and it cost one query.

**What it actually is — read from the site selector, not guessed.** `build-pipeline-trigger` fires an
agent whose selector is:

```sql
… ORDER BY wi.created_at ASC, wi.priority ASC, wi.id ASC LIMIT 1
```

plus `AND NOT EXISTS (… status='claimed' on that site)`. So each 30s tick picks **the single site
owning the globally oldest eligible item**, skipping sites with one already in flight. finetuning.uk
is **rank 15 of 31** waiting sites (oldest eligible 05:03:29) — **not excluded, just behind fourteen
older sites.** Nothing was broken.

> **⚠ THE ASSUMPTION THIS LANE HAS BEEN WORKING UNDER IS WRONG: `priority` DOES NOT MAKE A SITE GET
> PICKED SOONER.** Site selection orders by **`created_at` first**; `priority` is only a tiebreak
> among items sharing a timestamp. So an item's dispatch time is governed by **the age of the OLDEST
> item on its site**, which is somebody else's row. Once the site IS selected, the *dispatcher's*
> own `ORDER BY wi.priority ASC, wi.created_at ASC` does honour priority — so priority works
> **within** a site and is inert **between** sites. I set 30 and 35 on every item this lane has
> dispatched believing otherwise; last night's fast turnarounds were a shallow queue, not the
> priority.
>
> The tiebreak is not vacuous, which is why the distinction has to be stated precisely rather than
> as "priority is ignored": `[MEASURED 2026-08-26]` of 1,158 eligible items, **381 share a
> `created_at` with another** — batch inserts like this lane's own 22-row transaction at 19:17:57.
> Within such a batch, priority decides.

### ~~⚠ AND A FLEET-WIDE FINDING THAT IS NOT THIS LANE'S: the outage mitigation is still on, 6 hours after the outage ended~~ — **REFUTED 2026-08-26, see the correction below**

`build-pipeline-trigger-2` — *"Sibling dispatch turn #2 (dispatch_throughput 582-584, N=2). Same
selector/loop as build-pipeline-trigger"* — is **`enabled = false`**, and its row changed at
**2026-08-26 08:51:17**, eight seconds after its final run. That is inside the credit outage, and it
matches exactly the mitigation the `loanzy_uk_example_site` lane put to the owner ("the option to
pause build-pipeline-trigger ×2 until top-up").

**Credit was restored at ~09:00 and verified (124/124 in that hour). The pause has outlived its cause
by ~6 hours**, halving fleet dispatch throughput precisely while a 622-item backlog drains and 31
sites queue behind one lane.

**Not re-enabled by me.** It is a fleet-wide throughput switch and it appears to encode an owner
decision; flipping it back is the owner's call, not a side effect of my lane's canary being slow.
Surfaced to the owner and to the lane that proposed the pause. **This is the outage's second
residue** — the first being the 21 items that burned their attempts. Both share a shape: *the
visible half of an incident ends, and the mitigations it justified stay on.*

> **⚠ REFUTED the same afternoon by the `loanzy_uk_example_site` lane, and they are right — verified
> here first-hand rather than accepted on report.** `build-pipeline-trigger-2` was **not** a
> mitigation left running. It was **deliberately retired at 08:51Z under an owner ruling** by the
> `dispatch_throughput` lane: migration **`637_dispatch_lever_B_interval_not_sibling.sql`**, whose
> own first line reads *"B = retire the sibling row, set the ORIGINAL row's interval_seconds
> 60 → 30"*, and whose post-check `RAISE`s *"% enabled trigger rows, expected exactly 1 (ruling B)"*.
> Their commit `a5fd1651e` carries the measured case: fire cadence p50 90s → 60s, lost claims
> **12.9%** against 58–60% trailing-24h, because the sibling pair was **co-picking the same site 94%
> of the time** under a phase lock. **A single dispatch lane at 30s IS the ruled configuration**, and
> migration 584's VERIFY now `RAISE`s by name on *"a second enabled trigger row OR interval<30"* —
> so re-enabling it would not merely have been unwise, it would have tripped a guard written to hold
> that ruling.
>
> **The timing match that convinced me — disabled mid-outage, eight seconds after its last run — is
> a coincidence.**
>
> **⚠ AND THE DISCONFIRMING EVIDENCE WAS IN MY OWN QUERY OUTPUT.** I printed both rows together:
> `build-pipeline-trigger | t | 30` and `build-pipeline-trigger-2 | f | 60`. **A pause leaves the
> survivor's configuration untouched; only a deliberate reconfiguration changes it.** That interval
> asymmetry — 30 on the live row against the 60 the retired row is frozen at — is exactly migration
> 637's arithmetic, sitting on my screen, and I read past it because I was already reading the row
> for something else (`enabled`).
>
> **Why I made the error, which is the transferable part:** I had just been *right* about an outage
> residue — the 21 items that burned their attempts — and had written that up in the words *"the
> visible half of an incident ends, and the mitigations it justified stay on."* **A frame that fits
> one case primes you to fit the next one to it.** The second case had a timing coincidence to offer
> and I took it. What made the wrong inference cost nothing was **not** my judgement: it was
> declining to flip a fleet-wide switch I did not own. *Check-before-flip is what converts a wrong
> inference into a message instead of an incident.*
>
> And what actually caught it was not reasoning at all — the peer says a **memory line naming
> trigger-2 as that lane's designated ROLLBACK LEVER** made "deliberate retirement" a live
> alternative worth ruling out. **A durable record beat both our inferences**, which is the argument
> for writing the boring ones down.

### 2026-08-26 ~15:2xZ — the 5½-hour wait RESOLVED: it was a real defect, now `bugs_open/413`. This lane PARKS, deliberately

The `dispatch_throughput` lane took my undiagnosed handover and filed
**`bugs_open/413_HANDOFF_2026-08-26_selector_and_loader_disagree_on_ordering_so_a_pinned_item_starves_younger_sites.md`**
(090 run `250188a7` in flight as the independent check).

**The mechanism, theirs not mine:** the site selector ranks sites by their oldest eligible row's
**age**; the loader then serves the picked site by **priority**, max 5. So an old row at a poor
priority (e.g. `audit_tool` @140, "run last within the site") **keeps winning site selection for its
site while never being loaded** — better-priority rows keep arriving ahead of it. The site's age
freezes at the pinned row, and every younger site is starved behind it.

> **⚠ CORRECTED 2026-08-26 ~17:2xZ — "gets nothing" is TOO STRONG, and the `dispatch_throughput` lane
> caught the same overstatement in their own filed symptom and fixed it there too.** Fall-through
> service does exist: a younger site is reached when the sites above it are simultaneously busy. The
> precise claim, which all the evidence supports, is **"served only by fall-through, with no bound on
> the wait"** — and this lane's own site is the measurement behind it: **one build-dispatch-loop in
> 12 hours.** The defect is the *absence of a bound*, not the absence of service, which is why an age
> floor is the candidate that addresses it.
>
> **090 verdict on `bugs_open/413`: UNVERIFIABLE** — *"NOT confirmed (stopped: iteration-cap)"*,
> no fix proposed, neither confirmed nor refuted. **The mechanism is NOT amended**, so nothing above
> needs revising. Their reconciliation, for the record: the loop sampled `status='detected'` rows as
> "oldest rows" — a population the selector never sees, since it requires `triaged`/`approved` — and
> its "a dozen+ sites cycling" counter-point is what 413 *predicts* among sites at the old end of the
> order, refuting a monopoly claim the file does not make. 413 stands on declared first-hand
> verification under the 2026-07-31 ruling.

**The piece I could not have found:** it is invisible in `site_work_items`, where *"never selected"*
and *"selected but loaded zero"* look identical. `orchestration_states` separates them —
finetuning.uk had **exactly one** build-dispatch-loop in 12h, so the problem is selection-side. That
is the query I did not know to run, and it is why handing the observation over undiagnosed was
right rather than merely cautious.

**My rank-13/19/20 anomaly resolves too, and the resolution is a rolling-window trap I should have
suspected:** those sites held pre-05:03 rows *earlier* which have since drained or archived. They
stopped receiving loops at the exact moments their own old rows drained — which is the cross-check
that the model is right. I compared a *current* rank snapshot against *historical* service and read
the mismatch as an anomaly in the ordering; it was an artefact of comparing two different instants.

**Where this lane actually stands, measured:**

- `finetuning.uk` is **not itself pinned** — its oldest row is `generic_theme` @**60** (05:03:29),
  which would load fine. It is a **victim of pins on the seven sites ahead of it**; the peer confirms
  two of those (`gaswholesalers`, `loanandmortgagecalculator`) are starving the same way.
- **My canary is load position 1** in finetuning.uk's own queue — priority 30, ahead of the @60
  block. **The moment the site wins a turn, it runs.** Nothing about my dispatch is wrong.

> **DECISION: PARK. Do not fire the rerender directly.** The estate's direct-fire remedy is for a
> **dead** queue; this queue is doing **265–278 claims/h**. Firing round a *starved* site is exactly
> MISSTEP 9 in a new costume — and that time I also had a real-sounding justification.
>
> **The deciding fact is that there is NO live damage.** Verified at the artefact just now:
> `/tools/password-entropy.html` is archived and still returns **200** (absent control 404), and
> `/about.html` still resolves both its links to it. Archiving freezes a page and keeps serving it,
> so the intermediate state is **coherent, not broken**: the page is archived, still served, still
> linked, nothing dead. The remaining work is an *improvement*, not a repair, and it does not justify
> routing around a known defect another lane is actively fixing.
>
> The state is also **guarded against the one way it could go wrong**: a retraction that removed the
> file while inbound links remain would create dead links — and `retract_page_deployment` refuses
> exactly that case. So the sequence cannot be half-completed into damage by anyone.

**Adopted from this lane by `dispatch_throughput`:** the **per-site starvation floor** (worst-case
hours-since-last-claim across waiting sites) is now in their RUNBOOK and is a required part of
tomorrow's 24h post-B read — because the aggregate they were reading (265–278 claims/h, 13–17
distinct sites/h) is healthy *and* fully consistent with a mid-rank site receiving nothing for six
hours. My Phase 3 note is in 413 too, and correctly framed as cutting both ways (deeper loads reach
@140 rows sooner, but each pick holds a site longer) — to be measured, not assumed.

### 2026-08-26 ~15:2xZ — MISSTEP 15: I chose the one canary site that could not run — and two refinements from the throughput lane that correct me

**The canary moved to `leopardessconsulting.co.uk`.** Same test, site that is demonstrably alive.

> **MISSTEP 15, and it is a selection error, not a measurement one.** I picked `finetuning.uk` as the
> canary because it was **structurally simplest** — 0 `site_nav_items` rows, both nav flags false,
> fewest inbound references. That reasoning was sound on the criteria I used. **I never asked whether
> the site was being dispatched.** `[MEASURED 2026-08-26 15:23Z]` claims in the last three hours:
> leopardess **72**, aiao **111**, finetuning.uk **ZERO** (last claim 05:09:30). Of the three
> candidate sites I chose the only starved one, and then spent five and a half hours diagnosing the
> silence.
>
> **The check that would have caught it costs one query and belongs in every canary choice:**
> *is this target currently being serviced?* `SELECT max(claimed_at), count(*) FILTER (WHERE
> claimed_at > now() - interval '3 hours') FROM site_work_items WHERE site_id = …`. **"Simplest" is
> a property of the subject; "will actually run" is a property of the system**, and a canary needs
> both. Every canary criterion I had was about the page; none was about whether anything would pick
> it up.
>
> The five and a half hours were not wasted — they became `bugs_open/413` — but that was luck about
> where the silence led, not a good canary choice.

**finetuning.uk's archive STAYS.** Archived pages keep serving (verified: 200, absent control 404),
its inbound links still resolve, and its relink is **load position 1** whenever 413 clears. Parked
safe, nothing to undo.

**Two refinements from `dispatch_throughput`'s census `[MEASURED 2026-08-26 ~15:5xZ]`, both of which
correct something I told them** — 25 sites hold eligible work, **13 pinned / 12 victims**, and
finetuning.uk is confirmed a pure victim (`oldest_load_rank` 2, my canary at rank 1):

1. **Pin status is DYNAMIC.** Their own headline example — mortgagecalculator's `@140` trio —
   **unpinned itself within ~40 minutes** when its better-priority inflow paused and the queue
   drained below the load cap. So the pinned/victim discriminator I proposed produces **a dated
   snapshot, not a classification.** I offered it as though it partitioned sites; it partitions
   *this instant*.
2. **⚠ "Unpinning the pinned sites frees the victims for free" — which I said to them — is WRONG.**
   A site's own pin does not starve *itself*: **starvation is POSITIONAL**, i.e. being behind pins in
   age order. `loanandmortgagecalculator` is **pinned AND starving since 04:39**, because of pins
   *ahead* of it. So the neat two-group causal model (pinned sites cause, victims suffer) does not
   hold — a pinned site is also a victim of the pins above it. The consequence for their fix
   ranking is real: candidate 1 (make the two orderings agree) removes pins, but **candidate 2 (an
   age floor) is the only one that bounds the positional wait**, which is the harm both groups
   actually suffer.

*I generalised a two-case observation into a taxonomy, and the taxonomy did not survive a census.
Twelve victims and thirteen pinned sites is what the measurement says; "pinned sites starve victims"
is what my two examples suggested.*

### 2026-08-26 ~15:5xZ — canary ran. Hypothesis CONFIRMED in the DB; and the canary earned its keep by finding something the plan did not anticipate

The leopardess canary claimed 15:48:06, completed 15:48:48.

**THE HYPOTHESIS IS CONFIRMED, in the database.** Archiving *does* unblock the recompute. Both CTA
fields moved off the archived page at 15:48:34, and **both surfaces agree** — `content_data` and
`rendered_html` each carry `/tools/ai-agent-roi-estimator.html` and **neither carries
`password-entropy` anywhere** on any of the page's three components. So `loadResolverPageSet`'s
`status NOT IN ('deleted','archived')` is indeed what releases KEEP #2, exactly as read from the
source. **The three-step sequence archive → re-resolve → retract is sound.**

> **⚠ MISSTEP 16, and it is the same error a third time, now INSIDE the verification harness.** My
> first propagation poll compared `last-modified` against a **hardcoded epoch** — `1756223280`, which
> decodes to **2025**-08-26 15:48, a year early. Every possible `last-modified` exceeds it, so the
> loop reported **"ARTEFACT UPDATED"** on its first iteration **while the artefact had not changed at
> all**, and printed a stale `last-modified: 15:24:05` in the same breath as its own success message.
> *A check whose threshold is wrong in the permissive direction cannot fail.* The `<p>` count, the
> 80-char prefix, and now this — and this one is the worst of the three, because the other two were
> measuring the subject and this one was measuring **whether my measurement had happened yet.**
> **Compute a threshold, never type one:** `date -u -d '<the moment>' +%s`, and print the decoded
> value beside it so a wrong one is visible on sight.

**Artefact still pending at 15:52** — `last-modified: 15:24:05`, `cf-cache-status: DYNAMIC` (so not
an edge cache), against a deploy commit at ~15:48. The git half reported success with a sha; B2 sync
+ purge runs on push and takes time. Re-polling with a correctly computed threshold and a real
timeout. **I said "REFUTED at the served bytes" before checking this and that was wrong** — the
served bytes are the right place to verify, but only *after* confirming the artefact you are reading
is the one you deployed. **`last-modified` is the header that answers that**, and I read the body
before I read the header.

### ⚠ AND THE FINDING THAT CHANGES THE PLAN: the retirement's success criterion is satisfiable while the pages stay wrong

The recompute moved both CTAs to `/tools/ai-agent-roi-estimator.html`. Their labels are:

| slot | label | new destination |
|---|---|---|
| `hero` | *"Get in touch about a piece of work"* | `/tools/ai-agent-roi-estimator.html` |
| `call_to_action` | *"Write to leopardess@contactforsales.com"* | `/tools/ai-agent-roi-estimator.html` |

**A button reading "Write to leopardess@contactforsales.com" that opens an ROI estimator is wrong**,
and it is wrong in a *more* deceptive way than the password tool was — a plausible-looking tool link
is one nobody reports.

**I did NOT cause this, and the history says so.** `page_component_history` carries the identical
label/url mismatch back to **2026-08-24 18:33** at least — contact-intent copy already pointing at
`password-entropy` long before this lane touched the page. My recompute moved it **tool → tool**, not
contact → tool. This is `bugs_open/248`'s damage shape, and the KEEP #1 comment in
`rerender_page_sections_action.go` describes the mechanism exactly: *"a minted /contact.html whose
copy later went generic … took no keep at all and fell to the POSITIONAL PICK, which replaced a
working contact button with an unrelated tool page."*

**⇒ The consequence for step 2, and it is not small.** Re-resolving the ~60 label-less fields will
move them off `password-entropy` onto whichever tool ranks first — **satisfying this lane's stated
success criterion ("zero `password-entropy` references at the served bytes") while leaving
contact-intent buttons pointing at tools.** The `nav_order` demotion makes the pick the *right tool*;
it cannot make a tool the right *kind* of destination for copy that asks the reader to get in touch.

**So the criterion needs a second clause before step 2 runs at scale:** every rewritten CTA's label
and destination must agree in **kind** — a contact-intent label must resolve to a contact
destination, not merely to a live page. That is a check I can write against `content_data` before
dispatching anything, and it should gate the batch. **This is exactly what a canary is for, and it
is the second time this lane's canary has paid for itself.**

### 2026-08-26 ~15:5xZ — ⚠ FLEET-WIDE: the site deploy pipeline STALLED at ~15:25. Work items report `complete`, git is correct, nothing reaches the live sites

Chasing my own stale artefact located the boundary exactly, and it is not this lane's:

| stage | state |
|---|---|
| `page_components.content_data` | ✅ correct — CTAs moved off the archived page 15:48:34 |
| `page_components.rendered_html` | ✅ correct — both surfaces agree, `password-entropy` absent |
| git commit `f6e8734463ee`, 15:48:39Z | ✅ **correct** — `0` `password-entropy` hrefs, `5` `ai-agent-roi-estimator` refs (read from the GitHub API, not inferred) |
| **served bytes** | ❌ **STALE** — still 2 `password-entropy` hrefs, `last-modified 15:24:05` |

**The break is git → B2 → edge.** `gh run list --repo gqls/sites`: **27 of the last 60 runs are
`queued`**, oldest **15:12:51Z**, plus two `startup_failure`s. The runner pods are **Running** (3, up
126m) and **idle** — their logs show jobs completing normally until **15:25:52Z** on
`github-actions-runner` and **15:24:13Z** on `-vmsites`, then nothing. **Idle runners with a
27-deep queue is a stall, not capacity.**

And my page's `last-modified 15:24:05` is exactly the last successful deploy round before the stall,
which is the cross-check that the model is right.

**Control that makes this fleet-wide rather than mine:** five pages across three sites
(`leopardess/blog.html`, `loancash` ×2, `aiao/pricing.html`, plus mine) all reported `complete`
between 15:35–15:50; **none has a served artefact newer than 15:37.** ⚠ For four of those I have not
proved the content changed, so "no change ⇒ `b2 sync` uploads nothing ⇒ `last-modified` unchanged" is
a live alternative for them. **It is NOT an alternative for mine**, where the DB and the commit both
demonstrably changed and the served file still does not match the commit.

> **This is the estate's canonical trap in a new place — `complete` is not proof the work happened.**
> Every lane rerendering right now is getting a green work item, a real git commit, and an unchanged
> live site. Anything verified today *at the DB* is fine; anything verified *at the served bytes*
> after ~15:25 is reading a pre-stall artefact.

**NOT touched by me.** Restarting runner deployments is fleet-wide infrastructure with other
sessions' jobs in the queue, and "my canary is stale" is not a reason to bounce it — the same
check-before-flip that was right about `build-pipeline-trigger-2` this morning. **Raised to the owner
and to the lanes actively deploying.**

**What this does NOT change:** the hypothesis is still CONFIRMED — archiving unblocks the recompute,
proven in `content_data`, `rendered_html` **and the git commit**. Three of four stages are green and
the fourth is somebody else's outage. The `[REFUTED]` I wrote 20 minutes ago was wrong twice over:
wrong because I read the body before the `last-modified` header, and wrong again because the failure
is downstream of everything this lane controls.

### 2026-08-31 ~15:0xZ — five days on: **the archive drains the site by itself**, and the counts I handed over were stale by more than half

Picked the lane back up on a fresh chassis (**v1.0.1349**, read from the image tag — the binary grep
is useless here, its all-zeros control matches). **Nobody worked the lane in the interval**, and
`git log --since='2026-08-26'` over the resolver, the rerender action, `cta_label_agreement.go` and
the retraction returns exactly one commit (`b1190467c`, additive). **So the mechanism proven on
08-26 still holds** — worth checking before trusting a five-day-old proof, and cheap.

**⭐ THE FINDING. The remaining field population fell 41 → 25 with ZERO dispatch from this lane.**

Proven rather than inferred: all three fields still pointing at the tool on the two **archived** sites
sit on components whose `updated_at` **predates their site's archive** — `finetuning/ai-guides`
written 08-24 (archived 08-26), `finetuning/guides/llm-cost-calculator-guide` written 08-17,
`leopardess/blog/can-you-trust-ai-with-your-data` written 15:04 against a 15:25 archive **the same
day**, twenty-one minutes before. **Every component rewritten after its archive has already moved off
the tool on its own.** Any rerender, fired for any reason, recomputes the CTA — and archiving is what
releases KEEP #2, so the whole site drains as ordinary churn touches it.

> **⇒ Step 3b never needed a 60-item dispatch.** The expensive half of the plan was the archive
> doing the work for free. I had proven the mechanism on two canaries and then written a handoff that
> still assumed I would have to drive every field by hand — **I proved the release and did not follow
> it to its consequence.** The consequence was five days of free progress I did not predict.

**⚠ AND THE DRAIN CUTS BOTH WAYS, which is the part that matters.** The recompute sends a field to
whichever *tool* ranks first, so for contact-intent copy it manufactures exactly the
leopardess/careers defect. **`[MEASURED 2026-08-31]` one such case has already appeared on the
archived sites since 08-26.** Not hypothetical, not finished, and it will keep happening as pages
rerender.

**Re-measured population** — and note how the shape moved, not just the size:

| site | fields | contact-intent | no label |
|---|---|---|---|
| ai-agent-orchestration.com (**not archived**) | 22 | **19** | 2 |
| finetuning.uk (archived) | 2 | 1 | 0 |
| leopardessconsulting.co.uk (archived) | 1 | 0 | 1 |
| **total** | **25** | **20** | 3 |

**20 of 25 are now contact-intent, against 23 of 41 on 08-26** — the drain cleared the *easy* class
and left the class that needs a decision. **The blocker is now a larger share of a smaller problem**,
which is the honest way to state progress here.

> **This is the "a count carries the date it was counted" rule biting a count I wrote myself.** I
> handed over "41 fields, 23 contact-intent" as the size of the remaining work. Five days later both
> numbers are wrong and the *ratio* is wrong too. The handoff now says, in the checklist, **re-run the
> census — the count moves on its own.**

**Checked and rejected as a shortcut: the `bugs_open/399` audit cannot serve as this census.**
`CTA_LABEL_MISMATCH` is live and recording (**218 findings**, 19 on these three sites) and looks like
the right instrument. `[MEASURED 2026-08-31]` of all 218, **exactly one** is a contact-intent label
pointing at a tool. Verdicts: `no_opinion/ambiguous` 179, `contradicts` 35, `no_opinion/(none)` 4.
The judge is a **page-identity** test and *"Write to leopardess@…"* names no page — so it is blind to
our class, exactly as this lane told them and they documented. ⚠ Note its `context` nests findings
under `context->'findings'`; a flat `context->>'label'` returns empty and reads as "the audit records
nothing useful". I ran that query first and had to open one row to find the shape.

### 2026-08-31 ~16:0xZ — OWNER DECISION TAKEN (contact page), and a finding that splits the population in two

**Owner, 2026-08-31: "we can route the get in touch buttons to the contact page."** That is option 1
of §2 — the cheapest, no LLM, and what the copy already promises.

**Verified BEFORE writing that the repoint is durable, because a one-off that gets recomputed away is
worse than nothing.** `applyCTARecompute`'s **KEEP #1** *is* `bugs_open/248`'s fix (slug
`cta_recompute_clobbers_authored_contact_links`). It keeps — and deliberately **rewrites** — a stored
destination for which `storedCTADestinationIsAuthored()` holds:

| clause | our case |
|---|---|
| `ctaExcludedDestination(url)` | ✅ `contact` is in `areasExcludedFromCTA` (`resolve_internal_links_action.go:86-88`) |
| `validPages.Contains(url)` | ✅ `/contact.html` is 200 on all three sites, per-domain control 404 |
| `NOT CTAMintedCovers(...)` | ✅ none of the 20 fields carries a mint stamp naming `/contact.html` |

So a hand-written `/contact.html` is **exactly the shape KEEP #1 protects** — the supported way to say
*"this button is a contact button"*, not a patch. ⚠ And `__cta_minted` is deliberately **not**
touched: the predicate is url-specific so a stale stamp cannot cover the new url, and the mint map
merges **shallowly** (`cta_provenance.go:111-118`), so writing one field's stamp would replace the
whole record.

### ⭐ THE FINDING: the renderer DROPS a CTA whose destination is archived — so the two populations are not the same problem

Chasing why the canary page's served bytes were **clean while its stored component was dirty**
(`finetuning/guides/llm-cost-calculator-guide`: `content_data` and `rendered_html` both carry
password-entropy, written 08-17; served page last-modified **today 12:30**, **0** refs):

The served `call-to-action` section emits **only the secondary button**. The primary — *"Talk to us"*
→ the archived tool — **is not rendered at all**. One component, no duplicates, `build_status
deployed`. The renderer simply omits a CTA pointing at an archived page.

**⇒ My DB census over-reports the user-visible problem, and the split is the whole story:**

| | stored rows | what a visitor sees |
|---|---|---|
| **archived** sites (finetuning, leopardess) | dirty, and drain as pages rerender | **button absent** — no wrong link, but also no CTA |
| **NOT archived** (ai-agent-orchestration.com) | 19 contact-intent fields | **live and misdirecting** |

**Confirmed at the served bytes on the unarchived site — this is the real damage:**

```
/about.html        <a href="/tools/password-entropy.html" class="about-cta">Learn More About Us</a>
/services.html     <a href="/tools/password-entropy.html" class="cta-btn cta-btn-primary">Book a Technical Discovery Call</a>
/case-studies.html <a href="/tools/password-entropy.html" class="cta-btn cta-btn-primary">Book a Technical Discovery Call</a>
```

A visitor clicking *"Book a Technical Discovery Call"* lands on a password-strength toy. **That is
live now, on the one site this lane deliberately did not archive** — and the caution that kept it
unarchived (archiving would start the wrong-kind drain there) is exactly what left the wrong buttons
serving. Both readings were right; the second one is the one that matters to a visitor, and I had not
weighed them against each other until I looked at the artefact.

> **The instrument lesson, and it is this lane's own rule turned on itself:** I have been reporting
> "25 fields remaining" from a `page_components` census for a week. The number is correct about the
> **database** and wrong about **what is served** — inert on two sites, live on the third. *A census
> of stored state answers "what is stored", and I kept quoting it as "what visitors get".* The
> served-bytes check is three curls and I ran it only when a canary's before-state surprised me.

**Canary dispatched** on `aiao/services.html` — the field is repointed, and the rerender is the test
of KEEP #1 rather than merely a publish: **PASS** = the served button reads *"Book a Technical
Discovery Call"* → `/contact.html`; **FAIL** = password-entropy returns, meaning KEEP #1 did not hold
and the batch must stop.

### 2026-08-31 ~16:1xZ — CANARY PASSED, batch applied to the remaining 19 — and MISSTEP 17: my guard asked about the fleet when the change was 13 pages

**Canary PASSED end-to-end on `aiao/services.html`, both halves:**

1. **KEEP #1 holds in production, not just in my reading.** The `cta_links_stale` recompute ran
   (complete 16:09:07) and left the stored destination at **`/contact.html`** with the label
   unchanged. The recompute had every opportunity to send it back to a tool and did not.
2. **The artefact agrees.** `last-modified 16:09:23` — matching the deploy, checked **before** the
   body this time — and the page serves
   `<a href="/contact.html" class="cta-btn cta-btn-primary">Book a Technical Discovery Call</a>`.
   `password-entropy` hrefs **2 → 1**, the remainder being the **footer** (a `site_component` —
   retirement step 6, not this).

**Batch applied:** 19 fields across **13 pages**, one no-LLM rerender each. Post-condition asserted
inside the transaction: **zero** contact-intent fields still point at the tool.

> **⚠ MISSTEP 17 — the first run of the batch ABORTED, and the guard was wrong, not the change.**
> My mint-stamp guard counted stamps naming `/contact.html` **fleet-wide** and found **49**, so it
> raised and rolled the whole transaction back. Those 49 are pre-existing, correct, and none of my
> business: the resolver legitimately **mints** `/contact.html` through the label match, and a minted
> one is deliberately excluded from KEEP #1 (`storedCTADestinationIsAuthored`'s third clause) so it
> stays re-derivable. **The question I meant was "do the fields I am about to repoint carry such a
> stamp?" — scoped to 13 pages, the answer is 0.**
>
> **This is the same wrong-population error this lane has logged all week**, now inside a guard I
> wrote to protect a batch: a 9-row grouping enumerated as 6, two task rows read for one column,
> three runner pods read as two, a `page_components` census quoted as "what visitors get" — and now a
> fleet-wide predicate standing in for a 13-page one. **The instrument was correct; its population
> was not.** It failed safe, which is the only reason this is a note rather than an incident: a guard
> that aborts on the wrong population costs a re-run, one that passes on the wrong population costs
> the batch.
>
> **The check, stated so it generalises:** a guard's population must be **the set the change
> touches** — join it to the change, do not re-derive it from the world. Mine had a `repointed` temp
> table sitting in the same transaction and I did not join to it.

### 2026-08-31 ~16:4xZ — the batch PUBLISHED and verified; 25 → 5; and MISSTEP 18: my classifier's false-negative tail was only visible after clearing the class

**All 13 rerenders complete, verified at the served bytes** (headers before bodies): **every one of
the 14 repointed pages now serves a `/contact.html` CTA button.** `finetuning/guides/llm-cost-calculator-guide`
is at **0** `password-entropy` hrefs. The aiao pages sit at 1–2, and the residue is accounted for
exactly: **1 per page is the FOOTER** (a `site_component`, present on every page — retirement step 6),
and three pages carry one additional page-level field each.

**Population: 25 → 5.** What remains:

| site | page | field | label |
|---|---|---|---|
| aiao | `/about.html` | `cta_url` | *"Learn More About Us"* |
| aiao | `/case-study-kafka-consumer-group-remediation` | `cta_url` | *(none)* |
| aiao | `/enterprise-reference-deployment` | `cta_url` | *(none)* |
| finetuning | `/ai-guides.html` | `primary_cta_url` | ***"Start a Conversation"*** |
| leopardess | `/blog/can-you-trust-ai-with-your-data` | `cta_url` | *(none)* |

> **⚠ MISSTEP 18 — my contact-intent census under-reported the class by one, and the miss is
> structural.** *"Start a Conversation"* is a get-in-touch button by any reading and falls squarely
> under the owner's ruling. My regex —
> `get in touch|contact|write to|email|call us|book a|talk to|speak to|discovery call` — does not
> match it: **"conversation" does not contain "contact"**. So every count I have quoted for this class
> (23, then 20) was a floor, not a total.
>
> **The instructive part is when it became visible.** While 20 matched fields were in the way, the
> one unmatched field was indistinguishable from the rest of the residue. **Clearing the matched set
> is what made the tail legible** — a classifier's false negatives only surface once its true
> positives are removed. *So the last pass over a population is worth more than the first, and a
> regex census should be re-run after it has been acted on, not only before.*
>
> Repointed under the same ruling (`SQL_2026-08-31_route_start_a_conversation.sql`), with the
> mint-stamp guard **scoped to that page** this time — misstep 17's lesson applied on the next
> transaction rather than only written down.

**The three unlabelled fields need no decision.** With no copy to contradict, the positional pick
cannot produce a kind-mismatch: any relevant tool is a defensible destination, and the `nav_order`
demotion makes the pick sensible. They resolve on their own once `ai-agent-orchestration.com` is
archived — the drain proven in §"the archive drains the site by itself".

**⚠ ONE FIELD IS GENUINELY UNDECIDED and I am not guessing at it:** aiao `/about.html`
`cta_url` = *"Learn More About Us"*. Not contact-intent, so the ruling does not cover it; and the
positional pick would send it to a **tool**, which is the wrong kind for that copy — it reads as a
link to the about page, on the about page. **Flagged for the owner rather than resolved**; it is one
field and it is the only judgement left in the class.
