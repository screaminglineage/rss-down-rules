# rss-down-rules

Simple tool to generate RSS downloader rules JSON for Qbittorrent from a text file 

The Input File must be formatted according to the syntax:
```
[Search Term]|[Entry Title]|[Save Path]
```
For Example, `hunter x hunter|Hunter x Hunter|/home/$USER/Videos/Hunter x Hunter`

The terms must be entered in this order but skipped terms will be ignored and set to the following default values:
- 'Search Term' - Cannot be Skipped
- 'Entry Title' - Will be set to same as the Search Term
- 'Save Path' - Will be set to /home/$USER/Videos/Anime/[Search Term]


## TODO
- Look into refreshing the API token when it expires
- Implement auto detection of save path for both Linux and Windows
    - detect `$USER` directory for both
- maybe customize more features of the RSS feed generated



