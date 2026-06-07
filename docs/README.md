# docs/

This directory contains technical documentation for mlx-cgo.

## Contents

- **[ARCHITECTURE.md](ARCHITECTURE.md)** — Complete architecture reference:
  - Python → Go toolchain rewrite (what was ported and why)
  - Directory structure and data flow
  - Domain model and type system
  - CGo generator design (type mappings, safety, memory model)
  - All 10 bugs found and fixed
  - Test coverage analysis
  - CMake integration
  - End-to-end validation results against MLX v0.31.2
  - Project status summary

## Other Documentation

| File | Description |
|---|---|
| [`../README.md`](../README.md) | Project overview, quick start, build instructions |
| [`../CONTRIBUTING.md`](../CONTRIBUTING.md) | How to contribute |
| [`../TROUBLESHOOTING.md`](../TROUBLESHOOTING.md) | Common issues and solutions (Apple Silicon focused) |
| [`../MAINTENANCE.md`](../MAINTENANCE.md) | Upstream sync procedures and versioning policy |
| [`../CHANGELOG.md`](../CHANGELOG.md) | Version history and release notes |
| [`../REVIEW.md`](../REVIEW.md) | Human review checklist for v1.0 |
| [`../TEST_REPORT.md`](../TEST_REPORT.md) | Test metrics and verification results |
