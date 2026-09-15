package gl

import (
	"runtime"
	"slices"
	"strings"
	"testing"
	"unsafe"

	"github.com/goexlib/cgo"
)

// Call through the same native trampoline as GL without needing a display.
// GC during the foreign call exercises lifetime through uintptr conversion.
func TestTemporaryLocationNames(t *testing.T) {
	oldAttrib, oldBind, oldUniform := glGetAttribLocation, glBindAttribLocation, glGetUniformLocation
	defer func() {
		glGetAttribLocation, glBindAttribLocation, glGetUniformLocation = oldAttrib, oldBind, oldUniform
	}()
	var names []string
	read := func(p unsafe.Pointer) { runtime.GC(); names = append(names, cgo.GoString(p)) }
	glGetAttribLocation = cgo.Symbol(cgo.NewCallback(func(program uintptr, name unsafe.Pointer) uintptr { read(name); return 7 }))
	glBindAttribLocation = cgo.Symbol(cgo.NewCallback(func(program, index uintptr, name unsafe.Pointer) uintptr { read(name); return 0 }))
	glGetUniformLocation = cgo.Symbol(cgo.NewCallback(func(program uintptr, name unsafe.Pointer) uintptr { read(name); return 9 }))
	name := strings.Repeat("dynamic_name_", 100)
	if got := GetAttribLocation(Program{1}, name); got.Value != 7 {
		t.Fatal(got)
	}
	BindAttribLocation(Program{1}, 2, name)
	if got := GetUniformLocation(Program{1}, name); got.Value != 9 {
		t.Fatal(got)
	}
	if !slices.Equal(names, []string{name, name, name}) {
		t.Fatal("temporary names corrupted")
	}
}

func TestTemporaryShaderSources(t *testing.T) {
	old := glShaderSource
	defer func() { glShaderSource = old }()
	var got []string
	var nullEntry bool
	glShaderSource = cgo.Symbol(cgo.NewCallback(func(shader, count uintptr, sources unsafe.Pointer, lengths uintptr) uintptr {
		runtime.GC()
		got = nil
		for _, ptr := range unsafe.Slice((*unsafe.Pointer)(sources), int(count)) {
			if ptr == nil {
				nullEntry = true
			}
			got = append(got, cgo.GoString(ptr))
		}
		return 0
	}))
	want := []string{"#version 150\n", "", strings.Repeat("void main() {}\n", 100)}
	ShaderSource(Shader{1}, want...)
	if nullEntry || !slices.Equal(got, want) {
		t.Fatalf("shader strings: null=%v got=%q", nullEntry, got)
	}
	ShaderSource(Shader{1})
	if len(got) != 0 {
		t.Fatal("empty source list changed")
	}
}
