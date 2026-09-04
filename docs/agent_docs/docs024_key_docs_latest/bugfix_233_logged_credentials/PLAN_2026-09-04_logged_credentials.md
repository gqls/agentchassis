# PLAN — bugfix 233: what is left is a decision, not a design

The code fix landed 2026-08-09 and is live. This PLAN exists for the one open item, because it is a
decision with real trade-offs and it should not be made in a chat message that scrolls away.

## The decision

**Rotate the B2 application key pair and `CLIENTS_DB_PASSWORD`, or rule that it is not needed.**

## The case for rotating

`git log -S` dates the emission to `9260b86ed`, **2025-10-28**. Both credentials were written in
plaintext to pod logs for **over ten months**, readable by anything with `kubectl logs` on
`ai-persona-system`. The estate cannot enumerate who read them, and a credential that *might* have
been disclosed is treated as disclosed. Stopping the emission does not retract what was emitted.

## The cost, honestly

- **B2 pair** — touches `personae-default-secrets` **and** the GitHub-secrets copy the B2 CLI uses
  (per `LANDMINES.md`). Two places, and missing the second breaks deploys rather than the fleet.
- **`CLIENTS_DB_PASSWORD`** — touches every service's DB wiring. Materially bigger, and a partial
  rotation takes the estate down.

## Why it is safe to do NOW and was not before

The old blocker: `render-audit-adapter` ran a pre-fix image, so a rotation would have written the
**new** key into its log on the next restart — recreating the exposure the rotation was meant to end.
**Discharged**: it runs `v1.0.1360` and the leak string is absent from its binary. No image in the
fleet contains the emitting code, so **a rotation today cannot be re-leaked**.

⚠ That blocker was discharged on **2026-08-19** and this file's banner did not say so until
**2026-09-03**. Fifteen days of the right action looking premature.

## Sequencing, if the answer is yes

1. **B2 first** — smaller blast radius, and both copies must move together (secret + GitHub copy).
2. Confirm at the artefact that storage still works before touching the DB (a deploy and a page
   render exercise it).
3. **`CLIENTS_DB_PASSWORD` second**, as its own change, with every service's wiring updated in one
   move — a partial rotation is an outage.
4. Re-run the §5 binary probe afterwards. It must still read ABSENT: the point is that the *new*
   credential is never emitted.

## What this lane will NOT do

Rotate anything, or read any key value (owner ruling 2026-08-23 — probe from the pod, never read a
key into a session). Both are the owner's.

## Closing condition

The owner confirms the rotation, or rules it unnecessary. **Then** `bugs_open/233` moves to
`bugs_closed/`. Not before, and not on the strength of the code fix — that half has been demonstrably
live since 2026-08-09 and is not what has been holding the file open.
