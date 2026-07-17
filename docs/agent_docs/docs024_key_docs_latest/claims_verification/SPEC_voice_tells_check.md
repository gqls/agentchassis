# SPEC — voice_tells: flagging LLM-sounding copy (deterministic lane + prose lane)

**Purpose.** A checker that flags site copy that reads machine-written, per site, against that
site's own voice spec. Companion to the claims-verification layer and owned by the same
thread: claims answers "is it true?", voice_tells answers "does it read like a person wrote
it?". Requested by the owner 2026-07-17 after the leopardess plain-voice v2 decision.

**Status:** T0+T1 BUILT + TESTED 2026-07-17 (same day): engine `datahelpers/voicetells.go`, corpus tests green, `discovery_checks/check_voice_tells.go`, `cmd/voicescan`, leopardess voice_gate seeded. First live scan: 111 findings/85 components; v2 pages clean, v1-dense pages flagged — calibration verified. Deploy ships with next chassis image; then enable `voice_tells` in quality-discovery-agent checks (T2). T3 (V3 prose lane) + T4 (auto_rewrite) remain.
**Home:** this directory (`claims_verification/`) — it reuses the claims scan engine and the
same check→work-item→HITL grammar. Origin material:
`docs/leopardessconsulting/specs/PLAIN_VOICE_v2.md` (the keep/reject lists are the seed
rule-set) and `specs/VOICE_REWRITE_PROMPT.md` (v1 tells catalogue).

---

## 0. One-paragraph version

No agent in the platform flags LLM-sounding prose. `content-quality-auditor` checks tone
*alignment with the brief*; the 38 discovery checks are structural; nothing measures the
tells (banned phrases, em-dash rhythm, reflexive triads, uniform long sentences, zero
contractions). Meanwhile every site now has (or can have) a `voice` spec that *defines* its
register — leopardess's was just rewritten to v2 with quantified rules. This spec wires the
two together: a deterministic `voice_tells` discovery check that scores rendered copy against
the site's voice spec (regex + arithmetic only, V1-claims-style), plus a voice dimension in
the planned claims V3 prose auditor for what regex can't hear. Findings go to a human, or —
uniquely for this checker — can feed the PROVEN safe fixer: the framework rewrite path
(`page-build-handler` + `spec.suggestion` with the site's approved rewrite prompt).

## 1. Origin evidence (why this earns a build)

1. **The owner's original punch-list #10:** "Voice still reads LLM-written" — a human had to
   notice it, page by page. Twice (v1 pass, then the v2 re-direction).
2. **The tells are enumerable.** The leopardess journey produced a concrete catalogue:
   the balanced triad ("observability, fault isolation, and cost controls" — shipped, twice,
   on one page), "not X but Y" strawmen, em-dash asides as rhythm, summarising flourishes,
   title-case marketing headlines ("Production-Grade Multi-Agent AI Systems for UK
   Engineering Teams"), zero contractions, 30+ word packed sentences. Every one of these is
   regex- or arithmetic-detectable.
3. **The fixer already exists and is proven.** The v2 rollout (2026-07-17) rewrites pages
   through page-content-writer with an owner-approved prompt in `spec.suggestion`, gated by
   the claims layer. A voice checker's findings have a safe, existing remediation path —
   rare for a new check.
4. **Fleet-relevant.** Every chassis site ships LLM-written copy; leopardess is just the
   site where a human was watching.

## 2. What exists today (verified 2026-07-17)

| Agent/check | What it does | Voice-aware? |
|---|---|---|
| content-quality-auditor | tone *alignment* vs brief, gaps, CTAs (1 LLM call) | Alignment only — no tells concept |
| visual-design-auditor / component-quality-auditor | visual/structural | No |
| 38 discovery checks | structure, links, images, claims | No prose style |
| `site_specs` aspect `voice` | per-site register definition (leopardess: v2, quantified) | Defined but UNREAD by any checker |

## 3. Design (mirrors the claims layer deliberately)

### 3a. Deterministic lane — discovery check `voice_tells`

Pure Go over rendered text nodes (**reuse the claims engine's text extraction** in
`platform/orchestration/datahelpers/claims.go` — it already strips tags/attributes and
concatenates inline elements, solving the `<strong>`-split problem the hard way once).

Per page, compute and threshold:

| Signal | Method | Default trip |
|---|---|---|
| Banned phrases | site voice spec `banned_language` + a global AI-tells list ("dive into", "unlock", "leverage", "seamless", "in today's … landscape", "Whether you're", "not just X, but Y" regex) | any hit |
| Em-dash density | count / 1,000 words | > 3 |
| Reflexive triads | `X, Y(,) and Z` balanced-list regex, density per page | > 4 |
| Sentence length | share of sentences > 25 words; mean length | share > 30% or mean > 22 |
| Contraction ratio | contractions per 100 sentences (a plain-voice site with ~zero reads stiff) | 0 on a v2-register site |
| Summarising flourish | paragraph-final sentences opening "That is why / Ultimately / In short / And that is" | any hit |
| Title-case headings | headline caps pattern on non-title pages | per-site flag |

Output: a per-page score + the specific offending snippets (like claimscan's findings), so a
human sees *what* tripped, not a number. Thresholds live in a `voice_gate` block on the
site's `voice` spec — per-site, tunable, opt-in by presence (sites without a voice spec get
the global banned list only). Long-form pages (blog posts) get their own thresholds — essay
rhythm differs from landing copy.

### 3b. Prose lane — a voice dimension in the claims V3 auditor

The V3 claims auditor (one LLM call per page, discovery cadence) gains a second scoring
dimension: "does this page conform to the site's voice spec?" with the spec text injected.
Catches what regex can't: overall density, performance-register, hollow fluency. Findings
merge into the same work item, clearly labelled by lane.

### 3c. Remediation

- Default: `voice_tells` work item → **needs_human_review** (HITL, like claims).
- Optional per-site flag `voice_gate.auto_rewrite: true`: the handler fires the framework
  rewrite (`page-build-handler` + the site's stored approved rewrite prompt in
  `spec.suggestion`). Everything still passes `validate_page_content` including the claims
  gate. Ship this OFF; leopardess can pilot it after the v2 rollout proves the prompt.
- Build-time: add the deterministic scan to `validate_page_content` as **warning** severity
  only (style is softer than truth — never block a build on voice in v1 of this check).

## 4. What this checker must NOT do

- **Not detector-evasion.** The reference material behind v2 included AI-detector tricks
  (deliberate errors, slang, forced casualisms). The owner rejected those and so does this
  spec: the target is *readable and honest*, not "undetectable". No signal may reward errors.
- **Not rewrite anything itself.** Same rule as claims: flag → human (or the gated framework
  fixer). The checker never edits content_data.
- **Not judge quoted/third-party text** — exempt `<blockquote>`, quoted testimonials
  (when real ones exist), and code/tool UI strings.
- **Not run on locked components** (`locked_at`, per precedent).

## 5. Benchmark corpus (from leopardess history — all real)

| # | Text | Expected |
|---|---|---|
| V1 | services hero pre-fix: "…with observability, fault isolation, and cost controls — is an architecture problem…" | TRIP (triad + em-dash + 40-word sentence) |
| V2 | old index title "Production-Grade Multi-Agent AI Systems for UK Engineering Teams" | TRIP (title-case + hype compound) |
| V3 | v1-dense homepage hero ("Most of what we build is unglamorous, and that is the point…") | TRIP on flourish/sentence-length (this is the interesting calibration case — honest but dense) |
| V4 | v2 homepage hero ("We build systems that take over repetitive work. Each one has a clear job…") | PASS |
| V5 | who-we-help v2 cards ("A list that has to be checked against another list…") | PASS |
| V6 | A page of deliberate slang/typos (synthetic) | must NOT be rewarded — style errors are not the goal; PASS/FAIL on tells only |
| V7 | blog post hierarchical-guide (long-form, post-claims-fix) | PASS under long-form thresholds |

Acceptance per phase = corpus table in the notes, same convention as claims B1–B7.

## 6. Build order

1. **T0** — `voice_gate` config block schema on the `voice` aspect; leopardess seeded from
   PLAIN_VOICE_v2.md (one sitting).
2. **T1** — deterministic `voice_tells` discovery check + claimscan-style CLI mode; run
   read-only against leopardess + one other site; calibrate thresholds on the corpus.
3. **T2** — enable in quality-discovery-agent's checks array (HITL findings only).
4. **T3** — the V3 auditor voice dimension (build with claims V3, one auditor, two lanes).
5. **T4** — optional auto-rewrite handler behind `voice_gate.auto_rewrite` (leopardess pilot).

## 7. Landmines (inherited knowledge — do not relearn)

- Tag-split phrases: reuse the claims text-node extractor; literal greps miss
  `<strong>`-split text in both directions (proven 2026-07-16).
- Findings must carry snippets + locations, not just scores (claimscan precedent).
- Writer rebuild as fixer replaces the WHOLE page's llm fields and follows `pages.sections`
  — prune fabrication-prone sections from plans first (leadership-team / testimonials /
  case-studies-grid / system-stats class), and the claims gate must be live before
  auto_rewrite is ever enabled anywhere.
- content-gap-planner (2026-07-17) invented a full fake case study WITH metrics when asked to
  fill a content gap — any planner this checker's findings feed must not be allowed to "fix"
  voice by inventing new content. Rewrite guidance only, never gap-filling.
