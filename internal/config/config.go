// При Написание configs использовался pattern Builder(Строитель/функц. опции)
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
)

const (
	ModeDev   = "dev"
	ModeDebug = "debug"
	ModeProd  = "prod"
)

// Options хранит настройки приложения, собранные из флагов(высокий приоритет) командной строки,
// переменных окружения(средний приоритет), файла конфигурации(низкий приоритет)
// и значений по умолчанию.
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
	EnableHTTPS     bool
	ConfigFile      string
}

// FileConfig настройки приложения специально под JSON,
// имеет приоритет (ниже флагов и env).
type FileConfig struct {
	RunAddr         string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
}

// NewOptions создаёт Options со значениями по умолчанию и применяет
// переданные опции (pattern Builder / функциональные опции).
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

// Build собирает Options из флагов командной строки, переменных окружения
// и файла конфигурации. Приоритет: явный флаг > env > файл конфига > дефолт.
func Build() (*Options, error) {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	customFlagUsage(fs)

	var raw Options
	fs.StringVar(&raw.RunAddr, "a", "localhost:8080", "адрес и порт запуска HTTP-сервера, например: -a=localhost:8080")
	fs.StringVar(&raw.BaseURL, "b", "http://localhost:8080", "базовый адрес результирующего сокращённого URL, например: -b=http://localhost:8080")
	fs.StringVar(&raw.FileStoragePath, "f", "", "путь к файлу для хранения данных в формате JSON, например: -f=/tmp/short-url-db.json")
	fs.StringVar(&raw.DatabaseDSN, "d", "", "строка подключения к БД, например: -d=\"postgres://user:pass@localhost:5432/shortener\"")
	fs.StringVar(&raw.AuditFile, "audit-file", "", "путь к файлу-приёмнику аудита, например: -audit-file=audit.log")
	fs.StringVar(&raw.AuditURL, "audit-url", "", "URL удалённого сервера-приёмника аудита, например: -audit-url=http://localhost:9000/audit")
	fs.BoolVar(&raw.EnableHTTPS, "s", false, "включить HTTPS (true/false), например: -s=true | 1")
	fs.StringVar(&raw.ConfigFile, "c", "", "путь к файлу конфигурации в формате JSON, например: -c=config.json")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	explicit := map[string]bool{}
	fs.Visit(func(f *flag.Flag) {
		explicit[f.Name] = true
	})

	configPath := pickString(explicit["c"], "CONFIG", raw.ConfigFile, "")

	var fc FileConfig
	if configPath != "" {
		parsed, err := parseConfigFile(configPath)
		if err != nil {
			slog.Error("parse config file", "path", configPath, "error", err)
		} else {
			fc = *parsed
		}
	}

	opts := NewOptions()

	opts.RunAddr = pickString(explicit["a"], "SERVER_ADDRESS", raw.RunAddr, fc.RunAddr)
	opts.BaseURL = pickString(explicit["b"], "BASE_URL", raw.BaseURL, fc.BaseURL)
	opts.AuditFile = pickString(explicit["audit-file"], "AUDIT_FILE", raw.AuditFile, "")
	opts.AuditURL = pickString(explicit["audit-url"], "AUDIT_URL", raw.AuditURL, "")
	opts.EnableHTTPS = pickBool(explicit["s"], "ENABLE_HTTPS", raw.EnableHTTPS, fc.EnableHTTPS)
	opts.ConfigFile = configPath

	filePath := pickString(explicit["f"], "FILE_STORAGE_PATH", raw.FileStoragePath, fc.FileStoragePath)
	dsn := pickString(explicit["d"], "DATABASE_CONN_STRING", raw.DatabaseDSN, fc.DatabaseDSN)
	if filePath != "" && !explicit["d"] {
		dsn = ""
	}
	opts.FileStoragePath = filePath
	opts.DatabaseDSN = dsn

	return opts, nil
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

// WithEnableHTTPS задаёт режим работы протокола (http | https).
func WithEnableHTTPS(v bool) func(*Options) { return func(o *Options) { o.EnableHTTPS = v } }

// WithConfigFile задаёт конфигурацию приложения с помощью файла config.JSON.
func WithConfigFile(v string) func(*Options) { return func(o *Options) { o.ConfigFile = v } }

func pickString(explicitFlag bool, envStr string, flagVal string, fileVal string) string {
	if explicitFlag {
		return flagVal
	}
	if v := os.Getenv(envStr); v != "" {
		return v
	}
	if fileVal != "" {
		return fileVal
	}
	return flagVal
}

func pickBool(explicitFlag bool, envStr string, flagVal bool, fileVal bool) bool {
	if explicitFlag {
		return flagVal
	}
	if v := os.Getenv(envStr); v != "" {
		return v == "true" || v == "1"
	}
	return fileVal || flagVal
}

func parseConfigFile(fname string) (*FileConfig, error) {
	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var fc FileConfig
	if err := json.NewDecoder(file).Decode(&fc); err != nil {
		return nil, err
	}
	return &fc, nil
}

func customFlagUsage(fs *flag.FlagSet) {
	fmt.Fprintf(fs.Output(), `Использование: %s [флаги]

Приоритет источников конфигурации (от высокого к низкому):
  1. Флаг командной строки (если передан явно)
  2. Переменная окружения
  3. Файл конфигурации (JSON, путь через -c/-config или CONFIG)
  4. Значение по умолчанию

Флаги:
`, os.Args[0])
	fs.PrintDefaults()
	fmt.Fprintf(fs.Output(), `
Переменные окружения: SERVER_ADDRESS, BASE_URL, FILE_STORAGE_PATH,
  DATABASE_DSN, AUDIT_FILE, AUDIT_URL, ENABLE_HTTPS, CONFIG

Примеры:
  %s -a=localhost:8080 -b=http://localhost:8080
  %s -c=config.json
  SERVER_ADDRESS=localhost:9090 %s
`, os.Args[0], os.Args[0], os.Args[0])
}
