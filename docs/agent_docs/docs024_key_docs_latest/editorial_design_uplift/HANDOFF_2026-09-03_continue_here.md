# HANDOFF 2026-09-03 — editorial_design_uplift, continue here

**Supersedes `HANDOFF_2026-09-02_continue_here.md`.** That file is still worth reading ONCE, for its
§2–§4 (the migration 686 rollback, the real imagery finding, the planner-prompt answer) — none of
which changed today. This file replaces its §0, §5, §7 and §9.

**Branch:** `087_towards_multiple_domains`. Evidence is in `NOTES_editorial_design_uplift.md`,
2026-09-03 entry. Plain-prose account for the owner: `README_where_we_are.md`, same date.

---

## 0. Environment — state at handoff, 2026-09-03 ~12:45Z

1. **Kubeconfig is ALIVE.** The 09-02 expiry is resolved; `kubectl` works.
2. **The chassis is `v1.0.1359`** (pods up **13:28:18Z / 13:28:43Z**). THREE rolls happened across
   this session: `v1.0.1356` (08:57Z), `v1.0.1358` (12:06Z), `v1.0.1359` (13:28Z). **1356 and 1358
   are both stale — do not quote either.**
3. ~~⚠ `v1.0.1358` DOES NOT CARRY THIS SESSION'S CHANGE~~ **SUPERSEDED — `v1.0.1359` DOES, and the
   pre-registered probe in §4a confirmed it.** The 1358 gap is kept below because it is what made the
   confirmation decisive rather than merely reassuring:

   | UTC | event |
   |---|---|
   | 12:06:45Z | `3fa8a604c` "release v1.0.1358" |
   | 12:06:47Z | chassis pods start on `v1.0.1358` |
   | **12:09:56Z** | **`1007be27d` — this session's wiring commit** (missed 1358 by 3m 09s) |
   | 13:28:18Z | chassis pods start on `v1.0.1359` — **carries it** |

   **So 035 P1 direction 2 is LIVE IN THE FLEET as of `v1.0.1359`.**
   ⚠ **This session ran ~9 hours of wall clock** (12:09Z first code commit → 21:2xZ close), most of
   it idle between turns. If you are reading timestamps in this file that look implausibly spread
   apart, that is why — **there is no clock skew**: host and DB `now()` were re-checked against each
   other and agreed to one second.
4. **The migration dry-run owed after this roll HAS BEEN RUN** (`[MEASURED 2026-09-03 ~12:55Z]`),
   so the next session does not owe it again unless another roll lands:

   | reading | `v1.0.1356` (08:57Z) | `v1.0.1358` (12:06Z) | `v1.0.1359` (13:28Z) |
   |---|---|---|---|
   | `Pending (N)` | 177 | 180 | **183** |
   | `LIKELY ALREADY APPLIED; its own guard raised` | 36 | 37 | **36** |
   | `probe inconclusive` | — | 41 | **40** |

   The **already-applied** column is `bugs_open/426`'s figure — files applied by hand and never
   recorded, i.e. the replay hazard — and it is still climbing (34 on 09-02, 36 then 37 today).
   ⚠ The dry-run takes **over five minutes** and, piped through `tail`, prints nothing until it
   finishes — which reads exactly like a hang. Run it unpiped, in background. Both counts here come
   from the same method (`grep -c` over the run's own output), so they are comparable with each other
   and with 09-02's.
5. **`708_enable_unrendered_page_imagery.sql` is APPLIED**, so IMG-077 is no longer inert. The 114
   lane discharged it (`21e18a504`) — the 09-02 handoff's §5 and §9 say otherwise and are stale on
   that line only.
6. **⚠ A TIMEZONE TRAP THAT NEARLY INVERTED §0.3, and it is one command away from you.**
   `git log --date=format-local:'...Z'` renders the time in **BST** and lets you write a literal `Z`
   after it, so `1007be27d` printed as `13:09:56Z` when it is **12:09:56Z**. Here it inflated a
   three-minute miss into a comfortable-looking hour — the conclusion survived, the magnitude did
   not. **Reverse the roll and the commit and the same error flips the answer outright**, which is
   why it is worth the paragraph: this comparison decides "is my fix in the fleet", and it is exactly
   the comparison a one-hour offset is large enough to invert. Another lane logged the identical
   shape today (`8fc91d4ff`: *"my timestamp was BST written as UTC"*), so it is not a one-off.
   **The safe form is `%cI` (ISO-8601 WITH the offset) piped through `date -u -d`** — let `date(1)`
   do the conversion, never your own arithmetic:
   ```bash
   date -u -d "$(git log -1 --format=%cI <sha>)" +'%Y-%m-%dT%H:%M:%SZ'
   ```
   `kubectl`'s `.status.startTime` is already UTC, so the two are then directly comparable.

---

## 1. What happened today, in one paragraph

The imagery half of this lane is parked exactly where 09-02 left it (see that handoff's §2–§4 —
**do not restart imagery work at the component layer**). Today was the structure half: **035 P1
direction 2 is wired**. `recomposeAncestors` — written 08-31, reviewed over three council rounds,
committed, and never called — now runs, is guarded, is stamped, and has tests. The reason it was
never called turned out to be the finding (§2).

**Status in one line: on the branch, council-APPROVED, not yet in the fleet.** The change missed the
`v1.0.1358` roll by three minutes (§0.3) and ships on the next one. It is inert either way — 0 of
3,229 rows are parented — so nothing is waiting on that roll except the confirming probe in §4a.

---

## 2. THE FINDING: a parameter no caller could supply, justified by a paragraph that measured nothing

`recomposeAncestors` took `tx *sql.Tx`. Its own header explained why in capitals — *"THE db/tx SPLIT
IS FORCED, not stylistic"* — reasoning about which reads must see *"the uncommitted edit"* inside
`apply_section_edit`'s transaction.

`[MEASURED 2026-09-03]` **There is no transaction:**

```
grep -nE 'BeginTx|\.Begin\(|Commit\(\)|Rollback\(\)' platform/orchestration/actions/section_editor_actions.go
→ (no matches)
```

Every persist there runs through `updatePageComponentAfterEdit(ctx, params.DB, …)` on the autocommit
connection. **No call could compile.** The function sat uncalled for three days, Go's linker dropped
it, and the 09-02 binary probe reported the symbol ABSENT — which reads exactly like a missing
commit. (§9 of that handoff got the interpretation right; it just could not see the cause.)

**Three council rounds reviewed the design and none reached it. One grep did.** A comment can assert
a fact about its caller and nothing type-checks it. The claim was specific and technical — it
correctly named two functions that cannot take a `*sql.Tx` — and inferred a transaction from that,
which does not follow.

**Attempting the wiring then found three more defects, none of them visible in a design review**,
because they live one level below the seam the rounds argued about. The ancestor write:
carried **neither the tombstone nor the lock predicate** its sibling writes on that path carry;
read **zero rows affected as success**; and was **unstamped** though it writes a column the 357/552
triggers archive (`bugs_open/355` A1). All three fixed.

⚠ **The half with no detector, worth inheriting:** an ABSENT guard is invisible to this estate's
existing checks. `TestNoHandSpelledTombstonePredicate` catches a *wrong spelling* of the tombstone
predicate and never a *missing* one; `page_component_writer_coverage_test.go` asks only about the
floors. So a green suite says nothing about whether a NEW writer of `page_components.rendered_html`
is guarded — list the predicates on a sibling write and match them deliberately.

Recorded in `LANDMINES.md` (footprinted, verifier dispatched), `WRONG_CALLS.md`, and
`features_open/035` §5.

---

## 3. State of 035 P1 — the audit, so nobody re-derives it

| deliverable | state |
|---|---|
| `deriveRenderMode` third value (`composite`) | **DONE** `1f745e730` |
| membership helpers | **DONE** `bc8167100`, reached in production |
| direction 1 — refuse to render a composition parent alone | **DONE** `028c3e112` |
| flat-pass extraction | **DONE** `2a0bdb001`, `94f81cc60`, `22ed53ee7` |
| direction 2 — `recomposeAncestors`, **called** | **DONE today** `1007be27d` |
| `check_render_mode` routing arm | **REFUTED, not deferred** (`5542a76d6`) — nothing reads `render_mode`; P1's routing story cannot work as 035 wrote it |
| the walk in both render paths, §6.9's filter, register entry, live canary | **NOT DONE — this is what remains** |

**Still inert, re-measured today:** **0 of 3,229** `page_components` rows carry a
`parent_instance_id` `[MEASURED 2026-09-03]` (0 of 2,249 on 08-31; 0 of 2,005 on 08-24 — the table
grew ~1,000 rows in ten days and the parented count stayed at zero). Direction 2 adds **one indexed
SELECT per edit** to the live edit path and nothing else.

---

## 4. What to do next, in order

1. ~~Read the council verdict~~ **DONE — `cab931b1-8b45-461e-8a37-0dbdfa6aa928` came back APPROVED**
   (11 reviewers, 6 abstained, *"approved with 6 advisory objection(s) — none high-severity"*), was
   read, and was acted on in commit `3ba94508c` (trailer `Council-Reviewed:`). **One objection was a
   real defect I had argued the wrong way** — `writeRecomposedAncestor` returned `true` when
   `RowsAffected()` itself errored, which re-opens the door the row count exists to close; it now
   fails closed (case + mutation M5). A second commit-worthy pair (guardian: does
   `pageComponentAgentWritableSQL("")` degrade to always-true? reuse_agent: can two guarded-write
   styles drift?) is answered by one predicate-PARITY test asserting the rendered predicate is
   byte-identical in both writers and is not a tautology (mutation M6). Full adjudication of all six
   in `3ba94508c`'s message.
   ⚠ **One re-measurement the librarian seat's challenge forced, and it found a STALE count of my
   own:** `content_components` is **554** rows, not the **386** this lane and `features_open/035`
   were both quoting from 08-31. The inertness claim survives and is larger — **0 of 554** declare a
   `slots` key, **0 of 554** carry `render_mode='composite'` `[MEASURED 2026-09-03]`. Stale by
   ADDITION, exactly the class the dated-count rule exists for.
2. **THE READ PATH — the remaining core of P1.** `walkComponentHierarchy` still has no production
   caller, so a row that opted in today would render flat. Its own council round.
3. **Hazard 6.9's filter MUST land inside that change, not after it.** `loadStoredSections` selects
   `COALESCE(parent_instance_id::text,'')` but its WHERE is only `page_id = $1 AND <not removed>`, so
   every row comes back flat. The moment the walk renders children in a nested pass, children are in
   BOTH lists, every later section's `NextOccurrence` index shifts, and per-section figures attach to
   the wrong sections — rendering, deploying and looking correct. Read 035 §6.9 in full first: it also
   carries the `MergeLockedPageSlots` inverted-polarity trap for any plan-vs-live guard.
4. **Then the register entry and the live canary**, which are P1's actual acceptance.

   ⚠ **TWO CONDITIONS BIND THAT SUBMISSION, both named by seats that called them follow-ups rather
   than blocks. Do not let the read path land without them:**
   - **`stale_ancestor_slots` must be wired into something that FILES.** It is computed today and
     consumed by nothing — `bug_historian` (medium) named that as this platform's single most
     repeated failure shape (`bugs_open/083`, `/071`, 016b §9: *"a finding that outlives the state it
     describes is indistinguishable from a live one"*). It is harmless while the mechanism is inert;
     it becomes the incident the moment the read path makes composed pages real.
   - **Probe the DEPLOYED BINARY after the next roll** — **the BEFORE half is already run; see §4a
     for the table, the expectation and the exact command.** `recomposeAncestors` reads **absent** on
     `v1.0.1358` (correctly — that image predates the wiring by three minutes) and must read
     **PRESENT** on the first roll that carries `1007be27d`.
5. **`news-listing`** — same defect `article-body` had, still unwritten, still deliberately behind the
   09-02 handoff's §3 question.
6. **Do NOT restart imagery work at the component layer.** Unchanged from 09-02 §7.4; the live
   question is the planner's page composition and it belongs to `bugs_open/114`.

---

## 4a. The binary probe — the BEFORE half is DONE, and it is what makes the AFTER half decisive

`debug_historian` asked for the instrument that found the bug to be the one that confirms the fix.
Half of that is already run, on `v1.0.1358` `[MEASURED 2026-09-03 ~12:45Z]`, pod
`agent-chassis-554857f96f-kx69c`:

| symbol | v1.0.1358 | role |
|---|---|---|
| `PlanSectionsAction` | **PRESENT** | must-be-present control ✓ |
| `zzzInventedControl_NotInAnyBinary` | absent | must-be-absent control ✓ |
| `hierarchyChildrenOf` | **PRESENT** | **feature-specific** present control — direction 1's helper, called since `028c3e112` (08-31) |
| `recomposeAncestors` | **absent** | the subject: still uncalled in this build, so the linker drops it |

**The third row is the one that earns the table.** Generic controls prove the probe works; they do
not prove it can see THIS feature. `hierarchyChildrenOf` reading PRESENT rules out "the probe simply
cannot find hierarchy symbols", which is the alternative explanation a bare `recomposeAncestors:
absent` would leave open — and that is precisely the ambiguity the 09-02 probe had to write a
paragraph to resolve.

**So the AFTER half is now a one-bit test with a pre-registered expectation:** after the next roll,
`recomposeAncestors` must read **PRESENT**. If it does not, the roll did not carry `1007be27d` — not
that the wiring failed.

> ### ✅ DISCHARGED — the AFTER half ran on `v1.0.1359` and came back exactly as pre-registered
>
> `[MEASURED 2026-09-03 ~21:22Z]`, pod `agent-chassis-85c4984f77-nrqf7`:
>
> | symbol | v1.0.1358 | **v1.0.1359** | |
> |---|---|---|---|
> | `PlanSectionsAction` | PRESENT | **PRESENT** | control holds |
> | `zzzInventedControl_NotInAnyBinary` | absent | **absent** | control holds |
> | `hierarchyChildrenOf` | PRESENT | **PRESENT** | unchanged, as expected |
> | `recomposeAncestors` | absent | **PRESENT** | ← **the only symbol that moved** |
>
> **One bit changed, and it is the bit the expectation named.** Both controls held across the pair
> and the feature-specific control did not move either, so the reading cannot be explained by "the
> probe behaves differently on this image". `debug_historian`'s condition (corr `cab931b1`) is
> **satisfied**: the instrument that diagnosed the absence has confirmed the presence.
>
> **035 P1 direction 2 is live in the fleet.** It remains inert on live data — 0 of 3,229 rows carry
> a `parent_instance_id` — so "live" means reachable, not exercised. The first composed page is what
> exercises it, and that is still the read path's job.

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
for SYM in PlanSectionsAction zzzInventedControl_NotInAnyBinary recomposeAncestors hierarchyChildrenOf; do
  timeout 220 kubectl -n ai-persona-system exec "$POD" -- grep -aq "$SYM" /proc/1/exe
  RC=$?; case $RC in 0) echo "$SYM : PRESENT";; 1) echo "$SYM : absent";;
    124) echo "$SYM : *** TIMED OUT — NOT a negative, re-run ***";; *) echo "$SYM : error rc=$RC — NOT a negative";; esac
done
```

⚠ **Two things that make this probe lie if you shortcut them, both already paid for.** Each grep
needs **100–120s** (BusyBox, whole binary), so **a grep killed by a command timeout is
indistinguishable from a negative** — hence the explicit `rc=124` arm above, which is the difference
between a measurement and a guess. And **never `strings`**: it is absent from these debian-slim
images, and behind the customary `2>/dev/null` its failure looks exactly like "not stamped"
(`LANDMINES.md`).

⚠ **A roll is not evidence, and a peer lane re-confirmed that today**, independently, in
`e9274c1fa`: *"a roll happened, the tag advanced, my commit is an ancestor of HEAD — and the binary
does not carry it."* Ancestry proves the code is in the branch; only the artefact proves it is in
the fleet.

---

## 5. ⚠ HEAD is RED in a neighbouring package, and it is not this lane's

`go test ./platform/orchestration/...` fails
`discovery_checks/TestStylesheetGutted_TokenSetMatchesCanonicalCSSTokens`:
*"canonicalCSSTokens declares 4 token(s) this check does not police: [--color-accent-ink
--color-accent-text --color-cta-bg-ink --color-primary-ink]"*. Neither file is dirty, so it fails at
committed HEAD; `git log` names `0325ddebb` (2026-09-03 12:10, the 458 lane), which added the tokens
without extending `rendererGuaranteedTokens`. **Left alone deliberately** — the test message tells
that lane exactly what to do — but know it before you read your own `go test` output, and scope your
runs to `./platform/orchestration/actions/` if you want an unambiguous green.

**Still red when re-checked at 12:50Z**, four hours after it landed. ⚠ **This is a dated
observation, not a standing fact** — re-run it rather than repeating it, because the one thing this
handoff cannot tell you is whether that lane has since fixed it:
```bash
go test ./platform/orchestration/actions/discovery_checks/ -run TestStylesheetGutted_TokenSetMatchesCanonicalCSSTokens -count=1
```

---

## 5b. Both fleet-wide append-only files were swept by peers within two minutes — nothing lost

Recorded because it is the shared-tree trap CLAUDE.md documents, observed twice in one hour, and
because it means **the prescribed check does not survive a busy tree.** Today's LANDMINES entry went
into HEAD inside `4ffe30f0f` (the 427 lane) and the WRONG_CALLS row inside `e76129bf0` (the 450
lane) — both same-file passengers, both while this lane's own commits were being prepared. Both are
in HEAD and complete.

⚠ **The consequence for a reader:** commit `c9b95eb44`'s message lists a WRONG_CALLS row that the
commit does not contain, because the file was swept between `git diff --numstat` (which showed 39
added lines) and `git commit`, seconds later. `git log --follow` on the FILE finds the content;
`git log` on this lane's commits does not. **So resolve an append-only doc by grepping the file, not
by trusting the commit that claims it** — and note that "check numstat immediately before
committing", the remedy WRONG_CALLS itself prescribes for this class, is a race on a tree this busy
rather than a guarantee. Forward-only: not amended, corrected here.

---

## 6. Identifiers

**This session's five commits, in order:**

| sha | what |
|---|---|
| `1007be27d` | the wiring + guards + stamp + tests (`Council-Submitted:`) |
| `c9b95eb44` | lane docs: NOTES, README_where_we_are, this handoff, `features_open/035` |
| `6eb930e6a` | handoff §5b — correcting `c9b95eb44`'s message forward |
| `3ba94508c` | the council-driven fixes: fail-closed `RowsAffected`, predicate-parity test (`Council-Reviewed:`) |
| `c0a20cb7a` | the verdict record: adjudication of all six advisories, corrected `554` count |

- council corr `cab931b1-8b45-461e-8a37-0dbdfa6aa928` — **APPROVED**, and the orchestration itself
  reached `complete_approved: COMPLETED`, so the run terminated in the approved branch rather than
  merely leaving an artefact behind
- submission JSON: `editorial_design_uplift/COUNCIL_SUBMISSION_2026-09-03_035_p1_direction2_wiring.json`
- new writer stamp `action:recompose_ancestors`; new test file
  `platform/orchestration/actions/component_hierarchy_recompose_test.go` — **six** mutations recorded
  in its header (M1–M4 at authoring, M5–M6 after the council round), all killed
- `verify-head-builds.sh --test ./platform/orchestration/actions/` **green at HEAD `8fc91d4ff`**,
  i.e. against committed HEAD carrying all five commits above
- everything from the 09-02 handoff's §8 still stands: boxingonline site `d2aa5206-73bc-4707-a69c-2702c1eb9152`
  serving at `boxingonline.ugg2.com`; `article-body` `5835b2e1-50d7-4f20-8a9c-8da4d270ae3d` at md5
  `002cbcd9cada6a37bf4a5158fd1e5f22`; planner definition `f263eaa1-61e1-446e-9410-648e12b7875b`

---

## 7. NEW INBOUND, and this lane owns half of it: finetuning.uk homepage infographics

Arrived 2026-09-03 evening from the `finetuning` lane, relaying the owner: *"tidy up the components
and use more interesting ones for the cards, probably different carousel like structures… be
imaginative, research good alternatives and apply them"*, plus a 22:25 addendum, *"including
infographics wherever they will help the understanding of the concepts"*.

**The split, proposed by this lane and ACCEPTED by theirs:**

| half | owner |
|---|---|
| choosing + applying card/carousel components on `index.html` | **finetuning lane** |
| **infographics for that page** | **this lane** |
| the missing swap mechanism (no item type changes a slot's component) | **this lane** |

Their `design_critique_run` is filed (item `204f1ff7`, `design-critique-agent`, `triaged`) and its
report is the research input for both halves.

**Why the card half is not ours, stated because the routing looked wrong at first glance:** this
lane's PLAN scope sentence ("how the page family LOOKS") is fixed one paragraph above to the
**editorial** family. A marketing homepage is not that family. But nobody else owns it either —
`site-design-planner` does palette/layout/typography, the experience loop judges rather than applies,
`Staged component build` is CLOSED — so it is unowned work on a page they own, which makes them the
applier by elimination, not by territory.

**Slots scoped for imagery, `[MEASURED 2026-09-03]` at the live rows:**

| pos | `slot_name` | scoped? | why |
|---|---|---|---|
| 1 | `hero` | no | chrome |
| 2 | `features` | **YES** | "what fine-tuning is" — a concept diagram, **no quantities**, so nothing to source |
| 3 | `differentiators` | **YES** | the £99 vs ~$5,000 comparison — the strongest graphic on the page |
| 4 | `case-studies-grid` | **no, deliberately** | their canary slot; staying off it makes collision impossible |
| 5 | `departments-grid` | no | a taxonomy, not a quantity — an infographic there is decoration |
| 6 | `call-to-action` | no | chrome |

⚠ **`slot_name` is `differentiators`, NOT `differentiators-section`** (which is the component name).
Their brief had the component name in it — in the brief that warns against by-name matching. Resolve
by function.

**The sourcing question, and the answer that constrains the drawing.** Both numbers ARE registered:
the site's current `evidence_base` (10 facts, updated 2026-08-26) holds `ft-price-99` (99,
`tolerance: exact`) and `ft-market-anchor` (5000, **`tolerance: approximate`**, attested to a web
sweep finding $5k–$180k and a cheapest productised ~$4,800). **So the comparison is drawable — and
the `approximate` tolerance is a design constraint, not a footnote:** a crisp bar end presents as
precise a number the registry itself declines to call precise. Band it, or label it "from ~$5,000";
the £99 side may be crisp, and that asymmetry is the honest picture.

**⚠ I asserted the opposite of that first, and it was a truncated query, not a finding.** I told the
peer the site had no evidence base at all, on a list of "twelve current aspects" — the site has
**26**, my `| tail -12` ate the first fourteen, and `ORDER BY 1` had sorted `evidence_base` to the
top, so the one row the question was about was the first casualty. Retracted within the hour, and it
is in `WRONG_CALLS.md` and `LANDMINES.md` (footprint `site_specs`, `tail`, `evidence_base`). **Count
first, then list — a count cannot be truncated by a pipe.**

**Blocked on two things, neither of them technical:** the owner has not yet chosen between the narrow
route (hand-author this one page's `site_plan_imagery` rows at `kind='infographic'`) and the
fleet-wide planner-prompt change (§2 of the CONTRIB — planner owners' call, 18 remakes behind it);
and **this session does not have its user's authorisation to make live writes on finetuning.uk**. A
peer accepting a split is not that authorisation. Both are named to the peer, who is relaying the
route choice to the owner.

**THE CRITIQUE LANDED (21:38Z) and its two imagery findings are MEASURED, not accepted:**

1. **Hero reuse is 7× what the critic could see.** It named five pages; the site has **35 of 38 hero
   components on `/assets/images/hero.jpg`** and **2 distinct hero images in total**, across 58
   component-bearing pages `[MEASURED 2026-09-03]`. The critic was right in direction and bounded by
   its instrument — `design_critique_run` captures ≤8 pages × 2 viewports, so a site-wide claim in
   its prose is a **sample statistic**, and nothing marks which claims are sampled. Worth knowing
   before quoting any figure out of one of those reports.
2. **10 pages already hold a generated, deployed hero that nothing renders.** IMG-077 fired on this
   site today: 4 `unwired` + 6 `no_image_slot`, both items at `needs_human_review`. **So the cheapest
   imagery win here is WIRING, not generation.** ⚠ The six `no_image_slot` pages are the
   `article-body` shape — do NOT fix them by giving the component its own image field without first
   checking for a hero component on the page; that is migration **686**, applied and rolled back,
   because it renders the same image twice on 292 of 301 pages fleet-wide.
3. **The orange-left-border device stays.** The critic calls it "the strongest section of the site",
   so the comparison graphic sits WITH it, not instead of it — a crisp £99 against a **banded**
   $5,000, the banding forced by `ft-market-anchor`'s `approximate` tolerance.

⚠ **A stale detection that would have misdirected the peer's canary.** 11 `image_url_404` rows sit at
status `detected` on this site, dated 2026-07-26 → 2026-08-03, naming `/assets/images/case-study-*.jpg`
— the exact slot being canaried. **Probed before relaying: they return 200, and an invented-URL
control returns 404**, so the domain is not a catch-all and the 200s are real. The images are fine;
the ROWS are the defect. A month-old `detected` row is not evidence of current state.

**Papers:** the ask and its constraints —
`editorial_design_uplift/CONTRIB_2026-09-03_from_finetuning_owner_asks_for_more_imaginative_card_structures_on_the_homepage.md`;
this lane's answer — `finetuning_uk_service/CONTRIB_2026-09-03_from_editorial_design_uplift_answer_on_the_homepage_cards_and_infographics.md`
(commit `a85bcedea`).
