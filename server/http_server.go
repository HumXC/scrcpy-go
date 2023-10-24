package server

import (
	"fmt"
	"io"
	"net/http"

	"github.com/HumXC/scrcpy-go/server/utils"
	"github.com/gorilla/websocket"
)

func StartHttpServer(scrcpy *ScrcpyServer, addr string) error {
	http.HandleFunc("/ws/video", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("CONN")
		var upgrader = websocket.Upgrader{
			ReadBufferSize: 1024 * 1024 * 4,
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("Error upgrading to websocket:", err)
			return
		}
		defer conn.Close()
		video, err := scrcpy.Video()
		if err != nil {
			fmt.Println("Error getting video stream:", err)
			return
		}
		wc := &utils.WebsoctekWriteCloser{Conn: conn}
		io.Copy(wc, video)
	})
	http.HandleFunc("/video", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("CONN")
		video, err := scrcpy.Video()
		if err != nil {
			fmt.Println("Error getting video stream:", err)
			return
		}
		fmt.Println("GetVideo")
		w.Header().Set("Content-Type", "video/H264")
		_, err = io.Copy(w, video)
		if err != nil {
			fmt.Println(err)
		}
	})
	return http.ListenAndServe(addr, nil)
}
