# Owed: an AMENDMENT to the existing §9 entry in `016b_debugging_guide_8_consolidated.md`

`016b` **already carries a §9 entry for this case** (~line 4310-4334, written by the filing
session: *"the system accepts your configuration politely and behaves as though you never wrote
it"*), and its recommended fix pattern — move the workflow to where the runtime actually reads it,
rather than making the ignored field work — is exactly what was done. So this is **not a new
entry**; it is two additions to that one, plus a pointer correction.

**Why it is sitting here instead of in the guide.** At the time of writing (2026-07-26) another
session had **98 uncommitted lines** in `016b`. A pathspec commit cannot exclude a same-file
passenger, so pasting this in would have swept their in-flight work under this bug's message.
Apply it once their edit lands, and delete this file.

---

## 1. Pointer correction

```
Case: `bugs_open/074`.        →        Case: `bugs_closed/074` (closed 2026-07-26).
```

## 2. Append to that entry — the half the original account did not have

> **The receiving end DID support the field; the sending end could not express it.** This entry
> reads as "inline workflows are unsupported". They are not: the chassis honours one at
> `body.config.workflow` (`processor.go` `selectWorkflow`, Priority 1, ahead of group discovery),
> and 58 live orchestrations arrive that way from `DispatchFeedSourcesAction`. What no scheduled
> task can do is *reach* that field — the scheduler **builds** `config` itself from the row's
> columns and nests the whole `input_data` column beneath it, so the author's workflow lands at
> `body.input_data.config.workflow`, one level below the only reader. **When a config key seems
> unsupported, find who constructs the envelope before concluding the feature is missing: the two
> ends can disagree about depth while each looks correct in isolation.** Reading the reader first
> is what turned this fix from *lift the field* into *refuse the shape* — the field is supported,
> just not from there, and `bugs_closed/054` had already ruled that the sender must not reach into
> the payload.

> **Refuse at authorship, not at use.** This entry recommends a WARN, and one went in. But the
> load-bearing fix is a `CHECK` constraint (`migration 217`): live the moment it is applied, no
> image roll, and it fails the INSERT that *creates* the trap rather than the run that inherits
> it. **A guard that fires where the mistake is made beats one that fires nightly where the
> mistake is felt** — and it makes the runtime warning unreachable in a healthy database, which is
> the right relationship between the two.

Category tags to add: `envelope-owner-buries-payload-config`, `constraint-at-authorship`.
