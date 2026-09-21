//go:build ((darwin && !nucular_shiny) || nucular_gio) && !linux && !windows

package nucular

import (
	"image"
)

func (mw *masterWindow) setIcon(icon image.Image) {
	// not implemented
}
