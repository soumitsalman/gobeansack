package main

import (
	"fmt"
	"os"
	"time"

	"github.com/k0kubun/pp"
)

func flattenTags(url string, data []string) []TagData {
	var tags []TagData = make([]TagData, 0, len(data))
	for _, tag := range data {
		tags = append(tags, TagData{URL: url, Tag: tag})
	}
	return tags
}

func prepareTags(digests []Digest) map[string][]TagData {
	// rough initialization
	results := map[string][]TagData{
		"categories": make([]TagData, 0, 3*len(digests)),
		"sentiments": make([]TagData, 0, 3*len(digests)),
		"regions":    make([]TagData, 0, 3*len(digests)),
		"entities":   make([]TagData, 0, 3*len(digests)),
		"gist":       make([]TagData, 0, len(digests)),
	}
	for _, digest := range digests {
		results["categories"] = append(results["categories"], flattenTags(digest.URL, digest.Categories)...)
		results["sentiments"] = append(results["sentiments"], flattenTags(digest.URL, digest.Sentiments)...)
		results["regions"] = append(results["regions"], flattenTags(digest.URL, digest.Regions)...)
		results["entities"] = append(results["entities"], flattenTags(digest.URL, digest.Entities)...)
		results["gist"] = append(results["gist"], TagData{URL: digest.URL, Tag: digest.Gist})
	}
	return results
}

func hydrateTestDB(ds *Ducksack) {
	var importnum int64 = 50000
	ds.StoreBeans(getTestBeans(importnum))
	embeddings := getTestEmbeddings(importnum)
	ds.StoreEmbeddings(embeddings)
	ds.RectifyExtendedFields(embeddings, 3, 0.43)
	ds.StoreChatters(getTestChatters(400000))
	ds.StoreSources(getTestSources(importnum))
	digests := getTestDigests(importnum)
	tags := prepareTags(digests)

	if regions, ok := tags["regions"]; ok && len(regions) > 0 {
		ds.StoreTags(regions, BEAN_REGIONS)
	}
	if entities, ok := tags["entities"]; ok && len(entities) > 0 {
		ds.StoreTags(entities, BEAN_ENTITIES)
	}
	if gist, ok := tags["gist"]; ok && len(gist) > 0 {
		ds.StoreTags(gist, BEAN_GISTS)
	}
}

func testQueryDistinctValues(ds *Ducksack) {
	pp.Println("REGIONS", ds.DistinctRegions()[:3])
	pp.Println("ENTITIES", ds.DistinctEntities()[:3])
	pp.Println("SOURCES", ds.DistinctSources()[:3])
	pp.Println("CATEGORIES", ds.DistinctCategories()[:3])
}

func testQueryChatterStats(ds *Ducksack) {
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

	pp.Println("CHATTERS", ds.QueryChatters(urls)[:5])
	pp.Println("AGGREGATES", ds.QueryChatterAggregates(urls))
}

func testQueryBeanExtensions(ds *Ducksack) {
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

	pp.Println("CATEGORIES", ds.QueryCategories(urls))
	pp.Println("SENTIMENTS", ds.QuerySentiments(urls))
	pp.Println("RELATED", ds.QueryClusters(urls))
	pp.Println("REGIONS", ds.QueryRegions(urls))
	pp.Println("ENTITIES", ds.QueryEntities(urls))
	pp.Println("GISTS", ds.QueryGists(urls))
	pp.Println("BEAN AGGREGATES", ds.QueryBeans(urls))
}

func main() {
	// initialize database if needed
	init, err := os.ReadFile("./factory/init.sql")
	noerror(err)
	dim := 384
	dbpath := ".cache/test.db"
	initsql := fmt.Sprintf(string(init), dim, dim, "./factory/categories.parquet", dim, "./factory/sentiments.parquet")

	ds := NewDucksack(dbpath, initsql, dim)
	defer ds.Close()

	// hydrateTestDB(ds)
	// testDistinctQueries(ds)
	// testChatterStats(ds)
	// testQueryBeanExtensions(ds)
	pp.Println(ds.StreamBeans(NEWS, time.Now().AddDate(0, 0, -7), 10, 10))

	// titles := []string{
	// 	"Synthflow AI is bringing 'conversational' voice agents to call centers. Read the pitch deck that it used to raise $20 million.",
	// 	"Lahore high court halts transfer of Brazilian monkeys to Lahore zoo",
	// 	"Introducing the Klaviyo Partner Portal: our newest investment in helping our ecosystem grow with us",
	// 	"25 journalists jailed in Turkey in first quarter of 2025: report",
	// 	"KP police gets advanced anti drone tech to target terrorist drones",
	// }
	// urls := []string{
	// 	"https://greekreporter.com/2025/06/25/top-20-greek-islands-travel-greece/",
	// 	"https://www.businessinsider.com/synthflow-ai-pitch-deck-funding-voice-2025-6",
	// 	"https://analyticsindiamag.com/ai-news-updates/coralogix-introduces-olly-an-advanced-ai-powered-observability-agent-in-india/",
	// 	"https://minutemirror.com.pk/lahore-high-court-halts-transfer-of-brazilian-monkeys-to-lahore-zoo-406544/",
	// 	"https://www.klaviyo.com/blog/introducing-klaviyo-partner-portal",
	// 	"https://www.turkishminute.com/2025/06/25/turkey-relieved-as-iran-israel-truce-reduces-risk-of-regional-turmoil/",
	// 	"https://www.carsandhorsepower.com/featured/exploring-local-old-car-shows-near-you",
	// 	"https://philarchive.org/rec/SCHTCC-30",
	// }
	// beans := ds.QueryBeans(urls)
	// pp.Println(beans)
	// categories := ds.MatchCategories(urls, 3)
	// categories := ds.QueryTags(urls, BEAN_CATEGORIES)
	// pp.Println(categories)
	// sentiments := ds.MatchSentiments(urls, 3)
	// sentiments := ds.QueryTags(urls, BEAN_SENTIMENTS)
	// pp.Println(sentiments)
	// // clusters := ds.MatchClusters(urls, 0.5, 3)
	// // pp.Println(clusters)

	// // gists := ds.QueryTags(urls, BEAN_GISTS)
	// // pp.Println(gists)
	// chatters := ds.QueryChatters(urls)
	// pp.Println(chatters)
	// regions := ds.QueryTags(urls, BEAN_REGIONS)
	// pp.Println(regions)
	// entities := ds.QueryTags(urls, BEAN_ENTITIES)
	// pp.Println(entities)

	// urls := []string{
	// 	"https://www.slashgear.com/1896648/lifesaber-emergency-tool-usb-powered-features/",
	// 	"https://www.wusa9.com/article/news/nation-world/trump-big-bill-may-have-political-cost/507-0c07fc4b-248b-4a3b-96c0-81d75a511228",
	// 	"https://issuepay.app",
	// 	"https://minutemirror.com.pk/lahore-high-court-halts-transfer-of-brazilian-monkeys-to-lahore-zoo-406544/",
	// 	"https://jameshard.ing/pilot",
	// 	"https://jobsbyreferral.com/",
	// 	"https://llmapitest.com/",
	// 	"https://htmlrev.com/",
	// }

	// embeddings := ds.QueryBeanEmbeddings(urls)
	// for _, embedding := range embeddings {
	// 	// pp.Println(embedding.URL, embedding.Embedding)

	// 	similars := ds.VectorSearchBeans(embedding.Embedding, 5)
	// 	pp.Println(embedding.URL, datautils.Transform(similars, func(s *EmbeddingData) string {
	// 		return s.URL
	// 		// res, err := json.MarshalIndent(s, "", "  ")
	// 		// noerror(err)
	// 		// return string(res)
	// 	}))
	// }

	// pp.Println("CHATTERS", ds.QueryChatters(urls)[:5])
	// pp.Println("AGGREGATES", ds.QueryChatterAggregates(urls))
	// pp.Println("UPDATES", ds.QueryChatterUpdates(urls, 1))

	// pp.Println("CATEGORIES", ds.QueryCategories(urls))
	// pp.Println("SENTIMENTS", ds.QuerySentiments(urls))
	// pp.Println("RELATED", ds.QueryClusters(urls))
	// pp.Println("REGIONS", ds.QueryRegions(urls))
	// pp.Println("ENTITIES", ds.QueryEntities(urls))
	// pp.Println("GISTS", ds.QueryGists(urls))

	// start := time.Now()
	// res := ds.QueryBeansWithExtensions(urls)
	// pp.Println("BEAN AGGREGATES", res, time.Since(start))

}
