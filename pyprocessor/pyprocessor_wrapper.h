#ifndef PYPROCESSOR_WRAPPER_H
#define PYPROCESSOR_WRAPPER_H

#include <Python.h>

static inline void initPython() {
    Py_Initialize();
}

static inline int pyIsInitialized() {
    return Py_IsInitialized();
}

static inline PyGILState_STATE pyGILStateEnsure() {
    return PyGILState_Ensure();
}

static inline void pyGILStateRelease(PyGILState_STATE state) {
    PyGILState_Release(state);
}

static inline PyObject* pyObjectGetAttrString(PyObject *o, const char *attr_name) {
    return PyObject_GetAttrString(o, attr_name);
}

static inline int pyUnicodeCheck(PyObject *o) {
    return PyUnicode_Check(o);
}

static inline const char* pyUnicodeAsUTF8(PyObject *unicode) {
    return PyUnicode_AsUTF8(unicode);
}

static inline PyObject* importModule(const char* moduleName) {
    return PyImport_ImportModule(moduleName);
}

static inline PyObject* getClass(PyObject* module, const char* className) {
    return PyObject_GetAttrString(module, className);
}

static inline PyObject* createInstance(PyObject* class) {
    return PyObject_CallObject(class, NULL);
}

static inline PyObject* pyTupleNew(Py_ssize_t size) {
    return PyTuple_New(size);
}

static inline int pyTupleSetItem(PyObject *p, Py_ssize_t pos, PyObject *o) {
    return PyTuple_SetItem(p, pos, o);
}

static inline PyObject* pyObjectCallObject(PyObject *callable, PyObject *args) {
    return PyObject_CallObject(callable, args);
}

#endif // PYPROCESSOR_WRAPPER_H
