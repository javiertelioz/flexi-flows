package comment

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type FunctionMetadata struct {
	Type        string                  `json:"type"`
	Description string                  `json:"description"`
	Parameters  []ParameterMetadata     `json:"parameters"`
	Returns     []ReturnMetadata        `json:"returns"`
	Hooks       map[string]HookMetadata `json:"hooks,omitempty"`
}

type ParameterMetadata struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type ReturnMetadata struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type HookMetadata struct {
	Type        string              `json:"type"`
	Description string              `json:"description"`
	Parameters  []ParameterMetadata `json:"parameters"`
	Returns     []ReturnMetadata    `json:"returns"`
}

func ParseComments(dir string) (map[string]FunctionMetadata, error) {
	metadata := make(map[string]FunctionMetadata)
	hookMetadata := make(map[string]HookMetadata)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			err = parseFile(path, metadata, hookMetadata)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Associate hooks with their corresponding tasks
	for fnName, fnMeta := range metadata {
		fnMeta.Hooks = make(map[string]HookMetadata)
		for name, hook := range hookMetadata {
			if strings.HasPrefix(name, fnName) {
				fnMeta.Hooks[name] = hook
			}
		}
	}

	return metadata, nil
}

func parseFile(path string, metadata map[string]FunctionMetadata, hookMetadata map[string]HookMetadata) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	for _, f := range node.Decls {
		if fn, isFn := f.(*ast.FuncDecl); isFn {
			if fn.Doc != nil {
				meta := FunctionMetadata{}
				for _, comment := range fn.Doc.List {
					parseComment(comment.Text, &meta, hookMetadata, fn.Name.Name)
				}
				if meta.Type == "task" {
					meta.Hooks = make(map[string]HookMetadata)
					for name, hook := range hookMetadata {
						if strings.HasPrefix(name, fn.Name.Name) {
							meta.Hooks[name] = hook
						}
					}
					metadata[fn.Name.Name] = meta
				} else if meta.Type == "pre-hook" || meta.Type == "post-hook" {
					hookMetadata[meta.Type+"_"+fn.Name.Name] = HookMetadata{
						Type:        meta.Type,
						Description: meta.Description,
						Parameters:  meta.Parameters,
						Returns:     meta.Returns,
					}
				}
			}
		}
	}

	return nil
}

func parseComment(comment string, meta *FunctionMetadata, hookMeta map[string]HookMetadata, funcName string) {
	reType := regexp.MustCompile(`@type:\s+(.*)`)
	reDescription := regexp.MustCompile(`@description\s+(.*)`)
	reInput := regexp.MustCompile(`@input\s+(\w+)\s+\((.*?)\):\s+(.*)`)
	reOutput := regexp.MustCompile(`@output\s+(\w+):\s+(.*)`)
	reBefore := regexp.MustCompile(`@before\s+(.*)`)
	reAfter := regexp.MustCompile(`@after\s+(.*)`)

	if reType.MatchString(comment) {
		meta.Type = reType.FindStringSubmatch(comment)[1]
	}
	if reDescription.MatchString(comment) {
		meta.Description = reDescription.FindStringSubmatch(comment)[1]
	}
	if reInput.MatchString(comment) {
		for _, match := range reInput.FindAllStringSubmatch(comment, -1) {
			param := ParameterMetadata{
				Type:        match[2],
				Description: match[3],
			}
			meta.Parameters = append(meta.Parameters, param)
		}
	}
	if reOutput.MatchString(comment) {
		for _, match := range reOutput.FindAllStringSubmatch(comment, -1) {
			ret := ReturnMetadata{
				Type:        match[1],
				Description: match[2],
			}
			meta.Returns = append(meta.Returns, ret)
		}
	}
	if reBefore.MatchString(comment) {
		hookMeta["pre-hook_"+funcName] = HookMetadata{Type: "pre-hook", Description: meta.Description, Parameters: meta.Parameters, Returns: meta.Returns}
	}
	if reAfter.MatchString(comment) {
		hookMeta["post-hook_"+funcName] = HookMetadata{Type: "post-hook", Description: meta.Description, Parameters: meta.Parameters, Returns: meta.Returns}
	}
}
