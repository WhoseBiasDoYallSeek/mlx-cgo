#!/usr/bin/env bash
# regenerate.sh — Regenerate all C binding files using the Go mlxgen tool.
#
# Usage:
#   ./scripts/regenerate.sh [--mlx-source <path>]
#
# Options:
#   --mlx-source <path>   Path to the upstream MLX source tree (needed for
#                         C header generation from C++ headers). If not set,
#                         only the static generators (vector, map, closure,
#                         private) are run.
#
# Prerequisites:
#   - Go 1.21+
#   - clang (Apple Clang or LLVM clang on PATH)
#   - clang-format on PATH (or available via xcrun on macOS)
#
# The generated files are formatted with clang-format using the repo's
# .clang-format config before being written to disk.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
MLX_SOURCE=""

while [[ $# -gt 0 ]]; do
  case $1 in
    --mlx-source)
      MLX_SOURCE="$2"
      shift 2
      ;;
    *)
      echo "Unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

# Find clang-format
CFMT="${CLANG_FORMAT:-}"
if [[ -z "$CFMT" ]]; then
  if command -v clang-format &>/dev/null; then
    CFMT="clang-format"
  elif command -v xcrun &>/dev/null; then
    CFMT="$(xcrun --find clang-format 2>/dev/null)" || true
  fi
fi
if [[ -z "$CFMT" ]]; then
  echo "error: clang-format not found. Install it or set CLANG_FORMAT env var." >&2
  exit 1
fi

# Build mlxgen
echo "Building mlxgen..."
cd "$REPO_ROOT"
go build -o /tmp/mlxgen ./cmd/mlxgen

fmt_write() {
  local outfile="$1"
  local assume
  assume="$(basename "$outfile")"
  /tmp/mlxgen "${@:2}" | "$CFMT" --style=file --assume-filename="$assume" > "$outfile"
  echo "  wrote $outfile"
}

# ── Static generators ─────────────────────────────────────────────────────────
echo "Generating static files..."

fmt_write "$REPO_ROOT/mlx/c/vector.h"   --type=vector
fmt_write "$REPO_ROOT/mlx/c/vector.cpp" --type=vector --implementation
fmt_write "$REPO_ROOT/mlx/c/map.h"      --type=map
fmt_write "$REPO_ROOT/mlx/c/map.cpp"    --type=map    --implementation
fmt_write "$REPO_ROOT/mlx/c/closure.h"  --type=closure
fmt_write "$REPO_ROOT/mlx/c/closure.cpp" --type=closure --implementation
fmt_write "$REPO_ROOT/mlx/c/private/closure.h" --type=closure-private

# Private headers — one per opaque type.
# Each entry is: ctype|cpptype|nocopy|using
# Using '|' as separator to avoid conflicts with '::' in C++ types.
declare -a PRIVATE_TYPES=(
  "mlx_array|mlx::core::array||"
  "mlx_stream|mlx::core::Stream||"
  "mlx_device;mlx_device_info|mlx::core::Device;std::unordered_map<std::string, std::variant<std::string, size_t>>||;mlx_device_info_cpp"
  "mlx_closure_value_and_grad|std::function<std::pair<std::vector<mlx::core::array>, std::vector<mlx::core::array>>(const std::vector<mlx::core::array>&)>||"
  "mlx_closure_custom|std::function<std::vector<mlx::core::array>(std::vector<mlx::core::array>,std::vector<mlx::core::array>,std::vector<mlx::core::array>)>||"
  "mlx_closure_custom_jvp|std::function<std::vector<mlx::core::array>(std::vector<mlx::core::array>,std::vector<mlx::core::array>,std::vector<int>)>||"
  "mlx_closure_custom_vmap|std::function<std::pair<std::vector<mlx::core::array>, std::vector<int>>(std::vector<mlx::core::array>,std::vector<int>)>||"
  "mlx_vector_array|std::vector<mlx::core::array>||"
  "mlx_vector_int|std::vector<int>||"
  "mlx_vector_string|std::vector<std::string>||"
  "mlx_map_string_to_array|std::unordered_map<std::string, mlx::core::array>||"
  "mlx_map_string_to_string|std::unordered_map<std::string, std::string>||"
  "mlx_distributed_group|mlx::core::distributed::Group|no-copy|"
)

echo "Generating private headers..."
for entry in "${PRIVATE_TYPES[@]}"; do
  IFS='|' read -r ctype cpptype nocopy using <<< "$entry"
  # Use only the first ctype (before ';') to derive the output filename
  first_ctype="${ctype%%;*}"
  shortname="${first_ctype#mlx_}"
  outfile="$REPO_ROOT/mlx/c/private/${shortname}.h"
  args=(--type=private --ctype="$ctype" --cpptype="$cpptype")
  [[ "$nocopy" == "no-copy" ]] && args+=(--no-copy)
  [[ -n "$using" ]] && args+=(--using="$using")
  fmt_write "$outfile" "${args[@]}"
done

# ── C++ header-driven generators (require MLX source) ────────────────────────
if [[ -z "$MLX_SOURCE" ]]; then
  echo ""
  echo "Skipping C++ header-based generators (--mlx-source not provided)."
  echo ""
  echo "To regenerate C bindings from MLX C++ headers, first build the project"
  echo "(which fetches MLX via FetchContent), then run:"
  echo ""
  echo "  ./scripts/regenerate.sh --mlx-source <build_dir>/_deps/mlx-src/mlx"
  echo ""
  exit 0
fi

echo ""
echo "Generating C bindings from $MLX_SOURCE..."

# Each entry: headername|input_header(s)|docstring
# Using indexed array + pipe-separated fields for bash 3.x compatibility.
dynamic_headers=(
  "ops|$MLX_SOURCE/ops.h;$MLX_SOURCE/einsum.h|Core array operations"
  "fft|$MLX_SOURCE/fft.h|FFT operations"
  "linalg|$MLX_SOURCE/linalg.h|Linear algebra operations"
  "random|$MLX_SOURCE/random.h|Random number operations"
  "fast|$MLX_SOURCE/fast.h|Fast custom operations"
  "transforms|$MLX_SOURCE/transforms.h|Transform operations"
  "compile|$MLX_SOURCE/compile.h;$MLX_SOURCE/compile_impl.h|Compilation operations"
  "distributed|$MLX_SOURCE/distributed/ops.h|Distributed collectives"
  "metal|$MLX_SOURCE/backend/metal/metal.h|Metal specific operations"
  "cuda|$MLX_SOURCE/backend/cuda/cuda.h|Cuda specific operations"
  "graph_utils|$MLX_SOURCE/graph_utils.h|Graph Utils"
  "io|$MLX_SOURCE/io.h|IO operations"
  "export|$MLX_SOURCE/export.h|Function serialization"
)

include_root="$(dirname "$MLX_SOURCE")"

for entry in "${dynamic_headers[@]}"; do
  IFS='|' read -r headername input docstring <<< "$entry"
  echo "  $headername..."
  doc_args=()
  [[ -n "$docstring" ]] && doc_args=(--docstring="$docstring")
  fmt_write "$REPO_ROOT/mlx/c/${headername}.h" \
    --header="$input" --headername="$headername" --include-dir="$include_root" "${doc_args[@]}"
  fmt_write "$REPO_ROOT/mlx/c/${headername}.cpp" \
    --header="$input" --headername="$headername" --implementation --include-dir="$include_root"
done

echo ""
echo "Done."
