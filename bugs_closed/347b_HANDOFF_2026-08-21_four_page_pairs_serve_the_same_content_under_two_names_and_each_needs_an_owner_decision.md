> **⚠ NUMBER COLLISION — there are two 347s, resolve by SLUG not by number.** This one is
> the twin-pairs file (`…_four_page_pairs_serve_the_same_content_under_two_names…`). The
> other is `bugs_closed/347_HANDOFF_2026-08-21_webdesign_couk_head_component_is_a_fragment_with_no_head_element.md`,
> filed and closed by another lane the same day. This file was itself renumbered from 346
> after a *different* session took that number two and a half minutes before me — three
> sessions filed into the same two numbers within an hour, which is what this tree does.
> `git log` the FILE PATH, never the number.

# 347 — four page pairs serve one page under two live names; the minting is fixed, the existing duplicates are not, and each pair needs an owner decision rather than a fix

**Filed 2026-08-21**, spun out of `bugs_closed/215` when that bug closed. 215's
defect — the framework *minting* a second identity for one page — is fixed, live
and behaviourally proven. **This file is the damage that predates the fix**, and
it is deliberately separated because what remains is not engineering: it is four
per-pair decisions only the owner can make.

Continues the "O2" remediation begun in the brochure lane. **Three of the original
seven pairs are resolved** (`gripper-payload-calculator` and
`gripper-cycle-time-estimator` on robot-hands, completed 08-14/08-15 with full
retraction; one on ai-agent-orchestration.com no longer presents as a pair). **Four
remain.**

## The four, measured 2026-08-21

All eight URLs HTTP-tested the same day: **all serve 200 with real content**, against
a fabricated-URL 404 control of 2,886b on robot-hands. So every one of these is a
live page a visitor or a search engine can reach — not a phantom.

| site | pair | components | policy | bytes served |
|---|---|---|---|---|
| finetuning.uk | `ai-readiness-quiz` `/ai-readiness-quiz.html` | 2 | generic | 39,551 |
| | `tool-ai-readiness-quiz` `/tools/tool-ai-readiness-quiz.html` | 1 | **owned** | 36,367 |
| fundamentallyai.com | `automation-savings-estimator-guide` `/blog/…` | 3 | generic | 27,854 |
| | `tool-automation-savings-estimator-guide` `/guides/…` | 3 | generic | (200, content) |
| fundamentallyai.com | `model-approach-selector-guide` `/blog/…` | 3 | generic | 27,010 |
| | `tool-model-approach-selector-guide` `/guides/…` | 3 | generic | 32,934 |
| robot-hands.com | `matchmatrix` `/matchmatrix.html` | 4 | generic | 79,520 |
| | `tool-matchmatrix` `/tools/matchmatrix/index.html` | 1 | **owned** | 100,552 |

> ⚠ **Do NOT choose the survivor by component count.** `tool-matchmatrix` has **one**
> component and serves **more** bytes than its four-component twin. A `page_components`
> count is a container count, not a content measure — the landmine is on file, and this
> pair is a textbook instance of it. Open the served page.

> ⚠ One fetch returned `HTTP 000` on the first attempt and **200 on retry**. A single
> failed curl is not evidence a page is dead; re-fetch before recording an absence.

## What the decision is, per pair

**Which name owns the page**, and therefore which URL survives. That is all. The
framework will not choose: when both spellings are realised and both are in the plan,
every twin-identity layer deliberately REFUSES (register `PLAN-048`'s landmine), because
snapping would hand the writer two entries with one name and the dedup would then evict a
live page.

**The constraint that should be read BEFORE choosing, not after:**

- **There is no redirect mechanism on this estate.** `link_registry` and `redirects` are
  empty fleet-wide. So retiring one name of a pair makes that URL a **permanent 404** for
  anyone holding the link — a bookmark, an inbound link, a search result. This is the
  single fact most likely to change a decision, and it is why this is an owner call.
- **`finetuning.uk` is doubly constrained.** It is on the decomposed-site list
  (`bugs_open/204`), so its pages carry positional slot names and it must not be replanned
  until 204 is fixed. Its pair is the one to leave until last.
- **An archive is durable now, but was not.** `bugs_closed/266` (four producers rebuilding
  archived pages) is fixed, live and behaviourally proven, so archiving one side now holds
  — that was the blocker when this remediation stalled, and it is discharged.
- **Retraction is a separate step from archiving.** Archiving removes the page from the
  rebuild population; it does **not** delete the served file. The worked, twice-validated
  procedure is the runbook's 8 steps, and step 6 is the `delete_file` retraction.

## How to execute a decision once made

`docs/agent_docs/docs024_key_docs_latest/brochure_component_library/RUNBOOK_2026-08-11_duplicate_page_identity_remediation.md`
— eight steps, proven end-to-end on two pairs. Its two procedural corrections are load-bearing:
"open" work items means `workItemClosedStatuses` (not `workItemTerminalStatuses`), and step 4
is not durable alone because the fleet sweep re-queues a rerender per **active** page.

**Run the inbound census before mutating anything.** It carries its own positive control and
has correctly predicted both a refusal and a clean pass — it is the step that tells you whether
retiring a name will break a link that something else on the site depends on.

## How to verify this file is finished

Zero rows from the pair census (both sides `active` + `deployed_at IS NOT NULL` sharing a
role-prefix-stripped stem), **and** each retired URL HTTP-tested to 404 at the site's control
byte size, **and** the survivor unchanged to the byte. Re-run the census before believing a
zero: this file's own population dropped from 7 to 4 partly by remediation and partly by
causes nobody recorded.

## Relations

`bugs_closed/215` (the minting defect, fixed — this is its residual damage);
register `PLAN-048` (the seam that refuses these pairs by design, and why);
`bugs_open/204` (constrains finetuning.uk); `bugs_closed/266` (archive durability, the
discharged blocker); `bugs_open/098` → `RFC_011` (retiring a live page has no mechanism —
the reason a redirect does not exist).

---

## CORRECTION 2026-08-21, before anyone acted on this file — **NONE OF THE FOUR ARE DUPLICATES.** This file's own title is wrong, and "which name survives" is the wrong question

Written the same day, after opening all eight pages instead of trusting the pairing. The
table above is accurate about the *names*; it is wrong about what the pages **are**.

| pair | what one side actually is | what the other actually is |
|---|---|---|
| fundamentallyai `model-approach-selector-guide` | `/blog/…` — *"How the Model Approach Selector weighs fine-tuning, RAG, and prompting"*: a guide **to the tool**, walking through what it asks and why | `/guides/tool-…` — *"Prompting, RAG, or fine-tuning: a decision guide that starts with your actual use case"*: a **standalone topic guide** that barely mentions the tool |
| fundamentallyai `automation-savings-estimator-guide` | same shape: `/blog/…` explains the estimator | `/guides/tool-…` is the topic guide |
| finetuning.uk `ai-readiness-quiz` | `/ai-readiness-quiz.html` — a **5-question** readiness assessment ("Answer five questions… 0 / 5") | `/tools/tool-…` — a **6-question** quiz ("Answer 6 quick questions… Question 1 of 6") — a *different instrument* |
| robot-hands `matchmatrix` | `/matchmatrix.html` — the overview/pitch page for MatchMatrix | `/tools/matchmatrix/index.html` — the **working selection matrix** itself |

**So retiring either side of any of these destroys live, distinct content.** The remediation
this file was filed to schedule must not be run as written.

### Why they were paired, and why that is not a bug in the pairing

`PageItemStem` strips a leading `tool-`, so `matchmatrix` and `tool-matchmatrix` reduce to
one stem and the census reports them as a pair. That census is a **name** census and it
never claimed to be a content census — the estate's own documentation says the stem key
"is the one layer that can pair two genuinely different pages", which is exactly why
`stem_twin_snap` is separately gated and off by default. The pairing did its job; **the
error was mine, in reading a name-shaped signal as a content-shaped conclusion** and
writing a file that asked "which survives".

**Two things this invalidates in the record above:** the component counts and byte sizes in
the table are real but were offered as *survivor-selection evidence*, which they never
were; and the framing "one page under two live names" is false for all four. What remains
true: all eight URLs serve 200, and there is no redirect mechanism.

### What the real question is, per pair

Not "which name survives" but **"are these two pages the site should have, and are they in
the right places?"** — which is an editorial question, not a remediation one. Concretely:

- **fundamentallyai (both):** two legitimate guides each, split across `/blog/` and
  `/guides/`. The owner's instruction of 2026-08-21 — *"those guides should be under the
  guides directory"* — is about that split and is satisfiable **without retiring anything**:
  move the `/blog/` guide under `/guides/`. Cost: the old `/blog/` URL 404s permanently
  (no redirect mechanism). Residual oddity if only that is done: the two pages already at
  `/guides/` carry a `tool-` prefix on a `blog-post`, which is a canonicalisation artefact
  and reads badly in a URL — renaming them is a second, separate decision that costs two
  further 404s, and it needs new slugs because their bare names would collide with the
  pages being moved in.
- **finetuning.uk:** two different quizzes (5 vs 6 questions). Somebody has to decide
  whether the site wants both, and that is a product question. Also still constrained by
  `bugs_open/204`.
- **robot-hands:** a pitch page and the tool it pitches. Arguably correct as-is; the only
  defect is that they share a stem and so keep appearing in twin censuses.

### The transferable lesson, and it is not "check the content"

The pairs census, the component counts, the byte sizes and the HTTP statuses were all
**measured, dated, and true** — and together they still supported the wrong conclusion,
because every one of them measured the *container* and none of them measured the *page*.
The estate already has a landmine saying a `page_components` count is a container count and
not a content measure; this is the same failure one level up, where **an entire evidence
table can be container-shaped**. The check that broke it was the cheapest available and I
did it last instead of first: open the page and read the first paragraph.

---

# CLOSED 2026-08-22 — NO ACTION, by owner decision. This file's value is now the correction it carries, not the work it proposed

Owner decision, 2026-08-22: **leave it.** Closed with nothing done to any of the eight
pages, deliberately and on the record.

**Why that is the right outcome rather than a deferral.** This file was opened to schedule
remediation of "four page pairs serving one page under two live names". That premise is
false — see the 2026-08-21 correction above. All eight pages are distinct, live and
serving:

- **fundamentallyai.com** ×2 — a guide *to the tool* at `/blog/…` and a standalone *topic
  guide* at `/guides/tool-…`. Different articles.
- **finetuning.uk** — a **5**-question readiness assessment and a **6**-question quiz.
  Different instruments.
- **robot-hands.com** — the MatchMatrix pitch page and the working selection matrix.

**There is no defect here to fix.** What remained were editorial questions, and the owner
has answered them: leave the pages alone. The one instruction that *did* land — *"those
guides should be under the guides directory"* (2026-08-21) — was superseded by the same
decision once it was clear that acting on it would cost two permanent 404s (no redirect
mechanism exists) to move articles that are each legitimately where a reader can find them.

## What a future reader must NOT do

**Do not re-open this from a twin census.** These four will keep appearing in any
stem-based pairing report for as long as they exist, because `PageItemStem` strips a
leading `tool-` and `matchmatrix` / `tool-matchmatrix` reduce to one stem. **That is the
census working correctly and it is not evidence of a duplicate.** The estate's own
documentation says the stem key "is the one layer that can pair two genuinely different
pages", which is exactly why `stem_twin_snap` ships gated and default-OFF.

If a later census raises these again, the answer is in this file: **they were checked, page
by page, on 2026-08-21, and they are not duplicates.** Re-checking costs one `curl` and a
read of the first paragraph — do that before proposing anything, which is the step whose
omission produced this file's original false premise.

## What this leaves behind that is worth keeping

1. **The minting defect is fixed** — `bugs_closed/215` (all three modes, live, behaviourally
   proven: 19 phantoms → 0 on a real replan) and `bugs_closed/340` (the preservation-set
   gap). Nothing in this file was ever blocking those.
2. **Three of the original seven pairs WERE genuine** and were remediated in full by the
   brochure lane in August (two on robot-hands with retraction verified at the artefact).
   So the pairing signal is not worthless — it is simply a name signal, and it needs a
   content check before it becomes a decision.
3. **The lesson**, logged in `WRONG_CALLS.md` (2026-08-21): an entire evidence table can be
   container-shaped. Component counts, policies, byte sizes and HTTP statuses were all true
   and all measured the container; none measured the page, and their agreement with each
   other read as corroboration when it was only common blindness.

**Status: CLOSED, no action taken, no page altered.** The four pairs remain live and
serving exactly as they were.
