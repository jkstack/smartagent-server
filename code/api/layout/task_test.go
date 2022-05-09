package layout

import "testing"

func TestParseIf(t *testing.T) {
	var info taskInfo
	err := info.parseIf("1=1")
	if err != nil {
		t.Fatal(err)
	}
	err = info.parseIf("1 = 1")
	if err != nil {
		t.Fatal(err)
	}
	err = info.parseIf("$a=$b")
	if err != nil {
		t.Fatal(err)
	}
	err = info.parseIf("$a>=$b")
	if err != nil {
		t.Fatal(err)
	}
}
