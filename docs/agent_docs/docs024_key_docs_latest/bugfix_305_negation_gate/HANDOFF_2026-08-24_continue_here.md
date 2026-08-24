# HANDOFF 2026-08-24 — continue here (`bugfix_305_negation_gate`)

**Supersedes `docs/agent_docs/docs024_key_docs_latest/bugfix_305_negation_gate/HANDOFF_2026-08-23_continue_here.md`.**
Read that one for the council rounds and how the ceiling was found; read this one for state.

> ## ▶ ONE-LINE STATE
> **`305` is CLOSED** →
> `bugs_closed/305_HANDOFF_2026-08-18_v2_voice_does_not_suppress_define_by_negation.md`.
> Every defect this lane owned is fixed, LIVE and **demand-proven** (re-probed on `v1.0.1333`).
> Nothing here is inert. The residuals are filed and owned elsewhere: ONE page needing a rerender
> that is already queued behind another lane's claims work, plus decisions `D2`–`D5`.
> **`D3` must still NOT be decided** — see §4.

---

## 1. Verified today `[MEASURED 2026-08-24 ~10:35Z]`

| thing | state | how it was checked |
|---|---|---|
| §26 accounting fix | **LIVE** | binary probe BOTH replicas: `no_answer_for_target` **1**, control `no_answer_for_targez` **0**, known-present `rewrite_negations` **8** |
| §27 ceiling fix (`569`) | **LIVE + DEMAND-PROVEN** | **124 calls at `max_tokens=16000`, `cut = 0`** (was 4/34 = 11.8%). Zero `repair_unavailable` since the apply. 6-, 7- and 8-target pages now `repaired` |
| post-roll reconciliation | **22/22 markers reconcile** | 0 not-reconciling, 0 account-for-none, 0 over-counted — see §2 for the honest limit |
| the three pages | **repairable 6 → 2** | shipped-scanner canary, §3 |
| **the demand control §2 flagged as MISSING** | **SATISFIED 14:00Z** — `no_answer_for_target` fired **43** times | see §2a; the gap closed itself in four hours |
| fix present on the 13:12Z roll (`v1.0.1333`) | **re-probed, still live** | `no_answer_for_target` 1, control 0, known-present 8 |
| council `f3046f0c` / `4829bd48` | both **APPROVED** | `unreadable=0`, all seats voted, on both |

## 2. ⚠ THE ONE THING THAT IS *NOT* PROVEN, AND IT IS A DEMAND-CONTROL GAP

The three-era split (bug file §28a) shows something neither bug file predicted — **the two defects
were causally linked**:

| era | markers | non-reconciling |
|---|---|---|
| A — old ceiling, old accounting | 37 | **40.5%** |
| B — NEW ceiling only | 137 | **15.3%** |
| C — new ceiling + new accounting | 22 | **0.0%** |

Raising the ceiling **alone** took 40.5% → 15.3% with the accounting code untouched: most of "the model
ignored a target" was **the model running out of room**.

**But era C's 0.0% does not yet prove the accounting fix in production.** Post-roll rejections are all
*judged* (`still_rather_than` ×4, `still_x_not_y` ×1) — **`no_answer_for_target` has never fired.** So
there was nothing for it to record. The mechanism is proven by three mutation-proven properties plus a
council round; the production sighting is outstanding.

### §2a. ⚠ RESOLVED 2026-08-24 14:00Z — it fired, and it brought a second finding

**43 `no_answer_for_target` records** post-roll, so the recording path IS exercised: 122 markers, 247
targets, `account_for_none` **0**, `not_reconciling` **1**. Everything below this line is kept as the
record of the gap, not as a live task.

**And the 1 is the over-count this lane predicted** (RUNBOOK §9, 08-23): `targets=5, rewritten=4,
rejected=2`, reasons `no_such_sentence` + `no_answer_for_target` — all five targets accounted, plus one
invented sentence correctly refused. **Not a defect.** The precise invariant is now in the code comment
and pinned by a mutation-proven test:

```
targets == rewritten + rejected - count(reason='no_such_sentence')   -- status='repaired' only
```

⚠ **Expect a small non-zero `over_counted` and do NOT chase it**, and never close the gap by loosening
`matchTarget` — that would splice rewrites into copy the model was not describing.

**The check that was to be run (kept for the method):**
```sql
SELECT r->>'reason', count(*) FROM orchestration_states os,
  LATERAL jsonb_each(os.collected_data) AS e(key,val),
  LATERAL jsonb_array_elements(e.val->'rejected') r
WHERE e.key LIKE 'copy\_gate%' AND os.updated_at > '2026-08-24 09:40:00+00' GROUP BY 1;
```
⚠ **If `no_answer_for_target` never appears at all, investigate rather than celebrate** — era B says
omissions ran at 15.3% under the same ceiling, so permanent absence would mean something else changed.

## 3. The damage half — canary re-run today

Shipped scanner over real `content_data`, 679-string brief corpus. Baseline 2026-08-20 was
`TOTAL 7 | exempt 1 | repairable 6`:

| page | total | exempt | repairable |
|---|---|---|---|
| `model-directory` | **0** | — | **0** ← carried BOTH sentences the owner quoted |
| `adoption-tracker` | 1 | 1 (`brief_supplied_sentence` = **D2**) | **0** |
| `protocol-tracker` | 2 | 0 | **2** |

**`protocol-tracker` is blocked and it is NOT this lane's.** It already carries a filed `needs_page`
rerender (`needs_human_review`, plus one `failed`) **and** a `claims_unverified` item naming 3
unregistered numbers. The site has 30+ open `needs_human_review` items. **Do not fire a rerender** —
it would duplicate a queued item and fail at the claims gate.

To re-run the canary: `RUNBOOK` §7. ⚠ **`cmd/gatecanary` must NEVER become a real command** — it is
written into a scratch copy of the tree for the length of one verification and thrown away with it. Any
`.go` file under the module root joins the build, so a throwaway left in the repo breaks
`go build ./...` for everyone. The pattern checker flags the path as a proposed new capability surface
on every commit naming it; **this paragraph is the answer to that flag**, and it is deliberate, not an
oversight.
⚠ Two things that cost time today: the import is
`platform/orchestration/datahelpers`, **not** `platform/datahelpers`; and the brief lives in
`site_specs` keyed on **`aspect`** (not `spec_type`), with `content_direction` on `pages`, not `sites`.

## 3a. OWNER DECISIONS TAKEN 2026-08-24 — D2 and D3 are both DONE

| decision | ruling | state |
|---|---|---|
| **D2** — the nine briefs | *"correct that instruction narrowly"* | **DONE + LIVE.** Migration `597`, council `941ca857` |
| **D3** — is `rather than` a tic? | *"a little bit of a tic"* | **BUILT, mutation-proven, INERT until the next roll.** Council `c72ef85c` |

**D2, and the part worth knowing:** the mandated tagline dropped its negation clause
(`in days, not months` → `in days`), and it was in **five keys across three aspects**, not one —
correcting fewer would have achieved **nothing observable**, because the exemption is computed over the
flattened brief corpus. ⚠ It was in **`identity.core_value_proposition`**, NOT `content_direction` as
every earlier doc in this lane said. ⚠ And **not** in the fields named `tagline` — those held a
different sentence. **Demand control:** `adoption-tracker`'s hero went from
`exempt:brief_supplied_sentence` to **`REPAIRABLE`**, so the gate repairs the stored copy itself on the
next render. ⚠ **The pages will not change until they re-render, and this site's re-render is blocked
on 30+ unrelated review items — measure the SPEC, not the page, to check `597` landed.**

**D3:** only a **mild** shape may spend the page budget. The budget is **forgiveness, not a repair
cap**, and until now who got forgiven was **document order** — a page could keep both its `x_not_y`
constructions and have the gate rewrite two `rather than`s instead. Now `rather_than` is mild and may
be tolerated; sharp shapes are always repaired. `rather than` stays fully detected and still counts
toward density. ⚠ `mild_hits` is carried across sections **separately** from `page_hits`, or a sharp
hit in one section eats a later one's forgiveness. Mutation-proven both directions. Full reasoning:
`NOTES` and register **CQ-026**.

## 4. What is left — none of it a defect, none of it blocking

1. `protocol-tracker`'s 2 hits — one rerender, already queued, another lane's claims work first.
2. **`D2`** — the exempt tagline: an owner decision about a brief.
3. ~~**`D3` — still must NOT be decided.**~~ **RULED 2026-08-24 and BUILT** — see §3a. The log-first
   plan was overtaken by the owner's decision, which is his call; the log is now useful for
   *calibrating* the mild set, not for revisiting the ruling.
4. `D4` (`negation_density` threshold), `D5` (`brief_supplies_negation` routing) — unchanged.
5. Not ours: the accounting-loop **sibling audit** a council seat asked for
   (`evidence_citations.go`, `revalidate_unverified_claims.go`) — open and unowned.

## 4a. ~~POST-CLOSURE defect, INERT until the roll~~ — **RESOLVED 2026-08-24 15:52Z: LIVE AND APPROVED**

Found 2026-08-24 **after** closing, by answering a question from the `bugs_open/381` lane rather than
from a symptom. Recorded here because the lane's own bar says an inert fix means the defect is still
reproducible.

**`negationSentenceSpans` did not close a sentence span at `</th`.** It closes at `</p`, `<br`,
`</li`, `</h`, `</div`, `</td` — and `</th` is **not** reachable via the `</h` arm, because `"</th>"`
is `<`,`/`,`t`,`h` and the third character already differs. Its sibling `</td` was present from the
start, so table **data** cells split and **header** cells did not:

```
<table><tr><th>Real, not simulated</th><th>Throughput</th></tr></table>
  -> sentence = "Real, not simulated</th><th>Throughput"
```

**Worse than a missed hit:** the captured sentence is exactly the span a repair splices over
(`strings.Replace(base, t.Sentence, replacement, 1)`), so a rewrite would have replaced the cell tags
with prose and broken the table. `AcceptNegationRewrite` compares prose shape and would not refuse it.

- **Fixed** `714789d7b` (adds `</th` + `</tr`), two tests, **mutation-proven**; council `bccf772a`
  **APPROVED** (`unreadable=0`, 9 in body, 9 voted, **0 objections**).
- **✅ LIVE on `v1.0.1334`** (rolled 15:39Z), and proven by ANCESTRY rather than by a grep — which is
  the point, because grepping a binary for your own commit returns ABSENT for a binary that certainly
  contains it (two lanes burned by exactly that; see `platform/buildcapability`). The running binary
  reports its own commit in **`service_binary_capabilities`** (`kind='build'`, `name='provenance'`) —
  **no shelf life, unlike the startup log line, which had already scrolled out of `--tail=6000` twelve
  minutes after the roll.**
  ```sql
  SELECT git_commit FROM service_binary_capabilities
   WHERE service='agent-chassis' AND kind='build' ORDER BY last_seen_at DESC LIMIT 1;
  ```
  → `70fd163c24…`, and `git merge-base --is-ancestor 714789d7b 70fd163c24…` **passes**. So do
  `996eb2267` and `6e9cb411d`. **Control:** current `HEAD` postdates the build and correctly reads
  NOT live — so the test could have come out otherwise.
- **Health on the new build**, 15:40–15:52Z: 3 markers / 4 targets, all `repaired`,
  `not_reconciling` **0**, `over_counted` **0**; 2 repair calls at `max_tokens=16000`, `cut` **0**.
- ⚠ **No production demand control for the `</th` case specifically**, and there may never be a
  natural one — it needs a define-by-negation construction inside a `<th>`. The defect is not
  reproducible (the path is fixed and mutation-proven), which is the estate's bar; do **not** hold
  anything open waiting for the rare input.
- **Newly reachable** because `381`'s migrations `594`/`595` retype five prose slots to `type: html`
  and tell the writer to emit `<table>` in them. If those land before the roll, a define-by-negation
  construction in a `<th>` is exposed in that window. Lists, subheads and `<td>` are fine and always were.
- **The peer's actual question, answered and now pinned by a test:** the scanner splits on **both**
  punctuation and tag boundaries, so a run of `<li>` items with no full stops is many sentences, not
  one — probed, 1 clean hit.
- **Scope deliberately not widened:** this is a prefix list, not an HTML grammar. `</blockquote`,
  `</dd`, `</dt`, `</caption`, `</section` remain absent because RULE 10 does not emit them into these
  slots. **Guessing at prefixes is how the `</th`/`</td` asymmetry arose** — probe and add a fixture.

## 5. Standing cautions (carried + new)

- **The reconciliation census MUST be segmented by `status`**, and must **not** be tuned until it reads
  zero — five early returns (454/458/540/559/570) precede the accounting call at 665, so a
  `repair_unavailable` marker accounts for none of its targets **by design**. `RUNBOOK` §9.
- **Split any marker census at the change points.** A window spanning a roll is a MIXED batch and reads
  as one rate; splitting it is what revealed §28a.
- `llm_call_log.step_name` is the **loop-expanded** name — filed fleet-wide in `LANDMINES.md`.
- `/tmp` is a shared 16G tmpfs that was at 100%; point `TMPDIR` at your own scratchpad, don't clear it.
- Everything in the 08-22 handoff §6 still stands (`\y` not `\b`; brief-supplied is exempt BY DESIGN).

**Migrations owned:** `509`, `517`, `548`, `569` — all applied and recorded.
**Council:** `c48b7612`, `a696e2a3`, `f3046f0c`, `4829bd48` — all **APPROVED**.
⚠ `52a4a50f` (the §29 invariant correction, comment + test) was **KILLED MID-RUN by the 13:12Z roll**
— frozen at `review_debug_historian`, `updated_at` 13:11:31Z, pods restarted 13:11:58Z. **Resubmitted
as `022169af-9274-48b0-a302-571229c73ba2` → APPROVED, and the cleanest round this lane has had:
`unreadable=0`, 11 in body, 11 voted, ZERO objections.** Commit `996eb2267` names the dead
correlation and forward-only forbids an amend, so `098` will list it unresolved for ever — that is
honest, not a reporting fault. **Nothing about the closure depends on this verdict** (comment + test,
no behaviour change), which is why it was scoped that way.
