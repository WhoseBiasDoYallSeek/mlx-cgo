# mlx-cgo: Maintenance & Upstream Sync

## Overview

mlx-cgo is a **thin wrapper** over [ml-explore/mlx-c](https://github.com/ml-explore/mlx-c).

This document covers:
- How to sync with upstream MLX-C updates
- When to update generated C++ code
- Versioning strategy
- CI/CD automation

---

## Versioning Strategy

### mlx-cgo Version Format
```
v{MAJOR}.{MINOR}.{PATCH}
v1.0.0
  ↑     ↑    ↑
  |     |    └─ Patch: Bug fixes, small improvements
  |     └────── Minor: New features, new C++ files
  └──────────── Major: Breaking changes, incompatible APIs
```

### MLX-C Dependency
```
mlx-cgo v1.y.z → MLX v0.31.2 (locked)
mlx-cgo v2.y.z → MLX v0.32.x (when available)
```

**Policy**: Each mlx-cgo major version locks to one MLX major version.

---

## Monitoring Upstream

### Step 1: Watch ml-explore/mlx-c Releases

```bash
# Manual check
open https://github.com/ml-explore/mlx-c/releases

# Or via GitHub CLI:
gh release list --repo ml-explore/mlx-c
```

### Step 2: Check What Changed

```bash
# Compare against your locked version
git fetch --tags origin
git log v0.31.2..v0.32.0 --oneline -- mlx/c/

# Or diff headers:
git diff v0.31.2..v0.32.0 -- mlx/c/include/mlx/c/
```

### Step 3: Assess Impact

| Change | Impact | Action |
|--------|--------|--------|
| New header file | MEDIUM | Add to generators, regenerate |
| New function | LOW | Auto-generated, needs testing |
| Function signature change | HIGH | May break existing code, manual review |
| Type definition change | HIGH | May affect generated Go wrappers |
| Internal refactor (no API change) | LOW | Regenerate, verify binary identical |

---

## Update Process (Manual)

### Phase 1: Prepare Update Branch

```bash
# Create update branch
git checkout -b update/mlx-v0.32.0

# Update MLX submodule
cd build/_deps/mlx-src
git fetch --all
git checkout v0.32.0
cd ../..

# Commit submodule update
git add build/_deps/mlx-src
git commit -m "Update MLX-C to v0.32.0"
```

### Phase 2: Regenerate Bindings

```bash
# Build generator
go build ./cmd/mlxgen

# Regenerate all C++ files
./scripts/regenerate.sh

# Check what changed
git status mlx/c/

# Review changes (sample ~10 files)
git diff mlx/c/ops.h | head -100
git diff mlx/c/ops.cpp | head -100
```

### Phase 3: Test

```bash
# Compile C++ code
cmake -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build -j

# Run all tests
go test ./... -v -race

# Check binary size (should be similar)
ls -lh build/mlx/c/*.o | awk '{sum += $5} END {print "Total:", sum}'

# Compare binary sizes before/after
# Should be within ±5%
```

### Phase 4: Check for Breaking Changes

```bash
# If any test fails:
# 1. Check if MLX header changed API
# 2. Review git diff
# 3. Update generator if needed

# Example: If function signature changed:
git log v0.31.2..v0.32.0 --oneline -- mlx/c/include/mlx/c/ops.h

# Check mlx-cgo codegen
grep -r "mlx_array_" internal/codegen/ | head -5
```

### Phase 5: Commit & Tag

```bash
# If all tests pass:
git add -A
git commit -m "Sync MLX-C to v0.32.0

- Updated submodule to v0.32.0
- Regenerated 20+ C++ binding files
- All 71 tests passing
- Binary size ±3% (baseline)"

# Tag new version
git tag -a v1.1.0 -m "mlx-cgo v1.1.0: MLX-C v0.32.0 support"

# Push
git push origin update/mlx-v0.32.0
git push origin v1.1.0
```

### Phase 6: GitHub Release

Create release notes:
```markdown
## mlx-cgo v1.1.0

### New
- Support for MLX-C v0.32.0
- 3 new functions (cumsum, attention, custom_kernel v2)

### Fixed
- Generator handles new `mlx_function_exporter` type
- Export.h now properly generated for new types

### Changed
- Minimum Clang version still 14
- Minimum Go version still 1.19

### Breaking
- None (backwards compatible)

### Docs
- Updated ARCHITECTURE.md with new APIs
- Added examples/cumsum_example.go

See [CHANGELOG.md](./CHANGELOG.md) for full details.
```

---

## Automated Update Strategy (CI/CD)

### Proposed GitHub Actions Workflow

Create `.github/workflows/upstream-sync.yml`:

```yaml
name: Upstream MLX-C Sync

on:
  schedule:
    # Check weekly for updates
    - cron: '0 0 * * MON'
  workflow_dispatch:  # Manual trigger

jobs:
  check-upstream:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - name: Check MLX-C releases
        id: check
        run: |
          CURRENT=$(cd build/_deps/mlx-src && git describe --tags)
          LATEST=$(curl -s https://api.github.com/repos/ml-explore/mlx-c/releases/latest | jq -r .tag_name)
          
          echo "current=$CURRENT" >> $GITHUB_OUTPUT
          echo "latest=$LATEST" >> $GITHUB_OUTPUT
          
          if [ "$CURRENT" != "$LATEST" ]; then
            echo "UPDATE_AVAILABLE=true" >> $GITHUB_OUTPUT
          else
            echo "UPDATE_AVAILABLE=false" >> $GITHUB_OUTPUT
          fi

      - name: Create update PR
        if: steps.check.outputs.UPDATE_AVAILABLE == 'true'
        run: |
          git config user.name "mlx-cgo-bot"
          git config user.email "bot@mlx-cgo.local"
          
          LATEST=${{ steps.check.outputs.latest }}
          git checkout -b update/mlx-${LATEST}
          
          # Update submodule
          cd build/_deps/mlx-src
          git fetch --all
          git checkout ${LATEST}
          cd ../..
          
          git add build/_deps/mlx-src
          git commit -m "chore: Update MLX-C to ${LATEST}"
          
          git push origin update/mlx-${LATEST}
          
          # Create PR (requires GH_TOKEN)
          gh pr create \
            --title "chore: Update MLX-C to ${LATEST}" \
            --body "Automated upstream sync" \
            --base main \
            --head update/mlx-${LATEST}
```

### CI Pipeline for Update PR

When update PR is created, run:

```yaml
  test-update:
    runs-on: macos-latest
    needs: check-upstream
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Build generator
        run: go build ./cmd/mlxgen

      - name: Regenerate bindings
        run: ./scripts/regenerate.sh

      - name: Run tests
        run: go test ./... -v -race

      - name: Check binary sizes
        run: |
          echo "Generated file sizes:"
          du -sh mlx/c/ mlx/c/private/
          
          # Warn if >10% growth
          # Could fail PR if too large

      - name: Generate diff report
        run: |
          echo "## Changes in MLX-C update" >> $GITHUB_STEP_SUMMARY
          echo '```' >> $GITHUB_STEP_SUMMARY
          git diff --stat HEAD~1 mlx/c/ >> $GITHUB_STEP_SUMMARY
          echo '```' >> $GITHUB_STEP_SUMMARY
```

---

## 10 Problems: Maintenance Perspective

### Problem #1: Empty Array Panic
**Status**: ✅ FIXED in v1.0  
**Upstream**: MLX doesn't change slice handling, but if they add new array functions:
```bash
./mlxgen --all  # Will auto-generate with guards
go test ./...   # Verify safe
```

### Problem #2: C String Memory Leak
**Status**: ✅ FIXED in v1.0  
**Maintenance**: Every string param gets `defer C.free()`
**Test**: `grep "defer C.free" mlx/c/*.go | wc -l` should grow with updates

### Problem #3: Apple Silicon Only
**Status**: ✅ INTENTIONAL  
**No change needed** — MLX is Apple-first too

### Problem #4: Clang Version Mismatch
**Status**: Documented in TROUBLESHOOTING.md  
**CI Check**: Add to GitHub Actions:
```bash
clang --version | grep -q "14\|15\|16\|17" || exit 1
```

### Problem #5: MLX Version Mismatch
**Status**: This document (lock to v0.31.2)  
**CI Lock**: CMakeLists.txt has:
```cmake
ExternalProject_Add(mlx
  GIT_REPOSITORY https://github.com/ml-explore/mlx.git
  GIT_TAG v0.31.2  # Locked here
)
```

### Problem #6: Python Artifacts
**Status**: ✅ Deleted  
**Maintenance**: Verify stays deleted:
```bash
[ ! -d python ] || exit 1  # CI check
```

### Problem #7: Concurrency Race Conditions
**Status**: User responsibility (sync.Mutex pattern)  
**CI Test**: `go test -race` catches these

### Problem #8: Type Mismatches
**Status**: Go's type system (user converts explicitly)  
**No maintenance needed**

### Problem #9: Forgot to Free
**Status**: User responsibility (defer pattern)  
**CI Check**: `go test -race` detects leaks sometimes

### Problem #10: Stale Generated Code
**Status**: `scripts/regenerate.sh` fixes this  
**CI Prevention**: Run regenerate, commit if changed
```bash
./scripts/regenerate.sh
git diff --exit-code mlx/c/  # Fail if any diff
```

---

## Maintenance Checklist

### Monthly
- [ ] Check for new MLX-C releases (`gh release list --repo ml-explore/mlx-c`)
- [ ] Review any breaking change announcements

### When MLX-C Updates
- [ ] Create update branch
- [ ] Run `./scripts/regenerate.sh`
- [ ] Verify all tests pass (`go test -race`)
- [ ] Review C++ diffs (100+ line changes)
- [ ] Update CHANGELOG.md
- [ ] Tag new version
- [ ] Create GitHub release

### Quarterly
- [ ] Audit dependencies (go mod tidy)
- [ ] Check Go version (update if new LTS)
- [ ] Check Clang support (Xcode updates)
- [ ] Review GitHub Issues

---

## Troubleshooting Sync Issues

### Issue: Generator Can't Parse New Header
```bash
# Symptom: "couldn't find mlx_array_cumsum"
# Cause: New function in v0.32.0

# Fix:
git log v0.31.2..v0.32.0 -- mlx/c/include/mlx/c/ops.h | head -20
# See what changed in header

# Update variants.go if new function needs special handling
# Then regenerate:
./scripts/regenerate.sh
```

### Issue: Test Fails After Update
```bash
# Symptom: "TestFmt_Ops: expected X, got Y"
# Cause: Generated code differs from golden file

# Fix:
# 1. Check if test golden file needs update
# 2. If MLX API changed legitimately, update golden
# 3. If generator bug, fix generator
git diff internal/codegen/testdata/
```

### Issue: Binary Grew 50%
```bash
# Symptom: mlx/c/ops.o was 5MB, now 8MB
# Cause: Lots of new inline functions in headers

# Check:
du -sh mlx/c/*.o
# If significant growth, review MLX release notes
# May indicate new ops, normal growth
```

---

## Long-term Maintenance Strategy

### Year 1 (2026)
- Lock to MLX v0.31.2 (current stable)
- Monthly sync checks
- Fix bugs as reported
- Build community (README examples, etc)

### Year 2 (2027)
- Monitor MLX v0.32, v0.33 releases
- Plan upgrade path (v2.0.0 for v0.32+)
- Gather user feedback
- Optimize performance

### Year 3+ (2028+)
- Keep synchronized within major version (v1.x → all MLX v0.31.y)
- Support 2 major versions (e.g., v1.x and v2.x simultaneously)
- Community contributors take on maintenance

---

## Contributing Updates

If someone wants to port mlx-cgo to new MLX version:

1. **Create fork**: `git clone https://github.com/WhoseBiasDoYallSeek/mlx-cgo.git`
2. **Branch**: `git checkout -b update/mlx-v0.32.0`
3. **Update**: Follow steps 1-4 above
4. **Test**: All tests must pass
5. **PR**: Submit with detailed test results
6. **Review**: I'll verify binary safety + performance
7. **Merge**: Once approved

---

## Summary

**mlx-cgo sync strategy:**

✅ **Manual**: Update monthly, lock to MLX v0.31.2  
✅ **Automated**: CI detects updates, suggests PR  
✅ **Versioned**: v1.x = MLX v0.31.y, v2.x = MLX v0.32.y  
✅ **Tested**: All 71 tests must pass before merge  
✅ **Documented**: CHANGELOG.md tracks all updates  

**You stay synchronized with upstream MLX-C, but control the pace.**
