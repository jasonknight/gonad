package main
import (
	"testing"
	"os"
	"reflect"
	"net"
	"time"
	"bufio"
	"io/ioutil"
)
func getField(env Env,f string) string {
	ref := reflect.ValueOf(env)
	v := reflect.Indirect(ref).FieldByName(f)
	return string(v.String())
}
func TestPathExists(t *testing.T) {
	if pathExists("./main_test.go") != true {
		t.Errorf("failed to test if path exists")
	}
}
func TestCreateEnv(t *testing.T) {
	table := []struct {
		field string
		name string
		value string
	}{
		{"Host","GONAD_HOST","testing_value"},
		{"Port","GONAD_PORT","testing_value"},
		{"Destination","GONAD_DESTINATION","testing_value"},
		{"DestinationPath","GONAD_DESTINATION_PATH","testing_value"},
	}
	var old string
	for _,test_e := range table {
		old = os.Getenv(test_e.name)
		os.Setenv(test_e.name,test_e.value)
		env,err := createEnv()
		if err != nil {
			t.Errorf("%s",err)
			return
		}
		tval := getField(env,test_e.field)
		if tval != test_e.value {
			t.Errorf("%s was not %s but is %s for env %s\n",test_e.field, test_e.value,tval, test_e.name)
		}
		os.Setenv(test_e.name,old)
	}
}
func TestHandleAccept(t *testing.T) {
	defer func () {
		if pathExists("./test.golden") {
			os.Remove("./test.golden")
		}
	}()
	// Credits to Freman from SO for this idea
	s,c := net.Pipe()
	defer c.Close()
	defer s.Close()
	old_dest := os.Getenv("GONAD_DESTINATION")
	old_path := os.Getenv("GONAD_DESTINATION_PATH")
	os.Setenv("GONAD_DESTINATION","file")
	os.Setenv("GONAD_DESTINATION_PATH","./test.golden")
	env,_ := createEnv()
	defer destroyEnv(env)
	setOutput(os.Stdout)
	os.Setenv("GONAD_DESTINATION",old_dest)
	os.Setenv("GONAD_DESTINATION_PATH",old_path)

	test_bytes := []byte("hello world\n")
	go func(wr *bufio.Writer) {
		n,err := wr.Write(test_bytes)
		if err != nil {
			t.Errorf("in go write %s",err)
		}
		if n == 0 {
			t.Errorf("n=%d",n)
		}
		wr.Flush()
	}(bufio.NewWriter(c))
	go handleAccept(env,s,5 * time.Second)
	time.Sleep(1 * time.Second)
	contents,_ := ioutil.ReadFile("./test.golden")
	rs := contents[:len(test_bytes)]
	if string(test_bytes[:]) != string(rs) {
		t.Errorf("'%s' != '%s' from: %s\n",test_bytes[:],rs, contents)
	}
}
