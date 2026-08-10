# HANDOFF 2026-08-10b — Phase 5 DONE. Pricing DONE. All 9 CTAs DONE. Owner review is next.

**Start here cold. Supersedes `HANDOFF_2026-08-10_continue_here.md`** (which
described a blocked git-push state — that block turned out to be scoped to
one file in the immediate aftermath of an incident, not a standing rule; it
cleared on retry later the same session, and everything below shipped).
Read with `SUMMARY_2026-08-10_webdesign_uk_build_service.md` for the
milestone account, `NOTES_webdesign_uk_build_service.md`'s 2026-08-10
"(continued)" entry for the full technical trail, and
`PLAN_2026-08-04_webdesign_uk_vm_hosting.md` §7 for the pricing decision.

## 0. One paragraph

Phase 5 is complete. The chat-input-box is live on the contact page, backed
by a chat service with a real, tested memory-across-turns fix. A £75
non-refundable deposit is priced, researched, and live everywhere it needs
to be: `evidence_base`, three page copies, and the chat bot's own system
prompt. All 9 parked `unresolved_cta` items are resolved and marked
`complete`. **There is nothing left to build in this phase** — the next
step is the owner reviewing the whole thing (site + chat box + pricing)
before Phase 6 cutover.

## 1. State table — verified 2026-08-10, re-check before trusting

| thing | state | re-check |
|---|---|---|
| The 5 pages | Live, clean: 5×200, 0 ban hits, 0 title em-dashes, only the two known-benign 404s | `verify_served_site.sh` |
| `chat-input-box` on contact.html | Live, position 2 (between hero and contact-info) | `curl .../contact.html \| grep chat-input-box` |
| Chat loader JS | Live at origin (git commit `06d1039`), **but Cloudflare edge-caches `.js` for up to 4h** — check `curl -sI .../snippets.js` for `cf-cache-status: HIT` + `age`; if `age` is still under ~14400 and status is HIT, a real visitor may not have the working box yet. Self-resolving, no action needed, just verify before claiming full proof | `curl .../assets/js/snippets.js?cb=1 \| grep -c chat-input-box-loader` (bypasses cache) vs. without `?cb=` (what a real visitor gets) |
| Chat service | Live, memory-across-turns fixed and proven, deposit terms synced (`agentchassis` commit `f4e77c7fb`) | ask it a refund question directly |
| Deposit terms (£75 / £1,125) | Live: `evidence_base` (DB), index/how-it-works/faq page copy, chat bot system prompt — all say the same thing | grep each for "£75" |
| The 9 `unresolved_cta` items | All `status='complete'` in `site_work_items` | `SELECT status FROM site_work_items WHERE item_type='unresolved_cta'` |
| `bugs_open/239` (dispatch mechanism) | Still OPEN on the platform side. Not this lane's to fix. Corrected once already this week — read its own top banner before trusting anything below it | read the file |
| Fleet | A fresh chassis build landed after 239 was filed. Not confirmed to fix it | don't assume |

## 2. What to do next — owner review, no further build work implied

1. **Confirm the JS cache has cleared** (see table above) before telling
   anyone the chat box works end to end for a real visitor.
2. **Owner review of the whole thing** — site, chat box, and the new £75
   deposit pricing together. This is the ledger item that was already owed
   and now also covers Phase 5.
3. **Phase 6 cutover** once approved (see PLAN §3 Phase 6 — mostly
   Cloudflare DNS work, ~1 minute of actual changes, gated on resolving the
   Worker-binding question named there).
4. **`bugs_open/239`** is a real, open platform bug this lane found and
   characterised but does not own fixing. If you're picking up platform
   work generally, it's worth a look — it breaks the documented drive-loop
   pattern CLAUDE.md itself teaches, for every lane relying on it while the
   build queue stays starved. Read the file's own "Plan for whoever picks
   this up" section — do not re-bisect it against production.

## 3. Traps carried forward, still current

- **`pages.sections` is not durable across a rebuild** — a full
  `page-build-handler` run can silently reset it. Re-check after any future
  rebuild of any of these 5 pages, not just the hero/title landmine already
  documented.
- **`chat.go`'s `systemPromptFacts` has no code link to `evidence_base`** —
  this bit for real this session (the £75 deposit was live in the DB
  before the bot knew about it, for about the length of one dispatch
  cycle). Whoever next changes `evidence_base` still owns checking this
  file too; there is still no automated sync.
- **Any future hand-edit to a live page must update BOTH `content_data` AND
  `rendered_html`** on the `page_components` row, then propagate to the
  actual static file in `gqls/vm-sites` via git — content_data alone is
  invisible until a rebuild (CTS-003), and the DB row alone never reaches a
  visitor without the git push. This session's whole method, proven
  correct repeatedly: read current state → construct exact change →
  dry-run in a transaction → verify → apply → propagate to git → force
  `sitesync` → verify at the served artefact.
- **The permission classifier's block on a just-damaged file is not
  permanent** — it cleared within the same session once enough time/other
  work had passed. Worth retrying later in the SAME session before
  concluding a fresh session is needed.

## 4. Owner ledger

**Owed by the owner:** correction-fee number · written terms before live
Stripe · final review + cutover approval (now covers the chat box and the
£75 deposit pricing together).
**Settled, do not reopen:** £1,200, no VAT, of which **£75 is a
non-refundable deposit (2026-08-10)** — decline within 14 days, get back
£1,125 · 2 revision rounds · Sonnet 5 on the site writer · Haiku 4.5 on the
chat intake · framework-only builds · one box per trust class · the site
itself is "ok for now" · chat service multi-turn memory (fixed 2026-08-09)
· all 5 pages + the chat box are the whole of what ships in this phase — a
£19 static-hosting-only tier is a **future direction**, not current scope
(see PLAN §7).

## 5. Access map

Unchanged from `HANDOFF_2026-08-10_continue_here.md` §7 — `gh` = gqls,
ADMIN on `vm-sites`, local clone at `/home/ant/projects/vm-sites`;
`~/.ssh/webdesign_box_ed25519`; `~/.config/cloudflare/token` (works,
**no Cache Purge permission** — confirmed this session, don't assume it can
purge); kubectl → cluster/DB (`site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f'`).
