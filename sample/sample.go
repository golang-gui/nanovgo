//go:build !js
// +build !js

package main

/*
	#define WIN32_LEAN_AND_MEAN 1
	#include <windows.h>
	#include <stdlib.h>
	static HMODULE ogl32dll = NULL;
	static void* getProcAddress(const char* name) {
		void* pf = wglGetProcAddress((LPCSTR) name);
		if (pf) {
			return pf;
		}
		if (ogl32dll == NULL) {
			ogl32dll = LoadLibraryA("opengl32.dll");
		}
		return GetProcAddress(ogl32dll, (LPCSTR) name);
	}
*/
import "C"
import (
	"fmt"
	"github.com/goexlib/cgo"
	"github.com/golang-gui/nanovgo"
	"github.com/golang-gui/nanovgo/gl"
	"github.com/golang-gui/nanovgo/perfgraph"
	"github.com/goxjs/glfw"
	"log"
	"runtime"
	"sample/demo"
)

var blowup bool
var premult bool

func key(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyEscape && action == glfw.Press {
		w.SetShouldClose(true)
	} else if key == glfw.KeySpace && action == glfw.Press {
		blowup = !blowup
	} else if key == glfw.KeyP && action == glfw.Press {
		premult = !premult
	}
}

type wglContext int

func (c wglContext) GetProcAddress(name string) (uintptr, error) {
	cName := cgo.CString(name)
	proc := C.getProcAddress((*C.char)(cName))
	runtime.KeepAlive(cName)
	return uintptr(proc), nil
}

type contextWatcher int

func (contextWatcher) OnMakeCurrent(context interface{}) {}

func (contextWatcher) OnDetach() {}

func main() {
	err := glfw.Init(contextWatcher(0))
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	// demo MSAA
	glfw.WindowHint(glfw.Samples, 4)

	window, err := glfw.CreateWindow(1000, 600, "NanoVGo", nil, nil)
	if err != nil {
		panic(err)
	}
	window.SetKeyCallback(key)
	window.MakeContextCurrent()

	err = gl.Init(func(symbol string) (fn uintptr, err error) {
		cName := cgo.CString(symbol)
		proc := C.getProcAddress((*C.char)(cName))
		runtime.KeepAlive(cName)
		return uintptr(proc), nil
	})
	if err != nil {
		panic(err)
	}

	ctx, err := nanovgo.NewContext(wglContext(0), 0 /*nanovgo.AntiAlias | nanovgo.StencilStrokes | nanovgo.Debug*/)
	defer ctx.Delete()

	if err != nil {
		panic(err)
	}

	demoData := LoadDemo(ctx)

	glfw.SwapInterval(0)

	fps := perfgraph.NewPerfGraph("Frame Time", "sans")

	for !window.ShouldClose() {
		t, _ := fps.UpdateGraph()

		fbWidth, fbHeight := window.GetFramebufferSize()
		winWidth, winHeight := window.GetSize()
		mx, my := window.GetCursorPos()

		pixelRatio := float32(fbWidth) / float32(winWidth)
		gl.Viewport(0, 0, fbWidth, fbHeight)
		if premult {
			gl.ClearColor(0, 0, 0, 0)
		} else {
			gl.ClearColor(0.3, 0.3, 0.32, 1.0)
		}
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		gl.Enable(gl.CULL_FACE)
		gl.Disable(gl.DEPTH_TEST)

		ctx.BeginFrame(winWidth, winHeight, pixelRatio)

		demo.RenderDemo(ctx, float32(mx), float32(my), float32(winWidth), float32(winHeight), t, blowup, demoData)
		fps.RenderGraph(ctx, 5, 5)

		ctx.EndFrame()

		gl.Enable(gl.DEPTH_TEST)
		window.SwapBuffers()
		glfw.PollEvents()
	}

	demoData.FreeData(ctx)
}

func LoadDemo(ctx *nanovgo.Context) *demo.DemoData {
	d := &demo.DemoData{}
	for i := 0; i < 12; i++ {
		path := fmt.Sprintf("images/image%d.jpg", i+1)
		d.Images = append(d.Images, ctx.CreateImage(path, 0))
		if d.Images[i] == 0 {
			log.Fatalf("Could not load %s", path)
		}
	}

	d.FontIcons = ctx.CreateFont("icons", "entypo.ttf")
	if d.FontIcons == -1 {
		log.Fatalln("Could not add font icons.")
	}
	d.FontNormal = ctx.CreateFont("sans", "Roboto-Regular.ttf")
	if d.FontNormal == -1 {
		log.Fatalln("Could not add font italic.")
	}
	d.FontBold = ctx.CreateFont("sans-bold", "Roboto-Bold.ttf")
	if d.FontBold == -1 {
		log.Fatalln("Could not add font bold.")
	}
	return d
}
