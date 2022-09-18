package main

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

type config struct {
	BindAddr             string        `json:"bindAddr"`
	Port                 int           `json:"port"`
	StreamURLaddr        string        `json:"streamURLaddr"`
	StoragePath          string        `json:"storagePath"`
	MaxAgeTorrents       time.Duration `json:"maxAgeTorrents"`
	CachePath            string        `json:"cachePath"`
	CacheAgeXD           time.Duration `json:"cacheAgeXD"`
	RedisAddr            string        `json:"redisAddr"`
	RedisCreds           string        `json:"redisCreds"`
	BaseURLyts           string        `json:"baseURLyts"`
	BaseURLtpb           string        `json:"baseURLtpb"`
	BaseURL1337x         string        `json:"baseURL1337x"`
	BaseURLibit          string        `json:"baseURLibit"`
	BaseURLrarbg         string        `json:"baseURLrarbg"`
	BaseURLrd            string        `json:"baseURLrd"`
	BaseURLad            string        `json:"baseURLad"`
	BaseURLpm            string        `json:"baseURLpm"`
	LogLevel             string        `json:"logLevel"`
	LogFoundTorrents     bool          `json:"logFoundTorrents"`
	RootURL              string        `json:"rootURL"`
	ExtraHeadersXD       []string      `json:"extraHeadersXD"`
	SocksProxyAddrTPB    string        `json:"socksProxyAddrTPB"`
	WebConfigurePath     string        `json:"webConfigurePath"`
	IMDB2metaAddr        string        `json:"imdb2metaAddr"`
	UseOAUTH2            bool          `json:"useOAUTH2"`
	OAUTH2authorizeURLpm string        `json:"oauth2authURLpm"`
	OAUTH2tokenURLpm     string        `json:"oauth2tokenURLpm"`
	OAUTH2clientIDpm     string        `json:"oauth2clientIDpm"`
	OAUTH2clientSecretPM string        `json:"oauth2clientSecretPM"`
	OAUTH2encryptionKey  string        `json:"oauth2encryptionKey"`
	EnvPrefix            string        `json:"envPrefix"`
}

func parseConfig(logger *zap.Logger) config {
	result := config{}

	// Flags
	var (
		bindAddr             = flag.String("bindAddr", "https://deflix-egd-stremio.herokuapp.com/", `Local interface address to bind to. "localhost" only allows access from the local host. "0.0.0.0" binds to all network interfaces.`)
		port                 = flag.Int("port", 443, "Port to listen on")
		streamURLaddr        = flag.String("streamURLaddr", "https://deflix-egd-stremio.herokuapp.com/:442", "Address to be used in a stream URL that's delivered to Stremio and later used to redirect to , AllDebrid and Premiumize. If you enable OAuth2 handling this will also be used to determine whether the state cookie is a secure one or not.")
		storagePath          = flag.String("storagePath", "", `Path for storing the data of the persistent DB which stores torrent results. An empty value will lead to 'os.UserCacheDir()+"/deflix-stremio/badger"'.`)
		maxAgeTorrents       = flag.Duration("maxAgeTorrents", 7*24*time.Hour, "Max age of cache entries for torrents found per IMDb ID. The format must be acceptable by Go's 'time.ParseDuration()', for example \"24h\". Default is 7 days.")
		cachePath            = flag.String("cachePath", "", `Path for loading persisted caches on startup and persisting the current cache in regular intervals. An empty value will lead to 'os.UserCacheDir()+"/deflix-stremio/cache"'.`)
		cacheAgeXD           = flag.Duration("cacheAgeXD", 24*time.Hour, "Max age of cache entries for instant availability responses from RealDebrid, AllDebrid and Premiumize. The format must be acceptable by Go's 'time.ParseDuration()', for example \"24h\".")
		redisAddr            = flag.String("redisAddr", "", `Redis host and port, for example "localhost:6379". It's used for the redirect and stream cache. Keep empty to use in-memory go-cache.`)
		redisCreds           = flag.String("redisCreds", "", `Credentials for Redis. Password for Redis version 5 and older, username and password for Redis version 6 and newer. Use the colon character (":") for separating username and password. This implies you can't use a colon in the password when using Redis version 5 or older.`)
		baseURLyts           = flag.String("baseURLyts", "https://yts.mx", "Base URL for YTS")
		baseURLtpb           = flag.String("baseURLtpb", "https://apibay.org", "Base URL for the TPB API")
		baseURL1337x         = flag.String("baseURL1337x", "https://1337x.to", "Base URL for 1337x")
		baseURLibit          = flag.String("baseURLibit", "https://ibit.am", "Base URL for ibit")
		baseURLrarbg         = flag.String("baseURLrarbg", "https://torrentapi.org", "Base URL for RARBG")
		baseURLrd            = flag.String("baseURLrd", "https://api.real-debrid.com", "Base URL for RealDebrid")
		baseURLad            = flag.String("baseURLad", "https://api.alldebrid.com", "Base URL for AllDebrid")
		baseURLpm            = flag.String("baseURLpm", "https://www.premiumize.me/api", "Base URL for Premiumize")
		logLevel             = flag.String("logLevel", "error", `Log level to show only logs with the given and more severe levels. Can be "debug", "info", "warn", "error".`)
		logFoundTorrents     = flag.Bool("logFoundTorrents", true, "Set to true to log each single torrent that was found by one of the torrent site clients (with DEBUG level)")
		rootURL              = flag.String("rootURL", "https://www.deflix.tv", "Redirect target for the root")
		extraHeadersXD       = flag.String("extraHeadersXD", "", `Additional HTTP request headers to set for requests to RealDebrid, AllDebrid and Premiumize, in a format like "X-Foo: bar", separated by newline characters ("\n")`)
		socksProxyAddrTPB    = flag.String("socksProxyAddrTPB", "", "SOCKS5 proxy address for accessing TPB, required for accessing TPB via the TOR network (where \"127.0.0.1:9050\" would be typical value)")
		webConfigurePath     = flag.String("webConfigurePath", "", "Path to the directory with web files for the '/configure' endpoint. If empty, files compiled into the binary will be used")
		imdb2metaAddr        = flag.String("imdb2metaAddr", "", "Address of the imdb2meta gRPC server. Won't be used if empty.")
		useOAUTH2            = flag.Bool("useOAUTH2", false, "Flag for indicating whether to use OAuth2 for Premiumize authorization. This leads to a different configuration webpage that doesn't require API keys. It requires a client ID to be configured.")
		oauth2authURLpm      = flag.String("oauth2authURLpm", "https://www.premiumize.me/authorize", "URL of the OAuth2 authorization endpoint of Premiumize")
		oauth2tokenURLpm     = flag.String("oauth2tokenURLpm", "https://www.premiumize.me/token", "URL of the OAuth2 token endpoint of Premiumize")
		oauth2clientIDpm     = flag.String("oauth2clientIDpm", "", "Client ID for deflix-stremio on Premiumize")
		oauth2clientSecretPM = flag.String("oauth2clientSecretPM", "", "Client secret for deflix-stremio on Premiumize")
		oauth2encryptionKey  = flag.String("oauth2encryptionKey", "", "OAuth2 data encryption key")
		envPrefix            = flag.String("envPrefix", "", "Prefix for environment variables")
	)

	flag.Parse()

	if *envPrefix != "" && !strings.HasSuffix(*envPrefix, "_") {
		*envPrefix += "_"
	}
	result.EnvPrefix = *envPrefix

	// Only overwrite the values by their env var counterparts that have not been set (and that *are* set via env var).
	var err error
	if !isArgSet("bindAddr") {
		if val, ok := os.LookupEnv(*envPrefix + "BIND_ADDR"); ok {
			*bindAddr = val
		}
	}
	result.BindAddr = *bindAddr

	if !isArgSet("port") {
		if val, ok := os.LookupEnv(*envPrefix + "PORT"); ok {
			if *port, err = strconv.Atoi(val); err != nil {
				logger.Fatal("Couldn't convert environment variable from string to int", zap.Error(err), zap.String("envVar", "PORT"))
			}
		}
	}
	result.Port = *port

	if !isArgSet("streamURLaddr") {
		if val, ok := os.LookupEnv(*envPrefix + "STREAM_URL_ADDR"); ok {
			*streamURLaddr = val
		}
	}
	result.StreamURLaddr = *streamURLaddr

	if !isArgSet("storagePath") {
		if val, ok := os.LookupEnv(*envPrefix + "STORAGE_PATH"); ok {
			*storagePath = val
		}
	}
	result.StoragePath = *storagePath

	if !isArgSet("maxAgeTorrents") {
		if val, ok := os.LookupEnv(*envPrefix + "MAX_AGE_TORRENTS"); ok {
			if *maxAgeTorrents, err = time.ParseDuration(val); err != nil {
				logger.Fatal("Couldn't convert environment variable from string to time.Duration", zap.Error(err), zap.String("envVar", "CACHE_AGE_TORRENTS"))
			}
		}
	}
	result.MaxAgeTorrents = *maxAgeTorrents

	if !isArgSet("cachePath") {
		if val, ok := os.LookupEnv(*envPrefix + "CACHE_PATH"); ok {
			*cachePath = val
		}
	}
	result.CachePath = *cachePath

	if !isArgSet("cacheAgeXD") {
		if val, ok := os.LookupEnv(*envPrefix + "CACHE_AGE_XD"); ok {
			if *cacheAgeXD, err = time.ParseDuration(val); err != nil {
				logger.Fatal("Couldn't convert environment variable from string to time.Duration", zap.Error(err), zap.String("envVar", "CACHE_AGE_XD"))
			}
		}
	}
	result.CacheAgeXD = *cacheAgeXD

	if !isArgSet("redisAddr") {
		if val, ok := os.LookupEnv(*envPrefix + "REDIS_ADDR"); ok {
			*redisAddr = val
		}
	}
	result.RedisAddr = *redisAddr

	if !isArgSet("redisCreds") {
		if val, ok := os.LookupEnv(*envPrefix + "REDIS_CREDS"); ok {
			*redisCreds = val
		}
	}
	result.RedisCreds = *redisCreds

	if !isArgSet("baseURLyts") {
		if val, ok := os.LookupEnv(*envPrefix + "BASE_URL_YTS"); ok {
			*baseURLyts = val
		}
	}
	result.BaseURLyts = *baseURLyts

	if !isArgSet("baseURLtpb") {
		if val, ok := os.LookupEnv(*envPrefix + "BASE_URL_TPB"); ok {
			*baseURLtpb = val
		}
	}
	result.BaseURLtpb = *baseURLtpb

	if !isArgSet("baseURL1337x") {
		if val, ok := os.LookupEnv(*envPrefix + "BASE_URL_1337X"); ok {
			*baseURL1337x = val
		}
	}
	result.BaseURL1337x = *baseURL1337x

	if !isArgSet("baseURLibit") {
		if val, ok := os.LookupEnv(*envPrefix + "BASE_URL_IBIT"); ok {
			*baseURLibit = val
		}
	}
	result.BaseURLibit = *baseURLibit

	if !isArgSet("baseURLrarbg") {
		if val, ok := os.LookupEnv(*envPrefix + "BASE_URL_RARBG"); ok {
			*baseURLrarbg = val
		}
	}
	result.BaseURLrarbg = *baseURLrarbg

	if !isArgSet("baseURLrd") {
		if val, ok := os.LookupEnv(*envPrefix + "BASE_URL_RD"); ok {
			*baseURLrd = val
		}
	}
	result.BaseURLrd = *baseURLrd

	if !isArgSet("baseURLad") {
		if val, ok := os.LookupEnv(*envPrefix + "BASE_URL_AD"); ok {
			*baseURLad = val
		}
	}
	result.BaseURLad = *baseURLad

	if !isArgSet("baseURLpm") {
		if val, ok := os.LookupEnv(*envPrefix + "BASE_URL_PM"); ok {
			*baseURLpm = val
		}
	}
	result.BaseURLpm = *baseURLpm

	if !isArgSet("logLevel") {
		if val, ok := os.LookupEnv(*envPrefix + "LOG_LEVEL"); ok {
			*logLevel = val
		}
	}
	result.LogLevel = *logLevel

	if !isArgSet("logFoundTorrents") {
		if val, ok := os.LookupEnv(*envPrefix + "LOG_FOUND_TORRENTS"); ok {
			if *logFoundTorrents, err = strconv.ParseBool(val); err != nil {
				logger.Fatal("Couldn't convert environment variable from string to bool", zap.Error(err), zap.String("envVar", "LOG_FOUND_TORRENTS"))
			}
		}
	}
	result.LogFoundTorrents = *logFoundTorrents

	if !isArgSet("rootURL") {
		if val, ok := os.LookupEnv(*envPrefix + "ROOT_URL"); ok {
			*rootURL = val
		}
	}
	result.RootURL = *rootURL

	if !isArgSet("extraHeadersRD") {
		if val, ok := os.LookupEnv(*envPrefix + "EXTRA_HEADERS_RD"); ok {
			*extraHeadersXD = val
		}
	}
	if *extraHeadersXD != "" {
		headers := strings.Split(*extraHeadersXD, "\n")
		for _, header := range headers {
			header = strings.TrimSpace(header)
			if header != "" {
				result.ExtraHeadersXD = append(result.ExtraHeadersXD, header)
			}
		}
	}

	if !isArgSet("socksProxyAddrTPB") {
		if val, ok := os.LookupEnv(*envPrefix + "SOCKS_PROXY_ADDR_TPB"); ok {
			*socksProxyAddrTPB = val
		}
	}
	result.SocksProxyAddrTPB = *socksProxyAddrTPB

	if !isArgSet("webConfigurePath") {
		if val, ok := os.LookupEnv(*envPrefix + "WEB_CONFIGURE_PATH"); ok {
			*webConfigurePath = val
		}
	}
	result.WebConfigurePath = *webConfigurePath

	if !isArgSet("imdb2metaAddr") {
		if val, ok := os.LookupEnv(*envPrefix + "IMDB_2_META_ADDR"); ok {
			*imdb2metaAddr = val
		}
	}
	result.IMDB2metaAddr = *imdb2metaAddr

	if !isArgSet("useOAUTH2") {
		if val, ok := os.LookupEnv(*envPrefix + "USE_OAUTH2"); ok {
			if *useOAUTH2, err = strconv.ParseBool(val); err != nil {
				logger.Fatal("Couldn't convert environment variable from string to bool", zap.Error(err), zap.String("envVar", "USE_OAUTH2"))
			}
		}
	}
	result.UseOAUTH2 = *useOAUTH2

	if !isArgSet("oauth2authURLpm") {
		if val, ok := os.LookupEnv(*envPrefix + "OAUTH2_AUTH_URL_PM"); ok {
			*oauth2authURLpm = val
		}
	}
	result.OAUTH2authorizeURLpm = *oauth2authURLpm

	if !isArgSet("oauth2tokenURLpm") {
		if val, ok := os.LookupEnv(*envPrefix + "OAUTH2_TOKEN_URL_PM"); ok {
			*oauth2tokenURLpm = val
		}
	}
	result.OAUTH2tokenURLpm = *oauth2tokenURLpm

	if !isArgSet("oauth2clientIDpm") {
		if val, ok := os.LookupEnv(*envPrefix + "OAUTH2_CLIENT_ID_PM"); ok {
			*oauth2clientIDpm = val
		}
	}
	result.OAUTH2clientIDpm = *oauth2clientIDpm

	if !isArgSet("oauth2clientSecretPM") {
		if val, ok := os.LookupEnv(*envPrefix + "OAUTH2_CLIENT_SECRET_PM"); ok {
			*oauth2clientSecretPM = val
		}
	}
	result.OAUTH2clientSecretPM = *oauth2clientSecretPM

	if !isArgSet("oauth2encryptionKey") {
		if val, ok := os.LookupEnv(*envPrefix + "OAUTH2_ENCRYPTION_KEY"); ok {
			*oauth2encryptionKey = val
		}
	}
	result.OAUTH2encryptionKey = *oauth2encryptionKey

	return result
}

func (c *config) validate(logger *zap.Logger) {
	if c.StoragePath == "" {
		userCacheDir, err := os.UserCacheDir()
		if err != nil {
			logger.Fatal("Couldn't determine user cache directory via `os.UserCacheDir()`", zap.Error(err))
		}
		// Add two levels, because even if we're in `os.UserCacheDir()`, on Windows that's for example `C:\Users\John\AppData\Local`
		c.StoragePath = filepath.Join(userCacheDir, "deflix-stremio/badger")
	} else {
		c.StoragePath = filepath.Clean(c.StoragePath)
	}
	// If the dir doesn't exist, BadgerDB creates it when writing its DB files.

	if c.CachePath == "" {
		userCacheDir, err := os.UserCacheDir()
		if err != nil {
			logger.Fatal("Couldn't determine user cache directory via `os.UserCacheDir()`", zap.Error(err))
		}
		// Add two levels, because even if we're in `os.UserCacheDir()`, on Windows that's for example `C:\Users\John\AppData\Local`
		c.CachePath = filepath.Join(userCacheDir, "deflix-stremio/cache")
	} else {
		c.CachePath = filepath.Clean(c.CachePath)
	}
	// If the dir doesn't exist, it's created when the files are written.

	if c.UseOAUTH2 &&
		(c.OAUTH2authorizeURLpm == "" || c.OAUTH2clientIDpm == "" || c.OAUTH2clientSecretPM == "" || c.OAUTH2encryptionKey == "" || c.OAUTH2tokenURLpm == "") {
		logger.Fatal("Using OAuth2 requires setting all OAuth2 config values")
	}
}

// isArgSet returns true if the argument you're looking for is actually set as command line argument.
// Pass without "-" prefix.
func isArgSet(arg string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == arg {
			found = true
		}
	})
	return found
}
