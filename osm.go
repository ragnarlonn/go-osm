package osm

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

/*
{
	"place_id": 128726,
	"licence": "Data © OpenStreetMap contributors, ODbL 1.0. https://osm.org/copyright",
	"osm_type": "node",
	"osm_id": 25929985,
	"boundingbox": [
		"59.1651172",
		"59.4851172",
		"17.9110935",
		"18.2310935"
	],
	"lat": "59.3251172",
	"lon": "18.0710935",
	"display_name": "Stockholm, Stockholms kommun, Stockholms län, Svealand, 111 29, Sverige",
	"class": "place",
	"type": "city",
	"importance": 0.840175301943447,
	"icon": "https://nominatim.openstreetmap.org/images/mapicons/poi_place_city.p.20.png",
	"address": {
		"city": "Stockholm",
		"municipality": "Stockholms kommun",
		"state": "Stockholms län",
		"region": "Svealand",
		"postcode": "111 29",
		"country": "Sverige",
		"country_code": "se"
	}
},
*/

type OSMAddress struct {
	City         string `json:"city"`
	Municipality string `json:"municipality"`
	State        string `json:"state"`
	Region       string `json:"region"`
	Postcode     string `json:"postcode"`
	Country      string `json:"country"`
	CountryCode  string `json:"country_code"`
}

type OSMItem struct {
	Lat        string     `json:"lat"`
	Long       string     `json:"long"`
	Class      string     `json:"class"`
	Type       string     `json:"type"`
	Importance float64    `json:"importance"`
	Address    OSMAddress `json:"address"`
}

type OSMResponse []OSMItem

type Client struct {
	httpClient *http.Client
}

var (
	ErrWrongStatusCode = errors.New("Openstreetmap API returned wrong status code")
	ErrNoMatch         = errors.New("No match for location")
)

func NewClient(c *http.Client) *Client {
	if c == nil {
		c = &http.Client{}
	}
	return &Client{httpClient: c}
}

// Use OpenStreetmap to look up an address and map it to a country
// XXX TODO: also try using the user name, to find ethnicity/country origin of user
func (c *Client) Location2CountryCode(location string) (string, error) {

	//fmt.Printf("      ... using OpenStreetmap to look up \"%s\" : ", s)
	query := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s&format=json&addressdetails=1", location)
	body, resp, err := c.httpGet(query, nil)
	if err != nil {
		//fmt.Printf("Error: %s\n", err.Error())
		return "", err
	}
	if resp.StatusCode != 200 {
		//fmt.Printf("OSM returned status code %d\n", resp.StatusCode)
		return "", ErrWrongStatusCode
	}
	data := OSMResponse{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		//fmt.Printf("LookupLocation(): failed to parse JSON data: %s\n", string(body))
		return "", err
	}
	if len(data) < 1 {
		//fmt.Printf("LookupLocation(): no match for \"%s\"\n", location)
		return "", ErrNoMatch
	}
	cc := data[0].Address.CountryCode
	return cc, nil
}

// HTTPGet returns response body ([]byte), HTTP headers (map[string][]string) and error
func (c *Client) httpGet(url string, basicAuth []string) ([]byte, *http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if len(basicAuth) > 0 {
		req.SetBasicAuth(basicAuth[0], basicAuth[1])
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return []byte{}, resp, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, resp, err
	}
	return body, resp, nil
}
