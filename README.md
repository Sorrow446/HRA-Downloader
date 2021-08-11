# HRA-Downloader
Go port of my HIGHRESAUDIO tool.
![](https://i.imgur.com/IcMjhCo.png)
[Windows, Linux, macOS and Android binaries](https://github.com/Sorrow446/HRA-Downloader/releases)

# Setup
Input credentials into config file.
Configure any other options if needed.
|Option|Info|
| --- | --- |
|email|Email address.
|password|Password.
|outPath|Where to download to. Path will be made if it doesn't already exist.
|trackTemplate|Track filename template. Vars: album, albumArtist, artist, copyright, genre, isrc, title, track, trackPad, trackTotal, upc, year.
|downloadBooklets|Download digital booklets when available.
|maxCoverSize|Fetch covers in their max sizes. true = max, false = smaller.
|keepCover|Keep covers in album folders.
|language|Metadata language, en or de.

# Usage
Args take priority over the config file.

Download two albums:   
`hra_dl_x64.exe https://www.highresaudio.com/en/album/view/swsct3/norah-jones-til-we-meet-again-live https://www.highresaudio.com/en/album/view/pdqrig/supertramp-even-in-the-quietest-moments-remastered`

Download a single album and from two text files:   
`hra_dl_x64.exe https://www.highresaudio.com/en/album/view/swsct3/norah-jones-til-we-meet-again-live G:\1.txt G:\2.txt`

```
 _____ _____ _____    ____                _           _
|  |  | __  |  _  |  |    \ ___ _ _ _ ___| |___ ___ _| |___ ___
|     |    -|     |  |  |  | . | | | |   | | . | .'| . | -_|  _|
|__|__|__|__|__|__|  |____/|___|_____|_|_|_|___|__,|___|___|_|

Usage: main.exe [--outpath OUTPATH] URLS [URLS ...]

Positional arguments:
  URLS

Options:
  --outpath OUTPATH, -o OUTPATH
  --help, -h             display this help and exit
  ```
 
# Disclaimer
- I will not be responsible for how you use HRA Downloader.    
- HIGHRESAUDIO brand and name is the registered trademark of its respective owner.    
- HRA Downloader has no partnership, sponsorship or endorsement with HIGHRESAUDIO.
