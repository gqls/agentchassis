1. Human says: "the redesign made it worse, go back"

2. Rollback the design_intent aspect:
   UPDATE site_specs SET is_current = false, superseded_at = NOW()
   WHERE site_id = $1 AND aspect = 'design_intent' AND is_current = true;

   UPDATE site_specs SET is_current = true, superseded_at = NULL
   WHERE id = '<previous_version_id>';

3. Create work items to rebuild from the restored spec:
   INSERT INTO site_work_items (site_id, item_type, handler_agent, ...)
   VALUES ($1, 'needs_design_review', 'webdesign-agent', ...);

4. Dispatch loop picks it up. Webdesign-agent reads the now-restored
   design_intent from site_specs and generates CSS accordingly.

5. Rerender deploys the pages with the restored design.
   For content rollback: same pattern. Restore the content_direction aspect, create content_rewrite work items for affected pages, dispatch rebuilds them.
   For a full site rollback (everything): restore all aspects to a point in time, create work items for every page. This is heavier but follows the same pattern. Git history provides the additional safety net — the deployed HTML is in git with commit-per-work-item, so you can also revert at the git level for immediate effect while the spec-driven rebuild catches up.
   

# On the business context:
   The framework's value is the pipeline itself — the ability to create, maintain, and improve sites autonomously. The sites are the output, but the pipeline is the product. This means:

Sites need to be genuinely good (not AI-slop) because they demonstrate what the framework can do
The spec-driven approach matters for selling/licensing because it's the interface between human intent and automated execution — a buyer or licensee adjusts the spec, the system does the rest
Portability (export to WordPress/Laravel) is a future feature of the deployment layer, not the spec layer — the spec describes what the site is, the deployer puts it somewhere
The agent registry + feasibility checking is the mechanism for "cutdown" versions — a licensed instance with fewer agents still works, it just has more items in blocked status