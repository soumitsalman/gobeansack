package beansack

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/k0kubun/pp"
	datautils "github.com/soumitsalman/data-utils"
)

const (
	_TIMEOUT        = 5
	_POOL_SIZE      = 32
	_CONN_LIFETIME  = 5
	_CONN_IDLE_TIME = 5
)

const (
	_ORDER_BY_LATEST   = "created DESC"
	_ORDER_BY_TRENDING = "trend_score DESC"
	_ORDER_BY_DISTANCE = "distance ASC"
)

const (
	_TRENDING_BEANS_VIEW = "trending_beans_view"
)

type PGSack struct {
	db *pgxpool.Pool
}

func NewPGSack(ctx context.Context, connString string) *PGSack {
	config, err := pgxpool.ParseConfig(connString)
	NoError(err)

	config.MinConns = 1
	config.MaxConns = _POOL_SIZE
	config.MaxConnLifetime = time.Minute * _CONN_LIFETIME
	config.MaxConnIdleTime = time.Minute * _CONN_IDLE_TIME
	config.HealthCheckPeriod = time.Minute * _CONN_LIFETIME
	config.ConnConfig.ConnectTimeout = time.Minute * _TIMEOUT

	db, err := pgxpool.NewWithConfig(ctx, config)
	NoError(err)
	NoError(db.Ping(ctx)) // Quick health check on startup

	return &PGSack{db: db}
}

func (p *PGSack) QueryLatestBeans(ctx context.Context, conditions Condition, page Pagination, columns []string) ([]Bean, error) {
	items, err := fetchBeans(ctx, p.db, BEANS, conditions, _ORDER_BY_LATEST, page, columns)
	if err != nil {
		return nil, err
	}
	return datautils.Transform(items, func(item *dataRow) Bean { return item.toBean() }), nil
}

func (p *PGSack) QueryTrendingBeans(ctx context.Context, conditions Condition, page Pagination, columns []string) ([]BeanAggregate, error) {
	items, err := fetchBeans(ctx, p.db, _TRENDING_BEANS_VIEW, conditions, _ORDER_BY_TRENDING, page, columns)
	if err != nil {
		return nil, err
	}
	return datautils.Transform(items, func(item *dataRow) BeanAggregate { return item.toBeanAggregate() }), nil
}

// TODO: how to pass in text/tag/keyword based search
func (p *PGSack) QueryPublishers(ctx context.Context, conditions Condition, page Pagination, columns []string) ([]Publisher, error) {
	select_expr := buildPGSelect(PUBLISHERS, columns)
	where_expr, where_params := buildPGWhere(conditions)
	limit_offset_expr, limit_offset_params := buildPGLimitOffset(page)
	query, args := buildPGSQL(select_expr, nil, where_expr, where_params, "", limit_offset_expr, limit_offset_params)
	items, err := fetchAll[dataRow](ctx, p.db, query, args)
	if err != nil {
		return nil, err
	}
	return datautils.Transform(items, func(item *dataRow) Publisher { return item.toPublisher() }), nil
}

func (p *PGSack) QueryChatters(ctx context.Context, conditions Condition, page Pagination, columns []string) ([]Chatter, error) {
	select_expr := buildPGSelect(CHATTERS, columns)
	where_expr, where_params := buildPGWhere(conditions)
	limit_offset_expr, limit_offset_params := buildPGLimitOffset(page)
	query, args := buildPGSQL(select_expr, nil, where_expr, where_params, "", limit_offset_expr, limit_offset_params)
	items, err := fetchAll[dataRow](ctx, p.db, query, args)
	if err != nil {
		return nil, err
	}
	return datautils.Transform(items, func(item *dataRow) Chatter { return item.toChatter() }), nil
}

func (p *PGSack) DistinctCategories(ctx context.Context, page Pagination) ([]string, error) {
	select_expr := "SELECT DISTINCT unnest(categories) AS category FROM beans WHERE categories IS NOT NULL ORDER BY category"
	limit_offset_expr, limit_offset_params := buildPGLimitOffset(page)
	query, args := buildPGSQL(select_expr, nil, "", nil, "", limit_offset_expr, limit_offset_params)
	return fetchAllScalar[string](ctx, p.db, query, args)
}

func (p *PGSack) DistinctSentiments(ctx context.Context, page Pagination) ([]string, error) {
	select_expr := "SELECT DISTINCT unnest(sentiments) AS sentiment FROM beans WHERE sentiments IS NOT NULL ORDER BY sentiment"
	limit_offset_expr, limit_offset_params := buildPGLimitOffset(page)
	query, args := buildPGSQL(select_expr, nil, "", nil, "", limit_offset_expr, limit_offset_params)
	return fetchAllScalar[string](ctx, p.db, query, args)
}

func (p *PGSack) DistinctEntities(ctx context.Context, page Pagination) ([]string, error) {
	select_expr := "SELECT DISTINCT unnest(entities) AS entity FROM beans WHERE entities IS NOT NULL ORDER BY entity"
	limit_offset_expr, limit_offset_params := buildPGLimitOffset(page)
	query, args := buildPGSQL(select_expr, nil, "", nil, "", limit_offset_expr, limit_offset_params)
	return fetchAllScalar[string](ctx, p.db, query, args)
}

func (p *PGSack) DistinctRegions(ctx context.Context, page Pagination) ([]string, error) {
	select_expr := "SELECT DISTINCT unnest(regions) AS region FROM beans WHERE regions IS NOT NULL ORDER BY region"
	limit_offset_expr, limit_offset_params := buildPGLimitOffset(page)
	query, args := buildPGSQL(select_expr, nil, "", nil, "", limit_offset_expr, limit_offset_params)
	return fetchAllScalar[string](ctx, p.db, query, args)
}

func (p *PGSack) DistinctSources(ctx context.Context, page Pagination) ([]string, error) {
	select_expr := "SELECT source FROM publishers ORDER BY source"
	limit_offset_expr, limit_offset_params := buildPGLimitOffset(page)
	query, args := buildPGSQL(select_expr, nil, "", nil, "", limit_offset_expr, limit_offset_params)
	return fetchAllScalar[string](ctx, p.db, query, args)
}

func (p *PGSack) CountRows(ctx context.Context, table string, conditions Condition) (int64, error) {
	select_expr := buildPGSelect(table, []string{"count(*)"})
	where_expr, where_params := buildPGWhere(conditions)
	query, args := buildPGSQL(select_expr, nil, where_expr, where_params, "", "", nil)
	return fetchOneScalar[int64](ctx, p.db, query, args)
}

func (p *PGSack) Close() {
	if p != nil && p.db != nil {
		p.db.Close()
	}
}

func fetchBeans(ctx context.Context, db *pgxpool.Pool, table string, conditions Condition, order string, page Pagination, columns []string) ([]dataRow, error) {
	// TODO: add vector similarity search conditions
	select_expr := buildPGSelect(table, columns)
	where_expr, where_params := buildPGWhere(conditions)
	limit_offset_expr, limit_offset_params := buildPGLimitOffset(page)
	query, args := buildPGSQL(select_expr, nil, where_expr, where_params, buildPGOrderBy(order), limit_offset_expr, limit_offset_params)
	return fetchAll[dataRow](ctx, db, query, args)
}

// SQL query string builder utilities
// TODO: add function for building query
func (p *PGSack) buildSQL(table string, conditions Condition, order string, page Pagination, columns []string) (string, pgx.NamedArgs) {
	select_expr := buildPGSelect(table, columns)
	where_expr, where_params := buildPGWhere(conditions)
	limit_offset_expr, limit_offset_params := buildPGLimitOffset(page)
	return buildPGSQL(select_expr, nil, where_expr, where_params, buildPGOrderBy(order), limit_offset_expr, limit_offset_params)
}

// TODO: merge this with buildSQL
func buildPGSQL(select_expr string, select_params pgx.NamedArgs, where_expr string, where_params pgx.NamedArgs, order_expr string, limit_offset_expr string, limit_offset_params pgx.NamedArgs) (string, pgx.NamedArgs) {
	expr_builder := strings.Builder{}
	expr_builder.WriteString(select_expr)
	if where_expr != "" {
		expr_builder.WriteString(" ")
		expr_builder.WriteString(where_expr)
	}
	if order_expr != "" {
		expr_builder.WriteString(" ")
		expr_builder.WriteString(order_expr)
	}
	if limit_offset_expr != "" {
		expr_builder.WriteString(" ")
		expr_builder.WriteString(limit_offset_expr)
	}
	merged_params := pgx.NamedArgs{}
	for _, m := range []pgx.NamedArgs{select_params, where_params, limit_offset_params} {
		for k, v := range m {
			merged_params[k] = v
		}
	}
	expr := expr_builder.String()
	pp.Println(expr, merged_params)
	return expr, merged_params
}

// NOTE: Can cause sql injection if not used carefully. Only use with trusted input or after proper sanitization.
func buildPGSelect(table string, columns []string) string {
	fields := "*"
	if len(columns) > 0 {
		fields = strings.Join(columns, ", ")
	}
	return fmt.Sprintf("SELECT %s FROM %s", fields, table)
}

func buildPGWhere(conditions Condition) (string, pgx.NamedArgs) {
	parts := make([]string, 0, 10) // preallocate for expected conditions
	args := pgx.NamedArgs{}

	if len(conditions.Urls) > 0 {
		parts = append(parts, "url = ANY(@urls)")
		args["urls"] = conditions.Urls // pgx handles []string as array automatically
	}

	if conditions.Kind != "" {
		parts = append(parts, "kind = @kind")
		args["kind"] = conditions.Kind
	}

	if !conditions.Created.IsZero() {
		parts = append(parts, "created >= @created_from")
		args["created_from"] = conditions.Created
	}

	if !conditions.Collected.IsZero() {
		parts = append(parts, "collected >= @collected_from")
		args["collected_from"] = conditions.Collected
	}

	if !conditions.Updated.IsZero() {
		parts = append(parts, "updated >= @updated_from")
		args["updated_from"] = conditions.Updated
	}

	if len(conditions.Categories) > 0 {
		parts = append(parts, "categories && @categories")
		args["categories"] = conditions.Categories
	}

	if len(conditions.Regions) > 0 {
		parts = append(parts, "regions && @regions")
		args["regions"] = conditions.Regions
	}

	if len(conditions.Entities) > 0 {
		parts = append(parts, "entities && @entities")
		args["entities"] = conditions.Entities
	}

	if len(conditions.Tags) > 0 {
		parts = append(parts, "tags @@ plainto_tsquery('simple', @tags_query)")
		args["tags_query"] = strings.Join(conditions.Tags, " & ") // "tag1 & tag2 & tag3"
	}

	if len(conditions.Sources) > 0 {
		parts = append(parts, "source = ANY(@sources)")
		args["sources"] = conditions.Sources
	}

	if len(conditions.Extra) > 0 {
		parts = append(parts, conditions.Extra...)
	}

	if len(parts) == 0 {
		return "", nil
	}
	return fmt.Sprintf("WHERE %s", strings.Join(parts, " AND ")), args
}

func buildPGOrderBy(order ...string) string {
	if len(order) == 0 {
		return ""
	}
	return "ORDER BY " + strings.Join(order, ", ")
}

func buildPGLimitOffset(page Pagination) (string, pgx.NamedArgs) {
	parts := make([]string, 0, 2)
	args := pgx.NamedArgs{}

	if page.Limit > 0 {
		parts = append(parts, "LIMIT @limit")
		args["limit"] = page.Limit
	}
	if page.Offset > 0 {
		parts = append(parts, "OFFSET @offset")
		args["offset"] = page.Offset
	}

	if len(parts) == 0 {
		return "", nil
	}
	return strings.Join(parts, " "), args
}

// PGX fetch helpers
func fetchOne[T any](ctx context.Context, db *pgxpool.Pool, query string, args ...any) (T, error) {
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		var zero T
		return zero, err
	}
	defer rows.Close()
	return pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[T])
}

func fetchOneScalar[T any](ctx context.Context, db *pgxpool.Pool, query string, args ...any) (T, error) {
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		var zero T
		return zero, err
	}
	defer rows.Close()
	return pgx.CollectOneRow(rows, pgx.RowTo[T])
}

func fetchAll[T any](ctx context.Context, db *pgxpool.Pool, query string, args ...any) ([]T, error) {
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByNameLax[T])
}

func fetchAllScalar[T any](ctx context.Context, db *pgxpool.Pool, query string, args ...any) ([]T, error) {
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowTo[T])
}

// PGX marshalling and unmarshalling for custom types

type dataRow struct {
	URL         sql.NullString `db:"url"`
	Kind        sql.NullString `db:"kind"`
	Title       sql.NullString `db:"title"`
	Summary     sql.NullString `db:"summary"`
	Content     sql.NullString `db:"content"`
	Author      sql.NullString `db:"author"`
	Source      sql.NullString `db:"source"`
	ImageUrl    sql.NullString `db:"image_url"`
	Created     sql.NullTime   `db:"created"`
	Embedding   []float32      `db:"embedding"`
	Gist        sql.NullString `db:"gist"`
	Categories  []string       `db:"categories"`
	Sentiments  []string       `db:"sentiments"`
	Regions     []string       `db:"regions"`
	Entities    []string       `db:"entities"`
	Related     []string       `db:"related"`
	ClusterId   sql.NullString `db:"cluster_id"`
	ClusterSize int            `db:"cluster_size"`
	Updated     sql.NullTime   `db:"updated"`
	Likes       int            `db:"likes"`
	Comments    int            `db:"comments"`
	Subscribers int            `db:"subscribers"`
	Shares      int            `db:"shares"`
	Distance    float64        `db:"distance"`
	TrendScore  float64        `db:"trend_score"`
	ChatterURL  sql.NullString `db:"chatter_url"`
	Forum       sql.NullString `db:"forum"`
	Collected   sql.NullTime   `db:"collected"`
	BaseURL     sql.NullString `db:"base_url"`
	SiteName    sql.NullString `db:"site_name"`
	Description sql.NullString `db:"description"`
	Favicon     sql.NullString `db:"favicon"`
	RSSFeed     sql.NullString `db:"rss_feed"`
}

// Conversion methods from dataRow to public types
func (r *dataRow) toBean() Bean {
	return Bean{
		URL:        nullStringToString(r.URL),
		Kind:       nullStringToString(r.Kind),
		Title:      nullStringToString(r.Title),
		Summary:    nullStringToString(r.Summary),
		Content:    nullStringToString(r.Content),
		Author:     nullStringToString(r.Author),
		Source:     nullStringToString(r.Source),
		ImageUrl:   nullStringToString(r.ImageUrl),
		Created:    nullTimeToTime(r.Created),
		Embedding:  r.Embedding,
		Gist:       nullStringToString(r.Gist),
		Categories: r.Categories,
		Sentiments: r.Sentiments,
		Regions:    r.Regions,
		Entities:   r.Entities,
	}
}

func (r *dataRow) toBeanAggregate() BeanAggregate {
	bean := r.toBean()
	return BeanAggregate{
		Bean:        bean,
		Related:     r.Related,
		ClusterId:   nullStringToString(r.ClusterId),
		ClusterSize: r.ClusterSize,
		Updated:     nullTimeToTime(r.Updated),
		Likes:       r.Likes,
		Comments:    r.Comments,
		Subscribers: r.Subscribers,
		Shares:      r.Shares,
		Distance:    r.Distance,
		TrendScore:  r.TrendScore,
	}
}

func (r *dataRow) toPublisher() Publisher {
	return Publisher{
		Source:      nullStringToString(r.Source),
		BaseURL:     nullStringToString(r.BaseURL),
		SiteName:    nullStringToString(r.SiteName),
		Description: nullStringToString(r.Description),
		Favicon:     nullStringToString(r.Favicon),
		RSSFeed:     nullStringToString(r.RSSFeed),
		Collected:   nullTimeToTime(r.Collected),
	}
}

func (r *dataRow) toChatter() Chatter {
	return Chatter{
		ChatterURL:  nullStringToString(r.ChatterURL),
		URL:         nullStringToString(r.URL),
		Source:      nullStringToString(r.Source),
		Forum:       nullStringToString(r.Forum),
		Collected:   nullTimeToTime(r.Collected),
		Likes:       r.Likes,
		Comments:    r.Comments,
		Subscribers: r.Subscribers,
	}
}

func (r *dataRow) toBeanChatter() BeanChatter {
	return BeanChatter{
		URL:         nullStringToString(r.URL),
		Collected:   nullTimeToTime(r.Collected),
		Likes:       r.Likes,
		Comments:    r.Comments,
		Subscribers: r.Subscribers,
		Shares:      r.Shares,
	}
}
