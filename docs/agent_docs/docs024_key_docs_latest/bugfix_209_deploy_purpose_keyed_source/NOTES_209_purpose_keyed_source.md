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
