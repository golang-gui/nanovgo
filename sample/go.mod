module sample

go 1.24

replace github.com/golang-gui/nanovgo => ../

require (
	github.com/goexlib/cgo v0.0.1
	github.com/golang-gui/nanovgo v0.0.0-00010101000000-000000000000
	github.com/goxjs/glfw v0.0.0-20230704040236-622eb27e272a
)

require (
	github.com/ebitengine/purego v0.10.0 // indirect
	github.com/go-gl/glfw/v3.3/glfw v0.0.0-20250301202403-da16c1255728 // indirect
	github.com/gopherjs/gopherjs v1.20.1 // indirect
	honnef.co/go/js/dom v0.0.0-20250304181735-b5e52f05e89d // indirect
)
