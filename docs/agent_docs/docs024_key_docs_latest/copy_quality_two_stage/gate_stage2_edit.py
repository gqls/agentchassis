#!/usr/bin/env python3
"""gate_stage2_edit.py — grade ONE proposed stage-2 copy edit before it is applied.

WHY THIS EXISTS. Stage 2 rewrites live prose. Every guarantee it makes — same facts,
same links, same markup, same declared types — is stated in PROSE today, in the
component's own `llm_guidance` ("Preserve every factual claim, figure, and internal
link present in the existing content"). That instruction is already live on the proof
case, and the proof case is a page that LOST SIX LINKS anyway. A prose instruction to
preserve a SET is not reliably followed; this lane has now measured that four times.

So the contract is asserted here, at the boundary, against the component's OWN prior
content and its OWN declared schema — never against a brief, and never by comparing
prose to prose (`bugs_open/278` §8: same generator, same inputs, twice → 2 of 4 card
bodies diverged with nothing wrong; a prose-diff gate would fail that pair).

WHAT IT CHECKS, all five mechanical (arrays included — their items are flattened
to text so B–E apply to them too):

  A. TYPES     every field named in the proposal matches the type the component
               DECLARES (`content_components.input_schema`), read in BOTH dialects —
               house `{"fields":{…}}` and legacy `{"properties":{…}}`. `bugs_open/260`
               is what an unchecked retype does: a `range` over a string kills the
               whole component and every correctly-written field beside it.
  B. LINKS     no href present before may disappear, and every link the PAGE declares
               in `content_direction.required_links` must be present after.
  C. MARKUP    no class attribute may disappear, and no structural element count may
               fall (`bugs_open/253` is a markup-level loss a text-level check missed).
  D. FACTS     every number/currency figure present before must survive, and no NEW
               figure may appear (stage 2 may reorder and rewrite; it may not invent).
  E. VOLUME    a field may shrink ONLY as de-duplication: if it loses >10% of its
               words, every figure and link removed must still be reachable elsewhere
               on the page, or it fails. A cut below 25% of the original fails outright.
               (It began life as a flat 90% floor and could not tell a gutted section
               from a deliberately de-duplicated one — which is half of stage 2's remit.
               Fixed by making it discriminate, not by relaxing it.)

WHAT IT DELIBERATELY DOES NOT CHECK. Prose quality — that is the human's call under
D2, and the reason stage 2's output queues for review rather than applying itself. It
also does not forbid ADDED links or ADDED classes: adding the six missing links is the
proof case's whole point.

COVERAGE IS REPORTED, NOT ASSUMED (PLAN §10 addendum). A field the component does not
declare cannot be type-checked; a legacy-dialect component carries no `maxItems`/
`enum`/`pattern` through the projection. Both are printed as coverage lines, because a
gate that reports "clean" while silently declining to evaluate half a declaration is
the armed-but-inert shape this lane has already hit twice.

Run:
  gate_stage2_edit.py --component <page_component_id> --proposal <field_updates.json>
  gate_stage2_edit.py --component <id> --proposal <f.json> --self-test   # MUST fail
  gate_stage2_edit.py --item <work_item_id>      # read the proposal from a review item
"""
import argparse
import json
import re
import subprocess
import sys
from collections import Counter

PSQL = ["kubectl", "-n", "ai-persona-system", "exec", "-i", "postgres-clients-0", "--",
        "psql", "-U", "clients_user", "-d", "clients_db"]

# Structural elements whose count must not fall. Chosen because each one is a thing a
# reader loses if it goes: a card, a list item, a heading, a tool grid.
STRUCTURAL = ["section", "div", "ul", "ol", "li", "h1", "h2", "h3", "h4", "p", "a", "table", "tr"]

# A figure is a currency amount, a percentage, or a bare number of 2+ digits. Single
# digits are excluded deliberately: "one of three ways" is prose, not a fact, and
# including them made the check fire on every rewrite that changed a sentence.
FIGURE_RE = re.compile(r'£\s?[\d,]+(?:\.\d+)?|\d+(?:\.\d+)?\s?%|\b\d{2,}(?:,\d{3})*(?:\.\d+)?\b')


def psql(sql, want_json=False):
    r = subprocess.run(PSQL + ["-t", "-A", "-c", sql], capture_output=True, text=True, timeout=180)
    if r.returncode != 0:
        sys.exit(f"psql failed:\n{r.stderr.strip()}")
    out = r.stdout.strip()
    return json.loads(out) if (want_json and out) else out


def declared_fields(schema):
    """Normalise BOTH dialects onto {name: {type, ...}}, mirroring
    datahelpers.SchemaContentFields. Returns (fields, from_legacy).

    ⚠ Written against both on purpose. A gate written against `properties` alone sees
    4 of 191 components; one written against `fields` alone is blind to those 4 — and
    those 4 include `mechanism-flow`, the only component with a proven live failure
    (PLAN §10). The legacy dialect was declared extinct on 2026-07-21 and four
    components have been created in it SINCE, most recently 2026-08-10.
    """
    if not schema:
        return {}, False
    if isinstance(schema.get("fields"), dict):
        return schema["fields"], False
    if isinstance(schema.get("properties"), dict):
        out = {}
        for name, spec in schema["properties"].items():
            spec = dict(spec or {})
            t = spec.get("type")
            # SchemaContentFields' own special cases: string→text, minItems→min_items
            if t == "string":
                spec["type"] = "text"
            if "minItems" in spec:
                spec["min_items"] = spec["minItems"]
            out[name] = spec
        return out, True
    return {}, False


def json_type_of(v):
    if isinstance(v, list):
        return "array"
    if isinstance(v, dict):
        return "object"
    if isinstance(v, bool):
        return "boolean"
    if isinstance(v, (int, float)):
        return "number"
    return "text"


# What a declared type ACCEPTS. `html`/`text`/`string`/`markdown` are all prose fields
# carrying a JSON string; `array`/`list` are the pair that matters for 260 — an
# `array`-only filter never sees `list`, which is a fifth declared type (PLAN §10).
ACCEPTS = {
    "html": {"text"}, "text": {"text"}, "string": {"text"}, "markdown": {"text"},
    "url": {"text"}, "image": {"text"}, "number": {"number"}, "integer": {"number"},
    "boolean": {"boolean"}, "array": {"array"}, "list": {"array"}, "object": {"object"},
}


def flatten(v):
    """Field value -> comparable text. An array field is not exempt from the content
    checks just because it is not a string: `features` is a list of objects whose
    bodies carry the links, figures and classes this gate exists to protect, and the
    first version of this function skipped them entirely while reporting the field
    'type-checked' — true, and misleading, which is the shape this lane calls
    armed-but-inert.
    """
    if isinstance(v, str):
        return v
    if isinstance(v, list):
        return " ".join(flatten(x) for x in v)
    if isinstance(v, dict):
        return " ".join(flatten(x) for x in v.values())
    return "" if v is None else str(v)


def items_of(v):
    return len(v) if isinstance(v, list) else None


def hrefs(html):
    return Counter(re.findall(r'href="([^"]*)"', html or ""))


def classes(html):
    out = Counter()
    for attr in re.findall(r'class="([^"]*)"', html or ""):
        for c in attr.split():
            out[c] += 1
    return out


def elements(html):
    out = Counter()
    for tag in STRUCTURAL:
        out[tag] = len(re.findall(r'<%s[\s>]' % tag, html or "", re.I))
    return out


def figures(html):
    text = re.sub(r'<[^>]+>', ' ', html or "")
    return Counter(re.sub(r'\s', '', f) for f in FIGURE_RE.findall(text))


def words(html):
    return len(re.sub(r'<[^>]+>', ' ', html or "").split())


def load_page_text(page_id, exclude_component_id):
    """Every OTHER component's stored text on the same page, flattened.

    This is what makes a deliberate CUT distinguishable from a silent gutting: text
    removed because it was restated elsewhere is still findable elsewhere. Without
    this the gate can only measure that a field got shorter, which is the same reading
    for the edit stage 2 exists to make and for the bug it exists to prevent.
    """
    out = psql(f"""SELECT COALESCE(string_agg(content_data::text, ' '), '')
                     FROM page_components
                    WHERE page_id = '{page_id}' AND id <> '{exclude_component_id}';""")
    return out or ""


def load_component(component_id):
    row = psql(f"""
        SELECT json_build_object(
          'page_id', pc.page_id::text,
          'slot_name', pc.slot_name,
          'locked_at', pc.locked_at,
          'content_data', pc.content_data,
          'schema', cc.input_schema,
          'component', cc.name,
          'required_links', COALESCE(p.content_direction->'required_links', '[]'::jsonb),
          'page_name', p.name,
          'domain', s.domain)
        FROM page_components pc
        JOIN pages p ON p.id = pc.page_id
        JOIN sites s ON s.id = p.site_id
        LEFT JOIN content_components cc ON cc.id = pc.component_id
        WHERE pc.id = '{component_id}';""", want_json=True)
    if not row:
        sys.exit(f"no page_component {component_id}")
    return row


def report(ok, label, detail=""):
    print(f"{'ok  ' if ok else 'FAIL'} {label}{(': ' + detail) if detail else ''}")
    return 0 if ok else 1


def grade(component_id, proposal, induce=None):
    c = load_component(component_id)
    before_data = c["content_data"] or {}
    fields, from_legacy = declared_fields(c["schema"])
    failures = 0

    if induce == "types":
        # `bugs_open/260` in one line: the writer answers a declared array with prose,
        # or — as here — answers a declared prose field with a list. Either way the
        # renderer's `range` meets the wrong thing and the whole component dies.
        proposal = {k: ([{"body": v}] if isinstance(v, str) else v) for k, v in proposal.items()}

    print(f"subject: {c['domain']} /{c['page_name']} · {c['slot_name']} · "
          f"component {c['component']}{' · LEGACY DIALECT' if from_legacy else ''}")
    if c["locked_at"]:
        print(f"FAIL lock: component is LOCKED ({c['locked_at']}) — stage 2 must not select it")
        return 1
    print()

    # ── A. TYPES ────────────────────────────────────────────────────────────────
    unchecked = []
    for name, value in proposal.items():
        decl = fields.get(name)
        if not decl:
            unchecked.append(name)
            continue
        want = str(decl.get("type", "")).lower()
        got = json_type_of(value)
        accepted = ACCEPTS.get(want)
        if accepted is None:
            unchecked.append(f"{name} (undeclared type {want!r})")
            continue
        failures += report(got in accepted, f"type {name}",
                           f"declared {want}, proposal sent {got}" if got not in accepted
                           else f"declared {want}, got {got}")
    print(f"     coverage: {len(proposal) - len(unchecked)} of {len(proposal)} proposed "
          f"field(s) type-checked" + (f"; NOT checked: {', '.join(unchecked)}" if unchecked else ""))
    if from_legacy:
        print("     coverage: legacy dialect — maxItems/enum/pattern are NOT carried by the "
              "projection and were NOT evaluated (PLAN §10 addendum)")

    # ── B–E, per field, before vs after. Arrays included: flattened to text so the
    # link/markup/fact checks apply to their item bodies too. ────────────────────
    page_text = None
    for name, after_raw in proposal.items():
        before_raw = before_data.get(name)
        before, after = flatten(before_raw), flatten(after_raw)
        b_items, a_items = items_of(before_raw), items_of(after_raw)
        if b_items is not None or a_items is not None:
            print(f"     items {name}: {b_items} -> {a_items}"
                  + ("  (ITEMS REMOVED — the removal test below decides)" if (a_items or 0) < (b_items or 0) else ""))
        if induce and not isinstance(after, str):
            after = flatten(after_raw)
        if induce == "links":
            after = re.sub(r'<li><a href="[^"]*"[^>]*>.*?</a></li>', '', after, count=1)
        if induce == "markup":
            after = after.replace('class="tool-grid"', '', 1)
        if induce == "facts":
            after = after + "<p>Rates from 4.5% on loans up to £25,000.</p>"
        if induce == "volume":
            after = " ".join(after.split()[:20])

        b_href, a_href = hrefs(before), hrefs(after)
        dropped = sorted(set(b_href) - set(a_href))
        failures += report(not dropped, f"links {name} (no drops)",
                           f"{len(dropped)} link(s) dropped: {' '.join(dropped[:6])}" if dropped
                           else f"{len(b_href)} distinct href(s) preserved")

        required = [str(x).split()[0] for x in (c["required_links"] or [])]
        missing = [u for u in required if f'href="{u}"' not in after]
        failures += report(not missing, f"links {name} (page's declared set)",
                           f"{len(missing)} of {len(required)} required absent: {' '.join(missing[:6])}"
                           if missing else f"all {len(required)} declared links present")

        b_cls, a_cls = classes(before), classes(after)
        lost = sorted(k for k in b_cls if a_cls[k] < b_cls[k])
        failures += report(not lost, f"markup {name} (classes)",
                           f"{len(lost)} class attr(s) lost or reduced: {' '.join(lost[:6])}"
                           if lost else f"{sum(b_cls.values())} class attr(s), none lost")

        b_el, a_el = elements(before), elements(after)
        shrunk = sorted(f"{k} {b_el[k]}→{a_el[k]}" for k in b_el if a_el[k] < b_el[k])
        failures += report(not shrunk, f"markup {name} (structure)",
                           f"element count fell: {', '.join(shrunk)}" if shrunk
                           else "no structural element count fell")

        b_fig, a_fig = figures(before), figures(after)
        gone = sorted(set(b_fig) - set(a_fig))
        new = sorted(set(a_fig) - set(b_fig))
        failures += report(not gone and not new, f"facts {name}",
                           (f"{len(gone)} figure(s) lost: {' '.join(gone[:5])}; " if gone else "")
                           + (f"{len(new)} NEW figure(s): {' '.join(new[:5])}" if new else "")
                           or f"{len(b_fig)} figure(s), unchanged")

        # ── VOLUME, and why it is not a word-count floor ─────────────────────
        # The floor this started as (90% of before) could not tell a section that had
        # been GUTTED from one that had been deliberately DE-DUPLICATED — and cutting
        # restated copy is half of stage 2's remit, so the check failed the editor for
        # doing its job. It is not weakened here; it is made to discriminate.
        #
        # The mechanical discriminator is REDUNDANCY, not the rationale (a prose
        # justification would let the agent talk its way past its own gate): every
        # figure and link in the removed text must still be reachable — elsewhere in
        # this field, or in another component ON THE SAME PAGE. Content that survives
        # somewhere the reader can still reach it is a cut; content that vanishes from
        # the page is a loss, and stays a failure.
        b_w, a_w = words(before), words(after)
        if a_w >= int(b_w * 0.9):
            failures += report(True, f"volume {name}", f"{a_w} words (was {b_w})")
        else:
            if page_text is None:
                page_text = load_page_text(c["page_id"], component_id)
            removed_fig = sorted(set(figures(before)) - set(figures(after)))
            removed_href = sorted(set(hrefs(before)) - set(hrefs(after)))
            orphaned = [f for f in removed_fig if f not in re.sub(r"\s", "", page_text)] \
                     + [h for h in removed_href if f'href=\\"{h}\\"' not in page_text and h not in page_text]
            pct = round((1 - a_w / b_w) * 100) if b_w else 0
            failures += report(not orphaned, f"volume {name} (cut is redundancy?)",
                               (f"{pct}% shorter and {len(orphaned)} item(s) exist NOWHERE else on the page: "
                                + " ".join(map(str, orphaned[:5])))
                               if orphaned else
                               (f"{a_w} words, {pct}% shorter (was {b_w}) — every removed figure/link is "
                                f"still on the page, so this reads as de-duplication. ⚠ REVIEW THE PROSE"))
            # A catastrophic cut stays a failure whatever the redundancy test says: at
            # some point "de-duplicated" stops being a credible account of the edit.
            if b_w and a_w < b_w * 0.25:
                failures += report(False, f"volume {name} (absolute floor)",
                                   f"{a_w} of {b_w} words — more than three quarters removed")

    print(f"\n{'PASS' if failures == 0 else 'FAIL'} — {failures} check(s) failing.")
    return 1 if failures else 0


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--component", help="page_component_id the proposal edits")
    ap.add_argument("--proposal", help="JSON file holding the field_updates object")
    ap.add_argument("--item", help="read component+proposal from a needs_human_review item")
    ap.add_argument("--self-test", action="store_true",
                    help="run every induced control; each MUST fail or the gate is inert")
    a = ap.parse_args()

    if a.item:
        # A stage-2 proposal carries review_data.edits[], one entry per component —
        # page-scoped read, section-scoped write, so N edits are normal. Each is graded
        # against its OWN component's live row; the exit code is the worst of them.
        row = psql(f"""SELECT spec->'review_data'->'edits'
                         FROM site_work_items WHERE id = '{a.item}';""", want_json=True)
        if not row:
            sys.exit(f"item {a.item} carries no review_data.edits — not a stage-2 proposal")
        rc = 0
        for i, edit in enumerate(row):
            cid, fu = edit.get("page_component_id"), edit.get("field_updates")
            if not cid or not isinstance(fu, dict):
                print(f"FAIL edit {i}: missing page_component_id or field_updates")
                rc = 1
                continue
            if i:
                print()
            rc |= grade(cid, fu)
        return rc
    else:
        if not (a.component and a.proposal):
            sys.exit("need --component and --proposal, or --item")
        component_id = a.component
        proposal = json.load(open(a.proposal))

    if not isinstance(proposal, dict):
        sys.exit("proposal must be a JSON object of field_updates")

    if a.self_test:
        # DIALECT CONTROL, first and separately: the type check is worthless if the
        # normaliser only reads one dialect. Assert the SAME field, declared each way,
        # lands on the same type — the house `fields` dialect and the legacy
        # `properties` dialect that was declared extinct on 2026-07-21 and has had four
        # components created in it since (PLAN §10).
        house, house_legacy = declared_fields({"fields": {"steps": {"type": "array"}}})
        legacy, legacy_legacy = declared_fields({"properties": {"steps": {"type": "array"}}})
        if house.get("steps", {}).get("type") != "array" or legacy.get("steps", {}).get("type") != "array":
            sys.exit("CONTROL FAILED: the dialect normaliser does not read both dialects.")
        if house_legacy or not legacy_legacy:
            sys.exit("CONTROL FAILED: fromLegacy is not reported correctly per dialect.")
        strings, _ = declared_fields({"properties": {"body": {"type": "string"}}})
        if strings.get("body", {}).get("type") != "text":
            sys.exit("CONTROL FAILED: the legacy string→text remap is not applied.")
        print("CONTROL OK: both dialects normalise, fromLegacy reported, string→text remapped.\n")

        # Each control mutates the proposal in ONE way the gate claims to catch. A
        # control that does not fail means the gate is not reading what it says it is.
        bad = []
        for induced in ("types", "links", "markup", "facts", "volume"):
            print(f"\n===== induced control: {induced} =====")
            if grade(component_id, proposal, induce=induced) == 0:
                bad.append(induced)
        if bad:
            sys.exit(f"\nCONTROL FAILED: the gate PASSED with induced damage: {', '.join(bad)}. "
                     "It is not checking what it reports. Do not trust a green run.")
        print("\nCONTROL OK: every induced defect was caught.")
        return 0

    return grade(component_id, proposal)


if __name__ == "__main__":
    sys.exit(main())
