package main

import "os"

//HGNConfig struct used to consume configuration details
type HGNConfig struct {
	CertFile    string
	CertKeyFile string
	UseSSL      string

	BotName  string
	MasterID string
}

//DBConfig struct used to consume Configuration details
//regarding database access
type DBConfig struct {
	DBHost string
	DBUser string
	DBName string
	DBPass string
}

//loadConfig specifcially loads configuration information
//for the bot as opposed to the database.
//I'm pretty sure there is a better way to use one function
//to load and distribute configs where they are needed, but
//I haven't quite figreud out the way to do so just yet.
func initConfig() (config HGNConfig) {

	return HGNConfig{
		CertFile:    os.Getenv("HGNOTIFY_CERT_FILE"),
		CertKeyFile: os.Getenv("HGNOTIFY_CERT_KEY_FILE"),
		UseSSL:      os.Getenv("HGNOTIFY_USE_SSL"),

		BotName:  os.Getenv("HGNOTIFY_BOT_NAME"),
		MasterID: os.Getenv("HGNOTIFY_MASTER_GID"),
	}
}

//loadDBConfig specifically loads configuration information
//for the database as oppposed to the bot/connections
func initDBConfig() (config DBConfig) {
	return DBConfig{
		DBHost: os.Getenv("HGNOTIFY_DB_HOST"),
		DBUser: os.Getenv("HGNOTIFY_DB_USER"),
		DBName: os.Getenv("HGNOTIFY_DB_NAME"),
		DBPass: os.Getenv("HGNOTIFY_DB_PASS"),
	}
}
