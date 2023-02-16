package linkedin

import (
	"context"
	"encoding/json"
	"log"

	"fmt"
	"io"
	"net/http"

	"github.com/m7shapan/njson"
	"github.com/vitalis-virtus/simple-oauth/models"
	"github.com/vitalis-virtus/simple-oauth/utils"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/linkedin"
)

var (
	State = "linkedin_state"
)

func Callback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	stateCheck := r.FormValue("state")
	if State != stateCheck {
		http.Error(w, fmt.Sprintf("wrong state string: expected: %s, got: %s", State, stateCheck), http.StatusBadRequest)
		return
	}

	token, err := GetLinkedInConfig().Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		http.Error(w, fmt.Sprintf("wrong code: %s", r.FormValue("code")), http.StatusBadRequest)
		return
	}

	client := GetLinkedInConfig().Client(context.Background(), token)

	req, err := http.NewRequest("GET", "https://api.linkedin.com/v2/me", nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req.Header.Set("Bearer", token.AccessToken)

	res, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer res.Body.Close()

	content, err := io.ReadAll(res.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not parse response: %s", err.Error()), http.StatusInternalServerError)
	}

	var mainUserInfo models.MainProfileInfo
	err = json.Unmarshal(content, &mainUserInfo)
	if err != nil {
		log.Fatal("error unmarshaling json: ", err)
	}

	fmt.Fprint(w, mainUserInfo)

	// getting profile picture urls

	reqPic, err := http.NewRequest("GET", "https://api.linkedin.com/v2/me?projection=(id,first-name,last-name,email-address,profilePicture(displayImage~digitalmediaAsset:playableStreams))", nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req.Header.Set("Bearer", token.AccessToken)

	resPic, err := client.Do(reqPic)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer res.Body.Close()

	_, err = io.ReadAll(resPic.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not parse response: %s", err.Error()), http.StatusInternalServerError)
	}

	var picsResult models.ProfilePictureInfo
	err = njson.Unmarshal(content, &picsResult)
	if err != nil {
		log.Fatal("error unmarshaling json: ", err)
	}
	log.Printf("picsResult: %+v", picsResult)

	// fmt.Fprint(w, string(contentPic))

}

func GetLinkedInConfig() *oauth2.Config {
	return &oauth2.Config{
		RedirectURL:  utils.GoDotEnvVariable("LINKEDIN_REDIRECT_URL"),
		ClientID:     utils.GoDotEnvVariable("LINKEDIN_CLIENT_ID"),
		ClientSecret: utils.GoDotEnvVariable("LINKED_IN_SECRET"),
		Scopes:       []string{"r_emailaddress", "r_liteprofile"},
		Endpoint:     linkedin.Endpoint,
	}
}
