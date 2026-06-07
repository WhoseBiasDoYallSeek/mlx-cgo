# Copyright © 2023-2024 Apple Inc.

cmake_minimum_required(VERSION 3.16)

# ----------------------------- Setup -----------------------------
project(mlx.data LANGUAGES CXX C)
set(CMAKE_CXX_STANDARD 20)

# ----------------------------- Configuration -----------------------------
if(NOT MLX_C_VERSION)
  set(MLX_C_VERSION 0.6.0)
endif()

option(BUILD_SHARED_LIBS "Build mlx C as a shared library" OFF)
option(MLX_C_BUILD_EXAMPLES "Build examples for mlx C" ON)
option(MLX_C_USE_SYSTEM_MLX "Use system MLX" OFF)
option(MLX_C_REGENERATE_BINDINGS
       "Regenerate C bindings using mlxgen (Go toolchain) instead of using committed files"
       OFF)

# ----------------------------- mlx -----------------------------

include(FetchContent)
# Avoid warning about DOWNLOAD_EXTRACT_TIMESTAMP in CMake 3.24:
if(POLICY CMP0135)
  cmake_policy(SET CMP0135 NEW)
endif()

# mlx
set(MLX_BUILD_TESTS OFF)
set(MLX_BUILD_EXAMPLES OFF)
set(MLX_BUILD_BENCHMARKS OFF)
set(MLX_BUILD_PYTHON_BINDINGS OFF)

if(MLX_C_USE_SYSTEM_MLX)
  find_package(MLX REQUIRED)
else()
  FetchContent_Declare(
    mlx
    GIT_REPOSITORY "https://github.com/ml-explore/mlx.git"
    GIT_TAG v0.31.2)
  FetchContent_MakeAvailable(mlx)
endif()

# ----------------------------- Code generation (optional) -----------------------------
#
# When MLX_C_REGENERATE_BINDINGS=ON, all C binding files in mlx/c/ are regenerated
# from scratch using the Go toolchain (mlxgen) before compilation.
# By default this is OFF — the committed files in mlx/c/ are used as-is.
#
# Usage:
#   cmake -B build -DMLX_C_REGENERATE_BINDINGS=ON
#
# Requirements (only when ON):
#   - Go 1.21+  (go build)
#   - clang-format (for post-generation formatting)

if(MLX_C_REGENERATE_BINDINGS)
  find_program(GO_EXECUTABLE go DOC "Go compiler")
  if(NOT GO_EXECUTABLE)
    message(FATAL_ERROR
      "MLX_C_REGENERATE_BINDINGS=ON requires Go to be installed and on PATH.\n"
      "Install Go from https://go.dev/dl/ or disable with -DMLX_C_REGENERATE_BINDINGS=OFF")
  endif()

  find_program(CLANG_FORMAT_EXECUTABLE
    NAMES clang-format
    DOC "clang-format for post-generation formatting")
  if(NOT CLANG_FORMAT_EXECUTABLE)
    find_program(XCRUN_EXECUTABLE xcrun)
    if(XCRUN_EXECUTABLE)
      execute_process(
        COMMAND ${XCRUN_EXECUTABLE} --find clang-format
        OUTPUT_VARIABLE CLANG_FORMAT_EXECUTABLE
        OUTPUT_STRIP_TRAILING_WHITESPACE
        ERROR_QUIET)
    endif()
  endif()
  if(NOT CLANG_FORMAT_EXECUTABLE)
    message(FATAL_ERROR
      "MLX_C_REGENERATE_BINDINGS=ON requires clang-format.\n"
      "Install via Homebrew (brew install clang-format) or set CLANG_FORMAT_EXECUTABLE.")
  endif()

  message(STATUS "mlx-cgo: code generation enabled")
  message(STATUS "  Go:           ${GO_EXECUTABLE}")
  message(STATUS "  clang-format: ${CLANG_FORMAT_EXECUTABLE}")

  # Determine MLX source path (set by FetchContent or system install)
  if(MLX_C_USE_SYSTEM_MLX)
    get_target_property(_mlx_iface_includes mlx INTERFACE_INCLUDE_DIRECTORIES)
    list(GET _mlx_iface_includes 0 _mlx_src_base)
  else()
    # mlx_SOURCE_DIR is set by FetchContent_MakeAvailable(mlx)
    set(_mlx_src_base "${mlx_SOURCE_DIR}/mlx")
  endif()

  # Build mlxgen binary into the build tree
  set(MLXGEN_BIN "${CMAKE_BINARY_DIR}/mlxgen")
  set(MLXGEN_STAMP "${CMAKE_BINARY_DIR}/mlxgen.stamp")

  file(GLOB_RECURSE _mlxgen_go_srcs
    "${CMAKE_CURRENT_LIST_DIR}/cmd/mlxgen/*.go"
    "${CMAKE_CURRENT_LIST_DIR}/internal/*.go"
    "${CMAKE_CURRENT_LIST_DIR}/internal/**/*.go")

  add_custom_command(
    OUTPUT "${MLXGEN_STAMP}"
    COMMAND "${GO_EXECUTABLE}" build -o "${MLXGEN_BIN}"
            "${CMAKE_CURRENT_LIST_DIR}/cmd/mlxgen"
    COMMAND "${CMAKE_COMMAND}" -E touch "${MLXGEN_STAMP}"
    DEPENDS ${_mlxgen_go_srcs} "${CMAKE_CURRENT_LIST_DIR}/go.mod"
    WORKING_DIRECTORY "${CMAKE_CURRENT_LIST_DIR}"
    COMMENT "Building mlxgen"
    VERBATIM)

  add_custom_target(mlxgen_build DEPENDS "${MLXGEN_STAMP}")

  # ── Static files (vector, map, closure, all private headers) ──────────────
  # Delegate entirely to regenerate.sh which already knows all the type details.
  # The stamp file tracks whether regeneration has run.
  set(MLXGEN_STATIC_STAMP "${CMAKE_BINARY_DIR}/mlxgen_static.stamp")

  set(_regen_script "${CMAKE_CURRENT_LIST_DIR}/scripts/regenerate.sh")

  add_custom_command(
    OUTPUT "${MLXGEN_STATIC_STAMP}"
    COMMAND env
      "CLANG_FORMAT=${CLANG_FORMAT_EXECUTABLE}"
      bash "${_regen_script}"
    COMMAND "${CMAKE_COMMAND}" -E touch "${MLXGEN_STATIC_STAMP}"
    DEPENDS mlxgen_build "${_regen_script}"
    WORKING_DIRECTORY "${CMAKE_CURRENT_LIST_DIR}"
    COMMENT "Generating static C binding files (vector, map, closure, private headers)"
    VERBATIM)

  add_custom_target(mlxgen_static DEPENDS "${MLXGEN_STATIC_STAMP}")

  # ── Dynamic files (ops, fft, linalg, etc. — from MLX C++ headers) ─────────
  set(MLXGEN_DYNAMIC_STAMP "${CMAKE_BINARY_DIR}/mlxgen_dynamic.stamp")

  add_custom_command(
    OUTPUT "${MLXGEN_DYNAMIC_STAMP}"
    COMMAND env
      "CLANG_FORMAT=${CLANG_FORMAT_EXECUTABLE}"
      bash "${_regen_script}" --mlx-source "${_mlx_src_base}"
    COMMAND "${CMAKE_COMMAND}" -E touch "${MLXGEN_DYNAMIC_STAMP}"
    DEPENDS mlxgen_build mlxgen_static "${_regen_script}"
    WORKING_DIRECTORY "${CMAKE_CURRENT_LIST_DIR}"
    COMMENT "Generating dynamic C binding files from MLX C++ headers"
    VERBATIM)

  add_custom_target(mlxgen_dynamic DEPENDS "${MLXGEN_DYNAMIC_STAMP}")

  # Master target
  add_custom_target(mlxgen_all DEPENDS mlxgen_static mlxgen_dynamic)

endif() # MLX_C_REGENERATE_BINDINGS

# ----------------------------- lib -----------------------------

set(mlxc-src
    ${CMAKE_CURRENT_LIST_DIR}/mlx/c/array.cpp
    ${CMAKE_CURRENT_LIST_DIR}/mlx/c/closure.cpp
    ${CMAKE_CURRENT_LIST_DIR}/mlx/c/compile.cpp
    ${CMAKE_CURRENT_LIST_DIR}/mlx/c/device.cpp
    ${CMAKE_CURRENT_LIST_DIR}/mlx/c/distributed.cpp
    ${CMAKE_CURRENT_LIST_DIR}/mlx/c/distributed_group.cpp
    ${CMAKE_CURRENT_LIST_DIR}/mlx/c/error.cpp
    ${CMAKE_CURRENT_LIST_DIR}/mlx/c/export.cpp
    ${CMAKE_CURRENT_LIST_DIR}/mlx/c/fast.cpp
    ${CMAKE_CURRENT_LIST_DIR}/mlx/c/fft.cpp
    ${CMAKE_CURRENT_LIST_DIR}/mlx/c/graph_utils.cpp
    ${CMAKE_CURRENT_LIST_DIR}/mlx/c/io.cpp
    ${CMAKE_CURRENT_LIST_DIR}/mlx/c/io_types.cpp
    ${CMAKE_CURRENT_LIST_DIR}/mlx/c/linalg.cpp
    ${CMAKE_CURRENT_LIST_DIR}/mlx/c/map.cpp
    ${CMAKE_CURRENT_LIST_DIR}/mlx/c/memory.cpp
    ${CMAKE_CURRENT_LIST_DIR}/mlx/c/metal.cpp
    # ${CMAKE_CURRENT_LIST_DIR}/mlx/c/object.cpp
    ${CMAKE_CURRENT_LIST_DIR}/mlx/c/ops.cpp
    ${CMAKE_CURRENT_LIST_DIR}/mlx/c/random.cpp
    ${CMAKE_CURRENT_LIST_DIR}/mlx/c/stream.cpp
    ${CMAKE_CURRENT_LIST_DIR}/mlx/c/string.cpp
    ${CMAKE_CURRENT_LIST_DIR}/mlx/c/transforms.cpp
    ${CMAKE_CURRENT_LIST_DIR}/mlx/c/transforms_impl.cpp
    ${CMAKE_CURRENT_LIST_DIR}/mlx/c/vector.cpp
    ${CMAKE_CURRENT_LIST_DIR}/mlx/c/version.cpp)

add_library(mlxc ${mlxc-src})

# When regenerating, the library must wait for generated files
if(MLX_C_REGENERATE_BINDINGS)
  add_dependencies(mlxc mlxgen_all)
endif()

target_link_libraries(mlxc PRIVATE mlx)
target_include_directories(mlxc
                           PUBLIC $<BUILD_INTERFACE:${CMAKE_CURRENT_LIST_DIR}>)
set_property(TARGET mlxc PROPERTY POSITION_INDEPENDENT_CODE ON)

# Windows DLL symbol exports
if(WIN32)
  set_target_properties(mlxc PROPERTIES WINDOWS_EXPORT_ALL_SYMBOLS TRUE)
endif()

if(MLX_C_BUILD_EXAMPLES)
  add_subdirectory(${CMAKE_CURRENT_LIST_DIR}/examples)
endif()

# ----------------------------- Installation -----------------------------
include(GNUInstallDirs)

# Install library
install(
  TARGETS mlxc
  EXPORT MLXCTargets
  LIBRARY DESTINATION ${CMAKE_INSTALL_LIBDIR}
  ARCHIVE DESTINATION ${CMAKE_INSTALL_LIBDIR}
  RUNTIME DESTINATION ${CMAKE_INSTALL_BINDIR}
  INCLUDES
  DESTINATION ${CMAKE_INSTALL_INCLUDEDIR})

# Install headers
install(
  DIRECTORY ${CMAKE_CURRENT_LIST_DIR}/mlx
  DESTINATION ${CMAKE_INSTALL_INCLUDEDIR}
  COMPONENT headers
  FILES_MATCHING
  PATTERN "*.h"
  PATTERN "private" EXCLUDE)

# Install cmake config
set(MLX_C_CMAKE_BUILD_CONFIG ${CMAKE_BINARY_DIR}/MLXCConfig.cmake)
set(MLX_C_CMAKE_BUILD_VERSION_CONFIG
    ${CMAKE_BINARY_DIR}/MLXCConfigVersion.cmake)
set(MLX_C_CMAKE_INSTALL_MODULE_DIR share/cmake/MLXC)

install(
  EXPORT MLXCTargets
  FILE MLXCTargets.cmake
  DESTINATION ${MLX_C_CMAKE_INSTALL_MODULE_DIR})

include(CMakePackageConfigHelpers)

write_basic_package_version_file(
  ${MLX_C_CMAKE_BUILD_VERSION_CONFIG}
  COMPATIBILITY SameMajorVersion
  VERSION ${MLX_C_VERSION})

configure_package_config_file(
  ${CMAKE_CURRENT_LIST_DIR}/mlx-c.pc.in ${MLX_C_CMAKE_BUILD_CONFIG}
  INSTALL_DESTINATION ${MLX_C_CMAKE_INSTALL_MODULE_DIR}
  NO_CHECK_REQUIRED_COMPONENTS_MACRO
  PATH_VARS CMAKE_INSTALL_LIBDIR CMAKE_INSTALL_INCLUDEDIR
            MLX_C_CMAKE_INSTALL_MODULE_DIR)

install(FILES ${MLX_C_CMAKE_BUILD_CONFIG} ${MLX_C_CMAKE_BUILD_VERSION_CONFIG}
        DESTINATION ${MLX_C_CMAKE_INSTALL_MODULE_DIR})
