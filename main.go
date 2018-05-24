package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"sort"
	"time"
)

type departure struct {
	Route       string `json:"route"`
	StationName string
	LastStation string `json:"lastStation"`
	Minutes     int    `json:"minutes"`
}

type departures []departure

func (s departures) Len() int {
	return len(s)
}
func (s departures) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s departures) Less(i, j int) bool {
	return s[i].Minutes < s[j].Minutes
}

type update struct {
	key   string
	value departures
}

type aggregator struct {
	Stations   map[string]departures
	Departures departures
}

func (a *aggregator) Update(update update) {
	a.Stations[update.key] = update.value

	departs := departures{}

	for _, value := range a.Stations {
		departs = append(departs, value...)
	}

	sort.Sort(departures(departs))

	a.Departures = departs
}

var (
	stationCodes = [7]string{"1/142", "2/230", "4/420", "A/A38", "E/E01", "J/M23", "R/R27"}
	updates      = make(chan update)
	display      = make(chan bool)
	handler      = make(chan departures)
)

func getDepartures(stationCode string) error {
	for {
		getTime := struct {
			StationName string `json:"stationName"`
			Direction1  struct {
				Times departures `json:"times"`
			} `json:"direction1"`
			Direction2 struct {
				Times departures `json:"times"`
			} `json:"direction2"`
		}{}

		if resp, err := http.Get("https://mtasubwaytime.info/getTime/" + stationCode); err == nil {
			defer resp.Body.Close()
			if body, err := ioutil.ReadAll(resp.Body); err == nil {
				if err := json.Unmarshal(body, &getTime); err == nil {
					departures := departures{}

					for _, departure := range append(getTime.Direction1.Times, getTime.Direction2.Times...) {
						if getTime.StationName != departure.LastStation {
							departure.StationName = getTime.StationName
							departures = append(departures, departure)
						}
					}

					updates <- update{stationCode, departures}
				}
			}
		}

		time.Sleep(15 * time.Second)
	}
}

func main() {
	go func() {
		for _, stationCode := range stationCodes {
			go getDepartures(stationCode)
			time.Sleep(2 * time.Second)
		}
	}()

	go func() {
		index := aggregator{Stations: make(map[string]departures), Departures: departures{}}
		for {
			select {
			case <-display:
				handler <- index.Departures
			case update := <-updates:
				index.Update(update)
			}
		}
	}()

	t := template.Must(template.New("/").Parse(`{{range .}}
    <tr><td><img src='http://subwaytime.mta.info/img/{{.Route}}_sm.png'></td><td>{{.StationName}}</td><td>{{.LastStation}}</td><td align='right'>{{.Minutes}}</td></tr>{{end}}`))

	http.HandleFunc("/departures", func(w http.ResponseWriter, r *http.Request) {
		display <- true
		t.Execute(w, <-handler)
	})
	http.Handle("/", http.FileServer(http.Dir("public")))
	http.ListenAndServe(":8080", nil)
}
