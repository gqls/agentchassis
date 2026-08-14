# PLAN 2026-08-13 — fleet fix for bugs_open/268 (CTA destination keys lost on regeneration)

Objective, in order (the order is load-bearing — repairing before fixing re-arms
the loss, proven on webdesign.uk 2026-08-12):

1. **Diagnose** — 090 run pointed at `page_component_history`, mechanism stated
   as a question. Gate for everything after it.
2. **Fix** — stop regeneration destroying renderer-sourced keys. Candidate 1
   (route renderer/static fields through the missing-field path so the 238
   carry applies) is favoured going in; final shape follows the 090 verdict and
   the 2026-08-02 owner ruling (new authority on a shared seam = opt-in field,
   unsafe default OFF; per RFC_022 that shape is not architecture-scope while
   no live consumer names it — enumerate the consumers with the query, not by
   assertion).
3. **Council gate** before/alongside the commit (`Council-Submitted:` trailer
   if committing first). Register the seam in the concept register in the same
   commit. Name the consumers: 20 sites; producing agents page-build-handler +
   page-rerender.
4. **Canary** — one site, two pages that disagree (heavy + light), end-to-end
   `content_rewrite` with `mode=edit_live`, invariant diff (href counts) both
   sides.
5. **Repair** — 214 components / 18 unprotected sites, from
   `page_component_history` (or `resolve_internal_links` re-run if candidate 2
   ships). Write `content_data`, then re-render. Never patch `rendered_html`.
   webdesign.uk stays OUT of the sweep (locked, already repaired).
6. **Verify fleet** — census query from `bugs_open/268` §2 → expect
   label-without-URL count to fall to ~0; invariant diff on the canaries.
7. **Unlock webdesign.uk** — final step, only after the fix is live and proven
   (filing lane's RUNBOOK carries the unlock/edit/relock recipe).

## Decisions and corrections

- **2026-08-13: candidate mechanism REFINED against HEAD (code-read, pending
  090 confirmation).** The handoff's §1 OPEN candidate said
  `sourceResolver.resolve` short-circuits renderer fields to `(nil, true)` so
  `handleMissingField` never runs, and cautioned that `carryStored` might skip
  renderer fields too. Reading HEAD (`a3fee59b8`):
  - The bypass is **earlier**: a `renderer`/`static` field takes the early
    branch at `plan_sections_action.go:2362-2369` — writes only a declared
    `fallback` (these fields have none) and `continue`s. It never reaches
    `resolver.resolve` (:2372) at all.
  - Had it reached resolution, `resolve` (:622-626) returns `(nil, true)`, and
    the caller's `found && value != nil` test (:2374) sends nil values to
    `handleMissingField` anyway.
  - The handoff's `carryStored` caution is **inverted**: its guard (:2124)
    returns false only for `""`/`"llm"` sources — a renderer field reaching it
    WOULD be carried. The whole defect is that it never reaches it.
  So the minimal fix shape is: let renderer/static fields fall through to (or
  explicitly invoke) `handleMissingField`/`carryStored` instead of the bare
  `continue` at :2368 — subject to the 090 verdict and to whatever the early
  branch exists to protect (find out WHY it was written before deleting it).
- **2026-08-13: chassis rolled v1.0.1291 → v1.0.1295 (stamp `69612d692`) during
  session 1.** Only one commit in that window touches the 268 code paths:
  `0c8e08ccb` (fix 253, per-slot component floor in save_page_sections). It
  guards markup-structure loss on rewrite, scope-gated to slots with ≥10 class
  attributes — hero/call-to-action components are small, so it likely does NOT
  protect the CTA case [INFERRED — check a real hero row's attribute count
  before repeating]. Adjacent, not a fix and not an obstacle.

## Constraints

- webdesign.uk: 8 CTA-bearing components `lock_type='permanent'` — leave out of
  sweeps; a `lock_blocked_change` from them is correct behaviour.
- Sibling `webdesign_uk_build_service` lane owns the contact page
  chat-input-box lock. Do not touch.
- Go half of any fix is inert until an image rolls; schema/config half is live
  immediately on 20 sites — canary first.
