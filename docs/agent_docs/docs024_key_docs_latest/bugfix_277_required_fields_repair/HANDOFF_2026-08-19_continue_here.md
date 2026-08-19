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
2. **~2026-08-25: close `083`** once `444`/`458`'s doors have held a week (owner decision 5). Move
   with **both paths on the commit** (`git mv` landmine) and verify at HEAD with `git ls-tree`.
   ⚠ Before closing, check `479`'s reclaim arm has fired at least once — it still never has, so it
   is shipped-but-unexercised, and the close should say so rather than imply it works.
3. **`277`'s remaining half** — the `no_content_data` repair. This is also Tier 2.
   **✅ THE CONVERSATION IS OPEN — do not re-open it, and do not design ahead of the reply.**
   CONTRIB filed and committed (`7574482c7`) at
   `docs024_key_docs_latest/copy_quality_two_stage/CONTRIB_2026-08-19_from_the_277_083_lane_your_stage_2_output_is_the_shape_two_stuck_repair_queues_need.md`.
   See §5.
4. **`314`** — owner's call between the four candidates; candidate 1 is one line plus a credit cost
   somebody should size.
5. ~~**Two loose ends nobody owns**, both `[UNMEASURED]`~~ **BOTH MEASURED 2026-08-19 by
   `agentchassis-22` — see §10. (a)'s PREMISE WAS VOID; (b) has a population but no mechanism.**
   - ~~`page-rerender` saves to owned pages ~3,754 times without refusal while `page-build-handler` is
     refused every time. Same guard. One of those needs explaining.~~ **There was no asymmetry to
     explain.**
   - a page named/URL'd `tool-…` carrying `rebuild_policy='generic'` looks like a data defect;
     nobody has counted how many.

> ⚠ **CORRECTED 2026-08-19 — `copy-editor` is owned by the `copy_quality_two_stage` lane, NOT `loanandmortgagecalculator_couk`.** I got the wrong lane from a `grep -rl "copy-editor"` hit in LMC's `README_where_we_are.md` — a *mention* — and read it as ownership. `scripts/who-owns.py` exists to separate those two, and I did not run it. The defining evidence is what the commits shipping migrations `447`/`462` actually touch: `docs024_key_docs_latest/copy_quality_two_stage/`. Register entry **CQ-024**. A CONTRIB is filed in their lane dir (`CONTRIB_2026-08-19_from_the_277_083_lane_…`, commit `7574482c7`).


## 5. THE OUTBOUND ASK — filed 2026-08-19, awaiting a reply

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

## 9. OWED WORK from the peer exchange — one migration, and one measured warning already sent

### 9a. OWED: an explicit `copy_edit_proposed` exclusion in the promoter's `pre_query`, citing D2

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

### 9b. SENT: `473` will be refused on the 16 owned pages — warning delivered before the apply

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

### 9c. The discriminator worth keeping, from `agentchassis-8d`

> **Is success expressible as a diff assertion?**

"Are the asterisks gone and the text otherwise byte-identical" fully is; "does this read like a
person wrote it" cannot be. **A repair whose success condition is completely assertable does not need
a human-review posture — it needs a gate that fails loudly.** That is a sharper test than anything
this lane had written for when a repair needs HITL, and it independently argues for `473` over any
LLM route for this class.

---

## 10. §4.5 ANSWERED — and (a)'s premise was VOID, including in a figure I repeated

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

### (b) Population established, damage NOT — and they withdrew their own figure

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

## 7. Session-start checklist
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

## 8. The escalation clock, and other additive detail kept from a concurrent edit of this file

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
