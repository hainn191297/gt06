package config

type Config struct {
	TCPServer         string   `json:"TCPServer" yaml:"TCPServer"`
	MongoURI          string   `json:"MongoURI" yaml:"MongoURI"`
	DBName            string   `json:"DBName" yaml:"DBName"`
	LogLevel          string   `json:"LogLevel" yaml:"LogLevel"`
	Timeout           int      `json:"Timeout" yaml:"Timeout"`
	ScyllaHosts       []string `json:"ScyllaHosts" yaml:"ScyllaHosts"`
	ScyllaKeyspace    string   `json:"ScyllaKeyspace" yaml:"ScyllaKeyspace"`
	ScyllaConsistency string   `json:"ScyllaConsistency" yaml:"ScyllaConsistency"`
}

// Default returns a Config with default values
func Default() Config {
	return Config{
		TCPServer:         "0.0.0.0:8000",
		MongoURI:          "mongodb://localhost:27017",
		DBName:            "gt06",
		LogLevel:          "info",
		Timeout:           10,
		ScyllaHosts:       []string{"localhost:9042"},
		ScyllaKeyspace:    "gt06",
		ScyllaConsistency: "LOCAL_ONE",
	}
}
