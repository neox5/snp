#!/usr/bin/env bash

# Binary size analysis for Go projects
# Run from repository root

set -euo pipefail

BINARY="snp-analysis"

# Cleanup on exit
trap 'rm -f "$BINARY" "${BINARY}-stripped"' EXIT

echo "=== Go Binary Size Analysis ==="
echo

# Verify we're in the right place
if [ ! -f "go.mod" ]; then
    echo "Error: go.mod not found. Run from repository root."
    exit 1
fi

MODULE=$(go list -m)
echo "Module: $MODULE"
echo

# ============================================================================
# Step 1: Build unstripped binary
# ============================================================================
echo "Step 1: Building unstripped binary"
echo "---"
if ! go build -o "$BINARY" ./cmd/snp 2>&1; then
    echo "Build failed"
    exit 1
fi
echo "✓ Build complete"
echo

# ============================================================================
# Step 2: Overall size
# ============================================================================
echo "Step 2: Total binary size"
echo "---"
SIZE_BYTES=$(stat -f%z "$BINARY" 2>/dev/null || stat -c%s "$BINARY")
SIZE_MB=$(awk "BEGIN {printf \"%.2f\", $SIZE_BYTES / 1048576}")
echo "Size: $SIZE_MB MB ($SIZE_BYTES bytes)"
echo

# ============================================================================
# Step 3: Section breakdown
# ============================================================================
echo "Step 3: Binary sections"
echo "---"
if command -v size >/dev/null 2>&1; then
    size "$BINARY"
    echo
    
    # Parse and calculate percentages
    SECTIONS=$(size "$BINARY" | tail -1)
    TEXT=$(echo "$SECTIONS" | awk '{print $1}')
    DATA=$(echo "$SECTIONS" | awk '{print $2}')
    BSS=$(echo "$SECTIONS" | awk '{print $3}')
    
    awk -v text="$TEXT" -v data="$DATA" -v bss="$BSS" -v total="$SIZE_BYTES" 'BEGIN {
        printf "  text (code+rodata): %7.2f MB (%5.1f%%)\n", text/1048576, text*100/total
        printf "  data (initialized):  %7.2f MB (%5.1f%%)\n", data/1048576, data*100/total
        printf "  bss  (zero-init):    %7.2f MB (%5.1f%%)\n", bss/1048576, bss*100/total
    }'
else
    echo "Command 'size' not available"
fi
echo

# ============================================================================
# Step 4: Top packages by symbol size
# ============================================================================
echo "Step 4: Symbol size by package (top 20)"
echo "---"
go tool nm -size "$BINARY" 2>/dev/null | awk '
    NF == 4 {
        size = $2
        symbol = $4
        
        # Extract package name (before first dot or opening paren)
        if (match(symbol, /^([a-zA-Z0-9_\/.-]+)\./, arr)) {
            pkg = arr[1]
            pkgs[pkg] += size
            total += size
        }
    }
    END {
        # Sort by size descending
        for (pkg in pkgs) {
            sizes[pkgs[pkg]] = sizes[pkgs[pkg]] ? sizes[pkgs[pkg]] " " pkg : pkg
        }
        
        n = asorti(sizes, sorted, "@ind_num_desc")
        for (i = 1; i <= n && i <= 20; i++) {
            size = sorted[i]
            split(sizes[size], pkg_list, " ")
            for (j in pkg_list) {
                pkg = pkg_list[j]
                if (pkgs[pkg] == size) {
                    printf "%8.2f MB  (%5.1f%%)  %s\n", size/1048576, size*100/total, pkg
                    break
                }
            }
        }
    }
'
echo

# ============================================================================
# Step 5: Largest individual symbols
# ============================================================================
echo "Step 5: Largest symbols (top 20)"
echo "---"
go tool nm -size "$BINARY" 2>/dev/null | awk '
    NF == 4 {
        size = $2
        symbol = $4
        sizes[NR] = size
        symbols[NR] = symbol
        count = NR
    }
    END {
        # Bubble sort to find top 20
        for (i = 1; i <= count && i <= 20; i++) {
            max_idx = i
            for (j = i + 1; j <= count; j++) {
                if (sizes[j] > sizes[max_idx]) {
                    max_idx = j
                }
            }
            # Swap
            if (max_idx != i) {
                tmp_size = sizes[i]
                tmp_sym = symbols[i]
                sizes[i] = sizes[max_idx]
                symbols[i] = symbols[max_idx]
                sizes[max_idx] = tmp_size
                symbols[max_idx] = tmp_sym
            }
            printf "%8.2f MB  %s\n", sizes[i]/1048576, symbols[i]
        }
    }
' || echo "Failed to analyze symbols"
echo

# ============================================================================
# Step 6: Dependencies
# ============================================================================
echo "Step 6: Dependencies"
echo "---"

# Get all dependencies
ALL_DEPS=$(go list -deps ./cmd/snp 2>/dev/null)
TOTAL=$(echo "$ALL_DEPS" | wc -l | tr -d ' ')

# Count stdlib (no slash in import path)
STDLIB=$(echo "$ALL_DEPS" | grep -v "/" | wc -l | tr -d ' ')

# Count external (non-standard)
EXTERNAL=$(go list -f '{{if not .Standard}}{{.ImportPath}}{{end}}' -deps ./cmd/snp 2>/dev/null | wc -l | tr -d ' ')

echo "Total packages: $TOTAL"
echo "  Standard library: $STDLIB"
echo "  External: $EXTERNAL"
echo

# ============================================================================
# Step 7: External dependencies
# ============================================================================
echo "Step 7: External dependencies"
echo "---"
go list -f '{{if not .Standard}}{{.ImportPath}}{{end}}' -deps ./cmd/snp 2>/dev/null | sort | grep -v "^$MODULE"
echo

# ============================================================================
# Step 8: Stripped vs unstripped
# ============================================================================
echo "Step 8: Impact of stripping"
echo "---"
go build -ldflags="-s -w" -o "${BINARY}-stripped" ./cmd/snp 2>/dev/null

STRIPPED_BYTES=$(stat -f%z "${BINARY}-stripped" 2>/dev/null || stat -c%s "${BINARY}-stripped")
STRIPPED_MB=$(awk "BEGIN {printf \"%.2f\", $STRIPPED_BYTES / 1048576}")
SAVED=$(awk "BEGIN {printf \"%.2f\", ($SIZE_BYTES - $STRIPPED_BYTES) / 1048576}")
SAVED_PCT=$(awk "BEGIN {printf \"%.1f\", ($SIZE_BYTES - $STRIPPED_BYTES) * 100 / $SIZE_BYTES}")

echo "Unstripped: $SIZE_MB MB"
echo "Stripped:   $STRIPPED_MB MB"
echo "Saved:      $SAVED MB ($SAVED_PCT%)"
echo

echo "=== Analysis Complete ==="
