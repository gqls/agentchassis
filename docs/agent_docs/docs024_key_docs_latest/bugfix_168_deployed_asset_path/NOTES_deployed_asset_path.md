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

## LIVE on `v1.0.1250` — and a CORRECTION that undercuts this adoption's whole justification

### First, the deploy verification, which did hold

**Pod-verified, both replicas, `v1.0.1250` (pods started 2026-08-04T10:29Z).** The dated 0→1
transition worked exactly as designed: `re-observed filled` was **0 on both replicas of
`v1.0.1244`** yesterday and is **1 on both of `v1.0.1250`** today. First adopter's
`re-observed healthy` = 1 as positive control.

⚠ **My first probe reported the cap-fix string as ABSENT, and it was my probe that was wrong.**
The log line contains an **em-dash**, and passing it through `kubectl exec -- sh -c "grep '…—…'"`
mangled the multi-byte character, so `grep -c` returned 0. Re-probed with ASCII-only substrings
(`filing stops here`, `retraction still runs`) → **1** on both replicas. **Grep the pod for
ASCII-only substrings; never let a non-ASCII character cross the `exec` boundary** — the failure
is silent and reads exactly like "my change did not ship".

⚠ Also worth saying plainly: **the negative control I reached for was incapable of failing.** The
old cap log line is a strict PREFIX of the new one, so grepping for it matches either binary.
There is no valid negative control for this change, which is what "purely additive" means. The
dated 0→1 transition is the whole proof, and it only exists because it was taken before the roll.

### > **CORRECTED 2026-08-04 — the central claim of this adoption is FALSE.**

> Everything above and below in the previous section asserting that
> `required_fields_missing` **"could not close by ANY mechanism except a human hand on the
> database"** is **wrong**, and it was the main justification for adopting the seam here — it is
> in the commit message (`ba3aae47f`), in the council rationale, and in the handoff.
>
> **`platform/orchestration/actions/revalidate_review_queue_action.go:161` already does this:**
> `"required_fields_missing": revalidateNamedFields("missing_fields")`. It selects on
> `status = 'needs_human_review'`, which is the status every one of these items is **born in**,
> so it reaches **100%** of the population I called unreachable. Its predicate is the same as
> mine — *"every field this item reports missing (%s) is populated on the deployed component"* —
> and it has been running since at least **2026-07-27**.
>
> **In one respect it is better than mine**: it refuses when `content_data` is empty
> (*"the component renders from a template, a DERIVED source or a static fallback… content_data
> cannot answer the question, so we do not pretend it can"*) — a refusal I did not have. It also
> deliberately keys on `(page_name, slot)` and **never** on `spec.component_id`, having measured
> that component ids are not stable across re-renders (11 of 45 `required_fields_missing` items
> resolved to a component that no longer exists when keyed that way).
>
> **What caught it:** re-running the sizing after the roll. My 6 retractable items were **0**, and
> they had gone `complete` at **08:37Z, two hours before my pods started**, carrying a
> `result.revalidation` block. Full write-up, including the ten-second query that would have
> caught it before I wrote a line, is in `WRONG_CALLS.md` (2026-08-04).

### What is actually true now

- **The change is redundant, not harmful.** `resolveWorkItems` skips rows already in
  `workItemClosedStatuses`, so the two closers cannot double-close or clobber one another. Nothing
  is at risk while this sits in the tree.
- **It has retracted nothing and, on current populations, will not.** [MEASURED 2026-08-04] 0
  retractable fleet-wide; the revalidator gets there first, on a cadence that is not mine.
- **The one real gap is narrow and empty today.** The revalidator only sees `needs_human_review`;
  `resolveWorkItems` deliberately also closes `unresolved` and `failed` (RFC_010 Decision 2, on
  the ground that neither means "this stopped being a problem"). **0 of 95** rows are in those
  statuses today, so the gap is real and unexercised.
- **The `§3` producer check has a hole, and this is the finding worth keeping.** `HANDOFF_2026-08-03b`
  makes counting *producers* mandatory before adopting the seam. It says nothing about counting
  **closers** — and a second closer is the same hazard from the other end. I ran the producer
  check thoroughly, two independent ways, and it could never have surfaced this.

### What I am NOT doing, and why

**Not reverting unilaterally.** A revert is also a change, it needs its own justification, and the
honest position is that I do not know whether the narrow `unresolved`/`failed` gap is worth a
second closer or whether that gap should be closed by teaching the *revalidator* those statuses
instead — which is very likely the better design, since it already has the sharper refusals. That
is an owner call and it is written up in the handoff.

**Not resubmitting to the council.** The verdict was APPROVED on a rationale containing a false
premise; that is a defect in the submission, not in the verdict, and no resubmission un-approves
it. The correction belongs where the change lives, which is here, in `WRONG_CALLS.md`, and in the
handoff — per the standing rule that a refuted claim is recorded as a visible correction naming
what caught it.

## OWNER OPTION A — revert the duplicate, fix the gap at the better mechanism

Owner ruled **option A** on 2026-08-04. Both halves committed in `b4c64f433`, council
`1cec55d2-5928-4785-8598-dfd7870a39d8` submitted before the commit.

### Half 1 — the revert, done as a revert

`git checkout ba3aae47f~1 --` over both files, then verified: **`git diff ba3aae47f~1` over the
pair is EMPTY.** Byte-identical to the pre-adoption state, not a hand-unpick that leaves
fragments. That also restores the per-pass cap's `return result, nil` — which is **correct again**
the moment the check cannot retract, exactly as `check_image_source_unsatisfiable` is correct
today. So the `LANDMINES.md` cap census was corrected: **two** sites are armed-but-inert now, not
one. Leaving that entry saying "fixed" would have been the more comfortable option and a false
one.

### Half 2 — the gap, and it is sharper than "the gap is empty"

The handoff justified option A on a gap that was **empty today**. Reading the mechanism made it
concrete and *imminent* instead:

`revalidate_review_queue` selects `status='needs_human_review'` only. Every close it makes writes
`complete`, which feeds `insertWorkItem`'s two-strike counter (`status IN ('complete','failed') AND
created_at > NOW() - INTERVAL '7 days'`, `load_work_item_actions.go:1237`). Discovery re-raises the
finding next pass. **After the second close inside seven days the third re-raise is BORN
`unresolved`** — a status the sweep could not see. **The sweep's own success rate generates the
rows it then goes blind to.**

Not hypothetical: **no discovery check sets `recurrenceExpected`** (checked — only a comment in
`remit.go` mentions it), so the counter is not skipped for these items, and **5
`required_fields_missing` item_keys already sit at 1 strike** [MEASURED 2026-08-04]. One close
each from the blind state.

### THREE gates, not one — and the second and third are invisible from the first

`status` is checked in **three** places: the selection in `loadParkedReviewItems`, and **two
write-time CAS guards** in `recordRevalidation` (they re-check status so a row that moved
underneath the sweep is not clobbered). **Widening only the selection would select the new rows
and then silently update none of them** — the sweep would report `scanned: N, closed: 0` and read
as "nothing to do". That is the `input_mapping`-vs-`RETURNING` shape already in `LANDMINES.md`: a
dispatcher with two gates, where fixing the one you can see leaves the key dropped. The
SessionStart hook put that entry in front of me this session; it earned its place.

All three now interpolate one package-level `workItemRevalidatableStatuses` via the estate's
existing `sqlInList` idiom, placed beside its two siblings because that file's own comment says
the lists sit together **so the differences are visible rather than discovered**.

### ⚠ `failed` IS EXCLUDED — and measuring the OTHER consumers is the only reason I noticed

**My first draft included it**, straight from RFC_010 Decision 2, which pairs `unresolved` and
`failed`. Then I measured the blast radius across **all four** covered types instead of the one I
was reasoning from:

| item_type | `unresolved` | `failed` |
|---|---|---|
| required_fields_missing | 0 | 0 |
| needs_section_data | 0 | 0 |
| unresolved_cta | 0 | 0 |
| **needs_page** | **1** | **17** |

Those 17 are **precisely the population this action's own header defers by name**: *"failures
parked by `FailWorkItemAction`'s `status_override` branch, which does not increment attempt_count
so they neither retry nor age out. Real defect, **open owner decision (033 D2), not this
sweep's**."* Including `failed` would have quietly overruled a stated deferral and an open owner
decision **from inside an unrelated change** — the `bugs_closed/124` shape, and I would have done
it while writing the landmine warning against exactly that.

It was never needed either: **the two-strike counter brands items `unresolved`, never `failed`.**
The argument only ever supported one of the pair; I had copied the other by association.

**Blast radius as narrowed: 1 row.** Preventive, not a drain. Said plainly rather than dressed up.

⚠ **This is the second time in two days that reasoning from MY consumer nearly damaged ANOTHER
consumer of a shared mechanism** — first the missed closer, now the `failed` population. The
transferable form: *when you widen a shared mechanism, the blast radius is rarely on the consumer
you are thinking about.* Both instances are now in `LANDMINES.md`.

### Mutation round — and the first one was INVALID

Four mutations, each required to fail a named test. **The first run's harness was broken and I
nearly recorded its output.** Two mutations reported NOT CAUGHT; the cause was that the pattern
`"unresolved",` **also matches `workItemTerminalStatuses` twenty lines above**, so
`str.replace(old,new,1)` mutated the **wrong list** and the test correctly passed. Re-anchored on
the whole `var` block: **all four caught**, each by its named test.

⚠ Worse, the first harness lost its backups (the scratchpad had been cleared under `/tmp`
pressure) and left **two mutations applied in the tree**, so a later mutation ran against an
already-red baseline and its "CAUGHT" was meaningless. Repaired by hand — `git checkout` would
have discarded the real work too. **A mutation harness must (a) hold its baseline in memory, not
in `/tmp`, (b) re-assert the baseline is GREEN before each mutation, and (c) anchor on a pattern
that cannot match elsewhere in the file.** The rewritten harness does all three.

### State

`b4c64f433`. Build clean, `platform/orchestration/...` suite green. **Not live** — both replicas
run `v1.0.1251`, which still carries the reverted adoption (`re-observed filled` greps 1). The
revert and the widening both need the next roll.

## Council `1cec55d2` — **APPROVED**, and four objections that were checked rather than filed

### `editquality` (MEDIUM) — my own negative control was the shape I keep writing landmines about

*"`strings.Count(body, "status = 'needs_human_review'") != 0` is exactly the shape LANDMINES.md
warns about — the comment explaining a removal makes the removed symbol's negative control
non-zero. If any surviving doc comment describes the old predicate in prose, this test fails on a
correct fix."*

**Right, and it needed code.** Nothing in the file carries that literal in prose today — but the
test's whole job is to survive future edits, and **the most likely future edit is a comment
explaining why the literal went.** I would have written that comment myself.

Fixed by stripping `//` lines before the negative count. **Proven in BOTH directions**, which is
the part worth copying: appended a comment containing the old literal → test still **passes**;
reverted a real code gate to the literal → test **fails**, naming both the count and the survivor.
A one-directional proof here would have been worthless.

### `guardian` (MEDIUM) — "is any of the four types filed by more than one producer?" — IT IS

The seat asked the co-dedup question the previous round's landmine names, and it **found
something** [MEASURED 2026-08-04]:

| type | Go producers | rows my widening exposes |
|---|---|---|
| `required_fields_missing` | 1 | 0 |
| `needs_section_data` | 0 via `ItemType:` (inline INSERT — `plan_sections`' `createDeferredItems`) | 0 |
| `unresolved_cta` | 0 via `ItemType:` (inline — `resolve_internal_links_action.go:257`) | 0 |
| **`needs_page`** | **5** | **1** |

**`needs_page` is genuinely multi-producer, and it is the only one of the four with a live row in
scope.** Stated precisely: the risk is **pre-existing in kind, not introduced here** — the
revalidator already applies `revalidateNeedsPage` to that type for `needs_human_review` rows, and
this change extends its reach by exactly **one** row. But the guardian is right that the
submission did not say so, and "measured needs_page's revalidator difference" is not the same as
"counted its producers".

⚠ **Recorded for the `bugs_open/187` lane, which owns `needs_page`'s revalidator** (added
2026-08-03): one `unresolved` row of your type becomes closable by your revalidator after the next
roll. Not asking you to act — telling you, per the owner ruling that other consumers must be told
rather than merely measured.

### `debug_historian` (MEDIUM) — "no deploy-verification step is named"

Fair: the plan named none. Taken now, and **this change finally has a REAL negative control** —
the first in this lane — because the revert **removes** a string rather than only adding one.

**Pre-roll baseline, dated 2026-08-04, `v1.0.1251`, both replicas:**
`re-observed filled` = **1** (the reverted adoption is in the running binary) and must read **0**
after the roll. Positive control `auto:revalidated` = **2** on both, and must stay non-zero — it
proves the grep and the pipeline, so a 0/0 result reads as "broken probe", not "clean revert".

This is the inverse of the 08-03 measurement, where the change was purely additive and I had to
say plainly that no negative control existed. **A revert is the one change shape where the
`bugs_open/153` recipe applies in full.**

### The two answered from the tree

`sqlInList` exists (`work_items_common.go:151`) — the seat could not see it and said so. And the
revert's byte-identity claim: `git diff ba3aae47f~1` over both files returns **0 lines**, which is
stronger than the "cheap existence check" asked for.

⚠ `guardian` and `tooling_provenance` both recorded that they **could not** verify things I can
verify in seconds — Go source and a separate docs commit. **Third round running.** The pattern is
now established enough to state as a rule: *the gate cannot read function bodies or other commits,
so anything resting on either must be measured by the submitter and quoted, not asserted.*

## OPTION A IS LIVE AND PROVEN — `v1.0.1252`, both replicas, 2026-08-05

**The first change in this lane with a full positive+negative control on BOTH halves.** Baseline
was taken 2026-08-04 on `v1.0.1251` and dated; this completes it.

| probe | baseline (`v1.0.1251`) | now (`v1.0.1252`) | verdict |
|---|---|---|---|
| `re-observed filled` (revert removed it) | **1** | **0** | ✅ negative control fired |
| `findResolvedRequiredFields` | present | **0** | ✅ symbol gone |
| `auto:revalidated` (untouched) | **2** | **2** | ✅ positive control — the probe works |
| `AND status IN (%s)` (widening added it) | 0 | **3** | ✅ and see below |
| `AND status = 'needs_human_review'` (widening removed it) | present | **0** | ✅ |

⚠ **The `3` is the part worth keeping.** `AND status IN (%s)` greps **exactly three** times in the
shipped binary — which is the same assertion `TestAllThreeStatusGatesUseTheSharedList` makes about
the source, now confirmed **in the artefact that is actually running**. A unit test proves the
source; this proves the binary. They are different claims and this lane has been bitten by
conflating them.

⚠ **A revert is the one change shape where `bugs_open/153`'s recipe applies in full**, because it
removes a string. Every other change in this lane was additive and had to rely on a dated
before-measurement instead. Worth remembering when planning a verification: *what shape is my
change?* decides which proof is available, and you must decide before the roll.

### The sweep does act — checked rather than assumed

`diagnosis-review-queue-revalidator`, step `sweep`, **`dry_run: false`** [MEASURED 2026-08-05].
Not the InputSpec default (`true`), so this needed reading the live row rather than the code —
the standing "seed is not the system" lesson. **33 rows closed `auto:revalidated` all-history**,
latest 2026-08-04 08:37, i.e. **before** today's roll. So the widening has not yet had a pass.

### The number that vindicates excluding `failed`

| | 2026-08-04 | 2026-08-05 |
|---|---|---|
| `needs_page` **`failed`** | 17 | **21** |
| `needs_page` `unresolved` | 1 | 1 |
| `required_fields_missing` keys at 1 strike | 5 | 5 |

**The deferred 033 D2 population grew by 4 in a day.** Had I kept `failed` in the list — as my
first draft did, copying RFC_010 Decision 2's pairing — this widening would now be pointing an
auto-closer at **21 rows of an actively-growing population that an open owner decision covers**,
and growing. The measurement that stopped it was checking the *other* consumers rather than the
one I was reasoning from. **That is the third time in three days that the blast radius of a shared
mechanism was somewhere other than where I was looking**, and the first time I caught it before
shipping rather than after.

### What is NOT yet proven

The widening is proven **present**, not proven **effective**: no sweep has run since the roll, so
nothing has yet been closed from `unresolved`. The honest state is "live and structurally correct,
behaviourally unexercised". Expect the next sweep to close **at most 1 row** (the single
`needs_page` `unresolved`). **If it closes more than 1, something is wrong** — that is the check,
and it is disconfirmable, which is the point.

## 2026-08-05 evening — option A survived a second roll, and the mechanism is UNDRIVEN

**Still live on `v1.0.1254`** (two rolls after the one that shipped it): widening greps **3**, old
literal **0**, reverted string **0**, both replicas. Re-checked because *a roll is not evidence
your fix shipped* — an image can predate the commit — and this one did not.

### ⚠ The disconfirmable check returned 0, and 0 IS NOT A PASS HERE

I wrote in the handoff: *"expect AT MOST 1 newly-closed row; more than one means something is
wrong."* The result is **0 rows** — which satisfies the letter of that check and means **nothing**,
because the mechanism never ran. `auto:revalidated` is still **33 all-history, latest 2026-08-04
08:37:47**, unchanged across two rolls and ~36 hours.

**I wrote a one-sided check.** "At most 1" can only catch the mechanism being too WIDE; it cannot
distinguish "correctly closed the single eligible row" from "never executed". That is the standing
*a gate's 0 findings has two causes with opposite fixes* trap, and I walked into it while writing
the check that was supposed to prevent guessing. **The check should have been "exactly 1, AND the
sweep's own run count increased".** Corrected in the handoff.

### Why it never ran: there is no schedule for it

```
SELECT name, target_agent_type, enabled, interval_seconds, last_triggered_at
FROM scheduled_tasks WHERE target_agent_type ILIKE '%revalidat%' OR name ILIKE '%revalidat%';
-- (0 rows)
```

**No `scheduled_tasks` row exists for `diagnosis-review-queue-revalidator`, by name or by agent
type** [MEASURED 2026-08-05]. The scheduler is emphatically alive — **27 enabled tasks, latest
trigger 2026-08-05 20:47**, minutes before I looked — so this is not a dead scheduler. This agent
is simply not in it, and its 33 lifetime closes were hand-dispatched.

So the standing lesson lands exactly: **a silent mechanism is usually UNDRIVEN, not missing.**
Option A is structurally correct, council-approved, pod-verified — and inert until somebody runs
the sweep. The bug this lane set out to fix (a queue the sweep builds and cannot drain) is now
*fixable* but not yet *being fixed*.

### What I deliberately did NOT do

**Did not add the `scheduled_tasks` row.** It would make a `dry_run: false` auto-closer run
fleet-wide on a timer — new standing authority over work-item lifecycle across four item types,
one of which (`needs_page`) has **5 producers** and an open owner decision (033 D2) sitting in its
`failed` population. Config is live immediately with no build, so there is no ordering constraint
forcing my hand. **17 of the 44 scheduled tasks are disabled**, which suggests scheduling here is
managed deliberately rather than by whoever passes through. That is an owner call.

**Did not hand-dispatch a sweep either**, and this is the closer judgement. Its *closing* blast
radius is 1 row, measured — but gate 2 stamps `result.revalidation` on every row it scans and does
not close, so an unfiltered run writes to ~50 rows across the fleet. That is exactly what the sweep
is for and it is reversible, but it is a bigger action than "prove one row closes", and it wants a
site filter and a fresh coverage check rather than the tail end of a long session. The exact
command, with both, is in the handoff.

### The needs_page `failed` population is volatile, which is itself informative

17 (08-04) → 21 (08-05 morning) → **19** (08-05 evening). It moves in both directions, so it is
actively churning, not a static backlog. **That strengthens the decision to exclude `failed`**: a
timer-driven auto-closer pointed at a population that gains and loses rows daily, under an open
owner decision, is precisely the combination that produces an unexplainable audit trail.

## 2026-08-06 — DISPATCHED BY HAND: the widening is PROVEN EFFECTIVE, exactly as predicted

Owner asked for one hand-dispatch. Run `19d0ccbd-c2df-4beb-bd1a-f7e41599eb5f`, correlation
`12c88a1e-01f6-489e-ac8e-279a4a8bcfc1`.

**Scoped deliberately**, not fleet-wide: `{"site_id":"199733a8-…","item_type":"needs_page",
"dry_run":false,"max_items":50}`. The shared trigger
(`review_queue_drain/TRIGGER_revalidate_review_queue_v1.sh`) hardcodes `INPUT_DATA='{}'`, so I
replayed its envelope with a filter rather than editing another lane's script. Pre-checks first:
0 claimed items fleet-wide, pods 39 minutes old (past the ~300s silent-drop window), binary
re-verified on `v1.0.1256` (widening greps 3, old literal 0).

⚠ **First publish failed loudly and sent nothing** — the kcat pod name contained an uppercase
letter (`kcat-optA-…`), which k8s rejects as a non-RFC-1123 name. Worth knowing because the
adjacent landmine is that **`kcat -P` exits 0 having sent nothing**; here the *pod* was rejected so
it was visible, but the habit that saved it was verifying by the orchestration row rather than by
exit code.

### Both arms of the check, and this time the check could fail

Dispatch was immediate — **1 poll**, not the ~30 minutes the trigger warns about. One sample.

| arm | before | after | |
|---|---|---|---|
| **1** — newly closed rows since 08-05 | 0 | **exactly 1** | ✅ as predicted |
| **2** — `auto:revalidated` lifetime | 33, latest 2026-08-04 08:37:47 | **34**, latest 2026-08-06 08:05:19 | ✅ it really ran |

Arm 2 is the one that mattered: yesterday's 0 was vacuous because the sweep had never run, and a
one-armed check could not tell that from success. Now both moved.

**The closed row is the exact one the widening made reachable**: `needs_page:self-correction-
leopardessconsulting` on `fundamentallyai.com`, raised 2026-07-20, sitting at **`unresolved`** —
a status the sweep **could not see before this change**. Closed `auto:revalidated` with a positive
reason: *"page … is active and every section it declares (hero, generic-text-block, swipeable…)"*.
`unresolved` population across all four covered types is now **0**.

**This is a direct behavioural proof, not an inference.** Before the change that row was
unreachable by construction; after it, one sweep closed it on positive evidence.

### The discrimination control — it judged three and closed one

The site had 3 eligible rows. It closed 1 and **refused another with a stated reason**:
`needs_page:platform-log-index`, verdict **`unknown`**, judged 08:05:15, *"page 'platform-log-index'
resolves no sections from any source, so there is no evidence"*. Left open, correctly. A mechanism
that closed both would have looked the same in arm 1's count.

### ⚠ ANOMALY — one eligible row was neither closed nor stamped, and I cannot explain it

The third row was **not judged at all**: no `result.revalidation` key, so it did not even get
gate 2's record.

- `item_type = 'needs_page'`, `status = 'needs_human_review'`, created **2026-08-05 13:44** — i.e.
  well before the run, so "created after the sweep" does not explain it.
- **Its `item_key` is `page_rerender:tools`** — the prefix disagrees with its `item_type`. That is
  exactly the drift `workItemKey`'s doc comment describes ("hand-rolled prefixes are how the keys
  drifted from their item_type in the first place").

Whether the mismatch *causes* the skip is **[UNVERIFIED]** — the selection is on `item_type`, which
matches, so on a first reading it should have been loaded and judged. **Not guessing.** It is
recorded as an open observation in the handoff with the query that found it. It is not a regression
from this change: nothing in option A touches loading or key handling, and a skipped row is the
safe direction (nothing closed).

### What this run did NOT prove

**The `failed` exclusion was not exercised.** That site has **no** `needs_page` rows at `failed`
(8 complete, 2 needs_human_review, and the 1 just closed), so the exclusion was never put to the
test — it was *not contradicted*, which is a weaker thing. The 19 `failed` rows fleet-wide sit on
other sites and this run was site-scoped. Saying so because "the exclusion held" would be a claim
this run cannot support.

## 2026-08-06 — SCHEDULED (owner approved), verified end-to-end, and it exposed two things I had wrong

`scheduled_tasks` row **`review-queue-revalidate-daily`**, `interval_seconds=86400`,
`concurrency_group='review-queue-revalidate'`, `max_concurrent=1`, enabled. Config, so live
immediately — no build.

### Verified BY EFFECT, not by the row's own timestamp

The standing trap is that `last_triggered_at` is written by the scheduler at publish time and is
**not** evidence the agent ran (cf. the thunder-reaper landmine: *`enabled` + a fresh tick ≠ ever
RUN*). So the proof chain was taken end to end:

1. `last_triggered_at` = **2026-08-06 08:37:50.088** — the scheduler published.
2. Orchestration **`7e107c8a-65bd-4253-8b04-102989473f64`**, created **08:37:50.254** (0.17s later),
   `current_step=complete`, `status=COMPLETED` — the agent actually ran.
3. Its own counters: **`scanned 500, resolved 0, still_holds 31, unknown 469, dry_run false`**.

**`resolved: 0` is correct, not a failure** — the hand-run at 08:05 had already closed the only
eligible row, and `auto:revalidated` stayed at 34. A schedule that fires, runs in live mode, judges
500 items and closes nothing *because nothing was closable* is the right outcome. Distinguishing
that from "it did not run" is exactly why arm 2 of the check exists.

### ⚠ WRONG #1 — my `max_items` was INERT, and I had already written it into a landmine as the fix

I set `input_data = {"dry_run": false, "max_items": 1000}` specifically to defeat the starvation
described below. The run reported **`capped_at: 500`**.

`revalidate_review_queue_action.go:178` reads
`datahelpers.GetIntField(config, "max_items", 50)` — from the **step config**, not `input_data` —
and the `sweep` step has **no `input_mapping`**. So nothing in `scheduled_tasks.input_data` reaches
it. The 500 is the agent definition's step config.

**This is the two-gates shape for the third time in this lane** (`input_mapping` vs a claim query's
`RETURNING`; the sweep's selection vs its two CAS guards; now `input_data` vs step config). **The
gate you can reach is not the one that decides.**

Two corrections made rather than left:
- The `scheduled_tasks` row's `input_data` is back to `{}` and its `description` now says outright
  that the row cannot tune the sweep and where the real knob is. **A dead config key that looks
  live is its own defect** — leaving `max_items: 1000` sitting there would have taught the next
  reader something false.
- The `LANDMINES.md` entry I had written **minutes earlier** claimed "the scheduled row uses 1000
  for exactly this reason". Corrected in place, same day, with the measurement.

⚠ **I wrote a landmine asserting a fix I had not verified, and the verification was one query
away.** The entry was accurate about the *hazard* and wrong about the *remedy* — which is the more
dangerous half, because a remedy is what a reader copies.

### ⚠ WRONG #2 — the starvation is real, present, and NOT fixed

The first scheduled run measures it precisely: **500 scanned, 469 `unknown`** — 94% of the batch
was types with no revalidator, which return `unknown` for ever and stay parked, stay oldest, and
get re-selected next run. **279 of the 779 parked rows were never reached at all.**

So the daily sweep will re-judge substantially the same 500-row head every day. It still works —
it found and judged 31 covered items — but it is doing 94% wasted work and cannot see the tail.
**Fixing it needs an `item_type` filter or a bigger step-config batch, and that is an
`agent_definitions` change, not a schedule change.** Recorded as an open item in the handoff rather
than fixed at the end of a long session.

### State

Schedule live and proven. `auto:revalidated` = **34**. Next natural fire ~2026-08-07 08:37Z.

---

## 2026-08-06 (next session) — the open item, measured harder and then fixed twice

Picked up the lane cold from `HANDOFF_2026-08-04_continue_here.md` §0. Everything below is that
one item. Two changes shipped: a stopgap that is **live now**, and the durable fix, which is
**committed and inert until a roll**.

### The measurement the handoff did not take

§0 recorded the waste (`scanned 500 · unknown 469`, 94%) and that "279 rows are never reached at
all". It did **not** say whether any of the unreached rows were rows the sweep could actually
judge. That is the question that decides whether this is untidy or harmful, so I ranked the parked
set by `created_at` and split it on both axes at once:

```sql
WITH parked AS (
  SELECT id, item_type, created_at, row_number() OVER (ORDER BY created_at ASC) AS rn
  FROM site_work_items WHERE status IN ('needs_human_review','unresolved')
)
SELECT (rn <= 500) AS in_oldest_500,
       item_type IN ('required_fields_missing','needs_section_data','unresolved_cta','needs_page') AS covered,
       count(*)
FROM parked GROUP BY 1,2 ORDER BY 1 DESC, 2 DESC;
```

```
 in_oldest_500 | covered | count
---------------+---------+-------
 t             | t       |   104
 t             | f       |   396
 f             | t       |    64     <-- the finding
 f             | f       |   215
```

**[MEASURED 2026-08-06]** **64 judgeable rows sat beyond the cap** — `required_fields_missing` 48,
`needs_page` 8, `needs_section_data` 7, `unresolved_cta` 1; oldest filed **2026-07-24**. 38% of the
168 rows the sweep exists to judge. And it does not self-correct: 396 of the 500 head slots hold
rows that return `unknown`, which is non-terminal, so only ~104 slots ever turn over. **Not
reached slowly. Never reached.**

Disconfirmable, which the last two entries in this file are a reminder to check: had the covered
rows all been old, they would have been inside the head and the count would have been 0. It came
out otherwise.

⚠ **The starved rows are the NEWEST covered ones** — `rn > 500` under `ORDER BY created_at ASC` is
by definition the young tail. That **inverts the selection's own stated rationale** ("the oldest
items are the ones most likely to be describing a page state that no longer exists"): a finding
filed last week is the one a recent re-render is likeliest to have already fixed, and it was the
one guaranteed never to be looked at.

### ⚠ WRONG #3 — §0's recommended fix cannot work, and it falls into §0's own trap

§0 recommends "(1), as several scheduled rows — one per covered type." **That route does not
work.** `item_type` is read the same way `max_items` is:

```go
maxItems   := datahelpers.GetIntField(config, "max_items", 50)   // config = params.StepConfig.Config
typeFilter, _ := config["item_type"].(string)                     // ...the STEP config, not input_data
```

The sweep step has **no `input_mapping`** (verified against the live `agent_definitions` row — its
step keys are `action`, `config`, `next_step`, `description`, `output_field`). So *n* scheduled
rows would all read this one step config and behave identically — you would get four identical
sweeps a day, not four filtered ones. Making it work needs either four near-duplicate agent
definitions or an `input_mapping` added first.

**This is the exact trap §0 documents two paragraphs earlier** ("YOU CANNOT FIX THIS FROM
`scheduled_tasks`. I tried and measured it."). The lane's own NOTES had it right — "that is an
`agent_definitions` change, not a schedule change" — and the handoff's recommendation contradicted
its own notes. **A trap you have just documented is not thereby disarmed for your next paragraph.**

### What shipped

**1. Stopgap, LIVE.** `321_review_queue_sweep_reaches_every_parked_row.sql`, applied 09:04 UTC,
`max_items` 500 → 1500, commit `b14609e05`. Config, so live immediately, no roll. Verified at the
live row (1500, `dry_run` still false) and in `schema_migrations`. The guard block asserts the cap
still exceeds the live parked count **and** that no covered row remains beyond it, so a later apply
into an outgrown queue fails loudly rather than under-reaching in silence. The NOTICE read:
`779 parked, 168 judgeable, 0 now beyond the cap (was 64)`.

⚠ Parked went **772 → 779 during the session** (~20 min). Sized 1500 not 800 for that reason:
parked rows accumulated **205 in three days** against 55–81 in each prior week, so a cap sized to
today's queue is back in starvation within a week.

⚠ **`--apply` takes every pending file, and 24 were pending from other threads.** Scoped with
`MIGRATIONS_DIR=<scratch dir holding only my file>`. The dry run then showed `Pending (1)`, which
is the check that the scoping actually worked — not the assumption that it did.

**2. Durable fix, committed `0e4e79124`, inert until a roll.** The selection filters to the types
`reviewRevalidators` covers, **derived from the map** so registering a revalidator widens the
selection in the same edit and the two cannot drift. The coverage gap is now one `GROUP BY` over
the whole parked set instead of an accumulation of whatever fell inside the cap — the old shape
reported the gap as **smallest exactly when the backlog was worst**. Council `f64da546`, submitted
before the commit, trailer `Council-Submitted:`.

Two things closed in the same change because the fix would otherwise have created them:
- The filter would make an operator's uncovered `item_type` match **nothing**, and `scanned 0`
  reads identically to "the queue is drained". `validateTypeFilter` refuses and names the covered
  set. (Before the filter that request was merely useless; after it, it would have been *silent*.)
- `cap_binding` is now logged at WARN when the pass fills its cap. `capped_at` was **always** in
  the payload — §0 quotes it — and it still took a fortnight to notice, because nobody compares
  two numbers in a blob nobody reads. **A number in a payload is not a signal.**

### ⚠ Misstep — my own test asserted a behaviour Go does not have

`TestCoveredTypeFilterIsAccepted` called `loadParkedReviewItems(ctx, nil, ...)` expecting a nil
`*sql.DB` to return an error. It **panics**: `database/sql.(*DB).conn` dereferences immediately.
Caught by running it, one minute after writing it.

The fix was better than the test: I extracted the refusal into `validateTypeFilter`, a pure
function, so both directions are testable with no DB at all — and then made the accept-side test
loop over **every** covered type, because a blanket refusal that rejected everything would have
passed the refusal test on its own. **A refusal test needs its accept case in the same commit or
it only proves the code can say no.**

### State

Stopgap live; durable fix committed, awaiting council `f64da546` and a roll. `auto:revalidated`
= 34 at session start. **The next scheduled run (~2026-08-07 08:37Z) is the first that can reach
all 168** — it should report `scanned` ≈ the full judgeable count, not 500.

### Council `f64da546` — APPROVED round 1, 11 seats, 4 advisory objections, none high

`gated_by_truncation: false`, `unreadable: 0`, 6 abstained (relevance-gated). **Checked the
objections rather than filing them** — four were checkable and two found real facts.

**1. `bug_historian` "missing": is the schedule actually ENABLED?** The best objection of the
round, because a no would have made the whole change moot. **[MEASURED]** `enabled=t`,
`interval_seconds=86400`, `last_triggered_at=2026-08-06 08:37:50Z`. It is live. (The seat was
citing the landmine about a disabled task showing a fresh timestamp — worth the query every time.)

**2. `guardian` (low): "does not name every caller — blast radius on manual invocation is asserted,
not verified."** Fair, and now verified rather than argued:

```sql
SELECT type, jsonb_path_query_array(default_config, '$.**.config.item_type')
FROM agent_definitions WHERE default_config::text LIKE '%revalidate_review_queue%'
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
--  diagnosis-review-queue-revalidator | []
```

**Zero configured callers pass an `item_type` anywhere.** The only other invocation path is
`TRIGGER_revalidate_review_queue_v1.sh`, which sends `input_data {}` — and `input_data` is inert on
this step regardless. So the new refusal is **purely defensive today**: it cannot fire for any
existing caller. That is a stronger result than the submission claimed.

**3. `prior_art_librarian` (medium): the no-consumer absence claim was raw grep, which has known
blind spots here.** Re-ran by a different route — searching for the payload keys as identifiers
across all Go, then separately across `.sql/.sh/.py/.json/.yaml`. **No Go consumer of
`uncovered_types`, `uncovered_backlog` or `cap_binding` exists anywhere outside the action.** The
only `revalidation` hits outside it are **SQL comments** in
`review_queue_drain/seed_review_queue_revalidator.sql` — human verification queries, not code.
⚠ Those queries still work for covered types; for **uncovered** types they now return nothing,
which is exactly `bug_historian`'s second objection (below). The review_queue_drain lane is being
told.

**4. `constitution` (medium): the `IN` list is interpolated via `sqlInList`, not bound.** The rule
is written unconditionally, so the objection is correct on its face. **Not actioned, deliberately.**
`lib/pq` IS in `go.mod`, but this package **deliberately avoids `pq.Array`** — two files say so in
terms (`fixloop_digest_action.go:488`, `resolve_composition_helpers.go:398`). Binding only my half
would leave one query with its status list interpolated and its type list bound, which is worse to
read and would break the source-scan test's count. The values are Go map keys — compile-time
constants — and the operator-supplied filter continues through a bound `$n`. The seat itself noted
this "is an extension of a pre-existing practice rather than a newly introduced defect".

**5. `editquality` (medium): "`validateTypeFilter` is defined but nothing shows it being called."**
A gap in my **sketch**, not the code — `loadParkedReviewItems` calls it before building the type
predicate. Worth recording as a submission lesson: **an extracted helper needs its call site in the
sketch, or a reviewer must assume it is dead code**, and they are right to.

**6. `bug_historian` (medium), the one I cannot fully close, recorded honestly.** The new
`cap_binding` WARN and `uncovered_backlog` have **no confirmed consumer** — pod logs rotate within
minutes and nothing reads the payload key. *"A signal nobody reads is silent in practice, just with
extra steps."* Partial mitigation, stated as partial: after this ships `cap_binding` should be
structurally **false** (168 judgeable against a 1500 cap), so it is a **tripwire**, not a routine
signal — and the payload query is now written into the handoff (§0a-verify) rather than left for
someone to invent. **That still depends on a human running it.** Not solved; named.

**7. `bug_historian` (low), a real consequence I had not listed as a risk.** Uncovered rows now get
**no stamp at all**, so "has the sweep ever looked at this row?" becomes ambiguous — absence no
longer distinguishes *never scanned* from *no revalidator exists*. Previously the `unknown` stamp
answered it per row. The aggregate answers it per type, which is the more useful altitude, but it
is a genuine loss of a per-row fact and it belongs in the record.

### 2026-08-06, post-roll — the durable fix IS LIVE on `v1.0.1257`

**[MEASURED] Both replicas, `agent-chassis-5b9fd84984-{hqc5d,qvzkg}`, started 09:52Z:**

| needle | baseline `v1.0.1256` (pre-roll) | now `v1.0.1257` |
|---|---|---|
| `judgeable rows were left unexamined` (cap warning) | **0** | **1** |
| `so this sweep cannot judge it` (the refusal) | **0** | **1** |
| `auto:revalidated` (positive control) | 2 | **2** |

**This is the whole proof, and it only exists because the 0 was taken before the roll.** The change
is purely string-additive, so nothing it removed can be greped for 0 — there is no valid negative
control, and §2 of the handoff records this lane nearly pretending otherwise. The positive control
is the load-bearing companion: it is non-zero on *both* binaries, so a 0/0 reading would have meant
"the change did not ship", not "the probe broke". ASCII-only needles throughout (§2's em-dash
mangling across `exec -- sh -c`).

⚠ **The roll was not mine and did not mention my commit.** `v1.0.1257` was built by another session;
my commit `0e4e79124` simply happened to be at HEAD when it built. That is the standing fleet
lesson — *a roll is not evidence your fix shipped* — and it is exactly why the grep above is the
evidence and the version bump is not.

### PROVEN BY EFFECT — one sweep on `v1.0.1257` closed 20 rows the old code could never reach

Hand-dispatched `TRIGGER_revalidate_review_queue_v1.sh` (orchestration
`267fe850-5c22-43fe-b1d1-266e261a3a40`, 10:02Z, >300s after the pod restart per the trigger's own
rebalance gotcha). Payload:

```
scanned 168 · capped_at 1500 · cap_binding false · resolved 20 · still_holds 36 · unknown 112
closed 20 · uncovered_backlog 611
```

**Every number is a check that could have come out otherwise:**

- **`scanned` = 168, the exact judgeable count** — not 500 (the old head), not 779 (the parked
  total). The selection filter is doing precisely what it claims, and nothing more.
- **`cap_binding` = false** — no judgeable work left behind. 20 + 36 + 112 = 168, so the arithmetic
  closes with no rows unaccounted for.
- **`uncovered_backlog` = 611** — the full coverage gap. The old code **structurally could not
  report this number**; it could only report the unjudgeable rows that happened to fall inside the
  cap, which is why the gap looked smallest when the backlog was worst.
- **`unknown` = 112 now means something different and honest**: a revalidator looked and could not
  tell. Not "no revalidator exists".

**The decisive one — `resolved` 20, against 0 on the last pre-fix scheduled run:**

```sql
SELECT count(*), count(*) FILTER (WHERE created_at >= '2026-07-24') AS in_the_starved_tail
FROM site_work_items WHERE resolution_path='auto:revalidated' AND completed_at > '2026-08-06 10:00Z';
--  20 | 20     (created 2026-08-03 .. 2026-08-05; 17 required_fields_missing, 3 needs_page)
```

**All 20 were in the starved tail** — created 2026-08-03 to 2026-08-05, far beyond the old rank-500
cutoff, i.e. rows the previous code could **never** have judged. Fleet `auto:revalidated`: **34 →
54**. This is not "the mechanism still works"; it is 20 findings closed that were permanently
invisible to it yesterday.

⚠ And it confirms the *newest-rows* inversion was the real harm: the sweep's oldest-first rule was
starving exactly the rows most likely to be closable, which is why one pass over the previously
unreachable tail had a **12% close rate** where the reachable head had been returning 0.

### §0b's loose thread — RESOLVED, and its suspected cause REFUTED

§0b flagged one `needs_page` row on `fundamentallyai.com` (`2d669d7b`, created 2026-08-05) that got
**no `result.revalidation` key at all** while its siblings were judged, and suspected its
**mismatched `item_key` prefix** (`page_rerender:tools` under `item_type='needs_page'`) — marked
`[UNVERIFIED]`.

**It was not the prefix.** That same row, prefix unchanged, was judged `resolved` at
`2026-08-06T10:03:15Z` by the run above. It was simply **unreachable**: created 2026-08-05, it sat
in the starved tail. Its siblings that *were* judged (`88c7f89e`, `d53566f9`) are both from
**2026-07-21** — old enough to be inside the 500-row head. **§0b was a symptom of the same bug**,
not a second defect, and the `workItemKey` drift it suspected is a red herring here.

The general shape is worth keeping: **a row that is missing a stamp its siblings have looks like a
per-row skip, and "which rows got skipped" and "which rows were never loaded" are indistinguishable
from the row itself.** Check reachability before theorising about a predicate — the sibling
comparison that made the prefix look guilty was really an age comparison.

### ⚠ WRONG #4 (mine) — I wrote "this lane has nothing open" without checking §5.3/§5.4, then mis-measured the thing I had not checked

Two errors, an hour apart, both in the closing paperwork rather than the work.

**(a) The status claim.** Having closed §0 and §0b I wrote "**This lane now has nothing open**" into
the handoff's top box. §5.4 lists Decision 2's dedup half as open-and-blocked and §5.3 lists an
armed-but-inert cap; I had read neither that session. **Closing the item a handoff was reopened for
is not the same as closing the lane**, and the top box is precisely where a later reader takes the
status on trust. Corrected in place.

**(b) The re-measurement, which was worse, because it produced a confident number.** Checking
§5.4's blocker I wrote the exclusion list **from memory of the phrase "terminal statuses"**:
`status NOT IN ('complete','verified','cancelled','rejected')`. That returned **75 pairs / 227
rows** and I was one keystroke from recording "the blocker has grown ~56%". The real predicate,
read from the live index —

```sql
SELECT pg_get_indexdef(oid) FROM pg_class WHERE relname='idx_swi_dedup';
-- ... WHERE item_key IS NOT NULL AND status <> ALL (ARRAY[
--     'complete','verified','rejected','wont_fix','failed','cancelled','unresolved'])
```

— excludes **three more statuses** (`wont_fix`, `failed`, `unresolved`). My set silently included
them, inflating the count. Measured against the *proposed* predicate (today's index minus
`unresolved`, which is what Decision 2 actually changes): **53 pairs / 180 rows**, against 48 / 135
recorded on 2026-08-03. It has grown — modestly, and by nothing like what I nearly wrote.

**The cheap check, and it is the same one this file keeps re-learning:** `pg_get_indexdef` is one
query and it is the *only* authority on what an index excludes. **A status list reconstructed from
a name is a guess wearing a filter's clothing** — and here the guess and the truth differ by three
statuses in the direction that flatters the finding. Related: the standing lesson that re-running
someone's SQL inherits its blindness; this is the sibling where *not* re-running it, and inventing
the predicate instead, inherits your own.

**What caught it:** deciding the comparison had to be apples-to-apples before asserting growth —
i.e. asking "is my filter the same as theirs?" *before* the number went into a document, not after.

---

## 2026-08-08 — the unattended proof, two days and five builds later

The 2026-08-06 entry ended with a prediction: *"the next scheduled run ~2026-08-07 08:37Z should
show the same shape unattended."* Tested, and it holds.

**Runs happened with nobody watching, twice** — measured from the closures, not inferred:

```sql
SELECT completed_at::date, count(*) FROM site_work_items
WHERE resolution_path='auto:revalidated' GROUP BY 1 ORDER BY 1 DESC;
--  2026-08-08 | 3      <- unattended
--  2026-08-07 | 1      <- unattended
--  2026-08-06 | 21     (20 of them my one hand-dispatched drain)
--  2026-08-04 | 33     (pre-fix history)
```

Fleet `auto:revalidated`: **34 → 58**. Today's scheduled run (`1ac359c4`, 08:38:29Z, unprompted):

```
scanned 151 · capped_at 1500 · cap_binding false · resolved 3 · still_holds 37 · unknown 111
uncovered_backlog 625
```

**Three independent cross-checks, each of which could have failed:**

- **`scanned` 151 is the judgeable population, not the queue.** Live now: 148 judgeable, 625
  uncovered, **773 parked**. `151 − 3 resolved = 148`. The sweep looked at 151 of 773 — the exact
  right 151.
- **`uncovered_backlog` 625 matches the live uncovered count to the row.** Not a sampled or
  cap-truncated figure, which is what the old code could only ever produce.
- **`cap_binding` false**, with 151 against a cap of 1500.

⚠ **Do not read `resolved 3` as degradation from 20.** The 20 was a one-off drain of a backlog that
had been unreachable for a fortnight; 1 and 3 are steady state — the sweep is keeping up rather than
catching up. A trickle after a flush is the *correct* shape, and the number to watch is
`cap_binding`, not `resolved`.

**No regression across five builds.** Verified on `v1.0.1262` (was `v1.0.1257` when first proven):
both replicas still carry the cap warning and the refusal string, positive control non-zero. Worth
re-checking precisely because none of these builds were mine — the change has simply ridden along
at HEAD, which is the normal condition on this tree, not a special case.

⚠ **The 2026-08-06 evidence is GONE from the database.** `orchestration_states` holds **1** reval
run now; the 08:05, 08:37 and 10:03 rows from 08-06, including the one carrying the `scanned 168 /
resolved 20` payload, have aged out at ~24h retention. **The only surviving record of the decisive
run is what was written into these docs at the time.** That is the whole argument for recording a
payload the day you take it, and it is the standing `~24h retention` lesson arriving on schedule.

---

## 2026-08-08 (second session) — §4.3 opened: `voice_tells` is the sweep's fifth covered type

**Lane re-verified first, across two more rolls I did not make.** `v1.0.1264` then `v1.0.1266`,
both replicas: `capwarn=1 refusal=1 positive_control=2` each time. Latest scheduled run
(`1ac359c4`, 08:38:28Z) `scanned 151 · capped_at 1500 · cap_binding false · resolved 3 ·
uncovered_backlog 625`. Closures by day 3 / 1 / 21 / 33 — steady state, unattended.

**Dedup half re-measured [MEASURED 2026-08-08]: 55 colliding pairs across 184 rows**, against the
53 / 180 recorded in the morning's handoff and 48 / 135 on 08-03. Contributors unchanged:
`undeployed_asset` 48, `improve_tool` 30, `needs_internal_links` 29. **The index definition was READ,
not reconstructed** — `pg_get_indexdef` — which is the trap the handoff records costing a 56% growth
figure that was an artefact of a remembered exclusion list. It still drifts upward; §4.1 stays
blocked and gets more expensive.

### Candidate triage — and two of the three named candidates were wrong

The handoff named `content_rewrite` (~34), `voice_tells` (~25), `needs_sprite_css` (~10). **The
CLOSER check separated them immediately, and it is the check this lane exists to have learned:**

| candidate | closer census | verdict |
|---|---|---|
| `content_rewrite` | 51 `complete`, carrying `deploy_result` payloads — *"prose rewritten in voice H … verified at the served page"* | **REJECTED.** A real fix pipeline already drains it. Weakest retraction candidate of the three, not the largest one |
| `needs_sprite_css` | zero closed — but all 10 rows are `unresolved`, and its own source comment says *asset-deployer's sprite_css mode re-runs* | **DEFERRED.** A re-run path may exist; needs its own producer/closer pass |
| `voice_tells` | **zero rows have ever reached `complete`/`verified`** | **ADOPTED** |

**`voice_tells` had no closer of any kind.** `HandlerAgent: ""` by design, no dispatch path, no
revalidator: 25 items, all `needs_human_review`, all filed **2026-07-17**, all on
leopardessconsulting.co.uk, still parked 22 days later. Parked for ever by construction.

### The objection I expected to kill it, and why it did not

`bugs_open/033` quotes `check_voice_tells.go:142` — ***"never an unreviewed auto-rewrite"*** — as
evidence the type was filed correctly for human review. That reads like a prohibition on exactly
what I was about to build. **It is not: it governs the FIX path, not the retraction path.**
Retraction never edits copy and never dispatches a rewrite; it withdraws a claim the current page no
longer supports. The human decision about *how* to fix machine-written prose is untouched.

Checked both consuming bug files rather than assuming: **083 classifies `voice_tells` under
*advisory / machine-fixable in principle* (186 items), NOT under *needs a human ANSWER* (50).**
033 lists it among ~175 *deliberate, documented escalations*. **Neither places it under an open
owner decision** — which is the check that once stopped this lane pointing an auto-closer at 21 rows
mid-decision.

### What I built, and the one design choice that matters

`actions` imports `discovery_checks`, one-way, so shared code had to go in `discovery_checks`.
Extracted the emit side's query + scan into exported `ScanVoiceTells`; **both ends now call it**, so
the two ends of an item's life cannot answer *"does this read machine-written?"* differently. Same
precedent as `revalidateNeedsPage` sitting beside its own resolver. Registered as CQ-020.

**The subtle part is not the predicate — it is that `len(Findings) == 0` collapses three states:**

1. components were examined and are clean → **the prose was fixed** (the only state that may close)
2. **nothing was read at all** — page deleted, not `active`/`deployed`, no rendered components
3. **only human-LOCKED components were read** — the emit side has always skipped those

States 2 and 3 produce an empty findings list identical to state 1. So `VoicePageScan` reports
`ComponentsExamined` and `ComponentsSkippedLocked`, and the ladder refuses on both. This is the
no-op-case rule arriving in a new costume: **I checked what could break before I checked what would
be a silent no-op, and the no-op is the one that closes a live human-review row.**

### Measurements taken before writing the arms, not after

[MEASURED 2026-08-08, live clients_db] All 25 items: `page_missing 0 · page_not_live 0 ·
no_unlocked_components 0 · has_locked_components 0`. **So the two locked-component arms are
UNEXERCISED on today's data — reasoned and unit-tested, NOT observed in production.** Said plainly
in the code comment, the submission's risks block and CQ-020, because "handled" and "seen to work"
are different claims and this lane has paid for conflating them.

[MEASURED] **13 of the 25 pages have a `page_components.updated_at` later than their item's
`created_at`** — so this judges real change rather than running inert. That was the disconfirming
check: had it come back 0, registering the type would have added scan cost and closed nothing, and
I would have said so rather than shipping it.

[MEASURED] The site is still opted in — `aspect='voice'`, `is_current=true`,
`voice_gate.enabled=true` — so the gate loads and the refusal arm is not the common path.

### Missteps this session

- **I read the wrong handoff first.** Opened `HANDOFF_2026-08-04_continue_here.md` because it was
  the path I was handed; its own top banner says it is superseded. Cost nothing because the banner
  is there, which is the argument for putting one on every superseded file.
- **I called a build failure another session's WIP, on one truncated error.** `go build` failed with
  `undefined: verificationOutcome`; I saw `verifiers.go` dirty in the tree and said so. The symbol
  was defined 24 lines below its call site, and the next build passed with no intervention — the
  tree had simply been read mid-write. **The real lesson is not "it was theirs" but that I piped
  the error through `head -10` and then reasoned from the survivor**, and that my `&& echo "BUILD
  OK"` printed after a FAILED build because `head` exited 0. A shell idiom that reports success on
  a failed command is a landmine I built myself, in the same session I was writing landmine notes.
  Fixed by checking `${PIPESTATUS[0]}` and by building against `git archive HEAD` plus my files
  only, which is what the lane's own trap list says to do when the shared tree is broken.
- **Mutation testing caught a guard passing in series.** Deleting the *examined nothing* guard did
  not fail the locked-only ladder case — it fell through to the *some components locked* arm and
  produced a plausible-looking `unknown`. Only the separate property test
  (`TestVoiceTellsNeverResolvesWithoutReadingSomething`) caught the pure 0/0 case. **A per-case
  table cannot prove an invariant; state the invariant separately.** All three guards mutated to
  `if false` one at a time, each caught, green on restore.

### State

Committed `ef80216be`, council `4d430ca8-7e34-479a-95f3-71fdc12fdef6` submitted alongside with
`Council-Submitted:` (verdict pending at time of writing — **read it, and act on a REVISE**).
**Go change: inert until the next chassis roll.** The verification recipe is the same 0→1 transition
shape as the selection filter, and **the baseline must be taken BEFORE the roll** — see the handoff.

### Council verdict + the correction it forced (same day, ~1h after the section above)

**APPROVED r1, 13 seats, 5 advisory objections, none high, `gated_by_truncation: false`.** Full
answers with the checks run: `OBJECTIONS_2026-08-08_voice_tells_council.md`. Four seats
independently asked the same thing (show the second-producer search, do not assert it) and one
(`debug_historian`) landed a real hit.

> **CORRECTED 2026-08-08: the population is 32, not the 25 recorded above.** Seven more
> `voice_tells` rows were filed **today** by `quality-discovery-agent` while I was building the
> revalidator. I found them only because objection 1 sent me back to the provenance table — nothing
> in my own process would have caught it. **The check is actively filing, so this type GROWS; it is
> not a fixed backlog of 25.** Churn re-measured over all 32 is still 13 (the 7 new rows were filed
> today, so none can have changed since filing) — the ratio moved, the absolute count did not.
> My figure survived about four hours. A population count is a measurement with a timestamp.

> **`debug_historian` was RIGHT about the mechanism and wrong about the consequence, and the two
> must be scored separately.** It flagged `p.status IN ('active','deployed')` as the `pages.status`
> vs `build_status` landmine. Enumerated: `active` 585, `archived` 29 — **`'deployed'` NEVER OCCURS
> in `pages.status`, so half that disjunct is dead.** The seat was right. But all 32 items sit on
> `status='active'`, so the revalidator is **not inert** and the feared WRONG_CALLS-shaped failure
> does not happen. The dead literal is inherited verbatim from the emit side and **stays**:
> narrowing the shared predicate inside a retraction commit would change what the CHECK matches, for
> no behavioural gain. Recorded, not silently fixed.

⚠ **The provenance census produced a FALSE POSITIVE in the direction of alarm.** Two `created_by`
values (`generic` 25, `quality-discovery-agent` 7) look exactly like two producers. They are two
AGENTS running one check — `created_by` is `dctx.AgentType`, not the filing code. This is the
landmine's own point arriving as a live example: **`created_by` cannot answer "is there a second
producer"**, and the answer is the Go call-site census (one) plus the config census (one row, and it
is enablement rather than a filing path).

The crowd-out objection (guardian, medium) is refuted by headroom: 151 scanned against a cap of
1500, `cap_binding false`; +32 is 12% of budget. It would have been CORRECT at the old cap of 500 —
it is the 2026-08-06 stopgap that makes it moot, which is a dependency worth knowing rather than a
bad objection.

## 2026-08-08 (evening) — LIVE on `v1.0.1268`, and a correction to this file's own 08-06 entry

### The voice_tells revalidator is LIVE. The 0 -> 1 transition completed.

> **POST-ROLL `v1.0.1268`, BOTH replicas:** `opting out is not evidence the copy was fixed` = **1** ·
> `the scan read nothing` = **1** · `an unserved page is not evidence the prose was fixed` = **1** ·
> positive control `auto:revalidated` = **2**
>
> Against the **BASELINE 2026-08-08T17:13:45Z on `v1.0.1267`: 0 / 0 / 0 / 2.**

That is the whole proof for a string-additive change, and it exists only because the 0 was taken
before the roll. **The build was not mine** — `v1.0.1268` is another session's; the change rode
along at HEAD, which is the normal condition here.

**Not yet proven by EFFECT.** The next scheduled sweep is ~08:37Z tomorrow. See below for why I did
not force it.

### ⚠ CORRECTION to the 2026-08-06 entry in this file: that hand-dispatch was NOT scoped

The 08-06 entry says, in bold: *"**Scoped deliberately**, not fleet-wide:
`{"site_id":"199733a8-…","item_type":"needs_page","dry_run":false,"max_items":50}`"*.

**Both filters are inert from `input_data`. Read the code:**

```go
config := params.StepConfig.Config          // revalidate_review_queue_action.go:266
siteFilter, _ := config["site_id"].(string) // :275
typeFilter, _ := config["item_type"].(string)   // :276
```

Confirmed against the live `diagnosis-review-queue-revalidator` definition: the `sweep` step has
**no `input_mapping`**, and its whole config is `{"dry_run": false, "max_items": 1500}` — no
`site_id`, no `item_type`. **So that run was fleet-wide across all four covered types.** It closed
exactly 1 row because exactly 1 was closable, which is indistinguishable from a filter working.

**Why this matters more than a tidy-up.** This lane had *already documented the trap* — "`item_type`
is read from `config` = `params.StepConfig.Config`, **the step config, not `input_data`**" — and the
08-06 session then wrote `input_data` filters two entries later and recorded the result as scoped.
That is the *second* instance in this lane of the same shape, and the first is already written down
as **"A trap you have just documented is not disarmed for your next paragraph."** Twice is a
pattern, not a slip: **documenting a trap creates a feeling of having handled it.**

The 08-06 conclusions are **unaffected** — both arms moved, the closed row was the predicted one,
and a fleet-wide run was a superset of the intended one. Only the *characterisation* was wrong, and
it would have misled the next session into believing a scoped dispatch is available. It is not.

### What I deliberately did NOT do, and why

**Did not hand-dispatch a sweep to prove the new revalidator by effect.** Not caution for its own
sake — the scoping I would have used does not exist:

- A dispatch **cannot be filtered** to `voice_tells` or to one site (above). It runs fleet-wide over
  **five** covered types now.
- Gate 2 stamps `result.revalidation` on every row it scans and does not close, so an unfiltered run
  writes to ~150 rows.
- The blast radius now includes my own untested-in-production arms on 32 live rows.
- **The scheduled run does exactly this at ~08:37Z anyway**, unattended, which is the condition the
  change was reviewed for. Waiting costs ~12 hours and nothing else.

The precedent is this file's own: the 08-05 session declined for the same reason, and the 08-06
dispatch happened **because the owner asked for one**. That is the bar; I have not been asked.

---

## 2026-08-09 — §0b CONFIRMED, but the number the handoff nominated could never have shown it

**The voice_tells revalidator works, proven behaviourally and unattended.** First post-roll
scheduled sweep, `2026-08-09 08:38:53Z`, nobody watching:

| field | pre-roll (08-08 08:38Z) | post-roll (08-09 08:38Z) |
|---|---|---|
| `scanned` | 151 | **186** (+35) |
| `cap_binding` | false | **false** ✓ |
| `resolved` | 3 | 2 |
| `uncovered_backlog` | 625 | **625 — UNCHANGED** |

`scanned` decomposes exactly: `required_fields_missing` 63 + `needs_section_data` 42 +
`unresolved_cta` 34 + **`voice_tells` 32** + `needs_page` 15 = 186. **All 32 live `voice_tells`
rows were scanned**, up from zero — the type was invisible to the sweep the day before.

And one closed: `voice:ecfd0bfd-bc5c-4ed4-9c45-7ba9143e72c8`, page `ai-readiness-quiz`,
`resolution_path='auto:revalidated'`, carrying this code's own resolved-arm reason string
("re-scanned all 3 rendered component(s) … against this site's current voice gate"). That is
behavioural proof, not a strings grep.

> **⚠ CORRECTED — `HANDOFF_2026-08-08b` §0b told the next session to confirm by watching
> `uncovered_backlog` FALL by ~32. It did not move, and it never could have been the right check.**
> The total is a sum over ~40 types. `voice_tells` left `uncovered_types` **entirely** (25 →
> absent), and nine other types grew by **exactly 25** in the same window (`claims_unverified` +5,
> `content_rewrite` +5, `lock_blocked_change` +5, `save_refused_incomplete` +4,
> `empty_internal_href` +2, and five more at +1). The coincidence is incidental; **any** inflow
> makes the total uninformative about one type.
>
> Had I trusted the nominated check I would have recorded "the adoption did nothing" on the day it
> worked perfectly. **Confirm at the per-type map (`uncovered_types['<type>']` must be ABSENT, not
> smaller) and at `scanned` decomposed by type.** In LANDMINES.md, and in CQ-020/CQ-021's
> verify-later so the next adopter cannot inherit the bad recipe.
>
> Second trap inside the fix: the per-row stamp key is **`result->'revalidation'->>'at'`**, not
> `checked_at`. I guessed `checked_at` first and got **0 rows**, which reads exactly like "nothing
> was scanned". Enumerate the keys before believing a zero.

### The pods rolled twice more while I worked — 1268 → 1269 → 1270

The handoff records `v1.0.1268`. At 22:20Z the pods were on **1269** (started 22:02Z, not my
build); by 09:36Z they were on **1270**. I re-greped the voice needles at 1269 rather than
assuming the change rode along: **1/1/1 on both replicas, positive control 2.** A roll is not
evidence your fix shipped, and that cuts both ways — it is also not evidence it *survived*.

### Dedup: 47 pairs / 168 rows, and the upward-drift claim does not hold

Re-measured against the PROPOSED predicate, with the exclusion list READ from
`pg_get_indexdef(oid) WHERE relname='idx_swi_dedup'` rather than reconstructed:

```sql
WITH cand AS (
  SELECT site_id, item_key FROM site_work_items
  WHERE item_key IS NOT NULL
    AND status <> ALL (ARRAY['complete','verified','rejected','wont_fix','failed','cancelled'])
), dup AS (SELECT site_id, item_key, count(*) n FROM cand GROUP BY 1,2 HAVING count(*)>1)
SELECT count(*) AS colliding_pairs, sum(n) AS rows_involved FROM dup;
-- 2026-08-09 00:2xZ: 47 | 168
```

> **⚠ CORRECTED — the handoff (§3.1) and `README_where_we_are` both say it "drifts upward roughly
> two pairs a day, so this gets more expensive with every day it waits."** Four points —
> 48/135 (08-03), 53/180, 55/184 (08-08), **47/168 (08-09)** — are noisy, not monotonic. It **fell
> 8 pairs in ~14 hours.** The point measurement stands; **the trend, and the urgency argument
> resting on it, does not.** §3.1 stays blocked on the owner's which-copy-do-I-keep judgement
> either way — that was always the real blocker, and dressing it up as a worsening problem was
> pressure the evidence did not support.

### `image_url_404` looked like the next adoption and is UNREACHABLE — the sweep has a status ceiling

It has the exact shape the census rewards: 26 open, **0 closed ever**, 0 `revalidation` blocks, 0
`deploy_result` blocks, flag-only, and its question ("does an active asset deploy to this path?")
is answerable from the DB with no HTTP via the same `storage.DeployedWebPath` the check uses.

It cannot be adopted. Its rows are `blocked` (23) and `detected` (3), and
`workItemRevalidatableStatuses` (`work_items_common.go:140`) is **`['needs_human_review',
'unresolved']`** — so the sweep never selects them. Worse, `reportUncoveredBacklog` counts with
**the same list**, so the type is not reported as uncovered either; it is simply **absent** from
`uncovered_types`.

That generalises. Parked rows outside the two revalidatable statuses, measured 2026-08-08:

| status | rows | types |
|---|---|---|
| `triaged` | 249 | 12 |
| `detected` | 114 | 20 |
| `deferred` | 66 | 10 |
| `blocked` | 34 | 3 |
| `claimed` | 3 | 3 |
| `awaiting_experience_plan` | 1 | 1 |
| **total invisible** | **467** | |

So the 625 this lane steers by is scoped by the same list that scopes the selection — **the
coverage-gap report cannot see the gap it exists to name.** `workItemRevalidatableStatuses`'
comment argues carefully for the two statuses it has and is silent on the six it excludes:
unargued, not deliberately argued.

**I did not act on this and did not assert it.** A cross-cutting structural claim about a shared
reporting mechanism is exactly what CLAUDE.md says goes through the loop first. Diagnosis queue
checked (empty, no duplicate), filed:

- intake `0c9b44d2-5c74-4322-aa78-7dd206f92689` · **run `f3d18013-0b78-472f-b2cb-5bf5e4e893b8`**
- item_key `needs_diagnosis:uncovered-backlog-status-ceiling`
- ⚠ the trigger warned that it reads `origin/087_towards_multiple_domains`, and local HEAD is
  ahead. I checked the diff for the two files the symptom names: `work_items_common.go` is
  identical and `revalidate_review_queue_action.go` differs by exactly the 7-line `voice_tells`
  registration, which does not touch the mechanism. The diagnosis sees the code I described.

**A REFUTED verdict here is a success.** Widening that list would be architecture-scope anyway —
it is interpolated in three places, and per its own comment widening the selection alone selects
rows the write-time CAS guards then silently refuse to update.

### Adopted `claims_unverified` instead — CQ-021, commit `4030cadb9`

Re-ran the CLOSER census over every uncovered type with selectable rows. Zero-closer types:
`claims_unverified` 23 · `lock_blocked_change` 23 · `image_source_unsatisfiable` 18 ·
`needs_sprite_css` 10 · `dead_control` 8 · `stale_evidence` 5.

Picked `claims_unverified`: widest site spread (7), `spec.page_id` on all 23 rows, HITL-terminal,
0 closed ever, 0 `deploy_result`, locked-components skipped by the emit side, and an opt-in
register that gives it the same "site opted out" arm the voice gate has.

Two things differ from the voice_tells case and both are written into CQ-021 rather than
discovered later:

1. **TWO producing checks converge on one `item_type`** (`check_unverified_claims.go` and
   `check_unverified_claims_stats.go`). That is the owner's 2026-08-02 §1 case — no RFC needed
   *provided* the producer set and shared `item_key` shape are stated in the register entry. They
   are.
2. **`ScanDeployedClaims` has no page-status filter** where `ScanVoiceTells` restricts to
   `active`/`deployed`. Preserved deliberately: the revalidator must judge by the emit side's exact
   predicate. Consequence stated in the landmine — an item can resolve on an archived page.

Disconfirming check run **before** writing code: **16 of 23** pages have a component updated since
filing. A 0 would have killed the change. Arm reachability measured in the same pass —
`page_missing 2` and `site_has_no_evidence_base 1`, so those two refusal arms are **exercised** on
today's data (voice_tells had none exercised); `has_locked_components 0`, so the two locked arms
are reasoned and unit-tested, marked [UNEXERCISED] rather than dressed up.

Council `b67eb26a-14ef-45d7-b755-3e489fd57ef0` submitted alongside; committed with
`Council-Submitted:`, no `Council-Reviewed:` until I have read a verdict.

**Misstep worth recording:** the 097 trigger rejected my first submission —
`.plan.edits[].operation` takes `modify|add|remove|config_change`, and I wrote `create` for the two
new files. Client-side refusal, no credits spent. `add` is the spelling for a new file.

**Not mine, flagged:** `TestValidDocSubjectTypes_LockstepWithMigrationCheck` and
`TestEveryCheckProducedItemTypeIsClassified` fail in this package at HEAD. Reproduced on a clean
`git archive HEAD` extraction with none of my edits, so they predate this change — they belong to
`e1628f7df` (RFC_015 decision records, another lane). The second wants `decision_regression`
registered with a verifier or acknowledged as a gap.

### The 090 came back UNVERIFIABLE — on TOOLING, and the reason is worth more than the run

Run `f3d18013-…`: **`UNVERIFIABLE — stopped: iteration-cap`.** Its own words:

> *"Two repeated symbol/content searches for the literal declaration
> (`workItemRevalidatableStatuses = []string{...}`) returned 0 rows — per the index-staleness
> caveat this is **unknown, not proof** that the list omits those statuses."*

It never saw the list. I checked why, and it is structural rather than staleness:

```sql
SELECT kind, count(*) FROM code_symbols GROUP BY 1 ORDER BY 2 DESC;
-- func 3592 | method 1114 | struct 973 | alias 40 | interface 36   -- and NOTHING else
SELECT symbol, kind FROM code_symbols WHERE path='platform/orchestration/actions/work_items_common.go';
-- sqlInList, countDispatchableWorkItems, workItemKey, resolveWorkItems  -- 4 funcs, 0 vars
```

**The code index has no `var` or `const` kind at all.** So `workItemRevalidatableStatuses`,
`workItemTerminalStatuses`, `reviewRevalidators`, `itemTypesWithoutVerifiers` — every registry,
status list, threshold and allow-list this codebase argues about — is unreachable via `SEED_SCOPE`.
I named one and paid a full run for it.

⚠ **`UNVERIFIABLE` is not `REFUTED`.** It says the loop could not answer, not that the premise is
false. The premise is independently first-hand verified: `work_items_common.go:140-143` is literally
`{"needs_human_review", "unresolved"}`, and `image_url_404` has 26 open rows and appears nowhere in
the live `uncovered_types` map. What was missing was an *independent* read, which is the whole point
of filing — so I re-filed rather than declaring myself right.

Re-filed **once**, function-scoped, every seed symbol confirmed present in `code_symbols` first:
run **`a174b184-dac2-47a1-95ca-df2d192e183a`**, seeds `reportUncoveredBacklog` /
`loadParkedReviewItems` / `coveredItemTypes`, and the symptom re-framed around the observable
(`uncovered_types` omits types that have open rows) so it no longer depends on fetching an
unfetchable symbol. Landmine added with the pre-flight check.

**Also done:** dated CONSUMER NOTICEs for `claims_unverified` into `bugs_open/033` and
`bugs_open/083`, and the `voice_tells` notices in both updated to say it is now LIVE with a first
closure. `083` additionally gets the status-ceiling warning, since "findings never reach a handler"
is precisely the file whose subject those 467 invisible rows are.

### Council REVISE — and the misstep that earned it

`b67eb26a-…` came back **REVISE**, 15 seats, gated by `editquality` (HIGH). Full read-out and every
answer with its check: `OBJECTIONS_2026-08-09_claims_unverified_council.md`. Resubmitted round 2
under the same trail; corrections committed `6ab7ff594`.

**The misstep, which is the entry worth keeping.** I asserted **"TWO converging producers"** and
then invoked the owner ruling of 2026-08-02 §1 as my authority for shipping a shared-vocabulary
change without an RFC. **There is one producer.** `check_unverified_claims_stats.go` registers no
check, has no `init()`, and emits no `WorkItemSpec`; its `scanStoredStatClaims()` has exactly two
production call sites, both inside `ScanDeployedClaims`. What misled me was its own header line —
*"reuses the existing `claims_unverified` item type, so it incurs no [new type]"* — which describes
contributing findings to a type, and which I read as filing items under it.

**The cheap check that would have caught it, and which I skipped:** before calling anything a
producer, grep it for the emission, not for the type name.

```bash
grep -n "WorkItemSpec\|ItemKey\|ItemType\|func init\|Register(" <candidate.go>   # no output ⇒ not a producer
grep -rn "<itsScanFunc>(" --include=*.go platform/ | grep -v "func <itsScanFunc>"  # where is it actually called?
```

This is the same family as the entry two above about `created_by` — **"two things mention the type"
is not "two things file the type"**, and both times I reached for a name rather than a call graph.
Recorded there as well.

⚠ **It is worse than a wording slip, because the false claim had already propagated into the
concept register**, which the next council round reads as ground truth (bugfix 161's lesson: the
register both instructs the writer and vouches for the claim). Corrected visibly in all five places
— code header, map comment, coverage test, CQ-021 entry, index row — rather than quietly dropped.

**The seat's actual question was better than my error**, and its answer is favourable: it asked
whether the shared scan reproduces the *second producer's logic*, not merely its item_key shape.
It does, structurally — ONE scan, both halves inside it — so the feared "judged by the wrong
predicate" outcome cannot occur.

**Measurements the other seats asked for**, all folded into `grounded_in`:

| seat | check | result |
|---|---|---|
| `guardian` | anything else close this type? | `ever_closed 0 · deploy_result 0 · handler_agent 0 · distinct resolution_paths 0` |
| `bug_historian` | archived pages in the population? | **active 19 · archived 2 · deleted 2** — exposure is real, 2 of 23 |
| `guidelines` | does adding a type starve the budget? | `max_items 1500` vs `scanned 186`, `cap_binding false` — ~14% of cap |
| `debug_historian` | deploy verification? | existed, wasn't in the submission; baseline already spent (§0c) |

⚠ **AND A TRAP READING THE VERDICT.** CLAUDE.md's documented command —
`SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1` —
returned **another lane's verdict** (`46f87e4c-…`, `bugs_open/228`). On a tree this concurrent,
"most recent" is almost never yours, and the note reads perfectly plausibly until you notice it
discusses contact forms. Read by **your own correlation**, and note the objections live in
`diagnosis_artifacts.body`, not in `metadata` (which carries only `decision`/`reviewers`/`abstained`):

```sql
SELECT body FROM diagnosis_artifacts
WHERE correlation_id='<YOUR_CORR>' AND kind='council_report';
```

**Left for the owner, deliberately:** the `compliance` HIGH objection that auto-closing a
factual-claims HITL type is a policy change. Three seats, three mandates, all routed it to a human.
Options costed in `README_where_we_are`; recommendation is the ~4-line "only close if the copy
changed since filing" gate, which converts the objection into a mechanical guarantee. **Not
resubmitted around** — a policy veto is not answered with better measurements.
