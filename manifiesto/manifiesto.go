package manifiesto

import (
 "encoding/json"
 "fmt"
 "os"
 "github.com/yecharlot/AlsetOS/zyrion"
)

type ConfiguracionZyrion struct{Estado string `json:"estado"`;Estrategia string `json:"estrategia,omitempty"`}
type ModuloWASM struct{Ruta string `json:"ruta"`;Funcion string `json:"funcion,omitempty"`}
type Manifiesto struct{
 Version string `json:"version,omitempty"`
 Nombre string `json:"nombre"`
 LispAI []string `json:"lispai,omitempty"`
 Zyrion ConfiguracionZyrion `json:"zyrion"`
 Capacidades []string `json:"capacidades,omitempty"`
 Recursos []string `json:"recursos,omitempty"`
 Genes []string `json:"genes,omitempty"`
 WASM []ModuloWASM `json:"wasm,omitempty"`
 Memoria string `json:"memoria,omitempty"`
 Eventos string `json:"eventos,omitempty"`
 Agentes []string `json:"agentes,omitempty"`
}
func Cargar(ruta string)(Manifiesto,error){b,err:=os.ReadFile(ruta);if err!=nil{return Manifiesto{},fmt.Errorf("leer manifiesto: %w",err)};var d Manifiesto;if err:=json.Unmarshal(b,&d);err!=nil{return Manifiesto{},fmt.Errorf("analizar manifiesto: %w",err)};if err:=d.Validar();err!=nil{return Manifiesto{},err};return d,nil}
func(d Manifiesto)Validar()error{if d.Nombre==""{return fmt.Errorf("el manifiesto requiere nombre")};if d.Zyrion.Estado==""{return fmt.Errorf("el manifiesto requiere estado Zyrion")};if _,err:=d.EstadoZyrion();err!=nil{return err};for _,m:=range d.WASM{if m.Ruta==""{return fmt.Errorf("módulo WASM sin ruta")}};return nil}
func(d Manifiesto)EstadoZyrion()(zyrion.Valor,error){switch d.Zyrion.Estado{case "no":return zyrion.No,nil;case "si":return zyrion.Si,nil;case "incierto":return zyrion.Incierto,nil;default:return zyrion.Incierto,fmt.Errorf("estado Zyrion desconocido: %q",d.Zyrion.Estado)}}
func(d Manifiesto)Canonico()([]byte,error){if err:=d.Validar();err!=nil{return nil,err};b,err:=json.Marshal(d);if err!=nil{return nil,fmt.Errorf("serializar manifiesto canónico: %w",err)};return b,nil}
