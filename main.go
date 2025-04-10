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

const defaultRssFeedUrl = "https://subsplease.org/rss/?r=1080"

// Default Values (if not specified)

// TODO: get $USER from env
const animeDirLinux = "/home/aditya/Videos/Anime/"
const animeDirWindows = `D:\\Libraries\\Videos\\Anime\\`
const defaultPlatform = "all"

var downloadDirMap = map[string]string{
	"linux":   animeDirLinux,
	"windows": animeDirWindows,
}

func printHelp(verbose bool) {
	help :=
		`usage: rss-down-rules [-l|-w|-r|-d|-h|-H] <input-file>

Automatically generates a JSON file containing RSS downloader rules for qbittorrent`

	verboseHelp :=
		`The Input File must be formatted according to the syntax
    [Search Term]|[Entry Title]|[Save Path]

    Eg:-
    hunter x hunter|Hunter x Hunter|/home/$USER/Videos/Hunter x Hunter

    (Note that $USER is not automatically replaced by the username, it's
    just used as a placeholder value here)

    The terms must be entered in this order but skipped terms will
    be ignored and set to these default values --

    'Search Term' - Cannot be Skipped
    'Entry Title' - Will be set to same as the Search Term
    'Save Path' - Will be set to /home/$USER/Videos/Anime/[Search Term]`

	options :=
		`positional arguments:
    <input-file>        file to create JSON from (must come after flags)

options:
    -w 			generate file with download paths for windows
    -l 			generate file with download paths for linux
    -d, <filepath>	specify download path
    -r, <URL>		specify custom RSS feed URL
    -mal                generate input file from MyAnimeList plan to watch for next season
    -c                  get list for current season instead of next season when using '-mal'
    -h			show help message
    -H			show verbose help message`

	fmt.Println(help)
	if verbose {
		fmt.Println()
		fmt.Println(verboseHelp)
	}
	fmt.Println()
	fmt.Println(options)
}

func generateAndWriteJson(inputFilePath string, rssFeedUrl string, platform string, downloadDir string) {
	data, err := os.ReadFile(inputFilePath)
	if err != nil {
		printHelp(false)
		fmt.Println()
		log.Fatalf("%s\n", err.Error())
	}
	log.Println("Reading from file: ", inputFilePath)
	fileData := parseFile(string(data))

	outputFile := strings.TrimSuffix(inputFilePath, path.Ext(inputFilePath))

	if platform == "all" {
		linuxDownloadDir := downloadDirMap["linux"]
		windowsDownloadDir := downloadDirMap["windows"]
		if downloadDir != "" {
			linuxDownloadDir = downloadDir
			windowsDownloadDir = downloadDir
		}
		writeJson(fmt.Sprintf("%s_linux.json", outputFile), rssFeedUrl, linuxDownloadDir, fileData)
		writeJson(fmt.Sprintf("%s_windows.json", outputFile), rssFeedUrl, windowsDownloadDir, fileData)
	} else {
		platformDownloadDir := downloadDirMap[platform]
		if downloadDir != "" {
			platformDownloadDir = downloadDir
		}
		writeJson(fmt.Sprintf("%s.json", outputFile), rssFeedUrl, platformDownloadDir, fileData)
	}
}

func main() {
	log.SetFlags(0)

	platformLinux := flag.Bool("l", false, "")
	platformWindows := flag.Bool("w", false, "")
	help := flag.Bool("h", false, "")
	helpVerbose := flag.Bool("H", false, "")
	downloadDir := flag.String("d", "", "")
	rssFeedUrl := flag.String("r", defaultRssFeedUrl, "")
	getPlanToWatch := flag.Bool("mal", false, "")
	currentSeason := flag.Bool("c", false, "")
	flag.Usage = func() { fmt.Println(); printHelp(false) }
	flag.Parse()

	if *help || *helpVerbose {
		printHelp(*helpVerbose)
		os.Exit(0)
	}

	// Generating input file from Plan to Watch list
	// Cannot generate the RSS feed directly as the search terms
	// may not be always correct and need manual intervention
	if *getPlanToWatch {
		planToWatch := GetPlanToWatchAnime(*currentSeason)
		var planToWatchString strings.Builder
		for _, anime := range planToWatch {
			fmt.Fprintf(&planToWatchString, "%s|%s\n", anime, anime)
		}
		outputFile := "anime.txt"
		if err := os.WriteFile(outputFile, []byte(planToWatchString.String()), 0666); err != nil {
			log.Fatal(err)
		}
		log.Printf("Successfully Generated: %s", outputFile)
		os.Exit(0)
	}

	if *currentSeason {
		fmt.Printf("Option `-c` can only be applied to option `-mal`\n\n")
		printHelp(false)
		os.Exit(1)
	}

	// Generating rss feed from input file
	if flag.NArg() == 0 {
		fmt.Printf("No input file provided!\n\n")
		printHelp(false)
		os.Exit(1)
	}
	inputFilePath := flag.Args()[0]

	platform := defaultPlatform
	if *platformLinux && !*platformWindows {
		platform = "linux"
	} else if *platformWindows && !*platformLinux {
		platform = "windows"
	}

	generateAndWriteJson(inputFilePath, *rssFeedUrl, platform, *downloadDir)
}

type DownloadTitle struct {
	SearchTerm, Title, SavePath string
}

func parseFile(contents string) []DownloadTitle {
	downloadTitles := make([]DownloadTitle, 0)

	// Supports windows new lines
	re := regexp.MustCompile(`\r?\n`)
	for _, line := range re.Split(contents, -1) {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		line := strings.SplitN(line, "|", 3)
		searchTerm := line[0]
		var title, savePath string

		if len(line) >= 2 {
			title = line[1]
		}

		if len(line) == 3 {
			savePath = line[2]
		}
		downloadTitles = append(downloadTitles, DownloadTitle{searchTerm, title, savePath})
	}
	return downloadTitles
}

func writeJson(outputFilePath string, rssFeedUrl string, downloadDir string, downloadTitles []DownloadTitle) {
	jsonString := make([]byte, 0)

	jsonString = append(jsonString, "{"...)
	for i, title := range downloadTitles {
		if title.Title == "" {
			title.Title = title.SearchTerm
		}
		if title.SavePath == "" {
			title.SavePath = downloadDir + title.Title
		}
		jsonData := fmt.Sprintf(
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
            "torrentParams": {
                "category": "",
                "download_limit": -1,
                "download_path": "",
                "inactive_seeding_time_limit": -2,
                "operating_mode": "AutoManaged",
                "ratio_limit": 1,
                "save_path": "%s",
                "seeding_time_limit": -2,
                "share_limit_action": "Remove",
                "skip_checking": false,
                "ssl_certificate": "",
                "ssl_dh_params": "",
                "ssl_private_key": "",
                "tags": [
                ],
                "upload_limit": -1,
                "use_auto_tmm": false
            },
            "useRegex": false
        }`, title.Title, rssFeedUrl, title.SearchTerm, title.SavePath, title.SavePath)
		jsonString = append(jsonString, jsonData...)

		// Skip Trailing Comma
		if i < len(downloadTitles)-1 {
			jsonString = append(jsonString, ",\n"...)
		}
	}
	jsonString = append(jsonString, "\n}\n"...)

	err := os.WriteFile(outputFilePath, jsonString, 0666)
	if err != nil {
		log.Fatalf("%s\n", err.Error())
	}
	log.Printf("Successfully Generated: %s", outputFilePath)
}
