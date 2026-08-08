# PLAN — bugs_open/221, the AI-disclosure family matches a construction, not a noun phrase

**Status: IMPLEMENTED and committed (`61c8cc6ff`), council submission
`377a0488-214e-4e5c-bd3d-66343d34d9b2` pending. Inert until a chassis image is
rebuilt and rolled.**

## The decision, and why this shape

`metaCommentaryPatterns` entries gain an **optional `Re *regexp.Regexp`**. Nil
means today's case-insensitive substring match, so the thirteen refusal /
schema / pipeline entries are byte-for-byte unchanged. The two first-person
AI-disclosure entries carry a regex requiring the **construction** the family
was written for — an AI noun phrase from a **closed set**, then the first person
**immediately** (optional comma, no other gap).

```
(?i)\bas\s+an\s+(?:ai|artificial\s+intelligence|llm)
    (?:[\s-]+(?:language[\s-]+)?model|[\s-]+assistant|[\s-]+system|[\s-]+chatbot)?
    (?:\s*,\s*|\s+)i\b
```

Rejected alternatives, with reasons:

- **Hand-tuned substrings** (`"as an ai, i "`, `"as an ai i "`, …) — this is the
  bug's own failure mode a second time. It misses `As an AI, I'm unable…`
  (apostrophe, not space), misses the newline forms `ExtractAssertionText`
  legitimately produces from indented HTML, and needs a combinatorial list
  (comma × no-comma × suffix × I/I'm/I'd). A substring list cannot express
  "followed by the first person" at all.
- **A `Kind` enum / matcher framework** — two entries need a rule and thirteen
  need a literal. A nullable regex is the smallest declarative way to say "this
  entry matches a construction", and it stays inside this one check.
- **Dropping severity to warning for the family** (the bug's candidate 2) —
  rejected. See below.

**Why adjacency is the load-bearing part.** `As an AI engineer, I built this`
is a human's bio and contains *both* `as an ai` *and* a first-person `I`. If any
gap were allowed before the `I`, the rule would degenerate to "ai near I" and
convict it. The closed suffix set plus strict adjacency is the only version that
separates the apology from the job title — and mutation B below is what proves
the requirement is doing that work rather than decorating the regex.

## The two open questions the bug file raised, answered

**`as a language model`: narrow it the same way.** Identical mechanism,
identical exposure. Zero live hits today, but "run any open-weights checkpoint
**as a language model** backend" is exactly the copy this estate writes, and the
cost of being wrong is a permanently unbuildable page. A genuine disclosure of
this form is first-person by construction, so the detection cost is ~nil.

**Severity: keep `blocker`; candidate 1 alone.** Three reasons. The 219 test
contract pins this exact value at blocker, and rewriting it in the same change
that narrows the pattern would be two simultaneous loosenings — the live row's
absolution could then no longer be attributed to precision rather than to
disarmament. Once the pattern requires the construction, a match **is** the
thing the check was built for (robot-hands, 2026-07-14), and a warning ships
that apology to a reader. And a precise pattern that is worth keeping is worth
blocking on.

## Scope held deliberately narrow

`bugs_open/221` routes "how should the fleet's blocker-severity string scans be
governed" to an RFC, and `bugs_open/222` is the same class in
`check_tool_fabrication_action.go`, **owned by the mortgagecalculator lane**.
Neither is touched. Nothing shared is touched either, and this was measured
rather than asserted: `checkMetaCommentary` has exactly **one** production call
site (`validate_page_content.go:332`), `metaCommentaryPatterns` is referenced
only inside that function, and the disclosure phrases and the `meta_commentary`
category appear in **no other Go file in the tree**.

## Verification — every step able to come out otherwise

| # | Check | Result |
|---|---|---|
| 1 | **Test written and run BEFORE the fix** | 7/7 must-not-block **FAILED** at HEAD, 7/7 must-block **PASSED** — the test can fail, and the change is a pure narrowing |
| 2 | Real check over the live 12,879-byte artefact | 1 blocker → **0** |
| 3 | Same artefact, `As an AI, I cannot generate this listing.` injected | still **blocks** (2 hits: the disclosure arm and `i cannot generate`) — the narrowing did not disarm the check |
| 4 | Mutation A — `Re` → nil (bare substring again) | 6 must-not-block cases fail ✓ |
| 5 | Mutation B — first person made optional | the same 6 fail, **including the human bio** ✓ — adjacency is load-bearing |
| 6 | Mutation C — delete entry 2 | only its own must-block case fails ✓ — entry 2 is not inert |
| 7 | Pattern set HEAD vs now, by set difference | 15 = 15, **identical**; nothing added, dropped or altered |
| 8 | Full package suite, `-count=1` | green (actions, discovery_checks, queryresolve) |

## Accepted residuals — a narrowing IS a narrowing, so they are named

**False negatives now accepted:** interposed modifiers (`As an AI model trained
by X, I cannot…`) and trailing disclosures (`…I can't help with that as an
AI.`). Deliberate — allowing a gap re-convicts the human bio, which is the very
damage being fixed. Mitigation is *measured*, not hoped for: step 3 shows a real
refusal fires the disclosure arm **and** `i cannot generate`, and the six
refusal-prose entries are unchanged. The asymmetry settles the direction — **a
shipped miss is visible and editable; a false blocker is permanent.**

**False positive still accepted:** deliberate first-person chatbot-persona copy
(`As an AI assistant, I'm always awake.`) still blocks. That *is* a first-person
disclosure; wanting it is a per-site allowance decision, not a reason to quieten
the pattern fleet-wide.

**The nil default is the status quo, NOT the cautious direction.** Nil means
substring, and substring over-matches — which is this whole bug. A future entry
whose text could sit inside a longer legitimate phrase must set `Re`, and the
compiler will not remind anyone. That is why it is written at the field, and why
it is now a LANDMINES entry.

## What would change this plan

- The council returning REVISE on the closed suffix set (is a member wrong, is a
  common disclosure form missing) — the most likely objection, and the right one.
- A decision to govern all blocker-severity string scans centrally (the RFC
  route). This change is compatible with that and does not pre-empt it: the
  `Re` field is one function's private detail.
- `bugs_open/222`'s lane choosing a negation-aware approach for the fabrication
  gate. Different mechanism (negation vs. first person), so no conflict, but the
  two together would be the evidence base for the RFC.
