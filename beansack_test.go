package main

import (
	"context"
	"testing"
	"time"

	"github.com/k0kubun/pp"
)

const (
	DEFAULT_DB_NAME  = "master"
	DEFAULT_CONN_STR = "mongodb://localhost:27017"
)

var testCtx = context.Background()

func setupTestDB(t *testing.T) *Beansack {
	catalog := "postgresql://postgres:localpass@localhost:5432/beansackdb"
	storage := ".beansack/"
	return NewReadonlyBeansack(catalog, storage)
}

func TestQueryDistinctValues(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	pp.Println("REGIONS", db.DistinctRegions()[:5])
	pp.Println("ENTITIES", db.DistinctEntities()[:5])
	pp.Println("SOURCES", db.DistinctSources()[:5])
	pp.Println("CATEGORIES", db.DistinctCategories()[:5])
}

func TestQueryChatterStats(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	urls := []string{
		"https://www.slashgear.com/1896648/lifesaber-emergency-tool-usb-powered-features/",
		"https://www.wusa9.com/article/news/nation-world/trump-big-bill-may-have-political-cost/507-0c07fc4b-248b-4a3b-96c0-81d75a511228",
		"https://issuepay.app",
		"https://minutemirror.com.pk/lahore-high-court-halts-transfer-of-brazilian-monkeys-to-lahore-zoo-406544/",
		"https://jameshard.ing/pilot",
		"https://jobsbyreferral.com/",
		"https://llmapitest.com/",
		"https://htmlrev.com/",
	}

	pp.Println("CHATTERS", db.GetChatters(urls))
	pp.Println("AGGREGATES", db.GetBeanChatters(urls))
}

func TestQueryBeanExtensions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	urls := []string{
		"https://www.slashgear.com/1896648/lifesaber-emergency-tool-usb-powered-features/",
		"https://www.wusa9.com/article/news/nation-world/trump-big-bill-may-have-political-cost/507-0c07fc4b-248b-4a3b-96c0-81d75a511228",
		"https://issuepay.app",
		"https://minutemirror.com.pk/lahore-high-court-halts-transfer-of-brazilian-monkeys-to-lahore-zoo-406544/",
		"https://jameshard.ing/pilot",
		"https://jobsbyreferral.com/",
		"https://llmapitest.com/",
		"https://htmlrev.com/",
	}

	pp.Println("CATEGORIES", db.GetCategories(urls))
	pp.Println("SENTIMENTS", db.GetSentiments(urls))
	pp.Println("RELATED", db.GetRelated(urls))
	pp.Println("REGIONS", db.GetRegions(urls))
	pp.Println("ENTITIES", db.GetEntities(urls))
	pp.Println("GISTS", db.GetGists(urls))
	pp.Println("BEAN CLUSTERS", db.GetRelated(urls))
}

// func TestVectorSearch(t *testing.T) {
// 	ds := setupTestDB(t)
// 	query_emb := Float32Array{
// 		// ... embedding values from original test
// 	}

// 	sources := []string{"techstartups", "techradar"}
// 	beans, err := ds.VectorSearchBeansWithSelectFields(query_emb, 0.25, NEWS, time.Time{}, nil, nil, nil, sources, nil, nil, 0, 5, nil)
// 	noerror(err, "VECTOR SEARCH ERROR")
// 	datautils.PrintTable(
// 		beans,
// 		[]string{"kind", "title", "categories", "entities", "created", "source"},
// 		func(b *Bean) []string {
// 			return []string{b.Kind, b.Title, strings.Join(b.Categories, ", "), strings.Join(b.Entities, ", "), b.Created.Format(time.RFC3339), b.Source}
// 		},
// 	)
// }

func TestQueryBeans(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	categories := []string{"Artificial Intelligence", "Cloud Computing"}
	entities := []string{"ChatGPT", "Elon Musk"}

	beans, err := db.QueryLatestBeans(nil, []string{NEWS}, nil, nil, time.Now().AddDate(0, 0, -3), time.Time{}, categories, nil, entities, nil, 0, nil, nil, 5, 0, []string{DIGEST_COLUMNS})
	noerror(err, "QUERY BEANS ERROR")
	pp.Println("DIGESTS", beans)
}

func TestQueryRelated(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	urls := []string{
		"https://decrypt.co/328057/cloudflare-hits-kill-switch-ai-crawlers-entire-industry-cheers",
		"https://uk.pcmag.com/ai/158866/cloudflare-to-block-ai-crawlers-from-scraping-websites-unless-they-pay",
		"https://www.pcmag.com/news/cloudflare-to-block-ai-crawlers-from-scraping-websites-unless-they-pay",
		"https://www.engadget.com/ai/cloudflare-experiment-will-block-ai-bot-scrapers-unless-they-pay-a-fee-1215233…",
		"https://dailygalaxy.com/2025/07/red-jellyfish-bursts-storm-outer-space/",
	}
	pp.Println(db.GetRelated(urls))
}
