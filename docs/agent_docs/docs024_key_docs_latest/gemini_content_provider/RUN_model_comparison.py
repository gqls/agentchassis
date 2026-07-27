#!/usr/bin/env python3
"""Same prompt, same material, both models, N runs each, mechanically scored.

The point is to answer "is it the model or the prompt?" with data instead of two
non-comparable n=1 samples. Everything is held constant except the provider:
identical prompt text (the REAL page-content-writer prompt_template), identical
material, identical visible budget (8000). Gemini additionally gets the thinking
reserve our client adds, because that is what our client does.

Mechanical scores only. Whether the prose is GOOD is not mechanically decidable
and is left to the reader — the samples are written out for that.
"""
import json, os, re, sys, time, urllib.request

PROMPT = open(sys.argv[1], encoding="utf-8").read()
N = int(sys.argv[2]) if len(sys.argv) > 2 else 5
OUT = sys.argv[3]

GEM_KEY = os.environ["GEMINI_API_KEY"]
ANT_KEY = os.environ["ANTHROPIC_API_KEY"]
VISIBLE_BUDGET = 8000
RESERVE = 8192

FILLER = ["crucially", "seamless", "seamlessly", "robust", "leverage", "delve",
          "landscape", "realm", "testament", "unlock", "harness", "navigate",
          "empower", "cutting-edge", "transformative", "holistic", "synergy",
          "in essence", "at its core", "furthermore", "moreover"]


def post(url, body, headers):
    req = urllib.request.Request(url, data=json.dumps(body).encode(),
                                 headers={"Content-Type": "application/json", **headers})
    t0 = time.time()
    with urllib.request.urlopen(req, timeout=300) as r:
        return json.loads(r.read()), time.time() - t0


def run_gemini():
    body = {"contents": [{"role": "user", "parts": [{"text": PROMPT}]}],
            "generationConfig": {"maxOutputTokens": VISIBLE_BUDGET + RESERVE}}
    d, secs = post("https://generativelanguage.googleapis.com/v1beta/models/"
                   "gemini-pro-latest:generateContent", body, {"x-goog-api-key": GEM_KEY})
    c = d["candidates"][0]
    txt = "".join(p.get("text", "") for p in c["content"]["parts"] if not p.get("thought"))
    u = d.get("usageMetadata", {})
    return txt, {"secs": round(secs, 1), "finish": c.get("finishReason"),
                 "visible_tok": u.get("candidatesTokenCount", 0),
                 "thinking_tok": u.get("thoughtsTokenCount", 0),
                 "input_tok": u.get("promptTokenCount", 0)}


def run_claude():
    body = {"model": "claude-sonnet-4-6", "max_tokens": VISIBLE_BUDGET,
            "messages": [{"role": "user", "content": PROMPT}]}
    d, secs = post("https://api.anthropic.com/v1/messages", body,
                   {"x-api-key": ANT_KEY, "anthropic-version": "2023-06-01"})
    txt = "".join(b.get("text", "") for b in d.get("content", []) if b.get("type") == "text")
    u = d.get("usage", {})
    return txt, {"secs": round(secs, 1), "finish": d.get("stop_reason"),
                 "visible_tok": u.get("output_tokens", 0), "thinking_tok": 0,
                 "input_tok": u.get("input_tokens", 0)}


def score(txt):
    s = {}
    raw = txt.strip()
    s["fenced"] = raw.startswith("```")
    body = re.sub(r"^```[a-z]*\n?|```$", "", raw, flags=re.M).strip()
    try:
        d = json.loads(body)
        s["valid_json"] = True
        s["all_keys"] = {"headline", "subheadline", "body", "cta_text"} <= set(d)
        prose = " ".join(str(v) for v in d.values())
        head = str(d.get("headline", ""))
    except Exception:
        s["valid_json"] = False; s["all_keys"] = False
        prose = body; head = body[:80]
    s["em_dash"] = prose.count("—")
    s["bang"] = prose.count("!")
    s["filler"] = sum(1 for w in FILLER if re.search(rf"\b{re.escape(w)}\b", prose, re.I))
    s["contractions"] = len(re.findall(r"\b\w+'(s|re|ve|ll|t|d)\b", prose))
    s["neg_open"] = bool(re.match(r"^\s*(not |it'?s not|this isn'?t|forget |we don'?t)", head, re.I))
    # the exact tic the 2026-07-24 Claude test surfaced ("That second part matters")
    s["that_x_matters_tic"] = len(re.findall(r"\bThat\b[^.]{0,40}\bmatters\b", prose, re.I))
    sents = [x for x in re.split(r"(?<=[.!?])\s+", prose) if x.strip()]
    s["sentences"] = len(sents)
    s["mean_words"] = round(sum(len(x.split()) for x in sents) / max(len(sents), 1), 1)
    s["chars"] = len(prose)
    return s, prose


results = {}
samples = {}
for name, fn in (("gemini-pro-latest", run_gemini), ("claude-sonnet-4-6", run_claude)):
    rows, texts = [], []
    for i in range(N):
        try:
            txt, meta = fn()
        except Exception as e:
            print(f"{name} run {i+1}: ERROR {e}", file=sys.stderr)
            continue
        sc, prose = score(txt)
        rows.append({**meta, **sc})
        texts.append(prose)
        print(f"{name} run {i+1}/{N}: json={sc['valid_json']} em={sc['em_dash']} "
              f"filler={sc['filler']} think={meta['thinking_tok']} {meta['secs']}s")
    results[name] = rows
    samples[name] = texts

json.dump({"results": results, "samples": samples}, open(OUT, "w"), indent=2)
print("\nwritten", OUT)
