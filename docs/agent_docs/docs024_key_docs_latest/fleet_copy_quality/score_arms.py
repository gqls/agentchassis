#!/usr/bin/env python3
"""Score each arm's PROPOSED copy on the same rules used for the live baseline.

Input: llm_call_log response_text rows (JSON with a "content" HTML string).
Reports per arm: sentence-length distribution, ceiling breaches, and the
style markers that differ between the arms (contractions, British spelling,
phrasal verbs, banned modals).
"""
import re, json, sys, os, importlib.util, statistics

SP = os.path.dirname(os.path.abspath(__file__))
spec = importlib.util.spec_from_file_location("audit", os.path.join(SP, "ste_audit.py"))
A = importlib.util.module_from_spec(spec); spec.loader.exec_module(A)

def sections_from_dump(path):
    raw = open(path, encoding='utf-8', errors='replace').read()
    out = []
    for block in raw.split("----SECTION----"):
        block = block.strip()
        if not block: continue
        lines = block.split("\n", 1)
        if len(lines) < 2: continue
        name, body = lines[0].strip(), lines[1].strip()
        html = None
        try:
            html = json.loads(body).get("content", "")
        except Exception:
            m = re.search(r'"content"\s*:\s*"(.*)"\s*}\s*$', body, re.S)
            if m:
                try: html = json.loads('"' + m.group(1) + '"')
                except Exception: html = m.group(1)
        if html: out.append((name, html))
    return out

def prose_sentences(html):
    h = re.sub(r'(?is)<(script|style)\b[^>]*>.*?</\1>', ' ', html)
    sents = []
    for m in re.finditer(r'(?is)<(p|li|h1|h2|h3)\b[^>]*>(.*?)</\1>', h):
        t = re.sub(r'(?is)<[^>]+>', '', m.group(2))
        for a, b in (('&amp;','&'),('&#39;',"'"),('&pound;','£'),('&nbsp;',' '),('&quot;','"')):
            t = t.replace(a, b)
        t = re.sub(r'\s+', ' ', t).strip()
        if len(t) > 15:
            for s in A.sentences(t):
                if A.wc(s) >= 3: sents.append(s)
    return sents

def score(label, path):
    secs = sections_from_dump(path)
    sents = []
    for name, html in secs: sents.extend(prose_sentences(html))
    if not sents:
        print(f"\n### {label}: NO PROSE EXTRACTED ({len(secs)} sections)"); return None
    lens = sorted(A.wc(s) for s in sents)
    print(f"\n### {label}")
    print(f"  sections {len(secs)} · sentences {len(sents)}")
    print(f"  words/sentence: mean {statistics.mean(lens):.1f}  median {statistics.median(lens)}  max {max(lens)}")
    def pct(rx, name):
        hit = [s for s in sents if re.search(rx, s.lower())]
        print(f"  {name:26} {len(hit):>3} / {len(sents)}  ({100*len(hit)/len(sents):.0f}%)")
        return hit
    over20 = [s for s in sents if A.wc(s) > 20]
    over25 = [s for s in sents if A.wc(s) > 25]
    over30 = [s for s in sents if A.wc(s) > 30]
    print(f"  {'over 20 words':26} {len(over20):>3} / {len(sents)}  ({100*len(over20)/len(sents):.0f}%)")
    print(f"  {'over 25 words (STE cap)':26} {len(over25):>3} / {len(sents)}  ({100*len(over25)/len(sents):.0f}%)")
    print(f"  {'over 30 words (arm4 cap)':26} {len(over30):>3} / {len(sents)}  ({100*len(over30)/len(sents):.0f}%)")
    pct(A.CONTRACTIONS, "contractions")
    pct(A.BRITISH, "British spelling")
    pct(A.PHRASAL, "phrasal verbs")
    pct(A.MODALS, "banned modals")
    pct(A.ING_VERB, "-ing as verb")
    if over30:
        print("  LONGEST:")
        for s in sorted(sents, key=A.wc, reverse=True)[:2]:
            print(f"    ({A.wc(s)}w) {s[:150]}")
    return sents

if __name__ == '__main__':
    for label, fn in [("ARM 3 — raw STE", "arm3_ste.txt"),
                      ("ARM 4 — house voice + 4 mechanisms", "arm4_house_plus.txt")]:
        p = os.path.join(SP, fn)
        if os.path.exists(p) and os.path.getsize(p) > 50:
            score(label, p)
        else:
            print(f"\n### {label}: not available yet")
