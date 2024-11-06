package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
)

const RSS_FEED = "https://subsplease.org/rss/?r=1080"

// Default Values (if not specified)
const DEFAULT_OUTPUT_FILE = "anime.json"

// TODO: get $USER from env
const ANIME_DIR_LINUX = "/home/aditya/Videos/Anime/"
const ANIME_DIR_WINDOWS = "D:\\\\Libraries\\\\Videos\\\\Anime\\\\"
const DEFAULT_PLATFORM = "all"

var platformDownloadDir = map[string]string{
	"linux":   ANIME_DIR_LINUX,
	"windows": ANIME_DIR_WINDOWS,
}

func help(verbose bool) {
	help :=
		`usage: rss-down-rules [input-file]
    	Automatically generates a JSON file containing RSS downloader rules for qbittorrent`

	verbose_help :=
		`The Input File must be formatted according to the syntax
    	[Search Term]|[Entry Title]|[Save Path]
		
	Eg:-
	hunter x hunter|Hunter x Hunter|/home/$USER/Videos/Hunter x Hunter

    	The terms must be entered in this order but skipped terms will
    	be ignored and set to the default values --

    	'Search Term' - Cannot be Skipped
    	'Entry Title' - Will be set to same as the Search Term
    	'Save Path' - Will be set to /home/aditya/Videos/Anime/[Search Term]`

	options :=
		`positional arguments:
  input-file                 file to create JSON from
 
options:
  -w, --windows				generate file with download paths for windows
  -l, --linux				generate file with download paths for linux
  -d, --download <filepath>	specify download path
  -r, --rss <URL>			specify custom rss URL
  -h, --help				show help message (--help shows verbose help)`

	fmt.Println(help)
	if verbose {
		fmt.Println()
		fmt.Println(verbose_help)
	}
	fmt.Println()
	fmt.Println(options)
}

func writeJSON(output_file_path string, fileData []DownloadTitle, platform string) {
	json := generateJSON(fileData, platformDownloadDir[platform])

	err := os.WriteFile(DEFAULT_OUTPUT_FILE, json, 0666)
	if err != nil {
		log.Fatalf("%s\n", err.Error())
	}
	log.Printf("Succesfully Generated: %s.json", output_file_path)
}

func generateAndWriteJSON(input_file_path string, platform string) {
	data, err := os.ReadFile(input_file_path)
	if err != nil {
		help(false)
		fmt.Println()
		log.Fatalf("%s\n", err.Error())
	}
	log.Println("Reading from file: ", input_file_path)

	fileData := parseFile(string(data))
	output_file_path := strings.TrimSuffix(input_file_path, path.Ext(input_file_path))
	if platform == "all" {
		writeJSON(output_file_path, fileData, "linux")
		writeJSON(output_file_path, fileData, "windows")
	} else {
		writeJSON(output_file_path, fileData, platformDownloadDir[platform])
	}
}

func main() {
	log := log.New(os.Stderr, "", 0)
	cmdArgs := make(map[string]string)
	i := 1
	for i < len(os.Args) {
		arg := os.Args[i]

		if arg == "-h" {
			help(false)
			os.Exit(0)
		} else if arg == "--help" {
			help(true)
			os.Exit(0)
		} else if arg == "-w" || arg == "--windows" {
			cmdArgs["platform"] = "windows"
		} else if arg == "-l" || arg == "--linux" {
			cmdArgs["platform"] = "linux"
		} else if arg == "-d" || arg == "--download" {
			if i+1 >= len(os.Args) {
				help(false)
				log.Fatalf("Expected download path after argument: `%s`", arg)
			}
			cmdArgs["downloadPath"] = os.Args[i+1]
			i += 1
		} else if arg == "-r" || arg == "--rss" {
			if i+1 >= len(os.Args) {
				help(false)
				log.Fatalf("Expected URL after argument: `%s`", arg)
			}
			cmdArgs["rssURL"] = os.Args[i+1]
			i += 1

		} else {
			inputFilePath := arg
			if strings.HasPrefix(inputFilePath, "-") {
				help(false)
				if inputFilePath == "-help" {
					log.Fatalf("Unexpected argument found: `%s`, try `--help` instead for verbose help\n", inputFilePath)
				}
				log.Fatalf("Unexpected argument found: `%s`\n", inputFilePath)
			}
			cmdArgs["inputFilePath"] = inputFilePath
		}
		i += 1
	}
	if inputFilePath, ok := cmdArgs["inputFilePath"]; ok {
		platform, ok := cmdArgs["platform"]
		if !ok {
			platform = DEFAULT_PLATFORM
		}
		generateAndWriteJSON(inputFilePath, platform)

	} else {
		help(true)
		log.Fatal("Expected input file path")
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

func generateJSON(download_titles []DownloadTitle, downloadDir string) []byte {
	json_string := make([]byte, 0)

	json_string = append(json_string, "{"...)
	for i, title := range download_titles {
		if title.title == "" {
			title.title = title.search_term
		}
		if title.save_path == "" {
			title.save_path = downloadDir + title.title
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
