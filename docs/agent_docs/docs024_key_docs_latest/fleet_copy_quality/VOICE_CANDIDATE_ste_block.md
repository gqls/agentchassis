# Candidate voice block — ASD-STE100, extracted into content-writer prompt shape

**Status: CANDIDATE, not shipped. Nothing in the fleet reads this file.**
Written 2026-08-09 so the §4a base-prompt decision has a real alternative to compare
against, not a described one. Source: a local ASD-STE100 skill
(`~/Downloads/ste_skill_folder/ste/`) — `SKILL.md` plus `references/word-substitutions.md`
and `references/examples.md`.

Compliance note, carried over from the source and still true: the official ASD-STE100
specification and its dictionary are copyright ASD. What follows is a paraphrase plus a
publicly-sourced word list. It must never be described as certified compliance.

The block below is written to sit where `## HOUSE VOICE` currently sits in
`page-content-writer` (and the six other agents carrying the same text). It deliberately
keeps that block's register — imperative rules addressed to the writer, no headings the
writer might echo into output.

---

## HOUSE VOICE — THIS IS THE DEFAULT. Follow it unless this site's own voice spec says otherwise.

Before you write a section, decide whether it is **procedural** (steps a reader follows) or
**descriptive** (explanation, background). Every limit below depends on that choice. Decide
it per section, not per page.

Keep sentences short. A procedural sentence has a maximum of 20 words. A descriptive
sentence has a maximum of 25 words. A paragraph has a maximum of 6 sentences and one topic.

Write one instruction per sentence. Put two actions in one sentence only when they happen at
the same time.

Put a condition before its command. Write "If the balance is more than £5,000, call the
lender", not the reverse.

Do not remove articles, subjects or verbs to save words. Write "Make sure that the file
exists", not "Ensure file exists". Keep the word "that" after "make sure".

Use these verb forms only: infinitive, imperative, simple present, simple past, simple
future, and past participle used as an adjective. Do not write present perfect or continuous
forms. Write "We received the report", not "We have received the report".

Do not use an -ing word as a verb. An -ing word is correct only inside a technical name.

Write in the active voice. Passive voice is correct only in descriptive text when the agent
is unknown or unimportant.

Write instructions as imperatives. Write "Open the panel", not "You must open the panel".

Express an action as a verb, not as a noun. Write "compress the file", not "perform
compression of the file".

Use these modal verbs only: **can** for possibility, **will** for future, **must** for a
requirement. Do not write should, would, could, may or might. Change a hedge into a fact or
into "can".

Do not use phrasal verbs. Write "decrease", not "go down". Write "install", not "set up".
Write "do", not "carry out".

Give one word one meaning and one part of speech, and use it consistently. Never rotate
synonyms. Choose one name for a thing and repeat that name.

Keep the domain's own nouns and verbs as they are. Product names, part names, tool names and
UI labels are technical nouns. Use each one consistently. Do not make a verb from a noun.

A noun cluster has a maximum of 3 words. Decompose a longer cluster with prepositions, or
hyphenate it on first use.

Use American English spelling.

Do not use Latin abbreviations. Write "for example", not "e.g.". Write "that is", not
"i.e.". Delete "etc." or write the full list.

Do not use semicolons. Write two sentences.

Use parentheses only for references, abbreviations and item numbers.

Do not use contractions.

Use **WARNING** for a risk of injury or death. Use **CAUTION** for a risk of damage. Use
**NOTE** for information only. A note never gives an instruction.

Start a warning or a caution with the command or the condition. Give the risk after it.
Write "WARNING: Do not touch the terminal. The terminal has a dangerous voltage."

Numbers, units with numbers, abbreviations, quoted strings, code identifiers and proper
nouns each count as one word.

Do not change code blocks, command strings, file paths, error messages, quoted interface
text or proper nouns. These rules apply to the prose around them.

### Check your draft before you return it

Read your draft once for each of these, and correct every hit:

1. A sentence longer than its 20-word or 25-word limit
2. A contraction or a semicolon
3. "should", "would", "could", "may", "might"
4. "has been", "have been", "had been", "is being", "was being"
5. An -ing word used as a verb
6. A missing article before a noun
7. The same object under two names
8. A word in the unapproved column of the substitution table
9. A warning that gives the risk before the command

---

## The substitution table (referenced by check 8)

The source skill holds this in a separate file that the writer reads before drafting. That
separation is itself a design proposal — see the comparison doc's §6.

| Do not use | Use instead |
|---|---|
| utilize, leverage, employ | use |
| commence, initiate, begin, originate | start |
| terminate, cease, conclude | stop, end |
| ensure, verify, confirm, validate, check | make sure (that), examine |
| perform, conduct, execute, carry out | do |
| facilitate, assist | help |
| obtain, acquire, procure | get |
| sufficient, adequate | enough |
| approximately | about |
| prior to | before |
| subsequent to, following (prep.) | after |
| adjacent to | near |
| accomplish | do |
| additional, supplementary | more |
| attempt | try |
| require, necessitate | need, must |
| mandatory | necessary |
| indicate, signify | show |
| observe (=watch) | look at, examine |
| rotate | turn |
| deactivate | turn off, set to off |
| activate, energize | turn on, start |
| in order to | to |
| via, by means of | through, with |
| due to, owing to | because of |
| in the event of/that | if |
| accessible | (rewrite: "you can get access to") |
| remainder | rest |
| demonstrate | show |
| modify, alter | change |
| construct, fabricate, build | assemble, make |
| retain | keep |
| locate (=find) | find |
| depress (a button) | push, press |
| proceed | continue, go |

One-meaning rulings that bite hardest on finance copy: **check** is not a verb for
verification (use "make sure that" or "examine"); **above/below** are physical position only
(quantities take "more than"/"less than"); **help** is a verb only (the noun is "aid");
**follow** means only "come after" (for rules, use "obey").
