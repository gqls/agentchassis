#!/usr/bin/env python3
"""
04_eval_iter0.py — flywheel D runner

For each prompt in a held-out JSONL, generate iter_0's response and write
(source_log_id, prompt, claude_response, iter0_response, timing) to a
results JSONL. Streamed: each result is flushed immediately so an
interrupted run loses at most one row.

Use the same Unsloth + adapter loading pattern as 03_inference_test.py.

Usage:
    python 04_eval_iter0.py \\
        --adapter ~/lora_iter0_full \\
        --cases  ~/held_out_cases_v1.jsonl \\
        --output ~/iter0_eval_results_v1.jsonl \\
        --n 20
"""
from __future__ import annotations

# Unsloth import order matters — must come before transformers/peft
from unsloth import FastLanguageModel
from unsloth.chat_templates import get_chat_template

import argparse
import json
import os
import time
from pathlib import Path

import torch


def parse_args() -> argparse.Namespace:
    p = argparse.ArgumentParser(description=__doc__.split("\n", 1)[0])
    p.add_argument("--adapter", required=True, help="Path to LoRA adapter directory")
    p.add_argument("--cases", required=True, help="Held-out JSONL of cases to evaluate")
    p.add_argument("--output", required=True, help="Where to write results JSONL")
    p.add_argument("--base-model", default="unsloth/Llama-3.3-70B-Instruct-bnb-4bit")
    p.add_argument("--max-seq", type=int, default=4096)
    p.add_argument("--max-new-tokens", type=int, default=512,
                   help="generation cap. 512 covers all observed iter_0 outputs comfortably.")
    p.add_argument("--n", type=int, default=0,
                   help="cap rows (0 = all). Use --n 20 for a first round.")
    p.add_argument("--skip", type=int, default=0, help="skip first N cases")
    p.add_argument("--temperature", type=float, default=0.0,
                   help="0.0 = greedy. Use small >0 if you want sampling diversity.")
    p.add_argument("--seed", type=int, default=3407)
    return p.parse_args()


def load_cases(path: str, skip: int, n: int) -> list[dict]:
    out: list[dict] = []
    with open(path) as f:
        for i, line in enumerate(f):
            line = line.strip()
            if not line:
                continue
            if i < skip:
                continue
            out.append(json.loads(line))
            if n > 0 and len(out) >= n:
                break
    return out


def main() -> None:
    args = parse_args()

    output_path = Path(args.output)
    output_path.parent.mkdir(parents=True, exist_ok=True)

    # ---- Resume support: skip rows already in output --------------------------
    already_done: set[str] = set()
    if output_path.exists():
        with open(output_path) as f:
            for line in f:
                try:
                    already_done.add(json.loads(line)["source_log_id"])
                except Exception:
                    pass
        if already_done:
            print(f"Resume: {len(already_done)} rows already in {output_path}, will skip.")

    # ---- Load model + adapter -------------------------------------------------
    torch.cuda.reset_peak_memory_stats()
    print(f"GPU: {torch.cuda.get_device_name(0)}")
    print(f"Loading base model: {args.base_model}")
    model, tokenizer = FastLanguageModel.from_pretrained(
        model_name=args.base_model,
        max_seq_length=args.max_seq,
        load_in_4bit=True,
        dtype=None,
    )

    print(f"Loading adapter: {args.adapter}")
    model.load_adapter(args.adapter)
    tokenizer = get_chat_template(tokenizer, chat_template="llama-3.1")

    # Set pad token explicitly (avoids the "pad token same as eos" warning we
    # saw in the smoke and produces deterministic batched-or-not generation).
    if tokenizer.pad_token_id is None or tokenizer.pad_token_id == tokenizer.eos_token_id:
        pad_id = tokenizer.convert_tokens_to_ids("<|finetune_right_pad_id|>")
        if pad_id is not None and pad_id != tokenizer.unk_token_id:
            tokenizer.pad_token_id = pad_id
        # if that token isn't in the vocab for some Llama variant, leave as-is

    FastLanguageModel.for_inference(model)

    # ---- Load cases ----------------------------------------------------------
    cases = load_cases(args.cases, args.skip, args.n)
    cases = [c for c in cases if c.get("metadata", {}).get("source_log_id") not in already_done]
    print(f"Cases to process: {len(cases)}")

    # ---- Generate ------------------------------------------------------------
    do_sample = args.temperature > 0.0
    fout = open(output_path, "a", buffering=1)  # line-buffered

    overall_start = time.time()
    for i, case in enumerate(cases, 1):
        meta = case.get("metadata", {})
        source_log_id = meta.get("source_log_id", f"unknown_{i}")
        msgs = case["messages"]
        user_prompt = msgs[0]["content"]
        claude_response = msgs[1]["content"]

        prompt_text = tokenizer.apply_chat_template(
            [{"role": "user", "content": user_prompt}],
            tokenize=False,
            add_generation_prompt=True,
        )
        inputs = tokenizer(prompt_text, return_tensors="pt").to("cuda")
        prompt_tokens = inputs["input_ids"].shape[1]

        gen_start = time.time()
        with torch.inference_mode():
            outputs = model.generate(
                **inputs,
                max_new_tokens=args.max_new_tokens,
                do_sample=do_sample,
                temperature=args.temperature if do_sample else 1.0,
                pad_token_id=tokenizer.pad_token_id or tokenizer.eos_token_id,
            )
        gen_time_s = time.time() - gen_start

        # Slice off the prompt tokens — keep only the generated continuation
        gen_tokens = outputs[0][inputs["input_ids"].shape[1]:]
        iter0_response = tokenizer.decode(gen_tokens, skip_special_tokens=True).strip()
        gen_token_count = int(gen_tokens.shape[0])

        result = {
            "source_log_id":     source_log_id,
            "orchestration_id":  meta.get("orchestration_id"),
            "created_at":        meta.get("created_at"),
            "prompt":            user_prompt,
            "claude_response":   claude_response,
            "iter0_response":    iter0_response,
            "prompt_tokens":     prompt_tokens,
            "generation_tokens": gen_token_count,
            "generation_time_s": round(gen_time_s, 2),
        }
        fout.write(json.dumps(result, ensure_ascii=False) + "\n")
        fout.flush()

        elapsed = time.time() - overall_start
        eta = elapsed / i * (len(cases) - i) if i > 0 else 0
        print(f"[{i}/{len(cases)}] {source_log_id[:8]}… "
              f"gen={gen_time_s:.1f}s ({gen_token_count} tok) "
              f"elapsed={elapsed:.0f}s eta={eta:.0f}s")

    fout.close()
    peak_vram = torch.cuda.max_memory_allocated() / 1e9
    print(f"\nDone. peak_vram={peak_vram:.1f}GB  total={time.time()-overall_start:.0f}s")
    print(f"Results: {output_path}")


if __name__ == "__main__":
    main()
