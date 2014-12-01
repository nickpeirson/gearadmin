package gearadmin_test

import (
	"bufio"
	"fmt"
	. "github.com/nickpeirson/gearadmin"
	"io"
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
			if err == io.EOF {
				break
			}
			fmt.Println("Read error: ", err)
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
	want := strings.Split(string(statusResp), "\n")
	firstLine, _ := NewStatusLine(want[0])
	if got[0] != firstLine {
		t.Errorf("First line: got %#v want %#v", got[0], want[0])
	}
	lastLine, _ := NewStatusLine(want[len(want)-3])
	if got[len(got)-1] != lastLine { //Strip trailing ".\n" from want
		t.Errorf("Last line: got %#v want %#v", got[len(got)-1], want[len(want)-3])
	}
}

func TestExcludingLines(t *testing.T) {
	c := setupClient()
	got, err := c.StatusFiltered(func(line StatusLine) bool { return false })
	if err != nil {
		t.Error(err)
	}
	if len(got) != 0 {
		t.Error("Expected no lines, got ", got)
	}
}

func testSorting(by By, ascending bool, expected StatusLine, t *testing.T) {
	c := setupClient()
	got, err := c.Status()
	if err != nil {
		t.Error(err)
	}
	got.Sort(by, ascending)
	if got[0] != expected {
		t.Errorf("First line: got %#v want %#v", got[0], expected)
	}
}

func TestOrderByNameAsc(t *testing.T) {
	testSorting(ByName, true, StatusLine{"A-lastJob", "5", "4", "3"}, t)
}

func TestOrderByNameDesc(t *testing.T) {
	testSorting(ByName, false, StatusLine{"F-GearmanJob", "6", "5", "4"}, t)
}

func TestOrderByQueuedAsc(t *testing.T) {
	testSorting(ByQueued, true, StatusLine{"E-GearmanJob-stale", "1", "6", "5"}, t)
}

func TestOrderByQueuedDesc(t *testing.T) {
	testSorting(ByQueued, false, StatusLine{"F-GearmanJob", "6", "5", "4"}, t)
}

func TestOrderByWorkersAsc(t *testing.T) {
	testSorting(ByWorkers, true, StatusLine{"C-anotherGearmanJob-stale", "3", "2", "1"}, t)
}

func TestOrderByWorkersDesc(t *testing.T) {
	testSorting(ByWorkers, false, StatusLine{"D-anotherGearmanJob", "2", "1", "6"}, t)
}

func TestOrderByRunningAsc(t *testing.T) {
	testSorting(ByRunning, true, StatusLine{"D-anotherGearmanJob", "2", "1", "6"}, t)
}

func TestOrderByRunningDesc(t *testing.T) {
	testSorting(ByRunning, false, StatusLine{"E-GearmanJob-stale", "1", "6", "5"}, t)
}
