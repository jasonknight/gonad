package main
import (
	"testing"
	"os"
	"reflect"
	"net"
	"time"
	"bufio"
	"io/ioutil"
	"fmt"
)
func getField(env Env,f string) string {
	ref := reflect.ValueOf(env)
	v := reflect.Indirect(ref).FieldByName(f)
	return string(v.String())
}
func createServerAndClient() (net.Listener, net.Conn, net.Conn) {
	listener, err := net.Listen("tcp", ":1337")
	if err != nil {
		panic("could not create server")
	}
	sender,err := net.Dial("tcp",":1337")
	if err != nil {
		panic("could not create client")
	}
	lc,err := listener.Accept()
	if err != nil {
		panic(err)
	}
	return listener,lc,sender
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
		{"Host","GONNAD_HOST","testing_value"},
		{"Port","GONNAD_PORT","testing_value"},
		{"Destination","GONNAD_DESTINATION","testing_value"},
		{"DestinationPath","GONNAD_DESTINATION_PATH","testing_value"},
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
	l,s,c := createServerAndClient()
	defer c.Close()
	defer s.Close()
	defer l.Close()
	old_dest := os.Getenv("GONNAD_DESTINATION")
	old_path := os.Getenv("GONNAD_DESTINATION_PATH")
	os.Setenv("GONNAD_DESTINATION","file")
	os.Setenv("GONNAD_DESTINATION_PATH","./test.golden")
	env,_ := createEnv()
	setOutput(os.Stdout)
	os.Setenv("GONNAD_DESTINATION",old_dest)
	os.Setenv("GONNAD_DESTINATION_PATH",old_path)

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
		time.Sleep(1 * time.Second)
	}(bufio.NewWriter(c))
	go handleAccept(env,s,5 * time.Second)
	time.Sleep(1 * time.Second)
	env.Fd.Close()
	contents,err := ioutil.ReadFile("./test.golden")
	if err != nil {
		t.Errorf("%s",err)
	}
	fmt.Printf("contents[%s]",contents)
	rs := contents[:len(test_bytes)]
	if string(test_bytes[:]) != string(rs) {
		t.Errorf("'%s' != '%s' from: %s\n",test_bytes[:],rs, contents)
	}
}
