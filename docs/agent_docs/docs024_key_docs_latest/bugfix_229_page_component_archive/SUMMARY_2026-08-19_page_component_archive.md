# SUMMARY 2026-08-19 — page component archive (bug 229): the lane's first and closing read-out

**What we're trying to do.** Make it impossible for anything — the pipeline, an
admin screen, or someone typing at the database — to silently destroy a page
section's stored HTML. Chrome (headers/footers) got this protection first (bug
226); this lane extended it to ordinary page content, which is both hotter (every
save rewrites it by delete-and-replace) and where the platform's two recorded
losses of interactive tools actually happened.

**Where we've come from.** Filed 2026-08-08 at the council's own direction — a
reviewer objected that fixing chrome alone was "a mechanism-scoped fix, not a
pattern-scoped one". The owner ruled for extending the existing archive shape
rather than inventing a new one. Built and shipped 2026-08-09 in one day: a
fail-closed database trigger pair that archives every destruction (including raw
psql, which no Go code can see), fingerprint stamps in the four rebuild writers,
and review tickets when hand-made work is destroyed. Council approved after four
rounds — the code was clean from round two; the later rounds were making the
submission tell the truth about it, which is its own recorded lesson. Proven the
same evening by a production fire-drill: break a page on purpose, watch the
archive catch it, the alarm ring once, and an untouched rebuild stay silent.

**What we've done (this session, 2026-08-19).** Three things. First, re-verified
the whole closure ten days on, from scratch: fix commits in the running binary
(by ancestry against the pod's own build stamp), triggers enabled, zero wrongful
blocks, twenty real tickets raised — then moved the bug file to `bugs_closed/`
under the owner's restored fixed-AND-live bar. Second, the volume watch had
tripped — the archive grew 30MB→63MB in nine days, four times projection,
overwhelmingly copies of machine-regenerable content — so the retention design
the owner delegated to "once we have real numbers" was taken: a daily task
(migration 489) now discards only machine-reproducible copies older than 30
days, keeps everything irreplaceable for ever, writes a receipt every run
including zero-runs, and structurally cannot discard anything before ~09-08 (the
archive is younger than the window), giving a three-week free-abort runway.
Applied, ledger-recorded, self-probing, and confirmed scheduler-driven the same
hour. Third, verified the day's fresh chassis build (stamp `590ca3a20`, binary-
probed with controls) still carries everything.

**Where we are now.** LANE COMPLETE. The bug is closed and moved; the mechanism
is live, exercised, and bounded in cost. Residuals, all parked with named owners
and dates: (1) ~2026-09-09, one query confirms the first real prune and the
growth curve flattening — recorded in NOTES with the query; (2) register STY-056
(d) — whether the save-path snapshot should stop double-embedding artefacts now
the trigger owns recovery — an open design question, no pressure; (3) STY-056
(e) — surfacing losses from unlisted writers — deliberately parked until bug 083
gives detection a drain (the query is written; wiring it today would feed a
queue nothing empties, and the 20 existing tickets already sit undrained, which
is 083's cost, not this lane's).

**Where we're going.** Nothing further on this lane. The pattern's third
adopter, if one ever appears, triggers the shared-abstraction RFC by the
architecture seat's recorded condition; the retention watch hands over to the
daily receipts; and the next real work in this territory is bug 083's drain,
which is another lane's.
