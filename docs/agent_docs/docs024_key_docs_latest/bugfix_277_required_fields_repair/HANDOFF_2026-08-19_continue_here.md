# HANDOFF — 2026-08-19, fresh chat starts here: the roll landed, both changes are live and proven, and neither bug can close yet

**Supersedes `HANDOFF_2026-08-18d_continue_here.md`.** That file's §0 state table and §1 are stale
(both council verdicts are in, the roll has happened, and two of its numbers are corrected below).
Its §2 (`bugs_open/300`) and §3 (`bugs_open/314`) still read true. **Read this from disk, then
`NOTES_required_fields_repair.md` from the bottom.**

---

## 0. THE SHORT ANSWER TO "CAN WE CLOSE THIS LANE": no, and the reasons are specific

| bug | state | what blocks the close |
|---|---|---|
| **`bugs_open/277`** | router BUILT, LIVE, council-approved r5, **and doing its job** | **its own verify criterion, clause 1: the worked example must be REPAIRED. It is classified, not repaired.** Nothing repairs this type at all — see §3 |
| **`bugs_open/083`** | fix complete + artefact-proven | the door soak, ~2026-08-25 (owner decision 5). Also: `479`'s reclaim arm has still never fired on a real row |
| `bugs_open/300` | **fix LIVE on `v1.0.1314`**, council APPROVED r1 | behaviourally unexercised — nothing has dispatched this type since 08-18 |
| `bugs_open/314` | filed 08-18, unfixed | it is a proposal for the gate; owner's call which candidate |

Neither of this lane's two bugs is closeable today. **`277` is the interesting one** — it is much
closer to done than it looks, and it is blocked on something that is not routing.

---

## 1. THE ROLL — both changes live, proven at the artefact, behaviourally unverified

`agent-chassis:v1.0.1314`, pods `-l5h6l` (07:52Z) and `-nxmkf` (08:05Z). The `build provenance`
startup line had already scrolled out of `--tail=3000` — **that means "not in range", never
"unstamped"** — so a single-pass binary probe on **both replicas**:

```sh
kubectl -n ai-persona-system exec <pod> -- sh -c \
 "grep -aoE 'owned_page_refusal_status|resolveStatusRepairComponent|OWNED_PAGE_GUARD|ZZQQ_NEEDLE_THAT_MUST_NOT_EXIST' /proc/1/exe | sort -u"
```
`owned_page_refusal_status` **PRESENT** · `resolveStatusRepairComponent` **PRESENT** ·
`OWNED_PAGE_GUARD` **PRESENT** (long-lived control — the probe works) · nonsense needle **ABSENT**
(the probe discriminates). Config half intact.

**⚠ Neither is behaviourally verified, and do not let the quiet read as success.** Zero owned-page
refusals since the roll — **and zero `page-build-handler` orchestrations either**, so the zero is a
DEMAND artefact. Same for `300`: no `page_component_status_drift` dispatch since 08-18.

**Do not induce one.** Refusals occur at ~4/hour on live traffic (`bugs_open/301` measured 59 in
14 h); inducing costs exactly the wasted LLM chain that `301` exists about. **Just re-run the
RUNBOOK query tomorrow**, and it needs BOTH controls — refusals landing `wont_fix` with
`result ? 'owned_page_refusal'`, **and** genuine save failures still landing `failed` without the
stamp. A zero on the control means no genuine failures happened in the window, not that the split
works.

---

## 2. TWO NUMBERS OF MINE WERE WRONG, and the second changes what Tier 1 is worth

**Both are corrected in place in `bugs_open/301`, register WII-019, NOTES and WRONG_CALLS.** They
are here because a reader of the old handoffs will otherwise inherit them.

**(1) `phantom_internal_link` is 62.7%, not 47%.** Lifetime, live+archive, terminal only: generic
**101/46 = 68.7%**, owned **0/14**, total **101/60 = 62.7%**. The two component figures I quoted
beside it were right all along; the blend was arithmetic I got wrong and carried into four
documents. **The floor is 25%, so crossing it from there needs 243 more failures** — "one bad
stretch from switching off" was overstated.

**(2) "owned page + failed" IS NOT "ownership refusal", and this refutes a remedy I nearly
proposed.** Discriminating by the guard's own error text rather than by `pages.rebuild_policy`: of
**87** `owned`+`failed` rows, **85 name the guard and 2 do not** — and those 2 are
`placeholder_contact`'s, whose error is `step process_sections_loop_iter_0_generate_content failed`,
i.e. the **content generator** failing, not the guard refusing.

**So Tier 1 releases NOTHING that is held today**, and would not have even applied retroactively:

| held pair | why it is held | Tier 1 touches it? |
|---|---|---|
| `literal_markdown` | 3 ok / 16 REAL failures — still below floor with refusals excluded | no (`bugs_open/184`) |
| `placeholder_contact` | never completed one; its owned failures are generator errors | no |
| `dead_fragment_link` | never completed one — awaiting a hand canary | no |
| `missing_conversion_path` | never completed one; `bugs_open/255` — handler cannot read its spec | no |

**Its value is PREVENTIVE and still real** — 85 identified refusals already sit in the `failed`
bucket, and ~134 findings are queued behind the refusal on owned pages, every one of which would
otherwise enter a denominator it has nothing to do with. **But it is not restorative. Do not tell
the owner a pair was rescued.**

---

## 3. `bugs_open/277` — the half that is missing, and it is not routing

Measured against **this bug's own** verify criterion, not mine.

**Clause 2, MET.** The router is live and the type moves: 130 complete / 30 `needs_human_review`,
handler active as recently as 08-19 08:45. **All 30 parked rows carry a route** —
`no_content_data` 27, `asset_sourced` 2, `no_plan_owned` 1, **zero unrouted**. Nothing strands
unclassified any more, which is what the bug was filed about.

**Clause 1, NOT MET.** *"The gas converter's three items go `needs_human_review` → repaired → the
page serves real content."* Its item sits at `needs_human_review`, route `no_plan_owned`, updated
today. **Classified, not repaired.**

**And the general form:** *nothing repairs this type.* Completions in the live table are **44
`auto:revalidated`** (a sweep noticed the defect had gone — the page got content by some other
route), **37 `build-dispatch-loop`**, and **0 by the router**. The queue looks healthier than the
pages are.

**This is not a criticism of the router**, which does what it was built and approved to do. It is
that the owner's ruling — *"create a repair handler fleet wide"* — is half-delivered: routing
exists and is proven, repairing does not exist for `no_content_data`, which is 27 of the 30.

> **⚠ And the missing half is probably the SAME missing piece as Tier 2** — a finding-to-edit
> converter. `copy-editor` already emits `apply_section_edit`'s exact input shape
> (`{page_component_id, slot_name, field_updates, rationale}`) from a component's `content_data`,
> rendered HTML and declared schema. **It is ONE DAY OLD and owned by the
> `loanandmortgagecalculator_couk` lane** (migrations `447`/`462`). **Talk to them before designing
> anything** — a design written tonight against a contract that changed twice in two days is stale
> before it is read.

---

## 4. WHAT IS LEFT, in the order I would do it

1. **Tomorrow: re-run the two post-roll checks** (RUNBOOK), each with both controls. Minutes. This
   is the only thing standing between `300` and "proven", and between Tier 1 and the same.
   **This is now the top item — everything else on this list is either answered, owner's-call, or
   dated.**
2. **~2026-08-25: close `083`** once `444`/`458`'s doors have held a week (owner decision 5). Move
   with **both paths on the commit** (`git mv` landmine) and verify at HEAD with `git ls-tree`.
   ⚠ Before closing, check `479`'s reclaim arm has fired at least once — it still never has, so it
   is shipped-but-unexercised, and the close should say so rather than imply it works.
3. **`277`'s remaining half** — the `no_content_data` repair. This is also Tier 2.
   **✅ ASKED AND ANSWERED SAME DAY — the answer is "DIFFERENT AGENT". Do not re-open it, and do not
   design around `copy-editor`.** Stage 2's whole value is the *page-scoped read*; for "this
   component, this defect" that read is pure cost (78KB of prompt, a ~100s call, to change a field
   you already know is wrong), and its three-edit budget is a **symptom of page scope, not a feature
   to inherit**. §5–§7. CONTRIB `7574482c7`, their reply
   `bugfix_277_required_fields_repair/CONTRIB_2026-08-19_reply_different_agent_and_check_473_before_you_build_anything.md`.
   ⚠ **AND CHECK `473` FIRST** — a deterministic, no-LLM repair for `literal_markdown` ships this
   week and would make an LLM route for that class redundant. See §7b for what it does and does not
   cover.
4. **`314`** — owner's call between the four candidates; candidate 1 is one line plus a credit cost
   somebody should size.
5. ~~**Two loose ends nobody owns**, both `[UNMEASURED]`~~ **BOTH MEASURED 2026-08-19 by
   `agentchassis-22` — see §8. (a)'s PREMISE WAS VOID; (b) has a population but no mechanism.**
   - ~~`page-rerender` saves to owned pages ~3,754 times without refusal while `page-build-handler` is
     refused every time. Same guard. One of those needs explaining.~~ **There was no asymmetry to
     explain.**
   - ~~a page named/URL'd `tool-…` carrying `rebuild_policy='generic'` looks like a data defect;
     nobody has counted how many.~~ **COUNTED, AND THE DAMAGE IS ZERO — recommendation: DO NOT FILE.**
     ~69 real tools are inconsistently marked, and **inconsistent is not damaged**: no tool page has
     been clobbered, and the seed case serves its tool fully intact. §8.

> ⚠ **CORRECTED 2026-08-19 — `copy-editor` is owned by the `copy_quality_two_stage` lane, NOT `loanandmortgagecalculator_couk`.** I got the wrong lane from a `grep -rl "copy-editor"` hit in LMC's `README_where_we_are.md` — a *mention* — and read it as ownership. `scripts/who-owns.py` exists to separate those two, and I did not run it. The defining evidence is what the commits shipping migrations `447`/`462` actually touch: `docs024_key_docs_latest/copy_quality_two_stage/`. Register entry **CQ-024**. A CONTRIB is filed in their lane dir (`CONTRIB_2026-08-19_from_the_277_083_lane_…`, commit `7574482c7`).


## 5. THE OUTBOUND ASK — filed 2026-08-19, ANSWERED same day (see §6)

**To the `copy_quality_two_stage` lane** (register **CQ-024**; migrations `447` seed + `462`
budget). ⚠ **NOT `loanandmortgagecalculator_couk`** — see the correction boxes above; I had the
wrong lane for a day.

**The question asked:** can stage 2 be aimed at ONE named component with ONE named defect, rather
than at a whole page for editorial quality — is that something the design can accommodate,
deliberately excludes, or has already rejected?

**Why it is phrased as a question and not a proposal, which matters if you pick this up.** Their
handoff states two things are deliberate, and both are exactly what this use would press on:

1. *"Nothing dispatches `copy-editor`. No item_type routes to it, by choice."* — routing a finding
   at it **is** a dispatch.
2. It **cannot write to a page** — no step can, and the migration RAISEs if one is added. Output
   goes to `copy_edit_proposed` at `needs_human_review`; both proposals to date were owner-approved
   and **applied by hand**. A queue of ~134 wants less than one human approval each, which is a
   change to their safety posture — **theirs to rule on, not ours to assume.**

**"Build your own, here is what we learned" is named in the CONTRIB as a complete and welcome
answer.** If that is the reply, the parts that look hardest-won from outside are: enumerating the
page's required links as **data** (their own comment: *"a prose instruction to preserve a set is not
reliably followed"*), the declared-schema-in / same-type-out rule, and the **3-edit budget** added
by `462` after run 2 broke stage 2 on a harder page.

**What was offered back, so this is not purely extractive:** their two proposals are applied by
hand, and `apply_section_edit` consumes their exact output shape at **220/5 = 98%** on owned pages.
⚠ Stated to them as a **near-match with the gaps named from the spec**, not a drop-in: `edit_type`
is **Required** and their payload does not carry one, and the **RFC_015 citation gate**
(`acknowledges_decision` / `supersedes_decision`) is a control against precisely what an automated
editorial pass could get wrong. Whether they compose is their call.

> ⚠ **CORRECTED same day — "not a reachable peer" was wrong, and the distinction matters.** They
> ARE reachable; their peer name is **derived** (`agentchassis-8d`), not lane-shaped, so searching
> `ListAgents` for "copy quality two stage" misses it. Resolve names via `~/.claude/sessions/*.json`
> (`name` + `sessionId` → `af212352`), then map to transcripts in
> `~/.claude/projects/-home-ant-projects-agentchassis/`. **`ListAgents`' `[ref]` is NOT the
> session-id prefix**, which is what makes the obvious lookup fail. My claim should have been "I
> could not find the name", which is a different thing. (Told to me by `agentchassis-22`; verified
> before acting on it.)

**✅ ANSWERED — they reached me independently within the hour and their reply cites the CONTRIB**, so
the durable file was still the right thing to file. **See §6 for what came back, including a live
hazard they found in the promoter.** The committed CONTRIB remains the channel of record.
That is the estate's normal convention and it demonstrably works: their own handoff records two
inbound asks from other lanes, both answered. **Check for a reply in their lane dir, in
`bugs_open/277`, or in this lane's directory before assuming silence** — and give it more than a
day before reading silence as a no.

---

## 6. WHAT CAME BACK — and a hazard the other lane found in OUR promoter

**`copy_quality_two_stage` replied the same hour** (peer `agentchassis-8d`). Their answer to the
CONTRIB is not yet the narrow yes/no, but they volunteered something more important.

**`copy_edit_proposed` / `human-review` must NEVER be promoted or auto-dispatched — owner decision
D2 (2026-08-12): stage 2's output queues for human review, no unreviewed auto-rewrite.** Migration
`447` asserts the no-page-write property **at apply time**, with a guarded `DO` that RAISEs if any
step's action is one of six page-writing actions — a guarantee held by a check rather than a comment.

**Their question was whether our promoter needs it recorded as never-promotable. Measured answer:
safe today by TWO barriers, but the second one is rotten.**

1. `checkpoint_for_review` files at `needs_human_review` (`:223`); `scored` selects `status='detected'`
   only. Never looked at.
2. At `detected` it would fail `handler_ok` — `human-review` is not a live `agent_definitions` row (0).

⚠ **But a held row's reason is *"handler not a live agent"*, which reads as a broken routing config —
and `held-pair-canary-escalation` (ours) escalates after 3 days ASKING A HUMAN TO CANARY THE PAIR.**
So our own machinery would invite someone to break D2. **Safe by accident is indistinguishable from
safe by design until somebody acts on the hold reason.**

**Recommended to them (their call, not ours):** move the label into `spec` and file with
`handler_agent = ''`, which is what `voice_tells` does (43 rows) and which `scored` excludes
**outright** — the pre_query's own comment says such rows belong at `detected` permanently and
"holding is not what is happening to them". **I explicitly declined to add an exclusion list to the
promoter** — a second roster to maintain is the drift class this estate keeps filing bugs about, and
an empty handler makes the bad state unrepresentable rather than merely refused. If they push back,
the alternative offered is an explicit `item_type` exclusion in the `pre_query` **with D2 cited next
to it**. `LANDMINES.md` entry added jointly; verifier dispatched.

**Their two gifts, both worth keeping:** an all-history denominator can flip a before/after rate
depending on the window (same family as our "lifetime meant 7 days"); and they hand-file
`section_edit` rows born `triaged` then self-claim before publishing, because `section_edit` claim
latency is **1,695s mean with a 21,757s tail (n=172)**. That pattern is invisible to our promoter
(it selects `detected`) and needs no change.

**§4.5 is handed to `agentchassis-22`** — both loose ends, read-only, their own dated appendix. What
they were given that is not otherwise written down: the 3,754 figure is **inherited, not ours** and
must be re-derived; classify by the guard's **error text**, never `pages.rebuild_policy`; and
`page-rerender`'s `save_sections` has **no `error_step`**, so a refusal there fails the workflow
rather than routing — a candidate explanation for refusals being *invisible* on that route rather
than absent.

⚠ **One observation from this session that had never reached a file, now handed over with it:**
**0 `page-build-handler` orchestrations since the roll, while 20 of its work items were updated at
08:45:58.** A handler apparently acting with no orchestration to show for it. Noticed, not chased,
not recorded until now — **treat my zero as unverified.**

---

## 7. OWED WORK from the peer exchange — one migration, and one measured warning already sent

### 7a. OWED: an explicit `copy_edit_proposed` exclusion in the promoter's `pre_query`, citing D2

**I changed my mind and this is owed.** I first declined (see §6) on "a second place to maintain is
the drift class we keep filing bugs about". `agentchassis-8d` came back with two facts that beat the
objection:

- **`handler_agent=''` is not available to them.** `checkpoint_for_review_action.go:202` hardcodes
  the literal in the INSERT (`… $7, 'human-review', 'needs_human_review', $8`). Making it settable is
  a Go change to a shared action (4 live agents, 9 items) — council + roll, not a config edit.
- **`status` is hardcoded in that same INSERT**, so `copy_edit_proposed` **cannot be born `detected`
  through its only producer**. Barrier 1 is structural, not conventional — better than my reading —
  and the residual is now precisely characterised: **a human hand-filing such a row at `detected`,
  whom `held-pair-canary-escalation` would then invite to canary the pair.**

My objection was aimed at a *roster of types*; this is **one named exclusion guarding a
characterised residual**, with no second half to drift from. That is a different thing and it is
cheaper than the failure it prevents.

**⚠ NOT done, and deliberately not done quietly.** It is a config change to a live shared scheduled
task; this lane already has four migrations on that `pre_query`. It needs a numbered migration with
a guard + `_ROLLBACK.sql`, exercised in a rolled-back transaction first, **and it goes past the owner
— not in on a peer's say-so, however good the reasoning.** Cite **owner decision D2 (2026-08-12):
stage 2's output queues for human review, no unreviewed auto-rewrite.** If someone gets there first,
the full argument is in the `LANDMINES.md` entry.

### 7b. SENT: `473` will be refused on the 16 owned pages — warning delivered before the apply

Raised by `agentchassis-8d`, measured by me, filed to the 184 lane at `e7b009483` and messaged to
the live `bug 184` peer. **`473` is not applied** (absent from `schema_migrations`), so it landed in
time.

`473` re-routes `literal_markdown` onto **`page-rerender`** and contains no `rebuild_policy`
handling. All **74** `literal_markdown` rows route at `page-build-handler` today, so the re-route
moves the whole population onto an unused route — and **that route is not exempt from the guard:
16 `failed` + 1 `cancelled` `page_rerender` items on `owned` pages name it, most recent 2026-08-18.**
So `473` should fix the generic pages (the real target, 13% today) and hit the identical refusal on
the **16 owned** ones.

⚠ **Stated limit, which matters if you follow up:** the 17 rows prove the route *can* be refused, not
that an item arriving via `473`'s new `spec.reason` condition *will* be. `page-rerender`'s
`save_sections` resolves its page from `input_data.spec.page_name` and has an **early return
reporting `success:true, skipped:true`** when that misses — a third path reaching neither guard nor
repair. **A one-item canary on a known owned page settles it; check the served page, not the item
status.**

⚠ **And do not use the 1,769 `page-rerender` completions on owned pages as evidence of anything** —
whether they are real writes or that silent-skip path is unestablished, and is precisely §4.5(a),
now with `agentchassis-22`.

**Why this matters to us and not only to them:** `literal_markdown` is the largest held pair
(3 ok / 16 real failures) and 16 of its refusals are ours. If `473` lands, its floor arithmetic
improves with nothing of ours changing — **which is exactly why I declared that incentive to them
while telling them it will not cover 16 rows.**

### 7c. PRE-REGISTERED PREDICTION for `473`'s apply — written while it is STILL UNAPPLIED

**Why pre-registered rather than "check afterwards".** This lane has been bitten four times this week
by classifying a population *after* seeing it — a policy column standing in for a refusal, an
inherited count standing in for saves, an aggregate standing in for rows. **A prediction written
before the event cannot be reverse-fitted**, and it costs two minutes. `473` is confirmed absent from
`schema_migrations` at the time of writing.

**BASELINE, `literal_markdown` by the page's policy [MEASURED 2026-08-19, PRE-APPLY]:**

| policy | failed | unresolved | detected | needs_human_review | complete | total |
|---|---|---|---|---|---|---|
| `generic` | 16 | 10 | 2 | 2 | 3 | **33** |
| **`owned`** | 8 | 24 | 8 | 1 | 0 | **41** |

**PREDICTION.** After `473` applies and these dispatch, each owned-page item lands in one of three
outcomes — and **2 and 3 are INDISTINGUISHABLE at the work-item row**, which is the whole reason to
write this down first:

1. **REFUSED** — `failed`, `error LIKE '%rebuild_policy=owned%'`. **My prediction**, because `473`
   puts this population on the branch that calls `save_page_sections`, and the escalation path that
   would divert it needs a section with *no* `content_data` — which a `literal_markdown` section by
   definition has.
2. **COMPLETES WITHOUT WRITING** — `complete`, but the served page still carries the asterisks.
   **This is `agentchassis-22`'s `check_escalated` mechanism
   (`rerender_page_sections_action.go:401-419`), currently `[INFERRED]` from code and never
   measured.** If this is what happens, their inference becomes a measurement.
3. **ACTUALLY REPAIRED** — complete, and the asterisks are gone at the served page.

**THE CHECK, and it must reach the served page:**
```sql
-- item level: outcome 1 separates cleanly; 2 and 3 do NOT
SELECT COALESCE(p.rebuild_policy,'(gone)') AS policy, wi.status,
       count(*) FILTER (WHERE wi.error LIKE '%rebuild_policy=owned%') AS refused,
       count(*) AS rows
FROM site_work_items wi LEFT JOIN pages p ON p.id = wi.page_id
WHERE wi.item_type='literal_markdown' AND wi.updated_at > '<the apply>'
GROUP BY 1,2 ORDER BY 1,4 DESC;
```
Then for anything `complete` on an `owned` page, **fetch the page and look for the asterisks**. That
is the only thing separating 2 from 3; `page_components.updated_at` is a corroborator, **not** a
substitute (`bugs_open/315`: a rerender completes and stamps without writing bytes).

**Read the `generic` column FIRST — it is the positive control.** Expected there: mostly outcome 3.
If generic pages also land on 2, the migration is not working at all and the owned-page question is
moot.

⚠ **Record it either way.** Outcome 1 closes my warning to the 184 lane; outcome 2 promotes
`agentchassis-22`'s `[INFERRED]` mechanism to measured and matters well beyond `literal_markdown`.
**Neither of us owns `473` and neither may see it land**, which is why this is in a file rather than
in a session.

### 7c-RESULT. THE PREDICTION WAS WRONG — there was a FOURTH outcome, and my enumeration said "exactly one of three"

**`473`/`474` applied 2026-08-19 10:34:26Z / 10:34:35Z, ledger-recorded.** The 184 lane took all
three suggestions (scope statement in the header, the `rebuild_policy` caveat, and the verify block's
at-apply split — which read **30 generic / 41 owned open items**, my 16 having been terminal refusals
only).

**The owned-page canary took NEITHER predicted branch.** Not the ownership refusal (outcome 1, mine),
not the silent skip (outcome 2, `agentchassis-22`'s). It was **refused by the COMPLETION VERIFIER**,
on a ported-page slot's `rendered_html` code_span: ported/tool slots have **no `content_data` to
strip** and carry their HTML through, so the verifier — `item_type`-keyed, runs on every completion —
blocked it honestly, **before `pageIsOwnedForGuard` was ever reached.**

**Two things this changes in our model, and the second matters more than the first:**

1. **An owned-page `literal_markdown` item can terminally fail without ever touching the ownership
   guard**, and that failure is *honest*. So a `failed` row on an owned page is not even reliably a
   refusal at the *route* level, let alone the policy-column level (§2). Another reason the error
   text is the only discriminator.
2. **The "completes having written nothing" path CANNOT false-green for this item_type** — the
   verifier re-scans both stored surfaces. That is a real floor under `315`'s failure mode, for
   the types that have a registered verifier.

> ⚠ **THE LESSON, and it is this week's shape again.** I wrote *"each owned-page item lands in
> exactly one of three outcomes"*. It landed in a fourth. **The enumeration felt exhaustive because I
> had derived it from the two mechanisms I happened to know about** — the guard I own and the
> escalation path a peer had just told me about — and I never asked what else runs on a completion.
> **A pre-registered prediction that is wrong is worth more than a vague one that is right**, and it
> is only because it was written down beforehand that the fourth outcome is legible as a finding
> rather than as "well, obviously".

### 7e. ⚠ CONSEQUENCE OF THE RE-ROUTE THAT IS OURS TO OWN: `473` created a new pair, and OUR gate holds it

[MEASURED 2026-08-19, post-apply, live+archive]

| `literal_markdown` pair | complete/verified | failed |
|---|---|---|
| `page-build-handler` (the old route) | 3 | 35 |
| **`page-rerender` (the NEW route)** | **0** | **3** |

**The promoter's `known_good` gate requires ≥1 lifetime `complete` or `verified` for the pair.
`(literal_markdown, page-rerender)` has none.** So it is held — reason *"pair has never completed one
(awaiting a hand canary)"* — and after 3 days `held-pair-canary-escalation` escalates it asking a
human to canary it.

**So even once the fleet roll carries `f3939f27d` and the mechanism works, our gate will not dispatch
this type on the new route until someone hand-canaries it.** Their fix is gated by our mechanism, and
they had no way to know.

**The good news, and it is what I told them:** the canary they will run anyway to prove the fix *is*
the unblock — **one** successful completion flips `known_good` and releases the rest.

> **The transferable property, which is bigger than this case: RE-ROUTING AN `item_type` TO FIX IT
> CREATES A NEW `(item_type, handler_agent)` PAIR WITH NO HISTORY, AND THE PROMOTER HOLDS IT BY
> DEFAULT.** The old pair's record does not transfer. Any lane that repairs a type by re-pointing its
> handler will hit this, will see "held — never completed one", and may read it as their fix having
> failed. **This is the promoter working as designed and it still needs saying out loud.**

⚠ **And do NOT read the 3 `page-rerender` failures as the route being wrong.** They are pre-roll: the
generic canaries showed the strip fires and the save succeeds, but news-listing items are
**query-resolved** and `plan.ResolvedData` merges last, re-imposing markdown from source in the same
run (root: `content_feed_items.source_summary` carries raw markdown in **~700 of 10,855 rows**).
Fixed by `f3939f27d`, **pending the next fleet roll.** The pair's ratio should recover after that
roll, not now.

### 7d. The discriminator worth keeping, from `agentchassis-8d`

> **Is success expressible as a diff assertion?**

"Are the asterisks gone and the text otherwise byte-identical" fully is; "does this read like a
person wrote it" cannot be. **A repair whose success condition is completely assertable does not need
a human-review posture — it needs a gate that fails loudly.** That is a sharper test than anything
this lane had written for when a repair needs HITL, and it independently argues for `473` over any
LLM route for this class.

---

## 8. §4.5 ANSWERED — and (a)'s premise was VOID, including in a figure I repeated

Measured by `agentchassis-22`, committed at `a43f99bb4` as a dated appendix in this lane's NOTES.
**Read their appendix, not this summary, before building on any of it.**

### (a) There was no asymmetry. The denominator was ~8x too big — and the figure was one I repeated.

**The 3,754 (now 3,818) figure counts WORK-ITEM OUTCOMES, NOT SAVES.** `page-rerender` gates
`save_sections` behind `check_rerender_mode`, whose condition is four reasons
(`image_landed`, `section_data_resolved`, `cta_links_stale`, `template_changed`); everything else
routes to `render_page` and **never calls `save_page_sections`**. Of 4,171 owned-page items,
**3,710 take that assemble-only branch. The guard is reached ~461 times, not 3,754.**

**And `page-rerender` IS refused — 81 times**: 81 of its 89 owned-page failures name the guard,
against **0** on generic pages. `page-build-handler` is 100 of 112 — **and is not "refused every
time" either: it completes 74 items on owned pages.**

**So both handlers call the same action and are refused in proportion to how often they reach it.**
`page-rerender` looked unrefused because the denominator counted items that never tried.

> ⚠ **I inherited that figure from another session's table in `bugs_open/301` and repeated it in a
> handoff with no marker. It was right as a COUNT and wrong as a MEANING** — which is the more
> dangerous failure, because re-deriving the number confirms it.

### THE REAL ANOMALY, which is sharper than the question asked — and it may bear on §3

Post-guard, on owned pages, through the **same** `save_sections` step: `section_data_resolved`
**122 complete / 0 refused** vs `cta_links_stale` **112 / 19**. Same agent, same step, same guard.

Candidate mechanism, **`[INFERRED]`** by them from code and not measured per row:
`rerender_page_sections_action.go:401-419` escalates a section with **no `content_data`** to the
writer, returns `escalated=true`, and `check_escalated` routes to `complete` — **skipping the save
entirely, so the item completes having written nothing.** `isSelfContainedSection` exempts tool
sections, which is how two reasons diverge through one step.

**That is `277`'s `no_content_data` population appearing on a second route**, so it plausibly bears
on §3's missing half — and it is a third way a repair can report success without touching the page.

### (b) ANSWERED 2026-08-19 — the damage is ZERO, and the recommendation is DO NOT FILE

**Route 2 (price the consequence, not the cause) was run and lands on a zero** (`3440e53b7`). The
clobber signature is interactive markup replaced by prose, so **a page still carrying its markup is
not clobbered by construction** — 72 of the 89 still carry it. Of the 17 possible: **7 archived, 11
on `loanzy.uk` all created 2026-08-18** (a site built that day, mid-build), and 3 older actives, each
checked at the **served page** with an owned-tool positive control and a fabricated-URL 404 control
in the same run. None is the defect — one is a `ported-prose` landing page *about* a tool (slots
`hero/ported-prose/faq/tool-cta`, correctly `generic`), one is a never-built page (404, zero
components — a different defect, see below), one is mid-build.

**And the seed case settles it the other way:** `tool-pet-treatment-cost-estimator`, the page this
question came from, serves **200, 21KB, 5 scripts, 2 inputs — its tool fully intact.** The page most
likely to show harm shows none.

⚠ **The "89 identically-shaped" figure is narrowed by its own author** — true of the two tests run,
**false as an implication**, because `ported-prose` landing pages legitimately sit under `/tools/`
with a `tool-` name. Counting only pages carrying interactive controls it is **69 of 89 against 83 of
95**. **The marking is inconsistent across ~69 real tools — and inconsistent is not damaged.**

**RECOMMENDATION, theirs and I agree: DO NOT FILE (b).** My own §8 criterion decides it — (b) has no
mechanism read at source, and now a **priced consequence of zero**. Filing would assert a
cross-cutting cause behind a harmless inconsistency, which is the 2026-07-31 ruling's target. Route 1
(find the writer) is now the wrong spend: it would buy a tidier database and no protection.
**What would reopen it is written down and the detector query is dated and re-runnable** — a
`generic`-marked interactive tool page serving prose where its tool used to be.

> ⚠ **The method note, which is the third instance in one day:** the aggregate looked damning
> (`generic` tool pages 81% carrying script vs `owned` 99%, and 2.5x the components — exactly the
> shape of a tool component replaced by prose sections) and **printing the 17 rows dissolved it**, because
> the proportion had folded a new site's in-flight build into a damage signal. **Do not trust a
> proportion before printing its rows.**

### (b), the earlier withdrawn figure — kept because the withdrawal is the useful part

**89** pages named `tool-%`, under `/tools/`, not guides, carry `generic` — against **95
identically-shaped** pages marked `owned`, plus 4 more under `/tools/` without the prefix. The 70
`tool-…-guide` pages are correctly generic, **so the name test alone over-counts and the two axes
must stay independent.** My seed case re-verified first-hand by them.

**They priced the damage at 107, then withdrew it in the same entry** — splitting by guard era showed
174 post-guard completions on the route they had classified as guarded, which is impossible if the
classification held. Two confounds they named: their per-agent route classification was an
assumption, and **`pages.rebuild_policy` is MUTABLE and read at query time**, so historical rows get
judged against today's marking. **Do not quote an exposure figure for (b) from anyone yet.**

### ⚠ A SECOND, INDEPENDENT REASON FOR THE ERROR-TEXT RULE — worth carrying beyond this lane

§2 says classify by the guard's error text because *owned + failed ≠ refusal*. There is a second
reason: **`rebuild_policy` can have CHANGED since the run.** The policy column tells you what the
page is marked as *now*; the error text is what the run itself *recorded*. For anything
retrospective, only the second is evidence.

---

## 8b. ⚠ THE TRAP THIS LANE KEEPS SETTING FOR ITSELF — read before writing any work-item count

**`site_work_items` is a ~7-DAY WINDOW.** `work-item-archiver` moves terminal rows to
`site_work_items_archive` — measured 2026-08-18 by another lane, **20,184 archived against 10,689
live**. Any count of work items over the live table alone silently answers *"in the last week"*.

**This lane diagnosed that, shipped the fix, and then fell for it the next day.** Migration
`465_promoter_reads_archived_history.sql` (`a62809d29`, 2026-08-18) exists precisely because the
promoter's "lifetime" success history was really 7 days; **`bugs_open/083`'s entire floor arithmetic
depends on that UNION.** On 2026-08-19 I wrote a work-item count into `bugs_open/315` without it and
was **4x low** (3 rerenders, actually 13; class-wide the real figure is **331 completed items against
55 pages with zero components**).

**The lesson is not "remember the archive".** I had bound the check to a **topic** — the promoter's
floor — rather than to an **operation**. When the topic changed, the check did not travel.

> **Bind it to the operation: every `count(*) FROM site_work_items` is a 7-day answer until proven
> otherwise, whatever it is about.**
> ```sql
> FROM (SELECT ... FROM site_work_items
>       UNION ALL SELECT ... FROM site_work_items_archive) x
> ```
> ⚠ **The archive does NOT carry every column** — `error` is absent, so the guard's-error-text
> discriminator (§2) only works on the live table. When you need both scope *and* the error text,
> say which one you traded away.

⚠ **And "was my measurement scoped correctly?" is ONE QUESTION PER TABLE, not one question.** In the
same query set that got this wrong, the counts over `pages` / `page_components` were right — those
tables are not archived. **The window-limited half and the unlimited half were wrong and right
together, in one breath.**

---

## 9. Session-start checklist
`git log --oneline -10` · re-read this file from disk · `scripts/who-owns.py` **by slug** for `277`,
`083`, `300`, `301`, `307`, and **`copy-editor` belongs to another lane** · re-measure §1's probe ·
then §4 step 1.

---

# ⚠ MERGE NOTE — 2026-08-19 ~10:30Z: this file was OVERWRITTEN by a concurrent session and restored

**Everything above is the original author's, restored verbatim from `917a5de9f`. It is the
authoritative account and supersedes the overwriting session's figures wherever they differ.**

What happened: a concurrent session wrote its own 08-19 handoff to this exact path with a shell
redirect (`cat >`), having never read the file — the precise failure CLAUDE.md warns about ("**read
before write on any file you did not create; prefer the Write tool, which refuses an unread file,
over a shell redirect, which does not**"). Nothing was lost, because git had it; recovered with
`git show 917a5de9f:<path>` and restored forward-only, no amend. Logged in `WRONG_CALLS.md`.

**Two corrections the restored file makes to the overwriting session's work — believe the version
above, not the NOTES entries written at 09:00Z:**

1. **`phantom_internal_link` is 62.7% lifetime, not 47%.** The 47% blend is arithmetic taken from
   `480`'s own header and repeated without re-deriving it. Crossing the 25% floor from 62.7% needs
   **243 more failures**, so "one bad stretch from switching off" was overstated.
2. **"owned page + `failed`" is NOT the same set as "ownership refusal"** — of 87 such rows, 85 name
   the guard and 2 are the content generator failing. Discriminate on the guard's error text, not on
   `pages.rebuild_policy`.

Also correct the closure dates from §0/§4 above, not from the overwriting session: **`083` closes
~2026-08-25** (owner decision 5, the door soak), and **`277` is blocked on its own verify clause 1 —
the worked example must be REPAIRED, and nothing repairs `no_content_data` at all** — which is a
harder blocker than the churn-guard clock, and the more useful thing to know.

## 10. The escalation clock, and other additive detail kept from a concurrent edit of this file

### A. The escalation clock: three dated ticks, and 08-19's ZERO IS CORRECT

`held-pair-canary-escalation` is daily and fires at **12:57 UTC**; at 09:23Z on 08-19 it had last run
**08-18 12:57:48**, so none of these had happened yet. Instrument: the `pre_query_result` line in
`kafka-scheduler` logs.

| when | expect | what it proves |
|---|---|---|
| **08-19 12:57** | **`escalated=0`, `watching=15`** | ⚠ **ZERO IS CORRECT — not a failed migration.** It is `466`(a) working (a `HAVING` that still speaks on an idle tick) |
| **08-20 12:57** | `placeholder_contact → page-build-handler` (3 rows) escalates, **canary** wording | first real escalation this mechanism has produced |
| **08-21 12:57** | `literal_markdown` (10, **floor**), `dead_fragment_link` (1), `missing_conversion_path` (1) | first use of `471`/`472`'s corrected floor remedy |

⚠ **A daily task with a 3-day predicate delivers 3–4 days**, because it can only act on its own tick:
`placeholder_contact`'s oldest row is 08-16 **19:17**, which at the 08-19 **12:57** tick is 6h20m
short of three days. Predicting the fire date from a DATE rather than the tick timestamp is off by a
full tick, and the miss shows up as a silent zero. Conditional on the held set not changing — the
clock keys on `min(created_at)` per PAIR.

⚠ **Do not canary `missing_conversion_path → content-gap-planner`** — `bugs_open/255` owns it
(diagnosis CONFIRMED first iteration: routed at a handler that cannot read its spec).

### B. `RFC_030` — the router engine, and its blocker is now CLEARED

`docs/agent_docs/docs024_key_docs_latest/router_engine/` (standing five) ·
`.../architecture_review/RFC_030_single_type_work_item_routers_want_one_engine.md`
(**RULED 2026-08-15 by the owner — scheduled as its own lane**).

Phase 1 (measure the live population per type) is **DONE**, in that lane's NOTES. **Phase 2 is a
council design round on shape A vs B, submitted as an RFC-shaped design, BEFORE building** — and
unlike this lane's config-only migrations it *is* architecture scope, so the gate is the right venue.

✅ **The PLAN's guarantee 8 was STALE and is FIXED (2026-08-19).** It said RFC_022's accumulation
counter was *"unbuilt"*; RFC_022 is **CLOSED**, the counter shipped 2026-08-13
(`cmd/config-key-audit --optional-key-budget`, register **WFA-013**), the owner **ruled N = 10** on
08-14, and a daily CronJob has enforced it since. That turns guarantee 8 from "volunteer a count
nobody consumes" into **a live budget with a ruled threshold**: the round must ask whether the chosen
shape makes each routed type accumulate optional keys on one shared action — i.e. whether the engine
walks toward N = 10 as it succeeds. **[UNVERIFIED] which of A or B has that property is what the
round must establish.** Full correction, with the hand-maintained-literal trap and the parity test to
run, is in the PLAN at guarantee 8.

Then: build the engine; migrate **`410` first** (its 8 routes define the contract, its 44-item
history is the regression fixture — evidence in *this* directory), then `397`'s two; retire the three
bespoke seeds; update CQ-023 and IMG-071; register the engine.

### C. Migration numbers collide on this tree and MUST NOT be renumbered

`453`, `454`, `462` and `471` each exist twice; `462`'s two halves are one applied, one pending.
**A number tells you neither author nor applied-state — ask the ledger by exact filename.** Renaming
orphans the ledger row, the file reads as pending, and the runner re-applies it.

### D. The landmine family this lane keeps hitting — now SIX, and the shape never varies

Every one is **a population or a domain assumed rather than enumerated**; none was caught by review
(twelve seats approved `444`):

1. `failed` rows carry **no `completed_at`**. 2. `status` has **two** terminal success states
(`complete`, `verified`). 3. The row set is **not stable**. 4. The row set is only a **~7-day
window** (`work-item-archiver`; the archive is bigger than the live table). 5. **A control that
cannot come out otherwise** — three tautological ones caught here; the test is not *"is this control
true?"* but ***"could it ever have come out non-zero?"*** 6. **The CLOCK** — see A.

Also live: **a same-tag rebuild ships the cached image** · **an aggregate-only SELECT with a `WHERE`
returns one row regardless — use `HAVING`** · **a pathspec commit still takes a same-file passenger**
· **backticks in `git commit -m` EXECUTE** · **`EXPLAIN` proves SQL parses, it cannot prove a path
inside a string exists** (that shipped `bugs_open/295` into a live payload 30 minutes after 295 moved
to `bugs_closed/`) · **and a `cat >` heredoc cannot tell you the file already existed** — this note.
