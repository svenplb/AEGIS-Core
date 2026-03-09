"""
Export the HuggingFace tokenizer to a standalone tokenizer.json.

The Go tokenizer in internal/nlp/tokenizer.go loads this file directly.
No ONNX export needed — the Go implementation handles tokenization natively.

Usage:
    python export_tokenizer.py --model_dir ./ner-model --output_dir ./models

This is a convenience script for cases where you need to extract just
the tokenizer without the full model export pipeline.
"""

import argparse
import json
import os

from transformers import AutoTokenizer


def parse_args():
    parser = argparse.ArgumentParser(description="Export tokenizer for Go loader")
    parser.add_argument(
        "--model_dir", type=str, default="xlm-roberta-base",
        help="Model name or path (default: xlm-roberta-base)",
    )
    parser.add_argument(
        "--output_dir", type=str, default="./models",
        help="Output directory for tokenizer.json",
    )
    return parser.parse_args()


def main():
    args = parse_args()
    os.makedirs(args.output_dir, exist_ok=True)

    print(f"Loading tokenizer from {args.model_dir}...")
    tokenizer = AutoTokenizer.from_pretrained(args.model_dir)

    # Save the full tokenizer (produces tokenizer.json among other files).
    tokenizer.save_pretrained(args.output_dir)

    # Verify tokenizer.json exists and inspect it.
    tok_path = os.path.join(args.output_dir, "tokenizer.json")
    if os.path.exists(tok_path):
        size_kb = os.path.getsize(tok_path) / 1024
        print(f"  tokenizer.json: {size_kb:.0f} KB")

        with open(tok_path) as f:
            tok_data = json.load(f)

        model_type = tok_data.get("model", {}).get("type", "unknown")
        vocab_size = len(tok_data.get("model", {}).get("vocab", []))
        print(f"  Model type: {model_type}")
        print(f"  Vocab size: {vocab_size}")

        # Show special tokens.
        added = tok_data.get("added_tokens", [])
        special = [t for t in added if t.get("special", False)]
        print(f"  Special tokens: {[t['content'] for t in special]}")
    else:
        print("  WARNING: tokenizer.json not found!")

    print(f"\nDone! tokenizer.json saved to {args.output_dir}/")


if __name__ == "__main__":
    main()
