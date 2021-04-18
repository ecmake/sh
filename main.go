package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/lukaspj/ecmake/pkg/buildfile"
	"os"
)

func (m *Module) GetMethods() []string {
	return []string{
		"Run",
		"RunV",
		"RunWith",
		"RunWithV",
		"Output",
		"OutputWith",
		"Exec",
	}
}

func (m *Module) Invoke(method string, args []interface{}) interface{} {
	m.logger.Info("Invoke called", "method", method, "args", args)
	switch method {
	case "Run":
		c, a, err := getCmdAndArgs(args)
		if err != nil {
			m.logger.Error("failed to parse args in call to method", "method", method, "err", err)
			return err
		}
		err = m.Run(c, a...)
		if err != nil {
			m.logger.Error("method failed", "method", method, "err", err)
			return map[string]interface{} {
				"error": err.Error(),
				"code": m.ExitStatus(err),
			}
		}
		return map[string]interface{} {
			"code": m.ExitStatus(err),
		}
	case "RunV":
		c, a, err := getCmdAndArgs(args)
		if err != nil {
			m.logger.Error("failed to parse args in call to method", "method", method, "err", err)
			return err
		}
		err = m.RunV(c, a...)
		if err != nil {
			m.logger.Error("method failed", "method", method, "err", err)
			return map[string]interface{} {
				"error": err.Error(),
				"code": m.ExitStatus(err),
			}
		}
		return map[string]interface{} {
			"code": m.ExitStatus(err),
		}
	case "RunWith":
		env := args[0].(map[string]string)
		c, a, err := getCmdAndArgs(args[1:])
		if err != nil {
			m.logger.Error("failed to parse args in call to method", "method", method, "err", err)
			return err
		}
		err = m.RunWith(env, c, a...)
		if err != nil {
			m.logger.Error("method failed", "method", method, "err", err)
			return map[string]interface{} {
				"error": err.Error(),
				"code": m.ExitStatus(err),
			}
		}
		return map[string]interface{} {
			"code": m.ExitStatus(err),
		}
	case "RunWithV":
		env := args[0].(map[string]string)
		c, a, err := getCmdAndArgs(args[1:])
		if err != nil {
			m.logger.Error("failed to parse args in call to method", "method", method, "err", err)
			return err
		}
		err = m.RunWithV(env, c, a...)
		if err != nil {
			m.logger.Error("method failed", "method", method, "err", err)
			return map[string]interface{} {
				"error": err.Error(),
				"code": m.ExitStatus(err),
			}
		}
		return map[string]interface{} {
			"code": m.ExitStatus(err),
		}
	case "Output":
		c, a, err := getCmdAndArgs(args[1:])
		if err != nil {
			m.logger.Error("failed to parse args in call to method", "method", method, "err", err)
			return err
		}
		output, err := m.Output(c, a...)
		if err != nil {
			m.logger.Error("method failed", "method", method, "err", err)
			return map[string]interface{} {
				"error": err.Error(),
				"code": m.ExitStatus(err),
			}
		}
		return map[string]interface{} {
			"output": output,
			"code": m.ExitStatus(err),
		}
	case "OutputWith":
		env := args[0].(map[string]string)
		c, a, err := getCmdAndArgs(args[1:])
		if err != nil {
			m.logger.Error("failed to parse args in call to method", "method", method, "err", err)
			return err
		}
		output, err := m.OutputWith(env, c, a...)
		if err != nil {
			m.logger.Error("method failed", "method", method, "err", err)
			return map[string]interface{} {
				"error": err.Error(),
				"code": m.ExitStatus(err),
			}
		}
		return map[string]interface{} {
			"output": output,
			"code": m.ExitStatus(err),
		}
	case "Exec":
		env := args[0].(map[string]string)
		c, a, err := getCmdAndArgs(args[1:])
		if err != nil {
			m.logger.Error("failed to parse args in call to method", "method", method, "err", err)
			return map[string]string {
				"error": err.Error(),
			}
		}
		stderr := bytes.Buffer{}
		stdout := bytes.Buffer{}
		ran, err := m.Exec(env, &stdout, &stderr, c, a...)
		if err != nil {
			m.logger.Error("method failed", "method", method, "err", err)
			return map[string]interface{} {
				"error": err.Error(),
				"code": m.ExitStatus(err),
			}
		}
		return map[string]interface{} {
			"ran": ran,
			"stdout": stdout.String(),
			"stderr": stderr.String(),
			"code": m.ExitStatus(err),
		}
	default:
		m.logger.Error("unknown method", "method", method)
		return map[string]interface{} {
			"error": "unknown method: " + method,
		}
	}
}

func getCmdAndArgs(arr []interface{}) (cmd string, args []string, err error) {
	if len(arr) == 0 {
		return "", nil, errors.New("not enough arguments")
	}
	var ok bool
	cmd, ok = arr[0].(string)
	if !ok {
		return "", nil, errors.New("cmd argument was not a string")
	}
	if len(arr) == 1 {
		return
	}
	for i, a := range arr[1:] {
		args = append(args, a.(string))
		if !ok {
			return "", nil, errors.New(fmt.Sprintf("argument %d (%v) to %s was not a string", i, a, cmd))
		}
	}
	return cmd, args, nil
}

var _ buildfile.Module = &Module{}


func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:            "ShModule",
		Level:           hclog.Trace,
		Output:          os.Stderr,
		JSONFormat:      true,
	})

	module := &Module{
		logger: logger,
	}

	gob.Register(map[string]interface{}{})

	var pluginMap = map[string]plugin.Plugin{
		"module": &buildfile.ModulePlugin{Impl: module},
	}

	logger.Debug("Sh Module initialized")

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: buildfile.HandshakeConfig,
		Plugins: pluginMap,
	})
}

