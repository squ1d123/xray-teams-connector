package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/deckarep/golang-set"
	"github.com/gorilla/mux"
	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type XrayMessage struct {
	Created     time.Time `json:"created"`
	TopSeverity string    `json:"top_severity"`
	WatchName   string    `json:"watch_name"`
	PolicyName  string    `json:"policy_name"`
	Issues      []struct {
		Severity          string    `json:"severity"`
		Type              string    `json:"type"`
		Provider          string    `json:"provider"`
		Created           time.Time `json:"created"`
		Summary           string    `json:"summary"`
		Description       string    `json:"description"`
		ImpactedArtifacts []struct {
			Name          string `json:"name"`
			DisplayName   string `json:"display_name"`
			Path          string `json:"path"`
			PkgType       string `json:"pkg_type"`
			Sha256        string `json:"sha256"`
			Sha1          string `json:"sha1"`
			Depth         int    `json:"depth"`
			ParentSha     string `json:"parent_sha"`
			InfectedFiles []struct {
				Name        string `json:"name"`
				Path        string `json:"path"`
				Sha256      string `json:"sha256"`
				Depth       int    `json:"depth"`
				ParentSha   string `json:"parent_sha"`
				DisplayName string `json:"display_name"`
				PkgType     string `json:"pkg_type"`
			} `json:"infected_files"`
		} `json:"impacted_artifacts"`
		Cve string `json:"cve"`
	} `json:"issues"`
}

type TeamsMessage struct {
	Context         string            `json:"@context"`
	Type            string            `json:"@type"`
	ThemeColor      string            `json:"themeColor"`
	Title           string            `json:"title"`
	Text            string            `json:"text"`
	Sections        []Section         `json:"sections"`
	PotentialAction []PotentialAction `json:"potentialAction"`
}

type Section struct {
	Title      string `json:"title"`
	StartGroup bool   `json:"startGroup"`
	Text       string `json:"text"`
	// Facts      []Fact `json:"facts"`
}

// type Fact struct {
//     Name  string `json:"name"`
//     Value string `json:"value"`
// }

type PotentialAction struct {
	Type    string   `json:"@type"`
	Name    string   `json:"name"`
	Targets []Target `json:"targets"`
}

type Target struct {
	Os  string `json:"os"`
	URI string `json:"uri"`
}

type Artifact struct {
	DisplayName string `json:"display_name"`
	PkgType     string `json:"pkg_type"`
}

// Transform message and return
func TransformXrayMessage(xrayMessage *XrayMessage) TeamsMessage {
	var sections []Section
	artifactsSet := mapset.NewSet()
	// Iterate through and add artifacts to the set
	for _, issue := range xrayMessage.Issues {
		for _, artifact := range issue.ImpactedArtifacts {
			artifactsSet.Add(Artifact{
				DisplayName: artifact.DisplayName,
				PkgType:     artifact.PkgType,
			},
			)
		}
	}

	var impactedArtifacts string
	it := artifactsSet.Iterator()
	for artifact := range it.C {
		// replace the display names to form the url
		replacedName := strings.Replace(artifact.(Artifact).DisplayName, "/", "~2F", -1)
		replacedName = strings.Replace(replacedName, ":", "/", -1)

		impactedArtifacts += fmt.Sprintf(
			"- [%s](%s/web/#/component/details/%s:~2F~2F%s)\n",
			artifact.(Artifact).DisplayName,
			opts.XrayUrl,
			strings.ToLower(artifact.(Artifact).PkgType),
			replacedName)
	}

	sections = append(sections, Section{
		Title:      "Affected Artifacts",
		StartGroup: true,
		Text:       impactedArtifacts,
	})

	return TeamsMessage{
		Context:    "https://schema.org/extensions",
		Type:       "MessageCard",
		ThemeColor: "FF0000",
		Title:      fmt.Sprintf("Xray Violation: Policy - %s", xrayMessage.PolicyName),
		Text:       "Click **Go to Xray** to learn more about the vulnerabilities",
		Sections:   sections,
		PotentialAction: []PotentialAction{
			PotentialAction{
				Type: "OpenUri",
				Name: "Go to Xray",
				Targets: []Target{
					Target{
						Os:  "default",
						URI: opts.XrayUrl,
					},
				},
			},
		},
	}
}

// create a new item
func SendXrayMessage(w http.ResponseWriter, r *http.Request) {
	// params := mux.Vars(r)
	var xrayMessage XrayMessage
	err := json.NewDecoder(r.Body).Decode(&xrayMessage)
	if err != nil {
		log.Println(err)
	}

	if len(xrayMessage.PolicyName) == 0 {
		log.Error("Could not parse post body")
	}

	keys, ok := r.URL.Query()["key"]

	if !ok || len(keys[0]) < 1 {
		log.Error("Url Param 'key' is missing")
		return
	}

	if len(xrayMessage.Issues) < 1 {
		log.Debug("There are no Issues so not sending a message")
		return
	}

	webHookUrl := fmt.Sprintf("https://outlook.office.com/webhook/%s", keys[0])
	teamsMessage := TransformXrayMessage(&xrayMessage)

	requestByte, _ := json.Marshal(teamsMessage)
	requestReader := bytes.NewReader(requestByte)
	log.Printf("Sending Post to %s\n", webHookUrl)
	res, err := http.Post(webHookUrl, "application/json", requestReader)

	if err != nil {
		log.Fatal(err)
	}

	rb, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Error(err)
	}
	log.Infof("Microsoft Teams response text: %s", string(rb))
	if res.StatusCode != http.StatusOK {
		if err != nil {
			log.Errorf("Failed reading Teams http response: %v", err)
			// fmt.Errorf("Failed reading Teams http response: %v", err)
		}
		// fmt.Errorf("Failed sending to the Teams Channel. Teams http response: %s",
		log.Errorf("Failed sending to the Teams Channel. Teams http response: %s",
			res.Status)
	}
	if err := res.Body.Close(); err != nil {
		log.Error(err)
	}
	fmt.Fprintf(w, string(rb))
	res.Body.Close()
}

func setLogLevel(l string) {
	switch l {
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	case "FATAL":
		log.SetLevel(log.FatalLevel)
	case "PANIC":
		log.SetLevel(log.PanicLevel)
	default:
		log.Fatal("Error: Invalid log-level")
	}
}

func init() {
	_, err := flags.Parse(&opts)
	if err != nil {
		panic(err)
	}
	setLogLevel(opts.LogLevel)
}

//Flag options
var opts struct {
	XrayUrl  string `short:"x" long:"xrayUrl" description:"The Url to where xray is located" required:"true"`
	LogLevel string `short:"l" long:"logLevel" default:"INFO" description:"The log level"`
}

// main function to boot up everything
func main() {
	log.Info("Starting up REST interface on port 8000")
	router := mux.NewRouter()
	router.HandleFunc("/webhook", SendXrayMessage).Methods("POST")
	log.Fatal(http.ListenAndServe(":8000", router))
}
