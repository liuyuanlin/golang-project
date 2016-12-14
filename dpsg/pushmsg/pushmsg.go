package pushmsg

import (
	"net/http"
	"fmt"
)

func PushMsg(strUser string, strTitle string, strContent string) {
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:1001/?user=%s&title=%s&content=%s", strUser, strTitle, strContent))
	if err != nil {
		return
	}
	defer resp.Body.Close()
}