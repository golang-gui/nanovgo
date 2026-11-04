package nanovgo

import (
	"errors"
	"testing"
)

type failingCreateParams struct {
	nvgParams
	err      error
	deleted  int
	textures int
}

func (p *failingCreateParams) renderCreate() error { return p.err }
func (p *failingCreateParams) renderDelete()       { p.deleted++ }
func (p *failingCreateParams) renderCreateTexture(nvgTextureType, int, int, ImageFlags, []byte) int {
	p.textures++
	return 0
}

func TestContextCreateFailure(t *testing.T) {
	// 原生初始化失败时必须返回原错误并清理，不能继续创建字体纹理或返回半成品。
	want := errors.New("shader compilation failed")
	p := &failingCreateParams{err: want}
	context, err := createInternal(p)
	if context != nil || !errors.Is(err, want) || p.deleted != 1 || p.textures != 0 {
		t.Fatalf("初始化失败未正确结束：context=%v err=%v deletes=%d textures=%d", context, err, p.deleted, p.textures)
	}
}
