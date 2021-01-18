package scanner

import (
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type worker struct {
	id      int
	client  *http.Client
	scanner *Scanner
}

func (w *worker) process(stop <-chan struct{}) {
	for {
		select {
		case <-stop:
			log.Printf("worker %d finished\n", w.id)
			return
		default:
			if page := w.scanner.queue.Dequeue(); page != nil {
				w.scan(page)
			}
		}
	}
}

func (w *worker) scan(p *page) {
	u, err := url.Parse(p.url)
	if err != nil {
		log.Printf("failed to parse url: %s\n", p.url)
		return
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		log.Printf("scheme %s not supported, skipping: %s\n", u.Scheme, p.url)
		return
	}

	log.Printf("worker %d, start scanning: %+v\n", w.id, p.url)

	var resp *http.Response
	for i := 0; i < w.scanner.maxRetries; i++ {
		req, _ := http.NewRequest(http.MethodGet, p.url, nil)
		req.Close = true
		resp, err = w.client.Do(req)
		if err != nil {
			sleep := time.Second
			log.Printf("failed to fetch: %s, retry after %s, err: %s\n", p.url, sleep, err)
			time.Sleep(sleep)
			continue
		}
	}
	if err != nil {
		log.Printf("broken link found: %s -> %s, err: %s\n", p.previous, p.url, err)
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		log.Printf("broken link found: %s -> %s, status code: %d, %s\n", p.previous, p.url, resp.StatusCode, http.StatusText(resp.StatusCode))
		return
	}
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(u.String(), w.scanner.root.String()) || !strings.Contains(contentType, "html") {
		log.Printf("skipping parse content: %s, %s\n", contentType, p.url)
		return
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("failed to parse content: %s\n", p.url)
		return
	}
	location := resp.Request.URL.String()
	doc.Find("a").Each(func(i int, selection *goquery.Selection) {
		if href, ok := selection.Attr("href"); ok {
			newPage := &page{previous: p.url, url: absURL(href, location)}
			w.scanner.queue.Enqueue(newPage)
		}
	})
	doc.Find("link").Each(func(i int, selection *goquery.Selection) {
		if href, ok := selection.Attr("href"); ok {
			newPage := &page{previous: p.url, url: absURL(href, location)}
			w.scanner.queue.Enqueue(newPage)
		}
	})
	doc.Find("script").Each(func(i int, selection *goquery.Selection) {
		if href, ok := selection.Attr("src"); ok {
			newPage := &page{previous: p.url, url: absURL(href, location)}
			w.scanner.queue.Enqueue(newPage)
		}
	})
	doc.Find("img").Each(func(i int, selection *goquery.Selection) {
		if href, ok := selection.Attr("src"); ok {
			newPage := &page{previous: p.url, url: absURL(href, location)}
			w.scanner.queue.Enqueue(newPage)
		}
	})
	log.Printf("worker %d, finished scanning: %+v\n", w.id, p.url)
}
