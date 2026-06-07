/* Copyright © 2023-2024 Apple Inc.                   */
/*                                                    */
/* This file is auto-generated. Do not edit manually. */
/*                                                    */

#ifndef MLX_CLOSURE_CUSTOM_JVP_PRIVATE_H
#define MLX_CLOSURE_CUSTOM_JVP_PRIVATE_H

#include "mlx/c/closure_custom_jvp.h"
#include "mlx/mlx.h"

inline mlx_closure_custom_jvp mlx_closure_custom_jvp_new_() {
  return mlx_closure_custom_jvp({nullptr});
}

inline mlx_closure_custom_jvp mlx_closure_custom_jvp_new_(
    const std::function<std::vector<
        array>(std::vector<array>, std::vector<array>, std::vector<int>)>& s) {
  return mlx_closure_custom_jvp({new std::function<std::vector<array>(
      std::vector<array>, std::vector<array>, std::vector<int>)>(s)});
}

inline mlx_closure_custom_jvp mlx_closure_custom_jvp_new_(
    std::function<std::vector<
        array>(std::vector<array>, std::vector<array>, std::vector<int>)>&& s) {
  return mlx_closure_custom_jvp({new std::function<std::vector<array>(
      std::vector<array>, std::vector<array>, std::vector<int>)>(
      std::move(s))});
}

inline mlx_closure_custom_jvp& mlx_closure_custom_jvp_set_(
    mlx_closure_custom_jvp& d,
    const std::function<std::vector<
        array>(std::vector<array>, std::vector<array>, std::vector<int>)>& s) {
  if (d.ctx) {
    *static_cast<std::function<std::vector<array>(
        std::vector<array>, std::vector<array>, std::vector<int>)>*>(d.ctx) = s;
  } else {
    d.ctx = new std::function<std::vector<array>(
        std::vector<array>, std::vector<array>, std::vector<int>)>(s);
  }
  return d;
}

inline mlx_closure_custom_jvp& mlx_closure_custom_jvp_set_(
    mlx_closure_custom_jvp& d,
    std::function<std::vector<
        array>(std::vector<array>, std::vector<array>, std::vector<int>)>&& s) {
  if (d.ctx) {
    *static_cast<std::function<std::vector<array>(
        std::vector<array>, std::vector<array>, std::vector<int>)>*>(d.ctx) =
        std::move(s);
  } else {
    d.ctx = new std::function<std::vector<array>(
        std::vector<array>, std::vector<array>, std::vector<int>)>(
        std::move(s));
  }
  return d;
}

inline std::function<std::vector<
    array>(std::vector<array>, std::vector<array>, std::vector<int>)>&
mlx_closure_custom_jvp_get_(mlx_closure_custom_jvp d) {
  if (!d.ctx) {
    throw std::runtime_error("expected a non-empty mlx_closure_custom_jvp");
  }
  return *static_cast<std::function<std::vector<array>(
      std::vector<array>, std::vector<array>, std::vector<int>)>*>(d.ctx);
}

inline void mlx_closure_custom_jvp_free_(mlx_closure_custom_jvp d) {
  if (d.ctx) {
    delete static_cast<std::function<std::vector<array>(
        std::vector<array>, std::vector<array>, std::vector<int>)>*>(d.ctx);
  }
}

#endif
