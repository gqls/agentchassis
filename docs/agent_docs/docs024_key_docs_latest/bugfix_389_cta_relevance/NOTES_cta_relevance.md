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
