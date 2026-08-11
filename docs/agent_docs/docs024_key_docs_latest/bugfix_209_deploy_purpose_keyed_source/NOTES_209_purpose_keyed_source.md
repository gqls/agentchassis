# NOTES — 209, `deploy_image_asset` purpose-keyed source

Append-only, newest at the bottom. Technical log: evidence, commands, and every
misstep — including my own claims in this file that turned out false.

---

## 2026-08-08 (late) — session 1: verification before any fix

Picked up from `bugfix_221_ai_disclosure_precision/HANDOFF_2026-08-08_221_done_209_is_next.md`.
That handoff set two preconditions: re-check ownership myself, and re-verify the
bug is live. Both done before touching code. The second one produced a different
answer from the one the bug file expected, so no fix was written this session.

### Ownership — clear, and the `who-owns.py` hit is a false positive

`scripts/who-owns.py 209` returns **OWNED**, naming
`bugfix_221_ai_disclosure_precision` — my own predecessor lane. That is the same
false-positive shape the 221 handoff already flagged for bug 223: the handoff
*cites* 209, and `who-owns` reads commits, so citing looks like owning.

Live-transcript sweep (the check that catches a session mid-fix, which
`who-owns` cannot see):

- `98b5904b` — my predecessor; it wrote the handoff (`bdaff9cad`). Ended 22:12.
- `0581eab4` — bug 220 lane, council round 2 approved. Ended 22:11. Unrelated.
- `693556a1` — bug 203 CTA-resolver lane. Its `findStorageURI` hits are incidental
  reads of the same `actions` package. Ended 19:45, committed clean.

**No session is working 209.** Also learned in passing: fleet credit came back
~22:06 (the 220 lane got a council verdict), so 221's outstanding landmine-verifier
re-fire is unblocked.

### The defect is still present at HEAD, and it is wider than the bug file says

`findStorageURI` Priority 2 is intact — `deploy_image_asset_action.go:454-458`.
Writers confirmed at `v3_site_actions.go:2852` and `generate_image_actions.go:994`
(the bug file's line numbers had drifted; it said 2810).

**[MEASURED] New: Priority 2 is not the only purpose-keyed route.** Priorities
**3–7** are *all* keyed on `purpose` too — `{purpose}_result.image_uri`,
`{purpose}_result.response.generate.response.image_uri`, `{purpose}_stored.asset_url`
and so on (`:460-495`). The bug file names only Priority 2. Any fix phrased as
"delete Priority 2 so the asset_id path is the only DB-free route" does not
achieve that: it falls through to five more purpose-keyed lookups.

### Is it live? Config census over every live definition

```sql
SELECT type FROM agent_definitions
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config::text LIKE '%deploy_image_asset%';
```
→ exactly **three**: `asset-deployer`, `pageflow-builder`, `site-work-orchestrator`.

(Deliberately a DB question, not a repo one — 221's lane was corrected by the
council for reporting a Go call-site count as if it were the live caller count.)

The trigger shape 209 needs is **two same-purpose assets stored, then deployed
from `collected_data`, in ONE run.** Step-by-step against each:

| definition | store steps | deploy steps | can it collide? |
|---|---|---|---|
| `asset-deployer` | none | 1 (`deploy_asset`, `input_fields` include `s3_uri` **and** `asset_id`) | no — one deploy per run |
| `pageflow-builder` | hero + logo | hero + logo | no — **different purposes**, so `hero_uri`/`logo_uri` are distinct slots |
| `site-work-orchestrator` | hero + logo | hero + logo | no — same |

`image-build-handler` is the suspect the bug file names, and it **does** have the
first half: two `store_asset` steps *both* with static `purpose: "hero"`
(`store_hero_asset`, `store_imagery_brand_asset`). But:

- they sit on **mutually exclusive conditional branches** — `check_item_type_imagery`
  routes `needs_imagery` to the imagery branch, everything else to `check_item_type`;
  `check_imagery_brand_update` then picks brand-vs-plain. One store per run.
- it has **no deploy step at all**. It delegates via `call_asset_deployer`, whose
  `input_mapping` sets `s3_uri: asset_stored.image_uri` — the source of the asset it
  just stored, carried by identity, not looked up by purpose.

`StoreAssetAction` does populate that: it returns both `image_uri` and `s3_uri`
when the storage URI is non-empty (`v3_site_actions.go:~2889`).

**[MEASURED] The child cannot reach a purpose slot even if `s3_uri` were empty.**
Enumerated the top-level `collected_data` keys across all 18 live `asset-deployer`
rows (enumerated keys rather than probing a path, per the jsonb landmine):
`input_data`, `deploy_asset`, `deploy_result`, `check_*`, `agent_config`, the `__*__`
infrastructure keys. **No `{purpose}_uri` key of any kind.** And
`ExtractNestedFieldString` → `ExtractNestedField` (`data_helpers.go:1199`) is a
**strict dotted-path walk** with one `.response` unwrap — *not* a recursive search.
So Priority 2 finds nothing there and the action skips safely
("no storage URI found"), which is also what the deliberate
`error_step: spawn_asset_deployer` path degrades to.

**Conclusion: the defect is real in code and currently unreachable in live
configuration.** That upgrades the bug file's `[UNMEASURED]` to a measured
negative. It does not make the code correct — it makes it latent.

### MISSTEP — I claimed "26 days" and it was wrong

I first wrote that `pageflow-builder` and `site-work-orchestrator` "have not run in
26 days", reasoning from `orchestration_states` holding rows back to 2026-07-13.
**That was an overstatement and I caught it by asking what the old tail actually
contains:**

```sql
SELECT CASE WHEN created_at > now() - interval '24 hours' THEN 'last_24h'
            WHEN created_at > now() - interval '7 days'  THEN '1-7d'
            ELSE 'older_than_7d' END AS age, status, count(*)
FROM orchestration_states GROUP BY 1,2;
```

Only **13** `COMPLETED` rows survive past 24h; past 7 days there are **zero** —
the tail is `CANCELLED`/`RUNNING`/`INITIALIZED` only. So the table is effectively a
**~24-hour window for completed runs**, and absence in it means "has not run
today", not "has not run in 26 days". The cheap check that would have caught it
first time: never read a retention window off `min(created_at)` for the whole
table — bucket it by status, because the survivors are selected, not
representative.

### A second measurement I had to throw away — the positive control failed

Tried `llm_call_log` for longer retention (it goes back to 2026-03-25, 50,861
rows). It returned **0 rows for all four agent types** — including `asset-deployer`
and `image-build-handler`, which I had *just watched run* 16 and 8 times today.
The positive control fails, so the query is blind to these agent types and proves
nothing in either direction. **Discarded rather than reported.** (Consistent with
the known `llm_call_log` agent-type traps.)

So the honest word for the legacy pair is **dormant, not dead**: no live agent
definition spawns or calls either of them (0 rows over every active definition's
steps), and neither ran today — but Go/script/topic dispatch has not been excluded,
and the completed-run window is only 24h.

### The finding that actually changes the plan: fix candidate 1 is unsafe

The bug file ranks first: *"Delete Priority 2 and make the asset_id path the only
DB-free route."* Both legacy deploy steps carry **no `input_fields`**, so
`deploy_image_asset` resolves their inputs through `ExtractActionInputs`
**Strategy 2** — `ExtractFields` → `extractSingleField` → **Strategy 4, aggressive
recursive search** (`unified_extractor.go:439`). `findFieldRecursive` walks
`for key, val := range m` (`:494`) — **Go randomises map iteration order.**

With both `hero_stored.asset_id` and `logo_stored.asset_id` in `collected_data`,
`asset_id` therefore resolves *nondeterministically*. Measured by running the real
helper 400 times on identical input
(`deploy_image_asset_purpose_source_test.go`):

```
asset_id resolutions over 400 identical inputs:
  hero 11111111-… : 344     <- WRONG asset for a logo deploy step
  logo 22222222-… :  56
```

**The logo deploy step resolved the hero's `asset_id` in 344/400 runs (86%).**

So Priority 2 is not merely legacy cruft — for the hero+logo workflows it is the
*correct* discriminator, because their two assets differ precisely by purpose.
Deleting it would replace a correct per-purpose lookup with an 86%-wrong
recursive one. **Fix candidate 1 would introduce the wrong-bytes bug it exists to
prevent.** Recorded here rather than only in my head because the bug file states
the opposite ranking.

What the legacy steps actually resolve through today, also measured: both carry
`uri_field` (`hero_result.image_uri` / `logo_result.image_uri`), which the spec's
`Deprecated` map bridges to `s3_uri` — so `inputs.Get("s3_uri")` is already
populated with the right per-purpose source *before* `findStorageURI` is reached.

### What exists now

`platform/orchestration/actions/deploy_image_asset_purpose_source_test.go` — four
characterisation tests, all passing, pinning: the last-write-wins branch; that
distinct purposes do not collide; the 86% asset_id instability; and which route
supplies the legacy source today. Written to characterise, not to assert a fix, so
a future naive deletion of Priority 2 trips over the recorded reason it was kept.

### Pre-existing HEAD failure in the same package — NOT mine, not fixed

`go test ./platform/orchestration/actions/` has exactly **one** failure,
`TestValidDocSubjectTypes_LockstepWithMigrationCheck`: migration 340 adds `decision`
to `doc_notes`' subject types but `validDocSubjectTypes`
(`doc_subjects_common.go:63`) does not carry it. Every input to that test is
committed and unmodified in this tree (`git status` clean for all three files), so
the result is identical to HEAD's — it is another lane's lockstep break
(`bugs_open/064`; checklist at `experience_register/design/subject_type_addition.md`),
not a consequence of anything here. Recorded so the next reader of this lane does
not spend the same five minutes on it. My four tests pass.

## 2026-08-08, 22:33Z — the landmine entry was independently verified

Fired `trigger-landmine-verifier.sh` on the new entry. **Verdict: `STILL_VALID`** —
all four footprint symbols (`ExtractActionInputs`, `ExtractFields`,
`extractSingleField`, `findFieldRecursive`) and both files resolved, and the verdict
confirms the call chain and the randomised-map-iteration hazard by reading the code
independently of my test.

That verdict is worth more than usual because it doubles as the **disconfirming
control** for something else measured in the same batch: the 221 lane's entry, fired
minutes earlier, came back `NEEDS_HUMAN_REVIEW` claiming its symbols "no longer
resolve as standalone symbols (possibly inlined or renamed)". They do exist, at
`validate_page_content.go:105` and `:1229` — they are just declared
`var X = []struct{…}`, and `code_symbols` holds **no `var` kind at all** (func 3592,
method 1114, struct 973, alias 40, interface 36; total 5,755). My entry's footprints
are all `func`, which is exactly why mine resolved and theirs could not. Filed into
`bugs_open/223` as a third failure mode; the 221 entry was **not** downgraded.

The transferable point for this lane: **a landmine footprint should name a `func`,
`method`, `struct`, `alias` or `interface` if you want the verifier to be able to
check it.** A footprint naming a package-level `var`, a table or a command is
unverifiable by construction today, and the verifier will not say so — it will
suggest a rename instead.

---

## 2026-08-09 morning — session 2: post-roll re-verification, and one claim sharpened

Chassis rolled to **v1.0.1270** (both replicas started 08:49Z; Makefile already
bumped to 1271 by the 228 lane for its next build). Nothing of this lane's is in
the binary — the commit was test+docs only — so there is no pod-grep to do for 209.
What the roll *did* do is re-apply seeds: all four relevant `agent_definitions`
rows show `updated_at = 08:49:01`, the pod start minute.

**[MEASURED] A bumped `updated_at` cannot tell "re-seeded identical" from
"changed", so the four deciding facts were re-read by content, not timestamp:**
legacy pair store/deploy purposes + `uri_field`s; `image-build-handler`'s two
branch conditions; its `call_asset_deployer` `input_mapping`; `asset-deployer`'s
deploy `input_fields`. **All byte-identical to the 08-08 census.** The LATENT
verdict stands on v1.0.1270.

Also re-run at new HEAD: the four characterisation tests pass; the instability
split this run was 348/400 (it moves run to run — randomised iteration — the
hazard is the plurality of values, not any particular ratio). `git log
ae990ee82..HEAD` over every 209-relevant file: empty. The package's one failing
test is still the pre-existing `TestValidDocSubjectTypes_LockstepWithMigrationCheck`.

### SHARPENED (not retracted): where fix candidate 1's 86% actually bites

My 08-08 write-up says deleting the purpose-keyed lookups "swaps a correct lookup
for an 86%-wrong one". Re-reading the resolution order, that sentence deserves
precision:

- The legacy steps' **primary** route is the `uri_field` bridge → `s3_uri`
  (`hero_result.image_uri` / `logo_result.image_uri`), which is per-purpose and
  correct, and which candidate 1 does not touch.
- The purpose-keyed `findStorageURI` lookups are those steps' **fallback** — they
  run only when the `*_result` map is absent or carries no `image_uri`.
- Candidate 1 therefore converts a **correct fallback** (purpose-keyed) into a
  **wrong-bytes fallback** (`asset_id` by recursive search, 86% the wrong sibling
  when both stores have run) — it does not break the happy path.

The ranking conclusion is unchanged — a fallback that deploys the wrong asset's
bytes on the day the primary route hiccups is strictly worse than one that stays
correct, and "the primary failed" is precisely the regime a fallback exists for.
But the exposure is **conditional**, and the earlier phrasing let it read as
unconditional. Recorded here and as a dated addendum in the bug file.

## 2026-08-09, mid-morning — owner ruling received; and the 064-recurrence traced on request

**Owner ruling (verbatim in README/PLAN/bug file):** the legacy pair is *"not
dead, but not being worked on"*; divergence via new actions/workflows is licensed;
explain the problem clearly before moving further. PLAN now carries the two
divergence shapes (opt-in field, recommended, per the 2026-08-02 opt-in-field
ruling; vs a new action) as **DECISION PENDING** — no code until the owner picks.

**The 064 question, answered.** Asked: which thread is working `bugs_open/064`.
Findings, evidence inline:

- 064 is **closed** (`bugs_closed/064`, 2026-07-24, fixed `c9cc95a5a`). The red
  test is a **recurrence of its shape** — the second (`7290433f2`, 07-31, was the
  first, missing `landmine`).
- Introduced by `e1628f7df` (08-08 20:21), RFC_015 decision records: migration
  340 adds `decision` to `doc_notes`' accepted types; `validDocSubjectTypes`
  (`doc_subjects_common.go:63`) not updated. The commit's non-platform files land
  in `idea_uk_vm_site/` — **that lane owns it.**
- **Nobody is fixing it**: `git log e1628f7df..HEAD -- …doc_subjects_common.go`
  is empty; the idea_uk_vm_site docs never mention the test (grep empty); the two
  lanes that hit it (this one; `bugfix_226` at 08-08 22:25, seen in its live
  transcript) both recorded it as pre-existing and stepped around.
- Lane told: `idea_uk_vm_site/CONTRIB_2026-08-09_rfc015_broke_the_doc_subject_lockstep_test.md`,
  pointing at the one-word fix and 064's own checklist (which has an ordering
  catch: the migration is already live, so the catch-up order matters).

`who-owns.py 064` names `experience_register` and `staged_component_build` as
likely owners — **both false positives for the current break**: they own the
*mechanism's history* (064's original fix and the `component` addition), not the
08-08 recurrence. Same lesson as the 209/223 hits: `who-owns` measures who has
*touched the subject*, and recency of citation is not authorship of the defect.

## 2026-08-09, late morning — costing "into line" found a live defect: the Defaults shadow (bugs_open/231)

Owner asked: how much surgery to bring the legacy pair into line rather than
diverge? Costing it required knowing how their deploy steps' `purpose` actually
resolves under the modern spec — and the probe answered with a bug.

**[MEASURED] The logo deploy steps of both legacy workflows resolve
`purpose="hero"` today.** Mechanism, each arm cited and *executed*:

- `ExtractActionInputs` applies `spec.Defaults` into `Values` **first**
  (`action_inputs.go:457-460`); `DeployImageAssetInputSpec` defaults
  `purpose: "hero"` (since `34d2315ce`, **2026-02-20**).
- Strategies 1/2/3 each skip a field already in `Values` (`:499`, `:511`, `:523`)
  — a defaulted field always is. Strategy 0 reads only **dotted** config paths
  (`:478`) — static `"logo"` is invisible to it.
- The action's own fallback (`deploy_image_asset_action.go:92-99`,
  `if purpose == "" { config["purpose"] }`) is **unreachable** — the default means
  purpose is never empty.
- Consequence if run: hero resize class, and `BuildAssetPaths` (`url_helpers.go:190`,
  `filename = purpose + ext`) commits the logo's bytes to the **hero's path** —
  clobbering the hero and never writing the `logo_url` the store step just
  promised in `content_data`.

Proven by running the real resolver + real spec (a probe first — deliberately
failing, removed from the shared tree within a minute of confirming — then three
committed characterisation tests that PASS by pinning the defect:
`TestLegacyLogoStep_StaticPurposeIsShadowedByDefault`,
`TestPurposeFieldBridge_DeadForDefaultedField`,
`TestStrategy0DottedPaths_DefeatTheDefaultAndTheRecursiveSearch`). The third also
proves the repair mechanism: a Strategy-0 dotted path defeats both the default
and the recursive search, deterministically (100/100).

`[UNMEASURED]` whether the shadow ever fired live — the pair has no dispatcher and
completed-run retention is ~24h. The honest state: **broken-if-run today, at the
resolver level.** Filed as `bugs_open/231` with the method-note substitution
stated per the 07-31 ruling; the **fleet class** (any static config value for any
spec-defaulted field, ~10+ specs carry `Defaults`) was NOT asserted — handed to
the loop: **090 fired, RUN_CORRELATION_ID `e952039b-dcf1-4218-864e-c8c7a1a23a85`**.
LANDMINES entry added (func-shaped footprints, per this lane's 08-08 lesson) and
synced.

**What the discovery does to the costing:** the pair needs the same 4-step config
edit *just to be runnable* ("not dead" — owner) that "into line" Phase 1 consists
of. So repair and alignment are one edit, and divergence-without-repair leaves
the pair logo-broken. Full costed comparison in README_where_we_are (2026-08-09,
late morning) — the chat answer and that entry are the same text.

**Blast-radius note for the eventual fix:** the two definitions carry **6
references each** to `hero_uri|logo_uri|hero_result|logo_result` beyond the
deploy steps (measured via `jsonb_each_text` sweep) — so retiring the
`{purpose}_uri` **writers** is NOT part of Phase 1/2; those references must be
classified first (Phase 3, optional, not needed for 209). Also noted in passing:
`generate_image_actions.go:999` writes `contentData[purpose+"_uri"]` into the
in-memory site record — 155-adjacent, not chased here.

## 2026-08-09, midday — PHASE 1 EXECUTED: migration 348 applied, row-verified, harness extended to the live shape

Owner approved into-line, Phase 1 first. Executed this session:

- **Pre-flight facts nailed before writing SQL:** StoreAssetAction is NOT
  shadowed (reads `config["purpose"]` directly, `v3_site_actions.go:2603-2611` —
  no ActionInputSpec), so the dotted paths carry true values. And the deploy-time
  definition stamp (all 175 active rows at 08:49:01) **preserves content** — the
  measured control is migration 341's `gate_next_item` surviving this morning's
  stamp; without that control a DB-only config migration could not be trusted
  across rolls.
- **Design decision recorded in the migration header:** `input_fields`
  deliberately EXCLUDES `s3_uri`. Strategy 0 still resolves it when present
  (Strategy 0 iterates SPEC fields, not input_fields), but on a store failure an
  input_fields entry would send Strategy 1's aggressive search hunting a URI and
  it can land on the SIBLING asset's — the 209 class through a side door. Excluded,
  the corner resolves s3_uri="" → asset_id (present even on insert-failure, row
  absent) → "" → safe visible skip. This deliberately NARROWS behaviour vs the old
  uri_field route (which would deploy the generator's URI, bypassing the asset
  row — the 152/155 identity-bypass pattern).
- **Discipline sequence:** post-verify DO/RAISE **induced first** (standalone vs
  unmigrated rows → raised "0 of 4"); scoped dry-run (doomed txn, clean); scoped
  apply via `MIGRATIONS_DIR` scratch dir (pending set held other lanes' 342/345/346
  — untouched); row-verify by content, store steps as negative control (untouched);
  `schema_migrations` row present (09:41:53, run-migrations.sh).
- **Two config keys surfaced by the full-row read** that my earlier
  two-column queries had hidden: `domain_field: site_record.domain` (pre-existing,
  now inert — Strategy-0 `domain` wins, Strategy-3 bridge skips) and
  `output_mapping` on the hero steps only. Both preserved by the surgical
  `- 'uri_field' || {...}` merge. Lesson re-learned from my own 08-08 entry: a
  column-selected read is a filtered read; verify against the FULL object.
- **Harness now 8/8**, extended to the exact live shape (incl. `domain_field`)
  plus the corner test `TestMigration348Shape_StoreFailureResolvesNoURI_NeverTheSibling`
  (100 iterations, sibling never leaks). The pre-348 shadow test's header now
  marks its config shape as history while noting the resolver behaviour it pins
  is still current for any config authored that way.

### The 090 verdict on the fleet class (run e952039b): UNVERIFIABLE, scope-not-narrowing — read precisely

Not a refutation. The loop **independently confirmed the helper mechanism**
("…is real and would silently drop a static non-dotted config literal for a
defaulted field") — that is the cross-cutting core corroborated by a second
reader. It declined the *instance* for two stated gaps: (1) it could not fetch
`DeployImageAssetInputSpec`'s **declaration** — only its two use sites — because
a package-level `var` is unrepresentable in `code_symbols`: **bug 223's
var-blindness biting the diagnosis loop's own lookup layer** (second consumer;
addendum added to my 223 block). My committed test answers this gap by
*executing* the real spec. (2) No runtime row shows the bug firing — already
`[UNMEASURED]` in 231, and expected: the workflows do not run. The fleet-class
CENSUS remains undone; 231 stays OPEN for it, plus candidates 2/3 and the
behavioural proof.

---

## 2026-08-09 (afternoon) — the behavioural proof ran, and it found live 231 damage across the fleet

### Setup

Sacrificial domain `cookly.uk` (owner-named), seeded as a `sites` row with an
object-shaped brief (no `github_repo` → `resolveGitRepoNameDB` defaults to `sites`).
Dispatched with `071b_new_build`'s "trigger just pageflow-builder" message, published
with the payload in the container COMMAND plus a `PUBLISH_OK` marker (the
`kubectl run -i | kcat -P` silent-drop trap). Correlation
`0562a667-122f-41ed-8fee-841180264367`, orchestration
`22fb157a-322b-4c56-bc00-51e549db3060`. Scripts in RUNBOOK §10.

### `[MEASURED]` Migration 348 works end-to-end, and `findStorageURI` is NOT on the live path

The **hero** deploy step is the only one that runs with BOTH assets present in
`collected_data` (order: logo-generate → logo-store → logo-deploy → hero-generate →
hero-store → hero-deploy), so it is the step where the 86% wrong-asset hazard could
bite. Its log:

```
13:11:39.238 Strategy 0: Resolved config path  field=domain   path=site_record.domain
13:11:39.238 Strategy 0: Resolved config path  field=s3_uri   path=hero_stored.s3_uri
13:11:39.238 Strategy 0: Resolved config path  field=purpose  path=hero_stored.purpose
13:11:39.238 Strategy 0: Resolved config path  field=asset_id path=hero_stored.asset_id
13:11:39.239 Downloading image from S3  purpose=hero key=images/…/fa752a71-…png
13:11:41.649 deploy_image_asset: recorded local url on asset  asset_id=f015cd0c-… url=/assets/images/hero.jpg
```

`fa752a71` is **hero_stored's own** object; the logo's is `bd7308a5`. The asset row
stamped is the hero's own. The same log ALSO shows `Found via aggressive search` for
`purpose` and `asset_id` — the randomised mechanism ran and its result was
**discarded**, because Strategy 0 had already populated those fields
(`action_inputs.go:499`, skip-if-already-resolved). The hazard firing and being
overridden in the same step is the strongest single line of evidence here.

**Consequence for Phase 2:** `inputs.Get("s3_uri")` was non-empty, so
`findStorageURI` was never reached. Deleting it cannot change this workflow.

> **CORRECTION to `HANDOFF_2026-08-09b`'s own wording.** It says 348 "deliberately
> excluded s3_uri". True of `input_fields` ONLY — `s3_uri` is in
> `DeployImageAssetInputSpec.Optional`, so **Strategy 0 resolves the config dotted
> path regardless of `input_fields`**. The exclusion suppresses the *aggressive
> search* (Strategy 1), not the explicit path. The shorthand reads as if the field
> were unresolved, which would invert the Phase 2 risk assessment.

### `[MEASURED]` The byte comparison both bug files specify is NOT disconfirmable — do not cite it alone

`hero.jpg` sha `b2a8368…` (140,534 B) vs `logo.png` sha `4e9a947…` (239,735 B):
distinct. **But that test could not have come out otherwise.** The deploy re-encodes
per purpose (hero→jpg, logo→png), so the two outputs differ in bytes *even when the
wrong source is fetched*. The disconfirmable measurements are the **downloaded
object key** and the **asset row stamped**, both above. Recorded because 209 §"How
to verify a fix" and 231 both name the byte comparison as the bar.

### `[MEASURED]` 231 DID fire live — 11 sites serve a logo that was processed as a HERO

231 records `[UNMEASURED] Whether this ever fired live`. It did. Census of every
committed logo in `gqls/sites`, with its producing commit subject:

- **`logo.png` × 4** — ai-agent-orchestration 02-22, finetuning 03-02, leopardess
  07-10 (hand-made brand commit), **cookly 08-09 (this run)**. All **400×400 PNG**,
  subject "Deploy **logo** image".
- **`logo.jpg` × 11** — gamesdesign 06-06, idea.uk 06-21, vonc 06-23, dartsonline
  07-06, robot-hands 07-10, vetcomparison 07-17, fundamentallyai 07-21, oufe and
  webdesign.co.uk 07-25, relojistas 07-29 (later hand-replaced), lendzy 08-02,
  webdesign.uk 08-04. All **JPEG at 1408×768 / 900×900 / 646×275**, subject
  "Deploy **hero** image".

Three independent signals agree `purpose == "hero"` at those deploys:

1. `deploy_image_asset_action.go:579` builds the message as
   `fmt.Sprintf("Deploy %s image for %s", purpose, domain)` — the subject *is* the
   resolved purpose.
2. `DeployedAssetPath` takes the **extension from purpose** and the **filename from
   asset_key** (`url_helpers.go:317-330`); `ImagePurposes["logo"]` is
   `{400,400,90,"png"}` (`:364`), so a `.jpg` is unreachable with purpose "logo".
3. `DownloadOptimizeAndPrepare(…, purpose, …)` drives the resize — and not one of
   the 11 is 400×400; they keep their generation dimensions.

**Worse than a naming slip:** JPEG carries no alpha channel, so each of those logos
is served with an opaque background, at up to 1408×768 instead of 400×400.

`[UNVERIFIED] — this is the 090's question:` **which caller/config produced each
one.** 231's "proven instance" (the legacy pair's static `purpose: "logo"`) predicts
the logo landing on **the hero's own path** (`hero.jpg`), because it assumed no
`asset_key`. The artefacts show `logo.jpg`, so an `asset_key` of "logo" WAS supplied
— the historic producer is **not** the shape 231 modelled. `assets` points at
`asset-deployer`: four rows are literally named `input-data.asset-key.jpg`, an
unresolved `input_data.asset_key` config path leaked in as the asset key, again with
hero's extension. Do not assert the producer without the 090.

**So migration 348 did not fix the fleet's logo problem.** It fixed the pair that was
not running. The exposure that actually shipped runs through whatever calls
`asset-deployer`, and it is still open.

### Traps paid for this session

- **The per-agent pods keep ~11 SECONDS of logs.** `kubectl logs --since=20m` on
  `agent-pageflow-builder` returned its earliest line at 13:11:38 for a fetch at
  ~13:12 — the logo deploy at 13:11:27 was already gone. Attach `logs -f` BEFORE
  dispatching, or accept that only the final step is evidenced.
- **Only `agent-chassis` is a Deployment**; the per-agent pods
  (`agent-pageflow-builder-…`) are spawned per run, so you cannot pre-tail one by
  name — poll for the pod and attach the moment it appears.
- The fleet rolled to **v1.0.1274 at 12:23 UTC**, *between* this session's config
  check and its dispatch. Re-verified the four 348 steps by CONTENT afterwards:
  intact, `updated_at` re-stamped to 12:22:37 — RUNBOOK §8's exact pattern.
- `sites` has **no `site_id` column** (it is `id`), and `content_data` is an
  **array** on some sites and an object on others, so `jsonb_object_keys` errors on
  the former. Both cost a query.

---

## 2026-08-09 (late afternoon) — SWO arm done after 3 self-inflicted failures; the 090 stalled and the producer was found by hand; filed 235

### The SWO dispatch took four attempts, and the first wrong call was mine

Attempts 1+2 failed on MY dangling `input_data.reviewed_brief` mapping; attempt 3
on the agent CONTRACT (`missing required fields: [input_data reviewed_brief]`);
attempt 4 mirrors the working pageflow shape (outer `load_site`,
`reviewed_brief: site_record.content_data`) and ran clean. I attributed attempt 1
to the spawn→call handshake race WITHOUT reading `orchestration_states.error` —
logged in `WRONG_CALLS.md` 2026-08-09. The error column named the cause all
along.

### `[MEASURED]` SWO arm (correlation `aab47560-…`, 13:50–13:56)

Both assets re-made from the run's OWN store outputs: `logo_stored.s3_uri` →
`af67d2c4-…png`, `hero_stored.s3_uri` → `e623ad6e-…png` (fresh objects, not run
A's `bd7308a5`/`fa752a71`); commits `b56599fe0` "Deploy **logo** image"
(400×400 PNG, sha `e38781c2…`, 163,792 B) and `d47cf0315` "Deploy **hero**
image" (sha `be35ba8d…`, 145,376 B) — both byte-different from run A's
artefacts, per-purpose properties correct, correct subjects, deploys stamped the
same asset rows the stores wrote. **Caveat, stated:** the pod-level
Strategy-0/object-key log lines were NOT captured on this arm — my capture loop
was pinned to attempt 1's dead pod name, and these pods keep ~11s of logs. The
arm is corroboration on artefact + row evidence; the log-level proof lives on
the pageflow arm.

### The 090 on the producer: UNVERIFIABLE again, and the answer was one query away

Run `fd7ef7a9-…`: scope-not-narrowing. It read the EMPTY
`task_workflow`/`orchestrator_workflow` columns instead of `default_config`, and
could not fetch the `ImagePurposes` var declaration — bug 223's var-blindness,
third consumer. Its "still needed" list was satisfiable first-hand in two
queries, and the result **refuted my own filed hypothesis in its mechanism
detail**: purpose does NOT fall to the spec Default — it resolves fine, to a
static `"hero"` on `image-build-handler.store_imagery_brand_asset`, a branch
whose own description says it handles *"logo or canonical index hero"*. One
static for two purposes; the item's `spec.purpose` says `logo` and is unread.
Work-item ↔ commit date join pins lendzy (08-02) and webdesign.uk (08-04).
Filed as **`bugs_open/235`** (with the `[INFERRED]` marker on the older nine,
and the `input-data.asset-key.jpg` literal-path shard); LANDMINES entry added +
synced; 231 carries the cross-reference.

### Also observed, not chased

- Run A eventually FAILED at `apply_site_design`: `CHILD_ORCHESTRATION_FAILED`,
  3 retries, no design-agent child row ever appeared. Asset proof unaffected.
- Run D's `fix_items_loop` spawned `image-build-handler` three times; all ended
  `complete_error` (items unknown — the loop moved on). On a sacrificial domain,
  so left alone; whoever owns 210's needs_logo file may care.
- The cookly.uk go-live (Cloudflare zone + Nominet NS) was requested by the
  owner mid-session and delegated to a sub-agent; its report goes in
  `domains_cloudflare_rollout/NOTES` (that lane's first real zone-create).

---

## 2026-08-10 (morning) — 235's fix is PROVEN LIVE, the site list was wrong, and the queue was dead for a reason outside this lane

### Cold start: all four checks pass

Ground unmoved (`git log b8509054a..HEAD` on the lane's paths: empty). All 7 lane
tests pass. Migration 360 still in force, re-read BY CONTENT: `store_imagery_brand_asset`
carries **no** static `purpose` and `purpose_field = input_data.spec.purpose`,
while `store_hero_asset`/`store_logo_asset` correctly keep their statics on all
three definitions that own them.

### 235 fix candidate 1 — BEHAVIOURALLY PROVEN on cookly.uk

Fired one `needs_imagery` item at cookly.uk: `brand_update:true, asset_key:"logo",
purpose:"logo"`, item `3c1c7c65-f8ae-4dc1-bd44-0a56fee1adb7`.

**The routing is the part that makes this disconfirmable**, and it is recorded in
the run's own `collected_data`:

```
check_imagery_brand_update = {"condition_met": true,
                              "next_step_override": "store_imagery_brand_asset"}
asset_stored              = {"purpose": "logo", "logo_url": "/assets/images/logo.png",
                             "asset_id": "5c351ebc-…", "stored": true}
```

Pre-360 that same input stored the logo with `purpose:"hero"`. Served artefact is
a **PNG** (`logo.png`, 400×218, sha `3fb6ad54…`, 189,044 B), replacing the
previous 400×400 `e38781c2…`; **no `logo.jpg` was created** (404, as before).

> **Two corrections to the handoff's stated acceptance bar.**
> 1. **"400×400 PNG" is wrong as a literal test.** Logo processing fits within a
>    400px box preserving aspect — this wordmark came out **400×218**, and
>    idea.uk's known-good logo is 400×218 too. The disconfirmable properties are
>    **PNG (not JPEG), ≤400px (not 900×900 or 1408×768), and `purpose='logo'` on
>    the stamped row**. Anyone asserting literal 400×400 will fail a correct run.
> 2. **No new asset row is created.** `store_asset` UPDATED the existing row
>    `5c351ebc` in place (its `storage_path` moved to today's date). So "assert
>    the NEW row" does not work either — assert the row's `purpose` and its
>    `storage_path` date.

Note cookly's *pre-existing* correct logo came from the LEGACY `needs_logo` item
type (`store_logo_asset`, static purpose, always correct) — a different step. So
this run is the **first time the brand branch has ever produced a correct logo**,
which is exactly why it was worth firing rather than reasoning about.

### The site list in the handoff is wrong in two places `[MEASURED]`

DB signature (`assets` where `asset_key='logo'`) cross-checked against what each
domain actually serves and what its homepage HTML references:

| site | row purpose | serves | homepage references | verdict |
|---|---|---|---|---|
| gamesdesign, vonc, dartsonline | hero | JPEG 900×900 | `logo.jpg` | affected |
| robot-hands, vetcomparison, fundamentallyai, oufe, lendzy | hero | JPEG 1408×768 | `logo.jpg` | affected |
| **relojistas.com** | hero | JPEG 646×275 | `logo.jpg` | **affected — MISSING from the handoff's list** |
| webdesign.co.uk | hero | JPEG 1408×768 | *(no logo ref on homepage)* | row bad, impact unclear |
| webdesign.uk | hero | 302 on both | — | row bad, unverifiable by HTTP |
| **idea.uk** | **logo** | **PNG 400×218** | **`logo.png`** | **NOT affected — wrongly listed** |
| cookly.uk | logo | PNG 400×218 | *(no logo ref)* | correct (proof site) |

So: **9 sites with confirmed visitor-visible damage**, plus 2 with a bad row and
unclear/unverifiable impact. `idea.uk` has only a stale unreferenced `logo.jpg`.

### The queue was dead, and it was NOT this lane's bug

The cookly item sat `triaged` for 20 minutes. `kafka-scheduler` was in
`CrashLoopBackOff`, **OOMKilled 132× in 13h**, taking the whole scheduled layer
to a ~14% duty cycle. Cause is `platform/kafka`'s shared transport leaving
`MetadataTopics` blank — kafka-go then refetches metadata for **all 25,042
topics** every ~3s, and 24,131 of those are orphaned `job.*` topics nothing
deletes. Full case: **`bugs_open/240`**; owner read-out in
`bugfix_240_kafka_metadata_storm/SUMMARY_2026-08-10_…`.

Confirmed causally, not just correlationally: after the orphan sweep took the
topic count down, the scheduler ran **11 minutes** (previously never past ~75 s)
at **58Mi** against the 121Mi peak that had been killing it, and the fleet's
short-interval tasks returned to ≤1.0 intervals overdue.

### Missteps this session — all mine, all caught before they did damage

- **Wrote a sample count into `bugs_open/240` before the probe finished.** Said
  "40 of 40 BUSY"; the probe had produced 18. Logged in `WRONG_CALLS.md`. The
  direction is the tell: I invented the *stronger* number.
- **Read a trend out of a truncating command.** `kafka-topics.sh --list` piped
  through `kubectl exec` returns a short list with **exit 0** (21,409 / 23,017 /
  5,809 on three reads; truth 24,131). I had already written a "correction" into
  240 based on the apparent decline. Retracted in place.
- **Filled the broker's 5 MB `/tmp`** with my own dumps, then diagnosed the
  resulting zero-byte reads as the tool truncating. It was ENOSPC I caused.
  Cleaned up; `--describe`'s behaviour is now recorded as *unknown*, not
  *truncating*, because every observation I have of it was taken against a full
  disk. Both traps are now a LANDMINE entry.
- **Nearly ran a 23k-topic deletion scoped by a truncated list.** It refused only
  because of a two-count-agreement guard I had added for an unrelated reason.
  That is luck, not process, so the guard is now explicit and the script verifies
  the transferred list line-for-line against an in-pod count.

### Also observed, not chased

- `webdesign.uk` holds a work item (`8793da9a`) inserted directly as
  `status='claimed'` with NULL `claimed_at`/`claimed_by`. The dispatch selector
  skips any site with a claimed item, so **webdesign.uk is permanently blocked
  from the build pipeline** regardless of the scheduler. It is one of the sites
  needing a logo re-drive. Not this lane's to clear — but whoever re-drives
  webdesign.uk must clear it first or the item will never dispatch.

### The re-render step took THREE attempts, and the first two both "succeeded"

This is the most reusable thing in the session, so it is written up in full as a
LANDMINE. The short version, in the order I got it wrong:

1. **Fired a `page_rerender` for gamesdesign's `index`.** It completed.
   `pages.deployed_at` advanced to 12:26:47. The served HTML was **byte-identical
   (49,326 B both sides)** and still referenced `/assets/images/logo.jpg`.
   *Why:* a `page_rerender` assembles the page from the **existing**
   `site_components`, and the logo URL lives in the `head` and `header` slots,
   untouched since 2026-08-09.
2. **Fired the site-level `needs_rerender`** (`rerender-pages`,
   `{"refresh_site_components": true}`). It completed, and this time
   `site_components.head`/`header` genuinely flipped to `logo.png` (12:29:24).
   **The served page was STILL byte-identical.** *Why:* that agent refreshes the
   components and then only **queues** 34 `page_rerender` items; they had not
   dispatched yet.
3. **Waited for the fanned-out items.** The homepage then served
   `/assets/images/logo.png`, with no `.jpg` reference left.

**Both failures presented as success**, with a completed work item and a moved
`deployed_at`. Nothing but diffing the served artefact would have caught either.
Had I skipped the canary and fired all nine sites at step 1, **251 page items
would have completed and changed nothing**, and every status would have agreed
it worked.

Note the byte sizes are equal before and after the real fix too — `logo.jpg` and
`logo.png` are the same length. So **size is not a usable change detector here**;
grep the reference.

`[MEASURED]` the mechanism, for whoever does this next:
- `needs_rerender` → `render_site_components` → `get_pages` → `create_rerender_items`
  (fan-out). One site item becomes N page items.
- `get_pages_for_rerender` is configured `include_statuses = ['deployed','active']`,
  so the 26 `needs_rebuild` + 17 `planned` pages across these sites are **not**
  touched. That backlog predates this work and is left exactly as found.

### Asset locks — the failure that was the system working

`robot-hands.com` refused: `assets.locked_at 2026-07-11`, `locked_by
user-b6-approval`, `lock_type permanent`. `store_asset` returned
`{stored:false, locked:true, reason:…}` and the deploy step then died on
`source path 'asset_stored.image_uri' not found for field 's3_uri'`.

`[MEASURED]` **no other affected site carries a lock** — checked across all 11
before concluding nothing owner-approved had been overwritten. `relojistas.com`
is also locked (owner, `bugs_open/131`) but had already been excluded as
`vm-sites`.

Two things worth carrying forward:
- **The error is three steps from the cause.** A locked asset surfaces as a
  missing-path input_mapping error, so the next person debugs a mapping bug that
  does not exist. `call_asset_deployer` should branch on `asset_stored.stored`.
- **A retry is the wrong instinct here.** The locked asset IS the defective one,
  but the lock protects approved *artwork*; regenerating discards it. The defect
  is encoding and size, not subject. Correct repair is to re-deploy the existing
  source object at `purpose='logo'` — a different operation, and an owner call.

## 2026-08-11 (morning) — migration 380: the dispatch mapping fix; the handoff's option 1 REFUTED before it was built

Cold-start checks all pass (7 tests green; scheduler 10Mi on v1.0.1284 —
**C3 is LIVE**, GOMEMLIMIT=192MiB visible in the deploy env, pod-verified).
Topic count 106 at 09:45Z, down from 1,236 at 16:33Z — `[UNVERIFIED]` what
swept; no sweep log exists, no commit touched the sweep script, so most likely
a manual run by another session or the owner around the morning roll.

**C4 has never fired.** The crontab entry is installed (RELOAD logged 18:05
BST 08-10) but the journal has NO entries at all in the 00:15–00:20 window —
the machine was asleep at 00:17 local. **User crontabs get no anacron
catch-up**, so on a laptop that sleeps overnight the 00:17 slot silently
misses every night; only the 12:17 slot is real. First actual firing due
12:17 BST today — check `~/kafka-sweep-240.log` after that.

### The handoff's preferred fix candidate could never have worked

`HANDOFF_2026-08-10b` option 1: deploy_asset gains
`"purpose_field": "input_data.spec.purpose"` via the spec's Deprecated bridge.
**REFUTED by the resolver before implementation** (`action_inputs.go`): the
bridge is Strategy 3, which skips any field already populated, and Defaults
populate first — so for a Default-carrying field the bridge is structurally
dead. `bugs_open/231`'s own mechanism section says exactly this, and
`TestPurposeFieldBridge_DeadForDefaultedField` pins it. The handoff proposed,
in preference order, a fix its own bug file's measurement notes refute — the
precise LANDMINES shape ("a bug file's FIX CANDIDATE can be refuted by that
same file's own MEASUREMENT NOTES"). Caught by reading the deciding arm first;
logged in WRONG_CALLS.

### What was actually done: migration 380, the dispatch-mapping fix

Only a Strategy-0 dotted path on `purpose` itself can beat the Default, and
that binding (`input_data.purpose`) is correct — it just has nothing to
resolve against on the `undeployed_asset` dispatch shape. Re-pointing it at
`input_data.spec.purpose` would break the image-build-handler path (maps
purpose top-level via `call_asset_deployer`). So the fix is at the dispatch:
**build-dispatch-loop `call_handler.input_mapping` gains
`"purpose?": "current_item.spec.purpose"`** — the exact idiom
site-work-orchestrator's `fix_items_loop` already carries
(`"purpose?": "current_fix_item.spec.purpose"`), which is the evidence this
was an omission, not a design choice.

Blast radius `[MEASURED]`, queries in RUNBOOK: exactly two live definitions
bind `input_data.purpose` — asset-deployer (fix target) and
image-build-handler's `check_logo_or_hero`, whose `purpose == 'logo'` arm is
half-dead today and ACTIVATES: needs_imagery brand-update logo items will now
route down the logo-generation branch (the condition's stated intent; 235
family). The 11 no-mode favicon/og_card items flip from latent hero-deploys
to clean 179-B refusals (the guard fires on the RESOLVED purpose). Items
without spec.purpose: `?` mapping skips silently, no-op. page-build-handler
binds no input_data.purpose — inert.

Execution: post-verify DO/RAISE induced first (raised "0 of 1"), scoped-dir
dry-run, applied + recorded 380 ~10:15Z, live row re-read by content
(`purpose?` present in the mapping). Council corr
`a46a4421-244f-4869-9c75-1b73a870371a` submitted BEFORE the apply (FORCE=1 —
single edit is a migration file; the scope filter only knows paths).
ROLLBACK sidecar alongside. Relojistas item `6084d849` reset
`complete`→`triaged` (the webdesign.uk-proven path) to re-trigger the deploy.

Also noted: relojistas carries a second `detected` undeployed_asset item
(`24c2fb3b`, purpose='icon', the phantom `input-data.asset-key.jpg` asset
`d3138254`) — left alone, it is 235's estate-audit quarry, not this repair's.

### 2026-08-11 ~10:10Z — council round 1: REVISE (gated by debug_historian), every objection answered by running its check

Verdict landed in ~10 minutes (queue was empty). Objections and what the checks
found:

- **debug_historian HIGH / guardian MED (unscoped WHERE vs the multi-active-row
  landmine):** the check is one query — build-dispatch-loop has EXACTLY ONE row
  (id `099b51e0`, version 1, active, not snapshot, not deleted). Not one of the
  four multi-row types. The post-verify's "1 of 1" could only ever pass at
  total=1, and the loaded row proved itself behaviourally (the deploy resolved
  'logo'). **Objection was right to ask, and the answer was already latent in
  the verify's own denominator.**
- **debug_historian MED (no backup / idempotency gate):** fair — took
  `agent_def_build_dispatch_loop_backup_20260811_post380` (1 row). Pre-state is
  losslessly reconstructable anyway: single additive key, ROLLBACK is one `#-`.
- **editquality MED (why does a top-level value survive Defaults-first, and
  where is the test):** the deciding distinction — Strategy 0 assigns
  UNCONDITIONALLY (`result.Values[field] = value`, no has-value skip);
  Strategies 1–4 skip populated fields. Wrote the test the seat asked for:
  `TestMigration380Shape_TopLevelPurposeBeatsTheDefault` (commit `be1cd6b9d`)
  pins BOTH arms — post-380 shape → 'logo', pre-380 control → 'hero' with
  asset_key/s3_uri still resolving from spec (the live incident's signature).
- **guardian LOW (a handler treating the new key as meaningful noise):** no
  mechanism exists — input_contracts contains only input_mapping.go (nothing
  rejects unknown keys), no input_data hashing/idempotency derivation in the
  dispatch path, work-item dedup keys on item_key.
- **architecture MED (record the CLASS, not just the instance):** the class is
  recorded — 231's general rule + the LANDMINES ActionInputSpec/Defaults entry
  + 231's open candidates 2/3 and the three-arm census. Pointed there.

Resubmitted round 2 on the same trail ~10:25Z (RESUBMIT_CORR honoured; run
orch `3044dbee`). The migration stays applied throughout — review here is
after the fact by design (owner ruling 2026-07-29).

### The fundamentallyai hot-link needed BOTH halves of the component patched — the NOTES' own rerender lesson, page-component face

The 08-10 session learned "a page_rerender ASSEMBLES from existing components;
site chrome needs refresh_site_components". Today's variant: for a
**page_component**, patching `content_data` and firing a page_rerender changes
NOTHING SERVED — the assembly reads the component's stored `rendered_html`,
which still carried the `.jpg`. The completed item + a served page still on
`.jpg` was the tell (verified by grepping the SERVED page, which is the only
check that catches this class). Repair: patch `content_data` (so any future
real re-render inherits the truth) AND `rendered_html` (so assembly serves
it), same surgical replace, backup row covers both halves' pre-state
(`page_components_bak_20260811_fundai_logolink`), then re-fire the
page_rerender. Also swept relojistas' own page/content components for
`/assets/images/logo.jpg`: **zero rows** — its reference lived only in chrome
(head + header), which `refresh_site_components` flipped correctly.

### 2026-08-11 ~10:35Z — round 2: APPROVED (2 advisory objections, both the same gap, both now closed by measurement)

`decided_by: approved with 2 advisory objection(s) — none high-severity`,
7 abstentions. Both mediums (guidelines, prior_art_librarian) named the same
thing: the round proved the input_mapping gate but never checked the SECOND
same-named gate, `agent_definitions.input_contract`. The check, run:

- `asset-deployer.input_contract` **declares `purpose` in `optional`**
  ("Optional purpose (default: hero) controls resize dimensions") — the gate
  admits the key by declaration.
- `ValidateInputContract` (input_mapping.go:245) **checks required-presence
  only** — it is not a filter and not a surplus-key allow-list; extra keys
  pass untouched for every handler. So image-build-handler's contract not
  listing top-level `purpose` is inert: the payload arrives whole and
  `check_logo_or_hero` reads it from the payload, not through the contract.
- The end-to-end behavioural proof (the relojistas deploy resolving 'logo'
  through the REAL claim → dispatch → call_agent → resolver chain) covers the
  claim-query RETURNING gate for this path: spec demonstrably surfaced.

debug_historian's low advisories (value- vs row-idempotency; no RETURNING on
the guarded UPDATE) noted for the next migration's authoring — harmless here.

Trailer discipline: commits before this point carry `Council-Submitted:`;
098 credits them at report time. Commits from here may carry
`Council-Reviewed: a46a4421` — the verdict has been READ (this section is
the reading).
