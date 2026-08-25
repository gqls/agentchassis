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
