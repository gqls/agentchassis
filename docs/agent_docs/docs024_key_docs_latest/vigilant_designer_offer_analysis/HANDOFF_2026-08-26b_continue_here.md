# HANDOFF — vigilant designer + offer analyser (2026-08-26b)

**COLD-START = this file + `HANDOFF_2026-08-26_continue_here.md` (still correct on everything except
§1 and §4, see below) + `PLAN_2026-08-25b` §8 **and §8f** + register `IMG-074` / `WII-033` / `CLM-024`.**

**Supersedes `HANDOFF_2026-08-26_continue_here.md` on TWO points only:**
- **§1 (the imagery fix) is DONE — and its predicate as written there was WRONG.** See §A.
- **§4 (ask the owner) is ANSWERED.** See §B.

Everything else in that file — §2 (gate 1c's unreachable negative control), §3 (supply), §5
(carried-forward work), the whole Watch-outs list, the Residuals — **stands unchanged and still
applies. Read it.** Its "re-run every number before acting" instruction was tested today: every
figure in it re-measured true, and the fleet still moved under me mid-measurement (see §D).

---

## §A — THE IMAGERY FIX SHIPPED, AND IT WAS TWO HALVES, NOT ONE

**Migration `644_planner_sees_imagery_and_illustrated_block_sources_an_illustration.sql`** — applied
2026-08-26, recorded in the ledger, **live now** (DB config, no image, no inert window). Register
**IMG-074**. Commits `d10952b3b` (migration + register) and `b3bddba60` (LANDMINES ×2).
`Council-Submitted: 08477888-b3e6-4ceb-911d-6e2a3c446755` — ⚠ **VERDICT NOT YET READ. That is the
first thing you owe.**

```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
 WHERE correlation_id='08477888-b3e6-4ceb-911d-6e2a3c446755' AND kind='council_report' ORDER BY created_at;
```

A REVISE/REJECTED must be acted on: **the change is already live on the shared branch AND applied to
the database.** There is no inert window to hide in.

**Half 1 — the missing word.** `component_expresses` gained a fifth `image` arm. It fed **three**
live planner menus four tokens and none was an image, so `Generic Text Block` and `Illustrated Text
Block` read identically and the planner picked the plain one (208 instances across 23 sites vs 6 on
one). Not a missing component — a missing WORD.

**Half 2 — the source.** `Illustrated Text Block.image_url` moved off `site_assets.image` (which
**aliases to the page's own hero, unconditionally**) onto `site_assets.illustration`; `image_alt`
moved off it onto `llm` (it was typed `text`, so the resolver handed it the image URL for a screen
reader to read out).

> ### ⚠ THE PREDICATE IN THE PREVIOUS HANDOFF AND IN `PLAN_2026-08-25b` §8c WOULD HAVE CANCELLED THE FIX
> Both propose `source = 'site_assets.image'`, **exact equality**. Half 2 moves the field off that
> value, so the exact predicate makes `Illustrated Text Block` invisible again — the two halves
> cancel, each provably correct alone, and **every guard still passes**. The shipped predicate is over
> `site_assets.%` (excluding `logo`), and 644 asserts BOTH ends so the cancellation cannot commit.
> Full correction: `PLAN_2026-08-25b` **§8f**.

**⚠ It was also preventing live damage, on a clock I did not know about.** apis.uk/index has an
active `hero_home`, so the alias resolved there, and *live resolution beats `carryStored`* — its six
distinct illustrations were one `plan_sections` run from becoming six copies of `hero-home.jpg`. The
apis.uk session confirmed on receipt that the page was at `needs_rebuild` **with a `stale_chrome`
re-render wave imminent**. Hours, not "someday". **Telling the owning lane is what dated the risk;
I could not have.**

## §B — THE OWNER'S ANSWER (2026-08-26)

Both questions from the previous handoff §4 are settled:

1. **"Between paragraphs" — section-level placement is ACCEPTABLE.** He answered *"either, whichever
   ships"*. **So a strictly mid-prose component is NOT owed.** Do not build one on the strength of
   the original wording; that question is closed.
2. **He chose to fix the source before shipping**, over shipping the one-liner or holding for supply.

## §C — WHAT TO DO NEXT, in order

1. **Read the council verdict** (query above). Act on REVISE/REJECTED — it is live and applied.
2. **Verify at the artefact once a planner has actually run.** ⚠ **A zero is NOT failure.** Nothing
   re-plans a page on its own, and a site with no `section/illustration` row renders the block as
   plain prose *by design*. **Read the DEMAND side first:**
   ```sql
   SELECT count(*) FROM orchestration_states
    WHERE owner_agent_type IN ('build-site-planner','site-planner','content-gap-planner')
      AND created_at > '2026-08-26 11:00+00';
   ```
   Only then ask whether any planner chose `Illustrated Text Block`. **apis.uk is the regression test
   case** — its six values must still be six distinct illustrations after its next rebuild.
3. **THE SUPPLY QUESTION — this is now the whole of the imagery ask, and nobody owns it.** §1 created
   no assets. `[MEASURED 2026-08-26]` the pipeline generates **heroes** (206 active across 28 sites)
   and **icons** (139 across 19) and barely any **illustrations** (26 across 5 sites; only **4**
   `section/illustration` plan rows across 3 sites). **This is the real answer to "why don't pages
   have pictures in them" — not placement, generation.** Bigger than today's work. Say the supply
   figure alongside any imagery claim or you repeat this lane's own `bugs_open/395` shape.
4. Everything in the previous handoff §5 (v2 batch, `features_open/034`, the `HandlerCanWriteField`
   drift audit) is **untouched and still carried forward.**

**No SUMMARY written today, deliberately.** Half the imagery ask shipped and the bigger half (supply)
is unowned, so "where we are now" is mid-stream — the rarity rule says wait for the turn rather than
file a near-duplicate of `SUMMARY_2026-08-25`. **Write one when supply is decided**; that is the
inflection.

## §D — NEW WATCH-OUTS (the previous handoff's list all still stands)

- **⚠ A CHANGED-ROW COUNT CANNOT TELL A WIDENING FROM A RESHUFFLE.** A variant arm that also
  suppressed `list` changed **the same 9 rows** while 3 silently lost a capability. Assert the SHAPE
  (nothing loses a token; nothing changes but by gaining the new one) and **induce both**. Now a
  LANDMINE.
- **⚠ DO NOT ASSERT A LITERAL COMPONENT COUNT in a migration guard, and do not take BEFORE/AFTER as
  two snapshots.** Five components were created by another lane *during this change's measurement
  run*; the two-snapshot control broke outright (381 vs 386). Compute both sides in ONE query. 644's
  guards are structural and population-independent for this reason.
- **⚠ `site_assets.<path>` IS ALIAS-RESOLVED, NOT A LITERAL KEY.** `image`, `background`, `banner`,
  `header_image`, `product_screenshot`, `screenshot` and more all mean **`hero`**. Nothing in a
  component's schema, template or guidance says so. LANDMINE entry has the check.
- **⚠ A REAL-LOOKING PER-SECTION VALUE IN `content_data` IS NOT PROOF THE SOURCE WORKS.** It may be
  hand-seeded and surviving only by `carryStored`. This is how six rows of live data appeared to
  refute a correct code reading — `WRONG_CALLS.md` 2026-08-26e. **When a code reading and a data
  sample disagree, widen the population until the sample could have come out either way.**
- **⚠ A CONTROL THAT ERRORS AND A CONTROL THAT PASSES ARE THE SAME NUMBER** if you only read the
  number. An `awk` guard here failed to parse and printed `0`.
- **⚠ `run-migrations.sh` (even a bare dry run) TIMES OUT** — it probes all 1,052 files. Read
  `schema_migrations` directly, and note lanes apply their own file by hand and out of order (642
  landed before 636), then `--record-only`.
- **⚠ `component_expresses` HAS NO GO CONSUMERS** — `grep` finds nothing. Its three consumers are
  embedded in `agent_definitions` config; the `jsonb_path_query` to find them is in the RUNBOOK.

## §E — RESIDUALS FROM TODAY, stated plainly

1. **Supply (§C3) is unowned** and is the larger half of the owner's actual ask.
2. **`section/illustration` resolution is FIRST-WINS BY KIND**, so several illustrated sections on one
   page all resolve to the SAME image. apis.uk has routed around it (`content_data` + lock) and has
   **offered itself as the worked test case** — six distinct instances — if anyone builds per-section
   mapping.
3. **6 `site_plan_imagery` rows at `scope='page', kind='illustration'` are read by NO resolver arm**
   and are inert (5 apis.uk, 1 pool-energy-utilities.internal). Named, not fixed.
4. **llm-authored alt text for a server-resolved image is a hallucination surface** — a model
   describing a picture it cannot see. True of all 13 existing alt fields, the estate's settled
   convention; this change brought the one outlier INTO it rather than overturning it.
5. **The heroes-included judgement is arguable and is deliberately on the record** — `hero` and
   `product-hero_pre_037` now advertise `image`. Only `site_assets.logo` is excluded. A council seat
   may reasonably want banners excluded too; the submission says so in terms.
6. **Every behavioural claim in IMG-074 is unverified** — nothing has re-planned a page yet.

## §F — WHO OWNS WHAT NEARBY (changed since yesterday)

**`bugs_open/381` is CLOSED and its lane wrapped up 2026-08-25** — so `component_expresses` has **no
live owner**; the register relation in IMG-074 is now the durable channel, not a session.
**`apis.uk`** holds the only live instances and is actively engaged (they verified at both ends).
**`loanzy_uk_example_site`** — corrected their own ADDENDUM 1 revalidation claim today; I checked
`WII-033` at their request and it does **not** repeat it, so no correction was owed on this side.
Their fix (`recordModeSilenceRule`, council `04a3ce1f`) is Go, inert until a roll — **until then the
operative truth is that verdict rows are cleared by humans only.**
**`portfolio_positioning`** — their uncommitted SEO-007 LANDMINES edit rode into `b3bddba60` as a
named same-file passenger; they were told, nothing lost.
