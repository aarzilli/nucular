package nucular

import "image"

func drawFillOver_SIMD_internal(base *uint8, i0, i1 int, stride, n int, adivm, sr, sg, sb, sa uint32)
func haveAVX2() bool

var useSIMD = haveAVX2()

func drawFillOver(dst *image.RGBA, r image.Rectangle, sr, sg, sb, sa uint32) {
	const m = 1<<16 - 1
	a := (m - sa) * 0x101

	if useSIMD {
		adivm := a / m
		i0 := dst.PixOffset(r.Min.X, r.Min.Y)
		i1 := i0 + r.Dx()*4
		drawFillOver_SIMD_internal(&dst.Pix[0], i0, i1, dst.Stride, r.Max.Y-r.Min.Y, adivm, sr, sg, sb, sa)
		return
	}

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
