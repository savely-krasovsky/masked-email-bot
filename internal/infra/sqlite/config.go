package sqlite

type Config struct {
	DBFile              string `env:"SQLITE_DB_FILE,default=file:masked_email_bot.db?cache=shared&mode=rwc"`
	Name                string `env:"SQLITE_NAME,default=masked_email_bot"`
	MigrationsSourceURL string `env:"SQLITE_MIGRATIONS_SOURCE_URL,default=file://migrations"`
}
