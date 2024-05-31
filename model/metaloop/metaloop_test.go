package metaloop

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
)

func createClient() {
	var tokenResponse Response
	url := "http://data.deepglint.com"

	MClient.Url = url
	MClient.Cli = resty.New()
	resp, err := MClient.Cli.R().SetQueryParams(map[string]string{"token": "7f5f6faa-9bf9-4cef-8c58-aa18efac6261"}).Get(url + "/api/v1/auth_by_token")
	if resp.StatusCode() != 200 || err != nil {
		panic(err)
	}
	json.Unmarshal(resp.Body(), &tokenResponse)

	MClient.Cli.SetCookie(&http.Cookie{Name: "jwt", Value: tokenResponse.Data.Token})
}

func TestCreateClient(t *testing.T) {

}
