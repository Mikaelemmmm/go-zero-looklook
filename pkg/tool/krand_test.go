package tool

import "testing"

func TestMd5ByString(t *testing.T) {
	s := Md5ByString("AAA")
	t.Log(s)
}
