package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/itchyny/gojq"
	lua "github.com/yuin/gopher-lua"
)

// LuaEngine manages Lua script execution
type LuaEngine struct {
	state *lua.LState
}

// Global Lua engine instance
var luaEngine *LuaEngine

// InitLuaEngine initializes the global Lua engine
func InitLuaEngine() error {
	luaEngine = &LuaEngine{}
	return luaEngine.Init()
}

// GetLuaEngine returns the global Lua engine
func GetLuaEngine() *LuaEngine {
	return luaEngine
}

// Init initializes the Lua state with ktray functions
func (e *LuaEngine) Init() error {
	e.state = lua.NewState()

	// Create scripts directory if it doesn't exist
	if err := os.MkdirAll(ScriptsDir(), 0755); err != nil {
		return fmt.Errorf("failed to create scripts directory: %w", err)
	}

	// Register ktray module
	e.registerKtrayModule()

	return nil
}

// Close cleans up the Lua state
func (e *LuaEngine) Close() {
	if e.state != nil {
		e.state.Close()
	}
}

// registerKtrayModule registers the ktray Lua module with exposed functions
func (e *LuaEngine) registerKtrayModule() {
	ktray := e.state.NewTable()

	// Clipboard functions
	e.state.SetField(ktray, "copy", e.state.NewFunction(luaCopy))
	e.state.SetField(ktray, "paste", e.state.NewFunction(luaPaste))

	// Browser/URL functions
	e.state.SetField(ktray, "open_url", e.state.NewFunction(luaOpenURL))

	// HTTP functions
	e.state.SetField(ktray, "http_get", e.state.NewFunction(luaHTTPGet))
	e.state.SetField(ktray, "http_post", e.state.NewFunction(luaHTTPPost))

	// Kerberos functions
	e.state.SetField(ktray, "get_token", e.state.NewFunction(luaGetToken))
	e.state.SetField(ktray, "get_spn", e.state.NewFunction(luaGetSPN))

	// Shell execution
	e.state.SetField(ktray, "exec", e.state.NewFunction(luaExec))
	e.state.SetField(ktray, "shell", e.state.NewFunction(luaShell))

	// Status/UI functions
	e.state.SetField(ktray, "set_status", e.state.NewFunction(luaSetStatus))
	e.state.SetField(ktray, "notify", e.state.NewFunction(luaNotify))

	// Utility functions
	e.state.SetField(ktray, "sleep", e.state.NewFunction(luaSleep))
	e.state.SetField(ktray, "env", e.state.NewFunction(luaEnv))
	e.state.SetField(ktray, "log", e.state.NewFunction(luaLog))
	e.state.SetField(ktray, "info", e.state.NewFunction(luaInfo))

	// Cache functions
	e.state.SetField(ktray, "cache_get", e.state.NewFunction(luaCacheGet))
	e.state.SetField(ktray, "cache_set", e.state.NewFunction(luaCacheSet))
	e.state.SetField(ktray, "cache_delete", e.state.NewFunction(luaCacheDelete))
	e.state.SetField(ktray, "cache_keys", e.state.NewFunction(luaCacheKeys))

	// Encoding functions
	e.state.SetField(ktray, "base64_encode", e.state.NewFunction(luaBase64Encode))
	e.state.SetField(ktray, "base64_decode", e.state.NewFunction(luaBase64Decode))
	e.state.SetField(ktray, "jwt_decode", e.state.NewFunction(luaJWTDecode))

	// JSON processing functions
	e.state.SetField(ktray, "jq", e.state.NewFunction(luaJQ))
	e.state.SetField(ktray, "json_parse", e.state.NewFunction(luaJSONParse))
	e.state.SetField(ktray, "json_encode", e.state.NewFunction(luaJSONEncode))

	// User input functions
	e.state.SetField(ktray, "prompt", e.state.NewFunction(luaPrompt))
	e.state.SetField(ktray, "prompt_secret", e.state.NewFunction(luaPromptSecret))
	e.state.SetField(ktray, "confirm", e.state.NewFunction(luaConfirm))

	// Register the module
	e.state.SetGlobal("ktray", ktray)
}

// RunScript executes a Lua script file with optional context variables
func (e *LuaEngine) RunScript(scriptName string, context map[string]string) (string, error) {
	scriptPath := ScriptPath(scriptName)

	// Check if script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return "", fmt.Errorf("script not found: %s", scriptPath)
	}

	// Create a new Lua state for this execution (isolation)
	L := lua.NewState()
	defer L.Close()

	// Copy ktray module to new state
	e.registerKtrayModuleToState(L)

	// Set context variables
	ctx := L.NewTable()
	for k, v := range context {
		L.SetField(ctx, k, lua.LString(v))
	}
	L.SetGlobal("ctx", ctx)

	// Create result variable
	L.SetGlobal("result", lua.LNil)

	// Execute script
	if err := L.DoFile(scriptPath); err != nil {
		return "", fmt.Errorf("script error: %w", err)
	}

	// Get result if set
	result := L.GetGlobal("result")
	if result != lua.LNil {
		return result.String(), nil
	}

	return "", nil
}

// registerKtrayModuleToState registers ktray functions to a specific Lua state
func (e *LuaEngine) registerKtrayModuleToState(L *lua.LState) {
	ktray := L.NewTable()

	L.SetField(ktray, "copy", L.NewFunction(luaCopy))
	L.SetField(ktray, "paste", L.NewFunction(luaPaste))
	L.SetField(ktray, "open_url", L.NewFunction(luaOpenURL))
	L.SetField(ktray, "http_get", L.NewFunction(luaHTTPGet))
	L.SetField(ktray, "http_post", L.NewFunction(luaHTTPPost))
	L.SetField(ktray, "get_token", L.NewFunction(luaGetToken))
	L.SetField(ktray, "get_spn", L.NewFunction(luaGetSPN))
	L.SetField(ktray, "exec", L.NewFunction(luaExec))
	L.SetField(ktray, "shell", L.NewFunction(luaShell))
	L.SetField(ktray, "set_status", L.NewFunction(luaSetStatus))
	L.SetField(ktray, "notify", L.NewFunction(luaNotify))
	L.SetField(ktray, "sleep", L.NewFunction(luaSleep))
	L.SetField(ktray, "env", L.NewFunction(luaEnv))
	L.SetField(ktray, "log", L.NewFunction(luaLog))
	L.SetField(ktray, "info", L.NewFunction(luaInfo))

	// Cache functions
	L.SetField(ktray, "cache_get", L.NewFunction(luaCacheGet))
	L.SetField(ktray, "cache_set", L.NewFunction(luaCacheSet))
	L.SetField(ktray, "cache_delete", L.NewFunction(luaCacheDelete))
	L.SetField(ktray, "cache_keys", L.NewFunction(luaCacheKeys))

	// Encoding functions
	L.SetField(ktray, "base64_encode", L.NewFunction(luaBase64Encode))
	L.SetField(ktray, "base64_decode", L.NewFunction(luaBase64Decode))
	L.SetField(ktray, "jwt_decode", L.NewFunction(luaJWTDecode))

	// JSON processing functions
	L.SetField(ktray, "jq", L.NewFunction(luaJQ))
	L.SetField(ktray, "json_parse", L.NewFunction(luaJSONParse))
	L.SetField(ktray, "json_encode", L.NewFunction(luaJSONEncode))

	// User input functions
	L.SetField(ktray, "prompt", L.NewFunction(luaPrompt))
	L.SetField(ktray, "prompt_secret", L.NewFunction(luaPromptSecret))
	L.SetField(ktray, "confirm", L.NewFunction(luaConfirm))

	L.SetGlobal("ktray", ktray)
}

// Lua function implementations

// luaCopy copies text to clipboard: ktray.copy(text)
func luaCopy(L *lua.LState) int {
	text := L.CheckString(1)
	err := copyToClipboard(text)
	if err != nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LTrue)
	return 1
}

// luaPaste gets text from clipboard: ktray.paste() -> string
func luaPaste(L *lua.LState) int {
	// Platform-specific paste implementation would go here
	// For now, return empty string
	L.Push(lua.LString(""))
	return 1
}

// luaOpenURL opens a URL in the browser: ktray.open_url(url)
func luaOpenURL(L *lua.LState) int {
	url := L.CheckString(1)
	err := openBrowser(url)
	if err != nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LTrue)
	return 1
}

// luaHTTPGet performs an HTTP GET request: ktray.http_get(url, headers, timeout_seconds, skip_verify) -> body, error
func luaHTTPGet(L *lua.LState) int {
	url := L.CheckString(1)
	headersTable := L.OptTable(2, nil)
	timeoutSec := L.OptNumber(3, 0)    // 0 means use default
	skipVerify := L.OptBool(4, false)  // Skip TLS certificate verification

	// Build headers map
	headers := make(map[string]string)
	if headersTable != nil {
		headersTable.ForEach(func(k, v lua.LValue) {
			headers[k.String()] = v.String()
		})
	}

	timeout := time.Duration(timeoutSec) * time.Second

	body, err := httpGet(url, headers, timeout, skipVerify)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LString(body))
	return 1
}

// luaHTTPPost performs an HTTP POST request: ktray.http_post(url, body, headers, timeout_seconds, skip_verify) -> response, error
func luaHTTPPost(L *lua.LState) int {
	url := L.CheckString(1)
	body := L.CheckString(2)
	headersTable := L.OptTable(3, nil)
	timeoutSec := L.OptNumber(4, 0)    // 0 means use default
	skipVerify := L.OptBool(5, false)  // Skip TLS certificate verification

	// Build headers map
	headers := make(map[string]string)
	if headersTable != nil {
		headersTable.ForEach(func(k, v lua.LValue) {
			headers[k.String()] = v.String()
		})
	}

	timeout := time.Duration(timeoutSec) * time.Second

	response, err := httpPost(url, body, headers, timeout, skipVerify)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LString(response))
	return 1
}

// luaGetToken gets a Kerberos token: ktray.get_token(spn_name) -> token, error
// If spn_name is provided, it looks up the SPN from config by name and requests a token for it
// If spn_name is not provided, it returns the current cached token
func luaGetToken(L *lua.LState) int {
	spnName := L.OptString(1, "")

	// If no SPN name provided, return the current token
	if spnName == "" {
		stateMutex.RLock()
		token := lastToken
		stateMutex.RUnlock()

		if token == "" {
			L.Push(lua.LNil)
			L.Push(lua.LString("no token available - select an SPN or pass SPN name"))
			return 2
		}
		L.Push(lua.LString(token))
		return 1
	}

	// Look up the SPN by name in config
	stateMutex.RLock()
	cfg := appConfig
	stateMutex.RUnlock()

	if cfg == nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("no configuration loaded"))
		return 2
	}

	// Find SPN entry by name (case-insensitive partial match)
	var spnValue string
	spnNameLower := strings.ToLower(spnName)
	for _, entry := range cfg.SPNs {
		if strings.ToLower(entry.Name) == spnNameLower ||
			strings.Contains(strings.ToLower(entry.Name), spnNameLower) {
			spnValue = entry.SPN
			break
		}
	}

	if spnValue == "" {
		L.Push(lua.LNil)
		L.Push(lua.LString("SPN not found: " + spnName))
		return 2
	}

	// Check cache first
	if cachedToken, found := GetCache().GetToken(spnValue); found {
		L.Push(lua.LString(cachedToken))
		return 1
	}

	// Request a new token for this SPN
	token, err := getServiceTicket(spnValue)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("failed to get token: " + err.Error()))
		return 2
	}

	// Encode and cache the token
	encodedToken := base64.StdEncoding.EncodeToString(token)
	GetCache().SetToken(spnValue, encodedToken, DefaultTokenExpiration)
	updateCacheMenu()

	L.Push(lua.LString(encodedToken))
	return 1
}

// luaGetSPN gets the current SPN: ktray.get_spn() -> spn
func luaGetSPN(L *lua.LState) int {
	stateMutex.RLock()
	spn := currentSPN
	stateMutex.RUnlock()

	L.Push(lua.LString(spn))
	return 1
}

// luaExec executes a command and returns output: ktray.exec(cmd, args...) -> output, error
func luaExec(L *lua.LState) int {
	cmdName := L.CheckString(1)
	var args []string
	for i := 2; i <= L.GetTop(); i++ {
		args = append(args, L.CheckString(i))
	}

	cmd := exec.Command(cmdName, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		L.Push(lua.LString(string(output)))
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LString(string(output)))
	return 1
}

// luaShell executes a shell command: ktray.shell(command) -> output, error
func luaShell(L *lua.LState) int {
	command := L.CheckString(1)

	var cmd *exec.Cmd
	if isWindows() {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		L.Push(lua.LString(string(output)))
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LString(string(output)))
	return 1
}

// luaSetStatus sets the status line: ktray.set_status(text)
func luaSetStatus(L *lua.LState) int {
	text := L.CheckString(1)
	mStatus.SetTitle(text)
	return 0
}

// luaNotify shows a notification (placeholder): ktray.notify(title, message)
func luaNotify(L *lua.LState) int {
	title := L.CheckString(1)
	message := L.OptString(2, "")

	// For now, just update status - could add proper notifications later
	if message != "" {
		mStatus.SetTitle(fmt.Sprintf("%s: %s", title, message))
	} else {
		mStatus.SetTitle(title)
	}
	return 0
}

// luaSleep pauses execution: ktray.sleep(milliseconds)
func luaSleep(L *lua.LState) int {
	ms := L.CheckInt(1)
	time.Sleep(time.Duration(ms) * time.Millisecond)
	return 0
}

// luaEnv gets an environment variable: ktray.env(name) -> value
func luaEnv(L *lua.LState) int {
	name := L.CheckString(1)
	value := os.Getenv(name)
	L.Push(lua.LString(value))
	return 1
}

// luaLog prints to debug log: ktray.log(message)
func luaLog(L *lua.LState) int {
	message := L.CheckString(1)
	LogDebug("[Lua] %s", message)
	return 0
}

// luaInfo prints to info log (always visible): ktray.info(message)
func luaInfo(L *lua.LState) int {
	message := L.CheckString(1)
	LogInfo("[Lua] %s", message)
	return 0
}

// Helper function to check if running on Windows
func isWindows() bool {
	return strings.Contains(strings.ToLower(os.Getenv("OS")), "windows")
}

// Cache functions

// luaCacheGet retrieves a value from cache: ktray.cache_get(key) -> value, found
func luaCacheGet(L *lua.LState) int {
	key := L.CheckString(1)

	value, found := GetCache().Get(key)
	if !found {
		L.Push(lua.LNil)
		L.Push(lua.LFalse)
		return 2
	}
	L.Push(lua.LString(value))
	L.Push(lua.LTrue)
	return 2
}

// luaCacheSet stores a value in cache: ktray.cache_set(key, value, ttl_seconds)
// ttl_seconds is optional, defaults to 10 minutes
func luaCacheSet(L *lua.LState) int {
	key := L.CheckString(1)
	value := L.CheckString(2)
	ttlSeconds := L.OptInt(3, 600) // Default: 10 minutes

	ttl := time.Duration(ttlSeconds) * time.Second
	GetCache().Set(key, value, ttl)

	// Update the cache menu to reflect the new entry
	updateCacheMenu()

	L.Push(lua.LTrue)
	return 1
}

// luaCacheDelete removes a value from cache: ktray.cache_delete(key)
func luaCacheDelete(L *lua.LState) int {
	key := L.CheckString(1)

	GetCache().Delete(key)

	// Update the cache menu to reflect the deletion
	updateCacheMenu()

	L.Push(lua.LTrue)
	return 1
}

// luaCacheKeys returns all cache keys: ktray.cache_keys() -> table
func luaCacheKeys(L *lua.LState) int {
	keys := GetCache().ListKeys()

	table := L.NewTable()
	for i, key := range keys {
		L.SetTable(table, lua.LNumber(i+1), lua.LString(key))
	}
	L.Push(table)
	return 1
}

// Encoding functions

// luaBase64Encode encodes a string to base64: ktray.base64_encode(data) -> encoded
func luaBase64Encode(L *lua.LState) int {
	data := L.CheckString(1)
	encoded := base64.StdEncoding.EncodeToString([]byte(data))
	L.Push(lua.LString(encoded))
	return 1
}

// luaBase64Decode decodes a base64 string: ktray.base64_decode(encoded) -> data, error
func luaBase64Decode(L *lua.LState) int {
	encoded := L.CheckString(1)

	// Try standard encoding first
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		// Try URL-safe encoding (used by JWTs)
		decoded, err = base64.RawURLEncoding.DecodeString(encoded)
		if err != nil {
			// Try with padding
			decoded, err = base64.URLEncoding.DecodeString(encoded)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
		}
	}

	L.Push(lua.LString(string(decoded)))
	return 1
}

// luaJWTDecode decodes a JWT without verification: ktray.jwt_decode(token) -> table, error
// Returns a table with 'header', 'payload', and 'signature' fields
// The header and payload are decoded JSON as Lua tables
func luaJWTDecode(L *lua.LState) int {
	token := L.CheckString(1)

	// Remove "Bearer " prefix if present
	token = strings.TrimPrefix(token, "Bearer ")
	token = strings.TrimPrefix(token, "bearer ")

	// Split JWT into parts
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		L.Push(lua.LNil)
		L.Push(lua.LString("invalid JWT format: expected 3 parts separated by '.'"))
		return 2
	}

	// Decode header
	headerJSON, err := base64URLDecode(parts[0])
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("failed to decode header: " + err.Error()))
		return 2
	}

	// Decode payload
	payloadJSON, err := base64URLDecode(parts[1])
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("failed to decode payload: " + err.Error()))
		return 2
	}

	// Parse header JSON
	var header map[string]interface{}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("failed to parse header JSON: " + err.Error()))
		return 2
	}

	// Parse payload JSON
	var payload map[string]interface{}
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("failed to parse payload JSON: " + err.Error()))
		return 2
	}

	// Create result table
	result := L.NewTable()

	// Convert header to Lua table
	headerTable := jsonToLuaTable(L, header)
	L.SetField(result, "header", headerTable)

	// Convert payload to Lua table
	payloadTable := jsonToLuaTable(L, payload)
	L.SetField(result, "payload", payloadTable)

	// Keep signature as base64 string
	L.SetField(result, "signature", lua.LString(parts[2]))

	// Also provide raw JSON strings for convenience
	L.SetField(result, "header_json", lua.LString(string(headerJSON)))
	L.SetField(result, "payload_json", lua.LString(string(payloadJSON)))

	L.Push(result)
	return 1
}

// base64URLDecode decodes a base64url encoded string (used by JWTs)
func base64URLDecode(s string) ([]byte, error) {
	// Add padding if needed
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}

	return base64.URLEncoding.DecodeString(s)
}

// jsonToLuaTable converts a Go map/slice to a Lua table
func jsonToLuaTable(L *lua.LState, v interface{}) lua.LValue {
	switch val := v.(type) {
	case nil:
		return lua.LNil
	case bool:
		return lua.LBool(val)
	case float64:
		// Check if it's an integer
		if val == float64(int64(val)) {
			return lua.LNumber(val)
		}
		return lua.LNumber(val)
	case string:
		return lua.LString(val)
	case []interface{}:
		table := L.NewTable()
		for i, item := range val {
			L.SetTable(table, lua.LNumber(i+1), jsonToLuaTable(L, item))
		}
		return table
	case map[string]interface{}:
		table := L.NewTable()
		for k, item := range val {
			L.SetField(table, k, jsonToLuaTable(L, item))
		}
		return table
	default:
		return lua.LString(fmt.Sprintf("%v", val))
	}
}

// JSON processing functions

// luaJQ executes a jq query on JSON data: ktray.jq(json_string, query) -> result, error
// Uses gojq for full jq compatibility
// Returns the result as a Lua value (table, string, number, boolean, or nil)
func luaJQ(L *lua.LState) int {
	jsonStr := L.CheckString(1)
	queryStr := L.CheckString(2)

	// Parse the jq query
	query, err := gojq.Parse(queryStr)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("jq parse error: " + err.Error()))
		return 2
	}

	// Parse the JSON input
	var input interface{}
	if err := json.Unmarshal([]byte(jsonStr), &input); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("JSON parse error: " + err.Error()))
		return 2
	}

	// Execute the query
	iter := query.Run(input)

	// Collect all results
	var results []interface{}
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, isErr := v.(error); isErr {
			L.Push(lua.LNil)
			L.Push(lua.LString("jq error: " + err.Error()))
			return 2
		}
		results = append(results, v)
	}

	// Return results
	if len(results) == 0 {
		L.Push(lua.LNil)
		return 1
	} else if len(results) == 1 {
		// Single result - return as appropriate Lua type
		L.Push(jsonToLuaTable(L, results[0]))
		return 1
	} else {
		// Multiple results - return as array
		L.Push(jsonToLuaTable(L, results))
		return 1
	}
}

// luaJSONParse parses a JSON string into a Lua table: ktray.json_parse(json_string) -> table, error
func luaJSONParse(L *lua.LState) int {
	jsonStr := L.CheckString(1)

	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("JSON parse error: " + err.Error()))
		return 2
	}

	L.Push(jsonToLuaTable(L, data))
	return 1
}

// luaJSONEncode encodes a Lua table to JSON string: ktray.json_encode(table, pretty) -> json_string, error
// If pretty is true, the output is indented for readability
func luaJSONEncode(L *lua.LState) int {
	value := L.CheckAny(1)
	pretty := L.OptBool(2, false)

	// Convert Lua value to Go interface
	goValue := luaToGoValue(L, value)

	var jsonBytes []byte
	var err error

	if pretty {
		jsonBytes, err = json.MarshalIndent(goValue, "", "  ")
	} else {
		jsonBytes, err = json.Marshal(goValue)
	}

	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("JSON encode error: " + err.Error()))
		return 2
	}

	L.Push(lua.LString(string(jsonBytes)))
	return 1
}

// luaToGoValue converts a Lua value to a Go interface{}
func luaToGoValue(L *lua.LState, lv lua.LValue) interface{} {
	switch v := lv.(type) {
	case *lua.LNilType:
		return nil
	case lua.LBool:
		return bool(v)
	case lua.LNumber:
		// Check if it's an integer
		f := float64(v)
		if f == float64(int64(f)) {
			return int64(f)
		}
		return f
	case lua.LString:
		return string(v)
	case *lua.LTable:
		// Determine if it's an array or object
		// Array: consecutive integer keys starting from 1
		// Object: string keys
		maxIndex := 0
		hasStringKeys := false

		v.ForEach(func(key, _ lua.LValue) {
			if keyNum, ok := key.(lua.LNumber); ok {
				idx := int(keyNum)
				if idx > maxIndex {
					maxIndex = idx
				}
			} else {
				hasStringKeys = true
			}
		})

		// If no string keys and has sequential integers, treat as array
		if !hasStringKeys && maxIndex > 0 {
			arr := make([]interface{}, maxIndex)
			v.ForEach(func(key, val lua.LValue) {
				if keyNum, ok := key.(lua.LNumber); ok {
					idx := int(keyNum) - 1 // Lua arrays are 1-indexed
					if idx >= 0 && idx < maxIndex {
						arr[idx] = luaToGoValue(L, val)
					}
				}
			})
			return arr
		}

		// Otherwise, treat as object
		obj := make(map[string]interface{})
		v.ForEach(func(key, val lua.LValue) {
			keyStr := ""
			switch k := key.(type) {
			case lua.LString:
				keyStr = string(k)
			case lua.LNumber:
				keyStr = fmt.Sprintf("%v", float64(k))
			default:
				keyStr = key.String()
			}
			obj[keyStr] = luaToGoValue(L, val)
		})
		return obj
	default:
		return lv.String()
	}
}

// User input functions

// luaPrompt shows a dialog asking for text input: ktray.prompt(title, message, default) -> value, ok
// Returns the entered text and true if OK was clicked, or empty string and false if cancelled
func luaPrompt(L *lua.LState) int {
	title := L.CheckString(1)
	message := L.OptString(2, "")
	defaultValue := L.OptString(3, "")

	value, ok := PromptForInput(title, message, defaultValue, false)
	L.Push(lua.LString(value))
	L.Push(lua.LBool(ok))
	return 2
}

// luaPromptSecret shows a dialog asking for secret input (masked): ktray.prompt_secret(title, message) -> value, ok
// Input is masked with bullets/dots. Returns the entered text and true if OK was clicked.
func luaPromptSecret(L *lua.LState) int {
	title := L.CheckString(1)
	message := L.OptString(2, "")

	value, ok := PromptForInput(title, message, "", true)
	L.Push(lua.LString(value))
	L.Push(lua.LBool(ok))
	return 2
}

// luaConfirm shows a Yes/No confirmation dialog: ktray.confirm(title, message) -> bool
// Returns true if Yes was clicked, false otherwise
func luaConfirm(L *lua.LState) int {
	title := L.CheckString(1)
	message := L.OptString(2, "")

	result := ConfirmDialog(title, message)
	L.Push(lua.LBool(result))
	return 1
}