package pyprocessor

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/Rovanta/rmodel/processor"
)

func LoadPythonProcessor(pyCodePath, moduleName, processorClassName string, constructorArgs map[string]interface{}) *ExecPyProcessor {
	return &ExecPyProcessor{
		pyCodePath:         pyCodePath,
		moduleName:         moduleName,
		processorClassName: processorClassName,
		constructorArgs:    constructorArgs,
		scriptPath:         "temp_exec_script.py",
		pythonCmd:          "/usr/bin/python3",
	}
}

type ExecPyProcessor struct {
	pyCodePath         string
	moduleName         string
	processorClassName string
	constructorArgs    map[string]interface{}
	scriptPath         string
	pythonCmd          string
}

func (p *ExecPyProcessor) Process(ctx processor.BrainContext) error {
	if l := ctx.GetCurrentNeuronLabels(); l != nil {
		if v, ok := l["python_cmd"]; ok {
			p.pythonCmd = v
		}
	}

	err := p.createTempPythonScript()
	if err != nil {
		return fmt.Errorf("create temp python script failed: %s", err)
	}
	defer os.Remove(p.scriptPath)

	return p.execPythonScript(fmt.Sprintf("%s.db", ctx.GetBrainID()))
}

func (p *ExecPyProcessor) Clone() processor.Processor {
	return &ExecPyProcessor{
		pyCodePath:         p.pyCodePath,
		moduleName:         p.moduleName,
		processorClassName: p.processorClassName,
		constructorArgs:    p.constructorArgs,
		scriptPath:         p.scriptPath,
		pythonCmd:          p.pythonCmd,
	}
}

func (p *ExecPyProcessor) createTempPythonScript() error {
	importPath := strings.ReplaceAll(p.pyCodePath, string(os.PathSeparator), ".")
	importPath = strings.TrimSuffix(importPath, ".")
	importPath = strings.TrimPrefix(importPath, ".")

	content := fmt.Sprintf(`
import sys
import json
import os

from %s.%s import %s
from rModel import BrainContext

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python script.py <db_path> <params_json>")
        sys.exit(1)

    db_path = sys.argv[1]
    params_json = sys.argv[2]

    params = json.loads(params_json)

    processor = %s(**params)
    ctx = BrainContext(db_path)
    processor.process(ctx)
`, importPath, p.moduleName, p.processorClassName, p.processorClassName)

	return os.WriteFile(p.scriptPath, []byte(content), 0644)
}

func (p *ExecPyProcessor) execPythonScript(sqliteDBPath string) error {
	paramsJSON, err := json.Marshal(p.constructorArgs)
	if err != nil {
		return fmt.Errorf("Parameter serialization error: %s", err)
	}

	cmd := exec.Command(p.pythonCmd, p.scriptPath, sqliteDBPath, string(paramsJSON))

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("Unable to obtain standard output pipe: %s", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("Unable to obtain standard error output pipe: %s", err)
	}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("Failed to start Python process: %s", err)
	}

	reader := io.MultiReader(stdoutPipe, stderrPipe)

	fmt.Println("python processor:")
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fmt.Println("  " + scanner.Text())
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("Python process execution error: %s", err)
	}

	return nil
}
