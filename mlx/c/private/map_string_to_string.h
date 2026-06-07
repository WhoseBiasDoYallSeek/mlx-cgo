/* Copyright © 2023-2024 Apple Inc.                   */
/*                                                    */
/* This file is auto-generated. Do not edit manually. */
/*                                                    */

#ifndef MLX_MAP_STRING_TO_STRING_PRIVATE_H
#define MLX_MAP_STRING_TO_STRING_PRIVATE_H

#include "mlx/c/map_string_to_string.h"
#include "mlx/mlx.h"

inline mlx_map_string_to_string mlx_map_string_to_string_new_() {
  return mlx_map_string_to_string({nullptr});
}

inline mlx_map_string_to_string mlx_map_string_to_string_new_(
    const std::unordered_map<std::string, std::string>& s) {
  return mlx_map_string_to_string(
      {new std::unordered_map<std::string, std::string>(s)});
}

inline mlx_map_string_to_string mlx_map_string_to_string_new_(
    std::unordered_map<std::string, std::string>&& s) {
  return mlx_map_string_to_string(
      {new std::unordered_map<std::string, std::string>(std::move(s))});
}

inline mlx_map_string_to_string& mlx_map_string_to_string_set_(
    mlx_map_string_to_string& d,
    const std::unordered_map<std::string, std::string>& s) {
  if (d.ctx) {
    *static_cast<std::unordered_map<std::string, std::string>*>(d.ctx) = s;
  } else {
    d.ctx = new std::unordered_map<std::string, std::string>(s);
  }
  return d;
}

inline mlx_map_string_to_string& mlx_map_string_to_string_set_(
    mlx_map_string_to_string& d,
    std::unordered_map<std::string, std::string>&& s) {
  if (d.ctx) {
    *static_cast<std::unordered_map<std::string, std::string>*>(d.ctx) =
        std::move(s);
  } else {
    d.ctx = new std::unordered_map<std::string, std::string>(std::move(s));
  }
  return d;
}

inline std::unordered_map<std::string, std::string>&
mlx_map_string_to_string_get_(mlx_map_string_to_string d) {
  if (!d.ctx) {
    throw std::runtime_error("expected a non-empty mlx_map_string_to_string");
  }
  return *static_cast<std::unordered_map<std::string, std::string>*>(d.ctx);
}

inline void mlx_map_string_to_string_free_(mlx_map_string_to_string d) {
  if (d.ctx) {
    delete static_cast<std::unordered_map<std::string, std::string>*>(d.ctx);
  }
}

#endif
