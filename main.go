package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	url := os.Args[1]
	fmt.Printf("Snatching %s\n", url)
	var title string

	if err := snatch(url, &title); err != nil {
		log.Fatal(err)
	}

	title = fmt.Sprintf("[%s](%s)", title, url)
	fmt.Printf("\t%s\n", title)

	if err := copyToClip(title); err != nil {
		log.Fatal(err)
	}
}

func copyToClip(text string) error {
	var command string

	switch runtime.GOOS {
	case "darwin":
		command = "pbcopy"
	default:
		return fmt.Errorf("OS '%s' not supported for copy to clipboard", runtime.GOOS)
	}

	if found, err := exec.LookPath(command); err != nil {
		return err
	} else {
		command = found
	}

	if _, err := exec.Command("bash", "-c", fmt.Sprintf("echo -n '%s' | %s", text, command)).Output(); err != nil {
		return fmt.Errorf("copy to clip board failed: %w", err)
	}

	fmt.Println("Copied to clipboard")

	return nil
}

func getAnchor(doc *goquery.Document, fragment string, text *string) (bool, error) {
	if fragment != "" {
		id := "#" + fragment
		doc.Find(id).EachWithBreak(func(i int, selection *goquery.Selection) bool {
			*text = selection.Text()
			return true
		})
	}

	return *text != "", nil
}

func getTitle(doc *goquery.Document, text *string) (bool, error) {
	doc.Find("title").EachWithBreak(func(i int, selection *goquery.Selection) bool {
		*text = selection.Text()
		return true
	})

	return *text != "", nil
}

func snatch(link string, title *string) error {
	parsed, err := url.Parse(link)
	if err != nil {
		return err
	}

	doc, err := getPage(link)
	if err != nil {
		return err
	}

	if ok, err := getTitle(doc, title); err != nil {
		return err
	} else if !ok {
		*title = parsed.Host
	}

	var sectionTitle string
	if ok, err := getAnchor(doc, parsed.Fragment, &sectionTitle); err != nil {
		return err
	} else if ok {
		*title = fmt.Sprintf("%s : %s", *title, sectionTitle)
	}

	return nil
}

func getPage(url string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Golang_Canonical_Model/1.0")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	bs, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	bss := string(bs)
	reader := strings.NewReader(bss)
	doc, err := goquery.NewDocumentFromReader(reader)

	return doc, err
}
