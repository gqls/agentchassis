RATIONALE (draft for the council submission — bugs_open/260 renderer half)

The defect. RenderTemplateReportingMissing (component_library.go:965) executes a component
template with Go text/template via executeGoTemplate (call_agent.go:1170). On ANY error it logs
at Warn and silently falls back to a regex renderer written for HANDLEBARS syntax
(component_library.go:1032-1057). That renderer substitutes {{.field}} but cannot execute
{{if}}, {{range}} or {{end}}, so it emits well-formed HTML with every control directive left in
and every field value already resolved. Values resolved plus directives surviving is the
diagnostic fingerprint: generation worked, and no template engine did the rendering.

Why now. 26 events, 7 domains, 25 work items, 2026-08-11 to 2026-08-18, and accelerating —
08-18 alone was 9 events across three domains. 24 of the 25 items are parked at
needs_human_review. Three lanes hit it independently in one week on unrelated sites, including a
greenfield build with no prior components at all.

Why the cost was under-read. The bug file's headline said "no live damage", from a sound
measurement of STORED content: nothing broken has ever been persisted, because validate_content
refuses the page first. Both halves of that re-verify today (0 of 1,789 page_components and 0 of
72 site_components leak a control directive, positive-controlled regex). But the defect's entire
effect is that NOTHING IS STORED, so a census of stored rows reports it as harmless by
construction — survivorship, not safety. What actually happens is that the page never exists
while the pages that did build keep linking to it: loanzy.uk serves a 404 at /your-rights.html,
remortgagecalculator.uk carries two dead nav links on every page it serves.

And the protection is narrower than the file states. "The gate is why" is true of the page-BUILD
path only. applyContentEdit (section_editor_actions.go:886) and applyComponentSwap (:996) render
through the same seam, and updatePageComponentAfterEdit (:1233) writes rendered_html straight to
an already-live page — grepping that file for validate/unrendered returns one unrelated comment.
Chrome (RenderHeader/RenderFooter/RenderHead) is the same shape. Both editor sites do carry an
`if rendered == ""` guard, and it cannot fire for this defect, because the fallback returns
mangled HTML rather than empty. That path is exercised: 271 content_rewrite/content_edit items,
117 complete, since 2026-04-08. So the correct risk statement is not "no live damage is
possible" but "the ungated path has not yet been unlucky."

Why deleting the fallback is safe, measured rather than argued. Three independent zeros, each
with a control that could have come out otherwise:
  (1) 0 of 251 active component templates fail to Parse. Parse is data-independent so no
      RenderContext replica is involved; the seven FuncMap names were extracted mechanically
      from executeGoTemplate rather than typed, because an undefined function is itself a parse
      error. Controls: an unclosed {{if}} must fail (it did), a valid nested template must pass
      (it did). So the fallback is never entered via a parse error — every occurrence is an
      EXECUTE error, i.e. a data type violation.
  (2) 0 of 1,778 stored sections fail to Execute against their own stored content_data.
      Faithful without a replica because contextToInterfaceMap merges ContentData at the top
      level (component_library.go:1266-1268) and missingkey=zero makes absent site fields safe.
      Conservative, not inflated: it cannot manufacture a failure from a missing site field.
      Controls: the bug's own A/B pair — a string where an array is ranged must fail, the
      coerced array-of-objects must render. Both fired.
  (3) 0 of 253 active components use the fallback's own dialect: no {{# handlebars blocks, no
      {{nav_items_html}}, no {{quick_links_html}}. Re-verified today on the grown population
      (the file's figure was taken at 255).

Together: deleting the fallback changes the behaviour of nothing that currently works. It is a
path nothing on the estate can be rendered by, and its only observable effect is to convert a
clear error into a broken page.

What the change does NOT do. It does not make any page build that currently fails succeed. Those
pages already fail — at validate_content, with 20 blockers that are a regex cap rather than a
measurement. What changes is that they fail at the component, immediately, with the real error
("range can't iterate over ..."), naming the field. On the editor and chrome paths, which have no
gate, it closes a route by which mangled markup reaches a live page with nothing to stop it.

Constraint the design must respect. The owner has ruled all sites should be capable of having
tools, and tool pages legitimately carry {{ }} literals in copy. One of the 26 recorded
occurrences is exactly that shape ({{ variable }}, spaces inside the braces — content about
templates, not a failed range). So the fix must distinguish "the renderer failed to execute"
from "this content contains braces", and its acceptance test needs a good tool page as a
positive control that must still pass. Note this argues FOR the seam fix and AGAINST tightening
checkUnrenderedTemplates: with the seam failing loud, no leaked HTML reaches that detector at
all.

Scope and registration. Shared rendering plumbing, so: council round before/alongside the
commit, and a concept-register entry in the same commit that ships it (owner ruling 2026-07-28
§2, condition 2). No ordering constraint is claimed (owner ruling 2026-07-29 §2 — on a shared
HEAD there is none to claim). Other consumers have been told rather than merely measured (owner
ruling 2026-07-29 §3): the copy_quality_two_stage lane owns the stage-2 executor that renders
through this seam and has a written notice naming what changes about its guarantee.
