package douban

import "fmt"
import "syscall"
import "os/exec"

func PlaySongByMPG123(songUrl string, cmd **exec.Cmd) bool {
	*cmd = exec.Command("mpg123", "-q", songUrl)
	err := (*cmd).Start()
	if err != nil {
		fmt.Println("Fail to play song:", err, "")
		fmt.Println("Please make sure mpg123 has installed and `MPG123_DIR` in config.cfg has been set correctly, press 'Enter' to exit...")
		return false
	}
	err = (*cmd).Wait()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus := exitError.Sys().(syscall.WaitStatus)
			if waitStatus.ExitStatus() != 1 {
				fmt.Println("Fail to paly song:", err)
				return false
			}
		} else {
			fmt.Println("Fail to paly song:", err)
			return false
		}
	}
	return true
}
