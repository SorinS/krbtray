package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

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

// luaHTTPGet performs an HTTP GET request: ktray.http_get(url, headers) -> body, error
func luaHTTPGet(L *lua.LState) int {
	url := L.CheckString(1)
	headersTable := L.OptTable(2, nil)

	// Build headers map
	headers := make(map[string]string)
	if headersTable != nil {
		headersTable.ForEach(func(k, v lua.LValue) {
			headers[k.String()] = v.String()
		})
	}

	body, err := httpGet(url, headers)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LString(body))
	return 1
}

// luaHTTPPost performs an HTTP POST request: ktray.http_post(url, body, headers) -> response, error
func luaHTTPPost(L *lua.LState) int {
	url := L.CheckString(1)
	body := L.CheckString(2)
	headersTable := L.OptTable(3, nil)

	// Build headers map
	headers := make(map[string]string)
	if headersTable != nil {
		headersTable.ForEach(func(k, v lua.LValue) {
			headers[k.String()] = v.String()
		})
	}

	response, err := httpPost(url, body, headers)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LString(response))
	return 1
}

// luaGetToken gets the current Kerberos token: ktray.get_token() -> token, error
func luaGetToken(L *lua.LState) int {
	stateMutex.RLock()
	token := lastToken
	stateMutex.RUnlock()

	if token == "" {
		L.Push(lua.LNil)
		L.Push(lua.LString("no token available"))
		return 2
	}
	L.Push(lua.LString(token))
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
	if debugMode {
		fmt.Printf("[Lua] %s\n", message)
	}
	return 0
}

// Helper function to check if running on Windows
func isWindows() bool {
	return strings.Contains(strings.ToLower(os.Getenv("OS")), "windows")
}