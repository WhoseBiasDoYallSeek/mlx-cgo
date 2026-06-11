/* Copyright © 2023-2024 Apple Inc.                   */
/*                                                    */
/* This file is auto-generated. Do not edit manually. */
/*                                                    */

#ifndef MLX_CLOSURE_CUSTOM_PRIVATE_H
#define MLX_CLOSURE_CUSTOM_PRIVATE_H

#include "mlx/c/closure_custom.h"
#include "mlx/mlx.h"

inline mlx_closure_custom mlx_closure_custom_new_() {
  return mlx_closure_custom({nullptr});
}

inline mlx_closure_custom mlx_closure_custom_new_(
    const std::function<std::vector<mlx::core::array>(
        std::vector<mlx::core::array>,
        std::vector<mlx::core::array>,
        std::vector<mlx::core::array>)>& s) {
  return mlx_closure_custom({new std::function<std::vector<mlx::core::array>(
      std::vector<mlx::core::array>,
      std::vector<mlx::core::array>,
      std::vector<mlx::core::array>)>(s)});
}

inline mlx_closure_custom mlx_closure_custom_new_(
    std::function<std::vector<mlx::core::array>(
        std::vector<mlx::core::array>,
        std::vector<mlx::core::array>,
        std::vector<mlx::core::array>)>&& s) {
  return mlx_closure_custom({new std::function<std::vector<mlx::core::array>(
      std::vector<mlx::core::array>,
      std::vector<mlx::core::array>,
      std::vector<mlx::core::array>)>(std::move(s))});
}

inline mlx_closure_custom& mlx_closure_custom_set_(
    mlx_closure_custom& d,
    const std::function<std::vector<mlx::core::array>(
        std::vector<mlx::core::array>,
        std::vector<mlx::core::array>,
        std::vector<mlx::core::array>)>& s) {
  if (d.ctx) {
    *static_cast<std::function<std::vector<mlx::core::array>(
        std::vector<mlx::core::array>,
        std::vector<mlx::core::array>,
        std::vector<mlx::core::array>)>*>(d.ctx) = s;
  } else {
    d.ctx = new std::function<std::vector<mlx::core::array>(
        std::vector<mlx::core::array>,
        std::vector<mlx::core::array>,
        std::vector<mlx::core::array>)>(s);
  }
  return d;
}

inline mlx_closure_custom& mlx_closure_custom_set_(
    mlx_closure_custom& d,
    std::function<std::vector<mlx::core::array>(
        std::vector<mlx::core::array>,
        std::vector<mlx::core::array>,
        std::vector<mlx::core::array>)>&& s) {
  if (d.ctx) {
    *static_cast<std::function<std::vector<mlx::core::array>(
        std::vector<mlx::core::array>,
        std::vector<mlx::core::array>,
        std::vector<mlx::core::array>)>*>(d.ctx) = std::move(s);
  } else {
    d.ctx = new std::function<std::vector<mlx::core::array>(
        std::vector<mlx::core::array>,
        std::vector<mlx::core::array>,
        std::vector<mlx::core::array>)>(std::move(s));
  }
  return d;
}

inline std::function<std::vector<mlx::core::array>(
    std::vector<mlx::core::array>,
    std::vector<mlx::core::array>,
    std::vector<mlx::core::array>)>&
mlx_closure_custom_get_(mlx_closure_custom d) {
  if (!d.ctx) {
    throw std::runtime_error("expected a non-empty mlx_closure_custom");
  }
  return *static_cast<std::function<std::vector<mlx::core::array>(
      std::vector<mlx::core::array>,
      std::vector<mlx::core::array>,
      std::vector<mlx::core::array>)>*>(d.ctx);
}

inline void mlx_closure_custom_free_(mlx_closure_custom d) {
  if (d.ctx) {
    delete static_cast<std::function<std::vector<mlx::core::array>(
        std::vector<mlx::core::array>,
        std::vector<mlx::core::array>,
        std::vector<mlx::core::array>)>*>(d.ctx);
  }
}

#endif
