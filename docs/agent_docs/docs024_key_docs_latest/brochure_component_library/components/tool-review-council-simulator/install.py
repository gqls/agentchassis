#!/usr/bin/env python3
"""Generate the install SQL for tool-review-council-simulator, as ONE transaction.

Run with --emit to print the SQL, or pipe it straight into psql:
  python3 install.py --emit | kubectl -n ai-persona-system exec -i postgres-clients-0 \
      -- psql -U clients_user -d clients_db

Why a generator rather than a .sql file: the template is 28KB of HTML/CSS/JS and
hand-quoting it into SQL is exactly how a stray quote silently truncates a
component (bugs_open/012). Python does the quoting.

Why one transaction with verifying SELECTs before COMMIT: this touches three
tables (content_components, pages, page_components) and a half-applied placement
is the failure mode that silently drops a section at the next re-render. A bad
count raises, the transaction rolls back, and nothing is left half-done.
"""
import json
import os
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
TEMPLATE = os.path.join(HERE, "template.html")

SITE_ID = "199733a8-ac9c-4c30-b2ce-65ecdac6f3bd"
DOMAIN = "fundamentallyai.com"
FUNCTION = "tool-review-council-simulator"
COMPONENT_NAME = FUNCTION + "-fundamentallyai-com"
PAGE_NAME = "review-council-simulator"
PAGE_URL = "/tools/review-council-simulator.html"

# Sections, in order. Mirrors the llm-cost-calculator page minus tool-guide-intro
# (that component is render_mode=agent and would need an LLM pass to fill; this
# page is deliberately self-contained).
SECTIONS = ["hero-tool", FUNCTION, "tool-cta"]

DESCRIPTION = (
    "Interactive simulator estimating how often a change passes an AI reviewer panel, "
    "given which seats sit on it, how severe an objection must be to block, and how many "
    "revision rounds are run. Calibrated on 362 real council runs (2026-07-10 to 2026-07-30): "
    "26 seats with measured per-seat objection rates at three severity thresholds."
)

META_DESCRIPTION = (
    "An interactive AI Review Council Simulator, free to run in the browser. Set the panel, "
    "the blocking threshold and the number of revision rounds, and see how often a change gets "
    "through. Calibrated on 362 real council runs."
)

# Every URL referenced below was checked against the pages table before being written
# here. There is NO /tools.html index page on this site, so the sibling page's
# "Explore All Tools" label has no valid target and is not copied.
HERO_DATA = {
    "badge_label": "Interactive Tool",
    "hero_headline": "How strict should your AI review panel be?",
    "hero_subheadline": (
        "Put a panel of AI reviewers in front of every change and some proportion of your work "
        "comes back. This tool estimates how much, using the measured objection rates of the 26 "
        "reviewer seats we run against our own platform. Move the threshold, the panel and the "
        "number of rounds, and watch the pass rate move with them."
    ),
    "stat_one_value": "362",
    "stat_one_label": "council runs measured",
    "stat_two_value": "26",
    "stat_two_label": "reviewer seats",
    "stat_three_value": "51%",
    "stat_three_label": "passed after our July fix",
    "cta_primary_label": "How our review council works",
    "cta_primary_url": "/multi-agent-review-council.html",
    "cta_secondary_label": "Talk to us about review",
    "cta_secondary_url": "/contact.html",
}

CTA_DATA = {
    "eyebrow_label": "Free Tools",
    "headline": "Size the cost of reviewing everything.",
    "description": (
        "This simulator runs on the real objection rates of the reviewer seats we put in front of "
        "changes to our own platform, measured over 362 council runs in July 2026. We built it to "
        "settle how strict our own gate should be. Treat the pass rates as a floor rather than a "
        "forecast: the model assumes reviewers act independently, and they do not."
    ),
    "trust_note": "No account required. Open access.",
    "tools_list_label": "Our other tools",
    "empty_state_text": "More tools are currently in development.",
    "primary_cta_label": "How our review council works",
    "primary_cta_url": "/multi-agent-review-council.html",
    "secondary_cta_label": "Talk to us",
    "secondary_cta_url": "/contact.html",
    "items": [
        {
            "url": "/tools/llm-cost-calculator.html",
            "name": "llm-cost-calculator",
            "image": "",
            "title": "LLM Provider Cost Comparison Calculator",
            "nav_label": "Tools / LLM Provider Cost Comparison Calculator",
            "meta_description": (
                "Compare monthly and annual API costs across language model providers for your "
                "projected token volume, with a self-hosting break-even."
            ),
        },
        {
            "url": "/tools/model-approach-selector/index.html",
            "name": "tool-model-approach-selector",
            "image": "",
            "title": "Fine-Tuning vs RAG vs Prompting Decision Guide",
            "nav_label": "Tools / Fine-Tuning vs RAG vs Prompting Decision Guide",
            "meta_description": (
                "Work out which of fine-tuning, retrieval or prompting fits the job you have, and "
                "why the cheap option is usually the right first move."
            ),
        },
    ],
}


def q(s):
    """Single-quote a SQL string literal."""
    return "'" + s.replace("'", "''") + "'"


def dollar(s, tag="TPL"):
    """Dollar-quote a body. Verifies the tag cannot appear inside it."""
    delim = "$%s$" % tag
    if delim in s:
        raise SystemExit("dollar-quote tag %s appears in the body; pick another" % delim)
    return delim + s + delim


def main():
    with open(TEMPLATE) as fh:
        tpl = fh.read()

    if "{{" in tpl:
        raise SystemExit("template contains Go-template delimiters; the renderer would eat them")

    sql = []
    a = sql.append

    a("\\set ON_ERROR_STOP on")
    a("BEGIN;")
    a("")

    a("-- 1. the widget component. One row, site-scoped name, shared function,")
    a("--    matching the sibling tool's metadata (category=interactive, level=tool).")
    a("""INSERT INTO content_components
  (name, function, display_name, description, html_template, render_mode,
   category, component_level, is_active, is_dark_section, created_from)
VALUES (%s, %s, %s, %s,
        %s,
        'template', 'interactive', 'tool', true, false, 'manual');""" % (
        q(COMPONENT_NAME), q(FUNCTION), q("AI Review Council Simulator"),
        q(DESCRIPTION), dollar(tpl)))
    a("")

    a("-- 2. the page. Mirrors the llm-cost-calculator page row; nav_order 202 follows")
    a("--    200 (approach selector) and 201 (cost calculator).")
    a("""INSERT INTO pages
  (site_id, name, url, title, page_type, status, meta_description,
   nav_label, nav_order, in_header, in_footer, sections, build_status, rebuild_policy)
VALUES (%s, %s, %s, %s, 'tool', 'active', %s,
        %s, 202, true, false, %s::jsonb, 'pending', 'generic');""" % (
        q(SITE_ID), q(PAGE_NAME), q(PAGE_URL),
        q("AI Review Council Simulator | Tools"), q(META_DESCRIPTION),
        q("Tools / AI Review Council Simulator"), q(json.dumps(SECTIONS))))
    a("")

    a("-- 3. the three placements. slot_name is set on every row: a NULL slot finds no")
    a("--    component at re-render, the action carries the row's (empty) stored HTML,")
    a("--    and the section silently vanishes from the assembled page.")
    a("--    hero-tool and tool-cta reuse the exact component rows the sibling tool page")
    a("--    uses, so there is no chance of picking a different fork.")
    for i, slot in enumerate(SECTIONS, start=1):
        if slot == FUNCTION:
            comp = "(SELECT id FROM content_components WHERE name = %s)" % q(COMPONENT_NAME)
            data = "{}"
        else:
            comp = ("(SELECT pc.component_id FROM pages p"
                    " JOIN page_components pc ON pc.page_id = p.id"
                    " JOIN content_components cc ON cc.id = pc.component_id"
                    " WHERE p.site_id = %s AND p.url = '/tools/llm-cost-calculator.html'"
                    " AND cc.function = %s)" % (q(SITE_ID), q(slot)))
            data = json.dumps(HERO_DATA if slot == "hero-tool" else CTA_DATA)
        a("""INSERT INTO page_components (page_id, component_id, position, slot_name, content_data, build_status)
VALUES ((SELECT id FROM pages WHERE site_id = %s AND name = %s),
        %s,
        %d, %s, %s::jsonb, 'pending');""" % (
            q(SITE_ID), q(PAGE_NAME), comp, i, q(slot), dollar(data, "J")))
        a("")

    a("-- 4. VERIFY BEFORE COMMIT. Each check raises on a wrong count, which rolls the")
    a("--    whole thing back rather than leaving a half-placed page behind.")
    a("""DO $CHK$
DECLARE
  n_comp int; n_page int; n_pc int; n_null_slot int; n_null_comp int; n_null_data int;
  tpl_len int; n_sections int;
BEGIN
  SELECT count(*), max(length(html_template)) INTO n_comp, tpl_len
    FROM content_components WHERE name = %s;
  IF n_comp <> 1 THEN RAISE EXCEPTION 'expected 1 component row, got %%', n_comp; END IF;
  IF tpl_len < 20000 THEN
    RAISE EXCEPTION 'template stored as only %% chars - truncated', tpl_len;
  END IF;

  SELECT count(*) INTO n_page FROM pages WHERE site_id = %s AND name = %s;
  IF n_page <> 1 THEN RAISE EXCEPTION 'expected 1 page row, got %%', n_page; END IF;

  SELECT jsonb_array_length(sections) INTO n_sections
    FROM pages WHERE site_id = %s AND name = %s;
  IF n_sections <> 3 THEN RAISE EXCEPTION 'expected 3 sections, got %%', n_sections; END IF;

  SELECT count(*),
         count(*) FILTER (WHERE slot_name IS NULL),
         count(*) FILTER (WHERE component_id IS NULL),
         count(*) FILTER (WHERE content_data IS NULL)
    INTO n_pc, n_null_slot, n_null_comp, n_null_data
    FROM page_components
   WHERE page_id = (SELECT id FROM pages WHERE site_id = %s AND name = %s);
  IF n_pc <> 3 THEN RAISE EXCEPTION 'expected 3 page_components, got %%', n_pc; END IF;
  IF n_null_slot <> 0 THEN RAISE EXCEPTION '%% rows have a NULL slot_name', n_null_slot; END IF;
  IF n_null_comp <> 0 THEN
    RAISE EXCEPTION '%% rows failed to resolve a component_id', n_null_comp;
  END IF;
  IF n_null_data <> 0 THEN
    RAISE EXCEPTION '%% rows have NULL content_data (would escalate to the content writer)',
      n_null_data;
  END IF;

  RAISE NOTICE 'verified: 1 component (%% chars), 1 page, 3 placements, no NULLs', tpl_len;
END
$CHK$;""" % (q(COMPONENT_NAME), q(SITE_ID), q(PAGE_NAME), q(SITE_ID), q(PAGE_NAME),
             q(SITE_ID), q(PAGE_NAME)))
    a("")
    a("COMMIT;")
    a("")
    a("-- post-commit read-back, so the run's own output shows what landed")
    a("""SELECT p.url, pc.position, pc.slot_name, cc.function,
       length(cc.html_template) AS tpl_len, pc.build_status
  FROM pages p
  JOIN page_components pc ON pc.page_id = p.id
  JOIN content_components cc ON cc.id = pc.component_id
 WHERE p.site_id = %s AND p.name = %s
 ORDER BY pc.position;""" % (q(SITE_ID), q(PAGE_NAME)))

    print("\n".join(sql))


if __name__ == "__main__":
    if "--emit" not in sys.argv:
        raise SystemExit(__doc__)
    main()
