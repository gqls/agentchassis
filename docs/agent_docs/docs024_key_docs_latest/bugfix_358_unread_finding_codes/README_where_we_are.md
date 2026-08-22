# Where we are — the unread finding codes (bug 358)

Plain-prose log for the owner. Append-only, newest at the bottom.

## 2026-08-22 — picked up, and the bug is still real

The bug, in a sentence: lots of parts of the platform, when they notice something wrong that
they won't fix themselves, write a note about it into a database table — and for sixteen
different kinds of note, nothing ever reads the note, and a cleanup job deletes it after
30 days. We pay to detect things and then throw the detection away unread.

Today I checked the bug is still valid before planning anything, because things move fast
here. It is. Two small things changed since it was filed this morning, and both are actually
good news: the other team working on content loss shipped their checker, and it is the
first thing EVER to use the table's "resolved" tick-box (45,000 rows and nobody had ever
ticked it before today). That team's checker is the model citizen: it writes its findings
AND reads them back AND acts on them. The sixteen orphaned kinds of note are all still
orphaned — and a seventeenth was added by another session this very morning, which is the
whole point of the bug: nothing stops a new one arriving without a reader.

Next: research how the one similar guard-rail we already have was built (the "optional key
budget" checker), then write the plan and put it past the council.
