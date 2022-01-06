package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/dariubs/percent"
	"github.com/jedib0t/go-pretty/v6/table"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type User struct {
	Login     string `json:"login"`
	Repos     int `json:"public_repos"`
	Followers int `json:"followers"`
}

type Repo struct {
	Name      string `json:"name"`
	Fork      int `json:"forks_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// JsonUser struct aggregates a user's data
// Once transformed in main() function.
// All users are then aggregated in a slice.
// The slice is finally converted into JSON format.
type JsonUser struct {
	Login     string `json:"login"`
	Repos     int `json:"public_repos"`
	Followers int `json:"followers"`
	Languages string `json:"languages_distribution"`
	Forks     int `json:"forks_count"`
	Activity  string `json:"activity"`
}

// ProcessFile opens a file t and appends each line of t to lines.
// Each line is a GitHub username.
// The file name is passed to ProcessFile as a command line argument.
func ProcessFile (t *string, lines *[]string) error {
	f, err := os.Open(*t)

	defer f.Close()
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		*lines = append(*lines, scanner.Text())
	}
	return nil
}

// ReturnBody makes a GET request to url and returns its body.
func ReturnBody(url string) []byte{
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := client.Do(request)
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return body
}

// GetUser makes a GET request to GitHub API to fetch a user's data.
// JSON data are then parsed into a User struct.
func (u *User) GetUser(user string) User{
	url := "https://api.github.com/users/" + user
	body := ReturnBody(url)
	if err := json.Unmarshal(body, &u); err != nil {
		fmt.Println("Can not unmarshal JSON")
	}
	return *u
}

// GetRepos makes a GET request to GitHub API to fetch a user's repositories data.
// JSON data are then parsed into a slice of Repos structs.
func GetRepos(user string) []Repo{
	url := "https://api.github.com/users/" + user + "/repos"
	body := ReturnBody(url)
	var repos []Repo
	if err := json.Unmarshal(body, &repos); err != nil {   // Parse []byte to go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}
	return repos
}

// GetLanguages makes a GET request to GitHub API to fetch programming languages in each user's repo.
// JSON data are then parsed into a map[string]int.
func GetLanguages(user string, repo Repo) map[string]int{
	url := "https://api.github.com/repos/" + user + "/" + repo.Name + "/languages"
	body := ReturnBody(url)
	result := make(map[string]int)
	if err := json.Unmarshal(body, &result); err != nil {   // Parse []byte to go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}
	return result
}

// MergeMaps merges the slices representing the proportion in which programming languages are used.
// A slice represents a repo. Proportions are given as integers.
// MergeMaps returns a map with the total utilisation of each language.
// In the returned map, each language is represented once.
func MergeMaps(user string, repos []Repo) map[string]int {
	output := make(map[string]int)
	for _, repo := range repos {
		languages := GetLanguages(user, repo)
		for i, j := range languages {
			if _, ok := output[i]; ok {
				output[i] += j
			} else {
				output[i] = j
			}
		}
	}
	return output
}

// ReturnOther creates a category 'others' if MergeMaps returns a map whose length > 5.
// The category includes all languages which are not in the 5 most-used languages.
// The proportions in which these languages are used are aggregated.
func ReturnOther(s []map[string]float64) float64{
	sum := 0.0
	for _, pair := range s[6:] {
		for _,j := range pair{
			sum += j
		}
	}
	return sum
}

// LanguagesDistribution returns distributions from integers in a map.
// Each integer represents the proportion in which a language is used by a user on GitHub.
// The category 'others' is also represented.
// The distributions are ordered in descending order.
func LanguagesDistribution(languages map[string]int) []map[string]float64{
	distribution := make(map[string]float64)
	sum := 0
	for i, _ := range languages {
		sum += languages[i]
	}
	for language, _ := range languages {
		distribution[language] = math.Floor((percent.PercentOf(languages[language], sum))*100)/100
	}

	keys := make([]string, 0, len(distribution))
	for key := range distribution {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		return distribution[keys[i]] > distribution[keys[j]]
	})

	s := make([]map[string]float64, len(distribution))

	for _, language := range keys {
		m := make(map[string]float64)
		m[language] = distribution[language]
		s = append(s, m)
	}
	orderedSlice := s[len(distribution):]

	return orderedSlice
}

// MergeLanguages returns a string of the languages most used by a user on GitHub.
// The string also includes the proportions (in percentages) in which each language is used.
// The input to MergeLanguages is a map associating each language with its proportion.
// MergeLanguages includes category 'Others' if created by function ReturnOther.
func MergeLanguages(f []map[string]float64) string{
	s := f

	if len(f) > 5 {
		s = f[:5]
		sum := ReturnOther(f)
		l := make(map[string]float64)
		l["Others"] = sum
		s = append(s, l)

	}

	ls := ""
	for _, maps := range s {
		for x, y := range maps {
			ls += x + ":" + strconv.FormatFloat(y, 'f', 2, 64) + "\n"
		}
	}
	ls = strings.TrimSuffix(ls, "\n")
	return ls
}

// ReturnForks returns a string of the total number of forks for a GitHub user.
func ReturnForks(user string, repos []Repo) int {
	repos = GetRepos(user)
	forks := make(map[string]int)
	for _, repo := range repos {
		if repo.Fork > 0 {
			forks[repo.Name] = repo.Fork
		}
	}

	sum := 0
	for _, fork := range forks {
		sum += fork
	}
	return sum
}

// GetActivity returns a map associating floats to each year between 2017 and 2021.
// If a repo is created or updated for the last time in year x,
// The value of key x is incremented by 1.
func GetActivity(repos []Repo) map[int]float64{
	activity := map[int]float64{2021:0, 2020:0, 2019:0, 2018:0, 2017:0, 0:0}
	for _, repo := range repos {
		c := repo.CreatedAt.Year()
		u := repo.UpdatedAt.Year()
		if _, ok := activity[c]; ok {
			activity[c] ++
		} else if _, ok := activity[u]; ok {
			activity[u] ++
		}
		if c < 2017 || u < 2017 {
			activity[0] ++
		}
	}
	return activity
}

// ActivityDistribution returns distributions from floats in a map.
// Each float represents the number of actions (creation or last update) during a year on GitHub.
// The distributions are ordered in descending order.
func ActivityDistribution(activity map[int]float64) []map[int]float64{
	distribution := make(map[int]float64)
	sum := 0.0
	for i, _ := range activity {
		sum += activity[i]
	}
	for year, _ := range activity {
		distribution[year] = math.Floor((percent.PercentOfFloat(activity[year], sum))*1000)/1000
	}

	keys := make([]int, 0, len(distribution))
	for key := range distribution {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] > keys[j]
	})

	s := make([]map[int]float64, len(distribution))

	for _, activity := range keys {
		m := make(map[int]float64)
		m[activity] = distribution[activity]
		s = append(s, m)
	}
	orderedActivity := s[len(distribution):]

	return orderedActivity
}

// MergeActivity returns a string of the activity of a GitHub user for the past 5 years.
// The activity is formatted as percentage.
// The input to MergeLanguages is a map associating each year with the number of actions (creation or last update).
func MergeActivity(actDistribution []map[int]float64) string{
	ls := ""
	s := actDistribution[:len(actDistribution)-1]
	for _, maps := range s {
		for x, y := range maps {
			xStr := strconv.Itoa(x)
			ls += xStr + ":" + strconv.FormatFloat(y, 'f', 2, 64) + "%\n"
		}
	}
	o := actDistribution[len(actDistribution)-1]
	others := o[0]
	if others != 0 {
		othersStr := strconv.FormatFloat(others, 'f', 2, 64)
		ls += "Others:" + othersStr + "%"
	}
	return ls
}

func main() {

	// users list aggregates all users' data once transformed.
	// The list is then converted into a new file users.json.
	var users []JsonUser

	var lines []string

	// Text file given as command line argument.
	files := os.Args[1:]

	for _, f := range files {
		err := ProcessFile(&f, &lines)
		if err != nil {
			log.Printf("file '%s' not found", f)
			continue
		}
	}

	// Report formatted as a table.
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"User", "Repositories", "Followers", "Programming languages", "Forks", "Activity"})

	// For each user in text file:
	for _, user := range lines{

		// Get user's repos
		repos := GetRepos(user)

		// Get user's forks
		forks := ReturnForks(user, repos)

		// Get user's languages distribution
		allLanguages := MergeMaps(user, repos)
		langDistribution := LanguagesDistribution(allLanguages)
		ls := MergeLanguages(langDistribution)

		// Get user's other data (login, followers)
		var u User
		u = u.GetUser(user)

		// Get user's activity distribution
		activity := GetActivity(repos)
		ActivityDistribution := ActivityDistribution(activity)
		activityAll := MergeActivity(ActivityDistribution)

		// New lines are replaced by spaces before converting to JSON
		re := regexp.MustCompile(`\r?\n`)
		inputActivity := re.ReplaceAllString(activityAll, " - ")
		inputActivity = strings.TrimSuffix(inputActivity, " - ")
		inputLanguages := re.ReplaceAllString(ls, " - ")
		inputLanguages = strings.TrimSuffix(inputLanguages, " - ")

		// New rows added to printed table
		t.AppendRow([]interface{}{u.Login, u.Repos, u.Followers, ls, forks, activityAll})
		t.AppendSeparator()

		// Create a new user to be marshalled
		NewUser := JsonUser {
			Login : u.Login,
			Repos : u.Repos,
			Followers : u.Followers,
			Languages : inputLanguages,
			Forks : forks,
			Activity : inputActivity,
		}

		users = append(users, NewUser)

	}

	// Marshal users list with transformed data for each user
	// And create a JSON file users.json with new data
	data := users
	file, _ := json.MarshalIndent(data, "", " ")
	_ = ioutil.WriteFile("users.json", file, 0644)

	// Pretty print report
	t.SetStyle(table.StyleLight)
	t.Render()

}