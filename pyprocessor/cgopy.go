//go:build ignore
// +build ignore

package pyprocessor

/*
#cgo pkg-config: python3
#cgo CFLAGS: -I/usr/include/python3.9
#include "pyprocessor_wrapper.h"
*/
import "C"
import (
	"fmt"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/Rovanta/rmodel/processor"
)

var (
	initPythonOnce    sync.Once
	pythonInitialized bool
)

func ensurePythonInitialized() {
	initPythonOnce.Do(func() {
		C.initPython()
		pythonInitialized = true
	})
}

type PyProcessor struct {
	instance *C.PyObject
	mutex    sync.Mutex
}

func (p *PyProcessor) Process(ctx processor.BrainContext) error {
	done := make(chan error, 1)
	go func() {
		done <- p.processInternal(ctx)
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(10 * time.Second):
		return fmt.Errorf("Processing timeout")
	}
}

func (p *PyProcessor) processInternal(ctx processor.BrainContext) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	gstate := C.pyGILStateEnsure()
	defer C.pyGILStateRelease(gstate)

	methodName := C.CString("process")
	defer C.free(unsafe.Pointer(methodName))

	processMethod := C.pyObjectGetAttrString(p.instance, methodName)
	if processMethod == nil {
		return fmt.Errorf("Python class does not have a 'process' method")
	}
	defer C.Py_DecRef(processMethod)

	dbPath := fmt.Sprintf("%s.db", ctx.GetBrainID())
	cDbPath := C.CString(dbPath)
	defer C.free(unsafe.Pointer(cDbPath))
	pyDbPath := C.PyUnicode_FromString(cDbPath)
	defer C.Py_DecRef(pyDbPath)

	cBrainContextModule := C.CString("brain_context")
	defer C.free(unsafe.Pointer(cBrainContextModule))
	brainContextModule := C.PyImport_ImportModule(cBrainContextModule)
	if brainContextModule == nil {
		return fmt.Errorf("Unable to import brain_context module")
	}
	defer C.Py_DecRef(brainContextModule)

	cBrainContextClass := C.CString("BrainContext")
	defer C.free(unsafe.Pointer(cBrainContextClass))
	brainContextClass := C.pyObjectGetAttrString(brainContextModule, cBrainContextClass)
	if brainContextClass == nil {
		return fmt.Errorf("Unable to get BrainContext class")
	}
	defer C.Py_DecRef(brainContextClass)

	args := C.PyTuple_New(1)
	C.PyTuple_SetItem(args, 0, pyDbPath)
	C.Py_IncRef(pyDbPath)
	pyBrainContext := C.PyObject_CallObject(brainContextClass, args)
	C.Py_DecRef(args)
	if pyBrainContext == nil {
		return fmt.Errorf("Unable to create BrainContext instance")
	}
	defer C.Py_DecRef(pyBrainContext)


	args = C.PyTuple_New(1)
	C.PyTuple_SetItem(args, 0, pyBrainContext)
	C.Py_IncRef(pyBrainContext)

	result := C.PyObject_CallObject(processMethod, args)
	C.Py_DecRef(args)

	if result == nil {
		return fmt.Errorf("Error while calling Python 'process' method")
	}
	defer C.Py_DecRef(result)

	if C.pyUnicodeCheck(result) != 0 {
		cstr := C.PyUnicode_AsUTF8(result)
		gostr := C.GoString(cstr)
		fmt.Printf("Python method returns: %s\n", gostr)
	} else {
		fmt.Println("Python Method does not return a string")
	}

	return nil
}

func (p *PyProcessor) Clone() processor.Processor {
	fmt.Println("Cloning PyProcessor")
	p.mutex.Lock()
	defer p.mutex.Unlock()
	C.Py_IncRef(p.instance)
	return &PyProcessor{instance: p.instance}
}

func (p *PyProcessor) Close() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.instance != nil {
		C.Py_DecRef(p.instance)
		p.instance = nil
	}
}

func  DeprecatedLoadPythonProcessor(pyCodePath, moduleName, processorClassName string) processor.Processor {
	ensurePythonInitialized()
	if !pythonInitialized {
		panic("Failed to initialize Python interpreter")
	}

	fmt.Println("Initializing Python interpreter")
	C.initPython()
	if C.pyIsInitialized() == 0 {
		panic("Failed to initialize Python interpreter")
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	gstate := C.pyGILStateEnsure()
	defer C.pyGILStateRelease(gstate)

	fmt.Println("Adding path to sys.path:", pyCodePath)
	cPath := C.CString("path")
	defer C.free(unsafe.Pointer(cPath))
	sysPath := C.PySys_GetObject(cPath)

	cPyCodePath := C.CString(pyCodePath)
	defer C.free(unsafe.Pointer(cPyCodePath))
	pyCodePathObj := C.PyUnicode_FromString(cPyCodePath)
	defer C.Py_DecRef(pyCodePathObj)

	C.PyList_Append(sysPath, pyCodePathObj)

	fmt.Println("Importing Python module:", moduleName)
	modName := C.CString(moduleName)
	defer C.free(unsafe.Pointer(modName))
	module := C.importModule(modName)
	if module == nil {
		pyErr := C.PyErr_Occurred()
		if pyErr != nil {
			C.PyErr_Print()
		}
		panic(fmt.Sprintf("Error importing Python module: %s", moduleName))
	}
	defer C.Py_DecRef(module)

	fmt.Printf("Module: %v\n", module)

	fmt.Println("Getting Python class:", processorClassName)
	classNameC := C.CString(processorClassName)
	defer C.free(unsafe.Pointer(classNameC))
	class := C.getClass(module, classNameC)
	if class == nil {
		panic(fmt.Sprintf("Error getting Python class: %s", processorClassName))
	}
	defer C.Py_DecRef(class)

	fmt.Printf("Class: %v\n", class)

	fmt.Println("Instantiating Python class:", processorClassName)
	instance := C.createInstance(class)
	if instance == nil {
		panic(fmt.Sprintf("Error instantiating Python class: %s", processorClassName))
	}

	fmt.Printf("Instance: %v\n", instance)
	fmt.Println("Python processor loaded successfully")
	proc := &PyProcessor{instance: instance, mutex: sync.Mutex{}}
	runtime.SetFinalizer(proc, (*PyProcessor).Close)
	return proc
}
