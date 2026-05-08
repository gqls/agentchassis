# The terminology

### "Smoke speed isn't representative."

A smoke test is a deliberately tiny end-to-end run whose only goal is to prove the pipeline doesn't catch fire — name comes from electronics ("plug it in, see if smoke comes out"). 20 rows × 1 epoch is a smoke test. The speed observed during a smoke test is a bad predictor of full-run speed because one-time costs get spread over too few steps.

### "First-step compilation overhead (CUDA graph capture, kernel autotune)"

When PyTorch runs an operation on the GPU — say, a matrix multiplication of a particular size — it has several different kernels (low-level GPU programs) it could use. The first time it encounters that exact shape and dtype, it briefly runs each candidate kernel and picks the fastest. That's kernel autotune. The result gets cached, so step 2 onward skips the benchmarking. Add to that CUDA graph capture, where PyTorch records the sequence of GPU calls into a replay-able graph (which subsequent steps replay much faster than re-issuing every call). Add to that bitsandbytes' own dequantisation kernels which have their own autotune. Step 1 carries all of that overhead on top of the actual compute. Steps 50, 100, 500 don't.

### "Amortized across only 3 steps."

Amortize means spread a one-time cost across many uses. If startup overhead is, say, 60 seconds, that's 20s/step over 3 steps but 0.08s/step over 735. Smoke tests look much slower per step because the fixed cost gets divided by a tiny number.

### "Steady-state speed"

Speed once all the one-time costs have been paid and the loop is just doing the real work. You usually hit it after 5-20 steps.

### "FA2 / Flash Attention 2"

The attention layer in a transformer normally computes an N×N similarity matrix between every pair of tokens, where N is the sequence length. At N=4096 that's 16 million entries, written to and read from the GPU's main memory (HBM) which is the bottleneck. Flash Attention 2 is a hand-tuned reorganisation of the same math that streams through the computation tile-by-tile in the GPU's much faster on-chip SRAM, never materialising the full matrix in HBM. Result: roughly 2-4× faster attention plus lower memory use. It's a separate package because it's compiled against specific CUDA + torch + Python combinations.

### "Standard SDPA"

torch.nn.functional.scaled_dot_product_attention — PyTorch's built-in attention function. It has multiple internal backends: a naive math version, a memory-efficient version, and a "flash" version. On A100 without flash-attn installed, the best available backend is "memory-efficient" — better than naive but materially slower than dedicated FA2 on long sequences.

### "seq 4096"

Sequence length 4096 — the maximum number of tokens in any one training example. Attention compute scales as O(N²) in sequence length, so 4096 is roughly 4× more attention work than 2048. We use 4096 because some of our prompts (the ones with full-page briefs and accumulated context) approach 4000 tokens, and truncating them throws away the information we want the model to learn from.