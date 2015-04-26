package sockjs

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/dchest/uniuri"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type XHR struct {
	Address          string
	TransportAddress string
	ServerID         string
	SessionID        string
	Inbound          chan []body
}

func paddedRandomIntn(max int) string {
	var (
		ml = len(strconv.Itoa(max))
		ri = rand.Intn(max)
		is = strconv.Itoa(ri)
	)

	if len(is) < ml {
		is = strings.Repeat("0", ml-is) + is
	}

	return is
}

func NewXHR(address string) (*XHR, error) {
	xhr := &XHR{
		Address:   address,
		ServerID:  paddedRandomIntn(999),
		SessionID: uniuri.New(),
	}
	xhr.TransportAddress = address + "/" + xhr.ServerID + "/" + xhr.SessionID
	if err := xhr.Init(); err != nil {
		return nil, err
	}
	if err := xhr.StartReading(); err != nil {
		return nil, err
	}

	return xhr
}

func (x *XHR) Init() error {
	req, err := http.NewRequest("POST", x.TransportAddress+"/xhr", nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if body[0] != 'o' {
		return errors.New("Invalid initial message")
	}

	return nil
}

func (x *XHR) StartReading() {
	go func() {
		client := &http.Client{
			Timeout: time.Minute,
		}

		for {
			req, err := http.NewRequest("POST", x.TransportAddress+"/xhr", nil)
			if err != nil {
				log.Print(err)
				continue
			}
			resp, err := client.Do(req)
			if err != nil {
				log.Print(err)
				continue
			}
			defer resp.Body.Close()

			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Print(err)
				continue
			}

			switch data[0] {
			case 'h':
				// Heartbeat
				continue
			case 'a':
				// Normal message
				x.Inbound <- data[1:]
			case 'c':
				// Session closed
				var v []interface{}
				if err := json.Unmarshal(data[1:], &v); err != nil {
					log.Printf("Closing session: %s", err)
					break
				}
				log.Printf("%v: %v", v[0], v[1])
				break
			}
		}
	}()
}

func (x *XHR) ReadJSON(v interface{}) error {
	message := <-x.Inbound
	return json.Unmarshal(message, v)
}

func (x *XHR) WriteJSON(v interface{}) error {
	message, err := json.Marshal(v)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", x.TransportAddress+"/xhr_send", bytes.NewReader(message))
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		return errors.New("Invalid HTTP code - " + resp.Status)
	}

	return nil
}
