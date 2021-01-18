package scanner

import "sync"

type page struct {
	previous string
	url      string
}

type queue struct {
	lock  *sync.Mutex
	pages []*page
	cache *sync.Map
}

func (q *queue) Enqueue(page *page) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if _, ok := q.cache.LoadOrStore(page.url, ""); !ok {
		q.pages = append(q.pages, page)
	}
}

func (q *queue) Dequeue() *page {
	q.lock.Lock()
	defer q.lock.Unlock()
	if len(q.pages) > 0 {
		page := q.pages[0]
		q.pages = q.pages[1:]
		return page
	}
	return nil
}