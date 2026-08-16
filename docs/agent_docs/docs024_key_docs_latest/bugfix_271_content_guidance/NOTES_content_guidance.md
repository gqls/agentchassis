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

## 2026-08-15 — council APPROVED round 1 (corr `b24608e8`), 3 advisory objections, 16 seats

`decided_by: approved with 3 advisory objection(s) — none high-severity`. Four
seats abstained (relevance filter). Approval was NOT the useful part; five seats
raised something worth acting on, and two of those were right about a real gap.

**Acted on:**

1. **guardian (medium) — "a comment-only fence is soft; nothing prevents a later
   edit from reading 'one narrow exception' as licence for a second, less
   disciplined one."** Correct, and the sharpest objection in the round. `T4`
   bounds the ALIAS; nothing bounded the LOADER. Added
   `TestLoadWorkItems_SpecPassesThroughExceptTheAlias`: the emitted spec map must
   deep-equal the stored JSON plus exactly the one aliased key, over a fixture
   carrying routing ids, a nested object, an array, a bool and a number.
   **Mutation-proven** — injecting a second `specMap[...] = ...` next to the
   alias call fails it with the message naming the hazard. A second exception
   added anywhere on that path now fails here even if its own tests pass.
2. **architecture (low) — "worth a doc-notes landmine entry so a future session
   doesn't rediscover this by grep."** Written and dispatched
   (`LANDMINES.md#a-live-agent-prompt-still-asks-the-llm-for-contentguidance…`,
   verification correlation `46be0bcb`). The entry's point is the one the seat
   missed: the grep hit a future session WILL find is `content-gap-planner`'s own
   prompt asking the LLM for the key, which reads exactly like a live consumer.
3. **editquality (medium) — "the risks section claims a concept-register entry
   but no edit writes one; either an omission or asserted-not-delivered."** Fair
   from the submission alone: the 097 script refuses docs client-side, so
   register edits cannot appear in an edit list. It was delivered as WDS-016 in
   `1a4e47b9b` — but the seat could not know that, and it landed one commit after
   the code, which is its own miss (logged in WRONG_CALLS).
4. **reuse_agent (low) — "check whether datahelpers already has a coalesce /
   alias-field utility before hand-rolling a fifth variant."** Checked rather
   than assumed: no such primitive exists. `GetBoolFieldLoud` and the
   `ExtractNestedField*` family READ; `syncCoreFieldsToInputData` is unexported
   in another package; `setRoutingField` writes to the ITEM map, not the spec,
   and is the analogy the new function is placed beside deliberately. Answered,
   no change.
5. **bug_historian (medium) — "the fix is emitter-specific, not
   mechanism-level; a fifth emitter choosing a THIRD spelling reproduces 271
   exactly, and a human should note the residual rather than let this be read as
   closing the class."** Accepted in full, and stated in WDS-016 and the bug
   file. The generic root — `input_mapping` resolves exactly one source path and
   has no coalesce, so any mismatched key vanishes silently — is untouched by
   this fix and by `bugs_open/154`'s before it. Closing THAT would need a
   declared vocabulary for item-spec keys, which does not exist (the same
   registry `bugs_open/279` names as its own residual).

**Declined, with reason:**

- **debug_historian (medium) — "add a pod-grep for the `aliasGuidanceIntoSuggestion`
  symbol to the close-out."** Not taken, because the recipe it recommends is the
  one CLAUDE.md **retired on 2026-08-11** after it produced three confidently
  wrong readings in a day: `strings` is absent from the debian-slim images, and
  behind the customary `2>/dev/null` its failure is indistinguishable from "not
  stamped". The checklist in §9 already does the current, stronger thing — ask
  the service for its own `build provenance` line and
  `git merge-base --is-ancestor 9a7d23c49 <stamp>`, per service, with the
  known-sha `/proc/1/exe` probe (present AND absent controls) as the fallback
  when the startup line has scrolled. The seat's underlying point — "unit tests
  are not deployment evidence" — is right and is exactly what that checklist is.

## 2026-08-15 — a fresh roll landed and it does NOT carry the fix (the landmine, live)

Told a fresh chassis build had been deployed, which was true — **v1.0.1303** —
and the obvious next move was to run the §9 canary against it. Checked first, and
it would have been wasted: the pods started **18:45:33Z / 18:45:58Z** and the fix
commit `9a7d23c49` is stamped **20:42:42Z**. The build predates the fix by 1h57m.
`deploy/agent-chassis` is settled on 1303 with nothing rolling, so no newer image
is inbound either.

Worth recording because the failure would have been *convincing*: a canary filed
against 1303 comes back with the sentinel absent from every prompt — which is
byte-for-byte the pre-fix symptom — and the natural read is "the fix does not
work" rather than "the fix is not there". One `kubectl get pods -o jsonpath`
against the commit timestamp separates those two worlds in a single command, and
it needs no provenance line at all.

**Method note, because the documented recipe did not apply cleanly:** the
`build provenance` startup line had already scrolled out of `--tail=400` on this
service (expected — CLAUDE.md measures it absent at `--tail=3000` on the chassis),
and an empty grep there means "not in range", never "unstamped". The cheapest
sound substitute when the question is only *"could this binary contain commit
X?"* is **pod start time vs commit time** — a container that started before a
commit existed cannot contain it, no stamp required. Keep the binary probe for
the harder question ("which of two candidate commits is in there"), and only ever
with a present-AND-absent control pair.

## 2026-08-15 — the landmine verifier returned NEEDS_HUMAN_REVIEW, and it is a STALENESS artefact not a refutation

Verdict on `LANDMINES.md#a-live-agent-prompt-still-asks-the-llm-for-contentguidance…`
(correlation `46be0bcb`, verdict written 20:52:51Z):

> Core footprint files (`load_work_item_actions.go`, `apply_gap_plan_action.go`)
> and all named agents/item types confirmed present at a85ad401; `suggestion` key
> usage consistent with the entry's claims; however, the
> `aliasGuidanceIntoSuggestion` function (introduced in 9a7d23c49, 2026-08-15)
> postdates the index snapshot (2026-08-12) and cann[ot be verified]

So everything it *could* check, it confirmed; the one thing it could not is the
function committed forty minutes earlier. **`[per the verifier, not independently
measured]` the code index snapshot is 2026-08-12, three days behind HEAD.** That
matches the known recurrence shape — the index pin has to name the LIVE working
branch or it silently drifts — and it means **any** landmine entry citing code
from the last three days will draw the same verdict for the same reason.

Recording it because the verdict word is alarming and the content is not: a
reader skimming for `NEEDS_HUMAN_REVIEW` would reasonably infer the entry is
wrong. It is not — it is unverifiable-in-part by a tool whose ground truth
predates the code. The honest close is that the entry stands on the parts
verified at `a85ad401` plus this session's own first-hand reads, and the
alias half re-verifies for free once the index is repinned.

## 2026-08-16 — VERIFIED LIVE at the artefact on v1.0.1304

The owner cut a release; `agent-chassis` moved 1303 → **v1.0.1304**, pods started
10:41:27Z / 10:41:48Z.

**Establishing the binary took two attempts and the first was wrong.** The
`build provenance` startup line was already out of retained logs — not just out of
`--tail=2000`, but absent from `--since-time=<pod start>` as well, so container log
rotation had eaten it. The chassis DOES emit it (`cmd/agent-chassis/main.go:53`),
so this is a retention limit, not a missing stamp.

### Misstep 4 — a discovery grep for "some hex string" nearly made me report the opposite

Falling back, I grepped the captured logs with
`grep -oiE '(commit|git_sha|revision)"?[: ]+"?[0-9a-f]{7,40}'` and got
`commit a85ad401`, five times over. `git merge-base --is-ancestor 9a7d23c49
a85ad401` then answered **"the fix is NOT in this build"** — a clean, confident,
completely wrong result I was one sentence from reporting.

`a85ad401` is the **code-index snapshot commit** (2026-08-12), quoted inside the
landmine-verifier's own verdict prose, which the chassis logs. It is not a build
stamp and never was. CLAUDE.md warns about exactly this — *never a discovery grep
for "some 40-hex string"* — and I used one anyway, on log text rather than a
binary, which is the same failure in a new place.

**The check:** a provenance read must be anchored to the LINE that emits it
(`"msg":"build provenance"`), never to a pattern that matches any hex-looking
token in the stream. Logs carry other systems' commit ids as CONTENT.

**What worked instead** — a symbol probe with both controls in one breath:

| probe | result |
|---|---|
| `aliasGuidanceIntoSuggestion` (subject) | PRESENT on both pods |
| `setRoutingField` (positive control, weeks old) | PRESENT |
| `aliasGuidanceIntoSuggestionZZZ` (negative control) | ABSENT |

This also **reverses my earlier decline of the council's `debug_historian`
objection**. That seat asked for a pod-level symbol probe and I declined it as a
retired recipe. The retired part was `strings` (absent from debian-slim) and
same-tag assumptions; `grep -a` on `/proc/1/exe` with a present-and-absent control
pair is CLAUDE.md's own sanctioned fallback, and when the log line has rotated it
is the ONLY method left. The seat was right and my reason for declining was too
broad.

### The canary — positive, at the artefact

Filed on `pool-energy-utilities.internal` (unserved internal pool site, 0 deployed
pages, quiet 6 days — nothing customer-facing at risk), page `faq`, brief in the
**DEAD spelling only** (`content_guidance`, no `suggestion` key — verified at
insert time), sentinel `heliotrope kettledrum`. Baseline: 1 writer call in the
prior 3h, **0** sentinel hits.

Result — claimed after 120s, sentinel in the prompt 135s later:

| measure | value |
|---|---|
| page-content-writer calls in window | 2 |
| prompts containing the sentinel | **2 of 2** |
| prompts containing `## Rewrite Guidance` | **2 of 2** |
| stored `page_components` containing the phrase | **2** |
| item final status | `complete`, no error |

A brief that could ONLY have arrived via the alias reached the prompt AND the
stored artefact. That is the fix working end to end on the live fleet.

### Two measurement notes worth keeping

- **`llm_call_log.work_item_id` is NULL on child-agent calls.** Fable's plan
  marked this `[UNVERIFIED]` and deliberately avoided relying on it; now measured
  — writer calls spawned under a `page-build-handler` item carry **no**
  `work_item_id`, so it cannot scope a canary. Scope by time window plus a
  content discriminator instead.
- **I contaminated my own negative control and caught it late.** The control
  (an item with NEITHER key, which must gain no `## Rewrite Guidance` heading) is
  measured over a time window — and I re-triaged 25 guidance-CARRYING items into
  that same window minutes later. Their prompts will legitimately contain the
  heading, so a window-scoped count would have read as the control failing.
  The window happened to still be empty when I noticed. **The check:** before
  filing a negative control, ask what else you are about to inject into its
  measurement window — a control is only a control if nothing else can produce
  its positive signal.
