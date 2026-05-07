#!/usr/bin/env python3
"""
03_inference_test.py
============================================================================
Quick local sanity test of a trained LoRA adapter. Loads the base model +
adapter and generates responses for a handful of prompts pulled from the
training dataset (or a held-out test file).

This is NOT a proper evaluation — that comes later with the flywheel D
comparison harness. This is just "does the trained model produce sensible
output for prompts it was trained on?"

Usage:
  python 03_inference_test.py --adapter /workspace/lora_iter0 \
      --data /workspace/training_iter0.jsonl \
      --n 5
============================================================================
"""

import argparse
import json
import sys
from pathlib import Path

from unsloth import FastLanguageModel
import torch


def parse_args():
    ap = argparse.ArgumentParser()
    ap.add_argument("--adapter", required=True, help="Path to LoRA adapter dir")
    ap.add_argument("--data", required=True, help="JSONL to sample prompts from")
    ap.add_argument("--n", type=int, default=5, help="Number of samples to run")
    ap.add_argument("--max-new-tokens", type=int, default=300,
                    help="Max tokens to generate per prompt (iter_0 outputs ~100-200)")
    ap.add_argument("--temperature", type=float, default=0.3,
                    help="Low temp for structured-JSON output")
    ap.add_argument("--skip", type=int, default=0,
                    help="Skip the first N rows (useful for held-out sampling)")
    return ap.parse_args()


def main():
    args = parse_args()

    # ── Load base + adapter ────────────────────────────────────────────────
    print(f"Loading adapter from {args.adapter}...")
    model, tokenizer = FastLanguageModel.from_pretrained(
        model_name=args.adapter,         # Unsloth auto-detects it's a LoRA
        max_seq_length=4096,
        dtype=None,
        load_in_4bit=True,
    )
    FastLanguageModel.for_inference(model)  # enables Unsloth's 2x faster inference

    # ── Sample prompts ─────────────────────────────────────────────────────
    prompts = []
    with open(args.data) as f:
        for i, line in enumerate(f):
            if i < args.skip:
                continue
            if len(prompts) >= args.n:
                break
            row = json.loads(line)
            user_msg = next(m for m in row["messages"] if m["role"] == "user")
            expected = next(m for m in row["messages"] if m["role"] == "assistant")
            prompts.append({
                "prompt": user_msg["content"],
                "expected": expected["content"],
                "source_log_id": row.get("metadata", {}).get("source_log_id", "?"),
            })

    # ── Run inference on each ──────────────────────────────────────────────
    for i, item in enumerate(prompts, 1):
        print()
        print("=" * 72)
        print(f"[{i}/{len(prompts)}] source_log_id={item['source_log_id']}")
        print("=" * 72)

        messages = [{"role": "user", "content": item["prompt"]}]
        inputs = tokenizer.apply_chat_template(
            messages,
            tokenize=True,
            add_generation_prompt=True,
            return_tensors="pt",
        ).to(model.device)

        with torch.inference_mode():
            outputs = model.generate(
                inputs,
                max_new_tokens=args.max_new_tokens,
                temperature=args.temperature,
                do_sample=args.temperature > 0,
                use_cache=True,
                pad_token_id=tokenizer.eos_token_id,
            )

        new_tokens = outputs[0, inputs.shape[1]:]
        generated = tokenizer.decode(new_tokens, skip_special_tokens=True)

        # Basic structural check
        try:
            parsed = json.loads(generated.strip())
            keys = sorted(parsed.keys()) if isinstance(parsed, dict) else "[non-object]"
            json_ok = True
        except Exception as e:
            keys = f"[parse failed: {e}]"
            json_ok = False

        print(f"Prompt length: {len(item['prompt'])} chars")
        print(f"Expected (first 200 chars):\n  {item['expected'][:200]}")
        print()
        print(f"Generated (first 400 chars):\n  {generated[:400]}")
        print()
        print(f"JSON valid: {json_ok}")
        print(f"Keys: {keys}")

    print()
    print("=" * 72)
    print("Inference test complete.")
    print("=" * 72)
    print()
    print("If outputs look like sensible JSON matching the expected schema,")
    print("the fine-tune worked. Proceed to proper evaluation.")


if __name__ == "__main__":
    main()
