package douban

import "fmt"
import "syscall"
import "os/exec"

func PlaySongByMPG123(songUrl string, cmd **exec.Cmd) bool {
	*cmd = exec.Command("mpg123", "-q", songUrl)
	err := (*cmd).Start()
	if err != nil {
		fmt.Println("Fail to play song:", err, "")
		return false
	}
	err = (*cmd).Wait()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus := exitError.Sys().(syscall.WaitStatus)
			fmt.Sprintln("Fail to paly song:", err, waitStatus.ExitStatus())
			return false
		}
	}
	return true
}
