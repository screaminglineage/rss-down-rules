package main

import (
	"fmt"
	"os"
	"encoding/json"
	"bytes"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
)

const MAL_API = "https://api.myanimelist.net/v2"

func generateCodeChallenge() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ.-_~")
	challengeBytes := make([]rune, 128)
	for i := range challengeBytes {
		challengeBytes[i] = letters[rand.Intn(len(letters))]
	}
	return string(challengeBytes)
}

func requestAccessToken() []byte {
	codeChallenge := generateCodeChallenge()
	authorizeUrl :=  "https://myanimelist.net/v1/oauth2/authorize"
	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id",  "f0329e8fef42bf30a44e42dd24e25675")
	params.Add("code_challenge", codeChallenge)
	params.Add("state", "RequestID2235")

	authUrl := authorizeUrl + "?" + params.Encode()
	fmt.Println(authUrl)
	fmt.Print("Enter auth token: ");
	var authCode string
	fmt.Scanf("%s", &authCode)

	oauthUrl :=  "https://myanimelist.net/v1/oauth2/token"
	params = url.Values{}
	params.Add("client_id", "f0329e8fef42bf30a44e42dd24e25675")
	params.Add("code", authCode)
	params.Add("code_verifier", codeChallenge)
	params.Add("grant_type", "authorization_code")
	
	res, err := http.Post(oauthUrl, "application/x-www-form-urlencoded", bytes.NewBufferString(params.Encode()))
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode > 299 {
		log.Fatalf("Response failed with code: %d and body: \n%s", res.StatusCode, body)
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully generated Access Token!")
	return body
}

type Season struct {
	// season must be one of "winter",
	// "summer", "spring", or "fall"
	Season string
	Year int
}

// Winter 2025
// [60108, 21, 55318, 58514, 56752, 58567, 54857, 57533, 51119, 59055, 57648, 59730, 59142, 53876, 58484, 59914, 55842, 54437, 54717, 58066, 56566, 6149, 56653, 1199, 56135, 8687, 59226, 56662, 59136, 59514, 57719, 52215, 57592, 56894, 56701, 57616, 966, 52995, 58572, 57181, 58939, 57066, 235, 49981, 1960, 57924, 32353, 57946, 53907, 59135, 58600, 59361, 53924, 55071, 55997, 58271, 57554, 58822, 8336, 57798, 59144, 50607, 37096, 42295, 4459, 58395, 58853, 60410, 59349, 54769, 50418, 58259, 57796, 58739, 59002, 59265, 58437, 58082, 56647, 56420, 57050, 59113, 2406, 60516, 59732, 59752, 54667, 59387, 60576, 60565, 59031, 58379, 59740, 59881, 58604, 58557, 60558, 60749, 56484, 60561, 60567, 59512, 59409, 60746, 60736, 60681, 60690, 60670, 60453, 60712, 60772, 60771, 60737, 60766, 60613, 60680, 60677, 60034, 60207, 60669, 59894, 60672, 60742, 60671, 59884, 59883, 58827, 59406, 60273, 59489, 60407, 49363, 54144, 60537, 58137, 30151, 59490, 60425, 59957, 59843, 58964, 54740, 59408, 60666, 48442, 52967, 59419, 60094, 57827, 59980, 30119, 60557, 22669, 59729, 38099, 57357, 41458, 54871, 59685, 38776, 59499, 60544, 10506, 57101, 60045, 58351, 18941, 60541, 53408]

func getPlanToWatch(accessToken string, season Season) []string {
	planToWatchAnime := make([]string, 0)

	params := url.Values{}
	params.Add("status", "plan_to_watch");
	params.Add("fields", "start_season");
	url := fmt.Sprintf("%s/users/@me/animelist?%s", MAL_API, params.Encode())
	for {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Add("Authorization", "Bearer " + accessToken)

		client := http.Client{}
		res, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()
	
		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode > 299 {
			log.Fatalf("Response failed with code: %d and body: \n%s", res.StatusCode, body)
		}
		if err != nil {
			log.Fatal(err)
		}

		// var jsonString bytes.Buffer
		// json.Indent(&jsonString, body, "", "\t")
		// fmt.Println(string(jsonString.Bytes()))

		var jsonData map[string]any
		json.Unmarshal(body, &jsonData)

		animeData := jsonData["data"].([]any)
		for i := range animeData {
			anime := animeData[i].(map[string]any)["node"].(map[string]any)
			seasonJson, ok := anime["start_season"].(map[string]any)
			
			if ok && int(seasonJson["year"].(float64)) == season.Year && seasonJson["season"].(string) == season.Season {
				planToWatchAnime = append(planToWatchAnime, anime["title"].(string))
			}
		}
		
		paging := jsonData["paging"].(map[string]any)
		nextUrl, found := paging["next"]
		if !found {
			break
		}
		url = nextUrl.(string)
	}
	return planToWatchAnime
}

type AuthToken struct {
	TokenType string    `json:"token_type"`
	ExpiresIn int       `json:"expires_in"`
	AccessToken string  `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

const TOKEN_FILE = "token.json"
func api_mal() {
	var tokenString []byte
	tokenString, err := os.ReadFile(TOKEN_FILE)
	if os.IsNotExist(err) {
		accessTokenString := requestAccessToken()
		// TODO: add a newline at the end of file
		err := os.WriteFile(TOKEN_FILE, accessTokenString, 0666)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Saved Access Token in `%s`\n", TOKEN_FILE)
		tokenString = []byte(accessTokenString)
	}

	var accessToken AuthToken 
	err = json.Unmarshal(tokenString, &accessToken)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Making Request")
	// TODO: get season from the current date
	planToWatchAnime := getPlanToWatch(accessToken.AccessToken, Season{"winter", 2025})
	fmt.Println(planToWatchAnime)
}
