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
3. **`277`'s remaining half** — the `no_content_data` repair. **Start with a conversation with the
   `loanandmortgagecalculator_couk` lane**, not a design. This is also Tier 2.
4. **`314`** — owner's call between the four candidates; candidate 1 is one line plus a credit cost
   somebody should size.
5. **Two loose ends nobody owns**, both `[UNMEASURED]`:
   - `page-rerender` saves to owned pages ~3,754 times without refusal while `page-build-handler` is
     refused every time. Same guard. One of those needs explaining.
   - a page named/URL'd `tool-…` carrying `rebuild_policy='generic'` looks like a data defect;
     nobody has counted how many.

## 5. Session-start checklist
`git log --oneline -10` · re-read this file from disk · `scripts/who-owns.py` **by slug** for `277`,
`083`, `300`, `301`, `307`, and **`copy-editor` belongs to another lane** · re-measure §1's probe ·
then §4 step 1.
