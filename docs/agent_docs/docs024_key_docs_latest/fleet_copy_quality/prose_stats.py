import re, sys, subprocess, statistics as st, collections

def text_of(url):
    h = subprocess.run(["curl","-s",url],capture_output=True,text=True,timeout=60).stdout
    if len(h) < 3000 or "DOCTYPE" not in h[:200].upper(): return None
    body = re.sub(r'(?is)<(script|style|nav|header|footer)[^>]*>.*?</\1>',' ',h)
    body = re.sub(r'(?is)<!--.*?-->',' ',body)
    paras = re.findall(r'(?is)<p[^>]*>(.*?)</p>', body)
    paras = [re.sub(r'<[^>]+>',' ',p) for p in paras]
    paras = [re.sub(r'\s+',' ',p).strip() for p in paras]
    return [p for p in paras if len(p) > 40]

HEDGE = re.compile(r"\b(can be wrong|may be wrong|is not advice|not financial advice|we cannot|we can't|does not constitute|no guarantee|check with|verify|indicative only|for guidance only|always seek)\b", re.I)

def report(label, urls):
    paras=[]
    for u in urls:
        t = text_of(u)
        if t: paras += t
    sents=[]
    for p in paras:
        sents += [s.strip() for s in re.split(r'(?<=[.!?])\s+', p) if len(s.strip())>1]
    if not sents:
        print(f"{label}: no prose"); return
    lens=[len(s.split()) for s in sents]
    firsts=collections.Counter(" ".join(s.split()[:2]).lower().strip(",.") for s in sents)
    popens=collections.Counter(" ".join(p.split()[:2]).lower().strip(",.") for p in paras)
    hedges=sum(len(HEDGE.findall(p)) for p in paras)
    print(f"\n{label}")
    print(f"  paragraphs {len(paras)}  sentences {len(sents)}")
    print(f"  sentence words: mean {st.mean(lens):.1f}  stdev {st.pstdev(lens):.1f}  "
          f"CV {st.pstdev(lens)/st.mean(lens):.2f}  min {min(lens)} max {max(lens)}")
    short=sum(1 for l in lens if l<=6); long=sum(1 for l in lens if l>=30)
    print(f"  very short (<=6 words) {100*short/len(lens):.0f}%   long (>=30) {100*long/len(lens):.0f}%")
    print(f"  top sentence openings: {', '.join(f'{k}×{v}' for k,v in firsts.most_common(4))}")
    print(f"  top paragraph openings: {', '.join(f'{k}×{v}' for k,v in popens.most_common(4))}")
    print(f"  hedge/disclaimer hits: {hedges} ({hedges/max(len(paras),1):.2f} per para)")

report("lendzy.co.uk (framework-built greenfield, 2026-08)", [
 "https://lendzy.co.uk/index.html","https://lendzy.co.uk/about.html"])
report("loancalculator.co.uk (other LLM, outside framework)", [
 "https://loancalculator.co.uk/guides/how-loans-are-calculated.html",
 "https://loancalculator.co.uk/guides/hidden-loan-fees.html",
 "https://loancalculator.co.uk/guides/secured-vs-unsecured.html"])
report("webdesign.co.uk (framework-built)", [
 "https://webdesign.co.uk/index.html","https://webdesign.co.uk/about.html"])
