# SUMMARY — 2026-08-19 — the phantom-source birth gate (bugfix-309 lane, first and closing read-out)

**What we're trying to do.** Stop AI-generated page components from declaring data
sources the platform can never supply. The visible symptom was one page —
fundamentallyai.com's article index rendering six complete-looking cards with no
links — but the defect was the door it walked through: a generated component could
name any data source it liked, and nothing checked the name meant anything.

**Where we've come from.** The bug was filed on 2026-08-18 by another thread with
the symptom measured but the cause undiagnosed. This lane picked it up, traced it
end to end the same day (the component asked for a `site_specs.blog` data store
that has never existed on any site; the platform's rule for a missing value is
"quietly leave it out"; the template hides the link when the URL is missing), and
then censused the fleet: eleven components reference ten invented data stores, and
seven more reference database queries that were never registered. A second session
reached the same bug in parallel; we split it — they took the page repair, this
lane took the class.

**What we've done.** Built and shipped the door-lock: when a generated component is
stored, every declared source is now checked against what actually exists — the
live list of data-store names, the registered query vocabulary (refactored so the
dispatcher and the validator answer from one structure and cannot drift), and the
known source kinds. A made-up source is refused with a message naming the real
options. The review council approved it on round two, after demanding — fairly —
proof that the check is genuinely called (we deleted the call and watched the test
fail, then restored it) and a permanent record whenever the check must run
half-blind. Everything is committed, council-approved, and verified live in the
v1.0.1314 build at the binary with positive and negative controls.

**Where we are now.** The lane's own work is COMPLETE and this is its closing
read-out. The gate is live but not yet fired in anger — the first real generation
that declares a phantom source will be its live proof, and its silence until then
is expected, not suspicious. The original page is still unfixed, and correctly so:
the other lane's repair is applied and proven, but re-rendering the page is blocked
by a different guard doing its job — five of the eight articles have no summary
text, which turned out to be a fleet-wide gap (407 of 731 pages, filed as
bugs_open/320) with no existing mechanism to fill it.

**Where we're going.** Bug 309 itself now waits on one owner decision: which of
bugs_open/320 §8's options to take for filling in the missing page summaries. Once
that lands, the remaining two steps are mechanical and written out in the bug
file's §10 — re-dispatch the proven re-render, then verify the served page shows
eight linked cards. The residual risks this lane leaves behind are written where
the next person will look: hand-written database seeds bypass the gate (LANDMINES
entry; a table-level check is the named escalation if that reoffends), and the
store action now carries five separate guards that someone should one day
consolidate (register CLC-018).
