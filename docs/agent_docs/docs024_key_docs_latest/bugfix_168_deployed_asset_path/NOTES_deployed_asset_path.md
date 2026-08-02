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
