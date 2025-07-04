package main

import (
	"time"
)

const (
	NEWS      = "news"
	BLOG      = "blog"
	POST      = "post"
	GENERATED = "generated"
	COMMENT   = "comment"
)

type EmbeddingData struct {
	URL       string    `db:"url" bson:"url"`
	Embedding []float32 `db:"embedding" bson:"embedding"`
}

type TagData struct {
	URL string `db:"url"`
	Tag string `db:"tag"`
}

type Bean struct {
	URL           string    `db:"url"`
	Kind          string    `db:"kind"`
	Title         string    `db:"title"`
	TitleLength   int       `db:"title_length" bson:"num_words_in_title"`
	Content       string    `db:"content"`
	ContentLength int       `db:"content_length" bson:"num_words_in_content"`
	Summary       string    `db:"summary"`
	SummaryLength int       `db:"summary_length" bson:"num_words_in_summary"`
	Author        string    `db:"author"`
	Source        string    `db:"source"`
	Created       time.Time `db:"created" bson:"created"`
	Updated       time.Time `db:"updated" bson:"updated"`
	Collected     time.Time `db:"collected" bson:"collected"`
}

type GeneratedBean struct {
	Bean
	Intro       string   `db:"intro"`
	Analysis    []string `db:"analysis"`
	Insights    []string `db:"insights"`
	Verdict     string   `db:"verdict"`
	Predictions []string `db:"predictions"`
}

type Chatter struct {
	ChatterURL  string    `db:"chatter_url"`
	BeanURL     string    `db:"bean_url" bson:"url"`
	Source      string    `db:"source"`
	Forum       string    `db:"forum" bson:"group"`
	Collected   time.Time `db:"collected"`
	Likes       int       `db:"likes"`
	Comments    int       `db:"comments"`
	Shares      int       `db:"shares"`
	Subscribers int       `db:"subscribers"`
}

type Source struct {
	Name        string `db:"name" bson:"site_name"`
	Description string `db:"description" bson:"description"`
	BaseURL     string `db:"base_url" bson:"site_base_url"`
	DomainName  string `db:"domain_name" bson:"source"`
	Favicon     string `db:"favicon" bson:"site_favicon"`
	RSSFeed     string `db:"rss_feed" bson:"site_rss_feed"`
}
