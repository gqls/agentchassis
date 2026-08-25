# PLAN — noted.co.uk account deletion (owner ask 2026-08-25: "plan the account deletion")

A plan, not yet an implementation — build starts on the owner's word. Two
decisions are his (§5); everything else is designed to be built in one engine
change + one editor change.

## 1. What is promised, and what exists

The live privacy page says: *"If you would rather none of it were here, you can
close your account, and what is in it goes with it."* No mechanism exists — the
engine has no deletion endpoint, the editor no control. Every smoke run also
leaves a throwaway account, so the gap grows on its own. Since 2026-08-25,
accounts can hold real photos/video/audio in B2, which raises what "goes with
it" must actually delete.

## 2. The engine half

**`DELETE /api/account`**, authenticated, with the password typed again in the
body. The session cookie alone must not be enough for an irreversible action:
re-entering the password proves the person at the keyboard, not just the
browser (an open tab on a shared machine is the realistic attack, not a
forged request — SameSite=Lax already blocks cross-site DELETEs).

What it deletes, in order:

1. Collect every media row's `storage_key`/`b2_file_id` for the account.
2. Delete each B2 object ("already gone" counts as success — the client's
   existing `Delete` semantics — so a retry after a half-done attempt
   converges). Any hard failure → 502 "try again", nothing else deleted,
   account intact. **Objects before rows**: a row without an object is a 404;
   an object without a row is an invisible, paid orphan.
3. One transaction: delete the account row. The schema already cascades —
   `sessions`, `notes`, `media` all carry `ON DELETE CASCADE` — so rows
   cannot half-survive.
4. Answer with a plain goodbye JSON; clear the cookie.

The response comes ONLY after everything is gone (the same honesty contract as
everywhere else on this product: no optimistic "deleted").

Freed identity: `email_canonical` uniqueness dies with the row, so the same
email can register again later — a fresh, empty account, which is the right
meaning.

**Tests** (same discipline as the rest of `engine_test.go`, mutation-verified):
- deleting account A leaves account B's notes/media/sessions untouched
  (the isolation property, again — it is the whole product);
- wrong password → 401, account intact;
- after deletion: sessions dead (old cookie → 401), rows gone (direct SQL
  count = 0 for notes/media/sessions), B2 stub object count 0, and the email
  re-registers as an empty account;
- B2 delete failure mid-way → 502, account fully intact, retry completes.

## 3. The editor half

A small "Account" control in the signed-in bar → a panel that states plainly
what deletion means (*"deletes your notes, recordings, photos and videos from
your account — there is no undo"*), asks for the password, and requires the
confirm. Deletion is claimed ONLY on the 2xx (contract family); failure is
loud and changes nothing. Afterwards: the signed-out view with one plain
sentence. New `#nw-*` ids only — never renames. Harness case + mutations
(optimistic "deleted", panel deleting without password, silent failure).

## 4. Housekeeping the mechanism enables

- **Smoke self-cleanup**: `smoke_live_editor.py` ends by deleting its own
  throwaway account — the accumulation stops the day this ships.
- **The existing throwaways can be drained**: their credentials are
  reconstructible (`noted-smoke-<epoch>@example.invalid` /
  `smoke-<epoch>-0123456789` — the epoch is in the email), so a one-off script
  can walk them through the real endpoint. Better than hand-run SQL: it
  exercises the mechanism and leaves no bespoke deletion path.

## 5. The owner's two decisions

1. **Immediate hard delete vs a grace period.** RECOMMENDED: immediate. It is
   what the privacy page already promises ("goes with it"), the product has no
   billing entanglement to unwind, and a grace period means a half-alive
   account state plus email infrastructure noted does not have. If a grace
   period is ever wanted, it can be added in front of the same endpoint.
2. **The backup sentence.** Deleted data persists in the encrypted nightly
   dumps until they age out (14-day box retention; the offsite copies carry a
   30-day object lock — the privacy draft's own open question from 08-12).
   Media in B2 is deleted at once (its "backup" was the pg dump only while
   media lived in Postgres; B2-era media is genuinely gone on delete — worth
   saying, because it makes the honest sentence SIMPLER: text lingers in
   backups briefly, media does not). Proposed sentence for "Removing things":
   *"Closing your account removes everything at once; encrypted backups of the
   text age out within 30 days."* — his wording call, per the draft's note.

## 6. Size and order

Engine ~150 lines + tests, editor ~80 lines + case, one deploy each — the
engine half can ride the SAME deploy as the B2/stage-2 binary if approved
before the owner runs the box commands, otherwise the next one. No platform
(cluster) code is touched; not council scope.
