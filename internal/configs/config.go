package configs

type Config struct {
	Port      int             `json:"port,default=12395"`
	DBFile    string          `json:"dbFile,default=data/data.db"`
	LogFile   string          `json:"logFile,default=logs/share.log"`
	MediaDir  string          `json:"mediaDir,default=media_dir"`
	DBType    string          `json:"dbType,default=sqlite"`
	MySQL     *MySQLConfig    `json:"mysql,omitempty"`
	Postgres  *PostgresConfig `json:"postgres,omitempty"`
	TaskEngine *TaskEngineConfig `json:"taskEngine,omitempty"`
}

type MySQLConfig struct {
	Host   string `json:"host,default=127.0.0.1"`
	Port   int    `json:"port,default=3306"`
	User   string `json:"user,default=root"`
	Pass   string `json:"pass,default=root"`
	DBName string `json:"dbName,default=share"`
}

type PostgresConfig struct {
	Host    string `json:"host,default=192.168.31.51"`
	Port    int    `json:"port,default=5432"`
	User    string `json:"user,default=postgres"`
	Pass    string `json:"pass,default=postgres"`
	DBName  string `json:"dbName,default=share"`
	SSLMode string `json:"sslMode,default=disable"`
}

type TaskEngineConfig struct {
	WorkerCount    int   `json:"workerCount,default=4"`
	BufferSize     int   `json:"bufferSize,default=89120"`
	ProcessTimeout int   `json:"processTimeout,default=1800"`
	MaxRetry      int   `json:"maxRetry,default=3"`
}

type RuntimeConfig struct {
	*Config
	BuildDate  string
	Commit     string
	GitBranch  string
	GitSummary string
	Version    string
}
