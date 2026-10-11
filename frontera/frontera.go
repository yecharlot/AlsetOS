// Package frontera: frontera de ejecución kernel ↔ espacio de usuario.
//
// Hito mínimo comprobable:
//   - El runtime del nodo (kernel lógico) ejecuta handlers de organismos de usuario
//     en un dominio aislado por recover + estado propio.
//   - Un pánico o error del organismo de usuario NO derriba el proceso del nodo:
//     se captura, se marca el organismo en Error y el kernel sigue vivo.
//
// Esto NO es paginación de hardware ni un hypervisor. Es el primer LEGO real de
// aislamiento a nivel de runtime Alset, medible con pruebas.
package frontera

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/yecharlot/AlsetOS/organismo"
)

// Handler es el cuerpo ejecutable de un organismo de usuario (app / ORGES).
type Handler func(o *organismo.Organismo) error

// Resultado de una invocación aislada.
type Resultado struct {
	OID     string
	Estado  organismo.Estado
	Error   string
	Panico  bool
	KernelVivo bool
}

// Runtime del lado kernel: supervisa organismos de usuario.
type Runtime struct {
	mu       sync.Mutex
	vivo     atomic.Bool
	ticks    atomic.Uint64
	fallos   atomic.Uint64
	usuarios map[string]*organismo.Organismo
}

// NuevoRuntime crea el supervisor de frontera (siempre nace vivo).
func NuevoRuntime() *Runtime {
	r := &Runtime{usuarios: make(map[string]*organismo.Organismo)}
	r.vivo.Store(true)
	return r
}

// Vivo reporta si el kernel lógico sigue operativo tras fallos de usuario.
func (r *Runtime) Vivo() bool { return r.vivo.Load() }

// Ticks cuenta invocaciones supervisadas.
func (r *Runtime) Ticks() uint64 { return r.ticks.Load() }

// FallosUsuario cuenta panicos/errores capturados en espacio usuario.
func (r *Runtime) FallosUsuario() uint64 { return r.fallos.Load() }

// Registrar incorpora un organismo de usuario (no fundacional).
func (r *Runtime) Registrar(o *organismo.Organismo) error {
	if o == nil {
		return fmt.Errorf("organismo nulo")
	}
	if o.Fundacional || o.Espacio == organismo.EspacioKernel {
		return fmt.Errorf("los fundacionales no se ejecutan como usuario en la frontera")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.usuarios[o.Nombre] = o
	return nil
}

// EjecutarUsuario corre el handler en aislamiento lógico.
// Precondiciones: organismo Listo/Ejecutando según contrato; definición verificada.
func (r *Runtime) EjecutarUsuario(nombre string, h Handler) (res Resultado) {
	res.KernelVivo = true
	res.OID = nombre
	r.ticks.Add(1)

	r.mu.Lock()
	o := r.usuarios[nombre]
	r.mu.Unlock()
	if o == nil {
		res.Error = "organismo no registrado en frontera"
		return res
	}
	res.OID = o.Nombre

	if err := o.RequisitosParaEjecutar(); err != nil {
		res.Error = err.Error()
		res.Estado = o.Estado
		return res
	}
	if o.Estado != organismo.Ejecutando {
		if err := o.Iniciar(); err != nil {
			res.Error = err.Error()
			res.Estado = o.Estado
			return res
		}
	}

	defer func() {
		if rec := recover(); rec != nil {
			r.fallos.Add(1)
			res.Panico = true
			res.Error = fmt.Sprintf("panico usuario: %v", rec)
			_ = o.Fallar(res.Error)
			res.Estado = o.Estado
			// Kernel permanece vivo: no re-panic.
			res.KernelVivo = r.Vivo()
		}
	}()

	if h == nil {
		res.Error = "handler nulo"
		_ = o.Fallar(res.Error)
		res.Estado = o.Estado
		r.fallos.Add(1)
		return res
	}

	if err := h(o); err != nil {
		r.fallos.Add(1)
		_ = o.Fallar(err.Error())
		res.Error = err.Error()
		res.Estado = o.Estado
		res.KernelVivo = r.Vivo()
		return res
	}

	res.Estado = o.Estado
	res.KernelVivo = r.Vivo()
	return res
}
