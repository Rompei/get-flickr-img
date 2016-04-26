package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/Rompei/imdl"
	simplejson "github.com/bitly/go-simplejson"
	fk "github.com/mncaudill/go-flickr"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"
)

const URLTemplate = "https://farm%s.staticflickr.com/%s/%s_%s.jpg"

func main() {
	output := flag.String("output", "output", "Output directory")
	imgNum := flag.Int("imgnum", 100, "The number of images.")
	isList := flag.Bool("list", false, "Whether output image list or not.")
	x := flag.Uint("x", 213, "Width")
	y := flag.Uint("y", 213, "Height")
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU() - 1)
	if *imgNum < 0 || *imgNum > 500 {
		fmt.Printf("imgNum is too many %d\n", *imgNum)
		os.Exit(1)
	}
	queries, err := getQueries()
	if err != nil {
		panic(err)
	}
	urls, err := getQueryURLs(queries, *imgNum)
	if err != nil {
		panic(err)
	}
	finishCh := make(chan bool, 1)
	saveImage(urls, *output, *isList, *x, *y, finishCh)
	<-finishCh
}

func getQueries() ([]string, error) {
	var queries []string
	fp := os.Stdin
	reader := bufio.NewReaderSize(fp, 4096)
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		queries = append(queries, string(line))
	}
	return queries, nil
}

func getQueryURLs(queries []string, imgNum int) ([]string, error) {
	var urls []string
	flickr := &fk.Request{
		ApiKey: os.Getenv("FLICKR_API_KEY"),
		Method: "flickr.photos.search",
		Args: map[string]string{
			"sort":           "relevance",
			"privacy_filter": "1",
			"content_type":   "1",
			"per_page":       strconv.Itoa(imgNum),
			"format":         "json",
			"nojsoncallback": "1",
		},
	}
	for _, q := range queries {
		part, err := getImageURLs(flickr, 1, q)
		if err != nil {
			return nil, err
		}
		urls = append(urls, part...)
		fmt.Fprintf(os.Stderr, "Collecting image urls for %s\n", q)
		time.Sleep(time.Second)
	}
	return urls, nil
}

func getImageURLs(req *fk.Request, page int, query string) ([]string, error) {

	req.Args["page"] = strconv.Itoa(page)
	req.Args["text"] = query

	resp, err := req.Execute()
	if err != nil {
		return nil, err
	}

	js, err := simplejson.NewJson(resp)
	if err != nil {
		return nil, err
	}
	photos := js.Get("photos").Get("photo").MustArray()
	urls := make([]string, len(photos))

	for i := range photos {
		photo := js.Get("photos").Get("photo").GetIndex(i).MustMap()
		urls[i] = fmt.Sprintf(URLTemplate, photo["farm"], photo["server"], photo["id"], photo["secret"])
	}

	return urls, nil
}

func saveImage(urls []string, dirName string, isList bool, x, y uint, finishCh chan bool) {
	fnameCh := make(chan string, runtime.NumCPU())
	errCh := make(chan error, 1)
	totalImage := len(urls)
	go imdlDaemon(fnameCh, errCh, finishCh, dirName, isList, totalImage)
	var m sync.Mutex
	for i := 0; i < totalImage; i += 10 {
		end := i + 10
		if end > totalImage {
			end = totalImage
		}
		for _, u := range urls[i:end] {
			go imdl.DownloadToPath(u, dirName, fnameCh, errCh, x, y, true, &m)
		}
		time.Sleep(time.Second)
	}
}

func imdlDaemon(fnameCh chan string, errCh chan error, finishCh chan bool, dirName string, isList bool, totalImage int) {
	imgCount := 0
	var fp *os.File
	var err error
	if isList {
		fp, err = os.Create("img_list.txt")
		if err != nil {
			panic(err)
			return
		}
		defer fp.Close()
	}
L:
	for {
		select {
		case fname := <-fnameCh:
			imgCount++
			fmt.Fprintf(os.Stderr, "\r%.1f%%, Downloaded %s", (float64(imgCount)/float64(totalImage))*100.0, fname)
			if isList {
				path, err := buildPath(dirName + "/" + fname)
				if err != nil {
					panic(err)
				}
				fp.WriteString(path + "\n")
			}
			if imgCount == totalImage {
				break L
			}
		case err := <-errCh:
			panic(err)
		default:
		}
	}
	fmt.Fprintln(os.Stderr, "\r100%")
	close(fnameCh)
	close(errCh)
	finishCh <- true
}

func buildPath(path string) (res string, err error) {
	res = filepath.Clean(path)
	if filepath.IsAbs(res) {
		return
	}
	p, err := os.Getwd()
	if err != nil {
		return "", err
	}
	res = p + "/" + path
	return
}
