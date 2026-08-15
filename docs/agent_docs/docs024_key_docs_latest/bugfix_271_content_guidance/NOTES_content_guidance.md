# NOTES — bugfix 271, `content_guidance` has no reader

Append-only, newest at the bottom. Missteps are the point: record them where they
were made, with the check that would have caught them.

---

## 2026-08-15 — session start: picking the bug, and why not 279

Asked to take the next open bug no other thread is on; the owner suggested
`bugs_open/279`. **279 is taken and nearly finished** — `scripts/who-owns.py 279`
returns OWNED (`vigilant_designer_offer_analysis`, ACTIVE, 69 commits/14d), and
the file's own tail is a POST-ROLL CHECKLIST: its Go half is committed
(`d6d56e540`) and waiting on a chassis roll. Nothing to take.

`who-owns.py` reads COMMITS, so it cannot see a session mid-fix. Cross-checked
against the live session transcripts (~45 `.jsonl` files touched in the last six
hours, tallying `bugs_open/NNN` mentions per session). That put a named session
on 248, 281, 263, 184, 160, 270, 260, 283, 122, 278, 276, 275, 268, 277, 033,
274, 213, 264, 114, 238, 258, 259, 272, 279, 040, 115, 149, 185, 224, 113, 245,
282, 189, 280, 284, 285 — and **nobody on 271**, the newest such bug.

`who-owns.py 271` still says OWNED, because the lane that FILED it
(`ai_site_selling_automation`) is active and cites it. Read their handoff before
believing that: all three mentions are *route-around* notes ("check
`bugs_open/271`'s state before relying on item-spec steering"), not work in
flight. Filed-by is not owned-by. Taking it.

## 2026-08-15 — the bug is still valid, and the bug file understates the fix

`grep -rn content_guidance --include=*.go platform/ internal/ pkg/` at HEAD:
four writers (`apply_gap_plan_action.go:209`, `tool_content_item.go:170`,
`create_tool_component_action.go:448`, `deploy_tool_action.go:575`), zero readers
on the work-item path. The one read — `apply_gap_plan_action.go:178` — takes the
string off the gap PLAN and copies it into the item spec, which is where it dies.
Still valid, unchanged since filing.

**What the bug file did not know: a working steering channel already exists under
a different spelling.** Chain proven end to end (queries in the RUNBOOK):

```
site_work_items.spec  →  LoadWorkItemsAction parses it to a map (load_work_item_actions.go ~740)
                      →  build-dispatch-loop: "spec": "current_item.spec"   [inside a loop sub_workflow]
                      →  page-build-handler.call_content_writer: "rewrite_guidance?": "input_data.spec.suggestion"
                      →  page-content-writer prompt: {{if .rewrite_guidance}}## Rewrite Guidance (IMPORTANT…)
```

So this is not "a field with no reader" in the abstract — it is **two spellings
of one channel, one of which is wired**. `suggestion` is the fleet convention
(175 of 231 `content_rewrite` rows; every other item type uses it exclusively),
and ~90 live rows carry a brief ONLY in the dead spelling. That reframes the fix
from 271's candidate 1 ("inject guidance into the section brief in
`plan_sections`") toward converging the spellings at a choke point — which is
also the cheaper and more general change. 271 §5 is not wrong, it was written
without the `suggestion` half.

### Misstep 1 — a `jsonb_each` over `workflow->steps` cannot see a dispatcher

Asked "who maps the whole item spec into a child run?" with a join over
`jsonb_each(default_config->'workflow'->'steps')` and got **zero rows, twice**,
for two different spellings. Zero rows read as "nothing does this", and the next
step would have been to conclude the spec reaches `input_data` some other way.

It does not: `jsonb_each` descends ONE level, and every dispatcher nests its real
steps under `config.sub_workflow.steps`. The mapping was there the whole time
(`build-dispatch-loop`). **The check that would have caught it:** when a
structural query about "does anything do X" returns zero on a platform that
demonstrably does X, re-ask it as a text pattern over the whole config before
believing the zero — `default_config::text ~ '"spec\??": "[a-z_.]*spec"'` found
it in one query. A nesting-blind query returns a well-formed answer about a
subset it never names.

### Misstep 2 — nearly took `mentions_content_guidance = true` as a reader

`content-gap-planner`'s config matches `%content_guidance%`, which looks like a
consumer. It is that agent's own PROMPT asking the LLM to emit the key — the
false affordance itself. A `LIKE` hit on a config blob proves the string is
present, never that anything reads it. Followed each hit to the step that
consumes it before counting it (§3 of the RUNBOOK).

## 2026-08-15 — plan handed to fable

Evidence pack (chain, census, the `setRoutingField` precedent from
`bugs_open/154` and its "the spec map is NEVER mutated" invariant) handed to a
fable-model agent to rank fix candidates and produce a council-shaped edit list.
Its brief explicitly requires it to re-verify my reads rather than inherit them,
and to confront the never-mutate invariant head-on rather than route around it.

## 2026-08-15 — fable's plan corrected three of my reads, two of them load-bearing

Handed the evidence pack over; got back a plan that amended me on three counts.
I re-derived all three myself before believing either of us:

1. **Three live readers of `spec.suggestion`, not one.** Mine was an
   `input_mapping`-only query, so the two PROMPT-TEXT readers
   (`content-gap-planner`, `css-patch-agent`) were structurally invisible to it.
   This is the misstep that mattered — the entire safety argument is "every
   reader renders prose, none gates a branch", and a mapping-only enumeration
   cannot support a claim about all readers. Logged in `WRONG_CALLS.md`.
   Re-derived: `default_config::text LIKE '%spec.suggestion%'` → exactly 3.
2. **`needs_content_page` has drifted to 113 rows** (34 guidance-only) since the
   number I quoted an hour earlier. A census on this fleet ages in hours.
3. **`needs_content_page` DOES reach a guidance reader** once aliased — all 113
   rows carry `handler_agent='page-build-handler'`, whose `rewrite_guidance`
   path is item-type-agnostic. I had assumed it might have no reader at all.

Fable also found the decisive fact I had not looked for: **all 90 guidance-only
rows are terminal or parked, ZERO in `triaged`/`approved`** — and the loader
selects only those two statuses (`:651`). That is what makes "nothing in flight
changes on roll day" a measurement instead of a hope, and it is why the fix
ships without a default-OFF switch.

## 2026-08-15 — shipped, and the two mutations that prove the tests

Committed `9a7d23c49` (6 files, explicit pathspec — two other sessions had edits
in the same directory, so the pathspec named files, never the directory).

Mutation-proved rather than trusting green:
- Deleted the CALL SITE → only `TestLoadWorkItems_AliasesGuidanceOnTheLoadPath`
  failed, naming the cause. The five unit tests stayed green, which is the whole
  point of having it: a helper with no callers looks exactly like a finished
  refactor.
- Removed the PRECEDENCE check → the never-overwrite and non-string tests
  failed.
Both restored, full package re-run green (including the other session's
untracked `plan_sections_item_fields_dialect_test.go`).

### Misstep 3 — I wrote a ratchet fixture that could not fail

The first version of the ratchet's "these lines must match" list contained
`spec["content_guidance"] = brief`, which the key-position regex does not match
— so I "fixed" it with `if !pattern.MatchString(bad) && !strings.Contains(bad,
"] =")`. That escape hatch made the case pass unconditionally: the exact
quiet-test antipattern the fixture exists to prevent, written into the fixture
itself. Corrected by widening the pattern to cover index-assignment writes
(`\[\s*"content_guidance"\s*\]\s*=[^=]`) and deleting the escape hatch, with
`.(string)` reads and `==` comparisons kept in the must-NOT-match list so the
guard still cannot convict the live gap-plan contract at
`apply_gap_plan_action.go:178`. **The check:** when a fixture case does not
match, fix the PATTERN or drop the case — never add a condition that makes the
assertion vacuous.

### Noted, not acted on: two pre-existing pattern-check advisories

The commit hook flagged `unrepaired-component-write` on
`create_tool_component_action.go` and `deploy_tool_action.go`, and
`logged-model-output` on `tool_content_item.go:141`. All three predate this
change — I renamed a map key in each file and touched none of those lines. They
belong to `bugs_open/136`; recorded here rather than silently widening this
lane's scope.
