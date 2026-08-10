# README_where_we_are — bugfix_153_build_provenance

*(the owner's running plain-prose log — append only, newest at the bottom)*

## 2026-08-10

Picked up `bugs_open/153` — a real, unowned bug filed 2026-07-29 that nobody had gotten back
to. The short version: when we build a docker image for one of our services, nothing inside
the image or the running binary says which git commit it was built from. So if someone bumps
the version tag and pushes/deploys without actually rebuilding, we get a container running
old code labelled with a new version number — and nothing anywhere notices. The person who
filed this bug actually walked into that trap themselves and burned most of a day chasing
ghosts because of it.

Checked today (10 August): still true. Live production is on `agent-chassis` version
`v1.0.1279` and the binary running in those pods still has zero trace of what commit built it.

The fix, in plain terms: bake the git commit hash into the binary and the image at build time
(this is a standard practice — most serious software does this), so anyone can ask a running
container "what code are you actually running?" and get an exact, unambiguous answer instead
of trusting the version label, which — this bug proves — can lie.

I had a planning pass done by our "fable" model to turn that idea into a concrete list of file
edits (the plan is in this folder, `PLAN_2026-08-10_build_provenance.md`), scoped
conservatively: do the stamping + verification part now, fleet-wide (it's cheap and safe),
and leave the more invasive options — like refusing to push an image that doesn't match what
was built — for a later decision, since those change how the deploy process works for
everyone and deserve a proper sign-off rather than one session deciding unilaterally.

Next: build it, get it reviewed by the automated council (our advisory review process for
platform-code changes), commit it, and prove it works on one service (`agent-chassis`) end to
end before leaving the other 13 services' matching changes for their next normal rebuild.
