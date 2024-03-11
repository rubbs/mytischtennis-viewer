package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
	log "github.com/sirupsen/logrus"
)

type MyTischtennisParser struct{}

const (
	colMatchName   = 0
	colMatchHome   = 2
	colMatchAway   = 4
	colMatchBalls1 = 6
	colMatchSet    = 10
	colMatchResult = 11
)

type Game struct {
	Date    string
	Home    string
	Away    string
	Result  string
	Matches []Match
	HomePoints int
	AwayPoints int
}

type Match struct {
	Name       string
	Home       string
	Away       string
	Balls      []string
	Set        string
	Result     string
}

func isMatchRow(tr *colly.HTMLElement) bool {
	classes := tr.ChildAttr("td", "class")
	if strings.Contains(classes, "match-name") {
		return true
	}
	if strings.Contains(classes, "players") {
		return true
	}
	return false
}

func Parse(url string) Game {

	log.Info("start mytt crawler")
	c := colly.NewCollector()

	foundRows := 0

	g := Game{
		Date:    "",
		Home:    "",
		Away:    "",
		Result:  "",
		Matches: []Match{},
	}

	c.OnHTML("h5.green", func(h *colly.HTMLElement) {
		splitted := strings.Split(h.Text, ":")
		g.Home = strings.TrimSpace(splitted[0])
		g.Away = strings.TrimSpace(splitted[1])
	})

	c.OnHTML("h6", func(h *colly.HTMLElement) {
		g.Date = strings.TrimSpace(h.Text)
	})

	c.OnHTML("tr.summary", func(h *colly.HTMLElement) {
		res := h.ChildText("td:nth-child(6)")
		if res != "" {
			g.Result = strings.TrimSpace(res)
			g.HomePoints,_ = strconv.Atoi(strings.Split(g.Result, ":")[0])
			g.AwayPoints,_ = strconv.Atoi(strings.Split(g.Result, ":")[1])
		}
	})

	c.OnHTML("tr", func(tr *colly.HTMLElement) {
		if isMatchRow(tr) {
			foundRows++
			if foundRows >= 16 {
				return
			}
			m := Match{}
			m.Balls = parseBalls(tr)
			tr.ForEach("td", func(tdi int, td *colly.HTMLElement) {

				log.Debugf("%d, %v", tdi, td)
				switch tdi {
				case colMatchName:
					m.Name = td.Text
				case colMatchHome:
					m.Home = parsePlayer(td)
				case colMatchAway:
					m.Away = parsePlayer(td)
				case colMatchSet:
					m.Set = td.Text
				case colMatchResult:
					m.Result = td.Text
				}

			})
			g.Matches = append(g.Matches, m)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		log.Infof("Visiting %v", r)
	})

	c.Visit(url)
	log.Infof("found rows: %d", foundRows)
	return g
}

func parsePlayer(el *colly.HTMLElement) string {
	splitted := strings.Split(el.Text, "/")

	if len(splitted) > 1 {
		// double
		return strings.TrimSpace(splitted[0]) + " / " + strings.TrimSpace(splitted[1])
	} else {
		// single
		return strings.TrimSpace(splitted[0])
	}
}

func parseBalls(el *colly.HTMLElement) []string {

	balls := make([]string, 5)
	for i := 0; i < 5; i++ {
		balls[i] = strings.TrimSpace(el.ChildText(fmt.Sprintf("td:nth-child(%d)", i + colMatchBalls1)))
	}
	return balls

}
