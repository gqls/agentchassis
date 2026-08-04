# NOTES — bugfix 191 chrome link policy

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-08-04 — picking the bug

55 files in `bugs_open/`. Tallied `bugs_open/NNN` mentions across every session
transcript touched in the last 5–10 hours. Contention is real and heavy: `098` in 14
sessions, `153` in 10, `192` in 9, `083`/`178` in 7.

Zero- or one-mention candidates that survived a `who-owns.py` check: `191`, `188`,
`194`, `085`, `120`.

**Then I checked the FILES rather than the bugs, which changed the answer.**
`render_site_components_action.go` — 7 sessions. `run_checks_action.go` (bug 188) —
2 sessions with 29 and 15 hits. `save_page_sections_action.go` (bug 194) — 5
sessions. Contention on the bug and contention on the file are different questions
and the second one is the one that costs you a passenger.

Took `191` anyway, on two grounds: the file was **clean in `git status`** (7 sessions
had read it; none had uncommitted edits in it), and the change is two swapped lines
in that file plus a new file of its own.

## 2026-08-04 — the bug is still real

Code first: `render_site_components_action.go:171-187` is byte-for-byte what the bug
file quoted on 08-03, so nobody has fixed it underneath.

DB and wire:

```
mortgagecalculator.co.uk /tools/stamp-duty/index.html        -> 404   (planned, deployed_at NULL)
lendzy.co.uk /tools/price-cap-checker/index.html             -> 200   (the predicted false positive)
```

The bug file predicted both, including that its own SQL over-reports. It does — and
by more than it said: **6 rows, not 2**, because the `LEFT JOIN` on a regex-extracted
href makes every header with *no* CTA a NULL-join hit. 4 of the 6 have an empty
`cta_href`. Corrected query in `RUNBOOK` R1.

## 2026-08-04 — fable's plan, and running its own commands against it

Asked `fable` for the plan (`Plan` agent, model fable). **First attempt stalled
mid-stream** — the agent tool returned partial output and said so. Re-dispatched with
a tighter, ordered reading list; second run returned a complete plan.

The plan was good and its §5 did the right thing: it gave **commands** for the blast
radius rather than counts, and named what would disconfirm it. Ran all of them:

- `loadResolverPageSet` — def + exactly 3 call sites. As claimed.
- `loadFetchablePageSet` — def + exactly 1 consumer. As claimed.
- hand-spelled `deployed_at IS NULL` outside `datahelpers` — 3 hits, **all comments
  or a human-readable fix string**, no code. Had I counted the grep instead of
  reading it, I would have reported 3 rogue predicates.
- `NavFetchableOnly` consumers — all chrome-shaped (nav baked into shipped HTML).

## 2026-08-04 — ⚠ THE NEAR-MISS: a count that answered a different question

fable's §5 said a crowd of established sites in the zero-shipped bucket would
**disconfirm** the first-build escape argument. Ran it:

> **19 of 38 sites have zero shipped pages.**

That is half the fleet. It reads exactly like the disconfirming result, and I was one
sentence from writing "19 of 38 sites take the escape" into the register entry as a
measured fact.

It is false. The query lumps together two populations that behave nothing alike.
Split on whether the site has any pages at all:

| bucket | sites |
|---|---|
| has shipped pages → strict | 19 |
| **no pages at all** (never built; chrome renders nothing either way) | **18** |
| has pages, none shipped → **actually takes the escape** | **1** (`webdesign.uk`) |

The real escape population is **one genuinely young site**. The design survives, but
it survived a check that had to be asked properly first. Logged in `WRONG_CALLS.md`.

The transferable form: `count(*) FILTER (shipped = 0)` silently answers "sites where
the filter would be disabled **or would have nothing to filter**". Those are not the
same claim, and only one of them is about my design.

## 2026-08-04 — implementation, and what the mutations caught

Wrote `ChromeLinkPolicy` (new file), refactored `applyNavVisibility` to consume it,
swapped the CTA's two checks, documented the boundary on `loadResolverPageSet`, added
`chrome_link_policy_test.go`.

Standing control held: **`nav_visibility_test.go` passes unedited**, all 10 subtests.
That is the behaviour-preservation proof for the refactor and it is worth more than
the new tests, because it is the one I could not tune to pass.

### The mutation run, and a misstep inside it

Built `git archive HEAD` + my five files into `/home/ant/.cache/claude-mut-191` so
the shared tree was never dirtied by an experiment.

- ⚠ **`/tmp` is a 16 GB tmpfs shared by every session and it was 100% full** (45 MB
  free). `go build` failed with `no space left on device`, which looks like a broken
  toolchain. The session scratchpad lives on the same tmpfs, so it was no escape
  either. Fixed with `TMPDIR=/home/ant/.cache/gotmp-191`.

- ⚠ **MISSTEP: my first mutation of the first-build escape deleted the `case` block
  with a crude slice, and the result was `[build failed]`.** A build failure is not a
  red test. If I had skimmed the output for the word FAIL I would have recorded a
  guard as proven when nothing had exercised it — the exact shape of "a quiet test
  passes when the RULE is gone". Redone as `case deployedPages == 0:` →
  `== -1:`, which still compiles, and the guard then genuinely failed.
  **Rule: mutate a CONDITION, never delete a block.**

All four mutations bite (table in `RUNBOOK` R4).

## 2026-08-04 — registration and submission

LNK-030 in `link-management.md`, index row, headline count 1,758 → 1,759, in the same
commit as the code per the platform-seams ruling.

⚠ `scripts/test-concept-register-drift-local.py` **reads at `HEAD`, not the working
tree**. It cheerfully reported "Clean, 1758/1758" while my entry sat uncommitted and
invisible to it. A pass from that harness says nothing about your own change — it is
a check on the committed state only. Noted in the index header itself so the next
person does not read its green as validation.

LANDMINES entry appended and `landmines-sync.py --apply` run (6 footprints
registered).

Council: `SUBMISSION_CORR = 78b0b7ff-f88d-402b-8f8f-ca4ae01c2d30`, round 1.
Committing before the verdict with `Council-Submitted:`, per the 07-30 rule.

## 2026-08-04 — what is still OWED

The bug is fixed in code and **not yet proven live**. Go changes are inert until an
image is rebuilt and rolled. Owed after the next roll: the four-way pod-grep on
**both** replicas (this change removes a string, so a real negative control exists —
take it rather than inventing one), a `nav-updater` re-run on
`mortgagecalculator.co.uk`, the corrected R1 query, and a curl of every survivor.

## 2026-08-04 — the council: four rounds, three of which changed the code

Verdicts on correlation `78b0b7ff-f88d-402b-8f8f-ca4ae01c2d30`:

| round | verdict | gated by |
|---|---|---|
| 1 | REVISE | `bug_historian` — the fix is inert for chrome already rendered |
| 2 | REVISE | `render_guardian` — show that a corrected slot reaches a DEPLOYED page |
| 3 | REVISE | `editquality` — edit 8 duplicated edit 2 (my artefact defect, not the code's) |
| 4 | **APPROVED** | 3 advisory, none high, 4 abstained |

**This is the run that justifies the gate.** I had a green build, mutation-proven
tests and a registered seam after round 0, and the change was still wrong twice.

- **Round 1** — I had written the repair for already-affected sites as a manual
  step in my own verify-later list. That is not a mechanism, it is me
  remembering, and a manual step in a bug file is exactly where such things die.
- **Round 2** — I answered by analogy ("same shape as 166") without showing the
  propagation. Reading the deciding arm settled it: `assemblePage`
  (`rerender_single_page_action.go:532-537`) calls `getSiteComponents`, which is a
  plain `SELECT ... FROM site_components` — chrome is re-read on every assembly,
  never baked in. **Cite the arm, not the function.**
- **Rounds 3/4 — the expensive lesson.** `reuse_agent` and `guardian` asked,
  independently, whether `bugs_open/166` had already built what I was adding.
  **It had, and I had not looked.** `repointRetiredChromeSlot` signals a needed
  re-render with `build_status='pending'` under `pageComponentAgentWritableSQL`.
  My design computed staleness above the loop and OR'd it into `force`, which
  invented a second force channel **and bypassed the lock guard** — a human-locked
  chrome slot would have been forced to re-render. The 069 gate downstream would
  have caught it, but only after the repoint path had already written.
  **No test of mine would have caught this**, because every test I wrote was
  written against my own design. Rewritten to extend 166's mechanism.

The LANDMINES entry for that exit says it outright — *"do NOT clear
`rendered_html` to force it... `build_status='pending'` is the supported signal"*
— and I had grepped LANDMINES for my own file paths at session start without
reading the entry for the mechanism I was about to extend.

## 2026-08-04 — re-verifying the blast radius by a method that is not grep

`prior_art_librarian` objected (medium, advisory) that the absence claims rested
on my own grep, which this estate documents as unreliable for exactly this class.
Fair. Re-ran by **compiler**: rename the helper in a `git archive HEAD` copy and
read the build errors — a method that cannot miss a caller.

```
loadResolverPageSet  -> resolve_internal_links_action.go:147
                        rerender_page_sections_action.go:626      (2, both page-CONTENT)
loadFetchablePageSet -> chrome_link_policy.go:61                  (1, this policy)
```

Corroborates the grep and sharpens it: `loadResolverPageSet` is down to 2 callers
because the chrome one is gone, and `loadFetchablePageSet`'s single consumer moved
from `applyNavVisibility` to the policy. Both symbols the reuse claim rests on
confirmed present **at HEAD**, not merely in the working tree.

Note this is a genuinely independent method, not the same check twice: grep reads
text, the compiler resolves symbols. Two greps with different spellings would have
agreed with each other and proved nothing.
