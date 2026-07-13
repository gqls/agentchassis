#!/usr/bin/env python3
"""
Minimal CPU fine-tuning example for Hugging Face transformers
This runs quickly on CPU nodes for validation.
"""

from datasets import load_dataset
from transformers import (
    AutoTokenizer,
    AutoModelForCausalLM,
    Trainer,
    TrainingArguments,
    DataCollatorForSeq2Seq,
)
import json
import os

MODEL_NAME = "sshleifer/tiny-gpt2"  # small model that fits on CPU
OUTPUT_DIR = "/workspace/output"

print("Loading dataset slice...")
dataset = load_dataset("ag_news", split="train[:0.1%]")

print("Loading model and tokenizer...")
tokenizer = AutoTokenizer.from_pretrained(MODEL_NAME)
model = AutoModelForCausalLM.from_pretrained(MODEL_NAME)


def preprocess(batch):
    enc = tokenizer(batch["text"], truncation=True, padding="max_length", max_length=128)
    enc["labels"] = enc["input_ids"].copy()
    return enc


print("Tokenizing dataset...")
tokenized = dataset.map(preprocess, batched=True, remove_columns=dataset.column_names)

collator = DataCollatorForSeq2Seq(tokenizer, return_tensors="pt")

args = TrainingArguments(
    output_dir=OUTPUT_DIR,
    num_train_epochs=1,
    per_device_train_batch_size=1,
    logging_steps=10,
    save_strategy="no",
)

print("Starting CPU training...")
trainer = Trainer(model=model, args=args, train_dataset=tokenized, data_collator=collator)
trainer.train()

print("Saving model to", OUTPUT_DIR)
trainer.save_model(OUTPUT_DIR)

print("✅ Done! Output files:")
print(json.dumps(os.listdir(OUTPUT_DIR)))
