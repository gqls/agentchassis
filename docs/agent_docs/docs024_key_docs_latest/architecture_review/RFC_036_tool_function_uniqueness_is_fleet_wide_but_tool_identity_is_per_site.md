# RFC 036 — `idx_cc_tool_function_unique` is FLEET-WIDE but a tool's identity is PER-SITE: two gates on one INSERT that see near-complementary sets

## STATUS: OPEN — filed 2026-08-17 by the `webdesign_tool_rebuilds` lane, after it cost a build. No code change proposed by the filing lane; the owner has taken the contained interim (see §6) and this RFC exists for the durable question.

## 1. What happened, plainly

The lane is replacing webdesign.co.uk's 63 imported ("ported") tools with framework-built ones at
the same URLs, one at a time. Rebuild #2 (`tool-ab-test-calculator`) was filed with the documented
precondition checked, ran, and **died inside `create_tool_component` at `save_tool`**:

```
duplicate key value violates unique constraint "idx_cc_tool_function_unique" (SQLSTATE 23505)
```

The work item nevertheless reported `complete` with `error` NULL; the real message was in
`orchestration_states.collected_data->'__step_error'`. Nothing was built.

## 2. The structural fact

`create_tool_component` has **two gates on the same INSERT, and they see nearly complementary sets.**

```sql
-- gate A: the action's own "already exists?" probe (create_tool_component_action.go ~197-217)
SELECT cc.id FROM content_components cc
JOIN page_components pc ON pc.component_id = cc.id
JOIN pages p ON pc.page_id = p.id
WHERE cc.function = $1 AND cc.component_level = 'tool'
  AND p.site_id = $2 AND cc.is_active = true LIMIT 1;

-- gate B: the constraint the INSERT actually hits
CREATE UNIQUE INDEX idx_cc_tool_function_unique ON content_components (function)
 WHERE component_level = 'tool' AND forked_from IS NULL AND is_active = true;
```

| | gate A (probe) | gate B (index) |
|---|---|---|
| scope | **this site** | **fleet-wide — no `site_id` at all** |
| sees a fork (`forked_from` set)? | **yes**, in any `build_status` | **no** — forks are exempt |
| sees an unplaced library template? | **no** — the inner join drops it | **yes** |
| failure mode | silent `already_exists`, run "succeeds", nothing written | hard 23505 after the LLM has already run |

**Satisfying either tells you nothing about the other.** In the worked case, deactivating the
withdrawn per-site fork (the documented remedy for gate A) moved the failure from a silent no-op to
a hard constraint violation — a better failure, still a failure, and the row actually holding the
slot was the *library template with no placement anywhere*, which gate A cannot see by construction.

## 3. The question for the owner

**Is a tool `function` a fleet-wide identifier or a per-site one?** The estate currently asserts both:

- **Fleet-wide**, via gate B and the library/fork model — one canonical template per function, forked
  per site, `forked_from` marking the copies. That is coherent.
- **Per-site**, via gate A, via `TL-033`'s finding that *"a ported tool's identity is its PAGE, not its
  component"*, and via the generator's own naming (`<function>-<domain-slug>`). Also coherent.

The two only collide when a **novel build** (which sets `forked_from = NULL`) targets a function some
library template already claims. That is exactly what "rebuild an imported tool natively" is.

## 4. Blast radius `[MEASURED 2026-08-17, live DB]`

- **116** `component_level='tool'` components fleet-wide; **76** occupy a unique slot
  (`forked_from IS NULL AND is_active`); **26** are forks.
- Of this lane's **62** remaining ported tools, **4** are blocked by gate B:

  | tool | blocking template | forks from it | live forks |
  |---|---|---|---|
  | `tool-ab-test-calculator` | `8c9a6e06…_pre_037` | 2 | **1 (idea.uk)** |
  | `tool-meme-generator` | `6ae53f32… tool-meme-generator` | 1 | **1** |
  | `tool-bg-remover` | `bdd2990a… tool-bg-remover` | **0** | 0 |
  | `tool-prompt-architect` | `2c941ec2…_pre_037` | **0** | 0 |

  So it is 4 of 62 here — small — but the same collision is waiting for **any** lane that rebuilds a
  tool whose name matches a library entry, on any site, and it will present as a `complete` work item.

## 5. Options, costed

1. **Add `site_id` to the index** (`UNIQUE (function, site_id) WHERE …`). Makes identity per-site,
   matching gate A and TL-033. Requires `content_components` to carry a site — **it does not today**,
   which is precisely why the library/fork model exists. Largest change; probably the honest one.
2. **Make a rebuild set `forked_from`.** Cheap and local: a rebuild of an existing tool is arguably a
   fork of the library entry. Changes generator semantics estate-wide, and lies slightly — the
   rebuilt HTML is not derived from the template.
3. **Reconcile the two gates without changing either predicate** — have `create_tool_component`
   pre-check gate B's exact predicate and return a *typed, loud* failure. Does not unblock anything;
   converts a 23505-after-LLM-spend into an early refusal. Strictly an improvement, not a fix.
4. **Deactivate the blocking library templates** case by case. What the owner chose as the interim,
   scoped to templates that are both unplaced and unforked (§6). Does not scale and does not close
   the class.
5. **Leave it.** Cost: 4 tools stay imported, and the trap stays live for the next lane.

Options 3 and (1 or 2) compose: 3 is worth doing whichever identity answer wins, because the current
failure is silent at the work-item layer.

## 6. What has already been done (interim, owner-directed 2026-08-17)

The owner took the contained option: **deactivate only the blocking templates that are both unplaced
and unforked** — `bdd2990a` (`tool-bg-remover`) and `2c941ec2` (`tool-prompt-architect`), verified
0 placements and 0 forks each, in one transaction with pre-asserts on both properties and a
post-assert that the two load-bearing templates (`8c9a6e06`, `6ae53f32`) remain active. `UPDATE 2`.
`tool-ab-test-calculator` and `tool-meme-generator` remain blocked and stay on their ported versions
pending this RFC.

## 7. Related

- `create_tool_component_action.go` (both gates live here) · `bugs_open/286` + register **TL-044**
  (the `adopt_existing_page` path this sits immediately in front of)
- **TL-033** — "a ported tool's identity is its PAGE, not its component" (the per-site reading)
- LANDMINES: *"The tool generator's 'already exists' probe ignores `build_status`…"* (gate A's own trap,
  filed by this lane 2026-08-16 — this RFC is the other half of the same INSERT)
- `docs024_key_docs_latest/webdesign_tool_rebuilds/` NOTES 2026-08-17 12:12Z (the failed build, in full)
  and `WRONG_CALLS.md` 2026-08-17 (the filing lane ran the precondition and reasoned past it with
  gate A's logic — evidence that a human reader does not reliably keep the two gates apart)
