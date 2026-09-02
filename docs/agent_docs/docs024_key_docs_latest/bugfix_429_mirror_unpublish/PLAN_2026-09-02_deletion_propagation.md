# PLAN — bugfix 429: the publish mirror cannot unpublish (deletion propagation for b2worker)

**Lane opened 2026-09-02.** Owns `bugs_open/429_HANDOFF_2026-09-02_the_publish_mirror_cannot_unpublish_a_retracted_page_persists_at_the_hosted_copy_forever.md`
(confirmed unowned by the filing lane's own handoff and by direct message with the
site_delivery_and_editor session, which filed it and holds the consumer-side close
criterion: a SERVED 404 at `https://boxingonline.ugg2.com/contact.html`).

## The design (decided 2026-09-02; full trail in the session plan + NOTES)

Fix candidate 1 from the bug file, hardened. `b2worker.Publish` gains the deletion
half: after the existing copy + destination-listing ETag verify, destination keys
absent from the source key set are deleted, then verified GONE at a fresh listing
(the copy half's discipline, mirrored). The caller's served-bytes acceptance
becomes a PAIR when the sweep deleted anything probe-worthy: one swept key must
serve 404 (under-deletion) AND a kept `.html` must still serve 200
(over-deletion), both cache-busted, both before `published_hash` is written.

### Decisions and their reasons

1. **Sweep lives in the backend (`Publish`), drift-gated** — not a separate
   retraction hook (bug file's candidate 2). A retraction-aware hook leaves
   orphans from any OTHER cause unswept; convergence cleans them all. Rejected
   alternative: sweeping on every no-drift tick (out-of-band mirror writes are
   outside the seam's threat model, and the no-drift path stays free).
2. **`ObjectStore` widened with `Delete`** (compile-time), not an optional
   type-asserted interface — silent degradation on a store without Delete is the
   mirror of this very bug. `*storage.S3Client` already has Delete (s3.go:142).
   Implementers enumerated 2026-09-02: S3Client, `fakeStore`, `pubFakeStore`
   (`zipFakeStore` embeds the latter).
3. **Empty-source refusal in the backend** — `len(req.Files)==0` errors, naming
   bugs 304/429. A sweep against an empty source is a delete-all keyed on an
   absence; that verb needs its own decision (bug 304's), not a side effect.
   `publish_site` keeps its empty-tree skip but the reason now names the standing
   hosted copy when `published_hash != ''`.
4. **Bulk floor**: >20 orphans AND >50% of the destination refuses without
   `AllowBulkUnpublish` (opt-in field, default OFF, the 2026-08-02 §2 shape; zero
   live consumers — the scheduled reconciler dispatch cannot pass it, which is
   deliberate). Fail-danger analysis: a silently truncated source listing needs
   an SDK-level fault (`ListObjectsV2Paginator` runs to exhaustion and propagates
   mid-page errors, s3.go:170) — and even then the wrongly-swept copies are
   restored by the next full-listing tick's drift (source authoritative and
   untouched): bounded staleness, never data loss. Missed deletes are fail-safe.
5. **Rollout is the `th1:`→`th2:` TreeHash prefix bump** — pre-fix orphans are
   invisible to drift (boxingonline's `published_hash` already reflects the
   post-retraction tree), and the prefix exists precisely so a change
   "republishes once, explicably". Old binary: no drift. New binary: both
   opted-in sites (`boxingonline.com`, `noted.co.uk` — measured 2026-09-02) drift
   exactly once and converge via the normal hourly one-site-per-tick rotation.
   NO forcing (the reconciler-force landmine), no migration, no ordering hazard.
   Zip deliverable keys carry only the last-12-hex sha tail, so zip naming is
   unaffected; the delivery lane confirmed zero cost for the live order and
   prefers th2 over a HOLD migration.
6. **`robots.txt` excluded from the 404 probe** (the edge rewrites it to a 200 —
   a robots-only sweep would wedge the site in a permanent quiet retry loop);
   probe the literal key, not its directory form (DGH-012 rewrite); `==404`
   verified correct against `scripts/cloudflare/worker.js` (missing object is a
   404 in both branches, never a redirect/catch-all). Status-code probes are
   immune to the beacon-injection landmine (that one bites byte-compares only).
7. **Result carries `Deleted`/`DeletedKeys`**; the action result caps the
   recorded key list at 20 + count (results land in `collected_data`).

### Reviews

- **site_delivery_and_editor lane** (2026-09-02, two-way): no veto; contributed
  the bulk floor and the acceptance pair; th2 preferred. Two pings owed: roll
  landed; contact.html 404 verified (then they strike handoff §1.5).
- **Adversarial fork review** (2026-09-02): 12 findings, all folded in — incl.
  `treehash_test.go` pinning `th1:` (edit in same commit or shared HEAD goes
  red), and the `check.py` optional-key literal `"publish_site": 3` → 4 with the
  parity test + post-commit overlay re-apply.
- **Council**: submission alongside the commit (`Council-Submitted:` trailer),
  7 code edits. Guarantee change stated plainly: the mirror becomes destructive
  at the destination (converges instead of only copying).

## Out of scope, stated

- Whole-site unpublish (empty source ⇒ delete the mirror) — bug 304's decision;
  `AllowBulkUnpublish` is the hook its fix can use.
- Mirror refresh latency (retraction → up to N hours until the site's rotation
  slot). 429 fixes "forever"; "minutes" would need a retraction-triggered
  dispatch — worth its own consideration, not smuggled in here.
- cfpages backend (unarmed; refuses loudly; does not inherit the empty-set
  refusal — per-backend guard, noted in the code comment).
