/* Copyright © 2023-2024 Apple Inc.                   */
/*                                                    */
/* This file is auto-generated. Do not edit manually. */
/*                                                    */

#ifndef MLX_VECTOR_STRING_PRIVATE_H
#define MLX_VECTOR_STRING_PRIVATE_H

#include "mlx/c/vector_string.h"
#include "mlx/mlx.h"

inline mlx_vector_string mlx_vector_string_new_() {
  return mlx_vector_string({nullptr});
}

inline mlx_vector_string mlx_vector_string_new_(
    const std::vector<std::string>& s) {
  return mlx_vector_string({new std::vector<std::string>(s)});
}

inline mlx_vector_string mlx_vector_string_new_(std::vector<std::string>&& s) {
  return mlx_vector_string({new std::vector<std::string>(std::move(s))});
}

inline mlx_vector_string& mlx_vector_string_set_(
    mlx_vector_string& d,
    const std::vector<std::string>& s) {
  if (d.ctx) {
    *static_cast<std::vector<std::string>*>(d.ctx) = s;
  } else {
    d.ctx = new std::vector<std::string>(s);
  }
  return d;
}

inline mlx_vector_string& mlx_vector_string_set_(
    mlx_vector_string& d,
    std::vector<std::string>&& s) {
  if (d.ctx) {
    *static_cast<std::vector<std::string>*>(d.ctx) = std::move(s);
  } else {
    d.ctx = new std::vector<std::string>(std::move(s));
  }
  return d;
}

inline std::vector<std::string>& mlx_vector_string_get_(mlx_vector_string d) {
  if (!d.ctx) {
    throw std::runtime_error("expected a non-empty mlx_vector_string");
  }
  return *static_cast<std::vector<std::string>*>(d.ctx);
}

inline void mlx_vector_string_free_(mlx_vector_string d) {
  if (d.ctx) {
    delete static_cast<std::vector<std::string>*>(d.ctx);
  }
}

#endif
