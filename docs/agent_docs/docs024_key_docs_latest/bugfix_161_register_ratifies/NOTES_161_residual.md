# NOTES — bug 161 residual (`bugs_open/456`). Append-only, newest at the bottom.

## 2026-09-03 — resumed 161, found its inverse, and got three things wrong on the way

**Starting question:** is `bugs_open/161` still valid, and is a thread on it?

**Answer: it is `bugs_closed/`, not `bugs_open/`,** closed 2026-08-24, untouched since. No
active thread. `scripts/who-owns.py` named lanes that merely cite it; the lane that owns the
*mechanism* is `register_guards_code_phase_b` (owner of `bugs_open/288`), last active 08-26 and
dormant, with Phase 3b planned next — a different piece, so no competition.

**Re-verified 161 at the artefact rather than trusting the close-out.** `gd-trials` corrected
and carrying the first real `artifact_check`; `verified_at` bumped to 2026-09-02, so the daily
check is still running and passing; canary component still holds `Math.min(val, 10000)` and
zero `Math.random`; six repaired pages serving clean with a 404 control.

### Misstep 1 — I composed the page URLs by hand and got 404 on all seven INCLUDING the control

I curled `https://gamesdesign.co.uk/<pages.name>/`. All seven 404'd, and so did my invented-URL
control. Because my control shared the broken URL *form*, a uniform 404 was ambiguous between
"the pages are gone" and "my URL shape is wrong" — and the alarming reading fitted the bug I was
re-checking. `pages.url` shows three different forms on this one site
(`/guides/skinner-box/index.html`, `/guides/tool-spawn-rate-balancer-guide.html`,
`/games/auto-battler/index.html`). Read from the DB, all seven serve 200. **Fifth recorded
occurrence of this trap**; `scripts/probe-page-url.sh` exists for it. WRONG_CALLS.

### Three claims in the close-out are false, and the third is the interesting one

`page_name` addressing (it shipped as `subject_key`); "fail direction unit-proven only" (an
induced live drift was proven the same day, 288 §5b); and **"the attestation nudge cannot fire
before ~2027-01"** — wrong on its own code. `checkAttestationStaleness` treats an **undated**
fact as due immediately, by design and by its own comment. It fired on 2026-09-01 for
boxingonline.com. The claim reasoned from the facts that existed on 08-24 and treated the
180-day threshold as the only route to the queue. **A projection over a population still being
written to is a statement about today's rows, not about the mechanism.**

### The finding I was not looking for

I ran the real parser over all 27 live registers to *size* the 27-unverifiable-facts population.
**Two registers do not parse at all** — `finetuning.uk` (since 08-24) and `noted.co.uk` (since
08-25) — so both sites' whole claims layer, `banned_claims` included, was off. 10 bans inert.
One text-valued fact (`"MIT"`, `"30 days"`) was enough, because `ParseEvidenceBase` decoded
`facts` as one array and all three gate callers read a parse error as "no register".

The demand control is what makes this evidence: same register, same binary, **one fact
repaired**, and the sentence *"end-to-end encrypted and fully GDPR compliant"* goes from
unexamined to REFUSED by two bans quoting the owner's own reasons.

**It is a missing capability, not carelessness.** finetuning.uk's string-valued fact count went
**0 → 3 → 7 → 8** over three days as an author kept registering licences.

### Misstep 2 — my census processed 1 of 27 domains and printed "FAILED: 0"

`kubectl exec -i` inside a `while read` loop **eats the loop's stdin**. Exit 0, no error. It
read as a clean fleet. Caught only because I already knew two sites were broken — **if they had
sorted first the number would have looked plausible and I would have believed it.** Fix:
`< /dev/null`, and assert the loaded count. WRONG_CALLS.

### Misstep 3 — "gofmt ok" printed while the file was unformatted

`gofmt -l` reports by *printing a path* and exits 0 either way; my unconditional `echo "fmt ok"`
sat underneath saying the opposite. Same habit as misstep 2, an hour apart: a cheerful echo
after a quiet-means-pass command. WRONG_CALLS.

### The shared tree fought back, twice, and one was nearly serious

`claims.go` reported "modified on disk" **between two of my own edits** — the RFC_060 lane was
live in it. More seriously, `refresh_evidence_base_action.go` carried an **uncommitted
`livespec` call site** from another session at line 1779, with the helper itself untracked.
Committing that file by pathspec would have taken their half-written work and **broken HEAD for
the whole fleet**, because `make build-*` builds from committed HEAD and a build was imminent.

Split the work: committed the parse fix alone (`3f221f99f`), held the sweep half until their
helper landed (`1802359a6`), then committed it (`e5b41dc31`). `verify-head-builds.sh --with` is
what made this diagnosable — it builds committed HEAD plus only your named files, so a peer's
dirty tree cannot fail your run, and when the failure named **my** file with a symbol I never
wrote, that was the tell.

### Misstep 4 — my own tests caught two defects in my first draft

(1) An assertion that an artifact fact's message must not contain "human's word" — but my own
remedy text legitimately ends *"…if a human's word is the honest source"*, so the needle could
not tell the two wordings apart. (2) An sqlmock expectation assuming `writeWorkItem` is one
returning query; it runs several probe queries and an `ExecContext`. Both fixed by making the
tests assert the **decision** (a transaction is begun on a real run, none on a dry run) rather
than mirroring a helper's internals. Recorded in the test file rather than quietly corrected.

### Misstep 5 — the diagnosis loop returned UNVERIFIABLE, and it was my symptom

`status = UNVERIFIABLE`, `stopped: scope-not-narrowing`. **No verdict on either half.** I put
two mechanisms in one symptom (the decode voiding the register AND the residue arm plus early
return). They share a subject and nothing else. The tell was the word "Separately" in my own
symptom text. A full run spent, and the thing I actually wanted — an independent look for a
cause I had not considered — is exactly what I did not get. WRONG_CALLS.

### The council approved at round 1 and was still worth reading in full

`c2d1d570`, APPROVED, 3 advisories. **Three of the four things I then acted on came out of a
round that had already approved the change**, including the sharpest correction of the day:
the `debug_historian` seat caught that my evidence lines said **"LIVE, after this change"** over
what was an **offline repro against a copy of the live register**. The fix is committed and
inert until the roll. The reasoning was sound; the label was false. "Live data" and "live
system" are one word apart and a deploy apart. Also actioned: two searches I had asserted
instead of running (no tolerant-decode helper exists; `required_fields_missing` is not a
substitute because it has an auto-closing handler), a retirement condition for the new item
type, and a comment naming both ways the embed-and-shadow decode breaks silently.

**State at end of session:** both halves committed and council-approved, **inert until the next
chassis roll**. Post-roll verification is owed and specified in `456` §8 — nothing in this lane
is proven in production yet, and no document of mine says otherwise.
