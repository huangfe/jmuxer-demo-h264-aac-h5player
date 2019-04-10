package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func wsPushStream(client *websocket.Conn, controlChan chan int32) {
	content, _ := ioutil.ReadFile("config.txt")
	strContent := strings.Trim(string(content), "\n\t ")
	contents := strings.Split(strContent, "\n")
	for i := range contents {
		contents1 := strings.Split(contents[i], " ")
		duration, _ := strconv.ParseInt(contents1[0], 10, 32)
		duration1 := []byte{byte(duration >> 24), byte((duration & 0xff0000) >> 16), byte((duration & 0xff00) >> 8), byte(duration & 0xff)}
		audio, _ := ioutil.ReadFile(contents1[1])
		video, _ := ioutil.ReadFile(contents1[2])

		client.WriteMessage(websocket.BinaryMessage, append(append(duration1, 0x1), audio...))
		client.WriteMessage(websocket.BinaryMessage, append(append(duration1, 0x0), video...))

		time.Sleep(time.Duration(5) * time.Second)
	}
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI != "/h5player" {
			w.Header().Add("Access-Control-Allow-Origin", "*")
			http.ServeFile(w, r, "./"+strings.TrimLeft(r.RequestURI, "/\\"))
			return
		}

		var upgrader = websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(r.RequestURI, "WS ACCEPT ERROR")
			return
		}
		defer conn.Close()

		for {
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				break
			}

			data := map[string]string{}
			if messageType == 1 {
				if err := json.Unmarshal(p, &data); err == nil {
					kind, ok := data["t"]
					if !ok {
						return
					}
					message, ok := data["m"]
					if !ok {
						return
					}

					switch kind {
					case "open":
						fmt.Printf("%s => %s\n", kind, message)
					case "close":
					case "userHasLogin":
						resp1, _ := json.Marshal(&map[string]string{
							"status":  "true",
							"message": "<sid>",
						})
						data = map[string]string{
							"t": "userHasLogin",
							"m": string(resp1),
						}
						resp, _ := json.Marshal(&data)
						conn.WriteMessage(websocket.TextMessage, resp)
					case "streamStart":
						resp1, _ := json.Marshal(&map[string]string{
							"status":  "true",
							"message": "",
						})
						data = map[string]string{
							"t": "streamStart",
							"m": string(resp1),
						}
						resp, _ := json.Marshal(&data)
						conn.WriteMessage(websocket.TextMessage, resp)

						controlChan := make(chan int32, 1)
						go wsPushStream(conn, controlChan)
					case "streamStop":
					}
				}
			}
		}
	})

	serverBindAddr := "0.0.0.0:20005"
	if err := http.ListenAndServe(serverBindAddr, nil); err != nil {
		fmt.Println(fmt.Sprintf("%s %s", serverBindAddr, err.Error()), "HTTP CAN NOT BIND")
	}
}

