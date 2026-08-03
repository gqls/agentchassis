# NOTES — bugfix 168, deployed asset path

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-08-02 — taking the bug, and checking it was still real

Picked 168 after checking ownership two ways: `scripts/who-owns.py 168` (only the filing
commit touches it) **and** a grep of every `.jsonl` session transcript modified in the last
4 hours for `bugs_open/NNN` references, because `who-owns` reads COMMITS and is blind to a
session mid-fix. 27 sessions were live; the bugs they were holding were 010, 018, 029, 043,
064, 081, 083, 092, 095, 099, 117, 119, 120, 129, 136, 137, 138, 139, 149, 150, 151, 154,
156, 157, 159, 162, 164, 165, 169, 171–175. 168 appeared in none of them.

Bug still valid: `platform/storage/url_helpers.go` unchanged since `d671fb2b2`, and
`DeployedWebPath` still returns `/assets/images/og_card.png`.

## MISSTEP 1 — I nearly inherited the bug file's mechanism without checking it

The filed root cause says an underscore purpose stored with `asset_key = purpose` "yields a
path with an underscore where the deployed file has a hyphen". I was one edit away from
implementing fix candidate 2 (swap `_`→`-` unconditionally) on that basis.

Then I read the writer. `deploy_image_asset_action.go:185` branches on the **identical**
condition:

```go
if assetKey != "" && assetKey != purpose {   // deployer
if assetKey == "" || assetKey == purpose {   // helper (the skip)
```

So for a deployer-published file the deployed name **also** has the underscore. Helper and
deployer agree; candidate 2 would have *broken* that agreement for every future underscore
purpose — introducing the drift it claims to remove. The check that caught it was reading
`DownloadOptimizeAndPrepare` (it returns `BuildAssetPaths(purpose, ext)` verbatim), which
took two minutes.

**Cheap check that generalises:** when a bug file names a mismatch between a reader and a
writer, read the WRITER before believing the direction of the mismatch. A helper documented
as "mirroring" something is a claim about behaviour, not the behaviour.

## MISSTEP 2 — a live-impact hypothesis I talked myself into, then refuted

The queue showed `undeployed_asset: Asset 'og_card' generated but not deployed to site`,
four of them `unresolved after 2 attempts`. I formed a satisfying theory: the repair route
is `asset-deployer` → `deploy_image_asset` → writes `og_card.png` → the head references
`og-card.png` → the 404 persists → hence four permanent unresolved items. That would have
made 168 live-biting rather than latent, and I would have written it up that way.

It is wrong. `check_undeployed_assets.go:256` excludes brand-head purposes from the generic
half — `AND NOT (COALESCE(a.purpose,'') = ANY($2::text[]))` — and routes them to
`needs_brand_head_assets`, whose repair is **re-derivation**, not deployment. So brand-head
never reaches the deployer. The unresolved rows predate the 142 lane's fix.

**What caught it:** reading the SQL of the check I was about to accuse, instead of reasoning
from the item's summary text. The summary said "not deployed to site"; the predicate said
which rows can produce that summary, and og_card is not among them any more.

## The diagnosis loop: REFUTED — and it was right about the thing that mattered

Filed `090` before asserting a structural cause, per the owner ruling of 2026-07-31.
Intake `62ab1470-2edc-46ce-a480-8deea38e4ed0`, run corr
`ae9404bd-dab7-4606-ade3-c439ebda93af`. Verdict **REFUTED**, 2 iterations.

Its correct and useful finding: `injectBrandHeadTags` **hardcodes** `/assets/images/
favicon.png` and `/assets/images/og-card.png` and never calls `DeployedWebPath` or reads
`assets.url`. So there is no render-time failure today — which corroborates the bug file's
own "Low today, latent" severity, independently of me. This is why the fix is framed as
removing a drift mechanism, not as repairing live breakage, and why the code comment says
"corrects a latent disagreement rather than moving live traffic".

**Where the loop itself was wrong, recorded rather than accepted:** it asserted
*"DeployedWebPath's only found call site (queryresolve.go's webPath)"*. There are **six**
(grep below). One of them — `check_image_url_404` — is exactly where the two derivations
DO meet, which is why the 128 lane had to add an `IsBrandHeadPurpose` branch there to avoid
reporting a 404 for the og card and favicon of every site in the fleet. Its refutation rests
on an incomplete census, so I did not treat REFUTED as "no defect here".

```
plan_sections_action.go:304,328,355,393,423   render_site_components_action.go:415
emit_sprite_css_action.go:136                 derive_card_asset_action.go:204
queryresolve/queryresolve.go:294,297          discovery_checks/check_image_url_404.go:426
```

**Genuine side-finding from the loop, not mine to fix:** several active `favicon`/`og_card`
rows carry `assets.url = '/assets/images/input-data.asset-key.jpg'` — an unresolved template
literal run through the `_`→`-` swap. Already documented in `check_undeployed_assets.go` and
owned by `bugs_open/152`. Left alone.

## Measurements, with the queries inline

Fleet census of the skip branch, re-run 2026-08-02 (the bug file measured it 2026-07-31 —
grounding a figure rather than carrying it forward):

```sql
SELECT purpose, COALESCE(asset_key,'<null>') AS asset_key, count(*) AS rows
  FROM assets WHERE status='active'
   AND (asset_key IS NULL OR asset_key='' OR asset_key=purpose)
 GROUP BY 1,2 ORDER BY 3 DESC;
```
```
og_card | og_card | 12      hero | hero | 5
favicon | favicon | 12      logo | logo | 4
```

Identical to 07-31. 267 active rows total; every other underscore purpose
(`content_hero` ×31, `sprite_sheet` ×1) carries a distinct key and takes the swap correctly.
So the risk set outside brand-head is empty **structurally**, not by luck.

Also measured, and it surprised me: **all 24 brand-head rows store a site-relative web path
in `assets.url`**, not a presigned S3 URI — `recordDerivedAsset` writes the path it just
committed. `url` is therefore polymorphic across writers (presigned S3 for generated assets,
web path for derived brand-head ones). Noted, not relied on; the helper's doc comment already
warns callers off `assets.url`.

## Guards proven by mutation, because a green run proves nothing

Three mutations, three confirmed failures, each reverted immediately:

1. Removed the brand-head branch from `DeployedAssetPath` →
   `TestDeployedWebPathExpressesBrandHeadPaths` fails: `= "/assets/images/og_card.png",
   want "/assets/images/og-card.png"`.
2. Broke `RelativeURL` = `"/"+FilePath` in `assetPathsForFilename` →
   `TestDeployedAssetPathFormsAreConsistent` fails on five inputs, plus the two older tests.
3. Made the deployer re-implement the derivation (`storage.BuildAssetPaths` + a reference to
   `AssetKeyFilename`) → all three arms of the source sensor fire, including the
   anti-vacuity arm.

The tautology I deliberately did **not** write: "the deployer's path equals the renderer's
path". Both now call one function, so that assertion cannot fail, and it would keep passing
if someone reintroduced a private copy behind a condition the test does not exercise. What
broke before was a STRUCTURE — two implementations held together by prose — so the sensor
reads the source. That is stated in the test itself, so nobody deletes it as "not a real
test".

## The fourth mutation — the one that actually mattered

Removing `check_image_url_404`'s `IsBrandHeadPurpose` branch made me uneasy: that branch was
the 128 lane's guard against a **fleet-wide false 404 on the og card and favicon of every
site**, and "it's redundant now" is exactly the sentence someone says before deleting a load
-bearing check. Post-commit, I went back and proved the protection transferred rather than
evaporated.

`TestImageURL404_BrandHeadPathsResolveThroughTheirOwnMap` turns out to be a **behavioural**
test, not a structural one — it feeds asset rows `{favicon,favicon}`/`{og_card,og_card}` and
chrome referencing `/assets/images/og-card.png`, and asserts zero findings. So it does not
care *how* the path is resolved, only that it comes out right. Mutating the helper (deleting
the brand-head branch from `DeployedAssetPath`, leaving the check exactly as I had rewritten
it) reproduces the original fleet-wide defect:

```
--- FAIL: TestImageURL404_BrandHeadPathsResolveThroughTheirOwnMap
    brand-head assets exist; nothing should be reported, got [image_url_404:og-card.png]
```

That is the finding this whole family exists to prevent, produced on demand, through the new
code path. The guard is intact and now sits one level deeper, where it covers all six
consumers instead of one.

**Generalises:** when a refactor removes a local guard on the grounds that a shared mechanism
now covers the case, do not argue it — **mutate the shared mechanism and require the local
test to fail**. If it stays green, the guard did evaporate and the argument was wrong. Cost:
about ninety seconds.

## The blast-radius claim I asserted before I measured it — checked afterwards, and it held

In the council submission I wrote that five of the six readers "pass non-brand-head purposes
(card, sprite_sheet, hero, logo, icon)". Three of those call sites pass a **literal**
(`"card"` ×2, `"sprite_sheet"`), which I had read. The other three pass a **variable**
`purpose` scanned out of a query — `plan_sections` ×5, `render_site_components`,
`queryresolve`'s `HeroPurpose` — and for those I had inferred the value from context rather
than measured it. That is asserting a blast radius, which the 2026-07-28 ruling says to
measure yourself rather than hand to the reviewer.

So I measured it. Every variable-purpose site resolves `purpose` from `assets` joined to
`site_plan_imagery` through the current site plan:

```sql
SELECT a.purpose, count(*) AS rows,
       count(*) FILTER (WHERE a.asset_key = a.purpose) AS takes_skip_branch
  FROM site_plan_imagery spi
  JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current = true
  JOIN assets a ON a.site_id = sp.site_id AND a.asset_key = spi.key AND a.status='active'
 GROUP BY 1 ORDER BY 2 DESC;
```
```
 hero         | 82 | 0        illustration | 3 | 0
 icon         | 77 | 0        sprite_sheet | 1 | 0
```

**Stronger than what I claimed.** Not only is no brand-head purpose reachable there (a
`WHERE a.purpose IN ('favicon','og_card')` variant returns 0 rows) — **not one of the 163
reachable rows takes the skip branch at all**, because every one carries an `asset_key`
distinct from its `purpose`. `sprite_sheet` is the only underscore purpose that gets there
and it is keyed `sprite_sheet_main`. So those call sites cannot reach the code I changed,
by data as well as by type.

⚠ **The zero needed a denominator and I ran one** — the same join without the purpose filter
returns **163 rows across 4 distinct purposes**. A `0` from a join that returns nothing at
all would have looked identical and meant nothing.

**One loose word in the submission, corrected here:** I listed `logo` as one of the purposes.
It is not — at these sites the logo is stored as **`asset_key='logo'` under `purpose='hero'`**
(10 rows fleet-wide; the same quirk `check_image_url_404_test.go` documents for
fundamentallyai.com). The claim survives, because `hero` is not brand-head either, but
"purpose" and "asset_key" are exactly the two things this bug is about and I should not have
blurred them in a submission about them.

## Council round 1 (`abd9b119`) — REVISE, gated by guardian. What each seat was right about

12 seats fired (relevance-gated). 8 approve, 4 object, gating objection from `guardian`.
Recording the whole disposition, because "which objections were real" is the part that does
not survive in a commit message.

**The one real CODE defect — `editquality`, edit 1, low.** `brandHeadAssetPathsFor` lifted the
*filename* out of the map's value and re-joined it under `DefaultAssetBasePath`. Correct for
both entries that exist; **silently wrong for the first brand-head artefact not served from
that directory**, and the seat named the exact realistic case — a favicon at the site root.
No test in my change could have caught it, because both real entries live under that
directory, which is why the seat's second sentence mattered more than the first: the forms
test only checked the three forms *against each other*. Fixed (takes the map's value whole,
refuses a non-absolute value), plus two new tests. Mutation-proven: restoring the old split
fails with `/assets/images/favicon.ico, want /favicon.ico`.

**Right about my COMMENT, not my code — `editquality`, edit 2, medium.** It read the map's
comment as citing `TestBrandHeadAssetPathsMatchTheDeriver`, "a guard that was never built".
The guard exists and passes — it is the 142 lane's, at
`check_undeployed_assets_test.go:342`, and I did not need to build it. But the council sees
only the edit list, so from where it sat the comment cited a test that appears nowhere. **That
is a defect in the comment, and the seat found it correctly from the evidence it had.** The
comment now names each writer's pin, its package, and whether this change added or found it.
*Lesson: when a doc comment cites a guard, say where it lives and whether you wrote it.*

**Right, and I had already done it but not told them — `debug_historian`, medium.** "Never a
pod-grep of the RUNNING chassis binary." I had run one (three controls, proving NOT live) —
*after* submitting. It was not in the submission, so the objection was fair on the record.

**Right on principle, with the precedent to prove it — `bug_historian`.** The disclosed
`deploy_path` exposure needed *tracking*, not prose. `bugs_open/168` exists **only** because
the 128 lane wrote "its own item" and never created one. Filed `bugs_open/179`.

**The gating one, and the most instructive — `guardian`, high.** "The landmine list carries an
entry specifically for this symbol that has not been read into this review." I had read it —
I *rewrote* it in this change — and cited it nowhere in the submission. The seat could not
tell the difference between "read and handled" and "walked into". **Its note says so
explicitly: "I need the actual doc_notes body before I can tell whether this plan already
accounts for it or is walking into it."** Handling a landmine and *showing* you handled it are
two different deliverables, and only the second is visible to a reviewer.

**A finding I owed them that nobody asked for.** Answering guardian's "is `IsBrandHeadPurpose`
called anywhere besides `check_image_url_404`?" turned up something worse than the question:
after my change it has **zero production call sites** — definition, two comments, four test
references. A helper with no callers looks exactly like a finished refactor. Disclosed in
round 2 rather than left to be found, kept deliberately (it is still the right predicate for
"which table holds the evidence?"), and `179` candidate 1 would give it a job again.

**Accepted without argument — `architecture`, medium ×2.** A companion RFC as a *standing
artifact*, not plan prose. Its sentence is the correction of the round: *"The author's own
framing is the correct self-assessment — but per the standing rule, **declaring it doesn't
relocate it**."* I had self-declared architecture scope, brought it to the gate deliberately,
and still owed a citable document. Filed `RFC_009`.

### The pattern across the four objections that landed

Three of the four were not "you did the wrong thing" but **"you did not show me you did the
right thing"** — the landmine, the pod-grep, and the RFC were all done or true, and none was
legible from the submission. The fourth (`editquality`'s path split) was a genuine defect that
only a careful reader of the *code* would find, and it was rated **low** while the three
legibility objections were rated medium and high. Worth remembering when writing a
submission: **the council reviews the artefact you hand it, not the work you did.** Evidence
you hold and do not cite is evidence you do not have.

## Council round 2 (`abd9b119`) — REVISE again, and this time the council was right about the CODE

Gated by `bug_historian` at **high**, `guardian` concurring at medium, both on one point:
*"after this change, routing a brand-head purpose through `deploy_image_asset` no longer
produces a harmless orphan file but SILENTLY OVERWRITES the deriver's favicon/og-card
artefact — a git commit with no error, no warning, no failed item"*, and the guard for it was
**tracked, not shipped**.

I had told them twice it was *"measured as currently unreachable"*. **It is reachable.**

### The measurement error, which is the thing to take from this lane

I measured two populations. Both measurements were correct, both carefully denominatored,
and **neither could answer the question**:

| what I measured | why it cannot answer |
|---|---|
| `check_undeployed_assets` no longer *raises* brand-head items (true since 142) | about items created **from now on** |
| every variable-purpose **reader** resolves through `site_plan_imagery`, no brand-head purpose in 163 rows | about **readers**; the exposure is in a **writer** |

The population that answers it is the **standing queue**, and it took one query:

```sql
SELECT status, spec->>'mode', spec->>'purpose' FROM site_work_items
 WHERE item_type='undeployed_asset' AND spec->>'purpose' IN ('og_card','favicon')
   AND status NOT IN ('complete','cancelled','rejected');
```

**11 rows. `mode` NULL on every one. Two at `detected`**, which `triage_detect_items` promotes
into the build queue. dartsonline ×2, robot-hands ×9, all from before 142's fix. And
`asset-deployer`'s `check_mode` only diverts `mode == "brand_head"` — so they fall through to
`deploy_asset` → `deploy_image_asset` with `purpose=og_card`.

**A predicate change stops the tap; it does not empty the bath.** Nothing in this platform
sweeps a queue for items whose defining predicate has since moved. And the scope-out I was
leaning on was *inherited from a closed bug*: 142 made those items safe **under the code that
was live when 142 shipped**. My change moved the code and silently un-scoped them. 142 had no
reason to mention it, because under the old code they were harmless litter.

**What actually caught it: running the query to PROVE the objection wrong.** I believed the
seats were being cautious about something I had already established, and went to demonstrate
it for round 3. It returned 11 rows. Had they accepted my round-1 assurance, this ships.

### The second wrong call in the same hour — my own paperwork as evidence

Answering `prior_art_librarian`'s "I cannot check the `deploy_path` claim from this seat", I
ran `... WHERE collected_data::text LIKE '%deploy_path%'` over `orchestration_states` and got
**9** — apparently falsifying the 128 lane's "zero orchestrations in history" that I had by
then repeated to the council twice. **All nine were my own council submissions**: a council
run stores the submission JSON in `collected_data`, and my rationale argues about
`deploy_path` at length. Matching the JSON *shape* (`"deploy_path":"`) returns **zero**, so
the inherited claim survives — but the harder I had argued the point, the more rows my own
argument generated. Both wrong calls are in `WRONG_CALLS.md`.

### What shipped, and why it is safe to ship the clobber at all

`deploy_image_asset` now **refuses** a brand-head purpose — a completed result carrying the
reason, not an error, so the 11 queued items *resolve* instead of retrying for ever against a
guard that will never pass them. Positioned before the URI resolution, the download and the
commit.

**There is no exposure window, and that is luck rather than design:** the derivation change is
not live either (pod-verified, `v1.0.1228`), so the clobber and its refusal go out in the same
image. Worth stating plainly rather than presenting as planning — if 168 had already rolled,
this would have been a live incident rather than a council objection.

**Mutation-proven both ways, and the second one is the point.** Deleting the guard fails the
test. **Moving it after `DownloadOptimizeAndPrepare` fails it too** — position is the property
here, not presence, because a guard that fires after the git commit is not a guard. That is a
trap this repo has already paid for once (LANDMINES: *"Guarding an asset's provenance UPSERT
is not guarding the asset — the git commit already ran"*), and asserting presence alone would
have reproduced it.

### The through-line across three rounds

Round 1: three of four objections were "you did the right thing and didn't show me".
Round 2: the gating objection was **"you are wrong"** — and it was right. The seats that kept
pressing on the same point across two rounds, against my measured-sounding rebuttal, were
doing exactly what an adversarial reviewer is for. **A confident measurement is not a correct
one, and the tell was that I never re-read what the objection was actually about**: guardian
said *a future caller*, and I kept answering about *current callers*, which is a different
sentence.

## Council round 3 (`abd9b119`) — **APPROVED**, "with 2 advisory objection(s) — none high-severity"

11 seats approve, 2 object at low/medium (advisory only). The verdict is approved, so nothing
below blocked the change — but four of the advisories were *checkable*, so they were checked
rather than filed. That is the whole point of the round: an advisory you discharge is worth
more than an advisory you record.

**`prior_art_librarian` asked for the load-bearing claim to be re-run before the round closed.**
Right to ask — the 11-row figure is what set edit 3's severity, and it was 8 hours old.
Re-run: **2 `detected` + 9 `unresolved` = 11.** Unchanged, so the urgency framing stands.

**`reuse_agent` asked whether "refuse with reason" matched a platform convention or invented
one. It invented one, and the seat was right.** `grep` for `"refused"` across the platform
returned **exactly one occurrence: mine.** The actual house style — named as such in
`ingest_staged_asset_action.go`, the *same asset-deployer agent's* fourth mode — is *"refusals-
as-results per house style"*: the action's own success flag set false plus a `reason`
(`{"ingested": false, …, "reason": …}`). And `deploy_image_asset` already declined its
no-storage-URI case that way. So I had asserted a convention I had not looked for, and then
added a key on top of the real one. **Fixed:** the bespoke `"refused": true` is dropped, the
reason string carries the distinction exactly as the sibling does, and the test now *fails* if
someone reintroduces the key. Mutation-proven.

**`bug_historian` (low) found a real silent-failure hole in my own fix for round 1's catch.**
`DeployedAssetPath` refuses to split a `BrandHeadAssetPaths` value that does not start with
`/`, and then **falls through silently** to the generic purpose-derived path — so an author
adding a relative entry by mistake gets no signal at all. A runtime log would be the wrong
remedy: the map is a *compile-time declaration*, so the malformed case can be made impossible
to **ship** rather than merely noisy when it runs.
`TestBrandHeadAssetPathsAreAbsolute` is that guard; mutation-proven by making the `favicon`
entry relative. Note the shape — **my round-1 fix introduced a new silent path while closing
another**, which is exactly why the "refusing beats inventing" branch needed a test and not
just a comment saying it refuses.

**`guardian` (low) wanted evidence the landmine actually reached `doc_notes`, not just that
`--check` passed afterwards.** Confirmed by reading the row back:
`SELECT … FROM doc_notes WHERE categories ? 'landmine' AND body LIKE '%FIXED AT HEAD, NOT YET
IN THE RUNNING FLEET%'` returns the entry.

**`prior_art_librarian` could not confirm the `IsBrandHeadPurpose` call-site count from a
declarations-only index and asked a human to grep.** Done:
`deploy_image_asset_action.go:136` (the refusal) and the definition. **Exactly one production
caller, and it is the guard** — the predicate 168 had orphaned now has the job the gating
objection created for it.

**`architecture` + `bug_historian` (both medium): `deploy_path` is tracked, not fixed, and it
undermines the "ONE derivation" contract.** Architecture put the knife in precisely:
*"The round already proved that 'currently unreachable' measurements on this exact mechanism
[can be wrong]."* Fair, and after round 2 I have no standing to wave that away. So I measured
it **the way I should have measured the clobber** — including the standing-queue population I
missed last time:

| population | count |
|---|---|
| work items carrying `deploy_path` in `spec` | **0** |
| active agent definitions setting a `deploy_path` **value** | **0** |
| orchestrations with a `deploy_path` **value** (JSON shape, not bare substring) | **0** |

Zero on all three, including the one whose omission caused round 2's error. It stays OPEN as
`bugs_open/179` finding A — measuring it empty is not fixing it, and that is the seats' actual
point.

**`editquality` (low): the ordering test scans source text, so a benign refactor could
false-fail or false-pass it.** Accepted as a stated trade-off rather than fixed. The property
being pinned — *the guard precedes the download and the commit* — is positional, and a
value-level test cannot see position without executing a real download and git commit. The
test says "repoint this rather than deleting it" for exactly this reason. Recorded as a known
limitation, not silently dismissed.

### The arc, for whoever reads this lane cold

Round 1: three of four objections were *"you did the right thing and didn't show me"*.
Round 2: *"you are wrong"* — and it was, about reachability, on a measurement I had defended
twice. Round 3: *"your fix for round 1 opened a smaller silent path, and the convention you
claimed doesn't exist"* — both true, both cheap, both fixed.

**Nine guards are now mutation-proven.** The count is not the point; the point is that four of
the nine exist because a reviewer disagreed with me and I checked instead of arguing.

## LIVE — `v1.0.1229`, both replicas, and the lane closes

The fix went live on a build **another session made**. I did not roll it: `make build-*`
builds from committed HEAD, so my commits rode out on someone else's tag bump. That is the
mechanism working exactly as designed — and it is also the concrete demonstration of why the
ordering exemption's condition (1) was retired on 2026-07-29. I could not have held this back
if I had wanted to, and the *first* commit went live before the council had finished
reviewing it.

**The verification, with a negative control, because a positive one cannot tell you your
binary shipped** (`bugs_open/153`). Same exec, both replicas, after `rollout status` settled:

| control | g7fbt | n8nbj | what it proves |
|---|---|---|---|
| positive `deploy_image_asset` | 5 | 5 | the grep pipeline works |
| **negative `Phase 2E: derived variant deploy path`** | **0** | **0** | the string this change REMOVED is gone ⇒ my binary, not a stale same-tag one |
| `derived asset deploy path` | 1 | 1 | the unified derivation shipped |
| `is a brand-head artefact published at` | 1 | 1 | the refusal guard shipped |

⚠ **`grep -c` returns exit 1 with no output when the count is 0**, so a naive
`$(... | tail -1)` captures an empty string, not `0`. In a loop that reads as "no data" rather
than "the good answer". Default it (`N=${N:-0}`) or you will misread the one control that
matters.

**No regression:** all **24** brand-head artefacts (12 sites × `favicon.png` + `og-card.png`)
serve HTTP 200 after the roll.

⚠ **Verify AFTER `rollout status`, not during.** My first attempt hit a roll in progress and
three of four pod names came back `NotFound` mid-cycle, with one pod answering. One replica
answering during a roll is not "both replicas verified" — and the pod list you enumerated
seconds earlier is already stale.

`bugs_open/168` → `bugs_closed/168`. `bugs_open/179` stays open for the `deploy_path` escape
hatch, measured empty across three populations including the standing queue that round 2
taught me to include.

## The stale items: repaired, and the class filed (owner-directed, 2026-08-02 evening)

**The repair.** All 11 were **false**, not merely stale — every artefact they named serves
HTTP 200, verified on the wire before touching anything. And the *current* check would raise
none of them: dartsonline's asset rows record the published paths exactly, so
`findBrandHeadAssetGaps` is silent; robot-hands' rows record the unresolved template literal
`/assets/images/input-data.asset-key.jpg`, which hits that function's deliberate **third
state** — observed as a `brandHeadProvenanceNote`, never filed, because claiming "never
generated" would be a false claim.

So the repair is `cancelled`, not `complete` (nothing did any work) and not re-derivation
(nothing needs deriving). Script:
`SQL_2026-08-02_retire_stale_brand_head_undeployed_items.sql` — dry run, then a guarded
transaction that **aborts unless exactly 11 rows change**, then verification. Result:
`NOTICE: OK: 11 stale brand-head undeployed_asset items cancelled`, and the follow-up query
returns 0 still open.

⚠ **The reason string points at `bugs_open/152`.** robot-hands' `assets.url` is still wrong;
that defect is real and belongs to the asset-URL rewrite lane. Cancelling the item would
otherwise have taken the only visible signal of it with it — **retiring a false finding must
not retire the true one hiding behind it.**

⚠ `\gset` + a `DO $$ … :'var' … $$` guard does **not** work — psql interpolates textually and
the `$$` quoting collides. Use one `DO $tag$ … GET DIAGNOSTICS n = ROW_COUNT; IF n <> expected
THEN RAISE EXCEPTION …$tag$` block instead. The guard matters more than it looks: between the
dry run you read and the update you run, another session can move the population.

**The class → `RFC_010`.** The framework question the owner asked is not "why were these
stale" but "why can nothing tell you they are". Answer, measured: **49 of 50 discovery checks
are monotonic** — they compute the current truth set on every run, file what they find, and
discard the complement that would let them close what no longer reproduces.

**Prior art existed and I nearly missed it.** `check_backend_unreachable` self-clears — and
gets the safety property right in a way worth copying: **it retracts on a POSITIVE observation
(the probe succeeded), never on an absence of findings.** A naive "close anything not in this
run's results" would be catastrophic, because a check that errored or was silently blinded
returns an empty set indistinguishable from a healthy one — the standing *a gate's 0 findings
has TWO causes* trap.

**A second, independent defect found while measuring:** `unresolved` is excluded from
`idx_swi_dedup` **and** from the one self-clearing check. So it is not terminal, not
deduplicated, and not retractable — 9 of the 11 items were duplicates under 2 `item_key`s for
exactly this reason. It behaves as a landfill and needs a decision of its own.

**Not implemented, deliberately.** `CheckResult` is shared by 50 checks and the runner sits on
the dispatch path; changing either from inside a bug lane is the `bugs_closed/124` shape. Filed
as a design question, with the 909/497/206 measurement stated as a measure of *ignorance* — not
a claim that 497 items are false.

## 2026-08-03 (successor session) — Decision 1's adoption: picking the adopter was the whole job

Picked up from `HANDOFF_2026-08-03_continue_here.md` §4.1. The seam was built, approved and
**inert**; the handoff's own framing is that adoption, not more mechanism, is next.

**Ownership checked two ways first.** `d2d29561` — the originating session — was still writing
at 11:36 (61 mentions of `check_undeployed_assets`), so I checked what it was holding before
touching anything: it wrote the handoff, the summary and the commits, and had gone quiet. I am
its successor, not a competitor. The narrow grep matters: an alternation of four terms including
`discovery_checks/` matched **20** live sessions and told me nothing.

### The candidate I rejected, and why it is the finding of the day

`check_undeployed_assets` was the obvious first adopter — **95 open items, the largest stale
population in the queue**, 50 of them `unresolved`, and its half-2 switch literally contains an
arm that observes `case rowAtPublishedPath:` / *"The deriver's own record of a successful commit.
Deployed."* and does nothing with it. A positive observation, computed and discarded, exactly as
the RFC asks for.

**It is the wrong adopter, and adopting it would have quietly closed real defects fleet-wide.**
`undeployed_asset` has a **second producer**: `write_render_audit_findings_action.go`, which
files `undeployed_asset:<asset_id>` under the *same* item_type, key namespace and handler —
`"deliberate co-dedup (see header)"`, blessed by the owner's 2026-08-02 ruling. Its finding is
*"this image serves broken on a real page"*. `check_undeployed_assets` treats *"a deployed
page_component's rendered_html references the filename"* as healthy.

**Those two are not merely different — they are positively correlated.** You cannot have a broken
`<img>` unless the HTML references its `src`. So *every* render-audit 404 finding sits on an asset
the other check reads as healthy, and adopting retraction there would have retracted every one of
them on the next sweep, with no signal.

The general form, now in `LANDMINES.md`: **the seam addresses rows by `(site_id, item_type,
item_key)`, which is coarser than the producer set that shares that key.** 13 item types have ≥2
Go producers today, and the ruling that blesses co-dedup means that number will grow. The check
before adopting is one grep, and "is my observation a *refutation* of the other producer's
predicate, or merely unrelated to it?" — unrelated is not good enough, because the UPDATE closes
the row either way.

### The adopter chosen: `check_empty_sections`, on evidence

Single Go producer (grepped). Enabled on `completeness-discovery-agent`. 47 open items, oldest
**2026-04-06**, 32 `unresolved`. And — the reason it is the right one — the file **already
contains the predicate**: `emptySectionVerdict`, pure and unit-tested, written for the completion
gate (`VerifyEmptySectionResolved`). So the retraction reuses the platform's existing answer to
"does this section render content?" instead of minting a second one. That is RFC_009's
one-derivation discipline, from this same lane, applied to a different question.

### The measurement, run BEFORE any code was written

| bucket | items | what it means |
|---|---|---|
| every deployed component in the slot renders content | **17** (6 sites, 15 `unresolved`) | **retract** |
| still empty | 19 | finding still reproduces |
| **slot holds no deployed component** | **10** | **ambiguous — must not retract** |
| **mixed slot** | **1** | **must not retract** |

**An absence rule — "close what this run did not find" — would have closed the 10 + 1 = 11 it has
no evidence about.** That is the seam's forbidden inference, with a live number against it.

### MISSTEP 6 — my first census was inflated by a join fan-out, and the tell was a rounding-sized discrepancy

The first pass reported **19** retractable. The earlier census said 47 open items; the join
returned **49 rows**. Two extra rows is exactly the size of discrepancy that reads as "the queue
moved under me" on a live system — and I nearly wrote it off as that.

It was `bugs_open/156`: **`(page_id, slot_name)` is not unique.** One live page holds **three**
deployed components in one slot. Per-item the true figure is **17**, and the fan-out also
exposed a case I had not designed for — a *mixed* slot, where one component of several is still
empty. The rule became "retract only if the slot holds components **and every one** renders
content", which is conservative in the right direction.

⚠ **Generalises: when a per-item count disagrees with its source population by a small number,
that is a JOIN FAN-OUT until proven otherwise, not queue drift.** The check costs one query
(`GROUP BY … HAVING count(*) > 1`). I found this by measurement and only then found the landmine
that already says so — which is the wrong order, and the reason the grep-LANDMINES-for-your-
symbols habit exists.

### The trap inside the file, which no test would have caught

`Run` opened with `if len(sections) == 0 { return &CheckResult{}, nil }`. Appending the
retraction to the end of the function — the natural, minimal edit — produces a mechanism that is
**green in every test that has a finding and inert on every site that has none**. A site with
zero empty sections is the only site that guard fires on, and it is precisely the site whose
stale items need closing. Removed, and mutation M1 is the proof. Also in `LANDMINES.md`, because
every monotonic check adopting `Resolved` has one of these.

### The interaction I nearly shipped without checking: retraction burns a two-strike strike

Grepping `LANDMINES.md` for my symbols (the habit, not an instinct) surfaced
`recurrenceExpected` — and through it `insertWorkItem`'s two-strike block, which counts
`status IN ('complete','failed') AND created_at > NOW() - INTERVAL '7 days'` and brands the third
item on a repeated `item_key` as `unresolved`.

**A retraction writes `complete` onto an existing row.** So it can add a strike, and at two
strikes the next genuine detection of that key is born `unresolved` and undispatchable — the
landfill, created by the mechanism meant to drain it.

Measured rather than argued: of the 17 items that retract, **0 were created within 7 days**
(oldest 2026-04-06, newest 2026-07-19), and **0 keys are already at 2 strikes**. So zero strikes
burn today and no within-cycle suppression fires.

⚠ **Stated in the submission as a live interaction, NOT dismissed as unreachable — because this
lane has already been wrong in exactly that way.** Round 2 of council `abd9b119` refuted a
"measured as currently unreachable" claim I had defended twice; the error was measuring the tap
and not the bath. The honest version: the interaction is real, it is empty *today*, and the
semantics are at least consistent with `CompleteWorkItemAction`, which also writes `complete` on
a fix and also counts as a strike.

### Guards: six mutations, six required failures

M1 reinstate the early return → the zero-findings test fails. M2 drop the no-component refusal →
the ambiguity test fails. M3 let a mixed slot retract → the mixed-slot test fails. M4 set
`AllOfType` → the wide-branch test fails. M5 propagate the retraction fault → the
findings-not-suppressed test fails. M6 ignore the verdict entirely (anti-vacuity) → the
still-empty test fails. All six failed as required and were reverted.

M5 is the one worth keeping in mind: the runner's `continue` on a check error drops that check's
**inserts** too, so propagating a retraction fault would trade a missed closure for a missed
defect. It warns and retracts nothing instead.

### What was deliberately NOT done

No third copy of the closed-status vocabulary. This package already holds two hand-rolled copies
and **they already disagree** — `check_truncated_component.go:163` includes `'cancelled'`,
`check_component_template_corrupted.go:141` omits it. `resolveWorkItems` owns the predicate; the
check owns the observation. A stale row read costs a no-op `UPDATE`, and the worst site names 14
distinct slots in its entire history.

### State

Committed `2287606d1`, council `97923026-2b2d-4925-b9a3-de6f70c49d2b` submitted before the
commit, trailer `Council-Submitted:`. **NOT live** — both replicas run `v1.0.1238`; HEAD's
`IMAGE_TAG` is `v1.0.1239`, bumped by another session.

⚠ **The roll is deliberately deferred: a roll kills an in-flight council**, and mine is running.
Verification (positive control + both replicas) and the first real retraction measurement follow
the verdict. ⚠ **This change removes no string literal, so it has no natural negative control** —
the `bugs_open/153` recipe does not fully apply and saying so is better than substituting a
weaker control and calling it the same thing.

⚠ Environmental, cost me a confusing build failure: **`/tmp` is a 16G tmpfs at 97%**, 14G of it
session scratchpads, and the Go linker writes there — `go build ./...` fails with
`mapping output file failed: no space left on device`, which reads like a code error. Fixed
non-destructively with `TMPDIR=/home/ant/.cache/buildtmp` (235G free on `/`) rather than deleting
other sessions' scratch.

## Council `97923026` — **APPROVED at round 1**, 15 seats, 3 advisory objections (none high)

Dispatch was immediate this time, not the measured 29-minute queue: `fix_plan` persisted 10:55:19
UTC, verdict at 11:03. Do not read that as the new normal — one sample.

**Four of the advisories were checkable, so they were checked.** The pattern from this lane's
earlier rounds held again: most objections were not "you did the wrong thing".

**`editquality` (medium) — the one that needed code, and it was right in the way that matters.**
The retraction read `spec->>'page_id'` while `site_work_items` carries a **first-class `page_id`
column**, which is the standing landmine about a reader blind to the column its creator populates.
Measured: over all **58** `empty_section` rows ever, **0** have a NULL column, **0** disagree with
the spec key, and that holds for all 47 open ones. So the objection is **empirically empty today**
— and the `COALESCE(page_id::text, spec->>'page_id')` went in anyway, because *measured empty is
not unreachable* and this lane has already been wrong in exactly that way (round 2 of `abd9b119`).
A future filing change that sets the column and forgets the spec key would otherwise silently stop
retracting, with no signal. Mutation-proven, M7.

**`prior_art_librarian` (medium) + `guardian` (low) — "the single-producer claim rests on a grep
the council cannot verify".** Fair: it was load-bearing for scoping this as a contained fix. So I
corroborated it from the **data** side, which is independent of my grep — every `empty_section` row
ever was created by `completeness-discovery-agent` or `generic`, i.e. **two agent types running one
check, not two producers**.

**`prior_art_librarian` (low) — "the seam-is-inert claim is your own measurement".** Re-verified:
**0** agents enable `backend_unreachable`, **0** items in all history, and **0** rows in
`site_work_items` have ever carried `result.resolved_at`.

**Four seats (`editquality`, `bug_historian`, `tooling_provenance`, `architecture`) — "you say you
will file the hazard to LANDMINES/register, but no edit does it".** Fair on the record and
**already discharged in fact**: the gate **refuses docs client-side**, so a docs edit could never
have appeared in the plan it reviewed. They shipped in `d983de570`.

⚠ **A structural note worth keeping: the council cannot see the half of this lane's work that the
gate refuses to accept.** Round 1's lesson was "evidence you hold and do not cite is evidence you
do not have"; this is its sharper form — evidence you *cannot* cite, because the submission schema
excludes it. Four seats independently objected to a gap that did not exist. The fix is not to
argue: it is to say in the RATIONALE that the docs exist and name their commit, since prose is the
only channel the gate leaves open for them.

**The `untouched-twin` pattern-check advisory, answered here because it fired at commit time and
forward-only forbids an amend.** It flagged that `findResolvedEmptySections` changed while its twin
`findEmptySections` did not. **Deliberate, and the twin does not share the defect**: the twin reads
`page_components` and never touches `site_work_items`, so it has no spec-versus-column question to
get wrong. It is the *writer* that populates both fields; the ambiguity exists only on the read
side, which is the side that changed.

### STILL OPEN, and owed to a human — three seats asked for it by name

`guardian` (medium), `improvement_guardian` (low) and `bug_historian` (low) **independently** ask
for explicit sign-off on the two-strike interaction: a retraction writes `complete` onto an
existing row and feeds `insertWorkItem`'s terminal-row counter **identically to a real fix**.
Guardian's wording is the precise one — acceptable *"only if a human signs off that
'resolved-by-observation' and 'resolved-by-handler' should count identically toward the strike
counter — otherwise this needs a follow-up, not silent acceptance."*

Measured **0 of 17 affected today**; structurally unprevented. **Not accepted silently — raised
with the owner.** Three seats converging on one question from three different remits is the
strongest signal this council produces, and this lane's round 2 is the standing proof that the
seats pressing hardest are usually pressing on something real.

## LIVE AND PROVEN — `v1.0.1243`, and the first four retractions in the platform's history

**The roll came from another session's build**, as it did for `bugs_closed/168` hours earlier.
Third time in two days; it is the mechanism working as designed, not a coincidence worth noting
again.

**Pod-verified, both replicas, and the negative-control problem solved rather than waved at.**
This change removes no string literal, so `bugs_open/153`'s positive+negative recipe does not
apply. The substitute, which is *stronger* for a purely additive change: **the dated BEFORE
measurement.** `re-observed healthy: all` grepped **0 on both replicas of `v1.0.1238`** (taken
deliberately before the roll) and **1 on both replicas of `v1.0.1243`**. A stale same-tag binary
would still read 0, so the transition is the proof. The post-council hardening string
(`COALESCE(page_id::text`) also greps 1, which proves the build postdates `27891fab8` and not
merely the first commit. Positive control 1, invented negative control 0, same exec.

⚠ **Generalises, and belongs to anyone verifying an additive change: if your change removes no
string, take the pod-grep BEFORE the roll and date it.** It costs one command and converts an
unfalsifiable "my string is present" into a 0→1 transition.

### The measurement that was the whole point

Dispatched a `completeness-discovery-agent` sweep at `leopardessconsulting.co.uk` (4 retractable,
pre-measured), correlation `4401d952-4b1b-472c-b364-4d9fedb369f1`. Pre-dispatch coverage check
first: 0 claimed/in-progress items on the candidate sites, 0 running orchestrations. Pods were 34
minutes old, well past the ~300s post-restart window in which a spawn is silently dropped.

**Fleet-wide `result ? 'resolved_at'`: 0 → 4.** All four `resolved_by: empty_sections`, each
carrying its reason. **The items were raised 2026-04-14 and 2026-04-23** — over three months
stale, on `info-card-grid` ×3 and `tool-cta`, and nothing in this platform could close them until
today.

**The discrimination control is the load-bearing half, and it is the number I would want if I
were reviewing this.** The same sweep left **6 of that site's 10** `empty_section` rows open — 3
`unresolved`, 2 `needs_human_review`, 1 `detected`. It closed what it had evidence for and
nothing else. A mechanism that had closed all 10 would have looked identical in the headline
figure and been the catastrophe the RFC's standing condition exists to prevent.

14 retractable items remain across 5 unswept sites; they will close on those sites' next sweeps,
which is the honest way to say "we expect it to keep working" without claiming it already has.

### Owner ruling recorded

**The two-strike interaction is ACCEPTED AS-IS and tracked as `RFC_010` Q1**, not fixed here —
`insertWorkItem` is on the insert path of every work item in the estate, and changing it from
inside a check adoption is the `bugs_closed/124` shape. Three seats asked for the decision; the
decision is made and written down where the next adopter will find it.

## SECOND ADOPTER — `check_required_fields_missing`, and the trap the first adopter's landmine would not have caught

Picking up `HANDOFF_2026-08-03b` §4.1. Took §4.2 first because it is one query: **fleet
`result ? 'resolved_at'` is still exactly 4**, all four the original
`leopardessconsulting.co.uk` retractions from sweep `4401d952`. The 14 retractable items on the
5 unswept sites have **not** closed. Per the handoff's own framing that is a **scheduling
question, not a defect** — those sites have not been swept since. Not chased; recorded so the
next session does not re-measure it hoping for a different number. [MEASURED 2026-08-03]

### Choosing the adopter — the survey, not the table

The handoff offers four single-producer candidates. Ran §3's mandatory producer grep on all
four: each has exactly **one** Go producer. Corroborated from the DATA side too, which is
independent of my grep and is the check the council's `prior_art_librarian` seat asked for last
round — `created_by` over every row ever:

| item_type | Go producers | created_by (all history) | open |
|---|---|---|---|
| `required_fields_missing` | 1 | `completeness-discovery-agent` 37, `generic` 33 | 59 |
| `cta_names_unknown_destination` | 1 | `completeness-discovery-agent` 46, `generic` 61 | 107 |
| `needs_sprite_css` | 1 | `design-discovery-agent` 2, `generic` 9 | 10 |
| `voice_tells` | 1 | `generic` 25 | 25 |

Same shape as `empty_section`: two agent types running one check, not two producers.

**Did not take the biggest population.** `cta_names_unknown_destination` has 107 open rows, but
`check_misdirected_cta` files **two** item types — the other is `page_rerender`, which is about
as multi-producer as this estate gets — and its retraction predicate would have to re-extract
anchors and re-run the whole match index. Its natural "fixed" signal is also an **absence** (the
anchor is gone from the HTML), which is the shape the seam exists to refuse. That is a real
adopter, but it is not the cheap second one, and doing it badly would be worse than not doing it.

**Sized `required_fields_missing` before writing a line of code**, simulating the predicate in
SQL over the live queue [MEASURED 2026-08-03]:

| verdict | items | sites |
|---|---|---|
| REFUSE: still missing | 50 | 4 |
| **RETRACTABLE** | **6** | **2** |
| REFUSE: no deployed component | 3 | 1 |

Six is a modest number and it is the honest one — the 50 are the discrimination control. The
argument for adopting here is not the six: **this check is FLAG-ONLY.** No handler agent, items
born `needs_human_review`, so `CompleteWorkItemAction` is never reached and until now a
`required_fields_missing` item could not close by **any** mechanism except a hand on the
database. For this type retraction is not a faster path, it is the *only* path.

### The trap, and why the existing landmine would not have caught it

`LANDMINES.md` already carries the first adopter's entry: *a monotonic check's
`if len(findings) == 0 { return }` makes retraction inert*. I followed it, read the top of
`Run`, and found **no leading guard**. The check looked safe.

It was not. The retraction-skipping `return result, nil` sits **in the middle of the row loop**,
fired by `maxRequiredFieldFlagsPerPass` — a noise cap at 25 findings:

```go
if emitted >= maxRequiredFieldFlagsPerPass { ...; return result, nil }   // was
if emitted >= maxRequiredFieldFlagsPerPass { ...; break }                // now
```

Same defect, different disguise: inert on exactly the badly-shaped sites carrying the most stale
items, green in every test that stays under 25 findings (all of them), and **invisible to a grep
for the documented shape**. The cap's `return` was correct while the check could only file.

⚠ **This is a correction to the scope of an existing landmine, not a new instance of it.** Filed
as its own entry: read EVERY exit between the scan and the retraction, not just the leading one.

### The refusal unique to this check: an unreadable schema computes to HEALTHY

The predicate is *"the schema declares these fields required and content_data lacks them"*. So a
component whose `input_schema` is NULL, unparseable, or in an unrecognised dialect yields **no
required fields at all**, and the naive reading of that is "nothing missing → filled". It is the
exact inverse: the observation could not be made. Copying the filing half's `continue` straight
across does it, because there `continue` means *do not file* and here it must mean *do not count
as observed*. Left unguarded, the retraction would fire hardest precisely when a component's
schema had been dropped — silent loss arriving by a route that reads as success.

Kept honest structurally rather than by vigilance: **`obs.healthy++` is reachable from exactly
one line** — where `missingRequiredValueFields` actually ran and returned nothing. Every other
path falls through to the `healthy != deployed` gate, so a future refusal is added by *not
counting* rather than by remembering to veto.

Also refused **runtime-fill shells**, which is deliberately NARROWER than the negation of the
filing predicate. The filing half skips them because a browser loader supplies the content —
that is a reason not to FILE; it is not evidence the fields arrived.

### Mutation round: 8 mutations, and the one that PASSED is the interesting one

Harness at `scratchpad/mutate.py` — applies each mutation, requires the named test to fail,
restores. First run: **7 of 8 caught, M2 passed.**

M2 deletes the `if !componentID.Valid { continue }` refusal — the "slot holds no deployed
component" guard. Every test stayed green. **It was caught by a guard in SERIES**: today's query
joins `content_components` *through* `page_components`, so a LEFT JOIN miss always arrives with a
NULL schema too, and the schema refusal shadowed it. The realistic test row could not tell the
two guards apart.

Fixed with a deliberately **synthetic** row — a join miss whose other columns look perfectly
healthy — which pins the componentID guard on its own, so a future change to the join cannot
quietly make correctness depend on the other guard. Second run: **8 of 8, each caught by its own
named test.** A guard nothing fails on is decoration, and "the mutation passed" is not the same
as "the guard is redundant".

### `page_id`: the same landmine as last round, pointing the other way

Council `97923026`'s `editquality` seat made the first adopter read the first-class `page_id`
column ahead of `spec->>'page_id'`. Checked the same thing here and it inverts: **all 70
`required_fields_missing` rows ever written have a NULL `page_id` column** and a populated spec
key, 0 disagreements — because this check's filing half never sets `WorkItemSpec.PageID`
[MEASURED 2026-08-03]. So for `empty_section` the COALESCE pins a preference; **here the spec
fallback is the arm that actually fires.** Column still read first: it is the first-class one, and
a future filing change that populates it would otherwise silently stop retraction with no signal.

**Deliberately NOT fixed here.** Setting `PageID` on the filing half is a change to what this
check *writes*, needs its own justification and test, and the retraction's correctness does not
depend on it. Recorded as a follow-up rather than smuggled into a retraction change — the
`bugs_closed/124` shape.

### State

Build clean, `go vet` clean, full `platform/orchestration/actions/...` suite green.
⚠ Two new LANDMINES entries synced (`--apply`, 948 rows). Both came back
`NEEDS_VERIFICATION`, and the `landmine-verifier` **cannot** discharge them: every symbol they
name (`findResolvedRequiredFields`, `TestRequiredFieldsRetractionSurvivesThePerPassCap`)
postdates the code index's freeze at `d98010e8b` (2026-07-28), so it reads them as ABSENT. That
is the known index landmine, not a defect in the entries. Saying so rather than running it for a
predictable NEEDS_HUMAN_REVIEW.
⚠ `/tmp` tmpfs still near-full — `TMPDIR=/home/ant/.cache/buildtmp` throughout, as §8 of the
handoff warns.

## Council `64430363` — **APPROVED at round 1**, 14 seats, 1 objecting seat, none high

Dispatch was immediate again (submitted ~20:13 UTC, verdict 20:23) — second sample in a row, but
still not the measured 29-minute norm. Two samples is not a rate.

**Three objections were checkable, so they were checked rather than filed.** That is this lane's
standing pattern and it held for a third round.

### `bug_historian` (MEDIUM) — "second per-call-site fix of the same landmine, and nothing audits the rest"

The sharpest objection of the round, and correct in shape: *"this is now the SECOND time this
exact landmine has been found and fixed per-call-site, with no stated intent to sweep other
per-pass-capped checks before they adopt the seam."* Its `missing` field asked the question
outright, so I answered it with a census rather than an intention [MEASURED 2026-08-03]:

| check | cap | exit | seam adopted? |
|---|---|---|---|
| `check_componentless_pages.go` | `componentlessMaxPerPass` | `break` | no |
| `check_component_template_corrupted.go` | `maxTemplateRegensPerPass` | `break` | no |
| `check_section_source_drift.go` | `maxSectionDriftFlagsPerPass` | `break` | no |
| `check_required_fields_missing.go` | `maxRequiredFieldFlagsPerPass` | ~~`return`~~ → `break` | **yes (this change)** |
| **`check_image_source_unsatisfiable.go`** | `maxUnsatisfiableFlagsPerPass` | **`return result, nil`** | no |

**Exactly one other site is armed**, and it is **inert today**: `check_image_source_unsatisfiable`
does not populate `Resolved`, and while a check can only file, its `return` is *correct*. So there
is nothing to fix there now — and deliberately nothing was. **Fixing it pre-emptively would be a
behaviour change with no benefit and no test**, on a check I am not otherwise touching. What was
owed is that the next person cannot miss it: the census is appended to the `LANDMINES.md` entry,
with the re-run command, and the entry now says plainly that **the adoption commit is the commit
that must also change that `return` to a `break`.**

### `bug_historian` (LOW) — "the single-increment discipline is held by review, not by mechanism"

Fair, and the fix was cheap enough that arguing would have cost more.
`TestHealthyIsIncrementedFromExactlyOnePlace` parses `check_required_fields_missing.go` with
**`go/ast`** and asserts `.healthy` has exactly one write site.

Deliberately not a grep: this file's comments discuss `obs.healthy` repeatedly, so a grep would
match prose and pass for the wrong reason. It also cannot pass vacuously — it `t.Fatal`s if it
cannot find the function, and again if it counts **zero** writes, because a needle that stopped
matching produces a green test that can never fail again. Mutation M9 (a second `obs.healthy++`
on the unreadable-schema branch) → fails with *".healthy is written from 2 places, want exactly
1"*; restored → green. **9 mutations now, 9 required failures.**

### `editquality` (LOW) — "`build_status='deployed'` is a documented trap on other tables"

The seat was right to ask (`pages.build_status` undercounts / outlives; `site_components` never
reaches `'deployed'` at all), and the answer turned out to be **structural, not statistical** —
which is a better answer than the row count I first reached for.

`\d page_components` [MEASURED 2026-08-03]: there is **no separate `status` column**. Its one
status column is `build_status`, whose CHECK constraint is
`deployed | pending | approved | removed | needs_rebuild` — **retirement is `removed`, INSIDE the
column the predicate filters on.** The `pages` trap needs *two* columns to drift apart (archiving
sets `pages.status` and leaves `pages.build_status` alone); `page_components` has only one, so the
trap cannot reproduce here. Live values: deployed 962, approved 216, pending 22.

Also measured the page-level interaction it implies: **all 59 open items sit on `pages.status =
'active'`**, 0 on archived pages. ⚠ **Deliberately did NOT add a `p.status='active'` guard**, and
the reason is symmetry rather than the zero: the FILING half joins `pages` with no status filter,
so it still files on an archived page. A retraction narrower than its own filer would refuse to
close exactly the things it would still raise — the wrong asymmetry, and a worse defect than the
one being guarded against.

### The two seats that asked a human to confirm something I could confirm myself

`tooling_provenance` and `prior_art_librarian` both recorded that they **could not** verify the
docs commit or the cited symbols — `doc_plans` is not in their schema and `code_checks` cannot
see function bodies. Confirmed here so the record is not left hanging: `b312c409a` landed the two
`LANDMINES.md` entries plus the lane's NOTES and README (3 files, 179 insertions), and
`findResolvedEmptySections`, `SchemaContentFields`, `missingRequiredValueFields` and
`CheckResult.Resolved` all exist and are the shapes cited.

`reuse_agent` asked whether this was "a straight resubmission with unresolved prior objections
carried forward" — it is not: no `RESUBMIT_CORR`, and council `97923026` was a **different
change** (`check_empty_sections`) in the same lane. The docs-gap objection it noticed in my
rationale was me *pre-empting* last round's finding, not carrying it forward.

### The `architecture` seat, and why its verdict is the reassuring one

`ARCHITECTURE_SIGNAL: point_fix`, approve. Its reasoning is the answer to the question this lane
has been circling since `bugs_closed/124`: *"no shared contract added, no exported symbol other
packages depend on (`findResolvedRequiredFields` is file-local), single-package single-file change,
fails every needs_rfc trigger."* And it explicitly rewarded the **declining**: *"the author
explicitly declined to fix the filing half's PageID gap inside this change, correctly naming that
as the `bugs_closed/124` shape — that is the right instinct for this seat to reward."*

⚠ Worth carrying: **the seat that vetoes scope creep also credits you for the thing you did NOT
do, but only if you SAY you did not do it and why.** A silent omission reads as an oversight; a
stated deferral reads as judgement. Same edit, different verdict.

### State

Committed `ba3aae47f` (code, `Council-Submitted:`), `b312c409a` (docs), `9cd9c5227` (handoff).
Verdict read and approved, so the follow-up commit carries `Council-Reviewed:`.
**NOT LIVE.** Both replicas run `v1.0.1244`.

⚠ **Pre-roll baseline TAKEN AND DATED — the one measurement that cannot be recovered afterwards.**
`re-observed filled: all` greps **0 on both replicas of `v1.0.1244`** (2026-08-03), with the first
adopter's `re-observed healthy: all` at **1** on both as a positive control proving the grep and
the pipeline work. After the next roll the same command must read **1**; a stale same-tag binary
would still read 0, so the 0→1 transition is the proof. This change removes no string literal, so
`bugs_open/153`'s positive+negative recipe does not apply — and the substitute only exists if you
take it BEFORE the roll, which is why it is here and not in the next session's plan.
