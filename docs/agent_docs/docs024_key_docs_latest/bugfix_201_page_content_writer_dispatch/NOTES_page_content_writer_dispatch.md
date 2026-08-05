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

### State at end of session

Committed `37afbb847`, council `Council-Submitted: 71523705-07d1-4067-9c5d-af371ba84b89`
(**verdict not read** — owed). **Inert until a chassis roll**; the fleet is on `v1.0.1252`,
which predates the commit. Symptom 2 untouched by 201 §2's own instruction. Verification traps
for the next session are in RUNBOOK **R6** — three of them, and the site-lock one returns
success with zero items.
