package version

var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

type Info struct {
	Version   string
	Commit    string
	BuildTime string
}

func Get() Info {
	return Info{
		Version:   version,
		Commit:    commit,
		BuildTime: buildTime,
	}
}
