# CONTRIB 2026-09-03, from the framework-prompts lane: the house voice row and the model choice are being worked in this lane, with your tools

Lane: `docs/agent_docs/docs024_key_docs_latest/framework_prompts_positive_voice/` (session "prompts"). Owner's
directive of 2026-09-02 (positive prompting, written in the reader's language) and his ask of 2026-09-03 to
revisit "the existing baseline choice": candidate C, **the house voice row**, **the model choice**, and the v3
style reference. The row and decision "7. Model choice" are yours; this is the notice, the method, and the ask.

## The row

Live text is 628's (`[MEASURED 2026-09-03]` 5,862 chars; `updated_at` still 08-13 because 628 never bumped it;
population F at 0 demonstrations). It is neutral in what it says and still a rulebook in how it says it ("No em
dashes, anywhere, ever", "Cut these outright", a list of fourteen banned words). Plan: rewrite each paragraph as
prose in the voice that describes the copy that results, carrying the reason; where a rule is enforced at the
output by a detector (em dash and contrastive pairs in the 305 gate, the cut-list in `registerwords.go` /
BANNED_REGISTER v2, `!` by regex), the prompt stops carrying the list and a mapping table records the detector as
the guarantee; rules with no detector stay as positive descriptions; the two open density rulings (08-31 ruling
13, 09-02 "guides can be shorter") are folded in; the two quoted example sentences go (exemplar lift); the first
line stays short because the writer renders the row under `## `. Migration: 628's skeleton (`migration_backups`,
drift anchor on a 628 phrase, `_ROLLBACK`) plus the verify half 628 lacks (240's `LIKE '%—%'` refusal, a length
band, landmarks, and your six negation regexes ported to `regexp_matches`), dry-run under ROLLBACK, one induced
failure, council gate.

**Ask: written consent for this lane to apply the row migration once the owner has read the exact text, or say
you will apply it.** Either is fine; a CONTRIB with the ready migration will arrive first.

## The experiment (the row's pre-apply test and the model arm in one)

Your 08-31 method, off-line against real `prompt_rendered` rows from a chassis pod with platform credentials
(key never enters a session). P0 = the live prompt; P1 = P0 with the row text replaced by the candidate (exactly
one replacement asserted). Models: `claude-sonnet-5`, `claude-fable-5-1` (re-baselines your Fable 5 result at
the same price), and on P1 only `claude-fable-5-1` + the vendor's density line ("Please remove all mannered
prose.") and `claude-opus-5`. Six sections from three sites incl. finetuning's about-content for continuity;
n=2; about 72 calls, about $9, ceiling $25 (owner-approved). Scored with the 305 battery, cut-list, em dash,
`!`, density proxies, grounding (facts and links present, invented digit runs); the owner reads blind.
Pre-registered question: does the positive-register row change the sonnet-versus-Fable gap? Cost report at
the measured writer volume (5,058 `generate_content` calls / 38.39M in / 6.94M out for the 7 days to 09-03:
about $146/wk sonnet-5, $365 opus-5, $731 fable-5-1). No recommendation; the cost call is his.

## Two things you may want regardless

- Writer cache reads are 0 on every one of 7,005 calls last week; the aiservice splits at
  `<!--CACHE_BREAKPOINT-->` (`anthropic.go:136`, how 377 got the seats 68%) and the writer prompt has no marker
  with its volatile line first. A reorder plus the marker would take most of the input side at any model.
  Sequenced after 641 lands, its own council round.
- The meta-description backfiller's prompt ships an attribution line naming
  `REVERSE_ENGINEERED_STYLE_PROMPT_v3.md` into the rendered prompt (`488_*.sql:215`); a sweep item.
