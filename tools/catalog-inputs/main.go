// catalog-inputs extracts source-observed inputs for handlers in the REST catalog.
// Run: go run ./tools/catalog-inputs -source /path/to/multica/server
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Endpoint struct {
	Method  string `json:"method"`
	Path    string `json:"path"`
	Handler string `json:"handler"`
	Query   string `json:"query_fields"`
	Body    string `json:"request_body"`
}
type Catalog struct {
	Source struct {
		Repository string `json:"repository"`
	} `json:"source"`
	Endpoints []Endpoint `json:"endpoints"`
}
type Result struct {
	Handler         string   `json:"handler"`
	File            string   `json:"file"`
	Line            int      `json:"line"`
	RequestStructs  []string `json:"request_structs"`
	JSONFields      []string `json:"json_fields"`
	QueryGet        []string `json:"query_get"`
	QueryHas        []string `json:"query_has"`
	UnresolvedTypes []string `json:"unresolved_types"`
}

// decodedRequestTypes extracts the concrete targets passed to Decoder.Decode.
// It handles inline literals, typed locals, and pointers to composite literals.
func decodedRequestTypes(body *ast.BlockStmt) []string {
	if body == nil { return nil }
	vars := map[string]string{}
	ast.Inspect(body, func(n ast.Node) bool {
		v, ok := n.(*ast.ValueSpec); if !ok { return true }
		name := ""; if len(v.Names)>0 { name=v.Names[0].Name }
		if name!="" && v.Type!=nil { vars[name]=typeName(v.Type) }
		if name!="" && len(v.Values)>0 { if u,ok:=v.Values[0].(*ast.UnaryExpr); ok { if c,ok:=u.X.(*ast.CompositeLit); ok { vars[name]=typeName(c.Type) } } }
		return true
	})
	set:=map[string]bool{}
	ast.Inspect(body, func(n ast.Node) bool {
		c,ok:=n.(*ast.CallExpr); if !ok || len(c.Args)==0 { return true }
		s,ok:=c.Fun.(*ast.SelectorExpr); if !ok || s.Sel.Name!="Decode" { return true }
		u,ok:=c.Args[0].(*ast.UnaryExpr); if !ok { return true }
		switch x:=u.X.(type) { case *ast.CompositeLit: set[typeName(x.Type)]=true; case *ast.Ident: set[vars[x.Name]]=true }
		return true
	})
	out:=[]string{}; for n:=range set { if n!="" { out=append(out,n) } }; sort.Strings(out); return out
}
func typeName(e ast.Expr) string { switch x:=e.(type) { case *ast.Ident:return x.Name; case *ast.StarExpr:return typeName(x.X); case *ast.SelectorExpr:return typeName(x.Sel); case *ast.ArrayType:return typeName(x.Elt) }; return "" }

func main(){
	root := flag.String("source", "", "checkout root of source repository")
	catalogPath := flag.String("catalog", "docs/rest-api-catalog.json", "input REST catalog")
	out := flag.String("out", "", "write JSON report here (stdout if omitted)")
	flag.Parse()
	if *root == "" {
		fmt.Fprintln(os.Stderr, "-source is required")
		os.Exit(2)
	}
	data, err := os.ReadFile(*catalogPath)
	check(err)
	var cat Catalog
	check(json.Unmarshal(data, &cat))
	if cat.Source.Repository == "" {
		fmt.Fprintln(os.Stderr, "catalog source.repository missing")
		os.Exit(2)
	}
	type fnInfo struct {
		file string
		decl *ast.FuncDecl
		fset *token.FileSet
	}
	funcs := map[string]fnInfo{}
	structs := map[string]*ast.StructType{}
	fset := token.NewFileSet()
	err = filepath.WalkDir(*root, func(path string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() && (d.Name() == ".git" || d.Name() == "vendor") {
			return filepath.SkipDir
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		f, e := parser.ParseFile(fset, path, nil, 0)
		if e != nil {
			return e
		}
		for _, dec := range f.Decls {
			if fn, ok := dec.(*ast.FuncDecl); ok && fn.Name != nil && strings.Contains(filepath.ToSlash(path), "/internal/handler/") {
				funcs[fn.Name.Name] = fnInfo{path, fn, fset}
			}
		}
		ast.Inspect(f, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSpec)
			if ok {
				if st, yes := ts.Type.(*ast.StructType); yes {
					structs[ts.Name.Name] = st
				}
			}
			return true
		})
		return nil
	})
	check(err)
	unique := map[string]bool{}
	results := []Result{}
	for _, ep := range cat.Endpoints {
		if ep.Handler == "" || ep.Query == "Not inferred; inspect linked handler source" && ep.Body == "Not inferred; inspect linked handler source" {
			continue
		}
		info, ok := funcs[ep.Handler]
		if !ok {
			unique[ep.Handler] = true
			continue
		}
		key := ep.Handler + "\x00" + info.file
		if unique[key] {
			continue
		}
		unique[key] = true
		r := Result{Handler: ep.Handler, File: rel(*root, info.file), Line: info.fset.Position(info.decl.Pos()).Line}
		seenTypes, seenFields, seenGet, seenHas := map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}
		var addType func(string)
		addType = func(name string) {
			name = strings.TrimPrefix(name, "*")
			name = strings.TrimPrefix(name, "[]")
			if i := strings.Index(name, "."); i >= 0 {
				name = name[i+1:]
			}
			if name == "" || seenTypes[name] {
				return
			}
			st, ok := structs[name]
			if !ok {
				r.UnresolvedTypes = append(r.UnresolvedTypes, name)
				return
			}
			seenTypes[name] = true
			r.RequestStructs = append(r.RequestStructs, name)
			for _, field := range st.Fields.List {
				if id, ok := field.Type.(*ast.Ident); ok {
					if _, yes := structs[id.Name]; yes {
						addType(id.Name)
					}
				}
				if field.Tag != nil {
					raw := strings.Trim(field.Tag.Value, "`")
					for _, part := range strings.Split(raw, " ") {
						if strings.HasPrefix(part, "json:") {
							v := strings.Trim(strings.TrimPrefix(part, "json:"), "\"")
							name := strings.Split(v, ",")[0]
							if name != "" && name != "-" {
								seenFields[name] = true
							}
						}
					}
				}
			}
		}
		for _, name := range decodedRequestTypes(info.decl.Body) { addType(name) }
		if info.decl.Type.Params != nil {
			for _, p := range info.decl.Type.Params.List {
				if id, ok := p.Type.(*ast.Ident); ok {
					addType(id.Name)
				}
			}
		}
		ast.Inspect(info.decl.Body, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.CallExpr:
				if s, ok := x.Fun.(*ast.SelectorExpr); ok {
					if s.Sel.Name == "Get" && len(x.Args) > 0 {
						if lit, ok := x.Args[0].(*ast.BasicLit); ok {
							seenGet[strings.Trim(lit.Value, "\"")] = true
						}
					}
					if s.Sel.Name == "Has" && len(x.Args) > 0 {
						if lit, ok := x.Args[0].(*ast.BasicLit); ok {
							seenHas[strings.Trim(lit.Value, "\"")] = true
						}
					}
				}
			case *ast.CompositeLit:
				if id, ok := x.Type.(*ast.Ident); ok {
					if _, yes := structs[id.Name]; yes {
						addType(id.Name)
					}
				}
			case *ast.UnaryExpr:
				if id, ok := x.X.(*ast.Ident); ok {
					if _, yes := structs[id.Name]; yes {
						addType(id.Name)
					}
				}
			}
			return true
		})
		for s := range seenFields {
			r.JSONFields = append(r.JSONFields, s)
		}
		for s := range seenGet {
			r.QueryGet = append(r.QueryGet, s)
		}
		for s := range seenHas {
			r.QueryHas = append(r.QueryHas, s)
		}
		for s := range seenTypes {
			_ = s
		}
		sort.Strings(r.RequestStructs)
		sort.Strings(r.JSONFields)
		sort.Strings(r.QueryGet)
		sort.Strings(r.QueryHas)
		sort.Strings(r.UnresolvedTypes)
		results = append(results, r)
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Handler != results[j].Handler {
			return results[i].Handler < results[j].Handler
		}
		return results[i].File < results[j].File
	})
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if *out != "" {
		f, e := os.Create(*out)
		check(e)
		defer f.Close()
		enc = json.NewEncoder(f)
		enc.SetIndent("", "  ")
	}
	check(enc.Encode(results))
}
func rel(root, path string) string {
	r, e := filepath.Rel(root, path)
	if e != nil {
		return path
	}
	return filepath.ToSlash(r)
}
func check(e error) {
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
