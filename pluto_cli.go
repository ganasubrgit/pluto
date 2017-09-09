package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/dustin/go-humanize"

	"github.com/ishanjain28/pluto/pluto"
)

func main() {

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sig
		fmt.Printf("Interrupt Detected, Shutting Down.")
		os.Exit(0)
	}()

	parts := flag.Int("part", 32, "Number of Download parts")
	verbose := flag.Bool("verbose", false, "Enable Verbose Mode")
	name := flag.String("name", "", "Path or Name of save File")

	flag.Parse()
	urls := []string{}

	for i, v := range os.Args {
		if i == 0 || strings.Contains(v, "-name=") || strings.Contains(v, "-part=") || strings.Contains(v, "-verbose") {
			continue
		}

		urls = append(urls, v)
	}

	if len(urls) == 0 {
		u := ""
		fmt.Printf("URL: ")
		fmt.Scanf("%s\n", &u)
		if u == "" {
			log.Fatalln("No URL Provided")
		}
		urls = append(urls, u)
	}
	for _, v := range urls {
		download(v, *name, *parts, *verbose)
	}
}

func download(u, filename string, parts int, verbose bool) {

	// This variable is used to check if an error occurred anywhere in program
	// If an error occurs, Then it'll not exit.
	// And if no error occurs, Then it'll exit after 10 seconds
	var errored bool

	defer func() {
		if errored {
			select {}
		} else {
			time.Sleep(10 * time.Second)
		}
	}()

	up, err := url.Parse(u)
	if err != nil {
		errored = true
		log.Println("Invalid URL")
		return
	}

	fname := strings.Split(filepath.Base(up.String()), "?")[0]
	fmt.Printf("Starting Download with %d connections\n", parts)

	fmt.Printf("\nDownloading %s\n", up.String())

	meta, err := pluto.FetchMeta(up)
	if err != nil {
		errored = true
		log.Printf("error in fetching information about url: %v", err)
		return
	}

	meta.Name = filename

	if meta.Name == "" {
		meta.Name = fname
	}

	saveFile, err := os.Create(meta.Name)
	if err != nil {
		errored = true
		log.Printf("error in creating save file: %v", err)
		return
	}

	config := &pluto.Config{
		Meta:       meta,
		Parts:      parts,
		RetryCount: 10,
		Verbose:    verbose,
		Writer:     saveFile,
	}

	a := time.Now()
	err = pluto.Download(config)
	if err != nil {
		errored = true
		log.Printf("%v", err)
		return
	}
	timeTaken := time.Since(a)
	fmt.Printf("Downloaded complete in %s. Avg. Speed - %s/s\n", timeTaken, humanize.IBytes(uint64(meta.Size)/uint64(timeTaken.Seconds())))
}
