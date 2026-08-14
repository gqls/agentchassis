# Testing the editor's failure behaviour by hand

**For the owner. Written 2026-08-14.** About fifteen minutes, in an ordinary
browser on the open internet. Nothing here touches the server or anyone's data —
you will be breaking your own browser's connection, not the service.

**What you are testing, in one sentence.** When a save cannot reach the server,
the editor must say so loudly and keep your text exactly as you typed it — and it
must never say "Saved" unless the server really has saved.

**Why this one is done by hand.** The automated suite (24 checks, run before this
was handed to you) already proves the editor's behaviour against a simulated
outage, and each protection was deliberately broken once to prove the tests catch
it. What automation cannot do is show *you*, in a real browser with the real
service, that the failure is loud enough and the recovery obvious enough for a
person who is mid-thought. That judgement is yours.

**Where.** `https://app.noted.co.uk/tools/write/`

---

## Part 1 — a normal save (two minutes)

1. Open the page. You should see the sign-in form.
2. Make a throwaway account: any email-shaped address, any password of ten or
   more characters, and press **Create an account**. (A wrong password on an
   existing account should tell you plainly — you can try that too.)
3. Type a title and a line of text. The status beside the Save button should
   read **Unsaved changes** the moment you type.
4. Press **Save**. Watch the status: **Saving…**, then **Saved ✓**.
   - The important thing: **Saved ✓ appears only after a moment** — it is the
     server's answer, not the button's. If it ever appears instantly while your
     connection is broken (Part 2), that is the exact lie the old app told, and
     it is a failure.

## Part 2 — the outage (five minutes)

You will now cut the editor off from the server and watch it fail.

5. With the editor still open and some **new unsaved text** typed, open the
   browser's developer tools: press **F12** (or right-click the page → Inspect).
6. Find the **Network** tab across the top of the developer tools panel.
7. In that tab there is a dropdown that reads **No throttling**. Change it to
   **Offline**. Your browser is now pretending the internet is gone — for this
   tab only.
8. Press **Save**. Within a second or two you should see, all three:
   - a **bordered banner** stating the save did not go through and that your
     text is still here, with a **Try again** button;
   - the status reading **NOT saved**;
   - your text, untouched, exactly as you typed it.
9. Judge it as a person: if you were mid-thought, would you have noticed? Is it
   clear nothing was lost, and clear what to do next? That is the real test.
10. Now put the dropdown back to **No throttling** and press **Try again**.
    Status: **Saving…**, then **Saved ✓**. The same text, now really saved.

## Part 3 — closing the tab with unsaved text (two minutes)

11. Type something new and do **not** save it.
12. Close the tab. The browser should ask whether you really want to leave.
    Stay, save, and then close — it should let you go without asking.
    - The wording of that prompt is the browser's own and cannot be changed;
      what is ours is *whether* it asks, and it must ask only when there is
      unsaved text.

## Part 4 — proof the save was real (one minute)

13. Open a different browser (or a private window), sign in with the same
    throwaway account, and check the note is there. That is the whole product —
    the note left your browser and came back on another.

---

## What to do with the result

- **All four parts as described** → the degraded states are verified by hand;
  say so and this launch condition (handoff §5, blocker 4) is closed.
- **Anything off** — Saved appearing while offline, a silent failure, text
  vanishing, no prompt on close — say exactly which step, and stop there.
  Every one of those has a specific automated check that should have caught it,
  so a by-hand failure means the suite has a gap, which is itself a finding.

One honest limit to note: switching devtools to Offline severs the connection
*before* the request leaves. A connection that dies *mid-request* (after the
bytes left, before the answer returned) can in rare cases mean the server saved
but you were never told — the editor will say NOT saved for something that did
save. Pressing Try again re-sends the same note and, because a retried save
updates rather than duplicates, this resolves itself. You may see it as a note
that saved despite the banner; that is this, not a ghost.
