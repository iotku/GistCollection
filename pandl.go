package main

import (
	"fmt"
	"github.com/mmcdole/gofeed"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
)

// Download videos from panopto.com RSS feed
func main() {
	if len(os.Args) < 3{
		showHelp()
		return
	}
	regex := regexp.MustCompile("[\\\\/:*?\"<>|,]")
	outdir := regex.ReplaceAllString(os.Args[1], "_")

	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(os.Args[2])
	for i, _ := range feed.Items {
		filename := fmt.Sprintf(feed.Items[i].Title + " - "+ feed.Items[i].Published)
		url := feed.Items[i].GUID
		// Clean text for usage in file paths
		filename = regex.ReplaceAllString(filename, "_")
		fmt.Println("Downloading", filename)
		downloadVideo(url, outdir, filename)
	}
	saveFeed(outdir, feed.String()) // save feed into folder for future reference
}

func showHelp() {
	fmt.Println("Usage: pandl \"foldername\" \"RSSURL\"")
}

func saveFeed(outdir, feed string) {
	f, err := os.Create(outdir+"/feed.json")
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	_, err = f.WriteString(feed)
	if err != nil {
		log.Fatalln(err)
	}
}

func downloadVideo(url, outdir, filename string) {
	// Make output folder to place videos
	if _, err := os.Stat(outdir); os.IsNotExist(err) {
		err := os.Mkdir(outdir, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	// If there's data, write it to file (hopefully real video data.)
	filepath := outdir+"/"+filename+".mp4"

	if _, err := os.Stat(filepath); !os.IsNotExist(err)  {
		log.Println(filename, "not downloading, already exists.")
		return
	}

	resp, err := http.Get(url)
	if err != nil {
		log.Println("Http Get Error, URL:", url)
		log.Println(err.Error())
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		log.Println("Filepath: ", filepath, err.Error())
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Println("Copy:", resp.Body, err.Error())
	}
}
