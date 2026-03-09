"""
Export the fine-tuned NER model to ONNX format with INT8 quantization.

Usage:
    python export_onnx.py --model_dir ./ner-model --output_dir ./output

Produces:
    output/ner.onnx           — optimized ONNX model
    output/tokenizer.json     — HuggingFace tokenizer config (for Go loader)
"""

import argparse
import os
import shutil

from optimum.onnxruntime import ORTModelForTokenClassification, ORTQuantizer
from optimum.onnxruntime.configuration import AutoQuantizationConfig
from transformers import AutoTokenizer


def parse_args():
    parser = argparse.ArgumentParser(description="Export NER model to ONNX")
    parser.add_argument("--model_dir", type=str, required=True, help="Path to fine-tuned model")
    parser.add_argument("--output_dir", type=str, default="./output", help="Output directory")
    parser.add_argument("--no_quantize", action="store_true", help="Skip INT8 quantization")
    return parser.parse_args()


def main():
    args = parse_args()
    os.makedirs(args.output_dir, exist_ok=True)

    print(f"Loading model from {args.model_dir}...")
    tokenizer = AutoTokenizer.from_pretrained(args.model_dir)

    print("Exporting to ONNX...")
    ort_model = ORTModelForTokenClassification.from_pretrained(
        args.model_dir,
        export=True,
    )
    ort_model.save_pretrained(args.output_dir)

    if not args.no_quantize:
        print("Applying INT8 dynamic quantization...")
        try:
            quantizer = ORTQuantizer.from_pretrained(args.output_dir)
            qconfig = AutoQuantizationConfig.avx2(is_static=False, per_channel=False)
            quantizer.quantize(save_dir=args.output_dir, quantization_config=qconfig)
            print("Quantization complete.")
        except Exception as e:
            print(f"WARNING: Quantization failed ({e}), using unquantized model.")

    tokenizer.save_pretrained(args.output_dir)

    # Rename ONNX model to ner.onnx.
    onnx_files = [f for f in os.listdir(args.output_dir) if f.endswith(".onnx")]
    if onnx_files:
        quantized = [f for f in onnx_files if "quantized" in f]
        src = quantized[0] if quantized else onnx_files[0]
        dst = os.path.join(args.output_dir, "ner.onnx")
        src_path = os.path.join(args.output_dir, src)
        if src_path != dst:
            shutil.move(src_path, dst)
        print(f"Model saved as: {dst}")

    # Verify output.
    print("\nOutput files:")
    for f in ["ner.onnx", "tokenizer.json"]:
        path = os.path.join(args.output_dir, f)
        if os.path.exists(path):
            size_mb = os.path.getsize(path) / (1024 * 1024)
            print(f"  {f}: {size_mb:.1f} MB")
        else:
            print(f"  WARNING: {f} not found!")

    print(f"\nDone! Send these two files from {args.output_dir}/:")
    print("  - ner.onnx")
    print("  - tokenizer.json")


if __name__ == "__main__":
    main()
