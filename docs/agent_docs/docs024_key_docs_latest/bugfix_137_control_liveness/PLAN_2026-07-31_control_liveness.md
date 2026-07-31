# PLAN — reconciling the two judges of "is this control alive" (`bugs_open/137`)

**Started** 2026-07-31. **Scope:** one bug, one coherent change, one council round.

## What the bug asked for

`bugs_open/137` was filed 2026-07-28 by the experience-register session **at the
explicit request of the council gate's `reuse_agent` seat**, which objected that
leaving two disagreeing judges of control-liveness unreconciled was itself the
reuse risk. The bug is unusual in that it does not claim a defect: both
mechanisms behave as designed, and the filing session said plainly that *which*
is right was genuinely open.

It offered three options and named its own preference:

1. one shared predicate — rejected in the bug's own text, because attribute
   assertion would inherit a **page-wide** exemption and be weakened everywhere;
2. **scope the exemption to the element, not the page** — *"This looks like the
   best answer and is the one I would start from"*;
3. declare them different questions — cheapest, leaves two answers standing,
   and is exactly what the `reuse_agent` seat objected to.

## What reading the code added to the bug's account

The bug locates the disagreement in `evaluateStaticCriteria`. Reading for the
fix found the disagreement is a **symptom of the exemption's scope**, and that
the scope problem is not confined to that function:

> `strings.Contains(html, "data-runtime-fill")` appears in **eight** places, and
> its correctness depends entirely on **how finely the caller happened to chunk
> its input**.

| caller passes | the line means | correct? |
|---|---|---|
| one section | "is this section a shell?" | yes, by accident of framing |
| an assembled page | "does any section on this page hydrate?" | **no** — one shell exempts every neighbour |
| a served page (+ chrome) | as above, wider still | **no** |

Nothing recorded the second behaviour, and no test pinned it, because at each
individual call site the line reads as obviously correct.

**Prior art confirms the reading rather than contradicting it.**
`save_sections_link_repair.go:67-71` already reached the same conclusion and
applied it by chunking its own input:

> *"Per-section application narrows RepairPageLinks' `data-runtime-fill`
> exemption from whole-document to whole-section. That is deliberate and
> correct: … its statically-linked neighbours no longer ride on its exemption."*

A previous thread got the right answer and fixed it **at one call site by
convention**, leaving every other caller to rediscover it. That is the drift
class this council reviews for, and it is why the fix belongs in the predicate.

## The design, and the boundary that is the substance of it

> **CORRECTED 2026-07-31 (council round 1) — the boundary below was drawn in the
> wrong place, and the correction is the most useful thing in this document.**
> It read "element scope for *is this control alive*, whole-input for *is this
> section a shell*". That is necessary but not sufficient: it silently put
> `RepairPageLinks` on the element-scoped side, and `RepairPageLinks` **acts** on
> what it sees. **The real axis is JUDGE vs WRITER.** A judge that sees more
> markup raises more findings, each escalated to a human — safe. A writer that
> sees more markup rewrites more markup, and this writer's action is to strip the
> `<a>` and keep the text, which is a landmine in its own right ("a dead internal
> link is REPAIRED into orphaned prose"). For a writer the wide exemption
> **under-repairs**, and under-repair is fail-safe: an unrepaired control stays
> visible and stays flagged; a repaired one becomes prose nobody can find.
> Caught by the council's `editquality` seat (gating HIGH) and `render_guardian`,
> not by me. Full account in NOTES, misstep 4.

**One marker, one meaning, two representations** (`datahelpers/runtime_fill.go`):

- `RuntimeFillSpans(html)` → byte ranges of every marked element **including its
  subtree**, for the string-world callers;
- `InRuntimeFillShell(sel)` → the goquery form, for the DOM-world callers;
- both derive from one `RuntimeFillMarker` constant.

**Byte spans rather than a DOM, deliberately.** `RepairPageLinks` guarantees
byte-identical output when it changes nothing. Re-serialising through goquery
would break that guarantee on every page it touches, so the string world never
gets parsed and re-emitted.

**The boundary between the two questions is the design decision:**

| role | question | scope | callers |
|---|---|---|---|
| **judge** | *is this control alive?* | **element** | `check_tool_acceptance` (the sweep), `check_dead_controls`, `check_phantom_internal_links`, `check_backend_entry_orphaned`, the attribute checks |
| **writer** | *is this control alive?* | **whole input, by decision** | `RepairPageLinks` (unlinks), `render_site_components_action` (drops the control) |
| — | *is this section a shell?* | **whole input, unchanged** | `check_empty_sections`, `check_component_standards`, `check_component_template_corrupted`, `sectionHasVisibleContent` |

The middle row is the correction above. Both writers carry their reasoning in
the code, plus **what would have to be decided first** to move them: whether
unlinking is the right repair for a dead control at all — a question
`check_dead_controls` has already answered for itself by routing to
`needs_human_review` with no handler, "because picking a fixer automatically
would guess".

The second group asks a genuine whole-section question and is untouched.
`HasRuntimeFillMarker` **names** that predicate so it is a choice rather than a
coincidence, and a test pins it so a later tidy-up cannot quietly redirect those
callers at the element-scoped one.

## The decision that actually resolves 137: SKIP, not PASS

`static_attribute_checks.go`'s own **rule 2** already confines refutation to
elements actually in the served HTML, because anything JS builds is Tier 4's
claim — and it cites the `shell-dead-controls` sweep as its precedent **without
applying it**. An element inside a hydrating subtree is that same claim one step
earlier: markup the loader is about to rewrite.

So a shell-enclosed element is **skipped**:

- **skip, not pass** — a skip can never satisfy `experienceVerdict`, so nothing
  vouches for markup that was not checked. A pass there would be precisely the
  vacuity that file exists to end (its rule 1).
- **per element, not per page** — every matched element *outside* a shell is
  still refuted exactly as before. This is what keeps the change a
  reconciliation rather than a blanket amnesty on any page containing a shell,
  and it is the objection the bug raised against option 1.
- **the exemption is disclosed** — the detail string reports what was asserted
  and how many elements were set aside, honouring that file's rule that every
  detail carries the element count.

## Degrade direction, chosen not incidental

An **unclosed** marked element spans to end of input. That matches what a
browser does with the same markup, and it means malformed HTML degrades to
**exactly today's whole-input exemption, never to a narrower one**. A fix for
over-exemption must not become a source of under-exemption on markup it cannot
parse — that would turn a parser limitation into a wave of false dead-control
findings.

## Ordering of the candidates — why the predicate comes first

Ranked by what makes the bad state unrepresentable:

- **fix the predicate's scope** (taken) — after it, an exemption cannot reach
  markup outside the shell that declares it, whatever any future caller passes;
- *chunk each caller's input* — what `save_sections_link_repair` did; correct
  and unenforceable, since it is a convention every new caller must rediscover;
- *document that callers must pass a section* — a schema defect in a
  documentation costume.

## One pinned expectation changed, and it is disclosed

`TestEvaluateStaticCriteria_AttributeChecksFlowThrough` asserted **FAILED** for
the shell-enclosed template row, with a comment noting the sweep was suppressed
on the same element. **That test encoded the 137 contradiction as an
expectation**, so the reconciliation could not land without changing it. It now
asserts **SKIPPED**, with the date and the reason written into the test rather
than quietly updated.

## Council

Submitted **before** the commit, correlation `4465f655-c6c6-49b4-a9b8-4ca7a5f647df`,
committed with `Council-Submitted:` per the 2026-07-30 rule. Not an ordering
exemption claim: nothing here needed to ship ahead of review, and per the owner
ruling of 2026-07-29 no thread can hold a change out of the fleet on this tree
anyway — review here is after the fact by design.


## What round 1 added: a gate, not a note

`bug_historian` objected that naming the other consumers was "a documentation fix
for a code-shape problem". Right — and it is this defect one level up: eight
copies accumulated because a ninth cost nothing and told nobody.
`TestNoRawRuntimeFillMarkerTestOutsideThisPackagesPredicate` reads the source
tree and fails the build for any raw marker test outside `runtime_fill.go`. It
does **not** judge which scope a site should use; it forces the choice through a
named predicate so the intended scope is visible in review.

**The gate's own first version was blind to a spelling** — it matched
`Contains`/`HasPrefix`/`HasSuffix`/`Index` and missed
``regexp.MustCompile(`(?i)data-runtime-fill`)``. That is the same defect it
exists to prevent, inside the fix for that defect. Caught by re-deriving the
call-site manifest from a literal grep rather than from the new test. NOTES,
misstep 5.
