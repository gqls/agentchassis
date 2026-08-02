# NOTES — `bugs_open/136` (section-editor slug), sibling link repair

Append-only, newest at the bottom. Technical log: what was tried, what the system actually
said, and every misstep.

---

## 2026-08-02 ~10:20 BST — picking the bug, and the two I put back

Swept `bugs_open/` (60 files) against every `.jsonl` transcript touched in the last 5–10
hours, counting `bugs_open/NNN` mentions per session. That is the check CLAUDE.md asks for
and it earned its keep twice:

- **`bugs_open/155`** (deploy_image_asset resolves by purpose) read like the ideal pick —
  high severity, root cause CONFIRMED by the diagnosis loop, fix candidates written. Session
  `693556a1` has **119** hits on `bugs_open/155` and 113 on `resolveStorageURIFromAsset`. It
  is being fixed right now. `scripts/who-owns.py 155` said "OWNED or recently active" but
  named the *doc-review* and *143* workstreams — i.e. **the lagging-indicator failure mode
  the memory entry describes**: who-owns reads commits, and a session mid-fix has none yet.
  The transcript grep is what actually found the owner.
- **`bugs_open/093`** (stat audit, one guarded call site) is the same *family* as this bug
  and reads as unowned. Its own last update (2026-07-27 triage) says it is **not a code task
  any more**: candidate 1 was built and shipped in v1.0.1172, and the check it was built
  into is fired only by `improvement-sweep`, which has been disabled since 2026-05. It is
  blocked on `bugs_open/083`, which is hot in two live sessions. Left alone.

Landed on `bugs_open/136` **(section-editor slug)**. `136` is one of the ambiguous numbers;
`who-owns.py` matched on the number and surfaced the *other* 136's commits, which is exactly
why CLAUDE.md says resolve by slug and `git log` the file path.

## 2026-08-02 ~10:40 BST — is it still valid, and is the file right about its own targets

```
grep -rn "repairSectionLinks\|RepairPageLinks" --include=*.go platform/ internal/
```
→ three call sites, all pre-existing: the build gate, the outbound rerender seam, the
section save. **Still valid**; nothing has closed it since 2026-07-28.

Two corrections to the bug file's own account, both from reading the code it names:

1. **The report path's prose really is raw.** `renderReportSection` escapes every
   deterministic field with `html.EscapeString` (`esc(...)` on ~25 fields), which had me
   about to write "no anchor can reach this HTML". Then line 413:
   `fmt.Fprintf(&b, `<div class="report-prose"><h2>%s</h2>%s</div>`, esc(sec.heading), body)`
   — `body` is `summary_html` / `candidates_html` / `integration_html` /
   `vendor_questions_html`, LLM-authored, **unescaped**. The bug file is right and my
   first read was wrong; the escaping is dense enough to look total.
2. **The blog listing is not a member of the class.** The bug file lists it as a writer with
   no repair (true) and rates it "lower risk but not zero". It is lower than that: its only
   anchor comes from the template, and the live template has exactly one, `href="{{.url}}"`,
   filled from `pages.url`:

   ```sql
   SELECT function, (SELECT count(*) FROM regexp_matches(html_template,'<a[^>]+href','gi')) AS anchors,
          (SELECT string_agg(m[1],' | ') FROM regexp_matches(html_template,'href="([^"]*)"','g') m) AS hrefs
   FROM content_components WHERE function='content-listing' AND is_active;
   -- content-listing | content-listing | 1 | {{.url}}
   ```

   Since the repair index is built from `pages.url` under a *looser* predicate
   (`status NOT IN ('deleted','archived')`) than the listing's own eligibility floor, every
   href it emits is in the index by construction. Repair there can only no-op. Excluded.

   **Misstep worth recording:** I nearly added the call anyway "for symmetry". A no-op call
   on a live path is not free — it is a `pages` query per rebuild and a new failure surface —
   and "the bug file lists it" is not evidence. The query above took 20 seconds.

## 2026-08-02 ~10:55 BST — the census the bug file marks `[UNMEASURED]`

The file says: *"No count of phantom links actually shipped through these paths … Do not
state a volume until that is measured."* Measured now, fleet-wide, against stored
`page_components.rendered_html` (query in `RUNBOOK_sibling_link_repair.md`). It mirrors
`NormalizePagePath` in SQL — lowercase, strip `#`/`?` tail, strip `index.html`, trim
trailing `/` — and applies `ClassifyLinkScope`'s exclusions (external, `mailto:`/`tel:`/
`javascript:`, `#`-fragment, asset extensions).

**30 distinct (page, href) rows on 7 sites**: 14 `rewrite` (a real `.html` target exists,
the writer just omitted the extension) and 16 `UNLINK` (no such page — a live 404). None of
the carrying components has a `data-runtime-fill` marker, so the whole-input exemption
would not spare any of them. Buckets: `page_type='content'` 14 hrefs / 6 components,
`landing` 2 / 1, `tool` 1 / 1 (by page type of the *carrying* page).

**What this number is and is not.** It is the standing stock of unrepaired internal links in
storage. It is **[NOT ATTRIBUTABLE]** to a writer — a stored href cannot say who wrote it,
and 079's fix is live, so some of this predates it. Anyone quoting "30" as "the sibling
writers produced 30 phantom links" would be inventing the attribution. What it *does*
establish is that the class is live rather than theoretical, and that 8 of the 30 sit in
tool slots (`tool-cta`, `tool-ai-readiness-quiz`, …) — which is why the tool writers are
named as the ranked next candidate.

**Exposure of the path I am actually fixing, stated honestly:** `orchestration_states` (2469
rows, retained from 2026-07-13) contains **zero** runs owned by `section-editor` or
`tool-improver`, and there are **zero** `pages` rows with `page_type='report'`. So both call
sites I am guarding are *reachable and documented* but have not run in the retained window.
This is prevention, not a bleed being stopped, and the bug file should say so.

## 2026-08-02 ~11:05 BST — why the swap branch's UPDATE moves to the caller

`ApplySectionEditAction` has two persist sites, not one: `content_edit` returns its HTML and
the caller writes it (`updatePageComponentAfterEdit`, line 354), while `component_swap`
writes its own row *inside* `applyComponentSwap` (line 798) and then also returns the HTML.
A repair placed at either point leaves the other unguarded — which is this bug's own shape,
one level in. Moving the swap's UPDATE out to the caller gives one repair call before one
persist `switch`.

The locked-component semantics are preserved exactly: `errComponentLocked` was already
handled for both branches at the call site (the swap propagated it out of
`applyComponentSwap` into the `errors.Is` check at line 338), and after the move both
branches return through the same check.
