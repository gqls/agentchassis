# RFC 004 — a deploy action that can refuse

**Raised 2026-07-29** by the council gate's `editquality` seat, at **medium**
severity, against an **APPROVED** round-1 verdict
(`SUBMISSION_CORR=82fce741-2e5c-424c-b2a2-a5c0fc592720`). Filed here rather than
answered with more measurements, per the owner ruling of 2026-07-28: *"A veto on
SCOPE is not answered by resubmitting with better measurements. It is a
judgement about how a capability reached production."* The verdict was approval,
not veto — but the objection is the same shape, and the honest response is the
same.

**Status: OPEN. The code is live.** Under the 2026-07-29 ruling §2 that is not a
confession, it is the design: HEAD is shared, `make build-*` builds from
committed HEAD, and any session's roll ships it. Review here is after the fact.

## What shipped

`DeployToolToSiteAction` gained a second refusal. It has always refused a tool
with an empty template:

```go
if !toolHTMLTemplate.Valid || toolHTMLTemplate.String == "" {
    return nil, fmt.Errorf("tool %s has no HTML template", toolFunction)
}
```

It now also refuses a tool whose script addresses element ids the template does
not contain:

```go
if orphans := datahelpers.OrphanElementRefs(toolHTMLTemplate.String); len(orphans) > 0 {
    return nil, fmt.Errorf("tool %s addresses %d element(s) its template does not contain (%s) — "+
        "the script would throw on load and the tool would render no controls", ...)
}
```

## The objection, verbatim

> *"DeployToolToSiteAction previously deployed any tool with a non-empty
> template; this edit gives it a genuinely new capability — refusing to deploy.
> That is a change to what the mechanism GUARANTEES, not an addition to a
> detection/visibility path, and it is a different shape of edit than the rest
> of the plan (which all add or widen observability). The author explicitly
> names this as the edit they most want argued and cites RFC_002 as the relevant
> precedent for scope concerns; the 'measured inert on 32 existing templates'
> argument shows it is currently safe, not that it belongs in a
> diagnosis-driven fix for a detection gap."*

**The seat is right about the distinction and I agree with it.** I raised the
same question in the submission's own `risks` block, so this is not a reviewer
catching something concealed — it is a reviewer confirming that a flagged
question was a real one. The rest of that submission adds observability; this
one edit changes an outcome.

## The argument for it, stated fairly

- **It is measured inert.** All 32 live library and fork tool templates pass the
  predicate, so no current caller's behaviour changes. It is a guard against a
  future bad template.
- **It is the same kind of statement as the refusal above it.** "This template
  has no HTML" and "this template's script addresses markup it does not have"
  are both facts about the artefact, decided the same way, at the same point,
  with the same consequence.
- **The alternative is a page that renders, deploys, returns 200 and does
  nothing** — which is precisely the state ten live pages are in, and the reason
  the detector was written.

## The argument against it, which is the one that needs a human

- **Inert today is not inert forever.** The measurement proves it cannot fire
  now; it says nothing about a template written next month. When it does fire,
  it fires as a hard error on a deploy path, and the operator sees a failed
  deploy rather than a flagged tool.
- **It is a different kind of authority.** Everything else in that change
  observes and reports. This one blocks. RFC_002's ruling was that a mechanism
  gaining the ability to REFUTE where it previously only confirmed is the RFC
  trigger — and "an action that always deployed can now refuse" is that shape.
- **A refusal that cannot be overridden is a policy, not a check.** There is no
  escape hatch: no config key, no force flag. That was deliberate (a default-OFF
  switch rots unexercised — the owner's own reason for not requiring one), but
  it means the only way past it is to fix the template, which is right if the
  check is right and a wedge if it is not.

## The question for a human

**Should a shared deploy action be allowed to refuse on a content-quality
predicate, or should it only ever record and let something downstream decide?**

Three options, costed:

1. **Keep it as a hard refusal.** Cheapest, strongest. Cost: the first false
   positive becomes a blocked deploy rather than a noisy work item, and the
   check has already produced one false positive in its short life (see
   `WRONG_CALLS.md`, 2026-07-29 — a page that computes its ids from data). That
   class is now handled, which is evidence the check improves under contact, not
   evidence it is finished.
2. **Downgrade to a loud warning plus a work item.** Deploy proceeds, the defect
   is recorded, a human decides. Cost: we knowingly ship a tool that cannot
   work, which is the exact behaviour this whole change exists to stop.
3. **Refuse, but with a named override** (`allow_orphan_refs: true` on the step
   config). Cost: the mechanism-rots-unexercised problem the owner has already
   ruled against, and an override is exactly what gets set once and forgotten.

**My recommendation is 1, and I am the wrong person to take it** — I wrote the
check, and I have already been wrong once this week about what it finds.

## Consumers to tell, not merely measure (2026-07-29 ruling §3)

`deploy_tool_to_site` is called by the tool-deployer path for every site that
adopts a library tool. The guarantee that changed, stated for them: **deploying a
tool can now fail for a reason that is about the tool's own markup.** No site is
affected today (0 of 32 templates flag). The failure is a normal action error
with the missing ids named in the message.

## Evidence

- Precision and safety measurements: `platform/orchestration/datahelpers/element_refs.go` header.
- Council report: `diagnosis_artifacts` kind=`council_report`, correlation `82fce741-2e5c-424c-b2a2-a5c0fc592720`.
- Register entries: `docs026_concept_register/register/tool-lifecycle.md` TL-032, TL-033.
- The false positive and how it survived four documents: `docs024_key_docs_latest/WRONG_CALLS.md`, 2026-07-29.

## The second, lower objection — recorded, not actioned

The same seat noted at **low** severity that the eligibility widening
(TL-033) closes the gap **only for single-blob ported tools**, leaving a smaller
residual of the same `cc.component_level='tool'` blindness for **idea.uk (3
pages)** and **leopardessconsulting.co.uk (1)**, which are multi-section tool
pages with no tool-level component. That is accurate and was disclosed in the
submission. It stays open deliberately: picking which of several sections on
those pages IS the tool requires a rule I do not yet have evidence for, and
guessing would key content sections as tools. Whoever needs those four pages
covered should decide the rule from their shape, not from this file.

A third note flagged that `check_tool_health` was left unwidened. That was a
disclosed product decision (it raises work items for cosmetic warnings, so
widening it drops ~71 items into the build queue in one pass) and the reasoning
lives in `tool_eligibility.go`. Revisit when the ported tools have PLANs.
