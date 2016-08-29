package nucular

import (
	"testing"
)

func asserteq(t *testing.T, s []rune, tgt []rune) {
	if len(s) != len(tgt) {
		t.Fatalf("mismatched length %d %d", len(s), len(tgt))
	}

	for i := range s {
		if s[i] != tgt[i] {
			t.Fatalf("mismatch at character %d output: %q target: %q", i, string(s), string(tgt))
		}
	}

}

func TestInsertEnd(t *testing.T) {
	s := []rune("something")
	s = strInsertText(s, len(s), []rune("zap"))
	asserteq(t, s, []rune("somethingzap"))
}

func TestInsertMid(t *testing.T) {
	s := []rune("something")
	s = strInsertText(s, 3, []rune("zap"))
	asserteq(t, s, []rune("somzapething"))
}

func TestInsertStart(t *testing.T) {
	s := []rune("something")
	s = strInsertText(s, 0, []rune("zap"))
	asserteq(t, s, []rune("zapsomething"))
}

func TestDeleteEnd(t *testing.T) {
	s := []rune("something")
	s = strDeleteText(s, len(s)-3, 3)
	asserteq(t, s, []rune("someth"))
}

func TestDeleteMid(t *testing.T) {
	s := []rune("something")
	s = strDeleteText(s, 3, 2)
	asserteq(t, s, []rune("somhing"))
}

func TestDeleteStart(t *testing.T) {
	s := []rune("something")
	s = strDeleteText(s, 0, 2)
	asserteq(t, s, []rune("mething"))
}

func TestInsertFromNone(t *testing.T) {
	s := []rune{}
	s = strInsertText(s, 0, []rune("something"))
	asserteq(t, s, []rune("something"))
}

func TestDeleteAll(t *testing.T) {
	s := []rune("something")
	s = strDeleteText(s, 0, len(s))
	asserteq(t, s, []rune{})
}
