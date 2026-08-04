// При Написание configs использовался pattern Builder(Строитель/функц. опции)
package config

import (
	"flag"
	"os"
)

const (
	ModeDev   = "dev"
	ModeDebug = "debug"
	ModeProd  = "prod"
)

// Options хранит настройки приложения, собранные из флагов командной строки,
// переменных окружения и значений по умолчанию.
// generate:reset
type Options struct {
	RunAddr         string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
	JWTSecretKey    string
	AuditFile       string
	AuditURL        string
	Mode            string
	EnableHTTPS     string
}

// NewOptions создаёт Options со значениями по умолчанию и применяет
// переданные опции (pattern Builder / функциональные опции).
//
// Пример использования:
//
//	cfg := config.NewOptions(
//	    config.WithBaseURL("http://localhost:8080"),
//	    config.WithFileStoragePath("storage.json"),
//	)
func NewOptions(opts ...func(*Options)) *Options {
	o := &Options{
		RunAddr:      "localhost:8080",
		BaseURL:      "http://localhost:8080",
		JWTSecretKey: "jwt_secret_key",
		Mode:         ModeDev,
	}

	for _, opt := range opts {
		opt(o)
	}
	return o
}

// WithRunAddr задаёт адрес и порт, на которых запускается сервер.
func WithRunAddr(v string) func(*Options) { return func(o *Options) { o.RunAddr = v } }

// WithBaseURL задаёт базовый URL, используемый при формировании коротких ссылок.
func WithBaseURL(v string) func(*Options) { return func(o *Options) { o.BaseURL = v } }

// WithFileStoragePath задаёт путь к файлу для файлового хранилища.
func WithFileStoragePath(v string) func(*Options) { return func(o *Options) { o.FileStoragePath = v } }

// WithDatabaseDSN задаёт строку подключения к Postgres.
func WithDatabaseDSN(v string) func(*Options) { return func(o *Options) { o.DatabaseDSN = v } }

// WithJWTSecretKey задаёт секретный ключ для подписи JWT-токенов.
func WithJWTSecretKey(v string) func(*Options) { return func(o *Options) { o.JWTSecretKey = v } }

// WithAuditFile задаёт путь к файлу-приёмнику аудита.
func WithAuditFile(v string) func(*Options) { return func(o *Options) { o.AuditFile = v } }

// WithAuditURL задаёт URL удалённого сервера-приёмника аудита.
func WithAuditURL(v string) func(*Options) { return func(o *Options) { o.AuditURL = v } }

// WithMode задаёт режим работы приложения (например, "dev" или "debug").
func WithMode(v string) func(*Options) { return func(o *Options) { o.Mode = v } }

// WithMode задаёт режим работы приложения (например, "dev" или "debug").
func WithEnableHTTPS(v string) func(*Options) { return func(o *Options) { o.EnableHTTPS = v } }

// Build собирает Options из флагов командной строки и переменных окружения.
// Флаг имеет приоритет, если задан явно; иначе используется переменная
// окружения; иначе — значение по умолчанию.
func Build() *Options {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	var raw Options
	fs.StringVar(&raw.RunAddr, "a", "localhost:8080", "address and port to run server")
	fs.StringVar(&raw.BaseURL, "b", "http://localhost:8080", "Base URL")
	fs.StringVar(&raw.FileStoragePath, "f", "", "file storage")
	fs.StringVar(&raw.DatabaseDSN, "d", "", "connection string to DB")
	fs.StringVar(&raw.AuditFile, "audit-file", "", "путь к файлу-приёмнику")
	fs.StringVar(&raw.AuditURL, "audit-url", "", "URL удаленного сервера-приёмника")
	fs.StringVar(&raw.EnableHTTPS, "s", "", "Включение HTTPS")

	if err := fs.Parse(os.Args[1:]); err != nil {
		panic(err)
	}

	explicit := map[string]bool{}
	fs.Visit(func(f *flag.Flag) {
		explicit[f.Name] = true
	})

	pick := func(flagName string, flagVal string, envStr string) string {
		if explicit[flagName] {
			return flagVal
		}
		if v := os.Getenv(envStr); v != "" {
			return v
		}
		return flagVal
	}

	filePath := pick("f", raw.FileStoragePath, "FILE_STORAGE_PATH")

	dsn := pick("d", raw.DatabaseDSN, "DATABASE_CONN_STRING")

	if filePath != "" && !explicit["d"] {
		dsn = ""
	}

	return NewOptions(
		WithRunAddr(pick("a", raw.RunAddr, "SERVER_ADDRESS")),
		WithBaseURL(pick("b", raw.BaseURL, "BASE_URL")),
		WithFileStoragePath(filePath),
		WithDatabaseDSN(dsn),
		WithAuditFile(pick("audit-file", raw.AuditFile, "AUDIT_FILE")),
		WithAuditURL(pick("audit-url", raw.AuditURL, "AUDIT_URL")),
		WithEnableHTTPS(pick("s", raw.EnableHTTPS, "ENABLE_HTTPS")),
	)
}
