package workflow

import (
	"fmt"
	"github.com/deanishe/awgo"
	"github.com/wangkezun/mvnrepo-go/service"
	"log"
	"strings"
)

// Workflow is the main API
var WF *aw.Workflow
var titleCache *aw.Cache

func init() {
	// Create a new Workflow using default settings.
	// Critical settings are provided by Alfred via environment variables,
	// so this *will* die in flames if not run in an Alfred-like environment.
	WF = aw.New()
	titleCache = aw.NewCache("titleCache")
}

//查询的总入口
func Query() {
	//只有一个参数时，无法判断这个参数究竟是artifactid还是什么，所以是调用"https://mvnrepository.com/search?q=kotlin"url来获取匹配数据
	// ，先不用artifact。因为一次404 一次200这种很烦
	args := WF.Args()
	if len(args) == 1 {
		log.Printf("into default search,arg:%v", args[0])
		searchResults, err := service.Search(args[0])
		if err != nil {
			WF.FatalError(err)
		}
		BuildSearchResultItem(searchResults)
	} else if len(args) == 2 {
		log.Printf("into artifact search,groupId:%v,artifactId:%v", args[0], args[1])

		artifactResults, err := service.Artifact(args[0], args[1])
		if err != nil {
			WF.FatalError(err)
		}
		BuildArtifactResultItem(artifactResults)
	}

	WF.SendFeedback()
}

func BuildSearchResultItem(results []*service.SearchResult) {
	for _, result := range results {
		makeName := fmt.Sprintf("%v >> %v", result.GroupId, result.ArtifactId)
		arg := fmt.Sprintf("%v %v", result.GroupId, result.ArtifactId)

		item := WF.NewItem(result.Title).Subtitle(makeName).Arg(arg).Valid(true).Icon(&aw.Icon{Value: fmt.Sprintf("icons/%v.png", strings.ToUpper(string(result.ArtifactId[0])))})
		titleCache.StoreJSON(makeName, result.Title)
		modifier := item.NewModifier(aw.ModCmd).Subtitle(result.Description)
		modifier.Arg(result.Url).Valid(true)
	}
}

func BuildArtifactResultItem(results []*service.ArtifactResult) {
	for _, result := range results {
		//因为title
		makeName := fmt.Sprintf("%v %v", result.GroupId, result.ArtifactId)
		var title string
		_ = titleCache.LoadJSON(makeName, title)
		makeArg := fmt.Sprintf("%v %v %v", result.GroupId, result.ArtifactId, result.Version)
		item := WF.NewItem(result.Version).Subtitle(makeName).Arg(makeArg)
		//这里其实可以用modifier做复制粘贴了，
		item.NewModifier(aw.ModCmd).Subtitle("Open in browser").Arg(result.Url).Valid(true)
		item.NewModifier(aw.ModCtrl).Subtitle("Copy as maven format, scope may wrong").Arg(BuildMavenArg(result)).Valid(true)
		item.NewModifier(aw.ModShift).Subtitle("Copy as gradle format, configurations may wrong").Arg(BuildGradleArg(result)).Valid(true)
	}
}

func BuildMavenArg(result *service.ArtifactResult) string {
	return fmt.Sprintf(`<dependency>
    <groupId>%v</groupId>
    <artifactId>%v</artifactId>
    <version>%v</version>
</dependency>`, result.GroupId, result.ArtifactId, result.Version)
}
func BuildGradleArg(result *service.ArtifactResult) string {
	return fmt.Sprintf("compile group: '%v', name: '%v', version: '%v'", result.GroupId, result.ArtifactId, result.Version)
}
