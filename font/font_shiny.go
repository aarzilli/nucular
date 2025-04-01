package font

import (
	"crypto/sha256"
	"sync"

	"golang.org/x/image/font"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

type Face struct {
	Face font.Face
}

var fontsMu sync.Mutex
var fontsMap = map[[sha256.Size]byte]*truetype.Font{}

// NewFace returns a new face by parsing the ttf font.
func NewFace(ttf []byte, size int) (Face, error) {
	key := sha256.Sum256(ttf)
	fontsMu.Lock()
	defer fontsMu.Unlock()

	fnt, _ := fontsMap[key]
	if fnt == nil {
		var err error
		fnt, err = freetype.ParseFont(ttf)
		if err != nil {
			return Face{}, err
		}
	}

	return Face{truetype.NewFace(fnt, &truetype.Options{Size: float64(size), Hinting: font.HintingFull, DPI: 72})}, nil
}

func (face Face) Metrics() font.Metrics {
	return face.Face.Metrics()
}
