package version

var (
	Version   = "0.0.1"
	GitCommit = "dev"
	BuildDate = "unknown"
	GoVersion = "unknown"
)

func Info() string {
	return "boiler version " + Version
}

func FullInfo() string {
	info := "boiler version " + Version + "\n"
	info += "Git commit: " + GitCommit + "\n"
	info += "Build date: " + BuildDate + "\n"
	info += "Go version: " + GoVersion
	return info
}
