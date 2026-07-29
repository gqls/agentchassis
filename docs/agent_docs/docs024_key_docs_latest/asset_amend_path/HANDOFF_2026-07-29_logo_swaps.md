# HANDOFF — the asset-amend path, and the three logo swaps that use it (29 July 2026)

Repo-root `CLAUDE.md` binds. Workstream docs live in this directory (PLAN · RUNBOOK · NOTES ·
README_where_we_are). Parent finding: `bugs_open/131` **og-card slug** (ambiguous number — the
other 131 is the vonc gauntlet; resolve by slug, `git log` the file PATH).

## 0. Sixty-second orientation

The platform never had a path for a human to supply corrected image bytes — every image arrived
from a generator, and three sites' stored "logo" is junk (relojistas: two-up spec sheet;
gaswholesalers: nine-up contact sheet with garbled lettering; idea.uk: row broken AND the
picture itself an AI mangle). The owner ruled: build the path, through the chassis, no
credentials to the operator.

**Built 2026-07-29** (this directory's PLAN has the full design):

| piece | where | state |
|---|---|---|
| `ingest_staged_asset` action | `platform/orchestration/actions/ingest_staged_asset_action.go` + registry | written, 6 unit tests pass |
| staging table | `docs/agent_docs/sql_for_agents/265_asset_ingest_staging.sql` | written, NOT applied |
| asset-deployer 4th mode | `docs/agent_docs/sql_for_agents/266_asset_deployer_ingest_mode.sql` | written, NOT applied — **image first** |
| operator loader | `scripts/amend-asset.sh` (one command, one transaction, `--dry-run`) | written, dry-run verified |
| register | IMG-065 in docs026 imagery register + index | done |
| council | — | **NOT yet submitted** |
| build/roll | — | **NOT done** |

Bytes flow: file → base64 → psql stdin → `asset_ingest_staging` (BYTEA) → work item carrying
only the staging id → asset-deployer `mode=ingest_upload` → sha + image-decode validation → S3
at a NEW key → assets row amended in place (id stable, `alterations` history, `storage_path`
always set) → staging marked consumed. Locked assets refuse. Kafka never carries bytes (1 MiB
writer cap — that's the binding limit, not the broker's 5 MiB).

## 1. What remains, in order

1. **Council submission** (platform change; measured blast radius is in NOTES/PLAN — mode
   string, table name, item_type, action name all free fleet-wide; this is the FOURTH
   asset-deployer mode added by the same migration shape). Budget ~30 min queue latency; find
   the run by payload corr, not printed id.
2. **Commit narrowly** (the six files + docs), trailer if APPROVED.
3. **Build + roll**: bump `IMAGE_TAG` (makefile ~16), `make build-agent-chassis`, push, deploy.
   Pod-grep BOTH pods: `strings /app/agent-chassis | grep -c ingest_staged_asset` ≥1, plus a
   negative control. The same roll carries `e9e345464` (favicon aspect fix + locks honoured
   before commit) from the parallel session — **one roll satisfies both workstreams**.
4. **Apply 265** (inert table), then **266** (live mode) — 266 strictly after the pod-grep.
5. **Failing-branch check once**: stage with a corrupted sha → expect staging `failed`, assets
   row untouched (RUNBOOK § failure branch).
6. **Then the swaps** (§2).

No orchestration dispatch within ~300s of a chassis pod (re)start — the spawn is silently
dropped.

## 2. The logo swaps — OWNERSHIP SPLIT, agreed 2026-07-29

> **SUPERSEDES the first draft of this section.** The owner split the work mid-morning and the
> two sessions agreed it in
> `bugfix_131_og_card/COORDINATION_2026-07-29_who_does_what.md` — **read that file (and its
> reply thread) before touching any site.** relojistas-5 keeps relojistas end-to-end (S3 master
> uploaded in-cluster, row LOCKED, header deployed to `gqls/sites`) plus the deriving-code fix
> (`e9e345464`, council corr `bfd73f71` in flight) plus leopardess. **This session
> (relojistas-4) owns gaswholesalers + idea.uk.**

### gaswholesalers.com — THIS session; candidate logo EXISTS and is eyeballed

relojistas-5 generated a candidate (teal flame + "Gas Wholesalers" navy sans on white, spelling
exact, looked at by eye) —
`s3://personae-prod-uk001-images/images/system/20260729/58f69a8f-74cc-4ab3-99c0-b4923809140c.png`
(prompt + presigned URL in `bugfix_131_og_card/NOTES_og_card.md` entry (5)). Nothing on the
site touched. Two routes to install:

- **Amend path (preferred once rolled)**: download the object, run `amend-asset.sh` — this is
  also the mechanism's E2E verification target. The row lands path-style HTTPS +
  `storage_path` automatically.
- **Hand-update fallback (if the council drags)**: `UPDATE assets SET url='<path-style
  HTTPS>', storage_path='<key>' …` — **path-style HTTPS, never bare `s3://`** (landmine 1:
  `presignedURLToS3URI` reads the first key segment as the bucket and derives NoSuchKey).

Then: deploy the header (`logo.png` — matches the site's reference, which currently 404s so
the header shows alt text) → **hold `brand_head` until the favicon fix is in a rolled image**
(landmine 3) → derive → Read every PNG → lock the row.

### idea.uk — THIS session; candidate NOT yet owner-approved

Candidate: "idea.uk" high-contrast serif, ink on parchment, rust diamond dots —
`s3://personae-prod-uk001-images/images/system/20260729/ffea1049-a14b-4709-9827-7671dbece6a7.png`
(`gen-idea.png`). **The owner has not seen it — show it before installing anything.** Its row
is the malformed WEB-PATH case (`/assets/images/logo.jpg`, no `storage_path`); the install IS
the row repair — do not hand-repair the row first, there is nothing worth pointing it at.
Landmine 6: its header references `logo.jpg` while a logo-purpose deploy writes `logo.png` —
mind the extension you commit.

### relojistas.com + leopardessconsulting.co.uk — NOT this session's

relojistas: relojistas-5 end-to-end. leopardess: relojistas-5 has backfilled LOCKED
og_card/favicon rows (armed once `e9e345464` rolls; until then its malformed logo row is the
real protection — never "tidy" it). If bugs_open/142's detector work later flags leopardess's
backfilled rows as undeployed, they are DELIBERATE — do not delete.

### Roll coordination — a roll KILLS in-flight councils

relojistas-5's council round (corr `bfd73f71`) and this workstream's own submission are both
in the queue at various times; `kubectl apply -k`/deploy restarts chassis pods and kills any
council mid-seat (it already killed their round 1 on the 08:19Z v1.0.1198 roll). **Before any
push/deploy: check the coordination file and the orchestration_states queue for open council
runs, and say in the coordination file that you are rolling.**

## 3. Landmines

- **`status='triaged'` AND `pipeline='build'`** on every hand-queued item, or it is never
  dispatched (an item at `detected` sits forever with attempt_count=0).
- **Image before seed**: 266 names an action; applying it before the roll breaks the
  asset-deployer default path for everything (the conditional would error on dispatch).
  265 is safe any time.
- **The dedup key** `amend_asset:<asset_key>` refuses a second in-flight amend for the same
  key. Working as designed.
- **A locked row refuses the amend.** Clearing a lock is a deliberate separate step. Never
  script it.
- **Read every image this path touches.** The workstream exists because job-complete + HTTP 200
  + valid-PNG + provenance-rows-written all said success while the picture was a spec sheet.
- **`purpose` predicts geometry, not content** (logo→400×400 png, hero→1600×900 jpg). Don't
  infer what a picture IS from its purpose field — measured wrong on 07-28.
- **The header may reference `logo.jpg` while a logo-purpose deploy writes `logo.png`** —
  check the served page's `<img src>` before judging whether the deploy landed.
- **Shared tree**: `e9e345464` landed mid-plan from a parallel session. `git log` the file, not
  your memory of it, before editing anything in `derive_brand_head_assets_action.go`.
