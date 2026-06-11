/* Copyright © 2023-2024 Apple Inc.                   */
/*                                                    */
/* This file is auto-generated. Do not edit manually. */
/*                                                    */

#ifndef MLX_CLOSURE_VALUE_AND_GRAD_PRIVATE_H
#define MLX_CLOSURE_VALUE_AND_GRAD_PRIVATE_H

#include "mlx/c/closure_value_and_grad.h"
#include "mlx/mlx.h"

inline mlx_closure_value_and_grad mlx_closure_value_and_grad_new_() {
  return mlx_closure_value_and_grad({nullptr});
}

inline mlx_closure_value_and_grad mlx_closure_value_and_grad_new_(
    const std::function<
        std::pair<std::vector<mlx::core::array>, std::vector<mlx::core::array>>(
            const std::vector<mlx::core::array>&)>& s) {
  return mlx_closure_value_and_grad({new std::function<
      std::pair<std::vector<mlx::core::array>, std::vector<mlx::core::array>>(
          const std::vector<mlx::core::array>&)>(s)});
}

inline mlx_closure_value_and_grad mlx_closure_value_and_grad_new_(
    std::function<
        std::pair<std::vector<mlx::core::array>, std::vector<mlx::core::array>>(
            const std::vector<mlx::core::array>&)>&& s) {
  return mlx_closure_value_and_grad({new std::function<
      std::pair<std::vector<mlx::core::array>, std::vector<mlx::core::array>>(
          const std::vector<mlx::core::array>&)>(std::move(s))});
}

inline mlx_closure_value_and_grad& mlx_closure_value_and_grad_set_(
    mlx_closure_value_and_grad& d,
    const std::function<
        std::pair<std::vector<mlx::core::array>, std::vector<mlx::core::array>>(
            const std::vector<mlx::core::array>&)>& s) {
  if (d.ctx) {
    *static_cast<std::function<
        std::pair<std::vector<mlx::core::array>, std::vector<mlx::core::array>>(
            const std::vector<mlx::core::array>&)>*>(d.ctx) = s;
  } else {
    d.ctx = new std::function<
        std::pair<std::vector<mlx::core::array>, std::vector<mlx::core::array>>(
            const std::vector<mlx::core::array>&)>(s);
  }
  return d;
}

inline mlx_closure_value_and_grad& mlx_closure_value_and_grad_set_(
    mlx_closure_value_and_grad& d,
    std::function<
        std::pair<std::vector<mlx::core::array>, std::vector<mlx::core::array>>(
            const std::vector<mlx::core::array>&)>&& s) {
  if (d.ctx) {
    *static_cast<std::function<
        std::pair<std::vector<mlx::core::array>, std::vector<mlx::core::array>>(
            const std::vector<mlx::core::array>&)>*>(d.ctx) = std::move(s);
  } else {
    d.ctx = new std::function<
        std::pair<std::vector<mlx::core::array>, std::vector<mlx::core::array>>(
            const std::vector<mlx::core::array>&)>(std::move(s));
  }
  return d;
}

inline std::function<
    std::pair<std::vector<mlx::core::array>, std::vector<mlx::core::array>>(
        const std::vector<mlx::core::array>&)>&
mlx_closure_value_and_grad_get_(mlx_closure_value_and_grad d) {
  if (!d.ctx) {
    throw std::runtime_error("expected a non-empty mlx_closure_value_and_grad");
  }
  return *static_cast<std::function<
      std::pair<std::vector<mlx::core::array>, std::vector<mlx::core::array>>(
          const std::vector<mlx::core::array>&)>*>(d.ctx);
}

inline void mlx_closure_value_and_grad_free_(mlx_closure_value_and_grad d) {
  if (d.ctx) {
    delete static_cast<std::function<
        std::pair<std::vector<mlx::core::array>, std::vector<mlx::core::array>>(
            const std::vector<mlx::core::array>&)>*>(d.ctx);
  }
}

#endif
