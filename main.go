package main

import (
	"os"
	"strings"
	"time"

	"github.com/k0kubun/pp"
	datautils "github.com/soumitsalman/data-utils"
)

func flattenTags(url string, data []string) []TagData {
	var tags []TagData = make([]TagData, 0, len(data))
	for _, tag := range data {
		tags = append(tags, TagData{URL: url, Tag: tag})
	}
	return tags
}

func prepareTags(beans []Bean) map[string][]TagData {
	// rough initialization
	results := map[string][]TagData{
		"categories": make([]TagData, 0, 3*len(beans)),
		"sentiments": make([]TagData, 0, 3*len(beans)),
		"regions":    make([]TagData, 0, 3*len(beans)),
		"entities":   make([]TagData, 0, 3*len(beans)),
		"gist":       make([]TagData, 0, len(beans)),
	}
	for _, bean := range beans {
		results["categories"] = append(results["categories"], flattenTags(bean.URL, bean.Categories)...)
		results["sentiments"] = append(results["sentiments"], flattenTags(bean.URL, bean.Sentiments)...)
		results["regions"] = append(results["regions"], flattenTags(bean.URL, bean.Regions)...)
		results["entities"] = append(results["entities"], flattenTags(bean.URL, bean.Entities)...)
		results["gist"] = append(results["gist"], TagData{URL: bean.URL, Tag: bean.Gist})
	}
	return results
}

func hydrateTestDB(ds *Ducksack) {
	var importnum int64 = 60000

	beans := getTestBeans(importnum)
	ds.StoreBeans(beans)
	embs := datautils.Filter(beans, func(b *Bean) bool {
		return len(b.Embedding) > 0
	})
	ds.StoreEmbeddings(embs)

	tags := prepareTags(beans)

	if categories, ok := tags["categories"]; ok && len(categories) > 0 {
		ds.StoreTags(categories, BEAN_CATEGORIES)
	}
	if sentiments, ok := tags["sentiments"]; ok && len(sentiments) > 0 {
		ds.StoreTags(sentiments, BEAN_SENTIMENTS)
	}
	if regions, ok := tags["regions"]; ok && len(regions) > 0 {
		ds.StoreTags(regions, BEAN_REGIONS)
	}
	if entities, ok := tags["entities"]; ok && len(entities) > 0 {
		ds.StoreTags(entities, BEAN_ENTITIES)
	}
	if gist, ok := tags["gist"]; ok && len(gist) > 0 {
		ds.StoreTags(gist, BEAN_GISTS)
	}

	// ds.RectifyExtendedFields(beans, 3, 0.43)
	ds.StoreChatters(getTestChatters(400000))
	ds.StoreSources(getTestSources(importnum))
	// digests := getTestDigests(importnum)
}

func testQueryDistinctValues(ds *Ducksack) {
	pp.Println("REGIONS", ds.DistinctRegions()[:5])
	pp.Println("ENTITIES", ds.DistinctEntities()[:5])
	pp.Println("SOURCES", ds.DistinctSources()[:5])
	pp.Println("CATEGORIES", ds.DistinctCategories()[:5])
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

	pp.Println("CHATTERS", ds.GetChatters(urls)[:5])
	pp.Println("AGGREGATES", ds.GetBeanChatters(urls))
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

	pp.Println("CATEGORIES", ds.GetCategories(urls))
	pp.Println("SENTIMENTS", ds.GetSentiments(urls))
	pp.Println("RELATED", ds.GetClusters(urls))
	pp.Println("REGIONS", ds.GetRegions(urls))
	pp.Println("ENTITIES", ds.GetEntities(urls))
	pp.Println("GISTS", ds.GetGists(urls))
	pp.Println("BEAN AGGREGATES", ds.GetBeans(urls))
	pp.Println("BEAN CLUSTERS", ds.GetClusters(urls))
}

func testStreamBeans(ds *Ducksack) {
	categories := []string{"Artificial Intelligence", "Cloud Computing"}
	entities := []string{"ChatGPT", "Elon Musk"}
	for i := int64(0); i < 5; i++ {
		datautils.PrintTable(
			ds.StreamBeans(NEWS, time.Now().AddDate(0, 0, -45), categories, []string{}, entities, i*5, 6),
			[]string{"kind", "title", "categories", "entities"},
			func(b *Bean) []string {
				return []string{b.Kind, b.Title, strings.Join(b.Categories, ", "), strings.Join(b.Entities, ", ")}
			},
		)
	}
}

func main() {
	// initialize database if needed
	init, err := os.ReadFile("./factory/init.sql")
	noerror(err, "READ SQL ERROR")
	var dim int = 384
	var cluster_eps float32 = 0.43
	dbpath := ".cache/test.db" // "./.ducklake/"
	initsql := string(init)    // fmt.Sprintf(string(init), dim)

	ds := NewBeansack(dbpath, initsql, dim, cluster_eps)
	defer ds.Close()

	// hydrateTestDB(ds)
	// testQueryDistinctValues(ds)
	// testQueryChatterStats(ds)
	// testQueryBeanExtensions(ds)
	testStreamBeans(ds)

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
