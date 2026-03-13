package utils

const (
	// Success messages
	MsgSnippetStored  = "✓ Stored snippet '%s' at %s\n"
	MsgStackStored    = "✓ Stored stack '%s' at %s\n"
	MsgSnippetAdded   = "✓ Added snippet '%s' to %s\n"
	MsgStackAdded     = "✓ Added stack '%s' to %s\n"
	MsgSnippetRemoved = "✓ Removed snippet '%s'\n"
	MsgStackRemoved   = "✓ Removed stack '%s'\n"

	// Warning messages
	MsgWarningSnippetExists = "⚠ Warning: snippet '%s' already exists in store"
	MsgWarningStackExists   = "⚠ Warning: stack '%s' already exists in store"

	// Prompt messages
	MsgPromptVersionOrOverwrite = "Do you want to create a new version (v) or overwrite it (o)? "
	MsgPromptConfirmRemove      = "Remove %s '%s'? (y/N): "
	MsgPromptConfirmCleanAll    = "⚠️  This will remove ALL snippets and stacks!"

	// Error messages
	ErrPathNotExist      = "path '%s' does not exist"
	ErrResourceNotFound  = "%s '%s' not found"
	ErrSnippetMustBeFile = "snippet must be a file, not a directory"
	ErrStackMustBeDir    = "stack must be a directory, not a file"
	ErrSnippetNeedExt    = "snippet file must have an extension"
	ErrFileAlreadyExists = "file '%s' already exists. Use --force to overwrite"
	ErrDestAlreadyExists = "destination '%s' already exists. Use --force to overwrite"

	// Info messages
	MsgCancelled        = "Cancelled"
	MsgNoSnippets       = "No snippets to clean"
	MsgNoStacks         = "No stacks to clean"
	MsgNoResourcesFound = "No resources found in store"
)
