# PLAN — bug 226: chrome rebuild silently discards hand-patched content

**Workstream started 2026-08-08.** Owning bug: `bugs_open/226`. Filed by the oufe
rerender-safety lane at the council's direction (trail `5c18ccaa`, round-2 gating
objection: the STY-052/053 carriage fix "re-armours a symptom rather than closing
the mechanism"). This lane closes the mechanism.

## The mechanism being closed

`site_components.rendered_html` is a stored artefact (bugs_open/117). Every
legitimate rebuild replaces it outright
(`render_site_components_action.go:938-943`) with no comparison against what it
is replacing. Content that exists ONLY in the artefact — a psql `replace()`, an
admin edit, a chrome-fix append — is destroyed silently, with no record of what
was lost. Measured twice on oufe.com (mig 268's footer note, FIX_2026-07-26's
CTA rewrite), both unnoticed for eight days.

## Decision 1 — bug 226's fix candidate 1 is NOT implementable as written (CORRECTION to the bug file, recorded there too)

The bug proposes: "re-render with the row's *stamped* inputs (117 stores them)
and compare to the stored HTML". **The stamp cannot do that.** `render_inputs`
(`platform/orchestration/datahelpers/chrome_render_inputs.go`) is a jsonb map of
**md5 digests** of the input stores — you can tell whether inputs changed; you
cannot recover the inputs to re-render from. The LANDMINES entry "a bug file's
FIX CANDIDATE can be refuted by that same file's own MEASUREMENT NOTES" applies:
the bug's own "What already exists" section describes the fingerprint correctly.

**The variant that IS implementable, and is strictly stronger:** stamp a digest
of the **artefact itself** (`rendered_html_digest = md5(rendered_html)`) in the
same UPDATE that stores the bytes — the same same-statement principle 117 used
for `render_inputs`. At the next overwrite, compare the stored bytes against the
stamp:

- digest matches ⇒ the bytes are exactly what the render path last wrote ⇒
  machine-made, reproducible from inputs ⇒ overwrite freely;
- digest differs ⇒ some other writer changed the artefact since ⇒ hand-patched;
- digest NULL ⇒ pre-fix row, cannot distinguish (converges as the fleet
  re-renders; 46 of 57 rows are unstamped today, 11 stamped by the 117 work).

This drops candidate 1's determinism requirement entirely — no re-render is
needed to classify.

## Decision 2 — the archive is a table trigger, not Go call-site edits

The complete writer inventory for `site_components.rendered_html` (measured
2026-08-08, `grep -rn "UPDATE site_components\|INSERT INTO site_components"`):

| writer | destroys? |
|---|---|
| `renderAndStoreSiteComponent` (render_site_components_action.go:938) | overwrite — THE destruction point |
| `relinkSiteComponent` (site_component_lock_guard.go:157) | sets NULL on repoint — erases |
| `setSiteComponentHTML` (site_component_lock_guard.go:119) | whole replace (chrome fixes) |
| `appendSiteComponentHTML` (site_component_lock_guard.go:137) | append (self-described TRANSIENT) |
| core-manager admin handlers (page_admin_handlers.go:1116) | human edit, different service |
| raw psql / migrations (the 268 incident class) | invisible to any Go guard |

Guarding call sites covers three of six and can never cover the sixth — the raw
SQL hand-patch is the very writer this bug is about. A
`AFTER UPDATE OF rendered_html` trigger with a
`WHEN (OLD.rendered_html IS NOT NULL AND OLD.rendered_html <> '' AND NEW.rendered_html IS DISTINCT FROM OLD.rendered_html)`
gate archives the outgoing bytes from **every** writer, present and future —
the bad state (silent destruction) becomes unrepresentable at the table, which
is the ranking rule from `order_fix_candidates_by_what_closes_the_door`.

Costs, named honestly:

- **Invisible to Go greps** (the bugfix-205 landmine class). Mitigations: a
  LANDMINES.md entry with footprint `site_components`, the STY-054 register
  entry, and a comment at the render-path store statement pointing at the
  trigger.
- **Fail-closed**: a failing history INSERT aborts the overwrite. Chosen
  deliberately (you cannot destroy what you failed to archive) — precedent
  RFC_017. The failure mode (history table broken ⇒ chrome rebuilds error) is
  loud, not silent.
- **Live before the Go half** — DB config is live immediately; the image rides
  the next chassis build. This ordering is CORRECT here: the trigger is
  self-contained, and it must be armed **before the 117 staleness wave** (built
  2026-08-08, rides the same next roll) rebuilds the 46 unstamped rows.

## Decision 3 — divergence is loud, never blocking (no new refusal authority)

On detecting a hand-patched artefact about to be overwritten, the render path
WARNs and files a `chrome_divergence_overwritten` work item
(needs_human_review), **then proceeds**. It does not hold: refusal is what 069
locks are for, and per the owner ruling of 2026-08-02, new authority on a
shared seam would need an opt-in field — this design adds none. The three legs
end up: **prevent** = 069 locks · **reproduce** = STY-050 carriage ·
**detect + recover** = this work (the leg 226 names as missing).

Classification for the work item happens in Go via one SELECT before the store
UPDATE (`rendered_html_digest IS NOT NULL`, `= md5(rendered_html)`). The tiny
read-then-write race affects only the *loudness* half; the *recovery* half (the
trigger) is atomic with the overwrite by construction.

**No digest backfill.** Backfilling `md5(rendered_html)` onto existing rows
would stamp every unknown hand-patch as machine-made — declaring the key
silences the detector (WRONG_CALLS class). Unstamped rows stay NULL and
converge through real renders.

## The edits

1. `docs/agent_docs/sql_for_agents/344_site_component_history_divergence_guard.sql`
   — `site_component_history` table (house shape from `page_component_history`:
   entity refs + payload + source discriminator; archives `rendered_html`,
   `render_inputs`, digest verdict, `application_name`),
   `site_components.rendered_html_digest` column, the archive trigger +
   function, induced-probe verify block (DO/RAISE — a SELECT cannot stop a
   COMMIT).
2. `..._ROLLBACK.sql` sidecar — drops trigger + function; deliberately KEEPS
   table + column (they hold archived artefacts; a rollback must not become the
   loss it guards against).
3. `render_site_components_action.go` — stamp `rendered_html_digest = md5($1)`
   in the guarded store UPDATE; classify + WARN + emit work item before it.
4. `site_component_divergence.go` (new) — classification query +
   `emitChromeDivergenceItem` (mirrors `emitLockBlockedChange`; severity high
   for header/footer, medium otherwise, matching the 069 emitter's argument).
5. `site_component_divergence_test.go` (new) — sqlmock behaviour tests
   (classification branches; item emitted only on hand_patched) + contract pin
   that the digest is stamped in the SAME statement as the bytes.
6. `styling-render-pipeline.md` register entry **STY-054** (same commit — the
   platform-seam registration rule), naming the trigger landmine and the
   `chrome_divergence_overwritten` producer + item_key shape (RFC_010 ruling 1).
7. `bugs_open/226` — visible CORRECTED note on candidate 1 + pointer here.
8. `LANDMINES.md` append (+ `landmines-sync.py --apply`) — the trigger's
   invisibility to Go greps, footprint `site_components`.

## Verification (from the bug file, unchanged)

Hand-patch a throwaway string into a test site's footer `rendered_html`,
trigger a chrome re-render, require: WARN + work item naming the divergence,
and the outgoing HTML recoverable from `site_component_history`. Negative
control: an unpatched (stamped, matching) slot rebuilds with no item and —
if byte-identical — no archive row.

## Consumers told, not merely measured (owner ruling 2026-07-29 §3)

- **117 lane** (`bugfix_117`): your wave now archives every artefact it
  replaces; first-wave rows will classify `unstamped` (expected, converges).
- **oufe rerender-safety lane** (filed 226): candidate 1 corrected as above;
  your 339 carriage work is untouched and remains the reproduce-leg.
- **chrome-fix writers** (`fix_component_template` overflow patch): your
  self-described TRANSIENT patches are now archived at the wipe and named in a
  work item — the wipe stops being silent, which is what "TRANSIENT" always
  warned about.
- **admin-dashboard lane**: a dashboard chrome edit already auto-locks its slot
  (069); if a lock is lifted later, the edit now archives instead of vanishing.
