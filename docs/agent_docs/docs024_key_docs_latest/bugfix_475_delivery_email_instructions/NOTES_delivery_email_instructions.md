# NOTES — bugfix 475, the delivery email's instructions promise (mechanism half)

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-09-04, session open (`475` [206e1f])

**Task as given:** look at 475; if an active thread holds it, leave it; if not, resume it here. Use
the diagnosis loop if useful. Plan a fix with **Fable**, preferring a framework-wide solution to the
individual case. Check with the council. Find and talk to the other threads this affects. Confirm the
bug is still valid. Commit what should ride the next chassis build.

### Ownership — checked, and the check was not sufficient on its own

`scripts/who-owns.py 475` → **OWNED**, `site_delivery_and_editor` (153 commits/14d), plus
`bugfix_477_delivery_followup` (9 commits/14d). Their session `[2f98a0]` was live and had committed
**three minutes** before I looked.

⚠ **But `who-owns.py` reads COMMITS, so it cannot see a session mid-fix, and it cannot distinguish
"this lane owns the bug" from "this lane is busy with five other things".** The lane's own handoff
had 475 as **item 1 of a NEXT list** — queued, not claimed. Those are different states and the tool
reports them identically.

**So I asked them directly rather than inferring.** That is the whole finding: on this tree the only
current ownership signal is a live session answering. Their reply: *"Nobody is implementing any of the
mechanism half. I have only ever done copy on 475."* Clean split agreed — they keep the words, I take
the placeholder, `LinkConfig`, the page route, the zipper and the guard.

Their own handoff already carried the precedent, written after they were bitten by it: 477 was
*"CLAIMED by another session — do not compete"*, with the note **"filing a bug is not the same as
queueing it"**. I was the second session to arrive at that lane's door in a day and get a clean
answer because somebody had written the entry.

### Is the bug still valid? Yes — and the check needed a timestamp to be worth anything

`[MEASURED 2026-09-04 14:40Z]` live `delivery-email-sender` row still carries *"The ZIP comes with
instructions that walk you through putting it on free hosting."*

⚠ **The trap I walked past without noticing, and two other people caught for me.** Migration `776`
was applied to **the same `body_template` jsonb path** at **12:05:25Z the same day**. A reader
glancing at "that template was migrated today" would read the bug as fixed. It was not — 776 removed
a *different* false-promise clause (*"so we stop reminding you"*, the 477 lane's bug) from a
*different* paragraph.

My read happens to post-date the apply by 2h35m, so it stands — **but that was luck, not method.** I
did not know 776 existed when I ran the query. Both peer lanes volunteered it unprompted, and the 477
lane confirmed the clause independently with `body_template LIKE '%ZIP comes with instructions%'`.
**Two lanes, two instruments, same answer.** The lesson: *a field that was migrated today is the field
most likely to have been migrated for a different reason.*

### The framework defect — what the case was actually hiding

I came in expecting to add a placeholder. What is actually wrong is structural:

**Two customer letters share ONE closed placeholder vocabulary (`fillTemplate`) and TWO
hand-maintained mirror guard lists.** `send_followup_email_action.go:150-164` says so in its own
comment — *"If a shared mechanism is ever built, these two are its first two callers"* and *"If
fillTemplate's vocabulary grows, this list must grow with it."*

That is a contract held by a comment, which OWNER RULING 2026-08-02 §2 has already ruled is not a
control on a shared tree.

**And we demonstrated it live this afternoon.** Two sessions, both careful, both correct, about to
ship **two names for one concept** — `{{instructions_link}}` and `{{instructions_url}}`. What caught
it was one lane writing a comment and the other thinking to send a message. Neither is a mechanism.
Half a day's difference in timing and the estate would have had two placeholders for one page.

**This is why the fix is the shared table, not the placeholder.** Design in
`PLAN_2026-09-04_delivery_email_instructions.md` §2.1.

### MISSTEP 1 — I nearly built the table in the wrong shape, and no test of mine would have caught it

My first sketch was **"placeholder → link"**. The 477 lane corrected it before a line was written:
the seven tokens have **three different provenances**, and `{{days}}` is **not a link at all** — it is
`AdvertisedWindowDays`, a constant we deliberately set to **30** while actually serving **42**.

A link-shaped table either forces `{{days}}` in or quietly drops it, **and dropping it is the
dangerous outcome, because it is exactly the entry whose correct value is not the obvious one.** My
own tests would all have passed: they would have covered the six links I *had* modelled.

**The cheap check that would have caught it:** enumerate the vocabulary and ask of each entry *"where
does this value come from?"* before choosing the type's shape — rather than reading six of seven,
seeing a pattern, and naming the type after it. **A type name is a hypothesis about its members.**

### MISSTEP 2 — I stated a real risk at the wrong severity, in the same voice as measured facts

I wrote that if the pending build reached delivery, *"the same false statement goes to a second
**customer**"*, and argued from it that the bug file's *"no delivery is imminent"* premise had
expired, and that an interim copy fix was now justified.

**The next build is the owner's own trial run.** The voucher was minted for *him*. The recipient of
that sentence would have been the owner, not a stranger.

I had the voucher's existence from the copy lane and inferred the rest. The inference sat in a
paragraph of `[MEASURED]` facts wearing the same voice, which is the specific failure the estate's
marker rule exists to prevent — and I had used the markers correctly everywhere *else* in the same
document, which is what made the unmarked claim invisible.

**Outcome: the owner ruled "leave it" with the risk in front of him — no interim.** The correction is
recorded in the plan at §1 and §3 rather than edited away, and in `WRONG_CALLS.md`.

**And the ruling was better for the work than what I asked for:** there is now no competing change
queued against that jsonb path, so the phase-4 migration has a clear run. My stop-gap would have been
a third migration against one field in two days.

### MISSTEP 3 (not mine, but it cost the session's plan) — Fable hit a usage limit

The owner asked for the plan to be prepared using Fable. The subagent was launched with the full
brief and **terminated on HTTP 429, `claude-fable-5-1`**, returning nothing. The plan is authored by
this session (Opus) instead, and the plan says so at the top rather than letting the provenance
quietly become "the plan we asked Fable for". Worth re-running as an adversarial read of §2 when
credits allow.

### Decisions that came back from the owner via the copy lane

1. **The page is GENERIC, on `webdesign.uk`, built by the framework.** My proposal, ruled for. It
   spends no exception against the 2026-08-04 "every site goes through the framework" ruling, because
   `webdesign.uk` is itself a framework site (18 pages, `/guides/` deployed).
   ⚠ **The copy lane verified that at the SERVED BYTES with a control** — a real guide page 200s, an
   invented `/guides/` path 404s. **My version of the claim was a page count in a DB table.** Theirs
   is the evidence; a row count cannot distinguish a live route from a host that 200s everything.
2. **No interim wording. "Leave it."**
3. **v3 is stable to build against**, with `{{domain_paragraph}}` moved off the page by the ruling.
4. **`{{live_until_date}}` must NOT be wired** — three candidate dates disagree (presign 7 / email 30
   / tokens 42). Stating whichever is nearest to hand would be this bug's own root cause a third time.

### The landmine aimed squarely at this lane's next step

`deploy_image_asset` **resolves its source by PURPOSE, not by the `asset_id` passed to it** — so the
second same-purpose asset on a site silently deploys as the first. There are **ten** screenshots and
they will share a purpose. Supply `spec.s3_uri` explicitly from the asset's **`storage_path`**, never
its `url` column (`bugs_open/152`), and **verify with `sha256sum` that the ten deployed files
differ.** Ten identical images all reporting `success: true` is exactly what the bug looks like.
