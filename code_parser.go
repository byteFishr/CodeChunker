package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/python"

)


var langMap map[string]*sitter.Language = map[string]*sitter.Language{
	".c": cpp.GetLanguage(),
	".cpp": cpp.GetLanguage(),
	".py": python.GetLanguage(),
	".go": golang.GetLanguage(),
	".js": javascript.GetLanguage(),
}


const (
	Break = iota
	Continue 
	Skip 
)

func DFSVisit(node *sitter.Node, f func(n *sitter.Node) int) {
	if node == nil{
		return
	}
	for i := 0; i < int(node.ChildCount()); i++{
		child := node.Child(i)
		res := f(node)
		switch res {
		case Continue:
			DfsVisit(child, f)
		case Break:
			return 
		case Skip:
		}

	}

}

type CodeParser struct{
	Root *sitter.Node
	FileExt string
	Content []byte
}

func NewCodeParser(content []byte, fileExt string) *CodeParser {
	return &CodeParser{
		FileExt: fileExt,
		Content: content,
	}
}


// parse file and get the rootNode
func (c *CodeParser) ParserAst() (*sitter.Node, error) {
	if _, ok := langMap[c.FileExt]; !ok{
		return nil, fmt.Errorf("unsupport file type: %s", c.FileExt)
	}
	lang := langMap[c.FileExt]
	parser := sitter.NewParser()
	parser.SetLanguage(lang)
	tree, err := parser.ParseCtx(context.Background(), nil, c.Content)
	if err != nil{
		return nil, fmt.Errorf("faild to parse the tree: %s", err)
	}
	c.Root = tree.RootNode()
	return tree.RootNode(), nil
}


// func extract_points_of_interest(node *sitter.Node, file_ext string) {
// 	node_types_of_interest = self.

// }

// GetNodeTypesOfInterest maps AST node types to human-readable labels by file extension
func (c *CodeParser) GetNodeTypeOfInterest(fileExt string) map[string]string{

	nodeTypes := map[string]map[string]string{
		"c":{
			"type_definition": "Type Define",
			"function_definition": "Function",
			"struct_specifier": "Struct",
			"preproc_def": "Micro Function",
			"preproc_function_def": "Micro Function",
			"declaration": "Var Declare",
		},
		"py":{
			"import_statement": "Import",
			"export_statement": "Export",
			"class_definition": "class",
			"function_definition": "Function",
		},

	}

	if val, ok := nodeTypes[fileExt]; ok{
		return val
	}else if fileExt == "jsx"{
		return nodeTypes["js"]
	} else if fileExt == "tsx" {
		return nodeTypes["ts"]
	}

	return make(map[string]string)

}

type PointerInterest struct {
	Node *sitter.Node
	Label string
}


func (c *CodeParser) ExtractPointsOfInterest(root *sitter.Node, fileExt string) []*PointerInterest {
	var result []*PointerInterest

	nodeTypes := c.GetNodeTypeOfInterest(fileExt)

	DFSVisit(root, func(n *sitter.Node) int {
		if label, ok := nodeTypes[n.Type()]; ok{
			result = append(result, &PointerInterest{Node: n, Label: label})
			return Skip
		}
		return Continue
	})

	return result

}

func(c *CodeParser) GetLinesForPointsOfInterest(content []byte) []int {
	points := c.ExtractPointsOfInterest(c.Root, c.FileExt)
	
	lineNumbers := make(map[string][]int)
	for _, p := range points {
		startLine := int(p.Node.StartPoint().Row) + 1
		if !contains(lineNumbers[p.Label], startLine){
			lineNumbers[p.Label] = append(lineNumbers[p.Label], startLine)
		}
	}

	var lines []int

	for _, lns := range lineNumbers{
		lines = append(lines, lns...)
	}

	return uniqueIntSlice(lines)

}

func contains(slice []int, item int) bool{
	for _, s := range slice{
		if s == item{
			return true
		}
	}
	return false
}

// Unique removes duplicates from int slice
func uniqueIntSlice(input []int) []int{
	keys := make(map[int]bool)
	list := []int{}
	for _, entry := range input{
		if _, value := keys[entry]; !value{
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func (c *CodeParser) PrintAllLineTypes(code []byte, fileExt string){
	root, err := c.ParserAst()
	if err != nil{
		fmt.Printf("failed to parse ast: %v\n", err)
		return 
	}

	linesToNodeTypes := make(map[int][]string)
	c.MapLineToNodeType(root, linesToNodeTypes)

	codeLines := strings.Split(string(code), "\n")

	for lineNum, types := range linesToNodeTypes{
		if lineNum <= len(codeLines){
			fmt.Printf("line %d: %v | Code: %s\n", lineNum, strings.Join(types, ","))
		}
	}

}

func (c *CodeParser) MapLineToNodeType(node *sitter.Node, lineToNodeType map[int][]string){
	if node == nil{
		return
	}
	startLine := int(node.StartPoint().Row) + 1

	if _, ok := lineToNodeType[startLine]; !ok{
		lineToNodeType[startLine] = []string{}
	}

	lineToNodeType[startLine] = append(lineToNodeType[startLine], node.Type())

	for i := 0; i < int(node.ChildCount()); i++{
		child := node.Child(i)
		c.MapLineToNodeType(child, lineToNodeType)

	}

}

