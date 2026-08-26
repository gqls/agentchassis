#!/usr/bin/env python3
"""Build and VALIDATE a sample dataset for the finetuning.uk playground.

WHY THIS EXISTS
---------------
The phase-0 run (RESULTS_2026-08-15) uploaded 300 rows and trained **295**. Five
were dropped by the trainer's response-marker filter because they were too long,
and the drop is reported in the trainer log rather than raised — so a dataset can
lose rows *after* the GPU is paid for, and the only signal is a row count in a
stage summary nobody diffs. Everything this script refuses, it refuses BEFORE the
upload, which is the only point where refusing is free.

It is deliberately strict about one thing the trainer is quiet about: a row whose
assistant turn is empty or whitespace trains the model to say nothing, and reads
in every dashboard as a perfectly good row.

FORMAT
------
One JSON object per line, chat shape, exactly one user turn and one assistant
turn (what the phase-0 dataset carries, checked 300/300):

    {"messages": [{"role": "user", "content": "..."},
                  {"role": "assistant", "content": "..."}]}

INPUT
-----
A dataset directory containing `pairs.jsonl` — the same shape, unsplit — plus a
`meta.json` naming the dataset, its task and its PROVENANCE. Provenance is
required and unvalidatable by machine, which is exactly why it is a required
field: it forces the author to write down whose words these are.

USAGE
    build_dataset.py <dataset-dir> [--heldout N] [--max-chars N] [--force]
    build_dataset.py --self-test

Writes `training.jsonl` and `heldout.jsonl`. Exit 1 on any refusal.
"""

import argparse
import json
import os
import random
import sys

# The phase-0 drop was a LENGTH filter. 12,000 characters is well inside what a
# 1.7B model's context takes at the batch shape that run used, and comfortably
# above the longest row that survived. Raise it only with a measurement.
DEFAULT_MAX_CHARS = 12_000
DEFAULT_HELDOUT = 10

REQUIRED_META = ("name", "task", "provenance", "provenance_detail")


class Refusal(Exception):
    """A reason not to upload. Every one of these is cheaper here than on a GPU."""


def load_pairs(path):
    rows = []
    with open(path, encoding="utf-8") as fh:
        for n, line in enumerate(fh, 1):
            line = line.strip()
            if not line:
                continue
            try:
                rows.append((n, json.loads(line)))
            except json.JSONDecodeError as exc:
                raise Refusal(f"{path}:{n}: not valid JSON — {exc}")
    if not rows:
        raise Refusal(f"{path}: no rows")
    return rows


def check_row(n, row, max_chars):
    """Return the row's character length, or raise with the line number."""
    msgs = row.get("messages")
    if not isinstance(msgs, list):
        raise Refusal(f"line {n}: no 'messages' list")
    roles = [m.get("role") for m in msgs]
    if roles != ["user", "assistant"]:
        # The trainer's response-marker filter needs the assistant turn to be
        # findable. Anything other than exactly user-then-assistant is a row
        # whose fate depends on a template, which is not a thing to guess at.
        raise Refusal(f"line {n}: roles are {roles}, want exactly ['user', 'assistant']")
    for m in msgs:
        content = m.get("content")
        if not isinstance(content, str) or not content.strip():
            raise Refusal(
                f"line {n}: the {m.get('role')} turn is empty or whitespace — "
                "this trains the model to say nothing and reads as a good row everywhere"
            )
    length = sum(len(m["content"]) for m in msgs)
    if length > max_chars:
        raise Refusal(
            f"line {n}: {length} chars > {max_chars} — this is the shape that "
            "silently lost 5 of 300 rows on the phase-0 run, AFTER the GPU was paid for"
        )
    return length


def check_meta(meta):
    missing = [k for k in REQUIRED_META if not str(meta.get(k, "")).strip()]
    if missing:
        raise Refusal(
            f"meta.json is missing {missing}. 'provenance' is required and cannot be "
            "machine-checked, which is why it is required: whose words these are is a "
            "question a person has to answer before anything is published."
        )
    allowed = {"our-own-material", "synthetic", "open-licensed", "customer-with-permission"}
    if meta["provenance"] not in allowed:
        raise Refusal(f"meta.json provenance={meta['provenance']!r}, want one of {sorted(allowed)}")


def build(dataset_dir, heldout_n, max_chars, force):
    meta_path = os.path.join(dataset_dir, "meta.json")
    if not os.path.exists(meta_path):
        raise Refusal(f"{meta_path} not found")
    with open(meta_path, encoding="utf-8") as fh:
        meta = json.load(fh)
    check_meta(meta)

    rows = load_pairs(os.path.join(dataset_dir, "pairs.jsonl"))
    lengths = [check_row(n, r, max_chars) for n, r in rows]

    if len(rows) <= heldout_n:
        raise Refusal(f"{len(rows)} rows cannot yield a {heldout_n}-row held-out set and a training set")

    # Deterministic split, seeded on the dataset name: a worked example must be
    # reproducible, and a random split would quietly change which rows the model
    # never saw every time this ran.
    order = list(range(len(rows)))
    random.Random(meta["name"]).shuffle(order)
    heldout_idx = set(order[:heldout_n])

    train_path = os.path.join(dataset_dir, "training.jsonl")
    held_path = os.path.join(dataset_dir, "heldout.jsonl")
    for p in (train_path, held_path):
        if os.path.exists(p) and not force:
            raise Refusal(f"{p} exists; pass --force to overwrite")

    with open(train_path, "w", encoding="utf-8") as tf, open(held_path, "w", encoding="utf-8") as hf:
        for i, (_, row) in enumerate(rows):
            out = hf if i in heldout_idx else tf
            out.write(json.dumps(row, ensure_ascii=False) + "\n")

    n_train = len(rows) - heldout_n
    print(f"{meta['name']}: {n_train} training + {heldout_n} held-out "
          f"(provenance: {meta['provenance']}; longest row {max(lengths)} chars, cap {max_chars})")
    return n_train, heldout_n


def self_test():
    """Every refusal is induced, because a validator nobody has seen fail is a
    validator nobody has seen."""
    import tempfile

    ok_row = {"messages": [{"role": "user", "content": "brief"},
                           {"role": "assistant", "content": "reply"}]}
    cases = [
        ("empty assistant turn",
         [{"messages": [{"role": "user", "content": "x"}, {"role": "assistant", "content": "   "}]}],
         "empty or whitespace"),
        ("wrong role order",
         [{"messages": [{"role": "assistant", "content": "a"}, {"role": "user", "content": "u"}]}],
         "want exactly"),
        ("over the length cap",
         [{"messages": [{"role": "user", "content": "x" * 20_000},
                        {"role": "assistant", "content": "y"}]}],
         "phase-0 run"),
        ("a third turn",
         [{"messages": [{"role": "user", "content": "u"}, {"role": "assistant", "content": "a"},
                        {"role": "user", "content": "u2"}]}],
         "want exactly"),
    ]
    failures = []
    for label, rows, needle in cases:
        with tempfile.TemporaryDirectory() as d:
            with open(os.path.join(d, "meta.json"), "w") as fh:
                json.dump({"name": "t", "task": "t", "provenance": "synthetic",
                           "provenance_detail": "self-test"}, fh)
            with open(os.path.join(d, "pairs.jsonl"), "w") as fh:
                for r in rows + [ok_row] * 12:
                    fh.write(json.dumps(r) + "\n")
            try:
                build(d, 2, DEFAULT_MAX_CHARS, True)
            except Refusal as exc:
                if needle not in str(exc):
                    failures.append(f"{label}: refused for the wrong reason — {exc}")
            else:
                failures.append(f"CONTROL FAILED: {label} was accepted")

    # And the positive control: a clean set must BUILD, or the refusals above
    # prove only that the script refuses everything.
    with tempfile.TemporaryDirectory() as d:
        with open(os.path.join(d, "meta.json"), "w") as fh:
            json.dump({"name": "t", "task": "t", "provenance": "our-own-material",
                       "provenance_detail": "self-test"}, fh)
        with open(os.path.join(d, "pairs.jsonl"), "w") as fh:
            for _ in range(14):
                fh.write(json.dumps(ok_row) + "\n")
        try:
            n_train, n_held = build(d, 2, DEFAULT_MAX_CHARS, True)
            if (n_train, n_held) != (12, 2):
                failures.append(f"CONTROL FAILED: clean split gave {n_train}/{n_held}, want 12/2")
        except Refusal as exc:
            failures.append(f"CONTROL FAILED: a clean dataset was refused — {exc}")

    # Provenance is required, and that is the one a hurried author will skip.
    with tempfile.TemporaryDirectory() as d:
        with open(os.path.join(d, "meta.json"), "w") as fh:
            json.dump({"name": "t", "task": "t"}, fh)
        with open(os.path.join(d, "pairs.jsonl"), "w") as fh:
            for _ in range(14):
                fh.write(json.dumps(ok_row) + "\n")
        try:
            build(d, 2, DEFAULT_MAX_CHARS, True)
        except Refusal as exc:
            if "provenance" not in str(exc):
                failures.append(f"missing provenance refused for the wrong reason — {exc}")
        else:
            failures.append("CONTROL FAILED: a dataset with no provenance was accepted")

    if failures:
        for f in failures:
            print("FAIL:", f, file=sys.stderr)
        return 1
    print("self-test OK: 4 refusals induced, 1 provenance refusal induced, 1 positive control built")
    return 0


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("dataset_dir", nargs="?")
    ap.add_argument("--heldout", type=int, default=DEFAULT_HELDOUT)
    ap.add_argument("--max-chars", type=int, default=DEFAULT_MAX_CHARS)
    ap.add_argument("--force", action="store_true")
    ap.add_argument("--self-test", action="store_true")
    args = ap.parse_args()

    if args.self_test:
        return self_test()
    if not args.dataset_dir:
        ap.error("give a dataset directory, or --self-test")
    try:
        build(args.dataset_dir, args.heldout, args.max_chars, args.force)
    except Refusal as exc:
        print(f"REFUSED: {exc}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
