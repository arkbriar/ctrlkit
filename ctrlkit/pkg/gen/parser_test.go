package gen

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func Test_ParseDoc(t *testing.T) {
	f, err := os.Open("../../example/cronjob.cm")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	r := bufio.NewReader(f)
	doc, err := ParseDoc(r)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.MarshalIndent(doc, "", "  ")
	fmt.Println(string(b))
}
