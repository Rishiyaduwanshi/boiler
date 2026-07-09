package constants

const (
	// Boiler Directories
	BoilerDirName   = "bl"
	StoreDirName    = "store"
	SnippetsDirName = "snippets"
	StacksDirName   = "stacks"
	LogsDirName     = "logs"
	BinDirName      = "bin"

	// Config Files
	GlobalConfigFileName = "boiler.conf.json"
	LocalConfigFileName  = "boiler.local.json"

	// URLs & Schemes
	SchemeHTTP  = "http://"
	SchemeHTTPS = "https://"

	DefaultRegistryURL      = "https://github.com/rishiyaduwanshi/boiler"
	GithubRawContentBaseURL = "https://raw.githubusercontent.com"

	// Syntax & Variables
	VarPrefix     = "bl__"
	DetectorStart = "bl__DETECTOR_START"
	DetectorEnd   = "bl__DETECTOR_END"

	// Environment Variables
	EnvHome        = "HOME"
	EnvUserProfile = "USERPROFILE" // Windows home
	EnvEditor      = "EDITOR"

	// Default Values
	DefaultBranch = "main"

	// Git Providers
	ProviderGithubHost    = "github.com"
	ProviderGithubWWW     = "www.github.com"
	ProviderGithubRawHost = "raw.githubusercontent.com"
	ProviderGithubAlias   = "github"

	ProviderGitlabHost  = "gitlab.com"
	ProviderGitlabWWW   = "www.gitlab.com"
	ProviderGitlabAlias = "gitlab"

	ProviderBitbucketHost  = "bitbucket.org"
	ProviderBitbucketWWW   = "www.bitbucket.org"
	ProviderBitbucketAlias = "bitbucket"

	// Provider APIs
	ProviderGithubAPI    = "api.github.com"
	ProviderGitlabAPI    = "api/v4/projects"
	ProviderBitbucketAPI = "api.bitbucket.org/2.0"

	// Patterns
	VarKeyPattern = `^[a-z_][a-z0-9_-]*$`
)
