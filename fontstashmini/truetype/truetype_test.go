package truetype

import (
	"os"
	"runtime"
	"testing"
)

func Test_TTC(t *testing.T) {
	filename := ""
	switch runtime.GOOS {
	case "windows":
		filename = "C:/Windows/Fonts/simhei.ttf"
	case "linux":
		filename = "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf"
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	count := GetFontCount(data)
	t.Log("font count:", count)

	for i := 0; i < count; i++ {
		name := GetFontName(data, i)
		t.Log("font name:", name)
	}
}
