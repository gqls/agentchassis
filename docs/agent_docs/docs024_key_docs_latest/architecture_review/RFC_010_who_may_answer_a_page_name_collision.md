# RFC 010 — Who may answer a `pages(site_id, name)` collision, and with what authority

**Status: RATIFIED 2026-08-02** (owner; the three answers are in §6, and all three are
implemented — commit `f0…` below). Raised the same day by the `bugfix_175_page_role_upsert`
lane, **at the council gate's `review_architecture` seat's explicit request** on corr
`e78c62e3-7f01-48f1-b083-924eaccd195a` (verdict: **approved**, 4 advisory objections, none
high-severity).

> **§2.1 was WITHDRAWN before the ruling** — it asked the owner to ratify a predicate that
> should never have been written, because the estate already had one. See that section; the
> story is in `WRONG_CALLS.md`.

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

### 2.1 ~~Is `deployed_at IS NOT NULL OR build_status IN ('deployed','needs_rebuild')` the right line?~~ — **WITHDRAWN 2026-08-02: it was not, and there was nothing to decide**

> **This question is withdrawn hours after it was raised, and the reason belongs in
> the RFC rather than in a quiet edit.** It asked the owner to ratify a predicate I
> had written. While sizing its blast radius to answer the owner's "what am I
> deciding?", I found `datahelpers.NeverDeployedPagePredicate` — the estate's
> existing, shared, tested definition, with three consumers and a test asserting
> that it **must not single out `needs_rebuild`**, because doing so had produced a
> 34-page false-positive class for the nav lane.
>
> Mine singled out `needs_rebuild`. Measured: **11 live rows are `needs_rebuild`
> with no `deployed_at`, no `last_built_at` and mostly zero components** — never
> built, never served — three of them `lendzy.co.uk` TOOL pages created that day,
> i.e. exactly what the tool arm collides with. My version would have refused all
> eleven. **The code now reads the shared predicate directly** (`NOT
> (datahelpers.NeverDeployedPagePredicate)`), so there is one definition and it
> cannot drift; a test fails if anyone inlines a restatement, and that test was
> proved by mutation (with the import kept referenced, so the assertion fires
> rather than the compiler).
>
> **Nothing is left for the owner here** — "converge on the definition that already
> exists and is tested" is not a judgement call. What remains, and is genuinely
> owner-scope, is the one line below.
>
> The residue worth ruling on: **`bugs_closed/081`'s guard still carries the narrow
> `build_status = 'deployed'` form** (`apply_gap_plan_action.go:590`). It is a
> different arm with a different question, and no live row is currently mistyped in
> a way that exposes it — so this is "fix at next touch", unless the owner wants it
> done now.

**Superseded evidence, kept as the record:**

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

1. **Nothing to ratify — withdrawn, see §2.1.** The code now uses
   `datahelpers.NeverDeployedPagePredicate`, which already was the estate-wide definition.
   The only residue is whether `bugs_closed/081`'s still-narrow guard
   (`apply_gap_plan_action.go:590`) should be converged now or at next touch.
2. **Ratify or amend** the ADOPT branch. As built it is licensed by the constant-role contract
   in review; the amendment is a per-call-site opt-in field.
3. **Rule on the producer-set question generally**: is converging N producers onto one
   `item_type`/`item_key` an architecture-scope act when the type has no automated consumer and
   no rows? The answer decides a class of future work-item changes, not just this one.


---

## 6. OWNER RULING 2026-08-02 — three answers, all implemented

**1. ADOPT becomes opt-in, default OFF.** *(§2.2 — the owner took the tightening.)*
`PageRoleUpsert.AdoptUnshippedRows bool`. The four arms that use it today declare it, so
behaviour is unchanged where it runs; a future caller inherits `false` and gets a REFUSAL
instead of a silent re-type. The reasoning the owner endorsed: the current behaviour is
right for all four callers, but *"the only thing standing between it and 081's rejected
design is a human reading a comment, and this tree has many sessions."* A field is visible
to a reviewer of the CALLER; a doc comment in the helper never is.

Left `false`, the branch **refuses and files nothing**. Refusing is the only option that
cannot corrupt — refreshing-without-retyping is precisely the partial update `bugs_open/175`
is about, so inheriting a default must not reintroduce it. No work item is filed because
`mistyped_deployed_page` is a decision about a LIVE artefact, and this row has never been
served; the refusal reason names the missing declaration instead.

**2. Converging producers onto one `item_type` does NOT need an RFC** — *provided the
producer set is named in the register entry and the shared `item_key` is stated*, which is
what PBP-027 does. *(§2.3.)* The alternative taxes every de-duplication improvement with an
architecture round, and de-duplication is the behaviour this estate wants to make cheap. The
condition is what keeps it honest: a future reader can see who files the type and how it
dedupes without reading four call sites.

**3. Refusal-becomes-error stands.** *(§3.)* Two arms hard-error where they previously
overwrote a live page; the two companion-guide arms log and continue. A visible failure on a
measured four-page surface beats a silent overwrite, and the filed decision is `item_key`-
deduped so it cannot flood the queue.

**Plus the residue from the withdrawn §2.1, which the owner asked for now rather than at
next touch:** `bugs_closed/081`'s guard (`apply_gap_plan_action.go`) no longer reads
`build_status = 'deployed'` — it asks `NOT (datahelpers.NeverDeployedPagePredicate)` like
the new seam. Re-measured before changing it: the mistyped population is **still 5 rows, all
`deployed`**, so **no live row changes treatment today**; the hole was prospective and so is
the fix.

> **And the widening was briefly a decoration, which is worth more than the fix.** Mutating
> the guard back to `existingBuild == "deployed"` left every test in `081`'s file GREEN —
> both predicates agree on the inputs those tests supply (`deployed`+shipped,
> `planned`+unshipped). The discriminating input is a `needs_rebuild` page that HAS shipped,
> where only the new predicate can act; `TestApplyNewPage_NeedsRebuildPageIsRefused` now
> supplies it, and the same mutation is red. **When a mutation passes, the test is not
> confirming the guard — it is failing to see it.**


---

## 7. Implementation went through a REVISE, and the gate was right

Round 2 (`e78c62e3`, 2026-08-02 23:11) came back **REVISE**, gated by `editquality`
[high]:

> *"Rationale claims 'all four current arms declare it' but only deploy_tool_action.go
> (tool + blog-post) is edited — two arms. Since AdoptUnshippedRows defaults to false, any
> arm not explicitly edited to set it true will silently flip from adopt to refusal in
> production."*

**The code had all four; the SUBMISSION showed two.** Two one-line declarations
(`create_tool_component_action.go`, `create_report_page_action.go`) were folded into prose
instead of appearing as edits, so no reviewer could confirm the claim. Verified on
resubmission: `grep -rn 'UpsertPageForRole(ctx'` → exactly 4 call sites; `AdoptUnshippedRows: true`
→ exactly 4 declarations. No fifth caller exists to flip.

**Recording it because the failure mode generalises, and it is the sharp edge of decision 1.**
A default-OFF field converts every *unedited* call site into a behaviour change. That is
what makes the design safe (nothing adopts by accident) and it is exactly what makes an
incomplete edit list dangerous — the reviewer cannot tell an intentional omission from an
oversight. **A change whose safety depends on an exhaustive call-site sweep must show the
sweep, not assert it.**

Three further seats (`bug_historian`, `reuse_agent`, `debug_historian`) asked the same
medium-severity question — *does another `pages` upsert helper hand-roll its own "is this
shipped" test?* The audit is now on record: the other three helpers make **no liveness
judgement at all**. What it did turn up is `bugs_open/181` — ~10 detectors select
`p.build_status = 'deployed'` and are blind to **28 live pages**. Filed with its census
rather than fixed in passing, because converging them changes what ten checks report and
the first consequence of fixing a false-negative is a burst of findings.
