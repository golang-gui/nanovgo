package nanovgo

import "testing"

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
