/* Copyright © 2023-2024 Apple Inc.                   */
/*                                                    */
/* This file is auto-generated. Do not edit manually. */
/*                                                    */

#ifndef MLX_CLOSURE_PRIVATE_H
#define MLX_CLOSURE_PRIVATE_H

#include "mlx/c/closure.h"
#include "mlx/mlx.h"

inline mlx_closure mlx_closure_new_() {
  return mlx_closure({nullptr});
}

inline mlx_closure mlx_closure_new_(
    const std::function<std::vector<array>(std::vector<array>)>& s) {
  return mlx_closure(
      {new std::function<std::vector<array>(std::vector<array>)>(s)});
}

inline mlx_closure mlx_closure_new_(
    std::function<std::vector<array>(std::vector<array>)>&& s) {
  return mlx_closure({new std::function<std::vector<array>(std::vector<array>)>(
      std::move(s))});
}

inline mlx_closure& mlx_closure_set_(
    mlx_closure& d,
    const std::function<std::vector<array>(std::vector<array>)>& s) {
  if (d.ctx) {
    *static_cast<std::function<std::vector<array>(std::vector<array>)>*>(
        d.ctx) = s;
  } else {
    d.ctx = new std::function<std::vector<array>(std::vector<array>)>(s);
  }
  return d;
}

inline mlx_closure& mlx_closure_set_(
    mlx_closure& d,
    std::function<std::vector<array>(std::vector<array>)>&& s) {
  if (d.ctx) {
    *static_cast<std::function<std::vector<array>(std::vector<array>)>*>(
        d.ctx) = std::move(s);
  } else {
    d.ctx =
        new std::function<std::vector<array>(std::vector<array>)>(std::move(s));
  }
  return d;
}

inline std::function<std::vector<array>(std::vector<array>)>& mlx_closure_get_(
    mlx_closure d) {
  if (!d.ctx) {
    throw std::runtime_error("expected a non-empty mlx_closure");
  }
  return *static_cast<std::function<std::vector<array>(std::vector<array>)>*>(
      d.ctx);
}

inline void mlx_closure_free_(mlx_closure d) {
  if (d.ctx) {
    delete static_cast<std::function<std::vector<array>(std::vector<array>)>*>(
        d.ctx);
  }
}

#endif
