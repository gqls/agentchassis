# NOTES — bugs_open/367 (running record, append-only, newest at the bottom)

## 2026-08-23 — re-verifying the bug before touching anything

Ran the live `classify` SQL (read from `agent_definitions`, both params bound) for item
`562788c3`. **`route=stale, component_id='', html_len=0`** — while component `0a1498b3`
sits at that (page, slot) with `build_status='pending'`, `locked_at` NULL, and **9,220
characters of `rendered_html`**. The finding is true: `n_still_empty=2` of `n_named=2`.
Bug confirmed still live.

Ownership checked: `scripts/who-owns.py 367` names `bugfix_342_absent_required`, but that
lane's own handoff banner says **LANE CLOSED**, names 367 as successor, and states *"Nothing
here needs picking up."* No competing session (`ListAgents` shows a `bugs_open/342` session
idle and a `bugs_open/344` session on a different residual).

## 2026-08-23 — MISSTEP 1: I measured `pages.sections` with the wrong shape and got a zero

To decide whether an "unresolvable target" route would ever be reachable, I asked how many
slots named in `pages.sections` have no `page_components` row. I unnested with
`jsonb_array_elements(...) s` and filtered `s->>'name' IS NOT NULL`.

**Result: 0 of 0.** I nearly wrote "the case is unreachable" into the design.

`pages.sections` is an array of **text** — elements look like `"hero"` — so `s->>'name'`
is NULL for every element and my filter discarded all 746 non-empty rows. The measurement
could not have come out any other way.

Redone with `jsonb_array_elements_text`, and vocabulary-checked before I trusted it (sampled
pages show `sections` entries and `page_components.slot_name` are the same vocabulary; 1,824
of 2,160 named slots DO resolve to a row, so the join is sound):

> **2,160** slots named in `pages.sections` on non-deleted pages; **336 (15.6%) have no
> `page_components` row at all**; 16 have a non-deployed, non-removed row.

So `count(comp)=0` is an ordinary state for about one planned slot in six — which is what
makes `close_stale`'s *"no longer exists"* a false statement quite apart from the
`build_status` narrowing. **Caught by:** the zero looked too round, so I ran a control on my
own query (`jsonb_typeof(sections->0)`) instead of on the world.

## 2026-08-23 — MISSTEP 2: I claimed a route "had fired exactly once ever" off a survivor artefact

I wrote that `stale` had been taken **once in the platform's history**, from
`SELECT min(created_at) FROM orchestration_states` reading 2026-07-19 — i.e. "the table goes
back a month, and there is only one `stale` in it".

**Refuted.** Retention is about **two days**: 08-22 has 1,324 rows, 08-23 has 3,299, and then
nothing at all until four days in July totalling **24 rows, every one `CANCELLED`** — stuck
rows the cleanup skips. There are **zero** rows for 2026-08-14→19.

**Caught by:** reading `bugfix_277_required_fields_repair/RUNBOOK_required_fields_repair.md`,
which records canary `332bb3f6` closing `complete/stale` on 2026-08-15 — a route my census
said had never happened. A minimum over a retained table is not a retention window.

The bug does not rest on this: the hand-run classify and the whole-population
re-classification are independent of the audit trail.

## 2026-08-23 — MISSTEP 3: I passed on a count that was my own `LIMIT`

I told an adversarial reviewer *"all 8 existing `content_rewrite:from_rfm:` conversions are
`failed`"*. The 8 was the `LIMIT 8` on my own sampling query. The real census is **31** — 28
`failed`, 2 `cancelled`, 1 `complete`. The direction held and the conclusion survived, but I
had passed a number off as a finding when it was an artefact of how I looked.

## 2026-08-23 — the design changed twice, and both changes came from measurement

I started on the bug file's own candidate 1 ("widen the `comp` CTE"). Two findings moved me
off it:

1. **Widening alone repairs nothing.** `file_rewrite` reads `spec.component_id`, `.page_id`,
   `.component_function`, `.reason`. Post-deploy items carry all four (62 of 62); render-time
   items carry none (0 of 3). Both `spec_paths` (`create_work_item_action.go:281,294`) and
   `item_key_suffix_field` (`:252-256`) are **deliberate hard errors** when unresolved. I had
   predicted a *degenerate colliding key*; the truth is better — it is a loud error, by a
   council-hardened design. Either way, widening buys a loud failure, not a repair.
2. **The repair arm is the wrong destination.** `content_rewrite`/`edit_live` runs at
   `page-build-handler`, whose `save_page_sections_action.go:823` DELETEs every
   agent-writable row on the page and `:1014` reinserts at `'deployed'`. 28 of 31 `from_rfm`
   conversions already fail there on the owned-page refusal (`bugs_open/333`, owned by the
   277 lane) — and the 367 page is `rebuild_policy=owned`.

So the rule became: **a disposer may close only on positive evidence of absence.** The
estate already states it at `revalidate_review_queue_action.go:684`.

## 2026-08-23 — the fix, proven before it was applied

Migration `574`. Everything below was run **inside a transaction that was then rolled back**,
so production was never used as the test rig:

- migration parses; its own verify block passes all 8 assertions incl. negative controls
- **C1** real item → `target_not_dispatchable`, `target_state=pending`, component RESOLVES
  (`0a1498b3`, `html_len=9220`, `n_still_empty=2`)
- **C2** retired slot (`tool-clip-path`/`ported-page`) → **still `stale`**,
  `target_state=component_retired`
- **C3** page that does not exist → **still `stale`**, `target_state=page_missing`
- whole population, all 65 items, old vs new: **exactly one route changes**
- **apply-then-rollback returns `default_config` BYTE-IDENTICAL**

Then applied for real, and re-verified by reading the query **back out of the live row**:
same three controls pass, same one-route delta. Council submitted first —
`d48c0a89-9ff8-4286-bfe9-2690dc13d5bc`.

Two SQL traps worth carrying: `snapshot_agent` is overloaded, so a bare literal gives
`function snapshot_agent(unknown) is not unique`; and `to_jsonb()` over adjacent
string literals is `unknown` — cast with `::text`.

## 2026-08-23 — what the fix does NOT do, stated so nobody reads more into it

The render-time population is now **visible and honest**, not repaired. It parks at
`needs_human_review` with the facts and the repair paths on the row. Repair needs
`bugs_open/333` plus a producer that writes the convert arm's read-set. Both named, neither
taken here. Do not let anyone write that 367 "made the render-time findings repairable".

## 2026-08-23 — my three WRONG_CALLS entries were swept into another lane's commit

Appended them, then committed by pathspec — and the pathspec matched nothing, because the file was
already clean. Another session had committed `WRONG_CALLS.md` in the gap between my append and my
commit, taking my three entries with it. They are at HEAD, in `bb1e144b5` (the `bugs_open/328`
lane's commit, whose subject is about anchor suppression and says nothing about any of this).

Nothing is lost and forward-only holds, so there is nothing to undo. Recording it because it is the
exact scenario CLAUDE.md describes — *"it cannot stop a session that still runs `git add -A` from
sweeping up yours"* — and because the very next commit on that file
(`a79a65b09`) is another lane writing up **the same thing happening to them**: *"I ran the
append-only gate right after APPENDING instead of right before COMMITTING, and swept four other
lanes' entries into cdfa3cb35."* Two lanes, one afternoon, same file.

The practical lesson for a shared append-only doc: **the window between appending and committing is
where your work belongs to everybody.** Append and commit in the same breath, and do not assume a
`git status` from thirty seconds ago still describes the file.

## 2026-08-23 — the council round: APPROVED round 1, and two of its objections were worth code

Correlation `d48c0a89-9ff8-4286-bfe9-2690dc13d5bc`. **APPROVED at round 1**, 14 advisory
objections, none high-severity. Approval is not "no objections", so here is what I did with
them.

**Acted on with code — migration `576`.** The `tomb` CTE matched the slot with
`COALESCE(pc2.slot_name,'') = COALESCE(item.spec->>'slot_name','')` — inherited from 410's
`comp` CTE. If BOTH sides are absent that compares `'' = ''` and matches, so an item with no
`slot_name` could be "retired" by an unrelated removed row with no `slot_name` on the same
page, and **closed** on it. The seat named the consequence exactly: *"an incorrect `stale`
close — exactly the outcome this migration exists to stop."*

I measured before patching, and it is **not reachable**: 0 of 38 removed rows and **0
`page_components` rows anywhere** have an empty `slot_name`, and the 2 items lacking
`slot_name` also lack `page_name`, so they classify `malformed` first. No row's disposition
changes. I shipped the guard anyway, because "unpopulated" and "unrepresentable" are not the
same thing and the difference is one clause.

**Acted on with code — the `_VERIFY` sidecar.** A seat pointed out that 574's verify block
checks only the *shape* of the rewritten JSON and never RUNS the SQL it assembles — the
estate's own *"embedded SQL is DATA to your migration's probe"* landmine. It was right about
the deeper thing too: I **had** run the behavioural checks, against the patched row inside a
rolled-back transaction, but they lived in my scratchpad, so a re-apply could not repeat them.
They are now
`574_required_fields_router_stops_closing_what_it_cannot_resolve_VERIFY.sql`, and I proved it
non-vacuous by applying the 574 ROLLBACK inside a transaction and requiring VERIFY to fail —
it did, loudly, then the transaction was discarded.

**Acted on with prose — the precedent I should have found myself.** A seat noted that
`bugs_closed/032` (2026-07-19), *"the completion verifier reads a DELETED component as a
successful fix"*, is functionally identical one layer over, and that its fix was *"return an
error, never a verdict, so the gate's fail-OPEN policy turns a false success into a visible
unknown"* — the same remedy in a verifier's vocabulary. **My plan was well-grounded and still
never asked whether the council had already ruled on this shape.** Now cited in `016b` §9. The
transferable bit: *grep the closed bugs for the SHAPE, not just the mechanism* — I grepped for
`build_status` and `required_fields_missing`, and 032 contains neither.

**Answered by measurement, no change needed.**
- *"Two active definition rows could exist and the UPDATE does not pin version"* (two seats):
  the migration's first premise is `count(*) <> 1 → RAISE`, so it aborts rather than guessing.
  Live count is 1.
- *"Does the revalidator actually cover a `needs_human_review` row filed by this router?"*:
  yes — `revalidate_review_queue_action.go:281` registers
  `"required_fields_missing": revalidateNamedFields("missing_fields")` and the whole file is
  the `needs_human_review` queue. The new park lands in a queue that IS re-validated.
- *"'removed' may have writers you did not find"*: exactly one Go path,
  `internal/core-manager/admin/page_admin_handlers.go:576` (which also locks what it removes),
  plus hand SQL such as `570_deactivate_testimonials`. My Risk 5 held.
- *"`pageComponentNotRemovedSQL` may not be the only equivalent"*: it is the only **named**
  one (`section_editor_actions.go:1537`) and my spelling matches it byte for byte — but there
  are **12** hand-typed `<> 'removed'` / `IS DISTINCT FROM 'removed'` occurrences in non-test
  Go as of 2026-08-23. The claim was precise and easy to misread; stating the 12 is the honest
  form.
- *"Are there OTHER disposers with the same unresolved⇒stale-close shape?"*: I did census this
  before designing — `required-fields-missing-handler` is the **only** live agent definition
  with a `stale` route and a `close_stale` step. The objection was that I had not *said* so.
  Now said, dated, and in `016b` §9.

**Not acted on, deliberately.** A seat asked for the workflow edits to be retagged
`config_change` rather than `add` in the submission metadata. Fair, and a real machine-checkable
improvement — but it is a change to how submissions are *tagged*, not to what shipped, and
retagging an approved round's plan after the fact would make the record less accurate, not more.
Worth doing on the next submission.

## 2026-08-24 — the fix survived the roll, and the last criterion was met by DRIVING it

**v1.0.1334 rolled 15:39Z.** The fix is DB config so a roll cannot carry or drop it, but a
re-seed of `410` would silently revert it — so the first thing I did was run the `_VERIFY`
sidecar against the live row. All three controls green. That is the whole point of having
built it: the post-roll check is one command instead of a re-derivation.

**No new item had been filed since 574 went in.** Zero `required_fields_missing` rows created
after 2026-08-23 18:00Z, and the router had not run since the wrong close at 17:09Z. That is
not a fault — the render-time producer only fires on a section edit, and the post-deploy check
on discovery rotation. But it meant the bug's closure criterion ("observed on a real item")
could sit unmet indefinitely while looking like patience.

**So I drove it, and it was a repair rather than a test.** I re-checked the finding first:
`headline` and `trust_note` still absent from `content_data`, component `0a1498b3` still
`pending`, still unlocked, still 9,220 chars, `updated_at` untouched since 2026-07-17. The row
was a **false negative sitting in the "actioned" bucket** — precisely the damage this bug
describes — so re-opening it is the correct disposal, not an experiment staged on production.

Guards checked before writing anything: the dedup key was free (0 non-terminal rows — the
wrong close had released it), and no other session had open `required_fields_missing` work on
that page (the 7 open items there are other types, none routed at this handler).

**Result, and the two runs are each other's control — same item, same component:**

| orchestration | when | `route` | `target_state` | `component_id` | `html_len` |
|---|---|---|---|---|---|
| `ab2cf74e` | 08-23 17:09Z | `stale` → closed `complete` | *(none)* | *(empty)* | 0 |
| `e2a6bb94` | 08-24 16:08Z | `target_not_dispatchable` → parked | `pending` | `0a1498b3…` | **9220** |

Lifecycle observed live: `triaged` → `claimed` → `needs_human_review` in about 100 seconds.
`attempt_count 1`, `triaged_by` the router, and — the part I care about most — **it HOLDS its
dedup key** (1 non-terminal row on the key). The close released it; the park holds it. That is
the anti-churn property as data rather than as intent.

**The canary is a committed, guarded file**
(`CANARY_2026-08-24_reopen_the_wrongly_closed_item.sql`), not a paste. It refuses if the
finding is no longer true, if any non-terminal row already holds the dedup key, or if the item
is not `complete`. So the next thread can re-run it, and it cannot file noise if the world has
moved on.

**367 → `bugs_closed/`.** Then I grepped `bugs_open/367` across the tree and retracted the four
places still calling it a live open defect (STY-057's landmine, the concept index row,
`bugs_closed/342`'s banner, the 342 lane's handoff). That is the estate's own lesson about a
closed blocker still being obeyed, and it costs one grep on the day you close rather than
someone else's fortnight later.

**What I did NOT claim, and it is worth repeating at the close:** the population is visible and
honest, **not repaired**. It parks for a human with three named resolutions on the row. Repair
needs `bugs_open/333` plus a producer that writes the convert arm's read-set.

## 2026-08-24 — MISSTEP 4: I swept another lane's register entry into my own commit

Committing `content-quality.md` by pathspec for a one-line CQ-023 status update, I also took
the `bugfix_381_inexpressive_composition` lane's entire new `### CQ-028` entry, which was
sitting uncommitted in the shared tree. It is at HEAD in `0a5c6b08e`, under a message about
bug 367 that says nothing about it.

**This is the one case a pathspec commit cannot prevent**, and CLAUDE.md says so in terms:
*"It cannot see a same-file passenger — if two sessions edit one file, whoever commits takes
both edits, and no hook can prevent that."* Yesterday my own `WRONG_CALLS` entries were swept
by someone else; today I did the sweeping. Same file class — a shared, append-heavy document
that many lanes touch.

**It also created a real collision**, which the pattern-check flagged *on my commit* — so it
reads as mine. `CQ-028` was already held by the 277 lane's `rendered_html_transform` entry
(`af0f00bb5`, 2026-08-20), referenced from ~10 files including `LANDMINES.md`, the concept
index, `bugs_open/333` and two council submissions. The incumbent keeps the id; next free is
`CQ-031`.

**What I did NOT do: renumber it.** Their entry names `bugs_open/381` and migrations `594`/`595`
as sources, that lane was mid-flight (four other register files dirty in the tree at that
moment), and renaming from outside would desync work I cannot see. I flagged it in place above
their heading, committed the correction separately (`3f6f5fcea`), and **messaged the 381
session directly** — because a note in a file they may not re-read is weaker than telling them,
and from their side the entry looks uncommitted.

**The check I should have run, and it is cheap:** `git status --short <file>` immediately before
a pathspec commit on any shared document, and read the diff you are about to make — `git diff
--cached <file>` after `git add`, or `git diff <file>` before. I had run `git status` minutes
earlier and treated it as current. On this tree it is a snapshot with a half-life of minutes,
which CLAUDE.md also says, and which I have now demonstrated in both directions in two days.

### Resolved same day — and the handling was the right call

The 381 lane renumbered to **CQ-031** within the hour, updated the index row and the one
cross-reference I could not have seen (`site-plan-and-reconciler.md`, PLAN-053's relations
line), and confirmed nothing of theirs shipped the string `CQ-028` across their lane docs, five
migrations (591–595) and `bugs_open/381`. Verified from here: `CQ-031` at line 357, incumbent
`CQ-028` untouched at 312, index row present.

**Their reply settles the question of whether flagging beat fixing**, and it is worth recording
because I was not certain at the time:

> *"You were right not to renumber from outside — PLAN-053 referenced it and you could not
> have seen that."*

That is the general form: **a rename from outside a lane is only safe if you can enumerate the
referrers, and you cannot enumerate another lane's uncommitted ones.** Flag, name the free
number, and hand it back.

On the sweep itself they said the telling was the whole available remedy — *"I'd have lost more
time discovering an empty diff on my own"* — which is the argument for messaging the session
rather than only writing a note in a file they may not re-read.

**One loop left open on purpose:** my flag block above their entry is now stale (it still says
"needs renumbering"), and I did NOT remove it, because `content-quality.md`,
`site-plan-and-reconciler.md` and `000_concept_index.md` were all dirty with their work at that
moment. Editing and committing that file is exactly how the sweep happened; doing it twice in
one afternoon to clean up after the first time would be absurd. Handed back to them with the
line numbers and a replacement one-liner, to take in the commit they were about to make anyway.
