package armory

import (
	"strings"
	"sync"
	"testing"

	"github.com/bishopfox/sliver/client/assets"
	"github.com/bishopfox/sliver/client/command/extensions"
)

func makeExtensionEntry(id, armoryPK, armoryName, extensionName, commandName, dependsOn string) pkgCacheEntry {
	return pkgCacheEntry{
		ID: id,
		ArmoryConfig: &assets.ArmoryConfig{
			PublicKey: armoryPK,
			Name:      armoryName,
		},
		Pkg: ArmoryPackage{
			Name:    extensionName,
			IsAlias: false,
		},
		Extension: &extensions.ExtensionManifest{
			Name: extensionName,
			ExtCommand: []*extensions.ExtCommand{{
				CommandName: commandName,
				DependsOn:   dependsOn,
			}},
		},
	}
}

func resetPkgCache(entries ...pkgCacheEntry) {
	pkgCache = sync.Map{}
	for _, entry := range entries {
		pkgCache.Store(entry.ID, entry)
	}
}

func TestResolveExtensionPackageDependencies_RespectsArmoryScope(t *testing.T) {
	root := makeExtensionEntry("root-a", "armory-a", "Armory A", "root-extension", "rootcmd", "dep-cmd")
	depFromOtherArmory := makeExtensionEntry("dep-b", "armory-b", "Armory B", "dep-extension", "dep-cmd", "")

	resetPkgCache(root, depFromOtherArmory)

	deps := make(map[string]*pkgCacheEntry)
	err := resolveExtensionPackageDependencies(&root, "armory-a", deps, map[string]string{})
	if err == nil {
		t.Fatalf("expected dependency resolution to fail when dependency only exists in another armory")
	}
	if !strings.Contains(err.Error(), "could not resolve dependency dep-cmd") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveExtensionPackageDependencies_UsesArmoryScopeForRecursion(t *testing.T) {
	root := makeExtensionEntry("root-a", "armory-a", "Armory A", "root-extension", "rootcmd", "dep-cmd")
	depFromSelectedArmory := makeExtensionEntry("dep-a", "armory-a", "Armory A", "dep-extension", "dep-cmd", "")
	depFromOtherArmory := makeExtensionEntry("dep-b", "armory-b", "Armory B", "dep-extension", "dep-cmd", "")

	resetPkgCache(root, depFromSelectedArmory, depFromOtherArmory)

	deps := make(map[string]*pkgCacheEntry)
	err := resolveExtensionPackageDependencies(&root, "armory-a", deps, map[string]string{})
	if err != nil {
		t.Fatalf("expected dependency resolution to succeed: %v", err)
	}

	resolved, ok := deps["dep-cmd"]
	if !ok {
		t.Fatalf("expected dep-cmd to be resolved")
	}
	if resolved.ArmoryConfig.PublicKey != "armory-a" {
		t.Fatalf("resolved dependency from unexpected armory: got %s, want armory-a", resolved.ArmoryConfig.PublicKey)
	}
}
