-- INSTALL ducklake;
-- ATTACH 'ducklake:.cache/catalog.db' AS beanlake (DATA_PATH '.cache/');
-- USE beanlake;

INSTALL vss;
LOAD vss;

CREATE TABLE IF NOT EXISTS bean_cores (
    url VARCHAR NOT NULL PRIMARY KEY,
    kind VARCHAR NOT NULL,
    title VARCHAR,
    title_length INTEGER DEFAULT 0,
    content TEXT,
    content_length INTEGER DEFAULT 0,
    summary TEXT,
    summary_length INTEGER DEFAULT 0,
    author VARCHAR,
    source VARCHAR,
    created TIMESTAMP,
    collected TIMESTAMP
);

CREATE TABLE IF NOT EXISTS generated_beans (
    url VARCHAR NOT NULL PRIMARY KEY,
    intro TEXT,
    analysis TEXT[],
    insights TEXT[],
    verdict TEXT,
    predictions TEXT[]
);
CREATE TABLE IF NOT EXISTS bean_embeddings (
    url VARCHAR NOT NULL PRIMARY KEY,
    embedding FLOAT[%d] NOT NULL
);

-- CREATE TABLE IF NOT EXISTS bean_clusters (
--     url VARCHAR NOT NULL,
--     related VARCHAR NOT NULL
-- );
CREATE TABLE IF NOT EXISTS bean_categories (
    url VARCHAR NOT NULL,
    category VARCHAR NOT NULL
);
CREATE TABLE IF NOT EXISTS bean_sentiments (
    url VARCHAR NOT NULL,
    sentiment VARCHAR NOT NULL
);
CREATE TABLE IF NOT EXISTS bean_gists (
    url VARCHAR NOT NULL,
    gist TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS bean_regions (
    url VARCHAR NOT NULL,
    region VARCHAR NOT NULL
);
CREATE TABLE IF NOT EXISTS bean_entities (
    url VARCHAR NOT NULL,
    entity VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS chatters (
    chatter_url VARCHAR NOT NULL,
    bean_url VARCHAR NOT NULL,
    collected TIMESTAMP NOT NULL,
    source VARCHAR NOT NULL,
    forum VARCHAR,
    likes INTEGER DEFAULT 0,
    comments INTEGER DEFAULT 0,
    subscribers INTEGER DEFAULT 0
);

CREATE VIEW IF NOT EXISTS bean_chatters AS
WITH 
    max_stats AS (
        SELECT 
            chatter_url,
            MAX(likes) as likes, 
            MAX(comments) as comments
        FROM chatters 
        GROUP BY chatter_url
    ),
    first_seen AS (
        SELECT 
            fs.chatter_url,
            MIN(fs.collected) as collected
        FROM chatters fs 
        LEFT JOIN max_stats mx ON fs.chatter_url = mx.chatter_url
        WHERE fs.likes = mx.likes AND fs.comments = mx.comments
        GROUP BY fs.chatter_url
    )
SELECT 
    bean_url as url,
    MAX(collected) as collected,
    SUM(likes) as likes, 
    SUM(comments) as comments, 
    SUM(subscribers) as subscribers,
    COUNT(chatter_url) as shares
FROM(
    SELECT 
        ch.chatter_url,
        ch.bean_url, 
        ch.collected, 
        ch.likes, 
        ch.comments,
        ch.subscribers
    FROM chatters ch 
    LEFT JOIN first_seen fs ON fs.chatter_url = ch.chatter_url  
    WHERE fs.collected = ch.collected
) 
GROUP BY bean_url;

CREATE TABLE IF NOT EXISTS sources (
    name VARCHAR,
    description TEXT,
    base_url VARCHAR NOT NULL,
    domain_name VARCHAR NOT NULL PRIMARY KEY,
    favicon VARCHAR,
    rss_feed VARCHAR
);

-- CREATE TABLE IF NOT EXISTS categories AS
-- SELECT 
--     _id as category,
--     embedding::FLOAT[] as embedding
-- FROM read_parquet('factory/categories.parquet');

-- CREATE VIEW IF NOT EXISTS category_mappings AS
-- SELECT 
--     url,
--     category, 
--     array_cosine_distance(b.embedding, c.embedding) as distance
-- FROM bean_embeddings b CROSS JOIN categories c;

-- CREATE TABLE IF NOT EXISTS sentiments AS
-- SELECT 
--     _id as sentiment,
--     embedding::FLOAT[] as embedding
-- FROM read_parquet('factory/sentiments.parquet');

-- CREATE VIEW IF NOT EXISTS sentiment_mappings AS
-- SELECT 
--     url,
--     sentiment, 
--     array_cosine_distance(b.embedding, s.embedding) as distance
-- FROM bean_embeddings b CROSS JOIN sentiments s;

CREATE VIEW IF NOT EXISTS bean_clusters AS 
SELECT 
    be1.url as url, 
    be2.url as related, 
    array_distance(be1.embedding, be2.embedding) as distance
FROM bean_embeddings be1 CROSS JOIN bean_embeddings be2
WHERE be1.url != be2.url AND distance < %f;

CREATE VIEW IF NOT EXISTS bean_aggregates AS
SELECT
    b.url, b.kind, b.title, b.title_length, b.summary, b.summary_length, b.author, b.source, b.created, b.collected,
    e.embedding,
    g.gist,    
    COALESCE(ch.collected, b.created) as updated,
    COALESCE(ch.likes, 0) as likes,
    COALESCE(ch.comments, 0) as comments,
    COALESCE(ch.subscribers, 0) as subscribers,
    COALESCE(ch.shares, 0) as shares,
    c.categories,
    s.sentiments,
    r.regions,
    n.entities
FROM bean_cores b
LEFT JOIN bean_embeddings e ON b.url = e.url
LEFT JOIN bean_gists g ON b.url = g.url
LEFT JOIN bean_chatters ch ON b.url = ch.url
-- LEFT JOIN bean_categories c ON b.url = c.url
-- LEFT JOIN bean_sentiments s ON b.url = s.url
-- LEFT JOIN bean_regions r ON b.url = r.url
-- LEFT JOIN bean_entities n ON b.url = n.url
LEFT JOIN (
	SELECT url, LIST(category) as categories FROM bean_categories GROUP BY url
) as c ON b.url = c.url
LEFT JOIN (
	SELECT url, LIST(sentiment) as sentiments FROM bean_sentiments GROUP BY url
) as s ON b.url = s.url
LEFT JOIN (
	SELECT url, LIST(region) as regions FROM bean_regions GROUP BY url
) as r ON b.url = r.url
LEFT JOIN (
	SELECT url, LIST(entity) as entities FROM bean_entities GROUP BY url
) as n ON b.url = n.url
;