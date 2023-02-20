package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/vitalis-virtus/simple-oauth/utils"
	"golang.org/x/oauth2"
)

var (
	State = "github_state"
)

func Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	githubAccessToken := getGithubAccessToken(code)

	githubData := getGithubData(githubAccessToken)

	fmt.Fprint(w, githubData)
}

func getGithubData(accessToken string) string {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)

	if err != nil {
		log.Panic("API Request creation failed")
	}

	authorizationHeaderValue := fmt.Sprintf("token %s", accessToken)
	req.Header.Set("Authorization", authorizationHeaderValue)

	resp, resperr := http.DefaultClient.Do(req)
	if resperr != nil {
		log.Panic("Request failed")
	}

	respbody, _ := ioutil.ReadAll(resp.Body)

	return string(respbody)
}

func getGithubAccessToken(code string) string {
	clientID := utils.GoDotEnvVariable("GITHUB_CLIENT_ID")
	clientSecret := utils.GoDotEnvVariable("GITHUB_CLIENT_SECRET")

	requestBodyMap := map[string]string{"client_id": clientID, "client_secret": clientSecret, "code": code}

	requestJSON, _ := json.Marshal(requestBodyMap)

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", bytes.NewBuffer(requestJSON))

	if err != nil {
		log.Panic("Request creation failed")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, resperr := http.DefaultClient.Do(req)
	if resperr != nil {
		log.Panic("Request failed")
	}

	respbody, _ := ioutil.ReadAll(resp.Body)

	// Represents the response received from Github
	type githubAccessTokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}

	var ghresp githubAccessTokenResponse
	json.Unmarshal(respbody, &ghresp)

	return ghresp.AccessToken

}

func GetGithubConfig() *oauth2.Config {
	return &oauth2.Config{
		RedirectURL:  utils.GoDotEnvVariable("GITHUB_REDIRECT_URL"),
		ClientID:     utils.GoDotEnvVariable("GITHUB_CLIENT_ID"),
		ClientSecret: utils.GoDotEnvVariable("GITHUB_CLIENT_SECRET"),
	}
}
