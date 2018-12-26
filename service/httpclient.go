package service

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type SearchResult struct {
	Title       string
	GroupId     string
	ArtifactId  string
	Url         string
	Description string
}

type ArtifactResult struct {
	GroupId    string
	ArtifactId string
	Version    string
	Url        string
}

type VersionResult struct {
	Maven  string
	Gradle string
	SBT    string
	Ivy    string
}

// "https://mvnrepository.com/search?q=kotlin" 走这个接口的请求
func Search(name string) (searchResults []*SearchResult, err error) {
	searchUrl := fmt.Sprintf("https://mvnrepository.com/search?q=%v", name)
	resp, _ := http.Get(searchUrl)
	defer resp.Body.Close()
	document, err := goquery.NewDocumentFromReader(resp.Body)
	searchResults, err = FilterSearchBody(document)
	return
}

func FilterSearchBody(doc *goquery.Document) (searchResults []*SearchResult, err error) {
	searchResults = make([]*SearchResult, 0)
	doc.Find("#maincontent div.im").Each(func(i int, s *goquery.Selection) {
		//获取到title
		find := s.Find("div.im-header > h2.im-title >a:nth-child(2)")
		title := find.Text()
		if len(title) > 0 {
			groupId := s.Find("div.im-header > p > a:nth-child(1)").Text()
			artifactId := s.Find("div.im-header > p > a:nth-child(2)").Text()
			h, _ := s.Find("div.im-description").Html()
			description := strings.Replace(strings.Split(h, "<div")[0], "\n", "", -1)
			log.Printf("%v", description)
			var url string
			href, exists := find.Attr("href")
			if exists {
				url = fmt.Sprintf("https://mvnrepository.com%v", href)
			}
			result := SearchResult{Title: title, GroupId: groupId, ArtifactId: artifactId, Description: description, Url: url}
			searchResults = append(searchResults, &result)
		}
	})
	return
}

func Artifact(groupId, artifactId string) (artifactResults []*ArtifactResult, err error) {
	artifactUrl := fmt.Sprintf("https://mvnrepository.com/artifact/%v/%v", groupId, artifactId)
	resp, _ := http.Get(artifactUrl)
	defer resp.Body.Close()
	document, err := goquery.NewDocumentFromReader(resp.Body)
	artifactResults, err = FilterArtifactBody(document, groupId, artifactId)
	return
}

func FilterArtifactBody(doc *goquery.Document, groupId, artifactId string) (artifactResults []*ArtifactResult, err error) {
	artifactResults = make([]*ArtifactResult, 0)
	doc.Find("#snippets a.vbtn").Each(func(i int, s *goquery.Selection) {
		version := s.Text()
		log.Printf("%v", version)
		result := &ArtifactResult{GroupId: groupId, ArtifactId: artifactId, Version: version, Url: fmt.Sprintf("https://mvnrepository.com/artifact/%v/%v/%v", groupId, artifactId, version)}
		artifactResults = append(artifactResults, result)
	})
	return
}

func Version(groupId, artifactId, version string) (versionResult *VersionResult, err error) {
	versionUrl := fmt.Sprintf("https://mvnrepository.com/artifact/%v/%v/%v", groupId, artifactId, version)
	resp, err := http.Get(versionUrl)
	defer resp.Body.Close()
	document, err := goquery.NewDocumentFromReader(resp.Body)
	versionResult, _ = FilterVersionBody(document)
	return
}

func FilterVersionBody(doc *goquery.Document) (versionResult *VersionResult, err error) {
	maven := dropComment(doc.Find("textarea#maven-a").Text())
	gradle := dropComment(doc.Find("textarea#gradle-a").Text())
	sbt := dropComment(doc.Find("textarea#sbt-a").Text())
	ivy := dropComment(doc.Find("textarea#ivy-a").Text())
	versionResult = &VersionResult{Maven: maven, Gradle: gradle, SBT: sbt, Ivy: ivy}
	return
}

func dropComment(s string) string {
	r, _ := regexp.Compile("https://mvnrepository.com*")
	if r.MatchString(s) {
		return strings.Join(strings.Split(s, "\n")[1:], "\n")
	} else {
		return s
	}

}
