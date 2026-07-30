# RESUME HERE — consolidation programme (supersedes the 07-29 door)

**Written 2026-07-30 ~13:30Z.** Anchor: `features_open/024`.
Sibling cold-start: `robot_hands_gripper_dossier/HANDOFF_RESUME_gripper_dossier.md`.
Read-out for the owner: `consolidation/SUMMARY_2026-07-30_the_last_mile_belongs_to_someone_else.md`.

**Read the 07-29 handoff for the evidence trail, not for the next action** — its §4
is corrected twice inline and its item 1 is done. Everything you need to *act* is
below.

---

## 1. State in one line

**Everything on this lane that is ours to do is done, committed and
council-approved. What remains is not code — it is one other thread's agreement,
and the blocker is DELIVERY, not persuasion:** the patch has been sitting in their
directory since 07-29 13:34 and their own cold-start doc still tells them *"nothing
owed yet"*, a line written five hours BEFORE it arrived.

Do not start by writing more code. Start at §4.

## 2. What is committed (mine, this lane, verified — not inherited)

| commit | 07-29 | what |
|---|---|---|
| `1f780faba` | 09:07 | `bugs_open/139` refuted in place — a visitor CANNOT choose the IP; the identity is a CONSTANT |
| `a0a22cfbf` | 10:24 | OWNER RULING recorded at the top of `features_open/024` + `SUMMARY_2026-07-29` |
| `31c684124` | 13:26 | **the code**: `httpguard.ClientIP` takes a required `FrontEnd`; PUB-002 + PUB-003 registered in the SAME commit |
| `171ff677c` | 13:34 | the adoption patch, written into the gauntlet lane's own directory |
| `c1e583267` | 13:37 | council APPROVED round 1 recorded; first `Council-Reviewed:` trailer |
| `df7f918b8` | 14:26 | the architecture seat's carried objection discharged — peer-gate boundary + real-socket tests |

**Re-verified on the current tree, 2026-07-30 13:2xZ** (do this again — the tree is
shared and moves hourly):

```
go test -count=1 ./platform/httpguard/ ./platform/mailer/
# ok  platform/httpguard  0.003s
# ok  platform/mailer     0.002s        <- -count=1 matters; without it both say "(cached)"
```

## 3. What is NOT live, and why that is not a deploy problem

**Both packages still have ZERO importers. Re-measured 2026-07-30:**

```
grep -rn 'agentchassis/platform/httpguard' --include=*.go . | grep -v '^./platform/httpguard/'   # NONE
grep -rn 'agentchassis/platform/mailer'    --include=*.go . | grep -v '^./platform/mailer/'      # NONE
```

So there is **nothing to roll and no pod to grep.** Library code with no importers
compiles into no binary; the usual "prove it at the artefact" discipline has no
artefact to point at, and that is not a gap in the verification — it is the finding.
`status: built, approved, called by NOTHING` is the honest state and it is what
PUB-002/PUB-003 say. The moment someone imports it, the normal rules resume: image
first, then grep the running pod for a string the change added.

**Do not read "zero importers" as "unfinished work of mine".** The owner ruled on
07-29 that adoption is in remit, and the reason it has not happened is in §4.

## 4. THE ONE OPEN ITEM — and it is a delivery failure, so treat it as one

Adoption steps 2 and 3 both land inside **`internal/tools-api`**, which belongs to
the **`gauntlet_dead_cta`** thread. `httpguard`'s only adoption target is that
service; `mailer`'s only queued consumer (the gripper dossier's public half) would
live there too. So both remaining steps are one conversation, not two.

**What I measured this morning, and it is the actionable bit:**

- The patch has been in **their** directory since 07-29 13:34 —
  `gauntlet_dead_cta/CONTRIB_2026-07-29_tools_api_client_identity_is_a_constant.md`
  (commit `171ff677c`).
- They committed **six** times after that (07-29 14:24 → 18:27). **None** touched
  `internal/tools-api/` or the CONTRIB. Last commit to that service at all:
  `a9a1b3556`, 07-29 09:34, their own 083 judge fix.
- No `internal/tools-api/clientip` package exists (`ls internal/tools-api/` →
  `api config db handlers httperr middleware store`).
- **The decisive one.** Their live handoff carries a "Consolidation ping" item
  ending *"Nothing owed yet"* — and `git log -S'Nothing owed yet'` dates that line
  to **`e304e3955`, 07-29 08:22**, i.e. **five hours before the CONTRIB landed**.
  It has survived four subsequent edits of that file untouched. Their next
  cold-start would therefore read "nothing owed" while a finished patch sat two
  files away.

**This is the LANDMINES/D10 gap again, on a different corpus: a file in the right
directory is AUTHORING, not DELIVERY.** Nothing tells a thread that a new file in
its lane applies to it. I filed the evidence correctly and it still did not arrive.

> **Worth knowing, and it landed the same morning I measured this.** `d02db14f4`
> (07-30 15:13, another lane) reports **D10(b) BUILT — `LANDMINES.md` now syncs into
> `doc_notes`, 14 entries / 27 footprinted rows, live.** So for *landmines* the
> delivery half now has a mechanism. **It does not cover CONTRIBs or handoffs**, which
> is where this lane's patch is stuck — but it is the precedent to argue from if
> anyone wants to close that gap properly rather than by remembering to nudge.
>
> > **SHARPENED a few minutes after writing the above — I checked, and part of the
> > finding HAS been delivered, by a third lane rather than by me.** The D10 lane's
> > own commit `11654d102` (07-30 13:28) read `bugs_open/139` and filed the substance
> > as a landmine of its own: *"A 'per-IP' limiter behind Cloudflare is probably one
> > global bucket — and `httpguard` reads as the fix"*, footprinted on
> > **`internal/tools-api/middleware/ratelimit.go`** and **`platform/httpguard`**.
> > With D10's session-start hook matching entries against a dirty working tree, a
> > gauntlet session that so much as opens that limiter file will now be told — which
> > is real delivery, and better than the note I appended to their handoff.
> >
> > **So state the residue precisely, because the original wording overclaimed.** What
> > got delivered is the **warning**. What still has no delivery path at all is the
> > **patch** — a landmine can tell you the ground is mined; it cannot hand you the
> > three edits that fix it. That distinction is the useful form of this lane's
> > finding, and it is the thing to say if anyone proposes extending D10 to CONTRIBs.

**What I did about it (2026-07-30, commit below):** appended a dated note under
their item 4 in `gauntlet_dead_cta/HANDOFF_2026-07-29_continue_here.md` — appended,
nothing of theirs edited — stating that the contact has arrived, that the finding is
about **their** service and does not depend on this programme at all, and what the
patch is. Their cold-start doc is the one place a fresh gauntlet thread is
guaranteed to read.

**If you pick this up, the next move is THEIR AGREEMENT, not more code.** Options,
in order of how much they cost them:

1. **Nothing.** The note is in their cold-start path; wait for a gauntlet thread to
   pick it up. Cheapest, and the patch does not rot — it is three small edits.
2. **Raise it with the owner** as a routing question, if it stays untouched. It is
   a live defect in a public endpoint's limiter, so "nobody got round to it" is a
   real cost, not a tidy-up.
3. **Do NOT apply it yourself.** `tools-api` is theirs, `bugs_open/083` (by slug:
   `gauntlet_engine_503_discards_the_error`) is open against it, and reaching in
   would be exactly what the CONTRIB convention exists to prevent.

**Do not re-argue the dead rationale.** `bugs_open/139`'s headline — a visitor can
choose their IP — is **REFUTED** and is not a reason to adopt `httpguard`. The real
defect (a constant identity) is, and the general case never depended on either:
three per-IP limiters with the weakest guarding the only public endpoint, four CORS
postures, a honeypot the next build planned to copy.

## 5. The httpguard change, in enough detail to defend it

`ClientIP(r *http.Request, front FrontEnd) string` — the `FrontEnd` is **required**.
Pre-declared: `Nginx()` (reproduces the old behaviour exactly, so `bugs_closed/090`'s
regression test is unchanged and still proven to fail when the defect is
reintroduced), `CloudflareTunnel()` (`CF-Connecting-IP`), `Direct()` (trusts
nothing).

**Why required rather than a default:** it previously hard-coded nginx's rules —
prefer `X-Real-IP`, else the rightmost `X-Forwarded-For` — justified in its own
docstring by nginx's `proxy_set_header` behaviour. **Measured 07-29: that
justification is false on the estate's other front-end.** Caddy does not set
`X-Real-IP` and forwards a client-supplied one verbatim; it *overwrites*
`X-Forwarded-For` with its own peer rather than appending. So on the island both old
rules resolve to user input or a constant, and adopting the old default into
`tools-api` would have keyed every visitor on `172.18.0.1` **while reading like a
fix**. A caller must not be able to inherit an assumption it never stated.

**Council `49392838-5ada-4c8e-baeb-94b01e5855b4`, round 1: APPROVED**, *"1 advisory
objection — none high-severity"*. 9 seats fired, 8 abstained on relevance, none
unreadable. The `architecture` seat settled the venue question the submission raised
against itself, returning `point_fix`: *"hardening a shared mechanism's own contract
before its first real consumer, which is the cheapest point in its life to do it,
not a shared-mechanism change smuggled in via a symptom fix."* **Useful precedent —
it is the inverse of `bugs_open/124`, and the distinguishing test is whether any
consumer's stated guarantee changes.** Both medium objections asked for confirmation
rather than argument and both were answered with evidence (register shipped in the
same commit; zero importers re-grepped after the verdict).

**Boundary fact anyone adopting this must know**, now pinned by test rather than
discovered in production: Go's `net.IP.IsPrivate` covers RFC1918 and RFC4193
**only**, so a proxy behind **CGNAT (`100.64.0.0/10`)** or on a link-local address
is **NOT** trusted and its headers are ignored. Fails in the safe direction (coarse
key, not spoofable) but it is a real constraint on where the package can deploy.

## 6. Debts, in the order they will bite

1. **OWED to the council, and it is the right ask.** `prior_art` approved the
   gripper round but asked that the contracts-gap claim — *no mechanism declares
   action-to-action `collected_data` fields; `input_contract` only fires at the
   `call_agent` boundary; `output_contract` has zero readers* — be **independently
   verified before anyone cites it as precedent.** I am still the only reader. Do
   not cite it as settled; verifying it is a contained, useful task.
2. **`[UNMEASURED]`** — whether bare `.(string)` assertions on LLM-parsed maps recur
   elsewhere unaudited (`bug_historian`'s surviving advisory). I showed its *premise*
   was wrong: `bugs_closed/076` is a truncation-tolerance mechanism, has **no**
   shared safe-extraction helper, and its headline "113 call sites" is a figure that
   file itself retracts to *"37 of 118"*. The underlying question is still real.
   **A count is needed before it is a claim — do not file it as a bug on a hunch.**
3. **Structurally unclosable locally, and stated in PUB-002 rather than left
   implied:** a real connection from a genuine *public* peer, to exercise the
   peer-gate reversion. Every address a dev machine can bind is loopback or RFC1918
   and lands on the trusted side by construction. Only a real direct-exposure
   deployment closes it. That is a deployment question, not a test one — **do not
   let a future thread "close" it with another unit test.**
4. **Owed cleanup in someone else's table:** three `manual-test` probe rows I left
   in tools-api's store, listed **by id** in `bugs_open/139`. Remove them when the
   gauntlet thread next touches that table, or ask them to.
5. **Two live fixture pages on robot-hands.com** still await the owner's read;
   cleanup owed once seen. **They were scored by the PRE-fix code** — do not use
   them as a reference for current behaviour.

## 7. The rest of the programme, unchanged

- **A2 / A3 are NOT done** — the owner ruling of 07-29 says adoption is the bar and
  both sit at zero importers. §4 is the whole of what remains.
- **Finish the pilot's public half** — `/api/v1/tools/gripper` **inside** tools-api.
  Do **not** write `cmd/gripper-intake/`; that would be the estate's fourth VM fork.
  Re-seed 208's `base_url` to `https://tools.apis.uk/api/v1/tools/gripper`.
- **A1 remains a WON'T-DO** — see the 07-28 handoff §5 and `features_open/024`.
  Reopen only if a second site wants a *physics* scorer.

## 8. Landmines this lane paid for (newest first)

**The first two are now filed in the fleet-wide `LANDMINES.md`** (2026-07-30, with
footprints, so D10(b)'s sync carries them into `doc_notes`) — they are repeated here
because this is the doc a successor to this lane reads first:

- **A CONTRIB file in the right directory is not delivery.** Their cold-start doc
  said "nothing owed" five hours before the patch arrived and stayed that way
  through four edits. If you file into another lane, put a pointer where their
  *next* thread reads, then say you did. Footprint: other lanes' `HANDOFF_*.md`.
- **`sed -i 's|a|b|'` against Go source containing `||` prints an error and then a
  green `ok`.** My first mutation proof of the peer-gate tests was worthless for
  exactly this reason — the file was never mutated and the suite passed honestly,
  and the available reading was "the guard is redundant". Mutate with a script that
  **asserts the anchor was found** before replacing, and diff the file.
- **"Absent" is not "never sent".** I nearly wrote a second wrong mechanism story
  (that `httpguard` would CREATE a spoof via `X-Real-IP`). Cloudflare strips it. The
  only thing that distinguished "stripped at the edge" from "the app ignores it" was
  sending an arbitrary `X-Zzz-Control` header **in the same request** as a control.
- **`decided_by` names the SEAT, not the objection.** Never inherit a council
  verdict through prose — print the severity table and quote the HIGH verbatim.
- **`doc_notes … ORDER BY created_at DESC LIMIT 1` is not YOUR verdict.** It handed
  me another submission's note (`0237eb64`, og-card). Filter on your correlation.
- **The `097` trigger wants a NESTED plan** — `plan{summary, edits[], grounded_in[],
  risks}`; a flat structure fails with `ERROR: .plan missing`. Build a resubmission
  from the persisted round-1 plan with `jq` (`SELECT
  collected_data->'input_data'->'plan'`) so the edits carry byte-exact and only the
  evidence changes — and put the jq program in a **file**, since an apostrophe
  breaks a single-quoted shell string.
- **A retag is not a rebuild**; verify by a string your change CREATED plus a
  positive AND a negative control, on the pod running *now*.
- **`error_step` persists under `config`, not at step level** — a query on the
  step-level twin returns NULL for every row and reads like "no error routing
  anywhere".
- **A duplication audit sees SHAPE, not USAGE.** Open both files and query live
  usage before calling anything a duplicate.

## 9. Commands worth keeping

```bash
# is either package actually used yet? (the whole state of this lane, in two lines)
grep -rn 'agentchassis/platform/httpguard' --include=*.go . | grep -v '^./platform/httpguard/'
grep -rn 'agentchassis/platform/mailer'    --include=*.go . | grep -v '^./platform/mailer/'

# tests, forced — "(cached)" is not a run
go test -count=1 ./platform/httpguard/ ./platform/mailer/

# has the gauntlet thread taken the patch?
git log --format='%h %ad %s' --date=format:'%m-%d %H:%M' -- internal/tools-api/ | head
ls internal/tools-api/            # a `clientip` dir appearing = they took it

# date a line in someone else's doc before believing it is current
git log -S'<the exact line>' --format='%h %ad %s' --date=format:'%m-%d %H:%M' -- <their file>

# the acceptance check for the adoption, if it happens (a presence check PASSES unfixed)
# kubectl … psql -c "SELECT count(DISTINCT client_ip_hash), count(*) FROM <their table>;"
```
