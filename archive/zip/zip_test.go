package zip

import "testing"

func TestCompressDir(t *testing.T) {
	CompressDir(`D:\work\0317_export\cad`, "./test.zip", false)
	CompressDir(`D:\work\0317_export\cad`, "./test1.zip", true)
}
