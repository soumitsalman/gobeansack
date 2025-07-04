package main

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ctx = context.Background()

func getMongoCollection(database string, collection string) *mongo.Collection {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	noerror(err)
	db := client.Database(database)
	return db.Collection(collection)
}

func mongoFind[T any](db string, collection string, filter interface{}, limit int64, projection interface{}) []T {
	if filter == nil {
		filter = map[string]interface{}{}
	}
	if projection == nil {
		projection = map[string]interface{}{}
	}
	coll := getMongoCollection(db, collection)
	find_options := options.Find().SetLimit(limit).SetProjection(projection)
	cursor, err := coll.Find(ctx, filter, find_options)
	noerror(err)
	defer cursor.Close(ctx)
	items := []T{}
	for cursor.Next(ctx) {
		var item T
		err := cursor.Decode(&item)
		noerror(err)
		items = append(items, item)
	}
	return items
}

// func getTestBeans(limit int64) []Bean {
// 	collection := getMongoCollection("20250701", "beans")
// 	find_options := options.Find().SetLimit(limit).SetProjection(map[string]interface{}{
// 		"_id":       0,
// 		"url":       1,
// 		"title":     1,
// 		"content":   1,
// 		"summary":   1,
// 		"created":   1,
// 		"updated":   1,
// 		"collected": 1,
// 		"kind":      1,
// 		"source":    1,
// 		"author":    1,
// 	})
// 	cursor, err := collection.Find(
// 		ctx,
// 		map[string]interface{}{
// 			"content": map[string]interface{}{
// 				"$exists": true,
// 			},
// 			"summary": map[string]interface{}{
// 				"$exists": true,
// 			},
// 		},
// 		find_options,
// 	)
// 	noerror(err)
// 	defer cursor.Close(ctx)
// 	beans := []Bean{}
// 	for cursor.Next(ctx) {
// 		var bean Bean
// 		err := cursor.Decode(&bean)
// 		noerror(err)
// 		beans = append(beans, bean)
// 	}
// 	return beans
// }

func getTestBeans(limit int64) []Bean {
	return mongoFind[Bean](
		"master",
		"beans",
		map[string]interface{}{
			"content": map[string]interface{}{
				"$exists": true,
			},
			"summary": map[string]interface{}{
				"$exists": true,
			},
		},
		limit,
		map[string]interface{}{
			"url":       1,
			"title":     1,
			"content":   1,
			"summary":   1,
			"created":   1,
			"updated":   1,
			"collected": 1,
			"kind":      1,
			"source":    1,
			"author":    1,
		},
	)
}

func getTestEmbeddings(limit int64) []EmbeddingData {
	return mongoFind[EmbeddingData](
		"master",
		"beans",
		map[string]interface{}{
			"embedding": map[string]interface{}{
				"$exists": true,
			},
		},
		limit,
		map[string]interface{}{
			"url":       1,
			"embedding": 1,
		},
	)
}

// func getTestEmbeddings(limit int64) []EmbeddingData {
// 	collection := getMongoCollection("20250702", "beans")
// 	find_options := options.Find().SetLimit(limit).SetProjection(map[string]interface{}{
// 		"_id":       1,
// 		"embedding": 1,
// 	})
// 	cursor, err := collection.Find(
// 		ctx,
// 		map[string]interface{}{
// 			"embedding": map[string]interface{}{
// 				"$exists": true,
// 			},
// 		},
// 		find_options,
// 	)
// 	noerror(err)
// 	defer cursor.Close(ctx)
// 	embeddings := []EmbeddingData{}
// 	for cursor.Next(ctx) {
// 		var item EmbeddingData
// 		err := cursor.Decode(&item)
// 		noerror(err)
// 		embeddings = append(embeddings, item)
// 	}
// 	return embeddings
// }

type Digest struct {
	URL        string   `bson:"url"`
	Categories []string `bson:"categories"`
	Sentiments []string `bson:"sentiments"`
	Gist       string   `bson:"gist"`
	Regions    []string `bson:"regions"`
	Entities   []string `bson:"entities"`
}

func getTestDigests(limit int64) []Digest {
	digests := mongoFind[Digest](
		"master",
		"beans",
		map[string]interface{}{
			"gist": map[string]interface{}{
				"$exists": true,
			},
		},
		limit,
		map[string]interface{}{
			"url":        1,
			"gist":       1,
			"regions":    1,
			"entities":   1,
			"categories": 1,
			"sentiments": 1,
		},
	)
	return digests
}

func getTestChatters(limit int64) []Chatter {
	chatters := mongoFind[Chatter](
		"master",
		"chatters",
		map[string]interface{}{
			"likes": map[string]interface{}{
				"$exists": true,
			},
			"comments": map[string]interface{}{
				"$exists": true,
			},
		},
		limit,
		nil,
	)
	// fmt.Println(chatters[3])
	return chatters
}

func getTestSources(limit int64) []Source {
	sources := mongoFind[Source]("espresso", "sources", nil, limit, nil)
	return sources
}
