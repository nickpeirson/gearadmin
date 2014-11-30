package gearadmin

import (
	"bufio"
	"net"
	"time"
)

type Client struct {
	addr string
	conn net.Conn
}

func New(host, port string) Client {
	return Client{addr: host + ":" + port}
}

func (c *Client) Connect() (err error) {
	if c.conn != nil {
		return
	}
	conn, err := net.DialTimeout("tcp", c.addr, 1*time.Second)
	if err != nil {
		return err
	}
	c.conn = conn
	return
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

func (c *Client) Status() (status []string, err error) {
	if err = c.sendCmd("status"); err != nil {
		return nil, err
	}
	return c.getResponse(), err
}
