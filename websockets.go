package sockjs

import (
	"encoding/json"
	"errors"
	"github.com/dchest/uniuri"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type WebSocket struct {
	Address          string
	TransportAddress string
	ServerID         string
	SessionID        string
	Connection       *websocket.Conn
	Inbound          chan []byte
}

func NewWebSocket(address string) (*WebSocket, error) {
	ws := &WebSocket{
		Address:   address,
		ServerID:  paddedRandomIntn(999),
		SessionID: uniuri.New(),
	}

	ws.TransportAddress = address + "/" + ws.ServerID + "/" + ws.SessionID + "/websocket"

	if err := ws.Init(); err != nil {
		return nil, err
	}

	ws.StartReading()

	return ws, nil
}

func (w *WebSocket) Init() error {
	ws, _, err := websocket.DefaultDialer.Dial(w.TransportAddress, http.Header{})
	if err != nil {
		return err
	}

	// Read the open message
	_, data, err := ws.ReadMessage()
	if err != nil {
		return err
	}

	if data[0] != 'o' {
		return errors.New("Invalid initial message")
	}

	w.Connection = ws

	return nil
}

func (w *WebSocket) StartReading() {
	go func() {
		for {
			id, data, err := w.Connection.ReadMessage()
			if err != nil {
				log.Print(err)
				return
			}

			if len(data) < 1 {
				continue
			}

			switch data[0] {
			case 'h':
				// Heartbeat
				continue
			case 'a':
				// Normal message
				w.Inbound <- data[1:]
			case 'c':
				// Session closed
				var v []interface{}
				if err := json.Unmarshal(data[1:], &v); err != nil {
					log.Printf("Closing session: %s", err)
					break
				}
				break
			}
		}
	}()
}

func (w *WebSocket) ReadJSON(v interface{}) error {
	message := <-w.Inbound
	return json.Unmarshal(message, v)
}

func (w *WebSocket) WriteJSON(v interface{}) error {
	x, e := json.Marshal(&v)
	if e != nil {
		return e
	}
	return w.Connection.WriteJSON(v)
}

func (w *WebSocket) Close() error {
	return w.Connection.Close()
}
