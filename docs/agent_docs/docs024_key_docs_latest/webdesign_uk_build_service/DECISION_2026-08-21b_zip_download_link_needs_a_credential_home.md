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
