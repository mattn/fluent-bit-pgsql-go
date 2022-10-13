package main

import (
	"C"
	"fmt"
	"time"
	"unsafe"

	"github.com/fluent/fluent-bit-go/output"
	"github.com/mattn/fluent-bit-pgsql-go/pgsql"

	"github.com/sirupsen/logrus"
)

var (
	pluginInstances []*pgsql.OutputPlugin
)

func addPluginInstance(ctx unsafe.Pointer) error {
	pluginID := len(pluginInstances)

	config := getConfiguration(ctx, pluginID)
	err := config.Validate()
	if err != nil {
		return err
	}

	instance, err := pgsql.NewOutputPlugin(config)
	if err != nil {
		return err
	}

	output.FLBPluginSetContext(ctx, pluginID)
	pluginInstances = append(pluginInstances, instance)

	logrus.SetLevel(logrus.DebugLevel)
	return nil
}

func getPluginInstance(ctx unsafe.Pointer) *pgsql.OutputPlugin {
	pluginID := output.FLBPluginGetContext(ctx).(int)
	return pluginInstances[pluginID]
}

//export FLBPluginRegister
func FLBPluginRegister(ctx unsafe.Pointer) int {
	return output.FLBPluginRegister(ctx, "pgsql-go", "PostgreSQL Logs Fluent Bit Plugin!")
}

func getConfiguration(ctx unsafe.Pointer, pluginID int) pgsql.OutputPluginConfig {
	config := pgsql.OutputPluginConfig{}
	config.PluginInstanceID = pluginID

	config.DSN = output.FLBPluginConfigKey(ctx, "dsn")
	logrus.Infof("[pgsql-go %d] plugin parameter dsn = '%s'", pluginID, config.DSN)

	config.Table = output.FLBPluginConfigKey(ctx, "table")
	logrus.Infof("[pgsql-go %d] plugin parameter table = '%s'", pluginID, config.Table)

	return config
}

//export FLBPluginInit
func FLBPluginInit(ctx unsafe.Pointer) int {
	logrus.Debug("A new higher performance PostgreSQL Logs plugin has been released; ")

	err := addPluginInstance(ctx)
	if err != nil {
		logrus.Error(err)
		return output.FLB_ERROR
	}
	return output.FLB_OK
}

//export FLBPluginFlushCtx
func FLBPluginFlushCtx(ctx, data unsafe.Pointer, length C.int, tag *C.char) int {
	var count int
	var ret int
	var ts interface{}
	var record map[interface{}]interface{}

	// Create Fluent Bit decoder
	dec := output.NewDecoder(data, int(length))

	plugin := getPluginInstance(ctx)

	fluentTag := C.GoString(tag)
	logrus.Debugf("[pgsql-go %d] Found logs with tag: %s", plugin.PluginInstanceID, fluentTag)

	for {
		// Extract Record
		ret, ts, record = output.GetRecord(dec)
		if ret != 0 {
			break
		}

		var timestamp time.Time
		switch tts := ts.(type) {
		case output.FLBTime:
			timestamp = tts.Time
		case uint64:
			timestamp = time.Unix(int64(tts), 0)
		default:
			timestamp = time.Now()
		}

		retCode := plugin.AddEvent(&pgsql.Event{Tag: fluentTag, Record: record, TS: timestamp})
		if retCode != output.FLB_OK {
			return retCode
		}
		count++
	}
	err := plugin.Flush()
	if err != nil {
		fmt.Println(err)
		// TODO: Better error handling
		return output.FLB_RETRY
	}

	logrus.Debugf("[pgsql-go %d] Processed %d events", plugin.PluginInstanceID, count)

	// Return options:
	//
	// output.FLB_OK    = data have been processed.
	// output.FLB_ERROR = unrecoverable error, do not try this again.
	// output.FLB_RETRY = retry to flush later.
	return output.FLB_OK
}

//export FLBPluginExit
func FLBPluginExit() int {
	return output.FLB_OK
}

func main() {
}
