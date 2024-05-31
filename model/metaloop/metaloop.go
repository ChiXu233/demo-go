package metaloop

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

type Response struct {
	Code int `json:"code"`
	Data struct {
		Token string `json:"token"`
	} `json:"data"`
}

func InitMetaloop(url, token string) error {
	if url == "" {
		url = "http://data.deepglint.com"
	}
	if token == "" {
		return errors.New("用户api token为空")
	}
	MClient.Url = url
	MClient.Cli = resty.New()

	go func() {
		// 每24小时刷新一次用户cookie
		ticker := time.NewTicker(time.Second * (60*60*24 - 1))
		var tokenResponse Response
		resp, err := MClient.Cli.R().SetQueryParams(map[string]string{"token": token}).Get(url + "/api/v1/auth_by_token")
		if resp.StatusCode() != 200 || err != nil {
			fmt.Println(err)
		}
		json.Unmarshal(resp.Body(), &tokenResponse)

		MClient.Cli.SetCookie(&http.Cookie{Name: "jwt", Value: tokenResponse.Data.Token})
		for {
			select {
			case <-ticker.C:

				resp, err := MClient.Cli.R().SetQueryParams(map[string]string{"token": token}).Get(url + "/api/v1/auth_by_token")
				if resp.StatusCode() != 200 || err != nil {
					fmt.Println(err)
				}
				json.Unmarshal(resp.Body(), &tokenResponse)

				MClient.Cli.SetCookie(&http.Cookie{Name: "jwt", Value: tokenResponse.Data.Token})
			}
		}
	}()

	return nil

}
