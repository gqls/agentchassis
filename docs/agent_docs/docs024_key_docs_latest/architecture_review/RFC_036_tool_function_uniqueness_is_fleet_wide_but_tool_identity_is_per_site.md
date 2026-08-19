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

## 8. OWNER DIRECTION 2026-08-17 (recorded verbatim; resolution still needs one clarification)

> "RFC_036 I'd like other sites to be able to use tools but they can fork the tool for their own use
> which is probably the best route for them."

**What this settles:** the estate keeps the library-and-fork model. A tool is offered once and each
consuming site takes its **own fork** rather than sharing a canonical row. So **option 1 (add
`site_id` to the index) is NOT the direction** — that would make identity per-site at the schema
level, which is the opposite of "one offering, forked per site". Option 5 (leave it) is also out; the
owner has engaged with the question rather than deferring it.

**What it points at:** **option 2** — a native rebuild of a tool that a library entry already names
should be recorded as a **fork of that entry** (`forked_from` set), which is both true to the model
and exempt from `idx_cc_tool_function_unique` by construction. Option 3 (a typed early refusal in
`create_tool_component` instead of a 23505 after the LLM has run) still composes with it and is worth
doing regardless.

**The clarification still owed, and why it is not safe to assume:** the direction says other sites
*can fork*. Deactivating a library template — the interim applied in §6, and the obvious next step for
the two still-blocked tools — is precisely what makes a tool **un-forkable in future**. Those two
(`tool-ab-test-calculator` `8c9a6e06…`, `tool-meme-generator` `6ae53f32…`) each already have a live
fork on another site. Existing forks are separate rows and would keep working; what is lost is any
*future* site's ability to fork them. So the direction reads as an argument for **keeping** those
templates active, which leaves the two tools blocked until option 2 is built.
**Do not deactivate `8c9a6e06` or `6ae53f32` on the strength of §8 alone.**

## 9. THE PATH (owner directed 2026-08-19: framework ownership of all 63 tools, so the 2 blocked tools must be unblocked)

### 9.1 There is no config-only interim, and here is the proof

`deploy_tool_action.go:11-12` defines a library tool in the platform's own words:

> `Library tool: content_components WHERE component_level='tool' AND forked_from IS NULL`
> `Site fork:    new content_components row with forked_from = library tool ID`

**That predicate is the index predicate.** `idx_cc_tool_function_unique` fires on
`component_level='tool' AND forked_from IS NULL AND is_active` — so **"this row is forkable by other
sites" and "this row blocks any rebuild of that tool" are the same condition.** You cannot free the
index without making the template un-forkable, which is exactly what the owner's direction
(2026-08-17: other sites fork) forbids. Every workaround this lane considered is therefore dead:
- deactivate the template → un-forkable (and it is `is_active` that the fork lookup requires);
- set `forked_from` on the template to dodge the index → it stops being a library tool by definition;
- rename the template → does not help; the index keys on `function`, not `name`.

### 9.2 The mechanism is already half-built — `deploy_tool_to_site` does the correct thing

`deploy_tool_action.go:294-312` already creates a **site-owned copy with `forked_from` set to the
library tool's id**, which is exempt from the index by construction. So the estate already has, and
relies on, exactly the shape option 2 proposes. What that path does NOT do is generate fresh HTML —
it copies the library template.

**Naming does not collide between the two paths** (checked, because it would have been a hidden
blocker): the fork builds `name = <library component NAME> + '-' + domainSlug`
(`tool-ab-test-calculator_pre_037-webdesign-co-uk`, matching the live fork `cd60486c`), while the
generator builds `name = <FUNCTION> + '-' + domainSlug` (`tool-ab-test-calculator-webdesign-co-uk`).
Different strings, so `content_components_name_key` does not stand in the way of option 2.

### 9.3 The change, stated so someone can pick it up

In `create_tool_component_action.go`, immediately before the INSERT: look up a library tool claiming
this `function` (`component_level='tool' AND forked_from IS NULL AND is_active`, no site filter). If
one exists, **set the new component's `forked_from` to its id**. Nothing else changes.

- **It is semantically true, not a dodge.** A site-specific native build of a tool the library also
  offers IS a site copy of that tool — which is precisely what `forked_from` means everywhere else.
- **It makes the index correct rather than bypassed.** After the change the index still guarantees
  what it is for: one canonical library entry per tool function. Site copies, however they were
  produced, are exempt — as they already are when `deploy_tool_to_site` makes them.
- **Blast radius is small and enumerable:** it only fires when a library entry already claims the
  function. Fleet-wide that is **4 of this site's 63 tools** today, 76 of 116 tool components hold a
  slot. Every other generation is unaffected because the lookup returns nothing.
- **Council + roll**: it is a shared-seam change on `create_tool_component`, so it needs the gate and
  a chassis roll before the 2 parked tools can build. It does not need an image for anything else in
  this lane.

### 9.4 What it unblocks

`tool-ab-test-calculator` and `tool-meme-generator` — the last 2 of the 63 that cannot be reached by
the proven recipe. Both currently serve their ported versions and are safe; they are parked, not broken.
Beyond this lane it unblocks **any** site rebuilding a tool whose name the library also carries, which
is the general form of the same wall.

### 9.5 If nobody builds it

The lane finishes 61 of 63 and stops. That is the honest fallback and it should be stated in any
"complete" claim: **"61 of 63, with 2 blocked on RFC_036"** — not "done".

---

## 10. CONTRIBUTION 2026-08-19 (portfolio_positioning lane) — the same wall exists at SECTION level, on a different writer, and §9.3 fixes only half of it

**`bugs_open/311`** is this RFC's design fact — *function is fleet-wide, identity is per-site* —
occurring on `store_generated_component` instead of `create_tool_component`. It was filed
2026-08-18 with a `090` verdict of **CONFIRMED on the first iteration**
(`8aa2e283-129f-41d1-93a0-6dcacbbabeae`). Until this note, neither file cited the other:
`grep 311` here and `grep 036` there both returned 0.

**Measured distinction, so nobody merges them by mistake** (2026-08-19):

- 311's blocked rows are `component_level='section'` (`mortgages-repayment`,
  `loans-credit-health-check`, `loans-car-finance-calculator`). `idx_cc_tool_function_unique`
  is partial on `component_level='tool'`, so **the index is not what refuses them** — the
  *regeneration field-contract guard* at `store_generated_component_action.go:397-412` is,
  because the generated schema drops field names the incumbent site's `content_data` is keyed
  on. It presents as a **`failed`** work item, not a `complete` one.
- This RFC's rows (`tool-ab-test-calculator`, `tool-meme-generator`) are `tool`-level and are
  not touched by that guard.

**Why it bears on §9.3.** The change proposed there — look up a library tool claiming this
`function` and set the new row's `forked_from` — is exactly 311's own fix candidate 1, and
§9.2's observation that `deploy_tool_to_site` already builds that shape holds for both writers.
But §9.3 is scoped to `create_tool_component_action.go` alone, and its blast-radius sentence
("it only fires when a **library entry** already claims the function") is a `component_level='tool'`
predicate. **A section-level incumbent claims nothing in the index, so the §9.3 change as
written would leave 311 entirely live.**

**The cost of fixing only this half, measured at the served artefact rather than inferred:**
`loanzy.uk` was built greenfield and lost **7 of 7** tool sections to 311; its page
`https://loanzy.uk/tools/loan-comparison-calculator/index.html` returns 200 with **zero
`<input>` elements** — a calculator page with no calculator, live, with nothing in the artefact
to show a reader that anything failed. The portfolio buildout plans ~140 finance domains whose
propositions share calculator function names by construction, so whichever site creates a name
first owns it and every later site ships that tool hollow.

**Suggested, not asserted — it is this lane's call:** since both are shared-seam changes on the
component write path needing the gate and a roll, one submission covering both writers is
cheaper than two rounds, and it removes the risk of a "tools are fixed" claim that is true for
`tool`-level and false for `section`-level. If the lane prefers to keep §9.3 narrow, **say so in
§9.5's honest-fallback sentence** — "2 blocked on RFC_036" would otherwise read as the whole of
the problem, and it is not.

## §10 — CROSS-LANE 2026-08-19: the SECTION-level half is now BUILT (CLC-020), and this RFC is the tracking home for both mechanisms (council architecture seat's ask)

The 311 fix lane built the section-writer half the same day this RFC's cross-lane note asked
for it: `resolveStorageIdentity` (`platform/orchestration/actions/component_storage_identity.go`,
commit `17d883333`, council **APPROVED r1** `fc3ac5f4`, register **CLC-020**). On a foreign
collision, `store_generated_component` now diverts to a fresh **base** row
`<function>-<domainSlug>` with `section_type` = the requested section name — deliberately NOT
§9.3's `forked_from` shape, and the reason is now MEASURED, not asserted (a council advisory
asked for exactly this demonstration):

- The index this RFC is about, read live 2026-08-19 from `\d content_components`:
  `"idx_cc_tool_function_unique" UNIQUE, btree (function) WHERE component_level = 'tool'::text
  AND forked_from IS NULL AND is_active = true` — **partial on `forked_from IS NULL`**, so at
  TOOL level a fork escapes the gate AND the deploy path links pages itself (fork invisibility
  to selectors is irrelevant there).
- At SECTION level every selection path filters `forked_from IS NULL`
  (`component_selector.go`, `loadSectionComponents` has no such filter but
  `check_unresolved_sections` and the selector do), and the diverted row must be FOUND by the
  requesting page's rebuild — so `forked_from` stays NULL and identity lives in the suffixed
  `function` + request-vocabulary `section_type`.

**The architecture and reuse_agent seats' advisory on the approved round (their words):** the
estate now carries two collision-resolution conventions on one underlying defect (no ownership
column on a globally-named library) and they should be tracked HERE rather than left parallel
ad hoc. So: whoever picks up §9.3, reuse `foreignDependents` (CLC-020's census —
`page_components→pages` UNION `site_components`, requester-excluded) rather than writing a
third census, and record in this section which convention the tool writer adopts and why. The
honest-fallback sentence this RFC asked for now reads: **the section-level half is fixed
(inert until a chassis roll ships it); the tool-level half (`create_tool_component`) remains
open and is this RFC's question.**
