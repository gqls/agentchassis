#!/usr/bin/env bash
# dispatch_228_rerenders.sh — propagate the contact-block content_components
# change to the two live pages that carry it, and verify against the
# DEPLOYED page HTML with deploy-window discipline (not a naive curl right
# after dispatch).
#
# Round-2 council objection this answers (render_guardian, HIGH,
# correlation 46f87e4c-05fc-4a5c-bd6a-93a073b63253): "Editing
# content_components.html_template does not touch page_components.
# rendered_html... Without an explicit scoped rerender + deployed-page
# verification step, this closes the DB row but can leave the live pages
# serving the old fabricated-success markup -- the bugs_open/024 false-green
# class."
#
# Round-3 council objection this answers (prior_art_librarian, HIGH gating,
# same correlation): the previous version of this script printed curl
# commands to run immediately after dispatch, landing on two documented
# LANDMINES.md traps: "Verifying a page straight after firing its rerender
# shows a 404 or stale copy, and both look like you broke the site" (poll
# pages.deployed_at, cache-bust, assert a REMOVED string + an ADDED string +
# an untouched CONTROL) and "`curl | grep` twice against the same URL during
# a deploy reports a regression that never happened" (fetch ONCE to a file,
# grep the file for every property, never re-curl per assertion). Both are
# fixed below.
#
# PRECONDITION: apply_228_contact_block_fix.sh has been run AND its apply SQL
# has been COMMITted (not just generated). Verify:
#   SELECT html_template LIKE '%action="{{.form_action}}"%' FROM content_components WHERE function='contact-block';
# must be true before running this.
#
# Mechanism: direct page-rerender orchestrate, bypassing any queue lag
# (see memory: single-page-deploy-bypasses-stalled-queue), with reason=
# section_data_resolved so the section re-renders from content_data through
# the CURRENT html_template (picking up the template change) rather than
# reassembling the stale stored rendered_html. Both target pages were
# confirmed to have non-NULL content_data on their contact-block section
# (2026-08-09 query), so this does NOT escalate to full LLM content
# regeneration (the 049b script's own documented gotcha).

set -euo pipefail

PSQL() {
  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
    psql -U clients_user -d clients_db "$@"
}

DISPATCH_ONE() {
  local page_id="$1" site_id="$2" domain="$3"
  docs/agent_docs/docs024_key_docs_latest/cta_link_integrity/scripts/049b_deploy_single_page.sh \
    "$page_id" "$site_id" "$domain" section_data_resolved
}

echo "== robot-hands.com/contact =="
DISPATCH_ONE 9880a8f3-72df-4c8f-b036-5912dfb10805 00ff3af5-dad8-4770-9f70-3edc267a3c92 robot-hands.com

echo "== leopardessconsulting.co.uk/ai-readiness-quiz =="
DISPATCH_ONE ecfd0bfd-bc5c-4ed4-9c45-7ba9143e72c8 4851f6fc-71cf-4160-a270-e03d6d3e0732 leopardessconsulting.co.uk

echo ""
echo "Each dispatch printed CORR=<uuid> above. The first orchestration_states"
echo "row can take several minutes to appear (measured 7-9 min elsewhere on"
echo "this platform) -- that is NOT evidence of a dropped dispatch; do not"
echo "re-fire on that basis alone."
echo ""
echo "Once you believe the dispatch has had time to land, run (each polls"
echo "pages.deployed_at / cache-busts / re-greps ONE fetch per iteration):"
echo ""
cat <<'VERIFY'
./verify_228_deployed_page.sh \
  'https://robot-hands.com/contact.html' \
  'action="mailto:robot-hands@contactforsales\.com' \
  'x-never-matches-x' \
  'cb-form|contact-block-section' \
  robot-hands.com contact

./verify_228_deployed_page.sh \
  'https://robot-hands.com/tools/assets/contact-block.js' \
  'Opening your email client' \
  'setTimeout|has been sent' \
  'cb-contact-form|validateEmail'

./verify_228_deployed_page.sh \
  'https://leopardessconsulting.co.uk/ai-readiness-quiz.html' \
  'action="mailto:leopardess@contactforsales\.com' \
  'x-never-matches-x' \
  'cb-form|contact-block-section' \
  leopardessconsulting.co.uk ai-readiness-quiz

./verify_228_deployed_page.sh \
  'https://leopardessconsulting.co.uk/tools/assets/contact-block.js' \
  'Opening your email client' \
  'setTimeout|has been sent' \
  'cb-contact-form|validateEmail'
VERIFY
