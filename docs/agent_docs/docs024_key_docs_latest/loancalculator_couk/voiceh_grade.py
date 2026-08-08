#!/usr/bin/env python3
"""Grade voice-H rewrites on loancalculator.co.uk, page by page.

Usage:  voiceh_grade.py <baseline.json> <page-name> [<page-name> ...]

The grading question is NOT "did the page change" — a carried no-op and a
successful rewrite both leave a page that looks fine. So every check here is
chosen to separate the rewrite from INACTION as well as from damage:

  row_replaced  the prose row's id differs from the pre-run backup. save_page_sections
                does DELETE+INSERT, so a new id is proof the save actually ran; a
                carry branch would keep the id and the length could still coincide.
  facts         every number, %, £ figure, statute year, internal link and heading in
                the baseline text must still be present. Additions are reported too,
                and are NOT automatically a pass: the brief says preserve, not improve.
  locks         a locked calculator row must keep its id AND its updated_at. Same id
                with a newer stamp means something rewrote it in place.
  served        the live page must carry the new opening AND return 0 for the baseline
                opening — a negative control on the artefact, not just a positive match.

The served fetch is guarded by size + DOCTYPE first: during a deploy window B2
returns an error blob at HTTP 200, and every grep against that blob reads clean.
"""
import json, re, subprocess, sys, time

SITE = '0162cde4-633e-45e9-8ca6-87a6b2fe1d26'
DOMAIN = 'loancalculator.co.uk'

NUM = re.compile(r'£[\d,]+(?:\.\d+)?|\b\d+(?:\.\d+)?%|\b(?:19|20)\d{2}\b|\b\d+\s*(?:day|month|year)s?\b')


WORDS = {'1': 'one', '2': 'two', '3': 'three', '4': 'four', '5': 'five', '6': 'six',
         '7': 'seven', '8': 'eight', '9': 'nine', '10': 'ten', '11': 'eleven',
         '12': 'twelve', '20': 'twenty', '30': 'thirty', '50': 'fifty'}


def norm(tok):
    """Singularise the unit so a rephrase is not reported as a lost fact.

    "the entire 3 or 4 years" -> "3 or 4 year agreement" keeps the fact and
    changes only the grammar; a plural-only pattern calls that a lost figure.
    """
    tok = tok.strip().rstrip(',.;:')
    return re.sub(r'\s+', ' ', tok.rstrip('s')) if re.search(r'(day|month|year)s?$', tok) else tok


def still_present(tok, text):
    """Is this baseline figure still stated, in ANY form the voice permits?

    The H voice tells the writer to speak numbers the way a person would, so it
    routinely turns "4-5 years" into "four to five years" and "£25k+" into
    "£25,000 or more". A digit-only comparison calls every one of those a LOST
    FACT — 5 of the 6 fact flags in batches 2a/2b were exactly this, and a
    checker that cries wolf five times out of six is a checker nobody reads.
    So: accept the spelled-out form and the k/comma variants, and report only
    what is absent under every one of them.
    """
    if tok in text:
        return True
    m = re.match(r'^(\d+)\s*(day|month|year)s?$', tok)
    if m and WORDS.get(m.group(1)):
        if re.search(rf'\b{WORDS[m.group(1)]}\b[^.]{{0,40}}\b{m.group(2)}s?\b', text):
            return True
    m = re.match(r'^£([\d,]+)$', tok)
    if m:
        digits = m.group(1).replace(',', '')
        variants = {f'£{digits}', f'£{int(digits):,}'}
        if len(digits) > 3 and digits.endswith('000'):
            variants.add(f'£{digits[:-3]}k')
        if any(v in text for v in variants):
            return True
        # "£25" extracted from "£25k+" is a fragment, not a figure of its own
        if re.search(rf'£{re.escape(m.group(1))}[k,\d]', text):
            return True
    return False


def psql(sql):
    return subprocess.run(
        ['kubectl', '-n', 'ai-persona-system', 'exec', '-i', 'postgres-clients-0', '--',
         'psql', '-U', 'clients_user', '-d', 'clients_db', '-t', '-A', '-c', sql],
        capture_output=True, text=True, check=True).stdout.strip()


def live_rows(page):
    # JSON, not column output: the prose contains newlines AND tabs, so any
    # delimiter-split parse silently shreds one row into several.
    out = psql(f"""SELECT coalesce(jsonb_agg(r ORDER BY (r->>'pos')::int), '[]'::jsonb)::text FROM (
                     SELECT jsonb_build_object('slot', pc.slot_name, 'id', pc.id,
                              'lock', pc.lock_type, 'updated', pc.updated_at,
                              'pos', pc.position,
                              'text', coalesce(pc.content_data->>'content','')) AS r
                     FROM pages p JOIN page_components pc ON pc.page_id=p.id
                     WHERE p.site_id='{SITE}' AND p.name='{page}') s;""")
    return json.loads(out)


def facts(s):
    text = re.sub(r'<[^>]+>', ' ', s)
    return (set(norm(t) for t in NUM.findall(text)),
            set(re.findall(r'href="([^"]+)"', s)),
            [(h[0], re.sub(r'\s+', ' ', re.sub(r'<[^>]+>', '', h[1])).strip())
             for h in re.findall(r'<h([1-6])[^>]*>(.*?)</h\1>', s, re.S)])


def opening(s, words=9):
    """First sentence of real body prose.

    MUST come from inside a single <p>, not from the whole stripped document.
    Stripping tags first concatenates the <h1> with the following subtitle into
    one pseudo-sentence that exists nowhere in the served HTML (tags sit between
    them) — which reads as "the new copy never reached the page" on a page that
    deployed perfectly. That false negative fired on 3 of 4 pages in batch 2a.
    """
    for para in re.findall(r'<p\b([^>]*)>(.*?)</p>', s, re.S):
        if 'subtitle' in para[0]:
            continue
        text = re.sub(r'\s+', ' ', re.sub(r'<[^>]+>', '', para[1])).strip()
        for sent in re.split(r'(?<=[.!?]) ', text):
            if len(sent.split()) >= words:
                return sent.strip()
    return ''


def main():
    baseline = json.load(open(sys.argv[1]))
    pages = sys.argv[2:]
    base = {}
    for r in baseline:
        base.setdefault(r['page'], {})[r['slot']] = r

    verdicts = []
    for page in pages:
        url = next(iter(base[page].values()))['url']
        rows = live_rows(page)
        problems, notes = [], []

        prose = [r for r in rows if r['slot'].startswith('prose')]
        for r in prose:
            b = base[page].get(r['slot'])
            if b is None:
                problems.append(f"{r['slot']}: no baseline row")
                continue
            if r['id'] == b['row_id']:
                problems.append(f"{r['slot']}: row id UNCHANGED — the save did not run")
            ob, nb = facts(b['text']), facts(r['text'])
            new_plain = re.sub(r'\s+', ' ', re.sub(r'<[^>]+>', ' ', r['text']))
            missing_n = sorted(t for t in ob[0] - nb[0] if not still_present(t, new_plain))
            missing_l = ob[1] - nb[1]
            if missing_n:
                problems.append(f"{r['slot']}: FACTS LOST {missing_n}")
            if missing_l:
                problems.append(f"{r['slot']}: LINKS LOST {sorted(missing_l)}")
            # STRUCTURE is what the brief protects (same headings, same levels,
            # same order). Wording inside a heading is prose and the voice may
            # touch it — but an h1 rewrite changes the page's title, so it is
            # called out separately rather than buried in a note.
            ol, nl = [h[0] for h in ob[2]], [h[0] for h in nb[2]]
            if ol != nl:
                problems.append(f"{r['slot']}: heading STRUCTURE changed {ol} -> {nl}")
            for (lvl, o_t), (_, n_t) in zip(ob[2], nb[2]):
                if o_t != n_t:
                    tag = 'H1 (page title) REWORDED' if lvl == '1' else f'h{lvl} reworded'
                    notes.append(f"{r['slot']}: {tag}: {o_t!r} -> {n_t!r}")
            added = nb[0] - ob[0]
            if added:
                notes.append(f"{r['slot']}: added {sorted(added)} (read it — additions are not licensed)")
            if re.search(r'<(form|input|button|script|select)\b', r['text']):
                problems.append(f"{r['slot']}: introduced a form control or script")

            # ── CSS survival ──────────────────────────────────────────────
            # 8 of this site's 51 "prose" slots are not prose at all: the
            # decomposer put the page's <style> block in prose-0. The writer's
            # invariant ("this block contains no element a script addresses, so
            # rewriting it cannot break a calculator") is TRUE and says nothing
            # about CSS. On tool-compare-loans the rewrite dropped the style
            # block, taking .comparison-wrapper/.loan-column/.stat-label/
            # .stat-value off a page whose markup still referenced two of them.
            # It kept the block on the other three such pages — so this is a
            # coin flip, not a rule, and it must be checked every time.
            old_sel = set(re.findall(r'([.#][\w-]+)\s*\{', b['text']))
            lost_sel = sorted(s for s in old_sel if not re.search(re.escape(s) + r'\s*\{', r['text']))
            if lost_sel:
                problems.append(f"{r['slot']}: CSS RULES LOST {lost_sel} — restore this row from "
                                f"page_components_bak_20260807_voiceh")

        for r in [r for r in rows if r['lock']]:
            b = base[page].get(r['slot'])
            if b and (r['id'] != b['row_id'] or r['updated'] != b['updated_at']):
                problems.append(f"{r['slot']}: LOCKED ROW MOVED (id or updated_at changed)")

        # served artefact, with the deploy-window guard
        served, ok = '', False
        for _ in range(3):
            served = subprocess.run(['curl', '-s', f'https://{DOMAIN}{url}?cb={int(time.time())}'],
                                    capture_output=True, text=True).stdout
            if len(served) > 2000 and '<!DOCTYPE' in served[:200]:
                ok = True
                break
            time.sleep(10)
        if not ok:
            problems.append(f"served page unusable ({len(served)} b, no DOCTYPE) — deploy window?")
        else:
            # Strip tags from the SERVED html before matching. The opening comes
            # from tag-stripped DB text, so any inline <strong>/<em> inside that
            # first sentence breaks a raw-html substring search and reads as "the
            # copy never shipped" on a page that shipped perfectly.
            served = re.sub(r'\s+', ' ', re.sub(r'<[^>]+>', '', served))
            for r in prose:
                b = base[page].get(r['slot'])
                if not b:
                    continue
                new_open, old_open = opening(r['text']), opening(b['text'])
                if new_open[:40] and new_open[:40] not in served:
                    problems.append(f"{r['slot']}: new opening NOT on the served page")
                if old_open[:40] and old_open[:40] in served and old_open[:40] != new_open[:40]:
                    problems.append(f"{r['slot']}: baseline opening STILL on the served page")

        verdicts.append((page, problems, notes, prose))
        mark = 'PASS' if not problems else 'FAIL'
        sizes = ', '.join(f"{r['slot']} {len(base[page][r['slot']]['text'])}->{len(r['text'])}b"
                          for r in prose if r['slot'] in base[page])
        print(f"[{mark}] {page:36s} {sizes}")
        for p in problems:
            print(f"        ✗ {p}")
        for n in notes:
            print(f"        · {n}")

    bad = [v[0] for v in verdicts if v[1]]
    print(f"\n{len(verdicts)-len(bad)}/{len(verdicts)} pass" + (f"; FAILED: {bad}" if bad else ""))
    return 1 if bad else 0


if __name__ == '__main__':
    sys.exit(main())
