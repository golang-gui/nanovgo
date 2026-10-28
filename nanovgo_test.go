package nanovgo

import (
	"testing"
)

func TestStateInit(t *testing.T) {
	c := Context{}
	c.Save()
	c.Reset()

	topState := c.getState()
	if topState.alpha != 1.0 {
		t.Errorf("initial alpha should be 1.0, but %f", topState.alpha)
	}
}

func TestStateSaveRestore(t *testing.T) {
	c := Context{}
	c.Save()
	c.Reset()

	topState := c.getState()
	topState.alpha = 0.5

	c.Save()

	nextState := c.getState()
	if nextState.alpha != 0.5 {
		t.Errorf("initial alpha should be same with parent's one, but %f", nextState.alpha)
	}
	nextState.alpha = 0.75

	c.Restore()

	topStateAgain := c.getState()
	if topStateAgain.alpha != 0.5 {
		t.Errorf("Restore() should set saved alpha, but %f", topStateAgain.alpha)
	}
}

func equal(a1, a2 TransformMatrix) bool {
	for i := 0; i < 6; i++ {
		if a1[i] != a2[i] {
			return false
		}
	}
	return true
}

func TestStateSaveRestore2(t *testing.T) {
	c := Context{}
	c.Save()
	c.Reset()

	topState := c.getState()
	topState.xform = TranslateMatrix(10, 5)

	c.Save()

	nextState := c.getState()
	if !equal(nextState.xform, TranslateMatrix(10, 5)) {
		t.Errorf("initial xform should be same with parent's one, but %v", nextState.xform)
	}
	nextState.xform = ScaleMatrix(20, 30)

	c.Restore()

	topStateAgain := c.getState()
	if !equal(topStateAgain.xform, TranslateMatrix(10, 5)) {
		t.Errorf("Restore() should set saved xform, but %v", topStateAgain.xform)
	}
}

// 命令排队后恢复状态，不能覆盖已排队的合成方式；不需要 GL 上下文。
func TestCompositeCommandsKeepTheirOwnState(t *testing.T) {
	params := &glParams{isEdgeAntiAlias: true, context: &glContext{flags: AntiAlias}}
	c := &Context{params: params}
	c.BeginFrame(100, 100, 1)
	c.BeginPath()
	c.Rect(2, 2, 20, 20)
	c.Fill()
	c.Save()
	c.SetCompositeOperation(DestinationOut)
	c.BeginPath()
	c.Rect(4, 4, 8, 8)
	c.Fill()
	c.Stroke()
	c.Restore()
	c.BeginPath()
	c.Rect(40, 40, 10, 10)
	c.Fill()
	calls := params.context.calls
	want := []CompositeOperation{SourceOver, DestinationOut, DestinationOut, SourceOver}
	if len(calls) != len(want) {
		t.Fatalf("commands=%d, want %d", len(calls), len(want))
	}
	for i, op := range want {
		if calls[i].composite != op {
			t.Fatalf("command %d: operation=%v want %v", i, calls[i].composite, op)
		}
	}
	c.SetCompositeOperation(DestinationOut)
	vertices := []nvgVertex{{x: 1, y: 1}, {x: 2, y: 1}, {x: 2, y: 2}}
	state := c.getState()
	params.renderTriangles(&state.fill, state.composite, &state.scissor, vertices)
	params.renderTriangleStrip(&state.fill, state.composite, &state.scissor, vertices)
	for _, call := range params.context.calls[4:] {
		if call.composite != DestinationOut {
			t.Fatal("triangle command lost operation")
		}
	}
	c.CancelFrame()
	if len(params.context.calls) != 0 {
		t.Fatal("cancel kept pending commands")
	}
	c.BeginFrame(100, 100, 2)
	if c.getState().composite != SourceOver {
		t.Fatal("BeginFrame retained composite operation")
	}
}
