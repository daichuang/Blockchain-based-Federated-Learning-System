
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sbinet/go-python"
)

//var (
//	pyLogModule *python.PyObject
//	pyTest      *python.PyObject
//)

func init() {
	err := python.Initialize()
	if err != nil {
		panic(err.Error())
	}
}
func main() {
	run()
}

func run() {
	dir, err := os.MkdirTemp("", "go-python-")
	if err != nil {
		panic(fmt.Errorf("could not create temp dir: %+v", err))
	}
	//defer os.RemoveAll(dir)

	fname := filepath.Join(dir, "data.csv")
	err = os.WriteFile(fname, []byte(data), 0644)
	if err != nil {
		panic(fmt.Errorf("could not create data file: %+v", err))
	}

	m := python.PyImport_ImportModule("sys")
	sysPath := m.GetAttrString("path")

	python.PyList_Insert(sysPath, 0, python.PyString_FromString(dir))
	fmt.Println(sysPath)

	err = os.WriteFile(filepath.Join(dir, "m.py"), []byte(module), 0644)
	if err != nil {
		panic(fmt.Errorf("could not create pandas module: %+v", err))
	}

	pyLogModule := python.PyImport_ImportModule("m")
	if pyLogModule == nil {
		panic(fmt.Errorf("could not import module"))
	}

	pyTest := pyLogModule.GetAttrString("test")
	if pyTest == nil {
		fmt.Println("Error importing function")
		panic(fmt.Errorf("could not import function"))
	}

	pyNumFeatures := pyTest.CallFunction(python.PyString_FromString(fname))
	if pyNumFeatures == nil {
		exc, val, tb := python.PyErr_Fetch()
		fmt.Printf("exc=%v\nval=%v\ntb=%v\n", exc, val, tb)
		panic(fmt.Errorf("could not call function"))
	}

	numFeatures := python.PyInt_AsLong(pyNumFeatures)
	fmt.Println(numFeatures)
}

const module = `
import pandas as pd

def test(fname):
	print("fname: %s, pandas: %s" % (fname, pd.__file__))
	return int(pd.read_csv(fname).shape[1])
`

const data = `1,2,3,4,hello
5,6,7,8,world
`

