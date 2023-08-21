package main

//Moolticute websocket connection

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	MOOLTICUTE_DAEMON_URL = "ws://localhost:30035"
)

type MsgData struct {
	Service         string `json:"service,omitempty"`
	FallbackService string `json:"fallback_service,omitempty"`
	NodeData        string `json:"node_data,omitempty"`
	Failed          bool   `json:"failed,omitempty"`
	ErrorMessage    string `json:"error_message,omitempty"`
	Login           string `json:"login,omitempty"`
	Password        string `json:"password,omitempty"`
	Description     string `json:"description,omitempty"`
	RequestId       string `json:"request_id,omitempty"`
	ProgressTotal   int    `json:"progress_total,omitempty"`
	ProgressCurrent int    `json:"progress_current,omitempty"`
	McCliVersion    string `json:"mc_cli_version,omitempty"`
}

type MoolticuteMsg struct {
	Msg      string  `json:"msg"`
	Data     MsgData `json:"data"`
	ClientId string  `json:"client_id,omitempty"`
}

type MoolticuteMsgRaw struct {
	Msg      string           `json:"msg"`
	Data     *json.RawMessage `json:"data"`
	ClientId string           `json:"client_id,omitempty"`
}

type ProgressCb func(total, current int)

func McSendQuery(m *MoolticuteMsg) (*MsgData, error) {
	return McSendQueryProgress(m, func(total, current int) {})
}

func McSendQueryProgress(m *MoolticuteMsg, progress ProgressCb) (*MsgData, error) {
	u, err := url.Parse(*mcUrl)
	if err != nil {
		log.Println("Unable to parse moolticute URL", *mcUrl)
		return nil, err
	}
	log.Printf("Moolticute: connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Print("Moolticute: dial:", err)
		return nil, err
	}
	defer c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	defer c.Close()

	if m.ClientId != "" {
		client_uuid := uuid.New()
		m.ClientId = client_uuid.String()
	}

	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	if err = c.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Println("Moolticute: write:", err)
		return nil, err
	}

	//Wait for answer
	for {
		_, data, err = c.ReadMessage()
		if err != nil {
			return nil, fmt.Errorf("moolticute: read: %v", err)
		}

		log.Println(string(data))

		var recv MoolticuteMsgRaw
		err = json.Unmarshal(data, &recv)
		if err != nil {
			return nil, fmt.Errorf("moolticute: unmarshal error: %v", err)
		}

		if recv.Msg == "progress" && recv.Data != nil {
			var recvData MsgData
			err = json.Unmarshal([]byte(*recv.Data), &recvData)
			if err != nil {
				log.Println("moolticute: unmarshal error:", err)
				return nil, fmt.Errorf("moolticute: %v", err)
			}
			progress(recvData.ProgressTotal, recvData.ProgressCurrent)
		}

		if recv.Msg != m.Msg {
			continue
		}

		var recvData MsgData
		err = json.Unmarshal([]byte(*recv.Data), &recvData)
		if err != nil {
			log.Println("moolticute: unmarshal error:", err)
			return nil, fmt.Errorf("moolticute: %v", err)
		}

		if recvData.Failed {
			log.Println("error from moolticute:", recvData.ErrorMessage)
			return nil, fmt.Errorf("moolticute: %v", recvData.ErrorMessage)
		}

		if m.ClientId == recv.ClientId {
			return &recvData, nil
		}

		log.Println("should not get here, something is wrong with moolticute answer")
		return nil, fmt.Errorf("something went wrong in Moolticute answer")
	}
}

// We store an array of bytes to the device
type McBinKeys [][]byte

func McLoadKeys() (keys *McBinKeys, err error) {
	keys = new(McBinKeys)

	u, err := url.Parse(*mcUrl)
	if err != nil {
		log.Println("Unable to parse moolticute URL", *mcUrl)
		return
	}
	log.Printf("Moolticute: connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Print("Moolticute: dial:", err)
		return
	}
	defer c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	defer c.Close()

	client_uuid := uuid.New()

	m := MoolticuteMsg{
		Msg:      "get_data_node",
		ClientId: client_uuid.String(),
		Data: MsgData{
			Service: "Moolticute SSH Keys",
		},
	}

	data, err := json.Marshal(m)
	if err != nil {
		return
	}
	if err = c.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Println("Moolticute: write:", err)
		return
	}

	b64data := ""

	for {
		_, data, err = c.ReadMessage()
		if err != nil {
			log.Println("Moolticute: read:", err)
			return
		}

		log.Println(string(data))

		var recv MoolticuteMsgRaw
		err = json.Unmarshal(data, &recv)
		if err != nil {
			log.Println("Moolticute: unmarshal error:", err)
			return
		}

		if recv.Msg != "get_data_node" {
			continue
		}

		var recvData MsgData
		err = json.Unmarshal([]byte(*recv.Data), &recvData)
		if err != nil {
			log.Println("Moolticute: unmarshal error:", err)
			return
		}

		// keys are not present, this is not an error
		if recvData.Failed {
			log.Println("Error getting node data from moolticute:", recvData.ErrorMessage)
			return keys, nil
		}

		if m.ClientId == recv.ClientId {
			b64data = recvData.NodeData
			break
		}

		log.Println("Should not get here, something is wrong with moolticute answer")
		return keys, fmt.Errorf("something went wrong in Moolticute answer")
	}

	//try to decode binary data from device
	//First Base64 decode
	bdec, err := base64.StdEncoding.DecodeString(b64data)
	if err != nil {
		log.Println("Failed to base64 decode data:", err)
		return
	}

	//To debug
	//	kf, _ := os.Create("keys.bin")
	//	kf.Write(bdec)
	//	kf.Close()

	buffer := bytes.NewBuffer(bdec)
	binDec := gob.NewDecoder(buffer)
	err = binDec.Decode(keys)
	if err != nil {
		log.Println("Failed to decode encoding/gob:", err)
		return
	}

	return
}

func McSetKeys(keys *McBinKeys) (err error) {
	u := url.URL{Scheme: "ws", Host: MOOLTICUTE_DAEMON_URL, Path: "/"}
	log.Printf("Moolticute: connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Print("Moolticute: dial:", err)
		return
	}
	defer c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	defer c.Close()

	var buffer bytes.Buffer
	binEnc := gob.NewEncoder(&buffer)
	if err = binEnc.Encode(keys); err != nil {
		return fmt.Errorf("failed to encode with encoding/gob: %v", err)
	}

	client_uuid := uuid.New()

	m := MoolticuteMsg{
		Msg:      "set_data_node",
		ClientId: client_uuid.String(),
		Data: MsgData{
			Service:  "Moolticute SSH Keys",
			NodeData: base64.StdEncoding.EncodeToString(buffer.Bytes()),
		},
	}

	data, err := json.Marshal(m)
	if err != nil {
		return
	}
	if err = c.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Println("Moolticute: write:", err)
		return
	}

	for {
		_, data, err = c.ReadMessage()
		if err != nil {
			return fmt.Errorf("moolticute: read: %v", err)
		}

		log.Println(string(data))

		var recv MoolticuteMsgRaw
		err = json.Unmarshal(data, &recv)
		if err != nil {
			return fmt.Errorf("moolticute: unmarshal error: %v", err)
		}

		if recv.Msg != "set_data_node" {
			continue
		}

		var recvData MsgData
		err = json.Unmarshal([]byte(*recv.Data), &recvData)
		if err != nil {
			return fmt.Errorf("moolticute: unmarshal error: %v", err)
		}

		// keys are not present, this is not an error
		if recvData.Failed {
			return fmt.Errorf("error getting node data from moolticute: %v", err)
		}

		if m.ClientId == recv.ClientId {
			break
		}

		log.Println("Should not get here, something is wrong with moolticute answer")
		return fmt.Errorf("something went wrong in Moolticute answer")
	}

	return
}
