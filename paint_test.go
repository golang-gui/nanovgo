package nanovgo

import (
	"slices"
	"testing"
)

func TestBoxShadowPaint(t *testing.T) {
	color := RGBA(10, 20, 30, 128)
	paint := BoxShadow(10, 20, 40, 30, 7, 12, color)
	if paint.kind != paintBoxShadow {
		t.Fatalf("kind = %v, want box shadow", paint.kind)
	}
	if paint.extent != [2]float32{20, 15} || paint.radius != 7 || paint.feather != 6 {
		t.Fatalf("unexpected geometry: extent=%v radius=%v sigma=%v", paint.extent, paint.radius, paint.feather)
	}
	if paint.innerColor != color || paint.outerColor != color || paint.image != 0 {
		t.Fatalf("unexpected color/image state: %+v", paint)
	}

	paint.setPaintColor(color)
	if paint.kind != paintDefault {
		t.Fatalf("solid color retained box shadow kind: %v", paint.kind)
	}
}

func TestBoxShadowVaryingUniforms(t *testing.T) {
	paint := BoxShadowVarying(10, 20, 40, 30, 1, 2, 3, 4, 8, RGBA(20, 40, 80, 128))
	if paint.shadowCorners != ([4]float32{1, 2, 3, 4}) || paint.feather != 4 {
		t.Fatal("四角或sigma错误", paint)
	}
	var uniform glFragUniforms
	context := new(glContext)
	if err := context.convertPaint(&uniform, &paint, &nvgScissor{extent: [2]float32{-1, -1}}, 1, 1, -1); err != nil {
		t.Fatal(err)
	}
	if len(uniform) != 44 || uniform[28] != 1 || uniform[29] != 2 || uniform[30] != 3 || uniform[31] != 4 {
		t.Fatal("阴影改变了uniform布局或半径映射", uniform)
	}
	paint.setPaintColor(RGBA(1, 2, 3, 255))
	if paint.shadowCorners != ([4]float32{}) {
		t.Fatal("普通画刷保留阴影几何")
	}
}

func TestCurvedFillUsesCoverageStencil(t *testing.T) {
	for _, tc := range []struct {
		name      string
		alpha     uint8
		points    int
		operation CompositeOperation
		want      glnvgCallType
	}{
		{"半透明曲线", 128, 8, SourceOver, glnvgFILL},
		{"擦除曲线", 255, 8, DestinationOut, glnvgFILL},
		{"不透明曲线快捷路径", 255, 8, SourceOver, glnvgCONVEXFILL},
		{"矩形快捷路径", 128, 4, SourceOver, glnvgCONVEXFILL},
	} {
		t.Run(tc.name, func(t *testing.T) {
			context := &glContext{flags: AntiAlias}
			params := &glParams{context: context}
			paint := Paint{innerColor: RGBA(20, 40, 60, tc.alpha), outerColor: RGBA(20, 40, 60, tc.alpha), xform: IdentityMatrix()}
			params.renderFill(&paint, tc.operation, &nvgScissor{extent: [2]float32{-1, -1}}, 1, [4]float32{0, 0, 10, 10}, []nvgPath{{convex: true, count: tc.points}})
			if len(context.calls) != 1 || context.calls[0].callType != tc.want {
				t.Fatal("覆盖率路径选择错误", context.calls)
			}
		})
	}
}

// 无需GL：尖端的面积覆盖数据必须随命令排队，后续复用路径缓存不能污染它。
func TestSharpFillCoverageCommand(t *testing.T) {
	params := &glParams{isEdgeAntiAlias: true, context: &glContext{flags: AntiAlias}}
	c := &Context{params: params}
	c.BeginFrame(160, 80, 1)
	c.BeginPath()
	c.MoveTo(0, 16)
	const k float32 = .5522847498307936
	c.BezierTo(0, 16*(1-k), 16*(1-k), 0, 16, 0)
	c.LineTo(120, 0)
	c.LineTo(120, 4)
	c.LineTo(16, 4)
	c.BezierTo(16*(1-k), 4, 0, 16-12*k, 0, 16)
	c.ClosePath()
	c.Fill()
	call := params.context.calls[0]
	if call.callType != glnvgFILL || call.coverageCount < 3 || call.coverageCount > maxAreaCoverageEdges || call.coveragePixel != 1 {
		t.Fatal("尖端没有记录面积覆盖命令", call)
	}
	edges := slices.Clone(params.context.coverageEdges)
	if len(edges) != call.coverageCount*4 {
		t.Fatal("覆盖边数与数据不符", len(edges), call.coverageCount)
	}
	// 验证记录的是原始轮廓，不是已移动的AA边缘；数据逐边首尾相连。
	for i := 0; i < call.coverageCount; i++ {
		j := (i + 1) % call.coverageCount
		if edges[i*4+2] != edges[j*4] || edges[i*4+3] != edges[j*4+1] {
			t.Fatal("覆盖轮廓没有闭合", i)
		}
	}
	c.BeginPath()
	c.Rect(40, 40, 20, 10)
	c.Fill()
	if !slices.Equal(edges, params.context.coverageEdges) || params.context.calls[1].coverageCount != 0 || params.context.calls[1].callType != glnvgCONVEXFILL {
		t.Fatal("复用缓存污染了尖端命令，或普通矩形失去快捷路径")
	}
	c.CancelFrame()
	if len(params.context.coverageEdges) != 0 || len(params.context.calls) != 0 {
		t.Fatal("取消帧保留了覆盖数据")
	}
}

func TestAreaCoverageSelection(t *testing.T) {
	corner := nvgPoint{x: 0, y: 0, flags: nvgPtLEFT | nvgPtBEVEL}
	triangle := []nvgPoint{corner, {x: 10, y: 1}, {x: 10, y: -1}}
	rectangle := []nvgPoint{{x: 20, y: 0}, {x: 24, y: 0}, {x: 24, y: 4}, {x: 20, y: 4}}
	for _, tc := range []struct {
		name  string
		paths []nvgPath
		want  bool
	}{
		{"普通矩形", []nvgPath{{points: make([]nvgPoint, 4)}}, false},
		{"无原始轮廓", []nvgPath{{count: 4}}, false},
		{"单尖端", []nvgPath{{winding: Solid, points: triangle}}, true},
		{"凹角", []nvgPath{{winding: Solid, points: []nvgPoint{{flags: nvgPtBEVEL}, {}, {}}}}, false},
		{"多轮廓", []nvgPath{{winding: Solid, points: triangle}, {winding: Solid, points: rectangle}}, true},
		{"孔洞保持旧填充", []nvgPath{{winding: Hole, points: triangle}}, false},
		{"重合轮廓", []nvgPath{{winding: Solid, points: triangle}, {winding: Solid, points: triangle}}, false},
		{"自交", []nvgPath{{winding: Solid, points: []nvgPoint{corner, {x: 10, y: 10}, {x: 0, y: 10}, {x: 10, y: 0}}}}, false},
		{"嵌套轮廓", []nvgPath{{winding: Solid, points: triangle}, {winding: Solid, points: []nvgPoint{{x: 8, y: 0}, {x: 9, y: .2}, {x: 9, y: -.2}}}}, false},
		{"超出uniform预算", []nvgPath{{points: append([]nvgPoint{corner}, make([]nvgPoint, maxAreaCoverageEdges)...)}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := needsAreaCoverage(tc.paths); got != tc.want {
				t.Fatal("覆盖率路径选择错误", got, tc.want)
			}
		})
	}
}
