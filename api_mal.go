package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"
	"unicode"
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
	authorizeUrl := "https://myanimelist.net/v1/oauth2/authorize"
	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", "f0329e8fef42bf30a44e42dd24e25675")
	params.Add("code_challenge", codeChallenge)
	params.Add("state", "RequestID2235")

	authUrl := authorizeUrl + "?" + params.Encode()
	fmt.Printf("Allow Access to MyAnimeList account using this URL:\n%s", authUrl)
	fmt.Print("\n\nPaste auth token: ")
	var authCode string
	fmt.Scanf("%s", &authCode)

	oauthUrl := "https://myanimelist.net/v1/oauth2/token"
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
	fmt.Println("\nSuccessfully generated Access Token!")
	return body
}

type Season struct {
	// season must be one of "winter",
	// "summer", "spring", or "fall"
	Season string
	Year   int
}

func planToWatchApiCall(accessToken string, season Season) []string {
	planToWatchAnime := make([]string, 0)

	params := url.Values{}
	params.Add("status", "plan_to_watch")
	params.Add("fields", "start_season")
	url := fmt.Sprintf("%s/users/@me/animelist?%s", MAL_API, params.Encode())
	for {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Add("Authorization", "Bearer "+accessToken)

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
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

const TOKEN_FILE = "token.json"

func getCurrentSeason() Season {
	year, month, _ := time.Now().Date()
	var season string
	switch {
	case time.January <= month && month <= time.March:
		season = "winter"
	case time.April <= month && month <= time.June:
		season = "spring"
	case time.July <= month && month <= time.September:
		season = "summer"
	case time.October <= month && month <= time.December:
		season = "fall"
	}

	return Season{season, year}
}

func getNextSeason() Season {
	year, month, _ := time.Now().Date()
	var season string
	switch {
	case time.January <= month && month <= time.March:
		season = "spring"
	case time.April <= month && month <= time.June:
		season = "summer"
	case time.July <= month && month <= time.September:
		season = "fall"
	case time.October <= month && month <= time.December:
		season = "winter"
	}

	if season == "winter" {
		year += 1
	}

	return Season{season, year}
}

func GetPlanToWatchAnime(currentSeason bool) []string {
	var tokenString []byte
	tokenString, err := os.ReadFile(TOKEN_FILE)
	if os.IsNotExist(err) {
		accessTokenString := requestAccessToken()
		accessTokenString = append(accessTokenString, '\n')
		err := os.WriteFile(TOKEN_FILE, accessTokenString, 0666)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Saved Access Token in `%s`\n", TOKEN_FILE)
		tokenString = accessTokenString
	}

	var accessToken AuthToken
	err = json.Unmarshal(tokenString, &accessToken)
	if err != nil {
		log.Fatal(err)
	}
	// Check here if the accessToken is still valid and if not then refresh it
	// https://myanimelist.net/blog.php?eid=835707

	var season Season
	if currentSeason {
		season = getCurrentSeason()
	} else {
		season = getNextSeason()
	}

	// only easy way to quickly convert a string to Titlecase
	fmt.Printf("Retrieving Plan To Watch List for %c%s %d from MyAnimeList...\n",
		unicode.ToUpper(rune(season.Season[0])), season.Season[1:],
		season.Year,
	)
	return planToWatchApiCall(accessToken.AccessToken, season)
}
