# The two builder flows, side by side — for the owner's decision, 2026-08-18

Written at the owner's request while builds are halted. **Both flows use the same entry script
and the same agent graph** (`082_submit_domain_unified.sh`, FRESH path → `domain-submitter` →
classifier → strategist → briefing → planner → design → pages). They are not different
pipelines. **They differ only in what is prepared before dispatch** — and that turns out to
decide which safety machinery exists at all.

Everything below is measured from the three sites that have actually run each flow.

## The two flows

| | **A — seeded + briefed** (pilot, build #1) | **B — prompt only** (`loanzy.uk`) |
|---|---|---|
| Register entry | yes, with neighbours + must-nots | none |
| Seed SQL | site row + **email**, **evidence_base**, **imagery_style_guide** | none |
| Mission | hand-written brief (~3,000 chars) from the register | a short prompt |
| Human time per domain | ~45–60 min (draft mission, adapt seed, verify guards) | ~2 min |
| Machine cost per domain | **identical** — ~$3.81 text + imagery | ~$3.81 text + imagery |

**The machine cost is the same. The whole difference is human time versus safety.**

## What each site actually ended up with (measured 2026-08-18)

| spec | remortgagecalculator.uk (A) | adversecreditmortgage.co.uk (A) | loanzy.uk (B) |
|---|---|---|---|
| `email` | ✅ | ✅ | **❌ NONE** |
| `evidence_base` | ✅ | ✅ | **❌ never existed (0 rows all-history)** |
| `imagery_style_guide` | ✅ | ✅ | **❌** |
| `mission_brief` | ✅ | ✅ | ❌ |
| identity / classification / strategy / briefing / content_direction / design_intent / vertical_landscape | ✅ | ✅ | ✅ |

The agent-written specs are identical across both flows — the classifier, strategist and
briefing agent do their work either way. **Only the seeded specs differ, and those are the ones
that carry the guards.**

## Where each fails, and whether you would notice

### Flow B fails silently, in three places — all currently live on a 20-page site

1. **No claims layer at all.** `loadEvidenceBase` returns nil when the aspect is absent and
   **every claims lane silently no-ops** (`validate_page_content.go:727-746`). `loanzy.uk` has
   **20 pages** and has never had an `evidence_base`. Nothing has checked a single assertion on
   any of them, and no `banned_claims` pattern exists to catch a guarantee, a superlative or an
   invented figure. On a finance-adjacent site that is the exposure the whole verification layer
   was built for. **It produces no error and no warning — the pages simply pass.**
2. **The hallucinated-email check fails OPEN with no email** (`bugs_open/063`). A fabricated
   address reached production for hours on another site exactly this way. `loanzy.uk` has none.
3. **`content_hero` generates unstyled without an `imagery_style_guide`** (`bugs_closed/027`) —
   cosmetic next to the other two, but it is why flow-A sites look coherent and this one may not.

### Flow A's failure mode is different in kind

It fails **loudly and early, at a human**: if nobody writes the mission, nothing is dispatched.
Its real cost is that it does not scale — 45–60 minutes per domain is ~100 hours across the
remaining fleet, and a rushed or thin mission degrades the site without announcing it.

Flow A is also where this lane's own two seeding bugs happened (six inert `banned_claims`
patterns on the pilot; an inverted verify assertion). **Both were caught by probing the guard
rather than counting it** — and neither could have occurred in flow B, because flow B has no
guard to get wrong.

## The differentiation question (this is what couples to RFC_037)

- Under **flow A**, the register's positioning reaches the build through the hand-written
  mission. `RFC_037` (classifier reads the register) is belt-and-braces.
- Under **flow B**, nothing carries it. The classifier sees a domain string and a short prompt,
  and — measured — seven finance sites already collapse to two classifications with `industry`
  null on all seven. **`RFC_037` stops being an improvement and becomes the only mechanism that
  would keep 140 sites from converging.**

## The options, honestly

**1. Flow A for everything.** Safest, proven twice, ~100 hours of drafting, and it is the
status quo. Rejected by scale rather than by quality.

**2. Flow B for everything.** Two minutes a domain, and today it ships sites with no claims
verification. **Not acceptable as it stands** — but the gap is a *seeding* gap, not a flow gap.

**3. Flow B plus an automatic seed — the recommendation.** Make the seed part of the pipeline
rather than a hand-written SQL file: `domain-submitter` (or a step beside it) writes a default
`evidence_base` (governing rule + the standing finance `banned_claims` + an empty facts roster),
an `imagery_style_guide` derived from `design_intent`, and enforces a contact email. That is
exactly what this lane has been doing by hand twice, and both times the *content* was
boilerplate — only the compliance specifics differed. **Then flow B has flow A's guards at flow
B's cost**, and `RFC_037` supplies the differentiation the mission used to carry.

Option 3 makes RFC_037 a precondition rather than a nicety, and adds a second small piece of
work (auto-seed). It is the only option that is both safe and scalable, and both pieces are
architecture-scope because they touch shared seams.

## What I would do next, on approval

1. Decide flow (recommend 3).
2. RFC_037 through architecture review — it is already filed.
3. A second RFC (or the same round) for the auto-seed, since it is the same shared seam.
4. Unlock the two halted sites and resume build #1 from `needs_strategy`, which is preserved.
