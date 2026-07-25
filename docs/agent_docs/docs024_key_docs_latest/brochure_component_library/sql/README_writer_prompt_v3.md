# page-content-writer voice prompt — refinement 3 (2026-07-25)

`page_content_writer_prompt_v3_2026-07-25.txt` is the **exact live
`prompt_template`** after this refinement (the nested config path is
`default_config.workflow.steps.process_sections_loop.config.sub_workflow.steps.generate_content.config.prompt_template`).
Backups: `bak_agent_definitions_pcw_20260724` (v2), `bak_agent_definitions_pcw_20260725`
and `_20260725b` (this round, before each write).

## Why a third refinement, and why the first two failed

Refinements 1 and 2 both restated the em-dash rule more forcefully. Neither
worked: live pages still carried 8 / 6 / 5 / 2 em dashes per page. Two causes,
both measured rather than guessed:

**1. The rule described the wrong shape.** Every em dash actually produced was an
*appositive gloss* (a noun, a dash, a phrase re-explaining the noun), not the long
parenthetical aside the rule warned about. The model had no reason to think the
rule applied. Fixed by naming the shape and showing two real failures from its own
output with their rewrites, per the v3 style-prompt finding that worked examples
move behaviour where rules do not.

**2. The prompt banning em dashes contained 17 of them.** Fourteen were in its own
instructional prose — headings (`## Voice & Style (how the copy must READ — follow
strictly)`), other sections' guidance, the anti-fabrication rules. The model was
shown fourteen examples of the banned style in the most authoritative text in its
context while being told never to produce it. All fourteen rewritten as full
stops, colons or commas.

**Exactly three em dashes remain, all deliberate**, and a future edit must keep
them: the one naming the character to scan for, and the two inside `WRONG:`
examples that exist to demonstrate the fault. A verification that simply counts
em dashes in this prompt should expect 3, not 0.

## Applying / re-applying

Do **not** hand-edit through `psql` string literals: the prompt contains newlines,
single quotes, Go template braces and the em dash itself. `\copy … FROM PROGRAM`
also fails, because JSON newlines are read as row separators (tried; the NOT NULL
constraint caught it, nothing was corrupted). Round-trip via base64:

```bash
# dump
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA -c "
SELECT default_config->'workflow'->'steps'->'process_sections_loop'->'config'->'sub_workflow'
       ->'steps'->'generate_content'->'config'->>'prompt_template'
FROM agent_definitions WHERE type='page-content-writer' AND is_active AND deleted_at IS NULL;" > prompt.txt
# (psql -tA appends one trailing newline: strip it before re-applying)

# apply
B=$(python3 -c "import io,base64;s=io.open('prompt.txt',encoding='utf-8').read();s=s[:-1] if s.endswith(chr(10)) else s;print(base64.b64encode(s.encode()).decode())")
# UPDATE … jsonb_set(default_config, '{…,prompt_template}',
#   to_jsonb(convert_from(decode('<B>','base64'),'UTF8')), false)
```

## Open

Effect is **UNMEASURED**: this is config, so it is live immediately, but no page
has been written since. The read is a fresh page build, then
`count(regexp_matches(rendered_html,'—','g'))` per component against the
pre-change baseline (index 6 / about 8 / capabilities 6 / fine-tuning 5 /
council 2). If a fourth round is needed, the mechanical post-pass is the
alternative the owner has been offered.
