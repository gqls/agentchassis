# RFC 010 — Who may answer a `pages(site_id, name)` collision, and with what authority

**Status: OPEN — raised 2026-08-02** by the `bugfix_175_page_role_upsert` lane, **at the
council gate's `review_architecture` seat's explicit request** on corr
`e78c62e3-7f01-48f1-b083-924eaccd195a` (verdict: **approved**, 4 advisory objections, none
high-severity).

> **THIS RFC IS RETROSPECTIVE, AND THAT IS THE POINT.** The change is committed
> (`cbbecb021`) and **live** — chassis `v1.0.1233`, both replicas pod-verified. This is the
> `bugs_closed/124` shape the owner ruled on: *the code stays and the precedent gets fixed*.
> The useful product is a rule that makes the next class-fix of this shape cheaper to judge,
> not a revert. The architecture seat said as much in its own note: *"objecting on record, not
> to block … the blast-radius/rollback framing belongs in the architecture_review track's own
> document, not only in a bug's risks block, so the next similar class-fix has a citable
> precedent instead of reinventing the judgment call each time."*

---

## 1. Problem + evidence

### 1.1 What was added

`UpsertPageForRole` (`platform/orchestration/actions/page_role_upsert.go`) — a four-outcome
collision resolver that four call sites now share, replacing four independently-written
`INSERT … ON CONFLICT (site_id, name) DO UPDATE SET <some columns>` statements:

| collision | outcome |
|---|---|
| none | **created** — every declared column |
| row holds the SAME role | **refreshed** — the caller's declared `Refresh` subset only |
| DIFFERENT role, never served | **adopted** — the arm takes the row over completely, `page_type` included |
| DIFFERENT role, has been served | **refused** — nothing mutated, `mistyped_deployed_page` filed `needs_human_review` |

Why it was needed is not in dispute. `bugs_open/175` censused four arms carrying the shape
`bugs_closed/081` had just fixed on a fifth; the statement silently turns a CREATE into a
PARTIAL update (this arm's content under the existing row's role, no error, an id returned
either way). 081 measured that loop running three months on one site. `pattern-check`'s
baseline is the recurrence rate: **4 hits across 1,120 Go files at HEAD, written by different
sessions in four files.**

### 1.2 What the architecture seat ruled

> *"UpsertPageForRole is a 4-outcome collision-resolution state machine that becomes the
> shared seam for at least 4 call sites today and is explicitly written to be the seam future
> arms 'reach for.' That is a shared mechanism, not a patch to one action — it meets the
> architecture trigger even though it is well-tested and scoped."*

And, separately:

> *"`mistyped_deployed_page` goes from one producer/consumer to three producers sharing one
> `item_key` shape and one filing function. That is a contract expansion on a work-item type
> other code will come to assume has a fixed producer set."*

The `guardian` seat **declined to veto** in the same round — *"Council's own `bug_historian`
requested this exact measurement-then-fix path, and the author did the measurement rather than
asserting safety"* — while asking for the same thing this RFC exists to give: an explicit
sign-off on the ADOPT branch's authority, rather than a footnote in a risks block.

## 2. The question this RFC actually asks

Not *"was the seam a good idea"* — three seats and the measurement say yes. The question is
narrower and reusable:

> **When an arm writes a page whose role it owns by construction, how much authority may it
> take over a row that already holds that name?**

The change answers it with a boundary that has been **served / never served**, and the RFC asks
the owner to ratify or move that boundary. Three sub-questions, each with the evidence:

### 2.1 Is `deployed_at IS NOT NULL OR build_status IN ('deployed','needs_rebuild')` the right line?

**Evidence, measured 2026-08-02:**

```
 build_status  | count | ever_deployed
---------------+-------+---------------
 deployed      |   491 |           490
 needs_rebuild |    46 |            35
 planned       |    42 |             0
```

`bugs_closed/081` guarded `build_status = 'deployed'` alone. `bugs_closed/037` is an entire
filed case about `needs_rebuild` falling outside exactly that predicate. **35 rows are live and
invisible to the narrow guard.** This RFC proposes the wide form as the estate-wide definition
of "this page has been served", now recorded in `LANDMINES.md`.

### 2.2 May a constant-role arm ADOPT (re-type + rewrite) a row that has never been served?

This is the widest new authority in the change and the one three seats singled out.

**The case for:** the row has never been served, so nothing a visitor can see changes; the arm
was *already* writing its content into that row (that is the bug); and leaving `page_type`
wrong is the defect itself, not a conservative choice. A partial adopt — type but not url —
mints a fresh hybrid, which is the same defect wearing different columns.

**The case against, stated at its strongest:** it is authority licensed by a *comment*. Nothing
in the type system stops a future caller passing an LLM-chosen `page_type` and thereby handing
a model-steered arm the power `081` deliberately declined (its fix candidate 1). The
`editquality`, `guardian`, `constitution`, `mission` and `architecture` seats each named this
independently.

**Option if the owner wants it tightened:** an explicit opt-in field
(`AdoptUnshippedRows bool`, default false) at each call site — one field, four lines, and the
authority becomes visible in the caller rather than in a doc comment.

### 2.3 May a bug fix expand a work-item type's producer set from one to three?

`mistyped_deployed_page` had **one** producer and, verified against the live tables rather than
the test docstring the submission originally cited (`prior_art_librarian`'s objection, and it
was right to insist): **0 active `agent_definitions` reference it in any config column, and 0
rows exist.** It now has three producers, one `item_key` shape, one filing function — so a
gap-plan refusal and a tool-deploy refusal on the same page dedupe onto a single open human
decision instead of two.

The `architecture` seat's point is that "0 rows today" is not the same as "no contract", citing
`bugs_closed/129` as the precedent where a contract+schema pairing shipped inside a two-day bug
fix. **The convergence is the RFC-grade decision here, not the row count.**

## 3. Blast radius, stated properly

- **Code:** one package (`platform/orchestration/actions`). Four call sites converted, one
  (`refuseDeployedPageTypeConflict`) converged onto the shared filing function.
- **Pipelines:** four — tool-deploy, tool-generation, report-build, content-gap-plan. The
  `guardian` seat asked that each pipeline's owner confirm the behaviour change; **two arms now
  return a hard error where they previously overwrote a live page** (`DeployToolToSiteAction`'s
  tool page, `CreateReportPageAction`), and the two companion-guide arms stay non-fatal and log.
- **Not touched:** `applyNewPage`'s own resolver (LLM-chosen role — deliberately excluded from
  the ADOPT branch), and the five arms that carry `page_type = EXCLUDED.page_type` deliberately
  (adoption, blog posts, site sync, verbatim adoption, the port CLI). `bugs_open/175` says
  explicitly that making the two camps identical is the wrong resolution.
- **Rollback:** the seam is additive at the call sites — reverting means restoring four
  `DO UPDATE` statements, which restores the defect. There is no partial rollback that keeps the
  refusal and drops the adoption; if the owner wants that, it is the opt-in field in §2.2, not a
  revert.

## 4. What is already done, so the RFC asks only for the judgement

- Registered as **PBP-027** in the concept register, in the commit that shipped it (the
  2026-07-29 ordering-exemption ruling's condition 2).
- Other consumers **named and told**: `discovery_checks/verifier_coverage_test.go`'s
  description of the item type was updated in the same commit; the live check in §2.3 replaces
  the docstring premise.
- Recurrence closed mechanically: `check_partial_page_upsert` in `scripts/pattern-check.py`,
  measured 4 → 0.
- Guards proved by mutation (five induced breaks, five red tests) rather than asserted.
- **Live and pod-verified:** `v1.0.1233`, both replicas — added strings present, positive
  control present, and the removed statement's spelling absent (0).

## 5. Recommendation

Ratify the boundary as built, with one amendment if the owner shares the seats' unease:

1. **Ratify** "has been served" = `deployed_at IS NOT NULL OR build_status IN
   ('deployed','needs_rebuild')` as the estate-wide predicate for *may I mutate this page*, and
   fix `bugs_closed/081`'s narrower guard to match at the next touch of that file.
2. **Ratify or amend** the ADOPT branch. As built it is licensed by the constant-role contract
   in review; the amendment is a per-call-site opt-in field.
3. **Rule on the producer-set question generally**: is converging N producers onto one
   `item_type`/`item_key` an architecture-scope act when the type has no automated consumer and
   no rows? The answer decides a class of future work-item changes, not just this one.
