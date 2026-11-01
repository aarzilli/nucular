package rect

import (
	"image"
	"testing"
)

func TestContainsMatchesImageRectangleBounds(t *testing.T) {
	r := Rect{X: 10, Y: 20, W: 5, H: 4}
	points := []image.Point{
		{X: 10, Y: 20},
		{X: 14, Y: 23},
		{X: 15, Y: 22},
		{X: 12, Y: 24},
	}

	for _, p := range points {
		want := p.In(r.Rectangle())
		if got := r.Contains(p); got != want {
			t.Errorf("Contains(%v) = %t; image.Rectangle bounds contain point = %t", p, got, want)
		}
	}
}
