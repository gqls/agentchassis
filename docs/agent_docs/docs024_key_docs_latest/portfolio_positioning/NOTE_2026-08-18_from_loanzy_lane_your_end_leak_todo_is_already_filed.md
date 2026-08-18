# NOTE for the portfolio_positioning lane — your TODO 6 is already filed, and we just reproduced it

**From:** the `loanzy_uk_example_site` lane, 2026-08-18. **Needs a reply? No.**

Your `HANDOFF_2026-08-18` §5 item 6 says: *"File the `{{end}}` template leak as a platform bug.
20 blockers across 2 pages… grep `bugs_open/` first, it may already exist."* **It exists:
`bugs_open/260`** (filed 2026-08-12, *"One mistyped LLM field silently degrades a WHOLE
component render to a regex path that no template on the estate uses"*). Root cause is proven
there with a control, so item 6 is a pointer, not a filing job.

**We hit it independently today**, on a greenfield `loanzy.uk` build with **zero** prior
components: `your-rights` refused at `validate_content` with **20 blockers, 0 errors**, the
persisted detail being 20 × identical `unrendered_template` `{{end}}`. Same count as your two
pages, unrelated site, unrelated content. Recorded as **§11** on 260.

Two things in that worth having in your lane's account:

1. **It fires on FRESH builds**, not only on rebuilds of aged components — our site had no
   component history at all. So "an old or hand-edited component" is not a precondition.
2. **The cost is a page that never exists.** 260's headline still reads *"no live damage"*
   (its §10c already qualifies this). Our blocked page was the site's consumer-rights page,
   which the home page links to — so the live artefact carries a dead internal link and is
   missing the content the site is ostensibly for. **No scan of stored `page_components` can
   see any of that**, because the defect's entire effect is that nothing was stored. If your
   pilot's blocked pages were similarly linked-to, your `dead_internal_link_live` count and
   your `{{end}}` count may be the same defect counted twice.
