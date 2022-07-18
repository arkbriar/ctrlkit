package gen

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func Test_GenerateStubCodes(t *testing.T) {
	const testFile = "../../example/cronjob.cm"
	f, err := os.Open(testFile)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	r := bufio.NewReader(f)
	doc, err := ParseDoc(r)
	if err != nil {
		t.Fatal(err)
	}
	doc.FileName = filepath.Base(testFile)

	s, err := GenerateStubCodes(doc, "tmp")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(s)
}

func Test_GetStrExpr(t *testing.T) {
	fmt.Println(getStrExpr("risingwave-${target.Name}", "s.target"))
}
