package add

import (
	"fmt"

	"github.com/rishiyaduwanshi/boiler/internal/store"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
)

// ResolveStoreResource attempts to resolve a resource from the store (local or remote).
// It honors the requested resourceType (stack, snippet, or auto-detect).
// For auto-detect without an extension, it gracefully falls back from stack to snippet.
func ResolveStoreResource(st *store.Store, resource string, resourceType ResourceType, isRemote bool) (resolvedName string, isSnippet bool, err error) {
	baseName, version, ext := store.ParseResourceName(resource)

	var locationType string
	if isRemote {
		locationType = "remote registry"
	} else {
		locationType = "local store"
	}

	fetchAsSnippet := ext != ""

	if resourceType == ResourceTypeSnippet {
		fetchAsSnippet = true
	}

	if resourceType == ResourceTypeStack {
		fetchAsSnippet = false
	}

	// Auto-detect fallback: if no extension, check if it matches a snippet but not a stack
	if resourceType == ResourceTypeAuto && ext == "" {
		_, stackExists := st.GetStack(resource)
		if !stackExists {
			stackMatches := utils.FindMatchingResources(st.ListStacks(), baseName, "")
			if len(stackMatches) == 0 {
				_, snippetExists := st.GetSnippet(resource)
				snippetMatches := utils.FindMatchingResources(st.ListSnippets(), baseName, "")
				if snippetExists || len(snippetMatches) > 0 {
					fetchAsSnippet = true
				}
			}
		}
	}

	if fetchAsSnippet {
		if version != "" && ext != "" && !isRemote {
			return baseName + "@" + version + ext, true, nil
		}

		if _, ok := st.GetSnippet(resource); ok {
			return resource, true, nil
		}

		matches := utils.FindMatchingResources(st.ListSnippets(), baseName, ext)
		if len(matches) == 0 {
			return "", true, fmt.Errorf("snippet '%s' not found in %s", resource, locationType)
		}

		lookupName := baseName
		if ext != "" {
			lookupName = baseName + ext
		}

		selected, err := utils.PickFromList(lookupName, matches)
		if err != nil {
			return "", true, err
		}

		return selected, true, nil
	}

	if resourceType == ResourceTypeSnippet {
		return "", false, fmt.Errorf("snippet '%s' not found in %s", resource, locationType)
	}

	if _, ok := st.GetStack(resource); ok {
		return resource, false, nil
	}

	matches := utils.FindMatchingResources(st.ListStacks(), baseName, "")
	if len(matches) == 0 {
		return "", false, fmt.Errorf("resource '%s' not found in %s", resource, locationType)
	}

	selected, err := utils.PickFromList(baseName, matches)
	if err != nil {
		return "", false, err
	}

	return selected, false, nil
}
