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
