package beansack

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

const (
	K_URL        = "url"
	K_KIND       = "kind"
	K_TITLE      = "title"
	K_SUMMARY    = "summary"
	K_CONTENT    = "content"
	K_AUTHOR     = "author"
	K_SOURCE     = "source"
	K_IMAGE_URL  = "image_url"
	K_CREATED    = "created"
	K_CATEGORIES = "categories"
	K_SENTIMENTS = "sentiments"
	K_REGIONS    = "regions"
	K_ENTITIES   = "entities"
	K_GIST       = "gist"
	K_EMBEDDING  = "embedding"
	K_TRENDSCORE = "trend_score"
)

type Bean struct {
	URL        string    `db:"url" json:"url,omitempty"`
	Kind       string    `db:"kind" json:"kind,omitempty"`
	Title      string    `db:"title" json:"title,omitempty"`
	Summary    string    `db:"summary" json:"summary,omitempty"`
	Content    string    `db:"content" json:"content,omitempty"`
	Author     string    `db:"author" json:"author,omitempty"`
	Source     string    `db:"source" json:"source,omitempty"`
	ImageUrl   string    `db:"image_url" bson:"image_url" json:"image_url,omitempty"`
	Created    time.Time `db:"created" bson:"created" json:"created,omitempty,omitzero"`
	Embedding  []float32 `db:"embedding" json:"embedding,omitempty"`
	Gist       string    `db:"gist" json:"gist,omitempty"`
	Categories []string  `db:"categories" json:"categories,omitempty"`
	Sentiments []string  `db:"sentiments" json:"sentiments,omitempty"`
	Regions    []string  `db:"regions" json:"regions,omitempty"`
	Entities   []string  `db:"entities" json:"entities,omitempty"`
}

type BeanAggregate struct {
	Bean
	Related     []string  `db:"related" json:"related,omitempty"`
	ClusterId   string    `db:"cluster_id" json:"cluster_id,omitempty"`
	ClusterSize int       `db:"cluster_size" json:"cluster_size,omitempty"`
	Updated     time.Time `db:"updated" json:"updated,omitempty,omitzero"`
	Likes       int       `db:"likes" json:"likes,omitempty"`
	Comments    int       `db:"comments" json:"comments,omitempty"`
	Subscribers int       `db:"subscribers" json:"subscribers,omitempty"`
	Shares      int       `db:"shares" json:"shares,omitempty"`
	Distance    float64   `db:"distance" json:"-"`
	TrendScore  float64   `db:"trend_score" json:"trend_score,omitempty"`
}

type Chatter struct {
	ChatterURL  string    `db:"chatter_url" bson:"chatter_url" json:"chatter_url"`
	URL         string    `db:"url" bson:"url" json:"url"`
	Source      string    `db:"source" json:"source,omitempty"`
	Forum       string    `db:"forum" bson:"group" json:"forum,omitempty"`
	Collected   time.Time `db:"collected" json:"-"`
	Likes       int       `db:"likes" json:"likes,omitempty"`
	Comments    int       `db:"comments" json:"comments,omitempty"`
	Subscribers int       `db:"subscribers" json:"subscribers,omitempty"`
}

type BeanChatter struct {
	URL         string    `db:"url"`                // url of the bean
	Collected   time.Time `db:"collected" json:"-"` // last time some chatter was collected
	Likes       int       `db:"likes"`
	Comments    int       `db:"comments"`
	Subscribers int       `db:"subscribers"`
	Shares      int       `db:"shares"`
}

type Publisher struct {
	Source      string    `db:"source" json:"source,omitempty"`
	BaseURL     string    `db:"base_url" json:"base_url,omitempty"`
	SiteName    string    `db:"site_name" json:"site_name,omitempty"`
	Description string    `db:"description" json:"description,omitempty"`
	Favicon     string    `db:"favicon" json:"favicon,omitempty"`
	RSSFeed     string    `db:"rss_feed" json:"-"`
	Collected   time.Time `db:"collected" json:"-"`
}
