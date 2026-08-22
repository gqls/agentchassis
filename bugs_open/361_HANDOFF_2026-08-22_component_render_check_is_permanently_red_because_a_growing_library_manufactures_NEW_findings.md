# 361 — `component-render-check` is permanently red, because a GROWING LIBRARY manufactures "NEW" findings

**Filed 2026-08-22** by the `bugs_open/260` lane, which was asked to find out why the job was
failing. **Status: OPEN, UNOWNED** — `scripts/who-owns.py 140` puts the originating lane
(`bugfix_140_contact_info_fabrication`) at **quiet 14d**. This lane is not working it.

> **On the 090 loop, stated plainly because the 2026-07-31 owner ruling requires it.**
> Not filed to the loop, and here is the substitute so the omission is declared rather than silent.
> The root cause is **one line I read and quote** (`rendercheck.go:759`), not a cause living
> somewhere the symptom does not point — which is the shape 090 exists for. The population claim is
> a query whose output is quoted below, run against the live library. The one claim that is neither
> — that the job will go red again shortly after any regeneration — is marked **[INFERRED]** where it
> appears, with its arithmetic shown and its disconfirming result named.

## 1. The one-paragraph version — nothing is broken

The job is **doing exactly what it was built to do, and reporting truthfully every day.** It renders
every active component with each referenced field removed, flags the fields whose absence produces a
visible hole, and compares that set against a **baseline** — a committed list of findings it already
knows about, embedded in the binary. It exits 1 on a **NEW** finding (or an UNCOVERED one), and with
`backoffLimit: 1` two exit-1 attempts make the Job read `Failed`.

**The comparison is a flat key-level set difference** (`rendercheck.go:759`):

```go
if !base[k] {
    newFindings = append(newFindings, f)
}
```

A component that did not exist when the baseline was cut has **no keys in the baseline at all**, so
*every* finding it produces is "NEW". The baseline has been regenerated **once, ever**. The library
has grown by 109 components since. So the ratchet now measures library growth, not regression.

## 2. Evidence

**(a) How long, and the last green run.** `[MEASURED 2026-08-22]`
`kubectl -n ai-persona-system get cronjob component-render-check -o json` →
`status.lastSuccessfulTime = 2026-08-09T06:55:21Z`. **Twelve consecutive red days.** The three
retained failed Jobs each show `BackoffLimitExceeded`, `failed: 2`, ~38 s after start — i.e. the
binary ran to completion and exited non-zero, twice. It is not crashing, not timing out, and not in
`ImagePullBackOff`.

⚠ **`failedJobsHistoryLimit: 3`, so `get jobs` shows only the last three failures.** Do not size the
outage from it — the job list looks like "failing since Thursday". The honest instrument is
`lastSuccessfulTime`, and the `doc_notes` series below.

**(b) It reports on every run, red or green** — `--report` writes one `doc_notes` row per attempt:

```sql
SELECT created_at, left(split_part(body, E'\n', 1), 200) FROM doc_notes
WHERE source='component_render_check' ORDER BY created_at DESC;
```

| day | active components | NEW | rows/day |
|---|---|---|---|
| 08-07 → 08-09 | 184 → 185 | **0** | 1 (green) |
| 08-10 | 189 | first red | 2 |
| 08-15 | 228 | 51 | 2 |
| 08-18 | 245 | 79 | 2 |
| 08-20 | 263 | 112 | 2 |
| 08-21 | 275 | 210 | 2 |
| 08-22 | 282 | **227** | 2 |

(Two rows a day *is* the failure: `backoffLimit: 1` ⇒ two attempts ⇒ two reports. One row a day is
the green shape. That is a cheaper red/green signal than the Job list, and it is retained.)

**(c) The decisive measurement: the NEW findings are new COMPONENTS, not regressions.**
Today's 227 NEW findings belong to **38 distinct components**. Joining those names to the library:

```sql
WITH new_names(name) AS (VALUES ('…'), …)   -- the 38, parsed from the doc_notes body
SELECT (c.created_at < '2026-08-04') AS existed_at_baseline, count(*)
FROM content_components c JOIN new_names n ON n.name = c.name
WHERE c.is_active GROUP BY 1;
```
→ **`false | 37`**, **`true | 1`**.

**37 of 38 were created after the baseline was cut.** The single exception is
`blog-listing_pre_037` (created 2026-04-08), and it is not a regression either: its template was
**rewritten**. Verified directly — `html_template LIKE '%post1_title%'` is now **false** and
`LIKE '%range .articles%'` is **true**, which is also why **31 of the 54 "fixed" keys are its**.
It gained exactly one NEW key, `.articles empty_block`, which is the new template's shape.

So: **zero of the 227 NEW findings is a component that existed at baseline time and got worse.**

⚠ **This does NOT mean the findings are false.** Each is a real component that can render a visible
hole. They are new debt arriving with new components — not regressions. The check is right; the
*ratchet* is measuring the wrong thing.

**(d) The clone rule already solved the adjacent case, and cannot solve this one.**
`findingKey` (`rendercheck.go:485-491`) maps a component whose template is byte-identical to another
onto that representative's key, so a clone inherits the baseline and is not NEW. That was added
2026-08-05 after the first unattended run produced 13 NEW findings that were all one duplicated
template (register `CGV-030`). **Today's run reports 0 inherited** — the summary line carries no
`, N inherited` clause — so all 227 are genuinely distinct templates. The existing protection is
working and is simply not the protection this needs.

**(e) Why regenerating alone will not hold. [INFERRED — arithmetic, not a measurement]**
`content_components` created per day since the baseline: 2, 6, 0, 0, 1, 4, 1, 4, 13, 14, 3, 17, 1,
2, 6, 13, 13, 8, 3 — **109 over 19 days (~5.7/day; ~8.6/day over the last five)**. Of the 109,
**37 produced a NEW key (34%)**, so roughly **two new NEW-producing components per day**.
A regeneration therefore buys on the order of **one day** of green.
**What would disconfirm this:** a regeneration that stays green for a week. If the estate's
component creation stops or the new components stop rendering holes, the projection is wrong and
this bug is just an operational chore. Check it by regenerating and watching `lastSuccessfulTime`.

## 3. Root cause

`cmd/component-render-check/rendercheck.go`, line numbers valid at **`f11851c27`** (file unmodified
in the tree):

- **:759** — `if !base[k] { newFindings = append(...) }`. The ratchet is a **key-level** set
  difference against a **component-level-blind** baseline. There is no notion of "this component is
  not baselined yet", so an unbaselined component is indistinguishable from a regressed one.
- **:893** — `if len(newFindings) > 0 || len(uncoveredKeys) > 0 { os.Exit(1) }`.
- The author saw the risk and wrote it down at **:744**: *"a baseline nobody regenerates is a
  baseline that slowly stops describing the tree"*. What is missing is not awareness — it is a
  **mechanism**. There is no regeneration cadence, no owner, and nothing that notices staleness.
  `--write-baseline` is documented in the `bugfix_140` RUNBOOK as *"banking a real fix"* — a
  deliberate act after fixing something. Nobody designed it as an act after the library **grows**,
  which is the thing that actually happens daily.

**And the design's own manifest predicted this outcome** — `base/cronjob.yaml`: *"a permanently-red
job is a job everybody learns to ignore, which is the mistake this check was sequenced to avoid."*
Twelve days is that arriving.

## 4. Fix candidates, ordered by what closes the door

1. **Make the ratchet component-scoped.** A finding whose component owns **zero** keys in the
   baseline is not NEW — it is *unbaselined*. Report those in their own bucket with their own count,
   and fail only on a NEW key in a component the baseline already knows. This makes a red mean
   *"something that used to be fine got worse"*, which is what a ratchet is for, and it preserves the
   whole point of the embedded baseline: a **regression in an existing component still fails**.
   ~6 lines. The cost is honest and must be stated in the report: a brand-new component that renders
   holes will not fail this job — that debt belongs to birth-time gating (CGV-029 / the component
   birth path), not to a regression ratchet.
2. **Regenerate on a cadence, mechanically** — the job writes a fresh baseline and raises it as a
   reviewable diff. Keeps the census honest but **conflicts with the embedded-baseline design**
   (`base/cronjob.yaml`: a baseline that can be edited without a diff anybody reviews is the quiet
   clearing this check exists to catch), and it would silently bank a real regression alongside the
   growth. Only viable if (1) lands first, so what gets banked is provably not a regression.
3. **Regenerate by hand, periodically.** This is the status quo intent, and **twelve red days is the
   evidence that an unassigned manual step is not a mechanism.** Listing it because somebody will
   propose it.

**Whatever wins, the immediate operational question is separate and is the owner's**: regenerating
today banks **227 real findings** as "already known". They are real, they are not regressions, and
banking them is what makes the job able to go green. That is a debt decision, not a code decision.

## 5. Two things to fix in passing, cheap

- **`000_concept_index.md:1193` is STALE.** The CGV-030 index row still reads
  *"built + calibrated, CronJob wiring owed"*, while the register entry it points at says
  **"DEPLOYED AND PROVEN IN THE CLUSTER, 2026-08-04"** and strikes the wiring out as DONE. The index
  is what sessions grep before concluding something does not exist. **Corrected by this lane, and the
  correction is LIVE at HEAD — but not in a commit of mine.** I held it back from `361`'s commit
  precisely because another session had that file dirty and a pathspec commit takes a same-file
  passenger; in the minutes that took, **they committed first and my line rode into `5fddba825`**
  (a TL-049 status commit). So the trap is symmetric, and waiting does not avoid it — it only
  decides which side of it you are on. Recorded here so that session can find the line they carried.
  Noted at all because the class recurs — `LANDMINES.md`, *"A concept-register STATUS
  line is a snapshot that outlives its truth"*.)
- **`CGV-030`'s `verify-later` asks "whether the first UNATTENDED 06:55 firing succeeds".** It did
  (2026-08-05), and then the job went red on 08-10 and stayed. Worth writing back into the entry, so
  the next reader does not have to re-derive twelve days of history from `doc_notes`.

## 6. How to verify a fix

- **The instrument is `lastSuccessfulTime`, not the Job list** (§2(a)) — and the `doc_notes` series,
  where **one row a day is green and two is red**.
- **Prove BOTH arms, by mutation** — the RUNBOOK's R11b arms are already written for this tool and
  are the right pattern. A component-scoped ratchet must still exit **1** when a component that IS in
  the baseline gains a key, and exit **0** when an unbaselined component produces one. A fix verified
  only on the second arm is a fix that turned the check off.
- **Do not verify with a self-comparison** (`--write-baseline` then `--compare` against it): 0 NEW,
  0 fixed is the no-op case and proves nothing. That is stated in the register entry too.
- **Deploying it: `make deploy-component-render-check` ships NOTHING on its own** — this service's
  tag is pinned in its OVERLAY, not taken from `IMAGE_TAG`, and both make and `kubectl apply` report
  success anyway. Bump `newTag` in the same commit as the rebuild, then read the artefact:
  `kubectl -n ai-persona-system get cronjob component-render-check -o jsonpath='{.spec.jobTemplate.spec.template.spec.containers[0].image}'`.
  Full trap in `LANDMINES.md`.
