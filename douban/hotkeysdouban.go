package douban

// #define WIN32_LEAN_AND_MEAN
// #include <windows.h>
import "C"
import "fmt"
import "os/exec"
import "os"
import "bufio"
import "strings"
import "time"

const CONFIG_FILE = "config.cfg"

const (
	NEXT_SONG = iota
	NEXT_AND_BYE_SONG

	RATE_SONG
	UNRATE_SONG
	BYE_SONG

	CHANGE_CHANNEL
	STOP_HOTKEY_LOOP
	STOP_NOW

	HELP
)

type HotKey struct {
	key     int
	mode    int
	explain string
}

type MainLoop struct {
	cmd                 *exec.Cmd
	currentChannelId    string
	currentChannelName  string
	shouldChangeChannel bool
	currentSongId       string
	shouldExit          bool
	stopping            bool
	config              map[string]string
	hotKeys             map[int]HotKey
	hotKeyProcessor     map[int]func()
	douban              Douban
}

func NewMainLoop() *MainLoop {
	this := MainLoop{cmd: nil,
		currentChannelId:    "0",
		currentChannelName:  "私人兆赫",
		shouldChangeChannel: false,
		currentSongId:       "",
		shouldExit:          false,
		stopping:            false}
	this.config = make(map[string]string)
	this.hotKeys = make(map[int]HotKey)
	this.hotKeyProcessor = make(map[int]func())
	this.douban = Douban{cookie: nil,
		version: DOUBAN_APP_VERSION,
		appName: DOUBAN_APP_NAME,
		baseUrl: DOUBAN_BASE_URI}
	return &this
}

func (loop *MainLoop) initConfig(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line[0] == '#' {
			continue
		}
		data := strings.Split(line, "=")
		key := strings.Trim(data[0], " ")
		value := strings.Trim(data[1], " ")
		loop.config[key] = value
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func (loop *MainLoop) registerHotKey() bool {
	loop.hotKeys[NEXT_SONG] = HotKey{C.VK_F3, C.MOD_CONTROL | C.MOD_SHIFT, "Ctrl+Shift+F3\t=>\t next song"}
	loop.hotKeys[NEXT_AND_BYE_SONG] = HotKey{C.VK_F1, C.MOD_CONTROL | C.MOD_SHIFT, "Ctrl+Shift+F1\t=>\t next and bye song"}

	loop.hotKeys[RATE_SONG] = HotKey{C.VK_F5, C.MOD_CONTROL | C.MOD_SHIFT, "Ctrl+Shift+F5\t=>\t rate song"}
	loop.hotKeys[UNRATE_SONG] = HotKey{C.VK_F12, C.MOD_CONTROL | C.MOD_SHIFT, "Ctrl+Shift+F12\t=>\t unrate song"}
	loop.hotKeys[BYE_SONG] = HotKey{C.VK_F10, C.MOD_CONTROL | C.MOD_SHIFT, "Ctrl+Shift+F11\t=>\t bye song"}

	loop.hotKeys[CHANGE_CHANNEL] = HotKey{C.VK_F7, C.MOD_CONTROL | C.MOD_SHIFT, "Ctrl+Shift+F7\t=>\t change channel"}
	loop.hotKeys[STOP_HOTKEY_LOOP] = HotKey{C.VK_F11, C.MOD_CONTROL | C.MOD_SHIFT, "Ctrl+Shift+F11\t=>\t exit"}
	loop.hotKeys[STOP_NOW] = HotKey{C.VK_F4, C.MOD_CONTROL | C.MOD_SHIFT, "Ctrl+Shift+F4\t=>\t exit"}

	loop.hotKeys[HELP] = HotKey{C.VK_F9, C.MOD_CONTROL | C.MOD_SHIFT, "Ctrl+Shift+F9\t=>\t print help info"}

	for k, v := range loop.hotKeys {
		if !RegisterHotKey(k, v.mode, v.key) {
			return false
		}
	}
	return true
}

func (loop *MainLoop) printHotKeys() {
	for _, v := range loop.hotKeys {
		fmt.Println(v.explain)
	}
	fmt.Println()
}

func (loop *MainLoop) registerHotKeyProcessor() {
	loop.hotKeyProcessor[NEXT_SONG] = func() {
		loop.cmd.Process.Kill()
		fmt.Println(">>>Play next song")
	}
	loop.hotKeyProcessor[NEXT_AND_BYE_SONG] = func() {
		loop.cmd.Process.Kill()
		loop.douban.ByeSong(loop.currentChannelId, loop.currentSongId)
		fmt.Println(">>>Play next song and bye this song")
	}
	loop.hotKeyProcessor[RATE_SONG] = func() {
		loop.douban.RateSong(loop.currentChannelId, loop.currentSongId)
		fmt.Println(">>>Rate this song")
	}
	loop.hotKeyProcessor[UNRATE_SONG] = func() {
		loop.douban.UnrateSong(loop.currentChannelId, loop.currentSongId)
		fmt.Println(">>>Unrate this song")
	}
	loop.hotKeyProcessor[BYE_SONG] = func() {
		loop.douban.ByeSong(loop.currentChannelId, loop.currentSongId)
		fmt.Println(">>>Say google to this song")
	}
	loop.hotKeyProcessor[CHANGE_CHANNEL] = func() {
		for {
			channels := loop.douban.GetChannels()
			fmt.Println("<<<Select channel by input channel id here>: ")
			var input string
			fmt.Scanln(&input)
			if channel, ok := channels[input]; ok {
				loop.currentChannelId = input
				loop.currentChannelName = channel
				fmt.Printf(">>>Channel [%s] selected\n", channel)
				break
			} else {
				fmt.Println("<<<Invalid channel id, input again>: ")
			}
		}
		loop.shouldChangeChannel = true
		loop.cmd.Process.Kill()
	}
	loop.hotKeyProcessor[STOP_HOTKEY_LOOP] = func() {
		loop.cmd.Process.Kill()
		loop.shouldExit = true
		fmt.Println(">>>Exit")
	}
	loop.hotKeyProcessor[STOP_NOW] = func() {
		if !loop.stopping {
			loop.cmd.Process.Kill()
			fmt.Println(">>>Stop now")
		} else {
			fmt.Println(">>>Continue next song")
		}
		loop.stopping = !loop.stopping
	}
	loop.hotKeyProcessor[HELP] = func() {
		loop.printHotKeys()
	}
}

func (loop *MainLoop) Start() {

	if !loop.initConfig(CONFIG_FILE) || !loop.registerHotKey() {
		fmt.Println("Press 'Enter' to exit...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	}

	loop.registerHotKeyProcessor()

	messages := make(chan string)
	go func() {
		ProcessHotKeyEvent(loop.hotKeyProcessor, STOP_HOTKEY_LOOP)
		for k, _ := range loop.hotKeys {
			UnregisterHotKey(k)
		}
		messages <- "ping"
	}()

	fmt.Println("###############################################################################")
	fmt.Println("#          Logining, please wait for a second. Has a nice day :-)             #")
	fmt.Println("###############################################################################")
	fmt.Println()
	if !loop.douban.Signin(loop.config["USER"], loop.config["PASSWD"]) {
		fmt.Println("Sorry, you need try it again, press 'Enter' to exit...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	}
	os.Chdir(loop.config["MPG123_DIR"])

EXIT:
	for {
		songList := loop.douban.GetSongList(loop.currentChannelId)
		for i := 0; i < len(songList); i++ {
			fmt.Printf("Playing: [%s] [%s] => [%s]\n",
				songList[i].Artist, songList[i].Title,
				loop.currentChannelName)
			fmt.Println("===============================================================================")
			loop.currentSongId = songList[i].SongId
			if !PlaySongByMPG123(songList[i].Url, &loop.cmd) {
				bufio.NewReader(os.Stdin).ReadBytes('\n')
				return
			}
			if loop.shouldChangeChannel {
				loop.shouldChangeChannel = false
				break
			}
			for loop.stopping {
				time.Sleep(1 * time.Second)
			}
			if loop.shouldExit {
				break EXIT
			}
		}
	}
	<-messages
}
