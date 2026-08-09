-- Source work item for the ONE thing owed by HANDOFF_2026-08-08b §2:
-- index/prose-2 still carries the register the owner struck one line above it.
--
-- Two constraints shaped this prompt, and both are load-bearing:
--
--  1. HANDOFF §3 — the writer sees ONE section and never its siblings, so any
--     instruction is applied by every section that can believe it qualifies.
--     Hence the "IF THIS SECTION IS ... OTHERWISE leave your section alone"
--     frame rather than page-level language.
--
--  2. NOTES 2026-08-08 (evening), arm B — this site's own H voice spec, given a
--     neutral brief, produced "the true cost of borrowing" UNPROMPTED. So voice
--     alone does not remove this phrase; the prompt has to police the claim by
--     name and give the owner's reason for it.
--
-- Deliberately contains NO verbatim replacement copy. Last night's prompt shipped
-- an approved block and three sections each pasted it; a block of copy is portable
-- across sections in a way a set of constraints is not. The framework writes the
-- words (owner ruling 2026-08-04); this states the job and the limits.
--
-- The DB row is the source of truth (HANDOFF §5 — editing site_specs changes
-- nothing). This file exists so the prompt text is reviewable in git.

INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, spec, page_id, status, created_by,
   handler_agent, item_key, priority)
SELECT '0162cde4-633e-45e9-8ca6-87a6b2fe1d26', 'voiceh-rollout', 'content_rewrite', 'medium',
       'index prose-2 strap line - drop the accuracy claim (owner register ruling 2026-08-08)',
       jsonb_build_object(
         'mode', 'edit_live',
         'page', 'index',
         'page_name', 'index',
         'page_id', p.id::text,
         'work_item_type', 'content_rewrite',
         'max_fix_attempts', 1,
         'description', 'Rewrite one strap line into the site voice, dropping an accuracy claim. Every other section of this page is finished and must not change.',
         'acceptance_test', 'The one-sentence introduction under the Standard Loan Calculator heading no longer claims exactness or a "true" cost, does not restate the opening block, and keeps its wrapping <p> and inline style. No other section changed.',
         'suggestion', $SUGGESTION$This is a targeted single-section fix on a finished page. Almost every section of this page is already correct and must not change.

IF THIS SECTION IS the one-sentence introduction sitting directly under the "Standard Loan Calculator" heading - it currently reads "Calculate your exact monthly repayments and see the true total cost of borrowing." - then rewrite that sentence, following the rules below.

OTHERWISE this instruction does not apply to your section. Leave your section's subject, structure and wording exactly as they are, and return it unchanged.

If it does apply, the job is to say what the calculator directly below is for, in the site's ordinary voice, WITHOUT claiming accuracy:

- Do not use "exact", "precise", "mathematically rigorous", "true cost of credit", "true total cost of borrowing", or any wording whose point is that the arithmetic here is unusually accurate or uniquely truthful.
- WHY this matters: a borrower already assumes the arithmetic works, on this site and on everyone else's. Leading on accuracy answers a question nobody asked, and this is the first line a reader meets under the heading. Let the calculator's output carry the authority instead of claiming it in advance.
- The opening block higher up this page already says "Work out what a personal loan will cost you in total, not just each month." Do NOT restate that. This line should do a different job: tell the reader what to do with the calculator below, or what it will show them that a monthly figure on its own does not.
- One sentence. Two at the very most. This is a strap line under a heading, not a paragraph.
- Keep the wrapping <p> element and its inline style attribute exactly as they are. Change only the words inside it.
- British English. Plain, warm, second person. Introduce no new numbers, rates, figures or factual claims of any kind.$SUGGESTION$
       ),
       p.id, 'needs_human_review', 'voiceh-rollout',
       'page-build-handler', 'index-prose2-source', 40
FROM pages p
WHERE p.site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND p.name='index' AND p.status='active'
RETURNING id, item_key, status;
