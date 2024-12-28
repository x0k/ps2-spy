package main

import (
	"encoding/json"
	"fmt"

	"github.com/x0k/ps2-spy/internal/lib/census2/streaming"
)

func main() {
	buffers := map[streaming.MessageType]streaming.Message{
		streaming.ServiceStateChangedType: streaming.ServiceStateChanged{},
		streaming.HeartbeatType:           streaming.Heartbeat{},
		streaming.ServiceMessageType:      streaming.ServiceMessage[json.RawMessage]{},
	}

	buff := buffers[streaming.ServiceStateChangedType]
	content := `{"detail":"EventServerEndpoint_Miller_10","online":"true","service":"event","type":"serviceStateChanged"}`

	err := json.Unmarshal([]byte(content), &buff)
	if err != nil {
		panic(err)
	}

	fmt.Println(buff)
}
