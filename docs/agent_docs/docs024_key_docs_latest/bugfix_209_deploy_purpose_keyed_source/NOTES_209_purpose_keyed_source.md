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
