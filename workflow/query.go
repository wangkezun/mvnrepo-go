package workflow

import (
	"fmt"
	"github.com/deanishe/awgo"
	"github.com/deanishe/awgo/update"
	"github.com/wangkezun/mvnrepo-go/service"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Workflow is the main API
var updateJobName = "checkForUpdate"
var WF *aw.Workflow
var titleCache *aw.Cache

func init() {
	// Create a new Workflow using default settings.
	// Critical settings are provided by Alfred via environment variables,
	// so this *will* die in flames if not run in an Alfred-like environment.
	WF = aw.New(update.GitHub("wangkezun/alfred-mvnrepository-workflow"))
	titleCache = aw.NewCache("titleCache")
}

//查询的总入口
func Query() {
	//只有一个参数时，无法判断这个参数究竟是artifactid还是什么，所以是调用"https://mvnrepository.com/search?q=kotlin"url来获取匹配数据
	// ，先不用artifact。因为一次404 一次200这种很烦
	args := WF.Args()
	if len(args) == 0 {
		//indicates that this is a check update flow
		WF.Configure(aw.TextErrors(true))
		log.Println("Checking for updates...")
		if err := WF.CheckForUpdate(); err != nil {
			WF.FatalError(err)
		}

		// Call self with "check" command if an update is due and a check
		// job isn't already running.
		if WF.UpdateCheckDue() && !WF.IsRunning(updateJobName) {
			log.Println("Running update check in background...")

			cmd := exec.Command(os.Args[0], "-check")
			if err := WF.RunInBackground(updateJobName, cmd); err != nil {
				log.Printf("Error starting update check: %s", err)
			}
		}

		// Only show update status if query is empty.
		if WF.UpdateAvailable() {
			// Turn off UIDs to force this item to the top.
			// If UIDs are enabled, Alfred will apply its "knowledge"
			// to order the results based on your past usage.
			WF.Configure(aw.SuppressUIDs(true))

			// Notify user of update. As this item is invalid (Valid(false)),
			// actioning it expands the query to the Autocomplete value.
			// "workflow:update" triggers the updater Magic Action that
			// is automatically registered when you configure Workflow with
			// an Updater.
			//
			// If executed, the Magic Action downloads the latest version
			// of the workflow and asks Alfred to install it.
			WF.NewItem("Update available!").
				Subtitle("↩ to open release page").
				Arg("https://github.com/wangkezun/alfred-mvnrepository-workflow/releases").
				Valid(true).
				Icon(&aw.Icon{Value: "icons/update.png"})
		}

		WF.WarnEmpty("No matching items", "Try a different query?")
		WF.SendFeedback()
	} else if len(args) == 1 {
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
	} else if len(args) == 3 {
		log.Printf("into version page, groupId:%v,artifactId:%v,version:%v", args[0], args[1], args[2])
		versionResult, err := service.Version(args[0], args[1], args[2])
		if err != nil {
			WF.FatalError(err)
		}
		BuildVersionResultItem(versionResult)
	}

	WF.SendFeedback()
}

func BuildSearchResultItem(results []*service.SearchResult) {
	if len(results) > 0 {
		for _, result := range results {
			makeName := fmt.Sprintf("%v >> %v", result.GroupId, result.ArtifactId)
			arg := fmt.Sprintf("%v %v", result.GroupId, result.ArtifactId)
			item := WF.NewItem(result.Title).Subtitle(makeName).Autocomplete(arg).Valid(true).Icon(&aw.Icon{Value: fmt.Sprintf("icons/%v.png", strings.ToUpper(string(result.ArtifactId[0])))})
			titleCache.StoreJSON(makeName, result.Title)
			modifier := item.NewModifier(aw.ModCmd).Subtitle(result.Description)
			modifier.Arg(result.Url).Valid(true)
		}
	} else {
		WF.NewWarningItem("NOT FOUND", "Please check your keyword").Icon(&aw.Icon{Value: "icons/404.png"})
	}
}

func BuildArtifactResultItem(results []*service.ArtifactResult) {
	for _, result := range results {
		//因为title
		makeName := fmt.Sprintf("%v %v", result.GroupId, result.ArtifactId)
		var title string
		_ = titleCache.LoadJSON(makeName, title)
		arg := fmt.Sprintf("%v %v %v", result.GroupId, result.ArtifactId, result.Version)
		item := WF.NewItem(result.Version).Subtitle(makeName).Autocomplete(arg).Arg(arg).Valid(false).Icon(&aw.Icon{Value: fmt.Sprintf("icons/%v.png", strings.ToUpper(string(result.ArtifactId[0])))})
		//这里其实可以用modifier做复制粘贴了，
		item.NewModifier(aw.ModCmd).Subtitle("Open in browser").Arg(result.Url).Valid(true)
		item.NewModifier(aw.ModCtrl).Subtitle("Copy as maven format, scope may wrong").Arg(BuildMavenArg(result)).Valid(true)
		item.NewModifier(aw.ModShift).Subtitle("Copy as gradle format, configurations may wrong").Arg(BuildGradleArg(result)).Valid(true)
	}
}

func BuildVersionResultItem(result *service.VersionResult) {
	WF.NewItem("Maven").Subtitle("Copy as Maven format").Arg(result.Maven).Valid(true)
	WF.NewItem("Gradle").Subtitle("Copy as Gradle format").Arg(result.Gradle).Valid(true)
	WF.NewItem("SBT").Subtitle("Copy as SBT format").Arg(result.SBT).Valid(true)
	WF.NewItem("Ivy").Subtitle("Copy as Ivy format").Arg(result.Ivy).Valid(true)
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
