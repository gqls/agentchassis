# PLAN — bugs_open/097's headline half: resolving the links stored in `content_data`

**Started** 2026-08-02 · **Bug** `bugs_open/097_HANDOFF_2026-07-26_cta_integrity_misses_card_links_to_unbuilt_pages.md`
**Council** submission `40c0c14d-636c-4d6f-b3a2-9316267d7367` · **Commit** `d78f70bf1`

---

## Why this workstream exists, and what it is NOT

097 was filed on 2026-07-26 with six broken homepage links on oufe.com. Between
then and now, **its repair half was fully built by other rounds** — `RepairPageLinks`
now runs at four seams (the build gate, the section persist, both rerender
outbound paths) and the deployed artefact is protected. What was never built is
the half the file calls its headline:

> *"Detection and repair are not substitutes: the repair now deletes a phantom
> link silently, and the authoring defect that wrote it goes unreported."*

So this workstream is **not** another repair pass. It is the missing resolution of
the **third representation** of a page's links.

## The three representations (the framing the fix rests on)

| copy | who writes it | who resolved its links, before this change |
|---|---|---|
| the deployed HTML string | the deploy path | `repairOutboundPageLinks` (both rerender paths) |
| `page_components.rendered_html` | `save_page_sections` + 3 single-component writers | `repairSectionLinks` (079) + `repairComponentHTMLBeforePersist` (136) |
| **`page_components.content_data`** | the writer/LLM, per component | **nothing** |

`content_data` is not a cache of the other two — it is what
`rerender_page_sections` **rebuilds each section FROM**, with no writer pass. A
dead href stored there is regenerated on every re-render, repaired again on the
way out, and never once reported.

`component_link_repair.go`, committed by the 136 lane on the morning of
2026-08-02, names this limit exactly and routes it here:

> *"content_data. Same limit as 079's fix … The deployed artefact stays covered by
> the outbound rerender seam (repairOutboundPageLinks, bugs_open/097)."*

## The decision that shaped everything: nominate by NAME, judge by VALUE

Every mechanism the platform had answered "is this field a link, and does it point
anywhere real" by **enumerating field names**:

- `ctaFieldNames` — 6 components × 2 named top-level fields.
- `DeriveCTAURLFields` (`ctafields.go`, the staged successor on council trail
  `2525f980`) — top-level `<stem>_url` **that has a label sibling**.

097's own fix-candidate ranking says why that shape keeps reopening:

> *"Candidate 1 makes the bad state detectable wherever it appears; 2 and 3 both
> require someone to remember to extend a list, which is the shape that produced
> this bug."*

**The rule chosen:** a value is a *candidate* when its field NAME says it holds a
url — at any depth, in any container — and it is *judged* only when
`ClassifyLinkScope` says the VALUE is an internal page link.

That division is the whole design. It is why the fix needs **no exclusion list**:
an `image_url` holds `/images/x.jpg` (`LinkScopeAsset`) and a `docs_url` holds
`https://…` (`LinkScopeExternal`), so neither is ever considered without this code
naming either of them. `ctafields.go` reaches the same end with a `site_assets.`
source guard read from the *schema*; moving the judgement to the *data* is what
makes nesting unable to hide it.

## The asymmetry between the two arms, and why it is deliberate

| arm | condition | action | live count |
|---|---|---|---|
| **rewrite** | the target page exists; the writer omitted the extension | write the stored `pages.url` back into `content_data` | 19 |
| **phantom** | no page row at all | **report only — the value is left alone** | 33 |

The `content_data` analogue of `link_repair.go`'s *unlink* arm is to blank the
field. That arm is recorded as unsettled in its own header:

> *"WHAT IS STILL OWED (deferred, not forgotten): decide whether unlink is the
> right repair action at all … Raised by the council's editquality and
> render_guardian seats, round 1 of correlation 4465f655…"*

Widening a **disputed** repair from the rendered copy to the **source of truth**,
where it is unrecoverable, is not a scope fix's to do. Under-repair is the
fail-safe direction here for the same reason it is there: the deployed artefact is
already protected, so what is owed on this path is a **report**, not a second
deletion. A test pins the non-mutation, so reversing this is a deliberate edit
rather than a drift.

## Decisions, with their reasons

1. **At the persistence chokepoint, not at the CTA resolver.** 097's candidate 1
   says "at the same point the CTA check runs". That is impossible for these
   links: `resolve_internal_links` runs *before* the writer, and
   `info-card-grid.cards` is `"source": "llm"` — the values do not exist yet.
   Persistence is where every body-section writer converges
   (`save_sections_link_repair.go`'s own argument for being there).
2. **One page index, shared with the markup pass.** Two loads are two chances to
   disagree about what a real page is — the divergence
   `validate_page_content.go:272-277` already refuses between its own check and
   its own repair.
3. **content_data first, markup second.** So a rewritten source field and the
   anchor rendered from it are corrected in the *same* save, not one save apart.
4. **Runtime-fill sections exempted in lockstep with the markup pass.** Their
   stored values are placeholders the client replaces; and skipping in lockstep is
   what stops the two representations of one section diverging.
5. **A work RECORD, not a work ITEM.** `writeLinkRepairLog`'s precedent verbatim:
   `bugs_open/083` (nothing drains `detected`) and `bugs_open/077` (no items whose
   handler has no remit). New code `CONTENT_DATA_LINK_AUDIT`; the existing two
   untouched, so every query already written against them is unaffected.
6. **No migration.** The 52 live findings clear as a side effect of ordinary
   operation — each page's next save. A one-off `UPDATE` over production
   `content_data` would be a larger, less reversible act for the same end.

## Explicitly out of scope

- **The staged CTA precedence flip** (`ctafields.go`, trail `2525f980`, the
  `cta_link_integrity` lane). Untouched. That round decides which fields a
  build-time resolver *writes*; this decides whether what was written, by anyone,
  at any depth, points at a real page. Named in the submission so the flip's
  reviewers are told rather than left to measure it.
- **The single-component `content_data` writers** of `bugs_open/136`. They do not
  pass through `SavePageSectionsAction`. Stated in the file header rather than
  left to be inferred.
- **A fleet dispatcher for `check_phantom_internal_links`.** 097 flags it as
  owner-adjacent; building a second undispatched check would repeat `093`'s
  outcome (correct code behind something that never runs).
