# PLAN — bug 229: page_components artefact archive + divergence stamp

**Workstream started 2026-08-09 (evening), on the owner's ruling: candidate 1 —
extend the 344 shape.** Owning bug: `bugs_open/229`. Direct continuation of the
`bugfix_226_chrome_divergence` lane (same session took both); chrome's build is
the precedent, but every page-side decision below is made on page-side
evidence, per the bug file's own instruction.

## The ruling and its scope condition

Owner 2026-08-09: candidate 1. The architecture seat's recorded RFC condition
is not tripped — `page_components` is the SECOND table adopting the shape
("fine at two instances; a third needs the shared-abstraction RFC"). A third
adopter must go through architecture review; this is written into the register
entry.

## Measurements that shaped the design (all live `clients_db`, 2026-08-09 ~18:00Z)

- **DELETE is the dominant lifecycle, not UPDATE**: all-time 20,210 inserts /
  19,054 deletes / only 4,928 updates (1,331 live rows). The DELETE+INSERT
  rebuild family is the main event, so the DELETE arm of the trigger is
  load-bearing, not an edge case. `[MEASURED]`
- **The table is warm, not hot**: 27–290 rows touched/day over the last week.
  The 226 scope-out's "orders of magnitude hotter" fear was about row count
  (1,331 vs 57) — the WRITE RATE does not support a fail-open concession.
  `[MEASURED]`
- **Artefact sizes**: avg 5,615 bytes, max 35KB, 7.3MB across the table.
  Growth projection for the archive: ~160 deletes/day × 5.6KB ≈ 0.9MB/day
  (~27MB/month) worst case, on a history table already at 29MB. Real growth
  will be lower (only differing overwrites archive; deletes archive
  unconditionally). No pruning shipped; open-review item, mirroring 344's.
  `[MEASURED, projection stated as such]`
- **`save_page_sections` already snapshots — and drops the artefact when it
  matters most**: its pre-overwrite history INSERT (14,831 of 14,863 rows)
  embeds `rendered_html` into `content_data` ONLY when `content_data` is NULL
  (a COALESCE); when content is present, the artefact is not archived. The
  house archive records the wrapper and loses the bytes — the bug's claim,
  verified in the writer itself. `[MEASURED + read]`
- **Page-cascade deletes make archiving structurally impossible in one case**:
  `page_components.page_id → pages ON DELETE CASCADE`, and
  `page_component_history.page_id` has a plain FK to `pages` — during a page
  cascade the parent row is already gone, so an archive row cannot reference
  it. 740 pages deleted all-time, so the case is real. The trigger SKIPS
  archiving when the page row no longer exists (deliberate full-page deletion
  is not the silent-section-wipe class; and the FK forbids it anyway). Stated
  hole, in the register landmine. `[MEASURED + schema]`
- **Zero rows have NULL page_id** (orphan concern empty). Locks page-side are
  real and already guarded: 30 locked rows; every automation writer goes
  through `pageComponentAgentWritableSQL` (the 058 gate). `[MEASURED]`

## Decision 1 — fail-closed, on page-side evidence

Same answer as chrome, different justification: the write rate is hundreds/day
(measured), the role situation is identical (`clients_user` owns everything;
no permissions path to accidental failure), and the failure mode — history
table broken ⇒ page writes refused — is loud, attributable, one-statement
recoverable (ROLLBACK sidecar). The alternative silently recreates the bug
exactly when the archive is most needed. RFC_017 precedent stands.

## Decision 2 — the trigger archives, a small Go set is loud

Trigger `trg_page_component_artefact_archive` on `page_components`:

- **UPDATE OF rendered_html**, WHEN old non-empty AND `IS DISTINCT FROM` new:
  archive outgoing bytes + digest verdict (op `overwrite`).
- **DELETE**, WHEN old rendered_html non-empty: archive unconditionally
  (op `delete`), UNLESS the parent page row is already gone (cascade case
  above — skip). No comparison is possible across DELETE+INSERT, which is
  precisely why STY-025's losses were deletes.

History rows go into the EXISTING `page_component_history` (the ruling's
"extend" is literal): new nullable columns `rendered_html`,
`rendered_html_digest`, `divergence`, `application_name`, `op` — existing
content_data rows stay valid, existing four app writers untouched, new
`source` value `'artefact_archive_trigger'` keeps the arms distinguishable.
The app-level snapshot in save_page_sections stays as-is (content provenance);
the trigger row is the byte-exact recovery arm. A rebuild therefore writes
both — redundancy accepted and stated (volume booked above).

## Decision 3 — who stamps, who flags (the digest's meaning is "reproducible from content_data")

Stamp `rendered_html_digest = md5(html)` same-statement, ONLY in writers whose
HTML is the pipeline's own render of content:

| writer | stamps? | why |
|---|---|---|
| `save_page_sections` INSERT arm (:868) | YES | the render path |
| `rebuild_blog_listing` UPDATE :322 + INSERT :357 | YES | renders listing from content |
| `section_editor_actions` both UPDATE arms (:1121, :1130) | YES | renders the edited content |
| `create_report_page` UPDATE/INSERT (:216) | YES | renders the dossier |
| `adopt_verbatim` (:510) | **NO** | ported bytes are NOT reproducible from content_data — stamping them would declare recoverable what is not |
| colour-fix rewriters, admin handlers, raw psql | **NO** | artefact patchers; their content must FLAG at the next overwrite (the TRANSIENT class, as chrome's fixer loop shows) |

Loud half (classify → WARN → `page_divergence_overwritten` item,
needs_human_review, no handler — same closure design as chrome's): in ALL
FOUR stamped writers.

> **REVISED after council round 1 (2026-08-09, corr `eee2888b`).** The
> original scope was the two rebuild paths with recorded losses only, with
> "the council can push to widen" written into the plan — and the
> `bug_historian` seat did exactly that, naming it the documented "one call
> site of a shared judgement gets the rigorous fix; the sibling stays
> heuristic" pattern (bugs_open/093's shape). Widened: `apply_section_edit`
> (classify before the persist switch, emit after success — the locked path
> returns early so a refused write cannot emit) and `create_report_page`'s
> UPDATE arm (emit after RowsAffected > 0) now emit too, each filtered to
> the single component they touch. Still quiet BY DESIGN: `adopt_verbatim`
> (operator-initiated port of an unstamped surface — nothing stamped can be
> destroyed there until a stamped row is adopted over, which the trigger
> still archives) and the non-Go writers (admin, raw psql — the trigger
> archives; no Go seam exists to emit from). The same round's other code ask
> — a test pinning predicate PARITY between the classifier and the DELETE it
> speaks for (two files, drift risk) — is
> `TestSavePageSectionsDeleteUsesSameWritablePredicate`.

`save_page_sections` classifies the rows its DELETE will remove, using the
SAME `pageComponentAgentWritableSQL` predicate the DELETE uses, so
locked-surviving rows are not counted as destroyed, and emits after the
DELETE reports rows gone; `rebuild_blog_listing`'s UPDATE arm emits after
RowsAffected > 0, the pattern its lock-refusal emit already uses.

item_key: `page_divergence_overwritten:page_component:<page-uuid-first-8>:<position>:<digest-first-12>`
(site-scoped by `idx_swi_dedup`; digest fragment for within-page repeat
suppression, the 226 round-2 lesson).

## The edits

1. `sql_for_agents/<next>_page_component_artefact_archive.sql` — five nullable
   columns on `page_component_history`, `page_components.rendered_html_digest`,
   trigger + function (fail-closed, cascade-skip), induced DO/RAISE probe that
   self-deletes.
2. `..._ROLLBACK.sql` — drops trigger + function, KEEPS columns and data (a
   rollback must not become the loss it guards against; 344 precedent).
3. `page_component_divergence.go` (new) — classify (predicate-matched) +
   `emitPageDivergenceItem` + WARN, mirroring `site_component_divergence.go`.
4. `save_page_sections_action.go` — stamp on INSERT; classify before DELETE,
   emit after.
5. `rebuild_blog_listing_action.go` — stamp both arms; classify+emit on the
   UPDATE arm.
6. `section_editor_actions.go`, `create_report_page_action.go` — stamp only.
7. `page_component_divergence_test.go` (new) — sqlmock branch tests + stamp
   same-statement pins via expected query text (not comments — the
   source-scan landmine).
8. Register STY-056 + `bugs_open/229` update, same commit as the migration.

## Verification (the bug file's protocol, one table over)

Hand-patch a throwaway string into a test page section's `rendered_html`
(stamped first by a save), run the section rebuild, require: archive row with
the patched bytes, WARN, item with the digest fragment; negative control: an
untouched byte-identical rebuild archives nothing on the UPDATE path and emits
nothing; DELETE recoverability: a rebuild's DELETE leaves every pre-delete
artefact recoverable from history (op='delete' rows). Mind: chassis log
retention is seconds — arm followers before dispatching.

## Consumers told (owner ruling 2026-07-29 §3)

> **WIDENED after council round 1** (guardian, medium: "it is not four named
> consumers, it is every writer of that column, present and future"). The
> fail-closed trigger's blast radius is the COMPLETE writer inventory of
> `page_components.rendered_html` — the bug file's eight classes, restated
> here so the told-set matches the affected-set: `save_page_sections` (full
> save, DELETE+INSERT), `rebuild_blog_listing` (UPDATE + INSERT arms),
> `section_editor` (content_edit + component_swap persists),
> `create_report_page` (UPDATE + INSERT arms), `adopt_verbatim` (ported-page
> replace), the colour-fix rewriters (`fix_harcoded_colours`,
> `fix_forced_text_colours`), core-manager admin handlers, and raw psql. A
> broken `page_component_history` halts ALL of them, loudly — that is the
> deliberate fail-closed trade (measured write rate, one-statement rollback),
> and it now says so here rather than only in the risks block. Any FUTURE
> writer of the column inherits both the archive and the halt.

- **STY-025's lane / interactive-tools owners**: rebuild-destroyed tools are
  now archived at the wipe (op='delete') and — where the wipe is
  save_page_sections, the blog rebuild, a section edit, or a report
  regeneration — flagged.
- **colour-fix writers**: your rewrites now archive their predecessors and
  will be flagged as hand-patched at the next rebuild — same TRANSIENT
  bargain as chrome.
- **admin-dashboard**: an admin page edit's predecessor is archived; the edit
  itself flags at the next rebuild unless the slot is locked.
- **229's own filer (the 226 council trail)**: the bug_historian seat's gating
  objection is answered by this build.
