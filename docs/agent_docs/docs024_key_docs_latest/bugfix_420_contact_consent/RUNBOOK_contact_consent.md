# RUNBOOK — 420 contact consent

## Census the READERS before moving a shared column
The delivery lane's condition for the split, and it found a writer the bug file had missed:
```bash
grep -rn "sites.email\|s\.email" --include=*.go platform/ internal/ pkg/ cmd/ | grep -v _test
```
Then read each hit **in context** and classify it: does this PUBLISH the value, or does it use it
to REACH THE CUSTOMER? That classification is the whole deliverable — a bare list does not
decide anything. ⚠ Record the count **with its date** ("4 writers / 14 readers as of
2026-08-31"): a census does not go wrong, it goes STALE, by addition, and reads as current for
ever.

## Is the delivery chain actually coupled to the column?
Do not assume from convention. The binding question is whether any CODE reads it:
```bash
grep -rn "sites.email\|FROM sites" --include=*.go platform/delivery/ platform/orchestration/actions/send_delivery_email_action.go
```
Empty ⇒ the coupling is a recipe, not a dependency, and the split is cheap.

## Prove the register leak, when no column ties the two together
`briefing.contact_email` and the seeded `evidence_base` fact have no join. Establish it by
**elimination** — rule out every other source:
```sql
-- did the customer's own brief contain it?
SELECT direction->>'objective' ILIKE '%<needle>%', length(direction->>'objective')
  FROM build_queue WHERE domain='<domain>';
-- did the identity spec carry a contact block?
SELECT data->'contact' FROM site_specs
 WHERE site_id='<id>' AND aspect='identity' AND is_current;
```
Both negative + the briefing agent reading specs with `aspect: "all"` ⇒ the register was the only
possible source.

## Verify a removal — the served page, never a DB sweep
⚠ **Four independent DB sweeps read clean during this incident while a fifth store still served
the address.** The set of places a value is baked into is NOT enumerable from the schema.
```bash
# enumerate from the DEPLOYED set, never a remembered list
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc \
 "SELECT p.url FROM pages p JOIN sites s ON s.id=p.site_id
   WHERE s.domain='<domain>' AND p.status='active' AND p.build_status='deployed';"
```
Then curl each and grep. **Three controls, or the zero is not a measurement:**
1. each body must CONTAIN the company name (proves you fetched real content);
2. an invented URL must 404 (a parked domain 200s every path);
3. the same grep against a pre-fix copy must FIND the needle (proves the needle fires).
And print `probed N of N deployed` — a skip must fail the check, not pass it silently.

⚠ Do not read a rerender's success as proof: a chrome-only change is invisible to the page
content hash, so a whole-site rerender no-ops and reports success.

## The trap when re-running a build after clearing contact details
```sql
SELECT direction->>'customer_email' FROM build_queue WHERE domain='<domain>';
```
Non-empty ⇒ **a re-seed will republish that address** (fill-only-if-empty inverts into a refill
once the column is legitimately empty). Clearing `sites.email` is not sufficient, and no sweep
will show this, because every sweep describes the state after the LAST seed.

## Mutation-test a guard rather than trusting a passing suite
```bash
cp <file> /tmp/f.bak
sed -i '<re-introduce the exact defect>' <file>
go test ./platform/orchestration/actions/ -run '<TestName>' -count=1   # MUST fail
cp /tmp/f.bak <file>
```
Run each mutation **alone**. A mutation that passes hit a guard in series — investigate it rather
than assuming the coverage doubled.
