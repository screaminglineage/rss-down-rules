# rss-down-rules

Simple tool to generate RSS downloader rules JSON for Qbittorrent from a text file 

The Input File must be formatted according to the syntax:
    	[Search Term]|[Entry Title]|[Save Path]
    	eg.
    	made in abyss|Made in Abyss|/home/username/Videos/Made in Abyss

    	The terms must be entered in this order but skipped terms will
    	be ignored and set to the default values --

    	'Search Term' - Cannot be Skipped
    	'Entry Title' - Will be set to same as the Search Term
    	'Save Path' - Will be set to /home/$USER/Videos/Anime/[Search Term]

## TODO
- Implement auto detection of save path for both Linux and Windows
    - detect $USER directory for both
- add command line flags to pass in save path/save path root
- add command line flag to pass in RSS feed URL
- maybe customize more features of the RSS feed generated



