package client

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (c *Client) FetchHistory() (_ []HistoryItem, err error) {
	req, err := c.newRequest(http.MethodGet, "/api/history?opened=true", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp, &err)
	if err := checkStatus(resp); err != nil {
		return nil, err
	}
	var response struct {
		Documents []HistoryItem `json:"documents"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	return response.Documents, err
}

func (c *Client) PostHistory(query, urlStr, title string) (err error) {
	body := historyRequest{URL: urlStr, Title: title, Query: query}
	data, _ := json.Marshal(body)
	req, err := c.newRequest("POST", "/api/history", strings.NewReader(string(data)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer closeBody(resp, &err)
	return checkStatus(resp)
}

func (c *Client) DeleteHistoryEntry(query, urlStr string) (err error) {
	body := historyRequest{URL: urlStr, Query: query, Delete: true}
	data, _ := json.Marshal(body)
	req, err := c.newRequest("POST", "/api/history", strings.NewReader(string(data)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer closeBody(resp, &err)
	return checkStatus(resp)
}
