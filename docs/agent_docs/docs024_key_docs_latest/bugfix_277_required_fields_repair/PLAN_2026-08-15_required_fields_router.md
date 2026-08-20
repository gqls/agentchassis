# PLAN — bugfix 277: the fleet-wide repair router for `required_fields_missing`

**Owner ruling 2026-08-15 (bugs_open/277):** "we should create a repair handler fleet wide."
This lane builds it, as the next concrete increment of bugs_open/033's owner-ruled reframing
(2026-07-25: the framework, not a person, resolves these classes). Session: "bugfix 033".

## Why no 090 diagnosis run (the 2026-07-31 ruling's named escape hatch, used openly)

The mechanism is not in dispute and was verified first-hand rather than through the loop:
"no handler in the fleet claims this type" is a grep (`HandlerAgent: ""` at
`check_required_fields_missing.go`, one producer, zero consumers in code or
`agent_definitions`) plus live queries. > **CORRECTED 2026-08-15 (caught by the council's
prior_art_librarian seat):** this doc and the round-1 rationale said *"exactly one item in
platform history ever terminal"*, quoting `revalidate_review_queue_action.go:64-66` — a
2026-07-25 measurement repeated as current. Queried on objection: **50 complete**, almost all
the 033 drain's own closes since 07-27. The design argument never rested on the figure, but
the figure was stale. Logged in `WRONG_CALLS.md` 2026-08-15. The change implements
an explicit owner ruling, not a diagnosed root cause; the design risk was pressure-tested by a
planning agent against the code (verified mechanics recorded in NOTES) and by the council gate
(submission `7b0e2833-715f-4a9a-897b-efd913073582`).

## Design decisions and their reasons

1. **Router, not repairer** (IMG-071 / seed 397 pattern). The population is heterogeneous and
   two classes MUST NOT be auto-repaired: blob components (content_data NULL under serving
   rendered_html — regeneration replaces served HTML, bugs_open/263) and owned/tool pages with
   no plan (the owned-page guard, itself an owner ruling — reconcile_site_plan decision 3).
   Census 2026-08-15 (n=44): no_content_data 35 / stale 6 / no_plan_generic 1 /
   no_plan_owned 1 (the gas converter) / partial 1.
2. **Park-in-place, not checkpoint_for_review, for the two human classes.** The checkpoint
   action writes no item_key (no dedup) and hardcodes `handler_agent='human-review'`
   (unregistered); completing the original releases its dedup key → producer re-raises →
   two-strike births endless `unresolved` rows. Parking at `needs_human_review` HOLDS the key:
   churn is structurally impossible, the row stays on the dashboard with the router's triage in
   the error column, and the revalidator remains a second close path. Parking one's own item is
   first-class: `complete_work_item`'s guard no-ops benignly (`load_work_item_actions.go:956-978`).
   **This was a correction to this session's own first design** (checkpoint-and-complete),
   found by the planning agent's review — recorded in NOTES.
3. **Classification key is (page_name, slot_name)**, the revalidator's own key — never
   `spec.component_id` (016b §9: 11/45 items resolved to nothing when keyed on component_id).
4. **Conversions are born `triaged`** — the `detected` promoter is disabled (bugs_open/083);
   a `detected` item is stranded. Conversion item_keys are stable (`content_rewrite:from_rfm:`
   + component_id) so the two-strike brake works.
5. **`partial` converts to `content_rewrite` with `mode='edit_live'`** (PBP-028's third
   emitter, clause updated) — the writer edits current prose rather than fabricating; bug
   238's resolver-key protections confirmed in the running binary (stamp `a2a6912…`, both fix
   commits ancestors).
6. **Producer flips to routed-from-birth** (`HandlerAgent=required-fields-missing-handler`,
   `Status='triaged'`) — Go, inert until a chassis roll; the seed + assignment carry the live
   half until then.

## Phasing

1. Census (done — output saved beside this file; the seed's exact embedded SQL re-run against
   the five canary candidates routed all five as the census predicted, before any apply).
2. Council submission `7b0e2833-715f-4a9a-897b-efd913073582` (submitted 2026-08-15 ~11:0x).
3. Commit + apply seed 410 (inert; verify block asserts 0 assigned).
4. Canary assignment (4 rows: stale 332bb3f6, partial 4fa5b019, blob e512af8a, gas converter
   483fb749) → verify each arm → fleet assignment.
5. Commit the Go producer change (rides next chassis roll; post-roll re-run the assignment
   UPDATE once for stragglers filed pre-roll).
6. 033/277 bug files updated; 277 stays OPEN until fixed-AND-live per the closing bar.

## Out of scope, recorded

Blob decomposition (staged_component_build's domain), tool-page rebuild (tool lane),
033's other remaining pieces (Retry-refuse for handler-less items; owner decisions B/D; D3
identity; revalidator v2; the other ~20 uncovered types).

---

## ADDENDUM 2026-08-20 — owner ruled YES on the rendered_html repair route; design

**Owner (chat, 2026-08-20, this session):** *"Do those seven findings get a repair route?
Building one means a transform that edits finished HTML directly — I think yes."* That answers
the open question in `HANDOFF_2026-08-20_continue_here.md` §7 / `bugs_open/277` §5.6.

### The design, and why each piece sits where it does

1. **The transform lives in `datahelpers`** (`ConvertLiteralCodeSpansInHTML`): same package as
   the single-sourced pattern set (`literal_markdown.go`) AND the detector's parser skip-set
   (`nonAssertionElements`, claims.go:311), so the repairer structurally cannot drift from the
   detector's view of the surface. Tokenizer-splice (`x/net/html` `NewTokenizer`), NOT
   parse+re-serialise: `html.Parse` wraps fragments in `<html><body>` and normalises the whole
   document, so the diff would exceed the edit; the tokenizer copies `Raw()` bytes verbatim and
   the output differs ONLY in converted text nodes. Conversion pattern is DERIVED from
   `MDCodeSpanRe` and STRICTLY NARROWER (interior also excludes `<`/`>`): conversion ⊆ detection,
   so anything detection flags that conversion cannot safely reach is left for the verifier to
   fail → attempts → human review. The safe direction, stated rather than discovered.
   **Why markdown→HTML is safe HERE and banned in the stripper** (literal_markdown.go header):
   that ban guards values feeding INTO the unescaping render pipe; this transform operates on
   the pipe's OUTPUT, inserts only the fixed strings `<code>`/`</code>` around bytes already
   being served as text, and never emits any byte of LLM-authored markup.

2. **The carrier is `apply_section_edit`, a third `edit_type: rendered_html_transform`** —
   reuse over new machinery. The action already owns every control the route needs: lock gate,
   RFC_015 decision gate, per-slot floors, regulated-identity refusal, link repair, ONE persist
   switch, reassembly + git_commit + Cloudflare deploy. The owned-page guard's own prose names
   section-editor as how owned pages are legitimately edited. New authority (writing
   rendered_html NOT derived from content_data) ships per the 2026-08-02 §2 ruling: **opt-in,
   unsafe default OFF** — config key `allow_rendered_html_transform`, enabled by migration 513
   on section-editor's apply_edit step only. Optional input `transform_name` (7th of 10 —
   budget checked 2026-08-20, `apply_section_edit` was at 6). Persist is HTML-only
   (`updatePageComponentAfterEdit` with nil content_data — the branch already exists):
   content_data is deliberately untouched, because for this population it is provenance
   metadata, not source.

3. **Routing is decided by the DETECTOR at filing time** — migration 499's test
   ("read the finding's source, then ask whether content_data can reproduce rendered_html"),
   automated: a page routes `section-editor` + the new edit_type IFF every finding is
   `source=rendered_html` ∧ `pattern=code_span` ∧ one single slot ∧ that slot's component
   CANNOT regenerate (template names ≥1 top-level field, content_data holds none of them).
   Anything else keeps today's route. item_type, item_key, verifier all unchanged — the
   whole-page verifier survives the re-route by construction (its header says so).

4. **The promoter holds the new pair** (`literal_markdown → section-editor`, 0 lifetime
   completes → held by 444's ≥1-complete door). **The bootstrap is the canary run this lane
   already knows how to do by hand** (083, commit 8d77196ad): after the roll, manually dispatch
   ONE of the 7 — that run is simultaneously the artefact-level proof the close needs and the
   completion that opens the promoter door for the other six.

### Order of operations
code + tests → migration 513 (+ROLLBACK) → concept register entry (same commit as the seam —
ordering-exemption condition 2) → council submission (Council-Submitted trailer) → commit by
pathspec → chassis ships on the next fleet roll → apply 513 (safe either side of the roll: old
binary never reads the key; new-shape items only exist after the new detector rolls; a
config-not-yet-applied window fails into the attempts ladder and self-heals) → canary one item
→ verify at the served bytes (`curl` + `<code>` present, backticks gone) → remaining 6 flow.

### Scope statement
code_span ONLY. bold/heading/md_link in rendered_html stay on today's route (escalation to a
human): the owner ruled on 7 code_span findings, and each further pattern is its own
transform with its own safety argument. The residual is stated, not silent.
