package client

import (
	"encoding/json"
	"net/url"
	"strings"
)

func (c *Client) FetchRules() (_ *RulesResponse, err error) {
	req, err := c.newRequest("GET", "/api/rules", nil)
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
	var data RulesResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	return &data, err
}

func (c *Client) SaveRules(skip, priority string, versioning ...string) (err error) {
	versioningRules := ""
	if len(versioning) > 0 {
		versioningRules = versioning[0]
	}
	formData := url.Values{"skip": {skip}, "priority": {priority}, "versioning": {versioningRules}}
	req, err := c.newRequest("POST", "/api/rules", strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer closeBody(resp, &err)
	return checkStatus(resp)
}

func (c *Client) AddAlias(keyword, value string) (err error) {
	formData := url.Values{"alias-keyword": {keyword}, "alias-value": {value}}
	req, err := c.newRequest("POST", "/api/add_alias", strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer closeBody(resp, &err)
	return checkStatus(resp)
}

func (c *Client) DeleteAlias(alias string) (err error) {
	formData := url.Values{"alias": {alias}}
	req, err := c.newRequest("POST", "/api/delete_alias", strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer closeBody(resp, &err)
	return checkStatus(resp)
}
