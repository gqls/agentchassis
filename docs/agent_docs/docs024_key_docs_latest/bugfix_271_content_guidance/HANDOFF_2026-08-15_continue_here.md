# HANDOFF — bugfix 271 (`content_guidance` has no reader) · cold start for a fresh session

**Written 2026-08-15 evening.** Everything is committed; nothing is in flight in
a dirty tree. **The work left is one thing: verify the fix at the artefact after
the NEXT chassis roll, then close the bug.** Read §1 and §5 and you can start.

---

## 1. State in one paragraph

`bugs_open/271` — a rewrite brief written into `site_work_items.spec.content_guidance`
was read by nothing, so the instruction never reached the writer's prompt and the
rewrite happened anyway, reporting `complete`. **Fixed in code, council-APPROVED,
NOT live.** The fix aliases the dead key onto `spec.suggestion`, which is the
channel that already works end to end. Four commits, all on
`087_towards_multiple_domains`:

| commit | what |
|---|---|
| `9a7d23c49` | the fix: alias + four emitter renames + tests + the invariant narrowing |
| `1a4e47b9b` | register WDS-016, bug-file §9, WRONG_CALLS, workstream docs |
| `fa065c045` | council follow-up: the loader fence test, WDS-016 residual, LANDMINE |
| `32232ce4f` | 016b §9 transferable pattern, README correction |

Council: **APPROVED round 1**, correlation `b24608e8-4fb1-4028-9512-86af2ef788b7`
(trailer already on `fa065c045`; the earlier two carry `Council-Submitted:`, which
`098` credits automatically).

## 2. ⚠ The 2026-08-15 evening roll does NOT carry it — do not skip this

A fresh build **was** deployed that evening. It is **v1.0.1303**, and it
**predates the fix by 1h57m**:

```
commit 9a7d23c49  = 2026-08-15 20:42:42 UTC
pods started      = 18:45:33Z / 18:45:58Z   image v1.0.1303   (deployment settled, nothing pending)
```

A canary run against 1303 returns the sentinel **absent from every prompt** —
which is byte-identical to the pre-fix symptom — and the natural misreading is
"the fix does not work" rather than "the fix is not there". **Establish the
binary before interpreting any canary.**

Cheapest sound check when the question is only *"could this binary contain commit
X?"*: `kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath=…startTime,image`
against `git log -1 --format=%cI <commit>` — a container that started before a
commit existed cannot contain it, and it needs no provenance line. The
`build provenance` startup line has already scrolled out of `--tail=400` on this
service; an empty grep there means **not in range**, never "unstamped".

**What is needed:** `IMAGE_TAG` bumped and a new build (v1.0.1304+) from a HEAD
containing `9a7d23c49`. Releases are whole-fleet and **the owner runs
`make release`** — do not roll a single service at its own tag.

## 3. What the fix actually is (so you can judge the canary)

The chain that already worked, all four hops verified live 2026-08-15:

```
site_work_items.spec
  → LoadWorkItemsAction parses spec to a map        (load_work_item_actions.go ~:744)
  → build-dispatch-loop:  "spec": "current_item.spec"      [inside a loop sub_workflow]
  → page-build-handler:   "rewrite_guidance?": "input_data.spec.suggestion"
  → page-content-writer:  {{if .rewrite_guidance}}## Rewrite Guidance (IMPORTANT…){{end}}
```

`aliasGuidanceIntoSuggestion` fills `suggestion` from `content_guidance` when
suggestion is absent — **in memory only, the DB row is never written**, so the
stored `spec` still shows the old key and a SQL-only reader sees no change. It
writes at most that one key, never over an existing value, never from an empty or
non-string guidance, and never materialises `""`.

## 4. Files, if you need to read the code

- `platform/orchestration/actions/load_work_item_actions.go` — `aliasGuidanceIntoSuggestion`
  (beside `setRoutingField`), its call site in the unmarshal branch, and the
  **narrowed invariant** comment (~:629, "ONE NARROW EXCEPTION").
- `platform/orchestration/actions/load_work_items_guidance_alias_test.go` — 8 tests.
  Two are the fences: `TestAliasGuidance_WritesAtMostOneKey` bounds the alias;
  `TestLoadWorkItems_SpecPassesThroughExceptTheAlias` bounds the **loader** (the
  emitted spec must deep-equal the stored JSON plus exactly the one key).
- The four emitters now writing `suggestion`: `apply_gap_plan_action.go`,
  `tool_content_item.go`, `create_tool_component_action.go`, `deploy_tool_action.go`.

## 5. THE REMAINING WORK — post-roll verification, in order

1. **Stamp.** `git merge-base --is-ancestor 9a7d23c49 <stamp>` per service (§2 for
   how to get the stamp when the log line has scrolled).
2. **Baseline first**, so the later result is disconfirmable:
   ```sql
   SELECT count(*) FROM llm_call_log
   WHERE agent_type='page-content-writer' AND prompt_rendered LIKE '%<sentinel>%';  -- must be 0
   ```
   Pick a sentinel phrase that greps zero in the site's register AND existing copy.
3. **Canary, in the DEAD spelling only** — this is what discriminates the ALIAS
   from the emitter renames:
   `spec = {"page_name":"<canary>","mode":"edit_live","content_guidance":"Include the exact phrase '<sentinel>' in this section."}`
   Item type `content_rewrite`, handler `page-build-handler`, status `triaged`.
4. **Assert at the artefact:** the sentinel appears in `llm_call_log.prompt_rendered`
   alongside `## Rewrite Guidance`, then on the served page (cache-busted).
   ⚠ Per §8 of the bug file, the item's own status may read `failed` while the page
   deployed fine (the spawn→call handshake race) — **verify at the artefact, never
   the status**.
5. **Negative control in the same window:** an item with neither key must gain no
   `## Rewrite Guidance` heading.
6. **Then close:** move `bugs_open/271…md` → `bugs_closed/` naming **both paths on
   the commit** (`git mv` + a pathspec commit silently ships a copy — see LANDMINES),
   verify with `git ls-tree -r --name-only HEAD -- bugs_open/ bugs_closed/ | grep 271`
   returning exactly one line, and promote **WDS-016**'s status from
   `built, NOT live` to `deployed` with the evidence inline.

## 6. What is deliberately NOT done, so you do not re-litigate it

- **No opt-in switch.** Measured: all 90 guidance-only rows are terminal or parked,
  **zero** in `triaged`/`approved`, and the loader selects only those two statuses.
  Nothing in flight changes on roll day. A default-OFF switch would re-create the
  dead channel behind a flag (owner ruling 2026-07-29 §2 declines to require one).
- **The class is NOT closed** (council `bug_historian`, accepted): `input_mapping`
  resolves exactly one source path and has no coalesce, so **any** mismatched key
  still vanishes silently — a fifth producer with a THIRD spelling reproduces 271
  exactly. Closing that needs a declared vocabulary for item-spec keys, which does
  not exist; `bugs_open/279` records the same missing registry as its residual.
- **`apply_gap_plan_action.go:178` still reads `content_guidance`** — deliberately.
  That read takes the value off the gap PLAN and is `content-gap-planner`'s LLM
  output contract. The ratchet's own fixture asserts it does not convict that line.
- **`plan_sections_action.go` untouched** — 271 §5's candidate 1 would have built a
  second reader below a channel that already works, and another session had
  uncommitted work on that file's symbols.

## 7. Watch-outs carried from this session

- **~25 of the 90 guidance-only rows are `failed`/`needs_human_review`.** If an
  operator re-triages one after the roll, it will act on its brief for the first
  time. That is intended, but it is a real behaviour change to rows that exist
  today — flagged to the owner in `README_where_we_are.md`.
- **A `LIKE` hit on an agent config is not a reader.** The hit for
  `content_guidance` is `content-gap-planner`'s own prompt instructing an LLM to
  emit the key. Follow every hit to the step that consumes it.
- **A `jsonb_each` over `workflow->steps` sees neither prompt text nor anything
  inside a loop's `sub_workflow`.** It returns clean, confident, incomplete
  answers — it cost this lane a wrong "only one consumer" claim (WRONG_CALLS).
  Enumerate with `default_config::text LIKE '%…%'` first.
- Two **pre-existing** pattern-check advisories fire on files this lane touched
  (`unrepaired-component-write` on `create_tool_component_action.go` and
  `deploy_tool_action.go`). They predate this change, belong to `bugs_open/136`,
  and per LANDMINES those two are **deliberately not allow-listed** — silencing
  them converts a live debt into a false all-clear.

## 8. The rest of the standing five

`PLAN` was folded into the bug file's §9 (the design decision and its reasons live
there, where the fixing session will look). `RUNBOOK_content_guidance.md` has every
query with its gotcha. `NOTES_content_guidance.md` is the technical log including
four missteps. `README_where_we_are.md` is the owner's plain-prose log. No SUMMARY
yet — by the cadence rule, the milestone is the fix going **live**, and that is the
next session's to write.
