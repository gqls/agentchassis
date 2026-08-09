# HANDOFF — loancalculator.co.uk · **the COPY / VOICE thread** · continue here (2026-08-09b)

> ⚠ **THERE ARE TWO LIVE HANDOFFS IN THIS DIRECTORY DATED 08-09 AND THEY ARE DIFFERENT JOBS.**
> This one is the **copy/voice** thread — the site's words. The other,
> `HANDOFF_2026-08-09_continue_here.md`, is the **`bugs_open/227` platform** thread
> (experience-planner writes another site's plan), and it is **untouched and still owed**.
> Do not read either as superseding the other. They diverged on 08-08 night when two
> concurrent sessions each wrote a handoff 14 minutes apart, both claiming to supersede
> `HANDOFF_2026-08-08`; reconciling that cost this session its first half hour.

**Supersedes `HANDOFF_2026-08-08b_continue_here.md`** — its §2 (the one thing owed) is
**CLOSED**, and its §3 carries a **correction you must read before writing any rewrite
guidance**. Its §4 (CSS trap), §5 (guidance lineage) and §6 (verification) still stand.

```
site      loancalculator.co.uk   0162cde4-633e-45e9-8ca6-87a6b2fe1d26
chassis   v1.0.1270  (pods rolled 08:49Z 08-09; both 08-08 handoffs say 1269 — stale)
voice     26 of 26 active pages. Nothing is left in the old voice.
live      26/26 serving · toolgolden 11/11 exact · 17 locked rows (12 tool + 4 CSS + 1 owner)
the site  DONE. Nothing is owed on the words.
```

## 1. What closed today, and the one number that changed

`index`/`prose-2` — the last line in the register the owner struck — was rewritten through
the framework (corr `26648f55`). It now reads:

> *"Enter your loan amount, rate and term below to see how the monthly figure and the total
> cost move together."*

Verified: struck register **0** on the wire across four spellings; the owner's approved
opening still appears **exactly once**; the four sibling rows **byte-identical** (id,
`updated_at` **and** content hash); 26/26 serving; toolgolden 11/11 exact.

I also swept all 26 pages across eight spellings for the whole struck claim family before
firing, rather than trusting "it is one line". One other hit —
`guide-car-finance-explained`'s *"understanding exactly what you're signing"* — is about
the reader's own contract, not a boast about our arithmetic. It stands.

`index`/`prose-0` is now **permanently locked** (`loancalculator_owner_approved_20260809`).
It is the owner's personally approved copy and this lane destroyed it once already.

## 2. ⚠ READ THIS BEFORE WRITING ANY REWRITE GUIDANCE — §3's remedy is NOT sufficient

`HANDOFF_2026-08-08b` §3 says: write guidance per-SECTION, conditionally. **I did exactly
that and it leaked anyway.** Comparing what the writer *proposed* against the pre-run
backup:

```
prose-0   1102 -> 1102   byte-identical   obeyed
prose-4   2813 -> 2813   byte-identical   obeyed
prose-1    133 ->  400   LEAKED — kept its <h1>, appended its own copy of prose-2's job
prose-2    117 ->  143   the intended change
```

**Why:** my condition read *"the introduction under the 'Standard Loan Calculator'
heading"* — and `prose-1` **is** that heading, so it read the instruction as addressed to
itself. **A conditional whose condition names a neighbouring landmark is ambiguous to the
neighbour.** Make the condition decidable from the section's own bytes (quote the sentence
to be replaced), and ship **no verbatim replacement copy** — a block of approved copy is
portable and any section can paste it, which is how 08-08 broke three sections at once.

**The rule now: phrase it conditionally AND LOCK EVERY SIBLING you are not targeting**,
leaving exactly one agent-writable row. That is what held. Full method, with the gotchas:
`RUNBOOK` → *"Targeted single-section rewrite: lock the siblings"*.

Two traps inside that method, both of which will read as success:
- **`locked_at` is the load-bearing column, not `locked_by`.** A row with a `locked_by`
  string and NULL `locked_at` is fully writable and looks locked in every listing.
- **A `lock_blocked_change` item is NOT evidence the copy differed.** It fires on any
  incoming section matching a locked slot and records no proposed content, so it is
  guaranteed non-zero by your own locking. The real check is `llm_call_log.response_text`
  against a **pre-run backup table** — which means taking the backup is mandatory, not
  tidy. Both filed in `LANDMINES.md`.

## 3. What is NOT owed, so nobody reopens it

- **The site's words are finished.** 26/26 in voice H, all serving.
- **The expanded copy stays** — owner ruling 08-08 evening. Decided, not "probably".
- **`debt-help-uk` leads with the free charities** — owner ruling, executed, nine facts
  checked on the wire.
- **The homepage opening is the owner's approved copy** and is now locked.
- **`bugs_open/227` is not this thread's job** and must not be started here — see §5.

## 4. What is actually next for THIS thread

### 4a. The fleet-wide base-prompt change — chosen by the owner 08-05, still not started

This is the real remaining work. The owner chose the **wide** option. The week's evidence
has changed its shape — see `fleet_copy_quality/SUMMARY_2026-08-08…`.

**[MEASURED 2026-08-09] The house voice block is SEVEN copies across seven live agents,
and they have drifted.** Confirmed today with a nesting-proof census:

```sql
SELECT type FROM agent_definitions
 WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
   AND default_config::text ILIKE '%size of the fact%' ORDER BY type;
```
→ `content-creator-about`, `content-creator-hero`, `content-creator-hero-without-research`,
`content-writer`, `grounded-explainer`, **`page-content-writer`**,
`simple-content-writer-with-approval`.

⚠ **Do NOT census prompts by iterating `default_config->'workflow'->'steps'`.** That is
the first query anyone writes for this job and it is **blind to loop-nested prompts**: it
returns **6** and silently omits `page-content-writer`, whose 12,813-char prompt lives at
`config->'sub_workflow'->'steps'->'generate_content'->'config'->'prompt_template'`. I made
exactly that mistake this morning and nearly wrote it up as a correction to a figure that
was right. Landmine filed; use the text search above.

**The finding that matters for scoping it:** `page-content-writer` — the agent this lane
has spent a week driving, and the one that wrote every page on this site — **is one of the
seven**. So this lane's evidence is directly in scope for the change, which strengthens the
wide option rather than arguing for a narrow one.

**Before designing it, read §2 above.** A base-prompt change is a rule applied uniformly by
every section that can believe it qualifies — which is this lane's one recurring failure,
now five instances deep. The NOTES 08-08 two-arm experiment is the other essential input:
the platform default house voice produced **softer** claims than this site's own H voice
spec, so "add more rules" is not obviously the fix.

### 4b. The sibling finance sites — still held on the owner's review

If they proceed, they need **guidance v3's lineage, not the spec** (`HANDOFF_2026-08-08b`
§5: editing `site_specs.content_direction` changes nothing — the prompt is copied from a
pinned work item), and each site's own incumbent `writing_rules` re-read for conflicts
first.

- `mortgagecalculator.co.uk` — has its own active lane and a newer cold start:
  `docs/agent_docs/docs024_key_docs_latest/mortgagecalculator_couk_adoption/HANDOFF_2026-08-08c_continue_here.md`
  ⚠ That lane is doing **arithmetic/adoption** work, not voice. Check with it before
  sending a rewrite at the site.
- `loancash.co.uk` — lane exists (`loancash_couk/`) with NOTES + RUNBOOK but **no handoff**.
- `lendzy` — named in `HANDOFF_2026-08-08b` §7, but **[UNVERIFIED] there is no lendzy
  directory under `docs024_key_docs_latest/`**. Confirm the site exists before planning it.

## 5. The other live thread in this same directory — do not merge them

**`bugs_open/227` — experience-planner hardcodes one site's diagnosis, so every other site
gets that site's plan.** Found by this lane, owned by this lane, and **entirely separate
from the copy work.** Its fix is **written and dry-run proven but NOT applied**.

Continue it from:
`docs/agent_docs/docs024_key_docs_latest/loancalculator_couk/HANDOFF_2026-08-09_continue_here.md`

## 6. Verification, every time

```bash
LANE=docs/agent_docs/docs024_key_docs_latest/loancalculator_couk
$LANE/check_site_serving.sh          # 26/26: HTTP 200 AND >=2000 B AND starts <!DOCTYPE
python3 $LANE/toolgolden.py --compare $LANE/acceptance/GOLDEN_2026-08-08_voice_h_complete.json \
  https://loancalculator.co.uk/index.html \
  https://loancalculator.co.uk/tools/{overpayment-calculator,consolidation,compare-loans,car-finance-calculator,loan-vs-savings,settlement-calculator,interest-rate-stress-test,credit-health-check,damage-checker,application-tracker}.html
```

**Never grep a served page without the size + DOCTYPE guard first** — a deploy-window fetch
returns a B2 error blob at HTTP 200 and every grep against it reads clean. That guard is
now baked into `check_site_serving.sh`, so use it rather than a bare `curl`.

To fire a rewrite: `./voiceh_rewrite_v3.sh <page>`, or for a one-off,
`SRC_ITEM=<uuid> KEY_PREFIX=<slug> SUMMARY='<what it is>' ./voiceh_rewrite_v3.sh <page>`
(added 08-09 so a one-off reuses the dispatch path instead of becoming a v4 copy).
⚠ **Grade and CLOSE items before re-firing** — `detected` is not in `idx_swi_dedup`'s
excluded list, so a re-fire fails on duplicate key until you close them.
⚠ **Assert the fired item carries YOUR prompt** — a mistyped `SRC_ITEM` silently gives you
a whole-page voice rewrite instead of your targeted one.
