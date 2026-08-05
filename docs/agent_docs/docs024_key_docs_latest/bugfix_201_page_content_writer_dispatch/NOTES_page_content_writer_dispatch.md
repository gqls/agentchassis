# NOTES — `bugs_open/201` lane

Running record, append-only, newest at the bottom. Evidence, what the system actually said,
and every misstep.

## 2026-08-05 — lane opened from the 194 lane, fix-1 committed

### How this lane started, which is itself the lesson

I was on `bugfix_194` running its two owed checks. Check 3b needed a live dispatch that reaches
`save_page_sections` inside `site-work-orchestrator`'s build loop. Chasing a target for it, I
hit three walls in succession — and the third one was this bug, already filed by another lane
the day before. **Grepping `bugs_open/` for the mechanism before filing anything is the only
reason this lane is a contribution rather than a duplicate investigation.**

### What I did NOT re-derive

201 §3 asks that its failure mechanism not be re-verified — error text, `087` citation and the
artefact-level proof are all confirmed with correlation ids. I took that at face value and
spent the budget on its **open** question (§1) instead. Worth saying because the temptation to
re-run someone else's evidence is strong and it buys nothing.

### The cause is sharper than the error message says — and that matters

The error reads *"planned its own sections and none are ready"*, which sounds like a
data-readiness problem: the sections exist, their inputs are not available. **It is not.**

`plan_sections_action.go:867-875` has an **empty-input early return**:

```go
if len(sectionNames) == 0 {
    return map[string]interface{}{ ..., "ready_count": 0, "reason": "no sections to plan" }, nil
}
```

`sectionNames` comes from `inputs.GetRaw("sections")`, and `page-content-writer`'s self-plan
step binds that to `input_data.current_page.sections`. So the real failure is **"no sections
were supplied"**. Measured, all 14 `page-content-writer` items in history:

| item_type | `spec ? 'sections'` | keys | n |
|---|---|---|---|
| `literal_markdown` | **false** | check, findings, fix, original_pipeline, page_id, page_name, page_url | 12 |
| `placeholder_contact` | **false** | check, findings, original_pipeline, page_id, page_name, patterns | 2 |

A message that describes the *symptom state* rather than the *input state* sent the reader
looking at component `input_schema` data requirements. The disconfirming question was one
`spec ? 'sections'` away.

### §1's open question, answered — empirically, which beats the code-read

201 §1 asked whether `page-build-handler` returns `ready_count > 0` for an already-built page,
and proposed reading `load_spec_sections` to find out. I read it (it sources from
`site_specs.site_plan`, authoritative, `pages.sections` fallback — so it never depends on the
caller's spec), **but the stronger evidence was in the outcomes table**:

`page-build-handler`, items whose page already has `page_components` rows —
`content_rewrite` **19 complete**, `empty_section` **12 complete**, `empty_internal_href` **1**.
Against `page-content-writer`'s **zero** genuine successes in 14 items.

A code-read tells you what should happen; 32 completed rows tell you what does. Both were
cheap; only one of them can be wrong about a live system.

### The third check was an outright inconsistency, and I nearly missed it

`check_component_standards:477` files `ItemType: "needs_content_page"` — **the same item type
`write_build_items` produces** — and `write_build_items` routes that type at
`page-build-handler` (`load_work_item_actions.go:242-243`, where *every* entry in
`availableBuilders` resolves to `page-build-handler`). One item type, two handlers, decided by
which producer happened to file it. That one needs no argument about writers at all.

### A candidate 201 did not list, which I chose against — recorded so it is not re-proposed blind

`section-editor` is live and is exactly the right *shape*: `apply_section_edit`, contract
`domain + edit_type + (page_component_id | page_name+slot_name) + edit params`, and
`fix_component_template_action.go:27` says content changes are supposed to go through it. The
`literal_markdown` item even carries `findings[].slot_name` (`hero`) and `field`
(`subheadline`), so the target maps cleanly.

**It cannot serve, for a reason visible only by reading its step list:** `ensure_site_record →
spawn_deployer → load_edit_context → apply_edit → deploy_page → update_page_status →
trigger_deploy → complete`. **No LLM step.** It applies an edit someone else composed. And the
item's `fix` is an *instruction* — "Rewrite the affected fields WITHOUT markdown syntax… if the
writer wanted emphasis, re-word so the words carry it" — not a replacement string. Routing
there would have needed a new compose-then-edit agent: a new shared mechanism, architecture
scope, to fix a routing bug. Full trade-off in PLAN.

### Missteps, in order

- **I wrote SQL before reading the schema, twice, and both times it errored rather than
  misled.** `orchestration_states` has no `agent_type` (it is `owner_agent_type`);
  `jsonb_object_keys` cannot sit bare beside an aggregate. CLAUDE.md says `\d <table>` first
  and I skipped it. Cost: two round-trips. **The reason this was cheap is luck, not care** — an
  errored query is loud. The same haste against a column that *exists but means something else*
  is the expensive version, and this lane's sibling (194) has a `WRONG_CALLS` entry for exactly
  that (`created_at` vs `occurred_at`).
- **`who-owns.py 201` named MY OWN lane as the likely owner.** It reads commits and doc
  mentions; I had just written 12 mentions of 201 into `bugfix_194`'s files an hour earlier. The
  tool is honest and the signal was an artefact of my own action. **An ownership check reflects
  what has been WRITTEN, not who is working** — I checked live `.jsonl` transcripts as well,
  which is what actually answered it.
- **Two other sessions were live in `page-content-writer` code while I worked** (transcripts
  modified ~15 min before I looked). Neither mentioned 201 (`grep -c` → 0): one is the `156`
  dedup lane in `save_page_sections_action.go`, the other the concept-register drift lane. So
  201 was genuinely free — but I only know that because I looked at the *tree and transcripts*,
  not just at `git log`.
- **My first instinct on finding the site lock was to ask whether it could be released.** It
  can, and doing so would have re-run a filed incident (`aee11cb90`, a live homepage rebuilt
  under a held lock on that same site). The lock is the control added *because of* that. Noted
  because the instinct was wrong and immediate.

### The consequence I am declaring rather than burying

After this change, **no producer anywhere files a `page-content-writer` work item** — grep
across `platform/ internal/ pkg/` returns zero. `site-work-orchestrator`'s `build_items_loop`
is gated on exactly that handler, so it is now permanently unfeedable.

I am **not** repointing that filter at `page-build-handler`. `build-dispatch-loop` already
consumes those rows; a second consumer of the same rows is a double-dispatch invitation, and
this estate has filed history in that class. The loop was already unreachable in practice — it
has never run (absent from `agent_run_stats`, which has no reaper and does track
orchestrator-shaped agents), and its only possible inputs were the items that hard-fail. **This
makes an existing deadness explicit; it does not create one.** Written into the bug file as an
explicit do-not-tidy, because it is precisely the kind of thing a later session "cleans up".

> **[INFERRED — not measured]** That loop being unfeedable *may* mean seed 312's
> `site-work-orchestrator` mapping (from `bugs_closed/194`) can never be exercised on that
> route. I have not run it and am not asserting it. Flagged in 194's NOTES and in the CONTRIB.

### The council verdict — APPROVED, and two of its five objections were substantive

`decided_by: "approved with 5 advisory objection(s) — none high-severity"`, 15 reviewers,
2 abstained, 0 unreadable, `gated_by_truncation: false`. The run dispatched in **under a
minute** (submitted ~11:41Z, first seat executing 11:41:45Z, verdict by ~11:47Z) — nothing like
the 29-minute queue CLAUDE.md warns about, so budget by the worst case but do not assume it.

Approval is not the interesting part. Acting on the objections is.

**1. `editquality` [medium] — WAS RIGHT, and I had understated the cost in my own PLAN.**
It cited a landmine I had not read: `page-build-handler`'s writer never sees the page's own
stored prose unless `spec.mode="recreate"`, which none of these three checks sets. Read it
(`LANDMINES.md:4433`) and it is confirmed root cause on `bugs_open/178`, not a suspicion:
`load_existing_content_action.go:64-69` no-ops with `{"has_existing": false, "reason":
"not_recreate"}`, and `load_page_record` carries only sections/title/page_type — no prose. **So
the repair rewrites the section from scratch and prior prose is lost.** My PLAN had called this
"heavier than the ideal repair", which is too soft; corrected in place there with the reasoning
for why the decision still stands (the alternative is a permanent defect, not a field edit).
**And setting `mode=recreate` is explicitly the wrong fix** — it sources the adoption-crawl
snapshot, i.e. stale content rather than none.

**2. `guardian` [medium ×2] + `bug_historian` [medium] — the SECOND dispatch gate. Checked; clear.**
The objection was the one this session's own SessionStart landmine warns about: a dispatcher has
two gates, and fixing the visible one leaves the key dropped. Verified rather than argued:

- `scripts/audit-relay-gaps.sh` — 175 agents decoded, **0 relay-gap findings**;
  `build-dispatch-loop` is not even among the 2 uncovered dispatcher-shaped relays, so it is
  asserted.
- `build-dispatch-loop.call_handler` `input_mapping` forwards `domain` ← `input_data.domain` and
  `site_id` ← `current_item.site_id` — **both** of `page-build-handler`'s `input_contract`
  required fields (`{required: [site_id, domain], optional: [page_name, page_id, sections]}`) —
  plus `spec`, `current_page`, `item_type`, `work_item_id`, `page_name?`. It is
  **item-type-agnostic**: there is no per-`item_type` allow-list to exclude the new types.
- `spawn_handler` uses `agent_type_field: current_item.handler_agent`, so the spawn follows the
  ROW. New rows carry the new handler; **existing rows do not** (RUNBOOK R6 trap 1).
- ⚠ **Gotcha worth recording:** reading these steps with a top-level `#>> '{workflow,steps,…}'`
  returned EMPTY for both, because they are nested in a loop `sub_workflow`. An empty result
  there reads exactly like "the step does not exist". `jsonb_path_query(…, '$.**.steps')` — the
  same trap as R2, hit again in the same session.

**3. `bug_historian`'s deeper point stands and is NOT closed by the above.** It observed that
this trades a LOUD hard-fail for a pipeline with filed history of *silent* partial success
(016b §9: sections deferred and dropped; "page build completes having built nothing"). My
evidence for `page-build-handler` was **by analogy** — different item types succeeding — not
these item types. That is a fair characterisation and I have not closed it. It is why RUNBOOK
R6 trap 3 requires `content_data` to change rather than accepting `complete`.

**4. `debug_historian` [medium] — no post-deploy pod-grep was proposed.** Correct. I had written
"inert until a roll" but never named the artefact check. Added to RUNBOOK R6 and the bug file:
grep the RUNNING pod's binary for the new literal, every replica, never git or the tag.

**5. `prior_art_librarian` [low] — my "site-work-orchestrator has never run" claim.** It rests
on absence from `agent_run_stats`, which the council cannot see, and there is a standing
landmine about "has agent X ever run" claims being false negatives from `orchestration_states`'
~24h retention. **I used `agent_run_stats` deliberately BECAUSE of that retention trap** (it has
no reaper) — but the seat is right that the claim is unverifiable from its side and should not
be repeated in a close-out unchecked. Marked `[UNVERIFIABLE-BY-COUNCIL]` where it appears.

**6. `architecture` [medium] — the RFC-shaped finding, filed as RFC_014.** This is the FIFTH
site of the same defect class, and the only guard checks that the string names a KNOWN agent,
not that the agent can CONSUME the filed spec shape. Seat's verdict: "Approve the edits as
written … ship it", but on record. `RFC_014_handleragent_is_a_stringly_typed_routing_contract.md`
costs three options, recommending the cheap floor (a narrower legal-direct-dispatch set) over
continuing to patch strings one site at a time.

`editquality`'s low objection — that edit 4 is comment-only and should not count as an edit — is
accepted without argument. It was documentation upkeep bundled into the count.

### State at end of session

Committed `37afbb847`; council **APPROVED** (`71523705-07d1-4067-9c5d-af371ba84b89`), verdict
read and its objections acted on above. The commit carries `Council-Submitted:` — correct for a
pre-verdict commit, and `098` credits it automatically now the correlation is approved; **no
amend, forward-only.** I have deliberately NOT written a `Council-Reviewed:` trailer onto a
later commit, because that trailer belongs to the commit carrying the reviewed code and
back-dating it is the report's MISMATCH surface.

**Inert until a chassis roll** — the fleet is on `v1.0.1252`, which predates the commit.
Symptom 2 untouched by 201 §2's own instruction. Verification traps for the next session are in
RUNBOOK **R6** (now four: the stale `handler_agent` on existing rows, the locked site returning
success-with-zero-items, `complete` not being proof, and **prior prose legitimately vanishing**)
plus **R7**, the post-deploy pod-grep the council asked for.

Filed out of this round: **RFC_014** (the stringly-typed routing contract, five recurrences).

### One last thing I got wrong today, worth its own line

I ran `who-owns.py 201` and it named **my own lane**. I noticed and discounted it, which is the
only reason it did not mislead me — but the mechanism is worth stating: I had written 12
mentions of 201 into `bugfix_194`'s files an hour earlier, and the tool ranks by mentions and
recent commits. **An action I took to be diligent (cross-referencing the bug thoroughly) made
the ownership check point at me.** The generalisable form is already in memory as "your action
moves you to the back of the selector"; this is the same shape pointing forward instead of back.
The check that actually answered ownership was grepping live `.jsonl` transcripts and finding
two concurrent sessions in `page-content-writer` code, neither of which mentioned 201.
