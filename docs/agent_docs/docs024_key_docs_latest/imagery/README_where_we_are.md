● Everything is persisted (your bulk commit v1.0.1114 swept in my Turn 34 notes and the emit action; the asset-deployer sprite_css mode is live in the DB). Big progress this turn:

The sprite sheet is live and gate-passed. After all three deploy-plumbing fixes, the re-drive produced a clean 768×768 JPEG at 75,745 bytes — under the 80KB budget — serving 200. The image is exactly right (nine flat
glyphs on charcoal, correct order), you confirmed the cell meanings, and I've written cell_names_verified: true into the plan row so the CSS emitter can trust the map. Notably, the phase's headline risk — the model
putting the wrong glyph in the wrong cell — never materialised; every problem was deploy/dispatch plumbing, all now fixed and documented.

I2.2 and I2.3 are built:
- emit_sprite_css — computes the stylesheet purely from the verified grid geometry (no image cropping): a .sprite base plus per-glyph .sprite-<name> classes for inline/icon/nav use, and themed ul.sprite-list bullets
  with a default glyph and per-item li.sprite-b-<name> overrides, all scaled to bullet size. It commits /assets/css/sprites.css, and it's guarded on cell_names_verified=true so it can never emit CSS keyed to an
  unverified (possibly mislabelled) sheet. The geometry math is unit-tested.
- Head link injection — sprites.css is linked into every <head>, guarded on an active sprite sheet existing (no 404 link on sites without one).
- asset-deployer sprite_css mode (migration already live) — mirrors the brand_head pattern, making the whole thing reusable fleet-wide via a needs_sprite_css work item.

One deploy carries the emit action + head-link Go. After it, the finish line for I2 is short: dispatch the needs_sprite_css item → sprites.css commits → a site-component re-render lands the head link → I wire
class="sprite-list" onto one robot-hands section's list → live gate (readable themed bullets, one small download). Then I2.4's fulfilment discovery check closes the phase.

Flag me when you deploy and I'll run that finish sequence.

---

● All recorded, and the docs are consistent. One thing worth knowing: your bulk commit ce0a03cb swept in the specificity fix and all three imagery docs — I verified the scoped override is in HEAD. The new
product handoff is still untracked, so it'll go in your next sweep.

Choices recorded

Both are now formal decisions in the PLAN's register (which stops at D8, so these are D9/D10), with the reasoning captured so a future session doesn't relitigate them:

- D9 — cosmetic: accepted. The charcoal cell backdrop behind each glyph stays. I recorded why it isn't a defect: JPG has no alpha, and that was the deliberate Turn 33 trade after a lossless PNG blew both
  the Kafka commit size limit and the 80KB budget. Flagged for revisit only if a transparent format ever fits the budget.
- D10 — container opt-in: approved, sequenced as I2.5, after the live gate and I2.4, exactly as you asked. emit_sprite_css will also emit .sprite-bullets ul>li::before (same geometry, and specificity-safe
  this time), with the class going on a component wrapper — article-body's .article-body__content first. The Turn 36 hand-wired classes then become redundant but harmless.

The imagery handoff, plan and running notes (Turns 36 + 37) all carry these, plus the deploy step that's still pending: re-dispatch needs_sprite_css to re-emit the corrected CSS.

The product-page handoff

docs/agent_docs/docs024_key_docs_latest/HANDOFF_2026-07-14_empty_product_sections.md

Investigating it properly changed what the fix actually is, so it's worth flagging before you open that chat. It isn't simply "product data is missing" — the platform already detected this and reported it
fixed. There are empty_section work items for exactly these sections, handled by page-build-handler, marked complete on 2026-07-10 — and the sections are still empty today. A fix loop closing without
fixing is the highest-value bug in there, and it's a much bigger deal than the product pages themselves; it overlaps directly with the fixloop workstream's "bug dissolves but isn't fixed" benchmark case.

The other findings behind it:

- Every value field (product_name, product_price, feature_1..4, product_sku) is absent from content_data, though the schema declares them required: true, source: llm. All 49 keys present are chrome and
  boilerplate. (Trap noted: input_schema uses a fields wrapper, not JSON-Schema properties — querying properties returns nothing and misleads you, as it briefly did me.)
- A sibling component stored the LLM's prose apology about not having query.affiliate_products as page content, and its declared on_missing: skip_section never fired.
- The empty shell survives assembly because sectionHasVisibleContent keeps anything over 10 characters of text — and the template is stuffed with static labels like "SKU:" and "Add to Cart". The filter
  measures text, not resolved data.

On your instinct that it's fleet-wide: the mechanism is generic and will recur on any site whose planner places a component needing a data source the site lacks — but today the blast radius is robot-hands
(6 instances). dartsonline's 2 product components render 14 real cards, which is the useful counter-example: the pipeline is sound when a data source exists, so this is the no-data path failing open
instead of failing safe. That's in the handoff so the next chat doesn't start by suspecting the wrong layer.

Worth deciding early in that chat: robot-hands is a spec/comparison site, so Add-to-Cart furniture is category-wrong for it regardless of whether the data gets filled.

----