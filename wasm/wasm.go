package wasm

import (
 "context"
 "fmt"
 "os"
 "time"
 "github.com/tetratelabs/wazero"
)

type Ejecutor struct{ Limite time.Duration }

func Nuevo()*Ejecutor{return &Ejecutor{Limite:5*time.Second}}

func(e *Ejecutor)Ejecutar(ctx context.Context,ruta,funcion string)(uint64,error){
 b,err:=os.ReadFile(ruta);if err!=nil{return 0,fmt.Errorf("leer WASM: %w",err)}
 if funcion==""{funcion="alset_main"}
 limite:=e.Limite;if limite<=0{limite=5*time.Second}
 ctx, cancel:=context.WithTimeout(ctx,limite);defer cancel()
 rt:=wazero.NewRuntime(ctx);defer rt.Close(ctx)
 mod,err:=rt.InstantiateWithConfig(ctx,b,wazero.NewModuleConfig().WithStartFunctions())
 if err!=nil{return 0,fmt.Errorf("instanciar WASM: %w",err)}
 fn:=mod.ExportedFunction(funcion);if fn==nil{return 0,fmt.Errorf("función WASM no exportada: %s",funcion)}
 valores,err:=fn.Call(ctx);if err!=nil{return 0,fmt.Errorf("ejecutar WASM: %w",err)}
 if len(valores)==0{return 0,nil};return valores[0],nil
}
