package environment

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestHandleEnv(t *testing.T) {
	content, _ := ioutil.ReadFile("../example/deployment.yaml")
	out, err := HandleEnv(content, "../example/merge.yaml")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(out))
}
