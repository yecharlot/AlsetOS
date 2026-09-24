package recursos
import("fmt";"strings";"github.com/yecharlot/AlsetOS/organismo")
type Permiso struct{Name string}
func Requerir(e *organismo.Organismo,n string)error{n=strings.TrimSpace(n);if n==""{return fmt.Errorf("recurso vacío")};if e.Capacidad["recurso:"+n]||e.Capacidad[n]{return nil};return fmt.Errorf("recurso no autorizado: %s",n)}
func Conceder(e *organismo.Organismo,n string){e.Capacidad["recurso:"+strings.TrimSpace(n)]=true}
