package sigeaserver2

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"time"
)

type HandleConn interface {
	Name() string
	Handler(conn Connection)
}

func (s *SigeaServer) AddTcpHandler(handler HandleConn) {
	s.tcpHandlers = append(s.tcpHandlers, handler)
}

func (s *SigeaServer) handleConn(conn Connection) {
	defer conn.Close()
	name, _ := conn.ReadString()

	handler := s.findHandler(name)

	if handler == nil {
		fmt.Printf("Unknown connection type [%v].\n", name)
		return
	}

	handler.Handler(conn)
}

func (s *SigeaServer) findHandler(name string) HandleConn {
	for _, v := range s.tcpHandlers {
		if v.Name() == name {
			return v
		}
	}
	return nil
}

type Connection struct {
	conn net.Conn
}

func (c *Connection) Close() {
	c.conn.Close()
}

func (c *Connection) ReadLong() (int, error) {
	buf, err := c.ReadNBytes(4)
	if err != nil {
		return 0, err
	} else {
		r := int(buf[0])*256*256*256 + int(buf[1])*256*256 + int(buf[2])*256 + int(buf[3])
		return r, nil
	}
}

func (c *Connection) ReadNBytes(q int) ([]byte, error) {
	buf := make([]byte, q)
	_, err := io.ReadFull(c.conn, buf)
	if err != nil {
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		fmt.Printf("Erro lendo [%v] bytes [%v]\n", q, err)
		return nil, err
	} else {
		return buf, nil
	}
}

func (c *Connection) ReadString() (string, error) {
	wrap := bufio.NewReader(c.conn)
	line, err := wrap.ReadBytes('\n')
	return string(line)[:len(line)-2], err
}

func (c *Connection) ReadStringPrefix() (string, error) {
	n, err := c.ReadLong()
	if err != nil {
		fmt.Printf("Erro ao ler string size [%v]\n", err)
		return "", err
	}
	if n > 1024*1024*50 {
		fmt.Printf("Tamanho de string suspeito %v\n", n)
		return "", fmt.Errorf("Tamanho de string invalido")
	}
	line, err := c.ReadNBytes(n)
	if err != nil {
		fmt.Println("Erro ao ler linha")
		return "", err
	}
	return string(line), err
}

func (c *Connection) WriteString(line string) {
	c.conn.Write([]byte(line))
}

func (c *Connection) Writestringf(line string, f ...interface{}) {
	s := fmt.Sprintf(line, f...)
	c.WriteString(s)
}

func (c *Connection) WriteByte(b []byte) {
	c.conn.Write(b)
}

func (c *Connection) WriteResult(result bool) {
	if result {
		c.conn.Write([]byte{0})
	} else {
		c.conn.Write([]byte{1})
	}
}
