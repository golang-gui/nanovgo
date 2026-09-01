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
