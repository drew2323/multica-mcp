package main

import (
 "go/ast"
 "go/parser"
 "go/token"
 "reflect"
 "testing"
)

func TestDecodedRequestTypes(t *testing.T) {
 src := `package p
type Create struct { Name string }
type Update struct { Value int }
func h(r *Request) { var req Create; json.NewDecoder(r.Body).Decode(&req); json.NewDecoder(r.Body).Decode(&Update{}) }
`
 f, err := parser.ParseFile(token.NewFileSet(), "fixture.go", src, 0); if err != nil { t.Fatal(err) }
 got := decodedRequestTypes(f.Decls[2].(*ast.FuncDecl).Body)
 want := []string{"Create", "Update"}; if !reflect.DeepEqual(got,want) { t.Fatalf("got %v want %v",got,want) }
}
