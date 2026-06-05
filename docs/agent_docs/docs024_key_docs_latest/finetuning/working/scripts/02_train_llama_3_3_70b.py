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
import tarfile
import time
from pathlib import Path

import torch
from datasets import load_dataset
from trl import SFTTrainer, SFTConfig
# transformers is already imported (transitively) by trl above; TrainerCallback is
# a plain base class with no monkey-patch interaction, so importing it here is safe.
from transformers import TrainerCallback
import requests  # present via huggingface_hub; used for presigned PUT/GET


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
    # ---- Phase A (B2 durability): all optional; absent = old behaviour exactly ----
    p.add_argument("--save-steps", type=int, default=0,
                   help="checkpoint every N optimiser steps (0 = no intermediate checkpoints; "
                        "save_strategy stays 'no', as before)")
    p.add_argument("--save-total-limit", type=int, default=2,
                   help="max local checkpoints kept on disk (only applies when --save-steps > 0)")
    p.add_argument("--upload-manifest", default="",
                   help="path to a JSON file of presigned URLs for checkpoint + final-adapter "
                        "upload and (optionally) resume. Absent = no uploads, no resume.")
    return p.parse_args()


# ============================================================================
# Phase A — checkpoint + final-adapter upload to B2, and resume.
#
# The VM never holds a B2 key. The launcher pre-mints single-object, write-only
# presigned PUT URLs (and, on a resume launch, one GET URL) into a JSON file and
# SSH-places it on the box; --upload-manifest points at it. Shape:
#   { "run_id": "...",
#     "checkpoints": [ {"index":1,"key":"...","url":"<PUT>"}, ... ],   # by SAVE index
#     "final":  {"key":"...","url":"<PUT>"},
#     "resume": {"url":"<GET>","index":N} }                            # present only on resume
# Integrity is the eval gate's job downstream, not these URLs'.
# ============================================================================

_OCTET = "application/octet-stream"          # MUST match the Content-Type the launcher signed
_HTTP_TIMEOUT = (30, 1800)                   # (connect, read) seconds — large bodies


def load_upload_manifest(path: str) -> dict | None:
    """Load the presigned-URL manifest, or None if not supplied / not found."""
    if not path:
        return None
    p = Path(path)
    if not p.is_file():
        print(f"WARNING: --upload-manifest {path} not found; uploads + resume DISABLED")
        return None
    with open(p) as f:
        m = json.load(f)
    print(f"Loaded upload manifest: {len(m.get('checkpoints', []))} checkpoint URL(s), "
          f"final={'yes' if m.get('final') else 'no'}, resume={'yes' if m.get('resume') else 'no'}")
    return m


def _tar_dir(src_dir: Path, dest_tar: Path, exclude_top: set[str] | None = None) -> None:
    """gzip-tar src_dir into dest_tar, preserving src_dir.name as the top-level entry.
    exclude_top drops named immediate children (e.g. the 'checkpoints' subdir)."""
    exclude_top = exclude_top or set()

    def _filter(ti: tarfile.TarInfo):
        parts = ti.name.split("/")
        if len(parts) >= 2 and parts[1] in exclude_top:
            return None
        return ti

    with tarfile.open(dest_tar, "w:gz") as tar:
        tar.add(str(src_dir), arcname=src_dir.name, filter=_filter)


def _put_file(url: str, path: Path) -> None:
    """PUT one file to a presigned URL as a fixed-length body (NOT chunked — requests
    derives Content-Length from the file's fstat). Content-Type must match the signature."""
    size = path.stat().st_size
    with open(path, "rb") as fh:
        resp = requests.put(url, data=fh, headers={"Content-Type": _OCTET}, timeout=_HTTP_TIMEOUT)
    if resp.status_code not in (200, 201):
        raise RuntimeError(f"PUT {resp.status_code}: {resp.text[:500]}")
    print(f"  uploaded {path.name} ({size / 1e9:.2f}GB) -> HTTP {resp.status_code}")


def _download_and_extract(url: str, into_dir: Path, workdir: Path) -> None:
    """GET a checkpoint tarball and extract it into into_dir (yielding into_dir/checkpoint-<step>/)."""
    into_dir.mkdir(parents=True, exist_ok=True)
    tmp = workdir / "resume.tar.gz"
    print(f"Downloading resume checkpoint -> {into_dir}")
    with requests.get(url, stream=True, timeout=_HTTP_TIMEOUT) as resp:
        if resp.status_code != 200:
            raise RuntimeError(f"resume GET {resp.status_code}: {resp.text[:500]}")
        with open(tmp, "wb") as fh:
            for chunk in resp.iter_content(chunk_size=8 << 20):
                fh.write(chunk)
    with tarfile.open(tmp, "r:gz") as tar:
        try:
            tar.extractall(str(into_dir), filter="data")   # py3.12+: guards path traversal
        except TypeError:
            tar.extractall(str(into_dir))                  # older Python
    tmp.unlink(missing_ok=True)
    print(f"  extracted resume checkpoint into {into_dir}")


class CheckpointUploader(TrainerCallback):
    """On every Trainer checkpoint, tar it and PUT it to the next presigned URL from the
    manifest, keyed by SAVE INDEX (0,1,2,...) not global_step. Best-effort by design:
    a failed upload is logged loudly but NEVER aborts training — the next checkpoint
    re-establishes durability, and the final-adapter upload is the hard gate."""

    def __init__(self, checkpoints: list[dict], workdir: Path):
        self._urls = checkpoints or []
        self._workdir = workdir
        self._save_index = 0

    def on_save(self, args, state, control, **kwargs):  # noqa: D401 — Trainer hook
        idx = self._save_index
        self._save_index += 1
        if idx >= len(self._urls):
            print(f"WARNING: checkpoint save #{idx} (step {state.global_step}) has no presigned "
                  f"URL in the manifest; leaving it on local disk only")
            return
        ckpt_dir = Path(args.output_dir) / f"checkpoint-{state.global_step}"
        tar_path = self._workdir / f"ckpt-{idx}.tar.gz"
        try:
            if not ckpt_dir.is_dir():
                print(f"WARNING: expected checkpoint dir {ckpt_dir} missing; skipping upload")
                return
            print(f"Uploading checkpoint #{idx} (step {state.global_step}) -> "
                  f"{self._urls[idx].get('key')}")
            _tar_dir(ckpt_dir, tar_path)
            _put_file(self._urls[idx]["url"], tar_path)
        except Exception as e:  # noqa: BLE001 — durability is best-effort; never crash training
            print(f"WARNING: checkpoint #{idx} upload failed (training continues): {e}")
        finally:
            tar_path.unlink(missing_ok=True)


def main() -> None:
    args = parse_args()

    output_path = Path(args.output)
    output_path.mkdir(parents=True, exist_ok=True)

    # ---- Phase A: upload manifest + checkpoint policy -----------------------
    upload_manifest = load_upload_manifest(args.upload_manifest)
    do_checkpoints = args.save_steps > 0          # gates save_strategy below
    resume_ckpt = None                            # True once a resume checkpoint is staged

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
            # Phase A: when --save-steps>0 we checkpoint to <output>/checkpoints so the
            # CheckpointUploader can ship them to B2 and a resume can pick them up. With
            # the default --save-steps 0 this is exactly the old behaviour (no checkpoints;
            # the only persisted artefact is the final model.save_pretrained() below).
            save_strategy=("steps" if do_checkpoints else "no"),
            save_steps=(args.save_steps if do_checkpoints else 500),  # ignored when strategy="no"
            save_total_limit=(args.save_total_limit if do_checkpoints else None),
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

    # ---- Phase A: stage a resume checkpoint, attach the uploader -------------
    # If the launcher handed us a resume URL, pull that checkpoint into <output>/checkpoints
    # so resume_from_checkpoint=True finds it (Trainer restores optimiser/scheduler/RNG/step).
    if upload_manifest and upload_manifest.get("resume"):
        _download_and_extract(
            upload_manifest["resume"]["url"], output_path / "checkpoints", output_path
        )
        resume_ckpt = True
    # Ship each new checkpoint to B2 as it's written (best-effort; never aborts training).
    if do_checkpoints and upload_manifest and upload_manifest.get("checkpoints"):
        trainer.add_callback(CheckpointUploader(upload_manifest["checkpoints"], output_path))

    # ---- Train ---------------------------------------------------------------
    print("Starting training")
    t0 = time.time()
    train_result = trainer.train(resume_from_checkpoint=resume_ckpt)
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

    # ---- Phase A: upload the final adapter ----------------------------------
    # This is the hard gate: the adapter is the run's deliverable, so unlike checkpoint
    # uploads a failure here MUST fail the run (non-zero exit -> run.sh emits no DONE
    # marker -> the monitor will not treat the box as cleanly finished). Excludes the
    # checkpoints/ subdir; tars adapter weights + tokenizer + this manifest.
    if upload_manifest and upload_manifest.get("final"):
        final = upload_manifest["final"]
        final_tar = output_path.parent / "final_adapter.tar.gz"
        print(f"Uploading final adapter -> {final.get('key')}")
        try:
            _tar_dir(output_path, final_tar, exclude_top={"checkpoints"})
            _put_file(final["url"], final_tar)
        finally:
            final_tar.unlink(missing_ok=True)

    print("Done.")


if __name__ == "__main__":
    main()
