package utils

import (
	"reflect"
	"strconv"

	"github.com/gorilla/websocket"
)

func SetValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.Bool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(v)
	case reflect.Int:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(v)
	case reflect.String:
		field.SetString(value)
	default:
		// 能走到这说明 ScrcpyOptions 结构体有问题
		panic("unsupported type: " + field.Kind().String())
	}
	return nil
}

type WebsoctekWriteCloser struct {
	Conn *websocket.Conn
}

func (w *WebsoctekWriteCloser) Close() error {
	err := w.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return err
	}
	return w.Conn.Close()
}

func (w *WebsoctekWriteCloser) Write(p []byte) (int, error) {
	err := w.Conn.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
