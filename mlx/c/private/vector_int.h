/* Copyright © 2023-2024 Apple Inc.                   */
/*                                                    */
/* This file is auto-generated. Do not edit manually. */
/*                                                    */

#ifndef MLX_VECTOR_INT_PRIVATE_H
#define MLX_VECTOR_INT_PRIVATE_H

#include "mlx/c/vector_int.h"
#include "mlx/mlx.h"

inline mlx_vector_int mlx_vector_int_new_() {
  return mlx_vector_int({nullptr});
}

inline mlx_vector_int mlx_vector_int_new_(const std::vector<int>& s) {
  return mlx_vector_int({new std::vector<int>(s)});
}

inline mlx_vector_int mlx_vector_int_new_(std::vector<int>&& s) {
  return mlx_vector_int({new std::vector<int>(std::move(s))});
}

inline mlx_vector_int& mlx_vector_int_set_(
    mlx_vector_int& d,
    const std::vector<int>& s) {
  if (d.ctx) {
    *static_cast<std::vector<int>*>(d.ctx) = s;
  } else {
    d.ctx = new std::vector<int>(s);
  }
  return d;
}

inline mlx_vector_int& mlx_vector_int_set_(
    mlx_vector_int& d,
    std::vector<int>&& s) {
  if (d.ctx) {
    *static_cast<std::vector<int>*>(d.ctx) = std::move(s);
  } else {
    d.ctx = new std::vector<int>(std::move(s));
  }
  return d;
}

inline std::vector<int>& mlx_vector_int_get_(mlx_vector_int d) {
  if (!d.ctx) {
    throw std::runtime_error("expected a non-empty mlx_vector_int");
  }
  return *static_cast<std::vector<int>*>(d.ctx);
}

inline void mlx_vector_int_free_(mlx_vector_int d) {
  if (d.ctx) {
    delete static_cast<std::vector<int>*>(d.ctx);
  }
}

#endif
