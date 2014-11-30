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
	want, err := ioutil.ReadFile("testAssets/status.txt")
	if err != nil {
		fmt.Println(err)
	}
	if got[0] != string(want[:strings.IndexByte(string(want), '\n')]) {
		t.Errorf("got %#v want %#v", got[0], string(want[:strings.IndexByte(string(want), '\n')]))
	}
}
