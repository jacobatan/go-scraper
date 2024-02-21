package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/gocolly/colly"
)

type item struct {
	ProjectName string   `json:"projectName"`
	URL         string   `json:"url"`
	TeamMembers []string `json:"teamMembers,omitempty"`
}

func main() {
	c := colly.NewCollector(
		colly.AllowedDomains("part4project.foe.auckland.ac.nz"),
	)

	var items []item
	var mu sync.Mutex

	c.OnHTML("a[name=Project_title]", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		absoluteURL := e.Request.AbsoluteURL(href)
		name := e.Text
		mu.Lock()
		items = append(items, item{
			ProjectName: name,
			URL:         absoluteURL,
		})
		mu.Unlock()
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	c.Visit("https://part4project.foe.auckland.ac.nz/home/projects/ece/2024")

	c.OnHTML("h4:contains('Team') + ul", func(e *colly.HTMLElement) {
		var members []string
		e.ForEach("li", func(_ int, el *colly.HTMLElement) {
			name := el.Text
			members = append(members, name)
		})
		currentURL := e.Request.URL.String()
		mu.Lock()
		for i, item := range items {
			if item.URL == currentURL {
				items[i].TeamMembers = members
				break
			}
		}
		mu.Unlock()
	})

	for _, item := range items {
		fmt.Println("Visiting", item.URL)
		fmt.Println("Project: ", item.ProjectName)
		err := c.Visit(item.URL)
		if err != nil {
			fmt.Println("Error visiting:", item.URL, err)
		}
	}

	toJSON(items)
	toCSV(items)

}

func toJSON(items []item) {
	jsonData, err := json.Marshal(items)
	if err != nil {
		fmt.Println("Error encoding JSON", err)
		return
	}

	err = os.WriteFile("output.json", jsonData, 0644)

}

func toCSV(items []item) {
	file, err := os.Create("output.csv")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"Name", "URL", "Team Members"}
	if err := writer.Write(header); err != nil {
		fmt.Println("Error writing header to CSV:", err)
		return
	}

	for _, item := range items {
		teamMembers := strings.Join(item.TeamMembers, "; ")
		record := []string{item.ProjectName, item.URL, teamMembers}

		if err := writer.Write(record); err != nil {
			fmt.Println("Error writing record to CSV:", err)
			return
		}
	}
}
