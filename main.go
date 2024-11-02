package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

const RSS_FEED = "https://subsplease.org/rss/?r=1080"

// Default Values (if not specified)
const OUTPUT_FILE = "anime.json"
const OUTPUT_FILE_WINDOWS = "anime_windows.json"

// TODO: get $USER from env
const ANIME_DIR = "/home/aditya/Videos/Anime/"
const ANIME_DIR_WINDOWS = "D:\\Libraries\\Videos\\Anime\\"

func help(verbose bool) {
	help :=
		`usage: rss-down-rules [input-file]
    	Automatically generates a JSON file containing RSS downloader rules for qbittorrent`

	verbose_help :=
		`The Input File must be formatted according to the syntax
    	[Search Term]|[Entry Title]|[Save Path]
    	eg.
    	made in abyss|Made in Abyss|/home/username/Videos/Made in Abyss

    	The terms must be entered in this order but skipped terms will
    	be ignored and set to the default values --

    	'Search Term' - Cannot be Skipped
    	'Entry Title' - Will be set to same as the Search Term
    	'Save Path' - Will be set to /home/aditya/Videos/Anime/[Search Term]`

	options :=
		`positional arguments:
	input-file                 file to create JSON from
 
options:
	--h, --help		show help message (--help shows verbose help)`

	fmt.Println(help)
	if verbose {
		fmt.Println()
		fmt.Println(verbose_help)
	}
	fmt.Println()
	fmt.Println(options)
}

func main() {
	log := log.New(os.Stderr, "", 0)
	if len(os.Args) < 2 {
		help(true)
		fmt.Println()
		log.Fatalln("Not enough arguments, expected input file")
	}

	if os.Args[1] == "-h" {
		help(false)
		os.Exit(0)
	}
	if os.Args[1] == "--help" {
		help(true)
		os.Exit(0)
	}

	input_file_path := os.Args[1]
	if strings.HasPrefix(input_file_path, "-") {
		help(false)
		if input_file_path == "-help" {
			log.Fatalf("Unexpected argument found: `%s`, try `--help` instead for verbose help\n", input_file_path)
		}
		log.Fatalf("Unexpected argument found: `%s`\n", input_file_path)
	}

	data, err := os.ReadFile(input_file_path)
	if err != nil {
		help(false)
		fmt.Println()
		log.Fatalf("%s\n", err.Error())
	}
	json := generateJSON(parseFile(string(data)))
	err = os.WriteFile(OUTPUT_FILE, json, 0666)
	if err != nil {
		log.Fatalf("%s\n", err.Error())
	}
}

type DownloadTitle struct {
	search_term, title, save_path string
}

func parseFile(contents string) []DownloadTitle {
	download_titles := make([]DownloadTitle, 0)

	// Supports windows new lines
	re := regexp.MustCompile(`\r?\n`)
	for _, line := range re.Split(contents, -1) {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		line := strings.SplitN(line, "|", 3)
		search_term := line[0]
		var title, save_path string

		if len(line) >= 2 {
			title = line[1]
		}

		if len(line) == 3 {
			save_path = line[2]
		}
		download_titles = append(download_titles, DownloadTitle{search_term, title, save_path})
	}
	return download_titles
}

func generateJSON(download_titles []DownloadTitle) []byte {
	json_string := make([]byte, 0)

	json_string = append(json_string, "{"...)
	for i, title := range download_titles {
		if title.title == "" {
			title.title = title.search_term
		}
		if title.save_path == "" {
			title.save_path = ANIME_DIR + title.title
		}
		json_data := fmt.Sprintf(
			`
			"%s": {
				"addPaused": null,
				"affectedFeeds": [
					"%s"
				],
				"assignedCategory": "",
				"enabled": true,
				"episodeFilter": "",
				"ignoreDays": 0,
				"lastMatch": "",
				"mustContain": "%s",
				"mustNotContain": "",
				"previouslyMatchedEpisodes": [
				],
				"savePath": "%s",
				"smartFilter": true,
				"torrentContentLayout": null,
				"useRegex": false
			}`, title.title, RSS_FEED, title.search_term, title.save_path)
		json_string = append(json_string, json_data...)

		// Skip Trailing Comma
		if i < len(download_titles)-1 {
			json_string = append(json_string, ",\n"...)
		}
	}
	json_string = append(json_string, "\n}\n"...)
	return json_string
}
