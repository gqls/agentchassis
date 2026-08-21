# The ZIP download link cannot be built as specified, and the reason is a standing owner directive

**Status:** BLOCKED on an owner decision. Not a technical gap, not a missing library, and
not something a session should route around. `/c/<token>` (the other half of the same
build) is DONE and unblocked; this is only about `/d/<token>`.

**Raised 2026-08-21** while building Phase 4's HTTP surface.

---

## 1. What was specified, and what stops it

The plan says `/d/<token>` **"mints a clamped presign and redirects"**: the customer holds
a token of ours for six weeks, and each click mints a fresh short-lived download URL
server-side. That design is right, and it is the only thing that delivers the owner's
"longest link we have" ruling — a presigned URL is capped at seven days by the signing
protocol and every caller already sits exactly on that ceiling.

**To mint a presign, a process needs object-store credentials. No standing service in
this estate has any, and that is deliberate.**

`[MEASURED 2026-08-21]` `B2_APPLICATION_KEY_ID` is unset in the running pods of
`agent-chassis`, `auth-service`, `core-manager` and `admin-dashboard`. The only manifests
carrying B2 credentials are adapters, the `database-backup` CronJob, and spawned-pod
templates.

**And the mechanism, which matters more than the measurement.** They were REMOVED from the
standing agent-chassis on 2026-08-11 under `bugs_open/245`, whose first line records the
owner's own words: *"the agent chassis shouldn't carry b2 credentials."* The spawn path
now injects `secretKeyRef` references into spawned pod specs rather than laundering
credential *values* through the spawner's environment. So object-store access lives in
spawned pods, on purpose, as of ten days ago.

**Putting B2 credentials into `core-manager` would hand a standing service the exact thing
another standing service had just had taken away, at the request of the person who would
be asked to approve it.** That is an architecture decision with a real blast radius. It is
not mine, and it is not a build detail.

> ### ✅ RESOLVED 2026-08-21 by the owner, and the answer was already in the estate
>
> Owner: *"The storage credentials are in the framework, please search how to use them.
> Create an agent container that uses the s3 client and you can add that container type to
> the spawn action to contain those credentials."*
>
> **He is right, and the mechanism is exactly `isStorageEnabledAgent`**
> (`platform/orchestration/actions/spawn_actions.go:3049`). Membership in that list is the
> sanctioned per-type grant: a spawned pod of a listed type gets `AWS_ACCESS_KEY_ID`,
> `AWS_SECRET_ACCESS_KEY`, `B2_APPLICATION_KEY_ID` and `B2_APPLICATION_KEY` by
> **`secretKeyRef` against `personae-storage-secrets`**, plus `S3_ENDPOINT` / `S3_REGION` /
> `IMAGE_BUCKET` from the `storage-config` ConfigMap. That is `bugs_open/245`'s fix: the
> credentials never pass through the spawner's own environment, so the standing chassis
> holds nothing and a missing key fails LOUD (a visible `CreateContainerConfigError`)
> rather than silently at first use. §1 above is therefore right that no STANDING service
> may hold the keys, and wrong to have read that as "there is no home for this work".
>
> **AND NO NEW CONTAINER TYPE IS NEEDED, because one already exists.** `[MEASURED
> 2026-08-21]` `zip-deliverer` is a live, active agent (`category: executor`, steps `zip`
> and `complete`) and is **already on the `isStorageEnabledAgent` list** — added for Phase
> 3 with the note *"zip_deliverable lists, reads and writes the portfolio-sites bucket ...
> same grant rationale as site-publisher"*. Its `zip` step already mints a presigned URL
> and returns it as `presigned_url` alongside `zip_key` and `expiry_minutes`.
>
> So **options A and B below are both unnecessary**: nothing new holds credentials, the
> spawn action is not edited, and `isStorageEnabledAgent` is not touched. The estate
> already had the container the owner described.
>
> **The remaining question is not WHERE but WHEN**, and it is a latency fact rather than a
> preference. A spawn is a pod start; a browser following a download link cannot wait for
> one, and dispatch is dropped entirely within ~300s of a chassis restart. So minting
> per-click is out, and the shape is **pre-mint and refresh**:
>
> 1. `zip-deliverer` already produces a presigned URL when it cuts the ZIP. Store it
>    against the `zip_download` token at mint time.
> 2. `/d/<token>` becomes **pure DB → 302**. No credentials in core-manager, which is what
>    §1 said was impossible and is now simply unnecessary.
> 3. A scheduled `zip-deliverer` run refreshes the stored URL before the 7-day signing
>    ceiling, giving the six-week window in ~7 hops.
> 4. **The silent-failure objection to option D still stands and must be answered in the
>    build, not in prose:** if the refresher stops, every link dies a week later and nobody
>    learns until a customer says so. `/d/` must compare the stored expiry against now and,
>    when stale, render an honest "your link is being refreshed" page AND file a work item.
>    That converts a silent death into a queue row somebody sees.
>
> **What this still does NOT answer:** whether core-manager may be publicly reachable at
> all. That is the guardian's gating objection on council `99b5af22` and it is a separate
> owner decision (§5 below).

## 2. The options, costed

| | What | Cost | The catch |
|---|---|---|---|
| **A** | Give `core-manager` B2 credentials | ~10 lines of kustomize | **Reverses the `bugs_open/245` directive.** Also widens the blast radius of the one service that already holds every site's data |
| **B** | A small dedicated `delivery-edge` service holding only a **read-only, deliverables-prefix-only** B2 key | A new deployment, a new key, a new thing to roll | Honest and narrow, and the key can be scoped so it cannot do damage. But it is a new standing service for one endpoint |
| **C** | Store the presigned URL at ZIP time and have `/d/` just redirect to it | Almost nothing | **Dies after 7 days.** The customer window is six weeks, so this fails the ruling it exists to satisfy, silently, five weeks in |
| **D** | A CronJob (which may hold credentials, as `database-backup` does) refreshes every live token's stored URL every ~6 days | Moderate | A moving part that fails SILENTLY: if the refresher stops, every link dies a week later and nobody knows until a customer says so |
| **E** | Serve deliverables from the public CDN path at an unguessable key | Small | No expiry and no revocation, so it contradicts the six-week ruling from the other side |

**Recommendation: B.** It is the only option that both respects the 245 directive and
delivers the six-week window, and the narrow scope is the point: a key that can read one
prefix and write nothing is a materially smaller thing to hold than what `core-manager`
would be given under A. **A is a reasonable owner answer too** — it is his directive to
qualify, and "core-manager may presign, nothing may write" is a coherent position. What a
session must not do is pick either one quietly.

**C, D and E are recorded so nobody re-proposes them.** Each looks cheap and each fails
the same ruling in a different direction, two of them silently.

## 3. What was built anyway, because it needed none of this

`/c/<token>` — the confirm-transfer click — is **pure database**. It is built, wired,
tested and mutation-proved (`internal/core-manager/handlers/delivery.go`). It is the half
that stops the weekly chase email, so Phase 4's email work is not blocked by any of the
above: the delivery email can carry the confirm link now and gain the download link when
§2 is answered.

## 4. A hazard in `/c/` that is worth the owner's attention, separately

`/c/<token>` is a GET that changes state, which is what an emailed confirmation link has
to be given the owner's ruling that **recording the click IS the state, with no form**.

**Mail scanners and link-preview bots fetch links inside email.** Such a fetch would stamp
a site as transfer-confirmed without the customer clicking anything, and, if the token
were minted single-use, would also SPEND it so the customer's own click then fails.

- **The lockout half is cheap to remove and should be:** mint confirm-transfer tokens
  **not** single-use. The stamp is already `COALESCE`d, so re-clicks cannot move the
  recorded date. This is a one-argument change at the minting site (not yet built).
- **The false-confirmation half cannot be fixed without a second click.** The standard
  remedy is a landing page whose button is a second link. That is arguably not "a form",
  so it may not conflict with the ruling, but it is the owner's call and it costs the
  customer a click.
- **What it actually risks:** we stop chasing a customer who never moved their site off
  our hosting, and we retract their live link on schedule believing they had moved. Small
  today (nothing retracts anything yet) and worth deciding before the chase email ships.

---

**Sources for §1:** `bugs_open/245_HANDOFF_2026-08-10_standing_chassis_carries_b2_credentials_it_should_not_hold.md`;
`deployments/kustomize/services/agent-chassis/overlays/production/uk_001/patch-deployment.yaml`
lines 102–139; pod env measured 2026-08-21.


---

## 5. The exposure question the council gated on, which is NOT the credential question

Council `99b5af22` came back **REVISE**, gated by the `guardian` seat at HIGH severity, and
its objection is correct:

> *"This is the FIRST publicly-reachable route on core-manager ... and it directly
> contradicts an invariant documented in the same file (sitefacts.go: 'NO PUBLIC EXPOSURE
> ... ClusterIP only, no Ingress'). The grounded_in cites an owner ruling that approves the
> CONFIRMATION MECHANISM (click-is-state, no form) but nothing that explicitly approves
> making core-manager reachable from outside the cluster at all. Those are two different
> decisions bundled into one submission."*

**It is right and I bundled them.** The owner approved how a confirmation should work; he
was never asked whether the service holding every site's data should become reachable from
the internet. The options are (a) approve that posture for named paths only, (b) put `/c/`
and `/d/` on a different service that is allowed to be public, or (c) keep both behind the
box and have the box hold the state. **Owner decision; nothing is deployed either way** —
the nginx block is written but not applied, so nothing is reachable today.

## 6. The other four objections, and what each is owed

| seat | severity | objection | owed |
|---|---|---|---|
| `editquality` | medium | Rate limit keyed on `$http_cf_connecting_ip`, a client-suppliable header, with no `real_ip_from`/trusted-proxy config shown | **Verify the trust boundary rather than argue it.** nginx listens on `127.0.0.1` only and ingress is a cloudflared tunnel dialling out, so a forged header would have to originate on the box. That is a real answer but it is currently only a comment in the file, and the objection asks for it to be enforced or proven |
| `editquality` + `guardian` | medium | The prefetch hazard is shipped unmitigated, deferring to an owner decision not yet made | **Fixable in the handler, and I had wrongly assumed it was not.** Refuse to mutate on requests carrying prefetch signals (`Sec-Purpose: prefetch`, `Purpose: prefetch`, `X-Purpose: preview`) and on HEAD. That is independent of how tokens are minted |
| `tooling_provenance` | medium | The `/d/` deferral lives only in this markdown file, not in `doc_notes`, so a future session has no DB-resident trail | Write the `doc_notes` entry |
| `reuse_agent` | low | No evidence I checked for an existing handler-deps testing convention in the package, or for another token-confirm-by-GET handler | Run both searches and say what they found |
| `guardian` | low | The nginx block is written but not deployed, and nothing stops someone applying it later without re-review | State the gate in the file itself |
| `architecture` | low (approve) | Watch the accumulation: a second and third publicly-proxied prefix should trigger a boundary review | Recorded here |
