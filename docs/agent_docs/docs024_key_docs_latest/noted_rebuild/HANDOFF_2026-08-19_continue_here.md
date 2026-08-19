# HANDOFF — noted.co.uk, continue here

**Written 2026-08-19 ~15:45 UTC. Supersedes `HANDOFF_2026-08-12_continue_here.md`
(which carries three dated banners — read it only for the history).**
Standalone: you should not need anything else to start. Then: `NOTES_noted_rebuild.md`
(the technical log incl. every misstep — long, but the 08-17→08-19 entries are the
live ones), `README_where_we_are.md` (owner's plain-prose log),
`RUNBOOK_cutover.md`, `SUMMARY_2026-08-16_noted_rebuild.md`.

## 1. What this is

noted.co.uk — a note-taking product (text, voice, photos) — is **LIVE as a
framework site** at the apex, cut over from the legacy B2 bucket on 2026-08-16.
Engine (accounts/notes/media API) at `app.noted.co.uk`, Go binary + Postgres on
the webdesign box, loopback-bound behind nginx + cloudflared. The legacy
browser-only app is retired (redirect → the rescue tool); its bytes survive in
three places. **Nothing here is hand-built; every page is framework-owned.**

## 2. CURRENT STATE (all verified at the artefact, 2026-08-19)

- **10 pages live** incl. `/privacy.html` (owner's copy VERBATIM, 22/22 — the
  wording is canonical in `evidence_base.supplied_copy.privacy` AND inline in
  `writer_block`, because **only the writer_block string reaches the writer's
  prompt**), the editor `/tools/write/` (the sign-in page), the rescue tool
  `/tools/legacy-rescue/` (reads the old app's IndexedDB, same origin — tested,
  mutation-verified), and two guides.
- **Header CTA "Get Started" → `/tools/write/index.html` BY DERIVATION**: the
  opt-in `sites.content_data->>'header_cta_url'` override shipped (commit
  `229e14e74`, in build `0b185bad2…`) and a chrome re-render under it proved the
  derivation. No interim patches remain load-bearing.
- **Contact email is `noted@contactforsales.com` everywhere** (sites row, chrome
  footer, all components, the privacy copy; "a person will answer" removed
  2026-08-18). ⚠ Cloudflare email-obfuscation rewrites addresses at the edge —
  **a live grep for an email tests the obfuscator, not the page**; measure at the
  box or decode the `data-cfemail` span.
- **CSS RESTORED 2026-08-19** after the 08-18 loss (§3): theme v23 = 20,190
  chars; DB == repo == box == live (20,367 bytes served).

## 3. THE CSS INCIDENT — the thing this handoff must not let get lost

**What happened (all links config- or git-visible, quoted in `bugs_open/198`'s
second-incident section):** noted's `css_themes` row was **born empty** on
08-15 — created 67 seconds before webdesign-agent committed the real 17,475-byte
stylesheet to git only. Three days of silent split: git full, DB empty, page
perfect. On 08-18, 21 contrast findings ran css-patch-agent 21 times in 16
minutes: each **guarded DB append was correct** (bug 198's candidate-3 fix, live
since 08-10, held) — but `deploy_css` commits the **whole DB value** to
`assets/css/styles.css`, so the file went 17,475 → 91 → … → 2,381 bytes. The
contrast auditor then kept filing findings against the page IT had unstyled —
the loop amplified its own damage, every run green.

**Why the framework missed it:** `bugs_open/198` (08-04, relojistas incident)
had named the shrink guard "absent on BOTH writers". The fix closed the **DB**
writer only; noted's incident walked through the **git** writer — the named,
unclosed door — seeded by the new birth-empty defect that 198's first incident
did not have. No check compares DB theme vs deployed file.

**How the framework fixes it** (candidates appended to 198, ordered by
door-closing; 198 is **owned by the vigilant_designer lane** — contribute, do
not compete):
1. **git-writer shrink guard** — refuse a `git_commit` payload ≪ the file it
   replaces (closes both known routes at the last writer);
2. **birth guard** — `fork_theme_from_site` / `install_site_composition` refuse
   or `needs_review` an empty `css_content` insert;
3. **drift check** — DB theme length vs deployed `styles.css` size, fleet-wide
   (the landmine's one-query tell, mechanised).
One link is `[INFERRED]` (that fork inserted empty `renderedCSS` on 08-15) with
its verifying step written in 198 — **a 090 run is the right first move for
whoever builds the guards.**

**Repair recipe if it recurs anywhere:** LANDMINE "A `css_themes` row can be
BORN EMPTY…" — last-good git version + provenance comment + accumulated patches
(the DB value IS the patch list), deploy via git-adapter with the site's real
`repo_name` (**`vm-sites` for noted — the default `sites` would `--delete` the
wrong estate**), verify DB==repo==box==live.

## 4. OPEN THREADS, in order

| # | thread | state |
|---|---|---|
| 1 | **Council round 2 on the CTA override** (`RESUBMIT` under corr `89f3331e-57f4-4f8f-8f58-de6222d17337`) | Round 1 REVISE was right about the plan (data edit omitted) and already false about the world; round 2 submitted 08-18 with the executed config change included. **Check the verdict**; commit `229e14e74` carries `Council-Submitted` and auto-credits |
| 2 | **Bug 198 guards** (§3) | Not built; owned by another lane; noted contributed the incident + candidates. 090 recommended before Go changes |
| 3 | **Account deletion on the engine** | Absent; smoke tests leave throwaway accounts; the privacy copy promises "close your account" — **that promise currently has no mechanism**. Pre-real-sign-ups work |
| 4 | **Experience patterns at `proposed`** | Both bound (all four bind doors passed); promotion to `bound`/`verified` needs the patterns out of `draft` + the experience loop's green run; 3 checks stay Playwright-only (runner cannot seed IndexedDB / order after sign-in) |
| 5 | **Three `detected` items** (`needs_composition`, `needs_design`, `evaluate_tools`, 08-10) | Inert — `detected` is not dispatchable; nothing acts on them. Known since the first handoff |
| 6 | **Mail routing for `noted@contactforsales.com`** | Owner's to confirm — the site now names it |

## 5. TRAPS — the ones a fresh session hits first

- **Verify at the artefact, never the status** — this lane's recurring lesson:
  `complete` work items that changed nothing (3×), a `deploy_result` naming a
  commit that never happened, and 21 green runs that destroyed the CSS.
- **`page_rerender` re-assembles; the section editor re-renders.** A content fix
  goes `content_data` → section-editor `content_edit` (074/074b/074c scripts);
  writing `rendered_html` alone dies on the next render; a REGENERATION replaces
  `content_data` (re-run 074b for privacy if that ever happens).
- **The estate dispatch queue is ONE line, oldest-first, fleet-wide** — a fresh
  item sits behind every older site's backlog. Measure *your* position
  (items older than yours), never fleet health. Hand-filed `page_rerender` items
  need the **`page_id` COLUMN** set (the handler ignores your spec) — copying a
  row without resetting it deploys the WRONG PAGE while reporting success.
- **`kubectl exec -i` inside a stdin-fed `while read` loop eats the loop's
  remaining input** — exit 0, partial work. Feed loops via fd 3.
- **Byte vs char**: `length()` in SQL is chars, files are bytes, `${#var}` is
  chars — this lane tripped twice.
- **A `.bak` left in `/etc/nginx/sites-enabled/` is glob-included** and breaks
  the next reload. Backups → `/root/nginx-backups/`.
- **cloudflared: `systemctl restart`, never `kill -HUP`** (terminates, no
  `Restart=`). Shopfront control (`Host: webdesign.uk` → `127.0.0.1:8080`)
  before/after every box change, against a baseline YOU take (that lane ships
  continuously).
- **`sites.github_repo='vm-sites'` is load-bearing safety** — the default routes
  deploys at the bucket serving... nothing now, but B2 still archives the legacy
  app; do not "tidy".

## 6. COMMANDS (each with its gotcha, fuller set in RUNBOOK files)

```bash
# box
ssh -i ~/.ssh/webdesign_box_ed25519 root@webdesign.vs.mythic-beasts.com
# live smoke of the editor (base URL arg — run against the APEX post-changes)
/home/ant/.venvs/vonc_pw/bin/python docs/.../noted_rebuild/editor_tool/smoke_live_editor.py https://noted.co.uk
# rescue-tool + origin premise (seeds via the rescue page since /legacy-app/ retired)
/home/ant/.venvs/vonc_pw/bin/python docs/.../noted_rebuild/legacy_tool/probe_origin_after_cutover.py
# privacy copy: edit the DRAFT, run the checker, evidence_base scripts, then 074b
python3 docs/.../noted_rebuild/COPY_2026-08-12_privacy_check.py
# theme health (the CSS landmine's one-query tell)
# SELECT s.domain, ct.version, length(ct.css_content) FROM sites s
#   JOIN style_collections sc ON sc.id=s.style_collection_id
#   JOIN css_themes ct ON ct.id=sc.css_theme_id WHERE s.domain='noted.co.uk';
```

## 7. Error record of the closing sessions — read before trusting the above

Every one is corrected in place in NOTES with what caught it: the sweep that did
1-of-6 silently (stdin drain); "no framework path to add one content page"
(wrong store — the plan is TABLES); the owned_page_review prediction (rule
order); the writer instruction pointing at a JSON path that never travels; the
instruction that MODELLED the banned phrase it forbade; two verbatim-checker
false alarms (tag-strip space, byte-vs-char); the live email grep testing
Cloudflare's obfuscator; and the CSS repair guard asserting bytes against chars.
The habit that caught all of them: **measure at the artefact, and when a check
fails, suspect the instrument before the subject — but prove it.**
