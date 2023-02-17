package linkedin

import (
	"context"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/vitalis-virtus/simple-oauth/models"
	"github.com/vitalis-virtus/simple-oauth/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/linkedin"
	"io"
	"net/http"
)

var (
	State        = "linkedin_state"
	emailInfoUrl = "https://api.linkedin.com/v2/emailAddress?q=members&projection=(elements*(handle~))&oauth2_access_token="
	userInfoUrl  = "https://api.linkedin.com/v2/me"
	userPicUrl   = "https://api.linkedin.com/v2/me?projection=(id,firstName,lastName,profilePicture(displayImage~:playableStreams))"
)

func Callback(w http.ResponseWriter, r *http.Request) {
	var UserProfileInfo models.ProfileInfo
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

	// get user email
	reqUserEmail, err := http.NewRequest("GET", emailInfoUrl, nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	reqUserEmail.Header.Set("Bearer", token.AccessToken)

	resUserEmail, err := client.Do(reqUserEmail)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer resUserEmail.Body.Close()

	contentUserEmail, err := io.ReadAll(resUserEmail.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not parse response: %s", err.Error()), http.StatusInternalServerError)
	}

	UserProfileInfo.Email = gjson.Get(string(contentUserEmail), "elements.0.handle~.emailAddress").String()

	// get user info
	reqUserInfo, err := http.NewRequest("GET", userInfoUrl, nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	reqUserInfo.Header.Set("Bearer", token.AccessToken)

	resUserInfo, err := client.Do(reqUserInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer resUserEmail.Body.Close()

	contentUserInfo, err := io.ReadAll(resUserInfo.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not parse response: %s", err.Error()), http.StatusInternalServerError)
	}

	UserProfileInfo.ID = gjson.Get(string(contentUserInfo), "id").String()
	UserProfileInfo.FirstName = gjson.Get(string(contentUserInfo), "localizedFirstName").String()
	UserProfileInfo.LastName = gjson.Get(string(contentUserInfo), "localizedLastName").String()

	// get user pic
	reqUserPic, err := http.NewRequest("GET", userPicUrl, nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	reqUserPic.Header.Set("Bearer", token.AccessToken)

	resUserPic, err := client.Do(reqUserPic)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer resUserPic.Body.Close()

	contentUserPic, err := io.ReadAll(resUserPic.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not parse response: %s", err.Error()), http.StatusInternalServerError)
	}

	UserProfileInfo.Picture = gjson.Get(string(contentUserPic), "profilePicture.displayImage~.elements.#.identifiers.0.identifier").String()

	fmt.Fprint(w, UserProfileInfo)

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
