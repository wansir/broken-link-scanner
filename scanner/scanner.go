package scanner

import (
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Scanner struct {
	root       *url.URL
	queue      queue
	workers    []*worker
	maxRetries int
}



func NewScanner(threads int, maxRetries int, timeout time.Duration) *Scanner {
	s := &Scanner{
		queue:      queue{lock: &sync.Mutex{}, pages: make([]*page, 0), cache: &sync.Map{}},
		workers:    make([]*worker, 0),
		maxRetries: maxRetries,
	}
	for i := 0; i < threads; i++ {
		worker := &worker{
			id: len(s.workers),
			client: &http.Client{
				Timeout: timeout,
			},
			scanner: s,
		}
		s.workers = append(s.workers, worker)
	}
	return s
}

func (s *Scanner) Scan(root string, stopChan <-chan struct{}) {
	s.root, _ = url.Parse(root)
	s.queue.Enqueue(&page{url: root})
	for _, worker := range s.workers {
		go worker.process(stopChan)
	}
	<-stopChan
}

func absURL(currURL, baseURL string) string {
	curr, err := url.Parse(currURL)
	if err != nil {
		return ""
	}
	if curr.Scheme != "" {
		return currURL
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	return base.ResolveReference(curr).String()
}
