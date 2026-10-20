package gl

import (
	"runtime"
	"testing"
	"unsafe"

	"github.com/goexlib/cgo"
)

// Exercise the outer GL call bridge with a real native write, without a GL
// context. Fresh stacks and caller depths expose addresses lost by forwarding.
func TestCallOutputAddressSurvivesStackGrowth(t *testing.T) {
	lib := cgo.NewLazyLibrary("libc.so.6", "msvcrt.dll", "/usr/lib/libSystem.B.dylib")
	fn := cgo.Symbol(lib.NewSymbol("memset").Addr())
	for depth := 0; depth < 96; depth++ {
		result := make(chan [258]byte)
		go func() { result <- fillAtDepth(fn, depth) }()
		got := <-result
		if got[0] != 0x11 || got[257] != 0x22 {
			t.Fatalf("depth=%d canary overwritten", depth)
		}
		for i := 1; i < 257; i++ {
			if got[i] != 0xa5 {
				t.Fatalf("depth=%d byte=%d got=%#x want=0xa5", depth, i, got[i])
			}
		}
	}
}

//go:noinline
func fillAtDepth(fn cgo.Symbol, depth int) (out [258]byte) {
	var padding [256]byte
	padding[0] = byte(depth)
	if depth > 0 {
		out = fillAtDepth(fn, depth-1)
	} else {
		out = fillOutput(fn)
	}
	runtime.KeepAlive(padding)
	return
}

//go:noinline
func fillOutput(fn cgo.Symbol) (out [258]byte) {
	out[0], out[257] = 0x11, 0x22
	call(fn, uintptr(unsafe.Pointer(&out[1])), 0xa5, 256)
	return
}
