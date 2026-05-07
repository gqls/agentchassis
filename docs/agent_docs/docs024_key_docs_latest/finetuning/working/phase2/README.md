add money to thunder

Get it on the VM:
bash# from your laptop
tnr scp ./flywheel_C/02_train_llama_3_3_70b.py 0:/home/ubuntu/

Smoke run on the VM (~5-10 min, 20 rows × 1 epoch — fast signal that the model loads, the chat template applies, and the training loop runs at all):
bashsource ~/unsloth_env/bin/activate
cd ~

python 02_train_llama_3_3_70b.py \
--data ~/training_iter0.jsonl \
--output ~/lora_smoke \
--limit 20 --epochs 1 \
2>&1 | tee ~/smoke_train.log

--

look at training data:
jq -c '.messages[1].content' training_iter0.jsonl | head -5

--

training done
Run summary numbers:
33,134 seconds = 9h 12m wall time
Peak VRAM 44.8GB (under our 80GB ceiling, headroom to spare)
Cost: 9.2h × $1.79 = $16.50 — matches the projection
Adapter at ~/lora_iter0_full/ with manifest

Loss curve, epoch by epoch:

Epoch 1: starts 1.49, ends ~0.27
Epoch 2: 0.34 → 0.18 (mean ~0.23)
Epoch 3: 0.14 → 0.10 (mean ~0.12)

That's textbook learning behaviour — sharp drop in epoch 1 as the model picks up the JSON schema and chat template, then steady refinement on content patterns through 2 and 3. final_loss=0.267 is the trailing average; the actual end-of-run loss values are ~0.10-0.13. No spikes, no NaN, gradient norms stayed sane (mostly 0.3-0.5 with two harmless excursions early on).
One mild caveat I'd flag, not act on: the gap between epoch 2 mean (~0.23) and epoch 3 mean (~0.12) means the model is increasingly memorising rather than generalising in epoch 3. This is the classic "should we have stopped at 2 epochs?" signal. We won't know without inference quality testing — overfitting on a structured-JSON task with consistent schema is less harmful than overfitting on freeform prose, because there's less variance to lose. Note for iter_1: a --epochs 2 ablation would be useful comparative data.

scp 03 script onto thunder
:~/projects/agentchassis/docs/agent_docs/docs024_key_docs_latest/finetuning/working/phase1$ tnr scp ./03_inference_test.py 0:/home/ubuntu
_inference_test.py 0:/home/ubuntu
Uploading ./03_inference_test.py to oecrda3u:/home/ubuntu
sending incremental file list
03_inference_test.py
4,949 100%    0.00kB/s    0:00:00 (xfr#1, to-chk=0/1)
Transfer complete

set up the env again
source ~/unsloth_env/bin/activate

1. Inference smoke test on the VM (3-5 min, $0.15) — verify the adapter actually does something useful before stopping the instance. Re-present 03_inference_test.py?
   bash python 03_inference_test.py --adapter ~/lora_iter0_full \
   --data ~/training_iter0.jsonl --n 5 --skip 1900
   What we want: clean JSON output, no code fences, no preamble, schema matches what the prompt asked for. If those hold on rows the model trained on, then a separate test on a novel brief tells us whether it generalises.
2. Get the adapter off the VM — Thunder storage is ephemeral. Stopping without exfiltrating means losing the $16 of work.
   bash # from laptop
   tnr scp 0:/home/ubuntu/lora_iter0_full ./lora_iter0_full
   ls -lh lora_iter0_full/
   Should be ~150-200MB total. Verify adapter_model.safetensors and adapter_config.json are present and the manifest reads cleanly.
3. (Snapshot first via tnr snapshot 0 unsloth-trainer-base if you want to skip the env-rebuild next time — saves ~15 min on iter_1.)
   Stop the instance once (1) and (2) are confirmed:
   bash tnr stop 0
  
the inference test notes:
What the script does: loads the 4-bit base model, applies your LoRA adapter on top, generates outputs for N training prompts, checks JSON validity and key presence. Single model under test — the adapter+base combination.
What it doesn't do: compare against the untrained base. There's no A/B in this script. So it answers "does the trained model produce well-formed output" but not "did training actually improve anything."
Why I'd argue that's actually fine for this step. The comparison "trained adapter vs untrained base" is a weak signal anyway:

The untrained Llama 3.3 70B Instruct will follow the JSON-schema instructions reasonably well on its own, just with more verbose-creative-marketing-speak filler than your training data. So the comparison ends up being aesthetic ("this one sounds more like our brand") rather than functional ("this one works and that one doesn't").
The actually useful comparison is adapter vs Claude Sonnet 4.6 — Claude is the source of the training labels, so it's the ceiling we're trying to approach. That's exactly what flywheel D is designed for, and it's a better use of evaluation effort.
Running base vs adapter would also cost an extra ~5 min and ~$0.15 on the GPU, plus risks the adapter-toggle path having its own subtle bugs.

--
Strong result. Reading what came back:
5/5 valid JSON, 5/5 schema-correct, no preamble, no code fences, no hallucinations. Several outputs are essentially identical to the expected text (cases 1, 3, 5) — that's the model recognising training prompts and reproducing what it learned. Case 4 is more interesting: paraphrased ("rather be building than spreadsheeting" vs "rather spend less time in spreadsheets") with the same structural pattern and CTA fields preserved. That's the model generalising rather than just memorising. Case 2 produced two-key minimal output where the expected was also two keys — schema awareness intact.
Two minor observations, neither blocking:
The pad-token-equals-eos warning in case 1 is a benign Llama-3 quirk (its tokenizer ships with <|finetune_right_pad_id|> set equal to <|end_of_text|>). It can cause sequence-end ambiguity for batch inference. Single-prompt generation like the smoke isn't affected. For flywheel D's eval harness, worth setting tokenizer.pad_token_id = tokenizer.convert_tokens_to_ids("<|finetune_right_pad_id|>") explicitly if it does batched inference. Note for the patch document.
The held-back caveat: smoke prompts are training rows (--skip 1900 is rows 1900-1904 from the same dataset). High-quality output here means the pipeline works end-to-end and the model didn't break itself, not necessarily that it generalises to novel briefs. That's exactly what flywheel D is for, so we don't need to chase that signal here.
Closing the loop
Three commands, then the GPU goes away. Run from your laptop:
bash# 1. Pull the adapter off the ephemeral VM
tnr scp 0:/home/ubuntu/lora_iter0_full ./lora_iter0_full

# 2. Confirm what arrived
ls -lh lora_iter0_full/
cat lora_iter0_full/manifest.json
Expect roughly 150-300MB total, with adapter_model.safetensors, adapter_config.json, and the manifest. If safetensors is missing or under 50MB, something went wrong with the save and we shouldn't tear down yet.
bash# 3. Optional: snapshot before stopping, so iter_1 starts in 90 seconds not 90 minutes
tnr snapshot 0 unsloth-trainer-base

# 4. Stop billing
tnr stop 0



## Cost recap for the patch document, since this turned out useful baseline data:
Setup + smoke~1.5h

Full training run (3 epochs, 1958 rows, max_seq 4096)9.2h

Inference smoke + adapter exfil~30 min

Total iter_0~11.2h

Snapshot ongoing


Worth flagging: we picked Production mode at the start of this session, which is why we're at $1.79/hr instead of the $0.78/hr Prototyping rate that gets quoted in Thunder's marketing. For phase-2's chassis-driven trainer, prototyping might be acceptable — training jobs are bounded-time, restart-tolerant, and we have snapshots to recover environment state. Worth testing on iter_1 whether prototyping mode actually delivers usable A100 80GB performance, or whether the virtualisation overhead degrades training speed enough to cancel the cost savings. That's a good experiment for next time, not this session.

