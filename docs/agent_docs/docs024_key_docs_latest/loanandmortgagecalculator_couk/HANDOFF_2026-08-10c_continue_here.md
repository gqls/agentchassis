# HANDOFF — decompose both sites so the framework controls the components. START HERE.

**Written 2026-08-10 (evening) by the orientation session.** This is the single
entry point for this lane. It supersedes **both** earlier 08-10 handoffs:

- `HANDOFF_2026-08-10_continue_here.md` (11:57) — bugfix 224/225 and the acceptance
  fences. Still accurate about that work; it is *finished* work, not a task list.
- `HANDOFF_2026-08-10_unlock_and_upgrade.md` (13:24) — migration 367. Part 1
  accurate; **its §4 blocker claim is stale — see §2 below before you act on it.**

**Owner instruction, 2026-08-10 (evening):** *"I'd like the sites to have their
components undergo decomposition so that we can fully control them from the
framework."*

Everything in §1 was measured by this session against the live DB and the running
pods. Nothing here is carried forward from a doc unchecked.

> **⚠ CORRECTED 2026-08-10 (~19:40Z), and it moves work between the tracks —
> migration 377.** Six of the pages 367 unlocked **carry calculators**, and were
> sitting `generic` + `needs_rebuild` + verbatim with an open `page_rerender` item
> when this file was written. They are back to `owned`. So the counts below are
> wrong as printed: **Track A is 17 pages, not 23**, and **Track B is 22, not 16**.
> The six are `/loans/compare-loans.html`, `/loans/interest-rate-stress-test.html`,
> `/loans/loan-vs-savings.html`, `/loans/settlement-calculator.html`,
> `/loans/damage-checker.html`, `/mortgages/fact-finder.html` — all Track B now.
> **What caught it:** cross-checking the unlocked set against `decompose_lmc.py`'s
> hand-authored `CALCULATOR_URLS` instead of trusting 367's classifier. 367 read
> `onclick=|addEventListener` only; these six bind `oninput=`/`onsubmit=`/
> `onchange=` and two load the shared `calculators.js`, whose listeners are in the
> external file, not in the stored HTML. **367's negative control used the same
> expression as its filter**, so it agreed with it by construction. Full working:
> §2b below and `LANDMINES.md`.

---

## 0. The task in one paragraph

**57 of the 59 pages across the two sites are still a single verbatim blob** — one
`page_components` row, `slot_name='ported-page'`, holding the whole page. The
framework cannot control a component it cannot see, so decomposition is the work:
split each page into real prose components plus (where there is one) a protected
tool component. Two pages are already done and are the worked examples. The 39
pages migration 367 "unlocked" are unlocked and **still verbatim** — unlocking
granted permission, not structure, and those are two different things.

---

## 1. State, measured 2026-08-10 evening

Chassis **v1.0.1280**, both replicas, started `2026-08-10T15:45:06Z`
(newer than the v1.0.1277 the 13:24 handoff was written against).

### Page shape — the number that defines the job

```sql
SELECT s.domain, p.rebuild_policy,
       CASE WHEN p.sections::text = '["ported-page"]'
            THEN 'verbatim single ported-page' ELSE 'decomposed / other' END AS shape,
       count(*)
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE p.site_id IN ('ed633ada-f8af-424b-b4d4-8af79160dbcd',
                    'ee4a8199-4f5b-4e2e-88ce-01e600721b74')
GROUP BY 1,2,3 ORDER BY 1,2,3;
```

| domain | policy | shape | as printed 18:00 | after mig 377 (19:40) |
|---|---|---|---|---|
| loanandmortgagecalculator.co.uk | generic | verbatim | ~~23~~ | **17** |
| loanandmortgagecalculator.co.uk | generic | decomposed | 1 | 1 |
| loanandmortgagecalculator.co.uk | owned | verbatim | ~~16~~ | **22** |
| loanandmortgagecalculator.co.uk | owned | decomposed | 1 | 1 |
| loancash.co.uk | generic | verbatim | **15** | 15 |
| loancash.co.uk | owned | verbatim | **3** | 3 |

**Still 57 to decompose — but the split moved: 32 prose + 25 tool** (was stated as
38 + 19). The total is unchanged because 377 changed no page's shape, only which
side of the safe/unsafe line six of them are counted on. The two already done:

- `loans-consolidation` — `["prose-0", "tool-1", "prose-2"]`, `owned`, tool row
  `lock_type='permanent'`. **The target shape for a tool page.**
- `guide-how-loans-affect-mortgage-affordability` — `["prose-0"]`, `generic`, no
  lock. **The target shape for a prose page**, and it is already there, which makes
  it the cheapest canary on the estate.

### Other live facts

- **`site_plans` = 0 rows for both sites.** Unchanged.
- **Exactly one locked `page_components` row across both sites** (consolidation's
  tool). Everything else is unlocked.
- Unlock state per migration 367 (applied `2026-08-10 12:22:05Z`): LMC 24 generic /
  17 owned, loancash 15 generic / 3 owned. Matches the 13:24 handoff exactly.

### ⚠ Page names are HYPHENATED, not slashed

`pages.name` is `loans-consolidation`, `mortgages-stamp-duty`,
`tool-price-cap-checker`. The 13:24 handoff's §3 table lists them as
`loans/consolidation` etc. **A query using the slashed form returns 0 rows and looks
like an empty result, not a typo** — it cost this session its first query. The
slashed form is the URL, not the key.

---

## 2. ⛔ CORRECTION — the blocker the last handoff was written around is GONE

The 13:24 handoff §4 says:

> *"`bugs_open/204` is exactly the trap here — `plan_sections` resolves a section by
> NAME/FUNCTION only, so a decomposed page can never be rebuilt … Read 204 before
> seeding."*

**That was already false when it was written.** Both 204 and its sibling
`bugs_open/189` were fixed, rolled and behaviourally verified on **2026-08-06**, four
days earlier. They are still in `bugs_open/` because of the owner direction of
2026-08-06 (*"please leave the bugs that you've found in bugs_open not in the closed
bug file"*), which overrides CLAUDE.md's bar — **so a file's presence in
`bugs_open/` is not evidence that the bug is open.** Read the file's tail, not its
directory.

Re-verified by this session at the current binary, **not inherited from the record**:

| instrument | m9fbr | swzhc |
|---|---|---|
| `stored_slot_name` (189) | 1 | 1 |
| `load page slot identities` (204) | 1 | 1 |
| `slot_name repeats with different component_ids` (204) | 1 | 1 |
| nonsense-string negative control | 0 | — |

Config half also still live: `page-content-writer`'s active `default_config`
contains `slot_name_from`. (Check it with a `::text LIKE` test, **not** a
`jsonb_path_query` — this session's path read returned `[]` on a key that is
present.)

**Two consequences, both of which change the plan:**

1. **A decomposed page CAN now be rebuilt from a plan.** The build path resolves
   sections by `page_components.component_id` first and falls back to name, so
   positional slot names (`prose-0`) resolve. Decomposition is no longer a one-way
   door out of the generic pipeline.
2. **189's standing prohibition is lifted.** It read *"never fire a build-path run
   on a page holding locked rows — it will duplicate them."* That was the correct
   rule until 08-06 and is the exact shape this work creates (prose rows + a locked
   tool row). Both halves are fixed; the build half was proven by 204's canary
   persisting `prose-0`/`prose-1` rather than renaming them to `ported-prose`.

**Do not take this section on trust either.** It is four days old as of writing and
the config half is live-mutable by any session. Re-run the pod-greps and the config
check at the top of your session — they are three commands and they are in §5.

---

## 2b. ⛔ SIX LIVE CALCULATORS WERE UNPROTECTED FOR ~7 HOURS — migration 377

Found 2026-08-10 ~19:30Z by the session that picked this file up, in the first ten
minutes, by cross-checking 367's "prose" set against `decompose_lmc.py`'s
`CALCULATOR_URLS`. Six of the 24 pages 367 unlocked are calculators.

**Why 367 missed them.** It classified with
`bool_or(rendered_html ~ 'onclick=|addEventListener')`. These six bind handlers as
`oninput=` / `onsubmit=` / `onchange=`, and `compare-loans` +
`interest-rate-stress-test` also load `/assets/js/calculators.js` — the
`addEventListener` calls are in the **external file**, not in the stored HTML the
detector read. A working calculator can carry neither string in `rendered_html`.

**Why its negative control did not catch it — the transferable half.** 367's control
asserted "17 LMC + 3 loancash tool pages are still `owned`" **using the same
expression as its filter**. It was deliberate, it was induced, and it fired on the
induction. It was still blind to exactly what the filter was blind to. A control
that shares its subject's classifier cannot disconfirm that classifier; inducing it
proves the `RAISE` works, not that the population was right.

**Why this was live damage waiting, not untidiness.** At the moment of the finding,
all six were simultaneously:

| condition | value | why it matters |
|---|---|---|
| `rebuild_policy` | `generic` | `get_pages_to_build`'s only ownership filter is `COALESCE(rebuild_policy,'generic') <> 'owned'` |
| `build_status` | `needs_rebuild` | that is what the selector selects on |
| `sections` | `["ported-page"]` | one verbatim row, calculator `<script>` inline — the shape 367 refused |
| open work item | `page_rerender:detected` | live demand touching the page |

and the generic full-rebuild path **had already run at these pages**: on 2026-08-09,
`needs_page:loans-compare-loans` and 19 siblings reached step `save_sections` and
died there with *"page loans-compare-loans is rebuild_policy=owned (tool/widget-owned):
a generic section save would clobber it … Refusing to overwrite."* That refusal is the
only reason those calculators still exist, and 367 removed it for six of them at
12:22Z. `attempt_count=1, max_attempts=3` on those items.

**One inherited claim did NOT survive checking, and it matters for Track B.** Both
earlier handoffs and RUNBOOK §14 say the composition loop *"commits LLM-written HTML
to the sites repo one step BEFORE the DB guard refuses"*. On this path it did not:
`git log` over the sites repo for 2026-08-08 20:00 → 2026-08-09 03:00 shows only the
`bugs_open/224` fix and a consolidation rerender — **no clobbering commit**, though
20 runs reached `save_sections`. So on the `page-build-handler` path the DB guard
fired before anything reached the repo. `[MEASURED for this path and this window
only.]` Do **not** relax anything on the strength of it: the guard is what saved the
pages, the ordering claim may still hold on the other two composition loops, and it
was never the plan to find out on a live calculator.

**Fixed by** `docs/agent_docs/sql_for_agents/377_relock_six_verbatim_tool_pages_missed_by_367.sql`
(applied by hand + `--record-only`; `_mig377_relocked_tool_pages` stamps the six so
the rollback is exact). Its detector ORs three independent spellings — handler
attributes/listeners/`calculators.js`, form controls, `getElementById|querySelector`
— and over all 38 generic verbatim pages on the two sites **the six match all three
and the other 32 match none**. Assertion 1 checks the stamped set against
`CALCULATOR_URLS`, a source that has never read the SQL. Both controls were induced
before applying: the expected-set check aborted on a claimed seventh page, and the
over-locking control aborted at 16/14 when two prose pages were swept in.

**The induction also found a flaw in the new check itself:** `pages.url` is not
unique across these two sites (both have `/guides/jargon-buster.html` and
`/legal.html`), so a url-only set assertion is not exact — a deliberate over-lock of
"one" page stamped **two** rows and the assertion called neither unexpected. It now
keys on `domain || '|' || url`. Review did not catch that; running the induction did.

## 3. What decomposition does and does not buy

State this plainly to the owner, because the previous pass had to correct a similar
overstatement:

- **It does buy** component-level control today. `page-rerender` and
  `section-editor` / `apply_section_edit` operate per component, so after
  decomposition you can rewrite one prose block without touching the calculator.
  This works on `owned` pages too — migration 164 deliberately leaves re-assembly of
  existing `page_components` un-gated, *"it is how owned pages deploy"*.
- **It does buy** protection. A locked tool row survives a rebuild that would
  otherwise replace the calculator with prose.
- **It does NOT by itself buy** wholesale rebuild-from-plan. That needs
  `rebuild_policy='generic'` **and** a `site_plan`, and there are **0 plan rows**.
  Decomposing 57 pages and seeding no plan leaves you with fine-grained editing and
  no automated rebuild — which may well be what you want. Treat the plan as a
  separate, later decision with its own risks, not as step 5 of this one.

---

## 4. The plan — three tracks, in this order

The ordering is deliberate: it puts the tooling through its paces on pages with no
calculator before it goes near one.

### Track A — LMC prose, 17 pages · LOW RISK · do this first

Already `generic`, no calculator, and `decompose_lmc.py` already handles all 41 LMC
pages (it was written for this site and its assertion suite was refuted and repaired
on real pages here). No flip needed — these pages are already unlocked.

Per page: decompose → `load_lmc.py --check` → `--apply` → assemble-only rerender →
diff against the prediction. RUNBOOK §12 is the command sequence.

**Start with `guide-how-loans-affect-mortgage-affordability`** — it is *already*
decomposed and generic, so use it to prove the rerender/edit loop end to end before
converting anything. It is the only page on the estate where a mistake costs nothing.

### Track B — LMC tool pages, 22 pages · HIGH RISK · one at a time, check between

The six that migration 377 moved here from Track A are the *easiest* of the 22 in one
respect — `compare-loans`, `interest-rate-stress-test`, `loan-vs-savings` and
`settlement-calculator` are class A/B tools already covered by `oracle.py`, so the
per-page arithmetic check after decomposition is a single command. `damage-checker`
and `fact-finder` are class C (no external right answer) and want `invariants.py`.

`loans-consolidation` is the worked example of the finished shape. For each page:

1. **Decompose** while the page is still `owned`. Decomposition writes rows; it does
   not invoke the generic pipeline, so this step is safe at any policy.
2. **Lock the tool row** — `lock_type='permanent'`. This is what makes the
   lock-aware DELETE in `save_page_sections_action.go:757`
   (`pageComponentAgentWritableSQL`) spare it.
3. **Confirm the tool slot is named in the incoming composition** — see the trap in
   §6, which is a real precondition but not the one the last handoff described.
4. **Then** flip that page to `'generic'`, as a migration with a `DO`/`RAISE` verify
   block (RUNBOOK §14).
5. **Re-run the arithmetic checks on that page** before starting the next one
   (§5). The failure mode here is silent and the blast radius is a live
   consumer-finance calculator.

⛔ **Never flip a page that is still a single verbatim `ported-page` row.**
`assemble_page → deploy_page(git_commit) → save_sections` commits LLM-written HTML to
the sites repo **one step before** the DB guard refuses, so the calculator is
replaced by prose and shipped live. `rebuild_policy='owned'` is the only thing
preventing that, and it is why 367 refused these 20 pages. This has not changed and
is not affected by §2.

### Track C — loancash.co.uk, 18 pages · NEW TOOLING REQUIRED

loancash has **no decomposition tooling at all** — `decompose_lmc.py` is written
against this site's chrome and asserts against it per page. Expect to port it, not
to run it. 15 prose pages (already `generic`) and 3 tool pages (`owned`), all 18
verbatim.

Do Track C **after** A and B: the LMC work will tell you which of
`decompose_lmc.py`'s assertions are site-general and which are LMC-specific, and
that is much cheaper to learn on a site whose tooling already works.

---

## 5. Commands

```bash
LANE=docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk
CHASSIS=$(kubectl get pods -n ai-persona-system -o name | grep agent-chassis | head -1 | cut -d/ -f2)

# ── run these THREE FIRST, every session — §2 depends on them and they are live-mutable
kubectl exec -n ai-persona-system $CHASSIS -- sh -c \
  'strings /app/agent-chassis | grep -c "stored_slot_name"; \
   strings /app/agent-chassis | grep -c "load page slot identities"; \
   strings /app/agent-chassis | grep -c "zzz_cannot_exist"'     # expect 1, 1, 0
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -c \
  "SELECT default_config::text LIKE '%slot_name_from%' FROM agent_definitions
   WHERE type='page-content-writer' AND is_active AND COALESCE(is_snapshot,false)=false
   AND deleted_at IS NULL;"                                      # expect t

# ── decompose / re-voice (RUNBOOK §12 has the full sequence and its gotchas)
python3 $LANE/decompose_lmc.py $DECOMP_WORK/manifest.json --verbose
DECOMP_WORK=... python3 $LANE/load_lmc.py --check <page>          # prediction only
DECOMP_WORK=... python3 $LANE/load_lmc.py --apply <page>          # CHANGES A LIVE PAGE
DECOMP_WORK=... python3 $LANE/load_lmc.py --restore <page>        # per-page rollback

# ── the calculator must still compute (Track B, after every page)
cd $LANE && python3 oracle.py --tools <the-tool-you-touched>
python3 oracle.py --selftest-parse && python3 oracle.py --mutate expectation --tools simple
#   ^ ALWAYS in the same session, or a green oracle run is not evidence

# ── after ANY repo edit on this site
python3 $LANE/gate_component_bytes.py --repair && python3 $LANE/gate_component_bytes.py
```

Site ids: LMC `ed633ada-f8af-424b-b4d4-8af79160dbcd`, loancash
`ee4a8199-4f5b-4e2e-88ce-01e600721b74`. Sites repo `/home/ant/projects/sites` → push
→ GH Actions → B2; wait ~100 s past the run or B2 serves a `NoSuchKey` blob at HTTP
200 that greps clean.

---

## 6. The re-slot trap — sharper than the last handoff stated it

The 13:24 handoff and RUNBOOK §14 both say: *"a positional `tool-1` never matches,
and `save_page_sections_action.go:855` moves the unmatched locked row to
`len(sections)+1` — the calculator lands at the BOTTOM of the page."*

**The mechanism is real; the trigger condition as stated is wrong.** Reading
`matchLockedRow` (`save_page_sections_action.go:1043`): it matches the locked row's
`slot` against **the incoming section name** — exact first, then kebab-normalised.
And `loans-consolidation`'s `pages.sections` is already `["prose-0","tool-1",
"prose-2"]`. So an incoming composition that carries `tool-1` **does** match, exactly,
on the first branch.

The trap fires when the incoming composition **omits the tool slot** — which is
precisely what a *seeded site plan* would do if it describes the page in semantic
section names. So:

> **The precondition is "the composition the writer is handed must name the tool
> slot", not "positional names never match".** If you seed a plan, the plan must
> carry the tool slot. If you never seed a plan and only rerender from
> `pages.sections`, the slot is already there and this trap does not arise.

`[INFERRED — read from `matchLockedRow` + the live `pages.sections` row. NOT measured
end to end; no writer run has been driven against a decomposed live page with a
locked row.]` **Measure it on ONE page before Track B goes wide.** The disconfirming
result would be a locked `tool-1` landing at `position = len(sections)+1` with a
`lock_blocked` work item raised — check for both after the first page, because the
work item is the only thing that is not silent.

---

## 7. Other traps that cost real time (all in LANDMINES)

- **`gate_component_bytes.py --repair` would have destroyed a decomposed page** —
  it compared every row against the whole repo file. Fixed: verbatim +
  single-component only. Relevant again the moment Track A starts producing
  multi-row pages.
- **§2 of RUNBOOK (`build_site.py`) is DEAD for any page that has been decomposed.**
  The DB is the render source from that moment; rebuilding from the build scripts
  and pushing would be overwritten by the next rerender, and the two fight silently.
- **`toolgolden --compare` reports RED on every decomposed calculator and the page
  is fine** — the golden fingerprints `id="content"`, which the old wrapper carried
  and the new one does not. Use `golden_compare_post.py`, and run its `--self-test`
  first or the comparator is inert.
- **A page that is still verbatim ignores `site_components` entirely**, so chrome can
  be broken, rerender 41 pages, report success and change nothing. Check the chrome
  RESOLVES, never that rows exist.
- **Fetch the served page ~90 s after the item says `complete`.** Inside the deploy
  window B2 returns a 7-line `NoSuchKey` JSON at HTTP 200 and every grep against it
  returns 0, which reads as a clean pass. Guard on byte count and a leading
  `<!DOCTYPE`.
- **Use a tab field separator with psql on this site** — every page title contains
  `" | LoanAndMortgageCalculator.co.uk"`, so the default `|` splits inside the data
  and reads as truncation.
- **On this estate, assume a red result from a checker you wrote today is the
  checker.** Five times in the 08-08/09 pass the red was the harness, not the site
  (all in `WRONG_CALLS.md`). Print the inputs it drove and the raw value it compared
  before believing it.

---

## 8. Not part of this task, but still owed

- **loancash.co.uk has no arithmetic oracle**, and two of its three tools hardcode
  dated FCA caps (0.8%/day, £15 default fee, 100% total cost) with nothing checking
  them against CONC 5A. Same shape as the SDLT bug (`bugs_open/225`), which was a tax
  rule 16 months out of date and under-quoted by £5,000. **Verify against the FCA's
  own source, not the page.** Highest-value unstarted item on the estate; independent
  of decomposition.
- **Fleet sweep for the 224 defect class** — *a guard that leaves a handler without
  writing the DOM*. `mortgagecalculator.co.uk` and `loancash.co.uk` share this
  family's ancestry and have never been checked. `grep -n "return;"` per calculator
  page, read three lines up; `alert(...); return;` is the highest-yield spelling.
- **Six LMC pages are fence-eligible but have no fence** (application-tracker,
  credit-health-check, damage-checker, fact-finder, investor, portfolio). Class C —
  no external right answer — so they want INVARIANT checks, not arithmetic.

## 9. Read in this order if you are starting fresh

1. this file
2. `RUNBOOK` §12 (decompose/re-voice), §13 (arithmetic checks), §14 (unlock)
3. `PLAN_2026-08-05_voice_rebuild_and_decomposition.md` — why decomposition works
   the way it does here
4. the tail of `bugs_open/189` and `bugs_open/204` — **the tails, not the titles**
5. `HANDOFF_2026-08-10_unlock_and_upgrade.md` §1–3 for the lock mechanism, with §2
   of this file held against its §4

**Council:** decomposition itself is site content, DB rows and lane tooling — out of
the gate's `platform/ internal/ pkg/` scope. If Track C ends up changing shared Go,
that half goes through the gate.
