Build path now:
... → set_default_components → render_site_components → get_pages_to_build → build_pages_loop
(loop: write → review → assemble_page [inject_head] → git_commit → save_sections → update_status)

Rerender path now:
rerender-site: ensure_site_record → spawn agents → render_site_components [force] → get_pages → loop(call page-rerender) → trigger_deploy
page-rerender: render_page → check_skipped → deploy_page → update_status