# README — where we are (bugfix 440)

**2026-09-02.** New lane. The bug in one sentence: when a page-rebuild instruction carries a
label the system doesn't recognise, it quietly does the cheapest thing (re-ship the old page)
and reports success — and today it is impossible to make it refuse instead, because the same
label field is also where humans legitimately write free-text notes to each other (eleven such
notes written just today). The plan: give the routing label its own field, keep the notes field
free forever, and then a label nobody understands becomes refusable — checked at the database
door itself, so even hand-written migration scripts can't slip past. Big design piece (RFC_062)
for the behaviour change; a small inert foundation ships now. The 404 team built the current
warning half and their review round just finished — we've left them a note so nothing is done
behind their back.

---

**2026-09-03.** You made all five calls and we've acted on every one. The design questions are
closed and written into the design document itself, so nobody has to re-litigate them: a
rejected instruction goes to the human review queue rather than being thrown away or blowing up
the job; the database itself will refuse a bad label, which is the layer hand-written scripts
can't sneak past; the free-text notes field stays free forever; and the 404 team co-signs the
one migration that touches their work.

With the wait lifted, the producer half is built and submitted for review. One thing in it was
worth the care: the obvious implementation would have quietly changed how one existing case
behaves — a particular kind of rebuild instruction that deliberately does nothing today would
have started doing something once we flip the switch, and it wouldn't have shown up until weeks
later, looking like the flip was broken. We found it by reading the vocabulary's own notes, and
the shipped version keeps the new label locked in step with the old one, so the flip provably
changes nothing for the cases that already work.

Remaining: the authoring safeguards for hand-written migration scripts, then the flip itself —
which needs one technical confirmation first (how the gate treats a missing label versus an
empty one), because getting that backwards would refuse everything already in flight.
