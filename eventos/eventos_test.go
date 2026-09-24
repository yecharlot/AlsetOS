package eventos
import("path/filepath";"testing")
func TestCadenaPersistente(t *testing.T){r,err:=Nuevo(filepath.Join(t.TempDir(),"eventos.jsonl"));if err!=nil{t.Fatal(err)};if _,err=r.Emitir(Evento{Tipo:"inicio",RootCID:"rootcid:x"});err!=nil{t.Fatal(err)};if _,err=r.Emitir(Evento{Tipo:"pulso",RootCID:"rootcid:x",Contenido:"ok"});err!=nil{t.Fatal(err)};r2,err:=Nuevo(r.ruta);if err!=nil{t.Fatal(err)};if r2.UltimoHash()==""{t.Fatal("cadena vacía")}}
