package types_test

import (
	"fmt"
	"github.com/WhoseBiasDoYallSeek/mlx-cgo/internal/types"
	"testing"
)

func check(t *testing.T, label, cpptype, wantCArg, wantCToCPP, wantAssign string) {
	t.Helper()
	ti, ok := types.FindByCPP(cpptype)
	if !ok {
		t.Errorf("❌ %s: NOT FOUND", label)
		return
	}
	carg := ""
	if ti.CArg != nil {
		carg = ti.CArg("a")
	}
	ctocpp := ""
	if ti.CToCPP != nil {
		ctocpp = ti.CToCPP("a")
	}
	assign := ""
	if ti.CAssignFromCPP != nil {
		assign = ti.CAssignFromCPP("res", "cpp_val")
	}
	if carg != wantCArg {
		t.Errorf("%s CArg   = %q, want %q", label, carg, wantCArg)
	}
	if ctocpp != wantCToCPP {
		t.Errorf("%s CToCPP = %q, want %q", label, ctocpp, wantCToCPP)
	}
	if assign != wantAssign {
		t.Errorf("%s Assign = %q, want %q", label, assign, wantAssign)
	}
	if !t.Failed() {
		fmt.Printf("✅ %s\n", label)
	}
}

func TestTypes(t *testing.T) {
	check(t, "mlx_array", "mlx::core::array",
		"const mlx_array a", "mlx_array_get_(a)", "mlx_array_set_(*res, cpp_val)")
	check(t, "int", "int",
		"int a", "a", "*res = cpp_val")
	check(t, "float", "float",
		"float a", "a", "*res = cpp_val")
	check(t, "std::vector<int>", "std::vector<int>",
		"const int* a, size_t a_num",
		"std::vector<int>(a, a + a_num)",
		"res = cpp_val.data(); res_num = cpp_val.size()")
	check(t, "mlx::core::Shape", "mlx::core::Shape",
		"const int* a, size_t a_num",
		"mlx::core::Shape(a, a + a_num)",
		"res = cpp_val.data(); res_num = cpp_val.size()")
	check(t, "std::string", "std::string",
		"const char* a", "std::string(a)", "res = cpp_val.c_str()")
	check(t, "void", "void",
		"", "", "cpp_val")
	check(t, "mlx::core::Dtype", "mlx::core::Dtype",
		"mlx_dtype a", "mlx_dtype_to_cpp(a)", "res = mlx_dtype_to_c((int)((cpp_val).val))")

	// Alt lookups
	ti, ok := types.FindByCPP("StreamOrDevice")
	if !ok || ti.CName != "mlx_stream" {
		t.Errorf("StreamOrDevice alt: ok=%v CName=%q, want mlx_stream", ok, ti.CName)
	} else {
		fmt.Println("✅ StreamOrDevice alt")
	}

	ti2, ok2 := types.FindByCPP("array")
	if !ok2 || ti2.CName != "mlx_array" {
		t.Errorf("array alt: ok=%v CName=%q, want mlx_array", ok2, ti2.CName)
	} else {
		fmt.Println("✅ array alt")
	}
}
