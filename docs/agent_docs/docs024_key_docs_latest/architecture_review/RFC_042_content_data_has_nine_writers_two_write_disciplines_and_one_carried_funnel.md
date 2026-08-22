# RFC 042 — `page_components.content_data` has nine writers, two write disciplines and one carried funnel; the split stays reachable and needs an owner decision

**Filed:** 2026-08-20 by the `bugfix_238_regeneration_key_loss` lane, discharging the council
gate's `architecture` seat instruction on PBP-039 (correlation `bd38df2e`, APPROVED with 8
advisory objections), which returned `needs_rfc` and said:

> *a new resolution tier on a function shared by two fleet-wide consumers, plus a deliberate
> cross-path semantic change, is RFC-shaped rather than bug-patch-shaped — and noted that **this
> compensates for the REPLACE-vs-MERGE split rather than closing it, so the divergence stays
> reachable by future writers**. Routed for a human call, with the underlying asymmetry named as
> the thing an RFC would actually be about.*

That routing never produced a file. The 238 lane's own summary recorded the same conclusion and
declined it as *"above this thread's pay grade"*. **This is that file, eighteen days later, and the
delay is itself evidence: in the interval the divergence WAS re-reached** — `bugs_open/268` lost 214
CTA anchors across 19 sites through a branch the carry did not cover.

**Status:** ~~DRAFT — no code proposed here. The decision asked of the owner is at §6.~~
**DECIDED 2026-08-22 — OWNER RULED option (c).** See the decision record at the end of §6.

> **UPDATE 2026-08-22 — option (c) is now scoped and its population is measured:
> [`bugs_open/355_HANDOFF_2026-08-22_eight_of_nine_content_data_writers_cannot_be_observed_losing_keys.md`](../../../../bugs_open/355_HANDOFF_2026-08-22_eight_of_nine_content_data_writers_cannot_be_observed_losing_keys.md).**
> The headline changes the decision at §6 and should be read before taking it: pairing every archived
> in-place write (the 380 transitions belonging to the eight uncarried writers) and diffing
> schema-declared non-LLM keys returns **ZERO losses** — against a demand control, the identical query
> over the funnel population, which returns **72**. So the query can see losses and there are none
> here. What keeps 354 open is not damage but four quantified blind spots in the instrument, the
> sharpest being that **the archive cannot name the writer at all** (`application_name` is the pgx
> connection default), which is precisely the question §4.3 says the detector exists to answer.

**Answer this together with [`RFC_008`](RFC_008_a_mandatory_write_seam_for_page_components_rendered_html.md)**
(*"a mandatory write seam for `page_components.rendered_html`"*, open since 2026-08-02). It is the
same question about the sibling column of the same table, it has already collected four seats saying
the same thing, and its strongest counter-argument (§4.2 below) applies here almost unchanged. Two
RFCs proposing two seams over two columns of one table is how a codebase acquires three disciplines
instead of one.

---

## 1. What is actually true today, in plain terms

A page's sections live in `page_components`. Each row carries `content_data` — a JSON object of the
field values that section was built from. Some of those values are **written by a language model**
(the prose). Others are **resolved from elsewhere in the system**: an image URL from the site's
assets, a link target from a spec, a contact e-mail, a listing built by a query. The component's
`input_schema` says, per field, which kind it is.

There are two ways a page's rows get rewritten, and they treat those resolved values differently:

- a **re-render** takes what is stored, resolves what it can afresh, and **merges** the two;
- a **regeneration** deletes the rows and inserts new ones — it **replaces**.

So a resolved value whose source has gone quiet survives every re-render and dies on the first
regeneration. That is not a hypothesis; it is why `bugs_open/238` happened. Eleven URL values had
been byte-identical on finetuning.uk's homepage since 2026-05-01 because every run in between was a
re-render. The first true regeneration shipped five `<img src="">` and silently deleted six controls.

**PBP-039 fixed that, and only for one funnel.** `planSection` now falls back to the page's own
deployed row when a non-LLM source resolves nothing. It sits at the one point the build and
re-render paths share, so those two cannot drift from each other. It does nothing for any other
writer, and it repairs nothing already broken.

## 2. The census — nine writers, and the carry protects one funnel

Measured 2026-08-20 (grep for `INSERT INTO page_components` / `UPDATE page_components` setting
`content_data`, non-test, plus the admin handler whose SET clause is built dynamically and so does
not appear in a literal scan — that one is worth noting on its own, because a census by SQL literal
would have reported eight and been wrong):

| writer | file | discipline | archives first? |
|---|---|---|---|
| the plan→save funnel | `save_page_sections_action.go` (DELETE + INSERT) | **REPLACE**, carried by PBP-039 upstream | yes (`page_component_history`, pre-DELETE) |
| section editor — field updates | `section_editor_actions.go` | MERGE | trigger only |
| section editor — replacement mode | `section_editor_actions.go` | **REPLACE** | trigger only |
| component swap | `section_editor_actions.go` | REPLACE | trigger only |
| dossier / report pages | `create_report_page_action.go` | REPLACE | trigger only |
| blog listing rebuild | `rebuild_blog_listing_action.go` | REPLACE | trigger only |
| ported pages | `adopt_verbatim.go` | REPLACE | no |
| tool widgets | `deploy_tool_action.go`, `create_tool_component_action.go` | REPLACE (writes `'{}'`) | no |
| admin component edit | `internal/core-manager/admin/page_admin_handlers.go` | REPLACE (whole object from the request body) | **yes**, explicitly |
| CLI importer | `cmd/webdesignport/import.go` | REPLACE | no |

**One funnel is carried. The rest are not, and most of them replace.** Whether that is a live
exposure or merely an untested one is the question §4 turns on, and the honest answer today is that
**nobody has measured it** — which is itself an argument for the detector option rather than the
guard option.

## 3. Why the compensating layer is not the same as closing the split

PBP-039's own register entry states the residual in its landmine, and it is worth quoting because it
is the clearest statement anyone has written of the current position:

> *after this change the only thing making the build path lossless is the carry inside
> `planSection`/`handleMissingField`. Do not remove it, do not reorder it ahead of the alias chains,
> and do not assume the rerender merge covers the build path: it never did, which is the bug.*

Three properties of that position matter for the decision:

1. **It is a convention, not a constraint.** A new writer of `content_data` that does not route
   through `planSection` inherits nothing. Nothing tells its author that.
2. **It has already failed once past the fix.** `bugs_open/268`: the field loop's own
   "resolved at render time" branch `continue`d before the carry could be offered — a branch older
   than the carry, which became a trap when the carry arrived and nobody revisited it. 214 anchors,
   19 sites, and every one of five green checks blind to it because they were assembled from the
   author's intent (`WRONG_CALLS` 2026-08-12).
3. **It is currently masking a different defect.** `bugs_open/312`: `select_sections` reads the
   resolver's output at a wrong jsonb path, so a fresh build uses the pre-resolver plan and the
   values that survive are *the carry re-shipping the stored row* — 26 of 26 resolver-minted
   `*_target_title` values discarded, fleet-wide, still live (the fix, `477`, is HELD on a
   dangerous interlock). **A safety net that is silently load-bearing is not a safety net; it is an
   undocumented dependency.**

## 4. The options, costed

### 4.1 (a) Status quo — the carry plus a registered landmine

**Blast radius** nil. **Who must be told** nobody. **Cost of not acting:** the divergence stays
reachable, and it has already been re-reached once *after* the landmine was written — so the
evidence that documentation bounds this class is negative, not absent. This is the option the estate
is on by default if this RFC is not decided.

### 4.2 (b) A shared write seam every producer must call

One `persistComponentContentData(...)` that preserves schema-declared non-LLM keys from the row it
replaces, unless the caller explicitly declares replace-all.

**Blast radius** 9 files, 11+ write sites, and it converges with RFC_008 — the same table, the
sibling column, the same shape of proposal. Build one seam or admit two.
**Who must be told** every writer above, by name, plus RFC_008's owner.
**Shape** Go, inert until a roll; adoptable one writer at a time, each conversion its own
council-visible commit; per the owner's 2026-08-02 ruling the merge-by-default flip on each existing
writer is the authority question, so recommend per-writer defaults that match today's behaviour and
flip them by explicit follow-up.

**The argument against, stated at full strength because RFC_008 already made it and it is good:** a
mandatory seam needs an opt-out for the writers that legitimately replace — `adopt_verbatim` stores a
crawled document verbatim and hashes it; the tool writers deliberately write `'{}'`; the admin
handler is a human replacing a value on purpose. And, in RFC_008's own words, *"an opt-out parameter
is a lint allow-list wearing a type signature."* The gain over a lint is that the allow-list is
enumerable at compile time rather than diff-scoped — real, but smaller than it first looks.

### 4.3 (c) One unified content-loss detector

A single post-write invariant differ: for any `content_data` write, compare the schema-declared
non-LLM keys before and after; record a durable finding on loss; refuse only where a caller opts in
(unsafe default OFF, per the owner's 2026-08-02 ruling).

**This option has a standing instruction behind it.** `bugs_open/178` left a stop sign against
adding a sixth refusing floor to `save_page_sections`, and PBP-039's entry records the corollary:
*a sixth is the trigger for a unified content-loss detector as its own submission.* This is that
submission's shape.

**The embryo already exists**: `writeContentDataRegressionLog` (`save_sections_metadata_source.go`,
`bugs_open/194`) already records content_data regression at the save — but **page-level and
all-or-nothing**, so it could not have seen 238 at all: that page lost 11 of 58 keys while every
section still carried content_data. Extending an existing record to per-key resolution is a smaller
change than nine call-site conversions, and it answers the question option (b) cannot: **which
writers actually lose keys in practice.**

**Blast radius** one new mechanism rather than nine guards; readers of `agent_error_log`.
**Cost of not acting:** losses at non-funnel writers stay invisible, so option (b) would be sized
from inference rather than measurement — which is the failure mode `WRONG_CALLS` 2026-07-25 names as
*counting the population a fix is ABOUT instead of the population it would ACT on.*

### 4.4 (d) Make `save_page_sections` itself merge non-LLM keys — **stays rejected**

Previously rejected by the `inline_guide_imagery` lane, and the rejection is upheld — but one of its
grounds was mis-aimed and the correction matters, because the mis-aimed version is the one a future
reader would use to reopen this.

> *"merge of two versions of prose is ill-defined; the writer legitimately restructures articles
> (238's card rewrite is the worked proof)."*

That is true **of LLM fields, which nobody proposes merging.** The proposal here was only ever about
non-LLM keys. The grounds that survive:

1. **It repairs the row after the HTML has already rendered without the value** — the page still
   ships `src=""`, and `rendered_html_digest` then claims a reproducibility that no longer holds
   (STY-056 / `bugs_open/229`).
2. **`bugs_open/178`'s stop sign**: five refusing floors already sit in that function; a sixth is
   authority taken on a prediction.
3. **It covers nothing option (b) does not.** The other eight writers do not route through this save.

### 4.5 (e) Declare the plan the sole complete producer, and register it as a contract

Finish PBP-039 — close the resolved-side holes in `planSection` where the carry is never offered
(two are filed for independent diagnosis as of 2026-08-20, run `68b3f9b6`) — and then register the
contract: *a `content_data` row is born complete at plan time; a writer composing a row from partial
data must consume `planSection` rather than re-deriving its own merge.*

**Blast radius** `plan_sections_action.go` plus a register entry. **Cost of not acting:** without the
contract, (e) decays into (a) — it is only a decision if it is written where the next writer's author
will meet it.

## 4.6 Evidence gathered AFTER filing, which strengthens (e) and weakens the case for urgency

Added 2026-08-20, same day, from the `090` run this RFC's filing lane fired
(`68b3f9b6-1674-41a0-bc9e-c251192daaa1`, verdict UNVERIFIABLE) and the query it asked for.

**The one thing that IS measured about this seam's real-world behaviour, and it is good news.**
Pairing consecutive archived generations of the same (page, slot) and counting fields that went
non-empty → blank: **66 non-LLM losses, 11 sites, all `renderer`/`static`-sourced, all between
2026-08-11 and 2026-08-14 18:36 UTC, and none since** — against a demand control of **3,033**
archived generation-pairs in the six days after. `renderer`/`static` is precisely the class the
`bugs_open/268` carry extension closed (`8f899cc8d`, 2026-08-14 09:13 BST). ⚠ The window bounds it:
the trigger archive begins 2026-08-09, and the older `save_page_sections_overwrite` rows cannot
widen it because they carry no `slot_name`.

**What that means for the options.** The carried funnel is now *demonstrated* lossless on ordinary
fleet traffic, not merely argued to be. That does not close the split — the other eight writers are
still uncarried and unmeasured, which is exactly what option (c) exists to find out — but it does
say the urgency is lower than the 268 incident implied, and that (b)'s nine-writer conversion would
today be sized from inference rather than from a single observed loss outside the funnel.

**And the two gaps option (e) would close are REAL IN THE CODE AND UNOBSERVED IN PRODUCTION.** Both
were read directly in `planSection`: a blank resolved value beats a good stored one (the generic
branch stores whenever `found && value != nil`, and `resolveSpecPath`/`resolveSpecAlias`/the
`site_assets` lookup can each return a present-but-empty string), and a `query.*` resolver ERROR
leaves the key absent from `resolvedData` entirely. Neither has a single observed instance: **0**
loss events for `site_specs.*`/`site_assets.*`/`query.*` sources across the whole window, only
**2** empty-string spec values behind declared sources fleet-wide, and the `090` independently
observed that every record it could find says *"no previously-built row held a value, so there was
nothing to carry"* — the opposite of the hypothesised scenario.

They were therefore **recorded rather than shipped**, and that is the decision this section exists
to expose to the owner: three lines of Go and a mutation-proved test would have been cheap, easy,
and sized from a code reading rather than evidence. **What would justify shipping them:** one loss
event with one of those source families in the pairing query, or a `STRUCTURAL_KEY_CARRY_MISS` row
whose page and slot DID hold the value in the prior generation. The query is in the 238 lane's
RUNBOOK.

## 5. Not part of this RFC, deliberately

**Making `RenderTemplate` itself the reporting form** — the 8 unguarded call sites named in
`dead_url_guard.go`'s own header self-audit (3 of 11 guarded). The council already ruled that
RFC-shaped, and it is a decision about the *render primitive every path flows through*, not about the
*write seam*. Conflating them invites a split verdict on two independent questions. It also now
overlaps `RFC_041` (the render seam's new error contract), which is the right place for it to land.

## 6. The decision asked of the owner

Pick one, and say whether it is answered jointly with `RFC_008`:

- **(a)** accept the carry-plus-landmine as the ceiling, and record that the class is bounded by
  convention;
- **(b)** commission the shared write seam (and state whether it merges with RFC_008 into one seam
  over both columns);
- **(c)** commission the unified content-loss detector — record-only, refusal per-caller opt-in;
- **(e)** contract plus the `planSection` completion only.

**This lane's recommendation: (e) now, (c) next, (b) only if (c)'s findings show non-funnel writers
actually losing keys.** ⚠ **Re-read in light of `bugs_open/355` (2026-08-22):** (c)'s findings, as
far as today's instrument can produce them, are **zero** — so the case for (b) is now weaker than
when this was written, and the cheapest genuinely new thing in 354 is not the detector at all but
its candidate **A1**, a one-line-per-writer stamp that makes the archive self-attributing. That is
worth doing under option (a) as much as under (c). The reasoning is one sentence: **build the detector before the guard, because
the guard's population is currently an inference and the detector is what would measure it** — and
(e) is already in flight, so it costs this decision nothing to take.

> **DECIDED 2026-08-22 — OWNER RULED option (c)**, in the session working `bugs_closed/238`'s
> successor, on being shown this RFC's options and `bugs_open/355`'s scoping. The ruling: commission
> the unified content-loss detector — record-only, refusal per-caller opt-in — executed in the order
> `bugs_open/355` §4 prescribes: **A1** (writer self-attribution via transaction-scoped
> `application_name`) first, then **A2** (the per-key differ, extending
> `writeContentDataRegressionLog`) with **A3** (its consumer) in the same commit; **A4** (refusal)
> stays unbuilt until A2/A3 produce a population.
> ~~**Not ruled:** whether this is answered jointly with `RFC_008` — the owner named option (c) only,
> so RFC_008 remains open and nothing here decides the `rendered_html` seam.~~ Implementation record:
> `bugs_open/355` (per its §8).
>
> **THE JOINT HALF WAS RULED LATER THE SAME DAY: yes, jointly — and `RFC_008` is ANSWERED with NO
> MANDATORY SEAM.** Both columns of `page_components` therefore run one discipline, stated once:
> **detect and attribute, do not refuse.** `RFC_008`'s decision record carries the reasoning and its
> four reopen triggers. So the concern this file opened with — *"two RFCs proposing two seams over
> two columns of one table is how a codebase acquires three disciplines instead of one"* — is
> discharged: there is one posture over both, and the guard on either column is now conditional on a
> detector producing a population, which on `content_data` is measured at zero.

> **IMPLEMENTED 2026-08-22, same day (see `bugs_open/355` §10 for the full record).** One deviation
> from the ruling's letter, recorded here because this block is what a future reader will quote:
> A2 was NOT built by extending `writeContentDataRegressionLog` — that function sits in the FUNNEL
> (the save path), so per-key-ifying it upgrades the one writer PBP-039 already protects and never
> sees the eight uncarried writers, which do not route through it. The shipped shape is a daily
> archive sweep (`cmd/content-loss-check`, PBP-046) over `page_component_history`, with migration
> `552` closing that archive's content-only blind spot and A1 (PBP-047) making its
> `application_name` column name the writer. Reader and detector are one binary (the A3 same-commit
> rule); refusal (A4) remains correctly unbuilt — first run measured the non-funnel population at
> zero against a 72-loss control, and 48 historical findings were graded healed and stamped
> resolved, the first resolved rows in `agent_error_log`'s history.

---

## Appendix — how this RFC's own figures were obtained

Recorded because the estate's rule is that a `[MEASURED]` figure is only evidence if it could have
come out otherwise, and two of these nearly did not.

- **"Nine writers"**: a grep for `INSERT INTO page_components|UPDATE page_components` setting
  `content_data` finds **eight**. The ninth — the admin handler — builds its SET clause dynamically
  and is invisible to any literal scan. **A census by SQL literal would have reported eight and read
  as complete.**
- **"The carry protects one funnel"**: `carryStored` is reachable only from `handleMissingField` and
  the renderer/static branch, both inside `planSection`; `plan_sections` is a step of the build and
  re-render agents only. Verified by reading the branches, not by counting call sites.
- **"28 carry-miss findings, no consumer"**: `SELECT count(*) … FROM agent_error_log WHERE
  error_code='STRUCTURAL_KEY_CARRY_MISS'` → 28, 2026-08-11 → 08-17. The disconfirming result would
  have been a non-zero count of anything reading them; there is no reader in Go and no discovery
  check covering the class.
- **"Still-live damage"**: verified at the served artefact rather than the row, because a
  `page_components` join is an upper bound on live damage, never a measurement of it
  (`WRONG_CALLS` 2026-08-08). `curl ai-agent-orchestration.com/index.html | grep -c 'src=""'` → 5.
