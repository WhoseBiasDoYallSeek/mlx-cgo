/* Copyright © 2023-2024 Apple Inc.                   */
/*                                                    */
/* This file is auto-generated. Do not edit manually. */
/*                                                    */

#ifndef MLX_MAP_STRING_TO_ARRAY_PRIVATE_H
#define MLX_MAP_STRING_TO_ARRAY_PRIVATE_H

#include "mlx/c/map_string_to_array.h"
#include "mlx/mlx.h"

inline mlx_map_string_to_array mlx_map_string_to_array_new_() {
  return mlx_map_string_to_array({nullptr});
}

inline mlx_map_string_to_array mlx_map_string_to_array_new_(
    const std::unordered_map<std::string, mlx::core::array>& s) {
  return mlx_map_string_to_array(
      {new std::unordered_map<std::string, mlx::core::array>(s)});
}

inline mlx_map_string_to_array mlx_map_string_to_array_new_(
    std::unordered_map<std::string, mlx::core::array>&& s) {
  return mlx_map_string_to_array(
      {new std::unordered_map<std::string, mlx::core::array>(std::move(s))});
}

inline mlx_map_string_to_array& mlx_map_string_to_array_set_(
    mlx_map_string_to_array& d,
    const std::unordered_map<std::string, mlx::core::array>& s) {
  if (d.ctx) {
    *static_cast<std::unordered_map<std::string, mlx::core::array>*>(d.ctx) = s;
  } else {
    d.ctx = new std::unordered_map<std::string, mlx::core::array>(s);
  }
  return d;
}

inline mlx_map_string_to_array& mlx_map_string_to_array_set_(
    mlx_map_string_to_array& d,
    std::unordered_map<std::string, mlx::core::array>&& s) {
  if (d.ctx) {
    *static_cast<std::unordered_map<std::string, mlx::core::array>*>(d.ctx) =
        std::move(s);
  } else {
    d.ctx = new std::unordered_map<std::string, mlx::core::array>(std::move(s));
  }
  return d;
}

inline std::unordered_map<std::string, mlx::core::array>&
mlx_map_string_to_array_get_(mlx_map_string_to_array d) {
  if (!d.ctx) {
    throw std::runtime_error("expected a non-empty mlx_map_string_to_array");
  }
  return *static_cast<std::unordered_map<std::string, mlx::core::array>*>(
      d.ctx);
}

inline void mlx_map_string_to_array_free_(mlx_map_string_to_array d) {
  if (d.ctx) {
    delete static_cast<std::unordered_map<std::string, mlx::core::array>*>(
        d.ctx);
  }
}

#endif
