So the sequence during the recent rebuild was:

call_site_planner — planner produced a site plan with a list of pages
sync_pages_to_db — creates/updates page records from that plan. Pages in the plan get build_status = 'needs_rebuild'. Pages NOT in the plan are left untouched.
get_pages_to_build — queries for planned or needs_rebuild only
build_pages_loop — builds only those pages

If the site planner didn't include "use-cases" in the new plan, it was never set to needs_rebuild. It sits in the pages table with build_status = 'deployed' and in_header = true from the old build — so it shows in nav but has stale content.

