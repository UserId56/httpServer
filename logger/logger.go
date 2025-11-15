package logger

import "fmt"

const (
	Info    = "Info"
	Warning = "Warning"
	Error   = "Error"
)

func Log(err error, message, typeNotification string) {
	if err != nil {
		fmt.Printf("%s. %s: %v\n", typeNotification, message, err)
		return
	}
	fmt.Printf("%s: %s\n", typeNotification, message)
}
