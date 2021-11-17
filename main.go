package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/dustin/go-humanize"
	"github.com/go-flac/flacpicture"
	"github.com/go-flac/flacvorbis"
	"github.com/go-flac/go-flac"
)

var (
	jar, _ = cookiejar.New(nil)
	client = &http.Client{Jar: jar, Transport: &MyTransport{}}
)

func (t *MyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add(
		"User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 "+
			"(KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36",
	)
	req.Header.Add(
		"Referer", "https://stream-app.highresaudio.com/dashboard",
	)
	return http.DefaultTransport.RoundTrip(req)
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Downloaded += uint64(n)
	percentage := float64(wc.Downloaded) / float64(wc.Total) * float64(100)
	wc.Percentage = int(percentage)
	fmt.Printf("\r%d%%, %s/%s ", wc.Percentage, humanize.Bytes(wc.Downloaded), wc.TotalStr)
	return n, nil
}

func getScriptDir() (string, error) {
	var (
		ok    bool
		err   error
		fname string
	)
	if filepath.IsAbs(os.Args[0]) {
		_, fname, _, ok = runtime.Caller(0)
		if !ok {
			return "", errors.New("Failed to get script filename.")
		}
	} else {
		fname, err = os.Executable()
		if err != nil {
			return "", err
		}
	}
	scriptDir := filepath.Dir(fname)
	return scriptDir, nil
}

func readTxtFile(path string) ([]string, error) {
	var lines []string
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, strings.TrimSpace(scanner.Text()))
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}
	return lines, nil
}

func contains(lines []string, value string) bool {
	for _, line := range lines {
		if strings.EqualFold(line, value) {
			return true
		}
	}
	return false
}

func processUrls(urls []string) ([]string, error) {
	var (
		processed []string
		txtPaths  []string
	)
	for _, url := range urls {
		if strings.HasSuffix(url, ".txt") && !contains(txtPaths, url) {
			txtLines, err := readTxtFile(url)
			if err != nil {
				return nil, err
			}
			for _, txtLine := range txtLines {
				if !contains(processed, txtLine) {
					processed = append(processed, txtLine)
				}
			}
			txtPaths = append(txtPaths, url)
		} else {
			if !contains(processed, url) {
				processed = append(processed, url)
			}
		}
	}
	return processed, nil
}

func readConfig() (*Config, error) {
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		return nil, err
	}
	var obj Config
	err = json.Unmarshal(data, &obj)
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func parseArgs() *Args {
	var args Args
	arg.MustParse(&args)
	return &args
}

func parseCfg() (*Config, error) {
	cfg, err := readConfig()
	if err != nil {
		return nil, err
	}
	args := parseArgs()
	if !(cfg.Language == "en" || cfg.Language == "de") {
		return nil, errors.New("Language must be en or de.")
	}
	if cfg.OutPath == "" {
		cfg.OutPath = "HRA downloads"
	}
	cfg.Urls, err = processUrls(args.Urls)
	if err != nil {
		errString := fmt.Sprintf("Failed to process URLs.\n%s", err)
		return nil, errors.New(errString)
	}
	return cfg, nil
}

func auth(email, pwd string) (string, error) {
	_url := "https://streaming.highresaudio.com:8182/vault3/user/login"
	req, err := http.NewRequest(http.MethodGet, _url, nil)
	if err != nil {
		return "", err
	}
	query := url.Values{}
	query.Set("username", email)
	query.Set("password", pwd)
	req.URL.RawQuery = query.Encode()
	do, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return "", errors.New(do.Status)
	}
	var obj Auth
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return "", err
	}
	if obj.ResponseStatus != "OK" {
		return "", errors.New("Bad response.")
	} else if !obj.HasSubscription {
		return "", errors.New("Subscription required.")
	}
	userData, err := json.Marshal(&obj)
	if err != nil {
		return "", err
	}
	return string(userData), nil
}

func getAlbumId(url string) (string, error) {
	req, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer req.Body.Close()
	if req.StatusCode != http.StatusOK {
		return "", errors.New(req.Status)
	}
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return "", err
	}
	bodyString := string(bodyBytes)
	regexString := `data-id="([a-z\d]{8}-[a-z\d]{4}-[a-z\d]{4}-[a-z\d]{4}-[a-z\d]{12})"`
	regex := regexp.MustCompile(regexString)
	match := regex.FindStringSubmatch(bodyString)
	if match == nil {
		return "", errors.New("No regex match.")
	}
	return match[1], nil

}

func getMeta(albumId, userData, lang string) (*AlbumMeta, error) {
	_url := "https://streaming.highresaudio.com:8182/vault3/vault/album/"
	req, err := http.NewRequest(http.MethodGet, _url, nil)
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("album_id", albumId)
	query.Set("userData", userData)
	query.Set("lang", lang)
	req.URL.RawQuery = query.Encode()
	do, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return nil, errors.New(do.Status)
	}
	var obj AlbumMeta
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return nil, err
	}
	if obj.ResponseStatus != "OK" {
		return nil, errors.New("Bad response.")
	}
	return &obj, nil
}

func makeDir(path string) error {
	err := os.MkdirAll(path, 0755)
	return err
}

func fileExists(path string) (bool, error) {
	f, err := os.Stat(path)
	if err == nil {
		return !f.IsDir(), nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func checkUrl(url string) bool {
	regexes := [2]string{
		`^https://www.highresaudio.com/(?:en|de)/album/view/[a-z\d]+/[a-z\d-]+$`,
		`^https://stream-app.highresaudio.com/album/[a-z\d]{8}-[a-z\d]{4}-[a-z\d]{4}-[a-z\d]{4}-[a-z\d]{12}$`,
	}
	for _, regexString := range regexes {
		regex := regexp.MustCompile(regexString)
		match := regex.MatchString(url)
		if match {
			return true
		}
	}
	return false
}

func parseAlbumMeta(meta *AlbumMeta) map[string]string {
	parsedMeta := map[string]string{
		"album":       meta.Data.Results.Title,
		"albumArtist": meta.Data.Results.Artist,
		"copyright":   meta.Data.Results.Copyright,
		"upc":         meta.Data.Results.UPC,
		"year":        strconv.Itoa(meta.Data.Results.ReleaseDate.Year()),
	}
	return parsedMeta
}

func parseTrackMeta(meta *TrackMeta, albMeta map[string]string, trackNum, trackTotal int) map[string]string {
	albMeta["artist"] = meta.Artist
	albMeta["genre"] = meta.Genre
	albMeta["isrc"] = meta.ISRC
	albMeta["title"] = meta.Title
	albMeta["track"] = strconv.Itoa(trackNum)
	albMeta["trackPad"] = fmt.Sprintf("%02d", trackNum)
	albMeta["trackTotal"] = strconv.Itoa(trackTotal)
	return albMeta
}

func sanitize(filename string) string {
	regex := regexp.MustCompile(`[\/:*?"><|]`)
	sanitized := regex.ReplaceAllString(filename, "_")
	return sanitized
}

func downloadCover(meta *Covers, path string, maxCover bool) error {
	_url := "https://"
	if maxCover {
		_url += meta.Master.FileURL
	} else {
		_url += meta.Preview.FileURL
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	req, err := client.Get(_url)
	if err != nil {
		return err
	}
	defer req.Body.Close()
	if req.StatusCode != http.StatusOK {
		return errors.New(req.Status)
	}
	_, err = io.Copy(f, req.Body)
	return err
}

func parseTemplate(templateText string, tags map[string]string) string {
	var buffer bytes.Buffer
	for {
		err := template.Must(template.New("").Parse(templateText)).Execute(&buffer, tags)
		if err == nil {
			break
		}
		fmt.Println("Failed to parse template. Default will be used instead.")
		templateText = "{{.trackPad}}. {{.title}}"
		buffer.Reset()
	}
	return buffer.String()
}

func downloadTrack(trackPath, url string) error {
	f, err := os.Create(trackPath)
	if err != nil {
		return err
	}
	defer f.Close()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Range", "bytes=0-")
	do, err := client.Do(req)
	if err != nil {
		return err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK && do.StatusCode != http.StatusPartialContent {
		return errors.New(do.Status)
	}
	totalBytes := uint64(do.ContentLength)
	counter := &WriteCounter{Total: totalBytes, TotalStr: humanize.Bytes(totalBytes)}
	_, err = io.Copy(f, io.TeeReader(do.Body, counter))
	fmt.Println("")
	return err
}

// Tracks already come tagged, but come with missing meta and smaller artwork.
func writeTags(trackPath, coverPath string, tags map[string]string) error {
	var (
		err     error
		imgData []byte
	)
	if coverPath != "" {
		imgData, err = ioutil.ReadFile(coverPath)
		if err != nil {
			return err
		}
	}
	delete(tags, "trackPad")
	f, err := flac.ParseFile(trackPath)
	if err != nil {
		return err
	}
	cmt, err := flacvorbis.ParseFromMetaDataBlock(*f.Meta[1])
	if err != nil {
		return err
	}
	f.Meta = f.Meta[:1]
	tag := flacvorbis.New()
	tag.Vendor = cmt.Vendor
	for k, v := range tags {
		tag.Add(strings.ToUpper(k), v)
	}
	tagMeta := tag.Marshal()
	f.Meta = append(f.Meta, &tagMeta)
	if imgData != nil {
		picture, err := flacpicture.NewFromImageData(
			flacpicture.PictureTypeFrontCover, "", imgData, "image/jpeg",
		)
		if err != nil {
			return err
		}
		pictureMeta := picture.Marshal()
		f.Meta = append(f.Meta, &pictureMeta)
	}
	err = f.Save(trackPath)
	return err
}

func downloadBooklet(path, url string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	req, err := client.Get("https://" + url)
	if err != nil {
		return err
	}
	defer req.Body.Close()
	if req.StatusCode != http.StatusOK {
		return errors.New(req.Status)
	}
	_, err = io.Copy(f, req.Body)
	return err
}

func checkAvail(availAt time.Time) bool {
	return time.Now().Unix() >= availAt.Unix()
}

func main() {
	fmt.Println(`                                                           
 _____ _____ _____    ____                _           _         
|  |  | __  |  _  |  |    \ ___ _ _ _ ___| |___ ___ _| |___ ___ 
|     |    -|     |  |  |  | . | | | |   | | . | .'| . | -_|  _|
|__|__|__|__|__|__|  |____/|___|_____|_|_|_|___|__,|___|___|_| 
`)
	scriptDir, err := getScriptDir()
	if err != nil {
		panic(err)
	}
	err = os.Chdir(scriptDir)
	if err != nil {
		panic(err)
	}
	cfg, err := parseCfg()
	if err != nil {
		errString := fmt.Sprintf("Failed to parse config file.\n%s", err)
		panic(errString)
	}
	userData, err := auth(cfg.Email, cfg.Password)
	if err != nil {
		panic(err)
	}
	fmt.Println("Signed in successfully.\n")
	err = makeDir(cfg.OutPath)
	if err != nil {
		errString := fmt.Sprintf("Failed to make output folder.\n%s", err)
		panic(errString)
	}
	albumTotal := len(cfg.Urls)
	for albumNum, url := range cfg.Urls {
		fmt.Printf("Album %d of %d:\n", albumNum+1, albumTotal)
		ok := checkUrl(url)
		if !ok {
			fmt.Println("Invalid URL:", url)
			continue
		}
		albumId, err := getAlbumId(url)
		if err != nil {
			fmt.Println("Failed to extract album ID.\n", err)
			continue
		}
		meta, err := getMeta(albumId, userData, cfg.Language)
		if err != nil {
			fmt.Println("Failed to get metadata.\n", err)
			continue
		}
		availAt := meta.Data.Results.AvailableFrom
		ok = checkAvail(availAt)
		if !ok {
			fmt.Printf("Album unavailable. Available at: %v", availAt)
			continue
		}
		parsedAlbMeta := parseAlbumMeta(meta)
		albFolder := parsedAlbMeta["albumArtist"] + " - " + parsedAlbMeta["album"]
		fmt.Println(albFolder)
		if len(albFolder) > 120 {
			fmt.Println("Album folder was chopped as it exceeds 120 characters.")
			albFolder = albFolder[:120]
		}
		albumPath := filepath.Join(cfg.OutPath, sanitize(albFolder))
		err = makeDir(albumPath)
		if err != nil {
			fmt.Println("Failed to make album folder.\n", err)
			continue
		}
		coverPath := filepath.Join(albumPath, "cover.jpg")
		err = downloadCover(&meta.Data.Results.Cover, coverPath, cfg.MaxCoverSize)
		if err != nil {
			fmt.Println("Failed to get cover.\n", err)
			coverPath = ""
		}
		bookletPath := filepath.Join(albumPath, "booklet.pdf")
		if meta.Data.Results.Booklet != "" && cfg.DownloadBooklets {
			fmt.Println("Downloading booklet...")
			err = downloadBooklet(bookletPath, meta.Data.Results.Booklet)
			if err != nil {
				fmt.Println("Failed to download booklet.\n", err)
			}
		}
		trackTotal := len(meta.Data.Results.Tracks)
		for trackNum, track := range meta.Data.Results.Tracks {
			trackNum++
			parsedMeta := parseTrackMeta(&track, parsedAlbMeta, trackNum, trackTotal)
			trackFname := parseTemplate(cfg.TrackTemplate, parsedMeta)
			trackPath := filepath.Join(albumPath, sanitize(trackFname)+".flac")
			exists, err := fileExists(trackPath)
			if err != nil {
				fmt.Println("Failed to check if track already exists locally.\n", err)
				continue
			}
			if exists {
				fmt.Println("Track already exists locally.")
				continue
			}
			// Bit depth isn't provided.
			fmt.Printf(
				"Downloading track %d of %d: %s - %s kHz FLAC\n",
				trackNum, trackTotal, parsedMeta["title"], track.Format,
			)
			err = downloadTrack(trackPath, track.URL)
			if err != nil {
				fmt.Println("Failed to download track.\n", err)
				continue
			}
			err = writeTags(trackPath, coverPath, parsedMeta)
			if err != nil {
				fmt.Println("Failed to write tags.\n", err)
			}
		}
		if coverPath != "" && !cfg.KeepCover {
			err = os.Remove(coverPath)
			if err != nil {
				fmt.Println("Failed to delete cover.\n", err)
			}
		}
	}
}
