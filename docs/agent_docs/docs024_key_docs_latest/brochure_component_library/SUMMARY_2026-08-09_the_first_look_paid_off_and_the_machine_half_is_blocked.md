# SUMMARY — the first look paid off, and the machine half is blocked (2026-08-09)

*Camera front of the brochure lane. Follows `SUMMARY_2026-08-03_someone_looks_and_the_switch_is_thrown.md`,
which is the last read-out on this front. The 08-08 summary covers the improvement-sweep
front and is parallel to this one, not superseded by it.*

## What we're trying to do

Catch the faults that only a human eye catches. The tests we run on a page are good at
structure and behaviour — is the button there, does it do something when you press it —
and blind to whether the page is actually *legible*. Every fault the owner has found by
opening a page himself has been on a page where every automated test passed. So the aim is
narrow and specific: photograph pages that pass, put the photographs somewhere a person
will see them, and find out whether looking is worth the trouble.

## Where we've come from

By the start of this month the camera existed and worked: a page that passed its tests had
its picture taken and filed. What we did not have was anyone looking. The pictures landed
as storage references inside technical notes — no page, no email, nothing that put an image
in front of a person. That was the honest position on 2 August: the machinery half done,
the human half not started.

Then somebody built the missing half — a single command that fetches the recent
photographs and lays them out as a page you can scroll. The first person to open it found
things immediately, which was the encouraging news, and one of those things was that the
camera itself had a fault.

## What we've done since

**Fixed the camera's timing.** It was photographing pages *after* the tests had finished
poking at them. A calculator that had just had its Clear button pressed was photographed
looking empty, as though it were broken. That is a false alarm waiting to be raised, and a
person looking at it hesitated — a machine would not have. The camera now takes the picture
of the page as a visitor first sees it, before anything touches it, and every photograph is
labelled with which state it shows and at what screen size. On the rare occasion it cannot
get the clean shot it says so rather than pretending.

**Put the contact sheet on a schedule.** It now rebuilds itself weekly rather than waiting
for someone to remember.

**Tried to add a machine eye, and hit a wall.** The owner's decision was that rather than
relying on a person to look, a model should look at every photograph and raise a repair
ticket when it sees something wrong. That was wired up — carefully, so that if the looking
step fails it cannot damage the test result it follows, which turned out to matter because
it has failed every time. It fails because the part of the system that does the looking
cannot reach the stored photographs. The obvious fix was to hand it the keys to the picture
store, and the owner has ruled that out: those keys should stay in one place rather than
being copied to every component that might want a picture. That ruling is right, and it
means the machine eye now waits on a design decision rather than on a bug. **It has never
successfully looked at anything.**

**And then the looking paid off.** The contact sheet had not been rebuilt in five days, so
a backlog had quietly built up. Rebuilding it and actually looking turned up a calculator
page on one of our sites that had passed every single one of its tests, and on which the
five headings above the form were invisible. Not faint — invisible. Tracing it back: the
site's "primary" colour had been set to a near-black that is almost exactly its own
background colour, and the component uses that colour for its headings. It is perfectly
fine as a button colour and catastrophic as a text colour, and nothing in the system
distinguishes those two uses.

Measuring that across every live site: **five of the fifteen we could measure have the same
problem**, and on one of them the primary colour is not merely close to the background but
*identical* to it. Fifty-three of our reusable components use that colour for text, so any
of them landing on any of those five sites is invisible. A further five sites could not be
measured at all, and are recorded as unknown rather than quietly counted as fine.

## Where we are now

The human looking loop works, runs weekly, and has now justified itself: it found a real,
measurable, multi-site fault that every automated test had passed. That was the open
question this whole effort existed to answer, and the answer is yes.

The machine looking loop is built, safely wired, and blind. It is blocked on an
architectural decision about where the keys to the picture store live — a decision that
belongs to the team that owns the critic, not to this lane.

The colour fault has been handed to the existing bug that already owns that territory,
with the sites named and the numbers attached. It was deliberately *not* filed as a new
problem: the class was already known and described; what was missing was which sites still
had it, and now that is written down.

## Where we're going

The immediate, cheap next step is to run the measuring tool across the three worst sites.
One component on one page produced seven failures; those sites will have considerably more,
and the tool names exactly which text and which colours each time.

Separately, most of our test runs still photograph nothing at all. The runs that check
individual components — as opposed to whole tool pages — never had the camera switched on,
so the camera currently covers under a quarter of them. That is one setting on the
instructions those runs are dispatched with, it costs nothing to produce, and the team that
dispatches them has been given the measurement rather than the argument.

The machine eye waits on the architecture call. And the lane's largest unbuilt piece is
unchanged and still waiting: teaching the planner to assign facts to sections up front, so
that sibling pages stop restating each other.
