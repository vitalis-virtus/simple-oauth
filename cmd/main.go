package main

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/vitalis-virtus/simple-oauth/linkedin"
	"github.com/vitalis-virtus/simple-oauth/utils"
	"golang.org/x/oauth2"

	// "golang.org/x/oauth2/linkedin"

	"golang.org/x/oauth2/google"
)

var (
	googleOAuthConfig   *oauth2.Config
	linkedInOAuthConfig *oauth2.Config
	randomState         = "random"
	linkedInState       = "ranodmLinkedInState"
)

func init() {
	googleOAuthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:8080/callback",
		ClientID:     utils.GoDotEnvVariable("GOOGLE_CLIENT_ID"),
		ClientSecret: utils.GoDotEnvVariable("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}

	// linkedInOAuthConfig = &oauth2.Config{
	// 	RedirectURL:  "http://localhost:8080/linkedincallback",
	// 	ClientID:     utils.GoDotEnvVariable("LINKEDIN_CLIENT_ID"),
	// 	ClientSecret: utils.GoDotEnvVariable("LINKED_IN_SECRET"),
	// 	Scopes:       []string{"r_basicprofile", "r_emailaddress"},
	// 	Endpoint:     linkedin.Endpoint,
	// }
}

func main() {
	http.HandleFunc("/", handleHome)

	// todo change to "login-..."
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/login-linkedin", handleLoginLinkedIn)
	//todo change to "callback-google"
	http.HandleFunc("/callback", handleCallback)
	http.HandleFunc("/callback-linkedin", linkedin.Callback)

	http.ListenAndServe(":8080", nil)
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	var htmlIndex = `
	<html>
		<body>
			<a href="/login">Google Log In</a>
			</br>
			<a href="/login-linkedin">LinkedIn Log In</a>
		</body>
	</html>
	`
	fmt.Fprint(w, htmlIndex)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	url := googleOAuthConfig.AuthCodeURL(randomState)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleLoginLinkedIn(w http.ResponseWriter, r *http.Request) {
	url := linkedin.GetLinkedInConfig().AuthCodeURL(linkedin.State)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	content, err := getUserGoogleInfo(r.FormValue("state"), r.FormValue("code"))

	if err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	fmt.Fprint(w, string(content))
}

// func handleLinkedInCallback(w http.ResponseWriter, r *http.Request) {
// 	content, err := getUserLinkedInInfo(r.FormValue("state"), r.FormValue("code"))
// 	if err != nil {
// 		fmt.Println(err.Error())
// 		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
// 		return
// 	}

// 	fmt.Fprint(w, string(content))
// }

func getUserGoogleInfo(state string, code string) ([]byte, error) {
	if state != randomState {
		return nil, fmt.Errorf("invalid oauth state")
	}

	fmt.Println("Code: ", code)

	token, err := googleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}

	res, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}

	defer res.Body.Close()

	content, err := io.ReadAll(res.Body)

	if err != nil {
		fmt.Printf("could not parse response: %s\n", err.Error())

		return nil, fmt.Errorf("could not parse resonse: %s", err.Error())
	}

	return content, nil
}

// func getUserLinkedInInfo(state string, code string) ([]byte, error) {
// 	if state != linkedInState {
// 		return nil, fmt.Errorf("invalid oauth state")
// 	}

// 	token, err := linkedInOAuthConfig.Exchange(context.Background(), code)
// 	if err != nil {
// 		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
// 	}

// 	client := linkedInOAuthConfig.Client(context.Background(), token)

// 	req, err := http.NewRequest("GET", "https://api.linkedin.com/v1/people/~:(email-address,first-name,last-name,id,headline)?format=json", nil)
// 	if err != nil {
// 		return nil, fmt.Errorf(err.Error())
// 	}

// 	req.Header.Set("Bearer", token.AccessToken)
// 	res, err := client.Do(req)

// 	// res, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
// 	}

// 	defer res.Body.Close()

// 	content, err := io.ReadAll(res.Body)

// 	if err != nil {
// 		fmt.Printf("could not parse response: %s\n", err.Error())

// 		return nil, fmt.Errorf("could not parse resonse: %s", err.Error())
// 	}

// 	return content, nil

// }
