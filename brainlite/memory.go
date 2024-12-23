package brainlite

import (
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"

	"github.com/Rovanta/rmodel/internal/errors"

	_ "github.com/mattn/go-sqlite3"
)

type BrainMemory struct {
	db             *sql.DB
	datasourceName string
	keepMemory     bool
}


func (m *BrainMemory)Init() error {
	db, err := sql.Open("sqlite3", m.datasourceName)
	if err != nil {
		return errors.Wrapf(err, "init memory failed")
	}
	m.db = db

	_, err = m.db.Exec(`CREATE TABLE IF NOT EXISTS memory (
		key INTEGER PRIMARY KEY,
		value JSON,
		type TEXT
	)`)
	if err != nil {
		return errors.Wrapf(err, "init memory table failed")
	}

	return nil
}

func (m *BrainMemory)Set(key, value any) error {
	var valueType string
	var valueJSON []byte
	var err error

	hashedKey, err := hashKey(key)
	if err != nil {
		return fmt.Errorf("Unable to hash key: %v", err)
	}

	switch value.(type) {
	case string:
		valueType = "string"
	case int, int32, int64, uint32:
		valueType = "int"
	case float64:
		valueType = "float"
	case bool:
		valueType = "bool"
	default:
		valueType = "json"
	}

	valueJSON, err = json.Marshal(value)
	if err != nil {
		return fmt.Errorf("Unable to serialize value: %v", err)
	}

	_, err = m.db.Exec("INSERT OR REPLACE INTO memory (key, value, type) VALUES (?, ?, ?)",
		hashedKey, valueJSON, valueType)
	if err != nil {
		return fmt.Errorf("Error while storing data: %v", err)
	}

	return nil
}

func (m *BrainMemory)Get(key any) (any, error) {
	hashedKey, err 	:= hashKey(key)
	if err != nil {
		return nil, fmt.Errorf("Unable to hash key: %v", err)
	}

	var valueJSON []byte
	var valueType string
	err = m.db.QueryRow("SELECT value, type FROM memory WHERE key = ?", hashedKey).Scan(&valueJSON, &valueType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("key not found '%v'", key)
		}
		return nil, fmt.Errorf("An error occurred while querying data: %v", err)
	}

	var value any
	switch valueType {
	case "string":
		err = json.Unmarshal(valueJSON, &value)
	case "int":
		var intValue int64
		err = json.Unmarshal(valueJSON, &intValue)
		if err != nil {
			return nil, err
		}
		switch {
		case intValue >= int64(math.MinInt32) && intValue <= int64(math.MaxInt32):
			value = int(intValue)
		default:
			value = intValue
		}
	case "float":
		var floatValue float64
		err = json.Unmarshal(valueJSON, &floatValue)
		value = floatValue
	case "bool":
		var boolValue bool
		err = json.Unmarshal(valueJSON, &boolValue)
		value = boolValue
	case "json":
		err = json.Unmarshal(valueJSON, &value)
	}

	if err != nil {
		return nil, fmt.Errorf("An error occurred while parsing the data: %v", err)
	}

	return value, nil
}

func (m *BrainMemory)Del(key any) error {
	hashedKey, err := hashKey(key)
	if err != nil {
		return fmt.Errorf("Unable to hash key: %v", err)
	}

	_, err = m.db.Exec("DELETE FROM memory WHERE key = ?", hashedKey)
	if err != nil {
		return fmt.Errorf("An error occurred while deleting data: %v", err)
	}

	return nil
}

func (m *BrainMemory)Clear() error {
	_, err := m.db.Exec("DELETE FROM memory")
	if err != nil {
		return fmt.Errorf("An error occurred while clearing data: %v", err)
	}

	return nil
}

func (m *BrainMemory)Close() error {
	if err := m.db.Close(); err != nil {
		return err
	}
	m.db = nil

	if !m.keepMemory {
		if err := os.Remove(m.datasourceName); err != nil {
			return fmt.Errorf("Error while deleting database file: %v", err)
		}
	}


	return nil
}	

func hashKey(key any) (int64, error) {
    switch key.(type) {
    case int, int32, int64, uint32, uint64, float64, string, []byte, byte:
    default:
        return 0, fmt.Errorf("unsupported key type %T", key)
    }

    h := sha256.New()
    h.Write([]byte(fmt.Sprintf("%v", key)))
    hashBytes := h.Sum(nil)
    return int64(binary.BigEndian.Uint64(hashBytes[:8])), nil
}