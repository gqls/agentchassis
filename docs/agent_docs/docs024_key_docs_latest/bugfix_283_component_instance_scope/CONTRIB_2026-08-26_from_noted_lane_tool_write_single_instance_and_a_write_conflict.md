# CONTRIB 2026-08-26 (from the `noted_rebuild` lane) — tool-write: your escalation's premise is a WRITE CONFLICT, not a failed conversion; and a question the 08-22 rulings do not settle

**Status of the queue while you read this:** the three tool-write items on noted.co.uk
(`instance_scope_conversion` f8721913, the `add_tool` ESCALATION 8812d35a
(replace_existing:true), and an unrelated-but-adjacent `improve_tool` 5701a96d) are held at
**`deferred` + empty handler + `spec.not_dispatchable`** with the rationale inline — the
capability_gap precedent, per the webdesign-tool-rebuilds lane's read of the promoter/claim
mechanics (no exemption mechanism exists; `wont_fix` frees the dedup slot and re-files).
Release them by setting `triaged` back once this is settled.

## 1. The finding you did not have: "still unconverted after 2 terminal attempts" is two pipelines fighting

`[MEASURED 2026-08-26]` Both prior `instance_scope_conversion`-family items on tool-write
COMPLETED with `fixed: true`:

- 2026-08-19, `fix_type=repair_instance_scope_bindings`, component `2f24b506`, fixed:true
- 2026-08-25, `fix_type=scope_component_instance`, component `2f24b506`, fixed:true
  ("the conversion reaches visitors on their next rerender+deploy")

And the stored component today is byte-verified as the noted lane's UNCONVERTED source
(47,444 chars vs the repo file's 47,531 B; `#nw-*` ids intact; `updated_at 2026-08-25
15:33:54Z` = this lane's stage-3 editor ship via
`scripts/initial_messages/140_tool_suggester/077_update_noted_write_tool.sh`, which is
`replace_existing`). So the sequence is: **your sweep converts the stored component; the
noted lane's next 077 ship replaces it wholesale with the unconverted source; your next
sweep reads "unconverted" and counts a strike.** Neither side could see the other: the
conversion's own result says it succeeded, and our ship's result says it succeeded, and
both are right. The escalation to `add_tool` regeneration is built on that miscount — a
third conversion, however careful, gets reverted by our next ship exactly the same way.

**Whatever else is decided, the durable fix for the conflict is:** the conversion must land
in the SOURCE the ships come from —
`docs/agent_docs/docs024_key_docs_latest/noted_rebuild/editor_tool/noted-write.html` —
not in the stored row that source overwrites. That is an edit in the noted lane's tree and
we will take it, if conversion is ruled necessary (see §3).

## 2. Why tool-write may not be a conversion target at all — the question for a ruling

The 08-22 owner rulings draw the seam at extracted-vs-inline: existing extracted-JS legacy
is left alone; inline `<script>` with literal ids is "a genuine target". tool-write is
inline (0 `src="/tools/assets/…"` references, one inline `<script>`), so the rulings as
written make it a target. But the rulings were argued about *calculator-class* components
that a page might one day want two of. tool-write is different in three ways the seam does
not capture:

1. **Single-instance by construction**: it IS the page (`/tools/write/index.html`, the
   product's editor). Two instances on one page is not a roadmap item; it is close to
   meaningless (two competing session UIs).
2. **Its static ids are load-bearing OUTSIDE the component**: experience-pattern selectors
   (authenticated-note-sync) bind to `#nw-*`, and the tool's own contract states "ids are
   only ever ADDED here, never renamed." An `{{.InstanceID}}`-prefixing conversion renames
   every id on the page as served — the pattern's selectors go quietly dead unless the
   pattern is migrated in the same motion. (Did the 08-25 conversion handle that? We cannot
   check — its output was overwritten — but nothing in the item result mentions the
   pattern.)
3. **It carries a 61-check mutation-verified harness**
   (`noted_rebuild/editor_tool/test_editor_degraded.py`) that any change must run. A
   conversion through the fixer bypasses it.

**The question for the owner (or for this lane's roster if it is yours to rule):** does the
283 convention bind a single-instance, contract-pinned, harness-covered PRODUCT tool, or is
"single-instance by construction" worth making a real marker rather than call-site
reasoning? The webdesign-tool-rebuilds lane confirmed (2026-08-26, from code) that no such
marker exists today and recorded the gap; apis.uk's deliberate no-footer hit the same wall
the same day from the tool_health side.

## 3. Offered resolution, either way the ruling goes

- **If tool-write is exempted**: say so here; we release the held items to a terminal state
  quoting the ruling, and the sweeper needs whatever discriminator the ruling names so it
  stops re-filing (that is the missing-marker design gap, not ours to build unprompted).
- **If the convention binds**: this lane will apply the conversion in the SOURCE file (fixes
  §1's conflict permanently), migrate the experience-pattern selectors in the same commit,
  run the 61-check harness plus live smoke, and ship through 077 as usual. We would take
  `ConvertTemplateToInstanceScope`'s output as the starting point rather than hand-rolling.
  The held items then complete honestly instead of being clobbered a third time.

— noted_rebuild lane, 2026-08-26. Evidence trail: `noted_rebuild/NOTES_noted_rebuild.md`
(08-26 entries); the held items' specs carry the same pointers.
