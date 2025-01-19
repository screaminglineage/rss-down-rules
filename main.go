package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
)

const RSS_FEED = "https://subsplease.org/rss/?r=1080"

// Default Values (if not specified)

// TODO: get $USER from env
const ANIME_DIR_LINUX = "/home/aditya/Videos/Anime/"
const ANIME_DIR_WINDOWS = `D:\\Libraries\\Videos\\Anime\\`
const DEFAULT_PLATFORM = "all"

var platformDownloadDir = map[string]string{
	"linux":   ANIME_DIR_LINUX,
	"windows": ANIME_DIR_WINDOWS,
}

func print_help(verbose bool) {
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
    -w 				generate file with download paths for windows
    -l 				generate file with download paths for linux
    -d, <filepath>	        specify download path
    -r, <URL>		        specify custom rss URL
    -h				show help message
    -H				show verbose help message`

	fmt.Println(help)
	if verbose {
		fmt.Println()
		fmt.Println(verbose_help)
	}
	fmt.Println()
	fmt.Println(options)
}

func writeJSON(output_file_path string, platformDownloadDir string, fileData []DownloadTitle) {
	json := generateJSON(fileData, platformDownloadDir)

	err := os.WriteFile(output_file_path, json, 0666)
	if err != nil {
		log.Fatalf("%s\n", err.Error())
	}
	log.Printf("Succesfully Generated: %s", output_file_path)
}

func generateAndWriteJSON(input_file_path string, platform string, download_dir string) {
	data, err := os.ReadFile(input_file_path)
	if err != nil {
		print_help(false)
		fmt.Println()
		log.Fatalf("%s\n", err.Error())
	}
	log.Println("Reading from file: ", input_file_path)
	fileData := parseFile(string(data))

	output_file := strings.TrimSuffix(input_file_path, path.Ext(input_file_path))

	linux_download_dir := platformDownloadDir["linux"]
	windows_download_dir := platformDownloadDir["windows"]
	if download_dir != "" {
		linux_download_dir = download_dir
		windows_download_dir = download_dir
	}
	if platform == "all" {
		writeJSON(fmt.Sprintf("%s_linux.json", output_file), linux_download_dir, fileData)
		writeJSON(fmt.Sprintf("%s_windows.json", output_file), windows_download_dir, fileData)
	} else {
		writeJSON(fmt.Sprintf("%s.json", output_file), download_dir, fileData)
	}
}

func main() {
	log.SetFlags(0)

	platform_linux := flag.Bool("l", false, "generate file with download paths for linux")
	platform_windows := flag.Bool("w", false, "generate file with download paths for windows")
	help := flag.Bool("h", false, "show help")
	help_verbose := flag.Bool("H", false, "show verbose help")
	download_dir := flag.String("d", "", "specify download path")
	flag.Usage = func() { fmt.Println(); print_help(false) }
	flag.Parse()

	if *help || *help_verbose {
		print_help(*help_verbose)
		return
	}

	if flag.NArg() == 0 {
		fmt.Printf("No input file provided\n\n")
		print_help(false)
		return
	}
	input_file_path := flag.Args()[0]

	platform := "all"
	if *platform_linux && !*platform_windows {
		platform = "linux"
	} else if *platform_windows && !*platform_linux {
		platform = "windows"
	}

	generateAndWriteJSON(input_file_path, platform, *download_dir)
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
			title.save_path = downloadDir + "/" + title.title
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
