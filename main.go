package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	igc "github.com/marni/goigc"
)

type webhook struct {
	WebhookURL      string //"webhookURL": "http://remoteUrl:8080/randomWebhookPath",
	MinTriggerValue int    //"minTriggerValue": 2

}
type ticker struct {
	T_latest   int      //"t_latest": <latest added timestamp>,
	T_start    int      //"t_start": <the first timestamp of the added track>, this will be the oldest track recorded
	T_stop     string   //"t_stop": <the last timestamp of the added track>, this might equal to t_latest if there are no more tracks left
	Tracks     []string //"tracks": [<id1>, <id2>, ...],
	Processing float64  //"processing": <time in ms of how long it took to process the request>

}
type igcTrack struct {
	H_date        string  //"H_date": <date from File Header, H-record>,
	Pilot         string  //"pilot": <pilot>,
	Glider        string  //"glider": <glider>,
	Glider_id     string  //"glider_id": <glider_id>,
	Track_length  float64 //"track_length": <calculated total track length>
	Track_src_url string  //"track_src_url": <the original URL used to upload the track, ie. the URL used with POST>
}

type API struct {
	Uptime  time.Time //"uptime": <uptime>
	Info    string    //"info": "Service for IGC tracks."
	Version string    //"version": "v1"

}
type igcFile struct {
	Url string //a valid igc URL
}

type igcDB struct {
	igcs map[string]igcFile
}

func (db *igcDB) add(igc igcFile, id string) {
	for _, file := range db.igcs {
		if igc == file {
			return
		}
	}
	db.igcs[id] = igc
}

func (db igcDB) Count() int {
	return len(db.igcs)
}

func (db igcDB) Get(idWanted string) igcFile {
	for id, file := range db.igcs {
		if idWanted == id {
			return file
		}
	}
	return igcFile{}
}

func (db igcDB) isInDb(fileW igcFile) bool {
	for _, file := range db.igcs {
		if file.Url == fileW.Url {
			return true
		}
	}
	return false
}

func getApi(w http.ResponseWriter, r *http.Request) {
	http.Header.Add(w.Header(), "content-type", "application/json")
	//io.WriteString(w, "Api information :\n")
	api := &API{}
	api.Uptime = time.Now()
	api.Info = "Service for IGC tracks."
	api.Version = "version : v1"
	//fmt.Fprintf(w, "%s\n%s\n%s", api.Uptime, api.Info, api.Version)
	json.NewEncoder(w).Encode(api)

}

func igcHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "POST":
		start := time.Now()
		{
			if pathTrack.MatchString(r.URL.Path) {
				if r.Body == nil {
					http.Error(w, "no JSON body", http.StatusBadRequest)
					return
				}
				var igc igcFile
				
				//TODO check correct igc URL
				err := json.NewDecoder(r.Body).Decode(&igc)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
				}
				if db.isInDb(igc) == true {
					fmt.Fprintln(w, "igc file already in database")
					return
				} else {
					long = len(times)
					timestamp = time.Now().Nanosecond()
					times = append(times, timestamp)
					fmt.Fprintf(w, "URL : %s\n", igc.Url)
					Idstr := "id"
					strValue := fmt.Sprintf("%d", idCount)
					newId := Idstr + strValue
					ids = append(ids, newId)
					idCount += 1
					db.add(igc, newId)
					json.NewEncoder(w).Encode(newId)
					elapsed = time.Since(start).Seconds()
					//send message to webhooks
					t_conv := strconv.Itoa(timestamp)
					e_conv := fmt.Sprintf("%f", elapsed)
					text := "{\"text\": \"Timestamp :" + t_conv + ", new track is " + newId + " (processing time is " + e_conv + ")\"}"
					payload := strings.NewReader(text)
					for _, wh := range whDB {
						client := &http.Client{Timeout: (time.Second * 30)}
						req, err := http.NewRequest("POST", wh.WebhookURL, payload)
						req.Header.Set("Content-Type", "application/json")
						resp, err := client.Do(req)
						if err != nil {
							fmt.Print(err.Error())
						}
						fmt.Println(resp.Status)
					}

				}

			} else {
				http.NotFound(w, r)
			}
		}
	case "GET":
		{
			//GET case
			http.Header.Add(w.Header(), "content-type", "application/json")
			parts := strings.Split(r.URL.Path, "/")
			switch {
			case pathTrack.MatchString(r.URL.Path):
				{
					//deal with the array
					json.NewEncoder(w).Encode(ids)
					//fmt.Fprintln(w, "case track")
				}

			case pathId.MatchString(r.URL.Path):
				{
					//fmt.Fprintln(w, "Information about the id")
					//deal with the id
					var igcWanted igcFile
					rgx, _ := regexp.Compile("^id[0-9]*")
					id := parts[4]
					if rgx.MatchString(id) == true {
						igcWanted = db.Get(id)

						//then encode the igcFile
						url := igcWanted.Url
						track, err := igc.ParseLocation(url)
						if err != nil {
							//fmt.Errorf("Problem reading the track", err)
						}
						igcT := igcTrack{}
						igcT.Glider = track.GliderType
						igcT.Glider_id = track.GliderID
						igcT.Pilot = track.Pilot
						igcT.Track_length = track.Task.Distance()
						igcT.H_date = track.Date.String()
						igcT.Track_src_url = igcWanted.Url
						json.NewEncoder(w).Encode(igcT)
					}
					if rgx.MatchString(id) == false {
						fmt.Fprintln(w, "Use format id0 or id21 for exemple")
					}
				}
			case pathField.MatchString(r.URL.Path):
				{
					//TODO parse field track_lenghtto float64, return the value asked
					//fmt.Fprintln(w, parts)
					var igcW igcFile
					infoWanted := parts[5]
					id := parts[4]
					igcW = db.Get(id)
					url := igcW.Url
					track, err := igc.ParseLocation(url)
					if err != nil {
						//fmt.Errorf("Problem reading the track", err)
					}
					igcT := igcTrack{}
					switch infoWanted {
					case "pilot":
						{
							igcT.Pilot = track.Pilot
							json.NewEncoder(w).Encode(igcT.Pilot)
						}
					case "glider":
						{
							igcT.Glider = track.GliderType
							json.NewEncoder(w).Encode(igcT.Glider)
						}
					case "glider_id":
						{
							igcT.Glider_id = track.GliderID
							json.NewEncoder(w).Encode(igcT.Glider_id)
						}
					case "H_date":
						{
							igcT.H_date = track.Date.String()
							json.NewEncoder(w).Encode(igcT.H_date)
						}
					case "track_src_url":
						{
							igcT.Track_src_url = igcW.Url
							json.NewEncoder(w).Encode(igcW.Url)
						}
					}
				}
			default:
				http.NotFound(w, r)
			}

		}

	default:

		http.Error(w, "not implemented yet", http.StatusNotImplemented)

	}
}

var path, _ = regexp.Compile("/paragliding[/]{1}$")
var pathTrack, _ = regexp.Compile("/paragliding/api/track[/]{1}$")
var pathId, _ = regexp.Compile("/paragliding/api/track/id[0-9]+$")
var pathField, _ = regexp.Compile("/paragliding/api/track/id[0-9]+/(pilot$|glider$|glider_id$|H_date$|track_src_url$)")
var pathwhID, _ = regexp.Compile("/paragliding/api/webhook/new_track/id[0-9]+$")

//var pathTicker, _ = regexp.Compile("/paragliding/api/ticker$")

func router(w http.ResponseWriter, r *http.Request) {
	switch {
	case path.MatchString(r.URL.Path):
		{
			getApi(w, r)
		}
	default:
		http.NotFound(w, r)
	}
}

func tickerHandler(w http.ResponseWriter, r *http.Request) {
	http.Header.Add(w.Header(), "content-type", "application/json")
	ticker := ticker{}
	ticker.T_latest = timestamp
	ticker.Tracks = ids
	ticker.Processing = elapsed
	if times == nil {
		ticker.T_start = 0
	} else {
		ticker.T_start = times[0]
	}

	json.NewEncoder(w).Encode(ticker)

}

func tickerHandlerLatest(w http.ResponseWriter, r *http.Request) {
	if timestamp == 0 {
		fmt.Fprintln(w, "No added timestamp")

	} else {
		fmt.Fprintln(w, timestamp)
	}

}
func webhookHandler(w http.ResponseWriter, r *http.Request) {
	http.Header.Add(w.Header(), "content-type", "application/json")
	parts := strings.Split(r.URL.Path, "/")
	switch r.Method {
	case "POST":
		{
			//fmt.Fprintln(w, "wh")
			var wh webhook
			//TODO check correct wh format
			err := json.NewDecoder(r.Body).Decode(&wh)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			Idstr := "id"
			strValue := fmt.Sprintf("%d", idCountwh)
			newId := Idstr + strValue
			idswh = append(idswh, newId)
			idCountwh += 1
			whDB[newId] = wh
			json.NewEncoder(w).Encode(newId)
		}
	case "GET":
		{
			if pathwhID.MatchString(r.URL.Path) {
				idWant := parts[5]
				for id, file := range whDB {
					if id == idWant {
						json.NewEncoder(w).Encode(file)
					}
				}

			}
		}
	case "DELETE":
		{
			if pathwhID.MatchString(r.URL.Path) {
				idWant := parts[5]
				for id, file := range whDB {
					if id == idWant {
						json.NewEncoder(w).Encode(file)
						delete(whDB, idWant)
					}
				}
			}
		}
	default:
		http.NotFound(w, r)

	}
}

func adminCount(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(db.Count())
}

func adminDel(w http.ResponseWriter, r *http.Request) {
	if r.Method == "DELETE" {
		delCount := 0
		ids = []string{}
		for id, _ := range db.igcs {
			delete(db.igcs, id)
			delCount += 1
		}
		json.NewEncoder(w).Encode(delCount)
	}
}

var db igcDB
var ids []string
var idswh []string
var times []int
var idCount int
var idCountwh int
var timestamp int
var start time.Time
var elapsed float64
var whDB map[string]webhook
var long int

func main() {

	db = igcDB{}
	whDB = map[string]webhook{}
	db.igcs = map[string]igcFile{}
	idCount = 0
	timestamp = 0
	times = nil
	long = 0
	ids = []string{}
	port := os.Getenv("PORT")
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				if len(times) != long {
					text := "{\"text\": \"New track added\"}"
					payload := strings.NewReader(text)
					for _, wh := range whDB {
						client := &http.Client{Timeout: (time.Second * 30)}
						req, err := http.NewRequest("POST", wh.WebhookURL, payload)
						req.Header.Set("Content-Type", "application/json")
						resp, err := client.Do(req)
						if err != nil {
							fmt.Print(err.Error())
						}
						fmt.Println(resp.Status)
					}
					long += 1
				}
			}
		}
	}()
	http.HandleFunc("/", router)
	http.HandleFunc("/paragliding/api", getApi)
	http.HandleFunc("/paragliding/api/track/", igcHandler)
	http.HandleFunc("/paragliding/api/ticker", tickerHandler)
	http.HandleFunc("/paragliding/api/ticker/latest", tickerHandlerLatest)
	http.HandleFunc("/paragliding/api/webhook/new_track/", webhookHandler)
	http.HandleFunc("/admin/api/tracks_count", adminCount)
	http.HandleFunc("/admin/api/tracks", adminDel)
	http.ListenAndServe(":"+port, nil)
}
