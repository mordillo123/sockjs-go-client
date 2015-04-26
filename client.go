package sockjs

import (
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/dchest/uniuri"
	"github.com/gorilla/websocket"
)

type Client struct {
	sync.Mutex

	Connection Connection

	WebSockets   bool
	Address      string
	ReadBufSize  int
	WriteBufSize int
}

func NewClient(address string, readBufSize, writeBufSize int) (*Client, error) {
	client := &Client{}
	client.Lock()

	client.Address = address
	client.ReadBufSize = readBufSize
	client.WriteBufSize = writeBufSize

	// Parse the address
	u, err := url.Parse(client.Address)
	if err != nil {
		return nil, err
	}

	// Get info whether WebSockets are enabled
	info, err := client.Info()
	if err != nil {
		return nil, err
	}
	client.WebSockets = info.WebSocket

	// Create a WS session (not a SJS one)
	if client.WebSockets {
		var conn net.Conn
		var host, port string

		// Determine if we're using a custom port
		pi := strings.Index(u.Host, ":")
		if pi == -1 {
			if u.Scheme == "https" {
				port = "443"
			} else {
				port = "80"
			}
		} else {
			host = u.Host[:pi]
			port = u.Host[pi+1:]
		}

		// Dial the server
		if u.Scheme == "https" {
			conn, err = tls.Dial("tcp", host+":"+port, &tls.Config{})
			if err != nil {
				return nil, err
			}
		} else {
			conn, err = tls.Dial("tcp", host+":"+port)
			if err != nil {
				return nil, err
			}
		}

		// Append /websocket
		u.Path += "/websocket"

		// Create a new WS client
		ws, _, err := websocket.NewClient(conn, u, http.Header{}, client.ReadBufSize, client.WriteBufSize)
		if err != nil {
			return nil, err
		}

		// Put it into the client
		client.Connection = ws
	} else {
		// XHR
		client.Connection = NewXHR(address)
	}

	return client
}

func (c *Client) Info() (*Info, error) {
	resp, err := http.Get(c.Address + "/info")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	var info *Info
	if err := dec.Decode(&info); err != nil {
		return nil, err
	}

	return info
}

func (c *Client) WriteMessage(p interface{}) error {

}

func (c *Client) ReadMessage(p interface{}) error {

}
