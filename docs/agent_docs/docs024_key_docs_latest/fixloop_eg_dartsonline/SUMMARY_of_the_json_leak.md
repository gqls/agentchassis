● Closing summary — "json leak" thread

What the problem was

Nine article bodies had silently vanished from live SEO pages across five client sites, and five more were printing raw JSON — { "content": " — above the first heading for readers to see. Only 2 of 16 were healthy.

The key finding: the handoff's diagnosis was wrong

The handoff I was given attributed this to malformed JSON escaping — unescaped newlines in the LLM's output. I captured all 14 broken payloads and classified them, and that theory held for only 1 of 14.

The real cause was output-token truncation. The content writer was capped at max_tokens: 2000, and the call log showed it hitting exactly 2000 repeatedly. Long articles were simply being cut off mid-sentence,
producing JSON that couldn't be parsed.

This mattered because it changed the fix. The handoff said "the lost words are all recoverable" and proposed salvaging them from the database. They weren't recoverable — they'd never been fully written. Salvage
would have quietly shipped truncated, half-finished articles to five client sites while reporting success.

What we did

Fixed the root cause: raised max_tokens from 2000 to 8000. Confirmed at runtime — articles now generate at ~4000 tokens, comfortably complete.

Hardened the failure path so this class of bug can't hide again:
- The renderer now refuses to ship a section when required content is missing, instead of silently rendering an empty div
- The re-render path escalates for regeneration rather than overwriting good content with a blank
- Page assembly now names the section it drops, instead of dropping it silently

Recovered all the content. You chose regeneration over salvage. All 13 broken pages were rebuilt and verified live on the served pages — not just in the database.

Closed a follow-up gap. One page defeated three automated attempts. The cause: its article ends with a contact link, and the model writes href="mailto:..." with unescaped quotes — breaking the JSON even though the
content was complete. That would have recurred on any article ending in a contact link. Fixed by teaching the parser to distinguish a quote that ends a string from a quote that's just HTML, by looking at what
follows it.

Where it stands

Everything is closed, deployed in v1.0.1126, and verified live in the running pod — checked by inspecting the deployed binary directly rather than trusting git. All 17 article bodies are healthy.

Three things worth carrying forward

1. The handoff was confidently wrong on its central claim. It was well-researched and detailed, which is exactly what made it persuasive. Checking the raw evidence before acting is what caught it — and later, the
   parser's own error messages proved misleading too. Only inspecting the actual bytes gave the truth.
2. A stale doc was actively misleading. A parallel workstream's handoff claimed four pages were still broken and a test still failing — both long since resolved. I corrected it in place so nobody redoes finished
   work.
3. This thread was one of several sessions in the same code. Another session committed my in-progress work mid-task and briefly edited the same files. It reconciled cleanly, but it's why "verify against the pod,
   never git" earned its place in your practice.

One genuinely open item: the guards are verified by inspecting the deployed binary, not yet by watching a real image landing go through end-to-end. Worth watching the first one — expect a writer escalation, not a
blank.

