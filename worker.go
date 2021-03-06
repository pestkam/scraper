package scraper

import (
	"net/http"
	"net/url"
)

func (s *Scraper) runWorker() {
	defer func() {
		s.wg.Done()
		<-s.concurrentCounter
	}()
	var httpClient *http.Client
	var currentProxy *proxyItem
	var currentProxyURL *url.URL
	for task := range s.queue {
	page:
		for try := 0; try < s.maxRetry; try++ {
			if len(s.proxyList) > 0 {
				currentProxy = s.getNextProxy()
				httpClient = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(currentProxy.url)}}
				currentProxyURL = currentProxy.url
			} else {
				currentProxy = nil
				httpClient = &http.Client{}
				currentProxyURL = nil
			}
			resp, err := httpClient.Get(task)
			if err != nil {
				errScraper := ScraperError{currentProxy.url, task, err.Error()}
				result := Response{http.Response{}, errScraper, currentProxyURL}
				if currentProxy != nil {
					currentProxy.gotFail()
				}
				s.Results <- result
			} else if resp.StatusCode >= 300 {
				errScraper := ScraperError{currentProxy.url, task, resp.Status}
				result := Response{*resp, errScraper, currentProxyURL}
				if currentProxy != nil {
					currentProxy.gotFail()
				}
				s.Results <- result
			} else {
				result := Response{*resp, nil, currentProxyURL}
				s.Results <- result
				if currentProxy != nil {
					currentProxy.gotSuccess()
				}
				break page
			}
		}
	}
}
