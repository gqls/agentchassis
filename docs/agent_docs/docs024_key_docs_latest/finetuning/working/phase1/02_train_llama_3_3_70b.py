#!/usr/bin/env python3
"""
02_train_llama_3_3_70b.py
============================================================================
QLoRA fine-tune of Llama 3.3 70B Instruct on our page-content-writer iter_0
dataset using Unsloth. Designed for a single H100 80GB or A100 80GB.

Input dataset format (one JSON object per line):
    {
      "messages": [
        {"role": "user", "content": "<prompt_rendered>"},
        {"role": "assistant", "content": "<cleaned response>"}
      ],
      "metadata": {...}
    }

This matches what 01_pull_dataset_from_postgres.sh produces.

Defaults tuned for ~1,958 training examples, narrow-task adaptation (format
and voice, not new knowledge). Training wall time expected: ~30-90 minutes
on H100 80GB. Adjust `num_train_epochs`, `per_device_train_batch_size`, and
`max_seq_length` if VRAM or speed pushes back.

Output: LoRA adapter saved to --output dir (~150MB). Merge later if desired.
============================================================================
"""

import argparse
import json
import os
import sys

# Unsloth MUST be imported before transformers/trl per its README
from unsloth import FastLanguageModel, is_bfloat16_supported
import torch
from datasets import load_dataset
from trl import SFTTrainer, SFTConfig


MODEL_NAME = "unsloth/Llama-3.3-70B-Instruct-bnb-4bit"
# Note: Unsloth hosts pre-quantized 4-bit versions of popular models; much
# faster download than quantizing locally.


def parse_args():
    ap = argparse.ArgumentParser(description="Fine-tune Llama 3.3 70B on our dataset")
    ap.add_argument("--data", required=True, help="Path to training JSONL")
    ap.add_argument("--output", required=True, help="Output directory for LoRA adapter")
    ap.add_argument("--max-seq-length", type=int, default=4096,
                    help="Max context length. Our p99 prompt + response fits in ~2048 tokens.")
    ap.add_argument("--epochs", type=int, default=3,
                    help="Epochs (3 is a solid default for narrow-task adaptation)")
    ap.add_argument("--batch-size", type=int, default=1,
                    help="Per-GPU batch size. For 70B QLoRA on 80GB, 1 is safe; 2 if seq-len is short.")
    ap.add_argument("--grad-accum", type=int, default=8,
                    help="Gradient accumulation steps. Effective batch = batch_size * grad_accum.")
    ap.add_argument("--lr", type=float, default=2e-4,
                    help="Learning rate — 2e-4 is standard for QLoRA on 70B")
    ap.add_argument("--lora-r", type=int, default=16,
                    help="LoRA rank. 16 for narrow tasks, 32 for broader adaptation.")
    ap.add_argument("--lora-alpha", type=int, default=16,
                    help="LoRA alpha. Typically = rank.")
    ap.add_argument("--seed", type=int, default=3407, help="Random seed")
    ap.add_argument("--limit", type=int, default=0,
                    help="Optional: limit to N examples for a quick smoke run")
    return ap.parse_args()


def format_example(example, tokenizer):
    """
    Convert one ChatML-shaped row into the model's expected conversational
    format. Llama 3.3 uses the Llama 3 instruct chat template.

    Input: example["messages"] is [{role, content}, {role, content}]
    Output: {"text": "<full templated conversation including answer>"}
    """
    messages = example["messages"]
    text = tokenizer.apply_chat_template(
        messages,
        tokenize=False,
        add_generation_prompt=False,  # False because we include the assistant's answer
    )
    return {"text": text}


def main():
    args = parse_args()

    if not os.path.exists(args.data):
        sys.exit(f"Training data file not found: {args.data}")

    os.makedirs(args.output, exist_ok=True)

    print("=" * 72)
    print("Llama 3.3 70B QLoRA fine-tune")
    print("=" * 72)
    print(f"Model:          {MODEL_NAME}")
    print(f"Data:           {args.data}")
    print(f"Output:         {args.output}")
    print(f"Max seq len:    {args.max_seq_length}")
    print(f"Epochs:         {args.epochs}")
    print(f"Batch / grad:   {args.batch_size} / {args.grad_accum}")
    print(f"LR:             {args.lr}")
    print(f"LoRA r / alpha: {args.lora_r} / {args.lora_alpha}")
    print("=" * 72)
    print()

    # ── Load model + tokenizer ──────────────────────────────────────────────

    print("Loading base model (4-bit) — this is the big download if first time...")
    model, tokenizer = FastLanguageModel.from_pretrained(
        model_name=MODEL_NAME,
        max_seq_length=args.max_seq_length,
        dtype=None,                  # auto: bf16 on Ampere/Hopper
        load_in_4bit=True,
    )

    # ── Attach LoRA adapters ────────────────────────────────────────────────

    print("Attaching LoRA adapters to all linear layers...")
    model = FastLanguageModel.get_peft_model(
        model,
        r=args.lora_r,
        target_modules=[
            "q_proj", "k_proj", "v_proj", "o_proj",
            "gate_proj", "up_proj", "down_proj",
        ],
        lora_alpha=args.lora_alpha,
        lora_dropout=0.0,            # 0 is Unsloth-optimised
        bias="none",
        use_gradient_checkpointing="unsloth",  # 30% less VRAM
        random_state=args.seed,
        use_rslora=False,
        loftq_config=None,
    )

    # ── Load + format dataset ───────────────────────────────────────────────

    print(f"Loading dataset from {args.data}...")
    dataset = load_dataset("json", data_files=args.data, split="train")
    print(f"  loaded {len(dataset)} examples")

    if args.limit > 0 and args.limit < len(dataset):
        dataset = dataset.select(range(args.limit))
        print(f"  limited to {len(dataset)} examples (smoke test mode)")

    dataset = dataset.map(
        lambda ex: format_example(ex, tokenizer),
        remove_columns=dataset.column_names,
        desc="Formatting conversations",
    )

    # Sanity — log the first formatted example's length so we know the
    # sequence profile matches our --max-seq-length.
    sample = dataset[0]["text"]
    sample_tokens = tokenizer(sample, return_tensors="pt")["input_ids"].shape[1]
    print(f"  first example: {sample_tokens} tokens (max_seq={args.max_seq_length})")

    # ── Trainer ─────────────────────────────────────────────────────────────

    sft_config = SFTConfig(
        output_dir=args.output,
        num_train_epochs=args.epochs,
        per_device_train_batch_size=args.batch_size,
        gradient_accumulation_steps=args.grad_accum,
        learning_rate=args.lr,
        lr_scheduler_type="linear",
        warmup_steps=5,
        logging_steps=5,
        save_strategy="epoch",
        optim="adamw_8bit",
        weight_decay=0.01,
        seed=args.seed,
        bf16=is_bfloat16_supported(),
        fp16=not is_bfloat16_supported(),
        dataset_num_proc=2,
        dataset_text_field="text",
        max_seq_length=args.max_seq_length,
        packing=False,               # keep off for clean per-example loss
        report_to=[],                # silence W&B etc. by default
    )

    trainer = SFTTrainer(
        model=model,
        tokenizer=tokenizer,
        train_dataset=dataset,
        args=sft_config,
    )

    # ── VRAM snapshot before training ───────────────────────────────────────
    start_gpu = torch.cuda.memory_allocated() / 1e9
    print(f"\nVRAM after load: {start_gpu:.1f} GB used")
    print()

    # ── Train ───────────────────────────────────────────────────────────────
    print("=" * 72)
    print("Training starting...")
    print("=" * 72)
    trainer_stats = trainer.train()

    print()
    print("=" * 72)
    print(f"Training complete: {trainer_stats.metrics.get('train_runtime', 0):.1f}s total")
    print(f"Final loss: {trainer_stats.metrics.get('train_loss', 'unknown')}")
    print(f"VRAM peak: {torch.cuda.max_memory_reserved() / 1e9:.1f} GB")
    print("=" * 72)

    # ── Save LoRA adapter ───────────────────────────────────────────────────
    print(f"\nSaving LoRA adapter to {args.output}...")
    model.save_pretrained(args.output)
    tokenizer.save_pretrained(args.output)
    print("Done. Adapter size:")
    os.system(f"du -sh {args.output}")

    # ── Write a small manifest for future reference ────────────────────────
    manifest = {
        "model_base": MODEL_NAME,
        "data_path": args.data,
        "epochs": args.epochs,
        "per_device_batch_size": args.batch_size,
        "grad_accum": args.grad_accum,
        "learning_rate": args.lr,
        "lora_r": args.lora_r,
        "lora_alpha": args.lora_alpha,
        "max_seq_length": args.max_seq_length,
        "train_loss_final": trainer_stats.metrics.get("train_loss"),
        "train_runtime_seconds": trainer_stats.metrics.get("train_runtime"),
        "vram_peak_gb": torch.cuda.max_memory_reserved() / 1e9,
        "dataset_examples": len(dataset),
    }
    with open(os.path.join(args.output, "training_manifest.json"), "w") as f:
        json.dump(manifest, f, indent=2)
    print(f"\nManifest written to {args.output}/training_manifest.json")


if __name__ == "__main__":
    main()
