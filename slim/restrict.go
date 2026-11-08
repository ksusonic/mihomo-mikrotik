// Apply the experimental, fail-closed VLESS/TCP/REALITY + TPROXY profile.
package main

import (
 "go/ast"
 "go/format"
 "go/parser"
 "go/token"
 "os"
 "path/filepath"
 "strconv"
 "strings"
)

func body(src string) *ast.BlockStmt {
 f, err := parser.ParseFile(token.NewFileSet(), "snippet.go", "package x; func f() {"+src+"}", 0)
 if err != nil { panic(err) }; return f.Decls[0].(*ast.FuncDecl).Body
}
func main() {
 root := os.Args[1]
 for _, name := range []string{"adapter/parser.go", "listener/parse.go", "adapter/outbound/vless.go", "listener/listener.go", "config/config.go"} {
  path := filepath.Join(root,name); fs := token.NewFileSet(); f,err := parser.ParseFile(fs,path,nil,parser.ParseComments); if err != nil {panic(err)}
  for _, decl := range f.Decls {
   fn,ok := decl.(*ast.FuncDecl); if !ok {continue}
   if name=="adapter/parser.go" || name=="listener/parse.go" {
    ast.Inspect(fn.Body,func(n ast.Node) bool {
     sw,ok:=n.(*ast.SwitchStmt); if !ok {return true}; id,ok:=sw.Tag.(*ast.Ident); if !ok || id.Name!="proxyType" {return true}
     kept:=[]ast.Stmt{}; for _, s:=range sw.Body.List { c:=s.(*ast.CaseClause); keep:=len(c.List)==0
      for _, e:=range c.List { v:=e.(*ast.BasicLit).Value; if name=="adapter/parser.go" {keep=keep || v==`"vless"` || v==`"direct"` || v==`"dns"` || v==`"reject"` || v==`"rematch"`} else {keep=keep || v==`"tproxy"`} }
      if keep {kept=append(kept,c)}
     }; sw.Body.List=kept; return false
    })
   }
   if name=="adapter/outbound/vless.go" {
    switch fn.Name.Name {
    case "StreamConnContext": fn.Body=body(`c, err = v.streamTLSConn(ctx,c,false); if err != nil {return nil,err}; return v.streamConnContext(ctx,c,metadata)`)
    case "dialContext": fn.Body=body(`return v.dialer.DialContext(ctx,"tcp",v.addr)`)
    case "Close": fn.Body=body(`return nil`)
    case "NewVless":
     kept:=[]ast.Stmt{}; for _,s:=range fn.Body.List { sw,ok:=s.(*ast.SwitchStmt); if ok { if e,ok:=sw.Tag.(*ast.SelectorExpr); ok && e.Sel.Name=="Network" {continue} }; kept=append(kept,s) }
     guard:=body(`if (option.Network!="" && option.Network!="tcp") || !option.TLS || option.RealityOpts.PublicKey=="" {return nil,fmt.Errorf("this build supports only VLESS TCP with REALITY")}`).List
     fn.Body.List=append(guard,kept...)
    }
   }
   if name=="listener/listener.go" {
    switch fn.Name.Name {
    case "ReCreateShadowSocks", "ReCreateVmess", "ReCreateTuic", "ReCreateTun", "ReCreateHTTP", "ReCreateSocks", "ReCreateRedir", "ReCreateMixed": fn.Body=body(``)
    }
   }
   if name=="config/config.go" && fn.Name.Name=="ParseRawConfig" {
    guard:=body(`if rawCfg.Port!=0 || rawCfg.SocksPort!=0 || rawCfg.RedirPort!=0 || rawCfg.MixedPort!=0 || rawCfg.ShadowSocksConfig!="" || rawCfg.VmessConfig!="" || rawCfg.Tun.Enable || rawCfg.TuicServer.Enable {return nil,fmt.Errorf("this build supports only TPROXY listeners")}`).List
    fn.Body.List=append(guard,fn.Body.List...)
   }
  }
  if name=="adapter/outbound/vless.go" {
   for _,d:=range f.Decls { if g,ok:=d.(*ast.GenDecl); ok { for _,s:=range g.Specs { if t,ok:=s.(*ast.TypeSpec); ok && t.Name.Name=="Vless" {st:=t.Type.(*ast.StructType); fields:=[]*ast.Field{}; for _,field:=range st.Fields.List {if len(field.Names)>0 && (field.Names[0].Name=="gunClient" || field.Names[0].Name=="xhttpClient") {continue}; fields=append(fields,field)}; st.Fields.List=fields} } } }
  }
  // Drop imports made unreachable by removal of transport branches.
  used:=map[string]bool{}; ast.Inspect(f,func(n ast.Node) bool {if s,ok:=n.(*ast.SelectorExpr); ok {if i,ok:=s.X.(*ast.Ident); ok {used[i.Name]=true}};return true})
  for _,d:=range f.Decls {if g,ok:=d.(*ast.GenDecl); ok && g.Tok==token.IMPORT {specs:=[]ast.Spec{}; for _,s:=range g.Specs {im:=s.(*ast.ImportSpec); p,_:=strconv.Unquote(im.Path.Value); alias:=filepath.Base(p); if im.Name!=nil {alias=im.Name.Name}; if alias=="_" || alias=="." || used[alias] {specs=append(specs,s)}};g.Specs=specs}}
  out,err:=os.Create(path);if err!=nil {panic(err)};if err=format.Node(out,fs,f);err!=nil {panic(err)};out.Close()
 }
 // A version-visible marker makes accidental installation of an unrestricted binary detectable.
 os.WriteFile(filepath.Join(root,"constant/features/vless_only.go"),[]byte("package features\nconst VlessOnly = true\n"),0644)
 _=strings.Contains
}
