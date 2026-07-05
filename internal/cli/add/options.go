package add

// ResourceType defines whether a resource is treated as a snippet, stack, or auto-detected.
type ResourceType int

const (
	ResourceTypeAuto ResourceType = iota
	ResourceTypeStack
	ResourceTypeSnippet
)

// Options holds the execution context for add operations.
type Options struct {
	Force     bool   // --force: overwrite existing files without confirmation
	Spread    bool   // --spread: copy stack contents directly into destination
	AsStack   bool   // --stack/-k: force stack mode
	AsSnippet bool   // --snippet/-n: force snippet mode
	Registry  string // --registry: custom registry URL override
	Name      string // --name/-m: rename snippet or stack in destination directory
}
