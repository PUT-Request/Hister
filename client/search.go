// SPDX-License-Identifier: AGPL-3.0-or-later

package client

import (
	"encoding/json"
	"io"
	"net/url"

	"github.com/asciimoo/hister/server/indexer"
)

func (c *Client) Search(q *indexer.Query) (_ *indexer.Results, err error) {
	qJSON, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}
	u := "/search?query=" + url.QueryEscape(string(qJSON))
	req, err := c.newRequest("GET", u, nil)
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var res *indexer.Results
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}
	return res, nil
}
