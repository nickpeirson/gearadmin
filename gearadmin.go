package gearadmin

import (
	"bufio"
	"net"
	"time"
	"strconv"
	"strings"
	"errors"
	"sort"
	"github.com/pmylund/sortutil"
)

type Client struct {
	addr string
	conn net.Conn
}

type StatusLine struct {
	Name    string
	Queued  string
	Running string
	Workers string
}


func NewStatusLine(line string) (StatusLine, error) {
	parts := strings.Fields(line)
	if len(parts) != 4 {
		return StatusLine{}, errors.New("Wrong number of fields")
	}
	return StatusLine{parts[0], parts[1], parts[2], parts[3]}, nil
}

type StatusLines []StatusLine
type By func(l1, l2 *StatusLine) bool
type StatusLineSorter struct {
	lines StatusLines
	by By
}

func (a StatusLineSorter) Len() int      { return len(a.lines) }
func (a StatusLineSorter) Swap(i, j int) { a.lines[i], a.lines[j] = a.lines[j], a.lines[i] }
func (a StatusLineSorter) Less(i, j int) bool {return a.by(&a.lines[i], &a.lines[j])}

func byName(l1, l2 *StatusLine)  bool {
	return strings.ToLower(l1.Name) < strings.ToLower(l2.Name)
}

func byQueued(l1, l2 *StatusLine) bool {
	l1q, _ := strconv.Atoi(l1.Queued)
	l2q, _ := strconv.Atoi(l2.Queued)
	return l1q < l2q
}

func byRunning(l1, l2 *StatusLine)  bool {
	l1r, _ := strconv.Atoi(l1.Running)
	l2r, _ := strconv.Atoi(l2.Running)
	return l1r < l2r
}

func byWorkers(l1, l2 *StatusLine) bool {
	l1w, _ := strconv.Atoi(l1.Workers)
	l2w, _ := strconv.Atoi(l2.Workers)
	return l1w < l2w
}


func (sl StatusLines) Sort(field string, ascending bool) {
	sorter := StatusLineSorter{lines: sl}
	switch field {
	case "name":
		sorter.by = byName
	case "queued":
		sorter.by = byQueued
	case "running":
		sorter.by = byRunning
	case "workers":
		sorter.by = byWorkers
	}
	if ascending {
		sort.Sort(sorter)
	} else {
		sortutil.SortReverseInterface(sorter)
	}
}

//Given a line ResponseFilter returns true to include or false to exclude
type StatusLineFilter func(line StatusLine) bool

func nopStatusLineFilter (line StatusLine) bool {
	return true
}

func New(host, port string) Client {
	return Client{addr: host + ":" + port}
}


func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
func (c *Client) ConnectTimeout(timeout time.Duration) (err error) {
	if c.conn != nil {
		return
	}
	conn, err := net.DialTimeout("tcp", c.addr, timeout)
	if err != nil {
		return err
	}
	c.conn = conn
	return
}

func (c *Client) Connect() (err error) {
	return c.ConnectTimeout(1*time.Second)
}

func readResponse(conn net.Conn, resp chan string) {
	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		if line == ".\n" {
			break
		}
		resp <- line[:len(line)-1]
	}
	close(resp)
}

func (c *Client) sendCmd(cmd string) (err error) {
	if err = c.Connect(); err != nil {
		return err
	}
	writer := bufio.NewWriter(c.conn)
	writer.WriteString(cmd + "\n")
	err = writer.Flush()
	return
}

func (c *Client) getResponse() (resp []string) {
	respChan := make(chan string)
	go readResponse(c.conn, respChan)
	for line := range respChan {
		resp = append(resp, line)
	}
	return
}

func (c *Client) Status() (status StatusLines, err error) {
	return c.StatusFiltered(nopStatusLineFilter)
}

func (c *Client) StatusFiltered(f StatusLineFilter) (status StatusLines, err error) {
	if err = c.sendCmd("status"); err != nil {
		return nil, err
	}
	for _, responseLine := range c.getResponse() {
		statusLine, err := NewStatusLine(responseLine)
		if err != nil {
			return nil, err
		}
		if !f(statusLine) {
			continue
		}
		status = append(status, statusLine) 
	}
	return
}
