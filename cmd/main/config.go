package main

type appConfig struct {
	poolSize      int
	adminServer   webServerConfig
	matchServer   webServerConfig
	clusterServer webServerConfig
	walWrite      walFileConfig
	walLoad       walFileConfig
	writeWalFiles bool
}

type webServerConfig struct {
	bindTo              string
	port                int
	useTLS              bool
	certFilePath        string
	keyFilePath         string
	credentialsFilePath string
	credentialsFilePwd  string
}

type walFileConfig struct {
	fileDirectory               string
	filePrefix                  string
	maxEntriesPerFile           int
	maxDurationPerFileInSeconds int
}
