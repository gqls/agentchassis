# Where we are — RFC 012 execution (findings that must survive an await)

Append-only, newest at the bottom. Plain prose.

## 2026-08-06 — your three calls, and what each one sets in motion

The background: when one of our agents both *works something out* and *asks an external
service to do something*, the reply from that service overwrites the agent's own workings
in the permanent record. Three separate teams have now hit this and each invented their
own workaround. You ruled: build the shared escape hatch properly (the database-backed
version that survived testing, not the in-memory one that didn't), run a census of
everything that reads those overwritten records (so the deeper fix — merging instead of
overwriting — becomes decidable later), and turn the one-off audit of how workflows route
their outputs into a standing check that runs inside the platform rather than a script
someone must remember.

All three are now in motion: the rulings are written into the RFC itself so they can't be
lost, and three research passes are running in parallel — one gathering everything needed
to build the shared helper, one finding the right vehicle for the standing check, and one
performing the census itself.

## 2026-08-06, evening — two of the three built; the third needs a fresh session

Two of your three calls are done and committed. The shared escape hatch exists: the one
writer that records an agent's findings now lives somewhere every part of the platform can
reach it, so the next team that needs findings to survive makes one call instead of
inventing a fourth private workaround — and it counts its own failures rather than
reporting success over a row that never landed.

The standing audit is built and, more to the point, **proven able to fail**: it catches the
outage-causing bug that both simpler versions of the same check miss entirely, and a
deliberate mutation makes the finding disappear, which is the only way to know it is really
reading the workflow graph. Run against the live fleet it found exactly the two known
problems out of 176 agents and nothing spurious. It now carries a short list of those two,
signed off, so it goes green until something NEW appears — a check that is permanently red
is a check people stop reading.

Two things remain, and one of them found a real problem worth telling you about: while
building the audit I had to re-derive the list of ways a workflow can route between steps,
and **the specification's own list was one short**. Trusting the prose would have left a
blind spot in the very check written to have none.

Still to do: converting the remaining eighteen copied database writes onto the new shared
one; wrapping the audit in the scheduled job that makes it "online" as you asked; and the
census itself, which is a full session's work on its own. I delegated two of those to
helpers and both ran out of quota mid-flight, so they are written up in detail for a fresh
session rather than half-done here.
