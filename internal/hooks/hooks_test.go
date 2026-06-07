package hooks

import (
	"strings"
	"testing"
)

// ─── Apply: registry dispatch ────────────────────────────────────────────────

func TestApply_UnknownFunction_NotHandled(t *testing.T) {
	r := Apply("mlx_nonexistent_function", false)
	if r.Handled {
		t.Error("expected Handled=false for unknown function")
	}
	if r.Code != "" {
		t.Errorf("expected empty Code for unknown function, got %q", r.Code)
	}
}

func TestApply_KnownFunction_IsHandled(t *testing.T) {
	known := []string{
		"mlx_fast_metal_kernel",
		"mlx_load_gguf",
		"mlx_save_gguf",
		"mlx_export_to_dot",
		"mlx_custom_function",
		"mlx_custom_vjp",
	}
	for _, name := range known {
		r := Apply(name, false)
		if !r.Handled {
			t.Errorf("Apply(%q, false): expected Handled=true", name)
		}
	}
}

func TestApply_AllHooksReturnNonEmptyCode(t *testing.T) {
	known := []string{
		"mlx_fast_metal_kernel",
		"mlx_load_gguf",
		"mlx_save_gguf",
		"mlx_export_to_dot",
		"mlx_custom_function",
		"mlx_custom_vjp",
	}
	for _, name := range known {
		for _, impl := range []bool{false, true} {
			r := Apply(name, impl)
			if r.Code == "" {
				t.Errorf("Apply(%q, impl=%v): expected non-empty Code", name, impl)
			}
		}
	}
}

// ─── mlx_load_gguf ───────────────────────────────────────────────────────────

func TestHookLoadGGUF_Header(t *testing.T) {
	r := Apply("mlx_load_gguf", false)
	if !r.Handled {
		t.Fatal("expected Handled=true")
	}
	if !strings.Contains(r.Code, "mlx_load_gguf") {
		t.Error("header should contain function name")
	}
	if !strings.Contains(r.Code, "mlx_io_gguf*") {
		t.Error("header should declare out-param mlx_io_gguf*")
	}
	if strings.Contains(r.Code, "extern \"C\"") {
		t.Error("header should not contain extern C")
	}
}

func TestHookLoadGGUF_Impl(t *testing.T) {
	r := Apply("mlx_load_gguf", true)
	if !strings.Contains(r.Code, "extern \"C\"") {
		t.Error("impl should contain extern C")
	}
	if !strings.Contains(r.Code, "mlx::core::load_gguf") {
		t.Error("impl should call mlx::core::load_gguf")
	}
	if !strings.Contains(r.Code, "mlx_error") {
		t.Error("impl should have error handling")
	}
}

// ─── mlx_save_gguf ───────────────────────────────────────────────────────────

func TestHookSaveGGUF_Header(t *testing.T) {
	r := Apply("mlx_save_gguf", false)
	if !strings.Contains(r.Code, "mlx_save_gguf") {
		t.Error("header should contain function name")
	}
	if strings.Contains(r.Code, "extern \"C\"") {
		t.Error("header should not contain extern C")
	}
}

func TestHookSaveGGUF_Impl(t *testing.T) {
	r := Apply("mlx_save_gguf", true)
	if !strings.Contains(r.Code, "mlx::core::save_gguf") {
		t.Error("impl should call mlx::core::save_gguf")
	}
	if !strings.Contains(r.Code, "mlx_error") {
		t.Error("impl should have error handling")
	}
}

// ─── mlx_export_to_dot ───────────────────────────────────────────────────────

func TestHookExportToDot_Header(t *testing.T) {
	r := Apply("mlx_export_to_dot", false)
	if !strings.Contains(r.Code, "mlx_node_namer") {
		t.Error("header should define mlx_node_namer struct")
	}
	if !strings.Contains(r.Code, "mlx_node_namer_new") {
		t.Error("header should declare mlx_node_namer_new")
	}
	if !strings.Contains(r.Code, "mlx_node_namer_free") {
		t.Error("header should declare mlx_node_namer_free")
	}
	if !strings.Contains(r.Code, "mlx_node_namer_set_name") {
		t.Error("header should declare set_name")
	}
	if !strings.Contains(r.Code, "mlx_node_namer_get_name") {
		t.Error("header should declare get_name")
	}
}

func TestHookExportToDot_Impl(t *testing.T) {
	r := Apply("mlx_export_to_dot", true)
	if !strings.Contains(r.Code, "mlx::core::NodeNamer") {
		t.Error("impl should use mlx::core::NodeNamer")
	}
	// Impl should contain all 4 function implementations
	for _, fn := range []string{"mlx_node_namer_new", "mlx_node_namer_free", "mlx_node_namer_set_name", "mlx_node_namer_get_name"} {
		if !strings.Contains(r.Code, fn) {
			t.Errorf("impl missing function: %s", fn)
		}
	}
}

// ─── mlx_custom_function ─────────────────────────────────────────────────────

func TestHookCustomFunction_Header(t *testing.T) {
	r := Apply("mlx_custom_function", false)
	if !strings.Contains(r.Code, "mlx_custom_function") {
		t.Error("header should declare mlx_custom_function")
	}
	if !strings.Contains(r.Code, "mlx_closure_custom") {
		t.Error("header should include mlx_closure_custom parameter")
	}
	if !strings.Contains(r.Code, "mlx_closure_custom_jvp") {
		t.Error("header should include jvp parameter")
	}
	if !strings.Contains(r.Code, "mlx_closure_custom_vmap") {
		t.Error("header should include vmap parameter")
	}
}

func TestHookCustomFunction_Impl(t *testing.T) {
	r := Apply("mlx_custom_function", true)
	if !strings.Contains(r.Code, "mlx::core::custom") {
		t.Error("impl should call mlx::core::custom")
	}
	if !strings.Contains(r.Code, "std::nullopt") {
		t.Error("impl should handle optional parameters with std::nullopt")
	}
	if !strings.Contains(r.Code, "mlx_error") {
		t.Error("impl should have error handling")
	}
}

// ─── mlx_custom_vjp ──────────────────────────────────────────────────────────

func TestHookCustomVJP_Header(t *testing.T) {
	r := Apply("mlx_custom_vjp", false)
	if !strings.Contains(r.Code, "mlx_custom_vjp") {
		t.Error("header should declare mlx_custom_vjp")
	}
	if !strings.Contains(r.Code, "mlx_closure_custom") {
		t.Error("header should include mlx_closure_custom parameter")
	}
}

func TestHookCustomVJP_Impl(t *testing.T) {
	r := Apply("mlx_custom_vjp", true)
	if !strings.Contains(r.Code, "mlx::core::custom_vjp") {
		t.Error("impl should call mlx::core::custom_vjp")
	}
	if !strings.Contains(r.Code, "mlx_error") {
		t.Error("impl should have error handling")
	}
}

// ─── mlx_fast_metal_kernel ───────────────────────────────────────────────────

func TestHookFastMetalKernel_Header(t *testing.T) {
	r := Apply("mlx_fast_metal_kernel", false)
	if strings.Contains(r.Code, "mlx_fast_metal_custom_kernel_config") {
		t.Error("Bug 1 regression: header must not contain '_custom_kernel_config'")
	}
	if !strings.Contains(r.Code, "mlx_fast_metal_kernel_config") {
		t.Error("header should define mlx_fast_metal_kernel_config struct")
	}
	if !strings.Contains(r.Code, "mlx_fast_metal_kernel_new") {
		t.Error("header should declare mlx_fast_metal_kernel_new")
	}
}

func TestHookFastMetalKernel_Impl(t *testing.T) {
	r := Apply("mlx_fast_metal_kernel", true)
	if !strings.Contains(r.Code, "mlx::core::fast::metal_kernel") {
		t.Error("impl should call mlx::core::fast::metal_kernel")
	}
}

// ─── ApplyModule ─────────────────────────────────────────────────────────────

func TestApplyModule_UnknownModule_ReturnsEmpty(t *testing.T) {
	code := ApplyModule("nonexistent_module", false)
	if code != "" {
		t.Errorf("unknown module should return empty string, got %q", code)
	}
}

func TestApplyModule_Export_Header(t *testing.T) {
	code := ApplyModule("export", false)
	if code == "" {
		t.Fatal("export module header should not be empty")
	}
	if !strings.Contains(code, "mlx_export_function") {
		t.Error("export header should contain mlx_export_function")
	}
	if !strings.Contains(code, "mlx_function_exporter") {
		t.Error("export header should contain mlx_function_exporter")
	}
	if !strings.Contains(code, "mlx_imported_function") {
		t.Error("export header should contain mlx_imported_function")
	}
}

func TestApplyModule_Export_Impl(t *testing.T) {
	code := ApplyModule("export", true)
	if code == "" {
		t.Fatal("export module impl should not be empty")
	}
	if !strings.Contains(code, "extern \"C\"") {
		t.Error("export impl should contain extern C functions")
	}
	if !strings.Contains(code, "mlx_export_function") {
		t.Error("export impl should implement mlx_export_function")
	}
}

func TestApplyModule_Export_HeaderHasNoExternC(t *testing.T) {
	code := ApplyModule("export", false)
	if strings.Contains(code, "extern \"C\"") {
		t.Error("export module header should not contain extern C")
	}
}

// ─── Header vs Impl contract ─────────────────────────────────────────────────

func TestAllHooks_HeaderHasNoExternC(t *testing.T) {
	known := []string{
		"mlx_load_gguf",
		"mlx_save_gguf",
		"mlx_custom_function",
		"mlx_custom_vjp",
	}
	for _, name := range known {
		r := Apply(name, false)
		if strings.Contains(r.Code, "extern \"C\"") {
			t.Errorf("%s: header must not contain extern C", name)
		}
	}
}

func TestAllHooks_ImplHasExternC(t *testing.T) {
	known := []string{
		"mlx_load_gguf",
		"mlx_save_gguf",
		"mlx_custom_function",
		"mlx_custom_vjp",
		"mlx_export_to_dot",
		"mlx_fast_metal_kernel",
	}
	for _, name := range known {
		r := Apply(name, true)
		if !strings.Contains(r.Code, "extern \"C\"") {
			t.Errorf("%s: impl should contain extern C", name)
		}
	}
}

// ─── Registry completeness ───────────────────────────────────────────────────

func TestRegistry_HasExpectedEntries(t *testing.T) {
	expected := []string{
		"mlx_fast_metal_kernel",
		"mlx_load_gguf",
		"mlx_save_gguf",
		"mlx_export_to_dot",
		"mlx_custom_function",
		"mlx_custom_vjp",
	}
	for _, name := range expected {
		if _, ok := registry[name]; !ok {
			t.Errorf("registry missing expected entry: %s", name)
		}
	}
}

func TestModuleRegistry_HasExportEntry(t *testing.T) {
	if _, ok := moduleRegistry["export"]; !ok {
		t.Error("moduleRegistry should contain 'export' entry")
	}
}
