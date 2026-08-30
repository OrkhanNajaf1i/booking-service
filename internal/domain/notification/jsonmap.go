// File: internal/domain/notification/jsonmap.go
package notification

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSONMap – jsonb sutunu ile isleyen sərbəst formali payload.
// sqlx-in Scan/Value-ni ozu bilmesi ucun her iki interfeys realize olunur.
type JSONMap map[string]interface{}

// Value – DB-ye yazilarken JSON-a cevrilir. Bos map "{}" kimi gedir ki,
// sutun NOT NULL DEFAULT '{}' ile uyusun.
func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	encoded, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("jsonmap marshal failed: %w", err)
	}
	return encoded, nil
}

// Scan – DB-den oxunanda []byte ve ya string ola biler (driver-den asili).
func (m *JSONMap) Scan(src interface{}) error {
	if src == nil {
		*m = JSONMap{}
		return nil
	}

	var raw []byte
	switch typed := src.(type) {
	case []byte:
		raw = typed
	case string:
		raw = []byte(typed)
	default:
		return fmt.Errorf("jsonmap: desteklenmeyen tip %T", src)
	}

	if len(raw) == 0 {
		*m = JSONMap{}
		return nil
	}

	decoded := JSONMap{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return fmt.Errorf("jsonmap unmarshal failed: %w", err)
	}
	*m = decoded
	return nil
}
