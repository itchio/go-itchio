package itchio

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CamelCase(t *testing.T) {
	assert.EqualValues(t, "hello", camelcase("hello"))
	assert.EqualValues(t, "helloWorld", camelcase("hello_world"))
	assert.EqualValues(t, "pOsx", camelcase("p_osx"))
	assert.EqualValues(t, "shortText", camelcase("short_text"))
}

func Test_Camelify(t *testing.T) {
	assert.EqualValues(t, "hello", camelify("hello"))

	m1 := make(map[string]any)
	m1["short_text"] = "short text"
	m1["min_price"] = 1200
	m1["p_osx"] = true

	var users []any

	u1 := make(map[string]any)
	u1["full_name"] = "John Doe"
	users = append(users, u1)

	u2 := make(map[string]any)
	u2["full_name"] = "Jane Fischer"
	users = append(users, u2)

	m1["user_list"] = users

	mm := camelify(m1)
	assert.EqualValues(t, "short text", mm.(map[string]any)["shortText"])
	assert.EqualValues(t, 1200, mm.(map[string]any)["minPrice"])
	assert.EqualValues(t, true, mm.(map[string]any)["pOsx"])
	assert.EqualValues(t, "John Doe", mm.(map[string]any)["userList"].([]any)[0].(map[string]any)["fullName"])
	assert.EqualValues(t, "Jane Fischer", mm.(map[string]any)["userList"].([]any)[1].(map[string]any)["fullName"])

	// blacklisted keys are renamed but their contents are left alone
	m2 := map[string]any{
		"launch_targets": []any{
			map[string]any{"path": "game.x86_64", "linux_info": map[string]any{"arch": "amd64"}},
		},
	}
	targets := camelify(m2).(map[string]any)["launchTargets"].([]any)
	assert.EqualValues(t, "amd64", targets[0].(map[string]any)["linux_info"].(map[string]any)["arch"])
}

func Test_CamelifyBlacklist(t *testing.T) {
	m1 := make(map[string]any)
	m1["cookie"] = map[string]any{
		"itchio":     "session-value",
		"other_name": "raw-value",
	}

	mm := camelify(m1).(map[string]any)
	cookie := mm["cookie"].(map[string]any)
	assert.EqualValues(t, "session-value", cookie["itchio"])
	assert.EqualValues(t, "raw-value", cookie["other_name"])
}
