WITH 
filtered_chatters AS (
    SELECT * FROM chatters WHERE bean_url IN ('https://www.rawstory.com/trump-deporting-us-citizens/', 'https://www.theguardian.com/technology/2025/jun/27/deepfakes-denmark-copyright-law-artificial-intelligence')
),
current_per_id AS (
    SELECT
        id,
        FIRST(bean_url) as bean_url,
        MAX(collected) as collected,
        MAX(likes) as likes,
        MAX(comments) as comments,
        MAX(subscribers) as subscribers
    FROM filtered_chatters
    GROUP BY id
),
before_per_id AS (
    SELECT
        id,
        FIRST(bean_url) as bean_url,
        MAX(collected) as collected,
        MAX(likes) as likes,
        MAX(comments) as comments,
        MAX(subscribers) as subscribers
    FROM filtered_chatters        
    GROUP BY id
    HAVING collected < CURRENT_TIMESTAMP - INTERVAL 3 DAY
),
current_agg AS (
    SELECT
        bean_url,
        MAX(collected) as collected,
        SUM(likes) as likes,
        SUM(comments) as comments,
        SUM(subscribers) as subscribers,
        COUNT(id) as shares
    FROM current_per_id
    GROUP BY bean_url
),
before_agg AS (
    SELECT
        bean_url,
        MAX(collected) as collected,
        SUM(likes) as likes,
        SUM(comments) as comments,
        SUM(subscribers) as subscribers,
        COUNT(id) as shares
    FROM before_per_id
    GROUP BY bean_url
)
SELECT
	ca.bean_url,
	ca.collected,
	COALESCE(ca.likes, 0) - COALESCE(ba.likes, 0) as likes,
	COALESCE(ca.comments, 0) - COALESCE(ba.comments, 0) as comments,
	COALESCE(ca.subscribers, 0) - COALESCE(ba.subscribers, 0) as subscribers,
	COALESCE(ca.shares, 0) - COALESCE(ba.shares, 0) as shares
FROM current_agg ca
LEFT JOIN before_agg ba
ON ca.bean_url = ba.bean_url
