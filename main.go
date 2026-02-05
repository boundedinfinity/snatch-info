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
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	urls := os.Args[1:]

	for _, url := range urls {
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
}

func copyToClip(text string) error {
	var shell string

	if found, err := exec.LookPath("bash"); err != nil {
		return err
	} else {
		shell = found
	}

	var command string

	switch runtime.GOOS {
	case "darwin":
		command = "pbcopy"
	default:
		return fmt.Errorf("OS clipboard program '%s' not found", runtime.GOOS)
	}

	if found, err := exec.LookPath(command); err != nil {
		return err
	} else {
		command = found
	}

	args := []string{"-c", fmt.Sprintf("echo -n '%s' | %s", text, command)}

	if output, err := exec.Command(shell, args...).Output(); err != nil {
		soutput := string(output)

		if soutput == "" {
			soutput = "<no command output>"
		}

		return fmt.Errorf("copy to clip board failed: %s : %w", soutput, err)
	}

	fmt.Println("Copied to clipboard")

	return nil
}

func getAnchor(doc *goquery.Document, link *url.URL, text *string) (bool, error) {
	if link != nil && link.Fragment != "" {
		id := "#" + link.Fragment
		doc.Find(id).EachWithBreak(func(i int, selection *goquery.Selection) bool {
			*text = selection.Text()
			return true
		})
	}

	return *text != "", nil
}

func getTitle(doc *goquery.Document, link *url.URL, text *string) (bool, error) {
	working := *text

	doc.Find("title").EachWithBreak(func(i int, selection *goquery.Selection) bool {
		working = selection.Text()
		return true
	})

	working = strings.TrimSpace(working)

	if hostname, ok := geHostname(link); ok {
		ltitle := strings.ToLower(working)
		hostname = strings.ToLower(hostname)
		i := strings.Index(ltitle, hostname)

		if i >= 0 {
			working = working[:i]
			working = strings.Trim(working, " -")
		}

		working = fmt.Sprintf("%s : %s", hostname, working)
	}

	*text = working
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

	if ok, err := getTitle(doc, parsed, title); err != nil {
		return err
	} else if !ok {
		*title = parsed.Host
	}

	var sectionTitle string
	if ok, err := getAnchor(doc, parsed, &sectionTitle); err != nil {
		return err
	} else if ok {
		*title = fmt.Sprintf("%s : %s", *title, sectionTitle)
	}

	return nil
}

func geHostname(link *url.URL) (string, bool) {
	var hostname string
	parts := strings.Split(link.Hostname(), ".")

	if len(parts) > 1 {
		slices.Reverse(parts)
		hostname = parts[1]

	}

	return hostname, hostname != ""
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
