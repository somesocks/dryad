package core

import (
	"dryad/internal/filepath"
	"dryad/internal/os"
	"encoding/json"
	"strings"

	"dryad/task"
)

type rootDevelopSelectedPaths struct {
	AssetsPath       string `json:"assets_path"`
	CommandsPath     string `json:"commands_path"`
	TraitsPath       string `json:"traits_path"`
	SecretsPath      string `json:"secrets_path"`
	DocsPath         string `json:"docs_path"`
	RequirementsPath string `json:"requirements_path"`
}

func rootDevelop_selectedPathsFile(workspacePath string) string {
	return filepath.Join(rootDevelop_devDir(workspacePath), "selected-paths.json")
}

func rootDevelop_resolveSelectedPaths(
	ctx *task.ExecutionContext,
	rootPath string,
	variantDescriptor string,
) (error, rootDevelopSelectedPaths) {
	rootRef := SafeRootReference{BasePath: rootPath}
	err, unsafeVariant := rootRef.VariantFromFilesystem(variantDescriptor)
	if err != nil {
		return err, rootDevelopSelectedPaths{}
	}

	err, variant := unsafeVariant.Resolve(ctx)
	if err != nil {
		return err, rootDevelopSelectedPaths{}
	}

	assetsPath := ""
	if variant.Assets != nil {
		assetsPath = variant.Assets.BasePath
	}
	commandsPath := ""
	if variant.Commands != nil {
		commandsPath = variant.Commands.BasePath
	}
	traitsPath := ""
	if variant.Traits != nil {
		traitsPath = variant.Traits.BasePath
	}
	secretsPath := ""
	if variant.Secrets != nil {
		secretsPath = variant.Secrets.BasePath
	}
	docsPath := ""
	if variant.Docs != nil {
		docsPath = variant.Docs.BasePath
	}
	requirementsPath := ""
	if variant.Requirements != nil {
		requirementsPath = variant.Requirements.BasePath
	}

	return nil, rootDevelopSelectedPaths{
		AssetsPath:       assetsPath,
		CommandsPath:     commandsPath,
		TraitsPath:       traitsPath,
		SecretsPath:      secretsPath,
		DocsPath:         docsPath,
		RequirementsPath: requirementsPath,
	}
}

func rootDevelop_writeSelectedPaths(
	workspacePath string,
	selectedPaths rootDevelopSelectedPaths,
) error {
	rawBytes, err := json.Marshal(selectedPaths)
	if err != nil {
		return err
	}

	return os.WriteFile(rootDevelop_selectedPathsFile(workspacePath), rawBytes, 0o644)
}

func rootDevelop_readSelectedPaths(workspacePath string) (error, *rootDevelopSelectedPaths) {
	selectedPathsFile := rootDevelop_selectedPathsFile(workspacePath)
	exists, err := fileExists(selectedPathsFile)
	if err != nil {
		return err, nil
	}
	if !exists {
		return nil, nil
	}

	rawBytes, err := os.ReadFile(selectedPathsFile)
	if err != nil {
		return err, nil
	}

	selectedPaths := rootDevelopSelectedPaths{}
	err = json.Unmarshal(rawBytes, &selectedPaths)
	if err != nil {
		return err, nil
	}

	return nil, &selectedPaths
}

func rootDevelop_defaultSelectedPaths(basePath string) rootDevelopSelectedPaths {
	return rootDevelopSelectedPaths{
		AssetsPath:       filepath.Join(basePath, "dyd", "assets"),
		CommandsPath:     filepath.Join(basePath, "dyd", "commands"),
		TraitsPath:       filepath.Join(basePath, "dyd", "traits"),
		SecretsPath:      filepath.Join(basePath, "dyd", "secrets"),
		DocsPath:         filepath.Join(basePath, "dyd", "docs"),
		RequirementsPath: filepath.Join(basePath, "dyd", "requirements"),
	}
}

func rootDevelop_targetBasePathForKey(
	key string,
	basePath string,
	selectedPaths *rootDevelopSelectedPaths,
) (string, bool) {
	paths := rootDevelop_defaultSelectedPaths(basePath)
	if selectedPaths != nil {
		paths = *selectedPaths
	}

	type mapping struct {
		RelPrefix string
		BasePath  string
	}

	mappings := []mapping{
		{RelPrefix: filepath.Join("dyd", "assets"), BasePath: paths.AssetsPath},
		{RelPrefix: filepath.Join("dyd", "commands"), BasePath: paths.CommandsPath},
		{RelPrefix: filepath.Join("dyd", "docs"), BasePath: paths.DocsPath},
		{RelPrefix: filepath.Join("dyd", "traits"), BasePath: paths.TraitsPath},
		{RelPrefix: filepath.Join("dyd", "secrets"), BasePath: paths.SecretsPath},
		{RelPrefix: filepath.Join("dyd", "requirements"), BasePath: paths.RequirementsPath},
	}

	for _, mapping := range mappings {
		relPath, err := filepath.Rel(mapping.RelPrefix, key)
		if err != nil {
			continue
		}
		if relPath == "." || strings.HasPrefix(relPath, ".."+string(filepath.Separator)) || relPath == ".." {
			continue
		}
		if mapping.BasePath == "" {
			return "", false
		}
		return filepath.Join(mapping.BasePath, relPath), true
	}

	return "", false
}
