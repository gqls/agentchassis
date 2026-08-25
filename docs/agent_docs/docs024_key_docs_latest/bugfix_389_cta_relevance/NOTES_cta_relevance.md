# NOTES — CTA destination relevance (`bugs_open/389`). Append-only, newest at the bottom.

## 2026-08-25 — phase 0: cause found, and two of my own claims corrected

**Owner report:** `/tools/password-entropy.html` offered as the CTA on an AI-orchestration
consultancy — *"not deliberate and actually wrong"*. Raised out of yesterday's 277 session, where
I had measured the URL across three domains and deliberately **not** filed, because "that seems
off to me" is taste, not evidence. The owner's confirmation is what made it a defect.

### The path that found it (recorded because the wrong turns were the informative part)
1. Grepped for a hardcoded `password-entropy` — found none in live code, but found
   `005_content_components.sql:8942`, *"Narrow password-entropy tool affinity"*, explaining the
   tool was pushed to four sites *"because the library only had 2 tools with templates"*.
   **This is history and I nearly stopped here.** It explains why the tool EXISTS on those sites;
   it does not explain why CTAs POINT at it. Two different questions.
2. Assumed semantic/tag matching next — **wrong**. `chooseCTATargets` has no semantic input at
   all. Reading it took two minutes and killed the hypothesis.
3. The real answer: `nav_order` ascending, `name` as tiebreak, take `[0]`. Simulated it against
   live `pages`; the predicted winner matched the stored value on all 3 sites.

### ⚠ MISSTEP 1 — I nearly filed a 13-site claim that is false
Measured 13 sites whose rank-1 CTA target has `in_header=false` and drafted it as "13 deliberate
contradictions: a human hid it, the system ignored them". **The disconfirming check I had not run:
what fraction of tool pages are `in_header=false` at all? 143 of 228 — 62.7%, the majority state.**
So the flag does not mean a human judged anything. Only leopardess is a real case, and only
because `L5_nav_and_ctas.sql:29` carries the comment *"a password tool doesn't belong in the
primary nav"*. **The check that would have caught it is one `GROUP BY` and it costs nothing** —
before reading any flag as intent, measure its base rate. Now in the RUNBOOK.

### ⚠ MISSTEP 2 — the served-bytes probe that found nothing
Curled the **home pages** of finetuning and leopardess for `password-entropy`: **0 refs**, which
for a moment read as "not actually live". It was the wrong page: the stored fields sit on
`/services.html`, `/technical-details.html`, blog posts. **Probe the page that holds the field, not
the page you assume.** The home page of the third site *did* carry it — four references, minted
today.

### The finding that makes it urgent rather than cosmetic
The `__cta_minted` stamp (LNK-035, live 2026-08-22) splits the 80 fields into 17 resolver-minted
(dated **08-23 → 08-25, i.e. today**), 24 stamped-but-superseded, 39 unstamped. ⚠ **NULL is "not
recorded", not "authored"** — there is no backfill by design, so anything older than 08-22 is
unattributable. Reading NULL as authored would have made this look historical and closed.

### The structural point, which is the part worth carrying
`pages.nav_order` serves two unrelated readers — the nav menu and the CTA chooser — and nothing at
either site says so. `in_header` is read by one and not the other. That is why a human's explicit
"don't make this prominent" was a no-op: the two mechanisms disagree about which column carries
the intent, and there is no column that carries it for CTAs at all.
