# Prompt audit — phase 1 findings (mechanical pre-scan), 2026-08-25

**What phase 1 is:** the ORDERING pass of `PLAN_2026-08-25_prompt_audit.md` — count what every
prompt DEMONSTRATES so the judgment pass reads the biggest teachers first. Tool:
`audit_prompt_demonstrations.py`; table: `PHASE1_2026-08-25_league_table.md` (599 strings scanned,
every row dated). **A low score is not a clean bill** — the pattern list's ceiling is measured
(REFRESH §3), and the wider-register columns are lexical proxies. Everything below is
`[MEASURED 2026-08-25]` unless marked.

## Finding 1 — what the writer actually reads: ~64 negation demonstrations per call

The per-call truth is the RENDERED prompt (population R), not any template. Three consecutive
`page-content-writer` calls minutes apart, ~48,000 chars each:

| per rendered writer call | count |
|---|---|
| `X, not Y` | 25–27 |
| `rather than` | 23 |
| negative-reveal (`isn't a…`) | 8 |
| `not just/only` | 5 |
| `instead of` | 2 |
| **negation demonstrations, total** | **63–65** |
| "plainly" | **14** |
| "honest…" | **10** |
| em dashes | 35–37 |

Volume: **6,452 calls in 30 days**, avg 26,696 chars rendered (`llm_call_log`). The 305 lane's
"16 demonstrations per call" (CONTRIB 08-20) was a partial count of the same object; the full
assembly is four times that. The demonstrations stack from four layers, each individually modest:

| layer | negation demos | where |
|---|---|---|
| writer template (`generate_content`, 14,898 chars) | 4 + "plainly" ×2 + candour-family ×9 | `agent_definitions` sub_workflow |
| house voice block (6,032 chars) | **17** (2.8/1k — the densest single prompt) | `agent_default_configs` `voice_style_block` |
| the site's brief `content_direction.formatted` | **12–31 per site**; 25 of 30 sites ≥10; median 19 | `site_specs` |
| per-field `llm_guidance` of the components in play | 0–8 per component; 30 of 140 ≥3 | `content_components` |

The house voice contains the rule against the tell written *as* the tell: *"Say what a thing IS
rather than what it is not."* Also *"apply the reason, not just the letter"*, *"persuaded rather
than sold to"*, *"a machine gun rather than a person"*. It is owner-approved v2 prose (08-14) and
it reads well as guidance — the evidence says guidance prose is also example (REFRESH §2).

## Finding 2 — the about-page premise the owner rejected was INSTRUCTED

The writer template carries, under **"STRICT RULE — NEVER PROMISE ACCURACY YOU CANNOT
GUARANTEE"** (migration `223_compliance_seat_overclaimed_reliability.sql`, owner direction
2026-07-26, carried by `599`), this remedy clause, verbatim:

> *"If the section calls for a statement about method, say what we DO — and with no recorded
> operating history that means ONLY how the content is sourced: we name our sources and their
> dates so a reader can check them — and, where it fits, say plainly that we can still be wrong.
> Where the copy describes an interactive tool, it must say the tool can give a wrong answer."*

Matched against what shipped:

| instruction text | homegarden.uk about.html (OWNER_REVIEW §4) / finetuning.uk |
|---|---|
| "we name our sources and their dates so a reader can check them" | headings *Sourced and dated*, *Sources named*, *Where the detail comes from*; the phrase he named, *"and names the source behind any figure it uses"* |
| "say plainly that we can still be wrong" | *Timing stated plainly*, *"say plainly"* (his: *"people just don't say that"*), *Why the plain answer matters more than a confident one*; finetuning's *"we'll say plainly if that changes"* |
| "Say 'we cannot tell you X'" (same block) | *What this site will not do* ×2, *What a comparison cannot settle for you* (comparisons page h-heading) |
| "say what we DO" as the whole content of a method statement | 14 of 17 about.html headings about methodology |

**Status of the claim:** the textual correspondence is measured; the CAUSAL claim is
`[INFERRED]`. Its test (phase 3): a section planned as "about/method" on a page, built with the
clause present vs. with the remedy replaced, compared at the served artefact. The ban half of
the rule (do not claim every figure is verified) exists for a real incident — oufe.com shipped
*"Every factual claim is sourced to a named, dated primary document"* — and is right; it is the
REMEDY half that teaches methodology self-description on every section of every page.

**The irony worth writing down:** a rule written to stop one site describing its own rigour has
instructed a milder self-description of rigour across the fleet. Prohibition displaced, again
(REFRESH §2).

## Finding 3 — the per-site briefs are the densest teachers fleet-wide

`content_direction.formatted` (the writer's only brief wire): **29 of 30 sites** demonstrate the
construction; **25 of 30 at ≥10**; the top five (vetcomparison, loancalculator,
mortgagecalculator, loanandmortgagecalculator, remortgagecalculator) at 26–31 each, at
1.3–2.7 per 1,000 chars — roughly double the platform prompts' density. homegarden.uk's own
brief: 14 negations + 6 candour-family hits. This is finetuning's local measurement (7–8 in its
brief) confirmed as the fleet norm, and it is why CQ-027's cron finds what it finds.

## Finding 4 — the tools that police the tell demonstrate it

- `copy-editor` (stage 2, this lane's own): 6 demonstrations in 3,779 chars (1.6/1k), including
  a heading in the shape — *"THE READER, NOT THE SITE."* — and *"rejected, not discussed"*,
  *"editing prose … not redesigning"*.
- house voice: Finding 1.
- (the 305 gate's re-ask prompt was designed to carry no rule text and no example precisely to
  avoid this — bugfix_305 CONTRIB 08-20 trap 3; it is the one prompt built on the right theory.)

## Finding 5 — `llm_guidance` carries a baked template family

30 of 140 guidance-bearing components demonstrate ≥3; a family of loanandmortgagecalculator tool
components each carries 5–8 `not just` — one template's phrasing copied into every sibling's
per-field guidance (the CQ-005 shape: site-specific copy baked into shared machinery). `product-specs`
8× `rather than`; `platform-comparison` 5×. These reach the writer as the strongest lever there
is (CQ-031: 76% vs 7%).

## Finding 6 — the top ABSOLUTE rows are reviewer prompts, and reachability demotes them

`feature-designer`/`council-gate`/`fix-proposer` review prompts score 23–25 negations and 83–117
em dashes each, but they write reviews, not site copy. They matter for the audit's question 6
only (their language defines what gets FLAGGED); they are not phase-2 priority.

## Census corrections, recorded so the next run does not repeat them

- **Population F added**: the house voice is NOT under any prompt-keyed path in
  `agent_definitions` — it is one `agent_default_configs` row, `config_name='voice_style_block'`,
  text at `config->>'text'` (`platform/voicestyle/voicestyle.go:35`). ~~The register's
  "`agent_default_configs.voice_style_block`" reads like a column and is not one~~ **CORRECTED
  same night: the register (CQ-022 :227/:230) says "row" and gives the exact query; what misled
  me was a sweep SUBAGENT's paraphrase of it as `agent_default_configs.voice_style_block` — a
  report is another doc, and I built a census on it without opening the entry.** The first run
  of the scanner missed the densest prompt in the fleet for that reason.
- **Population D is an UPPER BOUND**: the backtick regex captures spans between stray backticks
  in Go comments — the "30,906-char literal" in `v3_site_actions.go:7064` is a comment block.
  717,501 chars is the ceiling; a `go/ast` walk is phase 2's job before any Go figure is quoted.
- **Population E sized**: 5 strings in the workflow columns, 0 demonstrations — the census's
  "15 ILIKE rows" were plumbing.
- 30 plumbing strings (<100 chars) counted, not scanned.

## What phase 2 reads first (from the table, weighted by reach × volume)

1. `page-content-writer` `generate_content` template — the STRICT RULE block above, the two
   "plainly", the nine candour-family sentences.
2. The house voice block — every sentence written in the shape it forbids.
3. `content_direction.formatted` — the fleet's briefs, top five first; homegarden's for the
   owner's case.
4. `content-gap-planner` `plan_gaps` (759 calls/30d) and `build-site-planner` `plan_site`
   (rendered 87–118K chars, 34–59 negations, 119–140 em dashes) — the PLANNERS, because the
   premise layer (which sections exist, what an about page is FOR) is decided there.
5. `llm_guidance` top 30 components.
6. `copy-editor`'s own prompt.

## What this does NOT establish

Causation for Finding 2 (test named); anything about the wider register beyond ordering (the
proxies are proxies); Go literal volume (upper bound); and no "clean" verdict anywhere — the
owner's ear is above the pattern ceiling.
