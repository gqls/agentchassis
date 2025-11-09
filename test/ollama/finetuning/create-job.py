#!/usr/bin/env python3
"""
create_job.py
-------------
Generates:
  • train.py  — a minimal CPU-friendly fine-tuning example
  • job.yaml  — a Kubernetes Job that runs it in a Python container
Then applies the Job using kubectl (if available).
"""

import os, subprocess, textwrap

# --- Step 1. Write the training script ---------------------------------
train_script = textwrap.dedent("""\
    #!/usr/bin/env python3
    import os, json
    from datasets import load_dataset
    from transformers import AutoTokenizer, AutoModelForCausalLM, Trainer, TrainingArguments, DataCollatorForSeq2Seq

    MODEL_NAME = "sshleifer/tiny-gpt2"   # small CPU model
    OUTPUT_DIR = "/tmp/finetune-output"

    print("Loading dataset slice...")
    ds = load_dataset("ag_news", split="train[:0.1%]")
    tokenizer = AutoTokenizer.from_pretrained(MODEL_NAME)
    model = AutoModelForCausalLM.from_pretrained(MODEL_NAME)

    def preprocess(batch):
        enc = tokenizer(batch["text"], truncation=True, padding="max_length", max_length=128)
        enc["labels"] = enc["input_ids"].copy()
        return enc

    tok_ds = ds.map(preprocess, batched=True, remove_columns=ds.column_names)
    collator = DataCollatorForSeq2Seq(tokenizer, return_tensors="pt")

    args = TrainingArguments(
        output_dir=OUTPUT_DIR,
        num_train_epochs=1,
        per_device_train_batch_size=1,
        save_strategy="no",
        logging_steps=10
    )

    trainer = Trainer(model=model, args=args, train_dataset=tok_ds, data_collator=collator)
    print("Starting training on CPU...")
    trainer.train()
    trainer.save_model(OUTPUT_DIR)
    print("Saved model to", OUTPUT_DIR)
    print(json.dumps(os.listdir(OUTPUT_DIR)))
""")

with open("train.py", "w") as f:
    f.write(train_script)
print("✅ Wrote train.py")

# --- Step 2. Write the Kubernetes Job manifest --------------------------
job_yaml = textwrap.dedent("""\
    apiVersion: batch/v1
    kind: Job
    metadata:
      name: cpu-finetune-job
    spec:
      backoffLimit: 0
      template:
        spec:
          restartPolicy: Never
          containers:
            - name: finetuner
              image: python:3.10-slim
              command: ["/bin/bash", "-c"]
              args:
                - |
                  set -e
                  apt-get update && apt-get install -y git build-essential wget
                  pip install --upgrade pip
                  pip install torch --index-url https://download.pytorch.org/whl/cpu
                  pip install transformers datasets accelerate sentencepiece
                  python /workspace/train.py
              volumeMounts:
                - name: workspace
                  mountPath: /workspace
              resources:
                requests:
                  cpu: "4"
                  memory: "8Gi"
                limits:
                  cpu: "16"
                  memory: "32Gi"
          volumes:
            - name: workspace
              configMap:
                name: finetune-script
""")

with open("job.yaml", "w") as f:
    f.write(job_yaml)
print("✅ Wrote job.yaml")

# --- Step 3. Create ConfigMap + apply Job -------------------------------
def run(cmd):
    print(">", " ".join(cmd))
    subprocess.run(cmd, check=True)

try:
    run(["kubectl", "create", "configmap", "finetune-script",
         "--from-file=train.py", "--dry-run=client", "-o", "yaml"])
except FileNotFoundError:
    print("⚠️  kubectl not found — just created the files. Apply manually with:")
    print("   kubectl create configmap finetune-script --from-file=train.py")
    print("   kubectl apply -f job.yaml")
    raise SystemExit

# Pipe create→apply
create = subprocess.Popen(
    ["kubectl", "create", "configmap", "finetune-script",
     "--from-file=train.py", "--dry-run=client", "-o", "yaml"],
    stdout=subprocess.PIPE)
apply = subprocess.Popen(["kubectl", "apply", "-f", "-"], stdin=create.stdout)
apply.wait()

run(["kubectl", "apply", "-f", "job.yaml"])

print("""
✅ Job submitted.

To monitor:
  kubectl get pods -l job-name=cpu-finetune-job
  kubectl logs -f $(kubectl get pod -l job-name=cpu-finetune-job -o name)
""")
