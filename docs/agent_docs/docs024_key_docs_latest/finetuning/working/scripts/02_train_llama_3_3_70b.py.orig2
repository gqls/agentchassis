#!/usr/bin/env python3
"""
02_train_llama_3_3_70b.py — QLoRA fine-tune of Llama 3.3 70B Instruct
on JSONL produced by training_exports (each line = {messages, metadata}).

Defaults match HANDOFF_2026-04-23: 3 epochs, batch 1, grad_accum 8,
lr 2e-4, lora_r 16, max_seq 4096. Override any of them via CLI.

Outputs the LoRA adapter (~150MB) plus a manifest.json describing the run
(used downstream by the model-trainer agent in flywheel C phase 2).
"""
from __future__ import annotations

# Unsloth MUST be imported before transformers/peft/trl so its monkey-patches
# land before any of the modules being patched are loaded.
from unsloth import FastLanguageModel
from unsloth.chat_templates import get_chat_template, train_on_responses_only

import argparse
import json
import time
from pathlib import Path

import torch
from datasets import load_dataset
from trl import SFTTrainer, SFTConfig


def parse_args() -> argparse.Namespace:
    p = argparse.ArgumentParser(description=__doc__.split("\n", 1)[0])
    p.add_argument("--data", required=True, help="Path to training JSONL")
    p.add_argument("--output", required=True, help="Output dir for LoRA adapter + manifest")
    p.add_argument("--base-model", default="unsloth/Llama-3.3-70B-Instruct-bnb-4bit")
    p.add_argument("--epochs", type=int, default=3)
    p.add_argument("--batch", type=int, default=1, help="per-device batch size")
    p.add_argument("--grad-accum", type=int, default=8)
    p.add_argument("--lr", type=float, default=2e-4)
    p.add_argument("--lora-r", type=int, default=16)
    p.add_argument("--lora-alpha", type=int, default=16)
    p.add_argument("--max-seq", type=int, default=4096)
    p.add_argument("--limit", type=int, default=0, help="cap rows (0 = use all). For smoke tests.")
    p.add_argument("--seed", type=int, default=3407)
    p.add_argument("--warmup-steps", type=int, default=5)
    p.add_argument("--logging-steps", type=int, default=5)
    return p.parse_args()


def main() -> None:
    args = parse_args()

    output_path = Path(args.output)
    output_path.mkdir(parents=True, exist_ok=True)

    # ---- VRAM start ----------------------------------------------------------
    torch.cuda.reset_peak_memory_stats()
    free_start_gb = torch.cuda.mem_get_info(0)[0] / 1e9
    total_gb = torch.cuda.get_device_properties(0).total_memory / 1e9
    print(f"GPU: {torch.cuda.get_device_name(0)}  free={free_start_gb:.1f}GB / total={total_gb:.1f}GB")

    # ---- Base model + tokenizer ---------------------------------------------
    print(f"Loading base model: {args.base_model}")
    model, tokenizer = FastLanguageModel.from_pretrained(
        model_name=args.base_model,
        max_seq_length=args.max_seq,
        load_in_4bit=True,
        dtype=None,   # Unsloth picks bf16 on A100/H100, fp16 elsewhere
    )

    # Llama 3.3 uses the same chat template as 3.1
    tokenizer = get_chat_template(tokenizer, chat_template="llama-3.1")

    # ---- LoRA adapters -------------------------------------------------------
    print(f"Adding LoRA adapters (r={args.lora_r}, alpha={args.lora_alpha})")
    model = FastLanguageModel.get_peft_model(
        model,
        r=args.lora_r,
        target_modules=[
            "q_proj", "k_proj", "v_proj", "o_proj",
            "gate_proj", "up_proj", "down_proj",
        ],
        lora_alpha=args.lora_alpha,
        lora_dropout=0,
        bias="none",
        use_gradient_checkpointing="unsloth",
        random_state=args.seed,
        use_rslora=False,
        loftq_config=None,
    )

    # ---- Dataset -------------------------------------------------------------
    print(f"Loading dataset: {args.data}")
    ds = load_dataset("json", data_files=args.data, split="train")
    if args.limit > 0:
        ds = ds.select(range(min(args.limit, len(ds))))
    print(f"  rows: {len(ds)}")

    def to_text(example: dict) -> dict:
        return {
            "text": tokenizer.apply_chat_template(
                example["messages"],
                tokenize=False,
                add_generation_prompt=False,
            )
        }

    ds = ds.map(to_text, num_proc=2, remove_columns=ds.column_names)

    # ---- Trainer -------------------------------------------------------------
    print("Configuring trainer")
    trainer = SFTTrainer(
        model=model,
        tokenizer=tokenizer,
        train_dataset=ds,
        args=SFTConfig(
            output_dir=str(output_path / "checkpoints"),
            per_device_train_batch_size=args.batch,
            gradient_accumulation_steps=args.grad_accum,
            num_train_epochs=args.epochs,
            warmup_steps=args.warmup_steps,
            learning_rate=args.lr,
            logging_steps=args.logging_steps,
            optim="adamw_8bit",
            weight_decay=0.01,
            lr_scheduler_type="linear",
            seed=args.seed,
            bf16=torch.cuda.is_bf16_supported(),
            fp16=not torch.cuda.is_bf16_supported(),
            save_strategy="no",   # final save via model.save_pretrained() below
            report_to="none",
            dataset_text_field="text",
            max_seq_length=args.max_seq,
            dataset_num_proc=2,
            packing=False,
        ),
    )

    # Train only on assistant tokens — the user/system tokens contribute zero
    # loss. This is what makes finetuning on chat data behave well.
    trainer = train_on_responses_only(
        trainer,
        instruction_part="<|start_header_id|>user<|end_header_id|>\n\n",
        response_part="<|start_header_id|>assistant<|end_header_id|>\n\n",
    )

    # ---- Train ---------------------------------------------------------------
    print("Starting training")
    t0 = time.time()
    train_result = trainer.train()
    runtime_s = time.time() - t0
    final_loss = (
        float(train_result.training_loss) if train_result.training_loss is not None else None
    )
    peak_vram_gb = torch.cuda.max_memory_allocated() / 1e9

    print(f"Training complete: {runtime_s:.1f}s, final_loss={final_loss}, peak_vram={peak_vram_gb:.1f}GB")

    # ---- Save adapter --------------------------------------------------------
    print(f"Saving LoRA adapter to {output_path}")
    model.save_pretrained(str(output_path))
    tokenizer.save_pretrained(str(output_path))

    # ---- Manifest ------------------------------------------------------------
    manifest = {
        "base_model": args.base_model,
        "dataset_path": args.data,
        "n_examples": len(ds),
        "epochs": args.epochs,
        "per_device_batch_size": args.batch,
        "gradient_accumulation_steps": args.grad_accum,
        "effective_batch_size": args.batch * args.grad_accum,
        "learning_rate": args.lr,
        "lora_r": args.lora_r,
        "lora_alpha": args.lora_alpha,
        "max_seq_length": args.max_seq,
        "seed": args.seed,
        "limit": args.limit,
        "warmup_steps": args.warmup_steps,
        "train_runtime_s": round(runtime_s, 1),
        "final_loss": final_loss,
        "peak_vram_gb": round(peak_vram_gb, 1),
        "torch_version": torch.__version__,
        "completed_at": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
    }
    manifest_path = output_path / "manifest.json"
    with open(manifest_path, "w") as f:
        json.dump(manifest, f, indent=2)
    print(f"Wrote {manifest_path}")
    print("Done.")


if __name__ == "__main__":
    main()
