# NOTES — `bugs_open/175`, the page-role upsert seam

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-08-02, session 1 — picking the bug, and checking nobody else had it

`who-owns.py` was run on five candidates (098, 093, 085, 087, 084) and returned
**OWNED or recently active** for every one of them. That is not the script being
broken — it is the script being *broad*: it reports any workstream directory that
mentions the number, and on a tree with forty lanes almost every bug is mentioned
somewhere. **So its verdict line is a prompt to look, not an answer.** What
actually separated the candidates was reading the live `.jsonl` transcripts of the
22 sessions active in the last hour and counting each session's DOMINANT bug —
the number a session mentions 40+ times is the one it is working; a number
mentioned twice is a directory listing. 175 appeared in 5 transcripts and was
dominant in none.

`175` itself says **"OPEN, unowned. Filed as a finding with the census done, not a
fix."** Its own filing lane (`bugfix_081`) last committed on 2026-08-01 and moved
on to closing 081.

## The census in the bug file is INCOMPLETE, and it changes what "the fix" means

175 lists six `ON CONFLICT (site_id, name)` sites. Today's grep finds **eleven**
(nine `DO UPDATE`, two `DO NOTHING`). The five it does not list —
`site_db_actions.go:1141`, `create_blog_posts_action.go:219`,
`adopt_verbatim.go:470`, `cmd/webdesignport/import.go:182`, and the
`seed_content_sources` pair — turn out **not to change the answer**: every one of
the `DO UPDATE` arms among them already carries `page_type = EXCLUDED.page_type`,
so they are all in the *opposite* camp 175 describes, and 175's instruction "do
not resolve this file by making all six identical" applies to them unchanged.

Worth recording anyway, because a session reading 175 alone would believe the
census was exhaustive and it is not. `[MEASURED 2026-08-02]` by re-running the
bug file's own grep.

## Exposure — 175 left it `[UNMEASURED]` and asked for it before choosing a fix

Its suggested query is written around `LIKE '%-guide'`/`'%-report'` and returns 2
rows. That query is looser than the arms are, so I wrote one per arm — asking, for
each arm, "is there a page holding a name THIS arm would claim, under a different
type?":

```sql
SELECT 'guide-arm', s.domain, p.name, p.page_type, p.build_status
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE p.name LIKE '%-guide' AND COALESCE(p.page_type,'') <> 'blog-post'
UNION ALL SELECT 'tool-arm', ... WHERE p.name LIKE 'tool-%' AND page_type <> 'tool' AND name NOT LIKE '%-guide'
UNION ALL SELECT 'report-arm', ... WHERE p.name LIKE 'report-%' AND page_type <> 'report';
```

Four rows, **all `deployed`**: `robot-hands.com/gripper-selection-guide`,
`robot-hands.com/selection-guide` (both `content`), `idea.uk/report-example`,
`lendzy.co.uk/report-loan-shark`.

**And then the honest part, which took a second look to see.** The guide arm's
name is `<tool page name>-guide`, and tool page names are canonicalised to
`tool-<slug>`, so the reachable collision is `tool-gripper-selection-guide`, not
`gripper-selection-guide` — unless a tool is deployed with a `pageNameOverride` or
via `deploy_tool_action`'s `strings.TrimPrefix(toolFunction, "tool-")` branch,
which produces the bare form. So the surface is real and reachable, but **it needs
a specific tool name that does not exist today.** The report arm's names are
`report-<uuid>`, so its two rows are not reachable at all.

**Conclusion recorded as it stands: the shape is confirmed by reading, the surface
is real, no collision has been observed.** This is prevention. Saying so is
better than inflating it — `bugs_closed/081` earned its severity from three months
of a measured loop and this file has no equivalent.

## Design — why the shared helper, and the two places it deliberately differs from 081

The owner's standing instruction is a framework fix over the individual case, and
175 says candidate 2 is "the only candidate that stops a seventh arm being written
with the same bug". Five arms in four files by different sessions is a class.

Two divergences from 081, both deliberate, both measured:

1. **"Has been live" is wider.** 081 guards `build_status = 'deployed'`.
   `bugs_closed/037` is an entire filed case about `needs_rebuild` falling outside
   exactly that predicate, and:

   ```
    build_status  | count | ever_deployed
   ---------------+-------+---------------
    deployed      |   491 |           490
    needs_rebuild |    46 |            35
    planned       |    42 |             0
   ```

   35 of 46 `needs_rebuild` rows have a non-null `deployed_at`. They have shipped.
   So the helper uses `build_status IN ('deployed','needs_rebuild') OR deployed_at
   IS NOT NULL`.

2. **A never-served row of another role is ADOPTED, not left mistyped.** 081
   refreshes such a row and leaves `page_type` alone — which for a constant-role
   arm just recreates the defect on the `planned` half. 081's own declined
   widening was about *repairing existing mistypes* and was declined on a count (0
   planned rows were mistyped); this is a different question — what to do with a
   collision on a row nothing has ever served. The arm owns the role by
   construction, so it takes the row over **completely**. A partial adopt (type but
   not url) would mint a fresh hybrid, which is the same defect wearing different
   columns.

## Missteps, in the order I made them

**1. I measured my own new check against a tree that already carried the fix.**
The `pattern-check` rule reported **0 findings over 1,120 Go files** and I nearly
wrote that down as "no false positives". It was a tree I had copied my own fixed
files into ten minutes earlier — so 0 was the *only* number it could have
returned, and it says nothing about whether the check can fire at all. Re-extracted
a genuinely pristine `git archive HEAD` and got **4 findings, exactly 175's
census** (`create_report_page`, `create_tool_component`, `deploy_tool` ×2). This is
`LANDMINES.md`'s "a gate's 0 findings has TWO causes with opposite fixes", and I
walked straight into it while writing a gate. **The check: a new detector must be
seen to FIRE on the case that motivated it, on a tree that still contains the
defect — before you record its clean run as evidence.** Logged in `WRONG_CALLS.md`.

**2. I wrote "6 hits before the fix" into the check's comment before measuring.**
It was 4. The number came from counting the `DO UPDATE` sites in the census rather
than from running anything. Corrected in place, with the count and the corpus size
in the comment so the next reader can re-derive it.

**3. `go test` failed on a compile error I did not cause.** Another session has
`drop_dead_url_controls.go` uncommitted with an unused import, so the whole
`actions` package would not build in the working tree. Fixing it would have swept
their WIP into my commit; leaving it meant I could not test. Resolved with the
practice already in memory: `git archive HEAD | tar -x -C <scratch>`, copy in ONLY
my six files, build and test there. Everything in this file's evidence was
produced in that clean tree, not in the shared working tree.

## Proving the guards rather than asserting them

`bugs_closed/081`'s test file had to be corrected once because
`mock.ExpectationsWereMet()` was mistaken for "no database call happened", and an
`UPDATE pages` added to the refusal path did **not** fail the test. So every guard
here was broken and watched to fail, on the clean tree:

| mutation | test that went red |
|---|---|
| ADOPT stops writing `page_type` (the filed bug, restored) | `UnshippedPageOfAnotherRoleIsAdopted` |
| `pageHasBeenLive` narrowed to 081's `build_status == "deployed"` | `NeedsRebuildCountsAsLive`, `DeployedAtAloneCountsAsLive` |
| an extra `UPDATE pages` on the refusal path | `LivePageOfAnotherRoleIsRefused` |
| refresh writes every column instead of the declared subset | `SameRoleRefreshesDeclaredColumnsOnly` |
| INSERT put back to `ON CONFLICT ... DO UPDATE` | `CleanCreate` |
| *(control)* unmutated | all green |

The assertions that catch mutations 1, 4 and 5 read the **SQL the helper actually
built**, captured through a `sqlmock.QueryMatcherFunc`. That is not decoration: the
defect IS a column list, and no expectation on statement order can see whether
`page_type` appeared in a `SET` clause.

## Three pattern-check findings on the commit that are NOT mine

The commit hook reported `logged-model-output` in `create_report_page_action.go`
and `unrepaired-component-write` in both tool actions. All three pre-date this
change and fire because I *touched* those files; the second pair belongs to
`bugs_open/136`'s lane, which has an allow-list mechanism for exactly this. Not
fixed here — fixing another lane's finding inside a 175 commit is how a commit
stops being reviewable. Noted so the next reader does not think 175 introduced
them.

## Sibling NOT converted, stated rather than silently skipped

`create_tool_component_action.go:288` creates its own tool page with a plain
`INSERT` and **no `ON CONFLICT` clause at all**. A collision there raises a unique
violation, deletes the just-created component and returns an error — loud and
fail-closed, so it is outside 175's silent-partial-update class. Converting it
would make re-runs idempotent, which is a real behaviour change nobody asked for.
Recorded in the bug file as a follow-up.

## The two fleet-wide files were swept into another lane's commit — nothing lost

I appended to `LANDMINES.md` and `WRONG_CALLS.md`, then named both on my `git
commit` pathspec. The commit took **4 files, not 6**: between my append and my
commit, the `bugfix_136_sibling_link_repair` lane committed `c734dbc98`
("docs(180): …"), which carried both files — my two entries included — into its
own message.

Verified intact in HEAD rather than assumed: `git show HEAD:…/LANDMINES.md | grep
"bugfix_175_page_role_upsert lane"` and the closing lines of my `WRONG_CALLS`
entry are both present and complete. So this is the CLAUDE.md case working exactly
as documented — "committing per task stops *you* sweeping up *others'* WIP; it
cannot stop a session that still runs `git add -A` from sweeping up yours" — and
the remedy is the one already written down: nothing is lost, forward-only holds,
say so and move on. Recorded because my own commit message claims those two files
and `git show` on that commit will not show them.

**Practical note for the next thread:** on the two append-only fleet files, expect
your lines to land under someone else's commit. Verify at HEAD (`git show
HEAD:<file> | grep <your marker>`), not at your own commit.

## 2026-08-02, after closure — the correction that answering the owner's question produced

The owner asked me to explain what RFC_010 was asking them to decide. Sizing decision
1's blast radius for that answer (`grep` for other `build_status = 'deployed'` guards)
turned up **`datahelpers.NeverDeployedPagePredicate`** — the estate's existing, shared,
tested definition of "this page has never shipped", with three consumers, and a test
whose assertion is *"predicate must not single out needs_rebuild"*.

**My widened predicate singled out `needs_rebuild`. It was wrong on 11 live rows:**

| domain | rows | shape |
|---|---|---|
| robot-hands.com | 5 | `needs_rebuild`, `deployed_at` NULL, `last_built_at` NULL, 0 components |
| lendzy.co.uk | 3 | same — and **tool pages created 2026-08-02**, i.e. what the tool arm collides with |
| dartsonline.com | 2 | same |
| gaswholesalers.com | 1 | same |

Never built, never served. `pageHasBeenLive` would have **refused** all eleven: a filed
human decision plus a hard error on the tool deploy, for pages nothing has ever seen.
The shared predicate gets them right because it keys on `deployed_at IS NULL` rather
than naming the status — and naming the status is exactly what produced a 34-page
false-positive class for the nav lane, which is why its test forbids it.

**Fixed:** the locking `SELECT` now computes `NOT (datahelpers.NeverDeployedPagePredicate)`,
`pageHasBeenLive` is deleted, and a test fails if anyone inlines a restatement.
Corrected in place, visibly, everywhere the wrong claim was written: the file header,
PBP-027, `LANDMINES.md` (whose entry recommended the wrong remedy), RFC_010 §2.1
(withdrawn — there was nothing for the owner to decide), and `WRONG_CALLS.md`.

**Why I missed it, which is the transferable part.** The council's `prior_art_librarian`
seat objected that I had not checked whether a page-upsert **helper** already existed. I
checked, answered the objection, and stopped — the objection named a helper, so I
searched for a helper. **The reusable unit here was one line of SQL**, and a one-line
constant is precisely what nobody thinks to search for.
`grep -rn "deployed_at IS NULL" --include=*.go` finds it in seconds. A well-argued
widening is still a widening: name the rows your version newly captures and *look* at
them.

**And a second, smaller misstep in the same pass.** My first mutation of the correction
failed to **compile** — deleting the last use of `datahelpers` left an unused import —
and I nearly recorded that as "the test caught it". It did not; the compiler did. Re-ran
with the import kept referenced (`_ = datahelpers.NeverDeployedPagePredicate`) and both
real assertions fired. **A mutation that does not compile is not a mutation test.**

## 2026-08-02, round 2 — the owner's ruling, and a widening that was a decoration

Owner ruled all three RFC_010 questions plus the residue: **opt-in** for ADOPT;
producer-convergence needs **no RFC** given register disclosure; refusal-becomes-error
**stands**; and fix `bugs_closed/081`'s guard **now**, not at next touch.

**The opt-in's default is the interesting design choice.** `AdoptUnshippedRows` left
false could mean two things, and neither is neutral: *refresh without re-typing* is
literally the partial update `175` filed, and *refuse* costs a spurious decision. It
refuses — but files **no** work item, because `mistyped_deployed_page` is a decision
about a LIVE artefact and this row has never been served. The reason string names the
missing field, so the refusal is diagnosable rather than mysterious. Assertion, not
argument: `TestUpsertPageForRole_AdoptionIsOptIn` checks `ItemFiled == false` and greps
the captured SQL for `site_work_items`.

**The two tests I did NOT have to change told me the fixture was wrong.** Making
adoption opt-in turned both adopt tests red, because `toolPageWrite` did not set the
field the real arm sets. Fixing the fixture rather than the assertion is the whole point
— *a fixture that quietly differs from production is a test measuring something nobody
ships.*

### The misstep worth the most: my widening of 081's guard was a DECORATION

I changed `existingBuild == "deployed"` to the shared predicate, then mutated it back to
check the tests bit. **Every test in `apply_gap_plan_deployed_conflict_test.go` stayed
green.** Both predicates agree on every input those tests supply — `deployed`+shipped
and `planned`+unshipped — so not one of them could see the change. I had "verified" a
widening no test could distinguish from its predecessor.

The discriminating input is the one where **only** the new predicate can act: a
`needs_rebuild` page that HAS shipped. `TestApplyNewPage_NeedsRebuildPageIsRefused`
supplies it; the same mutation is now red. This is the 180 lane's lesson from
`WRONG_CALLS.md` arriving from the other direction — there, a mutation passed because a
second guard sat in series; here, because the test inputs could not tell the two
predicates apart. **Same rule either way: when a mutation passes, ask what input would
make only the mutated line matter, and go and write it.**

A smaller one from the same pass: mutation A (hand-rolling the predicate) first failed
to **compile**, because deleting the last use of `datahelpers` left an unused import. A
compile error is not a test result. Re-run with `_ = datahelpers.NeverDeployedPagePredicate`
keeping the import alive, and both real assertions fired.

## 2026-08-03 — round 3 APPROVED, and two advisories that were checks, not opinions

Verdict trail on one correlation: **approved → revise → approved** (`e78c62e3`). Round 3:
*"approved with 5 advisory objection(s) — none high-severity"*, 6 seats abstained.

Two advisories were **claims of mine that could be checked**, so I checked them, and both
changed something I had written down.

### `debug_historian` was right: I scoped a blast radius by a status column I never enumerated

`bugs_open/181`'s census read `... AND status='active'` and published **28**. The seat
called it as the generalised form of the "`sites.status` is informational — do not scope
blast radius by it" trap. Enumerated: `pages.status` has two values, and **7 further rows
are `archived` AND shipped-but-not-`deployed`**. That is not a rounding error, because
`bugs_open/098` says archiving does **not** retract the live page — those 7 may still be on
the wire, and a detector skipping them is blind to a live artefact, which is the bug. 181
now states both numbers and says which question each answers. **The habit to keep: run the
`GROUP BY` on any status column BEFORE you filter on it, not after someone asks.**

### `prior_art_librarian` was right to refuse to take "it is live on v1.0.1233" on trust

That seat has no deployed-binary check tier and said so instead of nodding it through. The
pod-grep it asked for found the picture had **moved**: `v1.0.1234` is running, and on both
replicas —

| grep | result | means |
|---|---|---|
| old hand-rolled predicate (`IN ('deployed','needs_rebuild')`) | **0 / 0** | the 11-row spurious-refusal defect is **GONE from production** |
| shared predicate (`deployed_at IS NULL AND COALESCE(build_status…`) | 4 / 4 | the correction **shipped** |
| `did not set AdoptUnshippedRows` | **0 / 0** | the RFC_010 opt-in round has **not** shipped |

So `v1.0.1234` was built from a commit between `023f6624a` (the correction) and
`4ee695cc1` (the ruling round). **I had told the owner twice that "nothing from the
correction onward is live"; that stopped being true when someone else rolled.** The
statement was accurate when made and stale within the hour — which is the whole reason the
rule is *grep the running pod*, not *reason from what you committed*. A liveness claim has
a shelf life measured in minutes on this tree.

### The three advisories I did NOT act on, and why

- `reuse_agent` + `architecture` [medium]: the two constants are still two, cross-referenced
  by comment rather than merged. Deferred to `bugs_open/181` fix candidate 2 **on purpose** —
  `queryresolve` documents a deliberate family of three eligibility fragments, and collapsing
  one of them in a bug-fix round is the "shipped a contract inside a two-day fix" shape the
  same council flags elsewhere. Their objection strengthens 181's case; it is recorded there.
- `guardian` [medium]: bundling 081's guard (a different pipeline) into this round widened
  the blast radius, and dragged three mocks in another lane's closed test file. Fair as
  architecture criticism — **and it was owner-directed** ("fix the guard"), not opportunistic
  bundling. Recorded so the provenance is on the record rather than inferred.
- `editquality` [low]: the `links.go` edit is comment-only and should not count as a covering
  fix. Correct, and it never claimed to be — it is documentation of a relationship, and the
  fix is 181.
