package gearadmin_test

import (
	"bufio"
	"fmt"
	. "github.com/nickpeirson/gearadmin"
	"io/ioutil"
	"net"
	"strings"
	"testing"
)

const (
	testHost = "localhost"
	testPort = "49151"
	testAddr = testHost + ":" + testPort
)

func handleConn(conn net.Conn) {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	for {
		cmd, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error", err)
			continue
		}
		resp, err := ioutil.ReadFile("testAssets/" + cmd[:len(cmd)-1] + ".txt")
		if err != nil {
			fmt.Println(err)
			continue
		}
		_, _ = writer.Write(resp)
		writer.Flush()
	}
}

func mockGearmand() {
	listener, _ := net.Listen("tcp", testAddr)
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Print(err)
			continue
		}
		go handleConn(conn)
	}
}
func init() {
	go mockGearmand()
}

func setupClient() Client {
	return New(testHost, testPort)
}

func TestCanConnect(t *testing.T) {
	c := setupClient()
	if err := c.Connect(); err != nil {
		t.Error(err)
	}
}

func TestCanGetStatus(t *testing.T) {
	c := setupClient()
	got, err := c.Status()
	if err != nil {
		t.Error(err)
	}
	statusResp, err := ioutil.ReadFile("testAssets/status.txt")
	if err != nil {
		fmt.Println(err)
	}
	want := strings.Split(string(statusResp),"\n")
	if got[0] != want[0] {
		t.Errorf("First line: got %#v want %#v", got[0], want[0])
	}
	if got[len(got)-1] != want[len(want)-3] { //Strip trailing ".\n" from want
		t.Errorf("Last line: got %#v want %#v", got[len(got)-1], want[len(want)-3])
	}
}

func TestExcludingLines(t *testing.T) {
	c := setupClient()
	got, err := c.StatusFiltered(func (line string) bool{ return false })
	if err != nil {
		t.Error(err)
	}
	if len(got) != 0 {
		t.Error("Expected no lines, got ", got)
	}
}
