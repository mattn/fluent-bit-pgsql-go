package pgsql

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"

	fluentbit "github.com/fluent/fluent-bit-go/output"
	"github.com/sirupsen/logrus"
)

// OutputPlugin is the CloudWatch Logs Fluent Bit output plugin
type OutputPlugin struct {
	PluginInstanceID int
	db               *sql.DB
	table            string
	events           []map[string]interface{}
}

// OutputPluginConfig is the input information used by NewOutputPlugin to create a new OutputPlugin
type OutputPluginConfig struct {
	PluginInstanceID int
	DSN              string
	Table            string
}

type Event struct {
	TS     time.Time
	Record map[interface{}]interface{}
	Tag    string
}

// Validate checks the configuration input for an OutputPlugin instances
func (config OutputPluginConfig) Validate() error {
	errorStr := "%s is a required parameter"
	if config.DSN == "" {
		return fmt.Errorf(errorStr, "dsn")
	}
	return nil
}

// NewOutputPlugin creates a OutputPlugin object
func NewOutputPlugin(config OutputPluginConfig) (*OutputPlugin, error) {
	logrus.Debugf("[pgsql-go %d] Initializing NewOutputPlugin", config.PluginInstanceID)

	db, err := sql.Open("postgres", config.DSN)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(fmt.Sprintf(`create table if not exists %s (data json);`, config.Table))
	if err != nil {
		return nil, err
	}

	return &OutputPlugin{
		db:     db,
		events: nil,
		table:  config.Table,
	}, nil
}

func (output *OutputPlugin) AddEvent(e *Event) int {
	m := map[string]interface{}{}
	for k, v := range e.Record {
		m[fmt.Sprint(k)] = v
	}
	output.events = append(output.events, m)
	return fluentbit.FLB_OK
}

// Flush sends the current buffer of records.
func (output *OutputPlugin) Flush() error {
	logrus.Debugf("[pgsql-go %d] Flush() Called", output.PluginInstanceID)

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(output.events)
	if err != nil {
		return err
	}
	tx, err := output.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := fmt.Sprintf(`with json_pack as (select json_array_elements('%s') as json_pack )INSERT INTO %s select * from json_pack;`, buf.String(), output.table)
	_, err = tx.Exec(query)
	if err != nil {
		log.Println(err)
	} else {
		tx.Commit()
	}
	output.events = nil
	return nil
}
