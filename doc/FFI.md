# GL pointer forwarding

The private GL `call` entry forwards uintptr arguments to cgo.Symbol.CallRaw.
It must have `//go:uintptrescapes` at the point where API wrappers convert their
Go buffer addresses. An annotation only on the inner cgo call cannot recover a
Go address after it has already become an integer at this outer entry.

This preserves output handles, buffer uploads, logs, uniforms and pointer tables
through synchronous native calls. Shader source strings remain separately rooted
through the existing typed pointer table and KeepAlive. GL calls that interpret
an argument as a buffer offset retain their numeric ABI semantics.

`gl/pointer_test.go` uses real libc memset writes from fresh stacks at varying
depths, checking independent output/canary values without needing OpenGL or a
desktop. Removing only this entry's annotation fails the regression even with
the repaired cgo entry. Existing string/GC callback regressions remain in place.

The repair is compatible with the existing Go 1.24 minimum. The FFI dependency's
raw entries also need their matching pointer contract for direct users; see
goexlib/cgo's PointerLifetime design. No ABI, exported GL API or native resource
ownership changes are introduced here.
