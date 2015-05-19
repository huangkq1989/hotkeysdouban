package douban

// #define WIN32_LEAN_AND_MEAN
// #include <windows.h>
import "C"
import "fmt"

func RegisterHotKey(hotkeyNo int, mod int, key int) bool {
	if rc := int(C.RegisterHotKey(nil, C.int(hotkeyNo), C.UINT(mod), C.UINT(key))); rc == 0 {
		fmt.Printf("HotKey registered failed, maybe already register")
		return false
	}
	return true
}

func UnregisterHotKey(hotkeyNo int) {
	C.UnregisterHotKey(nil, C.int(hotkeyNo))
}

func ProcessHotKeyEvent(eventProcessor map[int]func(), stopSignal int) {
	var msg C.MSG
	for int(C.GetMessage(&msg, nil, 0, 0)) != 0 {
		if msg.message == C.WM_HOTKEY {
			if processor, ok := eventProcessor[int(msg.wParam)]; ok {
				go processor()
			}
			if int(msg.wParam) == stopSignal {
				break
			}
		}
		C.TranslateMessage(&msg)
		C.DispatchMessageA(&msg)
	}
}
