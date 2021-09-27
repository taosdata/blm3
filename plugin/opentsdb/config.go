package opentsdb

type Config struct {
	OpenTSDB OpenTSDB
}

type OpenTSDB struct {
	DB      string
	Disable bool
}
