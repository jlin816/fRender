package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

//const masterAddress = "hello_world"
const blenderPath = "blender"

type Friend struct {
	me int
}

func initFriend() (*Friend) {
	friend := Friend{}
	friend.registerWithMaster()
	go friend.listenOnSocket()

	return &friend //help
}

func (fr *Friend) listenOnSocket() {
	// call receiveJob here somewhere??
}

func (fr *Friend) registerWithMaster() {

}

func (fr *Friend) receiveJob() {

}

func (fr *Friend) renderFrames(file string, start_frame int, end_frame int) {
	// blender -b bob_lamp_update_export.blend -s 0 -e 100 -o render_files/frame_##### -a

	binary, lookErr := exec.LookPath(blenderPath)
	if lookErr != nil {
		panic(lookErr)
	}

	output_folder := fmt.Sprintf("%v_frames/frame_#####", file)

	args := []string{
		blenderPath,
		fmt.Sprintf("-b %v", file),
		"-F PNG",
		fmt.Sprintf("-s %v", start_frame),
		fmt.Sprintf("-e %v", end_frame),
		fmt.Sprintf("-o %v", output_folder),
		"-a",
	}

	env := os.Environ()

	execErr := syscall.Exec(binary, args, env)
	if execErr != nil {
		panic(execErr)
	}
}
