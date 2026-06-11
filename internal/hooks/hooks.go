// Package hooks provides hardcoded C/C++ code snippets for functions that
// cannot be automatically generated from the C++ AST — e.g. functions with
// complex config objects or special calling conventions.
//
// This is a direct port of python/mlxhooks.py from the original mlx-c project.
//
// Each hook function receives the function name and an implementation flag.
// It returns (code string, handled bool).
// If handled is false the normal code-generation path is used instead.
// If handled is true the returned code (possibly empty) replaces the normal output.
package hooks

import "strings"

// Result is the output of a hook invocation.
type Result struct {
	// Code is the C/C++ source to emit. May be empty (e.g. when the hook
	// intentionally suppresses the function).
	Code string
	// Handled reports whether this hook took ownership of the function.
	// When false the caller should use normal code generation.
	Handled bool
}

// Apply looks up a hook for the given fully-qualified C function name and
// calls it. If no hook is registered, returns Result{Handled: false}.
func Apply(funcName string, implementation bool) Result {
	if fn, ok := registry[funcName]; ok {
		return fn(implementation)
	}
	return Result{Handled: false}
}

// ApplyModule returns module-level code to emit for a given headername
// (e.g. "export"). Called once per module regardless of parsed functions.
// Returns empty string if no module hook is registered.
func ApplyModule(headername string, implementation bool) string {
	if fn, ok := moduleRegistry[headername]; ok {
		return fn(implementation)
	}
	return ""
}

// hookFn is the type of a hook implementation.
type hookFn func(implementation bool) Result

var registry = map[string]hookFn{
	"mlx_fast_metal_kernel": hookFastMetalKernel,
	"mlx_load_gguf":         hookLoadGGUF,
	"mlx_save_gguf":         hookSaveGGUF,
	"mlx_export_to_dot":     hookExportToDot,
	"mlx_custom_function":   hookCustomFunction,
	"mlx_custom_vjp":        hookCustomVJP,
}

// moduleRegistry maps headername → function that emits module-level code.
var moduleRegistry = map[string]func(bool) string{
	"export": hookModuleExport,
}

// ─── mlx_fast_metal_kernel ───────────────────────────────────────────────────

// customKernelConfigHeader is emitted before the metal kernel hook.
const customKernelConfigHeader = `
typedef struct mlx_fast_BACKEND_kernel_config_ {
  void* ctx;
} mlx_fast_BACKEND_kernel_config;
mlx_fast_BACKEND_kernel_config mlx_fast_BACKEND_kernel_config_new(void);
void mlx_fast_BACKEND_kernel_config_free(mlx_fast_BACKEND_kernel_config cls);

int mlx_fast_BACKEND_kernel_config_add_output_arg(
    mlx_fast_BACKEND_kernel_config cls,
    const int* shape,
    size_t size,
    mlx_dtype dtype);
int mlx_fast_BACKEND_kernel_config_set_grid(
    mlx_fast_BACKEND_kernel_config cls,
    int grid1,
    int grid2,
    int grid3);
int mlx_fast_BACKEND_kernel_config_set_thread_group(
    mlx_fast_BACKEND_kernel_config cls,
    int thread1,
    int thread2,
    int thread3);
int mlx_fast_BACKEND_kernel_config_set_init_value(
    mlx_fast_BACKEND_kernel_config cls,
    float value);
int mlx_fast_BACKEND_kernel_config_set_verbose(
    mlx_fast_BACKEND_kernel_config cls,
    bool verbose);
int mlx_fast_BACKEND_kernel_config_add_template_arg_dtype(
    mlx_fast_BACKEND_kernel_config cls,
    const char* name,
    mlx_dtype dtype);
int mlx_fast_BACKEND_kernel_config_add_template_arg_int(
    mlx_fast_BACKEND_kernel_config cls,
    const char* name,
    int value);
int mlx_fast_BACKEND_kernel_config_add_template_arg_bool(
    mlx_fast_BACKEND_kernel_config cls,
    const char* name,
    bool value);
`

const customKernelConfigImpl = `
struct mlx_fast_BACKEND_kernel_config_cpp_ {
  std::vector<mlx::core::Shape> output_shapes;
  std::vector<mlx::core::Dtype> output_dtypes;
  std::tuple<int, int, int> grid;
  std::tuple<int, int, int> thread_group;
  std::vector<std::pair<std::string, mlx::core::fast::TemplateArg>>
      template_args;
  std::optional<float> init_value;
  bool verbose;
};

inline mlx_fast_BACKEND_kernel_config mlx_fast_BACKEND_kernel_config_new_() {
  return mlx_fast_BACKEND_kernel_config(
      {new mlx_fast_BACKEND_kernel_config_cpp_()});
}

inline mlx_fast_BACKEND_kernel_config_cpp_& mlx_fast_BACKEND_kernel_config_get_(
    mlx_fast_BACKEND_kernel_config d) {
  if (!d.ctx) {
    throw std::runtime_error(
        "expected a non-empty mlx_fast_BACKEND_kernel_config");
  }
  return *static_cast<mlx_fast_BACKEND_kernel_config_cpp_*>(d.ctx);
}

inline void mlx_fast_BACKEND_kernel_config_free_(mlx_fast_BACKEND_kernel_config d) {
  if (d.ctx) {
    delete static_cast<mlx_fast_BACKEND_kernel_config_cpp_*>(d.ctx);
  }
}

extern "C" mlx_fast_BACKEND_kernel_config mlx_fast_BACKEND_kernel_config_new(void) {
  try {
    return mlx_fast_BACKEND_kernel_config_new_();
  } catch (std::exception& e) {
    mlx_error(e.what());
  }
  return {nullptr};
}

extern "C" void mlx_fast_BACKEND_kernel_config_free(
    mlx_fast_BACKEND_kernel_config cls) {
  mlx_fast_BACKEND_kernel_config_free_(cls);
}

extern "C" int mlx_fast_BACKEND_kernel_config_add_output_arg(
    mlx_fast_BACKEND_kernel_config cls,
    const int* shape,
    size_t size,
    mlx_dtype dtype) {
  try {
    mlx_fast_BACKEND_kernel_config_get_(cls).output_shapes.push_back(
        mlx::core::Shape(shape, shape + size));
    mlx_fast_BACKEND_kernel_config_get_(cls).output_dtypes.push_back(
        mlx_dtype_to_cpp(dtype));
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
  return 0;
}
extern "C" int mlx_fast_BACKEND_kernel_config_set_grid(
    mlx_fast_BACKEND_kernel_config cls,
    int grid1,
    int grid2,
    int grid3) {
  try {
    mlx_fast_BACKEND_kernel_config_get_(cls).grid =
        std::make_tuple(grid1, grid2, grid3);
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
  return 0;
}
extern "C" int mlx_fast_BACKEND_kernel_config_set_thread_group(
    mlx_fast_BACKEND_kernel_config cls,
    int thread1,
    int thread2,
    int thread3) {
  try {
    mlx_fast_BACKEND_kernel_config_get_(cls).thread_group =
        std::make_tuple(thread1, thread2, thread3);
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
  return 0;
}
extern "C" int mlx_fast_BACKEND_kernel_config_set_init_value(
    mlx_fast_BACKEND_kernel_config cls,
    float value) {
  try {
    mlx_fast_BACKEND_kernel_config_get_(cls).init_value = value;
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
  return 0;
}
extern "C" int mlx_fast_BACKEND_kernel_config_set_verbose(
    mlx_fast_BACKEND_kernel_config cls,
    bool verbose) {
  try {
    mlx_fast_BACKEND_kernel_config_get_(cls).verbose = verbose;
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
  return 0;
}
extern "C" int mlx_fast_BACKEND_kernel_config_add_template_arg_dtype(
    mlx_fast_BACKEND_kernel_config cls,
    const char* name,
    mlx_dtype dtype) {
  try {
    mlx_fast_BACKEND_kernel_config_get_(cls).template_args.push_back(
        std::make_pair(std::string(name), mlx_dtype_to_cpp(dtype)));
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
  return 0;
}
extern "C" int mlx_fast_BACKEND_kernel_config_add_template_arg_int(
    mlx_fast_BACKEND_kernel_config cls,
    const char* name,
    int value) {
  try {
    mlx_fast_BACKEND_kernel_config_get_(cls).template_args.push_back(
        std::make_pair(std::string(name), value));
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
  return 0;
}
extern "C" int mlx_fast_BACKEND_kernel_config_add_template_arg_bool(
    mlx_fast_BACKEND_kernel_config cls,
    const char* name,
    bool value) {
  try {
    mlx_fast_BACKEND_kernel_config_get_(cls).template_args.push_back(
        std::make_pair(std::string(name), value));
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
  return 0;
}
`

const customKernelDefHeader = `
typedef struct mlx_fast_BACKEND_kernel_ {
  void* ctx;
} mlx_fast_BACKEND_kernel;
`

const customKernelDefImpl = `
struct mlx_fast_BACKEND_kernel_cpp_ {
  mlx::core::fast::CustomKernelFunction mkf;
  mlx_fast_BACKEND_kernel_cpp_(mlx::core::fast::CustomKernelFunction mkf)
      : mkf(mkf) {};
};
`

const customKernelApplyHeader = `
void mlx_fast_BACKEND_kernel_free(mlx_fast_BACKEND_kernel cls);

int mlx_fast_BACKEND_kernel_apply(
    mlx_vector_array* outputs,
    mlx_fast_BACKEND_kernel cls,
    const mlx_vector_array inputs,
    const mlx_fast_BACKEND_kernel_config config,
    const mlx_stream stream);
`

const customKernelApplyImpl = `
inline mlx::core::fast::CustomKernelFunction& mlx_fast_BACKEND_kernel_get_(
    mlx_fast_BACKEND_kernel d) {
  if (!d.ctx) {
    throw std::runtime_error("expected a non-empty mlx_fast_BACKEND_kernel");
  }
  return static_cast<mlx_fast_BACKEND_kernel_cpp_*>(d.ctx)->mkf;
}

inline void mlx_fast_BACKEND_kernel_free_(mlx_fast_BACKEND_kernel d) {
  if (d.ctx) {
    delete static_cast<mlx_fast_BACKEND_kernel_cpp_*>(d.ctx);
  }
}

extern "C" void mlx_fast_BACKEND_kernel_free(mlx_fast_BACKEND_kernel cls) {
  mlx_fast_BACKEND_kernel_free_(cls);
}

extern "C" int mlx_fast_BACKEND_kernel_apply(
    mlx_vector_array* outputs,
    mlx_fast_BACKEND_kernel cls,
    const mlx_vector_array inputs,
    const mlx_fast_BACKEND_kernel_config config,
    const mlx_stream stream) {
  try {
    auto config_ctx = mlx_fast_BACKEND_kernel_config_get_(config);
    mlx_vector_array_set_(
        *outputs,
        mlx_fast_BACKEND_kernel_get_(cls)(
            mlx_vector_array_get_(inputs),
            config_ctx.output_shapes,
            config_ctx.output_dtypes,
            config_ctx.grid,
            config_ctx.thread_group,
            config_ctx.template_args,
            config_ctx.init_value,
            config_ctx.verbose,
            mlx_stream_get_(stream)));
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
  return 0;
}
`

func implementCustomKernel(backend, customCode string, impl bool) string {
	rep := func(s string) string {
		return strings.ReplaceAll(s, "BACKEND", backend)
	}
	var sb strings.Builder
	if impl {
		sb.WriteString(rep(customKernelConfigImpl))
		sb.WriteString(rep(customKernelDefImpl))
		sb.WriteString(customCode)
		sb.WriteString(rep(customKernelApplyImpl))
	} else {
		sb.WriteString(rep(customKernelConfigHeader))
		sb.WriteString(rep(customKernelDefHeader))
		sb.WriteString(customCode)
		sb.WriteString(rep(customKernelApplyHeader))
	}
	return sb.String()
}

func hookFastMetalKernel(impl bool) Result {
	var customCode string
	if impl {
		customCode = `
inline mlx_fast_metal_kernel mlx_fast_metal_kernel_new_(
    const std::string& name,
    const std::vector<std::string>& input_names,
    const std::vector<std::string>& output_names,
    const std::string& source,
    const std::string& header,
    bool ensure_row_contiguous,
    bool atomic_outputs) {
  return mlx_fast_metal_kernel(
      {new mlx_fast_metal_kernel_cpp_(mlx::core::fast::metal_kernel(
          name,
          input_names,
          output_names,
          source,
          header,
          ensure_row_contiguous,
          atomic_outputs))});
}

extern "C" mlx_fast_metal_kernel mlx_fast_metal_kernel_new(
    const char* name,
    const mlx_vector_string input_names,
    const mlx_vector_string output_names,
    const char* source,
    const char* header,
    bool ensure_row_contiguous,
    bool atomic_outputs) {
  try {
    return mlx_fast_metal_kernel_new_(
        name,
        mlx_vector_string_get_(input_names),
        mlx_vector_string_get_(output_names),
        source,
        header,
        ensure_row_contiguous,
        atomic_outputs);
  } catch (std::exception& e) {
    mlx_error(e.what());
  }
  return {nullptr};
}
`
	} else {
		customCode = `
mlx_fast_metal_kernel mlx_fast_metal_kernel_new(
    const char* name,
    const mlx_vector_string input_names,
    const mlx_vector_string output_names,
    const char* source,
    const char* header,
    bool ensure_row_contiguous,
    bool atomic_outputs);
`
	}
	return Result{
		Code:    implementCustomKernel("metal", customCode, impl),
		Handled: true,
	}
}

// ─── mlx_load_gguf / mlx_save_gguf ───────────────────────────────────────────

func hookLoadGGUF(impl bool) Result {
	var code string
	if impl {
		code = `extern "C" int mlx_load_gguf(mlx_io_gguf* gguf, const char* file, const mlx_stream s) {
  try {
    auto cpp_gguf = mlx::core::load_gguf(file, mlx_stream_get_(s));
    mlx_io_gguf_set_(*gguf, std::move(cpp_gguf));
    return 0;
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
}`
	} else {
		code = "\nint mlx_load_gguf(mlx_io_gguf* gguf, const char* file, const mlx_stream s);"
	}
	return Result{Code: code, Handled: true}
}

func hookSaveGGUF(impl bool) Result {
	var code string
	if impl {
		code = `extern "C" int mlx_save_gguf(const char* file, mlx_io_gguf gguf) {
  try {
    auto cpp_gguf = mlx_io_gguf_get_(gguf);
        mlx::core::save_gguf(file, cpp_gguf.first, cpp_gguf.second);
    return 0;
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
}`
	} else {
		code = "int mlx_save_gguf(const char* file, mlx_io_gguf gguf);"
	}
	return Result{Code: code, Handled: true}
}

// ─── mlx_export_to_dot ───────────────────────────────────────────────────────

func hookExportToDot(impl bool) Result {
	var code string
	if impl {
		code = `extern "C" mlx_node_namer mlx_node_namer_new() {
  try {
    return mlx_node_namer_new_(mlx::core::NodeNamer());
  } catch (std::exception& e) {
    mlx_error(e.what());
  }
  return {nullptr};
}
extern "C" int mlx_node_namer_free(mlx_node_namer namer) {
  try {
    mlx_node_namer_free_(namer);
    return 0;
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
}
extern "C" int mlx_node_namer_set_name(
    mlx_node_namer namer,
    const mlx_array arr,
    const char* name) {
  try {
    mlx_node_namer_get_(namer).set_name(mlx_array_get_(arr), name);
    return 0;
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
}
extern "C" int mlx_node_namer_get_name(
    const char** name,
    mlx_node_namer namer,
    const mlx_array arr) {
  try {
    *name = mlx_node_namer_get_(namer).get_name(mlx_array_get_(arr)).c_str();
    return 0;
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
}
extern "C" int mlx_export_to_dot(
    FILE* os,
    const mlx_node_namer namer,
    const mlx_vector_array outputs) {
  try {
    mlx::core::export_to_dot(
        CFileOutputStream::as_lvalue(CFileOutputStream(os)),
        mlx_node_namer_get_(namer),
        mlx_vector_array_get_(outputs));
    return 0;
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
}`
	} else {
		code = `typedef struct mlx_node_namer_ {
  void* ctx;
} mlx_node_namer;

mlx_node_namer mlx_node_namer_new();
int mlx_node_namer_free(mlx_node_namer namer);
int mlx_node_namer_set_name(
    mlx_node_namer namer,
    const mlx_array arr,
    const char* name);
int mlx_node_namer_get_name(
    const char** name,
    mlx_node_namer namer,
    const mlx_array arr);
int mlx_export_to_dot(
    FILE* os,
    const mlx_node_namer namer,
    const mlx_vector_array outputs);
`
	}
	return Result{Code: code, Handled: true}
}

// ─── mlx_custom_function ────────────────────────────────────────────────────

func hookCustomFunction(impl bool) Result {
	var code string
	if impl {
		code = `extern "C" int mlx_custom_function(
    mlx_closure* res,
    const mlx_closure fun,
    const mlx_closure_custom fun_vjp,
    const mlx_closure_custom_jvp fun_jvp,
    const mlx_closure_custom_vmap fun_vmap) {
  try {
    using namespace mlx::core;
    auto fn = mlx_closure_get_(fun);
    auto vjp = fun_vjp.ctx ? std::make_optional(mlx_closure_custom_get_(fun_vjp)) : std::nullopt;
    auto jvp = fun_jvp.ctx ? std::make_optional(mlx_closure_custom_jvp_get_(fun_jvp)) : std::nullopt;
    auto vmap = fun_vmap.ctx ? std::make_optional(mlx_closure_custom_vmap_get_(fun_vmap)) : std::nullopt;
    auto result = mlx::core::custom_function(fn, vjp, jvp, vmap);
    mlx_closure_set_(*res, result);
    return 0;
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
}`
	} else {
		code = `int mlx_custom_function(
    mlx_closure* res,
    const mlx_closure fun,
    const mlx_closure_custom fun_vjp,
    const mlx_closure_custom_jvp fun_jvp,
    const mlx_closure_custom_vmap fun_vmap);
`
	}
	return Result{Code: code, Handled: true}
}

// ─── mlx_custom_vjp ─────────────────────────────────────────────────────────

func hookCustomVJP(impl bool) Result {
	var code string
	if impl {
		code = `extern "C" int mlx_custom_vjp(
    mlx_closure* res,
    const mlx_closure fun,
    const mlx_closure_custom fun_vjp) {
  try {
    using namespace mlx::core;
    auto fn = mlx_closure_get_(fun);
    auto vjp = mlx_closure_custom_get_(fun_vjp);
    auto result = mlx::core::custom_vjp(fn, vjp);
    mlx_closure_set_(*res, result);
    return 0;
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
}`
	} else {
		code = `int mlx_custom_vjp(
    mlx_closure* res,
    const mlx_closure fun,
    const mlx_closure_custom fun_vjp);
`
	}
	return Result{Code: code, Handled: true}
}

// ─── module: export ──────────────────────────────────────────────────────────

func hookModuleExport(impl bool) string {
	if impl {
		return `extern "C" int mlx_export_function(
    const char* file,
    const mlx_closure fun,
    const mlx_vector_array args,
    bool shapeless) {
  try {
    mlx::core::export_function(
        std::string(file),
        mlx_closure_get_(fun),
        mlx_vector_array_get_(args),
        shapeless);
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
  return 0;
}
extern "C" int mlx_export_function_kwargs(
    const char* file,
    const mlx_closure_kwargs fun,
    const mlx_vector_array args,
    const mlx_map_string_to_array kwargs,
    bool shapeless) {
  try {
    mlx::core::export_function(
        std::string(file),
        mlx_closure_kwargs_get_(fun),
        mlx_vector_array_get_(args),
        mlx_map_string_to_array_get_(kwargs),
        shapeless);
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
  return 0;
}

extern "C" mlx_function_exporter mlx_function_exporter_new(
    const char* file,
    const mlx_closure fun,
    bool shapeless) {
  try {
    return mlx_function_exporter_new_(
        mlx::core::exporter(std::string(file), mlx_closure_get_(fun), shapeless));
  } catch (std::exception& e) {
    mlx_error(e.what());
  }
  return {nullptr};
}

extern "C" int mlx_function_exporter_free(mlx_function_exporter xfunc) {
  try {
    mlx_function_exporter_free_(xfunc);
    return 0;
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
}

extern "C" int mlx_function_exporter_apply(
    const mlx_function_exporter xfunc,
    const mlx_vector_array args) {
  try {
    mlx_function_exporter_get_(xfunc)(mlx_vector_array_get_(args));
    return 0;
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
}

extern "C" int mlx_function_exporter_apply_kwargs(
    const mlx_function_exporter xfunc,
    const mlx_vector_array args,
    const mlx_map_string_to_array kwargs) {
  try {
    mlx_function_exporter_get_(xfunc)(
        mlx_vector_array_get_(args), mlx_map_string_to_array_get_(kwargs));
    return 0;
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
}

extern "C" mlx_imported_function mlx_imported_function_new(const char* file) {
  try {
    return mlx_imported_function_new_(mlx::core::import_function(std::string(file)));
  } catch (std::exception& e) {
    mlx_error(e.what());
  }
  return {nullptr};
}

extern "C" int mlx_imported_function_free(mlx_imported_function xfunc) {
  try {
    mlx_imported_function_free_(xfunc);
    return 0;
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
}

extern "C" int mlx_imported_function_apply(
    mlx_vector_array* res,
    const mlx_imported_function xfunc,
    const mlx_vector_array args) {
  try {
    mlx_vector_array_set_(
        *res, mlx_imported_function_get_(xfunc)(mlx_vector_array_get_(args)));
    return 0;
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
}

extern "C" int mlx_imported_function_apply_kwargs(
    mlx_vector_array* res,
    const mlx_imported_function xfunc,
    const mlx_vector_array args,
    const mlx_map_string_to_array kwargs) {
  try {
    mlx_vector_array_set_(
        *res,
        mlx_imported_function_get_(xfunc)(
            mlx_vector_array_get_(args), mlx_map_string_to_array_get_(kwargs)));
    return 0;
  } catch (std::exception& e) {
    mlx_error(e.what());
    return 1;
  }
}
`
	}
	// Header declarations
	return `int mlx_export_function(
    const char* file,
    const mlx_closure fun,
    const mlx_vector_array args,
    bool shapeless);
int mlx_export_function_kwargs(
    const char* file,
    const mlx_closure_kwargs fun,
    const mlx_vector_array args,
    const mlx_map_string_to_array kwargs,
    bool shapeless);

typedef struct mlx_function_exporter_ {
  void* ctx;
} mlx_function_exporter;
mlx_function_exporter mlx_function_exporter_new(
    const char* file,
    const mlx_closure fun,
    bool shapeless);
int mlx_function_exporter_free(mlx_function_exporter xfunc);
int mlx_function_exporter_apply(
    const mlx_function_exporter xfunc,
    const mlx_vector_array args);
int mlx_function_exporter_apply_kwargs(
    const mlx_function_exporter xfunc,
    const mlx_vector_array args,
    const mlx_map_string_to_array kwargs);

typedef struct mlx_imported_function_ {
  void* ctx;
} mlx_imported_function;
mlx_imported_function mlx_imported_function_new(const char* file);
int mlx_imported_function_free(mlx_imported_function xfunc);
int mlx_imported_function_apply(
    mlx_vector_array* res,
    const mlx_imported_function xfunc,
    const mlx_vector_array args);
int mlx_imported_function_apply_kwargs(
    mlx_vector_array* res,
    const mlx_imported_function xfunc,
    const mlx_vector_array args,
    const mlx_map_string_to_array kwargs);
`
}
