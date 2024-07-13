package workflow

import (
	"go/ast"
	"go/parser"
	"go/token"
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

func ParseComments(src string) (map[string]FunctionMetadata, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, src, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	functions := make(map[string]*FunctionMetadata)
	hooks := make(map[string]*HookMetadata)

	// Parse all comments and capture metadata
	for _, f := range node.Decls {
		if fn, isFn := f.(*ast.FuncDecl); isFn {
			if fn.Doc != nil {
				fnName := fn.Name.Name
				meta := &FunctionMetadata{}
				for _, comment := range fn.Doc.List {
					parseComment(comment.Text, meta, hooks, fnName)
				}
				if meta.Type == "task" {
					functions[fnName] = meta
				} else if meta.Type == "pre-hook" || meta.Type == "post-hook" {
					hooks[fnName] = &HookMetadata{
						Type:        meta.Type,
						Description: meta.Description,
						Parameters:  meta.Parameters,
						Returns:     meta.Returns,
					}
				}
			}
		}
	}

	// Associate hooks with their corresponding tasks
	for fnName, fnMeta := range functions {
		fnMeta.Hooks = make(map[string]HookMetadata)
		for hookName, hookMeta := range hooks {
			if strings.Contains(hookName, fnName) {
				fnMeta.Hooks[hookMeta.Type+"_"+hookName] = *hookMeta
			}
		}
	}

	result := make(map[string]FunctionMetadata)
	for fnName, fnMeta := range functions {
		result[fnName] = *fnMeta
	}

	return result, nil
}

func parseComment(comment string, meta *FunctionMetadata, hookMeta map[string]*HookMetadata, funcName string) {
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
		hookMeta["pre-hook_"+funcName] = &HookMetadata{Type: "pre-hook", Description: meta.Description, Parameters: meta.Parameters, Returns: meta.Returns}
	}
	if reAfter.MatchString(comment) {
		hookMeta["post-hook_"+funcName] = &HookMetadata{Type: "post-hook", Description: meta.Description, Parameters: meta.Parameters, Returns: meta.Returns}
	}
}
