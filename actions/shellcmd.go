// +build !windows

package actions

func getShellCmd() (cmd, args string) {
	return "sh", "-c"
}
