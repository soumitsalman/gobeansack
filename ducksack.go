package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/marcboeker/go-duckdb"
	_ "github.com/marcboeker/go-duckdb"
	_ "github.com/mattn/go-sqlite3"
	datautils "github.com/soumitsalman/data-utils"
)

const (
	BEANS           = "beans"
	BEAN_EMBEDDINGS = "bean_embeddings"
	BEAN_CLUSTERS   = "bean_clusters"
	BEAN_CATEGORIES = "bean_categories"
	BEAN_SENTIMENTS = "bean_sentiments"
	BEAN_GISTS      = "bean_gists"
	BEAN_REGIONS    = "bean_regions"
	BEAN_ENTITIES   = "bean_entities"
	CHATTERS        = "chatters"
	SOURCES         = "sources"
	CATEGORIES      = "categories"
	SENTIMENTS      = "sentiments"
)

const (
	MAX_COMPUTED_TAGS = 3
	MAX_RELATED_EPS   = 0.43
)

type Ducksack struct {
	connector *duckdb.Connector
	db        *sql.DB
	query     *sqlx.DB
	dim       int
}

////////// INITIALIZE DATABASE //////////

func NewDucksack(datapath string, initsql string, vectordim int) *Ducksack {
	conn, err := duckdb.NewConnector(fmt.Sprintf("%s?threads=%d", datapath, max(1, runtime.NumCPU()-1)), nil)
	noerror(err)

	// open connection
	db := sql.OpenDB(conn)
	if initsql != "" {
		_, err = db.Exec(initsql)
		noerror(err)
	}

	return &Ducksack{
		connector: conn,
		db:        db,
		query:     sqlx.NewDb(db, "duckdb"),
		dim:       vectordim,
	}
}

////////// STORING FUNCTIONS //////////

// func (ds *Ducksack) getAppender(table string) *duckdb.Appender {
// 	conn, err := ds.connector.Connect(context.Background())
// 	noerror(err)
// 	appender, err := duckdb.NewAppenderFromConn(conn, "", table)
// 	noerror(err)
// 	return appender
// }

func appendToTable[T any](ds *Ducksack, table string, data []T, getfieldvalues func(item T) []driver.Value) int {
	conn, err := ds.connector.Connect(context.Background())
	noerror(err)
	appender, err := duckdb.NewAppenderFromConn(conn, "", table)
	noerror(err)
	defer appender.Close()
	count := 0
	for _, item := range data {
		if err := appender.AppendRow(getfieldvalues(item)...); err != nil {
			log.Println(err)
		} else {
			count++
		}
	}
	return count
}

func prepareBeans(beans []Bean) []Bean {
	now := time.Now()
	for i := range beans {
		if beans[i].Created.IsZero() {
			beans[i].Created = now
		}
		// if beans[i].Updated.IsZero() {
		// 	beans[i].Updated = now
		// }
		if beans[i].Collected.IsZero() {
			beans[i].Collected = now
		}
		if beans[i].TitleLength == 0 {
			beans[i].TitleLength = len(strings.Fields(beans[i].Title))
		}
		if beans[i].ContentLength == 0 {
			beans[i].ContentLength = len(strings.Fields(beans[i].Content))
		}
		if beans[i].SummaryLength == 0 {
			beans[i].SummaryLength = len(strings.Fields(beans[i].Summary))
		}
	}
	return beans
}

func (ds *Ducksack) StoreBeans(beans []Bean) int {
	beans = prepareBeans(beans)
	return appendToTable(ds, BEANS, beans, func(bean Bean) []driver.Value {
		return []driver.Value{bean.URL, bean.Kind, bean.Title, bean.TitleLength, bean.Content, bean.ContentLength, bean.Summary, bean.SummaryLength, bean.Author, bean.Source, bean.Created, bean.Collected}
	})
}

func (ds *Ducksack) StoreEmbeddings(embeddings []EmbeddingData) int {
	return appendToTable(ds, BEAN_EMBEDDINGS, embeddings, func(embedding EmbeddingData) []driver.Value {
		return []driver.Value{embedding.URL, embedding.Embedding}
	})
}

func (ds *Ducksack) RectifyExtendedFields(embeddings []EmbeddingData, max_computed_tags int, max_cluster_eps float32) {
	urls := datautils.Transform(embeddings, func(e *EmbeddingData) string {
		return e.URL
	})

	const _SQL_INSERT_CATEGORIES = `
	INSERT INTO bean_categories (url, category)
	SELECT m1.url, m1.category FROM category_mappings m1
	WHERE
		m1.url IN (?) AND 
		m1.category IN (
			SELECT category FROM category_mappings m2
			WHERE m1.url == m2.url
			ORDER BY m2.distance LIMIT %d
		);`
	updateItems(ds, fmt.Sprintf(_SQL_INSERT_CATEGORIES, max_computed_tags), urls)

	const _SQL_INSERT_SENTIMENTS = `
	INSERT INTO bean_sentiments (url, sentiment)
	SELECT m1.url, m1.sentiment FROM sentiment_mappings m1
	WHERE
		m1.url IN (?) AND 
		m1.sentiment IN (
			SELECT sentiment FROM sentiment_mappings m2
			WHERE m1.url == m2.url
			ORDER BY m2.distance LIMIT %d
		);`
	updateItems(ds, fmt.Sprintf(_SQL_INSERT_SENTIMENTS, max_computed_tags), urls)

	const _SQL_INSERT_CLUSTERS = `
	INSERT INTO bean_clusters (url, related)
	SELECT url, related FROM cluster_mappings
	WHERE url IN (?) AND distance < %f;`
	updateItems(ds, fmt.Sprintf(_SQL_INSERT_CLUSTERS, max_cluster_eps), urls)
}

func (ds *Ducksack) StoreTags(tags []TagData, tag_table string) int {
	return appendToTable(ds, tag_table, tags, func(tag TagData) []driver.Value {
		return []driver.Value{tag.URL, tag.Tag}
	})
}

func prepareChatters(chatters []Chatter) []Chatter {
	now := time.Now()
	for i := range chatters {
		if chatters[i].Collected.IsZero() {
			chatters[i].Collected = now
		}
	}
	return chatters
}

func (ds *Ducksack) StoreChatters(chatters []Chatter) int {
	chatters = prepareChatters(chatters)
	return appendToTable(ds, CHATTERS, chatters, func(chatter Chatter) []driver.Value {
		return []driver.Value{chatter.ChatterURL, chatter.BeanURL, chatter.Collected, chatter.Source, chatter.Forum, chatter.Likes, chatter.Comments, chatter.Subscribers}
	})
}

func (ds *Ducksack) StoreSources(sources []Source) int {
	return appendToTable(ds, SOURCES, sources, func(source Source) []driver.Value {
		return []driver.Value{source.Name, source.Description, source.BaseURL, source.DomainName, source.Favicon, source.RSSFeed}
	})
}

////////// QUERY HELPERS //////////

func mustIn(query string, args ...any) (string, []any) {
	query, args, err := sqlx.In(query, args...)
	noerror(err)
	return query, args
}

func mustSelect[T any](ds *Ducksack, query string, args ...any) []T {
	var data []T
	noerror(ds.query.Select(&data, query, args...))
	return data
}

func queryItems[T any](ds *Ducksack, sql string, urls []string) []T {
	query, args := mustIn(sql, urls)
	var data []T
	noerror(ds.query.Select(&data, query, args...))
	return data
}

func updateItems(ds *Ducksack, expr string, urls []string) {
	query, args := mustIn(expr, urls)
	_, err := ds.db.Exec(query, args...)
	noerror(err)
}

////////// QUERY FUNCTIONS //////////

func (ds *Ducksack) Exists(urls []string) []string {
	return queryItems[string](ds, "SELECT url FROM beans WHERE url IN (?)", urls)
}

// func (ds *Ducksack) QueryBeans(urls []string) []Bean {
// 	return queryItems[Bean](ds, "SELECT * FROM beans WHERE url IN (?)", urls)
// }

func (ds *Ducksack) QueryBeans(urls []string) []ExtendedBean {
	return queryItems[ExtendedBean](ds, "SELECT * FROM bean_aggregates WHERE url IN (?);", urls)
}

func (ds *Ducksack) QueryEmbeddings(urls []string) []ExtendedBean {
	return queryItems[ExtendedBean](ds, "SELECT * FROM bean_embeddings WHERE url IN (?);", urls)
}

func (ds *Ducksack) QueryGists(urls []string) []ExtendedBean {
	return queryItems[ExtendedBean](ds, "SELECT * FROM bean_gists WHERE url IN (?);", urls)
}

func (ds *Ducksack) QueryRegions(urls []string) []ExtendedBean {
	const _SQL_QUERY_REGIONS = `
	SELECT url, LIST(region) AS regions FROM bean_regions
	WHERE url IN (?) GROUP BY url;`
	return queryItems[ExtendedBean](ds, _SQL_QUERY_REGIONS, urls)
}

func (ds *Ducksack) QueryEntities(urls []string) []ExtendedBean {
	const _SQL_QUERY_ENTITIES = `
	SELECT url, LIST(entity) AS entities FROM bean_entities
	WHERE url IN (?) GROUP BY url;`
	return queryItems[ExtendedBean](ds, _SQL_QUERY_ENTITIES, urls)
}

func (ds *Ducksack) QueryCategories(urls []string) []ExtendedBean {
	const _SQL_QUERY_CATEGORIES = `
	SELECT url, LIST(category) AS categories FROM bean_categories 
	WHERE url IN (?) GROUP BY url;`
	return queryItems[ExtendedBean](ds, _SQL_QUERY_CATEGORIES, urls)
}

func (ds *Ducksack) QuerySentiments(urls []string) []ExtendedBean {
	const _SQL_QUERY_SENTIMENTS = `
	SELECT url, LIST(sentiment) AS sentiments FROM bean_sentiments 
	WHERE url IN (?) GROUP BY url;`
	return queryItems[ExtendedBean](ds, _SQL_QUERY_SENTIMENTS, urls)
}

func (ds *Ducksack) QueryClusters(urls []string) []ExtendedBean {
	const _SQL_QUERY_CLUSTERS = `
	SELECT url, LIST(related) AS related FROM bean_clusters
	WHERE url IN (?) GROUP BY url;`
	return queryItems[ExtendedBean](ds, _SQL_QUERY_CLUSTERS, urls)
}

/////////// CHATTER QUERIES //////////

func (ds *Ducksack) QueryChatters(urls []string) []Chatter {
	query, args := mustIn("SELECT * FROM chatters WHERE bean_url IN (?) ORDER BY collected DESC", urls)
	return mustSelect[Chatter](ds, query, args...)
}

func (ds *Ducksack) QueryChatterAggregates(urls []string) []ChatterAggregate {
	query, args := mustIn("SELECT * FROM chatter_aggregates WHERE url IN (?)", urls)
	return mustSelect[ChatterAggregate](ds, query, args...)
}

////////// DISTINCT ITEMS //////////

func (ds *Ducksack) DistinctRegions() []string {
	const _SQL_GET_ALL_REGIONS = `SELECT DISTINCT region FROM bean_regions;`
	return mustSelect[string](ds, _SQL_GET_ALL_REGIONS)
}

func (ds *Ducksack) DistinctEntities() []string {
	const _SQL_GET_ALL_ENTITIES = `SELECT DISTINCT entity FROM bean_entities;`
	return mustSelect[string](ds, _SQL_GET_ALL_ENTITIES)
}

func (ds *Ducksack) DistinctCategories() []string {
	const _SQL_GET_ALL_CATEGORIES = `SELECT category FROM categories;`
	return mustSelect[string](ds, _SQL_GET_ALL_CATEGORIES)
}

func (ds *Ducksack) DistinctSentiments() []string {
	const _SQL_GET_ALL_SENTIMENTS = `SELECT sentiment FROM sentiments;`
	return mustSelect[string](ds, _SQL_GET_ALL_SENTIMENTS)
}

func (ds *Ducksack) DistinctSources() []string {
	const _SQL_GET_ALL_SOURCES = `SELECT base_url AS value FROM sources;`
	return mustSelect[string](ds, _SQL_GET_ALL_SOURCES)
}

//////////// STREAM QUERIES ///////////

func (ds *Ducksack) StreamBeans(kind string, created_after time.Time, offset int64, limit int64) []ExtendedBean {
	const _SQL_STREAM_BEANS = `
	SELECT * FROM bean_aggregates
	WHERE kind = ? AND created >= ?
	ORDER BY created DESC
	OFFSET ? LIMIT ?;`
	return mustSelect[ExtendedBean](ds, _SQL_STREAM_BEANS, kind, created_after, offset, limit)
}

// // first take the chatters ONLY for the filtered urls
// // then take the current chatters and group by id
// // then then add/agg per bean
// // take the ones that got updated in last 1 day
// // take the chatters from 1 day ago per id
// // then aggregate per bean
// // then subtract
// const _SQL_QUERY_CHATTER_UPDATES = `
// WITH
// filtered_chatters AS (
//     SELECT * FROM chatters WHERE bean_url IN (?)
// ),
// current_agg AS (
// 	SELECT
//         bean_url,
//         MAX(collected) as collected,
//         SUM(likes) as likes,
//         SUM(comments) as comments,
//         SUM(subscribers) as subscribers,
//         COUNT(chatter_url) as shares,

//     FROM (
// 		SELECT
// 			chatter_url,
// 			FIRST(bean_url) as bean_url,
// 			MAX(collected) as collected,
// 			MAX(likes) as likes,
// 			MAX(comments) as comments,
// 			MAX(subscribers) as subscribers
// 		FROM filtered_chatters
// 		GROUP BY chatter_url
// 	)
//     GROUP BY bean_url
// ),
// before_agg AS (
// 	SELECT
//         bean_url,
//         MAX(collected) as collected,
//         SUM(likes) as likes,
//         SUM(comments) as comments,
//         SUM(subscribers) as subscribers,
//         COUNT(chatter_url) as shares
//     FROM (
// 		SELECT
// 			chatter_url,
// 			FIRST(bean_url) as bean_url,
// 			MAX(collected) as collected,
// 			MAX(likes) as likes,
// 			MAX(comments) as comments,
// 			MAX(subscribers) as subscribers
// 		FROM filtered_chatters
// 		WHERE collected + INTERVAL 1 DAY < CURRENT_TIMESTAMP
// 		GROUP BY chatter_url
// 	)
//     GROUP BY bean_url
// )
// SELECT
// 	ca.bean_url as url,
// 	ca.collected as last_collected,
// 	COALESCE(ca.likes, 0) - COALESCE(ba.likes, 0) as total_likes,
// 	COALESCE(ca.comments, 0) - COALESCE(ba.comments, 0) as total_comments,
// 	COALESCE(ca.subscribers, 0) - COALESCE(ba.subscribers, 0) as total_subscribers,
// 	COALESCE(ca.shares, 0) - COALESCE(ba.shares, 0) as total_shares
// FROM current_agg ca
// LEFT JOIN before_agg ba
// ON ca.bean_url = ba.bean_url
// WHERE
// 	ca.collected + INTERVAL 1 day >= CURRENT_TIMESTAMP AND
// 	(total_likes > 0 OR total_comments > 0 OR total_subscribers > 0 OR total_shares > 0);
// `

// func (ds *Ducksack) QueryChatterUpdates(urls []string) []ChatterAggregate {
// 	query, args := mustIn(_SQL_QUERY_CHATTER_UPDATES, urls)
// 	return mustSelect[ChatterAggregate](ds, query, args...)
// }

////////// VECTOR SEARCH //////////

func (ds *Ducksack) VectorSearchBeans(embedding []float32, limit int) []EmbeddingData {
	const _SQL_VECTOR_SEARCH_BEANS = `
	SELECT * FROM bean_embeddings
	ORDER BY array_cosine_distance(embedding, ?::FLOAT[%d])
	LIMIT ?`
	sql := fmt.Sprintf(_SQL_VECTOR_SEARCH_BEANS, len(embedding))
	return mustSelect[EmbeddingData](ds, sql, Float32Array(embedding), limit)
}

func (ds *Ducksack) Close() {
	noerror(ds.query.Close())
	noerror(ds.db.Close())
}

func noerror(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func logerror(err error) {
	if err != nil {
		log.Println(err)
	}
}
