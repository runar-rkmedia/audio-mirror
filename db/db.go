package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/extra/bundebug"
)

type DBOptions struct {
	InMemory bool
	FilePath string
}

type DB struct {
	DB *bun.DB
}

func CreateDatabase(options DBOptions) (*DB, error) {
	if options.InMemory {
		options.FilePath = "file::memory:?cache=shared"
	}
	sqldb, err := sql.Open(sqliteshim.ShimName, options.FilePath)
	if err != nil {
		return nil, err
	}
	bundDB := bun.NewDB(sqldb, sqlitedialect.New())
	bundDB.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
		bundebug.FromEnv("BUNDEBUG"),
	))
	db := &DB{bundDB}
	err = db.createTables()

	return db, err
}

func (db DB) createTables() error {
	ctx := context.TODO()
	if _, err := db.DB.NewCreateTable().Model((*Channel)(nil)).IfNotExists().Exec(ctx); err != nil {
		return fmt.Errorf("failed to create table channels: %w", err)
	}
	if _, err := db.DB.NewCreateTable().Model((*Episode)(nil)).IfNotExists().Exec(ctx); err != nil {
		return fmt.Errorf("failed to create table episode: %w", err)
	}

	return nil
}

func (db DB) GetChannels(ctx context.Context) ([]Channel, error) {
	var channels []Channel
	err := db.DB.NewSelect().Model(&channels).OrderExpr("id ASC").Limit(10).Scan(ctx)
	if err != nil {
		return channels, fmt.Errorf("failed to retrieve channels %w", err)
	}
	return channels, err
}

func (db DB) CreateChannel(ctx context.Context, channel Channel) (Channel, error) {
	var out Channel
	_, err := db.DB.NewInsert().Model(channel).Exec(ctx, &out)
	if err != nil {
		return out, fmt.Errorf("failed to create channel %w", err)
	}
	return out, err
}

type Channel struct {
	bun.BaseModel     `bun:"table:channel,alias:c"`
	CreatedAt         time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt         time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	ID                string    `bun:",pk"`
	Title             string    `bun:",notnull"`
	Description       string    `bun:",nullzero"`
	Type              string    `bun:",notnull"`
	LastEpisodeDate   *time.Time
	Frequency_seconds uint64
	// The original source, often a rss-feed or a json
	Source []byte
}
type Episode struct {
	bun.BaseModel `bun:"table:episodes,alias:e"`
	ID            string `bun:",pk"`
	ChannelID     string
	Channel       *Channel `bun:"rel:belongs-to,join:channel_id=id"`
	Title         string
	Description   string
}
