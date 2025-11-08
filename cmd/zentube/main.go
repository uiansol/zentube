package main

import (
	"log"

	"github.com/uiansol/zentube/internal/config"
	"github.com/uiansol/zentube/internal/youtube"
)

func main() {
	if err := config.LoadEnv(); err != nil {
		panic(err)
	}

	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		panic(err)
	}

	if err := config.InjectEnvVariables(cfg); err != nil {
		panic(err)
	}

	yt, err := youtube.NewService(cfg.YouTube)
	if err != nil {
		log.Fatalf("youtube service: %v", err)
	}

	call := yt.Search.List([]string{"id", "snippet"}).
		Q("golang tutorial").
		Type("video").
		MaxResults(cfg.YouTube.MaxResults)

	resp, err := call.Do()
	if err != nil {
		log.Fatalf("search: %v", err)
	}

	for _, item := range resp.Items {
		if item.Id.VideoId != "" {
			link := "https://www.youtube.com/watch?v=" + item.Id.VideoId
			log.Printf("%s – %s", item.Snippet.Title, link)
		}
	}

	// Print resp.Items[0] structure with all fields for debugging
	if len(resp.Items) > 0 {
		item := resp.Items[0]
		log.Printf("First item: %+v", item)
	}
}
