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

Done — the bullets are correct on the live site, and I2.4 is built and registered.

The bullets are right

The re-emit ran first attempt and the guide page now renders four distinct glyphs: ⓘ info, ✓ check (the unclassed item — proving the default still works), the gauge dial, and the ⚠ warning triangle.
Legible at 20px, one 75,745-byte sheet for all of them.

Your gate: https://robot-hands.com/guides/tool-grip-force-friction-calculator-guide.html → "Safety Factor Selection Guidance". Hard-refresh — sprites.css is cached for an hour. The faint charcoal tile
behind each glyph is the D9 thing you already accepted, not a defect.

Two verification notes worth keeping. My first check of the CSS after the emit returned the old file and briefly looked like a failed deploy — it was CDN cache; a cache-busted fetch showed the new one,
and the canonical URL caught up seconds later. And I confirmed the deploy by grepping the running pod's binary for the scoped selector rather than trusting git, per the standing practice.

I2.4 — the fulfilment check

Building it surfaced a design question worth flagging, because it changed the shape of the work.

Half of I2.4 turned out to need no code: unfulfilled_imagery_plan already emits needs_imagery for any unfulfilled plan row regardless of kind, so "sheet planned but no asset" is covered. The real gap is
"asset exists but sprites.css doesn't".

The catch: discovery checks in this codebase are DB-only by convention (check_image_url_404's header says so explicitly — no HTTP). So a check has no way to know whether the CSS was ever emitted. It
couldn't — so emit_sprite_css now stamps the plan row after committing (style_hints.sprites_css = {emitted_at, sheet_path, signature}), and the check reads that.

That stamp buys more than presence. The signature encodes the grid and cell vocabulary (3x3:check,gauge,…), so if the sheet is regenerated at a new geometry or the cell names are re-verified in a
different order, the committed CSS is stale — slicing the wrong glyphs, which is worse than missing — and the check re-emits. It also re-emits if the sheet asset is newer than the stamp. I put the
signature function in imageryplan rather than duplicating it, since actions imports discovery_checks and can't be imported back — the same anti-drift argument that package already makes for
Classify/ItemKey.

Tests cover missing, stale-by-grid, stale-by-regeneration, bad timestamp, and — the one that matters most — fulfilled-so-don't-re-emit, which is what stops the check re-committing an identical stylesheet
on every discovery pass.

Registration SQL is already applied (backup + verify, 20 checks). It was safe to apply before the Go ships because an unregistered check name is just a warn-and-skip, not a failed pass — so it activates
by itself on your next deploy.

One thing to watch on that deploy, rather than assume: the live CSS was emitted by the pre-stamp binary, so the plan row has no stamp. The check should fire exactly once (reason missing), re-commit
identical CSS, stamp the row, then go silent. If it keeps emitting every pass, the stamp write is broken. That single cycle is the live proof I2.4 works.

Build and all three test packages are green. Incidentally, I noticed check_required_fields_missing.go and complete_work_item_verification.go have appeared in the working tree — looks like the product-page
handoff is already being acted on in your other chat; my build passes alongside them.

Once you've eyeballed the gate, I2.5 (the D10 container opt-in) closes the phase. Worth knowing before I start it: article-body is a global component, so the wrapper class lands on every site that uses it
— harmless, since the class is inert without a sprites.css, but I'll call it out in the change rather than let it surprise anyone.


---------

