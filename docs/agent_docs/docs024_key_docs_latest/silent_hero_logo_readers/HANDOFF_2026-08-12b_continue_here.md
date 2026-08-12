# HANDOFF — 2026-08-12b — continue here

**Lane:** `silent_hero_logo_readers` — commission item 2 (owner ruling 2026-08-10, *"2. yes."*), plus
the diagnosis-harness defect that item 2's investigation turned up.
**Supersedes** `HANDOFF_2026-08-12_continue_here.md`, which is closed at its head with two
corrections you should read if you touch its §3/§4.
**Read first:** `bugs_open/261` (the whole finding, with the verification recipe), then
`NOTES_…` 08-12 entries (three missteps, one of them a false claim I published), then `RUNBOOK_…`
for the commands.

---

## 1. State in one paragraph

**Two things are done and neither is live.** Item 2 (durable rows when a deployed image arrives
without a usable URL) is council-APPROVED, committed and **live on `agent-chassis:v1.0.1290`** — but
its *behavioural* proof has not happened, because nothing has deployed a hero or logo since the roll.
`bugs_open/261` (the diagnosis code tier could not resolve the symbol spellings its own index
produces) is found, measured, fixed, **council-APPROVED in six minutes** and committed as
`6911c2da4` — but it is Go, so it is **inert until the next fleet roll**, and the defect is still
costing runs right now. **Nothing in this lane is blocked on me.** The one thing that would move it
is a roll, which I cannot cause; the owner decisions in §5 are unchanged.

## 2. What shipped since the last handoff

| commit | what |
|---|---|
| `6911c2da4` | **the fix** — `splitReceiver` + `spanOf` searches `Values`, plus the regression test. Trailer `Council-Submitted:` (predated the verdict) |
| `dfb7ffbab` | `bugs_open/261`, the LANDMINES correction + new entry, `WRONG_CALLS.md`, and two corrections to my own claims in `6911c2da4` |
| `c22910260` | the council round read and answered; standing five brought up to date |
| *(this commit)* | the milestone summary, this handoff, and the predecessor closed |

⚠ **The LANDMINES edits are at HEAD but inside `d878aa8f7`** — another session's *finetuning.uk
audit* swept them between my write and my commit. Nothing lost; noted so nobody hunts for them in
mine.

## 3. The finding in six lines, because it generalises

The code index writes a method as `(*SagaCoordinator).applyResponseToState`
(`code_symbols_actions.go:598`). `analysis.spanOf` split on the last dot and compared the **raw**
`(*SagaCoordinator)` against `receiverType()`'s `SagaCoordinator` — never a match. All 1,170 indexed
methods were unreadable; so were 1,238 package-level `var`/`const`, which `spanOf` never searched at
all. The bundle then printed *"body unavailable — a TOOLING failure"* and the verdict's
cite-or-abstain rule acted on the absence. **335 lost bodies across 47 runs, all-history.**

**The two parts worth carrying to another lane:**

- **Our own documentation taught the failing input.** The LANDMINES entry added 08-11 tells a
  `code_request` author to *"Name it `(*Receiver).Method`"* — right for the index query, wrong for the
  body read, and nothing distinguished the two paths. Corrected in place.
- **The test asserted a spelling no producer emits** (`Type.Method`). Test and code were blind in the
  same direction and agreed with each other. That is why it survived from the beginning.

## 4. RECOMMENDED NEXT MOVE — and it needs a roll first

**Do not fire a third `090` at `bugs_open/236` before the fix is live.** It will fail identically;
that is two runs' credits already spent learning the same thing.

**After the next fleet roll**, in this order:

1. Confirm the fix is actually running, at the artefact:
   ```bash
   kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
   git merge-base --is-ancestor 6911c2da4 <the stamp>
   ```
   ⚠ startup line, it scrolls — an empty result means "not in range", not "unstamped". Per SERVICE,
   not per fleet.
2. Re-run the `090` on `236` §5's question. Its scope names three methods, so it exercises exactly
   what was broken.
3. **Verify POSITIVELY.** The pass condition is a bundle that **renders**
   `(*SagaCoordinator).applyResponseToState`'s body (4,746 chars) — *not* a falling failure count,
   which also falls if nobody asked. `261` §6 has the queries.

## 5. Decisions waiting on the owner

**None are blocked on me.** Unchanged from the predecessor except where noted.

1. **Which commission item next.** Item 1 (large, mostly investigation, design decision reserved to
   the owner) or item 3 (medium, spans three layers, needs a routing call). **The §4 code-tier work
   that was previously my recommendation is DONE**, so this choice is now clean.
2. **Item 3's routing:** council gate or architecture RFC? It changes a client return signature and
   adds a field to a shared adapter response payload. The commission: *"if in doubt write the RFC —
   the cost is one document."*
3. **Item 3's modelling question:** `deploy_commit` is per-component, but a page is many components
   across possibly several commits. Never answered whether the page level wants it too.
4. **Item 1's design decision** (Design 1 vs Design 2) is explicitly reserved. The census's *"0
   breaks"* premise for Design 2 is contradicted by production — **re-measure the baseline before
   scoring either.**

## 6. Item 2's own outstanding verification — carried forward, unchanged

| check | result | date |
|---|---|---|
| live on the chassis | **YES** — `v1.0.1290`, literal in `/proc/1/exe` on both replicas, control absent | 08-12 |
| rows with the new `error_code` | **0** | 08-12 |
| **DEMAND CONTROL** — `hero_deployed` / `logo_deployed` anywhere | **0 / 0** of 6,364 retained | 08-12 |

**The zero is unfalsifiable, not reassuring.** Per `bugs_open/236` §3, do not read it as an incidence
rate in either direction. The proof needs a site build that deploys a hero or logo, then the
RUNBOOK's paired query **within four hours** (`AWAITING_RESPONSES` is pruned on that clock — measured
first-hand off the live `scheduled_tasks.pre_query`).

## 7. Traps this lane paid for — the new ones only

The predecessor's list still stands. These are from 08-12:

- **A NAME IS NOT A KIND.** I classified 20 failing symbols by reading them, called all 20
  package-level values, and wrote *"not ONE was genuinely absent"* into a council submission **and a
  commit message** before checking. The query returned 19. The honest figure is 334/335 and the fix
  does not cover the twentieth (index staleness). Both statements are immutable; only a correction
  elsewhere can catch a reader. **The sentence existed to be the disconfirmable one, and I published
  it before letting it disconfirm me.**
- **A number you checked is a number you checked at that instant.** `260` was free when I checked and
  taken by the time I saved. The fix's commit and three source comments cite `260` and mean `261`.
  Resolve by slug; `git log` the file path.
- **A figure is a snapshot.** The census moved 321→335 failures in forty minutes. Re-run it.
- **Read ALL the iterations, not the one the verdict cited.** The bundle does not accumulate, and
  generalising from the last one is what produced the predecessor's wrong mechanism.
- **Prove a resolver defect in a `git archive HEAD` checkout.** Both the fail-then-pass and the
  mutation ran there, so the shared tree never held a mutation — the direct answer to this lane's
  `038211dd8` scare. It cost nothing. RUNBOOK §3-4.

## 8. If you change this code

`splitReceiver` widens the accepted **spelling**; it must never widen the **match**. Strip everything
before the dot and `(*Nope).Greet` silently returns another type's method. Four negative controls
guard that, and they were proven capable of failing by mutation — if you touch `spanOf`, re-run that
mutation rather than trusting the green suite.

**If you add a symbol KIND to the indexer, `spanOf` must learn it in the same commit** — that is the
drift that produced both halves of this bug, and it has now happened three times
(`Values` missed by `spanOf`, `Values` missed by `knownScopeIdentities`, and the receiver spelling).
The LANDMINES entry added today is the prospective version of that warning.
