#!/usr/bin/env python3
"""Measure loancalculator.co.uk's live prose against ASD-STE100's mechanical rules.

Scope note (what this CAN and CANNOT see):
  CAN  — sentence length, contractions, banned modals, perfect/continuous forms,
         -ing-as-verb, semicolons, British spelling, a NAMED list of phrasal verbs,
         a NAMED list of unapproved vocabulary, Latin abbreviations, there-is openers.
  CANNOT — synonym rotation, one-word-one-meaning violations, noun-cluster length
         beyond a crude heuristic, warning/caution ordering, article omission.
         So every count below is a FLOOR, never a ceiling.
"""
import re, os, sys, glob, json
from collections import Counter, defaultdict

SP = os.path.dirname(os.path.abspath(__file__))
PAGES = sys.argv[1] if len(sys.argv) > 1 else os.path.join(SP, "pages")
# usage: ste_audit.py <dir-of-fetched-html>
#   fetch first, WITH the guard:
#     curl -s https://loancalculator.co.uk/sitemap.xml | grep -o '<loc>[^<]*</loc>' \
#       | sed 's/<[^>]*>//g' | while read u; do ... done   (see COMPARISON doc §2)

# ---------- extraction ----------
def prose_blocks(html):
    """Return (tag, text) for prose-bearing elements, minus chrome."""
    h = re.sub(r'(?is)<(script|style|nav|footer|head)\b[^>]*>.*?</\1>', ' ', html)
    h = re.sub(r'(?is)<(code|pre)\b[^>]*>.*?</\1>', ' CODEBLOCK ', h)  # STE: do not touch
    out = []
    for m in re.finditer(r'(?is)<(p|li|h2|h3)\b[^>]*>(.*?)</\1>', h):
        tag, inner = m.group(1).lower(), m.group(2)
        t = re.sub(r'(?is)<[^>]+>', '', inner)
        for a, b in (('&amp;', '&'), ('&#39;', "'"), ('&rsquo;', "'"), ('&lsquo;', "'"),
                     ('&pound;', '£'), ('&nbsp;', ' '), ('&quot;', '"'), ('&mdash;', '—'),
                     ('&ndash;', '–'), ('&hellip;', '…'), ('&#163;', '£')):
            t = t.replace(a, b)
        t = re.sub(r'\s+', ' ', t).strip()
        if len(t) > 25:
            out.append((tag, t))
    return out

def sentences(text):
    # protect £1,000.50 / 10.5% / e.g. / i.e. / U.K.
    t = re.sub(r'(\d)\.(\d)', r'\1<DOT>\2', text)
    t = re.sub(r'\b(e|i)\.(g|e)\.', lambda m: m.group(0).replace('.', '<DOT>'), t, flags=re.I)
    parts = re.split(r'(?<=[.!?])\s+(?=["“\(]?[A-Z0-9£])', t)
    return [p.replace('<DOT>', '.').strip() for p in parts if p.strip()]

def wc(s):
    return len([w for w in re.split(r'\s+', s.strip()) if re.search(r'[A-Za-z0-9£%]', w)])

# ---------- rule tables ----------
CONTRACTIONS = r"\b(it's|isn't|don't|doesn't|didn't|you're|we're|they're|you'll|they'll|we'll|it'll|i'm|that's|there's|here's|what's|won't|can't|cannot't|couldn't|wouldn't|shouldn't|haven't|hasn't|hadn't|aren't|weren't|wasn't|you've|we've|they've|i've|you'd|we'd|they'd|it'd|let's|who's|he's|she's)\b"
MODALS = r"\b(should|would|could|may|might)\b"
PERFECT_CONT = r"\b((has|have|had) been|(is|are|was|were) being|(has|have|had) [a-z]+ed)\b"
ING_VERB = r"\b(am|is|are|was|were|be|been|being) [a-z]+ing\b"
# CORRECTED: the original `\w*ise[sd]?\b` catch-all matched promise/rises/advertised/
# raise/exercise/otherwise — all correct in American English too. Explicit list only.
BRITISH = r"(\bamortisation\b|\bamortise|\borganisation\b|\borganise|\brecognise|\brealise|\bminimise|\bmaximise|\bprioritise|\bspecialis|\bsummaris|\bauthoris|\bcategoris|\bstandardis|\bpersonalis|\banalyse|\bcolour|\bfavour|\bbehaviour|\blabour|\bdefence\b|\blicence\b|\bcentre\b|\bwhilst\b|\bprogramme\b|\bpractise\b|\bmetre\b|\bcheque\b|\btravell|\bmodell|\bcancell|\benrol\b|\bgrey\b|\btowards\b|\binstalment)"
PHRASAL = r"\b(pay(s|ing|ed)? (back|off|down)|paid (back|off|down)|chip(s|ping|ped)? away|hand(s|ing|ed)? (it |the [a-z]+ )?back|add(s|ing|ed)? up|com(e|es|ing) out|tak(e|es|ing) (out|on)|took (out|on)|work(s|ing|ed)? out|go(es|ing)? (over|with|up|down)|went (over|with)|break(s|ing)? down|broke down|set(s|ting)? up|carry(ing)? out|carried out|find(s|ing)? out|found out|end(s|ing|ed)? up|build(s|ing)? up|built up|knock(s|ing|ed)? off|sign(s|ing|ed)? up|put(s|ting)? (off|toward|towards)|turn(s|ing|ed)? (to|into)|run(s|ning)? out|ran out|keep(s|ing)? up|kept up|catch(es|ing)? up|caught up|bring(s|ing)? down|brought down|cut(s|ting)? down|shop(s|ping|ped)? around|stack(s|ing|ed)? up|wip(e|es|ing)? out|hold(s|ing)? back|held back)\b"
UNAPPROVED = {
    'utilize': 'use', 'utilise': 'use', 'leverage': 'use', 'employ': 'use',
    'commence': 'start', 'initiate': 'start', 'originate': 'start',
    'terminate': 'stop/end', 'cease': 'stop/end', 'conclude': 'stop/end',
    'ensure': 'make sure (that)', 'verify': 'make sure (that)', 'confirm': 'make sure (that)',
    'validate': 'make sure (that)',
    'perform': 'do', 'conduct': 'do', 'execute': 'do',
    'facilitate': 'help', 'assist': 'help',
    'obtain': 'get', 'acquire': 'get', 'procure': 'get',
    'sufficient': 'enough', 'adequate': 'enough',
    'approximately': 'about', 'accomplish': 'do', 'additional': 'more',
    'supplementary': 'more', 'attempt': 'try', 'require': 'need/must',
    'requires': 'need/must', 'required': 'need/must', 'necessitate': 'need/must',
    'mandatory': 'necessary', 'indicate': 'show', 'indicates': 'show',
    'signify': 'show', 'rotate': 'turn', 'deactivate': 'turn off',
    'activate': 'turn on', 'toxic': 'poisonous', 'accessible': '(rewrite)',
    'remainder': 'rest', 'demonstrate': 'show', 'modify': 'change', 'alter': 'change',
    'construct': 'assemble/make', 'fabricate': 'assemble/make', 'retain': 'keep',
    'locate': 'find', 'depress': 'push/press', 'proceed': 'continue/go',
}
UNAPPROVED_PHRASES = {
    'in order to': 'to', 'prior to': 'before', 'subsequent to': 'after',
    'adjacent to': 'near', 'by means of': 'through/with', 'due to': 'because of',
    'owing to': 'because of', 'in the event of': 'if', 'in the event that': 'if',
}
# CORRECTED: the original matched 'check' as a NOUN ("a credit check"), which STE allows,
# and clear/right/above/below in senses it permits. Verb uses only, plus 'via'.
ONE_MEANING = (r"((^|[.!?] ?)Check\b|\b(to|and|before|after|when|by|without) check(ing)?\b"
               r"|\bcheck (your|the|their|this|that|whether|if)\b|\bvia\b"
               r"|\bfollow (the|these|this|those) (steps|instructions|rules|guidance)\b)")
LATIN = r"(\be\.g\.|\bi\.e\.|\betc\.)"
THERE_IS = r"^(There (is|are|was|were)|there (is|are|was|were))\b"

def audit():
    per_page = {}
    all_sent = []
    for f in sorted(glob.glob(os.path.join(PAGES, "*"))):
        html = open(f, encoding='utf-8', errors='replace').read()
        if not html.lstrip().startswith("<!DOCTYPE"):
            print("SKIP (no doctype):", f); continue
        blocks = prose_blocks(html)
        page = os.path.basename(f)
        sents = []
        for tag, text in blocks:
            for s in sentences(text):
                if wc(s) >= 3:
                    sents.append((page, tag, s))
        per_page[page] = len(sents)
        all_sent.extend(sents)

    hits = defaultdict(list)
    for page, tag, s in all_sent:
        low = s.lower()
        n = wc(s)
        if n > 25: hits['over_25_words'].append((page, n, s))
        if n > 20: hits['over_20_words'].append((page, n, s))
        for name, rx in (('contractions', CONTRACTIONS), ('modals_should_would_could_may_might', MODALS),
                         ('perfect_or_continuous', PERFECT_CONT), ('ing_as_verb', ING_VERB),
                         ('british_spelling', BRITISH), ('phrasal_verbs', PHRASAL),
                         ('one_meaning_verb_uses', ONE_MEANING), ('latin_abbrev', LATIN)):
            m = re.findall(rx, low)
            if m: hits[name].append((page, len(m), s))
        if ';' in s: hits['semicolons'].append((page, s.count(';'), s))
        if re.match(THERE_IS, s): hits['there_is_opener'].append((page, 1, s))
        # STE exempts proper nouns. Mask them before the vocabulary pass, or
        # "Financial Conduct Authority" scores as an unapproved use of "conduct".
        masked = re.sub(r'Financial Conduct Authority|Consumer Credit Act|Consumer Duty|'
                        r'Bank of England|Companies House|Money Helper|MoneyHelper',
                        ' PROPERNOUN ', s)
        masked_low = masked.lower()
        for w in UNAPPROVED:
            if re.search(r'\b' + w + r'\b', masked_low):
                hits['unapproved_vocab'].append((page, w, s)); break
        for p in UNAPPROVED_PHRASES:
            if p in low:
                hits['unapproved_phrases'].append((page, p, s)); break
    return all_sent, per_page, hits

if __name__ == '__main__':
    all_sent, per_page, hits = audit()
    total = len(all_sent)
    print(f"PAGES AUDITED : {len(per_page)}")
    print(f"SENTENCES     : {total}\n")
    print(f"{'STE rule':42} {'sentences hit':>14} {'% of corpus':>12}")
    print("-" * 70)
    order = ['over_25_words', 'over_20_words', 'contractions', 'phrasal_verbs',
             'one_meaning_verb_uses', 'ing_as_verb', 'british_spelling',
             'modals_should_would_could_may_might', 'perfect_or_continuous',
             'unapproved_vocab', 'unapproved_phrases', 'there_is_opener',
             'semicolons', 'latin_abbrev']
    for k in order:
        uniq = {(p, s) for p, _, s in hits[k]}
        print(f"{k:42} {len(uniq):>14} {100*len(uniq)/total:>11.1f}%")

    # union: sentences failing at least one mechanical rule
    fail = set()
    for k in order:
        if k == 'over_20_words':   # avoid double counting; 25 is the generous cap
            continue
        for p, _, s in hits[k]:
            fail.add((p, s))
    print("-" * 70)
    print(f"{'AT LEAST ONE VIOLATION (descriptive cap)':42} {len(fail):>14} {100*len(fail)/total:>11.1f}%")
    print(f"{'CLEAN under the checked rules':42} {total-len(fail):>14} {100*(total-len(fail))/total:>11.1f}%")

    print("\n\n=== SAMPLES: the site's own sentences, by rule ===")
    for k in order:
        if not hits[k]: continue
        print(f"\n--- {k} ---")
        seen = set()
        for p, meta, s in hits[k]:
            if s in seen: continue
            seen.add(s)
            print(f"  [{p[:34]}] ({meta}) {s[:190]}")
            if len(seen) >= 3: break
