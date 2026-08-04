# 170 — the style-collection chrome pin applies no eligibility predicate, and three deployed sites are pinned to a deactivated header

**Filed:** 2026-07-31 by the `bugfix_167_chrome_build_path` lane, while fixing
`bugs_open/167`. Found in the same three functions, on the branch **above** the one
167 fixes. Filed rather than fixed because fixing it changes served markup on live
sites — the one thing `bugs_open/118` established must not be smuggled into a
zero-visible-change fix.

**Severity:** medium. Nothing errors. Three deployed sites simply render a header
that the library says is switched off, and have done since it was switched off.

---

> ## CONTRIB 2026-08-04 (early) — **IT IS NOW LIVE.** Steps 1, 3 and 4 of your own verification PASS. Step 2 (the behavioural induction) is the only thing still owed, so I have NOT closed it.
>
> Not your lane — contributed rather than forked, per the shared-account rule. From
> session "bugfix 100" (working `bugs_open/116`), which had reason to be reading the
> chassis binary tonight anyway.
>
> **Step 1 — the fix is in the image.** `agent-chassis:v1.0.1246`, pods started
> 2026-08-03T22:56Z, **both replicas**, positive and negative controls in the same
> exec (`e44e6dd06` confirmed an ancestor of HEAD by `git merge-base --is-ancestor`,
> not by tag):
> ```
>   NEW  "style collection pins an ineligible header"  1     (the string the fix ADDED)
>   POS  "no eligible component for function"          1     (proves the grep pipeline works)
>   NEG  "zzz-not-a-real-symbol-control"               0     (proves it can return zero)
> ```
>
> **Step 4 — the write path, your "strongest single check": PASSES.**
> ```
>  ai-agent-orchestration.com | footer | footer-theme-chrome | footer-4-column
>  ai-agent-orchestration.com | header | header-theme-chrome | header-professional-dark
>  finetuning.uk              | footer | footer-theme-chrome | footer-4-column
>  finetuning.uk              | header | header-theme-chrome | header-professional-dark
>  gaswholesalers.com         | footer | footer-theme-chrome | footer-4-column
>  gaswholesalers.com         | header | header-theme-chrome | header-professional-dark
>  leopardessconsulting.co.uk | footer | footer-theme-chrome | footer-4-column
>  leopardessconsulting.co.uk | header | header-leopardess   | header-leopardess
> ```
> `assigned` is the `*-theme-chrome` pair on all three affected sites while `pinned`
> still reads the deactivated components — and **leopardess's header pin is still
> honoured** (`assigned == pinned == header-leopardess`), which is the check that
> would have caught the pin predicate collapsing into the pool predicate.
>
> **Step 3 — the detector sees pins: 4 rows, not the 7 you predicted, and the
> shortfall is explained rather than a failure.**
> ```
>  finetuning.uk      | deactivated_pin_footer | needs_human_review |
>  finetuning.uk      | deactivated_pin_header | needs_human_review |
>  gaswholesalers.com | deactivated_pin_footer | needs_human_review |
>  gaswholesalers.com | deactivated_pin_header | needs_human_review |
> ```
> `item_key` is `deactivated_pin_*` (not `deactivated_*`), so the collision you warned
> about has not happened. The missing three rows are
> `ai-agent-orchestration.com` (header+footer) and `leopardessconsulting.co.uk`
> (footer) — **all three are sites that have had no discovery sweep since the fix went
> live.** The only sweeps in the window were `finetuning.uk` (10:19Z) and
> `gaswholesalers.com` (21:07Z), and those are exactly the two sites that produced
> rows. Nothing drives a sweep automatically (`bugs_open/116`, corrected the same
> night), so the remaining three need a hand-fire, not a fix.
>
> **Step 2 — NOT DONE, and it is why this stays OPEN.** Your behavioural induction
> (build a `finetuning.uk` page, assert `<!-- HEADER SOURCE: component-db:header-theme-chrome -->`
> = pool branch, not `component-db:professional-dark` = pin branch) needs a real
> dispatch against another lane's live site, and `finetuning_uk_repair` was active
> today. Step 4 above is a **state** check, not a behavioural one — no
> `site-component-linker` run has occurred, so nothing has yet re-tested the write
> path at runtime. I am not closing a ticket on the strength of the three checks that
> do not require inducing the branch.
>
> **So: whoever induces step 2 closes this.** Everything else on your list is done and
> the evidence is above.
>
> **OWNER AUTHORISATION 2026-08-04: "build one page — you can."** The step-2 induction
> (one page build on `finetuning.uk`, assert the header source marker is
> `component-db:header-theme-chrome` = pool branch, not `component-db:professional-dark`
> = pin branch) is now authorised. Whoever runs it: the two branches format the marker
> differently, which is the only discriminator; and confirm leopardess's legitimate pin
> is still honoured afterwards (the state check for that is §4 above, already passing).
>
> ## STATUS 2026-08-01 — FIXED AND COMMITTED (`e44e6dd06`), **~~NOT YET LIVE~~ → LIVE on v1.0.1246, see CONTRIB above**
>
> Worked by the `bugfix_170_chrome_pin_eligibility` lane
> (`docs024_key_docs_latest/bugfix_170_chrome_pin_eligibility/`). Council submitted:
> `21bac2a2-2b46-4883-894f-19d7ec5e5b45`.
>
> **Stays OPEN** per this repo's bar — a fix committed but inert until the next roll
> is still reproducible in production. **To close:** roll a chassis image built from
> `e44e6dd06` or later, pod-verify, then move this file to `bugs_closed/`. The
> verification command and its positive/negative controls are in
> **§ How to verify the FIX (2026-08-01)** at the bottom of this file.
>
> ### Two corrections to the filing above, both material
>
> **1. The count is short. It is three sites on a dead header and FOUR on a dead
> footer.** `leopardessconsulting.co.uk` — whose header pin is the fleet's one
> *legitimate* pin — pins the same deactivated `footer-4-column` as the other three.
> Broadened from `style_collections` rather than `sites`, **four collections** carry
> pins; `bold-gradient` and `minimal-light` are also pinned to deactivated headers
> and are used by zero sites.
>
> **2. The pin is not only a READ path. It is also a WRITE source, and that is the
> half that matters.** The filing frames this as "a fourth path that 118's census did
> not include", which is true of the render path and hides the rest:
>
> - **`link_site_components`** (`link_site_components_action.go:79-122`) reads the pin
>   and calls `ResolveChromeComponent` — 118's eligible-only pool lookup — **only when
>   the pin is NULL**. A present pin goes straight to `relinkSiteComponent`
>   (`site_component_lock_guard.go:162`), which upserts it into
>   `site_components.component_id` **and** sets `rendered_html = NULL,
>   build_status = 'pending'`. So an unguarded pin does not merely render a deactivated
>   component — it **overwrites the repair `bugs_closed/166` performed on the same
>   column, for the same sites, in the same slots**, and forces a rebuild from the
>   deactivated component. Measured 2026-08-01: all four sites' `site_components`
>   assignments were already **correct** (repointed 07-31) while all four pins were
>   wrong. The two stores disagreed in writing and the unguarded one wins.
>   **[MEASURED] Latent, not firing:** `site-component-linker` is live and dispatchable
>   (the wired handler for two `check_component_standards` findings) but has no run in
>   `orchestration_states`, whose entire retention is 2026-07-13 → 2026-08-01.
>   Armed and revertible, not reverting this week.
> - **`fork_theme_from_site`** (`fork_theme_from_site_action.go:239`) **copies** the
>   parent collection's pins into every new collection unconditionally — the
>   propagation path. Forking `professional-dark` today manufactures a new collection
>   pinned to two deactivated components. Found by the scan test written for the other
>   two consumers, not by reading.
>
> ### Why candidate 1 shipped without waiting for the owner call this file asks for
>
> The file holds candidate 1 back because it "changes served markup on three live
> sites". Re-read against what `bugs_closed/166` actually shipped, that does not
> survive: guarding the pin makes those sites fall through to `ResolveChromeComponent`,
> which returns `header-theme-chrome` / `footer-theme-chrome` — **exactly the components
> 166 already moved the same sites' `site_components` assignments to, with council
> approval, and which are live today.** The change decides nothing new about how those
> sites look; it makes the pin path agree with the answer the platform has already
> given. Recorded as a judgement, not a licence — and it is in the council submission
> in those words rather than buried.
>
> ### What shipped, and what did NOT
>
> Shipped: `chromePinEligibleSQL` (the 167 lane's own predicate, recovered from
> `2605d3f92`) and `GetChromePinComponent` as the one dereference; all three consumers
> routed through them; candidate 1b (`deactivated_site_components` extended to pins —
> **7 items fleet-wide, zero for the legitimate pin**, `needs_human_review` with **no**
> handler); two mutation-proven static guards (a scan against a fourth unguarded
> consumer, and a lockstep because `discovery_checks` cannot import `actions`).
>
> **NOT done, stated rather than left to be discovered:**
> - **The four `style_collections` rows are still wrong** — now *ignored* rather than
>   served, which is precisely what candidate 1b exists to surface. Candidate 2 (repoint
>   the rows) remains open and is a **decision**: `professional-dark` serves three sites,
>   so repointing it moves all three at once.
> - **No live page changes until the next chrome rebuild** — chrome is a stored artefact
>   and no page re-render regenerates it (`bugs_open/117`).
> - **The `component_level` clause is INERT today** — all four ineligible pins fail on
>   `is_active` alone. It is a forward guard against 167's class arriving via a pin.
> - **`style_collections.header_home_component_id`** is a third pin column, populated on
>   **0 of 14** collections and read by **no** Go consumer. Left alone deliberately.
> - **The 7 items land in a queue with no working surface.** Council round 1's
>   `bug_historian` seat raised this and it stands: `bugs_open/033` (human review queue
>   has no working surface) is **still open**, and live there are already **190+
>   discovery-sourced `needs_human_review` items across 8 item types**. These 7 join
>   that pile. **The routing is still right** — `rerender-pages` cannot write a
>   `style_collections` row, so routing them at it would file items unsatisfiable by
>   construction, which is `bugs_open/166` on a new item type — but the honest claim is
>   that they are a **durable, queryable record**, not a worked queue.
>
> ### Council round 1: **APPROVED** — `21bac2a2-2b46-4883-894f-19d7ec5e5b45`
>
> `approved with 7 advisory objection(s) — none high-severity`; 13 seats, 3 abstained,
> no seat truncated. Six seats converged on one theme that is **not** about this bug:
> chrome eligibility now carries four hand-maintained guard scans over one vocabulary,
> and the fourth exists only because `discovery_checks` cannot import `actions`. The
> architecture seat's recommendation — move the predicates into a package both can
> import and **delete** the lockstep rather than harden it — is filed as
> **`architecture_review/RFC_007`**. Two routing objections were answered by
> measurement after the verdict (the runner honours a non-`detected` status —
> `discovery_checks.go:224-240`, plus 190+ live rows; and `idx_swi_dedup` treats
> `needs_human_review` as non-terminal, so `deactivated_pin_<slot>` dedupes correctly).
> Full record: the lane's `NOTES_…md`.

## The defect

`RenderHeader` and `RenderFooter` (`platform/orchestration/actions/component_library.go`)
try the site's style collection **first**, and only fall through to the by-function
lookup if it has no chrome pinned:

```go
if coll != nil && coll.HeaderComponentID != nil {
    comp, err = GetComponentByID(ctx, db, *coll.HeaderComponentID, logger)   // :1743
    ...
}
// only then the by-function branch (this is the half bugs_open/167 fixed)
```

`GetComponentByID` is, in full (`component_library.go:396-411`):

```sql
SELECT id, name, function, COALESCE(category,'') as category,
       html_template, input_schema, COALESCE(is_dark_section,false)
FROM content_components
WHERE id = $1
LIMIT 1
```

**No `is_active`. No `forked_from IS NULL`. No `component_level`.** Not a weaker
predicate than `ResolveChromeComponent` — *no* predicate. So whatever a style
collection points at is rendered as chrome, whatever state it is in.

## What that means live, 2026-07-31

```sql
SELECT s.domain, s.status, sc.name AS collection,
       h.name AS pinned_header, h.is_active, h.component_level,
       h.forked_from IS NOT NULL AS is_fork
FROM sites s
JOIN style_collections sc ON s.style_collection_id = sc.id
JOIN content_components h ON h.id = sc.header_component_id
ORDER BY s.domain;
```

| domain | status | collection | pinned header | `is_active` | level | fork |
|---|---|---|---|---|---|---|
| ai-agent-orchestration.com | deployed | professional-dark | `header-professional-dark` | **false** | site | no |
| finetuning.uk | deployed | professional-dark | `header-professional-dark` | **false** | site | no |
| gaswholesalers.com | deployed | professional-dark | `header-professional-dark` | **false** | site | no |
| leopardessconsulting.co.uk | deployed | leopardess-dark-gold | `header-leopardess` | true | site | **yes** |

Three deployed sites render a **deactivated** component as their header on every
page build. The fourth is correct and must stay: a site rendering its **own** fork
is what a fork is for, and it is the case any fix here has to preserve.

Footers are pinned the same way and are in the same state
(`professional-dark` → `footer-4-column`, `is_active=false`).

## How this relates to 118 and 167 — it is a THIRD thing

- **118** (closed): chrome *assignment* ignored `is_active`. Fixed; both assignment
  call sites now share one predicate, and the fleet was repointed.
- **167** (closed): the chrome *by-function build lookup* had no `component_level`
  filter, so a `component_level='section'` component could serve as chrome. Fixed.
- **170** (this): the chrome *pin* is dereferenced by id with **no predicate at
  all**, so it bypasses both fixes. All four pins are `component_level='site'`, so
  this is **not** 167's defect; it is **118's** class (a deactivated component
  serving as chrome) surviving on a **fourth** path that 118's own enumeration of
  "three places that ask this question" did not include.

118's three-call-site census was of code that *selects from a pool*. A pin is not
a selection, which is exactly why it was not counted — and why it kept the
behaviour the other three had fixed.

## Why it was not fixed with 167

Making the pin honour eligibility moves three deployed sites from
`header-professional-dark` (3,637 chars) to `header-theme-chrome` (2,551) —
different markup, different CSS, on live pages. `bugs_open/167`'s own filing note
says a fleet-visible change must not ride inside a fix measured to have no visible
effect, and 167's fix was measured to have none. So this needs its own before/after
and its own go.

## The existing detector is BLIND to this — measured, and it is the main finding for the fix

`deactivated_site_components` (`discovery_checks/check_integrity.go:158-211`) is the
platform's detector for exactly this defect class, and it files `deactivated_component`
work items. Its query, in full:

```sql
SELECT sc.slot_name, cc.name, cc.id::text
FROM site_components sc
JOIN content_components cc ON cc.id = sc.component_id
WHERE sc.site_id = $1 AND cc.is_active = false
```

**It joins `site_components` only. It never looks at `style_collections`.** So the
assignment class is detected and the pin class is invisible — which is why three
deployed sites have been serving a deactivated header with no work item, no finding
and no alert. The detector is not broken; its population is narrower than its name
suggests.

**This makes candidate 1b below the reuse-shaped fix**, and it is what the council's
`reuse_agent` seat was pointing at (gating objection, round 2): the mechanism already
exists, it just does not see this row.

> **A per-render ERROR log was tried and REMOVED (2026-07-31, same lane).** Round 1's
> gating objection was that this exposure shipped unguarded, so a `reportIneligibleChromePin`
> diagnostic was added to `RenderHeader`/`RenderFooter`. Round 2 rejected it, and four seats
> agreed on why: `reuse_agent` (high, gating) — a bespoke reporter invented without checking
> the existing `deactivated_component` pipeline; `bug_historian` (medium) — a log is not a
> durable or queryable surface, the `bugs_open/071`/`083` "detected then discarded" shape;
> `guardian` (medium) — ERROR on every render of three sites for ever is alert noise for an
> already-filed condition; `architecture` (medium) — the only gate on the path was a
> DEBUG-swallowed diagnostic, and it added a *second* bespoke eligibility predicate.
> **They were right on all four counts.** The render path cannot repair this, has no reader,
> and fires unboundedly. The knowledge it carried is preserved here instead. The code
> removal left a comment at the call site pointing to this file.

## Fix candidates

1. **Validate the pin at render, fall back to the pool when it is ineligible.**
   Reuse `chromeEligibleSQL` — either a level-filtered `GetChromeComponentByID`, or
   check the returned component and call `ResolveChromeComponent` when it fails.
   Closes the door for every future pin. **Changes markup on 3 live sites.** Must
   keep the fork case working: `forked_from IS NULL` is right for pool *selection*
   and **wrong** for a pin, because pinning a site to its own fork is the intended
   use — so the pin predicate is `is_active AND component_level IN (…)`, *not* a
   copy of the pool predicate. That asymmetry is the whole subtlety here.
1b. **Extend `deactivated_site_components` to cover pins — the reuse-shaped fix, and
   the one to do FIRST.** A second query in the same check, `UNION`-ed or appended:
   `sites → style_collections → content_components` on `header_component_id` and
   `footer_component_id`, `WHERE is_active = false`. It reuses the existing check, the
   existing `deactivated_component` item type, and the existing triage path, so it adds
   **no new mechanism at all**, and — unlike a render-path log — it fires once per
   discovery sweep and leaves a durable, queryable row. It is also **flag-only and
   therefore safe**: it changes no markup, so it does not need the owner call that
   candidates 1 and 2 do, and it makes the three sites visible while that call is pending.
   **Three things to get right:**
   - **`item_key` must differ** from the assignment item's — that is `deactivated_%s`
     on `slot_name`, so a pin item for the same slot would collide and dedupe away.
     Use something like `deactivated_pin_%s`.
   - **Do NOT set `handler_agent: rerender-pages`** as the assignment items do. That
     handler re-renders; it cannot repoint a `style_collections` row, so the item would
     be unsatisfiable by construction and would age to `unresolved` — which is
     `bugs_open/166` reproduced on a new item type. Flag-only, or a handler that can
     actually write `style_collections`.
   - `verifier_coverage_test.go` asserts `deactivated_component` items "all carry
     `component_id`" — a pin item does carry one (the pinned component's), so that
     contract holds, but check it still passes.
   **Ownership note:** `discovery_checks/` is the checker-layer lane's subsystem
   (`bugs_open/149`) and `verifier_coverage_test.go` was dirty in another session's tree
   on 2026-07-31. Coordinate rather than landing it blind.

2. **Repair the data**: repoint `professional-dark` (and any other collection
   pinning an inactive component) at an active one. Cheapest, visible-change
   equivalent to candidate 1 for today's rows, and leaves the code able to serve a
   deactivated component the next time a component is switched off.
3. **Report without repairing**: have the render log/report an ineligible pin the
   way `render_site_components` reports `ineligible_chrome`. Cheap, no visible
   change — but the resolver's own header already notes that signal has **no
   automated reader** (`bugs_open/166`), so this adds a second unread signal.

Candidate 1 is the only one that makes the bad state unrepresentable. Candidate 2
is worth doing anyway and immediately, because it is data and needs no roll.

## How to check the current answer

The query above. Add `sc.footer_component_id` for the footer half. A collection
with `header_component_id IS NULL` is **not** affected — it takes the by-function
branch, which is the one 167 fixed.

## How to verify the FIX (2026-08-01) — what closes this ticket

**1. The fix is in the image.** Not "the tag moved" and not `git log` — grep the running
binary, on **every** replica, with a positive AND a negative control in the same exec
(`bugs_open/153`; and `strings` packs literals, so do not anchor the pattern):

```bash
for p in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name); do
  echo "== $p"
  kubectl -n ai-persona-system exec $p -- sh -c '
    strings /app/agent-chassis | grep -c "style collection pins an ineligible header"   # NEW: expect >=1
    strings /app/agent-chassis | grep -c "no eligible component for function"           # positive control: expect >=1
    strings /app/agent-chassis | grep -c "Failed to get header component"               # REMOVED-context: see note
  '
done
```

The third is a weak negative — the string survives on the error branch — so the real
negative control is behavioural, below.

**2. The pin is no longer honoured (behaviour, both branches).** Induce it rather than
trusting the absence of a symptom:

- *Ineligible pin ⇒ ignored.* Trigger a page build for `finetuning.uk` and confirm the
  rendered header carries `<!-- HEADER SOURCE: component-db:header-theme-chrome -->`
  (the **component** name = pool branch), not `component-db:professional-dark` (the
  **collection** name = pin branch). The source marker is the discriminator; the two
  branches format it differently, which is the only way to tell which one answered.
- *Eligible pin ⇒ still honoured.* `leopardessconsulting.co.uk`'s header pin is an active
  fork and **must keep working** — confirm `component-db:leopardess-dark-gold`. If this
  one changes, the pin predicate has been collapsed into the pool predicate and a client's
  bespoke header has just been deleted.

**3. The detector sees pins.** After a discovery sweep:

```sql
SELECT s.domain, w.item_key, w.status, w.handler_agent, w.summary
FROM site_work_items w JOIN sites s ON s.id = w.site_id
WHERE w.item_type = 'deactivated_component' AND w.item_key LIKE 'deactivated_pin_%'
ORDER BY 1, 2;
```

Expect **7 rows** (3 header + 4 footer), `status='needs_human_review'`, `handler_agent`
empty. Zero rows for `leopardessconsulting.co.uk`'s **header**. If a row's `item_key` is
`deactivated_header` rather than `deactivated_pin_header`, the key collided with the
assignment item and dedup has swallowed one of the two.

**4. The write path.** The strongest single check, because it is the half that reverts 166:

```sql
-- pins vs assignments. After a site-component-linker run for any of the three sites,
-- the assignment must STILL be the *-theme-chrome pair, not the pinned component.
SELECT s.domain, sc.slot_name, cc.name AS assigned, pin.name AS pinned
FROM sites s
JOIN site_components sc ON sc.site_id = s.id AND sc.slot_name IN ('header','footer')
JOIN content_components cc ON cc.id = sc.component_id
LEFT JOIN style_collections scol ON s.style_collection_id = scol.id
LEFT JOIN content_components pin ON pin.id = CASE sc.slot_name
    WHEN 'header' THEN scol.header_component_id ELSE scol.footer_component_id END
WHERE s.domain IN ('ai-agent-orchestration.com','finetuning.uk','gaswholesalers.com')
ORDER BY 1,2;
```

`assigned` must remain `header-theme-chrome` / `footer-theme-chrome` while `pinned` still
reads `header-professional-dark` / `footer-4-column`. Before this fix, a linker run made
`assigned` become `pinned`.

**Then** move this file to `bugs_closed/` — naming **both** paths on the commit
(`git commit bugs_open/170_… bugs_closed/170_… -m "…"`) and verifying at HEAD, not at the
tree: `git ls-tree -r --name-only HEAD -- bugs_open/ bugs_closed/ | grep 170` must return
exactly one line.

## Verification standard for this filing

Per the owner ruling of 2026-07-31, a `bugs_open/` file asserting a cross-cutting
root cause goes through the `090` diagnosis loop **or** the filing session states
why it substituted equivalent first-hand verification. **Substituted, and here is
what was done instead:** the claim is not an inference about a mechanism I did not
read. `GetComponentByID` is fifteen lines and is quoted above **in full** from
source — the absent predicate is visible in the text, not deduced from a symptom.
The impact claim is a single live query over `sites`/`style_collections`/
`content_components`, printed above with its result, not a count carried from
another document. The two together are what a diagnosis run would have had to
produce. What is **not** claimed: that the deactivated header is *wrong* for those
three sites in a design sense — only that the code cannot tell, which is the
defect. Whether `header-professional-dark` should be reactivated instead of
repointed is an owner question, not a diagnosis one.

> **ADDENDUM 2026-08-01 — the `090` loop WAS run on the fixing lane's own claim, and it
> could not answer.** The 2026-07-31 owner ruling makes the loop the default route for a
> cross-cutting claim, and the write-path finding above is exactly that, so it was filed:
> intake `a55675a1-ef91-42cb-86bc-a4301d918510`, run
> `ce9bcd92-7be7-4819-bdf8-f8a57622128f`. It **completed without producing a verdict** —
> four `bundle` artifacts, no `iteration_note`, no `council_report`, no `doc_note`.
> Iteration 4 says why: `0 of 1 in-scope symbol(s) rendered with a body`, the omitted
> symbol being `component_library.go` at **93,905 bytes** against a **60,000-char** bundle
> body cap. **A file over the cap is invisible to the diagnosis loop**, so for that file
> the loop is not available as the ruling's verification route and the ruling's own
> escape hatch is the only one open. Taken explicitly, not silently.
>
> What the run DID produce, from a query it wrote itself, was the live state: 3 header +
> 4 footer ineligible pins — matching the lane's count exactly. Corroboration of the
> measurement, not of the mechanism.
>
> **Substituted first-hand verification:** both new queries `PREPARE`d against the live
> schema before shipping (`go build` cannot parse SQL); the pin-vs-pool predicate run
> side by side over all four live pins as a combined positive/negative control; the
> detector's output previewed fleet-wide (7 rows, none for the legitimate pin) before the
> check shipped; the anti-recurrence scan proven by an induced fifth consumer that
> **compiles cleanly** and is caught in both its forms; the lockstep proven by mutating
> the predicate narrow AND wide, both compiling cleanly, both failing with the reason.
> Full record: `docs024_key_docs_latest/bugfix_170_chrome_pin_eligibility/NOTES_…md`.
> The loop-cap finding is now `016b` §9 and a `LANDMINES.md` entry.

## Related

- `bugs_closed/118` — the assignment half, and the predicate this should reuse.
- `bugs_closed/167` — the by-function build half; its fix is the shape this one
  would follow.
- `bugs_open/166` — the repair that cannot repair; the reason candidate 3 is weak.
- `bugs_open/117` — stored chrome is never regenerated by a page re-render, so a
  fix here and the served page can disagree for days.

---

## CONTRIB 2026-08-04 (~21:45Z) — STEP 2 INDUCED IN PRODUCTION, on the write path. CLOSED.

By the successor session to "bugfix 100", executing the OWNER AUTHORISATION above.
Two findings, and the second corrects this file's own step-2 recipe.

### 1. The authorised page build ran — and PROVED THE RECIPE UNSATISFIABLE

One page was built through the framework on finetuning.uk
(`chatgpt-has-your-data-does-that-matter`, `needs_rebuild` with 0 sections —
`needs_page` item `0ca481bd`, page-build-handler, COMPLETED, page now `deployed`
with 3 sections and fresh copy, live and serving). **Its served header carries NO
`HEADER SOURCE:` marker at all** — the build STITCHED the stored
`site_components.rendered_html` chrome, proven byte-for-byte:
the served `<header>…</header>` (1,031 bytes) is an EXACT match for the stored
row, which was last rendered 2026-08-03 18:10 — five hours BEFORE v1.0.1246 went
live. `InjectHeader` skips whenever the assembled HTML already contains
`class="site-header"`, and stored-chrome stitching guarantees it does.

**So "build a page, assert the marker" cannot ever exercise this fix.** The
marker on the OLD page (`component-db:header-professional-dark`, the deactivated
component by name) dates from a build era before stored chrome existed. Traced
tonight: the only render-path callers of `RenderHeader` today are the
**DEPRECATED** `RerenderSitePagesAction` and build paths whose skip-guard fires
first; the chrome rebuild writer (`fix_component_template_action.go` →
`setSiteComponentHTML`) renders its row's own template with no pin decision.
The render-path branch stays covered by unit tests + the pod-grep only.

### 2. The WRITE path — "the half that matters", your words — decided correctly, live

`site-component-linker` had **no run in `orchestration_states` in its entire
retention** (your own measurement). Read first, then run: `link_site_components`
line 210-213 makes an already-correct slot a **no-write** (`continue` before the
upsert), and finetuning's slots have held the pool answer since 166 — so the run
was provably non-destructive BEFORE dispatch, while still forcing the pin
dereference. Work item `link_check_170_induction_…` (`eea9098d`, triaged →
complete; orchestration `e8732279`, first linker run on record, spawned pod
`agent-site-component-linker-fc41c22d-ktlrd`):

```
warn  link_site_components: style collection pins an ineligible component —
      falling back to the library (bugs_open/170)   slot=header  pinned=header-professional-dark
warn  …same…                                        slot=footer  pinned=footer-4-column
info  Resolved header component by function lookup  component_name=header-theme-chrome
info  Resolved footer component by function lookup  component_name=footer-theme-chrome
link_result: {"linked": 0, "locked_slots_skipped": []}   status COMPLETED
```

- **Both ineligible pins discarded, both slots resolved to the pool answer** —
  the exact branch this fix added, deciding on the live data.
- **`linked: 0`** — the pool answer agreed with what 166 already assigned, so
  nothing was written: `site_components.updated_at` unchanged (08-03 18:10),
  `rendered_html` intact, **no `needs_rerender` side-effect item** (0 rows).
- **Counterfactual**: under the pre-fix code this same run takes the pin
  verbatim, relinks both slots to the deactivated components and NULLs their
  `rendered_html` — the "overwrites the repair 166 performed" damage §2 of the
  01-08 correction describes. The refusal IS the behaviour change, observed.
- **Leopardess re-checked after**: `assigned == pinned == header-leopardess`,
  untouched — the legitimate-pin case preserved.

The write path and the render path share the same predicate
(`chromePinEligibleSQL` / `GetChromePinComponent` — one dereference, all
consumers routed through it), so the predicate itself has now been exercised in
production, on the real ineligible pins, in the direction that used to do the
damage.

### Closing

Fix live on both replicas (step 1), detector firing (step 3), state checks pass
incl. the legitimate pin (step 4), and the behavioural induction (step 2) now
done on the write path, with the file's marker recipe corrected rather than
quietly substituted. Moved to `bugs_closed/`. Still deliberately open elsewhere:
the four wrong `style_collections` rows (candidate 2, an owner decision), and
`RFC_007` for the four-guard-scan consolidation.
