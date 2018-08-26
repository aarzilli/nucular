package nucular

import (
	"image"
	"testing"
)

func drawFillOver_Normal(dst *image.RGBA, r image.Rectangle, sr, sg, sb, sa uint32) {
	const m = 1<<16 - 1
	// The 0x101 is here for the same reason as in drawRGBA.
	a := (m - sa) * 0x101
	i0 := dst.PixOffset(r.Min.X, r.Min.Y)
	i1 := i0 + r.Dx()*4
	for y := r.Min.Y; y != r.Max.Y; y++ {
		for i := i0; i < i1; i += 4 {
			dr := &dst.Pix[i+0]
			dg := &dst.Pix[i+1]
			db := &dst.Pix[i+2]
			da := &dst.Pix[i+3]

			*dr = uint8((uint32(*dr)*a/m + sr) >> 8)
			*dg = uint8((uint32(*dg)*a/m + sg) >> 8)
			*db = uint8((uint32(*db)*a/m + sb) >> 8)
			*da = uint8((uint32(*da)*a/m + sa) >> 8)
		}
		i0 += dst.Stride
		i1 += dst.Stride
	}
}

func drawFillOver_NoPtr(dst *image.RGBA, r image.Rectangle, sr, sg, sb, sa uint32) {
	const m = 1<<16 - 1
	// The 0x101 is here for the same reason as in drawRGBA.
	a := (m - sa) * 0x101
	i0 := dst.PixOffset(r.Min.X, r.Min.Y)
	i1 := i0 + r.Dx()*4
	for y := r.Min.Y; y != r.Max.Y; y++ {
		for i := i0; i < i1; i += 4 {
			dst.Pix[i+0] = uint8((uint32(dst.Pix[i+0])*a/m + sr) >> 8)
			dst.Pix[i+1] = uint8((uint32(dst.Pix[i+1])*a/m + sg) >> 8)
			dst.Pix[i+2] = uint8((uint32(dst.Pix[i+2])*a/m + sb) >> 8)
			dst.Pix[i+3] = uint8((uint32(dst.Pix[i+3])*a/m + sa) >> 8)
		}
		i0 += dst.Stride
		i1 += dst.Stride
	}
}

func drawFillOver_SIMD(dst *image.RGBA, r image.Rectangle, sr, sg, sb, sa uint32) {
	const m = 1<<16 - 1
	a := (m - sa) * 0x101
	adivm := a / m
	i0 := dst.PixOffset(r.Min.X, r.Min.Y)
	i1 := i0 + r.Dx()*4
	drawFillOver_SIMD_internal(&dst.Pix[0], i0, i1, dst.Stride, r.Max.Y-r.Min.Y, adivm, sr, sg, sb, sa)
}

func clearImg(b *image.RGBA) {
	for i := 0; i < len(b.Pix); i += 4 {
		b.Pix[i+0] = 50
		b.Pix[i+1] = 50
		b.Pix[i+2] = 50
		b.Pix[i+3] = 255
	}
}

func checkUniform(t *testing.T, b *image.RGBA, tgtr, tgtg, tgtb, tgta uint8) {
	ok := true
	for i := 0; i < len(b.Pix); i += 4 {
		if b.Pix[i+0] != tgtr {
			ok = false
			t.Errorf("mismatch at pixel %d (red) %d %d\n", i/4, b.Pix[i+0], tgtr)
		}
		if b.Pix[i+1] != tgtg {
			ok = false
			t.Errorf("mismatch at pixel %d (green) %d %d\n", i/4, b.Pix[i+1], tgtg)
		}
		if b.Pix[i+2] != tgtb {
			ok = false
			t.Errorf("mismatch at pixel %d (blue) %d %d\n", i/4, b.Pix[i+2], tgtb)
		}
		if b.Pix[i+3] != tgta {
			ok = false
			t.Errorf("mismatch at pixel %d (alpha) %d %d\n", i/4, b.Pix[i+3], tgta)
		}
		if !ok {
			t.Fatal("previous errors")
		}
	}
	outr, outg, outb, outa := b.Pix[0], b.Pix[1], b.Pix[2], b.Pix[3]
	t.Logf("color %d %d %d %d\n", outr, outg, outb, outa)
}

type fillOverFunc func(b *image.RGBA, r image.Rectangle, sr, sg, sb, sa uint32)

func testFillOver(t *testing.T, b *image.RGBA, fo fillOverFunc) {
	clearImg(b)
	fo(b, b.Bounds(), 12850, 14906, 15677, 57825)
	checkUniform(t, b, 56, 64, 67, 255)
}

func TestDrawFillOver(t *testing.T) {
	b := image.NewRGBA(image.Rect(0, 0, 2550, 1400))
	testFillOver(t, b, drawFillOver_Normal)
	testFillOver(t, b, drawFillOver_NoPtr)
	testFillOver(t, b, drawFillOver_SIMD)
}

func benchFillOver(bnc *testing.B, fo fillOverFunc) {
	bnc.StopTimer()
	b := image.NewRGBA(image.Rect(0, 0, 2550, 1400))

	for n := 0; n < bnc.N; n++ {
		clearImg(b)
		bnc.StartTimer()
		fo(b, b.Bounds(), 12850, 14906, 15677, 57825)
		bnc.StopTimer()
	}

}

// go test -bench=DrawFillOver -run=NONE -v

func BenchmarkDrawFillOverNormal(bnc *testing.B) { // 18734046 ns/op
	benchFillOver(bnc, drawFillOver_Normal)
}

func BenchmarkDrawFillOverNoPtr(bnc *testing.B) { // 19357654 ns/op
	benchFillOver(bnc, drawFillOver_NoPtr)
}

func BenchmarkDrawFillOverSIMD(bnc *testing.B) { // 4644812 ns/op
	benchFillOver(bnc, drawFillOver_SIMD)
}
