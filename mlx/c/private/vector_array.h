/* Copyright © 2023-2024 Apple Inc.                   */
/*                                                    */
/* This file is auto-generated. Do not edit manually. */
/*                                                    */

#ifndef MLX_VECTOR_ARRAY_PRIVATE_H
#define MLX_VECTOR_ARRAY_PRIVATE_H

#include "mlx/c/vector_array.h"
#include "mlx/mlx.h"

inline mlx_vector_array mlx_vector_array_new_() {
  return mlx_vector_array({nullptr});
}

inline mlx_vector_array mlx_vector_array_new_(const std::vector<array>& s) {
  return mlx_vector_array({new std::vector<array>(s)});
}

inline mlx_vector_array mlx_vector_array_new_(std::vector<array>&& s) {
  return mlx_vector_array({new std::vector<array>(std::move(s))});
}

inline mlx_vector_array& mlx_vector_array_set_(
    mlx_vector_array& d,
    const std::vector<array>& s) {
  if (d.ctx) {
    *static_cast<std::vector<array>*>(d.ctx) = s;
  } else {
    d.ctx = new std::vector<array>(s);
  }
  return d;
}

inline mlx_vector_array& mlx_vector_array_set_(
    mlx_vector_array& d,
    std::vector<array>&& s) {
  if (d.ctx) {
    *static_cast<std::vector<array>*>(d.ctx) = std::move(s);
  } else {
    d.ctx = new std::vector<array>(std::move(s));
  }
  return d;
}

inline std::vector<array>& mlx_vector_array_get_(mlx_vector_array d) {
  if (!d.ctx) {
    throw std::runtime_error("expected a non-empty mlx_vector_array");
  }
  return *static_cast<std::vector<array>*>(d.ctx);
}

inline void mlx_vector_array_free_(mlx_vector_array d) {
  if (d.ctx) {
    delete static_cast<std::vector<array>*>(d.ctx);
  }
}

#endif
